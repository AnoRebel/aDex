<template>
  <div class="system-monitor" :class="{ 'compact': compact, 'loading': isLoading }">
    <!-- Header with status and controls -->
    <div class="monitor-header">
      <div class="monitor-status">
        <div class="status-indicator" :class="systemHealth?.status">
          <div class="status-dot"></div>
          <span class="status-text">
            {{ isMonitoring ? 'Monitoring Active' : 'Monitoring Paused' }}
          </span>
        </div>
        <div class="health-score" :class="performanceScore">
          <div class="score-circle">
            <span class="score-value">{{ systemHealth?.score || 0 }}</span>
          </div>
          <div class="score-label">{{ performanceScore }}</div>
        </div>
      </div>

      <div class="monitor-controls">
        <button
          @click="refreshMetrics"
          :disabled="isLoading"
          class="refresh-button"
          :class="{ 'loading': isLoading }"
          title="Refresh system metrics"
        >
          <svg class="icon" :class="{ 'spin': isLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{{ isLoading ? 'Refreshing...' : 'Refresh' }}</span>
        </button>

        <button
          @click="toggleMonitoring"
          :class="isMonitoring ? 'stop' : 'start'"
          :title="isMonitoring ? 'Stop monitoring' : 'Start monitoring'"
        >
          <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 00.555.894l3.197-2.132a1 1 0 00.555-.894V9.87a1 1 0 00-.555-.894zM15 4.5a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z" />
          </svg>
          <span>{{ isMonitoring ? 'Stop' : 'Start' }}</span>
        </button>

        <div class="refresh-interval">
          <label for="interval-select">Update every:</label>
          <select
            id="interval-select"
            v-model="selectedInterval"
            @change="updateRefreshInterval"
            class="interval-select"
          >
            <option value="500">0.5s</option>
            <option value="1000">1s</option>
            <option value="2000">2s</option>
            <option value="5000">5s</option>
            <option value="10000">10s</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Error display -->
    <div v-if="error" class="error-container">
      <div class="error-message">
        <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>{{ error }}</span>
      </div>
      <button @click="refreshMetrics" class="retry-button">Retry</button>
    </div>

    <!-- Main metrics grid -->
    <div v-else class="metrics-grid">
      <!-- CPU Metrics Card -->
      <div class="metric-card cpu-card">
        <div class="card-header">
          <h3 class="card-title">CPU</h3>
          <div class="cpu-info">
            <span class="cpu-model">{{ cpuMetrics?.model || 'Unknown' }}</span>
            <span class="cpu-cores">{{ cpuMetrics?.cores || 0 }} cores</span>
          </div>
        </div>
        <div class="card-content">
          <div class="usage-display">
            <div class="usage-circle" :style="{ '--usage': cpuUsage + '%' }">
              <span class="usage-value">{{ cpuUsage.toFixed(1) }}%</span>
            </div>
            <div class="usage-details">
              <div class="frequency-info">
                <span class="current-freq">{{ formatFrequency(cpuMetrics?.frequency || 0) }}</span>
                <span class="max-freq">{{ formatFrequency(cpuMetrics?.frequencyMax || 0) }}</span>
              </div>
              <div class="load-average" v-if="cpuMetrics?.loadAverage">
                <span>Load: {{ cpuMetrics.loadAverage[0]?.toFixed(2) || 'N/A' }}</span>
              </div>
            </div>
          </div>
          <div class="cpu-chart">
            <CpuChart
              :data="{ overall: cpuHistory, perCore: getPerCoreHistory() }"
              :height="120"
              :show-per-core="!compact"
              theme="auto"
            />
          </div>
        </div>
      </div>

      <!-- Memory Metrics Card -->
      <div class="metric-card memory-card">
        <div class="card-header">
          <h3 class="card-title">Memory</h3>
          <div class="memory-info">
            <span class="total-memory">{{ formatBytes(memoryMetrics?.total || 0) }}</span>
          </div>
        </div>
        <div class="card-content">
          <div class="memory-breakdown">
            <div class="memory-row">
              <span class="memory-label">Used</span>
              <div class="memory-bar">
                <div class="memory-progress used" :style="{ width: memoryUsage + '%' }"></div>
              </div>
              <span class="memory-value">{{ formatBytes(memoryMetrics?.used || 0) }}</span>
            </div>
            <div class="memory-row">
              <span class="memory-label">Free</span>
              <div class="memory-bar">
                <div class="memory-progress free" :style="{ width: (100 - memoryUsage) + '%' }"></div>
              </div>
              <span class="memory-value">{{ formatBytes(memoryMetrics?.free || 0) }}</span>
            </div>
            <div class="memory-row swap" v-if="memoryMetrics?.swapTotal > 0">
              <span class="memory-label">Swap</span>
              <div class="memory-bar">
                <div class="memory-progress swap" :style="{ width: swapUsage + '%' }"></div>
              </div>
              <span class="memory-value">{{ formatBytes(memoryMetrics?.swapUsed || 0) }}</span>
            </div>
          </div>
          <div class="memory-chart">
            <MemoryChart
              :data="{ used: memoryHistory, free: getMemoryFreeHistory(), swap: getSwapHistory() }"
              :height="120"
              :show-swap="memoryMetrics?.swapTotal > 0"
              theme="auto"
            />
          </div>
        </div>
      </div>

      <!-- Processes Card -->
      <div class="metric-card processes-card">
        <div class="card-header">
          <h3 class="card-title">Processes</h3>
          <div class="process-counts">
            <span class="total-count">{{ totalProcesses }} total</span>
            <span class="running-count">{{ runningProcesses }} running</span>
          </div>
        </div>
        <div class="card-content">
          <div class="process-summary">
            <div class="process-stats">
              <div class="stat-item">
                <span class="stat-value">{{ runningProcesses }}</span>
                <span class="stat-label">Running</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ sleepingProcesses }}</span>
                <span class="stat-label">Sleeping</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ getProcessCount('stopped') }}</span>
                <span class="stat-label">Stopped</span>
              </div>
            </div>
          </div>
          <div class="top-processes">
            <div class="processes-section">
              <h4>Top CPU</h4>
              <ProcessList
                :processes="getTopCPUProcesses(5)"
                :compact="true"
                :show-details="false"
                max-items="5"
              />
            </div>
            <div class="processes-section">
              <h4>Top Memory</h4>
              <ProcessList
                :processes="getTopMemoryProcesses(5)"
                :compact="true"
                :show-details="false"
                max-items="5"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Disk Usage Card -->
      <div class="metric-card disk-card" v-if="enableDisk">
        <div class="card-header">
          <h3 class="card-title">Disk Usage</h3>
          <div class="disk-info">
            <span class="total-space">{{ formatBytes(getTotalDiskSpace()) }}</span>
          </div>
        </div>
        <div class="card-content">
          <div class="disk-list">
            <div
              v-for="disk in getDiskUsageList()"
              :key="disk.device"
              class="disk-item"
              :class="{ 'high-usage': disk.usagePercent > 80 }"
            >
              <div class="disk-info">
                <span class="disk-device">{{ disk.device }}</span>
                <span class="disk-mountpoint">{{ disk.mountpoint }}</span>
              </div>
              <div class="disk-usage">
                <div class="disk-bar">
                  <div
                    class="disk-progress"
                    :style="{ width: disk.usagePercent + '%' }"
                    :class="{ 'critical': disk.usagePercent > 95 }"
                  ></div>
                </div>
                <div class="disk-details">
                  <span class="usage-percent">{{ disk.usagePercent.toFixed(1) }}%</span>
                  <span class="usage-text">{{ formatBytes(disk.used) }} / {{ formatBytes(disk.total) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Network Interface Card -->
      <div class="metric-card network-card" v-if="enableNetwork">
        <div class="card-header">
          <h3 class="card-title">Network</h3>
          <div class="network-info">
            <span class="interface-count">{{ networkMetrics?.interfaces?.length || 0 }} interfaces</span>
          </div>
        </div>
        <div class="card-content">
          <div class="network-summary">
            <div class="network-stats">
              <div class="stat-item">
                <span class="stat-value">{{ formatBytes(totalNetworkUsage?.sent || 0) }}</span>
                <span class="stat-label">Sent</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ formatBytes(totalNetworkUsage?.received || 0) }}</span>
                <span class="stat-label">Received</span>
              </div>
            </div>
          </div>
          <div class="network-interfaces">
            <div
              v-for="iface in getActiveNetworkInterfaces()"
              :key="iface.name"
              class="network-interface"
            >
              <div class="interface-status" :class="{ 'active': iface.isUp }"></div>
              <div class="interface-info">
                <span class="interface-name">{{ iface.name }}</span>
                <span class="interface-ips">{{ iface.ipAddresses.join(', ') }}</span>
                <span class="interface-speed" v-if="iface.speed">
                  {{ formatNetworkSpeed(iface.speed) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Temperature Card -->
      <div class="metric-card temperature-card" v-if="enableTemperature && temperatureMetrics">
        <div class="card-header">
          <h3 class="card-title">Temperature</h3>
          <div class="temperature-info">
            <span class="sensor-count">{{ temperatureMetrics.sensors.length }} sensors</span>
          </div>
        </div>
        <div class="card-content">
          <div class="temperature-list">
            <div
              v-for="sensor in getTemperatureSensors()"
              :key="sensor.name"
              class="temperature-item"
              :class="{
                'critical': sensor.temperature >= sensor.critical,
                'high': sensor.temperature >= sensor.high
              }"
            >
              <div class="sensor-info">
                <span class="sensor-name">{{ sensor.name }}</span>
                <span class="sensor-status" :class="getTemperatureStatus(sensor)">
                  {{ getTemperatureStatus(sensor) }}
                </span>
              </div>
              <div class="temperature-display">
                <div class="temperature-value">
                  {{ sensor.temperature.toFixed(1) }}°{{ sensor.unit }}
                </div>
                <div class="temperature-range">
                  <span>Min: {{ sensor.min.toFixed(1) }}°</span>
                  <span>Max: {{ sensor.max.toFixed(1) }}°</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer with last updated time -->
    <div class="monitor-footer">
      <div class="last-updated">
        <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>Last updated: {{ formatLastUpdated() }}</span>
      </div>
      <div class="health-indicator">
        <div class="health-status" :class="systemHealth?.status">
          {{ systemHealth?.status?.toUpperCase() }}
        </div>
        <div class="health-score">{{ systemHealth?.score }}/100</div>
      </div>
    </div>

    <!-- Loading overlay -->
    <div v-if="isLoading" class="loading-overlay">
      <div class="loading-spinner"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useDebounceFn, useIntervalFn } from '@vueuse/core'
