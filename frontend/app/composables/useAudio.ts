import { ref, computed, watch, readonly } from 'vue'
import { useWails } from './useWails'
import type {
  AudioEvent,
  AudioSettings,
  AudioPlaybackStatus,
  AudioStats,
  Soundpack,
  AudioEventType
} from '~/types/audio'

export const useAudio = () => {
  const wails = useWails()

  // State
  const settings = ref<AudioSettings>({
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
  })

  const playbackStatus = ref<AudioPlaybackStatus>({
    isPlaying: false,
    currentEvent: null,
    volume: 0.5,
    muted: false,
    loadedEvents: [],
    supportedFormats: ['mp3', 'wav', 'ogg', 'aac'],
  })

  const availableEvents = ref<AudioEvent[]>([])
  const soundpacks = ref<Soundpack[]>([])
  const stats = ref<AudioStats | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed properties
  const isEnabled = computed(() => settings.value.enabled && !settings.value.muted)
  const effectiveVolume = computed(() => {
    if (settings.value.muted) return 0
    return (settings.value.volume / 100) * settings.value.globalVolume
  })

  const eventsByCategory = computed(() => {
    const grouped: Record<AudioEventType, AudioEvent[]> = {
      system: [],
      interaction: [],
      notification: [],
      error: [],
      success: [],
    }

    availableEvents.value.forEach(event => {
      if (grouped[event.category]) {
        grouped[event.category].push(event)
      }
    })

    return grouped
  })

  const enabledEvents = computed(() => {
    return availableEvents.value.filter(event =>
      settings.value.enabledEvents.includes(event.id)
    )
  })

  const isEventEnabled = (eventId: string) => {
    return settings.value.enabled &&
           !settings.value.muted &&
           settings.value.enabledEvents.includes(eventId)
  }

  const currentSoundpack = computed(() =>
    soundpacks.value.find(sp => sp.id === settings.value.soundpack)
  )

  // Initialize audio system
  const initialize = async () => {
    if (!wails.isReady.value) {
      console.warn('Wails not ready for audio operations')
      return
    }

    try {
      isLoading.value = true
      error.value = null

      // Load audio settings, events, and soundpacks
      await Promise.all([
        loadSettings(),
        loadAvailableEvents(),
        loadSoundpacks(),
        loadPlaybackStatus(),
      ])

      console.log('Audio system initialized successfully')
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to initialize audio system'
      console.error('Failed to initialize audio system:', err)
    } finally {
      isLoading.value = false
    }
  }

  // Load audio settings
  const loadSettings = async () => {
    if (!wails.isReady.value) return

    try {
      const loadedSettings = await wails.call('audio.GetSettings')
      if (loadedSettings) {
        settings.value = { ...settings.value, ...loadedSettings }
      }
    } catch (err) {
      console.error('Failed to load audio settings:', err)
    }
  }

  // Load available audio events
  const loadAvailableEvents = async () => {
    if (!wails.isReady.value) return

    try {
      const events = await wails.call('audio.GetAvailableEvents')
      if (events && Array.isArray(events)) {
        availableEvents.value = events
      }
    } catch (err) {
      console.error('Failed to load audio events:', err)
    }
  }

  // Load available soundpacks
  const loadSoundpacks = async () => {
    if (!wails.isReady.value) return

    try {
      const packs = await wails.call('audio.GetSoundpacks')
      if (packs && Array.isArray(packs)) {
        soundpacks.value = packs
      }
    } catch (err) {
      console.error('Failed to load soundpacks:', err)
    }
  }

  // Load playback status
  const loadPlaybackStatus = async () => {
    if (!wails.isReady.value) return

    try {
      const status = await wails.call('audio.GetPlaybackStatus')
      if (status) {
        playbackStatus.value = { ...playbackStatus.value, ...status }
      }
    } catch (err) {
      console.error('Failed to load playback status:', err)
    }
  }

  // Save audio settings
  const saveSettings = async () => {
    if (!wails.isReady.value) return false

    try {
      settings.value.updatedAt = new Date().toISOString()
      const success = await wails.call('audio.UpdateSettings', settings.value)
      return success
    } catch (err) {
      console.error('Failed to save audio settings:', err)
      return false
    }
  }

  // Play audio event
  const playEvent = async (eventId: string) => {
    if (!wails.isReady.value || !isEventEnabled(eventId)) {
      return false
    }

    try {
      error.value = null
      const success = await wails.call('audio.PlayEvent', eventId)

      if (success) {
        // Update playback status
        playbackStatus.value.isPlaying = true
        const event = availableEvents.value.find(e => e.id === eventId)
        if (event) {
          playbackStatus.value.currentEvent = event
        }
      }

      return success
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to play audio event'
      console.error('Failed to play audio event:', err)
      return false
    }
  }

  // Play UI event sound (maps UI events to audio events)
  const playUIEvent = async (uiEvent: string) => {
    if (!wails.isReady.value || !settings.value.enabled) {
      return false
    }

    try {
      const success = await wails.call('audio.PlayUIEvent', uiEvent)
      return success
    } catch (err) {
      console.error('Failed to play UI event sound:', err)
      return false
    }
  }

  // Stop current playback
  const stopPlayback = async () => {
    if (!wails.isReady.value) return false

    try {
      const success = await wails.call('audio.StopPlayback')
      if (success) {
        playbackStatus.value.isPlaying = false
        playbackStatus.value.currentEvent = null
      }
      return success
    } catch (err) {
      console.error('Failed to stop playback:', err)
      return false
    }
  }

  // Update master volume
  const setVolume = async (volume: number) => {
    const clampedVolume = Math.max(0, Math.min(100, volume))
    settings.value.volume = clampedVolume
    return await saveSettings()
  }

  // Toggle mute state
  const toggleMute = async () => {
    settings.value.muted = !settings.value.muted
    return await saveSettings()
  }

  // Set mute state
  const setMute = async (muted: boolean) => {
    settings.value.muted = muted
    return await saveSettings()
  }

  // Update global volume
  const setGlobalVolume = async (volume: number) => {
    const clampedVolume = Math.max(0, Math.min(1, volume))
    settings.value.globalVolume = clampedVolume
    return await saveSettings()
  }

  // Update effects volume
  const setEffectsVolume = async (volume: number) => {
    const clampedVolume = Math.max(0, Math.min(1, volume))
    settings.value.effectsVolume = clampedVolume
    return await saveSettings()
  }

  // Update notification volume
  const setNotificationVolume = async (volume: number) => {
    const clampedVolume = Math.max(0, Math.min(1, volume))
    settings.value.notificationVolume = clampedVolume
    return await saveSettings()
  }

  // Change soundpack
  const changeSoundpack = async (soundpackId: string) => {
    settings.value.soundpack = soundpackId
    const success = await saveSettings()

    if (success) {
      // Reload available events for new soundpack
      await loadAvailableEvents()
    }

    return success
  }

  // Toggle audio system enabled state
  const toggleEnabled = async () => {
    settings.value.enabled = !settings.value.enabled
    return await saveSettings()
  }

  // Enable/disable specific events
  const toggleEvent = async (eventId: string, enabled?: boolean) => {
    const shouldEnable = enabled !== undefined ? enabled : !settings.value.enabledEvents.includes(eventId)

    if (shouldEnable) {
      if (!settings.value.enabledEvents.includes(eventId)) {
        settings.value.enabledEvents.push(eventId)
      }
    } else {
      const index = settings.value.enabledEvents.indexOf(eventId)
      if (index > -1) {
        settings.value.enabledEvents.splice(index, 1)
      }
    }

    return await saveSettings()
  }

  // Enable/disable multiple events at once
  const toggleEvents = async (eventIds: string[], enabled?: boolean) => {
    const shouldEnable = enabled !== undefined ? enabled : false

    if (shouldEnable) {
      eventIds.forEach(id => {
        if (!settings.value.enabledEvents.includes(id)) {
          settings.value.enabledEvents.push(id)
        }
      })
    } else {
      const toRemove = new Set(eventIds)
      settings.value.enabledEvents = settings.value.enabledEvents.filter(id => !toRemove.has(id))
    }

    return await saveSettings()
  }

  // Load audio statistics
  const loadStats = async () => {
    if (!wails.isReady.value) return

    try {
      const audioStats = await wails.call('audio.GetStats')
      if (audioStats) {
        stats.value = audioStats
      }
    } catch (err) {
      console.error('Failed to load audio stats:', err)
    }
  }

  // Reset statistics
  const resetStats = async () => {
    if (!wails.isReady.value) return false

    try {
      const success = await wails.call('audio.ResetStats')
      if (success) {
        await loadStats()
      }
      return success
    } catch (err) {
      console.error('Failed to reset audio stats:', err)
      return false
    }
  }

  // Get volume icon based on current volume and mute state
  const getVolumeIcon = (): string => {
    if (settings.value.muted || effectiveVolume.value === 0) {
      return 'volume-off'
    }

    if (effectiveVolume.value < 0.33) {
      return 'volume-low'
    }

    if (effectiveVolume.value < 0.66) {
      return 'volume-medium'
    }

    return 'volume-high'
  }

  // Get human-readable volume percentage
  const getVolumePercentage = (): number => {
    return Math.round(effectiveVolume.value * 100)
  }

  // Setup event listeners
  const setupEventListeners = () => {
    // Listen for playback events
    wails.on('audio.playback.started', (data) => {
      if (data.event) {
        playbackStatus.value.isPlaying = true
        playbackStatus.value.currentEvent = data.event
      }
    })

    wails.on('audio.playback.stopped', () => {
      playbackStatus.value.isPlaying = false
      playbackStatus.value.currentEvent = null
    })

    wails.on('audio.playback.error', (data) => {
      error.value = data.error || 'Playback error occurred'
      playbackStatus.value.isPlaying = false
      playbackStatus.value.currentEvent = null
    })

    // Listen for settings changes
    wails.on('audio.settings.changed', (data) => {
      if (data.settings) {
        settings.value = { ...settings.value, ...data.settings }
      }
    })

    // Listen for UI events that should trigger sounds
    wails.on('ui.event', async (data) => {
      if (data.event && typeof data.event === 'string') {
        await playUIEvent(data.event)
      }
    })
  }

  // Watch for Wails availability
  watch(() => wails.isReady.value, (isReady) => {
    if (isReady) {
      initialize()
      setupEventListeners()
    }
  })

  // Auto-initialize when Wails is ready
  if (wails.isReady.value) {
    initialize()
    setupEventListeners()
  }

  return {
    // State
    settings: readonly(settings),
    playbackStatus: readonly(playbackStatus),
    availableEvents: readonly(availableEvents),
    soundpacks: readonly(soundpacks),
    stats: readonly(stats),
    isLoading: readonly(isLoading),
    error: readonly(error),

    // Computed
    isEnabled,
    effectiveVolume,
    eventsByCategory,
    enabledEvents,
    currentSoundpack,

    // Methods
    initialize,
    loadSettings,
    saveSettings,
    loadAvailableEvents,
    loadSoundpacks,
    loadPlaybackStatus,
    loadStats,
    resetStats,

    // Playback controls
    playEvent,
    playUIEvent,
    stopPlayback,

    // Settings controls
    setVolume,
    toggleMute,
    setMute,
    setGlobalVolume,
    setEffectsVolume,
    setNotificationVolume,
    changeSoundpack,
    toggleEnabled,
    toggleEvent,
    toggleEvents,
    isEventEnabled,

    // Utilities
    getVolumeIcon,
    getVolumePercentage,
  }
}
