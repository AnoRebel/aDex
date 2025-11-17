export interface WailsEvent {
  name: string
  data: any
  timestamp: Date
}

export interface WailsResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  timestamp: Date
}

export interface WailsCallOptions {
  timeout?: number
  retries?: number
  context?: Record<string, any>
}

export interface WailsFileInfo {
  name: string
  size: number
  mode: string
  modTime: Date
  isDir: boolean
  path: string
}

export interface WailsDirEntry {
  name: string
  isDir: boolean
  size: number
  mode: string
  modTime: Date
}

export interface WailsSystemInfo {
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

export interface WailsProcess {
  pid: number
  name: string
  command: string
  status: string
  cpu: number
  memory: number
  user: string
  startTime: Date
  threads: number
  parentPid: number
}

export interface WailsNetworkInterface {
  name: string
  type: string
  isUp: boolean
  mtu: number
  speed: number
  mac: string
  addresses: {
    family: string
    address: string
    netmask: string
  }[]
  bytesReceived: number
  bytesSent: number
}

export interface WailsCpuInfo {
  usage: number
  cores: {
    id: number
    usage: number
    frequency: number
  }[]
  temperature: number
  loadAverage: number[]
}

export interface WailsMemoryInfo {
  total: number
  available: number
  used: number
  usage: number
  swapTotal: number
  swapUsed: number
  swapFree: number
}

export interface WailsDiskInfo {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usage: number
}

export interface WailsCommandResult {
  output: string
  error: string
  exitCode: number
}

export interface WailsTerminalSession {
  id: string
  shell: string
  workingDirectory: string
  environment: Record<string, string>
  isActive: boolean
  createdAt: Date
  lastActivity: Date
  rows?: number
  cols?: number
}

export interface WailsAudioDevice {
  id: string
  name: string
  type: string
  isDefault: boolean
  isEnabled: boolean
  volume: number
  isMuted: boolean
  channels: number
  sampleRate: number
  driver: string
}

export interface WailsTheme {
  name: string
  displayName: string
  description: string
  isDark: boolean
  colors: Record<string, string>
  typography: {
    fontFamily: string
    fontSize: Record<string, string>
  }
  effects: {
    glow: boolean
    animation: boolean
    shadows: boolean
  }
}

export interface WailsLogEntry {
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  timestamp: Date
  source: string
  context?: Record<string, any>
}

export interface WailsNotification {
  title: string
  message: string
  type: 'info' | 'warning' | 'error' | 'success'
  timeout?: number
  actions?: Array<{
    label: string
    action: string
  }>
}

export interface WailsFileDialogOptions {
  title: string
  defaultDirectory?: string
  defaultFilename?: string
  filters?: Array<{
    name: string
    patterns: string[]
  }>
  allowMultiple?: boolean
  canCreateDirectories?: boolean
  showHiddenFiles?: boolean
}

export interface WailsDirectoryDialogOptions {
  title: string
  defaultDirectory?: string
  canCreateDirectories?: boolean
  showHiddenFiles?: boolean
}

export interface WailsMessageDialogOptions {
  type: 'info' | 'warning' | 'error' | 'question'
  title: string
  message: string
  buttons?: string[]
  defaultButton?: string
  cancelButton?: string
}

export interface WailsWindowOptions {
  width: number
  height: number
  x?: number
  y?: number
  resizable: boolean
  fullscreen: boolean
  maximized: boolean
  minimized: boolean
  title: string
  center: boolean
  alwaysOnTop: boolean
}

export interface WailsClipboardData {
  text?: string
  image?: string
  files?: string[]
}

export interface WailsScreenInfo {
  width: number
  height: number
  scale: number
  isPrimary: boolean
  name: string
}

export interface WailsAppInfo {
  name: string
  version: string
  buildTime: Date
  commit: string
  description: string
  author: string
  license: string
}
