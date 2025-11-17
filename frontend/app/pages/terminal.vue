<template>
  <div class="terminal-page">
    <!-- Terminal Header -->
    <header class="terminal-header">
      <div class="header-left">
        <h1 class="page-title">Terminal</h1>
        <div class="session-tabs">
          <div 
            v-for="session in allSessions" 
            :key="session.id"
            class="session-tab"
            :class="{ active: session.id === activeSessionId }"
            @click="setActiveSession(session.id)"
          >
            <span>{{ session.shell }}</span>
            <button 
              class="close-tab"
              @click.stop="closeSession(session.id)"
            >
              ×
            </button>
          </div>
          <button class="new-tab-btn" @click="createNewSession">
            +
          </button>
        </div>
      </div>
      <div class="header-right">
        <div class="terminal-actions">
          <button class="action-btn" @click="clearOutput">
            Clear
          </button>
          <button class="action-btn" @click="copyOutput">
            Copy
          </button>
          <button class="action-btn settings-btn" @click="toggleSettings">
            ⚙️
          </button>
        </div>
      </div>
    </header>

    <!-- Terminal Content -->
    <main class="terminal-content">
      <div class="terminal-container">
        <div class="terminal-output" ref="outputRef">
          <div 
            v-for="(line, index) in currentOutput" 
            :key="index"
            class="output-line"
            :class="line.type"
          >
            <span class="line-content">{{ line.content }}</span>
            <span class="line-timestamp">
              {{ formatTimestamp(line.timestamp) }}
            </span>
          </div>
        </div>
        
        <div class="terminal-input-container">
          <div class="prompt">
            <span class="prompt-user">{{ currentSession?.workingDirectory || '~' }}</span>
            <span class="prompt-symbol">$</span>
          </div>
          <input 
            ref="inputRef"
            v-model="currentCommand"
            type="text"
            class="terminal-input"
            :placeholder="isProcessing ? 'Processing...' : 'Enter command...'"
            :disabled="isProcessing"
            @keydown="handleKeyDown"
            @keyup="handleKeyUp"
          />
        </div>
      </div>

      <!-- Command Suggestions -->
      <div v-if="suggestions.length > 0" class="command-suggestions">
        <div 
          v-for="(suggestion, index) in suggestions" 
          :key="suggestion"
          class="suggestion-item"
          :class="{ active: index === selectedSuggestion }"
          @click="selectSuggestion(suggestion)"
        >
          {{ suggestion }}
        </div>
      </div>
    </main>

    <!-- Terminal Settings Panel -->
    <Teleport to="body">
      <div v-if="showSettings" class="modal-overlay" @click="toggleSettings">
        <div class="modal-panel settings-panel" @click.stop>
          <div class="panel-header">
            <h3>Terminal Settings</h3>
            <button class="close-btn" @click="toggleSettings">×</button>
          </div>
          <div class="panel-content">
            <div class="setting-group">
              <label>Shell</label>
              <select v-model="terminalSettings.shell">
                <option value="/bin/bash">Bash</option>
                <option value="/bin/zsh">Zsh</option>
                <option value="/bin/fish">Fish</option>
                <option value="/bin/powershell">PowerShell</option>
              </select>
            </div>
            
            <div class="setting-group">
              <label>Font Size</label>
              <input 
                v-model.number="terminalSettings.fontSize"
                type="range"
                min="10"
                max="24"
              />
              <span>{{ terminalSettings.fontSize }}px</span>
            </div>
            
            <div class="setting-group">
              <label>Font Family</label>
              <select v-model="terminalSettings.fontFamily">
                <option value='"JetBrains Mono", monospace'>JetBrains Mono</option>
                <option value='"Fira Code", monospace'>Fira Code</option>
                <option value='"Courier New", monospace'>Courier New</option>
                <option value='"Monaco", monospace'>Monaco</option>
              </select>
            </div>
            
            <div class="setting-group">
              <label>Background Color</label>
              <input 
                v-model="terminalSettings.backgroundColor"
                type="color"
              />
            </div>
            
            <div class="setting-group">
              <label>Text Color</label>
              <input 
                v-model="terminalSettings.textColor"
                type="color"
              />
            </div>
            
            <div class="setting-group">
              <label>
                <input 
                  v-model="terminalSettings.cursorBlink"
                  type="checkbox"
                />
                Cursor Blink
              </label>
            </div>
            
            <div class="setting-group">
              <label>Scrollback Lines</label>
              <input 
                v-model.number="terminalSettings.scrollback"
                type="number"
                min="100"
                max="10000"
              />
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Side Panel with Command History -->
    <aside class="terminal-sidebar">
      <div class="sidebar-section">
        <h3>Command History</h3>
        <div class="history-list">
          <div 
            v-for="cmd in recentCommands" 
            :key="cmd.id"
            class="history-item"
            @click="runCommandFromHistory(cmd.command)"
          >
            <span class="history-command">{{ cmd.command }}</span>
            <span class="history-status" :class="cmd.status">
              {{ cmd.status }}
            </span>
            <span class="history-time">{{ formatTime(cmd.timestamp) }}</span>
          </div>
        </div>
      </div>
      
      <div class="sidebar-section">
        <h3>Quick Commands</h3>
        <div class="quick-commands">
          <button 
            v-for="cmd in quickCommands" 
            :key="cmd.name"
            class="quick-cmd-btn"
            @click="runCommand(cmd.command)"
          >
            {{ cmd.name }}
          </button>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { useTerminal } from '~/composables/useTerminal'
