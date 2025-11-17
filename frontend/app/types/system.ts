// System types for aDex-UI

// ============================================================================
// BASIC SYSTEM INTERFACES
// ============================================================================

export interface SystemInfo {
  hostname: string
  platform: string
  architecture: string
  osVersion: string
  kernelVersion: string
  uptime: number
  bootTime: Date
  timezone: string
  cpuCount: number
  totalMemory: number
  hostId: string
  username: string
  homeDir: string
}

export interface SystemStats {
  cpu: CpuUsage
  memory: MemoryUsage
  disk: DiskUsage[]
  network: NetworkUsage
  timestamp: Date
}

export interface CpuUsage {
  usage: number
  cores: CpuCore[]
  temperature: number
  loadAverage: number[]
}

export interface CpuCore {
  id: number
  usage: number
  frequency: number
}

export interface MemoryUsage {
  total: number
  available: number
  used: number
  usage: number
  swapTotal: number
  swapUsed: number
  swapFree: number
}

export interface DiskUsage {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usage: number
  label?: string
}

export interface NetworkUsage {
  interfaces: NetworkInterface[]
  bytesReceived: number
  bytesSent: number
  packetsReceived: number
  packetsSent: number
}

export interface NetworkInterface {
  name: string
  type: string
  isUp: boolean
  mtu: number
  speed: number
  mac: string
  addresses: NetworkAddress[]
  bytesReceived: number
  bytesSent: number
  packetsReceived: number
  packetsSent: number
}

export interface NetworkAddress {
  family: 'ipv4' | 'ipv6'
  address: string
  netmask: string
  broadcast?: string
}

export interface Process {
  pid: number
  name: string
  command: string
  status: 'running' | 'sleeping' | 'stopped' | 'zombie' | 'dead'
  cpu: number
  memory: number
  user: string
  startTime: Date
  threads: number
  parentPid: number
  children?: number
  priority?: number
  nice?: number
}

// ============================================================================
// DETAILED MONITORING INTERFACES
// ============================================================================

// Comprehensive metrics interface
export interface SystemMetrics {
  cpu: CPUMetrics
  memory: MemoryMetrics
  processes: ProcessMetrics
  disks: DiskMetrics
  network: NetworkMetrics
  temperature: TemperatureMetrics | null
  timestamp: string
}

// CPU metrics with detailed information
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

// Memory metrics with comprehensive details
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

// Process information
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

// Process metrics
export interface ProcessMetrics {
  processes: ProcessInfo[]
  totalProcesses: number
  runningProcesses: number
  sleepingProcesses: number
  stoppedProcesses: number
  zombieProcesses: number
  timestamp: string
}

// Disk information
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

// Disk metrics
export interface DiskMetrics {
  disks: DiskInfo[]
  totalSpace: number
  totalUsed: number
  totalFree: number
  timestamp: string
}

// Network interface information
export interface NetworkInterfaceInfo {
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

// Network metrics
export interface NetworkMetrics {
  interfaces: NetworkInterfaceInfo[]
  totalBytesSent: number
  totalBytesRecv: number
  timestamp: string
}

// Temperature sensor information
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

// ============================================================================
// UTILITY INTERFACES
// ============================================================================

export interface SystemService {
  name: string
  description: string
  status: 'active' | 'inactive' | 'failed' | 'activating' | 'deactivating'
  enabled: boolean
  loaded: boolean
  startup: boolean
}

export interface SystemResource {
  type: 'cpu' | 'memory' | 'disk' | 'network'
  name: string
  usage: number
  total: number
  available: number
  unit: string
}

export interface PerformanceMetric {
  name: string
  value: number
  unit: string
  timestamp: Date
  category: 'system' | 'process' | 'network' | 'disk'
}

export interface SystemAlert {
  id: string
  type: 'warning' | 'error' | 'info' | 'critical'
  title: string
  message: string
  source: string
  timestamp: Date
  acknowledged: boolean
  severity: 'low' | 'medium' | 'high' | 'critical'
  category?: 'cpu' | 'memory' | 'disk' | 'temperature' | 'network' | 'process'
  value?: number
  threshold?: number
}

export interface SystemCommand {
  command: string
  args: string[]
  workingDirectory?: string
  environment?: Record<string, string>
  timeout?: number
}

export interface CommandResult {
  success: boolean
  output: string
  error: string
  exitCode: number
  duration: number
}

// ============================================================================
// CHART DATA INTERFACES
// ============================================================================

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

// ============================================================================
// CONFIGURATION AND STATE INTERFACES
// ============================================================================

export interface SystemConfig {
  refreshInterval: number
  maxProcesses: number
  enableTemperature: boolean
  enableNetwork: boolean
  enableDisk: boolean
  enableDetailedInfo: boolean
  temperatureUnit: 'celsius' | 'fahrenheit'
}

export interface SystemHealth {
  status: 'good' | 'warning' | 'critical' | 'unknown'
  issues: string[]
  score: number
  recommendations: string[]
}

export interface PerformanceThresholds {
  cpu: { warning: number; critical: number }
  memory: { warning: number; critical: number }
  disk: { warning: number; critical: number }
  temperature: { warning: number; critical: number }
  swap: { warning: number; critical: number }
}

// ============================================================================
// TYPE GUARDS AND UTILITIES
// ============================================================================

// Type guards
export function isSystemMetrics(obj: any): obj is SystemMetrics {
  return obj && typeof obj === 'object' && 'cpu' in obj && 'memory' in obj && 'timestamp' in obj
}

export function isSystemStats(obj: any): obj is SystemStats {
  return obj && typeof obj === 'object' && 'cpu' in obj && 'memory' in obj && 'disk' in obj && 'timestamp' in obj
}

// Conversion utilities
export function systemStatsToMetrics(stats: SystemStats): SystemMetrics {
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