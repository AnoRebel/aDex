import { defineStore } from 'pinia'
import { WindowRuntime } from '~/lib/wailsjs/runtime'
import { IsStarted } from '~/lib/wailsjs/coordinator'
import { nowMs } from '~/utils/now'

interface AppAlert {
  id: string
  type: 'success' | 'warning' | 'error' | 'info'
  title: string
  message: string
  timestamp: Date
  acknowledged: boolean
  persistent?: boolean
  actions?: Array<{
    label: string
    action: () => void
    primary?: boolean
  }>
}

interface AppState {
  isInitialized: boolean
  isLoading: boolean
  version: string
  buildNumber: string
  environment: 'development' | 'production' | 'testing'
  alerts: AppAlert[]
  notifications: Array<{
    id: string
    type: 'system' | 'network' | 'filesystem' | 'terminal' | 'audio'
    title: string
    message: string
    timestamp: Date
    read: boolean
  }>
  settings: {
    autoStart: boolean
    minimizeToTray: boolean
    startInTray: boolean
    enableSounds: boolean
    enableNotifications: boolean
    enableAutoSave: boolean
    autoSaveInterval: number
    language: string
    debugMode: boolean
    performanceMode: boolean
  }
  performance: {
    memoryUsage: number
    cpuUsage: number
    uptime: number
    lastUpdate: Date
  }
  shortcuts: Array<{
    id: string
    name: string
    keys: string[]
    action: string
    enabled: boolean
  }>
  status: {
    backend: 'connected' | 'disconnected' | 'connecting' | 'error'
    database: 'connected' | 'disconnected' | 'error'
    filesystem: 'ready' | 'busy' | 'error'
    network: 'online' | 'offline' | 'limited'
  }
}

