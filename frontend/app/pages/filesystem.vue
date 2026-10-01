<template>
  <div class="filesystem-page">
    <!-- Filesystem Header -->
    <header class="filesystem-header">
      <div class="header-left">
        <h1 class="page-title">File Manager</h1>
        <div class="breadcrumb">
          <button 
            v-for="(part, index) in pathParts" 
            :key="index"
            class="breadcrumb-item"
            @click="navigateToPath(getPathUpTo(index))"
          >
            {{ part || 'Root' }}
          </button>
        </div>
      </div>
      <div class="header-right">
        <div class="view-controls">
          <button 
            class="view-btn"
            :class="{ active: viewMode === 'list' }"
            @click="setViewMode('list')"
            title="List View"
          >
            📋
          </button>
          <button 
            class="view-btn"
            :class="{ active: viewMode === 'grid' }"
            @click="setViewMode('grid')"
            title="Grid View"
          >
            📱
          </button>
          <button 
            class="view-btn"
            :class="{ active: viewMode === 'tree' }"
            @click="setViewMode('tree')"
            title="Tree View"
          >
            🌳
          </button>
        </div>
        <div class="navigation-controls">
          <button 
            class="nav-btn"
            :disabled="!canGoBack"
            @click="navigateBack"
            title="Back"
          >
            ⬅️
          </button>
          <button 
            class="nav-btn"
            :disabled="!canGoForward"
            @click="navigateForward"
            title="Forward"
          >
            ➡️
          </button>
          <button 
            class="nav-btn"
            @click="navigateUp"
            title="Up"
          >
            ⬆️
          </button>
          <button 
            class="nav-btn"
            @click="navigateHome"
            title="Home"
          >
            🏠
          </button>
        </div>
      </div>
    </header>

    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-left">
        <button class="toolbar-btn" @click="createDirectory">
          📁 New Folder
        </button>
        <button class="toolbar-btn" @click="createFile">
          📄 New File
        </button>
        <div class="toolbar-separator"></div>
        <button 
          class="toolbar-btn"
          :disabled="!hasSelection"
          @click="copyItems"
        >
          📋 Copy
        </button>
        <button 
          class="toolbar-btn"
          :disabled="!hasSelection"
          @click="cutItems"
        >
          ✂️ Cut
        </button>
        <button 
          class="toolbar-btn"
          :disabled="!clipboard"
          @click="pasteItems"
        >
          📌 Paste
        </button>
        <div class="toolbar-separator"></div>
        <button 
          class="toolbar-btn danger"
          :disabled="!hasSelection"
          @click="deleteSelectedItems"
        >
          🗑️ Delete
        </button>
      </div>
      <div class="toolbar-right">
        <div class="search-box">
          <input 
            v-model="searchQuery"
            type="text"
            placeholder="Search files..."
            class="search-input"
          />
          <button class="search-btn">🔍</button>
        </div>
        <div class="sort-controls">
          <select v-model="sortBy" @change="updateSorting">
            <option value="name">Name</option>
            <option value="size">Size</option>
            <option value="modified">Modified</option>
            <option value="type">Type</option>
          </select>
          <button 
            class="sort-order-btn"
            @click="toggleSortOrder"
          >
            {{ sortOrder === 'asc' ? '↑' : '↓' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <main class="filesystem-content">
      <!-- Sidebar with Tree View -->
      <aside class="filesystem-sidebar">
        <div class="sidebar-header">
          <h3>Folders</h3>
          <button class="refresh-btn" @click="refreshCurrentDirectory">
            🔄
          </button>
        </div>
        <div class="tree-view">
          <div 
            v-for="item in treeItems" 
            :key="item.path"
            class="tree-item"
            :class="{ expanded: item.expanded, active: item.path === currentPath }"
            @click="toggleTreeItem(item)"
          >
            <span class="tree-icon">{{ item.isExpanded ? '📂' : '📁' }}</span>
            <span class="tree-label">{{ item.name }}</span>
          </div>
        </div>
        
        <!-- Quick Access -->
        <div class="quick-access">
          <h4>Quick Access</h4>
          <div 
            v-for="location in quickLocations" 
            :key="location.path"
            class="quick-item"
            @click="navigateToPath(location.path)"
          >
            <span class="quick-icon">{{ location.icon }}</span>
            <span class="quick-label">{{ location.name }}</span>
          </div>
        </div>
      </aside>

      <!-- File List -->
      <div class="file-list-container">
        <div class="file-list-header">
          <div class="list-info">
            <span>{{ allItems.length }} items</span>
            <span v-if="selectedItems.length > 0">
              {{ selectedItems.length }} selected
            </span>
          </div>
          <div class="list-actions">
            <button class="action-btn" @click="selectAll">
              Select All
            </button>
            <button class="action-btn" @click="clearSelection">
              Clear Selection
            </button>
          </div>
        </div>

        <!-- List View -->
        <div v-if="viewMode === 'list'" class="file-list">
          <div class="list-header">
            <div class="header-cell name-cell">
              <input 
                type="checkbox"
                :checked="allItems.length > 0 && selectedItems.length === allItems.length"
                @change="toggleSelectAll"
              />
              Name
            </div>
            <div class="header-cell size-cell">Size</div>
            <div class="header-cell modified-cell">Modified</div>
            <div class="header-cell type-cell">Type</div>
            <div class="header-cell actions-cell">Actions</div>
          </div>
          <div class="list-body">
            <div 
              v-for="item in paginatedItems" 
              :key="item.path"
              class="list-row"
              :class="{ selected: selectedItems.includes(item.path) }"
              @click="toggleSelection(item.path)"
              @dblclick="openItem(item)"
            >
              <div class="list-cell name-cell">
                <input 
                  type="checkbox"
                  :checked="selectedItems.includes(item.path)"
                  @change.stop="toggleSelection(item.path)"
                />
                <span class="item-icon">{{ getItemIcon(item) }}</span>
                <span class="item-name">{{ item.name }}</span>
              </div>
              <div class="list-cell size-cell">
                {{ item.isDirectory ? '-' : formatFileSize(item.size) }}
              </div>
              <div class="list-cell modified-cell">
                {{ formatDate(item.modified) }}
              </div>
              <div class="list-cell type-cell">
                {{ item.isDirectory ? 'Directory' : item.type || 'Unknown' }}
              </div>
              <div class="list-cell actions-cell">
                <button 
                  class="action-icon-btn"
                  @click.stop="renameItem(item)"
                  title="Rename"
                >
                  ✏️
                </button>
                <button 
                  class="action-icon-btn"
                  @click.stop="deleteItem(item)"
                  title="Delete"
                >
                  🗑️
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Grid View -->
        <div v-else-if="viewMode === 'grid'" class="file-grid">
          <div 
            v-for="item in paginatedItems" 
            :key="item.path"
            class="grid-item"
            :class="{ selected: selectedItems.includes(item.path) }"
            @click="toggleSelection(item.path)"
            @dblclick="openItem(item)"
          >
            <div class="grid-item-icon">
              <span class="item-icon-large">{{ getItemIcon(item) }}</span>
              <input 
                type="checkbox"
                class="grid-checkbox"
                :checked="selectedItems.includes(item.path)"
                @change.stop="toggleSelection(item.path)"
              />
            </div>
            <div class="grid-item-name">{{ item.name }}</div>
            <div class="grid-item-info">
              {{ item.isDirectory ? 'Folder' : formatFileSize(item.size) }}
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div class="pagination">
          <button 
            class="page-btn"
            :disabled="currentPage === 1"
            @click="changePage(currentPage - 1)"
          >
            Previous
          </button>
          <span class="page-info">
            Page {{ currentPage }} of {{ totalPages }}
          </span>
          <button 
            class="page-btn"
            :disabled="currentPage === totalPages"
            @click="changePage(currentPage + 1)"
          >
            Next
          </button>
        </div>
      </div>
    </main>

    <!-- Modals -->
    <Teleport to="body">
      <!-- Create Directory Modal -->
      <div v-if="showCreateDirModal" class="modal-overlay" @click="closeCreateDirModal">
        <div class="modal-panel" @click.stop>
          <div class="modal-header">
            <h3>Create New Folder</h3>
            <button class="close-btn" @click="closeCreateDirModal">×</button>
          </div>
          <div class="modal-content">
            <div class="form-group">
              <label>Folder Name</label>
              <input 
                v-model="newDirName"
                type="text"
                placeholder="Enter folder name"
                class="form-input"
                @keydown.enter="confirmCreateDirectory"
              />
            </div>
            <div class="modal-actions">
              <button class="btn-secondary" @click="closeCreateDirModal">
                Cancel
              </button>
              <button 
                class="btn-primary"
                :disabled="!newDirName.trim()"
                @click="confirmCreateDirectory"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Create File Modal -->
      <div v-if="showCreateFileModal" class="modal-overlay" @click="closeCreateFileModal">
        <div class="modal-panel" @click.stop>
          <div class="modal-header">
            <h3>Create New File</h3>
            <button class="close-btn" @click="closeCreateFileModal">×</button>
          </div>
          <div class="modal-content">
            <div class="form-group">
              <label>File Name</label>
              <input 
                v-model="newFileName"
                type="text"
                placeholder="Enter file name"
                class="form-input"
                @keydown.enter="confirmCreateFile"
              />
            </div>
            <div class="modal-actions">
              <button class="btn-secondary" @click="closeCreateFileModal">
                Cancel
              </button>
              <button 
                class="btn-primary"
                :disabled="!newFileName.trim()"
                @click="confirmCreateFile"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Rename Modal -->
      <div v-if="showRenameModal" class="modal-overlay" @click="closeRenameModal">
        <div class="modal-panel" @click.stop>
          <div class="modal-header">
            <h3>Rename Item</h3>
            <button class="close-btn" @click="closeRenameModal">×</button>
          </div>
          <div class="modal-content">
            <div class="form-group">
              <label>New Name</label>
              <input 
                v-model="renameValue"
                type="text"
                placeholder="Enter new name"
                class="form-input"
                @keydown.enter="confirmRename"
              />
            </div>
            <div class="modal-actions">
              <button class="btn-secondary" @click="closeRenameModal">
                Cancel
              </button>
              <button 
                class="btn-primary"
                :disabled="!renameValue.trim()"
                @click="confirmRename"
              >
                Rename
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Delete Confirmation Modal -->
      <div v-if="showDeleteModal" class="modal-overlay" @click="closeDeleteModal">
        <div class="modal-panel" @click.stop>
          <div class="modal-header">
            <h3>Confirm Delete</h3>
            <button class="close-btn" @click="closeDeleteModal">×</button>
          </div>
          <div class="modal-content">
            <p>Are you sure you want to delete the following items?</p>
            <ul class="delete-list">
              <li v-for="item in itemsToDelete" :key="item.path">
                {{ item.name }}
              </li>
            </ul>
            <div class="modal-actions">
              <button class="btn-secondary" @click="closeDeleteModal">
                Cancel
              </button>
              <button class="btn-danger" @click="confirmDelete">
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useFilesystem } from '~/composables/useFilesystem'
import type { FileItem, DirectoryItem } from '~/types/filesystem'

// Composables
const filesystem = useFilesystem()

// Local state
const showCreateDirModal = ref(false)
const showCreateFileModal = ref(false)
const showRenameModal = ref(false)
const showDeleteModal = ref(false)
const newDirName = ref('')
const newFileName = ref('')
const renameValue = ref('')
const renameItemPath = ref('')
const itemsToDelete = ref<(FileItem | DirectoryItem)[]>([])
const currentPage = ref(1)
const itemsPerPage = 50

// Computed properties
const currentPath = computed(() => filesystem.currentPath.value)
const files = computed(() => filesystem.files.value)
const directories = computed(() => filesystem.directories.value)
const allItems = computed(() => filesystem.allItems.value)
const selectedItems = computed(() => filesystem.selectedItems.value)
const hasSelection = computed(() => filesystem.hasSelection.value)
const clipboard = computed(() => filesystem.clipboard.value)
const viewMode = computed(() => filesystem.viewMode.value)
const sortBy = computed(() => filesystem.sortBy.value)
const sortOrder = computed(() => filesystem.sortOrder.value)
const canGoBack = computed(() => filesystem.canGoBack.value)
const canGoForward = computed(() => filesystem.canGoForward.value)

const searchQuery = computed({
  get: () => filesystem.searchQuery.value,
  set: (value) => filesystem.setSearchQuery(value)
})

const pathParts = computed(() => {
  return currentPath.value.split('/').filter(Boolean)
})

const paginatedItems = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return allItems.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(allItems.value.length / itemsPerPage)
})

