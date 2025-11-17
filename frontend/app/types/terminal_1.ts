/**
 * Terminal-related type definitions for the eDEX-UI frontend
 * Provides comprehensive TypeScript interfaces for terminal functionality
 */

// ============================================================================
// Core Terminal Types
// ============================================================================

export interface TerminalSession {
  id: string
  pid?: number
  shell: string
  cwd: string
  env: Record<string, string>
  size: TerminalSize
  active: boolean
  createdAt: string
  updatedAt: string
  lastSeen: string
  user?: string
  title?: string
  metadata?: Record<string, any>
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
  user?: string
}

// ============================================================================
// Terminal Tab Types
// ============================================================================

export interface TerminalTab {
  id: string
  sessionId: string
  title: string
  active: boolean
  icon?: string
  color?: string
  badge?: string | number
  position: number
  pinned: boolean
  modified: boolean
  lastActivity: string
}

export interface TerminalTabConfig {
  showIcons: boolean
  showBadges: boolean
  allowReorder: boolean
  allowDuplicate: boolean
  allowClose: boolean
  maxTitleLength: number
  showSessionStatus: boolean
  colorizeTabs: boolean
  groupBySession: boolean
}

// ============================================================================
// Terminal Theme Types
// ============================================================================

export interface TerminalTheme {
  id: string
  name: string
  displayName?: string
  description?: string
  colors: TerminalThemeColors
  font: TerminalThemeFont
  cursor: TerminalThemeCursor
  background?: TerminalThemeBackground
  extensions?: Record<string, any>
  author?: string
  version?: string
  isBuiltIn?: boolean
  isDark?: boolean
}

export interface TerminalThemeColors {
  background: string
  foreground: string
  cursor: string
  cursorAccent?: string
  selection: string
  selectionForeground?: string
  black: string
  red: string
  green: string
  yellow: string
  blue: string
  magenta: string
  cyan: string
  white: string
  brightBlack: string
  brightRed: string
  brightGreen: string
  brightYellow: string
  brightBlue: string
  brightMagenta: string
  brightCyan: string
  brightWhite: string
  ansi?: {
    [key: number]: string
  }
}

export interface TerminalThemeFont {
  family: string
  size: number
  weight: string | number
  lineHeight: number
  letterSpacing?: number
  ligatures: boolean
  antialias: boolean
  hinting: 'none' | 'slight' | 'medium' | 'full'
}

export interface TerminalThemeCursor {
  style: 'block' | 'underline' | 'bar'
  blink: boolean
  blinkInterval?: number
  width?: number
  color?: string
  accent?: string
}

export interface TerminalThemeBackground {
  type: 'solid' | 'gradient' | 'image'
  value?: string
  opacity?: number
  blur?: number
  size?: 'cover' | 'contain' | 'auto'
  position?: string
}

// ============================================================================
// Terminal Profile Types
// ============================================================================

export interface TerminalProfile {
  id: string
  name: string
  description?: string
  shell: string
  args?: string[]
  workingDir: string
  env: Record<string, string>
  user?: string
  icon?: string
  color?: string
  theme?: string
  font?: TerminalThemeFont
  size: TerminalSize
  isDefault: boolean
  isSystem: boolean
  createdAt: string
  updatedAt: string
  shortcuts?: Record<string, string>
  aliases?: Record<string, string>
  functions?: string[]
}

// ============================================================================
// Terminal Command Types
// ============================================================================

export interface TerminalCommand {
  id: string
  sessionId: string
  command: string
  arguments?: string[]
  cwd: string
  startTime: string
  endTime?: string
  exitCode?: number
  duration?: number
  user?: string
  title?: string
  processId?: number
  parentId?: string
  workingDirectory?: string
  environment?: Record<string, string>
  output?: string[]
  error?: string
  status: CommandStatus
  type: CommandType
}

export enum CommandStatus {
  Pending = 'pending',
  Running = 'running',
  Completed = 'completed',
  Failed = 'failed',
  Cancelled = 'cancelled',
  Timeout = 'timeout'
}

export enum CommandType {
  Interactive = 'interactive',
  Background = 'background',
  System = 'system',
  User = 'user',
  Script = 'script'
}

export interface TerminalCommandHistory {
  sessionId: string
  commands: TerminalCommand[]
  maxSize: number
  current: number
  searchIndex: number
  filter?: string
  sortBy: 'time' | 'name' | 'duration'
  sortOrder: 'asc' | 'desc'
}

// ============================================================================
// Terminal Event Types
// ============================================================================

export interface TerminalEvent {
  type: string
  sessionId: string
  timestamp: string
  data: any
  source: 'user' | 'process' | 'system'
  id?: string
  correlationId?: string
}

export interface TerminalOutputEvent extends TerminalEvent {
  type: 'terminal.output'
  data: {
    content: string
    length: number
    isStderr?: boolean
    chunkIndex?: number
    isLastChunk?: boolean
  }
}

export interface TerminalInputEvent extends TerminalEvent {
  type: 'terminal.input'
  data: {
    content: string
    length: number
    isCommand?: boolean
  }
}

