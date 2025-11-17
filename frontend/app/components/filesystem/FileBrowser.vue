<template>
  <div class="file-browser">
    <div class="browser-header">
      <h3 class="glitch">FILE SYSTEM</h3>
      <div class="header-controls">
        <button @click="goUp" class="control-btn" :disabled="!canGoUp">
          <span class="btn-icon">↑</span>
        </button>
        <button @click="refreshDirectory" class="control-btn">
          <span class="btn-icon">↻</span>
        </button>
        <button @click="toggleViewMode" class="control-btn">
          <span class="btn-icon">{{ viewMode === 'grid' ? '☰' : '⊞' }}</span>
        </button>
      </div>
    </div>

    <!-- Breadcrumb navigation -->
    <div class="breadcrumb-nav">
      <div class="breadcrumb-path">
        <span
          v-for="(segment, index) in pathSegments"
          :key="index"
          class="breadcrumb-segment"
          @click="navigateToPath(getPathUpTo(index))"
        >
          <span class="segment-icon" v-if="index === 0">🏠</span>
          <span class="segment-text">{{ segment }}</span>
          <span class="segment-separator" v-if="index < pathSegments.length - 1">/</span>
        </span>
      </div>
    </div>

    <!-- File search -->
    <div class="search-bar">
      <div class="search-input-wrapper">
        <span class="search-icon">🔍</span>
        <input
          v-model="searchQuery"
          @input="filterFiles"
          type="text"
          placeholder="Search files..."
          class="search-input"
        />
        <button
          v-if="searchQuery"
          @click="clearSearch"
          class="clear-search-btn"
        >
          ✕
        </button>
      </div>
      <div class="search-filters">
        <label class="filter-label">
          <input
            v-model="showHiddenFiles"
            type="checkbox"
            @change="filterFiles"
          />
          Show hidden files
        </label>
      </div>
    </div>

    <!-- Directory stats -->
    <div class="directory-stats">
      <div class="stat-item">
        <span class="stat-label">Items:</span>
        <span class="stat-value">{{ filteredFiles.length }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Size:</span>
        <span class="stat-value">{{ formatBytes(directorySize) }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Path:</span>
        <span class="stat-value path-value">{{ currentPath }}</span>
      </div>
    </div>

    <!-- File content area -->
    <div class="file-content" :class="viewMode">
      <!-- Grid view -->
      <div v-if="viewMode === 'grid'" class="grid-view">
        <div
          v-for="(file, index) in filteredFiles"
          :key="file.name"
          class="file-item grid-item"
          :class="{
            focused: isKeyboardFocused && focusedFile?.name === file.name,
            selected: selectedFiles.includes(file.name)
          }"
          :tabindex="isKeyboardFocused && focusedFile?.name === file.name ? 0 : -1"
          @click="handleFileClick(file)"
          @contextmenu="showContextMenu($event, file)"
          @dblclick="handleFileDoubleClick(file)"
          @mouseenter="handleMouseEnter(index)"
          @keydown="handleKeyDown($event, index)"
          ref="fileItems"
        >
          <div class="file-icon">
            <span class="icon-emoji">{{ getFileIcon(file) }}</span>
            <div class="file-selected-overlay" v-if="selectedFiles.includes(file.name)"></div>
          </div>
          <div class="file-name" :title="file.name">
            {{ file.name }}
          </div>
          <div class="file-info">
            <span class="file-size">{{ formatFileSize(file) }}</span>
            <span class="file-date">{{ formatDate(file.modified) }}</span>
          </div>
        </div>
      </div>

      <!-- List view -->
      <div v-else class="list-view">
        <table class="file-table">
          <thead>
            <tr>
              <th class="sortable-header" @click="sortBy('name')">
                <span>Name</span>
                <span class="sort-icon" v-if="sortField === 'name'">
                  {{ sortOrder === 'asc' ? '↑' : '↓' }}
                </span>
              </th>
              <th class="sortable-header" @click="sortBy('size')">
                <span>Size</span>
                <span class="sort-icon" v-if="sortField === 'size'">
                  {{ sortOrder === 'asc' ? '↑' : '↓' }}
                </span>
              </th>
              <th class="sortable-header" @click="sortBy('modified')">
                <span>Modified</span>
                <span class="sort-icon" v-if="sortField === 'modified'">
                  {{ sortOrder === 'asc' ? '↑' : '↓' }}
                </span>
              </th>
              <th class="sortable-header" @click="sortBy('type')">
                <span>Type</span>
                <span class="sort-icon" v-if="sortField === 'type'">
                  {{ sortOrder === 'asc' ? '↑' : '↓' }}
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(file, index) in filteredFiles"
              :key="file.name"
              class="file-row"
              :class="{
                focused: isKeyboardFocused && focusedFile?.name === file.name,
                selected: selectedFiles.includes(file.name)
              }"
              :tabindex="isKeyboardFocused && focusedFile?.name === file.name ? 0 : -1"
              @click="handleFileClick(file)"
              @contextmenu="showContextMenu($event, file)"
              @dblclick="handleFileDoubleClick(file)"
              @mouseenter="handleMouseEnter(index)"
              @keydown="handleKeyDown($event, index)"
              ref="fileItems"
            >
              <td class="file-name-cell">
                <span class="file-emoji">{{ getFileIcon(file) }}</span>
                <span class="file-name-text">{{ file.name }}</span>
              </td>
              <td class="file-size-cell">
                {{ formatFileSize(file) }}
              </td>
              <td class="file-date-cell">
                {{ formatDate(file.modified) }}
              </td>
              <td class="file-type-cell">
                <span class="file-type-badge" :class="getFileTypeClass(file)">
                  {{ getFileType(file) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Context menu -->
    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      @click="hideContextMenu"
    >
      <div class="context-menu-item" @click="openFile">
        <span class="menu-icon">📂</span>
        <span>Open</span>
      </div>
      <div class="context-menu-item" @click="renameFile">
        <span class="menu-icon">✏️</span>
        <span>Rename</span>
      </div>
      <div class="context-menu-item" @click="copyFile">
        <span class="menu-icon">📋</span>
        <span>Copy</span>
      </div>
      <div class="context-menu-item" @click="moveFile">
        <span class="menu-icon">✂️</span>
        <span>Cut</span>
      </div>
      <div class="context-menu-separator"></div>
      <div class="context-menu-item danger" @click="deleteFile">
        <span class="menu-icon">🗑️</span>
        <span>Delete</span>
      </div>
    </div>

    <!-- File preview modal -->
    <div v-if="previewFile" class="preview-modal" @click="closePreview">
      <div class="preview-content" @click.stop>
        <div class="preview-header">
          <span class="preview-title">{{ previewFile.name }}</span>
          <button @click="closePreview" class="close-btn">✕</button>
        </div>
        <div class="preview-body">
          <pre v-if="isTextFile(previewFile)" class="file-preview-text">{{ previewContent }}</pre>
          <div v-else-if="isImageFile(previewFile)" class="file-preview-image">
            <img :src="previewContent" :alt="previewFile.name" />
          </div>
          <div v-else class="file-preview-binary">
            <span class="binary-icon">📄</span>
            <p>Binary file - cannot preview</p>
            <p>Size: {{ formatFileSize(previewFile) }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'

interface FileItem {
  name: string
  path: string
  type: 'file' | 'directory'
  size: number
  modified: Date
  permissions: string
  owner: string
  extension?: string
}

interface ContextMenu {
  visible: boolean
  x: number
  y: number
  file?: FileItem
}

interface KeyboardNavigationState {
  focusedIndex: number
  isNavigating: boolean
  lastInteractionTime: number
}

// Reactive data
const currentPath = ref<string>('/')
const files = ref<FileItem[]>([])
const selectedFiles = ref<string[]>([])
const searchQuery = ref<string>('')
const showHiddenFiles = ref<boolean>(false)
const viewMode = ref<'grid' | 'list'>('grid')
const sortField = ref<'name' | 'size' | 'modified' | 'type'>('name')
const sortOrder = ref<'asc' | 'desc'>('asc')
const contextMenu = ref<ContextMenu>({ visible: false, x: 0, y: 0 })
const previewFile = ref<FileItem | null>(null)
const previewContent = ref<string>('')

// Keyboard navigation state
const keyboardNav = ref<KeyboardNavigationState>({
  focusedIndex: -1,
  isNavigating: false,
  lastInteractionTime: 0
})

// Computed properties
const pathSegments = computed(() => {
  return currentPath.value.split('/').filter(segment => segment.length > 0)
})

const canGoUp = computed(() => {
  return currentPath.value !== '/'
})

const filteredFiles = computed(() => {
  let result = [...files.value]

  // Filter by search query
  if (searchQuery.value) {
    result = result.filter(file =>
      file.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }

  // Filter hidden files
  if (!showHiddenFiles.value) {
    result = result.filter(file => !file.name.startsWith('.'))
  }

  // Sort files
  result.sort((a, b) => {
    let aValue: any = a[sortField.value]
    let bValue: any = b[sortField.value]

    if (sortField.value === 'name') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }

    if (sortField.value === 'type') {
      aValue = a.type === 'directory' ? 'directory' : getFileExtension(a)
      bValue = b.type === 'directory' ? 'directory' : getFileExtension(b)
    }

    if (aValue < bValue) return sortOrder.value === 'asc' ? -1 : 1
    if (aValue > bValue) return sortOrder.value === 'asc' ? 1 : -1
    return 0
  })

  // Directories first
  result.sort((a, b) => {
    if (a.type === 'directory' && b.type !== 'directory') return -1
    if (a.type !== 'directory' && b.type === 'directory') return 1
    return 0
  })

  return result
})

const focusedFile = computed(() => {
  if (keyboardNav.value.focusedIndex >= 0 && keyboardNav.value.focusedIndex < filteredFiles.value.length) {
    return filteredFiles.value[keyboardNav.value.focusedIndex]
  }
  return null
})

const canNavigateUp = computed(() => {
  return keyboardNav.value.focusedIndex > 0
})

const canNavigateDown = computed(() => {
  return keyboardNav.value.focusedIndex < filteredFiles.value.length - 1
})

const isKeyboardFocused = computed(() => {
  return keyboardNav.value.isNavigating
})

const directorySize = computed(() => {
  return filteredFiles.value
    .filter(file => file.type === 'file')
    .reduce((total, file) => total + file.size, 0)
})

// Methods
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatFileSize = (file: FileItem): string => {
  if (file.type === 'directory') return '—'
  return formatBytes(file.size)
}

const formatDate = (date: Date): string => {
  return new Intl.DateTimeFormat('en-US', {
    year: '2-digit',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

const getFileIcon = (file: FileItem): string => {
  if (file.type === 'directory') return '📁'

  const ext = file.extension?.toLowerCase()
  const iconMap: Record<string, string> = {
    // Documents
    'txt': '📄',
    'md': '📝',
    'pdf': '📕',
    'doc': '📘',
    'docx': '📘',
    'odt': '📗',

    // Code
    'js': '🟨',
    'ts': '🔷',
    'html': '🌐',
    'css': '🎨',
    'json': '📋',
    'xml': '📄',
    'yml': '📄',
    'yaml': '📄',
    'py': '🐍',
    'go': '🐹',
    'rs': '🦀',
    'java': '☕',
    'cpp': '🔧',
    'c': '🔧',
    'h': '🔧',
    'cs': '🔷',

    // Images
    'png': '🖼️',
    'jpg': '🖼️',
    'jpeg': '🖼️',
    'gif': '🖼️',
    'svg': '🎨',
    'webp': '🖼️',
    'ico': '🖼️',

    // Audio
    'mp3': '🎵',
    'wav': '🎵',
    'flac': '🎵',
    'ogg': '🎵',

    // Video
    'mp4': '🎬',
    'avi': '🎬',
    'mkv': '🎬',
    'webm': '🎬',
    'mov': '🎬',

    // Archives
    'zip': '📦',
    'rar': '📦',
    'tar': '📦',
    'gz': '📦',
    '7z': '📦',

    // Default
    'default': '📄'
  }

  return iconMap[ext || ''] || iconMap.default
}

const getFileType = (file: FileItem): string => {
  if (file.type === 'directory') return 'Directory'
  return file.extension?.toUpperCase() || 'File'
}

const getFileTypeClass = (file: FileItem): string => {
  if (file.type === 'directory') return 'directory'
  return `type-${file.extension?.toLowerCase() || 'unknown'}`
}

const getFileExtension = (file: FileItem): string => {
  return file.extension || 'unknown'
}

const navigateToPath = (path: string) => {
  currentPath.value = path
  loadDirectory(path)
}

const getPathUpTo = (index: number): string => {
  return '/' + pathSegments.value.slice(0, index + 1).join('/')
}

const goUp = () => {
  if (!canGoUp.value) return

  const segments = pathSegments.value
  segments.pop()
  currentPath.value = segments.length > 0 ? '/' + segments.join('/') : '/'
  loadDirectory(currentPath.value)
}

const loadDirectory = async (path: string) => {
  try {
    // Simulate loading directory (replace with actual API call)
    const mockFiles: FileItem[] = [
      {
        name: 'Documents',
        path: path + '/Documents',
        type: 'directory',
        size: 4096,
        modified: new Date('2024-01-15T10:30:00'),
        permissions: 'drwxr-xr-x',
        owner: 'user'
      },
      {
        name: 'Downloads',
        path: path + '/Downloads',
        type: 'directory',
        size: 4096,
        modified: new Date('2024-01-16T14:20:00'),
        permissions: 'drwxr-xr-x',
        owner: 'user'
      },
      {
        name: 'Projects',
        path: path + '/Projects',
        type: 'directory',
        size: 4096,
        modified: new Date('2024-01-17T09:15:00'),
        permissions: 'drwxr-xr-x',
        owner: 'user'
      },
      {
        name: 'config.json',
        path: path + '/config.json',
        type: 'file',
        size: 2048,
        modified: new Date('2024-01-14T16:45:00'),
        permissions: '-rw-r--r--',
        owner: 'user',
        extension: 'json'
      },
      {
        name: 'readme.md',
        path: path + '/readme.md',
        type: 'file',
        size: 1024,
        modified: new Date('2024-01-13T11:30:00'),
        permissions: '-rw-r--r--',
        owner: 'user',
        extension: 'md'
      },
      {
        name: 'script.js',
        path: path + '/script.js',
        type: 'file',
        size: 5120,
        modified: new Date('2024-01-12T13:20:00'),
        permissions: '-rwxr-xr-x',
        owner: 'user',
        extension: 'js'
      }
    ]

    files.value = mockFiles
    selectedFiles.value = []
  } catch (error) {
    console.error('Failed to load directory:', error)
  }
}

const refreshDirectory = () => {
  loadDirectory(currentPath.value)
}

const toggleViewMode = () => {
  viewMode.value = viewMode.value === 'grid' ? 'list' : 'grid'
}

const filterFiles = () => {
  // Filtering is handled by computed property
}

const clearSearch = () => {
  searchQuery.value = ''
}

const sortBy = (field: 'name' | 'size' | 'modified' | 'type') => {
  if (sortField.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortOrder.value = 'asc'
  }
}

const handleFileClick = (file: FileItem) => {
  const index = selectedFiles.value.indexOf(file.name)
  if (index > -1) {
    selectedFiles.value.splice(index, 1)
  } else {
    selectedFiles.value.push(file.name)
  }
}

const handleFileDoubleClick = (file: FileItem) => {
  if (file.type === 'directory') {
    navigateToPath(file.path)
  } else {
    previewFile.value = file
    loadFilePreview(file)
  }
}

const showContextMenu = (event: MouseEvent, file: FileItem) => {
  event.preventDefault()
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    file
  }
}

const hideContextMenu = () => {
  contextMenu.value.visible = false
}

const isTextFile = (file: FileItem): boolean => {
  const textExtensions = ['txt', 'md', 'json', 'js', 'ts', 'html', 'css', 'xml', 'yml', 'yaml', 'py', 'go', 'rs', 'java', 'cpp', 'c', 'h', 'cs']
  return textExtensions.includes(file.extension?.toLowerCase() || '')
}

const isImageFile = (file: FileItem): boolean => {
  const imageExtensions = ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'ico']
  return imageExtensions.includes(file.extension?.toLowerCase() || '')
}

const loadFilePreview = async (file: FileItem) => {
  if (isImageFile(file)) {
    previewContent.value = `data:image/${file.extension};base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==`
  } else if (isTextFile(file)) {
    previewContent.value = `// Sample content for ${file.name}\n\nThis is a preview of the file content.\nIn a real implementation, this would load the actual file content.\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit.\nSed do eiusmod tempor incididunt ut labore et dolore magna aliqua.`
  } else {
    previewContent.value = ''
  }
}

const closePreview = () => {
  previewFile.value = null
  previewContent.value = ''
}

// Context menu actions
const openFile = () => {
  if (contextMenu.value.file) {
    handleFileDoubleClick(contextMenu.value.file)
  }
  hideContextMenu()
}

const renameFile = () => {
  // Implement rename functionality
  hideContextMenu()
}

const copyFile = () => {
  // Implement copy functionality
  hideContextMenu()
}

const moveFile = () => {
  // Implement move functionality
  hideContextMenu()
}

const deleteFile = () => {
  // Implement delete functionality
  hideContextMenu()
}

// Keyboard navigation methods
const startKeyboardNavigation = () => {
  if (filteredFiles.value.length > 0) {
    keyboardNav.value.isNavigating = true
    keyboardNav.value.focusedIndex = keyboardNav.value.focusedIndex < 0 ? 0 : keyboardNav.value.focusedIndex
    keyboardNav.value.lastInteractionTime = Date.now()
  }
}

const stopKeyboardNavigation = () => {
  keyboardNav.value.isNavigating = false
  keyboardNav.value.focusedIndex = -1
}

const navigateUp = () => {
  if (!isKeyboardFocused.value || !canNavigateUp.value) return

  keyboardNav.value.focusedIndex--
  keyboardNav.value.lastInteractionTime = Date.now()
  ensureVisibleElement()
}

const navigateDown = () => {
  if (!isKeyboardFocused.value || !canNavigateDown.value) return

  keyboardNav.value.focusedIndex++
  keyboardNav.value.lastInteractionTime = Date.now()
  ensureVisibleElement()
}

const navigatePageUp = () => {
  if (!isKeyboardFocused.value) return

  const itemsPerPage = viewMode.value === 'grid' ? 12 : 20
  keyboardNav.value.focusedIndex = Math.max(0, keyboardNav.value.focusedIndex - itemsPerPage)
  keyboardNav.value.lastInteractionTime = Date.now()
  ensureVisibleElement()
}

const navigatePageDown = () => {
  if (!isKeyboardFocused.value) return

  const itemsPerPage = viewMode.value === 'grid' ? 12 : 20
  const maxIndex = filteredFiles.value.length - 1
  keyboardNav.value.focusedIndex = Math.min(maxIndex, keyboardNav.value.focusedIndex + itemsPerPage)
  keyboardNav.value.lastInteractionTime = Date.now()
  ensureVisibleElement()
}

const navigateFirst = () => {
  if (!isKeyboardFocused.value || filteredFiles.value.length === 0) return

  keyboardNav.value.focusedIndex = 0
  keyboardNav.value.lastInteractionTime = Date.now()
  ensureVisibleElement()
}

const navigateLast = () => {
  if (!isKeyboardFocused.value || filteredFiles.value.length === 0) return

  keyboardNav.value.focusedIndex = filteredFiles.value.length - 1
  keyboardNav.value.lastInteractionTime = Date.now()
  ensureVisibleElement()
}

const selectFocusedFile = () => {
  if (!focusedFile.value) return

  const fileName = focusedFile.value.name
  const index = selectedFiles.value.indexOf(fileName)

  if (index > -1) {
    // Deselect if already selected
    selectedFiles.value.splice(index, 1)
  } else {
    // Select if not selected
    selectedFiles.value.push(fileName)
  }
}

const activateFocusedFile = () => {
  if (!focusedFile.value) return

  handleFileDoubleClick(focusedFile.value)
}

const toggleFocusedFileSelection = () => {
  if (!focusedFile.value) return

  const fileName = focusedFile.value.name

  // If no files are selected, select this one
  if (selectedFiles.value.length === 0) {
    selectedFiles.value = [fileName]
  }
  // If only this file is selected, deselect it
  else if (selectedFiles.value.length === 1 && selectedFiles.value[0] === fileName) {
    selectedFiles.value = []
  }
  // Otherwise, select only this file
  else {
    selectedFiles.value = [fileName]
  }
}

const selectAllFiles = () => {
  selectedFiles.value = filteredFiles.value.map(file => file.name)
}

const clearSelection = () => {
  selectedFiles.value = []
}

const ensureVisibleElement = () => {
  nextTick(() => {
    const focusedElement = document.querySelector('.file-item.focused, .file-row.focused')
    if (focusedElement) {
      focusedElement.scrollIntoView({
        behavior: 'smooth',
        block: 'nearest',
        inline: 'nearest'
      })
    }
  })
}

const handleKeyNavigation = (event: KeyboardEvent) => {
  // Ignore key navigation when typing in search input
  if (event.target === document.querySelector('.search-input')) {
    return
  }

  const key = event.key.toLowerCase()
  const ctrlOrMeta = event.ctrlKey || event.metaKey
  const shift = event.shiftKey

  // Prevent default for navigation keys
  if (['arrowup', 'arrowdown', 'arrowleft', 'arrowright', 'home', 'end', 'pageup', 'pagedown', 'enter', 'space', 'escape'].includes(key)) {
    event.preventDefault()
  }

  switch (key) {
    // Navigation
    case 'arrowup':
      if (shift) {
        navigatePageUp()
      } else {
        navigateUp()
      }
      break
    case 'arrowdown':
      if (shift) {
        navigatePageDown()
      } else {
        navigateDown()
      }
      break
    case 'arrowleft':
    case 'arrowright':
      // These can be used for breadcrumb navigation
      break
    case 'home':
      navigateFirst()
      break
    case 'end':
      navigateLast()
      break
    case 'pageup':
      navigatePageUp()
      break
    case 'pagedown':
      navigatePageDown()
      break

    // Actions
    case 'enter':
      if (ctrlOrMeta) {
        // Open in new tab/window
        activateFocusedFile()
      } else {
        // Activate file
        activateFocusedFile()
      }
      break
    case ' ':
      if (!isKeyboardFocused.value) {
        startKeyboardNavigation()
      } else {
        selectFocusedFile()
      }
      break
    case 'escape':
      stopKeyboardNavigation()
      clearSelection()
      break

    // Shortcuts
    case 'f2':
      if (focusedFile.value) {
        // Rename focused file
        renameFocusedFile()
      }
      break
    case 'delete':
    case 'backspace':
      if (selectedFiles.value.length > 0 && focusedFile.value) {
        // Delete selected files
        deleteSelectedFiles()
      }
      break
    case 'a':
      if (ctrlOrMeta) {
        event.preventDefault()
        selectAllFiles()
      }
      break
    case 'c':
      if (ctrlOrMeta) {
        event.preventDefault()
        // Copy selected files (implement copy functionality)
      }
      break
    case 'x':
      if (ctrlOrMeta) {
        event.preventDefault()
        // Cut selected files (implement cut functionality)
      }
      break
    case 'v':
      if (ctrlOrMeta) {
        event.preventDefault()
        // Paste files (implement paste functionality)
      }
      break
    case 'f':
      if (ctrlOrMeta) {
        event.preventDefault()
        // Focus search input
        const searchInput = document.querySelector('.search-input') as HTMLInputElement
        if (searchInput) {
          searchInput.focus()
        }
      }
      break
    case 'r':
      if (!ctrlOrMeta && !searchQuery.value) {
        // Refresh directory
        refreshDirectory()
      }
      break
    case 'g':
      if (!ctrlOrMeta && !searchQuery.value) {
        // Go up directory
        goUp()
      }
      break
    case 'h':
      if (!ctrlOrMeta && !searchQuery.value) {
        // Toggle hidden files
        showHiddenFiles.value = !showHiddenFiles.value
      }
      break
    case 't':
      if (!ctrlOrMeta && !searchQuery.value) {
        // Toggle view mode
        toggleViewMode()
      }
      break
  }
}

const renameFocusedFile = () => {
  if (focusedFile.value) {
    // Focus on the file name for renaming
    // This would typically open a rename dialog
    console.log('Rename file:', focusedFile.value.name)
  }
}

const deleteSelectedFiles = () => {
  if (selectedFiles.value.length > 0) {
    // Implement delete functionality for selected files
    console.log('Delete files:', selectedFiles.value)
  }
}

const handleKeyDown = (event: KeyboardEvent, index: number) => {
  // Start keyboard navigation if not already active
  if (!keyboardNav.value.isNavigating) {
    startKeyboardNavigation()
  }

  // Update focused index
  keyboardNav.value.focusedIndex = index
  keyboardNav.value.lastInteractionTime = Date.now()

  // Handle the key event
  handleKeyNavigation(event)
}

const handleMouseEnter = (index: number) => {
  // Stop keyboard navigation when mouse is used
  if (keyboardNav.value.isNavigating) {
    stopKeyboardNavigation()
  }
}

// Watch for filtered files changes to reset keyboard navigation
watch(filteredFiles, () => {
  if (keyboardNav.value.focusedIndex >= filteredFiles.value.length) {
    keyboardNav.value.focusedIndex = Math.max(0, filteredFiles.value.length - 1)
  }
}, { flush: 'post' })

// Watch for search query changes
watch(searchQuery, () => {
  if (searchQuery.value) {
    stopKeyboardNavigation()
  }
})

// Lifecycle
onMounted(() => {
  loadDirectory(currentPath.value)

  // Hide context menu when clicking outside
  document.addEventListener('click', hideContextMenu)

  // Add keyboard event listeners
  document.addEventListener('keydown', handleKeyNavigation)

  // Auto-stop keyboard navigation after inactivity
  const checkInactivity = setInterval(() => {
    if (keyboardNav.value.isNavigating &&
        Date.now() - keyboardNav.value.lastInteractionTime > 30000) { // 30 seconds
      stopKeyboardNavigation()
    }
  }, 5000)

  // Store interval ID for cleanup
  window._keyboardNavInterval = checkInactivity
})

onUnmounted(() => {
  document.removeEventListener('click', hideContextMenu)
  document.removeEventListener('keydown', handleKeyNavigation)

  // Clean up interval
  if (window._keyboardNavInterval) {
    clearInterval(window._keyboardNavInterval)
  }
})
</script>

<style scoped>
.file-browser {
  background: rgba(10, 10, 10, 0.95);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  padding: 20px;
  backdrop-filter: blur(10px);
  font-family: 'Fira Code', monospace;
  color: var(--text-primary);
  height: 100%;
  display: flex;
  flex-direction: column;
}

.browser-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--surface-border);
}

.header-controls {
  display: flex;
  gap: 8px;
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.control-btn:hover:not(:disabled) {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.control-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 14px;
}

.breadcrumb-nav {
  margin-bottom: 15px;
}

.breadcrumb-path {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px 12px;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
}

.breadcrumb-segment {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  padding: 2px 4px;
  border-radius: 4px;
}

.breadcrumb-segment:hover {
  background: var(--surface-elevated);
  color: var(--primary-400);
}

.segment-icon {
  font-size: 14px;
}

.segment-text {
  font-size: 12px;
}

.segment-separator {
  color: var(--text-secondary);
  font-size: 12px;
}

.search-bar {
  margin-bottom: 15px;
}

.search-input-wrapper {
  position: relative;
  margin-bottom: 10px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-secondary);
  font-size: 14px;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 12px;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: var(--primary-500);
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.2);
}

.clear-search-btn {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 12px;
  padding: 2px;
  border-radius: 2px;
  transition: all 0.2s ease;
}

.clear-search-btn:hover {
  color: var(--error);
  background: rgba(239, 68, 68, 0.1);
}

.search-filters {
  display: flex;
  gap: 15px;
}

.filter-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
}

