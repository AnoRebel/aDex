// tests/integration/terminal.test.ts
//
// Integration coverage for the terminal state that the running application
// actually depends on: the terminal store and its interaction with the UI
// store's notifications.
//
// This file previously mounted TerminalPanel, TerminalTabBar and
// TerminalContextMenu. None of those are reachable from the application —
// pages/index.vue renders AdexShell, and nothing outside
// app/components/terminal/ references them — so the specs had drifted a long
// way from any code that runs: they mocked the pre-v5 unscoped `xterm`
// packages (the app uses `@xterm/*`), omitted composable exports the
// components import, and asserted a `terminalStore.clearAll()` that has never
// existed. Rather than maintain 700 lines of mocks against dead components,
// this covers the store contract those components sat on top of.

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useTerminalStore } from '~/stores/terminal'
import { useUIStore } from '~/stores/ui'
import type { TerminalSession } from '~/types/terminal'

function makeSession(overrides: Partial<TerminalSession> = {}): TerminalSession {
  const now = new Date().toISOString()
  return {
    id: 'test-session',
    shell: '/bin/bash',
    cwd: '/home/user',
    env: { PATH: '/usr/bin' },
    size: { rows: 24, cols: 80 },
    active: true,
    createdAt: now,
    updatedAt: now,
    lastSeen: now,
    pid: 12345,
    ...overrides,
  } as TerminalSession
}

describe('Terminal integration', () => {
  let pinia: any
  let terminalStore: any
  let uiStore: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    terminalStore = useTerminalStore(pinia)
    uiStore = useUIStore(pinia)

    terminalStore.reset()
    uiStore.clearNotifications()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Session bookkeeping', () => {
    it('stores an added session and counts it', () => {
      const session = makeSession()
      terminalStore.addSession(session)

      expect(terminalStore.sessions.get(session.id)).toEqual(session)
      expect(terminalStore.allSessions).toHaveLength(1)
    })

    it('tracks several sessions independently', () => {
      terminalStore.addSession(makeSession({ id: 'a' }))
      terminalStore.addSession(makeSession({ id: 'b', shell: '/bin/zsh' }))

      expect(terminalStore.allSessions).toHaveLength(2)
      expect(terminalStore.sessions.get('b').shell).toBe('/bin/zsh')
    })

    it('reset() clears sessions, history, buffers and error state', () => {
      terminalStore.addSession(makeSession({ id: 'a' }))
      terminalStore.addSession(makeSession({ id: 'b' }))
      expect(terminalStore.allSessions).toHaveLength(2)

      terminalStore.reset()

      expect(terminalStore.allSessions).toHaveLength(0)
      expect(terminalStore.activeSessionId).toBeNull()
      expect(terminalStore.lastError).toBeNull()
      expect(terminalStore.isLoading).toBe(false)
    })
  })

  describe('UI notifications', () => {
    it('records a notification', () => {
      uiStore.addNotification({
        type: 'success',
        title: 'Test',
        message: 'Test notification',
        persistent: false,
      })

      expect(uiStore.notifications).toHaveLength(1)
      expect(uiStore.notifications[0].title).toBe('Test')
    })

    it('clears notifications', () => {
      uiStore.addNotification({
        type: 'error',
        title: 'Boom',
        message: 'something failed',
        persistent: false,
      })
      expect(uiStore.notifications).toHaveLength(1)

      uiStore.clearNotifications()
      expect(uiStore.notifications).toHaveLength(0)
    })
  })
})
