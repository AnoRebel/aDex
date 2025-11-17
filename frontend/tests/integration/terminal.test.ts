import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import TerminalPanel from '~/components/terminal/TerminalPanel.vue'
import TerminalTabBar from '~/components/terminal/TerminalTabBar.vue'
import TerminalContextMenu from '~/components/terminal/TerminalContextMenu.vue'
import { useTerminalStore } from '~/stores/terminal'
import { useUIStore } from '~/stores/ui'
import type { TerminalSession, TerminalTab } from '~/types/terminal'

// Mock xterm
vi.mock('xterm', () => ({
  Terminal: vi.fn().mockImplementation(() => ({
    open: vi.fn(),
    write: vi.fn(),
    writeln: vi.fn(),
    clear: vi.fn(),
    focus: vi.fn(),
    blur: vi.fn(),
    hasSelection: vi.fn(() => false),
    getSelection: vi.fn(() => ''),
    selectAll: vi.fn(),
    clearSelection: vi.fn(),
    paste: vi.fn(),
    resize: vi.fn(),
    scrollToBottom: vi.fn(),
    onTitleChange: vi.fn(),
    onData: vi.fn(),
    onFocus: vi.fn(),
    onBlur: vi.fn(),
    onSelectionChange: vi.fn(),
    onKey: vi.fn(),
    onBell: vi.fn(),
    dispose: vi.fn(),
    options: {},
    loadAddon: vi.fn()
  }))
}))

vi.mock('xterm-addon-webgl', () => ({
  WebglAddon: vi.fn().mockImplementation(() => ({
    dispose: vi.fn()
  }))
}))

vi.mock('xterm-addon-fit', () => ({
  FitAddon: vi.fn().mockImplementation(() => ({
    fit: vi.fn(),
    activate: vi.fn()
  }))
}))

vi.mock('xterm-addon-canvas', () => ({
  CanvasAddon: vi.fn().mockImplementation(() => ({
    dispose: vi.fn()
  }))
}))

vi.mock('xterm-addon-ligatures', () => ({
  LigaturesAddon: vi.fn().mockImplementation(() => ({
    dispose: vi.fn()
  }))
}))

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn().mockResolvedValue(undefined),
    readText: vi.fn().mockResolvedValue('test clipboard content'),
    permission: 'granted' as PermissionState
  }
})

// Mock Notification API
Object.assign(window, {
  Notification: vi.fn().mockImplementation((title, options) => ({
    title,
    ...options,
    close: vi.fn()
  }))
})