import CpuChart from './CpuChart.vue'
import MemoryChart from './MemoryChart.vue'
import ProcessList from './ProcessList.vue'
import { useSystemStore } from '~/stores/system'
import type { SystemMonitorProps } from '~/types/system'

// Props
interface Props extends /* @vue-ignore */ SystemMonitorProps {}

const props = withDefaults(defineProps<Props>(), {
  autoRefresh: true,
  refreshInterval: 1000,
  showCharts: true,
  showDetails: true,
  maxProcesses: 100,
  compact: false
})

// Store
const systemStore = useSystemStore()

// Local state
const selectedInterval = ref(props.refreshInterval.toString())

// Computed
const {
  systemMetrics,
  isLoading,
  error,
  lastUpdated,
  isMonitoring,
  cpuMetrics,
  memoryMetrics,
  processMetrics,
  diskMetrics,
  networkMetrics,
  temperatureMetrics,
  cpuUsage,
  memoryUsage,
  swapUsage,
  totalProcesses,
  runningProcesses,
  sleepingProcesses,
  systemHealth,
  performanceScore
} = storeToRefs(systemStore)

// Computed helpers
const enableDisk = computed(() => systemStore.enableDisk)
const enableNetwork = computed(() => systemStore.enableNetwork)
const totalNetworkUsage = computed(() => ({
  sent: networkMetrics.value?.totalBytesSent || 0,
  received: networkMetrics.value?.totalBytesRecv || 0
}))

