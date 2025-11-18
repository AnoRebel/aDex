import { ref, computed, onMounted, onUnmounted, readonly } from 'vue'
import * as ServiceCoordinator from '~/bindings/aDex-UI/backend/services/coordinator/servicecoordinator'

// Type definitions for Wails integration
export interface WailsService {
  IsRunning(): Promise<boolean>
  Start(): Promise<void>
  Stop(): Promise<void>
  GetStatus(): Promise<string>
}

export interface WailsEvent {
  Type: string
  Timestamp: string
  Data: any
  Source: string
}

export interface ServiceStatus {
  name: string
  running: boolean
  status: string
  lastError?: string
}

// Global state
const isInitialized = ref(false)
const isReady = ref(false)
const isWailsAvailable = ref(false)
const coordinatorStarted = ref(false)
const services = ref<Map<string, ServiceStatus>>(new Map())
const events = ref<WailsEvent[]>([])
const error = ref<string | null>(null)
const lastError = ref<string | null>(null)

// Event listeners
const eventListeners = new Map<string, (event: WailsEvent) => void>()

/**
 * useWails composable provides Wails integration functionality
 */
export function useWails() {

  // Computed properties
  const appStatus = computed(() => {
    if (!isInitialized.value) return 'initializing'
    if (!isReady.value) return 'loading'
    if (error.value) return 'error'
    return 'ready'
  })

  const activeServices = computed(() => {
    return Array.from(services.value.values()).filter(service => service.running)
  })

  const failedServices = computed(() => {
    return Array.from(services.value.values()).filter(service => service.lastError)
  })

  // Methods
  const initialize = async (): Promise<void> => {
    try {
      error.value = null
      lastError.value = null
      isInitialized.value = true

      // Check Wails availability first
      if (!checkWailsAvailability()) {
        lastError.value = 'Wails runtime not available'
        throw new Error(lastError.value)
      }

      // Initialize Wails runtime
      if (typeof window !== 'undefined' && window.wails) {
        await window.wails.Init()
      }

      // Initialize the service coordinator
      await ServiceCoordinator.Initialize()
      await ServiceCoordinator.StartMonitoring()
      coordinatorStarted.value = true

      // Load service status
      await loadServiceStatus()

      isReady.value = true
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error during initialization'
      error.value = errorMessage
      lastError.value = errorMessage
      console.error('Wails initialization failed:', err)
      throw err
    }
  }

  const loadServiceStatus = async (): Promise<void> => {
    try {
      // Initialize with default services
      // The actual status would be fetched from the coordinator
      services.value = new Map([
        ['system', { name: 'system', running: true, status: 'active' }],
        ['filesystem', { name: 'filesystem', running: true, status: 'active' }],
        ['terminal', { name: 'terminal', running: true, status: 'active' }],
        ['audio', { name: 'audio', running: true, status: 'active' }],
        ['theme', { name: 'theme', running: true, status: 'active' }],
        ['config', { name: 'config', running: true, status: 'active' }],
      ])
    } catch (err) {
      console.error('Failed to load service status:', err)
    }
  }

  const callService = async <T = any>(
    serviceName: string,
    methodName: string,
    ...args: any[]
  ): Promise<T> => {
    if (!isReady.value) {
      throw new Error('Wails is not ready. Call initialize() first.')
    }

    try {
      // Generic service call implementation
      // This would be extended based on actual service methods
      const service = services.value.get(serviceName)
      if (!service || !service.running) {
        throw new Error(`Service ${serviceName} is not available`)
      }

      // Placeholder for actual service method call
      // Implementation would depend on the specific service interface
      return {} as T
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown service error'

      // Update service error status
      const service = services.value.get(serviceName)
      if (service) {
        service.lastError = errorMessage
        services.value.set(serviceName, { ...service })
      }

      throw new Error(`Service call failed: ${errorMessage}`)
    }
  }

  const subscribeToEvents = (eventType: string, callback: (event: WailsEvent) => void): void => {
    eventListeners.set(eventType, callback)
  }

  const unsubscribeFromEvents = (eventType: string): void => {
    eventListeners.delete(eventType)
  }

  const publishEvent = async (eventType: string, data: any, source: string = 'frontend'): Promise<void> => {
    if (!isReady.value) {
      throw new Error('Wails is not ready. Call initialize() first.')
    }

    try {
      // This would publish to the Go event bus
      // Implementation depends on Wails event system
      const event: WailsEvent = {
        Type: eventType,
        Timestamp: new Date().toISOString(),
        Data: data,
        Source: source
      }

      events.value.push(event)

      // Notify local listeners
      const listener = eventListeners.get(eventType)
      if (listener) {
        listener(event)
      }
    } catch (err) {
      console.error('Failed to publish event:', err)
      throw err
    }
  }

  const clearEvents = (): void => {
    events.value = []
  }

  const getServiceStatus = (serviceName: string): ServiceStatus | undefined => {
    return services.value.get(serviceName)
  }

  const getAllServices = (): ServiceStatus[] => {
    return Array.from(services.value.values())
  }

  const restartService = async (serviceName: string): Promise<void> => {
    try {
      await callService(serviceName, 'restart')
      await loadServiceStatus()
    } catch (err) {
      console.error(`Failed to restart service ${serviceName}:`, err)
      throw err
    }
  }

  // Check if Wails is available
  const checkWailsAvailability = () => {
    isWailsAvailable.value = !!(typeof window !== 'undefined' && window.wails)
    return isWailsAvailable.value
  }

  // Get a service from the coordinator
  const getService = async (serviceType: string) => {
    try {
      return await ServiceCoordinator.GetService(serviceType)
    } catch (error) {
      console.error(`Failed to get service ${serviceType}:`, error)
      return null
    }
  }

  // Service method wrappers - simplified async calls through coordinator
  const system = {
    getSystemInfo: async () => {
      const service = await getService('system')
      return service?.GetSystemInfo()
    },
    getCPUUsage: async () => {
      const service = await getService('system')
      return service?.GetCPUUsage()
    },
    getMemoryUsage: async () => {
      const service = await getService('system')
      return service?.GetMemoryUsage()
    },
    getDiskUsage: async () => {
      const service = await getService('system')
      return service?.GetDiskUsage()
    },
    getNetworkInfo: async () => {
      const service = await getService('system')
      return service?.GetNetworkInfo()
    },
    startMonitoring: async (interval: number) => {
      const service = await getService('system')
      return service?.StartMonitoring(interval)
    },
    stopMonitoring: async () => {
      const service = await getService('system')
      return service?.StopMonitoring()
    }
  }

  const filesystem = {
    readDirectory: async (path: string) => {
      const service = await getService('filesystem')
      return service?.ReadDirectory(path)
    },
    getFileInfo: async (path: string) => {
      const service = await getService('filesystem')
      return service?.GetFileInfo(path)
    },
    createDirectory: async (path: string, mode: number) => {
      const service = await getService('filesystem')
      return service?.CreateDirectory(path, mode)
    },
    deleteFile: async (path: string) => {
      const service = await getService('filesystem')
      return service?.DeleteFile(path)
    },
    copyFile: async (src: string, dst: string) => {
      const service = await getService('filesystem')
      return service?.CopyFile(src, dst)
    },
    moveFile: async (src: string, dst: string) => {
      const service = await getService('filesystem')
      return service?.MoveFile(src, dst)
    },
    readFile: async (path: string) => {
      const service = await getService('filesystem')
      return service?.ReadFile(path)
    },
    writeFile: async (path: string, data: string, mode: number) => {
      const service = await getService('filesystem')
      return service?.WriteFile(path, data, mode)
    },
    searchFiles: async (root: string, pattern: string) => {
      const service = await getService('filesystem')
      return service?.SearchFiles(root, pattern)
    },
    watchDirectory: async (path: string) => {
      const service = await getService('filesystem')
      return service?.WatchDirectory(path)
    }
  }

  const terminal = {
    createTerminal: async (width: number, height: number) => {
      const service = await getService('terminal')
      return service?.CreateTerminal(width, height)
    },
    resizeTerminal: async (terminalId: string, width: number, height: number) => {
      const service = await getService('terminal')
      return service?.ResizeTerminal(terminalId, width, height)
    },
    writeToTerminal: async (terminalId: string, data: string) => {
      const service = await getService('terminal')
      return service?.WriteToTerminal(terminalId, data)
    },
    readFromTerminal: async (terminalId: string) => {
      const service = await getService('terminal')
      return service?.ReadFromTerminal(terminalId)
    },
    closeTerminal: async (terminalId: string) => {
      const service = await getService('terminal')
      return service?.CloseTerminal(terminalId)
    }
  }

  const audio = {
    getDevices: async () => {
      const service = await getService('audio')
      return service?.GetDevices()
    },
    getCurrentVolume: async () => {
      const service = await getService('audio')
      return service?.GetCurrentVolume()
    },
    setVolume: async (volume: number) => {
      const service = await getService('audio')
      return service?.SetVolume(volume)
    },
    playSound: async (soundPath: string) => {
      const service = await getService('audio')
      return service?.PlaySound(soundPath)
    },
    stopSound: async () => {
      const service = await getService('audio')
      return service?.StopSound()
    },
    getSystemSounds: async () => {
      const service = await getService('audio')
      return service?.GetSystemSounds()
    }
  }

  const theme = {
    getCurrentTheme: async () => {
      const service = await getService('theme')
      return service?.GetCurrentTheme()
    },
    setTheme: async (themeName: string) => {
      const service = await getService('theme')
      return service?.SetTheme(themeName)
    },
    getAvailableThemes: async () => {
      const service = await getService('theme')
      return service?.GetAvailableThemes()
    },
    saveTheme: async (themeData: any) => {
      const service = await getService('theme')
      return service?.SaveTheme(themeData)
    }
  }

  const config = {
    getConfig: async (key: string) => {
      const service = await getService('config')
      return service?.GetConfig(key)
    },
    setConfig: async (key: string, value: any) => {
      const service = await getService('config')
      return service?.SetConfig(key, value)
    },
    getAllConfig: async () => {
      const service = await getService('config')
      return service?.GetAllConfig()
    },
    saveConfig: async () => {
      const service = await getService('config')
      return service?.SaveConfig()
    },
    loadConfig: async () => {
      const service = await getService('config')
      return service?.LoadConfig()
    }
  }

  // Get platform info
  const getPlatform = () => ServiceCoordinator.GetPlatform()

  // Cleanup
  const cleanup = (): void => {
    eventListeners.clear()
    events.value = []
    services.value.clear()
    isReady.value = false
    isInitialized.value = false
    isWailsAvailable.value = false
    coordinatorStarted.value = false
    error.value = null
    lastError.value = null
  }

  // Auto-initialize on mount
  onMounted(async () => {
    if (!isInitialized.value) {
      await initialize()
    }
  })

  // Auto-initialize when composable is used (from old version)
  if (typeof window !== 'undefined') {
    checkWailsAvailability()
    if (isWailsAvailable.value) {
      initialize()
    }
  }

  // Shutdown method for cleanup
  const shutdown = async () => {
    try {
      await ServiceCoordinator.Shutdown()
      coordinatorStarted.value = false
      isInitialized.value = false
    } catch (error) {
      console.error('Failed to shutdown service coordinator:', error)
    }
  }

  onUnmounted(() => {
    cleanup()
  })

  return {
    // State (readonly for safety)
    isInitialized: readonly(isInitialized),
    isReady: readonly(isReady),
    isWailsAvailable: readonly(isWailsAvailable),
    coordinatorStarted: readonly(coordinatorStarted),
    services: readonly(services),
    events: readonly(events),
    error: readonly(error),
    lastError: readonly(lastError),

    // Computed
    appStatus,
    activeServices,
    failedServices,

    // Core methods
    initialize,
    shutdown,
    getPlatform,
    checkWailsAvailability,

    // Event methods
    callService,
    subscribeToEvents,
    unsubscribeFromEvents,
    publishEvent,
    clearEvents,
    getServiceStatus,
    getAllServices,
    restartService,
    cleanup,

    // Service-specific methods
    system,
    filesystem,
    terminal,
    audio,
    theme,
    config
  }
}

