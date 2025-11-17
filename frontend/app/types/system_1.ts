// System monitoring types for aDex-UI

// Core system metrics interface
export interface SystemMetrics {
  cpu: CPUMetrics
  memory: MemoryMetrics
  processes: ProcessMetrics
  disks: DiskMetrics
  network: NetworkMetrics
  temperature: TemperatureMetrics | null
  timestamp: string
}

// CPU metrics interface
export interface CPUMetrics {
  usagePercent: number
  cores: number
  model: string
  vendor: string
  frequency: number
  frequencyMax: number
  perCoreUsage: number[]
  loadAverage?: number[]
  timestamp: string
}

// Memory metrics interface
export interface MemoryMetrics {
  total: number
  available: number
  used: number
  free: number
  usagePercent: number
  cached?: number
  buffers?: number
  swapTotal: number
  swapUsed: number
  swapFree: number
  swapPercent: number
  timestamp: string
}

// Process information interface
export interface ProcessInfo {
  pid: number
  ppid: number
  name: string
  command: string
  user: string
  status: string
  cpuPercent: number
  memoryPercent: number
  memoryRSS: number
  memoryVMS: number
  createTime: string
  numThreads: number
  numFDs: number
  cwd?: string
  executable?: string
}

// Process metrics interface
export interface ProcessMetrics {
  processes: ProcessInfo[]
  totalProcesses: number
  runningProcesses: number
  sleepingProcesses: number
  stoppedProcesses: number
  zombieProcesses: number
  timestamp: string
}

// Disk information interface
export interface DiskInfo {
  device: string
  mountpoint: string
  fsType: string
  total: number
  used: number
  free: number
  usagePercent: number
  inodesTotal?: number
  inodesUsed?: number
  inodesFree?: number
  readOnly: boolean
}

// Disk metrics interface
export interface DiskMetrics {
  disks: DiskInfo[]
  totalSpace: number
  totalUsed: number
  totalFree: number
  timestamp: string
}

// Network interface information
export interface NetworkInterface {
  name: string
  isUp: boolean
  bytesSent: number
  bytesRecv: number
  packetsSent: number
  packetsRecv: number
  errin: number
  errout: number
  dropin: number
  dropout: number
  ipAddresses: string[]
  mac?: string
  speed?: number
  mtu?: number
}

// Network metrics interface
export interface NetworkMetrics {
  interfaces: NetworkInterface[]
  totalBytesSent: number
  totalBytesRecv: number
  timestamp: string
}

// Temperature sensor interface
export interface TemperatureSensor {
  name: string
  temperature: number
  min: number
  max: number
  critical: number
  high: number
  unit: string
  devicePath?: string
}

// Temperature metrics interface
export interface TemperatureMetrics {
  sensors: TemperatureSensor[]
  timestamp: string
}

// System health status
export type SystemHealthStatus = 'good' | 'warning' | 'critical' | 'unknown'

// System health assessment
export interface SystemHealth {
  status: SystemHealthStatus
  issues: string[]
  score: number
  recommendations: string[]
}

// Performance score categories
export type PerformanceScore = 'excellent' | 'good' | 'fair' | 'poor'

// System configuration interface
export interface SystemConfig {
  refreshInterval: number
  maxProcesses: number
  enableTemperature: boolean
  enableNetwork: boolean
  enableDisk: boolean
  enableDetailedInfo: boolean
}

// System statistics interface
export interface SystemStatistics {
  cpu: Record<string, any>
  memory: Record<string, any>
  processes: Record<string, any>
  runtime: {
    goVersion: string
    goOS: string
    goArch: string
    numCPU: number
    numGoroutine: number
    numCgoCall: number
  }
  timestamp: Date
}

// Chart data interfaces
export interface ChartDataPoint {
  value: number
  timestamp: Date
  label?: string
}

export interface CPUChartData {
  overall: ChartDataPoint[]
  perCore: {
    [coreId: number]: ChartDataPoint[]
  }
}

