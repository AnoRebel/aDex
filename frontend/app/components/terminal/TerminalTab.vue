<template>
  <div
    class="terminal-tab"
    :class="{
      'terminal-tab--active': isActive,
      'terminal-tab--modified': isModified,
      'terminal-tab--pinned': isPinned,
      'terminal-tab--dragging': isDragging
    }"
    :title="tooltip"
    @click="handleClick"
    @contextmenu.prevent="handleContextMenu"
    @mousedown="handleMouseDown"
    @dragstart="handleDragStart"
    @dragend="handleDragEnd"
    @dragover.prevent
    @drop.prevent="handleDrop"
    draggable="true"
  >
    <!-- Tab Status Indicator -->
    <div class="terminal-tab__status">
      <div
        v-if="session?.active"
        class="terminal-tab__indicator terminal-tab__indicator--active"
      ></div>
      <div
        v-else
        class="terminal-tab__indicator terminal-tab__indicator--inactive"
      ></div>
    </div>

    <!-- Tab Icon -->
    <div class="terminal-tab__icon" v-if="icon">
      <Icon :name="icon" />
    </div>

    <!-- Pin Indicator -->
    <div class="terminal-tab__pin" v-if="isPinned">
      <Icon name="carbon:pin" />
    </div>

    <!-- Tab Title -->
    <div class="terminal-tab__title">
      <span class="terminal-tab__title-text">{{ displayTitle }}</span>
      <span
        v-if="badge"
        class="terminal-tab__badge"
        :class="`terminal-tab__badge--${badgeType}`"
      >
        {{ badge }}
      </span>
    </div>

    <!-- Tab Actions -->
    <div class="terminal-tab__actions" v-if="showActions || isHovered">
      <button
        v-if="allowDuplicate && !isPinned"
        class="terminal-tab__action terminal-tab__action--duplicate"
        @click.stop="handleDuplicate"
        title="Duplicate tab"
      >
        <Icon name="carbon:copy" />
      </button>

      <button
        class="terminal-tab__action terminal-tab__action--close"
        @click.stop="handleClose"
        :title="closeTooltip"
        :disabled="isPinned && !forceClose"
      >
        <Icon :name="closeIcon" />
      </button>
    </div>

    <!-- Progress Bar (for long-running operations) -->
    <div
      v-if="showProgress"
      class="terminal-tab__progress"
      :style="{ width: `${progress}%` }"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { TerminalTab } from '~/composables/useTerminal'

// Props
interface Props {
  tab: TerminalTab
  isActive?: boolean
  showActions?: boolean
  allowDuplicate?: boolean
  allowReorder?: boolean
  allowClose?: boolean
  forceClose?: boolean
  maxTitleLength?: number
  showSessionStatus?: boolean
  showProgress?: boolean
  progress?: number
}

const props = withDefaults(defineProps<Props>(), {
  isActive: false,
  showActions: false,
  allowDuplicate: true,
  allowReorder: true,
  allowClose: true,
  forceClose: false,
  maxTitleLength: 30,
  showSessionStatus: true,
  showProgress: false,
  progress: 0
})

// Emits
const emit = defineEmits<{
  click: [tab: TerminalTab]
  close: [tab: TerminalTab]
  duplicate: [tab: TerminalTab]
  contextMenu: [event: MouseEvent, tab: TerminalTab]
  reorder: [fromIndex: number, toIndex: number]
  activate: [tab: TerminalTab]
}>()

// Refs
const isHovered = ref(false)
const isDragging = ref(false)

// Computed
const session = computed(() => props.tab)

const isModified = computed(() => props.tab.modified)

const isPinned = computed(() => props.tab.pinned)

const icon = computed(() => {
  if (props.tab.icon) return props.tab.icon

  // Return icon based on session properties
  if (!session.value) return 'carbon:terminal'

  if (session.value.user === 'root') return 'carbon:shield'
  if (session.value.shell?.includes('zsh')) return 'carbon:face-dissatisfied'
  if (session.value.shell?.includes('bash')) return 'carbon:face-smile'
  if (session.value.shell?.includes('fish')) return 'carbon:face-dizzy'

  return 'carbon:terminal'
})

