// composables/useAdexAudio.ts
//
// V2 audio cue system. Uses Howler (via @vueuse/sound) as the primary
// playback layer with a Web-Audio synthesizer fallback when WAV assets
// fail to load or when the user opts into a synth-only soundpack.
//
// Vocabulary: 13 canonical eDEX cues (`keyboard, panels, theme, stdin,
// stdout, granted, denied, error, info, alarm, expand, folder, scan`),
// plus aliases that route the existing aDex 12-event vocabulary
// (`button_click`, `notification`, ...) through the same cues so legacy
// callers don't need to migrate at once.
//
// Spec: openspec/changes/edex-parity-and-uplift/specs/audio-cue-system/spec.md

import { ref, computed, readonly, type ComputedRef, type Ref } from "vue";
import { useStorage } from "@vueuse/core";

/* Canonical cue vocabulary -----------------------------------------------
 *
 * Layout: 13 legacy eDex cues + 3 new aDex cues (keypress, destructive,
 * click). The new ones map to the asset files in `app/assets/audio/adex/`
 * and are the recommended cues for terminal-input audio, destructive
 * confirmations, and generic UI click feedback respectively.
 */
export const EDEX_CUES = [
  // Legacy eDex 13-cue vocabulary — backed by WAVs in assets/audio/adex/.
  "alarm",
  "denied",
  "error",
  "expand",
  "folder",
  "granted",
  "info",
  "keyboard",
  "panels",
  "scan",
  "stdin",
  "stdout",
  "theme",
  // New aDex cues — backed by WAVs in assets/audio/adex/.
  "keypress",     // mechanical-keyboard click for terminal typing
  "destructive",  // handgun-click (chambering) when destructive dialog OPENS
  "gunshot",      // gunshot bang when destructive dialog is CONFIRMED
  "click",        // click-tone for routine UI interactions
] as const;
export type EdexCue = (typeof EDEX_CUES)[number];

/* Cue category map -------------------------------------------------------
 *
 * Each cue belongs to one of four categories — that's what the per-
 * category toggle in Settings → Audio gates. Cues default to `system`
 * if not listed (boot sweeps, alarms, etc. — covered by the master
 * audio.enabled flag only).
 */
export type CueCategory = "keyboard" | "destructive" | "interface" | "system";

const CUE_CATEGORY: Partial<Record<EdexCue, CueCategory>> = {
  // Keyboard / terminal input
  keypress: "keyboard",
  keyboard: "keyboard",
  stdin:    "keyboard",
  // Destructive actions (kill process, quit app, delete file)
  destructive: "destructive",
  gunshot:     "destructive",
  denied:      "destructive",
  alarm:       "destructive",
  // Interface clicks (button presses, menu opens, panel toggles)
  click:   "interface",
  panels:  "interface",
  expand:  "interface",
  folder:  "interface",
  // Everything else → system (boot, error tones, granted, etc.)
};

export function categoryOf(cue: EdexCue): CueCategory {
  return CUE_CATEGORY[cue] ?? "system";
}

/* Aliases — existing aDex events route through one of the cues. --------- */

const LEGACY_ALIASES: Record<string, EdexCue> = {
  // Existing soundpacks.json events
  system_startup:        "scan",
  terminal_bell:         "alarm",
  command_success:       "granted",
  command_error:         "denied",
  // button_click now routes to the new click-tone cue rather than the
  // keyboard one — the keyboard cue is reserved for the typewriter
  // mechanical click.
  button_click:          "click",
  menu_open:             "panels",
  notification:          "info",
  file_complete:         "granted",
  file_error:            "error",
  network_connected:     "info",
  network_disconnected:  "alarm",
  system_alert:          "alarm",
};

export function resolveCue(name: string): EdexCue | null {
  if ((EDEX_CUES as readonly string[]).includes(name)) return name as EdexCue;
  return LEGACY_ALIASES[name] ?? null;
}

/* Settings ---------------------------------------------------------------- */

