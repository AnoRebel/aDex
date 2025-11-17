export interface TerminalSession {
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

export interface TerminalCommand {
  id: string
  command: string
  sessionId: string
  timestamp: Date
  status: 'running' | 'success' | 'error'
  result?: any
  duration?: number
}

export interface TerminalOutput {
  content: string
  type: 'normal' | 'error' | 'success' | 'warning' | 'command'
  timestamp: string
}

export interface TerminalConfig {
  shell: string
  fontSize: number
  fontFamily: string
  backgroundColor: string
  textColor: string
  cursorColor: string
  cursorBlink: boolean
  scrollback: number
  opacity: number
  bell: string
  shortcuts: Record<string, string>
}

export interface ProcessInfo {
  pid: number
  name: string
  command: string
  status: 'running' | 'stopped' | 'zombie'
  cpu: number
  memory: number
  user: string
  startTime: Date
  threads: number
  parentPid: number
}

export interface ShellEnvironment {
  variables: Record<string, string>
  path: string[]
  home: string
  user: string
  shell: string
  term: string
}

export interface TerminalEvent {
  type: 'output' | 'error' | 'close' | 'resize'
  sessionId: string
  data?: any
  timestamp: Date
}
