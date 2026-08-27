// composables/useCustomLayouts.ts
//
// User-defined layouts. Built-in presets are CSS files keyed on
// `data-layout`; users cannot add CSS to a bundled app, so custom layouts are
// DATA instead: a definition names which panels go in which region and in
// what order, and the shell renders from that.
//
// Built-in presets are untouched by this — a user with no custom layouts sees
// exactly the previous behaviour.

import { ref, computed, readonly, type Ref } from 'vue'
import { GetCustomLayouts, GetCustomLayoutsPath } from '~/lib/wailsjs/coordinator'

/** Regions a panel can be placed in. These map to the shell's containers. */
export const LAYOUT_REGIONS = ['left', 'centre', 'right', 'bottom'] as const
export type LayoutRegion = (typeof LAYOUT_REGIONS)[number]

/**
 * Panels a layout may reference, by stable id. The id is what users write in
 * their configuration file, so these names are a compatibility surface: renaming
 * one breaks existing user layouts.
 */
export const LAYOUT_PANELS = [
  'clock',
  'sysinfo',
  'hardware',
  'cpu',
  'ram',
  'toplist',
  'terminal',
  'filesystem',
  'netstat',
  'globe',
  'traffic',
  'keyboard',
] as const
export type LayoutPanel = (typeof LAYOUT_PANELS)[number]

export interface CustomLayout {
  /** Unique id, used as the persisted layout value. */
  id: string
  /** Shown in the settings selector. */
  displayName?: string
  /** Panels per region, in render order. */
  regions: Partial<Record<LayoutRegion, LayoutPanel[]>>
}

const customLayouts = ref<CustomLayout[]>([])
const configPath = ref<string>('')
const loadErrors = ref<string[]>([])

function isPanel(v: unknown): v is LayoutPanel {
  return typeof v === 'string' && (LAYOUT_PANELS as readonly string[]).includes(v)
}
function isRegion(v: string): v is LayoutRegion {
  return (LAYOUT_REGIONS as readonly string[]).includes(v)
}

/**
 * Validate one definition, dropping what the application cannot honour rather
 * than rejecting the whole layout. Returns null when nothing usable remains.
 * Problems are collected so the user can be told what to fix.
 */
export function validateLayout(raw: unknown, errors: string[]): CustomLayout | null {
  if (!raw || typeof raw !== 'object') {
    errors.push('layout entry is not an object')
    return null
  }
  const o = raw as Record<string, unknown>
  const id = typeof o.id === 'string' ? o.id.trim() : ''
  if (!id) {
    errors.push('layout is missing an "id"')
    return null
  }

  const regionsRaw = o.regions
  if (!regionsRaw || typeof regionsRaw !== 'object') {
    errors.push(`layout "${id}" has no "regions"`)
    return null
  }

  const regions: Partial<Record<LayoutRegion, LayoutPanel[]>> = {}
  for (const [region, panels] of Object.entries(regionsRaw as Record<string, unknown>)) {
    if (!isRegion(region)) {
      errors.push(`layout "${id}": unknown region "${region}"`)
      continue
    }
    if (!Array.isArray(panels)) {
      errors.push(`layout "${id}": region "${region}" is not a list`)
      continue
    }
    const valid = panels.filter((p) => {
      if (isPanel(p)) return true
      errors.push(`layout "${id}": unknown panel "${String(p)}"`)
      return false
    }) as LayoutPanel[]
    regions[region] = valid
  }

  if (Object.keys(regions).length === 0) {
    errors.push(`layout "${id}": no usable regions`)
    return null
  }

  return {
    id,
    displayName: typeof o.displayName === 'string' ? o.displayName : id,
    regions,
  }
}

export interface UseCustomLayoutsApi {
  layouts: Readonly<Ref<CustomLayout[]>>
  path: Readonly<Ref<string>>
  errors: Readonly<Ref<string[]>>
  load(): Promise<void>
  find(id: string): CustomLayout | undefined
  isCustom(id: string): boolean
}

export function useCustomLayouts(): UseCustomLayoutsApi {
  async function load() {
    const errors: string[] = []
    try {
      const [raw, path] = await Promise.all([GetCustomLayouts(), GetCustomLayoutsPath()])
      configPath.value = typeof path === 'string' ? path : ''

      const list = Array.isArray(raw) ? raw : []
      const parsed: CustomLayout[] = []
      const seen = new Set<string>()
      for (const entry of list) {
        const layout = validateLayout(entry, errors)
        if (!layout) continue
        if (seen.has(layout.id)) {
          errors.push(`duplicate layout id "${layout.id}" — keeping the first`)
          continue
        }
        seen.add(layout.id)
        parsed.push(layout)
      }
      customLayouts.value = parsed
    } catch (err) {
      // A missing or unreadable config is not an error condition: the user
      // simply has no custom layouts. Only record it for diagnostics.
      errors.push(err instanceof Error ? err.message : String(err))
      customLayouts.value = []
    }
    loadErrors.value = errors
    if (errors.length) {
      console.warn('[custom-layouts] problems in the layout configuration:', errors)
    }
  }

  return {
    layouts: readonly(customLayouts) as Readonly<Ref<CustomLayout[]>>,
    path: readonly(configPath) as Readonly<Ref<string>>,
    errors: readonly(loadErrors) as Readonly<Ref<string[]>>,
    load,
    find: (id: string) => customLayouts.value.find((l) => l.id === id),
    isCustom: (id: string) => customLayouts.value.some((l) => l.id === id),
  }
}