// Mock tree data (would be loaded from backend)
const treeItems = ref([
  { name: 'Home', path: '/home', isExpanded: true, isDirectory: true },
  { name: 'Documents', path: '/home/Documents', isExpanded: false, isDirectory: true },
  { name: 'Downloads', path: '/home/Downloads', isExpanded: false, isDirectory: true },
  { name: 'Pictures', path: '/home/Pictures', isExpanded: false, isDirectory: true },
  { name: 'Videos', path: '/home/Videos', isExpanded: false, isDirectory: true },
  { name: 'Music', path: '/home/Music', isExpanded: false, isDirectory: true },
])

const quickLocations = ref([
  { name: 'Home', path: '/home', icon: '🏠' },
  { name: 'Documents', path: '/home/Documents', icon: '📄' },
  { name: 'Downloads', path: '/home/Downloads', icon: '⬇️' },
  { name: 'Pictures', path: '/home/Pictures', icon: '🖼️' },
  { name: 'Videos', path: '/home/Videos', icon: '🎬' },
  { name: 'Music', path: '/home/Music', icon: '🎵' },
  { name: 'Desktop', path: '/home/Desktop', icon: '🖥️' },
])

// Methods
const navigateToPath = (path: string) => {
  filesystem.navigateToPath(path)
}

