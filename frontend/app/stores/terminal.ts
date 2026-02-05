import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { CreateTerminal, WriteToTerminal, ResizeTerminal, CloseTerminal } from '~~/bindings'
import type { TerminalSession, TerminalCommand, TerminalOutput } from '~/types/terminal'

// Import types from useTerminal composable for tab and theme support
interface TerminalTab {
  id: string
  sessionId: string
  title: string
  position: number
  active: boolean
  closable: boolean
  icon?: string
  customTitle?: string
}

interface TerminalTheme {
  id: string
  name: string
  colors: {
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
    brightBlack: string
    brightRed: string
    brightGreen: string
    brightYellow: string
    brightBlue: string
    brightMagenta: string
    brightCyan: string
    brightWhite: string
  }
  fontFamily?: string
  fontSize?: number
  lineHeight?: number
  cursorStyle?: 'block' | 'underline' | 'bar'
  cursorBlink?: boolean
}

interface TerminalSettings {
  defaultShell: string
  defaultWorkingDir: string
  defaultRows: number
  defaultCols: number
  enableBell: boolean
  enableNotifications: boolean
  maxOutputBuffer: number
  maxCommandHistory: number
  autoFocus: boolean
  confirmClose: boolean
  rememberWorkingDir: boolean
  copyOnSelect: boolean
  theme: string
  fontFamily: string
  fontSize: number
  lineHeight: number
  cursorStyle: 'block' | 'underline' | 'bar'
  cursorBlink: boolean
  scrollbackSize: number
  enableTransparency: boolean
  opacity: number
}

interface TerminalState {
  sessions: Map<string, TerminalSession>
  tabs: Map<string, TerminalTab>
  activeSessionId: string | null
  commandHistory: Map<string, string[]> // Per-session command history
  outputBuffer: Map<string, TerminalOutput[]> // Per-session output buffer
  availableThemes: TerminalTheme[]
  currentTheme: TerminalTheme | null
  settings: TerminalSettings
  config: {
    shell: string
    fontSize: number
    fontFamily: string
    backgroundColor: string
    textColor: string
    cursorColor: string
    cursorBlink: boolean
    scrollback: number
    opacity: number
  }
  isLoading: boolean
  lastError: string | null
}

// Default terminal settings
const defaultSettings = (): TerminalSettings => ({
  defaultShell: '/bin/bash',
  defaultWorkingDir: '/',
  defaultRows: 24,
  defaultCols: 80,
  enableBell: true,
  enableNotifications: true,
  maxOutputBuffer: 1000,
  maxCommandHistory: 1000,
  autoFocus: true,
  confirmClose: false,
  rememberWorkingDir: true,
  copyOnSelect: false,
  theme: 'dark',
  fontFamily: '"JetBrains Mono", "Fira Code", monospace',
  fontSize: 14,
  lineHeight: 1.2,
  cursorStyle: 'block',
  cursorBlink: true,
  scrollbackSize: 1000,
  enableTransparency: false,
  opacity: 0.95
})

