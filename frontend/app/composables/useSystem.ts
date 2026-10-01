import { ref, computed, watch, onMounted, onUnmounted, readonly } from 'vue'
import { useDebounceFn, useThrottleFn, useIntervalFn } from '@vueuse/core'
// import { useToast } from '@nuxt/ui'
import { useWails } from './useWails'
// import { ServiceCoordinator } from '~/wailsjs/go/backend/services/coordinator/ServiceCoordinator'
import type { SystemInfo, SystemStats, Process, SystemMetrics, CPUMetrics, MemoryMetrics, ProcessMetrics, DiskMetrics, NetworkMetrics, TemperatureMetrics } from '~/types/system'
// import type { SystemInfo as WailsSystemInfo, MemoryInfo, DiskInfo, NetworkInfo } from '../../bindings/aDex/internal/services/system/models'
import { handleBackendError, handleComponentError } from '~/utils/errorHandler'
import { createLoading, completeLoading } from '~/utils/loadingStates'

// System composable for system monitoring functionality
export function useSystem(options: {
  autoRefresh?: boolean
  refreshInterval?: number
  maxProcesses?: number
  enableTemperature?: boolean
  enableNetwork?: boolean
  enableDisk?: boolean
} = {}) {
  // Default options
  const {
    autoRefresh = true,
    refreshInterval: initialRefreshInterval = 2000,
    maxProcesses = 100,
    enableTemperature = true,
    enableNetwork = true,
    enableDisk = true
  } = options

  const refreshInterval = ref(initialRefreshInterval)

  // Wails integration
  const wails = useWails()
  // const { toast } = useToast()
  // const serviceCoordinator = new ServiceCoordinator()

  // Reactive state
  const systemMetrics = ref<SystemMetrics | null>(null)
  const systemInfo = ref<SystemInfo | null>(null)
  const systemStats = ref<SystemStats | null>(null)
  const processes = ref<Process[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const lastUpdated = ref<Date | null>(null)
  const isMonitoring = ref(false)
  const autoRefreshInterval = ref<NodeJS.Timeout | null>(null)

  // Individual metrics
  const cpuMetrics = computed<CPUMetrics | null>(() => systemMetrics.value?.cpu || null)
  const memoryMetrics = computed<MemoryMetrics | null>(() => systemMetrics.value?.memory || null)
  const processMetrics = computed<ProcessMetrics | null>(() => systemMetrics.value?.processes || null)
  const diskMetrics = computed<DiskMetrics | null>(() => systemMetrics.value?.disks || null)
  const networkMetrics = computed<NetworkMetrics | null>(() => systemMetrics.value?.network || null)
  const temperatureMetrics = computed<TemperatureMetrics | null>(() => systemMetrics.value?.temperature || null)

  // Backward compatibility computed properties
  const isSystemInfoLoaded = computed(() => !!systemInfo.value)
  const isStatsLoaded = computed(() => !!systemStats.value)
  const cpuUsage = computed(() => systemStats.value?.cpu?.usage || cpuMetrics.value?.usagePercent || 0)
  const memoryUsage = computed(() => systemStats.value?.memory?.usage || memoryMetrics.value?.usagePercent || 0)
  const diskUsage = computed(() => systemStats.value?.disk || diskMetrics.value?.disks || [])
  const totalProcesses = computed(() => processes.value.length || processMetrics.value?.totalProcesses || 0)
  const runningProcesses = computed(() =>
    processes.value.filter(p => p.status === 'running').length || processMetrics.value?.runningProcesses || 0
  )

  const systemUptime = computed(() => {
    const uptime = systemInfo.value?.uptime || 0
    const days = Math.floor(uptime / 86400)
    const hours = Math.floor((uptime % 86400) / 3600)
    const minutes = Math.floor((uptime % 3600) / 60)
    return { days, hours, minutes }
  })

  const performanceStatus = computed(() => {
    const cpu = cpuUsage.value
    const memory = memoryUsage.value

    if (cpu > 80 || memory > 90) return 'critical'
    if (cpu > 60 || memory > 75) return 'warning'
    if (cpu > 40 || memory > 50) return 'moderate'
    return 'good'
  })

  // Computed properties
  const systemHealth = computed(() => {
    if (!systemMetrics.value) return null

    const health = {
      status: 'good' as 'good' | 'warning' | 'critical',
      issues: [] as string[],
      score: 100
    }

    // CPU health
    if ((cpuMetrics.value?.usagePercent || 0) > 90) {
      health.status = 'critical'
      health.issues.push('Critical CPU usage')
      health.score -= 30
    } else if ((cpuMetrics.value?.usagePercent || 0) > 75) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High CPU usage')
      health.score -= 15
    }

    // Memory health
    if ((memoryMetrics.value?.usagePercent || 0) > 90) {
      health.status = 'critical'
      health.issues.push('Critical memory usage')
      health.score -= 30
    } else if ((memoryMetrics.value?.usagePercent || 0) > 75) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High memory usage')
      health.score -= 15
    }

    // Swap health
    if ((memoryMetrics.value?.swapPercent || 0) > 50) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High swap usage')
      health.score -= 10
    }

    // Disk health
    if ((diskMetrics.value?.disks || []).some(disk => disk.usagePercent > 90)) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('Low disk space')
      health.score -= 10
    }

    // Temperature health
    if ((temperatureMetrics.value?.sensors || []).some(sensor => sensor.temperature >= sensor.critical)) {
      health.status = 'critical'
      health.issues.push('Critical temperature')
      health.score -= 25
    } else if ((temperatureMetrics.value?.sensors || []).some(sensor => sensor.temperature >= sensor.high)) {
      if (health.status !== 'critical') health.status = 'warning'
      health.issues.push('High temperature')
      health.score -= 10
    }

    return health
  })

  // Convert Wails SystemInfo to our SystemInfo type
  const convertSystemInfo = (wailsInfo: WailsSystemInfo): SystemInfo => {
    return {
      hostname: wailsInfo.Hostname || 'Unknown',
      platform: wailsInfo.OS || 'Unknown',
      architecture: wailsInfo.Architecture || 'Unknown',
      osVersion: 'Unknown',
      kernelVersion: 'Unknown',
      uptime: wailsInfo.Uptime ? Math.floor(wailsInfo.Uptime / 1000000000) : 0,
      bootTime: new Date(),
      timezone: 'UTC',
      cpuCount: 0,
      totalMemory: 0,
      hostId: '',
      username: 'Unknown',
      homeDir: '',
    }
  }

  // Convert SystemStats to SystemMetrics for compatibility
  const systemStatsToMetrics = (stats: SystemStats): SystemMetrics => {
    return {
      cpu: {
        usagePercent: stats.cpu.usage,
        cores: stats.cpu.cores.length,
        model: '',
        vendor: '',
        frequency: 0,
        frequencyMax: 0,
        perCoreUsage: stats.cpu.cores.map(core => core.usage),
        loadAverage: stats.cpu.loadAverage,
        timestamp: stats.timestamp.toISOString()
      },
      memory: {
        total: stats.memory.total,
        available: stats.memory.available,
        used: stats.memory.used,
        free: stats.memory.available,
        usagePercent: stats.memory.usage,
        swapTotal: stats.memory.swapTotal,
        swapUsed: stats.memory.swapUsed,
        swapFree: stats.memory.swapFree,
        swapPercent: stats.memory.swapTotal > 0 ? (stats.memory.swapUsed / stats.memory.swapTotal) * 100 : 0,
        timestamp: stats.timestamp.toISOString()
      },
      processes: {
        processes: [],
        totalProcesses: 0,
        runningProcesses: 0,
        sleepingProcesses: 0,
        stoppedProcesses: 0,
        zombieProcesses: 0,
        timestamp: stats.timestamp.toISOString()
      },
      disks: {
        disks: stats.disk.map(disk => ({
          device: disk.device,
          mountpoint: disk.mountpoint,
          fsType: disk.fstype,
          total: disk.total,
          used: disk.used,
          free: disk.free,
          usagePercent: disk.usage,
          readOnly: false
        })),
        totalSpace: stats.disk.reduce((sum, disk) => sum + disk.total, 0),
        totalUsed: stats.disk.reduce((sum, disk) => sum + disk.used, 0),
        totalFree: stats.disk.reduce((sum, disk) => sum + disk.free, 0),
        timestamp: stats.timestamp.toISOString()
      },
      network: {
        interfaces: stats.network.interfaces.map(iface => ({
          name: iface.name,
          isUp: iface.isUp,
          bytesSent: iface.bytesSent,
          bytesRecv: iface.bytesReceived,
          packetsSent: iface.packetsSent,
          packetsRecv: iface.packetsReceived,
          errin: 0,
          errout: 0,
          dropin: 0,
          dropout: 0,
          ipAddresses: iface.addresses.map(addr => addr.address),
          mac: iface.mac,
          speed: iface.speed,
          mtu: iface.mtu
        })),
        totalBytesSent: stats.network.bytesSent,
        totalBytesRecv: stats.network.bytesReceived,
        timestamp: stats.timestamp.toISOString()
      },
      temperature: null,
      timestamp: stats.timestamp.toISOString()
    }
  }

  // Format functions
  const formatBytes = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let size = bytes
    let unitIndex = 0

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`
  }

  const formatFrequency = (mhz: number): string => {
    if (mhz < 1000) {
      return `${mhz.toFixed(0)} MHz`
    }
    return `${(mhz / 1000).toFixed(2)} GHz`
  }

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)

    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`
    } else {
      return `${minutes}m`
    }
  }

  // Wails API functions
  const loadSystemInfo = async () => {
    if (!wails.isReady.value) {
      console.warn('Wails not ready for system info')
      return
    }

    try {
      isLoading.value = true
      const info = await wails.system.getSystemInfo()

      if (info) {
        systemInfo.value = convertSystemInfo(info)
      }
    } catch (error) {
      console.error('Failed to load system info:', error)
    } finally {
      isLoading.value = false
    }
  }

  const loadSystemStats = async () => {
    if (!wails.isReady.value) {
      console.warn('Wails not ready for system stats')
      return
    }

    try {
      const [cpuUsageData, memoryUsageData, diskUsageData, networkUsageData] = await Promise.all([
        wails.system.getCPUUsage(),
        wails.system.getMemoryUsage(),
        wails.system.getDiskUsage(),
        wails.system.getNetworkInfo()
      ])

      systemStats.value = {
        cpu: {
          usage: cpuUsageData || 0,
          cores: [],
          temperature: 0,
          loadAverage: [0, 0, 0],
        },
        memory: memoryUsageData ? {
          total: memoryUsageData.Total || 0,
          available: memoryUsageData.Available || 0,
          used: memoryUsageData.Used || 0,
          usage: memoryUsageData.Percent || 0,
          swapTotal: 0,
          swapUsed: 0,
          swapFree: 0,
        } : {
          total: 0,
          available: 0,
          used: 0,
          usage: 0,
          swapTotal: 0,
          swapUsed: 0,
          swapFree: 0,
        },
        disk: diskUsageData ? diskUsageData.map((disk: DiskInfo) => ({
          device: disk.Mountpoint || '',
          total: disk.Total || 0,
          used: disk.Used || 0,
          free: disk.Free || 0,
          usage: disk.Percent || 0,
          filesystem: '',
        })) : [],
        network: networkUsageData ? {
          interfaces: networkUsageData.Interfaces.map((iface: any) => ({
            name: iface.Name || '',
            ipAddress: iface.IPAddress || '',
            isUp: iface.IsUp || false,
            bytesReceived: 0,
            bytesSent: 0,
            packetsReceived: 0,
            packetsSent: 0,
          })),
          bytesReceived: 0,
          bytesSent: 0,
        } : {
          interfaces: [],
          bytesReceived: 0,
          bytesSent: 0,
        },
        timestamp: new Date(),
      }

      lastUpdated.value = new Date()
    } catch (error) {
      console.error('Failed to load system stats:', error)
    }
  }

  // API calls with fallback support
  const fetchSystemMetrics = async (): Promise<SystemMetrics> => {
    try {
      // Try ServiceCoordinator first (new API)
      const metrics = await serviceCoordinator.GetSystemMetrics()
      return metrics
    } catch (serviceError) {
      console.warn('ServiceCoordinator failed, falling back to Wails API:', serviceError)

      // Fallback to Wails API and convert
      await loadSystemStats()
      if (systemStats.value) {
        return systemStatsToMetrics(systemStats.value)
      }

      throw new Error('Both ServiceCoordinator and Wails API failed')
    }
  }

  // Additional Wails-specific functions

  // Process management functions
  const executeCommand = async (command: string, args: string[] = []) => {
    if (!wails.isReady.value) {
      return { success: false, output: '', error: 'Wails not ready' }
    }

    try {
      isLoading.value = true
      // Command execution would need to be implemented in the backend
      return { success: false, output: '', error: 'Command execution not available' }
    } catch (error) {
      return {
        success: false,
        output: '',
        error: error instanceof Error ? error.message : 'Unknown error',
      }
    } finally {
      isLoading.value = false
    }
  }

  const killProcess = async (pid: number, signal: number = 15) => {
    if (!wails.isReady.value) {
      return false
    }

    try {
      // Process management would need to be implemented in the backend
      return false
    } catch (error) {
      console.error('Failed to kill process:', error)
      return false
    }
  }

  // Auto-refresh management
  const startAutoRefresh = (interval?: number) => {
    if (autoRefreshInterval.value) {
      clearInterval(autoRefreshInterval.value)
    }

    const refreshIntervalMs = interval || refreshInterval.value
    autoRefreshInterval.value = setInterval(async () => {
      await Promise.all([
        loadSystemStats(),
        refreshMetrics()
      ])
    }, refreshIntervalMs)
  }

  const stopAutoRefresh = () => {
    if (autoRefreshInterval.value) {
      clearInterval(autoRefreshInterval.value)
      autoRefreshInterval.value = null
    }
  }

  const setRefreshRate = (rate: number) => {
    refreshInterval.value = rate
    if (autoRefreshInterval.value) {
      startAutoRefresh(rate)
    }
  }

  // Process search and filtering
  const searchProcesses = (query: string) => {
    if (!query.trim()) return processes.value

    const lowerQuery = query.toLowerCase()
    return processes.value.filter(process =>
      process.name.toLowerCase().includes(lowerQuery) ||
      process.command.toLowerCase().includes(lowerQuery) ||
      process.user.toLowerCase().includes(lowerQuery) ||
      process.pid.toString().includes(query)
    )
  }

  const getProcessByPid = (pid: number) => {
    return processes.value.find(p => p.pid === pid)
  }

  // Event listeners setup
  const setupEventListeners = () => {
    wails.on('system.stats.update', (data) => {
      if (data) {
        systemStats.value = { ...systemStats.value, ...data, timestamp: new Date() }
        lastUpdated.value = new Date()
      }
    })

    wails.on('system.process.added', (process: Process) => {
      processes.value.push(process)
    })

    wails.on('system.process.removed', (pid: number) => {
      processes.value = processes.value.filter(p => p.pid !== pid)
    })

    wails.on('system.process.updated', (process: Process) => {
      const index = processes.value.findIndex(p => p.pid === process.pid)
      if (index !== -1) {
        processes.value[index] = process
      }
    })
  }

  // Watch for Wails availability
  watch(() => wails.isReady.value, (isReady) => {
    if (isReady) {
      loadSystemInfo()
      setupEventListeners()
      if (autoRefresh) {
        startAutoRefresh()
      }
    } else {
      stopAutoRefresh()
    }
  })

  // Core functions
  const refreshMetrics = async () => {
    if (isLoading.value) return

    try {
      isLoading.value = true
      error.value = null

      const metrics = await fetchSystemMetrics()
      systemMetrics.value = metrics
      lastUpdated.value = new Date()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch system metrics'
      console.error('Error fetching system metrics:', err)
    } finally {
      isLoading.value = false
    }
  }

  const debouncedRefresh = useDebounceFn(refreshMetrics, 100)
  const throttledRefresh = useThrottleFn(refreshMetrics, 500)

  // Monitoring control
  const startMonitoring = () => {
    if (isMonitoring.value) return

    isMonitoring.value = true

    if (autoRefresh) {
      // Initial refresh
      refreshMetrics()
    }
  }

  const stopMonitoring = () => {
    isMonitoring.value = false
  }

  // Auto-refresh interval
  let intervalPause: (() => void) | null = null

  if (autoRefresh) {
    intervalPause = useIntervalFn(() => {
      if (isMonitoring.value) {
        refreshMetrics()
      }
    }, refreshInterval.value)
  }

  // Watch for configuration changes
  watch([refreshInterval, () => maxProcesses, () => enableTemperature, () => enableNetwork, () => enableDisk], () => {
    if (intervalPause) {
      intervalPause()
    }

    if (autoRefresh) {
      intervalPause = useIntervalFn(() => {
        if (isMonitoring.value) {
          refreshMetrics()
        }
      }, refreshInterval.value)
    }
  })

  // Lifecycle hooks
  onMounted(() => {
    if (autoRefresh) {
      startMonitoring()
    }
  })

  onUnmounted(() => {
    stopMonitoring()
    if (intervalPause) {
      intervalPause()
    }
  })

  // Top processes helpers
  const getTopCPUProcesses = (limit: number = 5) => {
    if (!processMetrics.value) return []

    return [...processMetrics.value.processes]
      .sort((a, b) => b.cpuPercent - a.cpuPercent)
      .slice(0, limit)
  }

  const getTopMemoryProcesses = (limit: number = 5) => {
    if (!processMetrics.value) return []

    return [...processMetrics.value.processes]
      .sort((a, b) => b.memoryPercent - a.memoryPercent)
      .slice(0, limit)
  }

  const getProcessByPID = (pid: number) => {
    if (!processMetrics.value) return null

    return processMetrics.value.processes.find(p => p.pid === pid) || null
  }

  // Disk helpers
  const getHighUsageDisks = (threshold: number = 90) => {
    if (!diskMetrics.value) return []

    return diskMetrics.value.disks.filter(disk => disk.usagePercent >= threshold)
  }

  // Network helpers
  const getActiveNetworkInterfaces = () => {
    if (!networkMetrics.value) return []

    return networkMetrics.value.interfaces.filter(iface => iface.isUp)
  }

  // Temperature helpers
  const getHottestSensors = (threshold: number = 70) => {
    if (!temperatureMetrics.value) return []

    return temperatureMetrics.value.sensors
      .filter(sensor => sensor.temperature >= threshold)
      .sort((a, b) => b.temperature - a.temperature)
  }

  // Lifecycle hooks
  onMounted(() => {
    if (wails.isReady.value) {
      loadSystemInfo()
      setupEventListeners()
      if (autoRefresh) {
        startAutoRefresh()
      }
    }
  })

  onUnmounted(() => {
    stopMonitoring()
    stopAutoRefresh()
    if (autoRefreshInterval.value) {
      clearInterval(autoRefreshInterval.value)
    }
  })

  // Auto-initialize when Wails is ready
  if (wails.isReady.value) {
    loadSystemInfo()
    setupEventListeners()
    if (autoRefresh) {
      startAutoRefresh()
    }
  }

  // Export state and functions
  return {
    // State (readonly for safety)
    systemMetrics: readonly(systemMetrics),
    systemInfo: readonly(systemInfo),
    systemStats: readonly(systemStats),
    processes: readonly(processes),
    isLoading: readonly(isLoading),
    error: readonly(error),
    lastUpdated: readonly(lastUpdated),
    isMonitoring: readonly(isMonitoring),

    // Backward compatibility computed properties
    isSystemInfoLoaded,
    isStatsLoaded,
    diskUsage,
    systemUptime,
    performanceStatus,

    // Computed
    systemHealth,
    cpuMetrics,
    memoryMetrics,
    processMetrics,
    diskMetrics,
    networkMetrics,
    temperatureMetrics,
    cpuUsage,
    memoryUsage,
    totalProcesses,
    runningProcesses,

    // Core functions
    refreshMetrics: debouncedRefresh,
    startMonitoring,
    stopMonitoring,
    loadSystemInfo,
    loadSystemStats,

    // Wails-specific functions
    executeCommand,
    killProcess,
    searchProcesses,
    getProcessByPid,
    startAutoRefresh,
    stopAutoRefresh,
    setRefreshRate,

    // Helpers
    formatBytes,
    formatFrequency,
    formatUptime,
    getTopCPUProcesses,
    getTopMemoryProcesses,
    getProcessByPID,
    getHighUsageDisks,
    getActiveNetworkInterfaces,
    getHottestSensors
  }
}