const navigateBack = () => {
  filesystem.navigateBack()
}

const navigateForward = () => {
  filesystem.navigateForward()
}

const navigateUp = () => {
  filesystem.navigateUp()
}

const navigateHome = () => {
  filesystem.navigateHome()
}

const getPathUpTo = (index: number) => {
  return '/' + pathParts.value.slice(0, index + 1).join('/')
}

const setViewMode = (mode: 'list' | 'grid' | 'tree') => {
  filesystem.setViewMode(mode)
}

const toggleSelection = (path: string) => {
  filesystem.toggleSelection(path)
}

const selectAll = () => {
  filesystem.selectAll()
}

const clearSelection = () => {
  filesystem.clearSelection()
}

const toggleSelectAll = () => {
  if (selectedItems.value.length === allItems.value.length) {
    clearSelection()
  } else {
    selectAll()
  }
}

const openItem = (item: FileItem | DirectoryItem) => {
  if (item.isDirectory) {
    navigateToPath(item.path)
  } else {
    // Open file (would need backend implementation)
    console.log('Open file:', item.path)
  }
}

const createDirectory = () => {
  showCreateDirModal.value = true
  newDirName.value = ''
}

const closeCreateDirModal = () => {
  showCreateDirModal.value = false
  newDirName.value = ''
}

