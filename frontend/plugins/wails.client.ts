import { defineNuxtPlugin } from '#app'
import { useWails, useEvents, useService } from '~/composables/useWails'
import { useUIStore } from '~/stores/ui'

// Plugin options interface
interface WailsClientOptions {
  autoInitialize?: boolean
  enableEventLogging?: boolean
  enableErrorHandling?: boolean
  services?: string[]
}

export default defineNuxtPlugin(async (nuxtApp) => {
  // Default plugin options
  const options: WailsClientOptions = {
    autoInitialize: true,
    enableEventLogging: false,
    enableErrorHandling: true,
    services: ['coordinator', 'system', 'terminal', 'filesystem', 'audio', 'config', 'theme']
  }

  // Only run on client side
  if (process.client) {
    try {
      // Initialize Wails composables
      const wails = useWails()
      const events = useEvents()
      const uiStore = useUIStore()

      // Auto-initialize if enabled
      if (options.autoInitialize && !wails.isInitialized.value) {
        await wails.initialize()
      }

      // Setup global error handling if enabled
      if (options.enableErrorHandling) {
        setupGlobalErrorHandling(wails, events, uiStore)
      }

      // Setup event logging if enabled
      if (options.enableEventLogging) {
        setupEventLogging(events)
      }

      // Setup services
      await setupServices(options.services, wails, events)

      // Provide Wails instance globally
      nuxtApp.provide('wails', wails)
      nuxtApp.provide('events', events)

      // Add to Vue app context
      nuxtApp.vueApp.config.globalProperties.$wails = wails
      nuxtApp.vueApp.config.globalProperties.$events = events

      console.log('Wails client plugin initialized successfully')

    } catch (error) {
      console.error('Failed to initialize Wails client plugin:', error)

      // Provide fallback instances
      nuxtApp.provide('wails', null)
      nuxtApp.provide('events', null)
      nuxtApp.vueApp.config.globalProperties.$wails = null
      nuxtApp.vueApp.config.globalProperties.$events = null
    }
  }
})

/**
 * Setup global error handling for Wails operations
 */
function setupGlobalErrorHandling(wails: ReturnType<typeof useWails>, events: ReturnType<typeof useEvents>, uiStore: ReturnType<typeof useUIStore>) {
  // Listen to error events
  events.subscribe('error.occurred', (event) => {
    const errorData = event.Data
    if (errorData) {
      uiStore.addNotification({
        type: 'error',
        title: 'Application Error',
        message: errorData.user_message || errorData.message || 'An unexpected error occurred',
        persistent: errorData.type === 'panic'
      })
    }
  })

  // Setup Vue error handler
  nuxtApp.vueApp.config.errorHandler = (error, instance, info) => {
    console.error('Vue error:', error, info)

    events.publish('error.occurred', {
      error: error.message,
      type: 'vue_error',
      context: info,
      stack: error.stack
    }, 'frontend')
  }

  // Setup unhandled promise rejection handler
  window.addEventListener('unhandledrejection', (event) => {
    console.error('Unhandled promise rejection:', event.reason)

    events.publish('error.occurred', {
      error: event.reason?.message || 'Unhandled promise rejection',
      type: 'promise_rejection',
      context: 'global',
      stack: event.reason?.stack
    }, 'frontend')
  })
}

/**
 * Setup event logging for debugging
 */
function setupEventLogging(events: ReturnType<typeof useEvents>) {
  // Subscribe to all events
  events.subscribe('*', (event) => {
    console.log(`[Event] ${event.Type}:`, event.Data)
  })
}

/**
 * Setup services that should be initialized
 */
