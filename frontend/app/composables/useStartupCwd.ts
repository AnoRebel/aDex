// composables/useStartupCwd.ts
//
// Resolves the user's "Open in" preference (home / launch-cwd) to a real
// path. Reads the saved Settings → System → initialCwd value (via VueUse
// `useStorage`, so it stays reactive cross-tab) and pairs it with the
// backend's `GetStartupPaths()` response.
//
// Consumers:
//   - pages/index.vue  — calls `resolve()` on first mount to seed the
//                        filesystem store + the initial terminal tab.
//   - AdexFilesystem    — could use it for an explicit "Reset CWD" button.

import { computed, ref } from 'vue'
import { useStorage } from '@vueuse/core'
import { GetStartupPaths } from '~/lib/wailsjs/coordinator'

export type InitialCwdMode = 'home' | 'cwd'

/** Slice of the persisted settings.json this composable cares about. */
interface PersistedSettings {
  system?: {
    initialCwd?: InitialCwdMode
  }
}

const STORAGE_KEY = 'adex-settings'

// Module-level cache of the backend-resolved paths. The data is stable
// for the process lifetime so we only ask once.
const cachedPaths = ref<{ home: string; cwd: string } | null>(null)

async function loadPaths(): Promise<{ home: string; cwd: string }> {
  if (cachedPaths.value) return cachedPaths.value
  try {
    const paths = await GetStartupPaths()
    cachedPaths.value = paths
    return paths
  } catch {
    // Backend unreachable (jsdom, browser preview). Return safe defaults
    // so callers never have to handle null.
    cachedPaths.value = { home: '/', cwd: '/' }
    return cachedPaths.value
  }
}

export function useStartupCwd() {
  // Reactive view onto the persisted settings blob. We read the whole
  // `adex-settings` object (the same key the modal writes) and project
  // out the one field we care about. `useStorage` gives us the storage-
  // event sync for free, so updates from the Settings modal land here
  // without a page reload.
  const settings = useStorage<PersistedSettings>(STORAGE_KEY, {})

  const mode = computed<InitialCwdMode>(() => {
    const m = settings.value?.system?.initialCwd
    return m === 'cwd' || m === 'home' ? m : 'home'
  })

  /**
   * Resolve the saved preference to a concrete path. Always returns a
   * non-empty string — falls back to '/' if both home and cwd are blank.
   */
  async function resolve(): Promise<string> {
    const paths = await loadPaths()
    const picked = mode.value === 'cwd' ? paths.cwd : paths.home
    return picked || paths.home || paths.cwd || '/'
  }

  return { resolve, mode, loadPaths }
}
