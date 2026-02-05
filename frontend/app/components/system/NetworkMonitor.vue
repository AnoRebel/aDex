<template>
  <div class="network-monitor">
    <div class="network-header">
      <h3 class="glitch">NETWORK ANALYZER</h3>
      <div class="connection-indicator" :class="connectionStatus">
        <div class="indicator-dot"></div>
        <span>{{ connectionStatus.toUpperCase() }}</span>
      </div>
    </div>

    <div class="network-content">
      <!-- Real-time traffic chart -->
      <div class="traffic-chart">
        <div class="chart-header">
          <span class="chart-title">TRAFFIC MONITOR</span>
          <div class="chart-legend">
            <span class="legend-item download">
              <span class="legend-color"></span>
              DOWNLOAD
            </span>
            <span class="legend-item upload">
              <span class="legend-color"></span>
              UPLOAD
            </span>
          </div>
        </div>
        <div class="chart-container">
          <canvas
            ref="trafficChart"
            width="600"
            height="200"
          ></canvas>
        </div>
      </div>

      <!-- Connection stats -->
      <div class="connection-stats">
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-icon">📡</span>
            <span class="stat-title">CONNECTIONS</span>
          </div>
          <div class="stat-value">{{ networkData.connections?.total || 0 }}</div>
          <div class="stat-details">
            <span class="detail-item">
              <span class="detail-label">Active:</span>
              <span class="detail-value">{{ networkData.connections?.active || 0 }}</span>
            </span>
            <span class="detail-item">
              <span class="detail-label">Listening:</span>
              <span class="detail-value">{{ networkData.connections?.listening || 0 }}</span>
            </span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-icon">🌍</span>
            <span class="stat-title">LOCATION</span>
          </div>
          <div class="stat-value">{{ networkData.location?.city || 'Unknown' }}</div>
          <div class="stat-details">
            <span class="detail-item">
              <span class="detail-label">IP:</span>
              <span class="detail-value">{{ networkData.location?.ip || 'Unknown' }}</span>
            </span>
            <span class="detail-item">
              <span class="detail-label">ISP:</span>
              <span class="detail-value">{{ networkData.location?.isp || 'Unknown' }}</span>
            </span>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-icon">⚡</span>
            <span class="stat-title">LATENCY</span>
          </div>
          <div class="stat-value">{{ networkData.latency || 0 }}ms</div>
          <div class="stat-details">
            <span class="detail-item">
              <span class="detail-label">Quality:</span>
              <span class="detail-value" :class="latencyQuality">{{ latencyQualityText }}</span>
            </span>
          </div>
        </div>
      </div>

      <!-- Active connections table -->
      <div class="connections-table">
        <div class="table-header">
          <span class="table-title">ACTIVE CONNECTIONS</span>
          <button @click="refreshConnections" class="refresh-btn">
            <span class="refresh-icon">↻</span>
            REFRESH
          </button>
        </div>
        <div class="table-container">
          <table class="connections-table-content">
            <thead>
              <tr>
                <th>PROTOCOL</th>
                <th>LOCAL ADDRESS</th>
                <th>REMOTE ADDRESS</th>
                <th>STATE</th>
                <th>PID</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(connection, index) in activeConnections"
                :key="index"
                class="connection-row"
              >
                <td class="protocol-cell">
                  <span class="protocol-badge" :class="connection.protocol">
                    {{ connection.protocol }}
                  </span>
                </td>
                <td class="address-cell">{{ connection.localAddress }}</td>
                <td class="address-cell">{{ connection.remoteAddress }}</td>
                <td class="state-cell">
                  <span class="state-badge" :class="connection.state">
                    {{ connection.state }}
                  </span>
                </td>
                <td class="pid-cell">{{ connection.pid }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Bandwidth usage summary -->
      <div class="bandwidth-summary">
        <div class="summary-header">
          <span class="summary-title">BANDWIDTH SUMMARY</span>
          <div class="time-range">
            <button
              v-for="range in timeRanges"
              :key="range.value"
              @click="selectedTimeRange = range.value"
              class="range-btn"
              :class="{ active: selectedTimeRange === range.value }"
            >
              {{ range.label }}
            </button>
          </div>
        </div>
        <div class="summary-stats">
          <div class="summary-stat">
            <span class="summary-label">Total Download</span>
            <span class="summary-value download">{{ formatBytes(totalDownload) }}</span>
          </div>
          <div class="summary-stat">
            <span class="summary-label">Total Upload</span>
            <span class="summary-value upload">{{ formatBytes(totalUpload) }}</span>
          </div>
          <div class="summary-stat">
            <span class="summary-label">Peak Download</span>
            <span class="summary-value">{{ formatBytes(peakDownload) }}/s</span>
          </div>
          <div class="summary-stat">
            <span class="summary-label">Peak Upload</span>
            <span class="summary-value">{{ formatBytes(peakUpload) }}/s</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'

interface NetworkData {
  connections?: {
    total: number
    active: number
    listening: number
  }
  location?: {
    city: string
    country: string
    ip: string
    isp: string
  }
  latency?: number
  interfaces?: Array<{
    name: string
    type: string
    speed: number
    status: string
  }>
}

interface Connection {
  protocol: 'TCP' | 'UDP'
  localAddress: string
  remoteAddress: string
  state: string
  pid: number
  processName?: string
}

interface TrafficData {
  timestamp: number
  download: number
  upload: number
}

const trafficChart = ref<HTMLCanvasElement>()
const networkData = ref<NetworkData>({})
const activeConnections = ref<Connection[]>([])
const trafficHistory = ref<TrafficData[]>([])
const selectedTimeRange = ref('1h')

const maxHistoryPoints = 60

// Time range options
const timeRanges = [
  { label: '1H', value: '1h' },
  { label: '6H', value: '6h' },
  { label: '24H', value: '24h' },
  { label: '7D', value: '7d' }
]

// Computed properties
const connectionStatus = computed(() => {
  const latency = networkData.value.latency || 0
  if (latency < 50) return 'excellent'
  if (latency < 100) return 'good'
  if (latency < 200) return 'fair'
  return 'poor'
})

const latencyQuality = computed(() => {
  const latency = networkData.value.latency || 0
  if (latency < 50) return 'excellent'
  if (latency < 100) return 'good'
  if (latency < 200) return 'fair'
  return 'poor'
})

const latencyQualityText = computed(() => {
  const latency = networkData.value.latency || 0
  if (latency < 50) return 'Excellent'
  if (latency < 100) return 'Good'
  if (latency < 200) return 'Fair'
  return 'Poor'
})

const totalDownload = computed(() => {
  return trafficHistory.value.reduce((sum, data) => sum + data.download, 0)
})

const totalUpload = computed(() => {
  return trafficHistory.value.reduce((sum, data) => sum + data.upload, 0)
})

const peakDownload = computed(() => {
  return Math.max(...trafficHistory.value.map(data => data.download), 0)
})

const peakUpload = computed(() => {
  return Math.max(...trafficHistory.value.map(data => data.upload), 0)
})

// Methods
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const drawTrafficChart = () => {
  if (!trafficChart.value) return

  const ctx = trafficChart.value.getContext('2d')
  if (!ctx) return

  const width = trafficChart.value.width
  const height = trafficChart.value.height

  // Clear canvas
  ctx.clearRect(0, 0, width, height)

  // Set up grid
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.1)'
  ctx.lineWidth = 1

  // Draw horizontal grid lines
  for (let i = 0; i <= 5; i++) {
    const y = (height / 5) * i
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(width, y)
    ctx.stroke()
  }

  // Draw vertical grid lines
  for (let i = 0; i <= 10; i++) {
    const x = (width / 10) * i
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, height)
    ctx.stroke()
  }

  // Find max values for scaling
  const maxDownload = Math.max(...trafficHistory.value.map(d => d.download), 1)
  const maxUpload = Math.max(...trafficHistory.value.map(d => d.upload), 1)
  const maxValue = Math.max(maxDownload, maxUpload)

  const stepX = width / (maxHistoryPoints - 1)

  // Draw download line
  ctx.strokeStyle = '#10b981'
  ctx.lineWidth = 2
  ctx.shadowBlur = 10
  ctx.shadowColor = '#10b981'
  ctx.beginPath()

  trafficHistory.value.forEach((data, index) => {
    const x = index * stepX
    const y = height - (data.download / maxValue) * height

    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })

  ctx.stroke()

  // Draw upload line
  ctx.strokeStyle = '#f59e0b'
  ctx.shadowColor = '#f59e0b'
  ctx.beginPath()

  trafficHistory.value.forEach((data, index) => {
    const x = index * stepX
    const y = height - (data.upload / maxValue) * height

    if (index === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  })

  ctx.stroke()

  // Add fill effect
  ctx.fillStyle = 'rgba(16, 185, 129, 0.1)'
  ctx.beginPath()
  ctx.moveTo(0, height)

  trafficHistory.value.forEach((data, index) => {
    const x = index * stepX
    const y = height - (data.download / maxValue) * height
    ctx.lineTo(x, y)
  })

  ctx.lineTo(width, height)
  ctx.closePath()
  ctx.fill()
}

