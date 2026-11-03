// composables/useUpdateChecker.ts
//
// Frontend half of the in-app updater. The Go updater service does all the
// network, signature and install work; this composable decides WHEN to check
// (the user's frequency and skip list) and turns the updater's events into
// state the modal and Settings can render.
//
// Persisted (useStorage, survives restarts):
//   - settings.enabled         : automatic checks on/off
//   - settings.interval        : 'manual' | 'launch' | 'daily' | 'weekly'
//   - settings.skippedVersions : versions the user chose to skip
//   - state.lastChecked / lastAttempted / lastError
//
// Session only: the found release, progress and the flow state. An install
// always follows a check in the same session, so there is nothing to cache.

import { computed, readonly, ref, type Ref } from 'vue'
import { useStorage } from '@vueuse/core'
import { Browser, Events, Updater } from '@wailsio/runtime'
import * as UpdaterService from '../../bindings/aDex/internal/services/updater/service.js'
import type { Info, Release } from '../../bindings/aDex/internal/services/updater/models.js'

export type { Info, Release }

declare const __APP_VERSION__: string | undefined
const BUILD_VERSION: string = typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : '0.0.0'

export type CheckInterval = 'manual' | 'launch' | 'daily' | 'weekly'

export type UpdatePhase =
  | 'idle' | 'checking' | 'up-to-date' | 'available'
  | 'downloading' | 'verifying' | 'installing' | 'ready' | 'error'

export interface UpdateSettings {
  enabled: boolean
  interval: CheckInterval
  skippedVersions: string[]
}

export interface UpdateState {
  /** ISO timestamp of the last successful check. */
  lastChecked: string
  /** ISO timestamp of the last attempt, successful or not. */
  lastAttempted: string
  lastError: string
}

export interface UpdateProgress { written: number; total: number; rate: number }
export interface UpdateError { stage: string; message: string }

const DAY_MS = 24 * 60 * 60 * 1000

/* ---- Module-scope state: every consumer shares the same refs ---------- */

const settings = useStorage<UpdateSettings>(
  'adex.updates.settings',
  { enabled: true, interval: 'launch', skippedVersions: [] },
  undefined,
  { mergeDefaults: true },
)

const state = useStorage<UpdateState>(
  'adex.updates.state',
  { lastChecked: '', lastAttempted: '', lastError: '' },
  undefined,
  { mergeDefaults: true },
)

const phase = ref<UpdatePhase>('idle')
const release = ref<Release | null>(null)
const progress = ref<UpdateProgress | null>(null)
const error = ref<UpdateError | null>(null)
const info = ref<Info | null>(null)

let subscribed = false

function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

/** Download/verify/install progress only arrives as events, because
 *  Install() resolves once the whole flow has finished. */
function subscribe() {
  if (subscribed) return
  subscribed = true
  const E = Updater.Events
  Events.On(E.DownloadStarted, () => {
    phase.value = 'downloading'
    progress.value = null
    error.value = null
  })
  Events.On(E.DownloadProgress, (e: { data?: UpdateProgress }) => {
    if (e.data) progress.value = e.data
  })
  Events.On(E.Verifying, () => { phase.value = 'verifying' })
  Events.On(E.Installing, () => { phase.value = 'installing' })
  Events.On(E.UpdateReady, () => { phase.value = 'ready' })
  Events.On(E.Error, (e: { data?: UpdateError }) => {
    error.value = e.data ?? { stage: '', message: 'update failed' }
    phase.value = 'error'
  })
}

async function refreshInfo(): Promise<Info | null> {
  try {
    info.value = await UpdaterService.Info()
  } catch (err) {
    console.warn('[updates] Info failed:', err)
  }
  return info.value
}

/* ---- Public API ------------------------------------------------------- */

export function useUpdateChecker() {
  subscribe()

  const currentVersion = computed(() => info.value?.currentVersion || BUILD_VERSION)
  const canSelfUpdate = computed(() => info.value?.canSelfUpdate === true)
  const checking = computed(() => phase.value === 'checking')
  /** A newer release the user hasn't skipped. */
  const hasUpdate = computed(() =>
    !!release.value && !settings.value.skippedVersions.includes(release.value.version),
  )

  /** Check now, whatever the frequency setting says. Resolves to the
   *  newer release, or null when up to date or the check failed. */
  async function check(): Promise<Release | null> {
    if (checking.value) return release.value
    phase.value = 'checking'
    error.value = null
    const attempted = new Date().toISOString()
    state.value.lastAttempted = attempted
    try {
      const rel = await UpdaterService.Check()
      release.value = rel ?? null
      phase.value = rel ? 'available' : 'up-to-date'
      state.value.lastChecked = attempted
      state.value.lastError = ''
      return release.value
    } catch (err) {
      const msg = message(err)
      state.value.lastError = msg
      error.value = { stage: 'check', message: msg }
      phase.value = 'error'
      return null
    } finally {
      void refreshInfo()
    }
  }

  /** Run a check if the user's frequency setting says one is due. */
  async function maybeCheckOnLaunch(): Promise<void> {
    const { enabled, interval } = settings.value
    if (!enabled || interval === 'manual') return
    if (interval !== 'launch') {
      const last = state.value.lastChecked ? Date.parse(state.value.lastChecked) : 0
      const threshold = interval === 'daily' ? DAY_MS : 7 * DAY_MS
      if (Date.now() - (Number.isNaN(last) ? 0 : last) < threshold) return
    }
    await check()
  }

  /** Download, verify and stage the found release. Never runs for installs
   *  that can't replace themselves; those get openReleasePage instead. */
  async function install(): Promise<void> {
    if (!release.value) return
    const i = info.value ?? (await refreshInfo())
    if (!i?.canSelfUpdate) return
    error.value = null
    try {
      await UpdaterService.Install()
    } catch (err) {
      // The Error event normally carries the stage; this covers a refusal
      // that happens before the updater starts.
      if (phase.value !== 'error') {
        error.value = { stage: 'install', message: message(err) }
        phase.value = 'error'
      }
    }
  }

  async function restart(): Promise<void> {
    try {
      await UpdaterService.Restart()
    } catch (err) {
      error.value = { stage: 'install', message: message(err) }
      phase.value = 'error'
    }
  }

  function openReleasePage(): void {
    const url = release.value?.url
    if (url) void Browser.OpenURL(url).catch(err => console.warn('[updates] open failed:', err))
  }

  function skipVersion(version: string): void {
    if (version && !settings.value.skippedVersions.includes(version)) {
      settings.value.skippedVersions.push(version)
    }
  }

  function unskipAll(): void {
    settings.value.skippedVersions = []
  }

  if (!info.value) void refreshInfo()

  return {
    settings,
    state,
    phase: readonly(phase) as Readonly<Ref<UpdatePhase>>,
    release: readonly(release) as Readonly<Ref<Release | null>>,
    progress: readonly(progress) as Readonly<Ref<UpdateProgress | null>>,
    error: readonly(error) as Readonly<Ref<UpdateError | null>>,
    info: readonly(info) as Readonly<Ref<Info | null>>,
    currentVersion,
    canSelfUpdate,
    checking,
    hasUpdate,
    check,
    maybeCheckOnLaunch,
    install,
    restart,
    openReleasePage,
    skipVersion,
    unskipAll,
  }
}
