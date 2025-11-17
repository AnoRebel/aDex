import { ref, computed, onMounted, onUnmounted, watch, readonly } from 'vue'
import { useWails, useEvents } from './useWails'
import { useUIStore } from '~/stores/ui'
import type { TerminalSession, TerminalCommand, TerminalOutput } from '~/types/terminal'
// import * as TerminalService from '../../bindings/aDex-UI/backend/services/coordinator/servicecoordinator'

export interface TerminalOptions {
  shell?: string
  cwd?: string
  env?: Record<string, string>
  rows?: number
  cols?: number
  profile?: string
  user?: string
}

export interface TerminalTheme {
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
  font: {
    family: string
    size: number
    weight: string
    lineHeight: number
    ligatures: boolean
  }
}

export interface TerminalTab {
  id: string
  sessionId: string
  title: string
  active: boolean
  icon?: string
  color?: string
  badge?: string
  position: number
  pinned: boolean
  modified: boolean
  lastActivity: string
}

export function useTerminal() {
  const wails = useWails()
  const events = useEvents()
  const uiStore = useUIStore()

  // State
  const sessions = ref<Map<string, TerminalSession>>(new Map())
  const activeSessionId = ref<string | null>(null)
  const tabs = ref<Map<string, TerminalTab>>(new Map())
  const outputBuffer = ref<Map<string, TerminalOutput[]>>(new Map())
  const commandHistory = ref<Map<string, TerminalCommand[]>>(new Map())
  const availableThemes = ref<TerminalTheme[]>([])
  const currentTheme = ref<TerminalTheme | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const commandInput = ref('')
  const isProcessing = ref(false)
  const maxHistory = 1000

  // Computed properties
  const activeSession = computed(() => {
    if (!activeSessionId.value) return null
    return sessions.value.get(activeSessionId.value) || null
  })

  const allSessions = computed(() => {
    return Array.from(sessions.value.values())
  })

  const activeTab = computed(() => {
    if (!activeSessionId.value) return null
    for (const tab of tabs.value.values()) {
      if (tab.sessionId === activeSessionId.value && tab.active) {
        return tab
      }
    }
    return null
  })

  const tabsArray = computed(() => {
    return Array.from(tabs.value.values()).sort((a, b) => a.position - b.position)
  })

  const sessionOutput = computed(() => {
    return (sessionId: string) => outputBuffer.value.get(sessionId) || []
  })

  const hasActiveSession = computed(() => !!activeSession.value)

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

  // Methods
  const createSession = async (shellPath?: string, width: number = 80, height: number = 24): Promise<string | null> => {
    if (!wails.isReady.value) {
      console.error('Wails not ready')
      return null
    }

    try {
      isLoading.value = true
      error.value = null
      isProcessing.value = true

      let sessionId: string | null = null

      // Try Wails service first
      if (wails.isReady.value) {
        try {
          const terminalData = await TerminalService.CreateTerminal(width, height)
          if (terminalData) {
            sessionId = terminalData.ID
          }
        } catch (wailsError) {
          console.warn('Wails terminal service failed, falling back:', wailsError)
        }
      }

      // Fallback to HTTP API if Wails service fails
      if (!sessionId) {
        try {
          const response = await fetch('/api/terminal/sessions', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              title: `Terminal ${sessions.value.size + 1}`,
              shell: shellPath || '/bin/bash',
              workingDirectory: '/',
            }),
          })

          if (response.ok) {
            const sessionData = await response.json()
            sessionId = sessionData.id
          }
        } catch (httpError) {
          console.warn('HTTP API fallback failed:', httpError)
        }
      }

      if (!sessionId) {
        throw new Error('Failed to create terminal session using both Wails and HTTP API')
      }

      // Create session object
      const session: TerminalSession = {
        id: sessionId,
        shell: shellPath || '/bin/bash',
        workingDirectory: '/',
        environment: {},
        isActive: true,
        createdAt: new Date(),
        lastActivity: new Date(),
        rows: height,
        cols: width,
        title: `Terminal ${sessionId.substring(0, 8)}`,
      }

      // Store session
      sessions.value.set(sessionId, session)
      outputBuffer.value.set(sessionId, [])
      commandHistory.value.set(sessionId, [])

      // Create tab
      const tab: TerminalTab = {
        id: `tab-${sessionId}`,
        sessionId: sessionId,
        title: session.title || `Terminal ${sessionId.substring(0, 8)}`,
        active: true,
        position: tabsArray.value.length,
        pinned: false,
        modified: false,
        lastActivity: session.lastActivity
      }

      tabs.value.set(tab.id, tab)
      activeSessionId.value = sessionId

      // Setup event listeners for this session
      setupSessionEvents(sessionId)

      // Show notification
      uiStore.addNotification({
        type: 'success',
        title: 'Terminal Created',
        message: `New terminal session started in ${session.workingDirectory}`,
        persistent: false
      })

      return sessionId
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to create terminal session'
      error.value = errorMessage
      console.error('Failed to create terminal session:', error)
      throw error
    } finally {
      isLoading.value = false
      isProcessing.value = false
    }
  }

  const closeSession = async (sessionId: string): Promise<boolean> => {
    try {
      let success = false

      // Try Wails service first
      if (wails.isReady.value) {
        try {
          await TerminalService.CloseTerminal(sessionId)
          success = true
        } catch (wailsError) {
          console.warn('Wails close service failed, trying HTTP API:', wailsError)
        }
      }

      // Fallback to HTTP API
      if (!success) {
        try {
          const response = await fetch(`/api/terminal/sessions/${sessionId}`, {
            method: 'DELETE',
          })
          success = response.ok
        } catch (httpError) {
          console.warn('HTTP API close failed:', httpError)
        }
      }

      if (success) {
        // Remove session
        sessions.value.delete(sessionId)
        outputBuffer.value.delete(sessionId)
        commandHistory.value.delete(sessionId)

        // Remove tab
        const tabToRemove = Array.from(tabs.value.values()).find(tab => tab.sessionId === sessionId)
        if (tabToRemove) {
          tabs.value.delete(tabToRemove.id)
        }

        // If this was the active session, activate another one
        if (activeSessionId.value === sessionId) {
          const remainingSessions = Array.from(sessions.value.values())
          if (remainingSessions.length > 0) {
            await setActiveSession(remainingSessions[0].id)
          } else {
            activeSessionId.value = null
          }
        }

        // Show notification
        uiStore.addNotification({
          type: 'info',
          title: 'Terminal Closed',
          message: `Terminal session ended`,
          persistent: false
        })

        return true
      }

      throw new Error('Failed to close terminal session')
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to close terminal session'
      error.value = errorMessage
      console.error('Failed to close terminal session:', error)
      return false
    }
  }

  const executeCommand = async (command: string, sessionId?: string): Promise<void> => {
    const targetSessionId = sessionId || activeSessionId.value
    if (!targetSessionId || !wails.isReady.value) {
      error.value = 'No active terminal session'
      return
    }

    try {
      isProcessing.value = true

      // Add command to history
      const cmd: TerminalCommand = {
        id: Date.now().toString(),
        command,
        sessionId: targetSessionId,
        timestamp: new Date(),
        status: 'running',
      }

      const sessionHistory = commandHistory.value.get(targetSessionId) || []
      sessionHistory.push(cmd)
      commandHistory.value.set(targetSessionId, sessionHistory)

      // Keep history size manageable
      if (sessionHistory.length > maxHistory) {
        sessionHistory.splice(0, sessionHistory.length - maxHistory)
      }

      // Try Wails service first
      let success = false
      if (wails.isReady.value) {
        try {
          const commandWithNewline = command + '\n'
          await TerminalService.WriteToTerminal(targetSessionId, commandWithNewline)
          success = true
          cmd.status = 'success'
        } catch (wailsError) {
          console.warn('Wails execute service failed, trying HTTP API:', wailsError)
        }
      }

      // Fallback to HTTP API
      if (!success) {
        try {
          const response = await fetch(`/api/terminal/sessions/${targetSessionId}/execute`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ command }),
          })

          if (response.ok) {
            const result = await response.json()
            success = true
            cmd.status = 'success'

            // Add initial output if available
            if (result.output) {
              addOutput(targetSessionId, {
                type: 'stdout',
                content: result.output,
                timestamp: new Date(),
              })
            }
          }
        } catch (httpError) {
          console.warn('HTTP API execute failed:', httpError)
        }
      }

      if (!success) {
        cmd.status = 'error'
        throw new Error('Command execution failed')
      }

      // Update command status
      const updatedHistory = commandHistory.value.get(targetSessionId) || []
      const commandIndex = updatedHistory.findIndex(c => c.id === cmd.id)
      if (commandIndex !== -1) {
        updatedHistory[commandIndex] = cmd
      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Command execution failed'
      error.value = errorMessage
      console.error('Failed to execute command:', error)
      addOutput(targetSessionId!, {
        type: 'error',
        content: errorMessage,
        timestamp: new Date(),
      })
    } finally {
      isProcessing.value = false
    }
  }

  const writeToSession = async (sessionId: string, data: string): Promise<void> => {
    if (!wails.isReady.value) {
      throw new Error('Wails not ready')
    }

    try {
      await TerminalService.WriteToTerminal(sessionId, data)
    } catch (error) {
      console.error('Failed to write to terminal:', error)
      throw error
    }
  }

  const resizeSession = async (sessionId: string, rows: number, cols: number): Promise<boolean> => {
    try {
      let success = false

      // Try Wails service first
      if (wails.isReady.value) {
        try {
          await TerminalService.ResizeTerminal(sessionId, cols, rows)
          success = true
        } catch (wailsError) {
          console.warn('Wails resize service failed, trying HTTP API:', wailsError)
        }
      }

      // Fallback to HTTP API
      if (!success) {
        try {
          const response = await fetch(`/api/terminal/sessions/${sessionId}/resize`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ columns: cols, rows }),
          })
          success = response.ok
        } catch (httpError) {
          console.warn('HTTP API resize failed:', httpError)
        }
      }

      if (success) {
        // Update session dimensions
        const session = sessions.value.get(sessionId)
        if (session) {
          session.rows = rows
          session.cols = cols
          session.lastActivity = new Date()
        }
        return true
      }

      throw new Error('Failed to resize terminal')
    } catch (error) {
      console.error('Failed to resize session:', error)
      return false
    }
  }

  const setActiveSession = (sessionId: string) => {
    if (sessions.value.has(sessionId)) {
      // Deactivate all sessions
      sessions.value.forEach(session => {
        session.isActive = false
      })

      // Update tab states
      for (const tab of tabs.value.values()) {
        tab.active = tab.sessionId === sessionId
      }

      // Activate selected session
      const session = sessions.value.get(sessionId)
      if (session) {
        session.isActive = true
        session.lastActivity = new Date()
        activeSessionId.value = sessionId
      }
    }
  }

  const addOutput = (sessionId: string, output: TerminalOutput) => {
    const buffer = outputBuffer.value.get(sessionId) || []
    buffer.push(output)

    // Keep buffer size manageable
    if (buffer.length > 1000) {
      buffer.splice(0, buffer.length - 1000)
    }

    outputBuffer.value.set(sessionId, buffer)

    // Update session activity
    const session = sessions.value.get(sessionId)
    if (session) {
      session.lastActivity = new Date()
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

  const updateTabTitle = (sessionId: string, title: string): void => {
    const tab = Array.from(tabs.value.values()).find(t => t.sessionId === sessionId)
    if (tab) {
      tab.title = title
      tab.lastActivity = new Date().toISOString()
    }

    const session = sessions.value.get(sessionId)
    if (session) {
      session.title = title
      session.lastActivity = new Date()
    }
  }

  // Setup event listeners for a session
  const setupSessionEvents = (sessionId: string) => {
    // Wails event listeners
    if (wails.isReady.value) {
      wails.on(`terminal.output.${sessionId}`, (data) => {
        if (data && data.content) {
          addOutput(sessionId, {
            type: 'stdout',
            content: data.content,
            timestamp: new Date(),
          })
        }
      })

      wails.on(`terminal.error.${sessionId}`, (data) => {
        if (data && data.error) {
          addOutput(sessionId, {
            type: 'error',
            content: data.error,
            timestamp: new Date(),
          })
        }
      })

      wails.on(`terminal.closed.${sessionId}`, () => {
        closeSession(sessionId)
      })
    }

    // Global event listeners
    if (events) {
      events.subscribe('terminal.output', (event) => {
        if (event.Data.session_id === sessionId) {
          addOutput(sessionId, {
            type: 'stdout',
            content: event.Data.data,
            timestamp: new Date(),
          })
        }
      })

      events.subscribe('terminal.error', (event) => {
        if (event.Data.session_id === sessionId) {
          addOutput(sessionId, {
            type: 'error',
            content: event.Data.error,
            timestamp: new Date(),
          })
        }
      })

      events.subscribe('terminal.closed', (event) => {
        if (event.Data.session_id === sessionId) {
          closeSession(sessionId)
        }
      })
    }
  }

  // Theme management
  const loadThemes = async (): Promise<void> => {
    if (!wails.isReady.value) return

    try {
      const themes = await wails.callService<TerminalTheme[]>('theme', 'GetThemes')
      availableThemes.value = themes

      // Set default theme
      if (themes.length > 0 && !currentTheme.value) {
        currentTheme.value = themes[0]
      }
    } catch (err) {
      console.error('Failed to load terminal themes:', err)
    }
  }

  const setTheme = async (themeId: string): Promise<void> => {
    if (!wails.isReady.value) return

    try {
      const theme = availableThemes.value.find(t => t.id === themeId)
      if (!theme) {
        throw new Error(`Theme not found: ${themeId}`)
      }

      await wails.callService('theme', 'SetTheme', themeId)
      currentTheme.value = theme
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to set theme'
      throw err
    }
  }

  // Initialize
  onMounted(async () => {
    // Load existing sessions if any
    if (wails.isReady.value) {
      try {
        // Create default session if none exist
        if (sessions.value.size === 0) {
          await createSession()
        }
      } catch (error) {
        console.error('Failed to initialize terminal:', error)
      }

      // Load themes
      await loadThemes()
    }

    // Setup global event listeners
    if (events) {
      events.subscribe('terminal.created', (event) => {
        console.log('Terminal created:', event.Data)
      })

      events.subscribe('terminal.error', (event) => {
        error.value = event.Data.error
        uiStore.addNotification({
          type: 'error',
          title: 'Terminal Error',
          message: event.Data.error,
          persistent: false
        })
      })
    }
  })

  onUnmounted(() => {
    // Cleanup
  })

  return {
    // State (readonly for safety)
    sessions: readonly(sessions),
    activeSession,
    activeSessionId: readonly(activeSessionId),
    tabs: readonly(tabs),
    activeTab,
    allSessions,
    tabsArray,
    outputBuffer: readonly(outputBuffer),
    commandHistory: readonly(commandHistory),
    availableThemes: readonly(availableThemes),
    currentTheme: readonly(currentTheme),
    isLoading: readonly(isLoading),
    error: readonly(error),
    commandInput,
    isProcessing: readonly(isProcessing),
    hasActiveSession,
    sessionOutput,
    recentCommands,
    commandSuggestions,

    // Methods
    createSession,
    closeSession,
    executeCommand,
    writeToSession,
    resizeSession,
    setActiveSession,
    addOutput,
    clearOutput,
    clearHistory,
    updateTabTitle,
    loadThemes,
    setTheme
  }
}

/**
 * useTerminalTab composable for managing individual terminal tabs
 */
export function useTerminalTab(sessionId: string) {
  const terminal = useTerminal()

  const tab = computed(() => {
    return terminal.tabsArray.value.find(tab => tab.sessionId === sessionId)
  })

  const session = computed(() => {
    return terminal.sessions.value.get(sessionId)
  })

  const isActive = computed(() => {
    return tab.value?.active || false
  })

  const output = computed(() => {
    return terminal.sessionOutput(sessionId)
  })

  const history = computed(() => {
    const allCommands = Array.from(terminal.commandHistory.value.values()).flat()
    return allCommands.filter(cmd => cmd.sessionId === sessionId)
  })

  const write = async (data: string) => {
    await terminal.writeToSession(sessionId, data)
  }

  const resize = async (rows: number, cols: number) => {
    await terminal.resizeSession(sessionId, rows, cols)
  }

  const activate = async () => {
    await terminal.setActiveSession(sessionId)
  }

  const close = async () => {
    await terminal.closeSession(sessionId)
  }

  const clearOutput = () => {
    terminal.clearOutput(sessionId)
  }

  const setTitle = (title: string) => {
    terminal.updateTabTitle(sessionId, title)
  }

  const pin = () => {
    if (tab.value) {
      tab.value.pinned = !tab.value.pinned
    }
  }

  return {
    tab,
    session,
    isActive,
    output,
    history,
    write,
    resize,
    activate,
    close,
    clearOutput,
    setTitle,
    pin
  }
}

export default useTerminal