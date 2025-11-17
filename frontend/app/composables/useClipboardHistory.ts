import { ref, computed, watch } from 'vue'
import type { ClipboardHistory } from './useClipboard'

export function useClipboardHistory(maxSize: number = 50) {
  // State
  const history = ref<ClipboardHistory[]>([])
  const maxHistorySize = ref(maxSize)

  // Computed
  const hasHistory = computed(() => history.value.length > 0)
  const recentItems = computed(() => history.value.slice(0, 10))
  const itemCount = computed(() => history.value.length)

  // Methods
  const addToHistory = (item: Partial<ClipboardHistory>): void => {
    const historyItem: ClipboardHistory = {
      id: generateId(),
      content: item.content || '',
      timestamp: new Date().toISOString(),
      length: item.content?.length || 0,
      type: item.type || 'text',
      source: item.source
    }

    // Add to beginning of history
    history.value.unshift(historyItem)

    // Limit history size
    if (history.value.length > maxHistorySize.value) {
      history.value = history.value.slice(0, maxHistorySize.value)
    }

    // Save to localStorage
    saveToStorage()
  }

  const removeFromHistory = (id: string): void => {
    history.value = history.value.filter(item => item.id !== id)
    saveToStorage()
  }

  const clearHistory = (): void => {
    history.value = []
    saveToStorage()
  }

  const getHistoryItem = (id: string): ClipboardHistory | undefined => {
    return history.value.find(item => item.id === id)
  }

  const searchHistory = (query: string): ClipboardHistory[] => {
    if (!query.trim()) return history.value

    const searchTerm = query.toLowerCase()
    return history.value.filter(item =>
      item.content.toLowerCase().includes(searchTerm) ||
      item.source?.toLowerCase().includes(searchTerm)
    )
  }

  const filterByType = (type: string): ClipboardHistory[] => {
    return history.value.filter(item => item.type === type)
  }

  const filterByDateRange = (startDate: Date, endDate: Date): ClipboardHistory[] => {
    const start = startDate.getTime()
    const end = endDate.getTime()

    return history.value.filter(item => {
      const itemTime = new Date(item.timestamp).getTime()
      return itemTime >= start && itemTime <= end
    })
  }

  const generateId = (): string => {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  const saveToStorage = (): void => {
    try {
      const data = {
        history: history.value,
        maxSize: maxHistorySize.value
      }
      localStorage.setItem('terminal-clipboard-history', JSON.stringify(data))
    } catch (error) {
      console.warn('Failed to save clipboard history to localStorage:', error)
    }
  }

  const loadFromStorage = (): void => {
    try {
      const stored = localStorage.getItem('terminal-clipboard-history')
      if (stored) {
        const data = JSON.parse(stored)
        history.value = data.history || []
        maxHistorySize.value = data.maxSize || maxSize
      }
    } catch (error) {
      console.warn('Failed to load clipboard history from localStorage:', error)
    }
  }

  const exportHistory = (): string => {
    const data = {
      history: history.value,
      exportedAt: new Date().toISOString(),
      version: '1.0'
    }
    return JSON.stringify(data, null, 2)
  }

  const importHistory = (jsonString: string): boolean => {
    try {
      const data = JSON.parse(jsonString)

      if (data.history && Array.isArray(data.history)) {
        history.value = data.history
        maxHistorySize.value = data.maxSize || maxSize
        saveToStorage()
        return true
      }

      return false
    } catch (error) {
      console.error('Failed to import clipboard history:', error)
      return false
    }
  }

  const updateMaxSize = (newSize: number): void => {
    maxHistorySize.value = newSize

    // Trim history if necessary
    if (history.value.length > newSize) {
      history.value = history.value.slice(0, newSize)
    }

    saveToStorage()
  }

  // Auto-cleanup old items (older than 7 days)
  const cleanupOldItems = (): void => {
    const sevenDaysAgo = new Date()
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7)

    const filtered = history.value.filter(item => {
      const itemDate = new Date(item.timestamp)
      return itemDate > sevenDaysAgo
    })

    if (filtered.length !== history.value.length) {
      history.value = filtered
      saveToStorage()
    }
  }

  // Load from storage on initialization
  loadFromStorage()

  // Periodically clean up old items
  setInterval(cleanupOldItems, 24 * 60 * 60 * 1000) // Once per day

  // Watch for max size changes
  watch(maxHistorySize, (newSize) => {
    if (history.value.length > newSize) {
      history.value = history.value.slice(0, newSize)
      saveToStorage()
    }
  })

  return {
    // State
    history,
    maxHistorySize,

    // Computed
    hasHistory,
    recentItems,
    itemCount,

    // Methods
    addToHistory,
    removeFromHistory,
    clearHistory,
    getHistoryItem,
    searchHistory,
    filterByType,
    filterByDateRange,
    exportHistory,
    importHistory,
    updateMaxSize,
    cleanupOldItems
  }
}

export default useClipboardHistory