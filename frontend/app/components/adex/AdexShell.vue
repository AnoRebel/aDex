<template>
  <div class="main-shell" ref="shellRef">
    <!-- Tab Bar -->
    <div class="shell-tabs">
      <div
        v-for="(tab, index) in tabs"
        :key="tab.id"
        class="shell-tab"
        :class="{ active: tab.active }"
        @click="activateTab(index)"
        @contextmenu.prevent="onTabContextMenu(index, $event)"
      >
        <span>{{ tab.title }}</span>
      </div>
    </div>

    <!-- Terminal Containers -->
    <div class="shell-terminal" ref="terminalAreaRef">
      <div
        v-for="(tab, index) in tabs"
        :key="'term-' + tab.id"
        :ref="(el) => setTerminalRef(index, el as HTMLElement)"
        class="shell-terminal-pane"
        :class="{ visible: tab.active }"
      />
    </div>

    <!-- Status Bar -->
    <div class="shell-status">
      <div class="shell-status-left">
        <span class="status-process">{{ activeProcessInfo }}</span>
        <span class="status-separator">|</span>
        <span class="status-session">{{ activeSessionLabel }}</span>
      </div>
      <div class="shell-status-right">
        <span class="status-ping">{{ pingDisplay }}</span>
        <span class="status-separator">|</span>
        <span class="status-datetime">{{ dateTimeDisplay }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  reactive,
  computed,
  onMounted,
  onBeforeUnmount,
  nextTick,
  watch,
  shallowRef,
} from 'vue'
import { useTerminalStore } from '~/stores/terminal'
import { Events } from '~/lib/wailsjs/runtime'
import {
import { nowMs } from '~/utils/now'
  CreateTerminal,
  WriteToTerminal,
  ResizeTerminal,
  CloseTerminal,
} from '~/lib/wailsjs/coordinator'

// ---- Types ----

interface ShellTab {
  id: string
  index: number
  sessionId: string | null
  title: string
  active: boolean
  terminalInstance: any | null
  fitAddon: any | null
  initialized: boolean
}

// ---- Props / Emits ----

const props = defineProps<{
  initialCols?: number
  initialRows?: number
}>()

const emit = defineEmits<{
  (e: 'tab-changed', index: number): void
  (e: 'session-created', sessionId: string, tabIndex: number): void
  (e: 'ready'): void
}>()

// ---- Stores ----

const terminalStore = useTerminalStore()

// ---- Refs ----

const shellRef = ref<HTMLElement | null>(null)
const terminalAreaRef = ref<HTMLElement | null>(null)
const terminalRefs = ref<(HTMLElement | null)[]>([null, null, null, null, null])

// ---- Tab State ----

const tabs = reactive<ShellTab[]>([
  { id: 'tab-0', index: 0, sessionId: null, title: 'MAIN SHELL', active: true, terminalInstance: null, fitAddon: null, initialized: false },
  { id: 'tab-1', index: 1, sessionId: null, title: 'EMPTY', active: false, terminalInstance: null, fitAddon: null, initialized: false },
  { id: 'tab-2', index: 2, sessionId: null, title: 'EMPTY', active: false, terminalInstance: null, fitAddon: null, initialized: false },
  { id: 'tab-3', index: 3, sessionId: null, title: 'EMPTY', active: false, terminalInstance: null, fitAddon: null, initialized: false },
  { id: 'tab-4', index: 4, sessionId: null, title: 'EMPTY', active: false, terminalInstance: null, fitAddon: null, initialized: false },
])

const activeTabIndex = ref(0)

// ---- Status Bar State ----

const pingMs = ref<number | null>(null)
const currentTime = ref('')
const currentDate = ref('')

// ---- Dynamic xterm imports (SSR-safe) ----

let Terminal: any = null
let FitAddon: any = null
let WebglAddon: any = null

async function loadXtermModules(): Promise<boolean> {
  if (typeof window === 'undefined') return false
  try {
    const xtermModule = await import('@xterm/xterm')
    Terminal = xtermModule.Terminal
    const fitModule = await import('@xterm/addon-fit')
    FitAddon = fitModule.FitAddon
    try {
      const webglModule = await import('@xterm/addon-webgl')
      WebglAddon = webglModule.WebglAddon
    } catch {
      // WebGL addon optional
      WebglAddon = null
    }
    return true
  } catch (err) {
    console.error('Failed to load xterm modules:', err)
    return false
  }
}

// ---- Computed ----

const activeTab = computed(() => tabs[activeTabIndex.value])

