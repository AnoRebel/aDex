<template>
  <div ref="terminalRef" class="terminal-emulator"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'

interface TerminalSettings {
  fontSize: number
  fontFamily: string
  backgroundColor: string
  textColor: string
  cursorBlink: boolean
  cursorColor: string
  scrollback: number
}

interface Props {
  sessionId: string
  onData?: (data: string) => void
  onResize?: (cols: number, rows: number) => void
  onTitle?: (title: string) => void
  settings?: Partial<TerminalSettings>
}

const props = withDefaults(defineProps<Props>(), {
  settings: () => ({}),
})

const emit = defineEmits<{
  data: [data: string]
  resize: [cols: number, rows: number]
  title: [title: string]
  ready: []
}>()

// Template refs
const terminalRef = ref<HTMLElement>()

// Terminal instance (will be initialized later)
let terminal: any = null
let fitAddon: any = null

// Default terminal settings
const defaultSettings: TerminalSettings = {
  fontSize: 14,
  fontFamily: '"JetBrains Mono", monospace',
  backgroundColor: '#000000',
  textColor: '#00ff00',
  cursorBlink: true,
  cursorColor: '#00ff00',
  scrollback: 1000,
}

const terminalSettings = computed(() => ({ ...defaultSettings, ...props.settings }))

// Initialize terminal
const initializeTerminal = async () => {
  if (!terminalRef.value) return

  // Dynamic import of xterm
  try {
    const { Terminal } = await import('xterm')
    const { FitAddon } = await import('xterm-addon-fit')
    const { WebglAddon } = await import('xterm-addon-webgl')
    const { LigaturesAddon } = await import('xterm-addon-ligatures')

    // Create terminal instance
    terminal = new Terminal({
      cursorBlink: terminalSettings.value.cursorBlink,
      cursorStyle: 'block',
      cursorWidth: 1,
      fontSize: terminalSettings.value.fontSize,
      fontFamily: terminalSettings.value.fontFamily,
      theme: {
        background: terminalSettings.value.backgroundColor,
        foreground: terminalSettings.value.textColor,
        cursor: terminalSettings.value.cursorColor,
        selection: 'rgba(0, 255, 0, 0.3)',
        black: '#000000',
        red: '#ff0000',
        green: '#00ff00',
        yellow: '#ffff00',
        blue: '#0000ff',
        magenta: '#ff00ff',
        cyan: '#00ffff',
        white: '#ffffff',
      },
      scrollback: terminalSettings.value.scrollback,
      allowTransparency: false,
      rows: 24,
      cols: 80,
    })

    // Addons
    fitAddon = new FitAddon()
    const ligaturesAddon = new LigaturesAddon()
    
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(ligaturesAddon)

    // Try to enable WebGL renderer
    try {
      const webglAddon = new WebglAddon()
      terminal.loadAddon(webglAddon)
    } catch (e) {
      console.warn('WebGL addon not supported, using DOM renderer')
    }

    // Event handlers
    terminal.onData((data: string) => {
      emit('data', data)
      props.onData?.(data)
    })

    terminal.onResize((size: any) => {
      emit('resize', size.cols, size.rows)
      props.onResize?.(size.cols, size.rows)
    })

    terminal.onTitleChange((title: string) => {
      emit('title', title)
      props.onTitle?.(title)
    })

    // Mount terminal
    terminal.open(terminalRef.value)

    // Fit terminal to container
    await nextTick()
    if (fitAddon) {
      fitAddon.fit()
    }

    emit('ready')
  } catch (error) {
    console.error('Failed to initialize terminal:', error)
  }
}

// Write data to terminal
const write = (data: string) => {
  if (terminal) {
    terminal.write(data)
  }
}

// Clear terminal
const clear = () => {
  if (terminal) {
    terminal.clear()
  }
}

// Focus terminal
const focus = () => {
  if (terminal) {
    terminal.focus()
  }
}

// Resize terminal
const resize = (cols: number, rows: number) => {
  if (terminal && fitAddon) {
    try {
      terminal.resize(cols, rows)
      fitAddon.fit()
    } catch (e) {
      console.warn('Failed to resize terminal:', e)
    }
  }
}

// Handle window resize
const handleResize = () => {
  if (fitAddon) {
    nextTick(() => {
      fitAddon.fit()
    })
  }
}

// Lifecycle
onMounted(async () => {
  await initializeTerminal()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (terminal) {
    terminal.dispose()
  }
})

// Expose methods
defineExpose({
  write,
  clear,
  focus,
  resize,
})
</script>

<style scoped>
.terminal-emulator {
  width: 100%;
  height: 100%;
  background: #000000;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 0 20px rgba(0, 255, 255, 0.3);
}

/* Custom scrollbar */
:deep(.xterm-viewport) {
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 255, 255, 0.3) transparent;
}

:deep(.xterm-viewport)::-webkit-scrollbar {
  width: 8px;
}

:deep(.xterm-viewport)::-webkit-scrollbar-track {
  background: transparent;
}

:deep(.xterm-viewport)::-webkit-scrollbar-thumb {
  background: rgba(0, 255, 255, 0.3);
  border-radius: 4px;
}

:deep(.xterm-viewport)::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 255, 255, 0.5);
}

/* Terminal selection */
:deep(.xterm-selection) {
  background: rgba(0, 255, 255, 0.3) !important;
}

/* Custom cursor animations */
:deep(.xterm-cursor) {
  animation: cursor-blink 1s infinite;
}

@keyframes cursor-blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* Terminal focus outline */
:deep(.xterm:focus) {
  outline: none;
}

/* Font ligature support */
:deep(.xterm-rows) {
  font-feature-settings: "liga" 1, "calt" 1;
}
</style>