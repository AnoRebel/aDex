import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWails } from '~/composables/useWails'
import type { SystemInfo, SystemStats, Process, SystemMetrics, CPUMetrics, MemoryMetrics, ProcessMetrics, DiskMetrics, NetworkMetrics, TemperatureMetrics } from '~/types/system'
import { handleBackendError, handleComponentError } from '~/utils/errorHandler'
import { createLoading, completeLoading } from '~/utils/loadingStates'

interface SystemState {
  // Basic system information
  systemInfo: SystemInfo | null
  systemStats: SystemStats | null
  processes: Process[]

  // Detailed metrics for advanced monitoring
  systemMetrics: SystemMetrics | null

  // State management
  isLoading: boolean
  error: string | null
  lastUpdate: Date | null
  isMonitoring: boolean

  // History tracking for charts (keep last 60 data points)
  cpuHistory: number[]
  memoryHistory: number[]
  historyTimestamps: Date[]

  // Auto-refresh
  autoRefresh: boolean
  refreshInterval: number
  refreshTimer: NodeJS.Timeout | null

  // Monitoring configuration
  maxProcesses: number
  enableTemperature: boolean
  enableNetwork: boolean
  enableDisk: boolean

  // Alerts and notifications
  alerts: Array<{
    id: string
    type: 'warning' | 'error' | 'info' | 'critical'
    category: 'cpu' | 'memory' | 'disk' | 'temperature' | 'network' | 'process'
    title: string
    message: string
    source: string
    timestamp: Date
    acknowledged: boolean
    severity: 'low' | 'medium' | 'high' | 'critical'
    value?: number
    threshold?: number
  }>

  // User settings
  settings: {
    showHiddenProcesses: boolean
    sortBy: 'name' | 'cpu' | 'memory' | 'pid'
    sortOrder: 'asc' | 'desc'
    refreshRate: number
    enableNotifications: boolean
    temperatureUnit: 'celsius' | 'fahrenheit'
    enableDetailedInfo: boolean
  }
}

