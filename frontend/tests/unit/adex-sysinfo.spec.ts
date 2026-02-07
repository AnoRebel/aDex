import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdexSysinfo from '~/app/components/adex/AdexSysinfo.vue'
import { useSystemStore } from '~/stores/system'

// Mock the useWails composable used by the system store
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

// Mock the error handler utilities
vi.mock('~/utils/errorHandler', () => ({
  handleBackendError: vi.fn(),
  handleComponentError: vi.fn(),
}))

// Mock the loading states utilities
vi.mock('~/utils/loadingStates', () => ({
  createLoading: vi.fn(),
  completeLoading: vi.fn(),
}))

describe('AdexSysinfo Component', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date(2026, 1, 5, 14, 30, 0))
    setActivePinia(createPinia())
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function mountComponent() {
    // We need to pre-populate the system store before mounting the component
    const pinia = createPinia()
    setActivePinia(pinia)

    // Import the store after pinia is set
    const systemStore = useSystemStore()

    // Set test data on the store
    systemStore.systemInfo = {
      hostname: 'test-host',
      platform: 'linux',
      os: 'Linux',
      arch: 'x64',
      architecture: 'x86_64',
      uptime: 90061, // 1 day, 1 hour, 1 minute, 1 second
      kernel: '5.15.0',
      kernelVersion: '5.15.0-generic',
      numProcs: 250,
    }

    wrapper = mount(AdexSysinfo, {
      global: {
        plugins: [pinia],
      }
    })

    return { systemStore }
  }

  describe('Rendering', () => {
    it('renders with mod-sysinfo class', () => {
      mountComponent()
      expect(wrapper.find('.mod-sysinfo').exists()).toBe(true)
    })

    it('renders inside a mod-panel wrapper', () => {
      mountComponent()
      expect(wrapper.find('.mod-panel').exists()).toBe(true)
    })

    it('shows SYSTEM INFO section title', () => {
      mountComponent()
      expect(wrapper.find('.section-title').text()).toContain('SYSTEM INFO')
    })
  })

  describe('Information rows', () => {
    it('displays YEAR row', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      const labels = rows.map(r => r.find('.sysinfo-label').text())
      expect(labels).toContain('YEAR')
    })

    it('displays DATE row', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      const labels = rows.map(r => r.find('.sysinfo-label').text())
      expect(labels).toContain('DATE')
    })

    it('displays UPTIME row', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      const labels = rows.map(r => r.find('.sysinfo-label').text())
      expect(labels).toContain('UPTIME')
    })

    it('displays OS row', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      const labels = rows.map(r => r.find('.sysinfo-label').text())
      expect(labels).toContain('OS')
    })

    it('displays POWER row', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      const labels = rows.map(r => r.find('.sysinfo-label').text())
      expect(labels).toContain('POWER')
    })

    it('has exactly 5 sysinfo rows', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      expect(rows.length).toBe(5)
    })
  })

  describe('CSS classes', () => {
    it('has sysinfo-row class on each row', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      expect(rows.length).toBeGreaterThan(0)
      rows.forEach(row => {
        expect(row.classes()).toContain('sysinfo-row')
      })
    })

    it('has sysinfo-label class on each label', () => {
      mountComponent()
      const labels = wrapper.findAll('.sysinfo-label')
      expect(labels.length).toBe(5)
    })

    it('has sysinfo-value class on each value', () => {
      mountComponent()
      const values = wrapper.findAll('.sysinfo-value')
      expect(values.length).toBe(5)
    })
  })

  describe('Data display', () => {
    it('shows the current year', async () => {
      mountComponent()
      await wrapper.vm.$nextTick()
      const yearRow = wrapper.findAll('.sysinfo-row')[0]
      expect(yearRow.find('.sysinfo-value').text()).toBe('2026')
    })

    it('shows the current date in MM/DD format', async () => {
      mountComponent()
      await wrapper.vm.$nextTick()
      const dateRow = wrapper.findAll('.sysinfo-row')[1]
      // February 5 => 02/05
      expect(dateRow.find('.sysinfo-value').text()).toBe('02/05')
    })

    it('shows uptime in DD:HH:MM format', () => {
      mountComponent()
      const uptimeRow = wrapper.findAll('.sysinfo-row')[2]
      // 90061 seconds = 1 day, 1 hour, 1 minute, 1 second
      // days=1, hours=1, minutes=1 => "01:01:01"
      expect(uptimeRow.find('.sysinfo-value').text()).toBe('01:01:01')
    })

    it('shows OS information from store', () => {
      mountComponent()
      const osRow = wrapper.findAll('.sysinfo-row')[3]
      const osText = osRow.find('.sysinfo-value').text()
      // The component builds OS string from os/platform + arch + kernel
      expect(osText).toContain('Linux')
    })

    it('shows POWER as AC Power by default', () => {
      mountComponent()
      const powerRow = wrapper.findAll('.sysinfo-row')[4]
      expect(powerRow.find('.sysinfo-value').text()).toBe('AC Power')
    })
  })

  describe('Uptime with no system info', () => {
    it('shows 00:00:00 when no systemInfo is available', () => {
      const pinia = createPinia()
      setActivePinia(pinia)
      const store = useSystemStore()
      store.systemInfo = null

      wrapper = mount(AdexSysinfo, {
        global: { plugins: [pinia] }
      })

      const uptimeRow = wrapper.findAll('.sysinfo-row')[2]
      expect(uptimeRow.find('.sysinfo-value').text()).toBe('00:00:00')
    })
  })

  describe('OS display with missing fields', () => {
    it('shows Unknown when no systemInfo is available', () => {
      const pinia = createPinia()
      setActivePinia(pinia)
      const store = useSystemStore()
      store.systemInfo = null

      wrapper = mount(AdexSysinfo, {
        global: { plugins: [pinia] }
      })

      const osRow = wrapper.findAll('.sysinfo-row')[3]
      expect(osRow.find('.sysinfo-value').text()).toBe('Unknown')
    })
  })

  describe('Timers', () => {
    it('sets up timers on mount', () => {
      const setIntervalSpy = vi.spyOn(global, 'setInterval')
      const setTimeoutSpy = vi.spyOn(global, 'setTimeout')
      mountComponent()
      // The component sets a setTimeout for midnight refresh and a setInterval for uptime
      expect(setTimeoutSpy).toHaveBeenCalled()
      expect(setIntervalSpy).toHaveBeenCalled()
    })

    it('clears timers on unmount', () => {
      const clearIntervalSpy = vi.spyOn(global, 'clearInterval')
      mountComponent()
      wrapper.unmount()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })

    it('increments local uptime offset every 60 seconds', async () => {
      const { systemStore } = mountComponent()
      // Set a base uptime
      systemStore.systemInfo = {
        ...systemStore.systemInfo!,
        uptime: 0, // 0 seconds
      }
      await wrapper.vm.$nextTick()

      const uptimeRow = wrapper.findAll('.sysinfo-row')[2]
      expect(uptimeRow.find('.sysinfo-value').text()).toBe('00:00:00')

      // Advance by 60 seconds
      vi.advanceTimersByTime(60000)
      await wrapper.vm.$nextTick()

      // Now local offset should add 60 seconds => 0 + 60 = 60 seconds => 00:00:01
      expect(uptimeRow.find('.sysinfo-value').text()).toBe('00:00:01')
    })
  })
})