const activeProcessInfo = computed(() => {
  const tab = activeTab.value
  if (!tab || !tab.sessionId) return 'No active session'
  const session = terminalStore.sessions.get(tab.sessionId)
  if (session) {
    return `PID: ${session.id.slice(-6)} | ${session.shell || '/bin/bash'}`
  }
  return `Session: ${tab.sessionId.slice(-8)}`
})

const activeSessionLabel = computed(() => {
  const tab = activeTab.value
  if (!tab || !tab.sessionId) return 'Tab ' + (activeTabIndex.value + 1)
  return tab.title
})

const pingDisplay = computed(() => {
  if (pingMs.value === null) return 'PING: --ms'
  return `PING: ${pingMs.value}ms`
})

const dateTimeDisplay = computed(() => {
  return `${currentDate.value} ${currentTime.value}`
})

// ---- Terminal Ref Management ----

function setTerminalRef(index: number, el: HTMLElement | null) {
  terminalRefs.value[index] = el
}

// ---- Terminal Creation ----

async function createTerminalForTab(tabIndex: number): Promise<void> {
  const tab = tabs[tabIndex]
  if (tab.initialized || !Terminal || !FitAddon) return

  const container = terminalRefs.value[tabIndex]
  if (!container) return

  // Determine terminal theme from CSS variables
  const rootStyle = getComputedStyle(document.documentElement)
  const bg = rootStyle.getPropertyValue('--terminal_bg').trim() || '#05080d'
  const fg = rootStyle.getPropertyValue('--terminal_fg').trim() || 'rgb(170, 207, 209)'
  const cursor = rootStyle.getPropertyValue('--terminal_cursor').trim() || fg
  const selection = rootStyle.getPropertyValue('--terminal_selection').trim() || 'rgba(170, 207, 209, 0.3)'

  const term = new Terminal({
    allowTransparency: true,
    cursorBlink: true,
    cursorStyle: 'block' as const,
    fontFamily: rootStyle.getPropertyValue('--terminal_font').trim() || "'Fira Code', monospace",
    fontSize: 14,
    lineHeight: 1.2,
    scrollback: 5000,
    convertEol: true,
    theme: {
      background: bg,
      foreground: fg,
      cursor: cursor,
      selectionBackground: selection,
      black: '#000000',
      red: '#ff5555',
      green: '#50fa7b',
      yellow: '#f1fa8c',
      blue: '#bd93f9',
      magenta: '#ff79c6',
      cyan: '#8be9fd',
      white: '#f8f8f2',
      brightBlack: '#6272a4',
      brightRed: '#ff6e6e',
      brightGreen: '#69ff94',
      brightYellow: '#ffffa5',
      brightBlue: '#d6acff',
      brightMagenta: '#ff92df',
      brightCyan: '#a4ffff',
      brightWhite: '#ffffff',
    },
  })

  const fit = new FitAddon()
  term.loadAddon(fit)

  // Try WebGL renderer
  if (WebglAddon) {
    try {
      const webgl = new WebglAddon()
      term.loadAddon(webgl)
    } catch {
      // Fallback to canvas/DOM renderer
    }
  }

  term.open(container)

  await nextTick()
  try {
    fit.fit()
  } catch {
    // Container may not be visible yet
  }

  // Store references
  tab.terminalInstance = term
  tab.fitAddon = fit
  tab.initialized = true

  // Create backend terminal session
  try {
    const cols = term.cols || props.initialCols || 80
    const rows = term.rows || props.initialRows || 24
    const terminalData = await CreateTerminal(cols, rows)

    if (terminalData) {
      const sessionId = terminalData.id || terminalData.Id || `term-${Date.now()}`
      tab.sessionId = sessionId

      // Update tab title for main shell
      if (tabIndex === 0) {
        tab.title = 'MAIN SHELL'
      } else {
        tab.title = `SHELL ${tabIndex + 1}`
      }

      // Register session in the store
      terminalStore.addSession({
        id: sessionId,
        title: tab.title,
        shell: '/bin/bash',
        workingDirectory: '/',
        environment: {},
        isActive: tab.active,
        createdAt: new Date(),
        lastActivity: new Date(),
        cols: cols,
        rows: rows,
      })

      if (tab.active) {
        terminalStore.setActiveSession(sessionId)
      }

      emit('session-created', sessionId, tabIndex)
    }
  } catch (err) {
    console.error(`Failed to create backend terminal for tab ${tabIndex}:`, err)
    tab.title = tabIndex === 0 ? 'MAIN SHELL' : `SHELL ${tabIndex + 1}`
  }

  // Handle terminal data input -> send to backend
  term.onData((data: string) => {
    if (tab.sessionId) {
      WriteToTerminal(tab.sessionId, data).catch((err: Error) => {
        console.error('Failed to write to terminal:', err)
      })
    }
  })

  // Handle terminal title change
  term.onTitleChange((title: string) => {
    if (title) {
      tab.title = title.length > 20 ? title.slice(0, 20) + '...' : title
    }
  })

  // Handle resize
  term.onResize(({ cols, rows }: { cols: number; rows: number }) => {
    if (tab.sessionId) {
      ResizeTerminal(tab.sessionId, cols, rows).catch((err: Error) => {
        console.error('Failed to resize terminal:', err)
      })
    }
  })
}

