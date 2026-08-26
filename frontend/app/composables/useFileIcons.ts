// composables/useFileIcons.ts
//
// Wraps the upstream eDEX-UI file-icons assets (8000+ filename → icon
// mappings spanning Atom file-icons, FontAwesome brands, Devopicons,
// MFixx, Bytesize) into a Vue-friendly API.
//
// Sources:
//   frontend/app/assets/data/file-icons.json    (~3 MB, generated SVG bundle)
//   frontend/app/utils/file-icons-match.js      (~2.4k regex match table)
//
// The JSON is lazy-imported so it doesn't bloat the initial chunk; first
// call resolves the import, subsequent calls hit the in-memory cache.

import { ref, computed } from 'vue'

export interface FileIconRecord {
  width: number
  height: number
  /** Inner SVG markup (no <svg> wrapper). */
  svg: string
}

type IconBundle = Record<string, FileIconRecord>

const bundle = ref<IconBundle | null>(null)
const matcher = ref<((name: string) => string | undefined) | null>(null)
const loading = ref(false)
const lastError = ref<Error | null>(null)
const memo = new Map<string, FileIconRecord | null>()

async function ensureLoaded(): Promise<void> {
  if (bundle.value && matcher.value) return
  if (loading.value) {
    // Wait for the in-flight load.
    while (loading.value) {
      await new Promise((r) => setTimeout(r, 16))
    }
    return
  }
  loading.value = true
  lastError.value = null
  try {
    const [iconsMod, matchMod] = await Promise.all([
      import('~/assets/data/file-icons.json'),
      import('~/utils/file-icons-match.js'),
    ])
    // Vite/JSON imports normalize the module shape across dev and build:
    // both `default` and the JSON object root are valid.
    bundle.value = (iconsMod as { default?: IconBundle }).default ?? (iconsMod as unknown as IconBundle)
    const m = (matchMod as { default?: (name: string) => string | undefined }).default
    if (typeof m !== 'function') throw new Error('file-icons matcher missing default export')
    matcher.value = m
  } catch (err) {
    lastError.value = err instanceof Error ? err : new Error(String(err))
    throw lastError.value
  } finally {
    loading.value = false
  }
}

/**
 * Resolve a file icon for the given filename. Returns `null` when no
 * upstream rule matches — callers should fall back to a generic icon.
 *
 * The matcher applies path-aware rules (e.g. `node_modules`, `.git`)
 * AND extension rules, so callers should pass the full basename or
 * relative path rather than just the extension.
 */
export async function resolveFileIcon(filename: string): Promise<FileIconRecord | null> {
  if (!filename) return null
  if (memo.has(filename)) return memo.get(filename) ?? null
  await ensureLoaded()
  const id = matcher.value?.(filename)
  const rec = id && bundle.value ? bundle.value[id] ?? null : null
  memo.set(filename, rec)
  return rec
}

export function useFileIcons() {
  return {
    isLoading: computed(() => loading.value),
    lastError: computed(() => lastError.value),
    resolve: resolveFileIcon,
    /** Eagerly load the bundle. Useful for warming the cache while the
     *  user is on another panel. */
    preload: ensureLoaded,
  }
}
