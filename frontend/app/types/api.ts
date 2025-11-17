// API type definitions for aDex-UI frontend-backend communication

// ============================================================================
// Common Types
// ============================================================================

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: ApiError
  timestamp: string
}

export interface ApiError {
  code: string
  type: string
  message: string
  description?: string
  context?: Record<string, any>
  user_message?: string
  retryable: boolean
}

export interface PaginatedResponse<T> extends ApiResponse<T[]> {
  pagination: {
    page: number
    limit: number
    total: number
    totalPages: number
    hasNext: boolean
    hasPrev: boolean
  }
}

// ============================================================================
// System Types
// ============================================================================

export interface SystemInfo {
  cpu: CPUInfo
  memory: MemoryInfo
  disk: DiskInfo[]
  network: NetworkInfo[]
  processes: ProcessInfo[]
  uptime: number
  loadAverage: number[]
  timestamp: string
}

export interface CPUInfo {
  usage: number
  cores: number
  model: string
  frequency: number
  temperature?: number
  usagePerCore?: number[]
}

export interface MemoryInfo {
  total: number
  used: number
  free: number
  available: number
  usage: number
  swap: SwapInfo
}

export interface SwapInfo {
  total: number
  used: number
  free: number
  usage: number
}

export interface DiskInfo {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usage: number
  readonly: boolean
}

export interface NetworkInfo {
  interface: string
  status: 'up' | 'down'
  ip_addresses: string[]
  mac_address?: string
  speed?: number
  bytes_sent: number
  bytes_received: number
  packets_sent: number
  packets_received: number
  errors_in: number
  errors_out: number
}

export interface ProcessInfo {
  pid: number
  ppid: number
  name: string
  command: string
  cpu: number
  memory: number
  status: string
  user: string
  start_time: string
  threads: number
  nice?: number
}

// ============================================================================
// Terminal Types
// ============================================================================

export interface TerminalSession {
  id: string
  pid?: number
  shell: string
  cwd: string
  env: Record<string, string>
  size: TerminalSize
  active: boolean
  created_at: string
  last_activity: string
}

export interface TerminalSize {
  rows: number
  cols: number
}

export interface TerminalOptions {
  shell?: string
  cwd?: string
  env?: Record<string, string>
  rows?: number
  cols?: number
  profile?: string
}

export interface TerminalData {
  session_id: string
  data: string
  type: 'input' | 'output' | 'error'
  timestamp: string
}

export interface TerminalCommand {
  session_id: string
  command: string
  cwd: string
  timestamp: string
  exit_code?: number
  duration?: number
}

// ============================================================================
// File System Types
// ============================================================================

export interface FileEntry {
  path: string
  name: string
  type: 'file' | 'directory' | 'symlink'
  size: number
  permissions: string
  owner: string
  group: string
  modified: string
  accessed: string
  created: string
  hidden: boolean
  symlink_target?: string
  mime_type?: string
}

export interface FileOperation {
  id: string
  type: 'copy' | 'move' | 'delete' | 'create' | 'rename'
  source: string
  destination?: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  error?: string
  created_at: string
  completed_at?: string
}

export interface FileWatcher {
  id: string
  path: string
  recursive: boolean
  events: FileWatcherEvent[]
  active: boolean
  created_at: string
}

export interface FileWatcherEvent {
  type: 'created' | 'modified' | 'deleted' | 'moved'
  path: string
  old_path?: string
  timestamp: string
}

// ============================================================================
// Audio Types
// ============================================================================

export interface AudioDevice {
  id: string
  name: string
  type: 'input' | 'output'
  driver: string
  volume: number
  muted: boolean
  channels: number
  sample_rate: number
  default: boolean
  available: boolean
}

export interface AudioSession {
  id: string
  device_id: string
  process_id?: number
  process_name?: string
  volume: number
  muted: boolean
  created_at: string
  active: boolean
}

export interface AudioSettings {
  master_volume: number
  default_output: string
  default_input: string
  enable_notifications: boolean
  enable_system_sounds: boolean
}

// ============================================================================
// Configuration Types
// ============================================================================

export interface ConfigSection {
  key: string
  value: any
  type: 'string' | 'number' | 'boolean' | 'object' | 'array'
  description?: string
  default_value?: any
  required?: boolean
  validation?: ValidationRule[]
}

export interface ValidationRule {
  type: 'required' | 'min' | 'max' | 'pattern' | 'custom'
  value?: any
  message?: string
}

export interface ConfigFile {
  version: string
  sections: Record<string, ConfigSection>
  metadata: {
    created_at: string
    updated_at: string
    schema_version: string
  }
}

// ============================================================================
// Theme Types
// ============================================================================

export interface Theme {
  id: string
  name: string
  description?: string
  author: string
  version: string
  is_dark: boolean
  colors: ThemeColors
  fonts: ThemeFonts
  effects: ThemeEffects
  custom_properties?: Record<string, string>
}

