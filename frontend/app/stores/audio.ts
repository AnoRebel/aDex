import { defineStore } from 'pinia'
// V2 audio playback flows through `composables/useAdexAudio.ts` (Howler +
// synth fallback), not via Wails. The store intentionally does not call
// any backend audio bindings — `GetService('audio')` would marshal a Go
// struct that contains channels and an oto.Context, which Wails cannot
// JSON-encode, and would surface as a launch fatal:
//   FAT | json: unsupported type: func() error
//
// The methods below remain as stubs so existing callers continue to
// compile; real cue playback happens via `useAdexAudio().playCue(name)`.
import type {
  AudioEvent,
  AudioSettings,
  AudioPlaybackStatus,
  AudioStats,
  Soundpack,
  AudioEventType
} from '~/types/audio'

interface AudioState {
  // Sound effects system state
  settings: AudioSettings
  playbackStatus: AudioPlaybackStatus
  availableEvents: AudioEvent[]
  soundpacks: Soundpack[]
  stats: AudioStats | null
  isLoading: boolean
  error: string | null

  // UI state
  showSettings: boolean
  showEventTester: boolean

  // Cache for performance
  lastPlayedEvents: Record<string, number>
  preloadCache: Set<string>
}

export const useAudioStore = defineStore('audio', {
  state: (): AudioState => ({
    // Audio settings (disabled by default as per requirements)
    settings: {
      enabled: false,
      volume: 50,
      muted: false,
      soundpack: 'default',
      globalVolume: 0.5,
      effectsVolume: 0.5,
      notificationVolume: 0.3,
      muteInBackground: false,
      autoPlay: true,
      enabledEvents: [
        'system:startup',
        'terminal:bell',
        'terminal:command_error',
        'ui:button_click',
        'ui:notification',
        'file:operation_complete',
        'network:connected',
      ],
      version: '1.0.0',
      updatedAt: new Date().toISOString(),
    },

    // Playback status
    playbackStatus: {
      isPlaying: false,
      currentEvent: null,
      volume: 0.5,
      muted: false,
      loadedEvents: [],
      supportedFormats: ['mp3', 'wav', 'ogg', 'aac'],
    },

    // Available data
    availableEvents: [],
    soundpacks: [],
    stats: null,

    // Loading and error state
    isLoading: false,
    error: null,

    // UI state
    showSettings: false,
    showEventTester: false,

    // Performance cache
    lastPlayedEvents: {},
    preloadCache: new Set(),
  }),

  getters: {
    // Basic state getters
    isEnabled: (state) => state.settings.enabled && !state.settings.muted,
    effectiveVolume: (state) => {
      if (state.settings.muted) return 0
      return (state.settings.volume / 100) * state.settings.globalVolume
    },
    volumePercentage: (state) => {
      const effective = state.settings.muted ? 0 : (state.settings.volume / 100) * state.settings.globalVolume
      return Math.round(effective * 100)
    },

    // Event management
    enabledEvents: (state) => {
      return state.availableEvents.filter(event =>
        state.settings.enabledEvents.includes(event.id)
      )
    },
    eventsByCategory: (state) => {
      const grouped: Record<AudioEventType, AudioEvent[]> = {
        system: [],
        interaction: [],
        notification: [],
        error: [],
        success: [],
      }

      state.availableEvents.forEach(event => {
        if (grouped[event.category]) {
          grouped[event.category].push(event)
        }
      })

      return grouped
    },

    // Soundpack management
    currentSoundpack: (state) => {
      return state.soundpacks.find(sp => sp.id === state.settings.soundpack)
    },
    builtinSoundpacks: (state) => {
      return state.soundpacks.filter(sp => sp.isBuiltin)
    },
    customSoundpacks: (state) => {
      return state.soundpacks.filter(sp => !sp.isBuiltin)
    },

    // Event checking helpers
    //
    // OLD behavior gated on `enabledEvents.includes(eventId)` which
    // defaulted to an EMPTY list on first launch — so every cue was a
    // silent no-op until the user explicitly opted in (which there's
    // no UI for). The cue vocabulary is small and curated; the global
    // enabled/muted flags are the right gate, not a per-event
    // allowlist. If we want opt-out per cue later, do it by listing
    // DISABLED events instead of required-enabled events.
    isEventEnabled: (state) => {
      return (_eventId: string) => {
        return state.settings.enabled && !state.settings.muted
      }
    },
    // Per-event cooldown is enforced at the playback layer
    // (useAdexAudio's RATE_LIMIT_MS map) where it can be tuned per cue.
    // The store-level 100ms blanket cooldown was the main reason boot
    // and shutdown sequences dropped half their cues — five steps
    // firing in 350ms with the same eventId only emitted 1 audible
    // tone. Always allow at this layer; rate-limiting that matters
    // happens downstream.
    canPlayEvent: () => {
      return (_eventId: string, _cooldownMs: number = 0) => true
    },

    // Playback helpers
    currentlyPlaying: (state) => state.playbackStatus.currentEvent,
    hasLoadedEvents: (state) => state.playbackStatus.loadedEvents.length > 0,

    // Volume icon based on effective volume
    volumeIcon: (state) => {
      const effective = state.settings.muted ? 0 : (state.settings.volume / 100) * state.settings.globalVolume

      if (state.settings.muted || effective === 0) {
        return 'volume-off'
      }
      if (effective < 0.33) {
        return 'volume-low'
      }
      if (effective < 0.66) {
        return 'volume-medium'
      }
      return 'volume-high'
    },

    // Event lookup helpers
    eventById: (state) => {
      return (eventId: string) => state.availableEvents.find(e => e.id === eventId)
    },
    eventsByCategoryName: (state) => {
      return (category: AudioEventType) => state.availableEvents.filter(e => e.category === category)
    },
    soundpackById: (state) => {
      return (soundpackId: string) => state.soundpacks.find(sp => sp.id === soundpackId)
    },

    // Statistics helpers
    totalPlaytime: (state) => state.stats?.totalPlaytime || 0,
    mostPlayedEvent: (state) => state.stats?.mostPlayedEvent || '',
    totalPlays: (state) => state.stats?.totalPlays || 0,

    // Loading states
    isReady: (state) => !state.isLoading && !state.error,
    hasError: (state) => !!state.error,

    // Cache helpers
    isEventPreloaded: (state) => {
      return (eventId: string) => state.preloadCache.has(eventId)
    },
  },

  actions: {
    // Data loading actions
    async loadAvailableEvents(events: AudioEvent[]) {
      this.availableEvents = events
    },

    async loadSoundpacks(soundpacks: Soundpack[]) {
      this.soundpacks = soundpacks
    },

    async loadPlaybackStatus(status: Partial<AudioPlaybackStatus>) {
      this.playbackStatus = { ...this.playbackStatus, ...status }
    },

    async loadStats(stats: AudioStats) {
      this.stats = stats
    },

    // Settings management
    updateSettings(newSettings: Partial<AudioSettings>) {
      this.settings = { ...this.settings, ...newSettings, updatedAt: new Date().toISOString() }
      this.saveToStorage()
    },

    setVolume(volume: number) {
      const clampedVolume = Math.max(0, Math.min(100, volume))
      this.updateSettings({ volume: clampedVolume })
    },

    setGlobalVolume(volume: number) {
      const clampedVolume = Math.max(0, Math.min(1, volume))
      this.updateSettings({ globalVolume: clampedVolume })
    },

    setEffectsVolume(volume: number) {
      const clampedVolume = Math.max(0, Math.min(1, volume))
      this.updateSettings({ effectsVolume: clampedVolume })
    },

    setNotificationVolume(volume: number) {
      const clampedVolume = Math.max(0, Math.min(1, volume))
      this.updateSettings({ notificationVolume: clampedVolume })
    },

    setMuted(muted: boolean) {
      this.updateSettings({ muted })
    },

    toggleMute() {
      this.setMuted(!this.settings.muted)
    },

    setEnabled(enabled: boolean) {
      this.updateSettings({ enabled })
    },

    toggleEnabled() {
      this.setEnabled(!this.settings.enabled)
    },

    changeSoundpack(soundpackId: string) {
      this.updateSettings({ soundpack: soundpackId })
      // Clear preload cache when switching soundpacks
      this.preloadCache.clear()
    },

    // Event management
    enableEvent(eventId: string) {
      if (!this.settings.enabledEvents.includes(eventId)) {
        this.settings.enabledEvents.push(eventId)
        this.updateSettings({})
      }
    },

    disableEvent(eventId: string) {
      const index = this.settings.enabledEvents.indexOf(eventId)
      if (index > -1) {
        this.settings.enabledEvents.splice(index, 1)
        this.updateSettings({})
      }
    },

    toggleEvent(eventId: string) {
      if (this.settings.enabledEvents.includes(eventId)) {
        this.disableEvent(eventId)
      } else {
        this.enableEvent(eventId)
      }
    },

    enableMultipleEvents(eventIds: string[]) {
      eventIds.forEach(id => {
        if (!this.settings.enabledEvents.includes(id)) {
          this.settings.enabledEvents.push(id)
        }
      })
      this.updateSettings({})
    },

    disableMultipleEvents(eventIds: string[]) {
      const toRemove = new Set(eventIds)
      this.settings.enabledEvents = this.settings.enabledEvents.filter(id => !toRemove.has(id))
      this.updateSettings({})
    },

    enableAllEvents() {
      this.settings.enabledEvents = this.availableEvents.map(e => e.id)
      this.updateSettings({})
    },

    disableAllEvents() {
      this.settings.enabledEvents = []
      this.updateSettings({})
    },

    // Playback actions
    startPlayback(event: AudioEvent) {
      this.playbackStatus.isPlaying = true
      this.playbackStatus.currentEvent = event
      this.lastPlayedEvents[event.id] = Date.now()
    },

    stopPlayback() {
      this.playbackStatus.isPlaying = false
      this.playbackStatus.currentEvent = null
    },

    playbackError(error: string) {
      this.error = error
      this.stopPlayback()
    },

    // Cache management
    preloadEvent(eventId: string) {
      this.preloadCache.add(eventId)
    },

    preloadEvents(eventIds: string[]) {
      eventIds.forEach(id => this.preloadCache.add(id))
    },

    clearPreloadCache() {
      this.preloadCache.clear()
    },

    // UI state management
    setShowSettings(show: boolean) {
      this.showSettings = show
    },

    setShowEventTester(show: boolean) {
      this.showEventTester = show
    },

    // Loading and error state
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
    },

    setError(error: string | null) {
      this.error = error
    },

    clearError() {
      this.error = null
    },

    // Storage management
    saveToStorage() {
      if (typeof localStorage !== 'undefined') {
        try {
          localStorage.setItem('adex-audio-settings', JSON.stringify(this.settings))
        } catch (error) {
          console.error('Failed to save audio settings:', error)
        }
      }
    },

    loadFromStorage() {
      if (typeof localStorage !== 'undefined') {
        try {
          const saved = localStorage.getItem('adex-audio-settings')
          if (saved) {
            const settings = JSON.parse(saved)
            this.settings = { ...this.settings, ...settings }
          }
        } catch (error) {
          console.error('Failed to load audio settings:', error)
        }
      }
    },

    // Utility actions
    resetToDefaults() {
      this.settings = {
        enabled: false, // Keep disabled by default
        volume: 50,
        muted: false,
        soundpack: 'default',
        globalVolume: 0.5,
        effectsVolume: 0.5,
        notificationVolume: 0.3,
        muteInBackground: false,
        autoPlay: true,
        enabledEvents: [
          'system:startup',
          'terminal:bell',
          'terminal:command_error',
          'ui:button_click',
          'ui:notification',
          'file:operation_complete',
          'network:connected',
        ],
        version: '1.0.0',
        updatedAt: new Date().toISOString(),
      }
      this.saveToStorage()
    },

    // Statistics helpers
    incrementPlayCount(eventId: string) {
      if (this.stats) {
        this.stats.totalPlays++
        if (this.stats.eventsPlayed) {
          this.stats.eventsPlayed[eventId] = (this.stats.eventsPlayed[eventId] || 0) + 1
        }
        this.stats.lastPlayed = new Date()
      }
    },

    // Initialize store
    initialize() {
      this.loadFromStorage()
    },

    // Stubs — the legacy backend audio bindings (GetAvailableEvents /
    // GetSoundpacks / GetStats / PlayEvent) were never wired through the
    // coordinator and calling them via GetService('audio') crashes Wails
    // (see file header). Until a frontend-side composable replaces them
    // (section 6.7 wires the V2 audio composable into the Settings tab),
    // these methods leave state untouched.
    async fetchAvailableEvents(): Promise<void> {
      // Events are populated client-side from soundpacks.json by the
      // V2 audio composable (`useAdexAudio`).
    },

    async fetchSoundpacks(): Promise<void> {
      // Soundpacks are loaded by `composables/useAdexAudio.ts`.
    },

    async fetchStats(): Promise<void> {
      // Stats are an in-memory client concept now; no backend round-trip.
    },

    async playEvent(eventId: string): Promise<void> {
      // Gate on the master enable/mute flags only. The per-event
      // allowlist + 100ms cooldown previously enforced here were the
      // primary cause of dropped cues during the boot / shutdown
      // sequences — five cues firing in 350ms with overlapping IDs
      // would only emit 1-2 audible tones. Cue-level rate-limiting,
      // when needed, lives in useAdexAudio's RATE_LIMIT_MS map.
      if (!this.isEventEnabled(eventId)) return

      // We DELIBERATELY do not wrap playback in startPlayback/
      // stopPlayback (which mutate a single global "currentEvent"
      // ref). Web Audio supports concurrent playback — when cue B
      // fires while cue A is still ringing, stopPlayback() from cue
      // B's finally would mark "nothing playing" mid-A. That state
      // ref was load-bearing for nothing useful, so we just stop
      // touching it on the fast path. The fields remain for any
      // long-form playback UI a future settings panel might want.
      try {
        const event = this.eventById(eventId)
        if (!event) return
        this.lastPlayedEvents[eventId] = Date.now()
        this.incrementPlayCount(eventId)
        // Route through the V2 cue system — Howler-backed with a synth
        // fallback. Imported lazily so SSR doesn't try to resolve Howler.
        const { useAdexAudio } = await import('~/composables/useAdexAudio')
        useAdexAudio().playCue(eventId)
      } catch (error) {
        console.error('Failed to play audio event:', error)
        this.playbackError(
          `Failed to play event: ${error instanceof Error ? error.message : 'Unknown error'}`,
        )
      }
    },

    async initializeFromBackend(): Promise<void> {
      this.loadFromStorage()
      await Promise.all([
        this.fetchAvailableEvents(),
        this.fetchSoundpacks(),
        this.fetchStats()
      ])
    },

    // Cleanup
    cleanup() {
      this.stopPlayback()
      this.clearError()
      this.clearPreloadCache()
    },
  },
})