const badge = computed(() => props.tab.badge)

const badgeType = computed(() => {
  if (!badge.value) return 'default'

  const badgeStr = badge.value.toString().toLowerCase()
  if (badgeStr.includes('error') || badgeStr.includes('✗')) return 'error'
  if (badgeStr.includes('warning') || badgeStr.includes('⚠')) return 'warning'
  if (badgeStr.includes('success') || badgeStr.includes('✓')) return 'success'
  if (badgeStr.match(/^\d+$/) && parseInt(badgeStr) > 0) return 'count'

  return 'default'
})

const displayTitle = computed(() => {
  let title = props.tab.title

  // Truncate if too long
  if (title.length > props.maxTitleLength) {
    title = title.substring(0, props.maxTitleLength - 3) + '...'
  }

  return title
})

const tooltip = computed(() => {
  const parts = []

  if (session.value) {
    parts.push(`Session: ${session.value.id.substring(0, 8)}`)

    if (session.value.cwd) {
      parts.push(`Directory: ${session.value.cwd}`)
    }

    if (session.value.shell) {
      parts.push(`Shell: ${session.value.shell}`)
    }

    if (session.value.pid) {
      parts.push(`PID: ${session.value.pid}`)
    }

    if (session.value.user) {
      parts.push(`User: ${session.value.user}`)
    }

    parts.push(`Last activity: ${formatTime(session.value.lastSeen)}`)
  }

  parts.push(`Title: ${props.tab.title}`)

  if (isPinned.value) {
    parts.push('Pinned')
  }

  return parts.join('\n')
})

const closeIcon = computed(() => {
  if (isHovered.value) return 'carbon:close-filled'
  return 'carbon:close'
})

const closeTooltip = computed(() => {
  if (isPinned.value && !props.forceClose) {
    return 'Cannot close pinned tab (hold Shift to force)'
  }
  return 'Close tab'
})

// Methods
const handleClick = () => {
  emit('click', props.tab)
  emit('activate', props.tab)
}

const handleContextMenu = (event: MouseEvent) => {
  emit('contextMenu', event, props.tab)
}

const handleMouseDown = (event: MouseEvent) => {
  if (event.button === 1) { // Middle mouse button
    event.preventDefault()
    handleClose()
  }

  isHovered.value = true
}

const handleMouseUp = () => {
  isHovered.value = false
}

const handleClose = () => {
  if (!props.allowClose) return

  if (isPinned.value && !props.forceClose) {
    // Check if Shift key is held for force close
    if (window.event && !(window.event as MouseEvent).shiftKey) {
      return
    }
  }

  emit('close', props.tab)
}

const handleDuplicate = () => {
  if (props.allowDuplicate && !isPinned.value) {
    emit('duplicate', props.tab)
  }
}

const handleDragStart = (event: DragEvent) => {
  if (!props.allowReorder) {
    event.preventDefault()
    return
  }

  isDragging.value = true

  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', props.tab.id)

    // Create a custom drag image
    const dragImage = event.currentTarget?.cloneNode(true) as HTMLElement
    if (dragImage) {
      dragImage.style.opacity = '0.5'
      dragImage.style.transform = 'rotate(2deg)'
      document.body.appendChild(dragImage)
      event.dataTransfer.setDragImage(dragImage, event.offsetX, event.offsetY)
      setTimeout(() => document.body.removeChild(dragImage), 0)
    }
  }
}

const handleDragEnd = () => {
  isDragging.value = false
}

const handleDrop = (event: DragEvent) => {
  event.preventDefault()

  if (!props.allowReorder || !event.dataTransfer) return

  const draggedTabId = event.dataTransfer.getData('text/plain')
  if (draggedTabId && draggedTabId !== props.tab.id) {
    // Find the index of the dragged tab and this tab
    // This would need to be handled by the parent component
    // For now, just emit the event
    emit('reorder', 0, 0) // Placeholder indices
  }
}

