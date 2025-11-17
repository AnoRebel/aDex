import { ref, computed } from 'vue'
import { useUIStore } from '~/stores/ui'

export interface ClipboardOptions {
  onSuccess?: (message: string) => void
  onError?: (error: Error) => void
  enableNotifications?: boolean
  notificationDuration?: number
}

export interface ClipboardHistory {
  id: string
  content: string
  timestamp: string
  length: number
  type: 'text' | 'image' | 'file'
  source?: string
}

export function useClipboard(options: ClipboardOptions = {}) {
  const {
    onSuccess,
    onError,
    enableNotifications = true,
    notificationDuration = 3000
  } = options

  const uiStore = useUIStore()

  // State
  const isSupported = ref(!!navigator.clipboard)
  const isWriting = ref(false)
  const isReading = ref(false)
  const lastError = ref<Error | null>(null)
  const history = ref<ClipboardHistory[]>([])
  const maxHistorySize = ref(50)

  // Computed
  const canWrite = computed(() => isSupported.value && !isWriting.value)
  const canRead = computed(() => isSupported.value && !isReading.value)
  const hasHistory = computed(() => history.value.length > 0)
  const recentItems = computed(() => history.value.slice(0, 10))

  // Methods
  const writeText = async (text: string, source?: string): Promise<boolean> => {
    if (!canWrite.value) {
      const error = new Error('Clipboard write not supported or busy')
      handleError(error)
      return false
    }

    isWriting.value = true
    lastError.value = null

    try {
      await navigator.clipboard.writeText(text)

      // Add to history
      addToHistory({
        content: text,
        type: 'text',
        source
      })

      // Show notification
      if (enableNotifications) {
        const message = getSuccessMessage(text, source)
        onSuccess?.(message)

        uiStore.addNotification({
          type: 'success',
          title: 'Copied',
          message,
          persistent: false
        })
      }

      return true

    } catch (error) {
      handleError(error as Error)
      return false

    } finally {
      isWriting.value = false
    }
  }

  const readText = async (): Promise<string | null> => {
    if (!canRead.value) {
      const error = new Error('Clipboard read not supported or busy')
      handleError(error)
      return null
    }

    isReading.value = true
    lastError.value = null

    try {
      const text = await navigator.clipboard.readText()
      return text

    } catch (error) {
      handleError(error as Error)
      return null

    } finally {
      isReading.value = false
    }
  }

  const copySelection = async (element?: HTMLElement): Promise<boolean> => {
    try {
      let selectedText = ''

      if (element) {
        // Copy from specific element
        const selection = window.getSelection()
        if (selection) {
          const range = document.createRange()
          range.selectNodeContents(element)
          selection.removeAllRanges()
          selection.addRange(range)
          selectedText = selection.toString()
        }
      } else {
        // Copy current selection
        const selection = window.getSelection()
        selectedText = selection?.toString() || ''
      }

      if (!selectedText) {
        throw new Error('No text selected')
      }

      return await writeText(selectedText, 'selection')

    } catch (error) {
      handleError(error as Error)
      return false
    }
  }

  const copyElement = async (element: HTMLElement): Promise<boolean> => {
    try {
      const text = element.textContent || element.innerText || ''
      if (!text) {
        throw new Error('Element has no text content')
      }

      return await writeText(text, 'element')

    } catch (error) {
      handleError(error as Error)
      return false
    }
  }

  const paste = async (target?: HTMLInputElement | HTMLTextAreaElement): Promise<string | null> => {
    const text = await readText()

    if (text && target) {
      // Insert text at cursor position
      const start = target.selectionStart || 0
      const end = target.selectionEnd || 0
      const currentValue = target.value || ''

      target.value = currentValue.substring(0, start) + text + currentValue.substring(end)
      target.selectionStart = target.selectionEnd = start + text.length

      // Dispatch input event
      target.dispatchEvent(new Event('input', { bubbles: true }))
    }

    return text
  }

  const clearHistory = (): void => {
    history.value = []
  }

  const clearHistoryItem = (id: string): void => {
    history.value = history.value.filter(item => item.id !== id)
  }

  const getHistoryItem = (id: string): ClipboardHistory | undefined => {
    return history.value.find(item => item.id === id)
  }

  const copyFromHistory = async (id: string): Promise<boolean> => {
    const item = getHistoryItem(id)
    if (!item) {
      handleError(new Error('History item not found'))
      return false
    }

    return await writeText(item.content, 'history')
  }

  // Helper methods
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
  }

  const generateId = (): string => {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  const getSuccessMessage = (text: string, source?: string): string => {
    const lines = text.split('\n').length
    const chars = text.length

    if (source === 'selection') {
      if (lines > 1) {
        return `${lines} lines copied to clipboard`
      } else {
        return `Selection copied to clipboard`
      }
    } else if (source === 'element') {
      return `Content copied to clipboard`
    } else if (source === 'history') {
      return `Restored from clipboard history`
    } else {
      if (lines > 1) {
        return `${lines} lines (${chars} chars) copied to clipboard`
      } else if (chars > 50) {
        return `${chars} characters copied to clipboard`
      } else {
        return `"${text.substring(0, 30)}${chars > 30 ? '...' : ''}" copied to clipboard`
      }
    }
  }

  const handleError = (error: Error): void => {
    lastError.value = error
    console.error('Clipboard error:', error)

    if (enableNotifications) {
      let message = 'Failed to copy to clipboard'

      if (error.message.includes('permission')) {
        message = 'Clipboard access denied. Please grant permission.'
      } else if (error.message.includes('NotAllowedError')) {
        message = 'Clipboard access not allowed in this context.'
      } else if (error.message.includes('No text selected')) {
        message = 'No text selected to copy.'
      }

      onError?.(error)

      uiStore.addNotification({
        type: 'error',
        title: 'Copy Failed',
        message,
        persistent: false
      })
    }
  }

  // Utility methods for terminal-specific operations
  const copyTerminalOutput = async (output: string[]): Promise<boolean> => {
    const text = output.join('')
    if (!text) {
      handleError(new Error('No terminal output to copy'))
      return false
    }

    return await writeText(text, 'terminal')
  }

  const copyTerminalSelection = async (terminal: any): Promise<boolean> => {
    try {
      const selection = terminal.getSelection()
      if (!selection) {
        throw new Error('No terminal selection')
      }

      return await writeText(selection, 'terminal')
    } catch (error) {
      handleError(error as Error)
      return false
    }
  }

  const pasteToTerminal = async (terminal: any): Promise<boolean> => {
    try {
      const text = await readText()
      if (!text) {
        throw new Error('No clipboard content to paste')
      }

      terminal.paste(text)
      return true

    } catch (error) {
      handleError(error as Error)
      return false
    }
  }

  // Keyboard shortcuts
  const setupKeyboardShortcuts = (element?: HTMLElement): void => {
    const handleKeyDown = async (event: KeyboardEvent) => {
      // Ctrl+C / Cmd+C - Copy
      if ((event.ctrlKey || event.metaKey) && event.key === 'c' && !event.shiftKey) {
        const selection = window.getSelection()?.toString()
        if (selection) {
          event.preventDefault()
          await copySelection()
        }
      }

      // Ctrl+V / Cmd+V - Paste
      if ((event.ctrlKey || event.metaKey) && event.key === 'v') {
        event.preventDefault()
        await paste()
      }

      // Ctrl+Shift+C - Copy in terminal (when Ctrl+C is used for interrupt)
      if ((event.ctrlKey || event.metaKey) && event.shiftKey && event.key === 'C') {
        event.preventDefault()
        await copySelection()
      }

      // Ctrl+Shift+V - Paste in terminal
      if ((event.ctrlKey || event.metaKey) && event.shiftKey && event.key === 'V') {
        event.preventDefault()
        await paste()
      }
    }

    const targetElement = element || document
    targetElement.addEventListener('keydown', handleKeyDown)

    // Return cleanup function
    return () => {
      targetElement.removeEventListener('keydown', handleKeyDown)
    }
  }

  return {
    // State
    isSupported,
    isWriting,
    isReading,
    lastError,
    history,
    maxHistorySize,

    // Computed
    canWrite,
    canRead,
    hasHistory,
    recentItems,

    // Methods
    writeText,
    readText,
    copySelection,
    copyElement,
    paste,
    clearHistory,
    clearHistoryItem,
    getHistoryItem,
    copyFromHistory,

    // Terminal-specific
    copyTerminalOutput,
    copyTerminalSelection,
    pasteToTerminal,

    // Utilities
    setupKeyboardShortcuts
  }
}

export default useClipboard