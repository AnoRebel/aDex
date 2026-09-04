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

  // The component reads store.systemData, which is a computed view over
  // store.systemStats — so seed systemStats, not systemMetrics.
  function mountMonitor() {
    setActivePinia(createPinia())
    systemStore = useSystemStore()
    systemStore.systemStats = {
      cpu: { usage: 12.5, cores: [10, 20, 30, 40], coreCount: 4, modelName: 'Test CPU', frequency: 3200 },
      memory: { total: 16 * 1024 ** 3, used: 8 * 1024 ** 3, free: 8 * 1024 ** 3, usage: 50 },
      disk: [],
    } as any
    wrapper = mount(SystemMonitor)
    return { wrapper, systemStore }
  }

  beforeEach(() => {
    mountMonitor()
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.clearAllMocks()
  })

  describe('Component rendering', () => {
    it('renders the monitor container', () => {
      expect(wrapper.find('.system-monitor').exists()).toBe(true)
    })

    it('renders the header with a status indicator', () => {
      expect(wrapper.find('.monitor-header').exists()).toBe(true)
      expect(wrapper.find('.monitor-status').exists()).toBe(true)
      expect(wrapper.find('.status-dot').exists()).toBe(true)
    })

    it('reports whether monitoring is active or paused', () => {
      const status = wrapper.find('.monitor-status').text()
      expect(['ACTIVE', 'PAUSED']).toContain(status.trim())
    })

    it('renders a metric card per tracked resource', () => {
      expect(wrapper.find('.metrics-content').exists()).toBe(true)
      expect(wrapper.find('.cpu-card').exists()).toBe(true)
      expect(wrapper.find('.memory-card').exists()).toBe(true)
    })
  })

  describe('CPU card', () => {
    it('shows a CPU percentage', () => {
      expect(wrapper.find('.cpu-card .card-value').text()).toMatch(/^\d+\.\d%$/)
    })

    it('renders a progress bar sized to the usage', () => {
      const fill = wrapper.find('.progress-fill.cpu')
      expect(fill.exists()).toBe(true)
      expect(fill.attributes('style')).toContain('width')
    })

    it('lists the core count', () => {
      expect(wrapper.find('.cpu-card .card-details').text()).toContain('Cores:')
    })
  })

  describe('Memory card', () => {
    it('shows a memory percentage', () => {
      expect(wrapper.find('.memory-card .card-value').text()).toMatch(/^\d+\.\d%$/)
    })

    it('shows used and total', () => {
      const details = wrapper.find('.memory-card .card-details').text()
      expect(details).toContain('Used:')
      expect(details).toContain('Total:')
    })
  })

  describe('Error state', () => {
    it('shows the error container and hides the metrics when the store errors', async () => {
      systemStore.error = 'metrics unavailable'
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.error-container').exists()).toBe(true)
      expect(wrapper.find('.metrics-content').exists()).toBe(false)
      expect(wrapper.find('.error-message').text()).toContain('metrics unavailable')
    })

    it('offers a retry control while errored', async () => {
      systemStore.error = 'metrics unavailable'
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.retry-button').exists()).toBe(true)
    })
  })

  describe('Store integration', () => {
    it('reflects a CPU change from the store', async () => {
      systemStore.systemStats = { ...systemStore.systemStats, cpu: { ...systemStore.systemStats.cpu, usage: 75 } } as any
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.cpu-card .card-value').text()).toBe('75.0%')
    })

    it('reflects a memory change from the store', async () => {
      systemStore.systemStats = { ...systemStore.systemStats, memory: { ...systemStore.systemStats.memory, usage: 42 } } as any
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.memory-card .card-value').text()).toBe('42.0%')
    })
  })
})
