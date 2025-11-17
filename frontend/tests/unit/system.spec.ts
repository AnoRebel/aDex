import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import SystemMonitor from '~/app/components/system/SystemMonitor.vue'
import CpuChart from '~/app/components/system/CpuChart.vue'
import MemoryChart from '~/app/components/system/MemoryChart.vue'
import ProcessList from '~/app/components/system/ProcessList.vue'
import { useSystemStore } from '~/stores/system'
import type { SystemMetrics, CPUMetrics, MemoryMetrics, ProcessMetrics, DiskMetrics, NetworkMetrics } from '~/types/system'

// Mock the system composable
vi.mock('~/composables/useSystem', () => ({
  useSystem: () => ({
    systemMetrics: ref(mockSystemMetrics),
    isLoading: ref(false),
    error: ref(null),
    refreshInterval: ref(1000),
    lastUpdated: ref(new Date()),
    refreshMetrics: vi.fn(),
    startMonitoring: vi.fn(),
    stopMonitoring: vi.fn(),
    isMonitoring: ref(true)
  })
}))

// Mock data
const mockCPUMetrics: CPUMetrics = {
  usagePercent: 45.2,
  cores: 8,
  model: 'Intel Core i7-9700K',
  vendor: 'Intel',
  frequency: 3600,
  frequencyMax: 4900,
  perCoreUsage: [42.1, 45.8, 48.2, 41.5, 44.7, 46.3, 43.9, 47.1],
  timestamp: new Date().toISOString()
}

const mockMemoryMetrics: MemoryMetrics = {
  total: 16777216000, // 16GB
  available: 8388608000,  // 8GB
  used: 8388608000,      // 8GB
  free: 8388608000,      // 8GB
  usagePercent: 50.0,
  swapTotal: 4294967296, // 4GB
  swapUsed: 1073741824,  // 1GB
  swapFree: 3221225472,  // 3GB
  timestamp: new Date().toISOString()
}

const mockProcessMetrics: ProcessMetrics = {
  processes: [
    {
      pid: 1234,
      name: 'chrome',
      cpuPercent: 25.5,
      memoryPercent: 15.2,
      memoryRSS: 1073741824,
      memoryVMS: 2147483648,
      status: 'Running',
      user: 'user',
      command: '/opt/google/chrome/chrome'
    },
    {
      pid: 5678,
      name: 'node',
      cpuPercent: 12.3,
      memoryPercent: 8.7,
      memoryRSS: 614400000,
      memoryVMS: 1228800000,
      status: 'Running',
      user: 'user',
      command: 'node /path/to/app'
    }
  ],
  totalProcesses: 245,
  runningProcesses: 3,
  sleepingProcesses: 242,
  timestamp: new Date().toISOString()
}

const mockDiskMetrics: DiskMetrics = {
  disks: [
    {
      device: '/dev/sda1',
      mountpoint: '/',
      total: 500000000000,  // 500GB
      used: 250000000000,   // 250GB
      free: 250000000000,   // 250GB
      usagePercent: 50.0,
      fsType: 'ext4'
    }
  ],
  totalSpace: 500000000000,
  totalUsed: 250000000000,
  totalFree: 250000000000,
  timestamp: new Date().toISOString()
}

const mockNetworkMetrics: NetworkMetrics = {
  interfaces: [
    {
      name: 'eth0',
      bytesSent: 1000000000,
      bytesRecv: 2000000000,
      packetsSent: 1000000,
      packetsRecv: 2000000,
      errin: 0,
      errout: 0,
      dropin: 0,
      dropout: 0,
      ipAddresses: ['192.168.1.100'],
      isUp: true
    }
  ],
  totalBytesSent: 1000000000,
  totalBytesRecv: 2000000000,
  timestamp: new Date().toISOString()
}

const mockSystemMetrics: SystemMetrics = {
  cpu: mockCPUMetrics,
  memory: mockMemoryMetrics,
  processes: mockProcessMetrics,
  disks: mockDiskMetrics,
  network: mockNetworkMetrics,
  temperature: null,
  timestamp: new Date().toISOString()
}

