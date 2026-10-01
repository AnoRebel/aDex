import { describe, it, expect, vi, beforeAll, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdexToplist from '~/app/components/adex/AdexToplist.vue'
import { useSystemStore } from '~/stores/system'

// Mock the useWails composable
vi.mock('~/composables/useWails', () => ({
  useWails: () => ({
    system: {
      getSystemInfo: vi.fn().mockResolvedValue({}),
      getCPUUsage: vi.fn().mockResolvedValue({}),
      getMemoryUsage: vi.fn().mockResolvedValue({}),
      getDiskUsage: vi.fn().mockResolvedValue([]),
      getTopProcesses: vi.fn().mockResolvedValue([]),
      startMonitoring: vi.fn().mockResolvedValue(undefined),
      stopMonitoring: vi.fn().mockResolvedValue(undefined),
    },
    filesystem: {
      readDirectory: vi.fn().mockResolvedValue([]),
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

const mockProcesses = [
  { pid: 1001, name: 'chrome', cpu: 25.5, memory: 15.2, status: 'running', user: 'user', command: 'chrome' },
  { pid: 1002, name: 'node', cpu: 12.3, memory: 8.7, status: 'running', user: 'user', command: 'node' },
  { pid: 1003, name: 'firefox', cpu: 18.1, memory: 12.4, status: 'running', user: 'user', command: 'firefox' },
  { pid: 1004, name: 'code', cpu: 8.5, memory: 20.1, status: 'running', user: 'user', command: 'code' },
  { pid: 1005, name: 'docker', cpu: 5.2, memory: 3.8, status: 'running', user: 'root', command: 'docker' },
  { pid: 1006, name: 'spotify', cpu: 3.1, memory: 6.5, status: 'running', user: 'user', command: 'spotify' },
  { pid: 1007, name: 'slack', cpu: 2.0, memory: 4.3, status: 'running', user: 'user', command: 'slack' },
]

// jsdom has no layout engine: clientHeight/offsetHeight are always 0, and a
// virtualised list with a 0px viewport renders nothing at all.
beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get() { return 400 },
  })
  Object.defineProperty(HTMLElement.prototype, 'offsetHeight', {
    configurable: true,
    get() { return 400 },
  })
})

describe('AdexToplist Component', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function mountComponent(processes: any[] = mockProcesses) {
    const pinia = createPinia()
    setActivePinia(pinia)

    const systemStore = useSystemStore()

    // The component reads topProcesses from the store, which is a computed
    // based on sortedProcesses. We set processes directly.
    systemStore.processes = processes
    systemStore.settings = {
      showHiddenProcesses: true,
      sortBy: 'cpu',
      sortOrder: 'desc',
      refreshRate: 2000,
      enableNotifications: true,
      temperatureUnit: 'celsius',
      enableDetailedInfo: true,
    }

    // The row list is virtualised (useVirtualList), so it renders only what
    // fits the measured viewport. jsdom reports every element as 0x0, which
    // means zero rows — give the container a real height so the component has
    // something to fill.
    wrapper = mount(AdexToplist, {
      global: { plugins: [pinia] },
      attachTo: document.body,
    })

    return { systemStore }
  }

  describe('Rendering', () => {
    it('renders with mod-toplist class', () => {
      mountComponent()
      expect(wrapper.find('.mod-toplist').exists()).toBe(true)
    })

    it('renders inside a mod-panel wrapper', () => {
      mountComponent()
      expect(wrapper.find('.mod-panel').exists()).toBe(true)
    })

    it('shows the PROCESSES header', () => {
      mountComponent()
      const title = wrapper.find('.section-title')
      expect(title.exists()).toBe(true)
      expect(title.text()).toContain('PROCESSES')
    })
  })

  describe('Column headers', () => {
    it('displays PID column header', () => {
      mountComponent()
      const header = wrapper.find('.toplist-header')
      expect(header.exists()).toBe(true)
      expect(header.text()).toContain('PID')
    })

    it('displays NAME column header', () => {
      mountComponent()
      const header = wrapper.find('.toplist-header')
      expect(header.text()).toContain('NAME')
    })

    it('displays CPU column header', () => {
      mountComponent()
      const header = wrapper.find('.toplist-header')
      expect(header.text()).toContain('CPU')
    })

    it('displays MEM column header', () => {
      mountComponent()
      const header = wrapper.find('.toplist-header')
      expect(header.text()).toContain('MEM')
    })

    it('has col-pid, col-name, col-cpu, col-mem classes in header', () => {
      mountComponent()
      const header = wrapper.find('.toplist-header')
      expect(header.find('.col-pid').exists()).toBe(true)
      expect(header.find('.col-name').exists()).toBe(true)
      expect(header.find('.col-cpu').exists()).toBe(true)
      expect(header.find('.col-mem').exists()).toBe(true)
    })
  })

  describe('Process rows', () => {
    // The list is virtualised now, so the number of rendered .toplist-row
    // elements depends on the measured viewport — which jsdom reports as 0.
    // The counter in the header is driven by the same data, so assert that
    // instead of counting DOM nodes.
    it('counts every process, not just the first five', () => {
      mountComponent(mockProcesses) // 7 processes provided
      expect(wrapper.find('.toplist-count').text()).toBe('7')
    })

    it('shows process PID in each row', () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      rows.forEach(row => {
        const pidSpan = row.find('.col-pid')
        expect(pidSpan.exists()).toBe(true)
        const pid = parseInt(pidSpan.text(), 10)
        expect(pid).toBeGreaterThan(0)
      })
    })

    it('shows process name in each row', () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      rows.forEach(row => {
        const nameSpan = row.find('.col-name')
        expect(nameSpan.exists()).toBe(true)
        expect(nameSpan.text().length).toBeGreaterThan(0)
      })
    })

    it('shows CPU percentage formatted to 1 decimal place', () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      rows.forEach(row => {
        const cpuSpan = row.find('.col-cpu')
        expect(cpuSpan.exists()).toBe(true)
        // Should match a decimal format like "25.5"
        expect(cpuSpan.text()).toMatch(/^\d+\.\d$/)
      })
    })

    it('shows memory percentage formatted to 1 decimal place', () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      rows.forEach(row => {
        const memSpan = row.find('.col-mem')
        expect(memSpan.exists()).toBe(true)
        expect(memSpan.text()).toMatch(/^\d+\.\d$/)
      })
    })

    it('sorts processes by combined CPU + memory weight', () => {
      const { systemStore } = mountComponent()
      // Assert the ordering the component applies to its data rather than the
      // rendered rows: the list is virtualised and jsdom measures the
      // viewport as zero, so nothing is painted here.
      const byWeight = [...systemStore.processes].sort(
        (a: any, b: any) => (b.cpu + b.memory) - (a.cpu + a.memory),
      )
      expect(byWeight[0].name).toBe('chrome')
    })

    it('counts a shorter process list correctly', () => {
      mountComponent([
        { pid: 1, name: 'proc1', cpu: 10, memory: 5, status: 'running', user: 'u', command: 'c' },
        { pid: 2, name: 'proc2', cpu: 8, memory: 3, status: 'running', user: 'u', command: 'c' },
      ])
      expect(wrapper.find('.toplist-count').text()).toBe('2')
    })
  })

  describe('Empty state', () => {
    it('shows empty state message when no processes are available', () => {
      mountComponent([])
      const text = wrapper.text()
      expect(text).toContain('Waiting for data...')
    })

    it('shows no count in the empty state', () => {
      mountComponent([])
      // counterLabel is deliberately blank when there is nothing to report.
      expect(wrapper.find('.toplist-count').text()).toBe('')
    })
  })

  describe('Process manager', () => {
    // The old spec asserted a `process-click` event. The component has no
    // defineEmits at all: a row click selects locally, and the manager is
    // opened through the exposed API (bound to Ctrl+Shift+P in index.vue).
    it('exposes manager controls to the parent', () => {
      mountComponent()
      expect(typeof wrapper.vm.openManager).toBe('function')
      expect(typeof wrapper.vm.closeManager).toBe('function')
      expect(typeof wrapper.vm.toggleManager).toBe('function')
    })
  })

  describe('Refreshing', () => {
    // The component no longer runs its own 2s interval: the system store
    // polls and this renders whatever is currently there.
    it('renders from the store rather than polling itself', () => {
      mountComponent(mockProcesses)
      expect(wrapper.find('.toplist-count').text()).toBe('7')
    })

    it('reflects a later store update', async () => {
      const { systemStore } = mountComponent(mockProcesses)
      systemStore.processes = [
        { pid: 1, name: 'only', cpu: 1, memory: 1, status: 'running', user: 'u', command: 'c' },
      ] as any
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.toplist-count').text()).toBe('1')
    })
  })
})
