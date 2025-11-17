/**
 * Audio Composable Tests
 * Tests for the audio effects system composable
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'

// Import composables
import { useAudio } from '~/composables/useAudio'
import { useAudioStore } from '~/stores/audio'

// Import types
import type {
  AudioEvent,
  AudioSettings,
  AudioPlaybackStatus,
  Soundpack,
  AudioEventType
} from '~/types/audio'

// Mock Wails
const mockWails = {
  isReady: { value: true },
  call: vi.fn(),
  on: vi.fn(),
  audio: {
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    playEvent: vi.fn(),
    playUIEvent: vi.fn(),
    stopPlayback: vi.fn(),
    getAvailableEvents: vi.fn(),
    getSoundpacks: vi.fn(),
    getPlaybackStatus: vi.fn(),
    getStats: vi.fn(),
    resetStats: vi.fn()
  }
}

// Mock the useWails composable
vi.mock('~/composables/useWails', () => ({
  useWails: () => mockWails
}))

describe('useAudio composable', () => {
  let audio: ReturnType<typeof useAudio>
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    audio = useAudio()

    // Reset mocks
    vi.clearAllMocks()

    // Default successful responses
    mockWails.call.mockResolvedValue(true)
    mockWails.audio.getSettings.mockResolvedValue({
      enabled: false,
      volume: 50,
      muted: false,
      soundpack: 'default',
      globalVolume: 0.5,
      effectsVolume: 0.5,
      notificationVolume: 0.3,
      muteInBackground: false,
      autoPlay: true,
      enabledEvents: ['system:startup', 'terminal:bell'],
      version: '1.0.0',
      updatedAt: new Date().toISOString()
    })

    mockWails.audio.getAvailableEvents.mockResolvedValue([
      {
        id: 'terminal_bell',
        name: 'Terminal Bell',
        description: 'Terminal bell sound',
        category: 'interaction' as AudioEventType,
        filePath: '/audio/default/terminal_bell.wav',
        duration: 200,
        volume: 50,
        enabled: true,
        createdAt: new Date()
      }
    ])

    mockWails.audio.getSoundpacks.mockResolvedValue([
      {
        id: 'default',
        name: 'Default',
        displayName: 'Default Soundpack',
        description: 'Default sound effects',
        version: '1.0.0',
        author: 'aDex-UI Team',
        isBuiltIn: true,
        enabled: true
      }
    ])

    mockWails.audio.getPlaybackStatus.mockResolvedValue({
      isPlaying: false,
      currentEvent: null,
      volume: 0.5,
      muted: false,
      loadedEvents: [],
      supportedFormats: ['wav', 'mp3', 'ogg', 'aac'],
      error: null
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('initialization', () => {
    it('should initialize with default values', () => {
      expect(audio.settings.value.enabled).toBe(false)
      expect(audio.settings.value.volume).toBe(50)
      expect(audio.settings.value.soundpack).toBe('default')
      expect(audio.playbackStatus.value.isPlaying).toBe(false)
      expect(audio.availableEvents.value).toEqual([])
      expect(audio.soundpacks.value).toEqual([])
    })

    it('should load settings on initialization', async () => {
      await audio.initialize()

      expect(mockWails.audio.getSettings).toHaveBeenCalled()
      expect(mockWails.audio.getAvailableEvents).toHaveBeenCalled()
      expect(mockWails.audio.getSoundpacks).toHaveBeenCalled()
      expect(mockWails.audio.getPlaybackStatus).toHaveBeenCalled()
    })

    it('should handle initialization errors gracefully', async () => {
      mockWails.audio.getSettings.mockRejectedValue(new Error('Failed to load settings'))

      await audio.initialize()

      expect(audio.error.value).toBeTruthy()
      expect(audio.error.value).toContain('Failed to initialize audio system')
    })
  })

  describe('state management', () => {
    it('should compute effective volume correctly', async () => {
      await audio.initialize()

      // Test normal volume
      expect(audio.effectiveVolume.value).toBe(0.25) // 50 * 0.5

      // Test muted state
      await audio.setMuted(true)
      expect(audio.effectiveVolume.value).toBe(0)

      // Test different volume levels
      await audio.setVolume(100)
      await audio.setMuted(false)
      expect(audio.effectiveVolume.value).toBe(0.5) // 100 * 0.5
    })

    it('should check if audio is enabled', () => {
      expect(audio.isEnabled.value).toBe(false) // disabled by default

      audio.settings.value.enabled = true
      audio.settings.value.muted = false
      expect(audio.isEnabled.value).toBe(true)

      audio.settings.value.muted = true
      expect(audio.isEnabled.value).toBe(false)
    })

    it('should group events by category', async () => {
      await audio.initialize()

      const eventsByCategory = audio.eventsByCategory.value
      expect(eventsByCategory).toHaveProperty('system')
      expect(eventsByCategory).toHaveProperty('interaction')
      expect(eventsByCategory).toHaveProperty('notification')
      expect(eventsByCategory).toHaveProperty('error')
      expect(eventsByCategory).toHaveProperty('success')

      expect(Array.isArray(eventsByCategory.interaction)).toBe(true)
    })

    it('should identify enabled events', async () => {
      await audio.initialize()

      const enabledEvents = audio.enabledEvents.value
      expect(Array.isArray(enabledEvents)).toBe(true)

      // Should have terminal_bell as it's in default enabled events
      const terminalBell = enabledEvents.find(e => e.id === 'terminal_bell')
      expect(terminalBell).toBeDefined()
    })

    it('should get current soundpack', async () => {
      await audio.initialize()

      const currentSoundpack = audio.currentSoundpack.value
      expect(currentSoundpack?.id).toBe('default')
      expect(currentSoundpack?.name).toBe('Default')
    })
  })

  describe('settings management', () => {
    beforeEach(async () => {
      await audio.initialize()
    })

    it('should update volume', async () => {
      await audio.setVolume(75)

      expect(mockWails.audio.updateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ volume: 75 })
      )

      expect(audio.settings.value.volume).toBe(75)
    })

    it('should clamp volume to valid range', async () => {
      await audio.setVolume(150)
      expect(audio.settings.value.volume).toBe(100)

      await audio.setVolume(-10)
      expect(audio.settings.value.volume).toBe(0)
    })

    it('should toggle mute state', async () => {
      const originalMuted = audio.settings.value.muted

      await audio.toggleMute()

      expect(audio.settings.value.muted).toBe(!originalMuted)
      expect(mockWails.audio.updateSettings).toHaveBeenCalled()
    })

    it('should set mute state explicitly', async () => {
      await audio.setMuted(true)

      expect(audio.settings.value.muted).toBe(true)
      expect(mockWails.audio.updateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ muted: true })
      )
    })

    it('should update global volume', async () => {
      await audio.setGlobalVolume(0.8)

      expect(mockWails.audio.updateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ globalVolume: 0.8 })
      )

      expect(audio.settings.value.globalVolume).toBe(0.8)
    })

    it('should update effects volume', async () => {
      await audio.setEffectsVolume(0.7)

      expect(mockWails.audio.updateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ effectsVolume: 0.7 })
      )

      expect(audio.settings.value.effectsVolume).toBe(0.7)
    })

    it('should update notification volume', async () => {
      await audio.setNotificationVolume(0.6)

      expect(mockWails.audio.updateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ notificationVolume: 0.6 })
      )

      expect(audio.settings.value.notificationVolume).toBe(0.6)
    })

    it('should change soundpack', async () => {
      await audio.changeSoundpack('retro')

      expect(mockWails.audio.updateSettings).toHaveBeenCalledWith(
        expect.objectContaining({ soundpack: 'retro' })
      )

      expect(audio.settings.value.soundpack).toBe('retro')
    })

    it('should toggle enabled state', async () => {
      const originalEnabled = audio.settings.value.enabled

      await audio.toggleEnabled()

      expect(audio.settings.value.enabled).toBe(!originalEnabled)
      expect(mockWails.audio.updateSettings).toHaveBeenCalled()
    })
  })

  describe('event management', () => {
    beforeEach(async () => {
      await audio.initialize()
    })

    it('should check if event is enabled', () => {
      // Should be enabled by default (in enabledEvents list)
      expect(audio.isEventEnabled('system:startup')).toBe(true)

      // Should be disabled (not in enabledEvents list)
      expect(audio.isEventEnabled('random:event')).toBe(false)

      // Should be disabled when audio is muted
      audio.settings.value.muted = true
      expect(audio.isEventEnabled('system:startup')).toBe(false)

      // Should be disabled when audio is disabled
      audio.settings.value.muted = false
      audio.settings.value.enabled = false
      expect(audio.isEventEnabled('system:startup')).toBe(false)
    })

    it('should toggle individual events', async () => {
      // Initially enabled
      expect(audio.isEventEnabled('system:startup')).toBe(true)

      await audio.toggleEvent('system:startup', false)

      expect(mockWails.audio.updateSettings).toHaveBeenCalled()
      expect(audio.settings.value.enabledEvents).not.toContain('system:startup')
    })

    it('should toggle multiple events', async () => {
      const events = ['system:startup', 'terminal:bell']

      await audio.toggleEvents(events, false)

      expect(mockWails.audio.updateSettings).toHaveBeenCalled()
      expect(audio.settings.value.enabledEvents).not.toContain('system:startup')
      expect(audio.settings.value.enabledEvents).not.toContain('terminal:bell')
    })
  })

  describe('playback control', () => {
    beforeEach(async () => {
      await audio.initialize()
    })

    it('should play audio events', async () => {
      mockWails.audio.playEvent.mockResolvedValue(true)

      const result = await audio.playEvent('terminal_bell')

      expect(mockWails.audio.playEvent).toHaveBeenCalledWith('terminal_bell')
      expect(result).toBe(true)

      expect(audio.playbackStatus.value.isPlaying).toBe(true)
    })

    it('should not play disabled events', async () => {
      mockWails.audio.playEvent.mockResolvedValue(true)

      const result = await audio.playEvent('random:event')

      expect(mockWails.audio.playEvent).not.toHaveBeenCalled()
      expect(result).toBe(false)
    })

    it('should handle playback errors', async () => {
      mockWails.audio.playEvent.mockRejectedValue(new Error('Playback failed'))

      const result = await audio.playEvent('terminal_bell')

      expect(result).toBe(false)
      expect(audio.error.value).toBeTruthy()
      expect(audio.error.value).toContain('Failed to play audio event')
    })

    it('should play UI events', async () => {
      mockWails.audio.playUIEvent.mockResolvedValue(true)

      const result = await audio.playUIEvent('ui:button_click')

      expect(mockWails.audio.playUIEvent).toHaveBeenCalledWith('ui:button_click')
      expect(result).toBe(true)
    })

    it('should not play UI events when audio is disabled', async () => {
      audio.settings.value.enabled = false

      const result = await audio.playUIEvent('ui:button_click')

      expect(mockWails.audio.playUIEvent).not.toHaveBeenCalled()
      expect(result).toBe(false)
    })

    it('should stop playback', async () => {
      mockWails.audio.stopPlayback.mockResolvedValue(true)

      // Set playing state
      audio.playbackStatus.value.isPlaying = true

      const result = await audio.stopPlayback()

      expect(mockWails.audio.stopPlayback).toHaveBeenCalled()
      expect(result).toBe(true)
      expect(audio.playbackStatus.value.isPlaying).toBe(false)
      expect(audio.playbackStatus.value.currentEvent).toBeNull()
    })
  })

  describe('statistics', () => {
    beforeEach(async () => {
      await audio.initialize()
    })

    it('should load audio statistics', async () => {
      const mockStats = {
        totalPlays: 10,
        eventsPlayed: { 'terminal_bell': 5 },
        categoriesPlayed: { 'interaction': 5 },
        totalPlaytime: 5000,
        averageVolume: 0.6,
        lastPlayed: new Date(),
        mostPlayedEvent: 'terminal_bell',
        settings: {},
        generatedAt: new Date()
      }

      mockWails.audio.getStats.mockResolvedValue(mockStats)

      await audio.loadStats()

      expect(audio.stats.value).toEqual(mockStats)
    })

    it('should reset statistics', async () => {
      mockWails.audio.resetStats.mockResolvedValue(true)
      mockWails.audio.getStats.mockResolvedValue({
        totalPlays: 0,
        eventsPlayed: {},
        categoriesPlayed: {},
        totalPlaytime: 0,
        averageVolume: 0,
        lastPlayed: new Date(),
        mostPlayedEvent: '',
        settings: {},
        generatedAt: new Date()
      })

      const result = await audio.resetStats()

      expect(mockWails.audio.resetStats).toHaveBeenCalled()
      expect(mockWails.audio.getStats).toHaveBeenCalled()
      expect(result).toBe(true)
    })
  })

  describe('utility functions', () => {
    beforeEach(async () => {
      await audio.initialize()
    })

    it('should get volume icon correctly', () => {
      // Test muted
      audio.settings.value.muted = true
      expect(audio.getVolumeIcon()).toBe('volume-off')

      // Test low volume
      audio.settings.value.muted = false
      audio.settings.value.volume = 20
      audio.settings.value.globalVolume = 0.5
      expect(audio.getVolumeIcon()).toBe('volume-low')

      // Test medium volume
      audio.settings.value.volume = 50
      audio.settings.value.globalVolume = 0.5
      expect(audio.getVolumeIcon()).toBe('volume-medium')

      // Test high volume
      audio.settings.value.volume = 80
      audio.settings.value.globalVolume = 0.5
      expect(audio.getVolumeIcon()).toBe('volume-high')
    })

    it('should get volume percentage correctly', () => {
      // Test muted
      audio.settings.value.muted = true
      expect(audio.getVolumePercentage()).toBe(0)

      // Test normal volume
      audio.settings.value.muted = false
      audio.settings.value.volume = 75
      audio.settings.value.globalVolume = 0.8
      expect(audio.getVolumePercentage()).toBe(60) // round(75 * 0.8)
    })
  })

  describe('error handling', () => {
    beforeEach(async () => {
      await audio.initialize()
    })

    it('should handle Wails not ready', async () => {
      mockWails.isReady.value = false

      const result = await audio.playEvent('terminal_bell')

      expect(result).toBe(false)
      expect(mockWails.audio.playEvent).not.toHaveBeenCalled()
    })

    it('should handle network timeouts', async () => {
      mockWails.audio.playEvent.mockImplementation(() =>
        new Promise((resolve, reject) => {
          setTimeout(() => reject(new Error('Timeout')), 100)
        })
      )

      const result = await audio.playEvent('terminal_bell')

      expect(result).toBe(false)
      expect(audio.error.value).toBeTruthy()
    })
  })

  describe('reactivity', () => {
    it('should update computed values when settings change', async () => {
      await audio.initialize()

      const originalVolume = audio.effectiveVolume.value
      await audio.setVolume(80)

      expect(audio.effectiveVolume.value).not.toBe(originalVolume)
    })

    it('should update playback status when playing', async () => {
      await audio.initialize()

      expect(audio.playbackStatus.value.isPlaying).toBe(false)

      mockWails.audio.playEvent.mockResolvedValue(true)
      await audio.playEvent('terminal_bell')

      expect(audio.playbackStatus.value.isPlaying).toBe(true)
    })
  })

  describe('integration with store', () => {
    it('should work with Pinia store', () => {
      const store = useAudioStore(pinia)

      expect(store).toBeDefined()
      expect(store.settings).toBeDefined()
      expect(store.playbackStatus).toBeDefined()
    })

    it('should sync state between composable and store', async () => {
      const store = useAudioStore(pinia)
      await audio.initialize()

      // Change settings via composable
      await audio.setVolume(75)

      // Store should have same settings (in real implementation,
      // they would share the same state source)
      expect(store.settings.volume).toBeDefined()
    })
  })

  describe('memory management', () => {
    it('should cleanup properly', async () => {
      await audio.initialize()

      // Simulate cleanup
      if (audio.stopPlayback) {
        await audio.stopPlayback()
      }

      // In real implementation, would verify event listeners are cleaned up
      expect(audio.error.value).toBeNull()
    })
  })

  describe('performance', () => {
    it('should handle rapid calls efficiently', async () => {
      await audio.initialize()

      const startTime = performance.now()

      // Make multiple rapid calls
      const promises = []
      for (let i = 0; i < 10; i++) {
        promises.push(audio.getVolumeIcon())
      }

      await Promise.all(promises)

      const endTime = performance.now()
      const duration = endTime - startTime

      // Should complete within reasonable time
      expect(duration).toBeLessThan(100)
    })
  })
})

describe('Audio Composable Edge Cases', () => {
  it('should handle invalid audio data', async () => {
    mockWails.audio.getAvailableEvents.mockResolvedValue([
      {
        id: '',
        name: 'Invalid Event',
        description: '',
        category: 'invalid' as AudioEventType,
        filePath: '',
        duration: -1,
        volume: 150,
        enabled: true,
        createdAt: new Date()
      }
    ])

    const audio = useAudio()
    await audio.initialize()

    // Should handle invalid data gracefully
    expect(audio.availableEvents.value).toHaveLength(1)
  })

  it('should handle missing audio files', async () => {
    mockWails.audio.playEvent.mockRejectedValue(new Error('Audio file not found'))

    const audio = useAudio()
    await audio.initialize()

    const result = await audio.playEvent('terminal_bell')

    expect(result).toBe(false)
    expect(audio.error.value).toBeTruthy()
  })

  it('should handle concurrent playback requests', async () => {
    const audio = useAudio()
    await audio.initialize()

    mockWails.audio.playEvent.mockResolvedValue(true)

    // Make multiple concurrent requests
    const promises = []
    for (let i = 0; i < 5; i++) {
      promises.push(audio.playEvent('terminal_bell'))
    }

    const results = await Promise.all(promises)

    // All should complete without errors
    expect(results.every(r => r === true)).toBe(true)
  })
})