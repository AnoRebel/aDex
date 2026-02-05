<template>
  <div class="system-monitor">
    <!-- Header -->
    <div class="monitor-header">
      <h3 class="glitch">SYSTEM MONITOR</h3>
      <div class="monitor-status" :class="healthClass">
        <div class="status-dot"></div>
        <span>{{ isMonitoring ? 'ACTIVE' : 'PAUSED' }}</span>
      </div>
    </div>

    <!-- Error display -->
    <div v-if="error" class="error-container">
      <div class="error-message">
        <span class="error-icon">⚠️</span>
        <span>{{ error }}</span>
      </div>
      <button @click="refreshMetrics" class="retry-button">Retry</button>
    </div>

    <!-- Main metrics -->
    <div v-else class="metrics-content">
      <!-- CPU Section -->
      <div class="metric-card cpu-card">
        <div class="card-header">
          <span class="card-icon">🔲</span>
          <span class="card-title">CPU</span>
          <span class="card-value" :class="cpuClass">{{ cpuPercent.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill cpu" :style="{ width: `${cpuPercent}%` }"></div>
        </div>
        <div class="card-details">
          <span class="detail-item">Cores: {{ coreCount }}</span>
          <span class="detail-item" v-if="cpuFrequency">{{ (cpuFrequency / 1000).toFixed(2) }} GHz</span>
        </div>
        <!-- Per-core usage -->
        <div v-if="cpuCores.length > 0" class="cores-grid">
          <div 
            v-for="(coreUsage, index) in cpuCores.slice(0, 16)" 
            :key="index" 
            class="core-item"
            :title="`Core ${index}: ${coreUsage.toFixed(1)}%`"
          >
            <div class="core-bar">
              <div 
                class="core-fill" 
                :style="{ height: `${coreUsage}%` }"
                :class="getCoreClass(coreUsage)"
              ></div>
            </div>
            <span class="core-label">{{ index }}</span>
          </div>
        </div>
        <div v-if="cpuModelName" class="cpu-model">{{ cpuModelName }}</div>
      </div>

      <!-- Memory Section -->
      <div class="metric-card memory-card">
        <div class="card-header">
          <span class="card-icon">💾</span>
          <span class="card-title">Memory</span>
          <span class="card-value" :class="memoryClass">{{ memoryPercent.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill memory" :style="{ width: `${memoryPercent}%` }"></div>
        </div>
        <div class="card-details">
          <span class="detail-item">Used: {{ formatBytes(memoryUsed) }}</span>
          <span class="detail-item">Total: {{ formatBytes(memoryTotal) }}</span>
        </div>
      </div>

      <!-- Swap Section -->
      <div class="metric-card swap-card">
        <div class="card-header">
          <span class="card-icon">💿</span>
          <span class="card-title">Swap</span>
          <span class="card-value">{{ swapPercent.toFixed(1) }}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill swap" :style="{ width: `${swapPercent}%` }"></div>
        </div>
      </div>

      <!-- Process Summary -->
      <div class="metric-card process-card">
        <div class="card-header">
          <span class="card-icon">📊</span>
          <span class="card-title">Processes</span>
        </div>
        <div class="process-stats">
          <div class="process-stat">
            <span class="stat-value">{{ totalProcesses }}</span>
            <span class="stat-label">Total</span>
          </div>
          <div class="process-stat">
            <span class="stat-value running">{{ runningProcessCount }}</span>
            <span class="stat-label">Running</span>
          </div>
          <div class="process-stat">
            <span class="stat-value sleeping">{{ sleepingProcessCount }}</span>
            <span class="stat-label">Sleeping</span>
          </div>
        </div>
      </div>

      <!-- Top Processes -->
      <div class="metric-card top-processes-card">
        <div class="card-header">
          <span class="card-icon">🔝</span>
          <span class="card-title">Top Processes</span>
        </div>
        <div class="process-list">
          <div
            v-for="(process, index) in topProcesses.slice(0, 5)"
            :key="index"
            class="process-item"
          >
            <span class="process-name">{{ process.name }}</span>
            <div class="process-bars">
              <div class="mini-bar cpu">
                <div class="mini-fill" :style="{ width: `${process.cpu}%` }"></div>
              </div>
              <span class="process-cpu">{{ process.cpu.toFixed(1) }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- System Info -->
      <div class="metric-card info-card">
        <div class="card-header">
          <span class="card-icon">ℹ️</span>
          <span class="card-title">System</span>
        </div>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">OS</span>
            <span class="info-value">{{ osName }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Hostname</span>
            <span class="info-value">{{ hostname }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Uptime</span>
            <span class="info-value">{{ formatUptime(uptime) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Arch</span>
            <span class="info-value">{{ architecture }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="monitor-footer">
      <div class="footer-left">
        <span class="last-update">Last updated: {{ lastUpdateText }}</span>
      </div>
      <div class="footer-right">
        <button @click="toggleMonitoring" class="control-btn" :class="{ active: isMonitoring }">
          {{ isMonitoring ? '⏸ Pause' : '▶ Resume' }}
        </button>
        <button @click="refreshMetrics" class="control-btn refresh" :disabled="isLoading">
          ↻ Refresh
        </button>
      </div>
    </div>

    <!-- Loading overlay -->
    <div v-if="isLoading" class="loading-overlay">
      <div class="loading-spinner"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useSystemStore } from '~/stores/system'

// Store
const systemStore = useSystemStore()

// Local state
const lastUpdate = ref<Date>(new Date())
const refreshInterval = ref<NodeJS.Timeout | null>(null)

// Computed from store
const isLoading = computed(() => systemStore.isLoading)
const error = computed(() => systemStore.error)
const isMonitoring = computed(() => systemStore.isMonitoring)
const systemData = computed(() => systemStore.systemData)
const topProcesses = computed(() => systemStore.topProcesses)

// CPU metrics
const cpuPercent = computed(() => systemData.value?.cpu?.usage || 0)
const cpuCores = computed(() => systemData.value?.cpu?.cores || [])
const coreCount = computed(() => systemData.value?.cpu?.coreCount || cpuCores.value.length || 0)
const cpuModelName = computed(() => systemData.value?.cpu?.modelName || '')
const cpuFrequency = computed(() => systemData.value?.cpu?.frequency || 0)

// Helper for per-core coloring
const getCoreClass = (usage: number): string => {
  if (usage > 90) return 'critical'
  if (usage > 75) return 'warning'
  return 'good'
}

// Memory metrics
const memoryPercent = computed(() => systemData.value?.memory?.usagePercent || 0)
const memoryUsed = computed(() => systemData.value?.memory?.used || 0)
const memoryTotal = computed(() => systemData.value?.memory?.total || 0)

// Swap metrics
const swapPercent = computed(() => systemData.value?.swap?.usagePercent || 0)

// Process metrics
const totalProcesses = computed(() => systemData.value?.processes?.length || 0)
const runningProcessCount = computed(() => 
  systemData.value?.processes?.filter((p: any) => p.status === 'running').length || 0
)
const sleepingProcessCount = computed(() => 
  systemData.value?.processes?.filter((p: any) => p.status === 'sleeping' || p.status !== 'running').length || 0
)

// System info
const osName = computed(() => systemData.value?.os || 'Unknown')
const hostname = computed(() => systemData.value?.hostname || 'localhost')
const uptime = computed(() => systemData.value?.uptime || 0)
const architecture = computed(() => systemData.value?.architecture || 'x64')
const loadAvg = computed(() => systemData.value?.loadAvg || '0.00')

// Health class
const healthClass = computed(() => {
  if (cpuPercent.value > 90 || memoryPercent.value > 90) return 'critical'
  if (cpuPercent.value > 75 || memoryPercent.value > 75) return 'warning'
  return 'good'
})

const cpuClass = computed(() => {
  if (cpuPercent.value > 90) return 'critical'
  if (cpuPercent.value > 75) return 'warning'
  return 'good'
})

const memoryClass = computed(() => {
  if (memoryPercent.value > 90) return 'critical'
  if (memoryPercent.value > 75) return 'warning'
  return 'good'
})

const lastUpdateText = computed(() => {
  const now = new Date()
  const diff = Math.floor((now.getTime() - lastUpdate.value.getTime()) / 1000)
  if (diff < 60) return `${diff}s ago`
  return `${Math.floor(diff / 60)}m ago`
})

// Methods
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

const formatUptime = (seconds: number): string => {
  if (!seconds) return '0m'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

const refreshMetrics = async () => {
  lastUpdate.value = new Date()
  await systemStore.fetchSystemStats()
}

const toggleMonitoring = () => {
  if (isMonitoring.value) {
    systemStore.stopMonitoring()
  } else {
    systemStore.startMonitoring()
  }
}

// Lifecycle
onMounted(() => {
  systemStore.startMonitoring()
  
  // Set up auto-refresh
  refreshInterval.value = setInterval(() => {
    lastUpdate.value = new Date()
  }, 2000)
})

onUnmounted(() => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
  }
})
</script>

<style scoped>
.system-monitor {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: rgba(10, 10, 10, 0.95);
  border-radius: 8px;
  font-family: 'Fira Code', monospace;
  color: var(--text-primary);
  overflow: hidden;
  position: relative;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid var(--surface-border);
  background: var(--surface);
}

.monitor-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 10px;
  font-weight: bold;
  text-transform: uppercase;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.monitor-status.good {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.monitor-status.good .status-dot {
  background: #10b981;
}

.monitor-status.warning {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.monitor-status.warning .status-dot {
  background: #f59e0b;
}

.monitor-status.critical {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.monitor-status.critical .status-dot {
  background: #ef4444;
}

.error-container {
  padding: 15px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  margin: 15px;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ef4444;
  font-size: 12px;
}

.retry-button {
  padding: 6px 12px;
  background: #ef4444;
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.retry-button:hover {
  background: #dc2626;
}

.metrics-content {
  flex: 1;
  padding: 15px;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.metric-card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  transition: all 0.2s ease;
}

.metric-card:hover {
  border-color: var(--primary-500);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.card-icon {
  font-size: 16px;
}

.card-title {
  flex: 1;
  font-size: 12px;
  font-weight: bold;
  color: var(--primary-400);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.card-value {
  font-size: 18px;
  font-weight: bold;
}

.card-value.good {
  color: #10b981;
}

.card-value.warning {
  color: #f59e0b;
}

.card-value.critical {
  color: #ef4444;
}

.progress-bar {
  height: 6px;
  background: var(--surface-elevated);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 10px;
}

.progress-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
}

.progress-fill.cpu {
  background: linear-gradient(90deg, #0ea5e9, #38bdf8);
}

.progress-fill.memory {
  background: linear-gradient(90deg, #10b981, #34d399);
}

.progress-fill.swap {
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

.card-details {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--text-secondary);
}

/* Per-core CPU display */
.cores-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(20px, 1fr));
  gap: 4px;
  margin-top: 10px;
  padding: 8px;
  background: var(--surface-elevated);
  border-radius: 6px;
}

.core-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.core-bar {
  width: 14px;
  height: 30px;
  background: var(--surface);
  border-radius: 3px;
  overflow: hidden;
  display: flex;
  align-items: flex-end;
}

.core-fill {
  width: 100%;
  border-radius: 2px;
  transition: height 0.3s ease;
}

.core-fill.good {
  background: linear-gradient(180deg, #10b981, #34d399);
}

.core-fill.warning {
  background: linear-gradient(180deg, #f59e0b, #fbbf24);
}

.core-fill.critical {
  background: linear-gradient(180deg, #ef4444, #f87171);
}

.core-label {
  font-size: 8px;
  color: var(--text-secondary);
}

.cpu-model {
  margin-top: 8px;
  font-size: 9px;
  color: var(--text-secondary);
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.process-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  text-align: center;
}

.process-stat {
  padding: 8px;
  background: var(--surface-elevated);
  border-radius: 6px;
}

.stat-value {
  display: block;
  font-size: 18px;
  font-weight: bold;
  color: var(--text-primary);
}

.stat-value.running {
  color: #10b981;
}

.stat-value.sleeping {
  color: #0ea5e9;
}

.stat-label {
  font-size: 9px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.top-processes-card {
  grid-column: span 2;
}

.process-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.process-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  background: var(--surface-elevated);
  border-radius: 6px;
  font-size: 11px;
}

.process-name {
  flex: 1;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 150px;
}

.process-bars {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mini-bar {
  width: 60px;
  height: 4px;
  background: var(--surface);
  border-radius: 2px;
  overflow: hidden;
}

.mini-fill {
  height: 100%;
  background: linear-gradient(90deg, #8b5cf6, #a78bfa);
  border-radius: 2px;
}

.process-cpu {
  min-width: 45px;
  text-align: right;
  color: var(--accent-400);
  font-weight: bold;
}

.info-card {
  grid-column: span 2;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: 9px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 12px;
  color: var(--text-primary);
}

.monitor-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  border-top: 1px solid var(--surface-border);
  background: var(--surface);
}

.footer-left {
  font-size: 10px;
  color: var(--text-secondary);
}

.footer-right {
  display: flex;
  gap: 8px;
}

.control-btn {
  padding: 6px 12px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.control-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.control-btn.active {
  background: var(--primary-500);
  border-color: var(--primary-500);
}

.control-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.loading-spinner {
  width: 30px;
  height: 30px;
  border: 3px solid var(--surface-border);
  border-top-color: var(--primary-500);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Responsive */
@media (max-width: 600px) {
  .metrics-content {
    grid-template-columns: 1fr;
  }
  
  .top-processes-card,
  .info-card {
    grid-column: span 1;
  }
}

/* Glitch effect */
.glitch {
  position: relative;
  color: var(--primary-400);
  font-size: 14px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 2px;
}
</style>