const formatTime = (timeString?: string) => {
  if (!timeString) return 'Unknown'

  try {
    const date = new Date(timeString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()

    if (diffMs < 60000) { // Less than 1 minute
      return 'Just now'
    } else if (diffMs < 3600000) { // Less than 1 hour
      const minutes = Math.floor(diffMs / 60000)
      return `${minutes} minute${minutes > 1 ? 's' : ''} ago`
    } else if (diffMs < 86400000) { // Less than 1 day
      const hours = Math.floor(diffMs / 3600000)
      return `${hours} hour${hours > 1 ? 's' : ''} ago`
    } else {
      return date.toLocaleDateString()
    }
  } catch {
    return 'Unknown'
  }
}

// Lifecycle
onMounted(() => {
  document.addEventListener('mouseup', handleMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('mouseup', handleMouseUp)
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.terminal-tab {
  @apply relative flex items-center gap-2 px-3 py-2 cursor-pointer select-none transition-all duration-200;
  @apply bg-surface hover:bg-surface-hover border-b-2 border-transparent;
  @apply min-w-0 max-w-48;
  user-select: none;
}

.terminal-tab--active {
  @apply bg-surface-active border-primary;
  color: var(--terminal-tab-active-color, var(--color-text-primary));
}

.terminal-tab--modified {
  font-weight: 500;
}

.terminal-tab--modified::after {
  content: '';
  @apply absolute top-1 right-1 w-2 h-2 bg-primary rounded-full;
}

.terminal-tab--pinned {
  opacity: 0.8;
}

.terminal-tab--dragging {
  opacity: 0.5;
  transform: rotate(2deg);
}

.terminal-tab__status {
  @apply flex-shrink-0;
}

.terminal-tab__indicator {
  @apply w-2 h-2 rounded-full transition-colors duration-200;
}

.terminal-tab__indicator--active {
  @apply bg-green-500;
}

.terminal-tab__indicator--inactive {
  @apply bg-gray-500;
}

.terminal-tab__icon {
  @apply flex-shrink-0 text-sm;
  color: var(--terminal-tab-icon-color, var(--color-text-secondary));
}

.terminal-tab--active .terminal-tab__icon {
  color: var(--terminal-tab-active-icon-color, var(--color-text-primary));
}

.terminal-tab__pin {
  @apply flex-shrink-0 text-xs;
  color: var(--terminal-tab-pin-color, var(--color-warning));
}

.terminal-tab__title {
  @apply flex-1 min-w-0 flex items-center gap-2;
}

.terminal-tab__title-text {
  @apply truncate text-sm font-medium;
}

.terminal-tab__badge {
  @apply inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-medium rounded-full;
  min-width: 1.25rem;
  height: 1.25rem;
}

.terminal-tab__badge--default {
  @apply bg-gray-200 text-gray-700;
}

.terminal-tab__badge--error {
  @apply bg-red-100 text-red-700;
}

.terminal-tab__badge--warning {
  @apply bg-yellow-100 text-yellow-700;
}

.terminal-tab__badge--success {
  @apply bg-green-100 text-green-700;
}

.terminal-tab__badge--count {
  @apply bg-blue-100 text-blue-700;
}

.terminal-tab__actions {
  @apply flex items-center gap-1 opacity-0 transition-opacity duration-200;
}

.terminal-tab:hover .terminal-tab__actions,
.terminal-tab--active .terminal-tab__actions {
  opacity: 1;
}

.terminal-tab__action {
  @apply p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-text transition-colors duration-200;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
}

.terminal-tab__action:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.terminal-tab__action--close:hover {
  @apply bg-error text-white;
}

.terminal-tab__action--duplicate:hover {
  @apply bg-primary text-white;
}

.terminal-tab__progress {
  @apply absolute bottom-0 left-0 h-0.5 bg-primary transition-all duration-300;
}

/* Dark theme adaptations */
@media (prefers-color-scheme: dark) {
  .terminal-tab {
    @apply bg-surface border-transparent;
  }

  .terminal-tab__badge--default {
    @apply bg-gray-700 text-gray-300;
  }

  .terminal-tab__badge--error {
    @apply bg-red-900 text-red-300;
  }

  .terminal-tab__badge--warning {
    @apply bg-yellow-900 text-yellow-300;
  }

  .terminal-tab__badge--success {
    @apply bg-green-900 text-green-300;
  }

  .terminal-tab__badge--count {
    @apply bg-blue-900 text-blue-300;
  }
}
</style>