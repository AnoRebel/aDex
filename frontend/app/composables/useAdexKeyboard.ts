// composables/useAdexKeyboard.ts
//
// V2 keyboard composable: lazy-loads layouts from
// `assets/data/kb_layouts-index.json`, applies CTRLSEQ substitution at
// load time, exposes the active layout + modifier state, and resolves
// keystrokes to the bytes that should be written to the active terminal.
//
// Stays decoupled from the legacy `useKeyboard.ts` (which uses a simpler
// single-cmd schema). The Adex* on-screen keyboard component switches to
// this composable in section 5.F.

import { ref, computed, readonly, type ComputedRef, type Ref } from "vue";
import { useStorage } from "@vueuse/core";
import type {
  KbKey,
  KbLayout,
  KbLayoutIndex,
  KbModifiers,
} from "~/types/kb-layout";
import {
  EMPTY_MODIFIERS,
  isModifierToggle,
  preprocessLayout,
  resolveCmd,
  resolveName,
} from "~/types/kb-layout";

const STORAGE_KEY = "adex.kb.layout";
const DEFAULT_LAYOUT_ID = "en-US";
const ASSETS_BASE = "/assets/data/";

const indexState = ref<KbLayoutIndex>([]);
const layoutCache = new Map<string, KbLayout>();
const activeId = ref<string>(DEFAULT_LAYOUT_ID);
const activeLayout = ref<KbLayout | null>(null);
const modifiers = ref<KbModifiers>({ ...EMPTY_MODIFIERS });
const isLoading = ref(false);
const lastError = ref<Error | null>(null);

/* Every keyboard layout JSON, resolved at build time.
 *
 * These live under `app/assets/` — a build-time directory that is never served
 * over HTTP (only `frontend/public/` is) — so they must be bundled, not
 * fetched. Key shape: "/assets/data/kb_layouts/en-US.json". */
const layoutModules = import.meta.glob<{ default: unknown }>(
  "~/assets/data/kb_layouts/*.json",
);

async function fetchJson<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`fetch ${url}: HTTP ${res.status}`);
  // A path that does not exist returns index.html with a 200, so status alone
  // does not mean success.
  const type = res.headers.get("content-type") ?? "";
  if (!type.includes("json")) {
    throw new Error(`fetch ${url}: expected JSON, got ${type || "unknown"}`);
  }
  return (await res.json()) as T;
}

async function loadIndex(): Promise<KbLayoutIndex> {
  if (indexState.value.length > 0) return indexState.value;
  try {
    const idx = (await import("~/assets/data/kb_layouts-index.json")).default as KbLayoutIndex;
    indexState.value = idx;
    return idx;
  } catch {
    const idx = await fetchJson<KbLayoutIndex>(`${ASSETS_BASE}kb_layouts-index.json`);
    indexState.value = idx;
    return idx;
  }
}

async function loadLayout(id: string): Promise<KbLayout> {
  const cached = layoutCache.get(id);
  if (cached) return cached;
  const idx = await loadIndex();
  const entry = idx.find((e) => e.id === id);
  if (!entry) throw new Error(`unknown keyboard layout: ${id}`);

  // Resolve through the build-time glob. A runtime-computed dynamic import
  // cannot be resolved by the bundler and silently yields the SPA's
  // index.html; fetching `/assets/data/...` fails the same way, since these
  // files are not served over HTTP. import.meta.glob is statically
  // analysable, so the layouts are bundled and need no request at all.
  const loader = layoutModules[`/assets/data/${entry.file}`];
  if (!loader) throw new Error(`keyboard layout not bundled: ${entry.file}`);
  const raw: unknown = (await loader()).default ?? (await loader());
  const layout = preprocessLayout(raw as KbLayout);
  layoutCache.set(id, layout);
  return layout;
}

function readPersistedId(): string {
  if (typeof localStorage === "undefined") return DEFAULT_LAYOUT_ID;
  // VueUse useStorage gives us a reactive ref backed by localStorage; we
  // re-instantiate at read/write time so the per-test localStorage.clear()
  // in vitest survives module-level singleton reuse.
  const stored = useStorage<string>(STORAGE_KEY, DEFAULT_LAYOUT_ID);
  return stored.value || DEFAULT_LAYOUT_ID;
}

