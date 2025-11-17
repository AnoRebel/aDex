<template>
  <div class="file-list" :class="{ 'is-loading': isLoading }">
    <!-- Loading overlay -->
    <div v-if="isLoading" class="loading-overlay">
      <div class="loading-spinner">
        <div class="spinner"></div>
        <span class="loading-text">Loading files...</span>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredFiles.length === 0" class="empty-state">
      <div class="empty-icon">📁</div>
      <h3 class="empty-title">No files found</h3>
      <p class="empty-description">
        {{ searchQuery ? 'No files match your search criteria.' : 'This directory is empty.' }}
      </p>
      <button v-if="searchQuery" @click="$emit('clear-search')" class="clear-search-btn">
        Clear search
      </button>
    </div>

    <!-- Grid view -->
    <div v-else-if="viewMode === 'grid'" class="grid-view">
      <VirtualScroller
        :items="filteredFiles"
        :item-height="120"
        :buffer="10"
        class="grid-scroller"
      >
        <template #default="{ item, index }">
          <div
            :key="item.path"
            class="file-item grid-item"
            :class="{
              'is-selected': selectedFiles.includes(item.path),
              'is-directory': item.is_dir,
              'is-hidden': item.name.startsWith('.')
            }"
            @click="handleClick(item, $event)"
            @dblclick="handleDoubleClick(item)"
            @contextmenu="handleContextMenu(item, $event)"
            @mouseenter="handleMouseEnter(item)"
            @mouseleave="handleMouseLeave(item)"
          >
            <!-- File icon/preview -->
            <div class="file-icon-container">
              <FileIcon
                :file="item"
                :size="48"
                :show-preview="showPreviews"
                class="file-icon"
              />
              <div class="selection-indicator" v-if="selectedFiles.includes(item.path)">
                <svg width="16" height="16" viewBox="0 0 16 16">
                  <path fill="currentColor" d="M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z"/>
                </svg>
              </div>
            </div>

            <!-- File name -->
            <div class="file-name" :title="getFullName(item)">
              {{ item.name }}
            </div>

            <!-- File metadata -->
            <div class="file-metadata">
              <span class="file-size">{{ formatFileSize(item) }}</span>
              <span class="file-date">{{ formatDate(item) }}</span>
            </div>

            <!-- File tags or badges -->
            <div class="file-badges" v-if="getFileBadges(item).length > 0">
              <span
                v-for="badge in getFileBadges(item)"
                :key="badge.type"
                class="file-badge"
                :class="`badge-${badge.type}`"
                :title="badge.title"
              >
                {{ badge.text }}
              </span>
            </div>
          </div>
        </template>
      </VirtualScroller>
    </div>

    <!-- List view -->
    <div v-else class="list-view">
      <div class="list-header">
        <div class="list-header-row">
          <div
            v-for="column in visibleColumns"
            :key="column.key"
            class="list-header-cell"
            :class="{
              'is-sortable': column.sortable,
              'is-sorted': sortField === column.key,
              'is-ascending': sortOrder === 'asc'
            }"
            :style="{ width: column.width }"
            @click="column.sortable ? handleSort(column.key) : null"
          >
            <span class="column-title">{{ column.title }}</span>
            <svg
              v-if="column.sortable && sortField === column.key"
              class="sort-icon"
              width="12"
              height="12"
              viewBox="0 0 12 12"
            >
              <path
                fill="currentColor"
                :d="sortOrder === 'asc'
                  ? 'M3 4.5L6 1.5L9 4.5H7V8.5H5V4.5H3Z'
                  : 'M3 7.5L6 10.5L9 7.5H7V3.5H5V7.5H3Z'"
              />
            </svg>
          </div>
        </div>
      </div>

      <VirtualScroller
        :items="filteredFiles"
        :item-height="32"
        :buffer="20"
        class="list-scroller"
      >
        <template #default="{ item, index }">
          <div
            :key="item.path"
            class="file-item list-item"
            :class="{
              'is-selected': selectedFiles.includes(item.path),
              'is-directory': item.is_dir,
              'is-hidden': item.name.startsWith('.'),
              'is-focused': focusedItem === item.path
            }"
            @click="handleClick(item, $event)"
            @dblclick="handleDoubleClick(item)"
            @contextmenu="handleContextMenu(item, $event)"
            @mouseenter="handleMouseEnter(item)"
            @mouseleave="handleMouseLeave(item)"
          >
            <!-- Checkbox for multi-selection -->
            <div class="file-checkbox">
              <input
                type="checkbox"
                :checked="selectedFiles.includes(item.path)"
                @change="handleCheckboxChange(item, $event)"
                @click.stop
              />
            </div>

            <!-- File icon -->
            <div class="file-icon-container">
              <FileIcon
                :file="item"
                :size="16"
                class="file-icon"
              />
            </div>

            <!-- File name -->
            <div class="file-name-cell" :title="getFullName(item)">
              <span class="file-name">{{ item.name }}</span>
            </div>

            <!-- File size -->
            <div class="file-size-cell" v-if="isColumnVisible('size')">
              <span class="file-size">{{ formatFileSize(item) }}</span>
            </div>

            <!-- Modified date -->
            <div class="file-date-cell" v-if="isColumnVisible('modified')">
              <span class="file-date">{{ formatDate(item) }}</span>
            </div>

            <!-- File type -->
            <div class="file-type-cell" v-if="isColumnVisible('type')">
              <span class="file-type">{{ getFileType(item) }}</span>
            </div>

            <!-- Permissions -->
            <div class="file-permissions-cell" v-if="isColumnVisible('permissions')">
              <span class="file-permissions" :title="item.permissions">
                {{ formatPermissions(item) }}
              </span>
            </div>

            <!-- Owner -->
            <div class="file-owner-cell" v-if="isColumnVisible('owner')">
              <span class="file-owner">{{ item.owner || '-' }}</span>
            </div>

            <!-- Actions -->
            <div class="file-actions-cell">
              <div class="file-actions" @click.stop>
                <button
                  @click="handleAction('open', item)"
                  class="action-btn"
                  title="Open"
                >
                  <svg width="14" height="14" viewBox="0 0 14 14">
                    <path fill="currentColor" d="M2 2h10v10H2V2zm2 2v6h6V4H4z"/>
                  </svg>
                </button>
                <button
                  @click="handleAction('menu', item)"
                  class="action-btn"
                  title="More options"
                >
                  <svg width="14" height="14" viewBox="0 0 14 14">
                    <circle cx="7" cy="7" r="1"/>
                    <circle cx="3" cy="7" r="1"/>
                    <circle cx="11" cy="7" r="1"/>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </template>
      </VirtualScroller>
    </div>

    <!-- Context menu -->
    <ContextMenu
      v-if="contextMenu.visible"
      :x="contextMenu.x"
      :y="contextMenu.y"
      :file="contextMenu.file"
      :selected-files="selectedFiles"
      @close="closeContextMenu"
      @action="handleContextMenuAction"
    />

    <!-- File preview tooltip -->
    <div
      v-if="previewTooltip.visible"
      class="preview-tooltip"
      :style="{
        left: previewTooltip.x + 'px',
        top: previewTooltip.y + 'px'
      }"
      @mouseenter="keepPreviewTooltip"
      @mouseleave="hidePreviewTooltip"
    >
      <FilePreview
        :file="previewTooltip.file"
        :size="previewTooltip.size"
        :loading="previewTooltip.loading"
        @error="handlePreviewError"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import type { FileSystemEntry } from '~/types/filesystem'
