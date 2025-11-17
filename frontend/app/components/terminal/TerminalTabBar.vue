<template>
  <div
    class="terminal-tab-bar"
    :class="{
      'terminal-tab-bar--vertical': direction === 'vertical',
      'terminal-tab-bar--horizontal': direction === 'horizontal'
    }"
    @contextmenu.prevent="handleBarContextMenu"
  >
    <!-- Tab List -->
    <div
      ref="tabListRef"
      class="terminal-tab-bar__list"
      :class="{
        'terminal-tab-bar__list--scrollable': tabs.length > maxVisibleTabs
      }"
      @wheel="handleWheel"
    >
      <!-- Draggable Tabs -->
      <div
        v-for="tab in sortedTabs"
        :key="tab.id"
        :class="{
          'terminal-tab-bar__tab': true,
          'terminal-tab-bar__tab--active': tab.active,
          'terminal-tab-bar__tab--dragging': draggingTabId === tab.id,
          'terminal-tab-bar__tab--pinned': tab.pinned,
          'terminal-tab-bar__tab--modified': tab.modified
        }"
        :data-tab-id="tab.id"
        draggable="true"
        @click="handleTabClick(tab)"
        @contextmenu.prevent="handleTabContextMenu($event, tab)"
        @dragstart="handleDragStart($event, tab)"
        @dragend="handleDragEnd"
        @dragover.prevent="handleDragOver($event, tab)"
        @drop.prevent="handleDrop($event, tab)"
        @dblclick="handleTabDoubleClick(tab)"
      >
        <!-- Tab Content -->
        <div class="terminal-tab-bar__tab-content">
          <!-- Status Indicator -->
          <div class="terminal-tab-bar__tab-status">
            <div
              v-if="getSessionStatus(tab.sessionId)"
              :class="`terminal-tab-bar__tab-indicator terminal-tab-bar__tab-indicator--${getSessionStatus(tab.sessionId)}`"
            ></div>
          </div>

          <!-- Tab Icon -->
          <div class="terminal-tab-bar__tab-icon" v-if="tab.icon">
            <Icon :name="tab.icon" />
          </div>

          <!-- Pin Indicator -->
          <div class="terminal-tab-bar__tab-pin" v-if="tab.pinned">
            <Icon name="carbon:pin" />
          </div>

          <!-- Tab Title -->
          <div class="terminal-tab-bar__tab-title">
            <span class="terminal-tab-bar__tab-title-text">{{ tab.title }}</span>
          </div>

          <!-- Tab Badge -->
          <div
            v-if="tab.badge"
            class="terminal-tab-bar__tab-badge"
            :class="`terminal-tab-bar__tab-badge--${getBadgeType(tab.badge)}`"
          >
            {{ tab.badge }}
          </div>

          <!-- Close Button -->
          <button
            v-if="showCloseButtons && !tab.pinned"
            class="terminal-tab-bar__tab-close"
            @click.stop="handleTabClose(tab)"
            :title="`Close ${tab.title}`"
          >
            <Icon name="carbon:close" />
          </button>
        </div>
      </div>

      <!-- Drop Zone for new tabs -->
      <div
        v-if="showNewTabButton"
        class="terminal-tab-bar__new-tab"
        @click="handleNewTab"
        :title="newTabTooltip"
      >
        <Icon name="carbon:add" />
      </div>
    </div>

    <!-- Scroll Buttons -->
    <div
      v-if="showScrollButtons && canScrollLeft"
      class="terminal-tab-bar__scroll terminal-tab-bar__scroll--left"
      @click="scrollLeft"
    >
      <Icon name="carbon:chevron-left" />
    </div>

    <div
      v-if="showScrollButtons && canScrollRight"
      class="terminal-tab-bar__scroll terminal-tab-bar__scroll--right"
      @click="scrollRight"
    >
      <Icon name="carbon:chevron-right" />
    </div>

    <!-- Context Menu -->
    <TerminalContextMenu
      :visible="showContextMenu"
      :x="contextMenuPosition.x"
      :y="contextMenuPosition.y"
      :has-selection="false"
      :has-output="false"
      :show-advanced="false"
      @close="closeContextMenu"
      @copy-selection="() => {}"
      @copy-all="() => {}"
      @paste="() => {}"
      @clear-selection="() => {}"
      @clear-all="() => {}"
      @find="() => {}"
      @save-as="() => {}"
      @print="() => {}"
      @fullscreen="() => {}"
      @settings="() => {}"
    />

    <!-- New Tab Dropdown -->
    <div
      v-if="showNewTabDropdown"
      class="terminal-tab-bar__new-tab-dropdown"
      :style="newTabDropdownStyle"
      @click.stop
    >
      <div class="terminal-tab-bar__dropdown-section">
        <div class="terminal-tab-bar__dropdown-title">New Terminal</div>
        <button
          v-for="profile in availableProfiles"
          :key="profile.id"
          class="terminal-tab-bar__dropdown-item"
          @click="handleNewTabWithProfile(profile)"
        >
          <Icon :name="profile.icon || 'carbon:terminal'" />
          <span>{{ profile.name }}</span>
          <span class="terminal-tab-bar__dropdown-description">{{ profile.description }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useTerminalStore } from '~/stores/terminal'