export interface AudioSettings {
  /** "adex" (default WAV pack) or "synth" (Web-Audio synthesized fallback). */
  pack: "adex" | "synth";
  /** Global volume multiplier (0..1) applied on top of per-cue volume. */
  globalVolume: number;
  /** Master mute. */
  muted: boolean;
  /** Per-cue volumes (0..1). */
  volumes: Partial<Record<EdexCue, number>>;
  /**
   * Per-category opt-out. Each category defaults to `true` (audible).
   * When false, cues belonging to that category are silenced even if
   * the master `muted` flag is false. Wired into the Settings → Audio
   * panel so users can independently mute the typewriter / destructive
   * confirm tone / interface clicks without losing splash audio.
   */
  categories: Record<CueCategory, boolean>;
}

const STORAGE_KEY = "adex.audio.settings";
const DEFAULT_SETTINGS: AudioSettings = {
  pack: "adex",
  globalVolume: 0.6,
  muted: false,
  volumes: {
    keyboard:    0.4,
    keypress:    0.35, // typewriter — keep low so typing isn't deafening
    panels:      0.5,
    stdin:       0.3,
    stdout:      0.3,
    theme:       0.7,
    click:       0.35, // routine UI clicks — also kept low
    destructive: 0.7,  // the user WANTS to notice this one
    gunshot:     0.8,  // confirmation impact — even louder than the open cue
  },
  categories: {
    keyboard:    true,
    destructive: true,
    interface:   true,
    system:      true,
  },
};

/* Module state ----------------------------------------------------------- */

const settings = ref<AudioSettings>(loadSettings());
const lastFireMs = new Map<EdexCue, number>();
const synthFallbackLogged = new Set<EdexCue>();
const howlCache = new Map<EdexCue, unknown>(); // Howl | null
const synthCtx: { value: AudioContext | null } = { value: null };
const isInitialized = ref(false);
const _playCount = ref(0); // for tests

// External signals
const backgroundMuted = ref(false);

/* Settings persistence ---------------------------------------------------
 *
 * VueUse's useStorage gives us a reactive ref synced to localStorage with
 * graceful JSON (de)serialization built in. We instantiate per-call so
 * vitest's per-test localStorage.clear() doesn't desync from a captured ref. */

function loadSettings(): AudioSettings {
  if (typeof localStorage === "undefined") return { ...DEFAULT_SETTINGS };
  const stored = useStorage<AudioSettings>(STORAGE_KEY, DEFAULT_SETTINGS, undefined, {
    mergeDefaults: true,
    flush: "sync",
  });
  // Migration: the WAV pack was renamed from 'edex' to 'adex' when we
  // consolidated the asset folders. Anyone with the old saved value
  // would otherwise hit the `pack === "adex"` branch as a string-
  // mismatch and silently fall through to no preload (and the synth
  // path on every play). Patch the old value in place so they get the
  // new pack without having to touch settings.
  const raw = stored.value as AudioSettings & { pack?: string };
  if (raw.pack === "edex") {
    raw.pack = "adex";
    stored.value = { ...raw } as AudioSettings;
  }
  return { ...DEFAULT_SETTINGS, ...stored.value };
}

function persistSettings(): void {
  if (typeof localStorage === "undefined") return;
  const stored = useStorage<AudioSettings>(STORAGE_KEY, DEFAULT_SETTINGS, undefined, {
    mergeDefaults: true,
    flush: "sync",
  });
  stored.value = { ...settings.value };
}

/* Howler-based primary path --------------------------------------------- */

/* Howler instance type — minimal surface we actually call on it.
 * Declared once so we don't repeat the cast everywhere. */
interface HowlInstance {
  play(): number;
  volume(v: number): void;
  once(event: string, cb: () => void): void;
}

// Track which Howl instances already have an `unlock` listener wired
// from their onplayerror handler. Without this gate, every dropped
// playback under the autoplay policy would register a NEW `once`
// callback — and if `unlock` never fires (no user gesture), those
// callbacks accumulate forever on the Howl's internal event map.
// Set keys are the Howl instances themselves; capped at 13 (one per
// EdexCue) so this is naturally bounded.
const unlockListenerArmed = new WeakSet<HowlInstance>();

