// composables/useModules.ts
//
// Per-module visibility and ordering.
//
// The top-bar buttons toggle whole REGIONS (the left column, the bottom band).
// This is the finer-grained control: each individual panel can be hidden and
// the panels within a column can be reordered, from Settings.
//
// Deliberately separate from custom layouts: a custom layout defines a named
// arrangement in a file, whereas this is the user nudging the current one.

import { computed, readonly, type Ref } from 'vue'
import { useStorage } from '@vueuse/core'

export interface ModuleDef {
  /** Stable id — also the key used in persisted settings. */
  id: string
  /** Shown in the settings list. */
  label: string
  /** Which column the module lives in. */
  region: 'left' | 'right' | 'bottom'
}

/**
 * Every toggleable module, in default order.
 *
 * Ids are a compatibility surface: renaming one silently resets that module's
 * saved visibility and position, so treat them as fixed.
 */
export const MODULES: ModuleDef[] = [
  { id: 'clock',      label: 'Clock',          region: 'left' },
  { id: 'sysinfo',    label: 'System info',    region: 'left' },
  { id: 'hardware',   label: 'Hardware',       region: 'left' },
  { id: 'cpu',        label: 'CPU',            region: 'left' },
  { id: 'ram',        label: 'Memory',         region: 'left' },
  { id: 'toplist',    label: 'Processes',      region: 'left' },
  { id: 'netstat',    label: 'Network status', region: 'right' },
  { id: 'globe',      label: 'World view',     region: 'right' },
  { id: 'traffic',    label: 'Network traffic', region: 'right' },
  { id: 'filesystem', label: 'File manager',   region: 'bottom' },
  { id: 'keyboard',   label: 'On-screen keyboard', region: 'bottom' },
]

interface ModuleState {
  /** id -> visible. Absent means visible: a module added in a later release
   *  should appear rather than be hidden by an older saved state. */
  hidden: Record<string, boolean>
  /** Per-region id order. Ids missing from a saved order fall back to their
   *  position in MODULES, so a new module still lands somewhere sensible. */
  order: Partial<Record<'left' | 'right' | 'bottom', string[]>>
}

const state = useStorage<ModuleState>('adex.modules', { hidden: {}, order: {} })

export function useModules() {
  /** Modules for a region, in the user's order, hidden ones removed. */
  function visibleIn(region: 'left' | 'right' | 'bottom'): string[] {
    const defaults = MODULES.filter(m => m.region === region).map(m => m.id)
    const saved = state.value.order?.[region] ?? []
    // Saved order first, then anything new that is not in it yet.
    const ordered = [
      ...saved.filter(id => defaults.includes(id)),
      ...defaults.filter(id => !saved.includes(id)),
    ]
    return ordered.filter(id => !state.value.hidden?.[id])
  }

  /** All modules for a region in order, including hidden — for the settings list. */
  function allIn(region: 'left' | 'right' | 'bottom'): ModuleDef[] {
    const defaults = MODULES.filter(m => m.region === region)
    const saved = state.value.order?.[region] ?? []
    const byId = new Map(defaults.map(m => [m.id, m]))
    const ordered: ModuleDef[] = []
    for (const id of saved) {
      const m = byId.get(id)
      if (m) { ordered.push(m); byId.delete(id) }
    }
    for (const m of defaults) if (byId.has(m.id)) ordered.push(m)
    return ordered
  }

  function isHidden(id: string): boolean {
    return state.value.hidden?.[id] === true
  }

  function setHidden(id: string, hidden: boolean) {
    state.value = {
      ...state.value,
      hidden: { ...state.value.hidden, [id]: hidden },
    }
  }

  function toggle(id: string) {
    setHidden(id, !isHidden(id))
  }

  /** Move a module one place up or down within its own region. */
  function move(id: string, delta: -1 | 1) {
    const def = MODULES.find(m => m.id === id)
    if (!def) return
    const ids = allIn(def.region).map(m => m.id)
    const from = ids.indexOf(id)
    const to = from + delta
    if (from < 0 || to < 0 || to >= ids.length) return
    ids.splice(to, 0, ...ids.splice(from, 1))
    state.value = {
      ...state.value,
      order: { ...state.value.order, [def.region]: ids },
    }
  }

  function reset() {
    state.value = { hidden: {}, order: {} }
  }

  return {
    modules: MODULES,
    state: readonly(state) as Readonly<Ref<ModuleState>>,
    visibleIn,
    allIn,
    isHidden,
    setHidden,
    toggle,
    move,
    reset,
  }
}
