<template>
  <div
    v-if="visible"
    class="terminal-context-menu"
    :style="menuStyle"
    @click.stop
    @contextmenu.prevent
  >
    <!-- Copy Section -->
    <div class="terminal-context-menu__section">
      <div class="terminal-context-menu__label">Copy</div>

      <button
        v-if="hasSelection"
        class="terminal-context-menu__item"
        @click="handleCopySelection"
        :disabled="!canCopy"
      >
        <Icon name="carbon:copy" />
        <span>Copy Selection</span>
        <kbd class="terminal-context-menu__shortcut">{{ copyShortcut }}</kbd>
      </button>

      <button
        class="terminal-context-menu__item"
        @click="handleCopyAll"
        :disabled="!canCopyAll"
      >
        <Icon name="carbon:document" />
        <span>Copy All</span>
        <kbd class="terminal-context-menu__shortcut">Ctrl+Shift+A</kbd>
      </button>

      <button
        class="terminal-context-menu__item"
        @click="handleCopyOutput"
        :disabled="!hasOutput"
      >
        <Icon name="carbon:output" />
        <span>Copy Output</span>
      </button>
    </div>

    <!-- Paste Section -->
    <div class="terminal-context-menu__section">
      <div class="terminal-context-menu__label">Paste</div>

      <button
        class="terminal-context-menu__item"
        @click="handlePaste"
        :disabled="!canPaste || readonly"
      >
        <Icon name="carbon:paste" />
        <span>Paste</span>
        <kbd class="terminal-context-menu__shortcut">{{ pasteShortcut }}</kbd>
      </button>

      <button
        class="terminal-context-menu__item"
        @click="handlePasteWithConfirmation"
        :disabled="!canPaste || readonly"
      >
        <Icon name="carbon:paste-alt" />
        <span>Paste with Confirmation</span>
      </button>
    </div>

    <!-- Clear Section -->
    <div class="terminal-context-menu__section">
      <div class="terminal-context-menu__label">Clear</div>

      <button
        class="terminal-context-menu__item"
        @click="handleClearSelection"
        :disabled="!hasSelection"
      >
        <Icon name="carbon:close" />
        <span>Clear Selection</span>
        <kbd class="terminal-context-menu__shortcut">Escape</kbd>
      </button>

      <button
        class="terminal-context-menu__item terminal-context-menu__item--danger"
        @click="handleClearAll"
      >
        <Icon name="carbon:trash-can" />
        <span>Clear All</span>
      </button>
    </div>

    <!-- Advanced Section -->
    <div class="terminal-context-menu__section" v-if="showAdvanced">
      <div class="terminal-context-menu__label">Advanced</div>

      <button
        class="terminal-context-menu__item"
        @click="handleFind"
      >
        <Icon name="carbon:search" />
        <span>Find in Output...</span>
        <kbd class="terminal-context-menu__shortcut">Ctrl+F</kbd>
      </button>

      <button
        class="terminal-context-menu__item"
        @click="handleSaveAs"
      >
        <Icon name="carbon:download" />
        <span>Save As...</span>
      </button>

      <button
        class="terminal-context-menu__item"
        @click="handlePrint"
      >
        <Icon name="carbon:printer" />
        <span>Print...</span>
        <kbd class="terminal-context-menu__shortcut">Ctrl+P</kbd>
      </button>
    </div>

    <!-- History Section -->
    <div class="terminal-context-menu__section" v-if="clipboardHistory.length > 0">
      <div class="terminal-context-menu__label">Clipboard History</div>

      <div class="terminal-context-menu__history">
        <button
          v-for="(item, index) in recentHistory"
          :key="item.id"
          class="terminal-context-menu__history-item"
          @click="handleCopyFromHistory(item)"
          :title="item.content"
        >
          <div class="terminal-context-menu__history-preview">
            {{ formatHistoryPreview(item.content) }}
          </div>
          <div class="terminal-context-menu__history-meta">
            {{ formatHistoryMeta(item) }}
          </div>
        </button>
      </div>
    </div>

    <!-- Separator -->
    <div class="terminal-context-menu__separator"></div>

    <!-- Actions -->
    <div class="terminal-context-menu__section">
      <button
        class="terminal-context-menu__item"
        @click="handleFullscreen"
      >
        <Icon :name="isFullscreen ? 'carbon:view-off' : 'carbon:view'" />
        <span>{{ isFullscreen ? 'Exit Fullscreen' : 'Enter Fullscreen' }}</span>
        <kbd class="terminal-context-menu__shortcut">F11</kbd>
      </button>

      <button
        class="terminal-context-menu__item"
        @click="handleSettings"
      >
        <Icon name="carbon:settings" />
        <span>Terminal Settings...</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useClipboard } from '~/composables/useClipboard'