// Bundle-time URL resolution for cue WAVs.
//
// Explicit `?url` imports — one per cue. This is the most reliable
// pattern in Vite: each import becomes a static analyzable reference,
// Vite emits the file as a hashed asset under `_nuxt/`, and the
// imported value is the final URL the browser should fetch.
//
// History — what didn't work and why:
//   1. `new URL('../path/${cue}.wav', import.meta.url).href`
//      The Vite docs say this is the recommended pattern for dynamic
//      asset paths, and it works correctly in many setups. But under
//      Nuxt 4 + Wails the production bundle ended up with literal
//      strings like `audio/edex/scan.wav` instead of hashed
//      `_nuxt/scan.<hash>.wav` URLs — Vite's static analyzer didn't
//      transform the template under this build pipeline. The WAVs
//      were emitted to `_nuxt/` but the code referenced unhashed
//      relative paths that 404'd at runtime. EVERY cue silently fell
//      to synth. This was the root cause of "audio went completely
//      silent" after the URL refactor.
//   2. `import.meta.glob('/public/audio/...')` — matched nothing
//      because Vite excludes public/ from glob resolution.
//
// Why explicit imports always work:
//   They're plain ES module imports. Vite's resolver and asset
//   pipeline handle them via a single well-trodden path that doesn't
//   depend on template-literal static analysis. The cost is one
//   import line per cue, paid here at the module scope.
// All cue assets now live under a single `assets/audio/adex/` folder.
// We used to split between `edex/` (the 13 legacy eDEX samples) and
// `adex/` (the new descriptively-named files like mechanical-keyboard,
// handgun-click, gunshot, click-tone) but the split was internal noise:
// every cue is referenced individually by URL, the directory layout
// never mattered to the runtime, and "edex" as a pack name confused
// users into thinking it controlled WHICH cues played. One folder,
// one pack name.
import alarmUrl       from "~/assets/audio/adex/alarm.wav?url";
import deniedUrl      from "~/assets/audio/adex/denied.wav?url";
import errorUrl       from "~/assets/audio/adex/error.wav?url";
import expandUrl      from "~/assets/audio/adex/expand.wav?url";
import folderUrl      from "~/assets/audio/adex/folder.wav?url";
import grantedUrl     from "~/assets/audio/adex/granted.wav?url";
import infoUrl        from "~/assets/audio/adex/info.wav?url";
import keyboardUrl    from "~/assets/audio/adex/keyboard.wav?url";
import panelsUrl      from "~/assets/audio/adex/panels.wav?url";
import scanUrl        from "~/assets/audio/adex/scan.wav?url";
import stdinUrl       from "~/assets/audio/adex/stdin.wav?url";
import stdoutUrl      from "~/assets/audio/adex/stdout.wav?url";
import themeUrl       from "~/assets/audio/adex/theme.wav?url";
import keypressUrl    from "~/assets/audio/adex/mechanical-keyboard.wav?url";
import destructiveUrl from "~/assets/audio/adex/handgun-click.wav?url";
import gunshotUrl     from "~/assets/audio/adex/gunshot.wav?url";
import clickUrl       from "~/assets/audio/adex/click-tone.wav?url";

const CUE_URLS: Record<EdexCue, string> = {
  alarm:       alarmUrl,
  denied:      deniedUrl,
  error:       errorUrl,
  expand:      expandUrl,
  folder:      folderUrl,
  granted:     grantedUrl,
  info:        infoUrl,
  keyboard:    keyboardUrl,
  panels:      panelsUrl,
  scan:        scanUrl,
  stdin:       stdinUrl,
  stdout:      stdoutUrl,
  theme:       themeUrl,
  keypress:    keypressUrl,
  destructive: destructiveUrl,
  gunshot:     gunshotUrl,
  click:       clickUrl,
};