export const useTerminalStore = defineStore('terminal', () => {
  // State
  const sessions = ref<Map<string, TerminalSession>>(new Map())
  const tabs = ref<Map<string, TerminalTab>>(new Map())
  const activeSessionId = ref<string | null>(null)
  const commandHistory = ref<Map<string, TerminalCommand[]>>(new Map())
  const outputBuffer = ref<Map<string, TerminalOutput[]>>(new Map())
  const availableThemes = ref<TerminalTheme[]>([])
  const currentTheme = ref<TerminalTheme | null>(null)
  const settings = ref<TerminalSettings>(defaultSettings())
  const config = ref({
    shell: '/bin/bash',
    fontSize: 14,
    fontFamily: '"JetBrains Mono", monospace',
    backgroundColor: '#000000',
    textColor: '#00ff00',
    cursorColor: '#00ff00',
    cursorBlink: true,
    scrollback: 1000,
    opacity: 0.9,
  })
  const isLoading = ref(false)
  const lastError = ref<string | null>(null)

  // Computed properties (getters)
  const activeSession = computed(() => {
    return activeSessionId.value ? sessions.value.get(activeSessionId.value) : null
  })

  const allSessions = computed(() => {
    return Array.from(sessions.value.values())
  })

  const sessionOutput = computed(() => {
    return (sessionId: string) => outputBuffer.value.get(sessionId) || []
  })

  const hasActiveSession = computed(() => {
    return !!activeSessionId.value
  })

  const recentCommands = computed(() => {
    const allCommands = Array.from(commandHistory.value.values()).flat()
    return allCommands.slice(-50).reverse()
  })

  const commandSuggestions = computed(() => {
    return (partial: string) => {
      const allCommands = Array.from(commandHistory.value.values()).flat()
      return allCommands
        .filter(cmd => cmd.command.startsWith(partial))
        .map(cmd => cmd.command)
        .filter((cmd, index, arr) => arr.indexOf(cmd) === index)
        .slice(0, 10)
    }
  })

  // Actions
  const createSession = async (title?: string, shell?: string, workingDirectory?: string): Promise<string | null> => {
    try {
      setLoading(true)
      clearError()

      // Create terminal using Wails bindings
      const cols = settings.value.defaultCols
      const rows = settings.value.defaultRows
      const terminalData = await CreateTerminal(cols, rows)

      if (!terminalData) {
        throw new Error('Failed to create terminal session')
      }

      const session: TerminalSession = {
        id: terminalData.id || `terminal-${Date.now()}`,
        title: title || `Terminal ${sessions.value.size + 1}`,
        shell: shell || config.value.shell,
        workingDirectory: workingDirectory || '/',
        columns: cols,
        rows: rows,
        isActive: false,
        createdAt: new Date(),
        lastActivity: new Date(),
      }

      addSession(session)
      return session.id

    } catch (error) {
      console.error('Failed to create terminal session:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      return null
    } finally {
      setLoading(false)
    }
  }

  const closeSession = async (sessionId: string): Promise<boolean> => {
    try {
      // Close terminal using Wails bindings
      await CloseTerminal(sessionId)
      removeSession(sessionId)
      return true

    } catch (error) {
      console.error('Failed to close terminal session:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      return false
    }
  }

  const executeCommand = async (command: string, sessionId?: string): Promise<void> => {
    const targetSessionId = sessionId || activeSessionId.value
    if (!targetSessionId) {
      setError('No active terminal session')
      return
    }

    try {
      const response = await fetch(`/api/terminal/sessions/${targetSessionId}/execute`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command }),
      })

      if (!response.ok) {
        throw new Error(`Command execution failed: ${response.status}`)
      }

      const result = await response.json()

      // Add command to history
      addCommand({
        id: result.commandId,
        sessionId: targetSessionId,
        command,
        status: 'running',
        timestamp: new Date(),
      })

      // Add initial output if available
      if (result.output) {
        addOutput(targetSessionId, {
          type: 'stdout',
          content: result.output,
          timestamp: new Date(),
        })
      }

    } catch (error) {
      console.error('Failed to execute command:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
    }
  }

  const sendInput = async (input: string, sessionId?: string): Promise<void> => {
    const targetSessionId = sessionId || activeSessionId.value
    if (!targetSessionId) {
      setError('No active terminal session')
      return
    }

    try {
      // Write to terminal using Wails bindings
      await WriteToTerminal(targetSessionId, input)

    } catch (error) {
      console.error('Failed to send input:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
    }
  }

  const interrupt = async (sessionId?: string): Promise<boolean> => {
    const targetSessionId = sessionId || activeSessionId.value
    if (!targetSessionId) return false

    try {
      const response = await fetch(`/api/terminal/sessions/${targetSessionId}/interrupt`, {
        method: 'POST',
      })

      if (!response.ok) {
        throw new Error(`Failed to interrupt: ${response.status}`)
      }

      // Add interrupt signal to output
      addOutput(targetSessionId, {
        type: 'signal',
        content: '^C',
        timestamp: new Date(),
      })

      return true

    } catch (error) {
      console.error('Failed to interrupt process:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      return false
    }
  }

  const resizeSession = async (columns: number, rows: number, sessionId?: string): Promise<boolean> => {
    const targetSessionId = sessionId || activeSessionId.value
    if (!targetSessionId) return false

    try {
      // Resize terminal using Wails bindings
      await ResizeTerminal(targetSessionId, columns, rows)
      return true

    } catch (error) {
      console.error('Failed to resize terminal session:', error)
      setError(error instanceof Error ? error.message : 'Unknown error')
      return false
    }
  }

  const setActiveSession = (sessionId: string) => {
    if (sessions.value.has(sessionId)) {
      // Deactivate all sessions
      sessions.value.forEach(session => {
        session.isActive = false
      })

      // Activate selected session
      const session = sessions.value.get(sessionId)
      if (session) {
        session.isActive = true
        activeSessionId.value = sessionId
      }
    }
  }

  const addSession = (session: TerminalSession) => {
    sessions.value.set(session.id, session)
    outputBuffer.value.set(session.id, [])

    // Set as active if no other session exists
    if (!activeSessionId.value) {
      setActiveSession(session.id)
    }
  }

  const removeSession = (sessionId: string) => {
    sessions.value.delete(sessionId)
    outputBuffer.value.delete(sessionId)

    // If this was the active session, select another one
    if (activeSessionId.value === sessionId) {
      const remainingSessions = Array.from(sessions.value.keys())
      activeSessionId.value = remainingSessions.length > 0 ? remainingSessions[0] : null
    }
  }

  const addOutput = (sessionId: string, output: TerminalOutput) => {
    const buffer = outputBuffer.value.get(sessionId) || []
    buffer.push(output)

    // Keep buffer size manageable
    if (buffer.length > config.value.scrollback) {
      buffer.splice(0, buffer.length - config.value.scrollback)
    }

    outputBuffer.value.set(sessionId, buffer)

    // Update session activity
    const session = sessions.value.get(sessionId)
    if (session) {
      session.lastActivity = new Date()
    }
  }

  const addCommand = (command: TerminalCommand) => {
    const sessionHistory = commandHistory.value.get(command.sessionId) || []
    sessionHistory.push(command)
    commandHistory.value.set(command.sessionId, sessionHistory)

    // Keep history size manageable
    if (sessionHistory.length > 1000) {
      sessionHistory.splice(0, sessionHistory.length - 1000)
    }
  }

  const updateCommandStatus = (commandId: string, status: 'success' | 'error', result?: any) => {
    for (const [sessionId, commands] of commandHistory.value.entries()) {
      const command = commands.find(cmd => cmd.id === commandId)
      if (command) {
        command.status = status
        command.result = result
        break
      }
    }
  }

  const clearOutput = (sessionId?: string) => {
    const targetSessionId = sessionId || activeSessionId.value
    if (targetSessionId) {
      outputBuffer.value.set(targetSessionId, [])
    }
  }

  const clearHistory = () => {
    commandHistory.value.clear()
  }

  const updateConfig = (newConfig: Partial<typeof config.value>) => {
    config.value = { ...config.value, ...newConfig }

    // Save to localStorage
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('adex-terminal-config', JSON.stringify(config.value))
    }
  }

  const loadConfig = () => {
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('adex-terminal-config')
      if (saved) {
        try {
          const configData = JSON.parse(saved)
          config.value = { ...config.value, ...configData }
        } catch (error) {
          console.error('Failed to load terminal config:', error)
        }
      }
    }
  }

  const setLoading = (loading: boolean) => {
    isLoading.value = loading
  }

  const setError = (error: string | null) => {
    lastError.value = error
  }

  const clearError = () => {
    lastError.value = null
  }

  // Initialize store
  const initialize = async (): Promise<void> => {
    loadConfig()

    // Create default session if none exist
    if (sessions.value.size === 0) {
      await createSession()
    }
  }

  // Reset store
  const reset = () => {
    sessions.value.clear()
    activeSessionId.value = null
    commandHistory.value.clear()
    outputBuffer.value.clear()
    lastError.value = null
    isLoading.value = false
  }

  return {
    // State (readonly for safety)
    sessions: readonly(sessions),
    tabs: readonly(tabs),
    activeSessionId: readonly(activeSessionId),
    commandHistory: readonly(commandHistory),
    outputBuffer: readonly(outputBuffer),
    availableThemes: readonly(availableThemes),
    currentTheme: readonly(currentTheme),
    settings: readonly(settings),
    config: readonly(config),
    isLoading: readonly(isLoading),
    lastError: readonly(lastError),

    // Computed properties
    activeSession,
    allSessions,
    sessionOutput,
    hasActiveSession,
    recentCommands,
    commandSuggestions,

    // Actions
    createSession,
    closeSession,
    executeCommand,
    sendInput,
    interrupt,
    resizeSession,
    setActiveSession,
    addSession,
    removeSession,
    addOutput,
    addCommand,
    updateCommandStatus,
    clearOutput,
    clearHistory,
    updateConfig,
    loadConfig,
    setLoading,
    setError,
    clearError,
    initialize,
    reset
  }
})