import { useTerminal } from '~/composables/useTerminal'
import { useDragAndDrop } from '@vueuse/core'
import type { TerminalTab, TerminalProfile } from '~/types/terminal'

// Props
interface Props {
  tabs: TerminalTab[]
  direction?: 'horizontal' | 'vertical'
  showCloseButtons?: boolean
  showNewTabButton?: boolean
  showScrollButtons?: boolean
  maxVisibleTabs?: number
  allowReorder?: boolean
  allowDuplicate?: boolean
  showProfileSelection?: boolean
  persistSessions?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  direction: 'horizontal',
  showCloseButtons: true,
  showNewTabButton: true,
  showScrollButtons: true,
  maxVisibleTabs: 10,
  allowReorder: true,
  allowDuplicate: true,
  showProfileSelection: true,
  persistSessions: true
})

// Emits
const emit = defineEmits<{
  tabClick: [tab: TerminalTab]
  tabClose: [tab: TerminalTab]
  tabDuplicate: [tab: TerminalTab]
  tabActivate: [tab: TerminalTab]
  tabReorder: [fromIndex: number, toIndex: number]
  newTab: [profile?: TerminalProfile]
  contextMenu: [event: MouseEvent, tab?: TerminalTab]
}>()

// Composables
const terminalStore = useTerminalStore()
const { createSession } = useTerminal()

// Refs
const tabListRef = ref<HTMLElement>()
const draggingTabId = ref<string | null>(null)
const draggedOverTabId = ref<string | null>(null)

// Context Menu State
const showContextMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const contextMenuTab = ref<TerminalTab | null>(null)

// New Tab Dropdown State
const showNewTabDropdown = ref(false)
const newTabDropdownPosition = ref({ x: 0, y: 0 })

// Computed
const sortedTabs = computed(() => {
  return props.tabs.sort((a, b) => {
    // Pinned tabs first
    if (a.pinned && !b.pinned) return -1
    if (!a.pinned && b.pinned) return 1

    // Then by position
    return a.position - b.position
  })
})

const visibleTabs = computed(() => {
  return sortedTabs.value.slice(0, props.maxVisibleTabs)
})

const canScrollLeft = computed(() => {
  return props.tabs.length > props.maxVisibleTabs && tabListRef.value
})

const canScrollRight = computed(() => {
  return props.tabs.length > props.maxVisibleTabs && tabListRef.value
})

const availableProfiles = computed(() => {
  return terminalStore.getProfiles()
})

const newTabTooltip = computed(() => {
  return props.showProfileSelection ? 'New Terminal (Click for profiles)' : 'New Terminal'
})

