<template>
  <div ref="containerRef" class="edex-terminal-instance">
    <div ref="xtermRef" class="xterm-container" />
  </div>
</template>

<script setup lang="ts">
/**
 * EdexTerminal -- renders a single xterm.js terminal for one shell tab.
 *
 * Props:
 *   sessionId  – the backend terminal session id
 *   active     – whether this terminal is the currently visible one
 *
 * Exposes:
 *   write(data)  – write raw data to the terminal
 *   focus()      – focus the xterm element
 *   fit()        – refit the terminal to its container
 */
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useTerminalStore } from '~/stores/terminal'

const props = defineProps<{
  sessionId: string
  active: boolean
}>()

const terminalStore = useTerminalStore()
const containerRef = ref<HTMLElement | null>(null)
const xtermRef = ref<HTMLElement | null>(null)

// Dynamic module refs (xterm is not SSR-safe)
let Terminal: any = null
let FitAddon: any = null
let term: any = null
let fitAddon: any = null
let resizeObserver: ResizeObserver | null = null
let wailsEventCleanup: (() => void) | null = null

// ---- Init ----
async function initTerminal() {
  if (!xtermRef.value) return

  // Dynamically import xterm (client-only)
  const xtermMod = await import('xterm')
  Terminal = xtermMod.Terminal

  const fitMod = await import('xterm-addon-fit')
  FitAddon = fitMod.FitAddon

  // Read CSS vars for theme colors
  const style = getComputedStyle(document.documentElement)
  const bg = style.getPropertyValue('--terminal_bg').trim() || '#05080d'
  const fg = style.getPropertyValue('--terminal_fg').trim() || '#aacfd1'
  const cursor = style.getPropertyValue('--terminal_cursor').trim() || fg
  const selection = style.getPropertyValue('--terminal_selection').trim() || 'rgba(170,207,209,0.3)'

  term = new Terminal({
    fontFamily: style.getPropertyValue('--terminal_font').trim() || "'Fira Code', monospace",
    fontSize: 14,
    lineHeight: 1.2,
    cursorStyle: 'block' as const,
    cursorBlink: true,
    scrollback: 1500,
    allowProposedApi: true,
    theme: {
      background: bg,
      foreground: fg,
      cursor,
      selection,
      black: '#000000',
      red: '#e06c75',
      green: '#98c379',
      yellow: '#d19a66',
      blue: '#61afef',
      magenta: '#c678dd',
      cyan: '#56b6c2',
      white: '#abb2bf',
      brightBlack: '#5c6370',
      brightRed: '#e06c75',
      brightGreen: '#98c379',
      brightYellow: '#d19a66',
      brightBlue: '#61afef',
      brightMagenta: '#c678dd',
      brightCyan: '#56b6c2',
      brightWhite: '#ffffff',
    },
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)

  // Try WebGL renderer for performance
  try {
    const webglMod = await import('xterm-addon-webgl')
    term.loadAddon(new webglMod.WebglAddon())
  } catch {
    // WebGL not available, fall back to canvas
    try {
      const canvasMod = await import('xterm-addon-canvas')
      term.loadAddon(new canvasMod.CanvasAddon())
    } catch {
      // Use default renderer
    }
  }

  // Open and fit
  term.open(xtermRef.value)
  await nextTick()
  try {
    fitAddon.fit()
  } catch {
    // ignore fit errors during init
  }

  // Send user input to backend
  term.onData((data: string) => {
    terminalStore.sendInput(data, props.sessionId)
  })

  // Notify backend of size changes
  term.onResize(({ cols, rows }: { cols: number; rows: number }) => {
    terminalStore.resizeSession(cols, rows, props.sessionId)
  })

  // Listen for output from the Wails backend
  setupWailsEvents()

  // Observe container resize
  if (containerRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (props.active) {
        try {
          fitAddon?.fit()
        } catch {
          // ignore
        }
      }
    })
    resizeObserver.observe(containerRef.value)
  }

  // If this terminal is active, focus it
  if (props.active) {
    term.focus()
  }

  // Write welcome message for the first terminal
  term.writeln(`\x1b[36mWelcome to aDex-UI Terminal\x1b[0m`)
  term.writeln(`\x1b[90mSession: ${props.sessionId}\x1b[0m`)
  term.writeln('')
}

function setupWailsEvents() {
  // Wails v2 events -- listen for terminal output routed to this session
  const runtime = (window as any).runtime
  if (runtime?.EventsOn) {
    const handler = (data: any) => {
      // The event may carry { sessionId, data } or just raw data
      if (typeof data === 'object' && data.sessionId) {
        if (data.sessionId === props.sessionId && term) {
          term.write(data.data)
        }
      } else if (typeof data === 'string' && term) {
        term.write(data)
      }
    }
    runtime.EventsOn('terminal.output', handler)
    runtime.EventsOn(`terminal.output.${props.sessionId}`, handler)

    wailsEventCleanup = () => {
      runtime.EventsOff('terminal.output')
      runtime.EventsOff(`terminal.output.${props.sessionId}`)
    }
  }
}

// ---- Exposed methods ----
function write(data: string) {
  term?.write(data)
}

function focus() {
  term?.focus()
}

function fit() {
  try {
    fitAddon?.fit()
  } catch {
    // ignore
  }
}

defineExpose({ write, focus, fit })

// ---- Watchers ----
watch(
  () => props.active,
  (isActive) => {
    if (isActive) {
      nextTick(() => {
        fit()
        focus()
      })
    }
  }
)

// ---- Lifecycle ----
onMounted(() => {
  initTerminal()
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  wailsEventCleanup?.()
  term?.dispose()
  term = null
  fitAddon = null
})
</script>

<style scoped>
.edex-terminal-instance {
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;
}

.xterm-container {
  width: 100%;
  height: 100%;
}

/* xterm.js overrides for eDex-UI look */
.xterm-container :deep(.xterm) {
  height: 100%;
  padding: 0.3vh 0.3vw;
}

.xterm-container :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

.xterm-container :deep(.xterm-viewport::-webkit-scrollbar) {
  width: 0.3vw;
}

.xterm-container :deep(.xterm-viewport::-webkit-scrollbar-track) {
  background: transparent;
}

.xterm-container :deep(.xterm-viewport::-webkit-scrollbar-thumb) {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
}
</style>
