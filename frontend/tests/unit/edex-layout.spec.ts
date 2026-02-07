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
          EdexClock: { template: '<div class="mod-clock-stub">CLOCK</div>' },
          EdexSysinfo: { template: '<div class="mod-sysinfo-stub">SYSINFO</div>' },
          EdexHardware: { template: '<div class="mod-hardware-stub">HARDWARE</div>' },
          EdexCpuInfo: { template: '<div class="mod-cpuinfo-stub">CPUINFO</div>' },
          EdexRamWatcher: { template: '<div class="mod-ram-stub">RAM</div>' },
          EdexToplist: { template: '<div class="mod-toplist-stub">TOPLIST</div>' },
          EdexNetstat: { template: '<div class="mod-netstat-stub">NETSTAT</div>' },
          EdexGlobe: { template: '<div class="mod-globe-stub">GLOBE</div>' },
          EdexTraffic: { template: '<div class="mod-traffic-stub">TRAFFIC</div>' },
          EdexFilesystem: { template: '<div class="edex-filesystem-stub">FILESYSTEM</div>' },
          EdexKeyboard: { template: '<div class="edex-keyboard-stub">KEYBOARD</div>' },
          EdexTerminal: { template: '<div class="edex-terminal-stub">TERMINAL</div>' },
          EdexSettingsModal: { template: '<div class="edex-settings-stub">SETTINGS</div>' },
          // Stub EdexBootScreen to bypass boot animation in most tests
          EdexBootScreen: {
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

      expect(wrapper.find('.edex-app').exists()).toBe(true)
    })
  })

  describe('Background', () => {
    it('has edex-background div', async () => {
      mountLayout()
      // Complete boot first
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.edex-background').exists()).toBe(true)
    })
  })

  describe('Top bar', () => {
    it('has edex-topbar navigation', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.edex-topbar').exists()).toBe(true)
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

      const topbar = wrapper.find('.edex-topbar')
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
    it('has edex-main container for the three columns', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.edex-main').exists()).toBe(true)
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

      expect(wrapper.find('.edex-center').exists()).toBe(true)
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

    it('contains Hardware component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const left = wrapper.find('.mod-column.left')
      expect(left.find('.mod-hardware-stub').exists()).toBe(true)
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
    it('has edex-bottom container', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.edex-bottom').exists()).toBe(true)
    })

    it('contains Filesystem component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const bottom = wrapper.find('.edex-bottom')
      expect(bottom.find('.edex-filesystem-stub').exists()).toBe(true)
    })

    it('contains Keyboard component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const bottom = wrapper.find('.edex-bottom')
      expect(bottom.find('.edex-keyboard-stub').exists()).toBe(true)
    })
  })

  describe('Settings modal', () => {
    it('has EdexSettingsModal component', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.edex-settings-stub').exists()).toBe(true)
    })
  })

  describe('Top bar section navigation', () => {
    it('switches active section when a topbar section is clicked', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      // Find the SYSTEM section and click it
      const systemSection = sections.find(s => s.text() === 'SYSTEM')
      expect(systemSection).toBeDefined()

      await systemSection!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(systemSection!.classes()).toContain('active')
    })
  })

  describe('Shell tab management in topbar', () => {
    it('shows shell tab labels in the top bar', async () => {
      mountLayout()
      const bootScreen = wrapper.find('.boot-overlay-stub')
      await bootScreen.trigger('click')
      await wrapper.vm.$nextTick()

      // Advance for async tab creation
      await vi.advanceTimersByTimeAsync(500)
      await wrapper.vm.$nextTick()

      const sections = wrapper.findAll('.topbar-section')
      const shellSections = sections.filter(s =>
        s.text().includes('MAIN SHELL') || s.text().includes('SHELL')
      )
      // At least one shell tab should appear
      expect(shellSections.length).toBeGreaterThan(0)
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
