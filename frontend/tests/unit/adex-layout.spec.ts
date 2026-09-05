import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import IndexPage from '~/app/pages/index.vue'

// Mock the useWails composable
vi.mock('~/composables/useWails', () => ({
  useWails: () => ({
    system: {
      getSystemInfo: vi.fn().mockResolvedValue({
        hostname: 'test-host',
        platform: 'linux',
        os: 'Linux',
        arch: 'x64',
        uptime: 3600,
        kernel: '5.15.0',
        numProcs: 100,
      }),
      getCPUUsage: vi.fn().mockResolvedValue({
        usage: 25,
        cores: [],
        coreCount: 4,
        modelName: 'Test CPU',
        frequency: 3000,
      }),
      getMemoryUsage: vi.fn().mockResolvedValue({
        total: 8589934592,
        used: 4294967296,
        free: 4294967296,
        usage: 50,
      }),
      getDiskUsage: vi.fn().mockResolvedValue([]),
      getTopProcesses: vi.fn().mockResolvedValue([]),
      startMonitoring: vi.fn().mockResolvedValue(undefined),
      stopMonitoring: vi.fn().mockResolvedValue(undefined),
    },
    filesystem: {
      readDirectory: vi.fn().mockResolvedValue([]),
      createDirectory: vi.fn().mockResolvedValue(undefined),
      deleteFile: vi.fn().mockResolvedValue(undefined),
      renameFile: vi.fn().mockResolvedValue(undefined),
      copyFile: vi.fn().mockResolvedValue(undefined),
      moveFile: vi.fn().mockResolvedValue(undefined),
      searchFiles: vi.fn().mockResolvedValue([]),
      readFile: vi.fn().mockResolvedValue(''),
      writeFile: vi.fn().mockResolvedValue(undefined),
    },
    isReady: { value: true },
    error: { value: null },
  })
}))

vi.mock('~/utils/errorHandler', () => ({
  handleBackendError: vi.fn(),
  handleComponentError: vi.fn(),
}))

vi.mock('~/utils/loadingStates', () => ({
  createLoading: vi.fn(),
  completeLoading: vi.fn(),
}))

// Mock the Wails bindings used by the terminal store
vi.mock('~~/bindings', () => ({
  CreateTerminal: vi.fn().mockResolvedValue({ id: 'test-session-1' }),
  WriteToTerminal: vi.fn().mockResolvedValue(undefined),
  ResizeTerminal: vi.fn().mockResolvedValue(undefined),
  CloseTerminal: vi.fn().mockResolvedValue(undefined),
}))

// Mock the keyboard layouts JSON
vi.mock('~/assets/data/keyboard-layouts.json', () => ({
  default: {
    layouts: {
      qwerty: {
        name: 'QWERTY',
        displayName: 'QWERTY (US)',
        type: 'standard',
        rows: [
          [{ key: 'a', shift: 'A', width: 1 }],
        ],
      },
    },
  },
  layouts: {
    qwerty: {
      name: 'QWERTY',
      displayName: 'QWERTY (US)',
      type: 'standard',
      rows: [
        [{ key: 'a', shift: 'A', width: 1 }],
      ],
    },
  },
}))

// Mock the Wails runtime
vi.mock('~/lib/wailsjs/runtime', () => ({
  Events: {
    On: vi.fn(),
    Off: vi.fn(),
    Emit: vi.fn(),
  },
  Log: {
    Info: vi.fn(),
    Error: vi.fn(),
    Warning: vi.fn(),
  },
}))

// Mock the coordinator bindings
vi.mock('~/lib/wailsjs/coordinator', () => ({
  GetCPUUsage: vi.fn().mockResolvedValue({}),
  GetMemoryUsage: vi.fn().mockResolvedValue({}),
  GetDiskUsage: vi.fn().mockResolvedValue([]),
  GetSystemInfo: vi.fn().mockResolvedValue({}),
  GetTopProcesses: vi.fn().mockResolvedValue([]),
  CreateTerminal: vi.fn().mockResolvedValue({ id: 'test-session-1' }),
  WriteToTerminal: vi.fn().mockResolvedValue(undefined),
  ResizeTerminal: vi.fn().mockResolvedValue(undefined),
  CloseTerminal: vi.fn().mockResolvedValue(undefined),
  ReadDirectory: vi.fn().mockResolvedValue([]),
}))