import FileIcon from './FileIcon.vue'
import FilePreview from './FilePreview.vue'
import ContextMenu from './ContextMenu.vue'
import VirtualScroller from '~/components/ui/VirtualScroller.vue'

// Props
interface Props {
  files: FileSystemEntry[]
  selectedFiles: string[]
  viewMode: 'grid' | 'list'
  sortField: string
  sortOrder: 'asc' | 'desc'
  searchQuery?: string
  isLoading?: boolean
  showHidden?: boolean
  showPreviews?: boolean
  multiSelect?: boolean
  columns?: Array<{
    key: string
    title: string
    width: string
    sortable: boolean
    visible: boolean
  }>
}

const props = withDefaults(defineProps<Props>(), {
  searchQuery: '',
  isLoading: false,
  showHidden: false,
  showPreviews: true,
  multiSelect: true,
  columns: () => [
    { key: 'name', title: 'Name', width: 'flex: 1', sortable: true, visible: true },
    { key: 'size', title: 'Size', width: '100px', sortable: true, visible: true },
    { key: 'modified', title: 'Modified', width: '150px', sortable: true, visible: true },
    { key: 'type', title: 'Type', width: '80px', sortable: true, visible: false },
    { key: 'permissions', title: 'Permissions', width: '120px', sortable: true, visible: false },
    { key: 'owner', title: 'Owner', width: '100px', sortable: true, visible: false },
  ]
})