import { useClipboardHistory } from '~/composables/useClipboardHistory'
import type { ClipboardHistory } from '~/composables/useClipboard'

// Props
interface Props {
  visible: boolean
  x: number
  y: number
  hasSelection: boolean
  hasOutput: boolean
  readonly?: boolean
  terminal?: any
  output?: string[]
  showAdvanced?: boolean
  isFullscreen?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  readonly: false,
  showAdvanced: true,
  isFullscreen: false
})

// Emits
const emit = defineEmits<{
  close: []
  copySelection: []
  copyAll: []
  copyOutput: []
  paste: [text: string]
  clearSelection: []
  clearAll: []
  find: []
  saveAs: []
  print: []
  fullscreen: []
  settings: []
}>()

// Composables
const clipboard = useClipboard({
  enableNotifications: true
})

const { history: clipboardHistory, recentItems: recentHistory } = useClipboardHistory()

// Computed
const menuStyle = computed(() => ({
  position: 'fixed',
  left: `${props.x}px`,
  top: `${props.y}px`,
  zIndex: 9999
}))

const canCopy = computed(() => clipboard.canWrite.value && props.hasSelection)
const canCopyAll = computed(() => clipboard.canWrite.value && props.hasOutput)
const canPaste = computed(() => clipboard.canRead.value)

const copyShortcut = computed(() => {
  return navigator.platform.includes('Mac') ? '⌘C' : 'Ctrl+C'
})

const pasteShortcut = computed(() => {
  return navigator.platform.includes('Mac') ? '⌘V' : 'Ctrl+V'
})

// Methods
const handleCopySelection = async () => {
  if (!props.terminal) return

  const success = await clipboard.copyTerminalSelection(props.terminal)
  if (success) {
    emit('copySelection')
    closeMenu()
  }
}

const handleCopyAll = async () => {
  if (!props.output?.length) return

  const success = await clipboard.copyTerminalOutput(props.output)
  if (success) {
    emit('copyAll')
    closeMenu()
  }
}

const handleCopyOutput = async () => {
  if (!props.output?.length) return

  // Copy only the output without prompts
  const outputLines = props.output.filter(line => !line.includes('$') && !line.includes('>'))
  const success = await clipboard.copyTerminalOutput(outputLines)

  if (success) {
    emit('copyOutput')
    closeMenu()
  }
}

const handlePaste = async () => {
  if (!props.terminal || props.readonly) return

  const success = await clipboard.pasteToTerminal(props.terminal)
  if (success) {
    const text = await clipboard.readText()
    if (text) {
      emit('paste', text)
    }
    closeMenu()
  }
}

const handlePasteWithConfirmation = async () => {
  if (!props.terminal || props.readonly) return

  const text = await clipboard.readText()
  if (!text) return

  // Show confirmation dialog for large content
  if (text.length > 1000) {
    const confirmed = confirm(`Paste ${text.length} characters? This could be a large amount of text.`)
    if (!confirmed) return
  }

  const success = await clipboard.pasteToTerminal(props.terminal)
  if (success) {
    emit('paste', text)
    closeMenu()
  }
}

const handleClearSelection = () => {
  if (props.terminal) {
    props.terminal.clearSelection()
  }
  emit('clearSelection')
  closeMenu()
}

const handleClearAll = () => {
  if (confirm('Clear all terminal output? This cannot be undone.')) {
    if (props.terminal) {
      props.terminal.clear()
    }
    emit('clearAll')
    closeMenu()
  }
}

const handleFind = () => {
  emit('find')
  closeMenu()
}

const handleSaveAs = () => {
  emit('saveAs')
  closeMenu()
}

const handlePrint = () => {
  emit('print')
  closeMenu()
}

const handleFullscreen = () => {
  emit('fullscreen')
  closeMenu()
}

const handleSettings = () => {
  emit('settings')
  closeMenu()
}

const handleCopyFromHistory = async (item: ClipboardHistory) => {
  const success = await clipboard.writeText(item.content, 'history')
  if (success) {
    closeMenu()
  }
}

const formatHistoryPreview = (content: string): string => {
  const maxLength = 40
  const preview = content.replace(/\n/g, ' ').trim()
  return preview.length > maxLength ? preview.substring(0, maxLength) + '...' : preview
}

const formatHistoryMeta = (item: ClipboardHistory): string => {
  const lines = item.content.split('\n').length
  const time = new Date(item.timestamp)
  const now = new Date()
  const diffMs = now.getTime() - time.getTime()

  let timeAgo = ''
  if (diffMs < 60000) {
    timeAgo = 'just now'
  } else if (diffMs < 3600000) {
    timeAgo = `${Math.floor(diffMs / 60000)}m ago`
  } else {
    timeAgo = time.toLocaleTimeString()
  }

  return `${lines} line${lines > 1 ? 's' : ''} • ${timeAgo}`
}