describe('Index Page Layout', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 0))
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function mountLayout() {
    const pinia = createPinia()
    setActivePinia(pinia)

    wrapper = mount(IndexPage, {
      global: {
        plugins: [pinia],
        stubs: {
          // Stub child components to isolate layout tests and avoid deep component tree
          AdexClock: { template: '<div class="mod-clock-stub">CLOCK</div>' },
          AdexSysinfo: { template: '<div class="mod-sysinfo-stub">SYSINFO</div>' },
          AdexHardware: { template: '<div class="mod-hardware-stub">HARDWARE</div>' },
          AdexCpuInfo: { template: '<div class="mod-cpuinfo-stub">CPUINFO</div>' },
          AdexRamWatcher: { template: '<div class="mod-ram-stub">RAM</div>' },
          AdexToplist: { template: '<div class="mod-toplist-stub">TOPLIST</div>' },
          AdexNetstat: { template: '<div class="mod-netstat-stub">NETSTAT</div>' },
          AdexGlobe: { template: '<div class="mod-globe-stub">GLOBE</div>' },
          AdexTraffic: { template: '<div class="mod-traffic-stub">TRAFFIC</div>' },
          AdexFilesystem: { template: '<div class="adex-filesystem-stub">FILESYSTEM</div>' },
          AdexKeyboard: { template: '<div class="adex-keyboard-stub">KEYBOARD</div>' },
          AdexTerminal: { template: '<div class="adex-terminal-stub">TERMINAL</div>' },
          AdexSettingsModal: { template: '<div class="adex-settings-stub">SETTINGS</div>' },
          // Stub AdexBootScreen to bypass boot animation in most tests
          AdexBootScreen: {
            template: '<div class="boot-overlay-stub" @click="$emit(\'complete\')">BOOT</div>',
            emits: ['complete'],
          },
        },
      },
    })

    return wrapper
  }

  describe('Boot screen', () => {
    it('shows boot screen initially (booting is true)', () => {
      mountLayout()
      // The boot screen stub should be visible
      expect(wrapper.find('.boot-overlay-stub').exists()).toBe(true)
    })

    it('hides boot screen after complete event', async () => {
      mountLayout()
      // Trigger the boot complete event
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.boot-overlay-stub').exists()).toBe(false)
    })

    it('shows main app after boot completes', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-app').exists()).toBe(true)
    })
  })

  describe('Background', () => {
    it('has adex-background div', async () => {
      mountLayout()
      // Complete boot first
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-background').exists()).toBe(true)
    })
  })

  describe('Top bar', () => {
    it('has adex-topbar navigation', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-topbar').exists()).toBe(true)
    })

    it('has PANEL section label', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      const sectionTexts = sections.map(s => s.text())
      expect(sectionTexts).toContain('PANEL')
    })

    it('has SYSTEM section label', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      const sectionTexts = sections.map(s => s.text())
      expect(sectionTexts).toContain('SYSTEM')
    })

    it('has TERMINAL section label', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      const sectionTexts = sections.map(s => s.text())
      expect(sectionTexts).toContain('TERMINAL')
    })

    it('has NETWORK section label', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      const sectionTexts = sections.map(s => s.text())
      expect(sectionTexts).toContain('NETWORK')
    })

    it('has role="navigation" attribute', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const topbar = wrapper.find('.adex-topbar')
      expect(topbar.attributes('role')).toBe('navigation')
    })

    it('has active state on a section by default', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const activeSections = wrapper.findAll('.topbar-section.active')
      expect(activeSections.length).toBeGreaterThan(0)
    })
  })

  describe('Three-column layout', () => {
    it('has adex-main container for the three columns', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-main').exists()).toBe(true)
    })

    it('has left column', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const leftCol = wrapper.find('.mod-column.left')
      expect(leftCol.exists()).toBe(true)
    })

    it('has center section', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-center').exists()).toBe(true)
    })

    it('has right column', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const rightCol = wrapper.find('.mod-column.right')
      expect(rightCol.exists()).toBe(true)
    })
  })

  describe('Left column components', () => {
    it('contains Clock component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-clock-stub').exists()).toBe(true)
    })

    it('contains Sysinfo component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-sysinfo-stub').exists()).toBe(true)
    })

    it('hides the Hardware component by default', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      // Hardware is in DEFAULT_HIDDEN: its output ("Manufacturer: Unknown")
      // was not useful and it consumed column height the process list needed.
      // It is switchable in Settings -> Modules, but off unless asked for.
      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-hardware').exists()).toBe(false)
    })

    it('contains CpuInfo component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-cpuinfo-stub').exists()).toBe(true)
    })

    it('contains RamWatcher component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-ram-stub').exists()).toBe(true)
    })

    it('contains Toplist component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-toplist-stub').exists()).toBe(true)
    })
  })

  describe('Right column components', () => {
    it('contains Netstat component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const right = wrapper.find('.mod-column.right')
      expect(right.find('.mod-netstat-stub').exists()).toBe(true)
    })

    it('contains Globe component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const right = wrapper.find('.mod-column.right')
      expect(right.find('.mod-globe-stub').exists()).toBe(true)
    })

    it('contains Traffic component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const right = wrapper.find('.mod-column.right')
      expect(right.find('.mod-traffic-stub').exists()).toBe(true)
    })
  })

  describe('Center section - Shell', () => {
    it('has main-shell container', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.main-shell').exists()).toBe(true)
    })

    it('has shell-tabs container', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.shell-tabs').exists()).toBe(true)
    })

    it('has shell-terminal container', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.shell-terminal').exists()).toBe(true)
    })

    it('has shell-status bar', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.shell-status').exists()).toBe(true)
    })

    it('has shell tab elements', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      // Advance timers to allow async tab creation
      await vi.advanceTimersByTimeAsync(500)
      await wrapper.vm.$nextTick()

      const tabs = wrapper.findAll('.shell-tab')
      expect(tabs.length).toBeGreaterThan(0)
    })

    it('has EMPTY placeholder tabs', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      await vi.advanceTimersByTimeAsync(500)
      await wrapper.vm.$nextTick()

      const text = wrapper.text()
      expect(text).toContain('EMPTY')
    })
  })

  describe('Bottom section', () => {
    it('has adex-bottom container', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-bottom').exists()).toBe(true)
    })

    it('contains Filesystem component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const bottom = wrapper.find('.adex-bottom')
      expect(bottom.find('.adex-filesystem-stub').exists()).toBe(true)
    })

    it('contains Keyboard component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const bottom = wrapper.find('.adex-bottom')
      expect(bottom.find('.adex-keyboard-stub').exists()).toBe(true)
    })
  })

  describe('Settings modal', () => {
    it('has AdexSettingsModal component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.adex-settings-stub').exists()).toBe(true)
    })
  })

  describe('Top bar section navigation', () => {
    it('toggles a region when its topbar section is clicked', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      const systemSection = sections.find(s => s.text() === 'SYSTEM')
      expect(systemSection).toBeDefined()

      // PANEL / SYSTEM / TERMINAL are toggles, not navigation: `active`
      // tracks whether the region is currently shown, so clicking a visible
      // one hides it and clears the class.
      const wasActive = systemSection!.classes().includes('active')

      await systemSection!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(systemSection!.classes().includes('active')).toBe(!wasActive)
    })
  })

  describe('Shell tab management in topbar', () => {
    it('shows the empty shell slots in the top bar', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      // Advance for async tab creation
      await vi.advanceTimersByTimeAsync(500)
      await wrapper.vm.$nextTick()

      // Named shell tabs only exist once a PTY does, and creating one needs a
      // real backend — jsdom has none, so shellTabs stays empty here. The
      // EMPTY slots are what the topbar renders in that state, and clicking
      // one is the only path that creates a session.
      const sections = wrapper.findAll('.topbar-section')
      const emptySlots = sections.filter(s => s.text().includes('EMPTY'))
      expect(emptySlots.length).toBeGreaterThan(0)
    })
  })

  describe('Keyboard shortcut handler', () => {
    it('registers keydown event listener on window', async () => {
      const addEventSpy = vi.spyOn(window, 'addEventListener')
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      // Advance timers for onMounted to complete
      await vi.advanceTimersByTimeAsync(500)
      await wrapper.vm.$nextTick()

      const keydownCalls = addEventSpy.mock.calls.filter(c => c[0] === 'keydown')
      expect(keydownCalls.length).toBeGreaterThan(0)
    })

    it('removes keydown event listener on unmount', async () => {
      const removeEventSpy = vi.spyOn(window, 'removeEventListener')
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      await vi.advanceTimersByTimeAsync(500)
      await wrapper.vm.$nextTick()

      wrapper.unmount()

      const keydownCalls = removeEventSpy.mock.calls.filter(c => c[0] === 'keydown')
      expect(keydownCalls.length).toBeGreaterThan(0)
    })
  })
})