// ---- Tab Activation ----

async function activateTab(index: number): Promise<void> {
  if (index < 0 || index >= tabs.length) return

  // Deactivate all tabs
  tabs.forEach((tab) => {
    tab.active = false
  })

  // Activate selected tab
  tabs[index].active = true
  activeTabIndex.value = index

  // Lazy-create terminal if not initialized
  if (!tabs[index].initialized) {
    await nextTick()
    await createTerminalForTab(index)
  }

  // Update store active session
  const sessionId = tabs[index].sessionId
  if (sessionId) {
    terminalStore.setActiveSession(sessionId)
  }

  // Fit the now-visible terminal
  await nextTick()
  const tab = tabs[index]
  if (tab.fitAddon) {
    try {
      tab.fitAddon.fit()
    } catch {
      // Ignore fit errors
    }
  }
  if (tab.terminalInstance) {
    tab.terminalInstance.focus()
  }

  emit('tab-changed', index)
}

// ---- Tab Context Menu ----

function onTabContextMenu(index: number, event: MouseEvent) {
  // Close terminal session for non-main tabs
  if (index > 0 && tabs[index].initialized) {
    closeTab(index)
  }
}

async function closeTab(index: number): Promise<void> {
  if (index === 0) return // Cannot close main shell

  const tab = tabs[index]

  // Close backend session
  if (tab.sessionId) {
    try {
      await CloseTerminal(tab.sessionId)
      terminalStore.removeSession(tab.sessionId)
    } catch (err) {
      console.error('Failed to close terminal session:', err)
    }
  }

  // Dispose terminal instance
  if (tab.terminalInstance) {
    tab.terminalInstance.dispose()
  }

  // Reset tab state
  tab.sessionId = null
  tab.title = 'EMPTY'
  tab.terminalInstance = null
  tab.fitAddon = null
  tab.initialized = false

  // Switch to main shell if this was active
  if (tab.active) {
    await activateTab(0)
  }
}

// ---- Event Handling: Terminal Output from Backend ----

function handleTerminalOutput(sessionId: string, data: string) {
  const tab = tabs.find((t) => t.sessionId === sessionId)
  if (tab && tab.terminalInstance) {
    tab.terminalInstance.write(data)
  }
}

function handleTerminalClosed(sessionId: string) {
  const tab = tabs.find((t) => t.sessionId === sessionId)
  if (tab) {
    const idx = tab.index
    if (idx === 0) {
      // Main shell closed - recreate
      tab.initialized = false
      tab.sessionId = null
      tab.terminalInstance?.dispose()
      tab.terminalInstance = null
      tab.fitAddon = null
      tab.title = 'MAIN SHELL'
      if (tab.active) {
        nextTick(() => createTerminalForTab(0))
      }
    } else {
      // Non-main tab
      if (tab.terminalInstance) {
        tab.terminalInstance.dispose()
      }
      tab.sessionId = null
      tab.title = 'EMPTY'
      tab.terminalInstance = null
      tab.fitAddon = null
      tab.initialized = false
      if (tab.active) {
        activateTab(0)
      }
    }
  }
}

// ---- Resize Handling ----

let resizeObserver: ResizeObserver | null = null

function handleContainerResize() {
  const tab = tabs[activeTabIndex.value]
  if (tab && tab.fitAddon && tab.initialized) {
    try {
      tab.fitAddon.fit()
    } catch {
      // Ignore
    }
  }
}

// ---- Clock / Date ----

let clockInterval: ReturnType<typeof setInterval> | null = null

function updateClock() {
  const now = new Date()
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  currentTime.value = `${hours}:${minutes}:${seconds}`

  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  currentDate.value = `${year}-${month}-${day}`
}

// ---- Ping ----

let pingInterval: ReturnType<typeof setInterval> | null = null

function updatePing() {
  // Simple performance-based approximation
  const start = nowMs()
  requestAnimationFrame(() => {
    const elapsed = Math.round(nowMs() - start)
    pingMs.value = elapsed
  })
}

// ---- Exposed Methods ----

function writeToActiveTerminal(data: string) {
  const tab = tabs[activeTabIndex.value]
  if (tab && tab.sessionId) {
    WriteToTerminal(tab.sessionId, data).catch((err: Error) => {
      console.error('Failed to write to active terminal:', err)
    })
  }
}