// Emits
const emit = defineEmits<{
  'file-select': [file: FileSystemEntry, event: MouseEvent]
  'file-open': [file: FileSystemEntry]
  'file-rename': [file: FileSystemEntry]
  'file-delete': [file: FileSystemEntry]
  'file-copy': [file: FileSystemEntry]
  'file-move': [file: FileSystemEntry]
  'sort-change': [field: string, order: 'asc' | 'desc']
  'context-menu': [file: FileSystemEntry, x: number, y: number]
  'clear-search': []
  'preview-request': [file: FileSystemEntry]
}>()

// Reactive state
const focusedItem = ref<string>('')
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  file: null as FileSystemEntry | null
})
const previewTooltip = ref({
  visible: false,
  x: 0,
  y: 0,
  file: null as FileSystemEntry | null,
  size: 'small' as 'small' | 'medium' | 'large',
  loading: false
})
const previewTooltipTimeout = ref<NodeJS.Timeout>()

// Computed properties
const filteredFiles = computed(() => {
  let result = [...props.files]

  // Apply search filter
  if (props.searchQuery) {
    const query = props.searchQuery.toLowerCase()
    result = result.filter(file =>
      file.name.toLowerCase().includes(query) ||
      file.path.toLowerCase().includes(query)
    )
  }

  // Apply hidden file filter
  if (!props.showHidden) {
    result = result.filter(file => !file.name.startsWith('.'))
  }

  // Apply sorting
  result.sort((a, b) => {
    let aValue: any = a[props.sortField as keyof FileSystemEntry]
    let bValue: any = b[props.sortField as keyof FileSystemEntry]

    // Handle different field types
    switch (props.sortField) {
      case 'name':
        aValue = aValue.toLowerCase()
        bValue = bValue.toLowerCase()
        break
      case 'size':
        aValue = Number(aValue) || 0
        bValue = Number(bValue) || 0
        break
      case 'mod_time':
        aValue = new Date(aValue).getTime()
        bValue = new Date(bValue).getTime()
        break
      default:
        aValue = String(aValue).toLowerCase()
        bValue = String(bValue).toLowerCase()
    }

    // Directories first
    if (a.is_dir && !b.is_dir) return -1
    if (!a.is_dir && b.is_dir) return 1

    // Apply sort order
    if (aValue < bValue) return props.sortOrder === 'asc' ? -1 : 1
    if (aValue > bValue) return props.sortOrder === 'asc' ? 1 : -1
    return 0
  })

  return result
})

const visibleColumns = computed(() => {
  return props.columns.filter(col => col.visible)
})

// Methods
const handleClick = (file: FileSystemEntry, event: MouseEvent) => {
  emit('file-select', file, event)
}

const handleDoubleClick = (file: FileSystemEntry) => {
  emit('file-open', file)
}

const handleContextMenu = (file: FileSystemEntry, event: MouseEvent) => {
  event.preventDefault()
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    file
  }
  emit('context-menu', file, event.clientX, event.clientY)
}

const handleMouseEnter = (file: FileSystemEntry) => {
  if (props.viewMode === 'list') {
    showPreviewTooltip(file)
  }
}

const handleMouseLeave = () => {
  hidePreviewTooltip()
}