export const useSystemStore = defineStore('system', () => {
  // State
  const systemInfo = ref<SystemInfo | null>(null)
  const systemStats = ref<SystemStats | null>(null)
  const systemMetrics = ref<SystemMetrics | null>(null)
  const processes = ref<Process[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const lastUpdate = ref<Date | null>(null)
  const isMonitoring = ref(false)

  // History for charts
  const cpuHistory = ref<number[]>([])
  const memoryHistory = ref<number[]>([])
  const historyTimestamps = ref<Date[]>([])

  // Auto-refresh configuration
  const autoRefresh = ref(true)
  const refreshInterval = ref(2000)
  const refreshTimer = ref<NodeJS.Timeout | null>(null)

  // Monitoring configuration
  const maxProcesses = ref(100)
  const enableTemperature = ref(true)
  const enableNetwork = ref(true)
  const enableDisk = ref(true)

  // Alerts
  const alerts = ref<Array<{
    id: string
    type: 'warning' | 'error' | 'info' | 'critical'
    category: 'cpu' | 'memory' | 'disk' | 'temperature' | 'network' | 'process'
    title: string
    message: string
    source: string
    timestamp: Date
    acknowledged: boolean
    severity: 'low' | 'medium' | 'high' | 'critical'
    value?: number
    threshold?: number
  }>>([])

  // Settings
  const settings = ref({
    showHiddenProcesses: false,
    sortBy: 'cpu' as 'name' | 'cpu' | 'memory' | 'pid',
    sortOrder: 'desc' as 'asc' | 'desc',
    refreshRate: 2000,
    enableNotifications: true,
    temperatureUnit: 'celsius' as 'celsius' | 'fahrenheit',
    enableDetailedInfo: true
  })

  // Computed properties for basic interfaces (backward compatibility)
  const isSystemInfoLoaded = computed(() => !!systemInfo.value)
  const isStatsLoaded = computed(() => !!systemStats.value)
  const cpuUsage = computed(() => systemStats.value?.cpu || null)
  const memoryUsage = computed(() => systemStats.value?.memory || null)
  const diskUsage = computed(() => systemStats.value?.disk || [])
  const totalProcesses = computed(() => processes.value.length)
  const runningProcesses = computed(() =>
    processes.value.filter(p => p.status === 'running').length
  )

  // Computed properties for detailed interfaces
  const cpuMetrics = computed<CPUMetrics | null>(() => systemMetrics.value?.cpu || null)
  const memoryMetrics = computed<MemoryMetrics | null>(() => systemMetrics.value?.memory || null)
  const processMetrics = computed<ProcessMetrics | null>(() => systemMetrics.value?.processes || null)
  const diskMetrics = computed<DiskMetrics | null>(() => systemMetrics.value?.disks || null)
  const networkMetrics = computed<NetworkMetrics | null>(() => systemMetrics.value?.network || null)
  const temperatureMetrics = computed<TemperatureMetrics | null>(() => systemMetrics.value?.temperature || null)

  // System health assessment
  const systemHealth = computed(() => {
    const cpuPercent = cpuMetrics.value?.usagePercent || cpuUsage.value?.usage || 0
    const memoryPercent = memoryMetrics.value?.usagePercent || memoryUsage.value?.usage || 0
    const swapPercent = memoryMetrics.value?.swapPercent || 0

    const health = {
      status: 'good' as 'unknown' | 'good' | 'warning' | 'critical',
      issues: [] as string[],
      score: 100,
      recommendations: [] as string[]
    }

    // CPU health check
    if (cpuPercent > 90) {
      health.status = 'critical'
      health.issues.push('Critical CPU usage')
      health.recommendations.push('Consider closing CPU-intensive applications')
      health.score -= 30
    } else if (cpuPercent > 75) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High CPU usage')
      health.recommendations.push('Monitor CPU usage closely')
      health.score -= 15
    }

    // Memory health check
    if (memoryPercent > 90) {
      health.status = 'critical'
      health.issues.push('Critical memory usage')
      health.recommendations.push('Close memory-intensive applications immediately')
      health.score -= 30
    } else if (memoryPercent > 75) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High memory usage')
      health.recommendations.push('Monitor memory usage')
      health.score -= 15
    }

    // Swap health check
    if (swapPercent > 50) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High swap usage')
      health.recommendations.push('Consider adding more RAM')
      health.score -= 10
    }

    // Disk health check
    if (diskMetrics.value?.disks?.some(disk => disk.usagePercent > 95) ||
        diskUsage.value?.some(disk => disk.usage > 95)) {
      health.status = 'critical'
      health.issues.push('Critical disk space')
      health.recommendations.push('Free up disk space immediately')
      health.score -= 25
    } else if (diskMetrics.value?.disks?.some(disk => disk.usagePercent > 85) ||
               diskUsage.value?.some(disk => disk.usage > 85)) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('Low disk space')
      health.recommendations.push('Monitor disk usage')
      health.score -= 10
    }

    // Temperature health check
    if (temperatureMetrics.value?.sensors?.some(sensor => sensor.temperature >= sensor.critical)) {
      health.status = 'critical'
      health.issues.push('Critical temperature')
      health.recommendations.push('Check system cooling')
      health.score -= 25
    } else if (temperatureMetrics.value?.sensors?.some(sensor => sensor.temperature >= sensor.high)) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High temperature')
      health.recommendations.push('Monitor system temperature')
      health.score -= 10
    }

    return health
  })

  // System uptime
  const systemUptime = computed(() => {
    if (!systemInfo.value?.uptime) return null
    const uptime = systemInfo.value.uptime
    const days = Math.floor(uptime / 86400)
    const hours = Math.floor((uptime % 86400) / 3600)
    const minutes = Math.floor((uptime % 3600) / 60)
    return { days, hours, minutes }
  })

  // Performance status
  const performanceStatus = computed(() => {
    const cpuPercent = cpuMetrics.value?.usagePercent || cpuUsage.value?.usage || 0
    const memoryPercent = memoryMetrics.value?.usagePercent || memoryUsage.value?.usage || 0

    if (cpuPercent > 80 || memoryPercent > 90) return 'critical'
    if (cpuPercent > 60 || memoryPercent > 75) return 'warning'
    if (cpuPercent > 40 || memoryPercent > 50) return 'moderate'
    return 'good'
  })

  // Sorted processes
  const sortedProcesses = computed(() => {
    let processList = [...processes.value]

    if (!settings.value.showHiddenProcesses) {
      processList = processList.filter(p => !p.name.startsWith('.'))
    }

    return processList.sort((a, b) => {
      let comparison = 0

      switch (settings.value.sortBy) {
        case 'name':
          comparison = a.name.localeCompare(b.name)
          break
        case 'cpu':
          comparison = b.cpu - a.cpu
          break
        case 'memory':
          comparison = b.memory - a.memory
          break
        case 'pid':
          comparison = a.pid - b.pid
          break
      }

      return settings.value.sortOrder === 'asc' ? comparison : -comparison
    })
  })

  // Alert getters
  const unacknowledgedAlerts = computed(() =>
    alerts.value.filter(alert => !alert.acknowledged)
  )

  const criticalAlerts = computed(() =>
    alerts.value.filter(alert =>
      alert.type === 'error' && !alert.acknowledged
    )
  )

  // Actions
  const setSystemInfo = (info: SystemInfo) => {
    systemInfo.value = info
  }

  const setSystemStats = (stats: SystemStats) => {
    systemStats.value = stats
    lastUpdate.value = new Date()

    // Update history for charts
    if (stats.cpu && stats.memory) {
      updateHistory(stats.cpu.usage, stats.memory.usage)
    }
  }

  const setSystemMetrics = (metrics: SystemMetrics) => {
    systemMetrics.value = metrics
    lastUpdate.value = new Date()

    // Update history for charts
    if (metrics.cpu && metrics.memory) {
      updateHistory(metrics.cpu.usagePercent, metrics.memory.usagePercent)
    }

    // Check for alerts
    checkThresholds(metrics)
  }

  const setProcesses = (processList: Process[]) => {
    processes.value = processList.slice(0, maxProcesses.value)
  }

  const setLoading = (loading: boolean) => {
    isLoading.value = loading
  }

  const setError = (errorMessage: string | null) => {
    error.value = errorMessage
  }

  const updateHistory = (cpuUsage: number, memoryUsage: number) => {
    const now = new Date()

    // Add new data points
    cpuHistory.value.push(cpuUsage)
    memoryHistory.value.push(memoryUsage)
    historyTimestamps.value.push(now)

    // Keep only last 60 data points
    if (cpuHistory.value.length > 60) {
      cpuHistory.value.shift()
      memoryHistory.value.shift()
      historyTimestamps.value.shift()
    }
  }

  const checkThresholds = (metrics: SystemMetrics) => {
    const newAlerts = []

    // CPU threshold check
    if (metrics.cpu.usagePercent > 90) {
      newAlerts.push({
        id: `cpu-${Date.now()}`,
        type: 'critical',
        category: 'cpu',
        title: 'Critical CPU Usage',
        message: `CPU usage is ${metrics.cpu.usagePercent.toFixed(1)}%`,
        source: 'system-monitor',
        timestamp: new Date(),
        acknowledged: false,
        severity: 'critical',
        value: metrics.cpu.usagePercent,
        threshold: 90
      })
    } else if (metrics.cpu.usagePercent > 75) {
      newAlerts.push({
        id: `cpu-${Date.now()}`,
        type: 'warning',
        category: 'cpu',
        title: 'High CPU Usage',
        message: `CPU usage is ${metrics.cpu.usagePercent.toFixed(1)}%`,
        source: 'system-monitor',
        timestamp: new Date(),
        acknowledged: false,
        severity: 'high',
        value: metrics.cpu.usagePercent,
        threshold: 75
      })
    }

    // Memory threshold check
    if (metrics.memory.usagePercent > 90) {
      newAlerts.push({
        id: `memory-${Date.now()}`,
        type: 'critical',
        category: 'memory',
        title: 'Critical Memory Usage',
        message: `Memory usage is ${metrics.memory.usagePercent.toFixed(1)}%`,
        source: 'system-monitor',
        timestamp: new Date(),
        acknowledged: false,
        severity: 'critical',
        value: metrics.memory.usagePercent,
        threshold: 90
      })
    }

    // Disk threshold check
    for (const disk of metrics.disks.disks) {
      if (disk.usagePercent > 95) {
        newAlerts.push({
          id: `disk-${disk.device}-${Date.now()}`,
          type: 'critical',
          category: 'disk',
          title: 'Critical Disk Space',
          message: `Disk ${disk.device} is ${disk.usagePercent.toFixed(1)}% full`,
          source: 'system-monitor',
          timestamp: new Date(),
          acknowledged: false,
          severity: 'critical',
          value: disk.usagePercent,
          threshold: 95
        })
      }
    }

    // Temperature threshold check
    if (metrics.temperature?.sensors) {
      for (const sensor of metrics.temperature.sensors) {
        if (sensor.temperature >= sensor.critical) {
          newAlerts.push({
            id: `temp-${sensor.name}-${Date.now()}`,
            type: 'critical',
            category: 'temperature',
            title: 'Critical Temperature',
            message: `${sensor.name} temperature is ${sensor.temperature}°${sensor.unit}`,
            source: 'system-monitor',
            timestamp: new Date(),
            acknowledged: false,
            severity: 'critical',
            value: sensor.temperature,
            threshold: sensor.critical
          })
        }
      }
    }

    alerts.value.unshift(...newAlerts)

    // Keep only last 50 alerts
    if (alerts.value.length > 50) {
      alerts.value = alerts.value.slice(0, 50)
    }
  }

  const startMonitoring = () => {
    isMonitoring.value = true
    if (autoRefresh.value && refreshInterval.value > 0) {
      refreshTimer.value = setInterval(() => {
        // Refresh logic would be implemented here
      }, refreshInterval.value)
    }
  }

  const stopMonitoring = () => {
    isMonitoring.value = false
    if (refreshTimer.value) {
      clearInterval(refreshTimer.value)
      refreshTimer.value = null
    }
  }

  const updateSettings = (newSettings: Partial<typeof settings.value>) => {
    settings.value = { ...settings.value, ...newSettings }
    refreshInterval.value = settings.value.refreshRate
  }

  const acknowledgeAlert = (alertId: string) => {
    const alert = alerts.value.find(a => a.id === alertId)
    if (alert) {
      alert.acknowledged = true
    }
  }

  const clearAlerts = () => {
    alerts.value = []
  }

  const clearHistory = () => {
    cpuHistory.value = []
    memoryHistory.value = []
    historyTimestamps.value = []
  }

  const reset = () => {
    systemInfo.value = null
    systemStats.value = null
    systemMetrics.value = null
    processes.value = []
    isLoading.value = false
    error.value = null
    lastUpdate.value = null
    stopMonitoring()
    clearHistory()
    clearAlerts()
  }

  // Fetch methods using Wails bindings
  const fetchSystemInfo = async (): Promise<void> => {
    try {
      setLoading(true)
      setError(null)

      const { system } = useWails()
      const info = await system.getSystemInfo()

      if (info) {
        setSystemInfo(info as SystemInfo)
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch system info'
      setError(errorMessage)
      console.error('Failed to fetch system info:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchSystemStats = async (): Promise<void> => {
    try {
      const { system } = useWails()

      // Fetch CPU and memory usage
      const [cpuData, memoryData, diskData] = await Promise.all([
        system.getCPUUsage(),
        system.getMemoryUsage(),
        system.getDiskUsage()
      ])

      if (cpuData || memoryData || diskData) {
        const stats: SystemStats = {
          cpu: cpuData ? { usage: cpuData.usage || 0, cores: cpuData.cores || [] } : { usage: 0, cores: [] },
          memory: memoryData ? {
            total: memoryData.total || 0,
            used: memoryData.used || 0,
            free: memoryData.free || 0,
            usage: memoryData.usage || 0
          } : { total: 0, used: 0, free: 0, usage: 0 },
          disk: diskData || []
        }
        setSystemStats(stats)
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch system stats'
      setError(errorMessage)
      console.error('Failed to fetch system stats:', err)
    }
  }

  const startSystemMonitoring = async (interval?: number): Promise<void> => {
    try {
      const { system } = useWails()
      await system.startMonitoring(interval || refreshInterval.value)
      startMonitoring()
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to start monitoring'
      setError(errorMessage)
      console.error('Failed to start system monitoring:', err)
    }
  }

  const stopSystemMonitoring = async (): Promise<void> => {
    try {
      const { system } = useWails()
      await system.stopMonitoring()
      stopMonitoring()
    } catch (err) {
      console.error('Failed to stop system monitoring:', err)
    }
  }

  // Initialize store with data from backend
  const initialize = async (): Promise<void> => {
    await fetchSystemInfo()
    await fetchSystemStats()
  }

  return {
    // State
    systemInfo: systemInfo,
    systemStats: systemStats,
    systemMetrics: systemMetrics,
    processes: processes,
    isLoading: isLoading,
    error: error,
    lastUpdate: lastUpdate,
    isMonitoring: isMonitoring,
    cpuHistory: cpuHistory,
    memoryHistory: memoryHistory,
    historyTimestamps: historyTimestamps,
    autoRefresh: autoRefresh,
    refreshInterval: refreshInterval,
    maxProcesses: maxProcesses,
    enableTemperature: enableTemperature,
    enableNetwork: enableNetwork,
    enableDisk: enableDisk,
    alerts: alerts,
    settings: settings,

    // Computed properties
    isSystemInfoLoaded: isSystemInfoLoaded,
    isStatsLoaded: isStatsLoaded,
    cpuUsage: cpuUsage,
    memoryUsage: memoryUsage,
    diskUsage: diskUsage,
    totalProcesses: totalProcesses,
    runningProcesses: runningProcesses,
    cpuMetrics: cpuMetrics,
    memoryMetrics: memoryMetrics,
    processMetrics: processMetrics,
    diskMetrics: diskMetrics,
    networkMetrics: networkMetrics,
    temperatureMetrics: temperatureMetrics,
    systemHealth: systemHealth,
    systemUptime: systemUptime,
    performanceStatus: performanceStatus,
    sortedProcesses: sortedProcesses,
    unacknowledgedAlerts: unacknowledgedAlerts,
    criticalAlerts: criticalAlerts,

    // Actions
    setSystemInfo: setSystemInfo,
    setSystemStats: setSystemStats,
    setSystemMetrics: setSystemMetrics,
    setProcesses: setProcesses,
    setLoading: setLoading,
    setError: setError,
    startMonitoring: startMonitoring,
    stopMonitoring: stopMonitoring,
    updateSettings: updateSettings,
    acknowledgeAlert: acknowledgeAlert,
    clearAlerts: clearAlerts,
    clearHistory: clearHistory,
    reset: reset,

    // Wails integration methods
    fetchSystemInfo: fetchSystemInfo,
    fetchSystemStats: fetchSystemStats,
    startSystemMonitoring: startSystemMonitoring,
    stopSystemMonitoring: stopSystemMonitoring,
    initialize: initialize
  }
})