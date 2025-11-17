import { ref, computed, watch } from 'vue'
import { useStorage } from '@vueuse/core'
import { useTimeoutFn } from '@vueuse/shared'
import { useTerminalStore } from '~/stores/terminal'
import { useTerminal } from '~/composables/useTerminal'
import { useNotificationStore } from '~/stores/notification'
import type { TerminalSession, TerminalTab, TerminalOptions } from '~/types/terminal'

export interface SessionManagerConfig {
  autoSave: boolean
  saveInterval: number // ms
  maxSessions: number
  maxTabsPerSession: number
  persistSessions: boolean
  restoreOnStartup: boolean
  cleanupInactiveSessions: boolean
  sessionTimeout: number // ms
  enableSessionGroups: boolean
  defaultProfile: string
}

export interface SessionGroup {
  id: string
  name: string
  description?: string
  sessionIds: string[]
  color?: string
  icon?: string
  createdAt: string
  updatedAt: string
}

export interface SessionState {
  sessions: TerminalSession[]
  tabs: TerminalTab[]
  activeSessionId: string | null
  groups: SessionGroup[]
  lastActivity: string
  version: string
}

export function useSessionManager(config: Partial<SessionManagerConfig> = {}) {
  const terminalStore = useTerminalStore()
  const { createSession, closeSession } = useTerminal()
  const notificationStore = useNotificationStore()

  // Configuration
  const managerConfig = ref<SessionManagerConfig>({
    autoSave: true,
    saveInterval: 30000, // 30 seconds
    maxSessions: 10,
    maxTabsPerSession: 5,
    persistSessions: true,
    restoreOnStartup: true,
    cleanupInactiveSessions: true,
    sessionTimeout: 30 * 60 * 1000, // 30 minutes
    enableSessionGroups: false,
    defaultProfile: 'default',
    ...config
  })

  // State
  const sessions = ref<TerminalSession[]>([])
  const tabs = ref<TerminalTab[]>([])
  const activeSessionId = ref<string | null>(null)
  const sessionGroups = ref<SessionGroup[]>([])
  const lastActivity = ref<string>(new Date().toISOString())

  // Local storage
  const sessionState = useStorage<SessionState>('terminal-session-state', {
    sessions: [],
    tabs: [],
    activeSessionId: null,
    groups: [],
    lastActivity: new Date().toISOString(),
    version: '1.0.0'
  })

  // Computed
  const activeSession = computed(() => {
    return sessions.value.find(session => session.id === activeSessionId.value) || null
  })

  const activeTabs = computed(() => {
    return tabs.value.filter(tab => tab.sessionId === activeSessionId.value)
  })

  const inactiveSessions = computed(() => {
    return sessions.value.filter(session => !session.active)
  })

  const sessionCount = computed(() => sessions.value.length)
  const tabCount = computed(() => tabs.value.length)

  const groupedSessions = computed(() => {
    if (!managerConfig.value.enableSessionGroups) {
      return [{ id: 'default', name: 'Default', sessionIds: sessions.value.map(s => s.id) }]
    }

    return sessionGroups.value.map(group => ({
      ...group,
      sessions: sessions.value.filter(session => group.sessionIds.includes(session.id))
    }))
  })

  // Methods
  const createNewSession = async (options?: TerminalOptions, profileId?: string): Promise<TerminalSession> => {
    try {
      const sessionOptions = {
        profile: profileId || managerConfig.value.defaultProfile,
        ...options
      }

      const session = await createSession(sessionOptions)

      // Create default tab for the session
      const tab: TerminalTab = {
        id: `tab-${session.id}`,
        sessionId: session.id,
        title: session.title || `Terminal ${session.id.substring(0, 8)}`,
        active: true,
        position: tabs.value.length,
        pinned: false,
        modified: false,
        lastActivity: session.lastSeen
      }

      // Add to state
      sessions.value.push(session)
      tabs.value.push(tab)
      activeSessionId.value = session.id
      lastActivity.value = new Date().toISOString()

      // Save state
      if (managerConfig.value.autoSave) {
        saveState()
      }

      // Show notification
      notificationStore.addNotification({
        type: 'success',
        title: 'Session Created',
        message: `New terminal session started in ${session.cwd}`,
        persistent: false
      })

      return session

    } catch (error) {
      console.error('Failed to create session:', error)
      notificationStore.addNotification({
        type: 'error',
        title: 'Session Creation Failed',
        message: error instanceof Error ? error.message : 'Unknown error',
        persistent: false
      })
      throw error
    }
  }

  const closeSessionById = async (sessionId: string): Promise<void> => {
    try {
      await closeSession(sessionId)

      // Remove from state
      sessions.value = sessions.value.filter(session => session.id !== sessionId)
      tabs.value = tabs.value.filter(tab => tab.sessionId !== sessionId)

      // Update active session if needed
      if (activeSessionId.value === sessionId) {
        const remainingSessions = sessions.value.filter(session => session.active)
        if (remainingSessions.length > 0) {
          activeSessionId.value = remainingSessions[0].id
          // Activate first tab of the new active session
          const firstTab = tabs.value.find(tab => tab.sessionId === activeSessionId.value)
          if (firstTab) {
            firstTab.active = true
          }
        } else {
          activeSessionId.value = null
        }
      }

      lastActivity.value = new Date().toISOString()

      // Save state
      if (managerConfig.value.autoSave) {
        saveState()
      }

      // Show notification
      notificationStore.addNotification({
        type: 'info',
        title: 'Session Closed',
        message: 'Terminal session ended',
        persistent: false
      })

    } catch (error) {
      console.error('Failed to close session:', error)
      notificationStore.addNotification({
        type: 'error',
        title: 'Session Close Failed',
        message: error instanceof Error ? error.message : 'Unknown error',
        persistent: false
      })
      throw error
    }
  }

  const activateSession = (sessionId: string): void => {
    const session = sessions.value.find(s => s.id === sessionId)
    if (!session) return

    // Update tab active states
    tabs.value.forEach(tab => {
      tab.active = tab.sessionId === sessionId
    })

    activeSessionId.value = sessionId
    session.lastSeen = new Date().toISOString()
    lastActivity.value = new Date().toISOString()

    if (managerConfig.value.autoSave) {
      saveState()
    }
  }

  const createTab = (sessionId: string, title?: string): TerminalTab => {
    const session = sessions.value.find(s => s.id === sessionId)
    if (!session) {
      throw new Error(`Session ${sessionId} not found`)
    }

    const sessionTabs = tabs.value.filter(tab => tab.sessionId === sessionId)
    if (sessionTabs.length >= managerConfig.value.maxTabsPerSession) {
      throw new Error('Maximum tabs per session reached')
    }

    const tab: TerminalTab = {
      id: `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      sessionId,
      title: title || `Tab ${sessionTabs.length + 1}`,
      active: true,
      position: tabs.value.length,
      pinned: false,
      modified: false,
      lastActivity: new Date().toISOString()
    }

    // Deactivate other tabs for this session
    tabs.value.forEach(t => {
      if (t.sessionId === sessionId) {
        t.active = false
      }
    })

    tabs.value.push(tab)
    lastActivity.value = new Date().toISOString()

    if (managerConfig.value.autoSave) {
      saveState()
    }

    return tab
  }

  const closeTab = (tabId: string): void => {
    const tabIndex = tabs.value.findIndex(tab => tab.id === tabId)
    if (tabIndex === -1) return

    const tab = tabs.value[tabIndex]
    const sessionTabs = tabs.value.filter(t => t.sessionId === tab.sessionId)

    if (sessionTabs.length <= 1) {
      // This is the last tab, close the session
      closeSessionById(tab.sessionId)
      return
    }

    // Remove the tab
    tabs.value.splice(tabIndex, 1)

    // If this was the active tab, activate another tab for the same session
    if (tab.active) {
      const remainingTabs = tabs.value.filter(t => t.sessionId === tab.sessionId)
      if (remainingTabs.length > 0) {
        remainingTabs[0].active = true
      }
    }

    lastActivity.value = new Date().toISOString()

    if (managerConfig.value.autoSave) {
      saveState()
    }
  }

  const updateTab = (tabId: string, updates: Partial<TerminalTab>): void => {
    const tab = tabs.value.find(t => t.id === tabId)
    if (!tab) return

    Object.assign(tab, updates)
    tab.lastActivity = new Date().toISOString()

    if (managerConfig.value.autoSave) {
      saveState()
    }
  }

  const pinTab = (tabId: string): void => {
    updateTab(tabId, { pinned: true })
  }

  const unpinTab = (tabId: string): void => {
    updateTab(tabId, { pinned: false })
  }

  const duplicateTab = (tabId: string): TerminalTab => {
    const sourceTab = tabs.value.find(t => t.id === tabId)
    if (!sourceTab) {
      throw new Error(`Tab ${tabId} not found`)
    }

    return createTab(sourceTab.sessionId, `${sourceTab.title} (copy)`)
  }

  const reorderTabs = (fromIndex: number, toIndex: number): void => {
    const [movedTab] = tabs.value.splice(fromIndex, 1)
    if (movedTab) {
      tabs.value.splice(toIndex, 0, movedTab)

      // Update positions
      tabs.value.forEach((tab, index) => {
        tab.position = index
      })

      lastActivity.value = new Date().toISOString()

      if (managerConfig.value.autoSave) {
        saveState()
      }
    }
  }

  const createSessionGroup = (name: string, sessionIds: string[], options?: Partial<SessionGroup>): SessionGroup => {
    const group: SessionGroup = {
      id: `group-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      name,
      sessionIds,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      ...options
    }

    sessionGroups.value.push(group)

    if (managerConfig.value.autoSave) {
      saveState()
    }

    return group
  }

  const updateSessionGroup = (groupId: string, updates: Partial<SessionGroup>): void => {
    const group = sessionGroups.value.find(g => g.id === groupId)
    if (!group) return

    Object.assign(group, updates)
    group.updatedAt = new Date().toISOString()

    if (managerConfig.value.autoSave) {
      saveState()
    }
  }

  const deleteSessionGroup = (groupId: string): void => {
    sessionGroups.value = sessionGroups.value.filter(g => g.id !== groupId)

    if (managerConfig.value.autoSave) {
      saveState()
    }
  }

  const cleanupInactiveSessions = (): void => {
    const now = Date.now()
    const timeout = managerConfig.value.sessionTimeout

    const inactiveSessions = sessions.value.filter(session => {
      if (session.active) return false
      return now - new Date(session.lastSeen).getTime() > timeout
    })

    if (inactiveSessions.length === 0) return

    // Close inactive sessions
    inactiveSessions.forEach(session => {
      closeSessionById(session.id)
    })

    notificationStore.addNotification({
      type: 'info',
      title: 'Sessions Cleaned Up',
      message: `Closed ${inactiveSessions.length} inactive session${inactiveSessions.length > 1 ? 's' : ''}`,
      persistent: false
    })
  }

  const saveState = (): void => {
    if (!managerConfig.value.persistSessions) return

    try {
      const state: SessionState = {
        sessions: sessions.value,
        tabs: tabs.value,
        activeSessionId: activeSessionId.value,
        groups: sessionGroups.value,
        lastActivity: lastActivity.value,
        version: '1.0.0'
      }

      sessionState.value = state
    } catch (error) {
      console.error('Failed to save session state:', error)
    }
  }

  const loadState = (): void => {
    if (!managerConfig.value.persistSessions || !managerConfig.value.restoreOnStartup) return

    try {
      const state = sessionState.value

      if (state.version !== '1.0.0') {
        console.warn('Session state version mismatch, skipping restore')
        return
      }

      sessions.value = state.sessions || []
      tabs.value = state.tabs || []
      activeSessionId.value = state.activeSessionId
      sessionGroups.value = state.groups || []
      lastActivity.value = state.lastActivity || new Date().toISOString()

      // Validate state
      validateState()

    } catch (error) {
      console.error('Failed to load session state:', error)
      resetState()
    }
  }

  const validateState = (): void => {
    // Remove tabs that reference non-existent sessions
    const validSessionIds = new Set(sessions.value.map(s => s.id))
    tabs.value = tabs.value.filter(tab => validSessionIds.has(tab.sessionId))

    // Ensure active session exists
    if (activeSessionId.value && !sessions.value.find(s => s.id === activeSessionId.value)) {
      if (sessions.value.length > 0) {
        activeSessionId.value = sessions.value[0].id
      } else {
        activeSessionId.value = null
      }
    }

    // Ensure at least one tab is active per session
    sessions.value.forEach(session => {
      const sessionTabs = tabs.value.filter(tab => tab.sessionId === session.id)
      const activeTab = sessionTabs.find(tab => tab.active)
      if (!activeTab && sessionTabs.length > 0) {
        sessionTabs[0].active = true
      }
    })
  }

  const resetState = (): void => {
    sessions.value = []
    tabs.value = []
    activeSessionId.value = null
    sessionGroups.value = []
    lastActivity.value = new Date().toISOString()

    if (managerConfig.value.persistSessions) {
      saveState()
    }
  }

  const exportState = (): string => {
    const state = {
      sessions: sessions.value,
      tabs: tabs.value,
      activeSessionId: activeSessionId.value,
      groups: sessionGroups.value,
      lastActivity: lastActivity.value,
      config: managerConfig.value,
      exportedAt: new Date().toISOString(),
      version: '1.0.0'
    }

    return JSON.stringify(state, null, 2)
  }

  const importState = (jsonString: string): boolean => {
    try {
      const state = JSON.parse(jsonString)

      if (state.version !== '1.0.0') {
        throw new Error('Incompatible state version')
      }

      sessions.value = state.sessions || []
      tabs.value = state.tabs || []
      activeSessionId.value = state.activeSessionId
      sessionGroups.value = state.groups || []
      lastActivity.value = state.lastActivity || new Date().toISOString()

      // Update config if provided
      if (state.config) {
        managerConfig.value = { ...managerConfig.value, ...state.config }
      }

      validateState()
      saveState()

      notificationStore.addNotification({
        type: 'success',
        title: 'State Imported',
        message: 'Terminal session state imported successfully',
        persistent: false
      })

      return true

    } catch (error) {
      console.error('Failed to import state:', error)
      notificationStore.addNotification({
        type: 'error',
        title: 'Import Failed',
        message: error instanceof Error ? error.message : 'Unknown error',
        persistent: false
      })
      return false
    }
  }

  const getStats = () => {
    const now = Date.now()
    const oneHourAgo = now - 60 * 60 * 1000
    const oneDayAgo = now - 24 * 60 * 60 * 1000

    return {
      totalSessions: sessions.value.length,
      activeSessions: sessions.value.filter(s => s.active).length,
      totalTabs: tabs.value.length,
      activeTabs: tabs.value.filter(t => t.active).length,
      sessionsLastHour: sessions.value.filter(s => new Date(s.lastSeen).getTime() > oneHourAgo).length,
      sessionsLastDay: sessions.value.filter(s => new Date(s.lastSeen).getTime() > oneDayAgo).length,
      pinnedTabs: tabs.value.filter(t => t.pinned).length,
      modifiedTabs: tabs.value.filter(t => t.modified).length,
      lastActivity: lastActivity.value
    }
  }

  // Auto-save interval
  const { start: startAutoSave, stop: stopAutoSave } = useTimeoutFn(
    () => {
      if (managerConfig.value.autoSave) {
        saveState()
      }
      startAutoSave()
    },
    managerConfig.value.saveInterval,
    { immediate: false }
  )

  // Cleanup interval
  const { start: startCleanup, stop: stopCleanup } = useTimeoutFn(
    () => {
      if (managerConfig.value.cleanupInactiveSessions) {
        cleanupInactiveSessions()
      }
      startCleanup()
    },
    5 * 60 * 1000, // Every 5 minutes
    { immediate: false }
  )

  // Lifecycle
  loadState()

  if (managerConfig.value.autoSave) {
    startAutoSave()
  }

  if (managerConfig.value.cleanupInactiveSessions) {
    startCleanup()
  }

  return {
    // State
    sessions,
    tabs,
    activeSessionId,
    sessionGroups,
    managerConfig,

    // Computed
    activeSession,
    activeTabs,
    inactiveSessions,
    sessionCount,
    tabCount,
    groupedSessions,

    // Methods
    createNewSession,
    closeSessionById,
    activateSession,
    createTab,
    closeTab,
    updateTab,
    pinTab,
    unpinTab,
    duplicateTab,
    reorderTabs,
    createSessionGroup,
    updateSessionGroup,
    deleteSessionGroup,
    cleanupInactiveSessions,
    saveState,
    loadState,
    resetState,
    exportState,
    importState,
    getStats,

    // Controls
    startAutoSave,
    stopAutoSave,
    startCleanup,
    stopCleanup
  }
}

export default useSessionManager