// Chart data helpers
const cpuHistory = computed(() => systemStore.cpuHistory)
const memoryHistory = computed(() => systemStore.memoryHistory)

// Methods
const refreshMetrics = async () => {
  if (isLoading.value) return
  await systemStore.refreshMetrics()
}

const debouncedRefresh = useDebounceFn(refreshMetrics, 100)

const toggleMonitoring = async () => {
  if (isMonitoring.value) {
    systemStore.stopMonitoring()
  } else {
    systemStore.startMonitoring()
    if (!systemMetrics.value) {
      await refreshMetrics()
    }
  }
}

const updateRefreshInterval = () => {
  const interval = parseInt(selectedInterval.value)
  systemStore.setRefreshInterval(interval)
}

// Formatting functions
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

const formatFrequency = (mhz: number): string => {
  if (mhz < 1000) {
    return `${mhz.toFixed(0)} MHz`
  }
  return `${(mhz / 1000).toFixed(2)} GHz`
}

const formatNetworkSpeed = (bps: number): string => {
  if (bps < 1000) return `${bps} b/s`
  if (bps < 1000000) return `${(bps / 1000).toFixed(1)} Kb/s`
  if (bps < 1000000) return `${(bps / 1000000).toFixed(1)} Mb/s`
  return `${(bps / 1000000).toFixed(1)} Gb/s`
}

