import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import EdexRamWatcher from '~/app/components/edex/EdexRamWatcher.vue'
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

describe('EdexRamWatcher Component', () => {
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

  function mountComponent(memoryOverrides?: {
    total?: number
    used?: number
    available?: number
    free?: number
    swapTotal?: number
    swapUsed?: number
  }) {
    const pinia = createPinia()
    setActivePinia(pinia)

    const systemStore = useSystemStore()

    // Set memory stats via systemStats (the fallback path used by the component)
    const total = memoryOverrides?.total ?? 8589934592      // 8 GiB
    const used = memoryOverrides?.used ?? 4294967296         // 4 GiB
    const free = memoryOverrides?.free ?? 2147483648         // 2 GiB
    const available = memoryOverrides?.available ?? (total - used)

    systemStore.systemStats = {
      cpu: { usage: 50, cores: [], coreCount: 4, modelName: 'Test CPU', frequency: 3000 },
      memory: { total, used, free, usage: (used / total) * 100 },
      disk: [],
    }

    // Set metrics for the primary path (memoryMetrics computed)
    systemStore.systemMetrics = {
      cpu: { usagePercent: 50, cores: 4, model: 'Test CPU', vendor: 'Test', frequency: 3000, frequencyMax: 4000, perCoreUsage: [], timestamp: '' },
      memory: {
        total,
        available,
        used,
        free,
        usagePercent: (used / total) * 100,
        swapTotal: memoryOverrides?.swapTotal ?? 4294967296,
        swapUsed: memoryOverrides?.swapUsed ?? 1073741824,
        swapFree: 3221225472,
        swapPercent: memoryOverrides?.swapTotal != null
          ? (memoryOverrides.swapTotal > 0
              ? ((memoryOverrides?.swapUsed ?? 0) / memoryOverrides.swapTotal) * 100
              : 0)
          : 25,
        timestamp: '',
      },
      processes: { processes: [], totalProcesses: 0, runningProcesses: 0, sleepingProcesses: 0, timestamp: '' },
      disks: { disks: [], totalSpace: 0, totalUsed: 0, totalFree: 0, timestamp: '' },
      network: { interfaces: [], totalBytesSent: 0, totalBytesRecv: 0, timestamp: '' },
      temperature: null,
      timestamp: '',
    }

    wrapper = mount(EdexRamWatcher, {
      global: { plugins: [pinia] }
    })

    return { systemStore }
  }

  describe('Rendering', () => {
    it('renders with mod-ram class', () => {
      mountComponent()
      expect(wrapper.find('.mod-ram').exists()).toBe(true)
    })

    it('renders inside a mod-panel wrapper', () => {
      mountComponent()
      expect(wrapper.find('.mod-panel').exists()).toBe(true)
    })

    it('shows MEMORY header in the section title', () => {
      mountComponent()
      const title = wrapper.find('.section-title')
      expect(title.exists()).toBe(true)
      expect(title.text()).toContain('MEMORY')
    })
  })

  describe('RAM grid', () => {
    it('has exactly 440 ram-cell divs (40x11 grid)', () => {
      mountComponent()
      const cells = wrapper.findAll('.ram-cell')
      expect(cells.length).toBe(440)
    })

    it('contains a ram-grid container', () => {
      mountComponent()
      expect(wrapper.find('.ram-grid').exists()).toBe(true)
    })

    it('all cells have one of the state classes: used, available, or free', async () => {
      mountComponent()
      await wrapper.vm.$nextTick()
      const cells = wrapper.findAll('.ram-cell')
      cells.forEach(cell => {
        const classes = cell.classes()
        const hasState = classes.includes('used') || classes.includes('available') || classes.includes('free')
        expect(hasState).toBe(true)
      })
    })

    it('distributes cells based on memory usage proportions', async () => {
      // 50% used, 50% available (total - used)
      mountComponent({
        total: 8589934592,
        used: 4294967296,
        available: 4294967296,
        free: 0,
      })
      await wrapper.vm.$nextTick()

      const cells = wrapper.findAll('.ram-cell')
      const usedCount = cells.filter(c => c.classes().includes('used')).length
      const availableCount = cells.filter(c => c.classes().includes('available')).length
      const freeCount = cells.filter(c => c.classes().includes('free')).length

      // With 50% used and 50% available, we expect approximately:
      // usedCells ~= 220, availableCells ~= 220, freeCells ~= 0
      expect(usedCount).toBeGreaterThan(180)
      expect(usedCount).toBeLessThan(260)
      expect(availableCount + freeCount).toBe(440 - usedCount)
    })

    it('shows mostly used cells when memory is nearly full', async () => {
      mountComponent({
        total: 8589934592,
        used: 8000000000,
        available: 589934592,
        free: 0,
      })
      await wrapper.vm.$nextTick()

      const cells = wrapper.findAll('.ram-cell')
      const usedCount = cells.filter(c => c.classes().includes('used')).length
      // With ~93% usage, most cells should be "used"
      expect(usedCount).toBeGreaterThan(350)
    })

    it('shows mostly free/available cells when memory usage is low', async () => {
      mountComponent({
        total: 8589934592,
        used: 858993459, // ~10%
        available: 7730941133,
        free: 7000000000,
      })
      await wrapper.vm.$nextTick()

      const cells = wrapper.findAll('.ram-cell')
      const usedCount = cells.filter(c => c.classes().includes('used')).length
      // With ~10% usage, only about 44 cells should be "used"
      expect(usedCount).toBeLessThan(100)
    })
  })

  describe('Memory usage display', () => {
    it('shows used GiB in the header', () => {
      mountComponent({
        total: 8589934592,
        used: 4294967296,
      })
      const titleRight = wrapper.find('.title-right')
      expect(titleRight.exists()).toBe(true)
      // 4294967296 / (1024^3) = 4.0 GiB
      expect(titleRight.text()).toContain('4.0')
    })

    it('shows total GiB in the header', () => {
      mountComponent({
        total: 8589934592,
        used: 4294967296,
      })
      const titleRight = wrapper.find('.title-right')
      // 8589934592 / (1024^3) = 8.0 GiB
      expect(titleRight.text()).toContain('8.0')
      expect(titleRight.text()).toContain('GiB')
    })

    it('shows USING X OUT OF Y GiB format', () => {
      mountComponent()
      const titleRight = wrapper.find('.title-right')
      expect(titleRight.text()).toMatch(/USING .+ OUT OF .+ GiB/)
    })
  })

  describe('Swap bar', () => {
    it('renders a swap bar section', () => {
      mountComponent()
      expect(wrapper.find('.swap-bar').exists()).toBe(true)
    })

    it('shows SWAP label', () => {
      mountComponent()
      const swapBar = wrapper.find('.swap-bar')
      expect(swapBar.text()).toContain('SWAP')
    })

    it('has a bar track and bar fill', () => {
      mountComponent()
      expect(wrapper.find('.bar-track').exists()).toBe(true)
      expect(wrapper.find('.bar-fill').exists()).toBe(true)
    })

    it('shows swap percentage', () => {
      mountComponent()
      const swapBar = wrapper.find('.swap-bar')
      // Default swap is 25%
      expect(swapBar.text()).toContain('25')
      expect(swapBar.text()).toContain('%')
    })

    it('swap bar width reflects swap percentage', () => {
      mountComponent()
      const barFill = wrapper.find('.bar-fill')
      const style = barFill.attributes('style')
      // Should have width: 25% (from the mock swap percent)
      expect(style).toContain('width:')
      expect(style).toContain('25%')
    })

    it('shows 0% swap when no swap is configured', () => {
      mountComponent({
        total: 8589934592,
        used: 4294967296,
        swapTotal: 0,
        swapUsed: 0,
      })
      const swapBar = wrapper.find('.swap-bar')
      expect(swapBar.text()).toContain('0%')
    })
  })

  describe('Timer management', () => {
    it('creates a refresh timer on mount', () => {
      const setIntervalSpy = vi.spyOn(global, 'setInterval')
      mountComponent()
      // The component sets an interval for re-shuffling cells (1500ms)
      expect(setIntervalSpy).toHaveBeenCalledWith(expect.any(Function), 1500)
    })

    it('clears the refresh timer on unmount', () => {
      const clearIntervalSpy = vi.spyOn(global, 'clearInterval')
      mountComponent()
      wrapper.unmount()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })

    it('re-shuffles cells on timer fire', async () => {
      mountComponent()
      await wrapper.vm.$nextTick()

      // Get initial cell class distribution
      const getCellClasses = () => wrapper.findAll('.ram-cell').map(c => c.classes().filter(cls => cls !== 'ram-cell').join(','))
      const initial = getCellClasses()

      // Advance the timer so cells re-shuffle
      await vi.advanceTimersByTimeAsync(1500)
      await wrapper.vm.$nextTick()

      const afterShuffle = getCellClasses()

      // The total counts should be the same but the distribution order may differ
      // (Fisher-Yates shuffle randomizes positions)
      const countClass = (arr: string[], cls: string) => arr.filter(c => c.includes(cls)).length
      expect(countClass(initial, 'used')).toBeGreaterThan(0)
      expect(countClass(initial, 'used')).toBe(countClass(afterShuffle, 'used'))
      expect(countClass(initial, 'available')).toBe(countClass(afterShuffle, 'available'))
      expect(countClass(initial, 'free')).toBe(countClass(afterShuffle, 'free'))
    })
  })
})