const newTabDropdownStyle = computed(() => ({
  position: 'fixed',
  left: `${newTabDropdownPosition.value.x}px`,
  top: `${newTabDropdownPosition.value.y}px`,
  zIndex: 1000
}))

// Methods
const getSessionStatus = (sessionId: string): 'active' | 'inactive' | 'error' => {
  const session = terminalStore.getSession(sessionId)
  if (!session) return 'inactive'
  if (session.active) return 'active'
  return 'inactive'
}

const getBadgeType = (badge: string | number): 'default' | 'error' | 'warning' | 'success' | 'count' => {
  const badgeStr = badge.toString().toLowerCase()
  if (badgeStr.includes('error') || badgeStr.includes('✗')) return 'error'
  if (badgeStr.includes('warning') || badgeStr.includes('⚠')) return 'warning'
  if (badgeStr.includes('success') || badgeStr.includes('✓')) return 'success'
  if (badgeStr.match(/^\d+$/) && parseInt(badgeStr) > 0) return 'count'
  return 'default'
}

const handleTabClick = (tab: TerminalTab) => {
  emit('tabClick', tab)
  emit('tabActivate', tab)
}

const handleTabDoubleClick = (tab: TerminalTab) => {
  // Double-click to rename tab
  // This could open a rename dialog or inline editing
}

const handleTabClose = (tab: TerminalTab) => {
  emit('tabClose', tab)
}

const handleTabContextMenu = (event: MouseEvent, tab: TerminalTab) => {
  event.preventDefault()
  contextMenuTab.value = tab
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  showContextMenu.value = true
}

const handleBarContextMenu = (event: MouseEvent) => {
  event.preventDefault()
  contextMenuTab.value = null
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  showContextMenu.value = true
}

const closeContextMenu = () => {
  showContextMenu.value = false
  contextMenuTab.value = null
}

// Drag and Drop Handlers
const handleDragStart = (event: DragEvent, tab: TerminalTab) => {
  if (!props.allowReorder) return

  draggingTabId.value = tab.id
  event.dataTransfer?.setData('text/plain', tab.id)
  event.dataTransfer?.setDragImage(event.target as HTMLElement, 0, 0)
}

const handleDragEnd = () => {
  draggingTabId.value = null
  draggedOverTabId.value = null
}

const handleDragOver = (event: DragEvent, tab: TerminalTab) => {
  if (!props.allowReorder || draggingTabId.value === tab.id) return

  event.preventDefault()
  draggedOverTabId.value = tab.id
}

const handleDrop = (event: DragEvent, targetTab: TerminalTab) => {
  event.preventDefault()

  if (!draggingTabId.value || !props.allowReorder) return

  const draggedTab = props.tabs.find(tab => tab.id === draggingTabId.value)
  if (!draggedTab || draggedTab.id === targetTab.id) return

  const fromIndex = props.tabs.findIndex(tab => tab.id === draggingTabId.value)
  const toIndex = props.tabs.findIndex(tab => tab.id === targetTab.id)

  if (fromIndex !== -1 && toIndex !== -1) {
    emit('tabReorder', fromIndex, toIndex)
  }

  draggingTabId.value = null
  draggedOverTabId.value = null
}

// New Tab Handlers
const handleNewTab = (event?: MouseEvent) => {
  if (event && props.showProfileSelection) {
    // Show profile dropdown
    const rect = (event.target as HTMLElement).getBoundingClientRect()
    newTabDropdownPosition.value = {
      x: rect.left,
      y: rect.bottom + 4
    }
    showNewTabDropdown.value = true
  } else {
    emit('newTab')
  }
}

const handleNewTabWithProfile = (profile: TerminalProfile) => {
  emit('newTab', profile)
  showNewTabDropdown.value = false
}

// Scroll Handlers
const scrollLeft = () => {
  if (tabListRef.value) {
    tabListRef.value.scrollBy({
      left: -200,
      behavior: 'smooth'
    })
  }
}