const closeMenu = () => {
  emit('close')
}

// Keyboard shortcuts
const handleKeyDown = (event: KeyboardEvent) => {
  if (!props.visible) return

  switch (event.key) {
    case 'Escape':
      closeMenu()
      break
    case 'Enter':
      if (canCopy.value) {
        handleCopySelection()
      }
      break
  }
}

// Click outside handler
const handleClickOutside = (event: MouseEvent) => {
  if (props.visible) {
    const target = event.target as HTMLElement
    if (!target.closest('.terminal-context-menu')) {
      closeMenu()
    }
  }
}

// Lifecycle
onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
  document.removeEventListener('click', handleClickOutside)
})

// Watch for visibility changes to adjust menu position
watch(() => props.visible, (visible) => {
  if (visible) {
    // Ensure menu stays within viewport
    nextTick(() => {
      const menu = document.querySelector('.terminal-context-menu') as HTMLElement
      if (menu) {
        const rect = menu.getBoundingClientRect()

        // Adjust horizontal position if menu goes off screen
        if (props.x + rect.width > window.innerWidth) {
          menu.style.left = `${window.innerWidth - rect.width - 10}px`
        }

        // Adjust vertical position if menu goes off screen
        if (props.y + rect.height > window.innerHeight) {
          menu.style.top = `${window.innerHeight - rect.height - 10}px`
        }
      }
    })
  }
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.terminal-context-menu {
  @apply fixed bg-surface border border-border rounded-lg shadow-xl py-2 min-w-64;
  backdrop-filter: blur(8px);
  background: var(--terminal-context-menu-bg, rgba(30, 30, 30, 0.95));
  border: 1px solid var(--terminal-context-menu-border, rgba(255, 255, 255, 0.1));
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
}

.terminal-context-menu__section {
  @apply py-1;
}

.terminal-context-menu__section:not(:last-child) {
  border-bottom: 1px solid var(--terminal-context-menu-separator, rgba(255, 255, 255, 0.05));
}

.terminal-context-menu__label {
  @apply px-3 py-1 text-xs font-medium text-text-secondary uppercase tracking-wider;
  margin: 4px 0;
}

.terminal-context-menu__item {
  @apply w-full flex items-center justify-between px-3 py-2 text-sm text-text hover:bg-surface-hover transition-colors duration-150;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
}

.terminal-context-menu__item:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.terminal-context-menu__item--danger {
  @apply text-error hover:bg-error/10;
}

.terminal-context-menu__item:hover .terminal-context-menu__shortcut {
  @apply opacity-70;
}

.terminal-context-menu__shortcut {
  @apply text-xs text-text-secondary opacity-50;
  font-family: var(--font-family-mono);
  background: var(--terminal-context-menu-shortcut-bg, rgba(255, 255, 255, 0.1));
  padding: 2px 6px;
  border-radius: 4px;
}

.terminal-context-menu__separator {
  @apply mx-3 h-px bg-border;
  margin: 4px 0;
}

.terminal-context-menu__history {
  @apply max-h-48 overflow-y-auto;
}

.terminal-context-menu__history-item {
  @apply w-full text-left px-3 py-2 text-sm hover:bg-surface-hover transition-colors duration-150;
  border: none;
  background: none;
  cursor: pointer;
  display: block;
}

.terminal-context-menu__history-preview {
  @apply font-mono text-text truncate mb-1;
}

.terminal-context-menu__history-meta {
  @apply text-xs text-text-secondary;
}

/* Scrollbar styling for history */
.terminal-context-menu__history::-webkit-scrollbar {
  width: 6px;
}

.terminal-context-menu__history::-webkit-scrollbar-track {
  @apply bg-transparent;
}

.terminal-context-menu__history::-webkit-scrollbar-thumb {
  @apply bg-border rounded-full;
}

.terminal-context-menu__history::-webkit-scrollbar-thumb:hover {
  @apply bg-text-secondary;
}

/* Dark theme variations */
@media (prefers-color-scheme: dark) {
  .terminal-context-menu {
    background: var(--terminal-context-menu-bg, rgba(20, 20, 20, 0.98));
    border-color: var(--terminal-context-menu-border, rgba(255, 255, 255, 0.08));
  }

  .terminal-context-menu__item {
    color: var(--terminal-context-menu-item-color, #e0e0e0);
  }

  .terminal-context-menu__item:hover {
    background: var(--terminal-context-menu-item-hover, rgba(255, 255, 255, 0.05));
  }

  .terminal-context-menu__shortcut {
    background: var(--terminal-context-menu-shortcut-bg, rgba(255, 255, 255, 0.08));
    color: var(--terminal-context-menu-shortcut-color, #a0a0a0);
  }
}
</style>