async function getHowl(cue: EdexCue): Promise<HowlInstance | null> {
  const cached = howlCache.get(cue);
  if (cached !== undefined) return cached as HowlInstance | null;
  // Read the precomputed URL — same one the startup probe HEAD-checked.
  const url = CUE_URLS[cue];
  try {
    // Lazy import keeps Howler out of the initial chunk.
    const mod: { Howl: new (opts: object) => HowlInstance } =
      await import("howler");
    const howl = new mod.Howl({
      src: [url],
      preload: true,
      volume: 1.0,
      onloaderror: () => {
        // Load failure — cache null so we fall back to synth without
        // re-attempting the fetch on every play.
        howlCache.set(cue, null);
        if (!synthFallbackLogged.has(cue)) {
          // eslint-disable-next-line no-console
          console.info(`[audio] ${cue}.wav failed to load — falling back to synth`);
          synthFallbackLogged.add(cue);
        }
      },
      // CRITICAL: under the autoplay policy (Chrome/Wails WebView too)
      // Howler can fail to play even after load if the page hasn't
      // received a user gesture yet. The Howler docs document this
      // exact pattern: register `playerror`, wait for `unlock`, retry.
      // Without this, the first boot tones are silently dropped during
      // the splash animation and audio only "starts working" after
      // the user clicks something.
      //
      // Memory guard: arm the `once('unlock')` listener AT MOST once
      // per Howl instance. Otherwise every dropped play during a
      // long pre-gesture period would stack another callback on the
      // Howl's internal event map. unlockListenerArmed is a WeakSet
      // so it can't outlive the Howl instances themselves.
      onplayerror: function (this: HowlInstance) {
        if (unlockListenerArmed.has(this)) return;
        unlockListenerArmed.add(this);
        try {
          this.once("unlock", () => {
            // Allow re-arming after a successful unlock — if the
            // context gets suspended again later (e.g. autoSuspend
            // after idle), we want to be able to re-listen.
            unlockListenerArmed.delete(this);
            try { this.play(); } catch { /* swallow */ }
          });
        } catch {
          unlockListenerArmed.delete(this);
        }
      },
    });
    howlCache.set(cue, howl);
    return howl;
  } catch (err) {
    howlCache.set(cue, null);
    if (!synthFallbackLogged.has(cue)) {
      // eslint-disable-next-line no-console
      console.info(`[audio] howler unavailable for ${cue} — falling back to synth (${err})`);
      synthFallbackLogged.add(cue);
    }
    return null;
  }
}

/* Web-Audio synth fallback ---------------------------------------------- */

function getSynthCtx(): AudioContext | null {
  if (synthCtx.value) {
    // Resume on demand. Browsers may suspend the context after the
    // tab loses focus (Howler.autoSuspend is on by default). Calling
    // resume() on an already-running context is a documented no-op
    // and returns immediately, so this is cheap to do every play.
    if (synthCtx.value.state === "suspended") {
      // Fire-and-forget — we don't await because the cue scheduling
      // happens on `ctx.currentTime` which updates as soon as the
      // resume promise resolves. The first tone after a long idle
      // period may be ~10ms late, but it won't be dropped.
      synthCtx.value.resume().catch(() => undefined);
    }
    return synthCtx.value;
  }
  if (typeof window === "undefined") return null;
  type Win = Window & typeof globalThis & { webkitAudioContext?: typeof AudioContext };
  const Ctor = window.AudioContext ?? (window as Win).webkitAudioContext;
  if (!Ctor) return null;
  synthCtx.value = new Ctor();
  // Chrome/Wails create the context in `suspended` state under the
  // autoplay policy until a user gesture. We still try to resume
  // here so the FIRST cue plays as soon as a gesture lands, without
  // needing a second call to surface it.
  if (synthCtx.value.state === "suspended") {
    synthCtx.value.resume().catch(() => undefined);
  }
  return synthCtx.value;
}

interface SynthRecipe {
  freq: number;
  to?: number;
  type: OscillatorType;
  durationMs: number;
}

