<template>
  <div ref="terminalRef" class="terminal-emulator"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { useWails } from '~/composables/useWails'

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
  sessionId?: string
  onData?: (data: string) => void
  onResize?: (cols: number, rows: number) => void
  onTitle?: (title: string) => void
  settings?: Partial<TerminalSettings>
}

const props = withDefaults(defineProps<Props>(), {
  sessionId: 'default',
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
let backendTerminalId: string | null = null

// Wails terminal service
const { terminal: terminalService } = useWails()

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

// Import Wails runtime for events
import { Events } from '~/lib/wailsjs/runtime'

// Initialize terminal
const initializeTerminal = async () => {
  if (!terminalRef.value) return

  // Dynamic import of xterm
  try {
    const { Terminal } = await import('xterm')
    const { FitAddon } = await import('xterm-addon-fit')

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

    // Add fit addon
    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)

    // Try to create backend terminal
    try {
      const dims = fitAddon.proposeDimensions()
      const cols = dims?.cols || 80
      const rows = dims?.rows || 24
      
      const result = await terminalService.create(cols, rows)
      if (result && result.id) {
        backendTerminalId = result.id
        console.log('Backend terminal created:', backendTerminalId)
        
        // Subscribe to terminal output events from backend
        // Wails v2 EventsOn passes data as separate arguments
        Events.On('terminal.output', (...args: any[]) => {
          const event = args[0]
          if (event && event.terminalId === backendTerminalId) {
            let text = ''
            if (event.data) {
              if (event.data instanceof Uint8Array || Array.isArray(event.data)) {
                // Convert array of bytes back to string
                const bytes = new Uint8Array(event.data)
                text = new TextDecoder().decode(bytes)
              } else if (typeof event.data === 'string') {
                text = event.data
              }
            }
            if (text) {
              terminal?.write(text)
            }
          }
        })
      }
    } catch (e) {
      console.warn('Failed to create backend terminal:', e)
      // Continue with local-only terminal
    }

    // Event handlers
    terminal.onData(async (data: string) => {
      emit('data', data)
      props.onData?.(data)
      
      // Send to backend if available
      if (backendTerminalId) {
        try {
          await terminalService.write(backendTerminalId, data)
        } catch (e) {
          // Backend write failed, ignore
        }
      } else {
        // Local echo mode (demo)
        handleLocalInput(data)
      }
    })

    terminal.onResize(async (size: any) => {
      emit('resize', size.cols, size.rows)
      props.onResize?.(size.cols, size.rows)
      
      // Resize backend terminal if available
      if (backendTerminalId) {
        try {
          await terminalService.resize(backendTerminalId, size.cols, size.rows)
        } catch (e) {
          // Ignore resize errors
        }
      }
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

    // Write welcome message
    if (!backendTerminalId) {
      // Only write banner and fake prompt in local/demo mode
      terminal.writeln('\x1b[36m╔════════════════════════════════════════════════════════╗\x1b[0m')
      terminal.writeln('\x1b[36m║\x1b[0m  \x1b[1;32maDex-UI Terminal\x1b[0m                                      \x1b[36m║\x1b[0m')
      terminal.writeln('\x1b[36m║\x1b[0m  \x1b[33mAdvanced Desktop Environment\x1b[0m                          \x1b[36m║\x1b[0m')
      terminal.writeln('\x1b[36m╚════════════════════════════════════════════════════════╝\x1b[0m')
      terminal.writeln('')
      terminal.writeln('\x1b[90mLocal terminal mode. Type "help" for commands.\x1b[0m')
      terminal.writeln('')
      terminal.write('\x1b[32muser@adex\x1b[0m:\x1b[34m~\x1b[0m$ ')
    }
    // When connected to backend, the shell will send its own prompt

    emit('ready')
  } catch (error) {
    console.error('Failed to initialize terminal:', error)
  }
}

// Local input handler for demo mode
let currentLine = ''
const handleLocalInput = (data: string) => {
  if (!terminal) return
  
  // Handle Enter key
  if (data === '\r') {
    terminal.writeln('')
    handleCommand(currentLine.trim())
    currentLine = ''
    terminal.write('\x1b[32muser@adex\x1b[0m:\x1b[34m~\x1b[0m$ ')
  }
  // Handle Backspace
  else if (data === '\x7f') {
    if (currentLine.length > 0) {
      currentLine = currentLine.slice(0, -1)
      terminal.write('\b \b')
    }
  }
  // Handle regular characters
  else if (data >= ' ') {
    currentLine += data
    terminal.write(data)
  }
}

// Handle demo commands
const handleCommand = (cmd: string) => {
  if (!terminal) return
  
  const commands: Record<string, () => void> = {
    'help': () => {
      terminal.writeln('\x1b[1;36mAvailable commands:\x1b[0m')
      terminal.writeln('  \x1b[33mhelp\x1b[0m     - Show this help message')
      terminal.writeln('  \x1b[33mclear\x1b[0m    - Clear the terminal')
      terminal.writeln('  \x1b[33mdate\x1b[0m     - Show current date and time')
      terminal.writeln('  \x1b[33mwhoami\x1b[0m   - Show current user')
      terminal.writeln('  \x1b[33mhostname\x1b[0m - Show system hostname')
      terminal.writeln('  \x1b[33muptime\x1b[0m   - Show system uptime')
      terminal.writeln('  \x1b[33mecho\x1b[0m     - Echo text back')
      terminal.writeln('  \x1b[33mversion\x1b[0m  - Show aDex-UI version')
    },
    'clear': () => {
      terminal.clear()
    },
    'date': () => {
      terminal.writeln(new Date().toString())
    },
    'whoami': () => {
      terminal.writeln('user')
    },
    'hostname': () => {
      terminal.writeln('adex-desktop')
    },
    'uptime': () => {
      terminal.writeln('System up for 1 hour, 23 minutes')
    },
    'version': () => {
      terminal.writeln('aDex-UI v2.0.0')
    }
  }
  
  if (cmd === '') return
  
  if (cmd.startsWith('echo ')) {
    terminal.writeln(cmd.substring(5))
  } else if (commands[cmd]) {
    commands[cmd]()
  } else {
    terminal.writeln(`\x1b[31mCommand not found: ${cmd}\x1b[0m`)
    terminal.writeln('Type "help" for available commands')
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
  
  // Unsubscribe from terminal output events
  Events.Off('terminal.output')
  
  // Close backend terminal if exists
  if (backendTerminalId) {
    terminalService.close(backendTerminalId).catch(() => {})
  }
  
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