export interface ThemeColors {
  primary: string
  secondary: string
  accent: string
  background: string
  surface: string
  text: string
  text_secondary: string
  text_disabled: string
  border: string
  border_focus: string
  success: string
  warning: string
  error: string
  info: string
  terminal: {
    background: string
    foreground: string
    cursor: string
    selection: string
    black: string
    red: string
    green: string
    yellow: string
    blue: string
    magenta: string
    cyan: string
    white: string
    bright_black: string
    bright_red: string
    bright_green: string
    bright_yellow: string
    bright_blue: string
    bright_magenta: string
    bright_cyan: string
    bright_white: string
  }
}

export interface ThemeFonts {
  primary: string
  monospace: string
  sizes: {
    xs: string
    sm: string
    base: string
    lg: string
    xl: string
    '2xl': string
    '3xl': string
  }
}

export interface ThemeEffects {
  blur: boolean
  shadows: boolean
  animations: boolean
  transparency: boolean
  border_radius: string
}

// ============================================================================
// Event Types
// ============================================================================

export interface AppEvent {
  type: string
  data: any
  source: string
  timestamp: string
  id?: string
}

export interface SystemEvent extends AppEvent {
  type: 'system.info.updated' | 'system.alert' | 'system.process.started' | 'system.process.ended'
}

export interface TerminalEvent extends AppEvent {
  type: 'terminal.created' | 'terminal.closed' | 'terminal.resized' | 'terminal.output' | 'terminal.input' | 'terminal.command'
}

export interface FileEvent extends AppEvent {
  type: 'file.created' | 'file.modified' | 'file.deleted' | 'file.moved' | 'directory.created' | 'directory.deleted' | 'directory.modified'
}

export interface AudioEvent extends AppEvent {
  type: 'audio.device.changed' | 'audio.volume.changed' | 'audio.session.started' | 'audio.session.ended'
}

export interface ConfigEvent extends AppEvent {
  type: 'config.changed' | 'config.saved' | 'config.loaded' | 'config.reset'
}

export interface ThemeEvent extends AppEvent {
  type: 'theme.changed' | 'theme.loaded' | 'theme.created' | 'theme.deleted'
}

export interface NetworkEvent extends AppEvent {
  type: 'network.connected' | 'network.disconnected' | 'network.error'
}

export interface ErrorEvent extends AppEvent {
  type: 'error.occurred' | 'panic.occurred' | 'warning.issued'
}

// ============================================================================
// Service Types
// ============================================================================

export interface ServiceStatus {
  name: string
  status: 'stopped' | 'starting' | 'running' | 'stopping' | 'error'
  health: 'healthy' | 'unhealthy' | 'unknown'
  uptime?: number
  last_error?: string
  version?: string
  dependencies?: string[]
}

export interface ServiceConfig {
  name: string
  enabled: boolean
  auto_start: boolean
  config: Record<string, any>
  dependencies: string[]
  environment?: Record<string, string>
}

// ============================================================================
// UI Types
// ============================================================================

export interface PanelConfig {
  id: string
  type: 'terminal' | 'monitor' | 'filesystem' | 'network' | 'settings'
  visible: boolean
  position: { x: number; y: number }
  size: { width: number; height: number }
  config: Record<string, any>
}

export interface Shortcut {
  id: string
  keys: string[]
  action: string
  description: string
  global: boolean
  enabled: boolean
}

export interface NotificationSettings {
  enabled: boolean
  sound: boolean
  duration: number
  position: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'
  types: {
    info: boolean
    warning: boolean
    error: boolean
    success: boolean
  }
}

// ============================================================================
// Utility Types
// ============================================================================

export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P]
}

export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>

export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>

export type EventCallback<T = any> = (event: T) => void | Promise<void>

export type AsyncFunction<T = any, R = any> = (...args: T[]) => Promise<R>

export type ServiceMethod<T = any, R = any> = (...args: T[]) => Promise<R>

// ============================================================================
// API Client Types
// ============================================================================

export interface ApiClient {
  get<T>(endpoint: string, params?: Record<string, any>): Promise<ApiResponse<T>>
  post<T>(endpoint: string, data?: any): Promise<ApiResponse<T>>
  put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>>
  delete<T>(endpoint: string): Promise<ApiResponse<T>>
  subscribe<T>(event: string, callback: EventCallback<T>): () => void
  unsubscribe(event: string, callback: EventCallback<T>): void
  publish(event: string, data: any): Promise<void>
}

// ============================================================================
// Validation Types
// ============================================================================

export interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
}

export interface ValidationError {
  field: string
  message: string
  code: string
  value?: any
}

export interface ValidationSchema {
  [key: string]: ValidationRule[]
}

// ============================================================================
// Export all types for easy importing
// ============================================================================

export * from './terminal'
export * from './system'
export * from './filesystem'
export * from './audio'
export * from './theme'
export * from './config'