const SYNTH_RECIPES: Record<EdexCue, SynthRecipe> = {
  keyboard:    { freq: 1200, type: "square",   durationMs: 40 },
  panels:      { freq: 600,  to: 800,   type: "triangle", durationMs: 180 },
  theme:       { freq: 320,  to: 880,   type: "sawtooth", durationMs: 420 },
  stdin:       { freq: 1500, type: "sine",     durationMs: 30 },
  stdout:      { freq: 1100, type: "sine",     durationMs: 30 },
  granted:     { freq: 660,  to: 990,   type: "sine",     durationMs: 280 },
  denied:      { freq: 220,  to: 110,   type: "square",   durationMs: 220 },
  error:       { freq: 200,  to: 80,    type: "sawtooth", durationMs: 360 },
  info:        { freq: 880,  type: "sine",     durationMs: 220 },
  alarm:       { freq: 440,  to: 880,   type: "square",   durationMs: 900 },
  expand:      { freq: 700,  to: 950,   type: "triangle", durationMs: 160 },
  folder:      { freq: 540,  to: 720,   type: "triangle", durationMs: 220 },
  scan:        { freq: 100,  to: 1200,  type: "sawtooth", durationMs: 1000 },
  // New aDex synth fallbacks — used when the corresponding WAV in
  // assets/audio/adex/ fails to load. Tuned to roughly match the
  // sonic character of each real sample:
  //   - keypress: short crisp click, slightly lower than `keyboard`
  //   - destructive: sharp drop, evokes the handgun-click slide
  //   - click: bright two-pitch UI blip
  keypress:    { freq: 900,  type: "square",   durationMs: 25 },
  destructive: { freq: 1400, to: 200, type: "square",   durationMs: 90 },
  // Gunshot synth fallback — short impact-like noise envelope.
  gunshot:     { freq: 80,   to: 30,  type: "sawtooth", durationMs: 250 },
  click:       { freq: 1500, to: 1800, type: "triangle", durationMs: 60 },
};

function playSynth(cue: EdexCue, volume: number): void {
  const ctx = getSynthCtx();
  if (!ctx) return;
  const r = SYNTH_RECIPES[cue];
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = r.type;
  osc.frequency.setValueAtTime(r.freq, ctx.currentTime);
  if (r.to !== undefined) {
    osc.frequency.exponentialRampToValueAtTime(
      Math.max(1, r.to),
      ctx.currentTime + r.durationMs / 1000,
    );
  }
  gain.gain.setValueAtTime(0, ctx.currentTime);
  gain.gain.linearRampToValueAtTime(volume, ctx.currentTime + 0.005);
  gain.gain.linearRampToValueAtTime(0, ctx.currentTime + r.durationMs / 1000);
  osc.connect(gain).connect(ctx.destination);
  osc.start();
  osc.stop(ctx.currentTime + r.durationMs / 1000 + 0.02);
  // Explicit teardown when the tone ends.
  //
  // Why this matters: even though the spec says scheduled OscillatorNodes
  // become eligible for GC after their `ended` event fires, that only
  // holds if they have NO outgoing connections. Leaving `osc.connect(
  // gain).connect(ctx.destination)` wired up keeps the implementation's
  // internal graph references alive on some engines (notably Blink's
  // pre-2024 Web Audio refactor) and the nodes pile up across thousands
  // of cue plays during a long session. Disconnecting in `onended`
  // makes both nodes immediately collectible.
  //
  // We also null out `onended` itself so the handler closure doesn't
  // outlive the call — otherwise we'd be holding `cue`/`volume`
  // captures longer than necessary.
  osc.onended = () => {
    try { osc.disconnect(); } catch { /* already disconnected */ }
    try { gain.disconnect(); } catch { /* already disconnected */ }
    osc.onended = null;
  };
}

/* Public API ------------------------------------------------------------- */

const RATE_LIMIT_MS: Partial<Record<EdexCue, number>> = {
  stdin:    250,   // 4 Hz cap
  stdout:   250,   // 4 Hz cap
  // Typing produces one onData event per keystroke at ~10 cps fast typing.
  // 35ms is roughly the click duration — anything shorter just stacks
  // tones on top of each other and sounds like noise.
  keypress: 35,
  keyboard: 35,
  click:    80,    // routine clicks don't need rapid-fire
};

export interface UseAdexAudioApi {
  settings: Readonly<Ref<AudioSettings>>;
  isInitialized: Readonly<Ref<boolean>>;
  playCount: Readonly<Ref<number>>;
  initialize(): Promise<void>;
  playCue(name: string): void;
  setPack(pack: "adex" | "synth"): void;
  setMuted(muted: boolean): void;
  setGlobalVolume(volume: number): void;
  setCueVolume(cue: EdexCue, volume: number): void;
  setCategoryEnabled(category: CueCategory, enabled: boolean): void;
  setBackgroundMuted(muted: boolean): void;
}