export interface TerminalResizeEvent extends TerminalEvent {
  type: 'terminal.resized'
  data: {
    size: TerminalSize
    previousSize: TerminalSize
    reason: 'user' | 'auto' | 'system'
  }
}

export interface TerminalFocusEvent extends TerminalEvent {
  type: 'terminal.focus.changed'
  data: {
    focused: boolean
    reason: 'click' | 'keyboard' | 'system'
  }
}

export interface TerminalTitleChangeEvent extends TerminalEvent {
  type: 'terminal.title.changed'
  data: {
    title: string
    previousTitle: string
  }
}

export interface TerminalDirectoryChangeEvent extends TerminalEvent {
  type: 'terminal.directory.changed'
  data: {
    cwd: string
    previousCwd?: string
  }
}

export interface TerminalBellEvent extends TerminalEvent {
  type: 'terminal.bell'
  data: {
    visual: boolean
    audible: boolean
  }
}

export interface TerminalErrorEvent extends TerminalEvent {
  type: 'terminal.error'
  data: {
    error: string
    code?: string
    context?: string
    recoverable: boolean
  }
}

export interface TerminalNotificationEvent extends TerminalEvent {
  type: 'terminal.notification'
  data: {
    type: 'info' | 'warning' | 'error' | 'success'
    title: string
    message: string
    persistent?: boolean
    actions?: NotificationAction[]
  }
}

export interface NotificationAction {
  id: string
  label: string
  action: string
  style?: 'default' | 'primary' | 'danger'
}

// ============================================================================
// Terminal Settings Types
// ============================================================================

export interface TerminalSettings {
  // General
  defaultShell: string
  defaultWorkingDir: string
  defaultProfile: string
  confirmClose: boolean
  rememberWorkingDir: boolean

  // Appearance
  theme: string
  fontFamily: string
  fontSize: number
  lineHeight: number
  letterSpacing: number
  cursorStyle: 'block' | 'underline' | 'bar'
  cursorBlink: boolean
  opacity: number
  enableTransparency: boolean
  blurBackground: boolean

  // Behavior
  copyOnSelect: boolean
  autoFocus: boolean
  enableBell: boolean
  enableNotifications: boolean
  scrollOnOutput: boolean
  scrollOnKeystroke: boolean
  altGrIsMeta: boolean
  convertEol: boolean

  // Buffer and History
  scrollbackSize: number
  maxOutputBuffer: number
  maxCommandHistory: number
  clearSelectionOnCopy: boolean

  // Advanced
  rendererType: 'dom' | 'webgl' | 'canvas'
  gpuAcceleration: boolean
  enableLigatures: boolean
  enableSubpixelFontRendering: boolean
  fastScrollModifier: 'alt' | 'ctrl' | 'shift'
  wordSeparator: string

  // Accessibility
  screenReaderMode: boolean
  highContrast: boolean
  reduceMotion: boolean
  fontSizeAdjustment: number

  // Developer
  debugMode: boolean
  logLevel: 'debug' | 'info' | 'warn' | 'error'
  enablePerfMonitoring: boolean
}

export interface TerminalKeyBindings {
  [key: string]: string | KeyBindingAction
}

export interface KeyBindingAction {
  action: string
  args?: any[]
  when?: string
}

// ============================================================================
// Terminal Process Types
// ============================================================================

export interface TerminalProcess {
  pid: number
  ppid: number
  name: string
  command: string
  arguments: string[]
  cwd: string
  user: string
  startTime: string
  cpuUsage?: number
  memoryUsage?: number
  status: ProcessStatus
  children?: TerminalProcess[]
}

export enum ProcessStatus {
  Running = 'running',
  Sleeping = 'sleeping',
  Stopped = 'stopped',
  Zombie = 'zombie',
  Dead = 'dead'
}

// ============================================================================
// UI Component Types
// ============================================================================

export interface TerminalComponentConfig {
  showHeader: boolean
  showFooter: boolean
  showTabs: boolean
  showResizeHandle: boolean
  showStatusBar: boolean
  showScrollBars: boolean
  showLineNumbers: boolean
  showMinimap: boolean
  allowFullscreen: boolean
  allowSplitting: boolean
  maxTabs: number
}

export interface TerminalResizeConfig {
  direction: 'horizontal' | 'vertical' | 'both'
  minWidth: number
  minHeight: number
  maxWidth?: number
  maxHeight?: number
  step: number
  snapToGrid: boolean
  gridSize: number
  preserveAspectRatio: boolean
  aspectRatio?: number
}

export interface TerminalTabConfig {
  closable: boolean
  draggable: boolean
  duplicatable: boolean
  pinnable: boolean
  reorderable: boolean
  showIcons: boolean
  showBadges: boolean
  showCloseButtons: boolean
  maxTitleLength: number
  colorByStatus: boolean
  groupByProcess: boolean
}

// ============================================================================
// Terminal State Types
// ============================================================================