const confirmCreateDirectory = async () => {
  if (newDirName.value.trim()) {
    const success = await filesystem.createDirectory(newDirName.value.trim())
    if (success) {
      closeCreateDirModal()
    }
  }
}

const createFile = () => {
  showCreateFileModal.value = true
  newFileName.value = ''
}

const closeCreateFileModal = () => {
  showCreateFileModal.value = false
  newFileName.value = ''
}

const confirmCreateFile = async () => {
  if (newFileName.value.trim()) {
    const success = await filesystem.createFile(newFileName.value.trim())
    if (success) {
      closeCreateFileModal()
    }
  }
}

const renameItem = (item: FileItem | DirectoryItem) => {
  renameItemPath.value = item.path
  renameValue.value = item.name
  showRenameModal.value = true
}

const closeRenameModal = () => {
  showRenameModal.value = false
  renameValue.value = ''
  renameItemPath.value = ''
}

const confirmRename = async () => {
  if (renameValue.value.trim() && renameItemPath.value) {
    const success = await filesystem.renameItem(renameItemPath.value, renameValue.value.trim())
    if (success) {
      closeRenameModal()
    }
  }
}

const deleteItem = (item: FileItem | DirectoryItem) => {
  itemsToDelete.value = [item]
  showDeleteModal.value = true
}