export interface MemoryChartData {
  used: ChartDataPoint[]
  free: ChartDataPoint[]
  cached?: ChartDataPoint[]
  buffers?: ChartDataPoint[]
  swap?: ChartDataPoint[]
}

export interface ProcessChartData {
  topCPU: {
    [pid: number]: ChartDataPoint[]
  }
  topMemory: {
    [pid: number]: ChartDataPoint[]
  }
}

// Network chart data
export interface NetworkChartData {
  sent: ChartDataPoint[]
  received: ChartDataPoint[]
  interfaces: {
    [interfaceName: string]: {
      sent: ChartDataPoint[]
      received: ChartDataPoint[]
    }
  }
}

// Disk chart data
export interface DiskChartData {
  usage: {
    [mountpoint: string]: ChartDataPoint[]
  }
  throughput?: {
    [mountpoint: string]: {
      read: ChartDataPoint[]
      write: ChartDataPoint[]
    }
  }
}

// Temperature chart data
export interface TemperatureChartData {
  sensors: {
    [sensorName: string]: ChartDataPoint[]
  }
}

// Process sorting options
export type ProcessSortField = 'cpu' | 'memory' | 'pid' | 'name' | 'status' | 'threads' | 'fds'
export type ProcessSortOrder = 'asc' | 'desc'

// Process filter options
export interface ProcessFilter {
  user?: string
  status?: string
  name?: string
  minCpu?: number
  minMemory?: number
  maxCpu?: number
  maxMemory?: number
}

// Alert and warning interfaces
export interface SystemAlert {
  id: string
  type: 'warning' | 'critical' | 'info'
  category: 'cpu' | 'memory' | 'disk' | 'temperature' | 'network' | 'process'
  title: string
  message: string
  value?: number
  threshold?: number
  timestamp: Date
  acknowledged: boolean
}

// Performance thresholds
export interface PerformanceThresholds {
  cpu: {
    warning: number
    critical: number
  }
  memory: {
    warning: number
    critical: number
  }
  disk: {
    warning: number
    critical: number
  }
  temperature: {
    warning: number
    critical: number
  }
  swap: {
    warning: number
    critical: number
  }
}

// Historical data interface
export interface HistoricalData {
  timeframe: '1m' | '5m' | '15m' | '1h' | '6h' | '24h' | '7d'
  interval: number
  data: {
    cpu: CPUChartData
    memory: MemoryChartData
    network: NetworkChartData
    temperature?: TemperatureChartData
  }
}

// System events
export interface SystemEvent {
  id: string
  type: 'process_start' | 'process_end' | 'high_usage' | 'disk_full' | 'temperature_critical'
  title: string
  description: string
  data: Record<string, any>
  timestamp: Date
  severity: 'low' | 'medium' | 'high' | 'critical'
}

// System capabilities
export interface SystemCapabilities {
  temperatureMonitoring: boolean
  diskIoMonitoring: boolean
  networkMonitoring: boolean
  processMonitoring: boolean
  loadAverage: boolean
  swapMonitoring: boolean
  perCoreCpuUsage: boolean
}

// Export utility functions type
export interface SystemExportOptions {
  format: 'json' | 'csv' | 'xml'
  includeCharts: boolean
  timeframe: string
  metrics: string[]
}

// Component prop types
export interface SystemMonitorProps {
  autoRefresh?: boolean
  refreshInterval?: number
  showCharts?: boolean
  showDetails?: boolean
  maxProcesses?: number
  compact?: boolean
}

export interface CpuChartProps {
  data: CPUChartData
  showPerCore?: boolean
  showLoadAverage?: boolean
  height?: number
  width?: number
  theme?: 'light' | 'dark'
}

export interface MemoryChartProps {
  data: MemoryChartData
  showSwap?: boolean
  showBuffers?: boolean
  showCached?: boolean
  height?: number
  width?: number
  theme?: 'light' | 'dark'
}