import type { TerminalOutput } from '~/types/terminal'

// Composables
const terminal = useTerminal()

// Template refs
const outputRef = ref<HTMLElement>()
const inputRef = ref<HTMLInputElement>()

// Local state
const currentCommand = ref('')
const suggestions = ref<string[]>([])
const selectedSuggestion = ref(0)
const showSettings = ref(false)
const commandHistoryIndex = ref(-1)

const terminalSettings = ref({
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

const quickCommands = [
  { name: 'List Files', command: 'ls -la' },
  { name: 'Current Directory', command: 'pwd' },
  { name: 'System Info', command: 'uname -a' },
  { name: 'Disk Usage', command: 'df -h' },
  { name: 'Memory', command: 'free -h' },
  { name: 'Processes', command: 'ps aux' },
  { name: 'Network', command: 'ip addr' },
  { name: 'Environment', command: 'env' },
]

// Computed properties
const activeSession = computed(() => terminal.activeSession.value)
const allSessions = computed(() => terminal.allSessions.value)
const currentOutput = computed(() => {
  if (!activeSession.value) return []
  return terminal.sessionOutput(activeSession.value.id)
})
const recentCommands = computed(() => terminal.recentCommands.value)
const isProcessing = computed(() => terminal.isProcessing.value)

// Methods
const createNewSession = async () => {
  const sessionId = await terminal.createSession(terminalSettings.value.shell)
  if (sessionId) {
    terminal.setActiveSession(sessionId)
  }
}

const setActiveSession = (sessionId: string) => {
  terminal.setActiveSession(sessionId)
}

const closeSession = async (sessionId: string) => {
  await terminal.closeSession(sessionId)
}

const executeCommand = async (command: string) => {
  if (!command.trim() || !activeSession.value || isProcessing.value) {
    return
  }

  // Clear input
  currentCommand.value = ''
  suggestions.value = []
  commandHistoryIndex.value = -1

  // Execute command
  await terminal.executeCommand(command)
  
  // Scroll to bottom
  await nextTick()
  scrollToBottom()
}

const runCommand = (command: string) => {
  currentCommand.value = command
  executeCommand(command)
}

const runCommandFromHistory = (command: string) => {
  currentCommand.value = command
  inputRef.value?.focus()
}

const clearOutput = () => {
  if (activeSession.value) {
    terminal.clearOutput(activeSession.value.id)
  }
}

const copyOutput = () => {
  if (currentOutput.value.length > 0) {
    const text = currentOutput.value.map(line => line.content).join('\n')
    navigator.clipboard.writeText(text)
  }
}

const toggleSettings = () => {
  showSettings.value = !showSettings.value
}

const getSuggestions = (partial: string) => {
  if (partial.trim()) {
    suggestions.value = terminal.getCommandSuggestions(partial)
    selectedSuggestion.value = 0
  } else {
    suggestions.value = []
  }
}

const selectSuggestion = (suggestion: string) => {
  currentCommand.value = suggestion
  suggestions.value = []
  inputRef.value?.focus()
}

const scrollToBottom = () => {
  if (outputRef.value) {
    outputRef.value.scrollTop = outputRef.value.scrollHeight
  }
}

const formatTimestamp = (timestamp: string) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString()
}

const formatTime = (date: Date) => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

// Event handlers
const handleKeyDown = (event: KeyboardEvent) => {
  switch (event.key) {
    case 'Enter':
      event.preventDefault()
      executeCommand(currentCommand.value)
      break
      
    case 'Tab':
      event.preventDefault()
      if (suggestions.value.length > 0) {
        const suggestion = suggestions.value[selectedSuggestion.value]
        selectSuggestion(suggestion)
      }
      break
      
    case 'ArrowUp':
      event.preventDefault()
      if (suggestions.value.length > 0) {
        selectedSuggestion.value = Math.max(0, selectedSuggestion.value - 1)
      } else if (recentCommands.value.length > 0) {
        // Navigate command history
        if (commandHistoryIndex.value < recentCommands.value.length - 1) {
          commandHistoryIndex.value++
          currentCommand.value = recentCommands.value[recentCommands.value.length - 1 - commandHistoryIndex.value].command
        }
      }
      break
      
    case 'ArrowDown':
      event.preventDefault()
      if (suggestions.value.length > 0) {
        selectedSuggestion.value = Math.min(suggestions.value.length - 1, selectedSuggestion.value + 1)
      } else if (commandHistoryIndex.value > 0) {
        commandHistoryIndex.value--
        currentCommand.value = recentCommands.value[recentCommands.value.length - 1 - commandHistoryIndex.value].command
      } else if (commandHistoryIndex.value === 0) {
        commandHistoryIndex.value = -1
        currentCommand.value = ''
      }
      break
      
    case 'Escape':
      suggestions.value = []
      break
      
    case 'c':
      if (event.ctrlKey) {
        event.preventDefault()
        // Cancel current operation (would need backend implementation)
      }
      break
  }
}

const handleKeyUp = () => {
  if (currentCommand.value) {
    getSuggestions(currentCommand.value)
  } else {
    suggestions.value = []
  }
}

// Watch for output changes and scroll to bottom
watch(currentOutput, () => {
  nextTick(() => {
    scrollToBottom()
  })
})

// Watch for terminal settings changes
watch(terminalSettings, (newSettings) => {
  terminal.updateConfig(newSettings)
}, { deep: true })

// Initialize
onMounted(async () => {
  await terminal.initialize()
  
  // Create initial session if none exists
  if (!terminal.hasActiveSession.value) {
    await createNewSession()
  }
  
  // Focus input
  inputRef.value?.focus()
})

// SEO
useHead({
  title: 'Terminal - aDex-UI',
  meta: [
    { name: 'description', content: 'Advanced terminal emulator with multiple sessions and command history' }
  ]
})
</script>

<style scoped>
.terminal-page {
  display: grid;
  grid-template-columns: 1fr 300px;
  grid-template-rows: auto 1fr;
  height: 100vh;
  background: var(--color-background);
}

.terminal-header {
  grid-column: 1 / -1;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.page-title {
  margin: 0;
  color: var(--color-text);
  font-size: 1.5rem;
}

.session-tabs {
  display: flex;
  gap: 0.25rem;
  background: var(--color-background);
  padding: 0.25rem;
  border-radius: 0.5rem;
}

.session-tab {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 0.25rem;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.session-tab:hover {
  background: var(--color-surface);
}

.session-tab.active {
  background: var(--color-primary);
  color: var(--color-background);
}

.close-tab {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  margin-left: 0.5rem;
  opacity: 0.7;
}

.close-tab:hover {
  opacity: 1;
}

.new-tab-btn {
  background: transparent;
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.new-tab-btn:hover {
  background: var(--color-surface);
  color: var(--color-text);
}

.terminal-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.settings-btn {
  padding: 0.5rem 0.75rem;
}

.terminal-content {
  display: flex;
  flex-direction: column;
  position: relative;
  background: var(--color-background);
}

.terminal-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin: 1rem;
  background: v-bind('terminalSettings.backgroundColor');
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  overflow: hidden;
  font-family: v-bind('terminalSettings.fontFamily');
  font-size: v-bind('terminalSettings.fontSize + "px"');
  color: v-bind('terminalSettings.textColor');
}

.terminal-output {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  font-family: inherit;
  font-size: inherit;
}

.output-line {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.25rem;
  font-family: inherit;
  font-size: inherit;
  white-space: pre-wrap;
  word-break: break-word;
}

.line-content {
  flex: 1;
  font-family: inherit;
  font-size: inherit;
}

.line-timestamp {
  font-size: 0.75em;
  opacity: 0.6;
  margin-left: 1rem;
  font-family: inherit;
}

.output-line.command .line-content {
  color: var(--color-primary);
  font-weight: bold;
}

.output-line.error .line-content {
  color: var(--color-error);
}

.output-line.success .line-content {
  color: var(--color-success);
}

.output-line.warning .line-content {
  color: var(--color-warning);
}

.terminal-input-container {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-top: 1px solid var(--color-border);
  background: v-bind('terminalSettings.backgroundColor');
}

.prompt {
  display: flex;
  align-items: center;
  margin-right: 0.5rem;
  font-family: inherit;
  font-size: inherit;
}

.prompt-user {
  color: var(--color-accent);
  font-weight: bold;
}

.prompt-symbol {
  margin-left: 0.5rem;
  color: var(--color-primary);
}

.terminal-input {
  flex: 1;
  background: transparent;
  border: none;
  color: v-bind('terminalSettings.textColor');
  font-family: v-bind('terminalSettings.fontFamily');
  font-size: v-bind('terminalSettings.fontSize + "px"');
  outline: none;
}

.terminal-input::placeholder {
  color: var(--color-text-secondary);
  opacity: 0.6;
}

.command-suggestions {
  position: absolute;
  bottom: 100%;
  left: 1rem;
  right: 1rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  max-height: 200px;
  overflow-y: auto;
  z-index: 10;
}

.suggestion-item {
  padding: 0.5rem 1rem;
  cursor: pointer;
  color: var(--color-text);
  transition: background-color 0.2s ease;
}

.suggestion-item:hover,
.suggestion-item.active {
  background: var(--color-primary);
  color: var(--color-background);
}

.terminal-sidebar {
  background: var(--color-surface);
  border-left: 1px solid var(--color-border);
  padding: 1rem;
  overflow-y: auto;
}

.sidebar-section {
  margin-bottom: 2rem;
}

.sidebar-section h3 {
  margin: 0 0 1rem 0;
  color: var(--color-text);
  font-size: 1.1rem;
}

.history-list {
  max-height: 300px;
  overflow-y: auto;
}

.history-item {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 0.5rem;
  padding: 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  margin-bottom: 0.25rem;
  transition: background-color 0.2s ease;
}

.history-item:hover {
  background: var(--color-background);
}

.history-command {
  font-family: monospace;
  font-size: 0.9rem;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.history-status {
  font-size: 0.75rem;
  padding: 0.125rem 0.25rem;
  border-radius: 0.125rem;
  font-weight: bold;
}

.history-status.success {
  background: var(--color-success);
  color: var(--color-background);
}

.history-status.error {
  background: var(--color-error);
  color: var(--color-background);
}

.history-status.running {
  background: var(--color-warning);
  color: var(--color-background);
}

.history-time {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.quick-commands {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.quick-cmd-btn {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.quick-cmd-btn:hover {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 1rem;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.panel-header h3 {
  margin: 0;
  color: var(--color-text);
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 0.5rem;
}

.close-btn:hover {
  color: var(--color-text);
}

.panel-content {
  padding: 1.5rem;
}

.setting-group {
  margin-bottom: 1.5rem;
}

.setting-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--color-text);
  font-weight: 500;
}

.setting-group input,
.setting-group select {
  width: 100%;
  padding: 0.5rem;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  color: var(--color-text);
}

.setting-group input[type="range"] {
  padding: 0;
}

.setting-group input[type="checkbox"] {
  width: auto;
  margin-right: 0.5rem;
}

/* Responsive */
@media (max-width: 1024px) {
  .terminal-page {
    grid-template-columns: 1fr;
  }
  
  .terminal-sidebar {
    display: none;
  }
  
  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
}

/* Animations */
@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

.terminal-input {
  animation: blink 1s infinite;
}
</style>
