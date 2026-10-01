import { watch } from 'vue'
import { useStorage, useDebounceFn } from '@vueuse/core'
import { GetUISettings, SaveUISettings } from '~/lib/wailsjs/coordinator'

/**
 * File-backed settings persistence.
 *
 * Architecture (decided with the user): the Go side
 * (`internal/services/settings.UIStore`, file
 * `~/.config/aDex-UI/adex-ui-settings.json`) is the SOURCE OF TRUTH.
 * The `adex-settings` localStorage entry — which every settings-aware
 * component already reads via `useStorage('adex-settings')` — is only
 * a fast, reactive CACHE.
 *
 *   boot:   backend file  ──hydrate──▶  localStorage cache  ──▶  UI
 *   change: UI  ──▶  localStorage cache  ──debounce 500ms──▶  backend file
 *
 * On a boot conflict the BACKEND WINS: whatever is on disk overwrites
 * the cache. The only exception is first run on a machine that already
 * has a populated localStorage cache from a pre-persistence build —
 * there we migrate the cache UP to the backend instead of clobbering
 * it with defaults (see hydrate()).
 *
 * Wails-absent (plain `nuxt dev` in a browser, tests): every backend
 * call throws synchronously from getCoordinator(). We swallow that and
 * fall back to localStorage-only, i.e. the exact pre-persistence
 * behaviour. No feature is lost in browser dev; it just isn't durable.
 *
 * This composable is shape-agnostic: it round-trips the whole
 * `adex-settings` object as opaque JSON. The TYPED contract lives on
 * the Go side (models.UISettings) per the user's choice of typed
 * reconciliation over an opaque blob — that's where a malformed or
 * partial payload gets normalised against defaults.
 */

const STORAGE_KEY = 'adex-settings'
// Matches AdexSettingsModal's getDefaultSettings() debounce; long
// enough to coalesce a slider drag / rapid toggling into one write,
// short enough that a normal "change one thing and close" persists
// before the window can be quit.
const FLUSH_DEBOUNCE_MS = 500

// useStorage with no default + no mergeDefaults: this composable never
// owns the defaults. AdexSettingsModal already created the ref with
// getDefaultSettings() + mergeDefaults; we attach to the SAME key so
// we share that single reactive object and don't fight over shape.
type AnySettings = Record<string, unknown>

let hydrated = false
let backendAvailable = true

function readBackend(): Promise<AnySettings | null> {
  // GetUISettings returns the typed models.UISettings as a plain
  // object (Wails JSON-marshals it). It carries extra metadata keys
  // (version, updatedAt) the frontend simply ignores.
  return GetUISettings()
}

function writeBackend(value: unknown): Promise<void> {
  return SaveUISettings(JSON.stringify(value))
}

/**
 * Pull the on-disk settings into the localStorage cache exactly once
 * per app session. Call early in app boot (before settings-reading
 * panels mount their watchers) so the first paint reflects persisted
 * state, not stale cache.
 */
async function hydrate(): Promise<void> {
  if (hydrated) return
  hydrated = true

  const cache = useStorage<AnySettings>(STORAGE_KEY, {})

  let backend: AnySettings | null = null
  try {
    backend = await readBackend()
  } catch {
    // Wails not present (browser dev / unit test) — localStorage
    // remains the only store. This is the documented degraded mode.
    backendAvailable = false
    return
  }

  if (!backend || typeof backend !== 'object') {
    return
  }

  // First-run migration: the backend file is brand new (no UpdatedAt
  // yet, because the store only stamps it on the first real Save) but
  // the user already has a populated cache from a pre-persistence
  // build. Push the cache UP rather than overwriting it with the
  // backend's freshly-written defaults.
  const backendIsFresh = !backend.updatedAt
  const cacheHasUserData =
    cache.value && typeof cache.value === 'object' && Object.keys(cache.value).length > 0

  if (backendIsFresh && cacheHasUserData) {
    try {
      await writeBackend(cache.value)
    } catch {
      backendAvailable = false
    }
    return
  }

  // Normal path: backend wins. Strip the backend-only metadata so the
  // cache keeps the exact frontend shape every panel expects.
  const { version: _v, updatedAt: _u, ...frontendShape } = backend
  cache.value = frontendShape as AnySettings
}

/**
 * Install the debounced cache→backend flush. Idempotent per call site
 * but intended to be installed exactly once at app root. The deep
 * watch covers every nested settings panel because they all mutate
 * this one shared `adex-settings` object.
 */
function installFlush(): void {
  const cache = useStorage<AnySettings>(STORAGE_KEY, {})

  const flush = useDebounceFn(async () => {
    if (!backendAvailable) return
    try {
      await writeBackend(cache.value)
    } catch {
      // Lost Wails mid-session (shouldn't happen in the desktop app);
      // stop hammering a throwing binding.
      backendAvailable = false
    }
  }, FLUSH_DEBOUNCE_MS)

  watch(
    cache,
    () => {
      // Don't flush the hydration write back to the backend on the
      // same tick — hydrate() either already matches the backend or
      // just performed the migration write itself.
      if (!hydrated) return
      flush()
    },
    { deep: true },
  )
}

/**
 * One-call bootstrap: hydrate from disk, then start the debounced
 * flush. Safe to await; the flush watcher is installed only after
 * hydration so the backend's own value isn't immediately written
 * back as if it were a user edit.
 */
export function useSettingsPersistence() {
  async function init(): Promise<void> {
    await hydrate()
    installFlush()
  }

  // Force an immediate, un-debounced write — used right before the app
  // quits so an edit made in the last 500ms isn't lost to the pending
  // debounce timer.
  async function flushNow(): Promise<void> {
    if (!backendAvailable) return
    const cache = useStorage<AnySettings>(STORAGE_KEY, {})
    try {
      await writeBackend(cache.value)
    } catch {
      backendAvailable = false
    }
  }

  return { init, flushNow, get backendAvailable() { return backendAvailable } }
}