export interface ProcessListProps {
  processes: ProcessInfo[]
  sortBy?: ProcessSortField
  sortOrder?: ProcessSortOrder
  filter?: ProcessFilter
  maxItems?: number
  showDetails?: boolean
  autoRefresh?: boolean
  compact?: boolean
}

// Action types
export type SystemAction =
  | { type: 'SET_METRICS'; payload: SystemMetrics }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'START_MONITORING' }
  | { type: 'STOP_MONITORING' }
  | { type: 'SET_REFRESH_INTERVAL'; payload: number }
  | { type: 'SET_MAX_PROCESSES'; payload: number }
  | { type: 'SET_CONFIG'; payload: Partial<SystemConfig> }
  | { type: 'CLEAR_HISTORY' }
  | { type: 'RESET' }

// Composable return type
export interface UseSystemReturn {
  // State
  systemMetrics: Ref<SystemMetrics | null>
  isLoading: Ref<boolean>
  error: Ref<string | null>
  lastUpdated: Ref<Date | null>
  isMonitoring: Ref<boolean>

  // Computed
  systemHealth: Ref<SystemHealth>
  performanceScore: Ref<PerformanceScore>
  cpuMetrics: Ref<CPUMetrics | null>
  memoryMetrics: Ref<MemoryMetrics | null>
  processMetrics: Ref<ProcessMetrics | null>
  diskMetrics: Ref<DiskMetrics | null>
  networkMetrics: Ref<NetworkMetrics | null>
  temperatureMetrics: Ref<TemperatureMetrics | null>
  cpuUsage: Ref<number>
  memoryUsage: Ref<number>
  totalProcesses: Ref<number>
  runningProcesses: Ref<number>

  // Methods
  refreshMetrics: () => Promise<void>
  startMonitoring: () => void
  stopMonitoring: () => void
  setRefreshInterval: (interval: number) => void
  setMaxProcesses: (max: number) => void

  // Helpers
  formatBytes: (bytes: number) => string
  formatFrequency: (mhz: number) => string
  formatUptime: (seconds: number) => string
  getTopCPUProcesses: (limit?: number) => ProcessInfo[]
  getTopMemoryProcesses: (limit?: number) => ProcessInfo[]
  getProcessByPID: (pid: number) => ProcessInfo | null
}

// Store return type
export interface SystemStore {
  // State
  systemMetrics: SystemMetrics | null
  isLoading: boolean
  error: string | null
  lastUpdated: Date | null
  isMonitoring: boolean
  refreshInterval: number
  maxProcesses: number
  enableTemperature: boolean
  enableNetwork: boolean
  enableDisk: boolean

  // History
  cpuHistory: number[]
  memoryHistory: number[]
  historyTimestamps: Date[]

  // Actions
  setSystemMetrics: (metrics: SystemMetrics) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  startMonitoring: () => void
  stopMonitoring: () => void
  setRefreshInterval: (interval: number) => void
  setMaxProcesses: (max: number) => void
  updateConfig: (config: Partial<SystemConfig>) => void
  clearHistory: () => void
  reset: () => void
}

// Color themes for charts
export interface ChartTheme {
  colors: {
    primary: string
    secondary: string
    success: string
    warning: string
    danger: string
    info: string
    background: string
    grid: string
    text: string
  }
}

// Default themes
export const LIGHT_THEME: ChartTheme = {
  colors: {
    primary: '#3b82f6',
    secondary: '#64748b',
    success: '#22c55e',
    warning: '#f59e0b',
    danger: '#ef4444',
    info: '#06b6d4',
    background: '#ffffff',
    grid: '#e5e7eb',
    text: '#111827'
  }
}

export const DARK_THEME: ChartTheme = {
  colors: {
    primary: '#60a5fa',
    secondary: '#94a3b8',
    success: '#34d399',
    warning: '#fbbf24',
    danger: '#f87171',
    info: '#22d3ee',
    background: '#1f2937',
    grid: '#374151',
    text: '#f9fafb'
  }
}