const handleCheckboxChange = (file: FileSystemEntry, event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.checked) {
    // Selection is handled by parent
  } else {
    // Deselection is handled by parent
  }
}

const handleSort = (field: string) => {
  const newOrder = props.sortField === field && props.sortOrder === 'asc' ? 'desc' : 'asc'
  emit('sort-change', field, newOrder)
}

const handleAction = (action: string, file: FileSystemEntry) => {
  switch (action) {
    case 'open':
      emit('file-open', file)
      break
    case 'menu':
      // Show context menu at file position
      const element = document.querySelector(`[data-file-path="${file.path}"]`)
      if (element) {
        const rect = element.getBoundingClientRect()
        contextMenu.value = {
          visible: true,
          x: rect.right,
          y: rect.top,
          file
        }
      }
      break
  }
}

const closeContextMenu = () => {
  contextMenu.value.visible = false
}

const handleContextMenuAction = (action: string, file: FileSystemEntry) => {
  switch (action) {
    case 'open':
      emit('file-open', file)
      break
    case 'rename':
      emit('file-rename', file)
      break
    case 'delete':
      emit('file-delete', file)
      break
    case 'copy':
      emit('file-copy', file)
      break
    case 'move':
      emit('file-move', file)
      break
  }
  closeContextMenu()
}

const showPreviewTooltip = (file: FileSystemEntry) => {
  // Clear existing timeout
  if (previewTooltipTimeout.value) {
    clearTimeout(previewTooltipTimeout.value)
  }

  // Show tooltip after delay
  previewTooltipTimeout.value = setTimeout(() => {
    // Find mouse position or use file element position
    const element = document.querySelector(`[data-file-path="${file.path}"]`) as HTMLElement
    if (element) {
      const rect = element.getBoundingClientRect()
      previewTooltip.value = {
        visible: true,
        x: rect.right + 10,
        y: rect.top,
        file,
        size: 'small',
        loading: true
      }
      emit('preview-request', file)
    }
  }, 500)
}

const hidePreviewTooltip = () => {
  if (previewTooltipTimeout.value) {
    clearTimeout(previewTooltipTimeout.value)
  }
  previewTooltip.value.visible = false
}

const keepPreviewTooltip = () => {
  // Keep tooltip visible when mouse enters it
  if (previewTooltipTimeout.value) {
    clearTimeout(previewTooltipTimeout.value)
  }
}

const handlePreviewError = () => {
  previewTooltip.value.loading = false
}

const isColumnVisible = (key: string) => {
  return props.columns.some(col => col.key === key && col.visible)
}

const getFullName = (file: FileSystemEntry) => {
  return file.is_dir ? `${file.name}/` : file.name
}