function effectiveVolume(cue: EdexCue): number {
  if (settings.value.muted || backgroundMuted.value) return 0;
  // Per-category gate — silences whole groups of cues without touching
  // global mute. `categories` is optional in older saved settings, so
  // we default to true when the key is missing to avoid retroactively
  // silencing users on upgrade.
  const cat = categoryOf(cue);
  if (settings.value.categories?.[cat] === false) return 0;
  const perCue = settings.value.volumes[cue] ?? 1;
  return Math.max(0, Math.min(1, settings.value.globalVolume * perCue));
}

function rateLimited(cue: EdexCue): boolean {
  const cap = RATE_LIMIT_MS[cue];
  if (!cap) return false;
  // Check for the METHOD, not just the object. Some webview contexts expose a
  // partial `performance` without `now`, so `typeof performance !== "undefined"`
  // passed and the call then threw `performance.now is not a function` — which
  // propagated out of playCue and silenced the cue entirely.
  //
  // This only ever affected rate-limited cues (stdin, stdout, keypress,
  // keyboard, click), so interaction sounds went missing while un-limited ones
  // like the settings and boot cues kept playing.
  const now =
    typeof performance?.now === "function" ? performance.now() : Date.now();
  const last = lastFireMs.get(cue) ?? 0;
  if (now - last < cap) return true;
  lastFireMs.set(cue, now);
  return false;
}

export function useAdexAudio(): UseAdexAudioApi {
  async function initialize() {
    if (isInitialized.value) return;
    isInitialized.value = true;
    if (settings.value.pack === "adex") {
      // Best-effort preload of the most common cues. Failures fall back to synth.
      await Promise.allSettled(
        (["keyboard", "panels", "theme"] as EdexCue[]).map((c) => getHowl(c)),
      );
    }
  }

  function playCue(name: string): void {
    const cue = resolveCue(name);
    if (!cue) return;
    if (rateLimited(cue)) return;

    const vol = effectiveVolume(cue);
    if (vol <= 0) return;
    _playCount.value += 1;

    if (settings.value.pack === "synth") {
      playSynth(cue, vol);
      return;
    }

    // Howler primary; fall back to synth if Howler load failed.
    void getHowl(cue).then((howl) => {
      if (!howl) {
        playSynth(cue, vol);
        return;
      }
      howl.volume(vol);
      howl.play();
    });
  }

  function setPack(pack: "adex" | "synth") {
    settings.value = { ...settings.value, pack };
    persistSettings();
  }
  function setMuted(muted: boolean) {
    settings.value = { ...settings.value, muted };
    persistSettings();
  }
  function setGlobalVolume(v: number) {
    settings.value = { ...settings.value, globalVolume: Math.max(0, Math.min(1, v)) };
    persistSettings();
  }
  function setCueVolume(cue: EdexCue, v: number) {
    settings.value = {
      ...settings.value,
      volumes: { ...settings.value.volumes, [cue]: Math.max(0, Math.min(1, v)) },
    };
    persistSettings();
  }
  function setCategoryEnabled(category: CueCategory, enabled: boolean) {
    const current = settings.value.categories ?? {
      keyboard: true, destructive: true, interface: true, system: true,
    };
    settings.value = {
      ...settings.value,
      categories: { ...current, [category]: enabled },
    };
    persistSettings();
  }
  function setBackgroundMuted(muted: boolean) {
    backgroundMuted.value = muted;
  }

  return {
    settings: readonly(settings) as Readonly<Ref<AudioSettings>>,
    isInitialized: readonly(isInitialized) as Readonly<Ref<boolean>>,
    playCount: readonly(_playCount) as Readonly<Ref<number>>,
    initialize,
    playCue,
    setPack,
    setMuted,
    setGlobalVolume,
    setCueVolume,
    setCategoryEnabled,
    setBackgroundMuted,
  };
}

/* Test internals --------------------------------------------------------- */

export const _internals = {
  STORAGE_KEY,
  EDEX_CUES,
  LEGACY_ALIASES,
  reset() {
    settings.value = { ...DEFAULT_SETTINGS };
    lastFireMs.clear();
    synthFallbackLogged.clear();
    howlCache.clear();
    synthCtx.value = null;
    isInitialized.value = false;
    backgroundMuted.value = false;
    _playCount.value = 0;
    if (typeof localStorage !== "undefined") localStorage.removeItem(STORAGE_KEY);
  },
};