const updateNetworkData = async () => {
  try {
    // Simulate network data (replace with actual API calls)
    networkData.value = {
      connections: {
        total: Math.floor(Math.random() * 50) + 10,
        active: Math.floor(Math.random() * 20) + 5,
        listening: Math.floor(Math.random() * 10) + 2
      },
      location: {
        city: 'San Francisco',
        country: 'United States',
        ip: '192.168.1.100',
        isp: 'Fiber ISP'
      },
      latency: Math.floor(Math.random() * 150) + 10
    }

    // Update traffic history
    const currentTraffic: TrafficData = {
      timestamp: Date.now(),
      download: Math.random() * 1024 * 1024, // Random bytes
      upload: Math.random() * 1024 * 1024
    }

    trafficHistory.value.push(currentTraffic)
    if (trafficHistory.value.length > maxHistoryPoints) {
      trafficHistory.value.shift()
    }

    // Update active connections (sample data)
    activeConnections.value = [
      {
        protocol: 'TCP',
        localAddress: '192.168.1.100:54321',
        remoteAddress: '74.125.224.72:443',
        state: 'ESTABLISHED',
        pid: 1234,
        processName: 'chrome'
      },
      {
        protocol: 'TCP',
        localAddress: '192.168.1.100:54322',
        remoteAddress: '151.101.1.69:443',
        state: 'ESTABLISHED',
        pid: 1234,
        processName: 'chrome'
      },
      {
        protocol: 'UDP',
        localAddress: '192.168.1.100:53',
        remoteAddress: '8.8.8.8:53',
        state: 'ESTABLISHED',
        pid: 567,
        processName: 'systemd-resolve'
      }
    ]

    await nextTick()
    drawTrafficChart()
  } catch (error) {
    console.error('Failed to update network data:', error)
  }
}

