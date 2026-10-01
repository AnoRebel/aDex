<template>
  <div class="system-page">
    <!-- System Header -->
    <header class="system-header">
      <div class="header-left">
        <h1 class="page-title">System Monitor</h1>
        <div class="system-info">
          <span class="info-item">
            <strong>Hostname:</strong> {{ systemInfo?.hostname || 'Unknown' }}
          </span>
          <span class="info-item">
            <strong>OS:</strong> {{ systemInfo?.platform || 'Unknown' }} {{ systemInfo?.osVersion || '' }}
          </span>
          <span class="info-item">
            <strong>Uptime:</strong> {{ formatUptime(systemInfo?.uptime || 0) }}
          </span>
        </div>
      </div>
      <div class="header-right">
        <div class="refresh-controls">
          <button 
            class="refresh-btn"
            :class="{ active: autoRefresh }"
            @click="toggleAutoRefresh"
          >
            {{ autoRefresh ? '⏸️' : '▶️' }} Auto Refresh
          </button>
          <select v-model="refreshRate" @change="setRefreshRate(Number(refreshRate))">
            <option value="1000">1s</option>
            <option value="2000">2s</option>
            <option value="5000">5s</option>
            <option value="10000">10s</option>
          </select>
          <button class="refresh-btn" @click="refreshData">
            🔄 Refresh
          </button>
        </div>
        <div class="performance-status" :class="performanceStatus">
          <span class="status-indicator"></span>
          {{ performanceStatus.toUpperCase() }}
        </div>
      </div>
    </header>

    <!-- System Overview Cards -->
    <section class="overview-section">
      <div class="overview-grid">
        <!-- CPU Card -->
        <div class="overview-card cpu-card">
          <div class="card-header">
            <h3>CPU Usage</h3>
            <div class="card-icon">🖥️</div>
          </div>
          <div class="card-content">
            <div class="metric-display">
              <span class="metric-value">{{ cpuUsage?.usage.toFixed(1) || 0 }}%</span>
              <div class="metric-chart">
                <div 
                  class="chart-fill cpu-fill"
                  :style="{ width: (cpuUsage?.usage || 0) + '%' }"
                ></div>
              </div>
            </div>
            <div class="metric-details">
              <div class="detail-row">
                <span>Cores:</span>
                <span>{{ cpuUsage?.cores?.length || 0 }}</span>
              </div>
              <div class="detail-row">
                <span>Temperature:</span>
                <span>{{ formatTemperature(cpuUsage?.temperature || 0) }}</span>
              </div>
              <div class="detail-row">
                <span>Load Average:</span>
                <span>{{ formatLoadAverage(cpuUsage?.loadAverage || []) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Memory Card -->
        <div class="overview-card memory-card">
          <div class="card-header">
            <h3>Memory Usage</h3>
            <div class="card-icon">💾</div>
          </div>
          <div class="card-content">
            <div class="metric-display">
              <span class="metric-value">{{ memoryUsage?.usage.toFixed(1) || 0 }}%</span>
              <div class="metric-chart">
                <div 
                  class="chart-fill memory-fill"
                  :style="{ width: (memoryUsage?.usage || 0) + '%' }"
                ></div>
              </div>
            </div>
            <div class="metric-details">
              <div class="detail-row">
                <span>Used:</span>
                <span>{{ formatBytes(memoryUsage?.used || 0) }}</span>
              </div>
              <div class="detail-row">
                <span>Total:</span>
                <span>{{ formatBytes(memoryUsage?.total || 0) }}</span>
              </div>
              <div class="detail-row">
                <span>Available:</span>
                <span>{{ formatBytes(memoryUsage?.available || 0) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Processes Card -->
        <div class="overview-card processes-card">
          <div class="card-header">
            <h3>Processes</h3>
            <div class="card-icon">⚙️</div>
          </div>
          <div class="card-content">
            <div class="metric-display">
              <span class="metric-value">{{ totalProcesses }}</span>
              <div class="metric-label">Total</div>
            </div>
            <div class="metric-details">
              <div class="detail-row">
                <span>Running:</span>
                <span>{{ runningProcesses }}</span>
              </div>
              <div class="detail-row">
                <span>Sleeping:</span>
                <span>{{ totalProcesses - runningProcesses }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Disk Card -->
        <div class="overview-card disk-card">
          <div class="card-header">
            <h3>Disk Usage</h3>
            <div class="card-icon">💿</div>
          </div>
          <div class="card-content">
            <div class="disk-list">
              <div 
                v-for="disk in diskUsage?.slice(0, 3)" 
                :key="disk.device"
                class="disk-item"
              >
                <div class="disk-info">
                  <span class="disk-name">{{ disk.mountpoint }}</span>
                  <span class="disk-usage">{{ disk.usage.toFixed(1) }}%</span>
                </div>
                <div class="disk-chart">
                  <div 
                    class="chart-fill disk-fill"
                    :style="{ width: disk.usage + '%' }"
                  ></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Detailed Sections -->
    <section class="detailed-section">
      <!-- CPU Details -->
      <div class="detail-panel cpu-panel">
        <div class="panel-header">
          <h3>CPU Details</h3>
          <button class="expand-btn" @click="togglePanel('cpu')">
            {{ expandedPanels.cpu ? '▼' : '▶' }}
          </button>
        </div>
        <div v-show="expandedPanels.cpu" class="panel-content">
          <div class="cpu-cores">
            <div 
              v-for="core in cpuUsage?.cores" 
              :key="core.id"
              class="core-item"
            >
              <span class="core-label">Core {{ core.id }}</span>
              <div class="core-usage">
                <div class="usage-bar">
                  <div 
                    class="usage-fill"
                    :style="{ width: core.usage + '%' }"
                  ></div>
                </div>
                <span class="usage-text">{{ core.usage.toFixed(1) }}%</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Process List -->
      <div class="detail-panel processes-panel">
        <div class="panel-header">
          <h3>Processes</h3>
          <div class="panel-controls">
            <input 
              v-model="processSearch"
              type="text"
              placeholder="Search processes..."
              class="search-input"
            />
            <select v-model="sortBy" @change="updateSorting">
              <option value="name">Name</option>
              <option value="cpu">CPU</option>
              <option value="memory">Memory</option>
              <option value="pid">PID</option>
            </select>
            <button class="expand-btn" @click="togglePanel('processes')">
              {{ expandedPanels.processes ? '▼' : '▶' }}
            </button>
          </div>
        </div>
        <div v-show="expandedPanels.processes" class="panel-content">
          <div class="process-table">
            <div class="table-header">
              <div class="header-cell">PID</div>
              <div class="header-cell">Name</div>
              <div class="header-cell">CPU</div>
              <div class="header-cell">Memory</div>
              <div class="header-cell">Status</div>
              <div class="header-cell">Actions</div>
            </div>
            <div class="table-body">
              <div 
                v-for="process in filteredProcesses" 
                :key="process.pid"
                class="table-row"
              >
                <div class="table-cell">{{ process.pid }}</div>
                <div class="table-cell process-name">{{ process.name }}</div>
                <div class="table-cell">{{ process.cpu.toFixed(1) }}%</div>
                <div class="table-cell">{{ formatBytes(process.memory) }}</div>
                <div class="table-cell">
                  <span 
                    class="status-badge"
                    :class="process.status"
                  >
                    {{ process.status }}
                  </span>
                </div>
                <div class="table-cell">
                  <button 
                    class="action-btn kill-btn"
                    @click="killProcess(process.pid)"
                    title="Kill Process"
                  >
                    ⚠️
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- System Commands -->
      <div class="detail-panel commands-panel">
        <div class="panel-header">
          <h3>System Commands</h3>
          <button class="expand-btn" @click="togglePanel('commands')">
            {{ expandedPanels.commands ? '▼' : '▶' }}
          </button>
        </div>
        <div v-show="expandedPanels.commands" class="panel-content">
          <div class="command-interface">
            <div class="command-input-group">
              <input 
                v-model="systemCommand"
                type="text"
                placeholder="Enter system command..."
                class="command-input"
                @keydown.enter="executeSystemCommand"
              />
              <button 
                class="execute-btn"
                :disabled="!systemCommand.trim() || isCommandRunning"
                @click="executeSystemCommand"
              >
                Execute
              </button>
            </div>
            <div class="quick-commands">
              <button 
                v-for="cmd in quickCommands" 
                :key="cmd.name"
                class="quick-cmd-btn"
                @click="systemCommand = cmd.command"
              >
                {{ cmd.name }}
              </button>
            </div>
            <div v-if="commandResult" class="command-result">
              <div class="result-header">
                <span>Command: {{ commandResult.command }}</span>
                <span 
                  class="exit-code"
                  :class="{ success: commandResult.exitCode === 0 }"
                >
                  Exit: {{ commandResult.exitCode }}
                </span>
              </div>
              <div class="result-output">
                <pre>{{ commandResult.output }}</pre>
              </div>
              <div v-if="commandResult.error" class="result-error">
                <pre>{{ commandResult.error }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Alerts Section -->
    <section v-if="unacknowledgedAlerts.length > 0" class="alerts-section">
      <div class="alerts-header">
        <h3>System Alerts</h3>
        <button class="acknowledge-all-btn" @click="acknowledgeAllAlerts">
          Acknowledge All
        </button>
      </div>
      <div class="alerts-list">
        <div 
          v-for="alert in unacknowledgedAlerts" 
          :key="alert.id"
          class="alert-item"
          :class="alert.type"
        >
          <div class="alert-icon">
            {{ alert.type === 'error' ? '🚨' : alert.type === 'warning' ? '⚠️' : 'ℹ️' }}
          </div>
          <div class="alert-content">
            <div class="alert-title">{{ alert.title }}</div>
            <div class="alert-message">{{ alert.message }}</div>
            <div class="alert-time">{{ formatTime(alert.timestamp) }}</div>
          </div>
          <button class="alert-close" @click="acknowledgeAlert(alert.id)">
            ×
          </button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useSystem } from '~/composables/useSystem'
import type { Process } from '~/types/system'

// Composables
const system = useSystem()

// Local state
const autoRefresh = ref(true)
const refreshRate = ref(2000)
const expandedPanels = ref({
  cpu: true,
  processes: true,
  commands: false,
})
const processSearch = ref('')
const sortBy = ref('cpu')
const sortOrder = ref<'asc' | 'desc'>('desc')
const systemCommand = ref('')
const commandResult = ref<any>(null)
const isCommandRunning = ref(false)

const quickCommands = [
  { name: 'System Info', command: 'uname -a' },
  { name: 'Disk Space', command: 'df -h' },
  { name: 'Memory Info', command: 'free -h' },
  { name: 'CPU Info', command: 'lscpu' },
  { name: 'Network Status', command: 'ip addr show' },
  { name: 'Running Services', command: 'systemctl list-units --type=service --state=running' },
]

// Computed properties
const systemInfo = computed(() => system.systemInfo.value)
const systemStats = computed(() => system.systemStats.value)
const cpuUsage = computed(() => system.cpuUsage.value)
const memoryUsage = computed(() => system.memoryUsage.value)
const diskUsage = computed(() => system.diskUsage.value)
const totalProcesses = computed(() => system.totalProcesses.value)
const runningProcesses = computed(() => system.runningProcesses.value)
const performanceStatus = computed(() => system.getPerformanceStatus.value)
const unacknowledgedAlerts = computed(() => system.unacknowledgedAlerts.value)

const filteredProcesses = computed(() => {
  let processes = system.processes.value

  // Apply search filter
  if (processSearch.value.trim()) {
    const query = processSearch.value.toLowerCase()
    processes = processes.filter(process => 
      process.name.toLowerCase().includes(query) ||
      process.command.toLowerCase().includes(query) ||
      process.user.toLowerCase().includes(query) ||
      process.pid.toString().includes(query)
    )
  }

  // Apply sorting
  return processes.sort((a, b) => {
    let comparison = 0
    
    switch (sortBy.value) {
      case 'name':
        comparison = a.name.localeCompare(b.name)
        break
      case 'cpu':
        comparison = b.cpu - a.cpu
        break
      case 'memory':
        comparison = b.memory - a.memory
        break
      case 'pid':
        comparison = a.pid - b.pid
        break
    }
    
    return sortOrder.value === 'asc' ? comparison : -comparison
  }).slice(0, 50) // Limit to top 50 processes
})

// Methods
const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    system.startAutoRefresh(refreshRate.value)
  } else {
    system.stopAutoRefresh()
  }
}

const setRefreshRate = (rate: number) => {
  refreshRate.value = rate
  if (autoRefresh.value) {
    system.setRefreshRate(rate)
  }
}

const refreshData = async () => {
  await Promise.all([
    system.loadSystemInfo(),
    system.loadSystemStats(),
    system.loadProcesses(),
  ])
}

const togglePanel = (panel: keyof typeof expandedPanels.value) => {
  expandedPanels.value[panel] = !expandedPanels.value[panel]
}

const updateSorting = () => {
  // Sorting is handled by the computed property
}

const killProcess = async (pid: number) => {
  if (confirm(`Are you sure you want to kill process ${pid}?`)) {
    const success = await system.killProcess(pid)
    if (success) {
      system.addAlert({
        type: 'info',
        title: 'Process Killed',
        message: `Process ${pid} has been terminated`,
        timestamp: new Date(),
        acknowledged: false,
      })
    } else {
      system.addAlert({
        type: 'error',
        title: 'Failed to Kill Process',
        message: `Could not terminate process ${pid}`,
        timestamp: new Date(),
        acknowledged: false,
      })
    }
  }
}

const executeSystemCommand = async () => {
  if (!systemCommand.value.trim() || isCommandRunning.value) return

  isCommandRunning.value = true
  try {
    const result = await system.executeCommand(systemCommand.value)
    commandResult.value = {
      command: systemCommand.value,
      ...result,
      timestamp: new Date(),
    }
  } catch (error) {
    commandResult.value = {
      command: systemCommand.value,
      success: false,
      output: '',
      error: error instanceof Error ? error.message : 'Unknown error',
      exitCode: -1,
      timestamp: new Date(),
    }
  } finally {
    isCommandRunning.value = false
  }
}

const acknowledgeAlert = (alertId: string) => {
  system.acknowledgeAlert(alertId)
}

const acknowledgeAllAlerts = () => {
  system.acknowledgeAllAlerts()
}

// Utility functions
const formatBytes = (bytes: number): string => {
  return system.formatBytes(bytes)
}

const formatTemperature = (celsius: number): string => {
  return system.formatTemperature(celsius)
}

const formatUptime = (seconds: number): string => {
  return system.formatUptime(seconds)
}

const formatLoadAverage = (loadAverage: number[]): string => {
  return loadAverage.map(l => l.toFixed(2)).join(', ')
}

const formatTime = (date: Date): string => {
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

// Lifecycle
onMounted(async () => {
  await system.initialize()
  if (autoRefresh.value) {
    system.startAutoRefresh(refreshRate.value)
  }
})

onUnmounted(() => {
  system.stopAutoRefresh()
})

// SEO
useHead({
  title: 'System Monitor - aDex',
  meta: [
    { name: 'description', content: 'Real-time system monitoring with CPU, memory, and process tracking' }
  ]
})
</script>

<style scoped>
.system-page {
  padding: 1rem;
  max-width: 1600px;
  margin: 0 auto;
  background: var(--color-background);
  min-height: 100vh;
}

.system-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  gap: 2rem;
}

.header-left {
  flex: 1;
}

.page-title {
  margin: 0 0 1rem 0;
  color: var(--color-text);
  font-size: 2rem;
}

.system-info {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}

.info-item {
  color: var(--color-text-secondary);
}

.info-item strong {
  color: var(--color-text);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.refresh-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.refresh-btn {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-btn:hover,
.refresh-btn.active {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.refresh-controls select {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
}

.performance-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  font-weight: bold;
}

.performance-status.good {
  background: var(--color-success);
  color: var(--color-background);
}

.performance-status.moderate {
  background: var(--color-warning);
  color: var(--color-background);
}

.performance-status.warning {
  background: var(--color-warning);
  color: var(--color-background);
}

.performance-status.critical {
  background: var(--color-error);
  color: var(--color-background);
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse 2s infinite;
}

.overview-section {
  margin-bottom: 2rem;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.overview-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 1rem;
  padding: 1.5rem;
  transition: transform 0.2s ease;
}

.overview-card:hover {
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header h3 {
  margin: 0;
  color: var(--color-text);
}

.card-icon {
  font-size: 1.5rem;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.metric-display {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.metric-value {
  font-size: 2.5rem;
  font-weight: bold;
  color: var(--color-primary);
  min-width: 100px;
}

.metric-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.metric-chart {
  flex: 1;
  height: 12px;
  background: var(--color-background);
  border-radius: 6px;
  overflow: hidden;
}

.chart-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.cpu-fill {
  background: linear-gradient(90deg, var(--color-primary), var(--color-accent));
}

.memory-fill {
  background: linear-gradient(90deg, var(--color-info), var(--color-warning));
}

.disk-fill {
  background: linear-gradient(90deg, var(--color-warning), var(--color-error));
}

.metric-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.detail-row span:first-child {
  color: var(--color-text-secondary);
}

.detail-row span:last-child {
  color: var(--color-text);
  font-weight: 500;
}

.disk-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.disk-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.disk-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.disk-name {
  color: var(--color-text);
}

.disk-usage {
  color: var(--color-primary);
  font-weight: bold;
}

.disk-chart {
  height: 6px;
  background: var(--color-background);
  border-radius: 3px;
  overflow: hidden;
}

.detailed-section {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.detail-panel {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 1rem;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: var(--color-background);
  border-bottom: 1px solid var(--color-border);
}

.panel-header h3 {
  margin: 0;
  color: var(--color-text);
}

.panel-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.search-input {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
  min-width: 200px;
}

.panel-controls select {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem;
  border-radius: 0.25rem;
}

.expand-btn {
  background: none;
  border: none;
  color: var(--color-text);
  cursor: pointer;
  padding: 0.5rem;
  font-size: 1rem;
}

.panel-content {
  padding: 1.5rem;
}

.cpu-cores {
  display: grid;
  gap: 1rem;
}

.core-item {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.core-label {
  min-width: 80px;
  color: var(--color-text);
}

.core-usage {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.usage-bar {
  flex: 1;
  height: 8px;
  background: var(--color-background);
  border-radius: 4px;
  overflow: hidden;
}

.usage-fill {
  height: 100%;
  background: var(--color-primary);
  transition: width 0.3s ease;
}

.usage-text {
  min-width: 50px;
  text-align: right;
  color: var(--color-text);
  font-weight: 500;
}

.process-table {
  overflow-x: auto;
}

.table-header,
.table-row {
  display: grid;
  grid-template-columns: 60px 200px 80px 100px 100px 80px;
  gap: 1rem;
  align-items: center;
  padding: 0.75rem 0;
}

.table-header {
  border-bottom: 1px solid var(--color-border);
  font-weight: bold;
  color: var(--color-text-secondary);
}

.table-row {
  border-bottom: 1px solid var(--color-border);
}

.table-row:hover {
  background: var(--color-background);
}

.table-cell {
  color: var(--color-text);
  font-size: 0.875rem;
}

.process-name {
  font-family: monospace;
  font-weight: 500;
}

.status-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: bold;
  text-transform: uppercase;
}

.status-badge.running {
  background: var(--color-success);
  color: var(--color-background);
}

.status-badge.sleeping {
  background: var(--color-info);
  color: var(--color-background);
}

.status-badge.stopped {
  background: var(--color-warning);
  color: var(--color-background);
}

.status-badge.zombie {
  background: var(--color-error);
  color: var(--color-background);
}

.action-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  font-size: 1rem;
}

.kill-btn:hover {
  color: var(--color-error);
}

.command-interface {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.command-input-group {
  display: flex;
  gap: 1rem;
}

.command-input {
  flex: 1;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.75rem;
  border-radius: 0.25rem;
  font-family: monospace;
}

.execute-btn {
  background: var(--color-primary);
  border: none;
  color: var(--color-background);
  padding: 0.75rem 1.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-weight: bold;
  transition: background-color 0.2s ease;
}

.execute-btn:hover:not(:disabled) {
  background: var(--color-accent);
}

.execute-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.quick-commands {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.quick-cmd-btn {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.quick-cmd-btn:hover {
  background: var(--color-primary);
  color: var(--color-background);
  border-color: var(--color-primary);
}

.command-result {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  overflow: hidden;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  font-family: monospace;
  font-size: 0.875rem;
}

.exit-code {
  font-weight: bold;
}

.exit-code.success {
  color: var(--color-success);
}

.exit-code:not(.success) {
  color: var(--color-error);
}

.result-output,
.result-error {
  padding: 1rem;
  font-family: monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  overflow-x: auto;
}

.result-error {
  background: rgba(255, 0, 0, 0.1);
  color: var(--color-error);
}

.alerts-section {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 1rem;
  padding: 1.5rem;
}

.alerts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.alerts-header h3 {
  margin: 0;
  color: var(--color-text);
}

.acknowledge-all-btn {
  background: var(--color-primary);
  border: none;
  color: var(--color-background);
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.875rem;
}

.alerts-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.alert-item {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem;
  border-radius: 0.5rem;
  border-left: 4px solid;
}

.alert-item.error {
  background: rgba(255, 0, 0, 0.1);
  border-left-color: var(--color-error);
}

.alert-item.warning {
  background: rgba(255, 170, 0, 0.1);
  border-left-color: var(--color-warning);
}

.alert-item.info {
  background: rgba(0, 153, 255, 0.1);
  border-left-color: var(--color-info);
}

.alert-icon {
  font-size: 1.5rem;
}

.alert-content {
  flex: 1;
}

.alert-title {
  font-weight: bold;
  color: var(--color-text);
  margin-bottom: 0.25rem;
}

.alert-message {
  color: var(--color-text-secondary);
  margin-bottom: 0.5rem;
}

.alert-time {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.alert-close {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 1.25rem;
  padding: 0;
}

.alert-close:hover {
  color: var(--color-text);
}

/* Responsive */
@media (max-width: 1024px) {
  .system-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .overview-grid {
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  }

  .table-header,
  .table-row {
    grid-template-columns: 60px 150px 60px 80px 80px 60px;
    font-size: 0.75rem;
  }

  .command-input-group {
    flex-direction: column;
  }
}

/* Animations */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
