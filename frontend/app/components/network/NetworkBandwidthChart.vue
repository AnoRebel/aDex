<template>
  <div class="bandwidth-chart">
    <div class="chart-header">
      <h3>Bandwidth Usage</h3>
      <div class="chart-controls">
        <select v-model="selectedTimeRange" @change="refreshData" class="time-range-select">
          <option value="1">Last Hour</option>
          <option value="6">Last 6 Hours</option>
          <option value="24">Last 24 Hours</option>
        </select>
        <div class="interface-selector" v-if="availableInterfaces.length > 0">
          <select v-model="selectedInterface" @change="onInterfaceChange" class="interface-select">
            <option value="all">All Interfaces</option>
            <option v-for="iface in availableInterfaces" :key="iface.name" :value="iface.name">
              {{ iface.name }}
            </option>
          </select>
        </div>
      </div>
    </div>
    
    <div class="chart-container" ref="chartContainer">
      <canvas ref="chartCanvas" v-if="!loading && chartData"></canvas>
      <div v-if="loading" class="chart-loading">
        <Icon name="loader" class="spinning" />
        <p>Loading bandwidth data...</p>
      </div>
      <div v-if="!loading && !chartData" class="chart-empty">
        <Icon name="wifi-off" />
        <p>No bandwidth data available</p>
      </div>
    </div>

    <!-- Simple bandwidth display for immediate feedback -->
    <div class="bandwidth-display" v-if="!loading && hasData">
      <div class="bandwidth-item upload">
        <div class="bandwidth-label">Upload</div>
        <div class="bandwidth-value">{{ formatBandwidth(currentUpload) }}</div>
        <div class="bandwidth-bar">
          <div class="bandwidth-fill upload-fill" :style="{ width: uploadPercentage + '%' }"></div>
        </div>
        <div class="bandwidth-packets">{{ formatPackets(currentPacketsSent) }} packets</div>
      </div>
      <div class="bandwidth-item download">
        <div class="bandwidth-label">Download</div>
        <div class="bandwidth-value">{{ formatBandwidth(currentDownload) }}</div>
        <div class="bandwidth-bar">
          <div class="bandwidth-fill download-fill" :style="{ width: downloadPercentage + '%' }"></div>
        </div>
        <div class="bandwidth-packets">{{ formatPackets(currentPacketsRecv) }} packets</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useNetworkStore } from '~/stores/network'

// Chart.js will be imported dynamically
let Chart: any = null
let chartInstance: any = null

const props = defineProps<{
  data?: any[]
  loading?: boolean
}>()

const emit = defineEmits<{
  'interface-selected': [interfaceName: string]
}>()

// Store
const networkStore = useNetworkStore()

// Reactive data
const chartContainer = ref<HTMLElement>()
const chartCanvas = ref<HTMLCanvasElement>()
const selectedTimeRange = ref('1') // hours
const selectedInterface = ref('all')
const chartData = ref<any>(null)

// Computed properties
const availableInterfaces = computed(() => networkStore.metrics?.interfaces || [])

const currentUpload = computed(() => {
  if (!chartData.value || chartData.value.length === 0) return 0
  const latest = chartData.value[chartData.value.length - 1]
  return latest.upload || 0
})

const currentDownload = computed(() => {
  if (!chartData.value || chartData.value.length === 0) return 0
  const latest = chartData.value[chartData.value.length - 1]
  return latest.download || 0
})

const currentPacketsSent = computed(() => {
  if (!chartData.value || chartData.value.length === 0) return 0
  const latest = chartData.value[chartData.value.length - 1]
  return latest.packetsSent || 0
})

const currentPacketsRecv = computed(() => {
  if (!chartData.value || chartData.value.length === 0) return 0
  const latest = chartData.value[chartData.value.length - 1]
  return latest.packetsRecv || 0
})

const currentTotal = computed(() => currentUpload.value + currentDownload.value)

const peakUpload = computed(() => {
  if (!chartData.value) return 0
  return Math.max(...chartData.value.map((d: any) => d.upload || 0))
})

const peakDownload = computed(() => {
  if (!chartData.value) return 0
  return Math.max(...chartData.value.map((d: any) => d.download || 0))
})

const hasData = computed(() => {
  return networkStore.metrics && availableInterfaces.value.length > 0
})

const maxBandwidth = ref(100 * 1024 * 1024) // 100 Mbps in bytes

const uploadPercentage = computed(() => {
  return Math.min((currentUpload.value / maxBandwidth.value) * 100, 100)
})

const downloadPercentage = computed(() => {
  return Math.min((currentDownload.value / maxBandwidth.value) * 100, 100)
})

// Methods
const formatBandwidth = (bytes: number): string => {
  if (bytes === 0) return '0 Mbps'
  const mbps = (bytes * 8) / 1024 / 1024
  if (mbps < 1) {
    const kbps = mbps * 1024
    return `${kbps.toFixed(1)} Kbps`
  }
  if (mbps < 1000) {
    return `${mbps.toFixed(1)} Mbps`
  }
  const gbps = mbps / 1024
  return `${gbps.toFixed(2)} Gbps`
}

const formatPackets = (packets: number): string => {
  if (packets === 0) return '0'
  if (packets < 1000) return packets.toString()
  if (packets < 1000000) return `${(packets / 1000).toFixed(1)}K`
  return `${(packets / 1000000).toFixed(1)}M`
}

