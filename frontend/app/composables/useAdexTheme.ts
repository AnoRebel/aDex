// composables/useAdexTheme.ts
//
// V2 theme runtime. Lazy-loads themes from `assets/data/themes-index.json`,
// applies them by mutating CSS custom properties on the document root, sets
// the `data-layout` attribute that drives layout-preset stylesheets, and
// persists the active id to localStorage.
//
// This composable is intentionally small and decoupled from the legacy
// `useTheme` / theme store. The shell consumes it via `themeStore.activeLayout`
// (which we will wire here) and via the Settings → Theme panel (section 6).

import { ref, computed, readonly, type ComputedRef, type Ref } from "vue";
import { useStorage } from "@vueuse/core";
import type { AdexTheme, AdexThemeIndex, LayoutPreset } from "~/types/adex-theme";
import { parseAdexTheme } from "~/types/adex-theme";

const STORAGE_KEY = "adex.theme.id";
const DEFAULT_THEME_ID = "tron";

// `useStorage` is created lazily inside readPersistedId/persistId rather
// than at module scope. Module-level instantiation captures a snapshot of
// localStorage that never resyncs after tests call `localStorage.clear()`,
// causing flake. Per-call instantiation is cheap (VueUse caches the
// underlying ref by key under the hood).

/* Module-level singletons so multiple consumers share state. -------------- */
const indexState = ref<AdexThemeIndex>([]);
const themeCache = new Map<string, AdexTheme>();
const activeId = ref<string>(DEFAULT_THEME_ID);
const activeTheme = ref<AdexTheme | null>(null);
const isLoading = ref(false);
const lastError = ref<Error | null>(null);

/* The base URL where theme JSONs live. We resolve at module load via Vite's
 * import.meta.url so the path stays stable in both dev and `nuxt generate`. */
const ASSETS_BASE = "/assets/data/";

async function fetchJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`fetch ${url}: HTTP ${res.status}`);
  // Guard against the SPA fallback: a request for a path that does not exist
  // returns index.html with a 200, so status alone does not mean success.
  const type = res.headers.get("content-type") ?? "";
  if (!type.includes("json")) {
    throw new Error(`fetch ${url}: expected JSON, got ${type || "unknown"}`);
  }
  return (await res.json()) as T;
}

/* Every theme JSON, resolved at build time.
 *
 * These files live under `app/assets/` — a build-time directory that is never
 * served over HTTP (only `frontend/public/` is) — so they must be bundled
 * rather than fetched. Verified key shape: "/assets/data/themes/tron.json". */
const themeModules = import.meta.glob<{ default: unknown }>(
  "~/assets/data/themes/*.json",
);

async function loadIndex(): Promise<AdexThemeIndex> {
  if (indexState.value.length > 0) return indexState.value;
  // The index is small enough to ship inline via `import` so we don't pay a
  // second fetch on first paint. This works in both dev and prod because
  // Vite resolves `?raw` and JSON imports identically.
  try {
    const idx = (await import("~/assets/data/themes-index.json")).default as AdexThemeIndex;
    indexState.value = idx;
    return idx;
  } catch {
    // Fallback: fetch from the public asset path.
    const idx = await fetchJson<AdexThemeIndex>(`${ASSETS_BASE}themes-index.json`);
    indexState.value = idx;
    return idx;
  }
}

async function loadTheme(id: string): Promise<AdexTheme> {
  const cached = themeCache.get(id);
  if (cached) return cached;
  const idx = await loadIndex();
  const entry = idx.find((e) => e.id === id);
  if (!entry) throw new Error(`unknown theme: ${id}`);

  // Resolve through the build-time glob (see themeModules).
  //
  // Neither of the previous strategies worked. `import(/* @vite-ignore */
  // \`~/assets/data/${...}\`)` has a runtime-computed specifier the bundler
  // cannot resolve, so the request fell through to the SPA and the import
  // SUCCEEDED with index.html as its payload — the `catch` never ran, and the
  // HTML reached the parser as `Unexpected token '<' ... is not valid JSON`.
  // Fetching `/assets/data/...` fails the same way, because these files live
  // in `app/assets/` (a build-time directory) and are never served over HTTP;
  // only `frontend/public/` is. Both paths therefore returned the app shell.
  //
  // import.meta.glob is statically analysable, so every theme is bundled and
  // resolved by the bundler with no network request at all.
  const loader = themeModules[`/assets/data/${entry.file}`];
  if (!loader) throw new Error(`theme file not bundled: ${entry.file}`);
  const raw: unknown = (await loader()).default ?? (await loader());
  const theme = parseAdexTheme(raw);
  themeCache.set(id, theme);
  return theme;
}

