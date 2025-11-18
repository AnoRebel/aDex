import { defineStore } from 'pinia'
import { useWails } from '~/composables/useWails'
import type { FileItem, DirectoryItem, FileSystemStats, FileOperation } from '~/types/filesystem'

interface FilesystemState {
  currentPath: string
  files: FileItem[]
  directories: DirectoryItem[]
  selectedItems: string[]
  clipboard: {
    items: string[]
    operation: 'copy' | 'cut'
  } | null
  searchQuery: string
  sortBy: 'name' | 'size' | 'modified' | 'type'
  sortOrder: 'asc' | 'desc'
  viewMode: 'list' | 'grid' | 'tree'
  isLoading: boolean
  operations: FileOperation[]
  stats: FileSystemStats | null
  history: string[]
  historyIndex: number
  alerts: Array<{
    id: string
    type: 'warning' | 'error' | 'info'
    title: string
    message: string
    timestamp: Date
    acknowledged: boolean
  }>
  settings: {
    showHiddenFiles: boolean
    showPreview: boolean
    enableThumbnails: boolean
    itemsPerPage: number
    doubleClickAction: 'open' | 'rename' | 'properties'
    confirmDelete: boolean
  }
}

export const useFilesystemStore = defineStore('filesystem', {
  state: (): FilesystemState => ({
    currentPath: '/',
    files: [],
    directories: [],
    selectedItems: [],
    clipboard: null,
    searchQuery: '',
    sortBy: 'name',
    sortOrder: 'asc',
    viewMode: 'list',
    isLoading: false,
    operations: [],
    stats: null,
    history: ['/'],
    historyIndex: 0,
    alerts: [],
    settings: {
      showHiddenFiles: false,
      showPreview: true,
      enableThumbnails: true,
      itemsPerPage: 50,
      doubleClickAction: 'open',
      confirmDelete: true,
    },
  }),

  getters: {
    allItems: (state) => {
      const dirs = state.directories.map(dir => ({ ...dir, isDirectory: true }))
      const files = state.files.map(file => ({ ...file, isDirectory: false }))
      const combined = [...dirs, ...files]
      
      // Apply search filter
      let filtered = combined
      if (state.searchQuery.trim()) {
        const query = state.searchQuery.toLowerCase()
        filtered = combined.filter(item => 
          item.name.toLowerCase().includes(query)
        )
      }
      
      // Apply sorting
      return filtered.sort((a, b) => {
        let comparison = 0
        
        switch (state.sortBy) {
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
        
        return state.sortOrder === 'asc' ? comparison : -comparison
      })
    },

    currentDirectory: (state) => {
      return state.currentPath.split('/').filter(Boolean).pop() || '/'
    },

    parentDirectory: (state) => {
      const parts = state.currentPath.split('/').filter(Boolean)
      parts.pop()
      return '/' + parts.join('/')
    },

    selectedFiles: (state) => {
      return state.files.filter(file => state.selectedItems.includes(file.path))
    },

    selectedDirectories: (state) => {
      return state.directories.filter(dir => state.selectedItems.includes(dir.path))
    },

    hasSelection: (state) => state.selectedItems.length > 0,

    canGoBack: (state) => state.historyIndex > 0,
    
    canGoForward: (state) => state.historyIndex < state.history.length - 1,

    activeOperations: (state) => 
      state.operations.filter(op => op.status === 'running'),

    completedOperations: (state) => 
      state.operations.filter(op => op.status === 'completed'),

    failedOperations: (state) => 
      state.operations.filter(op => op.status === 'failed'),

    totalSize: (state) => {
      return [...state.files, ...state.directories].reduce((total, item) => total + item.size, 0)
    },
  },

  actions: {
    async fetchDirectory(path: string): Promise<void> {
      try {
        this.setLoading(true)

        // Call Go backend via Wails bindings
        const { filesystem } = useWails()
        const data = await filesystem.readDirectory(path)

        if (data) {
          // Separate directories and files from the response
          const directories: DirectoryItem[] = []
          const files: FileItem[] = []

          if (Array.isArray(data)) {
            data.forEach((item: any) => {
              if (item.isDir || item.type === 'directory') {
                directories.push({
                  name: item.name,
                  path: item.path,
                  size: item.size || 0,
                  modified: item.modTime || new Date().toISOString(),
                  type: 'directory'
                })
              } else {
                files.push({
                  name: item.name,
                  path: item.path,
                  size: item.size || 0,
                  modified: item.modTime || new Date().toISOString(),
                  type: item.type || 'file',
                  extension: item.name.includes('.') ? item.name.split('.').pop() : ''
                })
              }
            })
          }

          this.setDirectories(directories)
          this.setFiles(files)
        }

        this.setCurrentPath(path)

      } catch (error) {
        console.error('Failed to fetch directory:', error)
        this.addAlert({
          type: 'error',
          title: 'Directory Access Error',
          message: `Failed to access directory ${path}: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
      } finally {
        this.setLoading(false)
      }
    },

    async createDirectory(name: string, parentPath?: string): Promise<boolean> {
      try {
        const basePath = parentPath || this.currentPath
        const fullPath = basePath === '/' ? `/${name}` : `${basePath}/${name}`

        const { filesystem } = useWails()
        await filesystem.createDirectory(fullPath, 0o755)

        await this.fetchDirectory(this.currentPath)
        return true

      } catch (error) {
        console.error('Failed to create directory:', error)
        this.addAlert({
          type: 'error',
          title: 'Create Directory Error',
          message: `Failed to create directory: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
        return false
      }
    },

    async deleteItems(paths: string[]): Promise<boolean> {
      try {
        const { filesystem } = useWails()

        // Delete each item
        for (const path of paths) {
          await filesystem.deleteFile(path)
        }

        // Remove items from local state
        paths.forEach(path => {
          this.removeFile(path)
          this.removeDirectory(path)
        })

        this.clearSelection()
        return true

      } catch (error) {
        console.error('Failed to delete items:', error)
        this.addAlert({
          type: 'error',
          title: 'Delete Error',
          message: `Failed to delete items: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
        return false
      }
    },

    async renameItem(oldPath: string, newName: string): Promise<boolean> {
      try {
        const newPath = oldPath.substring(0, oldPath.lastIndexOf('/')) + '/' + newName

        const { filesystem } = useWails()
        await filesystem.renameFile(oldPath, newPath)

        await this.fetchDirectory(this.currentPath)
        return true

      } catch (error) {
        console.error('Failed to rename item:', error)
        this.addAlert({
          type: 'error',
          title: 'Rename Error',
          message: `Failed to rename item: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
        return false
      }
    },

    async copyItems(sourcePaths: string[], destinationPath: string): Promise<boolean> {
      try {
        const { filesystem } = useWails()

        // Copy each item to destination
        for (const sourcePath of sourcePaths) {
          const fileName = sourcePath.split('/').pop() || ''
          const destPath = destinationPath === '/' ? `/${fileName}` : `${destinationPath}/${fileName}`
          await filesystem.copyFile(sourcePath, destPath)
        }

        await this.fetchDirectory(this.currentPath)
        return true

      } catch (error) {
        console.error('Failed to copy items:', error)
        this.addAlert({
          type: 'error',
          title: 'Copy Error',
          message: `Failed to copy items: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
        return false
      }
    },

    async moveItems(sourcePaths: string[], destinationPath: string): Promise<boolean> {
      try {
        const { filesystem } = useWails()

        // Move each item to destination
        for (const sourcePath of sourcePaths) {
          const fileName = sourcePath.split('/').pop() || ''
          const destPath = destinationPath === '/' ? `/${fileName}` : `${destinationPath}/${fileName}`
          await filesystem.moveFile(sourcePath, destPath)
        }

        // Remove items from local state
        sourcePaths.forEach(path => {
          this.removeFile(path)
          this.removeDirectory(path)
        })

        await this.fetchDirectory(this.currentPath)
        return true

      } catch (error) {
        console.error('Failed to move items:', error)
        this.addAlert({
          type: 'error',
          title: 'Move Error',
          message: `Failed to move items: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
        return false
      }
    },

    async searchFiles(query: string, searchPath?: string): Promise<FileItem[] | DirectoryItem[]> {
      try {
        const { filesystem } = useWails()
        const path = searchPath || this.currentPath
        const results = await filesystem.searchFiles(path, query)
        return results || []

      } catch (error) {
        console.error('Search failed:', error)
        return []
      }
    },

    async getFileSystemStats(): Promise<void> {
      try {
        // Stats would be fetched from the system service
        // For now, use default stats
        this.setStats({
          totalSpace: 0,
          usedSpace: 0,
          freeSpace: 0,
          fileCount: this.files.length,
          directoryCount: this.directories.length
        })

      } catch (error) {
        console.error('Failed to get filesystem stats:', error)
      }
    },

    async getFileContent(path: string): Promise<string> {
      try {
        const { filesystem } = useWails()
        const content = await filesystem.readFile(path)
        return content || ''

      } catch (error) {
        console.error('Failed to read file:', error)
        throw error
      }
    },

    async saveFileContent(path: string, content: string): Promise<boolean> {
      try {
        const { filesystem } = useWails()
        await filesystem.writeFile(path, content, 0o644)
        return true

      } catch (error) {
        console.error('Failed to save file:', error)
        this.addAlert({
          type: 'error',
          title: 'Save Error',
          message: `Failed to save file: ${error instanceof Error ? error.message : 'Unknown error'}`,
          timestamp: new Date(),
          acknowledged: false
        })
        return false
      }
    },

    setCurrentPath(path: string) {
      this.currentPath = path

      // Update history
      if (this.history[this.historyIndex] !== path) {
        // Remove any forward history
        this.history = this.history.slice(0, this.historyIndex + 1)
        // Add new path
        this.history.push(path)
        this.historyIndex = this.history.length - 1

        // Keep history size manageable
        if (this.history.length > 50) {
          this.history.shift()
          this.historyIndex--
        }
      }

      // Clear selection when changing directory
      this.selectedItems = []
    },

    setFiles(files: FileItem[]) {
      this.files = files
    },

    setDirectories(directories: DirectoryItem[]) {
      this.directories = directories
    },

    addFile(file: FileItem) {
      this.files.push(file)
    },

    removeFile(path: string) {
      this.files = this.files.filter(file => file.path !== path)
      this.selectedItems = this.selectedItems.filter(item => item !== path)
    },

    addDirectory(directory: DirectoryItem) {
      this.directories.push(directory)
    },

    removeDirectory(path: string) {
      this.directories = this.directories.filter(dir => dir.path !== path)
      this.selectedItems = this.selectedItems.filter(item => item !== path)
    },

    setSelectedItems(items: string[]) {
      this.selectedItems = items
    },

    toggleSelection(path: string) {
      const index = this.selectedItems.indexOf(path)
      if (index > -1) {
        this.selectedItems.splice(index, 1)
      } else {
        this.selectedItems.push(path)
      }
    },

    selectAll() {
      this.selectedItems = this.allItems.map(item => item.path)
    },

    clearSelection() {
      this.selectedItems = []
    },

    setClipboard(items: string[], operation: 'copy' | 'cut') {
      this.clipboard = { items, operation }
    },

    clearClipboard() {
      this.clipboard = null
    },

    setSearchQuery(query: string) {
      this.searchQuery = query
    },

    setSorting(sortBy: FilesystemState['sortBy'], sortOrder: FilesystemState['sortOrder']) {
      this.sortBy = sortBy
      this.sortOrder = sortOrder
    },

    setViewMode(mode: FilesystemState['viewMode']) {
      this.viewMode = mode
    },

    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
    },

    addOperation(operation: FileOperation) {
      this.operations.push(operation)
      
      // Keep only last 100 operations
      if (this.operations.length > 100) {
        this.operations.splice(0, this.operations.length - 100)
      }
    },

    updateOperation(operationId: string, updates: Partial<FileOperation>) {
      const operation = this.operations.find(op => 
        op.source.includes(operationId) || op.destination === operationId
      )
      if (operation) {
        Object.assign(operation, updates)
      }
    },

    setStats(stats: FileSystemStats) {
      this.stats = stats
    },

    navigateBack() {
      if (this.canGoBack) {
        this.historyIndex--
        this.setCurrentPath(this.history[this.historyIndex])
      }
    },

    navigateForward() {
      if (this.canGoForward) {
        this.historyIndex++
        this.setCurrentPath(this.history[this.historyIndex])
      }
    },

    navigateUp() {
      if (this.currentPath !== '/') {
        this.setCurrentPath(this.parentDirectory)
      }
    },

    updateSettings(newSettings: Partial<FilesystemState['settings']>) {
      this.settings = { ...this.settings, ...newSettings }
      
      // Save to localStorage
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('adex-filesystem-settings', JSON.stringify(this.settings))
      }
    },

    loadSettings() {
      if (typeof localStorage !== 'undefined') {
        const saved = localStorage.getItem('adex-filesystem-settings')
        if (saved) {
          try {
            const settings = JSON.parse(saved)
            this.settings = { ...this.settings, ...settings }
          } catch (error) {
            console.error('Failed to load filesystem settings:', error)
          }
        }
      }
    },

    formatFileSize(bytes: number): string {
      const units = ['B', 'KB', 'MB', 'GB', 'TB']
      let size = bytes
      let unitIndex = 0
      
      while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024
        unitIndex++
      }
      
      return size.toFixed(1) + ' ' + units[unitIndex]
    },

    formatDate(date: Date | string): string {
      const d = new Date(date)
      return d.toLocaleDateString() + ' ' + d.toLocaleTimeString()
    },

    getFileIcon(item: FileItem | DirectoryItem): string {
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
    },

    addAlert(alert: Omit<FilesystemState['alerts'][0], 'id'>) {
      const id = Date.now().toString()
      this.alerts.push({
        id,
        ...alert,
      })

      // Keep only last 50 alerts
      if (this.alerts.length > 50) {
        this.alerts.splice(0, this.alerts.length - 50)
      }

      // Auto-remove info alerts after 15 seconds
      if (alert.type === 'info') {
        setTimeout(() => {
          this.removeAlert(id)
        }, 15000)
      }
    },

    removeAlert(id: string) {
      this.alerts = this.alerts.filter(alert => alert.id !== id)
    },

    acknowledgeAlert(id: string) {
      const alert = this.alerts.find(a => a.id === id)
      if (alert) {
        alert.acknowledged = true
      }
    },

    // Initialize store
    async initialize(): Promise<void> {
      this.loadSettings()

      // Initial directory fetch
      await this.fetchDirectory(this.currentPath)

      // Get filesystem stats
      await this.getFileSystemStats()
    },

    // Reset store
    reset() {
      this.currentPath = '/'
      this.files = []
      this.directories = []
      this.selectedItems = []
      this.clipboard = null
      this.searchQuery = ''
      this.operations = []
      this.stats = null
      this.history = ['/']
      this.historyIndex = 0
      this.alerts = []
    },
  },
})