const refreshConnections = () => {
  updateNetworkData()
}

let updateInterval: NodeJS.Timeout

onMounted(() => {
  updateNetworkData()
  updateInterval = setInterval(updateNetworkData, 2000)
})

onUnmounted(() => {
  if (updateInterval) {
    clearInterval(updateInterval)
  }
})
</script>

<style scoped>
.network-monitor {
  background: rgba(10, 10, 10, 0.95);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  padding: 20px;
  backdrop-filter: blur(10px);
  font-family: 'Fira Code', monospace;
  color: var(--text-primary);
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
}

.network-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid var(--surface-border);
}

.connection-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.connection-indicator.excellent {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.connection-indicator.excellent .indicator-dot {
  background: #10b981;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
}

.connection-indicator.good {
  background: rgba(14, 165, 233, 0.2);
  color: #0ea5e9;
}

.connection-indicator.good .indicator-dot {
  background: #0ea5e9;
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.5);
}

.connection-indicator.fair {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.connection-indicator.fair .indicator-dot {
  background: #f59e0b;
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
}

.connection-indicator.poor {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.connection-indicator.poor .indicator-dot {
  background: #ef4444;
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
}

.network-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.traffic-chart {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.chart-title {
  color: var(--primary-400);
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.chart-legend {
  display: flex;
  gap: 15px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  color: var(--text-secondary);
}

.legend-color {
  width: 12px;
  height: 3px;
  border-radius: 2px;
}

.legend-item.download .legend-color {
  background: #10b981;
}

.legend-item.upload .legend-color {
  background: #f59e0b;
}

.chart-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

.connection-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
}

.stat-card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  transition: all 0.3s ease;
}

.stat-card:hover {
  border-color: var(--primary-500);
  transform: translateY(-2px);
}

.stat-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.stat-icon {
  font-size: 16px;
}

.stat-title {
  color: var(--primary-400);
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--accent-400);
  margin-bottom: 8px;
}

.stat-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
}