/** Translate a theme into CSS-custom-property assignments and apply them
 *  to `:root`. Emits a `theme.changed` CustomEvent on `document` so other
 *  modules (xterm renderer, globe canvas) can re-paint without a full reload. */
function applyTheme(theme: AdexTheme): void {
  if (typeof document === "undefined") return;
  const root = document.documentElement;
  const c = theme.colors;

  // Legacy + canonical accent variables (kept in sync to avoid breakage in
  // older rules that still reference --color_r/g/b).
  root.style.setProperty("--color_r", String(c.r));
  root.style.setProperty("--color_g", String(c.g));
  root.style.setProperty("--color_b", String(c.b));
  root.style.setProperty("--accent-rgb", `${c.r}, ${c.g}, ${c.b}`);

  root.style.setProperty("--color_black", c.black);
  root.style.setProperty("--color_light_black", c.light_black);
  root.style.setProperty("--color_grey", c.grey);
  root.style.setProperty("--bg-deep", c.black);
  root.style.setProperty("--bg-rgb", parseHexToRgbTriple(c.light_black));

  root.style.setProperty("--font_main", theme.cssvars.font_main);
  root.style.setProperty("--font_main_light", theme.cssvars.font_main_light);

  root.style.setProperty("--terminal_font", theme.terminal.fontFamily);
  root.style.setProperty("--terminal_bg", theme.terminal.background);
  root.style.setProperty("--terminal_fg", theme.terminal.foreground);
  root.style.setProperty("--terminal_cursor", theme.terminal.cursor);
  root.style.setProperty("--terminal_selection", theme.terminal.selection);

  root.style.setProperty("--globe-base", theme.globe.base);
  root.style.setProperty("--globe-marker", theme.globe.marker);
  root.style.setProperty("--globe-pin", theme.globe.pin);
  root.style.setProperty("--globe-satellite", theme.globe.satellite);

  // Layout preset is applied to the .adex-app element rather than :root so
  // that the boot screen (which is outside .adex-app) doesn't shift.
  const shell = document.querySelector<HTMLElement>(".adex-app");
  if (shell) shell.setAttribute("data-layout", theme.layout);
  root.setAttribute("data-theme", theme.id);

  document.dispatchEvent(
    new CustomEvent("theme.changed", { detail: { id: theme.id, layout: theme.layout } }),
  );
}

function parseHexToRgbTriple(hex: string): string {
  // "#RRGGBB" -> "r, g, b". Returns "5, 8, 13" for the default fallback.
  const m = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(hex.trim());
  if (!m) return "5, 8, 13";
  return `${parseInt(m[1], 16)}, ${parseInt(m[2], 16)}, ${parseInt(m[3], 16)}`;
}

/* Persistence ------------------------------------------------------------- */

function readPersistedId(): string {
  if (typeof localStorage === "undefined") return DEFAULT_THEME_ID;
  // Use VueUse's useStorage as a reactive ref so consumers can watch the
  // active theme cross-tab via the storage event. Read-time instantiation
  // dodges the test-flake described above.
  const stored = useStorage<string>(STORAGE_KEY, DEFAULT_THEME_ID);
  return stored.value || DEFAULT_THEME_ID;
}

function persistId(id: string): void {
  if (typeof localStorage === "undefined") return;
  const stored = useStorage<string>(STORAGE_KEY, DEFAULT_THEME_ID);
  stored.value = id;
}