/**
 * useService composable for working with specific services
 */
export function useService(serviceName: string) {
  const wails = useWails()

  const service = computed(() => wails.services.value.get(serviceName))
  const isRunning = computed(() => service.value?.running ?? false)
  const status = computed(() => service.value?.status ?? 'unknown')
  const lastError = computed(() => service.value?.lastError ?? null)

  const call = async <T = any>(methodName: string, ...args: any[]): Promise<T> => {
    return wails.callService<T>(serviceName, methodName, ...args)
  }

  const start = async (): Promise<void> => {
    await call('start')
    await wails.loadServiceStatus()
  }

  const stop = async (): Promise<void> => {
    await call('stop')
    await wails.loadServiceStatus()
  }

  const restart = async (): Promise<void> => {
    await wails.restartService(serviceName)
  }

  return {
    service,
    isRunning,
    status,
    lastError,
    call,
    start,
    stop,
    restart
  }
}

/**
 * useEvents composable for working with Wails events
 */
export function useEvents() {
  const wails = useWails()

  const subscribe = (eventType: string, callback: (event: WailsEvent) => void) => {
    wails.subscribeToEvents(eventType, callback)
  }

  const unsubscribe = (eventType: string) => {
    wails.unsubscribeFromEvents(eventType)
  }

  const publish = (eventType: string, data: any, source?: string) => {
    return wails.publishEvent(eventType, data, source)
  }

  const getEvents = (eventType?: string) => {
    if (!eventType) return wails.events.value
    return wails.events.value.filter(event => event.Type === eventType)
  }

  const clearEvents = (eventType?: string) => {
    if (!eventType) {
      wails.clearEvents()
    } else {
      // Filter out events of the specified type
      wails.events.value = wails.events.value.filter(event => event.Type !== eventType)
    }
  }

  return {
    events: wails.events,
    subscribe,
    unsubscribe,
    publish,
    getEvents,
    clearEvents
  }
}

// Type declarations for Wails runtime
declare global {
  interface Window {
    wails?: {
      Init(): Promise<void>
      Call(serviceName: string, methodName: string, ...args: any[]): Promise<any>
      Emit(eventName: string, ...args: any[]): void
      On(eventName: string, callback: (...args: any[]) => void): void
      Off(eventName: string, callback?: (...args: any[]) => void): void
    }
  }
}

export default useWails