const scrollRight = () => {
  if (tabListRef.value) {
    tabListRef.value.scrollBy({
      left: 200,
      behavior: 'smooth'
    })
  }
}

const handleWheel = (event: WheelEvent) => {
  if (!props.showScrollButtons || !tabListRef.value) return

  event.preventDefault()
  tabListRef.value.scrollBy({
    left: -event.deltaY,
    behavior: 'smooth'
  })
}

// Click outside handlers
const handleClickOutside = (event: MouseEvent) => {
  if (showContextMenu.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.terminal-tab-bar') && !target.closest('.terminal-context-menu')) {
      closeContextMenu()
    }
  }

  if (showNewTabDropdown.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.terminal-tab-bar__new-tab-dropdown') && !target.closest('.terminal-tab-bar__new-tab')) {
      showNewTabDropdown.value = false
    }
  }
}

// Session Persistence
const saveSessionState = () => {
  if (!props.persistSessions) return

  try {
    const state = {
      tabs: props.tabs,
      activeTabId: props.tabs.find(tab => tab.active)?.id,
      timestamp: Date.now()
    }
    localStorage.setItem('terminal-tab-state', JSON.stringify(state))
  } catch (error) {
    console.warn('Failed to save tab state:', error)
  }
}

const loadSessionState = () => {
  if (!props.persistSessions) return

  try {
    const saved = localStorage.getItem('terminal-tab-state')
    if (saved) {
      const state = JSON.parse(saved)
      // Emit loaded state to parent component
      // This would be handled by the parent component
    }
  } catch (error) {
    console.warn('Failed to load tab state:', error)
  }
}

// Lifecycle
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadSessionState()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  saveSessionState()
})

// Watch for tab changes and persist
watch(() => props.tabs, () => {
  saveSessionState()
}, { deep: true })
</script>

<style scoped>
@reference "../../assets/css/main.css";
.terminal-tab-bar {
  @apply flex items-center bg-surface border-b border-border;
  min-height: 2.5rem;
  position: relative;
}

.terminal-tab-bar--horizontal {
  @apply flex-row;
}

.terminal-tab-bar--vertical {
  @apply flex-col w-48;
  border-right: 1px solid var(--color-border);
  border-bottom: none;
}

.terminal-tab-bar__list {
  @apply flex items-center flex-1 overflow-hidden;
  scroll-behavior: smooth;
}

.terminal-tab-bar__list--scrollable {
  @apply overflow-x-auto;
  scrollbar-width: none;
}

.terminal-tab-bar__list--scrollable::-webkit-scrollbar {
  display: none;
}

.terminal-tab-bar__tab {
  @apply relative flex items-center px-3 py-2 cursor-pointer select-none transition-all duration-200;
  min-width: 0;
  max-width: 200px;
  border-bottom: 2px solid transparent;
}

.terminal-tab-bar__tab:hover {
  @apply bg-surface-hover;
}

.terminal-tab-bar__tab--active {
  @apply bg-surface-active;
  border-bottom-color: var(--color-primary);
  color: var(--color-text-primary);
}

.terminal-tab-bar__tab--dragging {
  @apply opacity-50;
}

.terminal-tab-bar__tab--pinned {
  @apply opacity-80;
}

.terminal-tab-bar__tab--modified::after {
  content: '';
  @apply absolute top-1 right-1 w-2 h-2 bg-primary rounded-full;
}

.terminal-tab-bar__tab-content {
  @apply flex items-center gap-2 flex-1 min-w-0;
}

.terminal-tab-bar__tab-status {
  @apply flex-shrink-0;
}

.terminal-tab-bar__tab-indicator {
  @apply w-2 h-2 rounded-full;
}

.terminal-tab-bar__tab-indicator--active {
  @apply bg-green-500;
}

.terminal-tab-bar__tab-indicator--inactive {
  @apply bg-gray-500;
}

.terminal-tab-bar__tab-icon {
  @apply flex-shrink-0 text-sm;
}