const deleteSelectedItems = () => {
  const selected = allItems.value.filter(item => selectedItems.value.includes(item.path))
  itemsToDelete.value = selected
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  itemsToDelete.value = []
}

const confirmDelete = async () => {
  const paths = itemsToDelete.value.map(item => item.path)
  const success = await filesystem.deleteItems(paths)
  if (success) {
    closeDeleteModal()
  }
}

const copyItems = () => {
  filesystem.copyItems(selectedItems.value)
}

const cutItems = () => {
  filesystem.cutItems(selectedItems.value)
}

const pasteItems = async () => {
  const success = await filesystem.pasteItems()
  if (success) {
    clearSelection()
  }
}

const updateSorting = () => {
  filesystem.setSorting(sortBy.value, sortOrder.value)
}

const toggleSortOrder = () => {
  const newOrder = sortOrder.value === 'asc' ? 'desc' : 'asc'
  filesystem.setSorting(sortBy.value, newOrder)
}

const refreshCurrentDirectory = async () => {
  await filesystem.loadDirectory(currentPath.value)
}

const changePage = (page: number) => {
  currentPage.value = page
}

const toggleTreeItem = (item: any) => {
  item.isExpanded = !item.isExpanded
  // Load tree children if needed
}

// Utility functions
const getItemIcon = (item: FileItem | DirectoryItem): string => {
  return filesystem.getFileIcon(item)
}

const formatFileSize = (bytes: number): string => {
  return filesystem.formatFileSize(bytes)
}

const formatDate = (date: Date | string): string => {
  return filesystem.formatDate(date)
}

// Initialize
onMounted(() => {
  filesystem.initialize()
  refreshCurrentDirectory()
})

// SEO
useHead({
  title: 'File Manager - aDex',
  meta: [
    { name: 'description', content: 'Advanced file manager with multiple view modes and file operations' }
  ]
})
</script>

<style scoped>
.filesystem-page {
  display: grid;
  grid-template-rows: auto auto 1fr;
  height: 100vh;
  background: var(--color-background);
}