function persistId(id: string): void {
  if (typeof localStorage === "undefined") return;
  const stored = useStorage<string>(STORAGE_KEY, DEFAULT_LAYOUT_ID);
  stored.value = id;
}

export interface UseAdexKeyboardApi {
  index: Readonly<Ref<KbLayoutIndex>>;
  activeId: Readonly<Ref<string>>;
  activeLayout: Readonly<Ref<KbLayout | null>>;
  modifiers: Readonly<Ref<KbModifiers>>;
  isLoading: Readonly<Ref<boolean>>;
  lastError: Readonly<Ref<Error | null>>;
  rows: ComputedRef<KbKey[][]>;
  initialize(): Promise<void>;
  setLayout(id: string): Promise<KbLayout>;
  /**
   * Process a key press. Toggles modifiers internally for sticky keys; for
   * printable keys, returns the bytes that callers should inject into the
   * terminal (or `""` if the press was absorbed by a modifier toggle).
   */
  press(key: KbKey): string;
  /** Reset all sticky modifiers to off. */
  resetModifiers(): void;
  /** Convenience helpers for callers that just want a label. */
  labelOf(key: KbKey): string;
}

export function useAdexKeyboard(): UseAdexKeyboardApi {
  async function initialize() {
    if (activeLayout.value) return;
    isLoading.value = true;
    lastError.value = null;
    try {
      await loadIndex();
      const id = readPersistedId();
      const layout = await loadLayout(id).catch(async () => loadLayout(DEFAULT_LAYOUT_ID));
      activeId.value = id;
      activeLayout.value = layout;
    } catch (err) {
      lastError.value = err instanceof Error ? err : new Error(String(err));
      throw lastError.value;
    } finally {
      isLoading.value = false;
    }
  }

  async function setLayout(id: string): Promise<KbLayout> {
    isLoading.value = true;
    lastError.value = null;
    try {
      const layout = await loadLayout(id);
      activeId.value = id;
      activeLayout.value = layout;
      persistId(id);
      return layout;
    } catch (err) {
      lastError.value = err instanceof Error ? err : new Error(String(err));
      throw lastError.value;
    } finally {
      isLoading.value = false;
    }
  }

  function resetModifiers(): void {
    modifiers.value = { ...EMPTY_MODIFIERS };
  }

  function press(key: KbKey): string {
    // Sticky-modifier toggle? Update state and absorb the press.
    const toggle = isModifierToggle(key.cmd);
    if (toggle) {
      // Most modifiers are momentary in the original — they flip on, the
      // next key consumes them, then they flip off. CapsLock is the
      // exception (true toggle).
      modifiers.value = { ...modifiers.value, [toggle]: !modifiers.value[toggle] };
      return "";
    }

    const out = resolveCmd(key, modifiers.value);

    // Auto-release momentary modifiers after a printable press.
    const m = modifiers.value;
    if (m.shift || m.ctrl || m.alt || m.fn) {
      modifiers.value = {
        ...m,
        shift: false,
        ctrl: false,
        alt: false,
        fn: false,
        // CapsLock persists.
      };
    }
    return out;
  }

  function labelOf(key: KbKey): string {
    return resolveName(key, modifiers.value);
  }

  const rows = computed<KbKey[][]>(() => {
    const l = activeLayout.value;
    if (!l) return [];
    return [l.row_numbers, l.row_1, l.row_2, l.row_3, l.row_space];
  });

  return {
    index: readonly(indexState) as Readonly<Ref<KbLayoutIndex>>,
    activeId: readonly(activeId) as Readonly<Ref<string>>,
    activeLayout: readonly(activeLayout) as Readonly<Ref<KbLayout | null>>,
    modifiers: readonly(modifiers) as Readonly<Ref<KbModifiers>>,
    isLoading: readonly(isLoading) as Readonly<Ref<boolean>>,
    lastError: readonly(lastError) as Readonly<Ref<Error | null>>,
    rows,
    initialize,
    setLayout,
    press,
    resetModifiers,
    labelOf,
  };
}

/* Test internals --------------------------------------------------------- */
export const _internals = {
  STORAGE_KEY,
  DEFAULT_LAYOUT_ID,
  reset() {
    indexState.value = [];
    layoutCache.clear();
    activeId.value = DEFAULT_LAYOUT_ID;
    activeLayout.value = null;
    modifiers.value = { ...EMPTY_MODIFIERS };
    isLoading.value = false;
    lastError.value = null;
  },
};