.terminal-tab-bar__tab-pin {
  @apply flex-shrink-0 text-xs text-warning;
}

.terminal-tab-bar__tab-title {
  @apply flex-1 min-w-0;
}

.terminal-tab-bar__tab-title-text {
  @apply truncate text-sm font-medium;
}

.terminal-tab-bar__tab-badge {
  @apply inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-medium rounded-full;
  min-width: 1.25rem;
  height: 1.25rem;
}

.terminal-tab-bar__tab-badge--default {
  @apply bg-gray-200 text-gray-700;
}

.terminal-tab-bar__tab-badge--error {
  @apply bg-red-100 text-red-700;
}

.terminal-tab-bar__tab-badge--warning {
  @apply bg-yellow-100 text-yellow-700;
}

.terminal-tab-bar__tab-badge--success {
  @apply bg-green-100 text-green-700;
}

.terminal-tab-bar__tab-badge--count {
  @apply bg-blue-100 text-blue-700;
}

.terminal-tab-bar__tab-close {
  @apply p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-error transition-colors duration-200;
  border: none;
  background: none;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.terminal-tab-bar__tab:hover .terminal-tab-bar__tab-close {
  opacity: 1;
}

.terminal-tab-bar__new-tab {
  @apply flex items-center justify-center p-2 hover:bg-surface-hover cursor-pointer transition-colors duration-200;
  border: none;
  background: none;
  color: var(--color-text-secondary);
}

.terminal-tab-bar__new-tab:hover {
  color: var(--color-text-primary);
}

.terminal-tab-bar__scroll {
  @apply flex items-center justify-center p-1 hover:bg-surface-hover cursor-pointer transition-colors duration-200;
  border: none;
  background: none;
}

.terminal-tab-bar__scroll--left {
  @apply border-r border-border;
}

.terminal-tab-bar__scroll--right {
  @apply border-l border-border;
}

.terminal-tab-bar__new-tab-dropdown {
  @apply absolute bg-surface border border-border rounded-lg shadow-xl py-2 min-w-64;
  backdrop-filter: blur(8px);
}

.terminal-tab-bar__dropdown-section {
  @apply py-1;
}

.terminal-tab-bar__dropdown-title {
  @apply px-3 py-1 text-xs font-medium text-text-secondary uppercase tracking-wider;
  margin: 4px 0;
}

.terminal-tab-bar__dropdown-item {
  @apply w-full flex items-start gap-3 px-3 py-2 text-sm text-text hover:bg-surface-hover transition-colors duration-150;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.terminal-tab-bar__dropdown-description {
  @apply text-xs text-text-secondary;
  margin-top: 2px;
  line-height: 1.2;
}

/* Vertical orientation styles */
.terminal-tab-bar--vertical .terminal-tab-bar__list {
  @apply flex-col overflow-y-auto;
}

.terminal-tab-bar--vertical .terminal-tab-bar__tab {
  @apply w-full border-b-2 border-l-2 border-transparent;
  border-bottom: none;
}

.terminal-tab-bar--vertical .terminal-tab-bar__tab--active {
  @apply border-l-primary;
}

.terminal-tab-bar--vertical .terminal-tab-bar__tab-close {
  @apply opacity-100;
}

/* Dark theme adaptations */
@media (prefers-color-scheme: dark) {
  .terminal-tab-bar {
    @apply bg-surface;
  }

  .terminal-tab-bar__tab-badge--default {
    @apply bg-gray-700 text-gray-300;
  }

  .terminal-tab-bar__tab-badge--error {
    @apply bg-red-900 text-red-300;
  }

  .terminal-tab-bar__tab-badge--warning {
    @apply bg-yellow-900 text-yellow-300;
  }

  .terminal-tab-bar__tab-badge--success {
    @apply bg-green-900 text-green-300;
  }

  .terminal-tab-bar__tab-badge--count {
    @apply bg-blue-900 text-blue-300;
  }
}
</style>