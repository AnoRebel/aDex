import { beforeAll, vi } from 'vitest'

// Mock browser APIs
beforeAll(() => {
  // Mock ResizeObserver
  global.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn()
  }))

  // MutationObserver
  global.MutationObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    disconnect: vi.fn(),
    takeRecords: vi.fn(() => [])
  }))

  // IntersectionObserver
  global.IntersectionObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn()
  }))

  // matchMedia
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn()
    }))
  })

  // getComputedStyle
  Object.defineProperty(window, 'getComputedStyle', {
    writable: true,
    value: vi.fn().mockImplementation(() => ({
      getPropertyValue: () => '',
      appearance: ''
    }))
  })

  // requestAnimationFrame
  global.requestAnimationFrame = vi.fn((cb) => setTimeout(cb, 16))
  global.cancelAnimationFrame = vi.fn()

  // Performance API
  global.performance = {
    ...global.performance,
    now: vi.fn(() => Date.now())
  }

  // CustomEvent constructor
  global.CustomEvent = class CustomEvent extends Event {
    constructor(type: string, options?: CustomEventInit) {
      super(type, options)
    }
  } as any

  // URL.createObjectURL
  global.URL.createObjectURL = vi.fn(() => 'mock-url')
  global.URL.revokeObjectURL = vi.fn()

  // Blob
  global.Blob = class Blob {
    constructor(content: any[], options?: BlobPropertyBag) {}
  } as any

  // File
  global.File = class File {
    constructor(content: any[], name: string, options?: FilePropertyBag) {}
  } as any

  // FileReader
  global.FileReader = class FileReader {
    readAsText = vi.fn()
    readAsDataURL = vi.fn()
    readAsArrayBuffer = vi.fn()
    onload = null as any
    onerror = null as any
  } as any

  // Mock CSS.supports
  Object.defineProperty(window, 'CSS', {
    writable: true,
    value: {
      supports: vi.fn(() => false)
    }
  })

  // Mock localStorage
  const localStorageMock = (() => {
    let store: Record<string, string> = {}
    return {
      getItem: vi.fn((key) => store[key] || null),
      setItem: vi.fn((key, value) => {
        store[key] = value.toString()
      }),
      removeItem: vi.fn((key) => {
        delete store[key]
      }),
      clear: vi.fn(() => {
        store = {}
      }),
      get length() {
        return Object.keys(store).length
      },
      key: vi.fn((index) => {
        const keys = Object.keys(store)
        return keys[index] || null
      })
    }
  })()

  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock
  })

  // Mock sessionStorage
  const sessionStorageMock = (() => {
    let store: Record<string, string> = {}
    return {
      getItem: vi.fn((key) => store[key] || null),
      setItem: vi.fn((key, value) => {
        store[key] = value.toString()
      }),
      removeItem: vi.fn((key) => {
        delete store[key]
      }),
      clear: vi.fn(() => {
        store = {}
      }),
      get length() {
        return Object.keys(store).length
      },
      key: vi.fn((index) => {
        const keys = Object.keys(store)
        return keys[index] || null
      })
    }
  })()

  Object.defineProperty(window, 'sessionStorage', {
    value: sessionStorageMock
  })

  // Mock fetch
  global.fetch = vi.fn()

  // Mock WebSocket
  global.WebSocket = vi.fn().mockImplementation(() => ({
    close: vi.fn(),
    send: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
    CONNECTING: 0,
    OPEN: 1,
    CLOSING: 2,
    CLOSED: 3
  })) as any

  // Mock console methods
  console.error = vi.fn()
  console.warn = vi.fn()
  console.log = vi.fn()
  console.info = vi.fn()
  console.debug = vi.fn()
})