.detail-label {
  color: var(--text-secondary);
}

.detail-value {
  color: var(--text-primary);
}

.detail-value.excellent {
  color: #10b981;
}

.detail-value.good {
  color: #0ea5e9;
}

.detail-value.fair {
  color: #f59e0b;
}

.detail-value.poor {
  color: #ef4444;
}

.connections-table {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.table-title {
  color: var(--primary-400);
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.refresh-icon {
  font-size: 12px;
  transition: transform 0.3s ease;
}

.refresh-btn:hover .refresh-icon {
  transform: rotate(180deg);
}

.table-container {
  overflow-x: auto;
}

.connections-table-content {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
}

.connections-table-content th {
  text-align: left;
  padding: 8px;
  border-bottom: 1px solid var(--surface-border);
  color: var(--primary-400);
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.connections-table-content td {
  padding: 8px;
  border-bottom: 1px solid var(--surface-elevated);
}

.connection-row:hover {
  background: var(--surface-elevated);
}

.protocol-cell {
  width: 80px;
}

.protocol-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: bold;
  text-transform: uppercase;
}

.protocol-badge.TCP {
  background: rgba(14, 165, 233, 0.2);
  color: #0ea5e9;
}

.protocol-badge.UDP {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.address-cell {
  font-family: monospace;
  color: var(--text-secondary);
  min-width: 150px;
}

.state-cell {
  width: 100px;
}

.state-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 9px;
  font-weight: bold;
}

.state-badge.ESTABLISHED {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.pid-cell {
  width: 60px;
  color: var(--text-secondary);
}

.bandwidth-summary {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
}

.summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.summary-title {
  color: var(--primary-400);
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.time-range {
  display: flex;
  gap: 5px;
}

.range-btn {
  padding: 4px 8px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 9px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.range-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.range-btn.active {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

.summary-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 15px;
}

.summary-stat {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.summary-label {
  color: var(--text-secondary);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.summary-value {
  font-size: 16px;
  font-weight: bold;
  color: var(--text-primary);
}

.summary-value.download {
  color: #10b981;
}

.summary-value.upload {
  color: #f59e0b;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Responsive design */
@media (max-width: 768px) {
  .connection-stats {
    grid-template-columns: 1fr;
  }

  .summary-stats {
    grid-template-columns: repeat(2, 1fr);
  }

  .table-container {
    font-size: 10px;
  }

  .connections-table-content th,
  .connections-table-content td {
    padding: 6px 4px;
  }
}
</style>