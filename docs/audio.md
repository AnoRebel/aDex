# Audio Cue System

aDex's audio is a small, curated cue vocabulary played through
[Howler.js](https://howlerjs.com/) with a Web-Audio oscillator synth
fallback. The engine lives in
`frontend/app/composables/useAdexAudio.ts`.

## Architecture

```
caller → playCue(name)
            │  resolveCue: legacy alias → canonical cue
            │  rateLimited?  (per-cue cap, e.g. keypress 35ms)
            │  effectiveVolume = 0 if muted / category-off
            ▼
       pack === 'synth' ──► playSynth (oscillator recipe)
       pack === 'adex'  ──► Howler.play(<hashed wav url>)
                                 │ onloaderror / unavailable
                                 ▼
                            playSynth fallback
```

WAV assets live in `frontend/app/assets/audio/adex/` and are imported
with Vite's `?url` suffix so they emit as hashed bundle assets
(`_nuxt/<name>.<hash>.wav`) — this works identically under `wails3 dev`,
`wails3 task build`, and a standalone Nuxt deploy. (The earlier
`import.meta.glob` / `new URL(...)` approaches were unreliable under
the Nuxt 4 + Wails pipeline; explicit `?url` imports are the only
form that consistently transformed.)

## Cue vocabulary

16 canonical cues. Each belongs to a **category** (the per-category
mute switches in Settings → AUDIO gate whole groups):

| Cue           | Category    | File / synth | Used for |
|---------------|-------------|--------------|----------|
| `keyboard`    | keyboard    | `keyboard.wav` | legacy terminal-typing tick |
| `keypress`    | keyboard    | `mechanical-keyboard.wav` | per-keystroke in the terminal |
| `stdin`       | keyboard    | `stdin.wav` | terminal input stream activity |
| `destructive` | destructive | `handgun-click.wav` | destructive confirm dialog **opens** |
| `gunshot`     | destructive | `gunshot.wav` | destructive action **confirmed** |
| `denied`      | destructive | `denied.wav` | rejected action (e.g. tab limit) |
| `alarm`       | destructive | `alarm.wav` | resource threshold breach (mem ≥ 92%) |
| `click`       | interface   | `click-tone.wav` | routine UI clicks, tab switch, dialog dismiss |
| `panels`      | interface   | `panels.wav` | modal / panel open |
| `expand`      | interface   | `expand.wav` | expand/collapse |
| `folder`      | interface   | `folder.wav` | file-manager navigate |
| `scan`        | system      | `scan.wav` | boot sweep / shutdown wind-down |
| `granted`     | system      | `granted.wav` | success / boot complete |
| `error`       | system      | `error.wav` | error notification |
| `info`        | system      | `info.wav` | info notification |
| `theme`       | system      | `theme.wav` | theme change |
| `stdout`      | system      | `stdout.wav` | terminal output stream activity |

### Legacy aliases

Older event names route to a canonical cue via `LEGACY_ALIASES`:

| Legacy name           | → cue     |
|-----------------------|-----------|
| `system_startup`      | `scan`    |
| `terminal_bell`       | `alarm`   |
| `command_success`     | `granted` |
| `command_error`       | `denied`  |
| `button_click`        | `click`   |
| `menu_open`           | `panels`  |
| `notification`        | `info`    |
| `file_complete`       | `granted` |
| `file_error`          | `error`   |
| `network_connected`   | `info`    |
| `network_disconnected`| `alarm`   |
| `system_alert`        | `alarm`   |

## Where cues fire

| Surface | Cue |
|---------|-----|
| Boot splash start / steps / complete | `scan` / `stdin` / `granted` |
| Shutdown splash sequence | `scan` → per-step → `alarm` capstone |
| Terminal keystroke | `keypress` |
| Tab switch | `click` |
| Settings modal open | `panels` |
| Theme change | `theme` |
| Quit dialog opens | `destructive` |
| Quit confirmed | `gunshot` |
| Kill-process dialog opens | `destructive` |
| Kill confirmed | `gunshot` |
| Memory ≥ 92% (hysteresis re-arm < 85%) | `alarm` |
| Audio re-enabled in Settings | `click` |

## Settings

Settings → **AUDIO** tab:

- **Enable Audio** — master mute. Drives `useAdexAudio.setMuted`.
- **Master Volume** — 0–100, mapped to `setGlobalVolume(0..1)`.
- **Soundpack** — `adex` (WAV samples, recommended) or `synth`
  (Web-Audio oscillators, no asset load — for accessibility /
  asset-missing modes).
- **Mute in Background** — silences cues when the window loses focus.
- **Boot Splash Audio** / **Shutdown Splash Audio** — independent
  toggles for the two splash sequences.
- **Keyboard Click** / **Destructive Confirm Tone** /
  **Interface Clicks** — per-category mutes (mute one class without
  losing the others).

Persistence: the composable owns
`localStorage['adex.audio.settings']`. The Settings panel's per-
category toggles live in `adex-settings.audio.category*` and are
forwarded into the engine by a `watch` in `pages/index.vue`.

## Rate limiting

High-frequency cues are capped (`RATE_LIMIT_MS`): `keypress`/`keyboard`
35ms, `stdin`/`stdout` 250ms (4 Hz), `click` 80ms. Fast typing won't
machine-gun the engine.

## Browser autoplay policy

Web Audio is suspended until the first user gesture. The boot splash
plays before any interaction, so the first launch's boot tones are
silently dropped — this is a browser security boundary, not a bug.
The Howler `onplayerror` → `once('unlock')` retry (armed at most once
per Howl via a WeakSet) recovers playback after the first click. The
shutdown splash always has sound because the Quit click is itself the
unlocking gesture.

## Adding a cue

1. Add the WAV to `frontend/app/assets/audio/adex/`.
2. Add the cue name to `EDEX_CUES` in `useAdexAudio.ts`.
3. Add a `CUE_CATEGORY[<cue>]` mapping (or it defaults to `system`).
4. Add an explicit `?url` import + a `CUE_URLS` entry.
5. Add a `SYNTH_RECIPES[<cue>]` oscillator fallback.
6. Optionally add a `RATE_LIMIT_MS[<cue>]` cap and a default
   `volumes[<cue>]`.
7. Optionally register it in `soundpacks.json` so it shows in any
   future soundpack-editor UI.