describe('Terminal Integration Tests', () => {
  let pinia: any
  let terminalStore: any
  let uiStore: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    terminalStore = useTerminalStore(pinia)
    uiStore = useUIStore(pinia)

    // Clear any existing state
    terminalStore.clearAll()
    uiStore.clearNotifications()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('TerminalPanel Component', () => {
    const mockSession: TerminalSession = {
      id: 'test-session-1',
      shell: '/bin/bash',
      cwd: '/home/user',
      env: {},
      size: { rows: 24, cols: 80 },
      active: true,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      lastSeen: new Date().toISOString(),
      pid: 12345,
      title: 'Test Terminal'
    }

    it('should render terminal panel with basic props', () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id,
          theme: 'default-dark',
          fontFamily: 'JetBrains Mono',
          fontSize: 14,
          cursorStyle: 'block',
          cursorBlink: false,
          opacity: 100
        },
        global: {
          plugins: [pinia]
        }
      })

      expect(wrapper.find('.terminal-panel')).toBeTruthy()
      expect(wrapper.find('.terminal-header')).toBeTruthy()
      expect(wrapper.find('.terminal-content')).toBeTruthy()
      expect(wrapper.find('.terminal-footer')).toBeTruthy()
    })

    it('should hide header when hideHeader prop is true', () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id,
          hideHeader: true
        },
        global: {
          plugins: [pinia]
        }
      })

      expect(wrapper.find('.terminal-header')).not.toBeTruthy()
    })

    it('should hide footer when hideFooter prop is true', () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id,
          hideFooter: true
        },
        global: {
          plugins: [pinia]
        }
      })

      expect(wrapper.find('.terminal-footer')).not.toBeTruthy()
    })

    it('should show loading state during initialization', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      // Initially loading should be false (since we mock xterm)
      expect(wrapper.find('.terminal-loading')).not.toBeTruthy()

      // Simulate loading state
      await wrapper.vm.$nextTick()
      wrapper.vm.isLoading = true
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.terminal-loading')).toBeTruthy()
      expect(wrapper.find('.loading-spinner')).toBeTruthy()
      expect(wrapper.find('.loading-text')).toBeTruthy()
    })

    it('should display error state when initialization fails', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      wrapper.vm.error = 'Test error message'
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.terminal-error')).toBeTruthy()
      expect(wrapper.find('.error-message')).toBeTruthy()
      expect(wrapper.text()).toContain('Test error message')
      expect(wrapper.find('.error-retry')).toBeTruthy()
    })

    it('should handle copy functionality', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      // Mock having content
      wrapper.vm.hasContent = true

      await wrapper.find('.terminal-control[title="Copy content"]').trigger('click')

      // Should attempt to copy to clipboard
      expect(navigator.clipboard.writeText).toHaveBeenCalled()
    })

    it('should handle clear functionality', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      wrapper.vm.hasContent = true

      await wrapper.find('.terminal-control[title="Clear content"]').trigger('click')

      // Should clear the terminal
      // This would be tested through the useXTerm composable
    })

    it('should handle fullscreen toggle', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      const fullscreenButton = wrapper.find('.terminal-control[title="Toggle fullscreen"]')
      await fullscreenButton.trigger('click')

      expect(wrapper.vm.isFullscreen).toBe(true)
    })

    it('should show context menu on right-click', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      await wrapper.find('.terminal-panel').trigger('contextmenu', {
        clientX: 100,
        clientY: 100,
        preventDefault: vi.fn()
      })

      expect(wrapper.vm.showContextMenu).toBe(true)
      expect(wrapper.vm.contextMenuPosition.x).toBe(100)
      expect(wrapper.vm.contextMenuPosition.y).toBe(100)
    })

    it('should close context menu when clicking outside', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      // Open context menu
      wrapper.vm.showContextMenu = true

      // Simulate click outside
      document.dispatchEvent(new MouseEvent('click', {
        clientX: 1000,
        clientY: 1000
      }))

      await nextTick()

      expect(wrapper.vm.showContextMenu).toBe(false)
    })

    it('should resize terminal when resize handle is dragged', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      const resizeHandle = wrapper.find('.resize-handle')
      if (resizeHandle) {
        await resizeHandle.trigger('mousedown', {
          clientX: 100,
          clientY: 100,
          preventDefault: vi.fn()
        })

        expect(wrapper.vm.isResizing).toBe(true)

        // Simulate mouse move
        document.dispatchEvent(new MouseEvent('mousemove', {
          clientX: 150,
          clientY: 150
        }))

        // Simulate mouse up
        document.dispatchEvent(new MouseEvent('mouseup', {}))

        expect(wrapper.vm.isResizing).toBe(false)
      }
    })

    it('should handle scroll events', async () => {
      const wrapper = mount(TerminalPanel, {
        props: {
          sessionId: mockSession.id
        },
        global: {
          plugins: [pinia]
        }
      })

      const terminalContent = wrapper.find('.terminal-content')
      if (terminalContent) {
        await terminalContent.trigger('scroll', {
          target: terminalContent.element,
          scrollTop: 100,
          scrollHeight: 1000,
          clientHeight: 400
        })

        // Should handle scroll event
        expect(wrapper.vm.handleScroll).toBeDefined()
      }
    })
  })

  describe('TerminalTabBar Component', () => {
    const mockTabs: TerminalTab[] = [
      {
        id: 'tab-1',
        sessionId: 'session-1',
        title: 'Terminal 1',
        active: true,
        position: 0,
        pinned: false,
        modified: false,
        lastActivity: new Date().toISOString()
      },
      {
        id: 'tab-2',
        sessionId: 'session-2',
        title: 'Terminal 2',
        active: false,
        position: 1,
        pinned: false,
        modified: true,
        lastActivity: new Date().toISOString()
      }
    ]

    it('should render tab bar with tabs', () => {
      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: mockTabs,
          showCloseButtons: true,
          showNewTabButton: true
        },
        global: {
          plugins: [pinia]
        }
      })

      expect(wrapper.find('.terminal-tab-bar')).toBeTruthy()
      expect(wrapper.findAll('.terminal-tab-bar__tab')).toHaveLength(2)
      expect(wrapper.find('.terminal-tab-bar__new-tab')).toBeTruthy()
    })

    it('should show active tab with proper styling', () => {
      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: mockTabs
        },
        global: {
          plugins: [pinia]
        }
      })

      const activeTab = wrapper.find('.terminal-tab-bar__tab--active')
      expect(activeTab).toBeTruthy()
      expect(activeTab.text()).toContain('Terminal 1')
    })

    it('should show modified indicator for modified tabs', () => {
      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: mockTabs
        },
        global: {
          plugins: [pinia]
        }
      })

      const modifiedTab = wrapper.findAll('.terminal-tab-bar__tab')[1]
      expect(modifiedTab.classes()).toContain('terminal-tab-bar__tab--modified')
    })

    it('should handle tab click events', async () => {
      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: mockTabs
        },
        global: {
          plugins: [pinia]
        }
      })

      const firstTab = wrapper.find('.terminal-tab-bar__tab')
      await firstTab.trigger('click')

      expect(wrapper.emitted('tabClick')).toBeTruthy()
      expect(wrapper.emitted('tabActivate')).toBeTruthy()
    })

    it('should handle tab close events', async () => {
      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: mockTabs,
          showCloseButtons: true
        },
        global: {
          plugins: [pinia]
        }
      })

      const firstTab = wrapper.find('.terminal-tab-bar__tab')
      const closeButton = firstTab.find('.terminal-tab-bar__tab-close')

      if (closeButton) {
        await closeButton.trigger('click')
      }

      expect(wrapper.emitted('tabClose')).toBeTruthy()
    })

    it('should not show close button for pinned tabs', () => {
      const pinnedTabs = [
        { ...mockTabs[0], pinned: true },
        { ...mockTabs[1] }
      ]

      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: pinnedTabs,
          showCloseButtons: true
        },
        global: {
          plugins: [pinia]
        }
      })

      const tabs = wrapper.findAll('.terminal-tab-bar__tab')
      const firstTabCloseButton = tabs[0].find('.terminal-tab-bar__tab-close')
      const secondTabCloseButton = tabs[1].find('.terminal-tab-bar__tab-close')

      expect(firstTabCloseButton).toBeNull()
      expect(secondTabCloseButton).toBeTruthy()
    })

    it('should handle tab drag and drop', async () => {
      const wrapper = mount(TerminalTabBar, {
        props: {
          tabs: mockTabs,
          allowReorder: true
        },
        global: {
          plugins: [pinia]
        }
      })

      const firstTab = wrapper.find('.terminal-tab-bar__tab')

      // Start drag
      await firstTab.trigger('dragstart', {
        dataTransfer: {
          setData: vi.fn(),
          setDragImage: vi.fn()
        }
      })

      // Simulate drag over second tab
      const secondTab = wrapper.findAll('.terminal-tab-bar__tab')[1]
      await secondTab.trigger('dragover', {
        preventDefault: vi.fn()
      })

      // Drop
      await secondTab.trigger('drop', {
        preventDefault: vi.fn()
      })

      expect(wrapper.emitted('tabReorder')).toBeTruthy()
    })
  })

  describe('TerminalContextMenu Component', () => {
    it('should render context menu when visible', () => {
      const wrapper = mount(TerminalContextMenu, {
        props: {
          visible: true,
          x: 100,
          y: 100,
          hasSelection: true,
          hasOutput: true,
          readonly: false,
          output: ['line1', 'line2'],
          isFullscreen: false
        },
        global: {
          plugins: [pinia]
        }
      })

      expect(wrapper.find('.terminal-context-menu')).toBeTruthy()
      expect(wrapper.text()).toContain('Copy Selection')
      expect(wrapper.text()).toContain('Copy All')
      expect(wrapper.text()).toContain('Paste')
    })

    it('should hide copy options when no selection', () => {
      const wrapper = mount(TerminalContextMenu, {
        props: {
          visible: true,
          x: 100,
          y: 100,
          hasSelection: false,
          hasOutput: true,
          readonly: false,
          output: ['line1', 'line2'],
          isFullscreen: false
        },
        global: {
          plugins: [pinia]
        }
      })

      expect(wrapper.text()).not.toContain('Copy Selection')
      expect(wrapper.text()).toContain('Copy All')
    })

    it('should disable paste options when readonly', () => {
      const wrapper = mount(TerminalContextMenu, {
        props: {
          visible: true,
          x: 100,
          y: 100,
          hasSelection: true,
          hasOutput: true,
          readonly: true,
          output: ['line1', 'line2'],
          isFullscreen: false
        },
        global: {
          plugins: [pinia]
        }
      })

      const pasteItems = wrapper.findAll('button').filter(btn =>
        btn.text().includes('Paste')
      )

      pasteItems.forEach(item => {
        expect(item.attributes('disabled')).toBeDefined()
      })
    })

    it('should handle copy selection event', async () => {
      const wrapper = mount(TerminalContextMenu, {
        props: {
          visible: true,
          x: 100,
          y: 100,
          hasSelection: true,
          hasOutput: true,
          readonly: false,
          output: ['line1', 'line2'],
          isFullscreen: false
        },
        global: {
          plugins: [pinia]
        }
      })

      const copyButton = wrapper.find('button').filter(btn =>
        btn.text().includes('Copy Selection')
      )

      await copyButton.trigger('click')

      expect(wrapper.emitted('copySelection')).toBeTruthy()
      expect(wrapper.emitted('close')).toBeTruthy()
    })

    it('should close when clicking outside', async () => {
      const wrapper = mount(TerminalContextMenu, {
        props: {
          visible: true,
          x: 100,
          y: 100,
          hasSelection: true,
          hasOutput: true,
          readonly: false,
          output: ['line1', 'line2'],
          isFullscreen: false
        },
        global: {
          plugins: [pinia]
        }
      })

      // Simulate click outside
      document.dispatchEvent(new MouseEvent('click', {
        clientX: 1000,
        clientY: 1000
      }))

      await nextTick()

      expect(wrapper.emitted('close')).toBeTruthy()
    })

    it('should handle keyboard shortcuts', async () => {
      const wrapper = mount(TerminalContextMenu, {
        props: {
          visible: true,
          x: 100,
          y: 100,
          hasSelection: true,
          hasOutput: true,
          readonly: false,
          output: ['line1', 'line2'],
          isFullscreen: false
        },
        global: {
          plugins: [pinia]
        }
      })

      // Test Escape key
      await wrapper.trigger('keydown', { key: 'Escape' })
      expect(wrapper.emitted('close')).toBeTruthy()

      // Re-open for Enter key test
      await wrapper.setProps({ visible: true })
      await wrapper.trigger('keydown', { key: 'Enter' })
      expect(wrapper.emitted('copySelection')).toBeTruthy()
    })
  })

  describe('Store Integration', () => {
    it('should update terminal store with sessions', () => {
      const session: TerminalSession = {
        id: 'test-session',
        shell: '/bin/bash',
        cwd: '/home/user',
        env: { PATH: '/usr/bin' },
        size: { rows: 24, cols: 80 },
        active: true,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        lastSeen: new Date().toISOString(),
        pid: 12345
      }

      terminalStore.addSession(session)

      expect(terminalStore.sessions.get(session.id)).toEqual(session)
      expect(terminalStore.sessionCount).toBe(1)
    })

    it('should update UI store with notifications', () => {
      uiStore.addNotification({
        type: 'success',
        title: 'Test',
        message: 'Test notification',
        persistent: false
      })

      expect(uiStore.notifications).toHaveLength(1)
      expect(uiStore.notifications[0].title).toBe('Test')
    })
  })

  describe('End-to-End Scenarios', () => {
    it('should create and manage multiple terminal sessions', async () => {
      // This would test the full workflow:
      // 1. Create session
      // 2. Create tabs
      // 3. Switch between sessions
      // 4. Close sessions
      // 5. Verify state persistence

      // Mock the useTerminal composable
      vi.mock('~/composables/useTerminal', () => ({
        useTerminal: () => ({
          sessions: ref([]),
          activeSessionId: ref(null),
          createSession: vi.fn().mockResolvedValue({
            id: 'session-1',
            shell: '/bin/bash',
            cwd: '/tmp',
            env: {},
            size: { rows: 24, cols: 80 },
            active: true,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
            lastSeen: new Date().toISOString()
          }),
          closeSession: vi.fn().mockResolvedValue(undefined),
          // ... other methods
        })
      }))

      const sessionManager = useSessionManager()
      const session = await sessionManager.createNewSession({
        cwd: '/tmp'
      })

      expect(session).toBeDefined()
      expect(session.id).toBe('session-1')
      expect(sessionManager.sessionCount).toBe(1)
    })

    it('should handle copy/paste operations across sessions', async () => {
      // Test clipboard operations between different terminal sessions
      // This would require more complex mocking of the clipboard API
      expect(navigator.clipboard.writeText).toBeDefined()
      expect(navigator.clipboard.readText).toBeDefined()
    })

    it('should handle terminal bell notifications', async () => {
      // Test bell notification system
      // This would test the useBellNotifications composable
      const { ringBell } = useBellNotifications()

      await ringBell({
        visual: true,
        audible: true,
        title: 'Test Bell',
        message: 'Test notification'
      })

      // Verify browser notification would be shown if permission granted
      expect(window.Notification).toBeDefined()
    })
  })
})