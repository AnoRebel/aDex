import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { GeoIPInfo, GeoIPSettings, GeoIPCacheStats, GeoIPConfig } from '~/app/types/geoip'

export function useGeoIP() {
  const isEnabled = ref(false)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const cacheStats = ref<GeoIPCacheStats | null>(null)
  const settings = ref<GeoIPSettings>({
    enabled: false,
    provider: 'ipapi.co',
    showFlags: true,
    showCity: true,
    showISP: false,
    cacheTimeout: 24 * 60 * 60 * 1000, // 24 hours
    autoLookup: true,
    privacyMode: true
  })

  // Wails bindings will be generated, assuming they exist
  // These would be the actual Wails method calls
  const networkService = $ref<any>(null)

  // Computed properties
  const isGeoIPEnabled = computed(() => isEnabled.value && settings.value.enabled)
  const hasError = computed(() => !!error.value)
  const cacheHitRate = computed(() => {
    if (!cacheStats.value) return 0
    return cacheStats.value.size > 0 ? Math.min((cacheStats.value.size / cacheStats.value.maxSize) * 100, 100) : 0
  })

  // Methods
  const enableGeoIP = async () => {
    try {
      isLoading.value = true
      error.value = null

      if (networkService) {
        await networkService.EnableGeoIP()
        isEnabled.value = true
        settings.value.enabled = true

        // Emit event for other components
        if (process.client) {
          window.dispatchEvent(new CustomEvent('geoip:enabled'))
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to enable GeoIP service'
      console.error('Failed to enable GeoIP:', err)
    } finally {
      isLoading.value = false
    }
  }

  const disableGeoIP = async () => {
    try {
      isLoading.value = true
      error.value = null

      if (networkService) {
        await networkService.DisableGeoIP()
        isEnabled.value = false
        settings.value.enabled = false
        cacheStats.value = null

        // Emit event for other components
        if (process.client) {
          window.dispatchEvent(new CustomEvent('geoip:disabled'))
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to disable GeoIP service'
      console.error('Failed to disable GeoIP:', err)
    } finally {
      isLoading.value = false
    }
  }

  const getGeoIPData = async (ip: string): Promise<GeoIPInfo | null> => {
    if (!isGeoIPEnabled.value) {
      return null
    }

    try {
      isLoading.value = true
      error.value = null

      if (networkService) {
        const data = await networkService.GetGeoIPDataForIP(ip)
        return data
      }
      return null
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : `Failed to get GeoIP data for ${ip}`
      error.value = errorMessage
      console.error('GeoIP lookup failed:', err)
      return null
    } finally {
      isLoading.value = false
    }
  }

  const updateSettings = async (newSettings: Partial<GeoIPSettings>) => {
    try {
      isLoading.value = true
      error.value = null

      const updatedSettings = { ...settings.value, ...newSettings }

      // Convert to GeoIPConfig format for backend
      const config: GeoIPConfig = {
        enabled: updatedSettings.enabled,
        provider: updatedSettings.provider,
        cacheTimeout: updatedSettings.cacheTimeout,
        requestTimeout: 5000,
        maxCacheSize: 1000,
        updateInterval: 60 * 60 * 1000, // 1 hour
        apiKey: undefined
      }

      if (networkService) {
        await networkService.SetGeoIPConfig(config)
        settings.value = updatedSettings

        // Update enabled state
        if (updatedSettings.enabled !== settings.value.enabled) {
          isEnabled.value = updatedSettings.enabled
        }

        // Emit settings changed event
        if (process.client) {
          window.dispatchEvent(new CustomEvent('geoip:settings-changed', {
            detail: updatedSettings
          }))
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update GeoIP settings'
      console.error('Failed to update GeoIP settings:', err)
    } finally {
      isLoading.value = false
    }
  }

  const refreshCacheStats = async () => {
    try {
      if (networkService && isGeoIPEnabled.value) {
        const stats = await networkService.GetGeoIPCacheStats()
        cacheStats.value = stats
      }
    } catch (err) {
      console.error('Failed to get cache stats:', err)
    }
  }

  const clearCache = async () => {
    try {
      isLoading.value = true
      error.value = null

      if (networkService) {
        await networkService.ClearGeoIPCache()
        await refreshCacheStats()

        // Emit cache cleared event
        if (process.client) {
          window.dispatchEvent(new CustomEvent('geoip:cache-cleared'))
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to clear GeoIP cache'
      console.error('Failed to clear cache:', err)
    } finally {
      isLoading.value = false
    }
  }

  // Event listeners
  const handleGeoIPUpdated = (event: CustomEvent) => {
    const { ip, data } = event.detail
    console.log(`GeoIP data updated for ${ip}:`, data)

    // Refresh cache stats when data is updated
    refreshCacheStats()

    // Emit custom event for components
    if (process.client) {
      window.dispatchEvent(new CustomEvent('geoip:data-updated', {
        detail: { ip, data }
      }))
    }
  }

  const handleConnectionGeoIPUpdated = (event: CustomEvent) => {
    const { connection, geoip } = event.detail
    console.log('Connection GeoIP updated:', connection, geoip)

    // Emit custom event for components
    if (process.client) {
      window.dispatchEvent(new CustomEvent('geoip:connection-updated', {
        detail: { connection, geoip }
      }))
    }
  }

  // Lifecycle hooks
  onMounted(() => {
    // Check initial status
    if (networkService) {
      networkService.IsGeoIPEnabled().then((enabled: boolean) => {
        isEnabled.value = enabled
      }).catch(() => {
        // Ignore errors during initialization
      })

      refreshCacheStats()
    }

    // Add event listeners
    if (process.client) {
      window.addEventListener('geoip:updated', handleGeoIPUpdated as EventListener)
      window.addEventListener('network:geoip-updated', handleConnectionGeoIPUpdated as EventListener)
    }
  })

  onUnmounted(() => {
    // Remove event listeners
    if (process.client) {
      window.removeEventListener('geoip:updated', handleGeoIPUpdated as EventListener)
      window.removeEventListener('network:geoip-updated', handleConnectionGeoIPUpdated as EventListener)
    }
  })

  // Auto-refresh cache stats periodically
  let statsInterval: NodeJS.Timeout | null = null

  onMounted(() => {
    if (isGeoIPEnabled.value) {
      statsInterval = setInterval(refreshCacheStats, 30000) // Refresh every 30 seconds
    }
  })

  onUnmounted(() => {
    if (statsInterval) {
      clearInterval(statsInterval)
      statsInterval = null
    }
  })

  // Watch for enabled state changes
  watch(isGeoIPEnabled, (enabled) => {
    if (enabled) {
      statsInterval = setInterval(refreshCacheStats, 30000)
    } else {
      if (statsInterval) {
        clearInterval(statsInterval)
        statsInterval = null
      }
    }
  })

  return {
    // State
    isEnabled: readonly(isEnabled),
    isLoading: readonly(isLoading),
    error: readonly(error),
    cacheStats: readonly(cacheStats),
    settings: readonly(settings),

    // Computed
    isGeoIPEnabled,
    hasError,
    cacheHitRate,

    // Methods
    enableGeoIP,
    disableGeoIP,
    getGeoIPData,
    updateSettings,
    refreshCacheStats,
    clearCache
  }
}

// Helper composable for lazy loading GeoIP data
export function useLazyGeoIP() {
  const { getGeoIPData, isGeoIPEnabled } = useGeoIP()
  const geoIPCache = ref<Map<string, GeoIPInfo>>(new Map())
  const pendingRequests = ref<Map<string, Promise<GeoIPInfo | null>>>(new Map())

  const getGeoIPDataLazy = async (ip: string): Promise<GeoIPInfo | null> => {
    // Check cache first
    if (geoIPCache.value.has(ip)) {
      return geoIPCache.value.get(ip)!
    }

    // Check if request is already pending
    if (pendingRequests.value.has(ip)) {
      return pendingRequests.value.get(ip)!
    }

    // Only proceed if GeoIP is enabled
    if (!isGeoIPEnabled.value) {
      return null
    }

    // Create new request
    const request = getGeoIPData(ip).then((data) => {
      if (data) {
        geoIPCache.value.set(ip, data)
      }
      pendingRequests.value.delete(ip)
      return data
    }).catch((err) => {
      console.error(`Failed to get GeoIP data for ${ip}:`, err)
      pendingRequests.value.delete(ip)
      return null
    })

    pendingRequests.value.set(ip, request)
    return request
  }

  const preloadGeoIPData = async (ips: string[]) => {
    if (!isGeoIPEnabled.value) {
      return
    }

    const promises = ips
      .filter(ip => !geoIPCache.value.has(ip) && !pendingRequests.value.has(ip))
      .map(ip => getGeoIPDataLazy(ip))

    await Promise.allSettled(promises)
  }

  const clearCache = () => {
    geoIPCache.value.clear()
    pendingRequests.value.clear()
  }

  return {
    getGeoIPData: getGeoIPDataLazy,
    preloadGeoIPData,
    clearCache,
    cache: readonly(geoIPCache)
  }
}