const formatTime = (timestamp: number): string => {
  return new Date(timestamp).toLocaleTimeString()
}

const loadChart = async () => {
  try {
    // Dynamically import Chart.js
    const ChartModule = await import('chart.js/auto')
    Chart = ChartModule.default
  } catch (error) {
    console.error('Failed to load Chart.js:', error)
    return
  }
}

const createChart = async () => {
  if (!Chart || !chartCanvas.value) return

  await nextTick()

  // Destroy existing chart
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }

  const ctx = chartCanvas.value.getContext('2d')
  if (!ctx) return

  const data = prepareChartData()

  chartInstance = new Chart(ctx, {
    type: 'line',
    data: {
      labels: data.labels,
      datasets: [
        {
          label: 'Upload',
          data: data.uploadData,
          borderColor: '#f59e0b',
          backgroundColor: 'rgba(245, 158, 11, 0.1)',
          borderWidth: 2,
          tension: 0.4,
          fill: true,
          pointRadius: 0,
          pointHoverRadius: 4,
        },
        {
          label: 'Download',
          data: data.downloadData,
          borderColor: '#3b82f6',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          borderWidth: 2,
          tension: 0.4,
          fill: true,
          pointRadius: 0,
          pointHoverRadius: 4,
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        legend: {
          display: false, // Using simple display instead
        },
        tooltip: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          titleColor: '#fff',
          bodyColor: '#fff',
          padding: 12,
          displayColors: true,
          callbacks: {
            label: (context: any) => {
              const label = context.dataset.label || ''
              const value = formatBandwidth(context.parsed.y)
              return `${label}: ${value}`
            },
            title: (context: any) => {
              return `Time: ${formatTime(context[0].label)}`
            }
          }
        }
      },
      scales: {
        x: {
          display: true,
          grid: { display: false },
          ticks: { color: '#9ca3af', maxRotation: 0, autoSkip: true, maxTicksLimit: 8 }
        },
        y: {
          display: true,
          position: 'left',
          grid: { color: 'rgba(156, 163, 175, 0.1)' },
          ticks: {
            color: '#9ca3af',
            callback: (value: number) => formatBandwidth(value)
          },
          title: {
            display: true,
            text: 'Bandwidth (Mbps)',
            color: '#9ca3af'
          }
        }
      }
    }
  })
}

const prepareChartData = () => {
  if (!chartData.value) {
    return { labels: [], uploadData: [], downloadData: [] }
  }

  const labels = chartData.value.map((d: any) => d.timestamp)
  const uploadData = chartData.value.map((d: any) => d.upload || 0)
  const downloadData = chartData.value.map((d: any) => d.download || 0)

  return { labels, uploadData, downloadData }
}

const refreshData = async () => {
  try {
    if (selectedInterface.value === 'all') {
      chartData.value = props.data || []
    } else {
      const data = await networkStore.fetchBandwidthData(selectedInterface.value)
      chartData.value = data.bandwidthHistory || []
    }

    await createChart()
  } catch (error) {
    console.error('Failed to refresh bandwidth data:', error)
  }
}

const onInterfaceChange = () => {
  emit('interface-selected', selectedInterface.value)
  refreshData()
}

const resizeChart = () => {
  if (chartInstance) chartInstance.resize()
}

// Watchers
watch(() => props.data, () => {
  if (selectedInterface.value === 'all') {
    chartData.value = props.data || []
    createChart()
  }
}, { deep: true })

watch(chartData, () => createChart(), { deep: true })

// Lifecycle
onMounted(async () => {
  await loadChart()
  window.addEventListener('resize', resizeChart)
  refreshData()
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeChart)
  if (chartInstance) chartInstance.destroy()
})
</script>

<style scoped>
.bandwidth-chart {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 1rem;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.chart-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-controls {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.time-range-select,
.interface-select {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-primary);
  font-size: 0.875rem;
}

.time-range-select:focus,
.interface-select:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.chart-container {
  flex: 1;
  position: relative;
  min-height: 300px;
  background: var(--surface-primary);
  border-radius: 4px;
  border: 1px solid var(--border-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-loading,
.chart-empty {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  color: var(--text-secondary);
}

.chart-loading svg,
.chart-empty svg {
  width: 48px;
  height: 48px;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.bandwidth-display {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-top: 1rem;
  padding: 1rem;
  background: var(--surface-secondary);
  border-radius: 6px;
  border: 1px solid var(--border-secondary);
}

.bandwidth-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bandwidth-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.bandwidth-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
}

.bandwidth-bar {
  width: 100%;
  height: 8px;
  background: var(--surface-primary);
  border-radius: 4px;
  overflow: hidden;
}

.bandwidth-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.upload-fill {
  background: linear-gradient(90deg, #f59e0b, #d97706);
}

.download-fill {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.bandwidth-packets {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-align: center;
  margin-top: 0.25rem;
}

/* Responsive design */
@media (max-width: 768px) {
  .chart-header {
    flex-direction: column;
    gap: 0.75rem;
    align-items: stretch;
  }
  
  .chart-controls {
    justify-content: center;
  }
  
  .bandwidth-stats {
    grid-template-columns: 1fr;
  }
  
  .bandwidth-display {
    gap: 0.75rem;
  }
}
</style>