export const useAppStore = defineStore('app', {
  state: (): AppState => ({
    isInitialized: false,
    isLoading: false,
    version: '2.0.0',
    buildNumber: '2024.01.01',
    environment: 'development',
    alerts: [],
    notifications: [],
    settings: {
      autoStart: true,
      minimizeToTray: false,
      startInTray: false,
      enableSounds: true,
      enableNotifications: true,
      enableAutoSave: true,
      autoSaveInterval: 30000, // 30 seconds
      language: 'en',
      debugMode: false,
      performanceMode: false,
    },
    performance: {
      memoryUsage: 0,
      cpuUsage: 0,
      uptime: 0,
      lastUpdate: new Date(),
    },
    shortcuts: [
      {
        id: 'toggle-terminal',
        name: 'Toggle Terminal',
        keys: ['Ctrl', '`'],
        action: 'toggleTerminal',
        enabled: true,
      },
      {
        id: 'toggle-filebrowser',
        name: 'Toggle File Browser',
        keys: ['Ctrl', 'Shift', 'F'],
        action: 'toggleFileBrowser',
        enabled: true,
      },
      {
        id: 'toggle-system-monitor',
        name: 'Toggle System Monitor',
        keys: ['Ctrl', 'Shift', 'S'],
        action: 'toggleSystemMonitor',
        enabled: true,
      },
      {
        id: 'settings',
        name: 'Settings',
        keys: ['Ctrl', ','],
        action: 'openSettings',
        enabled: true,
      },
      {
        id: 'quit',
        name: 'Quit Application',
        keys: ['Ctrl', 'Q'],
        action: 'quit',
        enabled: true,
      },
    ],
    status: {
      backend: 'disconnected',
      database: 'disconnected',
      filesystem: 'ready',
      network: 'online',
    },
  }),

  getters: {
    isReady: (state) => state.isInitialized && !state.isLoading,

    unacknowledgedAlerts: (state) => state.alerts.filter(alert => !alert.acknowledged),

    criticalAlerts: (state) =>
      state.alerts.filter(alert =>
        alert.type === 'error' && !alert.acknowledged
      ),

    unreadNotifications: (state) =>
      state.notifications.filter(notification => !notification.read),

    systemStatus: (state) => {
      const statuses = Object.values(state.status)
      if (statuses.some(status => status === 'error')) return 'error'
      if (statuses.some(status => status === 'disconnected' || status === 'offline')) return 'warning'
      return 'healthy'
    },

    enabledShortcuts: (state) =>
      state.shortcuts.filter(shortcut => shortcut.enabled),

    systemInfo: (state) => ({
      version: state.version,
      build: state.buildNumber,
      environment: state.environment,
      uptime: state.performance.uptime,
      memory: state.performance.memoryUsage,
      cpu: state.performance.cpuUsage,
    }),

    canExit: (state) => true, // Always allow exit for now
  },

  actions: {
    async initialize(): Promise<void> {
      try {
        this.setLoading(true)

        // Load settings
        this.loadSettings()

        // Check backend connection
        await this.checkBackendConnection()

        // Initialize performance monitoring
        this.startPerformanceMonitoring()

        // Setup keyboard shortcuts
        this.setupKeyboardShortcuts()

        // Mark as initialized
        this.isInitialized = true

        this.addAlert({
          type: 'success',
          title: 'Application Started',
          message: `aDex-UI v${this.version} is ready`,
          persistent: false,
        })

      } catch (error) {
        console.error('Failed to initialize application:', error)
        this.addAlert({
          type: 'error',
          title: 'Initialization Failed',
          message: `Failed to start application: ${error instanceof Error ? error.message : 'Unknown error'}`,
          persistent: true,
        })
      } finally {
        this.setLoading(false)
      }
    },

    async checkBackendConnection(): Promise<boolean> {
      try {
        this.setStatus('backend', 'connecting')

        // Probe the backend through the generated bindings. Wails v3 has no
        // `window.go` IPC global; a failed call throws rather than the
        // binding being absent.
        {
          {
            const isStarted = await IsStarted()
            if (isStarted) {
              this.setStatus('backend', 'connected')
              return true
            }
          }
        }
        
        // Wails bindings not available yet, but that's okay in dev mode
        console.warn('Wails bindings not available, continuing in limited mode')
        this.setStatus('backend', 'disconnected')
        return false
      } catch (error) {
        console.error('Backend connection check failed:', error)
        // Don't show error alert here - let the app continue
        this.setStatus('backend', 'disconnected')
        return false
      }
    },

    startPerformanceMonitoring(): void {
      // Wails' WebView2 (Windows) and WebKit (macOS) ship a partial
      // `performance` polyfill — `performance.now` and `performance.memory`
      // can be undefined depending on the host build. Capture the start
      // time once via Date.now() and use it as the fallback so we always
      // return seconds since first measurement, never throw.
      const startedAt = Date.now()
      const hasNow = typeof performance !== 'undefined'
        && typeof performance.now === 'function'

      const updatePerformance = () => {
        if (typeof performance !== 'undefined' && 'memory' in performance) {
          const memory = (performance as any).memory
          if (memory && typeof memory.usedJSHeapSize === 'number') {
            this.performance.memoryUsage = memory.usedJSHeapSize / 1024 / 1024 // MB
          }
        }

        if (hasNow) {
          this.performance.uptime = nowMs() / 1000 // seconds since page load
        } else {
          this.performance.uptime = (Date.now() - startedAt) / 1000
        }
        this.performance.lastUpdate = new Date()
      }

      // Update every 5 seconds
      setInterval(updatePerformance, 5000)
      updatePerformance() // Initial update
    },

    setupKeyboardShortcuts(): void {
      if (typeof window === 'undefined') return

      const handleKeyDown = (event: KeyboardEvent) => {
        const keys = []
        if (event.ctrlKey) keys.push('Ctrl')
        if (event.shiftKey) keys.push('Shift')
        if (event.altKey) keys.push('Alt')
        if (event.metaKey) keys.push('Meta')
        if (event.key && !['Control', 'Shift', 'Alt', 'Meta'].includes(event.key)) {
          keys.push(event.key)
        }

        const shortcut = this.enabledShortcuts.find(s =>
          s.keys.length === keys.length &&
          s.keys.every(key => keys.includes(key))
        )

        if (shortcut) {
          event.preventDefault()
          this.executeShortcut(shortcut.action)
        }
      }

      document.addEventListener('keydown', handleKeyDown)

      // Store cleanup function (called in reset)
      this._cleanupKeyboardShortcuts = () => {
        document.removeEventListener('keydown', handleKeyDown)
      }
    },

    executeShortcut(action: string): void {
      switch (action) {
        case 'toggleTerminal':
          // Toggle terminal window
          window.dispatchEvent(new CustomEvent('toggle-terminal'))
          break

        case 'toggleFileBrowser':
          // Toggle file browser window
          window.dispatchEvent(new CustomEvent('toggle-filebrowser'))
          break

        case 'toggleSystemMonitor':
          // Toggle system monitor window
          window.dispatchEvent(new CustomEvent('toggle-systemmonitor'))
          break

        case 'openSettings':
          // Open settings modal
          window.dispatchEvent(new CustomEvent('open-settings'))
          break

        case 'quit':
          // Quit application
          this.quit()
          break

        default:
          console.warn('Unknown shortcut action:', action)
      }
    },

    setLoading(isLoading: boolean): void {
      this.isLoading = isLoading
    },

    setStatus(component: keyof AppState['status'], status: AppState['status'][typeof component]): void {
      this.status[component] = status
    },

    addAlert(alert: Omit<AppAlert, 'id' | 'timestamp'>): string {
      const id = Date.now().toString()
      const fullAlert: AppAlert = {
        id,
        timestamp: new Date(),
        acknowledged: false,
        persistent: false,
        ...alert,
      }

      this.alerts.push(fullAlert)

      // Keep only last 100 alerts
      if (this.alerts.length > 100) {
        this.alerts.shift()
      }

      // Auto-remove non-persistent info alerts after 10 seconds
      if (alert.type === 'info' && !alert.persistent) {
        setTimeout(() => {
          this.removeAlert(id)
        }, 10000)
      }

      // Auto-remove success alerts after 5 seconds
      if (alert.type === 'success' && !alert.persistent) {
        setTimeout(() => {
          this.removeAlert(id)
        }, 5000)
      }

      return id
    },

    removeAlert(id: string): void {
      this.alerts = this.alerts.filter(alert => alert.id !== id)
    },

    acknowledgeAlert(id: string): void {
      const alert = this.alerts.find(a => a.id === id)
      if (alert) {
        alert.acknowledged = true
      }
    },

    acknowledgeAllAlerts(): void {
      this.alerts.forEach(alert => {
        alert.acknowledged = true
      })
    },

    addNotification(notification: Omit<AppState['notifications'][0], 'id' | 'timestamp' | 'read'>): string {
      const id = Date.now().toString()
      const fullNotification = {
        id,
        timestamp: new Date(),
        read: false,
        ...notification,
      }

      this.notifications.push(fullNotification)

      // Keep only last 200 notifications
      if (this.notifications.length > 200) {
        this.notifications.shift()
      }

      return id
    },

    markNotificationAsRead(id: string): void {
      const notification = this.notifications.find(n => n.id === id)
      if (notification) {
        notification.read = true
      }
    },

    markAllNotificationsAsRead(): void {
      this.notifications.forEach(notification => {
        notification.read = true
      })
    },

    updateSettings(newSettings: Partial<AppState['settings']>): void {
      this.settings = { ...this.settings, ...newSettings }
      this.saveSettings()
    },

    saveSettings(): void {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('adex-app-settings', JSON.stringify(this.settings))
      }
    },

    loadSettings(): void {
      if (typeof localStorage !== 'undefined') {
        const saved = localStorage.getItem('adex-app-settings')
        if (saved) {
          try {
            const settings = JSON.parse(saved)
            this.settings = { ...this.settings, ...settings }
          } catch (error) {
            console.error('Failed to load app settings:', error)
          }
        }
      }
    },

    addShortcut(shortcut: Omit<AppState['shortcuts'][0], 'id'>): string {
      const id = Date.now().toString()
      this.shortcuts.push({ ...shortcut, id })
      return id
    },

    removeShortcut(id: string): void {
      this.shortcuts = this.shortcuts.filter(shortcut => shortcut.id !== id)
    },

    updateShortcut(id: string, updates: Partial<AppState['shortcuts'][0]>): void {
      const shortcut = this.shortcuts.find(s => s.id === id)
      if (shortcut) {
        Object.assign(shortcut, updates)
      }
    },

    async checkForUpdates(): Promise<{ hasUpdate: boolean; version?: string }> {
      try {
        // The Go coordinator exposes no CheckForUpdates method, so this never
        // had a backend to call: under v2 the `window.go` probe just failed
        // its guard and fell through to the same result returned here. Real
        // update checking runs in the frontend against GitHub Releases — see
        // useUpdateChecker.
        return { hasUpdate: false }
      } catch (error) {
        console.error('Failed to check for updates:', error)
        return { hasUpdate: false }
      }
    },

    async quit(): Promise<void> {
      try {
        await this.cleanup()

        // Ask the Wails runtime to quit. This begins the same teardown the
        // Go side runs: ShouldQuit, then each service's ServiceShutdown in
        // reverse registration order.
        if (typeof window !== 'undefined') {
          WindowRuntime.Quit()
        }
      } catch (error) {
        console.error('Error during quit:', error)
        if (typeof window !== 'undefined') {
          window.close()
        }
      }
    },

    async cleanup(): Promise<void> {
      // Cleanup keyboard shortcuts
      if (this._cleanupKeyboardShortcuts) {
        this._cleanupKeyboardShortcuts()
      }

      // Save settings
      this.saveSettings()

      // Acknowledge all alerts
      this.acknowledgeAllAlerts()
    },

    async restart(): Promise<void> {
      try {
        await this.cleanup()

        // Wails v3 has no WindowReloadApp; reloading the webview is a plain
        // browser reload, which re-runs the frontend against the still-running
        // backend services.
        if (typeof window !== 'undefined') {
          window.location.reload()
        }
      } catch (error) {
        console.error('Error during restart:', error)
      }
    },

    // Private cleanup function
    _cleanupKeyboardShortcuts(): void {
      // This will be set by setupKeyboardShortcuts
    },

    
    // Reset store
    reset(): void {
      if (this._cleanupKeyboardShortcuts) {
        this._cleanupKeyboardShortcuts()
      }

      this.isInitialized = false
      this.isLoading = false
      this.alerts = []
      this.notifications = []
      this.status = {
        backend: 'disconnected',
        database: 'disconnected',
        filesystem: 'ready',
        network: 'online',
      }
    },
  },
})

// Extend the interface to include the cleanup method
declare module 'pinia' {
  export interface PiniaCustomProperties {
    _cleanupKeyboardShortcuts?: () => void
  }
}