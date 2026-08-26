import { ref, computed, onMounted, onUnmounted, readonly } from 'vue'

// Import Wails runtime (local implementation for v2)
import { Events, Log } from '~/lib/wailsjs/runtime'

// Import coordinator bindings
import {
  GetCPUUsage,
  GetMemoryUsage,
  GetDiskUsage,
  GetSystemInfo,
  GetTopProcesses,
  CreateTerminal,
  WriteToTerminal,
  ResizeTerminal,
  CloseTerminal,
  ReadDirectory
} from '~/lib/wailsjs/coordinator'

// Type definitions for Wails v2 integration
export interface WailsEvent {
  name: string
  data?: any
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
const error = ref<string | null>(null)
const services = ref<Map<string, ServiceStatus>>(new Map())
const events = ref<WailsEvent[]>([])

// Event listeners map
const eventListeners = new Map<string, Set<(data: any) => void>>()

/**
 * useWails composable provides Wails v2 integration functionality
 */
export function useWails() {
  const appStatus = computed(() => {
    if (!isInitialized.value) return 'initializing'
    if (!isReady.value) return 'loading'
    if (error.value) return 'error'
    return 'ready'
  })

  const activeServices = computed(() => {
    return Array.from(services.value.values()).filter(service => service.running)
  })

  const initialize = async (): Promise<void> => {
    try {
      error.value = null
      isInitialized.value = true

      // Wails v2 runtime is automatically available
      // No explicit init needed
      
      // Setup event listeners
      setupEventListeners()
      
      isReady.value = true
      Log.info('Wails v2 initialized successfully')
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error during initialization'
      error.value = errorMessage
      Log.error('Wails initialization failed: ' + errorMessage)
      throw err
    }
  }

  const setupEventListeners = () => {
    // Listen to all Wails events
    Events.On('time.updated', (time: string) => {
      emit('time.updated', { time })
    })
  }

  const emit = (eventName: string, data?: any) => {
    const event: WailsEvent = { name: eventName, data }
    events.value.push(event)
    
    // Notify local listeners
    const listeners = eventListeners.get(eventName)
    if (listeners) {
      listeners.forEach(callback => callback(data))
    }
  }

  const subscribe = (eventName: string, callback: (data: any) => void) => {
    if (!eventListeners.has(eventName)) {
      eventListeners.set(eventName, new Set())
      
      // Also subscribe to Wails runtime events
      Events.On(eventName, callback)
    }
    eventListeners.get(eventName)!.add(callback)
  }

  const unsubscribe = (eventName: string, callback?: (data: any) => void) => {
    if (callback) {
      eventListeners.get(eventName)?.delete(callback)
    }
    // Wails v2 EventsOff takes event names (not callbacks)
    const listeners = eventListeners.get(eventName)
    if (!callback || !listeners?.size) {
      Events.Off(eventName)
      listeners?.clear()
    }
  }

  const publish = (eventName: string, data?: any) => {
    Events.Emit(eventName, data)
    emit(eventName, data)
  }

  const cleanup = () => {
    eventListeners.forEach((_listeners, eventName) => {
      Events.Off(eventName)
    })
    eventListeners.clear()
    events.value = []
    services.value.clear()
    isReady.value = false
    isInitialized.value = false
    error.value = null
  }

  onUnmounted(() => {
    cleanup()
  })

  // System service methods - calls the Go backend via Wails bindings
  const system = {
    async getSystemInfo() {
      try {
        const info: any = await GetSystemInfo()
        if (info) {
          // Backend now uses explicit camelCase JSON tags (`platform`,
          // `architecture`, `kernelVersion`, ...). The legacy
          // PascalCase fallbacks stay so a partially-regenerated
          // build can't break boot.
          return {
            hostname: info.hostname || info.Hostname || 'localhost',
            platform: info.platform || info.OS || info.os || 'linux',
            os: info.platform || info.OS || info.os || 'Linux',
            arch: info.architecture || info.Architecture || 'x64',
            uptime: info.uptime ? Number(info.uptime) / 1e9 : (info.Uptime ? Number(info.Uptime) / 1e9 : 0),
            kernel: info.kernelVersion || info.KernelVersion || info.kernel || '',
          }
        }
        return { hostname: 'localhost', platform: 'linux', os: 'Linux', arch: 'x64', uptime: 0, kernel: '' }
      } catch (e) {
        console.error('getSystemInfo error:', e)
        return { hostname: 'localhost', platform: 'linux', os: 'Linux', arch: 'x64', uptime: 0, kernel: '' }
      }
    },
    async getCPUUsage() {
      try {
        const result = await GetCPUUsage()
        console.log('CPU Usage result:', result)
        return { 
          usage: result?.usage || 0, 
          cores: result?.cores || [],
          coreCount: result?.coreCount || result?.cores?.length || 0,
          modelName: result?.modelName || '',
          frequency: result?.frequency || 0
        }
      } catch (e) {
        console.error('getCPUUsage error:', e)
        return { usage: 0, cores: [], coreCount: 0, modelName: '', frequency: 0 }
      }
    },
    async getMemoryUsage() {
      try {
        const result = await GetMemoryUsage()
        console.log('Memory Usage result:', result)
        return { 
          total: result?.total || 0, 
          used: result?.used || 0, 
          free: result?.free || 0, 
          usage: result?.usage || 0 
        }
      } catch (e) {
        console.error('getMemoryUsage error:', e)
        return { total: 0, used: 0, free: 0, usage: 0 }
      }
    },
    async getDiskUsage() {
      try {
        const result = await GetDiskUsage()
        return result || []
      } catch (e) {
        console.error('getDiskUsage error:', e)
        return []
      }
    },
    async getTopProcesses(metric: string = 'cpu', limit: number = 10) {
      try {
        return await GetTopProcesses(metric, limit)
      } catch (e) {
        console.error('getTopProcesses error:', e)
        return []
      }
    },
    async startMonitoring(interval: number) {
      try {
        // Monitoring is handled by the backend automatically
        console.log('System monitoring started')
      } catch (e) {
        console.error('startMonitoring error:', e)
      }
    },
    async stopMonitoring() {
      try {
        console.log('System monitoring stopped')
      } catch (e) {
        console.error('stopMonitoring error:', e)
      }
    }
  }

  // Terminal service methods
  const terminal = {
    async create(cols: number, rows: number) {
      try {
        return await CreateTerminal(cols, rows)
      } catch (e) {
        console.error('terminal.create error:', e)
        return { id: `term-${Date.now()}` }
      }
    },
    async write(terminalId: string, data: string) {
      try {
        await WriteToTerminal(terminalId, data)
      } catch (e) {
        console.error('terminal.write error:', e)
      }
    },
    async resize(terminalId: string, cols: number, rows: number) {
      try {
        await ResizeTerminal(terminalId, cols, rows)
      } catch (e) {
        console.error('terminal.resize error:', e)
      }
    },
    async close(terminalId: string) {
      try {
        await CloseTerminal(terminalId)
      } catch (e) {
        console.error('terminal.close error:', e)
      }
    }
  }

  // Filesystem service methods
  const filesystem = {
    async readDirectory(path: string) {
      try {
        return await ReadDirectory(path)
      } catch (e) {
        console.error('filesystem.readDirectory error:', e)
        return []
      }
    },
    async getCurrentDirectory() {
      return '/'
    }
  }

  return {
    isInitialized: readonly(isInitialized),
    isReady: readonly(isReady),
    error: readonly(error),
    services: readonly(services),
    events: readonly(events),
    appStatus,
    activeServices,
    initialize,
    emit,
    subscribe,
    unsubscribe,
    publish,
    cleanup,
    system,
    terminal,
    filesystem
  }
}

/**
 * useEvents composable for working with Wails events
 */
export function useEvents() {
  const wails = useWails()

  const subscribe = (eventName: string, callback: (data: any) => void) => {
    wails.subscribe(eventName, callback)
  }

  const unsubscribe = (eventName: string, callback?: (data: any) => void) => {
    wails.unsubscribe(eventName, callback)
  }

  const publish = (eventName: string, data?: any) => {
    return wails.publish(eventName, data)
  }

  return {
    events: wails.events,
    subscribe,
    unsubscribe,
    publish
  }
}

/**
 * useService composable for working with specific services
 * In Wails v2, services are bound Go structs
 */
export function useService(serviceName: string) {
  const wails = useWails()

  const isRunning = computed(() => {
    const service = wails.services.value.get(serviceName)
    return service?.running ?? false
  })

  const status = computed(() => {
    const service = wails.services.value.get(serviceName)
    return service?.status ?? 'unknown'
  })

  const call = async <T = any>(methodName: string, ...args: any[]): Promise<T> => {
    // In Wails v2, bound methods are available on window.go.{ServiceName}.{MethodName}
    const boundMethod = (window as any).go?.[serviceName]?.[methodName]
    if (!boundMethod) {
      throw new Error(`Method ${serviceName}.${methodName} not found`)
    }
    return boundMethod(...args)
  }

  return {
    isRunning,
    status,
    call
  }
}

// Default export
export default useWails