/* Public API -------------------------------------------------------------- */

export interface UseAdexThemeApi {
  index: Readonly<Ref<AdexThemeIndex>>;
  activeId: Readonly<Ref<string>>;
  activeTheme: Readonly<Ref<AdexTheme | null>>;
  activeLayout: ComputedRef<LayoutPreset>;
  isLoading: Readonly<Ref<boolean>>;
  lastError: Readonly<Ref<Error | null>>;
  initialize(): Promise<void>;
  setTheme(id: string): Promise<AdexTheme>;
  preview(id: string): Promise<AdexTheme>;
}

export function useAdexTheme(): UseAdexThemeApi {
  async function initialize() {
    if (activeTheme.value) return;
    isLoading.value = true;
    lastError.value = null;
    try {
      await loadIndex();
      const id = readPersistedId();
      const theme = await loadTheme(id).catch(async () => loadTheme(DEFAULT_THEME_ID));
      activeId.value = theme.id;
      activeTheme.value = theme;
      applyTheme(theme);
    } catch (err) {
      lastError.value = err instanceof Error ? err : new Error(String(err));
      throw lastError.value;
    } finally {
      isLoading.value = false;
    }
  }

  async function setTheme(id: string): Promise<AdexTheme> {
    isLoading.value = true;
    lastError.value = null;
    try {
      let theme: AdexTheme;
      try {
        theme = await loadTheme(id);
      } catch (err) {
        // An id that is not in the index must not abort theme application.
        // Persisted ids outlive the theme catalogue: settings files, saved
        // localStorage, and older builds can all name a theme that no longer
        // exists (this shipped as `default`, which was never a real id). The
        // old behaviour rethrew here, so a single stale value left the whole
        // interface unstyled — no colours, and no `data-layout` attribute, so
        // layout presets stopped applying too.
        //
        // Fall back to the default theme, record why, and carry on.
        if (id === DEFAULT_THEME_ID) throw err;
        console.warn(
          `[adex-theme] unknown theme "${id}" — falling back to "${DEFAULT_THEME_ID}"`,
        );
        lastError.value = err instanceof Error ? err : new Error(String(err));
        theme = await loadTheme(DEFAULT_THEME_ID);
      }
      activeId.value = theme.id;
      activeTheme.value = theme;
      // Persist the theme that actually loaded, so a stale id is repaired
      // rather than reported again on every start.
      persistId(theme.id);
      applyTheme(theme);
      return theme;
    } catch (err) {
      lastError.value = err instanceof Error ? err : new Error(String(err));
      throw lastError.value;
    } finally {
      isLoading.value = false;
    }
  }

  async function preview(id: string): Promise<AdexTheme> {
    // Apply without persisting — callers (Settings → Themes hover preview)
    // are responsible for calling setTheme on click or restoring the active
    // theme on cancel.
    const theme = await loadTheme(id);
    applyTheme(theme);
    return theme;
  }

  const activeLayout = computed<LayoutPreset>(
    () => activeTheme.value?.layout ?? "default",
  );

  return {
    index: readonly(indexState) as Readonly<Ref<AdexThemeIndex>>,
    activeId: readonly(activeId) as Readonly<Ref<string>>,
    activeTheme: readonly(activeTheme) as Readonly<Ref<AdexTheme | null>>,
    activeLayout,
    isLoading: readonly(isLoading) as Readonly<Ref<boolean>>,
    lastError: readonly(lastError) as Readonly<Ref<Error | null>>,
    initialize,
    setTheme,
    preview,
  };
}

/* Convenience exports for tests ------------------------------------------ */
export const _internals = {
  STORAGE_KEY,
  applyTheme,
  loadTheme,
  loadIndex,
  themeCache,
  reset() {
    indexState.value = [];
    themeCache.clear();
    activeId.value = DEFAULT_THEME_ID;
    activeTheme.value = null;
    isLoading.value = false;
    lastError.value = null;
  },
};