function sendKeyToActiveTerminal(key: string) {
  writeToActiveTerminal(key)
}

function focusActiveTerminal() {
  const tab = tabs[activeTabIndex.value]
  if (tab && tab.terminalInstance) {
    tab.terminalInstance.focus()
  }
}

function fitAllTerminals() {
  tabs.forEach((tab) => {
    if (tab.fitAddon && tab.initialized) {
      try {
        tab.fitAddon.fit()
      } catch {
        // Ignore
      }
    }
  })
}

function getActiveSessionId(): string | null {
  return tabs[activeTabIndex.value]?.sessionId || null
}

defineExpose({
  writeToActiveTerminal,
  sendKeyToActiveTerminal,
  focusActiveTerminal,
  fitAllTerminals,
  activateTab,
  closeTab,
  getActiveSessionId,
})

// ---- Wails Event Listeners ----

let unsubOutput: (() => void) | undefined
let unsubClosed: (() => void) | undefined
let unsubPing: (() => void) | undefined

function setupWailsEvents() {
  // Listen for terminal output from the Go backend
  unsubOutput = Events.On('terminal.output', (sessionId: string, data: string) => {
    handleTerminalOutput(sessionId, data)
  })

  // Listen for terminal closed events
  unsubClosed = Events.On('terminal.closed', (sessionId: string) => {
    handleTerminalClosed(sessionId)
  })

  // Listen for ping updates if backend provides them
  unsubPing = Events.On('network.ping', (ms: number) => {
    if (typeof ms === 'number') {
      pingMs.value = ms
    }
  })
}

function cleanupWailsEvents() {
  // Call the unsubscribe functions Events.On returned rather than
  // removing by event name: each call drops only this component's
  // listener, leaving any other subscriber to the same event intact.
  unsubOutput?.()
  unsubClosed?.()
  unsubPing?.()
  unsubOutput = unsubClosed = unsubPing = undefined
}

// ---- Lifecycle ----

onMounted(async () => {
  const loaded = await loadXtermModules()
  if (!loaded) {
    console.error('xterm modules failed to load')
    return
  }

  // Setup Wails event listeners
  setupWailsEvents()

  // Start clock
  updateClock()
  clockInterval = setInterval(updateClock, 1000)

  // Start ping
  updatePing()
  pingInterval = setInterval(updatePing, 5000)

  // Auto-create main shell (tab 0)
  await nextTick()
  await createTerminalForTab(0)

  // Setup resize observer
  if (terminalAreaRef.value) {
    resizeObserver = new ResizeObserver(() => {
      handleContainerResize()
    })
    resizeObserver.observe(terminalAreaRef.value)
  }

  emit('ready')
})

onBeforeUnmount(() => {
  // Cleanup clock
  if (clockInterval) {
    clearInterval(clockInterval)
    clockInterval = null
  }

  // Cleanup ping
  if (pingInterval) {
    clearInterval(pingInterval)
    pingInterval = null
  }

  // Cleanup resize observer
  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }

  // Cleanup Wails events
  cleanupWailsEvents()

  // Dispose all terminal instances
  tabs.forEach((tab) => {
    if (tab.terminalInstance) {
      tab.terminalInstance.dispose()
      tab.terminalInstance = null
    }
    tab.fitAddon = null
    tab.initialized = false
  })
})
</script>

<style scoped>
/*
 * The main-shell, shell-tabs, shell-tab, shell-terminal, shell-status classes
 * are primarily defined in main.css. These scoped styles provide additional
 * layout handling for the multi-pane terminal container.
 */

.shell-terminal-pane {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  pointer-events: none;
  overflow: hidden;
}

.shell-terminal-pane.visible {
  opacity: 1;
  pointer-events: auto;
  z-index: 1;
}

.shell-terminal-pane :deep(.xterm) {
  height: 100%;
  padding: 0.3vh 0.3vw;
}

.shell-terminal-pane :deep(.xterm-viewport) {
  overflow-y: auto;
}

.shell-terminal-pane :deep(.xterm-screen) {
  height: 100%;
}

/* Status bar sub-elements */
.shell-status-left,
.shell-status-right {
  display: flex;
  align-items: center;
  gap: 0.5vw;
}

.shell-status-left {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.shell-status-right {
  flex-shrink: 0;
}

.status-separator {
  opacity: 0.3;
  margin: 0 0.2vw;
}

.status-process {
  color: var(--color_accent);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status-session {
  color: var(--color_accent_dimmed);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status-ping {
  color: var(--color_accent);
  font-variant-numeric: tabular-nums;
}

.status-datetime {
  color: var(--color_accent);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.05em;
}
</style>
