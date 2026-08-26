// composables/useUpdateChecker.ts
//
// GitHub-backed update checker for aDex-UI.
//
// Hits the public Releases API (no auth — public repo, 60 req/hr per IP
// without a token, which is more than enough for occasional polls).
// Compares the latest release's tag against the current build version
// and surfaces a non-blocking modal when an update is available.
//
// State (persisted via useStorage so the user's preferences and recent
// check history survive restarts):
//   - settings.enabled         : auto-check toggle (default true)
//   - settings.interval        : 'manual' | 'launch' | 'daily' | 'weekly'
//   - settings.skippedVersions : list of tags the user clicked "Skip"
//   - state.lastChecked        : ISO timestamp of the most recent check
//   - state.latest             : last release the API returned (cached
//                                so we can render the modal again
//                                without re-hitting GitHub)
//
// The auto-check timer is driven by useIntervalFn on the page that
// mounts the checker (we don't run a global background timer here —
// pages/index.vue calls `maybeCheckOnLaunch()` once after boot and
// then schedules subsequent checks per interval).

import { computed, readonly, ref, type Ref } from "vue";
import { useStorage } from "@vueuse/core";

/* ---- Config -------------------------------------------------------- */

/**
 * Owner + repo of the GitHub project. Hardcoded — this checker is for
 * aDex-UI specifically, not a generic utility.
 */
const REPO_OWNER = "AnoRebel";
const REPO_NAME = "aDex-UI";

/**
 * Current app version — injected by Vite from wails.json's
 * info.productVersion at build time via `__APP_VERSION__`. Falls back
 * to '0.0.0' if the define isn't wired (jsdom tests, etc.) so we
 * never crash, just never claim "up to date" without a real version.
 */
declare const __APP_VERSION__: string | undefined;
export const CURRENT_VERSION: string =
  typeof __APP_VERSION__ !== "undefined" ? __APP_VERSION__ : "0.0.0";

export type CheckInterval = "manual" | "launch" | "daily" | "weekly";

export interface UpdateSettings {
  /** Master switch — false disables ALL update activity. */
  enabled: boolean;
  /** How often the auto-checker fires. `'manual'` = never. */
  interval: CheckInterval;
  /** Release tags the user clicked "Skip this version" on. */
  skippedVersions: string[];
}

/** Subset of the GitHub Releases API response we actually use. */
export interface GitHubRelease {
  tag_name: string;
  name: string;
  html_url: string;
  body: string;
  published_at: string;
  prerelease: boolean;
  draft: boolean;
  assets: Array<{
    name: string;
    browser_download_url: string;
    size: number;
    content_type: string;
  }>;
}

export interface UpdateState {
  /** ISO timestamp of the last successful check. */
  lastChecked: string;
  /** ISO timestamp of the last attempted check (even if failed). */
  lastAttempted: string;
  /** Last error message from the most recent attempt, if any. */
  lastError: string;
  /** Cached release payload so the modal can re-render offline. */
  latest: GitHubRelease | null;
}

const DEFAULT_SETTINGS: UpdateSettings = {
  enabled: true,
  interval: "launch",
  skippedVersions: [],
};

const DEFAULT_STATE: UpdateState = {
  lastChecked: "",
  lastAttempted: "",
  lastError: "",
  latest: null,
};

/* ---- Module-scope reactive state ----------------------------------- */
//
// We hoist `settings` and `state` out of the composable factory so
// every consumer reads the SAME refs. Otherwise the "check now"
// button in Settings and the auto-checker on the page would each
// keep their own copy and stale-state surprises would follow.

const settings = useStorage<UpdateSettings>(
  "adex.updates.settings",
  DEFAULT_SETTINGS,
  undefined,
  { mergeDefaults: true },
);

const state = useStorage<UpdateState>(
  "adex.updates.state",
  DEFAULT_STATE,
  undefined,
  { mergeDefaults: true },
);

/** True while a check is in flight — drives the spinner in Settings. */
const checking = ref(false);

/* ---- Helpers ------------------------------------------------------- */

/**
 * Strip a leading `v` from a tag and split into numeric parts.
 *   'v1.2.3'  → [1, 2, 3]
 *   '1.2.3-rc.1' → [1, 2, 3]  (pre-release tail ignored)
 *   'main'    → [0]            (unparseable falls to zero)
 */
function parseSemver(tag: string): number[] {
  if (!tag) return [0];
  const stripped = tag.replace(/^v/i, "").split(/[-+]/)[0] ?? "0";
  const parts = stripped.split(".").map((p) => {
    const n = Number(p);
    return Number.isFinite(n) ? n : 0;
  });
  return parts.length > 0 ? parts : [0];
}

/**
 * Returns >0 if `a` newer than `b`, <0 if older, 0 if equal.
 * Pads shorter lists with 0 so 1.2 ≡ 1.2.0.
 */
function compareSemver(a: string, b: string): number {
  const pa = parseSemver(a);
  const pb = parseSemver(b);
  const len = Math.max(pa.length, pb.length);
  for (let i = 0; i < len; i++) {
    const da = pa[i] ?? 0;
    const db = pb[i] ?? 0;
    if (da !== db) return da - db;
  }
  return 0;
}

/**
 * True if the cached release's tag is newer than the current build
 * AND the user hasn't explicitly skipped it. This is what the page
 * checks to decide whether to mount the UpdateAvailableModal.
 */
