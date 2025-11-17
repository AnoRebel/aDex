import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

// Keyboard layout types
export interface KeyData {
  key: string
  shift?: string
  width?: number
  type?: 'standard' | 'function' | 'special' | 'modifier' | 'space' | 'number' | 'operator'
  side?: 'left' | 'right'
}

export interface KeyboardLayout {
  name: string
  displayName: string
  type: 'standard' | 'alternative' | 'special'
  rows: KeyData[][]
}

export interface KeyboardSettings {
  defaultLayout: string
  keyRepeatDelay: number
  keyRepeatInterval: number
  autoShow: boolean
  hapticFeedback: boolean
  soundFeedback: boolean
  opacity: number
  theme: 'dark' | 'light' | 'cyberpunk'
}

export const useKeyboard = () => {
  // State
  const currentLayout = ref<string>('qwerty')
  const layouts = ref<Record<string, KeyboardLayout>>({})
  const settings = ref<KeyboardSettings>({
    defaultLayout: 'qwerty',
    keyRepeatDelay: 500,
    keyRepeatInterval: 50,
    autoShow: false,
    hapticFeedback: false,
    soundFeedback: false,
    opacity: 0.9,
    theme: 'dark'
  })

  const isShiftActive = ref(false)
  const isCapsLockActive = ref(false)
  const isCtrlActive = ref(false)
  const isAltActive = ref(false)
  const isVisible = ref(false)

  // Load keyboard layouts and settings
  const loadKeyboardData = async () => {
    try {
      const response = await fetch('/app/assets/data/keyboard-layouts.json')
      if (!response.ok) {
        throw new Error('Failed to load keyboard layouts')
      }

      const data = await response.json()
      layouts.value = data.layouts
      settings.value = { ...settings.value, ...data.settings }

      // Load saved settings from localStorage
      const savedSettings = localStorage.getItem('adex-keyboard-settings')
      if (savedSettings) {
        const parsed = JSON.parse(savedSettings)
        settings.value = { ...settings.value, ...parsed }
        currentLayout.value = parsed.currentLayout || settings.value.defaultLayout
      }
    } catch (error) {
      console.error('Failed to load keyboard data:', error)
      // Fallback to basic QWERTY
      loadFallbackLayout()
    }
  }

  const loadFallbackLayout = () => {
    layouts.value = {
      qwerty: {
        name: 'qwerty',
        displayName: 'QWERTY (US)',
        type: 'standard',
        rows: [
          // Minimal fallback layout
          [
            { key: 'Escape', width: 1, type: 'function' }
          ],
          [
            { key: '1', shift: '!', width: 1 },
            { key: '2', shift: '@', width: 1 },
            { key: '3', shift: '#', width: 1 },
            { key: 'Backspace', width: 2, type: 'special' }
          ]
        ]
      }
    }
  }

  // Computed properties
  const currentLayoutData = computed(() => layouts.value[currentLayout.value])

  const availableLayouts = computed(() => {
    return Object.values(layouts.value).map(layout => ({
      name: layout.name,
      displayName: layout.displayName,
      type: layout.type
    }))
  })

  // Key repeat functionality
  const repeatTimer = ref<NodeJS.Timeout | null>(null)
  const isRepeating = ref(false)

  const startKeyRepeat = (callback: () => void) => {
    if (isRepeating.value) return

    isRepeating.value = true

    // Initial delay
    repeatTimer.value = setTimeout(() => {
      callback()

      // Start interval
      repeatTimer.value = setInterval(() => {
        callback()
      }, settings.value.keyRepeatInterval)
    }, settings.value.keyRepeatDelay)
  }

  const stopKeyRepeat = () => {
    isRepeating.value = false
    if (repeatTimer.value) {
      clearTimeout(repeatTimer.value)
      clearInterval(repeatTimer.value)
      repeatTimer.value = null
    }
  }

  // Modifier state management
  const handleModifierPress = (key: string) => {
    switch (key.toLowerCase()) {
      case 'shift':
        isShiftActive.value = !isShiftActive.value
        break
      case 'capslock':
        isCapsLockActive.value = !isCapsLockActive.value
        break
      case 'ctrl':
        isCtrlActive.value = !isCtrlActive.value
        break
      case 'alt':
        isAltActive.value = !isAltActive.value
        break
    }
  }

  const handleModifierRelease = (key: string) => {
    // Only toggle Shift back on release if not CapsLock
    if (key.toLowerCase() === 'shift' && !isCapsLockActive.value) {
      isShiftActive.value = false
    }
  }

  // Keyboard actions
  const setLayout = (layoutName: string) => {
    if (layouts.value[layoutName]) {
      currentLayout.value = layoutName
      saveSettings()
    }
  }

  const showKeyboard = () => {
    isVisible.value = true
    document.body.style.overflow = 'hidden' // Prevent scrolling
  }

  const hideKeyboard = () => {
    isVisible.value = false
    document.body.style.overflow = ''
    resetModifiers()
  }

  const toggleKeyboard = () => {
    if (isVisible.value) {
      hideKeyboard()
    } else {
      showKeyboard()
    }
  }

  const resetModifiers = () => {
    isShiftActive.value = false
    isCapsLockActive.value = false
    isCtrlActive.value = false
    isAltActive.value = false
  }

  const updateSettings = (newSettings: Partial<KeyboardSettings>) => {
    settings.value = { ...settings.value, ...newSettings }
    saveSettings()
  }

  const saveSettings = () => {
    const settingsToSave = {
      ...settings.value,
      currentLayout: currentLayout.value
    }
    localStorage.setItem('adex-keyboard-settings', JSON.stringify(settingsToSave))
  }

  // Keyboard event handling for hardware keyboard integration
  const handleHardwareKeyDown = (event: KeyboardEvent) => {
    // Update modifier states
    if (event.shiftKey && !isShiftActive.value) {
      isShiftActive.value = true
    }
    if (event.ctrlKey && !isCtrlActive.value) {
      isCtrlActive.value = true
    }
    if (event.altKey && !isAltActive.value) {
      isAltActive.value = true
    }

    // Handle CapsLock toggle
    if (event.key === 'CapsLock') {
      isCapsLockActive.value = !isCapsLockActive.value
    }
  }

  const handleHardwareKeyUp = (event: KeyboardEvent) => {
    // Update modifier states on release
    if (!event.shiftKey && isShiftActive.value && !isCapsLockActive.value) {
      isShiftActive.value = false
    }
    if (!event.ctrlKey && isCtrlActive.value) {
      isCtrlActive.value = false
    }
    if (!event.altKey && isAltActive.value) {
      isAltActive.value = false
    }
  }

  // Accessibility helpers
  const getAriaLabel = (keyData: KeyData) => {
    if (keyData.type === 'space') return 'Space bar'
    if (keyData.type === 'modifier') {
      return `${keyData.key} modifier key (${keyData.side || ''})`
    }
    if (keyData.shift) {
      return `${keyData.key} with shift ${keyData.shift}`
    }
    return keyData.key
  }

  const getKeyName = (keyData: KeyData, shiftActive: boolean, capsLockActive: boolean) => {
    if (keyData.type === 'space') return ' '
    if (keyData.type === 'modifier') return keyData.key

    // Handle shift and caps lock logic
    const useShift = shiftActive !== capsLockActive
    if (useShift && keyData.shift) {
      return keyData.shift
    }

    return keyData.key
  }

  // Touch feedback
  const triggerHapticFeedback = (pattern?: number | number[]) => {
    if (!settings.value.hapticFeedback || !('vibrate' in navigator)) {
      return
    }

    try {
      if (pattern) {
        navigator.vibrate(pattern)
      } else {
        navigator.vibrate(10) // Default short vibration
      }
    } catch (error) {
      // Ignore vibration errors
    }
  }

  // Positioning helpers
  const calculateKeyboardPosition = (targetElement?: HTMLElement) => {
    if (!targetElement || typeof window === 'undefined') {
      return { bottom: '0', left: '0', right: '0' }
    }

    const rect = targetElement.getBoundingClientRect()
    const viewportHeight = window.innerHeight

    // Position keyboard below the target element
    return {
      bottom: `${viewportHeight - rect.bottom}px`,
      left: '0',
      right: '0'
    }
  }

  // Event listeners
  const addKeyboardListeners = () => {
    if (typeof window === 'undefined') return

    document.addEventListener('keydown', handleHardwareKeyDown)
    document.addEventListener('keyup', handleHardwareKeyUp)

    // Handle window resize
    const handleResize = () => {
      if (window.innerWidth < 768) {
        // Mobile: adjust keyboard size
        updateSettings({ theme: 'dark' }) // Force dark theme on mobile
      }
    }

    window.addEventListener('resize', handleResize)
  }

  const removeKeyboardListeners = () => {
    if (typeof window === 'undefined') return

    document.removeEventListener('keydown', handleHardwareKeyDown)
    document.removeEventListener('keyup', handleHardwareKeyUp)
  }

  // Lifecycle
  onMounted(() => {
    loadKeyboardData()
    addKeyboardListeners()
  })

  onUnmounted(() => {
    stopKeyRepeat()
    removeKeyboardListeners()
  })

  return {
    // State
    currentLayout,
    currentLayoutData,
    availableLayouts,
    settings,
    isVisible,
    isShiftActive,
    isCapsLockActive,
    isCtrlActive,
    isAltActive,
    isRepeating,

    // Actions
    setLayout,
    showKeyboard,
    hideKeyboard,
    toggleKeyboard,
    resetModifiers,
    updateSettings,
    saveSettings,

    // Key repeat
    startKeyRepeat,
    stopKeyRepeat,

    // Modifier handling
    handleModifierPress,
    handleModifierRelease,

    // Utilities
    getAriaLabel,
    getKeyName,
    triggerHapticFeedback,
    calculateKeyboardPosition
  }
}

// Export types for external use
export type { KeyData, KeyboardLayout, KeyboardSettings }