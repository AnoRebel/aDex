import { ref, computed, onMounted, onUnmounted, readonly, getCurrentInstance } from 'vue'

// Wails v3 runtime facade + generated coordinator bindings.
import { Events, Log } from '~/lib/wailsjs/runtime'
import * as CoordinatorBindings from '~/lib/wailsjs/coordinator'

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

// Type definitions for the Wails integration
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
// Wails v3 unsubscribe functions, one per subscribed event name.
const runtimeUnsubscribers = new Map<string, () => void>()

/**
 * useWails composable provides Wails integration functionality
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

      // The Wails v3 runtime is available as soon as the webview loads;
      // no explicit initialisation step is required.
      
      // Setup event listeners
      setupEventListeners()
      
      isReady.value = true
      Log.info('Wails runtime initialized successfully')
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

      // One runtime subscription per event name, fanned out to the local
      // callback set. Wails v3 returns an unsubscribe function, which we
      // keep so teardown removes exactly this subscription — important for
      // per-session names like `terminal.output.<id>`, where dropping a
      // shared listener would silence other panes.
      const off = Events.On(eventName, (data: any) => {
        eventListeners.get(eventName)?.forEach(cb => cb(data))
      })
      runtimeUnsubscribers.set(eventName, off)
    }
    eventListeners.get(eventName)!.add(callback)
  }

  const unsubscribe = (eventName: string, callback?: (data: any) => void) => {
    if (callback) {
      eventListeners.get(eventName)?.delete(callback)
    }
    // Drop the runtime subscription only once no local callbacks remain.
    const listeners = eventListeners.get(eventName)
    if (!callback || !listeners?.size) {
      runtimeUnsubscribers.get(eventName)?.()
      runtimeUnsubscribers.delete(eventName)
      listeners?.clear()
      eventListeners.delete(eventName)
    }
  }

  const publish = (eventName: string, data?: any) => {
    Events.Emit(eventName, data)
    emit(eventName, data)
  }

  const cleanup = () => {
    runtimeUnsubscribers.forEach(off => off())
    runtimeUnsubscribers.clear()
    eventListeners.clear()
    events.value = []
    services.value.clear()
    isReady.value = false
    isInitialized.value = false
    error.value = null
  }

  // Only register the hook when there IS a component to attach it to.
  // Stores call useWails() from actions, outside any setup(), where Vue warns
  // "onUnmounted is called when there is no active component instance" on
  // every invocation.
  if (getCurrentInstance()) {
    onUnmounted(() => {
      cleanup()
    })
  }

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
    // Wails v3 has no `window.go` IPC global: bound methods are reached
    // through the generated, typed bindings, which this module re-exports.
    // Resolving by string here keeps the historical dynamic-call shape, but
    // prefer importing the binding directly — that gets compile-time checking
    // on both the method name and its arguments.
    const bound = (CoordinatorBindings as Record<string, unknown>)[methodName]
    if (typeof bound !== 'function') {
      throw new Error(`Method ${serviceName}.${methodName} not found in generated bindings`)
    }
    return (bound as (...a: any[]) => Promise<T>)(...args)
  }

  return {
    isRunning,
    status,
    call
  }
}

// Default export
export default useWails