describe('SystemMonitor Component', () => {
  let wrapper: any
  let systemStore: any

  beforeEach(() => {
    setActivePinia(createPinia())
    systemStore = useSystemStore()

    wrapper = mount(SystemMonitor, {
      global: {
        components: {
          CpuChart,
          MemoryChart,
          ProcessList
        },
        stubs: {
          CpuChart: true,
          MemoryChart: true,
          ProcessList: true
        }
      }
    })
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.clearAllMocks()
  })

  describe('Component Rendering', () => {
    it('renders system monitor container', () => {
      expect(wrapper.find('.system-monitor').exists()).toBe(true)
    })

    it('renders monitoring status indicator', () => {
      expect(wrapper.find('.monitoring-status').exists()).toBe(true)
      expect(wrapper.find('.status-indicator').exists()).toBe(true)
    })

    it('displays system metrics grid', () => {
      expect(wrapper.find('.metrics-grid').exists()).toBe(true)
    })

    it('shows refresh button', () => {
      expect(wrapper.find('.refresh-button').exists()).toBe(true)
    })

    it('displays last updated timestamp', () => {
      expect(wrapper.find('.last-updated').exists()).toBe(true)
    })
  })

  describe('CPU Metrics Display', () => {
    it('displays CPU usage percentage', () => {
      expect(wrapper.find('.cpu-usage').exists()).toBe(true)
      expect(wrapper.text()).toContain('45.2%')
    })

    it('shows CPU model and cores', () => {
      expect(wrapper.text()).toContain('Intel Core i7-9700K')
      expect(wrapper.text()).toContain('8 Cores')
    })

    it('displays CPU frequency', () => {
      expect(wrapper.text()).toContain('3.60 GHz')
    })

    it('shows per-core usage chart', () => {
      expect(wrapper.findComponent(CpuChart).exists()).toBe(true)
    })
  })

  describe('Memory Metrics Display', () => {
    it('displays memory usage percentage', () => {
      expect(wrapper.find('.memory-usage').exists()).toBe(true)
      expect(wrapper.text()).toContain('50.0%')
    })

    it('shows memory usage in human readable format', () => {
      expect(wrapper.text()).toContain('8.0 GB')
      expect(wrapper.text()).toContain('16.0 GB')
    })

    it('displays swap usage if available', () => {
      expect(wrapper.text()).toContain('Swap')
      expect(wrapper.text()).toContain('1.0 GB')
    })

    it('shows memory chart component', () => {
      expect(wrapper.findComponent(MemoryChart).exists()).toBe(true)
    })
  })

  describe('Process Metrics Display', () => {
    it('displays total process count', () => {
      expect(wrapper.text()).toContain('245')
      expect(wrapper.text()).toContain('Total Processes')
    })

    it('shows running and sleeping process counts', () => {
      expect(wrapper.text()).toContain('3 Running')
      expect(wrapper.text()).toContain('242 Sleeping')
    })

    it('renders process list component', () => {
      expect(wrapper.findComponent(ProcessList).exists()).toBe(true)
    })
  })

  describe('Disk Metrics Display', () => {
    it('displays disk usage information', () => {
      expect(wrapper.text()).toContain('Disk Usage')
      expect(wrapper.text()).toContain('50.0%')
      expect(wrapper.text()).toContain('500.0 GB')
    })

    it('shows individual disk information', () => {
      expect(wrapper.text()).toContain('/dev/sda1')
      expect(wrapper.text()).toContain('/')
      expect(wrapper.text()).toContain('ext4')
    })
  })

  describe('Network Metrics Display', () => {
    it('displays network interface information', () => {
      expect(wrapper.text()).toContain('Network')
      expect(wrapper.text()).toContain('eth0')
      expect(wrapper.text()).toContain('192.168.1.100')
    })

    it('shows data transfer amounts', () => {
      expect(wrapper.text()).toContain('1.0 GB')
      expect(wrapper.text()).toContain('2.0 GB')
    })
  })

  describe('Loading States', () => {
    it('shows loading spinner when loading', async () => {
      systemStore.isLoading = true
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.loading-spinner').exists()).toBe(true)
      expect(wrapper.find('.loading-overlay').exists()).toBe(true)
    })

    it('disables refresh button during loading', async () => {
      systemStore.isLoading = true
      await wrapper.vm.$nextTick()

      const refreshButton = wrapper.find('.refresh-button')
      expect(refreshButton.attributes('disabled')).toBeDefined()
    })
  })

  describe('Error States', () => {
    it('displays error message when error occurs', async () => {
      systemStore.error = 'Failed to fetch system metrics'
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.error-message').exists()).toBe(true)
      expect(wrapper.text()).toContain('Failed to fetch system metrics')
    })

    it('shows retry button on error', async () => {
      systemStore.error = 'Connection failed'
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.retry-button').exists()).toBe(true)
    })
  })

  describe('Monitoring Controls', () => {
    it('shows monitoring status', () => {
      expect(wrapper.text()).toContain('Monitoring Active')
      expect(wrapper.find('.status-indicator.active').exists()).toBe(true)
    })

    it('shows paused status when monitoring stopped', async () => {
      systemStore.isMonitoring = false
      await wrapper.vm.$nextTick()

      expect(wrapper.text()).toContain('Monitoring Paused')
      expect(wrapper.find('.status-indicator.paused').exists()).toBe(true)
    })

    it('toggles monitoring when status is clicked', async () => {
      const toggleButton = wrapper.find('.monitoring-toggle')
      await toggleButton.trigger('click')

      expect(systemStore.startMonitoring).toHaveBeenCalled()
      expect(systemStore.stopMonitoring).toHaveBeenCalled()
    })
  })

  describe('Refresh Functionality', () => {
    it('calls refreshMetrics when refresh button is clicked', async () => {
      const refreshButton = wrapper.find('.refresh-button')
      await refreshButton.trigger('click')

      expect(systemStore.refreshMetrics).toHaveBeenCalled()
    })

    it('shows loading state during refresh', async () => {
      systemStore.refreshMetrics = vi.fn().mockImplementation(async () => {
        systemStore.isLoading = true
        await new Promise(resolve => setTimeout(resolve, 100))
        systemStore.isLoading = false
      })

      const refreshButton = wrapper.find('.refresh-button')
      await refreshButton.trigger('click')

      expect(wrapper.find('.loading-spinner').exists()).toBe(true)
    })
  })

  describe('Auto-refresh', () => {
    it('displays refresh interval controls', () => {
      expect(wrapper.find('.refresh-interval').exists()).toBe(true)
      expect(wrapper.find('.interval-select').exists()).toBe(true)
    })

    it('shows current refresh interval', () => {
      expect(wrapper.text()).toContain('1 second')
    })

    it('updates refresh interval when changed', async () => {
      const intervalSelect = wrapper.find('.interval-select')
      await intervalSelect.setValue('5000')

      expect(systemStore.refreshInterval).toBe(5000)
    })
  })

  describe('Timestamp Display', () => {
    it('displays last updated time', () => {
      expect(wrapper.find('.last-updated').exists()).toBe(true)
      expect(wrapper.text()).toContain('Last updated')
    })

    it('formats timestamp correctly', () => {
      const now = new Date()
      systemStore.lastUpdated = now

      // Check that time is displayed in readable format
      expect(wrapper.text()).toMatch(/\d{1,2}:\d{2}:\d{2}/)
    })

    it('updates timestamp when metrics refresh', async () => {
      const initialTime = systemStore.lastUpdated

      // Simulate metrics refresh
      systemStore.lastUpdated = new Date()
      await wrapper.vm.$nextTick()

      expect(systemStore.lastUpdated).not.toEqual(initialTime)
    })
  })

  describe('Responsive Design', () => {
    it('adapts layout for small screens', async () => {
      // Simulate small screen
      wrapper.vm.isSmallScreen = true
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.system-monitor').classes()).toContain('compact')
      expect(wrapper.find('.metrics-grid').classes()).toContain('grid-cols-1')
    })

    it('uses grid layout for larger screens', async () => {
      // Simulate large screen
      wrapper.vm.isSmallScreen = false
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.metrics-grid').classes()).toContain('grid-cols-2')
    })
  })

  describe('Performance Optimization', () => {
    it('debounces refresh requests', async () => {
      const refreshSpy = vi.spyOn(systemStore, 'refreshMetrics')

      // Rapidly click refresh button multiple times
      const refreshButton = wrapper.find('.refresh-button')
      for (let i = 0; i < 5; i++) {
        await refreshButton.trigger('click')
      }

      // Should only call refresh once due to debouncing
      expect(refreshSpy).toHaveBeenCalledTimes(1)
    })

    it('stops monitoring when component is unmounted', async () => {
      const stopSpy = vi.spyOn(systemStore, 'stopMonitoring')

      wrapper.unmount()

      expect(stopSpy).toHaveBeenCalled()
    })
  })

  describe('Accessibility', () => {
    it('has proper ARIA labels', () => {
      expect(wrapper.find('[aria-label="System monitoring status"]').exists()).toBe(true)
      expect(wrapper.find('[aria-label="Refresh system metrics"]').exists()).toBe(true)
    })

    it('supports keyboard navigation', async () => {
      const refreshButton = wrapper.find('.refresh-button')
      await refreshButton.trigger('keydown', { key: 'Enter' })

      expect(systemStore.refreshMetrics).toHaveBeenCalled()
    })

    it('announces status changes to screen readers', async () => {
      const statusElement = wrapper.find('.status-announcement')
      expect(statusElement.attributes('aria-live')).toBe('polite')
    })
  })
})

describe('SystemMonitor Integration', () => {
  it('integrates with system store correctly', () => {
    setActivePinia(createPinia())
    const systemStore = useSystemStore()

    const wrapper = mount(SystemMonitor, {
      global: {
        components: {
          CpuChart: true,
          MemoryChart: true,
          ProcessList: true
        }
      }
    })

    // Store should be accessible
    expect(wrapper.vm.systemStore).toBe(systemStore)
  })

  it('reacts to store changes', async () => {
    setActivePinia(createPinia())
    const systemStore = useSystemStore()

    const wrapper = mount(SystemMonitor, {
      global: {
        components: {
          CpuChart: true,
          MemoryChart: true,
          ProcessList: true
        }
      }
    })

    // Change store state
    systemStore.systemMetrics.cpu.usagePercent = 75.0
    await wrapper.vm.$nextTick()

    // Component should reflect the change
    expect(wrapper.text()).toContain('75.0%')
  })
})