export interface TerminalState {
  sessions: Map<string, TerminalSession>
  tabs: Map<string, TerminalTab>
  activeSessionId: string | null
  outputBuffers: Map<string, string[]>
  commandHistories: Map<string, TerminalCommandHistory>
  availableThemes: TerminalTheme[]
  currentTheme: TerminalTheme | null
  availableProfiles: TerminalProfile[]
  currentProfile: TerminalProfile | null
  settings: TerminalSettings
  isLoading: boolean
  error: string | null
  lastActivity: string
}

export interface TerminalUIState {
  isFullscreen: boolean
  isFocused: boolean
  showContextMenu: boolean
  contextMenuPosition: { x: number; y: number }
  showSearch: boolean
  searchQuery: string
  searchResults: SearchResult[]
  showSettings: boolean
  notifications: TerminalNotification[]
  shortcuts: TerminalShortcut[]
}

export interface SearchResult {
  sessionId: string
  line: number
  column: number
  content: string
  context: string[]
}

export interface TerminalNotification {
  id: string
  sessionId?: string
  type: 'info' | 'warning' | 'error' | 'success'
  title: string
  message: string
  timestamp: string
  persistent: boolean
  read: boolean
  actions?: NotificationAction[]
}

export interface TerminalShortcut {
  id: string
  key: string
  modifiers: string[]
  action: string
  description: string
  category: string
  enabled: boolean
  global: boolean
}

// ============================================================================
// Utility Types
// ============================================================================

export type TerminalEventMap = {
  'terminal.output': TerminalOutputEvent
  'terminal.input': TerminalInputEvent
  'terminal.resized': TerminalResizeEvent
  'terminal.focus.changed': TerminalFocusEvent
  'terminal.title.changed': TerminalTitleChangeEvent
  'terminal.directory.changed': TerminalDirectoryChangeEvent
  'terminal.bell': TerminalBellEvent
  'terminal.error': TerminalErrorEvent
  'terminal.notification': TerminalNotificationEvent
  'terminal.session.created': TerminalEvent
  'terminal.session.closed': TerminalEvent
  'terminal.command.started': TerminalEvent
  'terminal.command.completed': TerminalEvent
}

export type TerminalEventType = keyof TerminalEventMap

export type TerminalEventHandler<T extends TerminalEventType> = (
  event: TerminalEventMap[T]
) => void | Promise<void>

export type TerminalThemePreset = Omit<TerminalTheme, 'id' | 'isBuiltIn'>

export type TerminalFontFamily =
  | 'JetBrains Mono'
  | 'Fira Code'
  | 'Source Code Pro'
  | 'Cascadia Code'
  | 'IBM Plex Mono'
  | 'Space Mono'
  | 'Ubuntu Mono'
  | 'Consolas'
  | 'Monaco'
  | string

export type TerminalFontSize = number | `${number}px` | `${number}pt` | `${number}em`

export type TerminalThemeMode = 'light' | 'dark' | 'auto'

export type TerminalLogLevel = 'debug' | 'info' | 'warn' | 'error'

export type TerminalKeyModifier = 'ctrl' | 'alt' | 'shift' | 'meta' | 'cmd'

export type TerminalPlatform = 'windows' | 'macos' | 'linux' | 'unix'

export type TerminalArchitecture = 'x64' | 'arm64' | 'ia32'

// ============================================================================
// Error Types
// ============================================================================

export interface TerminalError {
  code: string
  message: string
  details?: any
  sessionId?: string
  timestamp: string
  recoverable: boolean
  suggestions?: string[]
}

export interface TerminalValidationError extends TerminalError {
  field: string
  value: any
  constraint: string
}

export interface TerminalConnectionError extends TerminalError {
  endpoint: string
  retryCount: number
  maxRetries: number
}

// ============================================================================
// Export Default Types
// ============================================================================

export type {
  // Core types
  TerminalSession as Session,
  TerminalSize as Size,
  TerminalOptions as Options,

  // UI types
  TerminalTab as Tab,
  TerminalTheme as Theme,
  TerminalProfile as Profile,
  TerminalSettings as Settings,

  // Event types
  TerminalEvent as Event,
  TerminalOutputEvent as OutputEvent,
  TerminalInputEvent as InputEvent,

  // Utility types
  TerminalCommand as Command,
  TerminalProcess as Process,
  TerminalNotification as Notification
}

export {
  // Enums
  CommandStatus,
  CommandType,
  ProcessStatus
}

export default {
  // Core
  TerminalSession,
  TerminalSize,
  TerminalOptions,

  // UI
  TerminalTab,
  TerminalTheme,
  TerminalProfile,
  TerminalSettings,

  // Events
  TerminalEvent,
  TerminalOutputEvent,
  TerminalInputEvent,
  TerminalResizeEvent,
  TerminalFocusEvent,
  TerminalTitleChangeEvent,
  TerminalDirectoryChangeEvent,
  TerminalBellEvent,
  TerminalErrorEvent,
  TerminalNotificationEvent,

  // Commands
  TerminalCommand,
  TerminalCommandHistory,

  // Processes
  TerminalProcess,

  // State
  TerminalState,
  TerminalUIState,

  // Enums
  CommandStatus,
  CommandType,
  ProcessStatus,

  // Utilities
  TerminalError,
  TerminalValidationError,
  TerminalConnectionError
}