.filesystem-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.page-title {
  margin: 0;
  color: var(--color-text);
  font-size: 1.5rem;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.breadcrumb-item {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  transition: color 0.2s ease;
}

.breadcrumb-item:hover {
  color: var(--color-primary);
}

.breadcrumb-item:not(:last-child)::after {
  content: '/';
  margin-left: 0.5rem;
  color: var(--color-text-secondary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.view-controls,
.navigation-controls {
  display: flex;
  gap: 0.25rem;
}

.view-btn,
.nav-btn {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.view-btn:hover,
.nav-btn:hover:not(:disabled) {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.view-btn.active,
.nav-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  gap: 1rem;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.toolbar-btn {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.875rem;
}

.toolbar-btn:hover:not(:disabled) {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.toolbar-btn.danger:hover:not(:disabled) {
  background: var(--color-error);
  border-color: var(--color-error);
}

.toolbar-separator {
  width: 1px;
  height: 20px;
  background: var(--color-border);
}

.search-box {
  display: flex;
  align-items: center;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  overflow: hidden;
}

.search-input {
  border: none;
  background: transparent;
  color: var(--color-text);
  padding: 0.5rem;
  outline: none;
  width: 200px;
}

.search-btn {
  background: none;
  border: none;
  padding: 0.5rem;
  cursor: pointer;
  color: var(--color-text-secondary);
}

.sort-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sort-controls select {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
}

.sort-order-btn {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  width: 32px;
}

.filesystem-content {
  display: grid;
  grid-template-columns: 250px 1fr;
  overflow: hidden;
}

.filesystem-sidebar {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--color-border);
}

.sidebar-header h3 {
  margin: 0;
  color: var(--color-text);
  font-size: 1rem;
}

.refresh-btn {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
}

.refresh-btn:hover {
  background: var(--color-background);
}

.tree-view {
  padding: 0.5rem;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  border-radius: 0.25rem;
  transition: background-color 0.2s ease;
}

.tree-item:hover {
  background: var(--color-background);
}

.tree-item.active {
  background: var(--color-primary);
  color: var(--color-background);
}

.tree-icon {
  font-size: 0.875rem;
}

.tree-label {
  font-size: 0.875rem;
}

.quick-access {
  padding: 1rem 0.5rem;
  border-top: 1px solid var(--color-border);
}

.quick-access h4 {
  margin: 0 0 0.5rem 0;
  color: var(--color-text);
  font-size: 0.875rem;
}

.quick-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  border-radius: 0.25rem;
  transition: background-color 0.2s ease;
}

.quick-item:hover {
  background: var(--color-background);
}

.quick-icon {
  font-size: 0.875rem;
}

.quick-label {
  font-size: 0.875rem;
  color: var(--color-text);
}

.file-list-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.file-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--color-background);
  border-bottom: 1px solid var(--color-border);
}

.list-info {
  display: flex;
  gap: 1rem;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.list-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.75rem;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.file-list {
  flex: 1;
  overflow-y: auto;
}

.list-header {
  display: grid;
  grid-template-columns: 40px 1fr 100px 150px 100px 100px;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  font-weight: bold;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.list-body {
  overflow-y: auto;
}

.list-row {
  display: grid;
  grid-template-columns: 40px 1fr 100px 150px 100px 100px;
  gap: 1rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.list-row:hover {
  background: var(--color-background);
}

.list-row.selected {
  background: var(--color-primary);
  color: var(--color-background);
}

.list-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  overflow: hidden;
}

.name-cell {
  font-weight: 500;
}

.item-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: background-color 0.2s ease;
}

.action-icon-btn:hover {
  background: var(--color-surface);
}

.file-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 1rem;
  padding: 1rem;
  overflow-y: auto;
}

.grid-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.grid-item:hover {
  background: var(--color-background);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--color-shadow);
}

.grid-item.selected {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.grid-item-icon {
  position: relative;
  margin-bottom: 0.5rem;
}

.item-icon-large {
  font-size: 2rem;
}

.grid-checkbox {
  position: absolute;
  top: -8px;
  right: -8px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
}

.grid-item-name {
  font-size: 0.875rem;
  font-weight: 500;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  width: 100%;
}

.grid-item-info {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-align: center;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
}

.page-btn {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.page-btn:hover:not(:disabled) {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
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

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.modal-header h3 {
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

.modal-content {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--color-text);
  font-weight: 500;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  color: var(--color-text);
  font-size: 1rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
}

.btn-primary,
.btn-secondary,
.btn-danger {
  padding: 0.75rem 1.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-primary {
  background: var(--color-primary);
  border: none;
  color: var(--color-background);
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-accent);
}

.btn-secondary {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
}

.btn-secondary:hover {
  background: var(--color-surface);
}

.btn-danger {
  background: var(--color-error);
  border: none;
  color: var(--color-background);
}

.btn-danger:hover {
  background: #cc0000;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.delete-list {
  max-height: 200px;
  overflow-y: auto;
  margin: 1rem 0;
  padding: 1rem;
  background: var(--color-background);
  border-radius: 0.25rem;
}

.delete-list li {
  padding: 0.25rem 0;
  color: var(--color-text);
}

/* Responsive */
@media (max-width: 1024px) {
  .filesystem-content {
    grid-template-columns: 1fr;
  }

  .filesystem-sidebar {
    display: none;
  }

  .list-header,
  .list-row {
    grid-template-columns: 40px 1fr 80px 120px 80px;
  }

  .toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 1rem;
  }

  .toolbar-left,
  .toolbar-right {
    justify-content: center;
  }
}

@media (max-width: 768px) {
  .filesystem-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  }

  .search-input {
    width: 150px;
  }
}
</style>
