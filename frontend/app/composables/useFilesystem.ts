import { ref, computed, watch } from 'vue'
import { useWails } from './useWails'
import type { FileItem, DirectoryItem, FileSystemStats, FileWatcher } from '~/types/filesystem'
import type { FileInfo, DirectoryEntry } from '../../bindings/aDex-UI/backend/services/filesystem/models'

export const useFilesystem = () => {
  const wails = useWails()
  
  // State
  const currentPath = ref('/')
  const files = ref<FileItem[]>([])
  const directories = ref<DirectoryItem[]>([])
  const isLoading = ref(false)
  const selectedItems = ref<string[]>([])
  const fileWatchers = ref<Map<string, FileWatcher>>(new Map())
  const clipboard = ref<{ items: string[], operation: 'copy' | 'cut' } | null>(null)
  const searchQuery = ref('')
  const sortBy = ref<'name' | 'size' | 'modified' | 'type'>('name')
  const sortOrder = ref<'asc' | 'desc'>('asc')

  // Computed properties
  const allItems = computed(() => {
    const dirs = directories.value.map(dir => ({ ...dir, isDirectory: true }))
    const files = files.value.map(file => ({ ...file, isDirectory: false }))
    const combined = [...dirs, ...files]
    
    // Apply search filter
    let filtered = combined
    if (searchQuery.value.trim()) {
      const query = searchQuery.value.toLowerCase()
      filtered = combined.filter(item => 
        item.name.toLowerCase().includes(query)
      )
    }
    
    // Apply sorting
    return filtered.sort((a, b) => {
      let comparison = 0
      
      switch (sortBy.value) {
        case 'name':
          comparison = a.name.localeCompare(b.name)
          break
        case 'size':
          comparison = (a.size || 0) - (b.size || 0)
          break
        case 'modified':
          comparison = new Date(a.modified).getTime() - new Date(b.modified).getTime()
          break
        case 'type':
          comparison = (a.type || '').localeCompare(b.type || '')
          break
      }
      
      return sortOrder.value === 'asc' ? comparison : -comparison
    })
  })

  const currentDirectory = computed(() => {
    return currentPath.value.split('/').filter(Boolean).pop() || '/'
  })

  const parentDirectory = computed(() => {
    const parts = currentPath.value.split('/').filter(Boolean)
    parts.pop()
    return '/' + parts.join('/')
  })

  const selectedFiles = computed(() => {
    return files.value.filter(file => selectedItems.value.includes(file.path))
  })

  const selectedDirectories = computed(() => {
    return directories.value.filter(dir => selectedItems.value.includes(dir.path))
  })

  const hasSelection = computed(() => selectedItems.value.length > 0)
  
  const canGoBack = computed(() => currentPath.value !== '/')
  const canGoForward = ref(false) // Would need navigation history

  // Convert DirectoryEntry to DirectoryItem
  const convertToDirectoryItem = (entry: DirectoryEntry): DirectoryItem => {
    return {
      name: entry.Name,
      path: entry.Path,
      size: entry.Size,
      modified: entry.ModTime.toISOString(),
      created: entry.ModTime.toISOString(),
      accessed: entry.ModTime.toISOString(),
      permissions: '', // Not available in DirectoryEntry
      owner: '', // Not available in DirectoryEntry
      group: '', // Not available in DirectoryEntry
      isHidden: entry.Name.startsWith('.'),
      itemCount: 0, // Not available in DirectoryEntry
    }
  }

  // Convert FileInfo to FileItem
  const convertToFileItem = (fileInfo: FileInfo): FileItem => {
    return {
      name: fileInfo.Name,
      path: fileInfo.Path,
      size: fileInfo.Size,
      modified: fileInfo.ModTime.toISOString(),
      created: fileInfo.ModTime.toISOString(),
      accessed: fileInfo.ModTime.toISOString(),
      permissions: fileInfo.Permissions,
      owner: fileInfo.Owner,
      group: fileInfo.Group,
      isHidden: fileInfo.Name.startsWith('.'),
      type: fileInfo.IsDirectory ? 'directory' : 'file',
      extension: fileInfo.IsDirectory ? '' : fileInfo.Name.split('.').pop() || '',
      mimeType: '', // Not available in FileInfo
    }
  }

  // Load directory contents
  const loadDirectory = async (path: string) => {
    if (!wails.isReady.value) {
      console.warn('Wails not ready for filesystem operations')
      return
    }

    try {
      isLoading.value = true
      const entries = await wails.filesystem.readDirectory(path)
      
      if (entries) {
        currentPath.value = path
        
        // Separate directories and files
        const dirs = entries.filter(entry => entry.IsDir)
        const files = entries.filter(entry => !entry.IsDir)
        
        directories.value = dirs.map(convertToDirectoryItem)
        files.value = files.map(convertToFileItem)
        
        // Clear selection when changing directory
        selectedItems.value = []
      }
    } catch (error) {
      console.error('Failed to load directory:', error)
    } finally {
      isLoading.value = false
    }
  }

  // Navigate to directory
  const navigateTo = (path: string) => {
    loadDirectory(path)
  }

  // Navigate up one level
  const navigateUp = () => {
    if (canGoBack.value) {
      navigateTo(parentDirectory.value)
    }
  }

  // Navigate to home directory
  const navigateHome = () => {
    navigateTo('/home')
  }

  // Navigate to path
  const navigateToPath = (path: string) => {
    if (path.startsWith('/')) {
      navigateTo(path)
    } else {
      navigateTo(currentPath.value + '/' + path)
    }
  }

  // Create directory
  const createDirectory = async (name: string) => {
    if (!wails.isReady.value || !name.trim()) {
      return false
    }

    try {
      const newPath = currentPath.value + '/' + name.trim()
      await wails.filesystem.createDirectory(newPath, 0o755)
      await loadDirectory(currentPath.value)
      return true
    } catch (error) {
      console.error('Failed to create directory:', error)
      return false
    }
  }

  // Create file
  const createFile = async (name: string, content: string = '') => {
    if (!wails.isReady.value || !name.trim()) {
      return false
    }

    try {
      const filePath = currentPath.value + '/' + name.trim()
      await wails.filesystem.writeFile(filePath, content, 0o644)
      await loadDirectory(currentPath.value)
      return true
    } catch (error) {
      console.error('Failed to create file:', error)
      return false
    }
  }

  // Delete items
  const deleteItems = async (paths: string[]) => {
    if (!wails.isReady.value || paths.length === 0) {
      return false
    }

    try {
      await Promise.all(
        paths.map(path => wails.filesystem.deleteFile(path))
      )
      
      await loadDirectory(currentPath.value)
      selectedItems.value = []
      return true
    } catch (error) {
      console.error('Failed to delete items:', error)
      return false
    }
  }

  // Copy files/directories
  const copyItems = (paths: string[]) => {
    clipboard.value = {
      items: [...paths],
      operation: 'copy'
    }
  }

  // Cut files/directories
  const cutItems = (paths: string[]) => {
    clipboard.value = {
      items: [...paths],
      operation: 'cut'
    }
  }

  // Paste items
  const pasteItems = async () => {
    if (!clipboard.value || !wails.isReady.value) {
      return false
    }

    try {
      const { items, operation } = clipboard.value
      
      if (operation === 'copy') {
        await Promise.all(
          items.map(async (item) => {
            const itemName = item.split('/').pop() || ''
            const destPath = currentPath.value + '/' + itemName
            await wails.filesystem.copyFile(item, destPath)
          })
        )
      } else if (operation === 'cut') {
        await Promise.all(
          items.map(async (item) => {
            const itemName = item.split('/').pop() || ''
            const destPath = currentPath.value + '/' + itemName
            await wails.filesystem.moveFile(item, destPath)
          })
        )
      }
      
      clipboard.value = null
      await loadDirectory(currentPath.value)
      return true
    } catch (error) {
      console.error('Failed to paste items:', error)
      return false
    }
  }

  // Rename item
  const renameItem = async (oldPath: string, newName: string) => {
    if (!wails.isReady.value || !newName.trim()) {
      return false
    }

    try {
      const newPath = oldPath.substring(0, oldPath.lastIndexOf('/')) + '/' + newName.trim()
      await wails.filesystem.moveFile(oldPath, newPath)
      await loadDirectory(currentPath.value)
      return true
    } catch (error) {
      console.error('Failed to rename item:', error)
      return false
    }
  }

  // Get file info
  const getFileInfo = async (path: string) => {
    if (!wails.isReady.value) {
      return null
    }

    try {
      const fileInfo = await wails.filesystem.getFileInfo(path)
      if (fileInfo) {
        return convertToFileItem(fileInfo)
      }
      return null
    } catch (error) {
      console.error('Failed to get file info:', error)
      return null
    }
  }

  // Read file content
  const readFile = async (path: string) => {
    if (!wails.isReady.value) {
      return null
    }

    try {
      const content = await wails.filesystem.readFile(path)
      if (content) {
        // Convert Uint8Array to string
        return new TextDecoder().decode(content)
      }
      return null
    } catch (error) {
      console.error('Failed to read file:', error)
      return null
    }
  }

  // Write file content
  const writeFile = async (path: string, content: string) => {
    if (!wails.isReady.value) {
      return false
    }

    try {
      await wails.filesystem.writeFile(path, content, 0o644)
      if (path.startsWith(currentPath.value)) {
        await loadDirectory(currentPath.value)
      }
      return true
    } catch (error) {
      console.error('Failed to write file:', error)
      return false
    }
  }

  // Watch file/directory for changes
  const watchPath = async (path: string) => {
    if (!wails.isReady.value) {
      return null
    }

    try {
      const watcher = await wails.filesystem.watchDirectory(path)
      
      if (watcher) {
        const fileWatcher: FileWatcher = {
          id: Date.now().toString(), // Generate unique ID
          path,
          isActive: true,
          createdAt: new Date(),
        }
        
        fileWatchers.value.set(path, fileWatcher)
        
        // Setup event listeners for the watcher
        // This would need backend implementation to send events
        
        return fileWatcher.id
      }
    } catch (error) {
      console.error('Failed to watch path:', error)
    }
    
    return null
  }

  // Stop watching path
  const unwatchPath = (path: string) => {
    const watcher = fileWatchers.value.get(path)
    if (watcher) {
      fileWatchers.value.delete(path)
      // Would need backend implementation to stop watching
    }
  }

  // Selection management
  const selectItem = (path: string) => {
    if (!selectedItems.value.includes(path)) {
      selectedItems.value.push(path)
    }
  }

  const deselectItem = (path: string) => {
    const index = selectedItems.value.indexOf(path)
    if (index > -1) {
      selectedItems.value.splice(index, 1)
    }
  }

  const toggleSelection = (path: string) => {
    if (selectedItems.value.includes(path)) {
      deselectItem(path)
    } else {
      selectItem(path)
    }
  }

  const selectAll = () => {
    selectedItems.value = allItems.value.map(item => item.path)
  }

  const clearSelection = () => {
    selectedItems.value = []
  }

  // Format file size
  const formatFileSize = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let size = bytes
    let unitIndex = 0
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    
    return size.toFixed(1) + ' ' + units[unitIndex]
  }

  // Format date
  const formatDate = (date: Date | string): string => {
    const d = new Date(date)
    return d.toLocaleDateString() + ' ' + d.toLocaleTimeString()
  }

  // Get file icon based on type
  const getFileIcon = (item: FileItem | DirectoryItem): string => {
    if ('isDirectory' in item || item.type === 'directory') {
      return 'folder'
    }
    
    const extension = (item as FileItem).extension?.toLowerCase()
    const iconMap: Record<string, string> = {
      'txt': 'file-text',
      'json': 'file-code',
      'js': 'file-code',
      'ts': 'file-code',
      'vue': 'file-code',
      'html': 'file-code',
      'css': 'file-code',
      'md': 'file-text',
      'pdf': 'file-pdf',
      'jpg': 'file-image',
      'jpeg': 'file-image',
      'png': 'file-image',
      'gif': 'file-image',
      'mp3': 'file-audio',
      'mp4': 'file-video',
      'zip': 'file-archive',
      'tar': 'file-archive',
      'gz': 'file-archive',
    }
    
    return iconMap[extension || ''] || 'file'
  }

  // Initialize filesystem
  const initialize = async () => {
    if (wails.isReady.value) {
      await loadDirectory(currentPath.value)
    }
  }

  // Watch for Wails availability
  watch(() => wails.isReady.value, (isReady) => {
    if (isReady) {
      initialize()
    }
  })

  // Auto-initialize when Wails is ready
  if (wails.isReady.value) {
    initialize()
  }

  return {
    // State
    currentPath: readonly(currentPath),
    files: readonly(files),
    directories: readonly(directories),
    isLoading: readonly(isLoading),
    selectedItems,
    searchQuery,
    sortBy,
    sortOrder,
    clipboard: readonly(clipboard),

    // Computed
    allItems,
    currentDirectory,
    parentDirectory,
    selectedFiles,
    selectedDirectories,
    hasSelection,
    canGoBack,
    canGoForward,

    // Navigation
    navigateTo,
    navigateUp,
    navigateHome,
    navigateToPath,
    loadDirectory,

    // File operations
    createDirectory,
    createFile,
    deleteItems,
    copyItems,
    cutItems,
    pasteItems,
    renameItem,
    getFileInfo,
    readFile,
    writeFile,

    // Watchers
    watchPath,
    unwatchPath,

    // Selection
    selectItem,
    deselectItem,
    toggleSelection,
    selectAll,
    clearSelection,

    // Utilities
    formatFileSize,
    formatDate,
    getFileIcon,
    initialize,
  }
}
