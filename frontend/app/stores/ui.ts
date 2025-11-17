import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWails, useEvents } from '../composables/useWails'

// UI State interfaces
export interface Theme {
  id: string
  name: string
  isDark: boolean
  colors: Record<string, string>
  author: string
  version: string
}

export interface WindowState {
  id: string
  title: string
  width: number
  height: number
  x: number
  y: number
  maximized: boolean
  minimized: boolean
  focused: boolean
}

export interface Notification {
  id: string
  type: 'info' | 'warning' | 'error' | 'success'
  title: string
  message: string
  timestamp: Date
  duration?: number
  persistent: boolean
  actions?: NotificationAction[]
}

export interface NotificationAction {
  label: string
  action: () => void
  primary?: boolean
}

export interface ModalState {
  id: string
  type: 'alert' | 'confirm' | 'prompt' | 'custom'
  title: string
  message: string
  visible: boolean
  closable: boolean
  data?: any
}

export interface MenuItem {
  id: string
  label: string
  icon?: string
  shortcut?: string
  disabled?: boolean
  separator?: boolean
  submenu?: MenuItem[]
  action?: () => void
}

export interface ContextMenuState {
  visible: boolean
  x: number
  y: number
  items: MenuItem[]
  target?: any
}

export interface PanelState {
  id: string
  type: 'terminal' | 'monitor' | 'filesystem' | 'network' | 'settings'
  visible: boolean
  active: boolean
  position: { x: number; y: number }
  size: { width: number; height: number }
  zIndex: number
  data?: any
}

// App State interface
export interface AppState {
  initialized: boolean
  ready: boolean
  online: boolean
  version: string
  buildDate: string
  platform: string
  arch: string
}

/**
 * UI Store - manages global UI state
 */