.filter-label input[type="checkbox"] {
  accent-color: var(--primary-500);
}

.directory-stats {
  display: flex;
  gap: 20px;
  margin-bottom: 15px;
  padding: 10px 15px;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.stat-value {
  color: var(--primary-400);
  font-size: 11px;
  font-weight: bold;
}

.path-value {
  color: var(--text-secondary);
  font-family: monospace;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-content {
  flex: 1;
  overflow: auto;
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  background: var(--surface);
}

/* Grid View */
.grid-view {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 12px;
  padding: 15px;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px;
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.grid-item:hover {
  background: var(--surface-elevated);
  border-color: var(--primary-500);
}

.grid-item.selected {
  background: rgba(14, 165, 233, 0.1);
  border-color: var(--primary-500);
}

.grid-item.focused {
  background: rgba(14, 165, 233, 0.15);
  border-color: var(--primary-400);
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.3);
  outline: 2px solid var(--primary-400);
  outline-offset: 2px;
}

.grid-item.focused.selected {
  background: rgba(14, 165, 233, 0.2);
  border-color: var(--primary-400);
  box-shadow: 0 0 15px rgba(14, 165, 233, 0.4);
}

.file-icon {
  position: relative;
  margin-bottom: 8px;
  font-size: 32px;
}

.icon-emoji {
  font-size: inherit;
}

.file-selected-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(14, 165, 233, 0.2);
  border: 2px solid var(--primary-500);
  border-radius: 4px;
}

.file-name {
  font-size: 11px;
  text-align: center;
  word-break: break-word;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.file-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  font-size: 9px;
  color: var(--text-secondary);
}

/* List View */
.list-view {
  height: 100%;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
}

.sortable-header {
  cursor: pointer;
  user-select: none;
  transition: all 0.2s ease;
}

.sortable-header:hover {
  background: var(--surface-elevated);
  color: var(--primary-400);
}

.sort-icon {
  font-size: 10px;
  color: var(--accent-400);
}

.file-row {
  border-bottom: 1px solid var(--surface-elevated);
  cursor: pointer;
  transition: all 0.2s ease;
}

.file-row:hover {
  background: var(--surface-elevated);
}

.file-row.selected {
  background: rgba(14, 165, 233, 0.1);
}

.file-row.focused {
  background: rgba(14, 165, 233, 0.15);
  border-color: var(--primary-400);
  box-shadow: 0 0 8px rgba(14, 165, 233, 0.3);
  outline: 2px solid var(--primary-400);
  outline-offset: -2px;
}

.file-row.focused.selected {
  background: rgba(14, 165, 233, 0.2);
  border-color: var(--primary-400);
  box-shadow: 0 0 12px rgba(14, 165, 233, 0.4);
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
}

.file-emoji {
  font-size: 16px;
}

.file-name-text {
  font-family: inherit;
}

.file-size-cell,
.file-date-cell,
.file-type-cell {
  padding: 8px 12px;
  color: var(--text-secondary);
}

.file-type-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: bold;
  text-transform: uppercase;
}

