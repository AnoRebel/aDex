import { ref, computed, onMounted, onUnmounted } from 'vue'

// Wails v3 API import
import { ServiceCoordinator } from '~/wailsjs/go/backend/services/coordinator/ServiceCoordinator'

export function useNetwork() {
  // Reactive state
  const metrics = ref<any>(null)
  const connections = ref<any[]>([])
  const alerts = ref<any[]>([])
  const statistics = ref<any>(null)
  const config = ref<any>(null)
  const isMonitoring = ref(false)
  const lastUpdateTime = ref<Date | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  const serviceCoordinator = new ServiceCoordinator()
  
  let refreshTimer: NodeJS.Timeout | null = null
  const refreshInterval = ref(2000) // 2 seconds

  // Computed properties
  const activeInterfaces = computed(() => {
    if (!metrics.value?.interfaces) return []
    return metrics.value.interfaces.filter((iface: any) => iface.isUp)
  })

  const activeConnections = computed(() => {
    return connections.value.filter(conn => conn.state === 'ESTABLISHED')
  })

  const activeAlerts = computed(() => {
    return alerts.value.filter(alert => !alert.resolved)
  })

  const totalBandwidth = computed(() => {
    if (!metrics.value) return 0
    return (metrics.value.totalBytesSent || 0) + (metrics.value.totalBytesRecv || 0)
  })

  const criticalAlerts = computed(() => {
    return alerts.value.filter(alert => !alert.resolved && alert.severity === 'error')
  })

  // Methods
  const fetchMetrics = async () => {
    try {
      const data = await serviceCoordinator.GetNetworkMetrics()
      metrics.value = data
      lastUpdateTime.value = new Date()
    } catch (err) {
      error.value = `Failed to fetch network metrics: ${err}`
      throw err
    }
  }

  const fetchConnections = async () => {
    try {
      const data = await serviceCoordinator.GetNetworkConnections()
      connections.value = data || []
      lastUpdateTime.value = new Date()
    } catch (err) {
      error.value = `Failed to fetch network connections: ${err}`
      throw err
    }
  }

  const fetchAlerts = async () => {
    try {
      const data = await serviceCoordinator.GetNetworkAlerts()
      alerts.value = data || []
      lastUpdateTime.value = new Date()
    } catch (err) {
      error.value = `Failed to fetch network alerts: ${err}`
      throw err
    }
  }

  const fetchStatistics = async () => {
    try {
      const data = await serviceCoordinator.GetNetworkStatistics()
      statistics.value = data
      lastUpdateTime.value = new Date()
    } catch (err) {
      error.value = `Failed to fetch network statistics: ${err}`
      throw err
    }
  }

  const fetchConfig = async () => {
    try {
      const data = await serviceCoordinator.GetNetworkConfig()
      config.value = data
    } catch (err) {
      error.value = `Failed to fetch network config: ${err}`
      throw err
    }
  }

  const fetchBandwidthData = async (interfaceName: string) => {
    try {
      const data = await serviceCoordinator.GetBandwidthData(interfaceName)
      return data
    } catch (err) {
      error.value = `Failed to fetch bandwidth data: ${err}`
      throw err
    }
  }

  const startMonitoring = async () => {
    try {
      await serviceCoordinator.StartNetworkMonitoring()
      isMonitoring.value = true
      startAutoRefresh()
    } catch (err) {
      error.value = `Failed to start monitoring: ${err}`
      throw err
    }
  }

  const stopMonitoring = async () => {
    try {
      await serviceCoordinator.StopNetworkMonitoring()
      isMonitoring.value = false
      stopAutoRefresh()
    } catch (err) {
      error.value = `Failed to stop monitoring: ${err}`
      throw err
    }
  }

  const resolveAlert = async (alertId: string) => {
    try {
      await serviceCoordinator.ResolveNetworkAlert(alertId)
      // Update local alert state
      const alert = alerts.value.find(a => a.id === alertId)
      if (alert) {
        alert.resolved = true
        alert.resolvedAt = new Date()
      }
    } catch (err) {
      error.value = `Failed to resolve alert: ${err}`
      throw err
    }
  }

  const clearAlerts = async () => {
    try {
      await serviceCoordinator.ClearNetworkAlerts()
      // Remove resolved alerts from local state
      alerts.value = alerts.value.filter(alert => !alert.resolved)
    } catch (err) {
      error.value = `Failed to clear alerts: ${err}`
      throw err
    }
  }

  const updateConfig = async (newConfig: any) => {
    try {
      await serviceCoordinator.UpdateNetworkConfig(newConfig)
      config.value = newConfig
    } catch (err) {
      error.value = `Failed to update config: ${err}`
      throw err
    }
  }

  const resetService = async () => {
    try {
      await serviceCoordinator.ResetNetworkService()
      // Reset local state
      metrics.value = null
      connections.value = []
      alerts.value = []
      statistics.value = null
      error.value = null
    } catch (err) {
      error.value = `Failed to reset service: ${err}`
      throw err
    }
  }

  const refreshAllData = async () => {
    loading.value = true
    error.value = null
    
    try {
      await Promise.all([
        fetchMetrics(),
        fetchConnections(),
        fetchAlerts(),
        fetchStatistics(),
        fetchConfig()
      ])
    } catch (err) {
      error.value = `Failed to refresh data: ${err}`
    } finally {
      loading.value = false
    }
  }

  const clearError = () => {
    error.value = null
  }

  // Auto-refresh functionality
  const startAutoRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
    }
    
    if (isMonitoring.value) {
      refreshTimer = setInterval(refreshAllData, refreshInterval.value)
    }
  }

  const stopAutoRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }

  const updateMonitoringStatus = async () => {
    try {
      isMonitoring.value = await serviceCoordinator.IsNetworkMonitoring()
    } catch (err) {
      error.value = `Failed to check monitoring status: ${err}`
    }
  }

  // Lifecycle
  onMounted(async () => {
    await refreshAllData()
    await updateMonitoringStatus()
    if (isMonitoring.value) {
      startAutoRefresh()
    }
  })

  onUnmounted(() => {
    stopAutoRefresh()
  })

  return {
    // State
    metrics,
    connections,
    alerts,
    statistics,
    config,
    isMonitoring,
    lastUpdateTime,
    loading,
    error,
    
    // Computed
    activeInterfaces,
    activeConnections,
    activeAlerts,
    totalBandwidth,
    criticalAlerts,
    
    // Methods
    fetchMetrics,
    fetchConnections,
    fetchAlerts,
    fetchStatistics,
    fetchConfig,
    fetchBandwidthData,
    startMonitoring,
    stopMonitoring,
    resolveAlert,
    clearAlerts,
    updateConfig,
    resetService,
    refreshAllData,
    clearError,
    updateMonitoringStatus
  }
}