export const useUIStore = defineStore('ui', () => {
  const wails = useWails()
  const events = useEvents()

  // App state
  const app = ref<AppState>({
    initialized: false,
    ready: false,
    online: navigator.onLine,
    version: '1.0.0',
    buildDate: '',
    platform: '',
    arch: ''
  })

  // Theme state
  const currentTheme = ref<Theme | null>(null)
  const availableThemes = ref<Theme[]>([])
  const customCSSVariables = ref<Record<string, string>>({})

  // Window state
  const windowState = ref<WindowState>({
    id: 'main',
    title: 'aDex-UI',
    width: 1200,
    height: 800,
    x: 0,
    y: 0,
    maximized: false,
    minimized: false,
    focused: true
  })

  // Notifications
  const notifications = ref<Notification[]>([])
  const notificationSound = ref(true)

  // Modals
  const modals = ref<Map<string, ModalState>>(new Map())

  // Context menu
  const contextMenu = ref<ContextMenuState>({
    visible: false,
    x: 0,
    y: 0,
    items: []
  })

  // Panels
  const panels = ref<Map<string, PanelState>>(new Map())
  const activePanel = ref<string | null>(null)

  // Loading states
  const loading = ref<Record<string, boolean>>({})
  const globalLoading = ref(false)

  // Error state
  const globalError = ref<string | null>(null)

  // Computed properties
  const isAppReady = computed(() => app.value.initialized && app.value.ready && wails.isReady.value)
  const isDarkTheme = computed(() => currentTheme.value?.isDark ?? true)
  const hasNotifications = computed(() => notifications.value.length > 0)
  const visiblePanels = computed(() => Array.from(panels.value.values()).filter(panel => panel.visible))
  const loadingCount = computed(() => Object.values(loading.value).filter(Boolean).length)

  // Actions
  const initializeApp = async () => {
    try {
      globalLoading.value = true
      globalError.value = null

      // Initialize Wails
      await wails.initialize()

      // Update app state
      app.value.initialized = true
      app.value.ready = true

      // Load UI settings
      await loadUISettings()

      // Setup event listeners
      setupEventListeners()

      // Apply initial theme
      if (currentTheme.value) {
        applyTheme(currentTheme.value)
      }

    } catch (error) {
      globalError.value = error instanceof Error ? error.message : 'Failed to initialize app'
      console.error('App initialization failed:', error)
    } finally {
      globalLoading.value = false
    }
  }

  const loadUISettings = async () => {
    try {
      // Load theme settings
      // This would load from the configuration service
      currentTheme.value = {
        id: 'default-dark',
        name: 'Default Dark',
        isDark: true,
        colors: {
          primary: '#3b82f6',
          secondary: '#1f2937',
          accent: '#10b981',
          background: '#111827',
          surface: '#1f2937',
          text: '#f3f4f6'
        },
        author: 'HackEAC',
        version: '1.0.0'
      }

      // Load panel configurations
      // This would load from the configuration service
      setupDefaultPanels()
    } catch (error) {
      console.error('Failed to load UI settings:', error)
    }
  }

  const setupDefaultPanels = () => {
    // Setup default panel layout
    panels.value.set('terminal', {
      id: 'terminal',
      type: 'terminal',
      visible: true,
      active: true,
      position: { x: 0, y: 0 },
      size: { width: 600, height: 400 },
      zIndex: 1,
      data: { sessionCount: 1 }
    })

    panels.value.set('monitor', {
      id: 'monitor',
      type: 'monitor',
      visible: true,
      active: false,
      position: { x: 620, y: 0 },
      size: { width: 580, height: 400 },
      zIndex: 2
    })

    activePanel.value = 'terminal'
  }

  const setupEventListeners = () => {
    // Listen to online/offline events
    window.addEventListener('online', () => {
      app.value.online = true
    })

    window.addEventListener('offline', () => {
      app.value.online = false
    })

    // Listen to Wails events
    events.subscribe('app.started', (event) => {
      app.value.ready = true
    })

    events.subscribe('app.shutdown', (event) => {
      cleanup()
    })

    events.subscribe('theme.changed', (event) => {
      const themeData = event.Data
      if (themeData) {
        setTheme(themeData)
      }
    })

    events.subscribe('system.alert', (event) => {
      const alertData = event.Data
      if (alertData) {
        addNotification({
          type: 'warning',
          title: 'System Alert',
          message: alertData.message || 'System alert occurred',
          persistent: false
        })
      }
    })
  }

  const setTheme = (theme: Theme) => {
    currentTheme.value = theme
    applyTheme(theme)
    saveThemePreference(theme.id)
  }

  const applyTheme = (theme: Theme) => {
    const root = document.documentElement

    // Apply CSS custom properties
    Object.entries(theme.colors).forEach(([key, value]) => {
      root.style.setProperty(`--color-${key}`, value)
    })

    // Apply dark/light class
    root.classList.toggle('dark', theme.isDark)
    root.classList.toggle('light', !theme.isDark)

    // Update custom CSS variables
    customCSSVariables.value = { ...theme.colors }
  }

  const saveThemePreference = async (themeId: string) => {
    try {
      // Save to configuration service
      await events.publish('config.changed', {
        key: 'ui.theme',
        oldValue: currentTheme.value?.id,
        newValue: themeId,
        section: 'ui'
      })
    } catch (error) {
      console.error('Failed to save theme preference:', error)
    }
  }

  const addNotification = (notification: Omit<Notification, 'id' | 'timestamp'>) => {
    const id = `notification-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const newNotification: Notification = {
      ...notification,
      id,
      timestamp: new Date()
    }

    notifications.value.unshift(newNotification)

    // Auto-remove non-persistent notifications
    if (!notification.persistent) {
      const duration = notification.duration ?? 5000
      setTimeout(() => {
        removeNotification(id)
      }, duration)
    }

    // Play notification sound if enabled
    if (notificationSound.value && notification.type !== 'success') {
      playNotificationSound()
    }
  }

  const removeNotification = (id: string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
  }

  const clearNotifications = () => {
    notifications.value = []
  }

  const playNotificationSound = () => {
    // Play notification sound
    // This would integrate with the audio service
    try {
      events.publish('audio.notification', { type: 'notification' })
    } catch (error) {
      console.error('Failed to play notification sound:', error)
    }
  }

  const showModal = (modal: Omit<ModalState, 'id'>): string => {
    const id = `modal-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const newModal: ModalState = {
      ...modal,
      id
    }

    modals.value.set(id, newModal)
    return id
  }

  const hideModal = (id: string) => {
    modals.value.delete(id)
  }

  const hideAllModals = () => {
    modals.value.clear()
  }

  const showContextMenu = (x: number, y: number, items: MenuItem[], target?: any) => {
    contextMenu.value = {
      visible: true,
      x,
      y,
      items,
      target
    }
  }

  const hideContextMenu = () => {
    contextMenu.value.visible = false
  }

  const addPanel = (panel: Omit<PanelState, 'id'>): string => {
    const id = `panel-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    const newPanel: PanelState = {
      ...panel,
      id,
      zIndex: Math.max(...Array.from(panels.value.values()).map(p => p.zIndex), 0) + 1
    }

    panels.value.set(id, newPanel)
    return id
  }

  const removePanel = (id: string) => {
    panels.value.delete(id)
    if (activePanel.value === id) {
      activePanel.value = null
    }
  }

  const updatePanel = (id: string, updates: Partial<PanelState>) => {
    const panel = panels.value.get(id)
    if (panel) {
      panels.value.set(id, { ...panel, ...updates })
    }
  }

  const setActivePanel = (id: string | null) => {
    activePanel.value = id
    if (id) {
      // Bring panel to front
      const panel = panels.value.get(id)
      if (panel) {
        const maxZ = Math.max(...Array.from(panels.value.values()).map(p => p.zIndex))
        updatePanel(id, { zIndex: maxZ + 1, active: true })
      }
    }
  }

  const setLoading = (key: string, loading: boolean) => {
    loading.value[key] = loading
  }

  const clearGlobalError = () => {
    globalError.value = null
  }

  const cleanup = () => {
    // Cleanup resources
    hideAllModals()
    hideContextMenu()
    clearNotifications()
    panels.value.clear()
    loading.value = {}
  }

  // Initialize on store creation
  if (typeof window !== 'undefined') {
    initializeApp()
  }

  return {
    // State
    app,
    currentTheme,
    availableThemes,
    customCSSVariables,
    windowState,
    notifications,
    notificationSound,
    modals,
    contextMenu,
    panels,
    activePanel,
    loading,
    globalLoading,
    globalError,

    // Computed
    isAppReady,
    isDarkTheme,
    hasNotifications,
    visiblePanels,
    loadingCount,

    // Actions
    initializeApp,
    setTheme,
    addNotification,
    removeNotification,
    clearNotifications,
    showModal,
    hideModal,
    hideAllModals,
    showContextMenu,
    hideContextMenu,
    addPanel,
    removePanel,
    updatePanel,
    setActivePanel,
    setLoading,
    clearGlobalError,
    cleanup
  }
})

export default useUIStore