import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
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

    wrapper = mount(AdexToplist, {
      global: { plugins: [pinia] }
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

    it('shows TOP PROCESSES header', () => {
      mountComponent()
      const title = wrapper.find('.section-title')
      expect(title.exists()).toBe(true)
      expect(title.text()).toContain('TOP PROCESSES')
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
    it('displays up to 5 processes', () => {
      mountComponent(mockProcesses) // 7 processes provided
      const rows = wrapper.findAll('.toplist-row')
      expect(rows.length).toBe(5)
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
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      // The component sorts by (cpu + memory) desc, takes top 5
      // Expected order by (cpu + memory):
      // chrome: 25.5+15.2=40.7, firefox: 18.1+12.4=30.5, code: 8.5+20.1=28.6,
      // node: 12.3+8.7=21.0, spotify: 3.1+6.5=9.6 (or docker: 5.2+3.8=9.0)
      const firstRowName = rows[0].find('.col-name').text()
      expect(firstRowName).toBe('chrome')
    })

    it('displays fewer rows when fewer processes are available', () => {
      mountComponent([
        { pid: 1, name: 'proc1', cpu: 10, memory: 5, status: 'running', user: 'u', command: 'c' },
        { pid: 2, name: 'proc2', cpu: 8, memory: 3, status: 'running', user: 'u', command: 'c' },
      ])
      const rows = wrapper.findAll('.toplist-row')
      expect(rows.length).toBe(2)
    })
  })

  describe('Empty state', () => {
    it('shows empty state message when no processes are available', () => {
      mountComponent([])
      const text = wrapper.text()
      expect(text).toContain('Waiting for data...')
    })

    it('shows exactly one toplist-row in empty state', () => {
      mountComponent([])
      const rows = wrapper.findAll('.toplist-row')
      expect(rows.length).toBe(1) // The "waiting for data" row
    })
  })

  describe('Events', () => {
    it('emits process-click when a process row is clicked', async () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      await rows[0].trigger('click')

      const emitted = wrapper.emitted('process-click')
      expect(emitted).toBeTruthy()
      expect(emitted!.length).toBe(1)
    })

    it('emits process data with the click event', async () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      await rows[0].trigger('click')

      const emitted = wrapper.emitted('process-click')
      expect(emitted).toBeTruthy()
      const payload = emitted![0][0] as any
      expect(payload).toHaveProperty('pid')
      expect(payload).toHaveProperty('name')
      expect(payload).toHaveProperty('cpu')
      expect(payload).toHaveProperty('memory')
    })

    it('emits the correct process for the clicked row', async () => {
      mountComponent()
      const rows = wrapper.findAll('.toplist-row')
      // Click the second row
      await rows[1].trigger('click')

      const emitted = wrapper.emitted('process-click')
      const payload = emitted![0][0] as any
      // Second process in sorted order (by cpu+mem desc) should be firefox
      expect(payload.name).toBe('firefox')
    })
  })

  describe('Timer management', () => {
    it('creates a refresh timer on mount', () => {
      const setIntervalSpy = vi.spyOn(global, 'setInterval')
      mountComponent()
      expect(setIntervalSpy).toHaveBeenCalledWith(expect.any(Function), 2000)
    })

    it('clears the refresh timer on unmount', () => {
      const clearIntervalSpy = vi.spyOn(global, 'clearInterval')
      mountComponent()
      wrapper.unmount()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })
  })
})