function isUpdateAvailable(): boolean {
  const rel = state.value.latest;
  if (!rel || !settings.value.enabled) return false;
  if (rel.draft) return false;
  // Pre-releases are advertised only if the user is already on one
  // (so stable users don't get pestered with betas).
  if (rel.prerelease && !/-/.test(CURRENT_VERSION)) return false;
  if (settings.value.skippedVersions.includes(rel.tag_name)) return false;
  return compareSemver(rel.tag_name, CURRENT_VERSION) > 0;
}

/**
 * Hit the Releases API. Stores the result in `state.latest` regardless
 * of whether it's newer — the "is update available" decision is made
 * downstream by isUpdateAvailable(). On failure we log to state.lastError
 * but DON'T throw, because the UI surface is non-critical.
 */
async function fetchLatestRelease(): Promise<GitHubRelease | null> {
  if (typeof fetch === "undefined") return null;
  const url = `https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest`;
  state.value.lastAttempted = new Date().toISOString();
  try {
    const res = await fetch(url, {
      headers: { Accept: "application/vnd.github+json" },
    });
    if (!res.ok) {
      // 404 = no releases yet. We don't treat that as an error in
      // user-facing messaging because it just means "no updates".
      if (res.status === 404) {
        state.value.lastError = "";
        return null;
      }
      state.value.lastError = `GitHub returned HTTP ${res.status}`;
      return null;
    }
    const data = (await res.json()) as GitHubRelease;
    state.value.latest = data;
    state.value.lastChecked = state.value.lastAttempted;
    state.value.lastError = "";
    return data;
  } catch (err) {
    state.value.lastError = err instanceof Error ? err.message : String(err);
    return null;
  }
}

/* ---- Public API ---------------------------------------------------- */

export interface UseUpdateCheckerApi {
  /** Reactive settings — write-throughs auto-persist. */
  settings: typeof settings;
  /** Reactive state (cached release, timestamps, last error). */
  state: typeof state;
  /** True while a network check is in flight. */
  checking: Readonly<Ref<boolean>>;
  /** Current app version (read-only). */
  currentVersion: string;
  /** True iff there's a non-skipped newer release cached. */
  hasUpdate: Readonly<Ref<boolean>>;

  /** Manually trigger a check now. Returns the latest release or null. */
  checkNow(): Promise<GitHubRelease | null>;

  /** Wrapper around `checkNow` that respects the interval setting.
   *  Call from pages/index.vue on boot. */
  maybeCheckOnLaunch(): Promise<void>;

  /** Mark a version as skipped — modal won't re-appear for it. */
  skipVersion(tag: string): void;

  /** Reset the skip list. Used when user wants to be reminded again. */
  unskipAll(): void;

  /** Open the latest release's html_url in the OS browser via Wails
   *  (or window.open as a fallback). */
  openReleasePage(): void;
}

export function useUpdateChecker(): UseUpdateCheckerApi {
  const hasUpdate = computed(() => isUpdateAvailable());

  async function checkNow() {
    if (checking.value) return state.value.latest;
    checking.value = true;
    try {
      return await fetchLatestRelease();
    } finally {
      checking.value = false;
    }
  }

  async function maybeCheckOnLaunch() {
    if (!settings.value.enabled) return;
    const mode = settings.value.interval;
    if (mode === "manual") return;
    if (mode === "launch") {
      await checkNow();
      return;
    }
    // Time-based modes: only fire if enough time has elapsed since the
    // last successful check. Cheaper than running a real cron because
    // most users open the app well within a "weekly" window anyway.
    const last = state.value.lastChecked
      ? new Date(state.value.lastChecked).getTime()
      : 0;
    const now = Date.now();
    const elapsedMs = now - last;
    const oneDay = 24 * 60 * 60 * 1000;
    const threshold = mode === "daily" ? oneDay : 7 * oneDay;
    if (elapsedMs >= threshold) {
      await checkNow();
    }
  }

  function skipVersion(tag: string) {
    if (!tag) return;
    if (settings.value.skippedVersions.includes(tag)) return;
    settings.value.skippedVersions.push(tag);
  }

  function unskipAll() {
    settings.value.skippedVersions = [];
  }

  function openReleasePage() {
    const url = state.value.latest?.html_url;
    if (!url) return;
    // Try the Wails BrowserOpenURL runtime first so the user gets the
    // OS default browser (not the WebView itself, which would navigate
    // away from the app). Fall back to window.open if Wails isn't
    // around (browser preview / jsdom tests).
    const win = window as unknown as {
      runtime?: { BrowserOpenURL?: (u: string) => void };
    };
    try {
      if (typeof win.runtime?.BrowserOpenURL === "function") {
        win.runtime.BrowserOpenURL(url);
        return;
      }
    } catch {
      /* fall through */
    }
    try {
      window.open(url, "_blank", "noopener,noreferrer");
    } catch {
      /* swallow */
    }
  }

  return {
    settings,
    state,
    checking: readonly(checking) as Readonly<Ref<boolean>>,
    currentVersion: CURRENT_VERSION,
    hasUpdate,
    checkNow,
    maybeCheckOnLaunch,
    skipVersion,
    unskipAll,
    openReleasePage,
  };
}
