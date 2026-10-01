import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdexSysinfo from '~/app/components/adex/AdexSysinfo.vue'
import { useSystemStore } from '~/stores/system'

/** Find a sysinfo row by its label, rather than by index — adding a row
 *  shifts every positional lookup and was why these broke. */
function rowByLabel(w: any, label: string) {
  return w.findAll('.sysinfo-row').find(
    (r: any) => r.find('.sysinfo-label').text() === label,
  )
}


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

    // YEAR, DATE, UPTIME, OS, KERNEL, POWER. KERNEL and POWER render
    // conditionally, so this asserts the full set the mock data produces.
    it('has exactly 6 sysinfo rows', () => {
      mountComponent()
      const rows = wrapper.findAll('.sysinfo-row')
      expect(rows.length).toBe(6)
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
      expect(labels.length).toBe(6)
    })

    it('has sysinfo-value class on each value', () => {
      mountComponent()
      const values = wrapper.findAll('.sysinfo-value')
      expect(values.length).toBe(6)
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
      const uptimeRow = rowByLabel(wrapper, 'UPTIME')
      // 90061 seconds = 1 day, 1 hour, 1 minute, 1 second
      // The display carries explicit unit labels — plain 01:01:01 was
      // ambiguous between H:M:S and D:H:M — and only shows the day field
      // once uptime exceeds a day.
      expect(uptimeRow.find('.sysinfo-value').text()).toBe('1d 01h 01m 01s')
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
      const powerRow = rowByLabel(wrapper, 'POWER')
      expect(powerRow.find('.sysinfo-value').text()).toBe('AC Power')
    })
  })

  describe('Uptime with no system info', () => {
    it("shows '--' when no systemInfo is available", () => {
      const pinia = createPinia()
      setActivePinia(pinia)
      const store = useSystemStore()
      store.systemInfo = null

      wrapper = mount(AdexSysinfo, {
        global: { plugins: [pinia] }
      })

      const uptimeRow = rowByLabel(wrapper, 'UPTIME')
      expect(uptimeRow.find('.sysinfo-value').text()).toBe('--')
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
    // The component no longer schedules its own uptime tick or midnight
    // refresh: YEAR / DATE / UPTIME all derive from VueUse's `useNow`, and
    // only the power source is polled directly. These tests assert the
    // behaviour (a live-updating clock) rather than which timer API backs it.
    it('polls the power source on an interval', () => {
      const setIntervalSpy = vi.spyOn(global, 'setInterval')
      mountComponent()
      expect(setIntervalSpy).toHaveBeenCalled()
    })

    it('clears timers on unmount', () => {
      const clearIntervalSpy = vi.spyOn(global, 'clearInterval')
      mountComponent()
      wrapper.unmount()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })

    it('advances the uptime display as time passes', async () => {
      const { systemStore } = mountComponent()
      // A real boot time an hour ago, so uptime is derived rather than
      // accumulated from a local offset.
      systemStore.systemInfo = {
        ...systemStore.systemInfo!,
        uptime: 3600,
      }
      await wrapper.vm.$nextTick()

      const uptimeRow = rowByLabel(wrapper, 'UPTIME')
      const before = uptimeRow.find('.sysinfo-value').text()
      expect(before).not.toBe('--')

      // useNow ticks every second; advancing the clock must move the display.
      vi.advanceTimersByTime(60_000)
      await wrapper.vm.$nextTick()

      expect(uptimeRow.find('.sysinfo-value').text()).not.toBe(before)
    })
  })
})