async function setupServices(serviceNames: string[], wails: ReturnType<typeof useWails>, events: ReturnType<typeof useEvents>) {
  const servicePromises = serviceNames.map(async (serviceName) => {
    try {
      const service = useService(serviceName)

      // Initialize service if it has an initialize method
      if (service.isRunning.value === false) {
        await service.call('initialize')
      }

      console.log(`Service ${serviceName} initialized successfully`)

      return { name: serviceName, success: true }
    } catch (error) {
      console.error(`Failed to initialize service ${serviceName}:`, error)

      return {
        name: serviceName,
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      }
    }
  })

  const results = await Promise.allSettled(servicePromises)

  // Log initialization results
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      const { name, success, error } = result.value
      if (success) {
        console.log(`✓ ${name} service ready`)
      } else {
        console.error(`✗ ${name} service failed:`, error)
      }
    } else {
      console.error(`✗ Service initialization failed:`, result.reason)
    }
  })

  // Publish service initialization complete event
  events.publish('app.services.initialized', {
    services: results.map((result, index) => ({
      name: serviceNames[index],
      success: result.status === 'fulfilled' ? result.value.success : false
    }))
  })
}

/**
 * Type declarations for Nuxt plugin
 */
declare module '#app' {
  interface NuxtApp {
    $wails: ReturnType<typeof useWails> | null
    $events: ReturnType<typeof useEvents> | null
  }
}

declare module 'vue' {
  interface ComponentCustomProperties {
    $wails: ReturnType<typeof useWails> | null
    $events: ReturnType<typeof useEvents> | null
  }
}

/**
 * Vue composables for easy access to Wails functionality
 */
export const useWailsClient = () => {
  const nuxtApp = useNuxtApp()
  return nuxtApp.$wails as ReturnType<typeof useWails> | null
}

export const useWailsEvents = () => {
  const nuxtApp = useNuxtApp()
  return nuxtApp.$events as ReturnType<typeof useEvents> | null
}

/**
 * Service factory composables
 */
export const useSystemService = () => {
  const wails = useWailsClient()
  if (!wails) {
    throw new Error('Wails client not available')
  }
  return useService('system')
}

export const useTerminalService = () => {
  const wails = useWailsClient()
  if (!wails) {
    throw new Error('Wails client not available')
  }
  return useService('terminal')
}

export const useFileSystemService = () => {
  const wails = useWailsClient()
  if (!wails) {
    throw new Error('Wails client not available')
  }
  return useService('filesystem')
}

export const useAudioService = () => {
  const wails = useWailsClient()
  if (!wails) {
    throw new Error('Wails client not available')
  }
  return useService('audio')
}

export const useConfigService = () => {
  const wails = useWailsClient()
  if (!wails) {
    throw new Error('Wails client not available')
  }
  return useService('config')
}

export const useThemeService = () => {
  const wails = useWailsClient()
  if (!wails) {
    throw new Error('Wails client not available')
  }
  return useService('theme')
}

/**
 * Error handling utilities
 */
export const handleWailsError = (error: any, context?: string) => {
  console.error(`Wails error${context ? ` in ${context}` : ''}:`, error)

  const uiStore = useUIStore()
  uiStore.addNotification({
    type: 'error',
    title: 'Wails Error',
    message: error?.message || 'An unexpected error occurred in Wails',
    persistent: false
  })
}

/**
 * Event utilities
 */
export const waitForEvent = (eventType: string, timeout = 5000): Promise<any> => {
  return new Promise((resolve, reject) => {
    const events = useWailsEvents()
    if (!events) {
      reject(new Error('Events not available'))
      return
    }

    const timeoutId = setTimeout(() => {
      events.unsubscribe(eventType)
      reject(new Error(`Timeout waiting for event: ${eventType}`))
    }, timeout)

    events.subscribe(eventType, (event) => {
      clearTimeout(timeoutId)
      resolve(event.Data)
    })
  })
}

/**
 * Utility to check if Wails is available
 */
export const isWailsAvailable = (): boolean => {
  return process.client && typeof window !== 'undefined' && !!window.wails
}

/**
 * Utility to wait for Wails to be ready
 */
export const waitForWails = (timeout = 10000): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (!isWailsAvailable()) {
      reject(new Error('Wails is not available'))
      return
    }

    const checkInterval = setInterval(() => {
      const wails = useWailsClient()
      if (wails?.isReady.value) {
        clearInterval(checkInterval)
        resolve()
      }
    }, 100)

    setTimeout(() => {
      clearInterval(checkInterval)
      reject(new Error('Timeout waiting for Wails to be ready'))
    }, timeout)
  })
}