const formatLastUpdated = (): string => {
  if (!lastUpdated.value) return 'Never'
  const now = new Date()
  const diff = now.getTime() - lastUpdated.value.getTime()

  if (diff < 60000) {
    return `${Math.floor(diff / 1000)}s ago`
  }
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}m ago`
  }
  return lastUpdated.value.toLocaleTimeString()
}

// Helper functions for computed data
const getPerCoreHistory = () => {
  // This would be implemented to get per-core historical data
  return []
}

const getMemoryFreeHistory = () => {
  return memoryHistory.value.map((used, index) => ({
    value: (100 - used),
    timestamp: new Date(Date.now() - (memoryHistory.value.length - index - 1) * 1000)
  }))
}

const getSwapHistory = () => {
  // This would be implemented to get swap historical data
  return []
}

const getTopCPUProcesses = (limit: number) => {
  if (!processMetrics.value) return []
  return [...processMetrics.value.processes]
    .sort((a, b) => b.cpuPercent - a.cpuPercent)
    .slice(0, limit)
}

const getTopMemoryProcesses = (limit: number) => {
  if (!processMetrics.value) return []
  return [...processMetrics.value.processes]
    .sort((a, b) => b.memoryPercent - a.memoryPercent)
    .slice(0, limit)
}

const getProcessCount = (status: string) => {
  if (!processMetrics.value) return 0
  switch (status.toLowerCase()) {
    case 'running': return processMetrics.value.runningProcesses
    case 'sleeping': return processMetrics.value.sleepingProcesses
    case 'stopped': return processMetrics.value.stoppedProcesses
    case 'zombie': return processMetrics.value.zombieProcesses
    default: return 0
  }
}

const getDiskUsageList = () => {
  if (!diskMetrics.value) return []
  return diskMetrics.value.disks
    .sort((a, b) => b.usagePercent - a.usagePercent)
}

const getTotalDiskSpace = () => {
  if (!diskMetrics.value) return 0
  return diskMetrics.value.totalSpace
}

const getActiveNetworkInterfaces = () => {
  if (!networkMetrics.value) return []
  return networkMetrics.value.interfaces.filter(iface => iface.isUp)
}

const getTemperatureSensors = () => {
  if (!temperatureMetrics.value) return []
  return [...temperatureMetrics.value.sensors]
    .sort((a, b) => b.temperature - a.temperature)
}

const getTemperatureStatus = (sensor: any) => {
  if (sensor.temperature >= sensor.critical) return 'critical'
  if (sensor.temperature >= sensor.high) return 'high'
  return 'normal'
}

// Auto-refresh
let intervalPause: (() => void) | null = null

onMounted(async () => {
  if (props.autoRefresh) {
    systemStore.startMonitoring()
    await refreshMetrics()

    intervalPause = useIntervalFn(() => {
      if (isMonitoring.value) {
        debouncedRefresh()
      }
    }, props.refreshInterval)
  }
})

onUnmounted(() => {
  if (intervalPause) {
    intervalPause()
  }
  systemStore.stopMonitoring()
})

// Watch for prop changes
watch(() => props.refreshInterval, (newInterval) => {
  selectedInterval.value = newInterval.toString()
  if (intervalPause) {
    intervalPause()
    intervalPause = useIntervalFn(() => {
      if (isMonitoring.value) {
        debouncedRefresh()
      }
    }, newInterval)
  }
})

watch(() => props.maxProcesses, (newMax) => {
  systemStore.setMaxProcesses(newMax)
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.system-monitor {
  @apply bg-white dark:bg-gray-900 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700;
  @apply overflow-hidden relative;
  @apply transition-all duration-200;
}

.system-monitor.loading {
  @apply opacity-75;
}

.monitor-header {
  @apply flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700;
  @apply bg-gray-50 dark:bg-gray-800;
}

.monitor-status {
  @apply flex items-center gap-4;
}

.status-indicator {
  @apply flex items-center gap-2;
}

.status-dot {
  @apply w-3 h-3 rounded-full;
  @apply transition-all duration-200;
}

.status-indicator.good .status-dot {
  @apply bg-green-500;
}

.status-indicator.warning .status-dot {
  @apply bg-yellow-500;
}

.status-indicator.critical .status-dot {
  @apply bg-red-500;
}

.status-indicator.unknown .status-dot {
  @apply bg-gray-500;
}

.status-text {
  @apply text-sm font-medium text-gray-700 dark:text-gray-300;
}

.health-score {
  @apply flex items-center gap-2;
}

.score-circle {
  @apply w-8 h-8 rounded-full flex items-center justify-center;
  @apply text-xs font-bold text-white;
}

.health-score.excellent .score-circle {
  @apply bg-green-500;
}

.health-score.good .score-circle {
  @apply bg-blue-500;
}

.health-score.fair .score-circle {
  @apply bg-yellow-500;
}

.health-score.poor .score-circle {
  @apply bg-red-500;
}

.score-label {
  @apply text-xs font-medium text-gray-600 dark:text-gray-400;
}

.monitor-controls {
  @apply flex items-center gap-3;
}

.refresh-button,
.monitoring-button {
  @apply flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium;
  @apply bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600;
  @apply hover:bg-gray-50 dark:hover:bg-gray-600;
  @apply transition-colors duration-200;
}

.refresh-button:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.refresh-button.loading {
  @apply text-blue-600 dark:text-blue-400;
}

.monitoring-button.start {
  @apply bg-green-500 text-white hover:bg-green-600 border-green-500;
}

.monitoring-button.stop {
  @apply bg-red-500 text-white hover:bg-red-600 border-red-500;
}

.icon {
  @apply w-4 h-4;
}

.icon.spin {
  @apply animate-spin;
}

.refresh-interval {
  @apply flex items-center gap-2;
}

.interval-select {
  @apply border border-gray-300 dark:border-gray-600 rounded-md px-2 py-1 text-sm;
  @apply bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300;
  @apply focus:outline-none focus:ring-2 focus:ring-blue-500;
}

.error-container {
  @apply p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800;
  @apply flex items-center justify-between;
}

.error-message {
  @apply flex items-center gap-2 text-red-700 dark:text-red-300;
}

.retry-button {
  @apply px-3 py-1 bg-red-500 text-white rounded-md text-sm;
  @apply hover:bg-red-600 transition-colors duration-200;
}

.metrics-grid {
  @apply grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 p-4;
}

.metric-card {
  @apply bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700;
  @apply overflow-hidden;
}

.card-header {
  @apply flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700;
}

.card-title {
  @apply text-lg font-semibold text-gray-900 dark:text-gray-100;
}

.card-content {
  @apply p-4 space-y-4;
}

/* CPU Card Styles */
.cpu-info {
  @apply text-sm text-gray-600 dark:text-gray-400 space-x-4;
}

.usage-display {
  @apply flex items-center justify-between;
}

.usage-circle {
  @apply relative w-16 h-16 rounded-full;
  @apply bg-gray-200 dark:bg-gray-700;
  @apply flex items-center justify-center;
  @apply before:content-[''];
  @apply before:absolute inset-0 rounded-full;
  @apply before:bg-gradient-to-r from-blue-500 to-cyan-500;
  @apply before:transition-all duration-300;
  @apply before:[clip-path:polygon(0%_0%,_var(--usage)_0%,_var(--usage)_100%,_0%_100%)];
}

.usage-value {
  @apply relative z-10 text-sm font-bold text-gray-900 dark:text-gray-100;
}

.frequency-info,
.load-average {
  @apply text-xs text-gray-600 dark:text-gray-400;
}

/* Memory Card Styles */
.memory-info {
  @apply text-sm text-gray-600 dark:text-gray-400;
}

.memory-breakdown {
  @apply space-y-2;
}

.memory-row {
  @apply flex items-center justify-between text-sm;
}

.memory-label {
  @apply w-12 text-gray-600 dark:text-gray-400;
}

.memory-bar {
  @apply flex-1 mx-2 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden;
}

.memory-progress {
  @apply h-full transition-all duration-300;
}

.memory-progress.used {
  @apply bg-blue-500;
}

.memory-progress.free {
  @apply bg-green-500;
}

.memory-progress.swap {
  @apply bg-orange-500;
}

.memory-value {
  @apply w-24 text-right text-gray-900 dark:text-gray-100;
}

.memory-row.swap .memory-label {
  @apply text-orange-600 dark:text-orange-400;
}

/* Process Card Styles */
.process-counts {
  @apply text-sm text-gray-600 dark:text-gray-400 space-x-4;
}

.process-summary {
  @apply space-y-3;
}

.process-stats {
  @apply grid grid-cols-3 gap-4 text-center;
}

.stat-item {
  @apply space-y-1;
}

.stat-value {
  @apply text-lg font-bold text-gray-900 dark:text-gray-100;
}

.stat-label {
  @apply text-xs text-gray-600 dark:text-gray-400;
}

.top-processes {
  @apply space-y-3;
}

.processes-section h4 {
  @apply text-sm font-medium text-gray-700 dark:text-gray-300 mb-2;
}

/* Disk Card Styles */
.disk-info {
  @apply text-sm text-gray-600 dark:text-gray-400;
}

.disk-list {
  @apply space-y-3;
}

.disk-item {
  @apply space-y-2;
}

.disk-item.high-usage {
  @apply border-l-4 border-orange-500 pl-2;
}

.disk-info {
  @apply flex items-center justify-between text-sm;
}

.disk-usage {
  @apply space-y-1;
}

.disk-bar {
  @apply h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden;
}

.disk-progress {
  @apply h-full transition-all duration-300;
}

.disk-progress {
  @apply bg-blue-500;
}

.disk-progress.critical {
  @apply bg-red-500;
}

.usage-percent {
  @apply text-xs font-medium;
}

.usage-text {
  @apply text-xs text-gray-600 dark:text-gray-400;
}

/* Network Card Styles */
.network-info {
  @apply text-sm text-gray-600 dark:text-gray-400;
}

.network-summary {
  @apply space-y-3;
}

.network-stats {
  @apply flex justify-between gap-4;
}

.network-stat {
  @apply flex items-center gap-2;
}

.stat-label {
  color: var(--primary-400);
  font-weight: bold;
}

.stat-value {
  color: var(--text-primary);
  font-size: 12px;
}

.network-interfaces {
  @apply space-y-2;
}

.network-interface {
  @apply flex items-center gap-2 p-2 rounded-md;
  @apply bg-gray-50 dark:bg-gray-700;
}

.interface-status {
  @apply w-2 h-2 rounded-full bg-gray-400;
}

.interface-status.active {
  @apply bg-green-500;
}

.interface-info {
  @apply flex-1 text-sm;
}

.interface-name {
  @apply font-medium text-gray-900 dark:text-gray-100;
}

.interface-ips {
  @apply text-gray-600 dark:text-gray-400 text-xs;
}

.interface-speed {
  @apply text-gray-500 text-xs;
}

/* Temperature Card Styles */
.temperature-info {
  @apply text-sm text-gray-600 dark:text-gray-400;
}

.temperature-list {
  @apply space-y-3;
}

.temperature-item {
  @apply p-3 rounded-md border;
  @apply border-gray-200 dark:border-gray-700;
}

.temperature-item.high {
  @apply border-yellow-500 bg-yellow-50 dark:bg-yellow-900/20;
}

.temperature-item.critical {
  @apply border-red-500 bg-red-50 dark:bg-red-900/20;
}

.sensor-info {
  @apply flex items-center justify-between mb-2;
}

.sensor-name {
  @apply text-sm font-medium text-gray-900 dark:text-gray-100;
}

.sensor-status {
  @apply px-2 py-1 rounded text-xs font-medium;
  @apply bg-gray-100 text-gray-600;
}

.sensor-status.normal {
  @apply bg-green-100 text-green-700;
}

.sensor-status.high {
  @apply bg-yellow-100 text-yellow-700;
}

.sensor-status.critical {
  @apply bg-red-100 text-red-700;
}

.temperature-display {
  @apply flex items-center justify-between;
}

.temperature-value {
  @apply text-lg font-bold;
}

.temperature-value.high {
  @apply text-yellow-600;
}

.temperature-value.critical {
  @apply text-red-600;
}

.temperature-range {
  @apply text-xs text-gray-600 dark:text-gray-400;
}

/* Footer Styles */
.monitor-footer {
  @apply flex items-center justify-between p-3 border-t border-gray-200 dark:border-gray-700;
  @apply bg-gray-50 dark:bg-gray-800 text-sm;
}

.last-updated {
  @apply flex items-center gap-2 text-gray-600 dark:text-gray-400;
}

.health-indicator {
  @apply flex items-center gap-3;
}

.health-status {
  @apply px-2 py-1 rounded text-xs font-medium;
  @apply uppercase;
}

.health-status.good {
  @apply bg-green-100 text-green-700;
}

.health-status.warning {
  @apply bg-yellow-100 text-yellow-700;
}

.health-status.critical {
  @apply bg-red-100 text-red-700;
}

.health-score {
  @apply text-gray-600 dark:text-gray-400;
}

/* Loading Overlay */
.loading-overlay {
  @apply absolute inset-0 bg-white/80 dark:bg-gray-900/80;
  @apply flex items-center justify-center;
  @apply z-50;
}

.loading-spinner {
  @apply w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full;
  @apply animate-spin;
}

/* Compact mode adjustments */
.system-monitor.compact .metrics-grid {
  @apply grid-cols-1;
}

.system-monitor.compact .card-content {
  @apply p-3;
}

.system-monitor.compact .card-header {
  @apply p-3;
}

.system-monitor.compact .monitor-header {
  @apply p-3;
}

.system-monitor.compact .monitor-footer {
  @apply p-2;
}

.system-monitor.compact .usage-circle {
  @apply w-12 h-12;
}

.system-monitor.compact .stat-value {
  @apply text-base;
}

.system-monitor.compact .process-stats {
  @apply grid-cols-2;
}

.system-monitor.compact .network-stats {
  @apply grid-cols-1;
}
</style>