.file-type-badge.directory {
  background: rgba(14, 165, 233, 0.2);
  color: #0ea5e9;
}

.file-type-badge.type-js {
  background: rgba(250, 204, 21, 0.2);
  color: #facc15;
}

.file-type-badge.type-py {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

/* Context Menu */
.context-menu {
  position: fixed;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 4px 0;
  min-width: 150px;
  z-index: 1000;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 11px;
}

.context-menu-item:hover {
  background: var(--surface-elevated);
}

.context-menu-item.danger {
  color: var(--error);
}

.context-menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.1);
}

.menu-icon {
  font-size: 14px;
}

.context-menu-separator {
  height: 1px;
  background: var(--surface-border);
  margin: 4px 0;
}

/* Preview Modal */
.preview-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(5px);
}

.preview-content {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  max-width: 80vw;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid var(--surface-border);
}

.preview-title {
  color: var(--primary-400);
  font-weight: bold;
  font-size: 14px;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 18px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  color: var(--error);
  background: rgba(239, 68, 68, 0.1);
}

.preview-body {
  flex: 1;
  padding: 20px;
  overflow: auto;
}

.file-preview-text {
  background: var(--background);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  padding: 15px;
  font-family: monospace;
  font-size: 12px;
  line-height: 1.4;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 60vh;
  overflow: auto;
}

.file-preview-image {
  display: flex;
  justify-content: center;
  align-items: center;
}

.file-preview-image img {
  max-width: 100%;
  max-height: 60vh;
  border-radius: 6px;
}

.file-preview-binary {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 15px;
  color: var(--text-secondary);
  text-align: center;
}

.binary-icon {
  font-size: 48px;
  opacity: 0.5;
}

/* Responsive design */
@media (max-width: 768px) {
  .grid-view {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 8px;
    padding: 10px;
  }

  .directory-stats {
    flex-direction: column;
    gap: 10px;
  }

  .stat-item {
    justify-content: space-between;
  }

  .file-table {
    font-size: 10px;
  }

  .file-name-cell,
  .file-size-cell,
  .file-date-cell,
  .file-type-cell {
    padding: 6px 8px;
  }
}
</style>