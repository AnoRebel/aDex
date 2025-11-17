import { ref, computed, watch } from 'vue'
import { usePreferredReducedMotion, useTimestamp } from '@vueuse/core'
import { useTimeoutFn, useIntervalFn } from '@vueuse/shared'
import { useNotificationStore } from '~/stores/notification'
import { useUIStore } from '~/stores/ui'
import type { TerminalSession } from '~/types/terminal'

export interface BellNotification {
  id: string
  sessionId: string
  timestamp: number
  visual: boolean
  audible: boolean
  message?: string
  title?: string
  source: 'terminal' | 'system' | 'user'
}

export interface BellOptions {
  visual?: boolean
  audible?: boolean
  vibration?: boolean
  duration?: number
  title?: string
  message?: string
  icon?: string
  source?: 'terminal' | 'system' | 'user'
}

export interface BellConfig {
  enabled: boolean
  visual: boolean
  audible: boolean
  vibration: boolean
  volume: number // 0-1
  frequency: number // Hz
  duration: number // ms
  maxNotifications: number
  cooldownPeriod: number // ms
  respectReducedMotion: boolean
  respectFocusState: boolean
  silentWhenActive: boolean
  customSound?: string
}

export function useBellNotifications(session?: Ref<TerminalSession | undefined>) {
  const notificationStore = useNotificationStore()
  const uiStore = useUIStore()

  // VueUse utilities
  const prefersReducedMotion = usePreferredReducedMotion()
  const now = useTimestamp()

  // State
  const isActive = ref(false)
  const notifications = ref<BellNotification[]>([])
  const lastBellTime = ref(0)
  const audioContext = ref<AudioContext | null>(null)
  const oscillator = ref<OscillatorNode | null>(null)
  const gainNode = ref<GainNode | null>(null)

  // Configuration
  const config = ref<BellConfig>({
    enabled: true,
    visual: true,
    audible: true,
    vibration: true,
    volume: 0.3,
    frequency: 800,
    duration: 100,
    maxNotifications: 5,
    cooldownPeriod: 1000,
    respectReducedMotion: true,
    respectFocusState: true,
    silentWhenActive: true,
    customSound: undefined
  })

  // Computed
  const isEnabled = computed(() => {
    if (!config.value.enabled) return false

    // Respect focus state
    if (config.value.respectFocusState && isActive.value && config.value.silentWhenActive) {
      return false
    }

    // Respect reduced motion for visual notifications
    if (config.value.respectReducedMotion && prefersReducedMotion.value && config.value.visual) {
      // Only allow audible notifications if motion is reduced
      return config.value.audible
    }

    return true
  })

  const canPlaySound = computed(() => {
    return config.value.audible && !prefersReducedMotion.value
  })

  const canVibrate = computed(() => {
    return config.value.vibration && 'vibrate' in navigator
  })

  const hasNotifications = computed(() => notifications.value.length > 0)

  const recentNotifications = computed(() =>
    notifications.value.slice(0, config.value.maxNotifications)
  )

  const isCooldownActive = computed(() => {
    return now.value - lastBellTime.value < config.value.cooldownPeriod
  })

  // Methods
  const initializeAudio = (): void => {
    try {
      if (!audioContext.value) {
        audioContext.value = new (window.AudioContext || (window as any).webkitAudioContext)()
      }
    } catch (error) {
      console.warn('Failed to initialize audio context:', error)
    }
  }

  const playAudibleBell = async (options: BellOptions = {}): Promise<void> => {
    if (!canPlaySound.value || !audioContext.value) return

    try {
      // Resume audio context if suspended
      if (audioContext.value.state === 'suspended') {
        await audioContext.value.resume()
      }

      // Create oscillator
      oscillator.value = audioContext.value.createOscillator()
      gainNode.value = audioContext.value.createGain()

      // Configure oscillator
      oscillator.value.type = 'sine'
      oscillator.value.frequency.setValueAtTime(
        options.visual ? config.value.frequency : config.value.frequency * 1.2,
        audioContext.value.currentTime
      )

      // Configure gain (volume)
      gainNode.value.gain.setValueAtTime(0, audioContext.value.currentTime)
      gainNode.value.gain.linearRampToValueAtTime(
        config.value.volume,
        audioContext.value.currentTime + 0.01
      )
      gainNode.value.gain.exponentialRampToValueAtTime(
        0.01,
        audioContext.value.currentTime + (options.duration || config.value.duration) / 1000
      )

      // Connect nodes
      oscillator.value.connect(gainNode.value)
      gainNode.value.connect(audioContext.value.destination)

      // Play sound
      oscillator.value.start(audioContext.value.currentTime)
      oscillator.value.stop(audioContext.value.currentTime + (options.duration || config.value.duration) / 1000)

    } catch (error) {
      console.warn('Failed to play audible bell:', error)
      // Fallback to system beep
      systemBeep()
    }
  }

  const playVibration = (options: BellOptions = {}): void => {
    if (!canVibrate.value) return

    const pattern = options.visual ? [100, 50, 100] : [100]
    navigator.vibrate(pattern)
  }

  const showVisualBell = (options: BellOptions = {}): void => {
    if (!config.value.visual) return

    // Add visual notification to the list
    const notification: BellNotification = {
      id: generateId(),
      sessionId: session?.value?.id || 'unknown',
      timestamp: now.value,
      visual: true,
      audible: options.audible || false,
      message: options.message,
      title: options.title || 'Terminal Bell',
      source: options.source || 'terminal'
    }

    notifications.value.unshift(notification)

    // Limit notifications
    if (notifications.value.length > config.value.maxNotifications) {
      notifications.value = notifications.value.slice(0, config.value.maxNotifications)
    }

    // Show browser notification if permission is granted
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification(notification.title || 'Terminal Activity', {
        body: notification.message || 'Terminal bell rang',
        icon: options.icon || '/terminal-bell-icon.png',
        tag: `terminal-bell-${notification.sessionId}`,
        requireInteraction: false
      })
    }

    // Update terminal UI state
    uiStore.addNotification({
      type: 'info',
      title: notification.title,
      message: notification.message || 'Terminal bell activity',
      persistent: false,
      timestamp: new Date(notification.timestamp).toISOString()
    })
  }

  const systemBeep = (): void => {
    // Create a system beep using Web Audio API as fallback
    try {
      const audio = new Audio()
      audio.volume = 0.1
      // Create a simple beep sound data URI
      audio.src = 'data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBTGH0fPTgjMGHm7A7+OZURE'
      audio.play().catch(() => {
        // Silent fail if audio can't be played
      })
    } catch (error) {
      // Final fallback - use console bell
      console.log('\u0007') // ASCII bell character
    }
  }

  const ringBell = async (options: BellOptions = {}): Promise<void> => {
    if (!isEnabled.value || isCooldownActive.value) return

    const mergedOptions = {
      visual: config.value.visual,
      audible: config.value.audible,
      vibration: config.value.vibration,
      duration: config.value.duration,
      ...options
    }

    lastBellTime.value = now.value

    // Play audible bell
    if (mergedOptions.audible) {
      await playAudibleBell(mergedOptions)
    }

    // Play vibration
    if (mergedOptions.vibration) {
      playVibration(mergedOptions)
    }

    // Show visual notification
    if (mergedOptions.visual) {
      showVisualBell(mergedOptions)
    }

    // Add to history
    addToHistory(mergedOptions)
  }

  const generateId = (): string => {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  const addToHistory = (options: BellOptions): void => {
    // Store in localStorage for analytics
    try {
      const history = JSON.parse(localStorage.getItem('terminal-bell-history') || '[]')
      history.push({
        timestamp: now.value,
        sessionId: session?.value?.id,
        options,
        isActive: isActive.value
      })

      // Keep only last 100 entries
      if (history.length > 100) {
        history.splice(0, history.length - 100)
      }

      localStorage.setItem('terminal-bell-history', JSON.stringify(history))
    } catch (error) {
      // Silent fail
    }
  }

  const clearNotifications = (): void => {
    notifications.value = []
  }

  const removeNotification = (id: string): void => {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }

  const markAsRead = (id: string): void => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification) {
      // Mark as read (could add read state if needed)
      removeNotification(id)
    }
  }

  const requestNotificationPermission = async (): Promise<boolean> => {
    if (!('Notification' in window)) {
      return false
    }

    if (Notification.permission === 'granted') {
      return true
    }

    if (Notification.permission !== 'denied') {
      const permission = await Notification.requestPermission()
      return permission === 'granted'
    }

    return false
  }

  const updateConfig = (newConfig: Partial<BellConfig>): void => {
    config.value = { ...config.value, ...newConfig }

    // Save to localStorage
    try {
      localStorage.setItem('terminal-bell-config', JSON.stringify(config.value))
    } catch (error) {
      console.warn('Failed to save bell config:', error)
    }
  }

  const loadConfig = (): void => {
    try {
      const saved = localStorage.getItem('terminal-bell-config')
      if (saved) {
        config.value = { ...config.value, ...JSON.parse(saved) }
      }
    } catch (error) {
      console.warn('Failed to load bell config:', error)
    }
  }

  const getStats = () => {
    try {
      const history = JSON.parse(localStorage.getItem('terminal-bell-history') || '[]')
      const last24h = history.filter((entry: any) =>
        now.value - entry.timestamp < 24 * 60 * 60 * 1000
      )

      return {
        total: history.length,
        last24h: last24h.length,
        lastWeek: history.filter((entry: any) =>
          now.value - entry.timestamp < 7 * 24 * 60 * 60 * 1000
        ).length,
        averagePerDay: last24h.length / 1 // Could calculate more sophisticated average
      }
    } catch (error) {
      return { total: 0, last24h: 0, lastWeek: 0, averagePerDay: 0 }
    }
  }

  const setActive = (active: boolean): void => {
    isActive.value = active
  }

  // Lifecycle
  loadConfig()

  // Initialize audio on user interaction
  const handleUserInteraction = () => {
    initializeAudio()
    document.removeEventListener('click', handleUserInteraction)
    document.removeEventListener('keydown', handleUserInteraction)
  }

  document.addEventListener('click', handleUserInteraction)
  document.addEventListener('keydown', handleUserInteraction)

  // Cleanup
  const cleanup = () => {
    if (oscillator.value) {
      try {
        oscillator.value.stop()
        oscillator.value.disconnect()
      } catch (error) {
        // Ignore cleanup errors
      }
    }

    if (gainNode.value) {
      gainNode.value.disconnect()
    }

    if (audioContext.value && audioContext.value.state !== 'closed') {
      audioContext.value.close()
    }

    document.removeEventListener('click', handleUserInteraction)
    document.removeEventListener('keydown', handleUserInteraction)
  }

  // Auto-cleanup on unmount
  onUnmounted(() => {
    cleanup()
  })

  return {
    // State
    notifications,
    config,
    isActive,

    // Computed
    isEnabled,
    canPlaySound,
    canVibrate,
    hasNotifications,
    recentNotifications,
    isCooldownActive,

    // Methods
    ringBell,
    clearNotifications,
    removeNotification,
    markAsRead,
    requestNotificationPermission,
    updateConfig,
    getStats,
    setActive,
    cleanup
  }
}

export default useBellNotifications