const formatFileSize = (file: FileSystemEntry): string => {
  if (file.is_dir) return '—'

  const bytes = file.size
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`
}

const formatDate = (file: FileSystemEntry): string => {
  return new Intl.DateTimeFormat('en-US', {
    year: '2-digit',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(file.mod_time))
}

const getFileType = (file: FileSystemEntry): string => {
  if (file.is_dir) return 'Directory'

  const ext = file.name.split('.').pop()?.toLowerCase()
  return ext?.toUpperCase() || 'File'
}

const formatPermissions = (file: FileSystemEntry): string => {
  return file.mode?.toString(8).padStart(3, '0') || '---'
}

const getFileBadges = (file: FileSystemEntry) => {
  const badges = []

  if (file.is_hidden) {
    badges.push({ type: 'hidden', text: 'H', title: 'Hidden file' })
  }

  if (file.is_executable) {
    badges.push({ type: 'executable', text: 'X', title: 'Executable' })
  }

  if (file.is_symlink) {
    badges.push({ type: 'symlink', text: '→', title: 'Symbolic link' })
  }

  return badges
}

// Watch for changes
watch(() => props.files, () => {
  focusedItem.value = ''
})

// Cleanup
onUnmounted(() => {
  if (previewTooltipTimeout.value) {
    clearTimeout(previewTooltipTimeout.value)
  }
})
</script>

<style scoped>
.file-list {
  position: relative;
  height: 100%;
  overflow: hidden;
}

/* Loading overlay */
.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 2px solid var(--surface-border);
  border-top: 2px solid var(--primary-500);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-text {
  color: var(--text-secondary);
  font-size: 14px;
}

/* Empty state */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: var(--text-primary);
}

.empty-description {
  color: var(--text-secondary);
  margin: 0 0 16px 0;
}

.clear-search-btn {
  padding: 8px 16px;
  background: var(--primary-500);
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.clear-search-btn:hover {
  background: var(--primary-600);
}

/* Grid view */
.grid-view {
  height: 100%;
}

.grid-scroller {
  height: 100%;
  padding: 16px;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  background: var(--surface);
  position: relative;
}

.grid-item:hover {
  background: var(--surface-elevated);
  border-color: var(--surface-border);
  transform: translateY(-2px);
}

.grid-item.is-selected {
  background: var(--primary-900);
  border-color: var(--primary-500);
}

.grid-item.is-directory {
  font-weight: 600;
}

.file-icon-container {
  position: relative;
  margin-bottom: 8px;
}

.selection-indicator {
  position: absolute;
  top: -4px;
  right: -4px;
  background: var(--primary-500);
  color: white;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--surface);
}

.file-name {
  font-size: 12px;
  text-align: center;
  word-break: break-word;
  line-height: 1.3;
  margin-bottom: 4px;
  color: var(--text-primary);
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.file-metadata {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 10px;
  color: var(--text-secondary);
  text-align: center;
}

.file-badges {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}

.file-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-hidden {
  background: var(--warning-500);
  color: white;
}

.badge-executable {
  background: var(--success-500);
  color: white;
}

.badge-symlink {
  background: var(--info-500);
  color: white;
}

/* List view */
.list-view {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.list-header {
  background: var(--surface-elevated);
  border-bottom: 1px solid var(--surface-border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.list-header-row {
  display: flex;
  align-items: center;
  height: 32px;
}

.list-header-cell {
  display: flex;
  align-items: center;
  padding: 0 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-right: 1px solid var(--surface-border);
}

.list-header-cell:last-child {
  border-right: none;
}

.list-header-cell.is-sortable {
  cursor: pointer;
  user-select: none;
}

.list-header-cell.is-sortable:hover {
  color: var(--text-primary);
  background: var(--surface);
}

.sort-icon {
  margin-left: 4px;
  opacity: 0.6;
}

.list-scroller {
  flex: 1;
}

.list-item {
  display: flex;
  align-items: center;
  height: 32px;
  padding: 0 8px;
  border-bottom: 1px solid var(--surface-border);
  cursor: pointer;
  transition: background-color 0.15s ease;
  gap: 8px;
}

.list-item:hover {
  background: var(--surface-elevated);
}

.list-item.is-selected {
  background: var(--primary-900);
}

.list-item.is-focused {
  outline: 1px solid var(--primary-500);
  outline-offset: -1px;
}

.file-checkbox {
  display: flex;
  align-items: center;
}

.file-icon-container {
  display: flex;
  align-items: center;
}

.file-name-cell {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-item.is-directory .file-name {
  font-weight: 600;
}

.file-size-cell,
.file-date-cell,
.file-type-cell,
.file-permissions-cell,
.file-owner-cell {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
}

.file-actions-cell {
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.list-item:hover .file-actions-cell,
.list-item.is-selected .file-actions-cell {
  opacity: 1;
}

.file-actions {
  display: flex;
  gap: 4px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: var(--surface-elevated);
  color: var(--text-primary);
}

/* Context menu and tooltip positioning */
.preview-tooltip {
  position: fixed;
  z-index: 1000;
  pointer-events: auto;
}

/* Responsive design */
@media (max-width: 768px) {
  .grid-item {
    padding: 8px;
  }

  .file-name {
    font-size: 11px;
  }

  .file-metadata {
    font-size: 9px;
  }

  .list-item {
    height: 40px;
  }

  .file-name {
    font-size: 12px;
  }

  .file-size-cell,
  .file-date-cell,
  .file-type-cell,
  .file-permissions-cell,
  .file-owner-cell {
    font-size: 11px;
  }
}
</style>