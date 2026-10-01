import { beforeAll, vi } from 'vitest'

// Stub the Wails runtime for every spec.
//
// There is no Wails host under Vitest, so any runtime call rejects — and
// because stores subscribe to events at init (ui.ts -> useWails/useEvents),
// that surfaced as an unhandled rejection that failed the whole run even
// with every test passing.
vi.mock('@wailsio/runtime', () => ({
  Events: {
    On: vi.fn(() => () => {}),
    Once: vi.fn(() => () => {}),
    Off: vi.fn(),
    Emit: vi.fn(async () => undefined),
  },
  Window: {
    Maximise: vi.fn(async () => undefined),
    UnMaximise: vi.fn(async () => undefined),
    IsMaximised: vi.fn(async () => false),
    Center: vi.fn(async () => undefined),
    Fullscreen: vi.fn(async () => undefined),
    UnFullscreen: vi.fn(async () => undefined),
    IsFullscreen: vi.fn(async () => false),
    SetFrameless: vi.fn(async () => undefined),
  },
  Application: { Quit: vi.fn(async () => undefined) },
}))

// Mock browser APIs
beforeAll(() => {
  // Observer mocks.
  //
  // These must be real constructors. `vi.fn().mockImplementation(() => ({}))`
  // returns a plain object and throws "is not a constructor" the moment
  // anything calls it with `new` — which VueUse's useResizeObserver does,
  // taking out every spec for a component that observes its own size.
  // Reports a fixed non-zero size to its callback. VueUse's useElementSize
  // (and therefore useVirtualList) derives its viewport from the observer
  // entry, not from the DOM — with no callback the viewport stays 0px and a
  // virtualised list renders no rows at all.
  class MockResizeObserver {
    private cb: ResizeObserverCallback
    constructor(cb: ResizeObserverCallback) { this.cb = cb }
    observe = vi.fn((target: Element) => {
      const box = { inlineSize: 800, blockSize: 400 }
      this.cb(
        [{
          target,
          contentRect: { width: 800, height: 400, top: 0, left: 0, bottom: 400, right: 800, x: 0, y: 0 },
          borderBoxSize: [box],
          contentBoxSize: [box],
          devicePixelContentBoxSize: [box],
        }] as unknown as ResizeObserverEntry[],
        this as unknown as ResizeObserver,
      )
    })
    unobserve = vi.fn()
    disconnect = vi.fn()
  }
  global.ResizeObserver = MockResizeObserver as unknown as typeof ResizeObserver

  class MockMutationObserver {
    observe = vi.fn()
    disconnect = vi.fn()
    takeRecords = vi.fn(() => [])
  }
  global.MutationObserver = MockMutationObserver as unknown as typeof MutationObserver

  class MockIntersectionObserver {
    observe = vi.fn()
    unobserve = vi.fn()
    disconnect = vi.fn()
    takeRecords = vi.fn(() => [])
    root = null
    rootMargin = ''
    thresholds: number[] = []
  }
  global.IntersectionObserver = MockIntersectionObserver as unknown as typeof IntersectionObserver

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

  // Mock fetch.
  //
  // This must resolve a Response-like object rather than `undefined`.
  // @wailsio/runtime calls `fetch(url).then(...)` at MODULE level (see
  // loadOptionalScript, which probes /wails/custom.js), so a bare vi.fn()
  // threw "Cannot read properties of undefined (reading 'then')" before any
  // test body ran — taking out every spec that imports the runtime, directly
  // or through a store.
  global.fetch = vi.fn(async () => ({
    ok: false,
    status: 404,
    headers: { get: () => null },
    json: async () => ({}),
    text: async () => '',
  })) as unknown as typeof fetch

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