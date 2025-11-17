<template>
  <div class="network-dashboard">
    <div class="dashboard-header">
      <h2 class="dashboard-title">
        <Icon name="wifi" class="title-icon" />
        Network Activity Monitor
      </h2>
      <div class="dashboard-controls">
        <button
          @click="toggleMonitoring"
          :class="['monitoring-btn', { active: isMonitoring }]"
          :disabled="loading"
        >
          <Icon :name="isMonitoring ? 'pause' : 'play'" />
          {{ isMonitoring ? 'Pause' : 'Start' }} Monitoring
        </button>
        <button @click="refreshData" :disabled="loading" class="refresh-btn">
          <Icon name="refresh" :class="{ spinning: loading }" />
          Refresh
        </button>
        <button @click="showSettings = true" class="settings-btn">
          <Icon name="settings" />
        </button>
      </div>
    </div>

    <!-- Network Overview Cards -->
    <div class="overview-cards">
      <div class="overview-card interfaces-card">
        <div class="card-header">
          <h3>Network Interfaces</h3>
          <Icon name="network" />
        </div>
        <div class="card-content">
          <div class="metric-value">{{ activeInterfacesCount }}</div>
          <div class="metric-label">Active Interfaces</div>
          <div class="metric-change" :class="{ positive: interfaceChange >= 0, negative: interfaceChange < 0 }">
            <Icon :name="interfaceChange >= 0 ? 'trending-up' : 'trending-down'" />
            {{ Math.abs(interfaceChange) }}
          </div>
        </div>
      </div>

      <div class="overview-card bandwidth-card">
        <div class="card-header">
          <h3>Current Bandwidth</h3>
          <Icon name="activity" />
        </div>
        <div class="card-content">
          <div class="metric-value">{{ formatBandwidth(currentBandwidth) }}</div>
          <div class="metric-label">Total Usage</div>
          <div class="metric-change" :class="{ positive: bandwidthTrend >= 0, negative: bandwidthTrend < 0 }">
            <Icon :name="bandwidthTrend >= 0 ? 'trending-up' : 'trending-down'" />
            {{ formatBandwidth(Math.abs(bandwidthTrend)) }}
          </div>
        </div>
      </div>

      <div class="overview-card connections-card">
        <div class="card-header">
          <h3>Active Connections</h3>
          <Icon name="link" />
        </div>
        <div class="card-content">
          <div class="metric-value">{{ activeConnectionsCount }}</div>
          <div class="metric-label">Connections</div>
          <div class="metric-change" :class="{ positive: connectionsChange >= 0, negative: connectionsChange < 0 }">
            <Icon :name="connectionsChange >= 0 ? 'trending-up' : 'trending-down'" />
            {{ Math.abs(connectionsChange) }}
          </div>
        </div>
      </div>

      <div class="overview-card alerts-card">
        <div class="card-header">
          <h3>Network Alerts</h3>
          <Icon name="alert-triangle" />
        </div>
        <div class="card-content">
          <div class="metric-value">{{ activeAlertsCount }}</div>
          <div class="metric-label">Active Alerts</div>
          <div class="metric-change" :class="{ positive: alertsChange <= 0, negative: alertsChange > 0 }">
            <Icon :name="alertsChange <= 0 ? 'trending-down' : 'trending-up'" />
            {{ Math.abs(alertsChange) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="dashboard-grid">
      <!-- Bandwidth Chart -->
      <div class="grid-item bandwidth-chart">
        <NetworkBandwidthChart
          :data="bandwidthData"
          :loading="loading"
          @interface-selected="onInterfaceSelected"
        />
      </div>

      <!-- Network Interfaces -->
      <div class="grid-item interfaces-list">
        <NetworkInterfacesList
          :interfaces="networkMetrics?.interfaces || []"
          :selected-interface="selectedInterface"
          @interface-selected="onInterfaceSelected"
          @refresh="refreshData"
        />
      </div>

      <!-- Active Connections -->
      <div class="grid-item connections-list">
        <NetworkConnectionsList
          :connections="networkConnections"
          :loading="loading"
          @connection-selected="onConnectionSelected"
        />
      </div>

      <!-- Network Alerts -->
      <div class="grid-item alerts-panel">
        <NetworkAlertsPanel
          :alerts="networkAlerts"
          :loading="loading"
          @alert-resolved="onAlertResolved"
          @alerts-cleared="onAlertsCleared"
        />
      </div>

      <!-- Network Statistics -->
      <div class="grid-item statistics-panel">
        <NetworkStatisticsPanel
          :statistics="networkStatistics"
          :loading="loading"
        />
      </div>

      <!-- Connection Details -->
      <div class="grid-item connection-details" v-if="selectedConnection">
        <NetworkConnectionDetails
          :connection="selectedConnection"
          @close="selectedConnection = null"
        />
      </div>
    </div>

    <!-- Settings Modal -->
    <NetworkSettingsModal
      v-if="showSettings"
      :config="networkConfig"
      @save="onSettingsSave"
      @close="showSettings = false"
    />

    <!-- Status Bar -->
    <div class="status-bar">
      <div class="status-item">
        <Icon name="clock" />
        Last Update: {{ formatTime(lastUpdateTime) }}
      </div>
      <div class="status-item" :class="{ 'status-error': !isMonitoring }">
        <Icon :name="isMonitoring ? 'check-circle' : 'x-circle'" />
        {{ isMonitoring ? 'Monitoring Active' : 'Monitoring Paused' }}
      </div>
      <div class="status-item">
        <Icon name="cpu" />
        Update Rate: {{ updateInterval }}s
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useNetworkStore } from '~/stores/network'
import { useToast } from '~/composables/useToast'
import { ServiceCoordinator } from '~/wailsjs/go/backend/services/coordinator/ServiceCoordinator'

// Components
import NetworkBandwidthChart from './NetworkBandwidthChart.vue'
import NetworkInterfacesList from './NetworkInterfacesList.vue'
import NetworkConnectionsList from './NetworkConnectionsList.vue'
import NetworkAlertsPanel from './NetworkAlertsPanel.vue'
import NetworkStatisticsPanel from './NetworkStatisticsPanel.vue'
import NetworkConnectionDetails from './NetworkConnectionDetails.vue'
import NetworkSettingsModal from './NetworkSettingsModal.vue'

// Store and API
const networkStore = useNetworkStore()
const { toast } = useToast()
const serviceCoordinator = new ServiceCoordinator()

// Reactive data
const loading = ref(false)
const showSettings = ref(false)
const selectedInterface = ref<string>('')
const selectedConnection = ref<any>(null)
const updateInterval = ref(2)
let refreshTimer: NodeJS.Timeout | null = null

// Computed properties - combining store and direct API for robustness
const networkMetrics = computed(() => networkStore.metrics)
const networkConnections = computed(() => networkStore.connections)
const networkAlerts = computed(() => networkStore.alerts)
const networkStatistics = computed(() => networkStore.statistics)
const networkConfig = computed(() => networkStore.config)
const isMonitoring = computed(() => networkStore.isMonitoring)
const lastUpdateTime = computed(() => networkStore.lastUpdateTime)

// Overview metrics
const bandwidthData = computed(() => {
  if (!networkMetrics.value?.interfaces) return []
  // Return bandwidth data based on selected interface
  return networkMetrics.value.interfaces.map((iface: any) => ({
    name: iface.name,
    upload: iface.bytesSent || 0,
    download: iface.bytesRecv || 0,
    packetsSent: iface.packetsSent || 0,
    packetsRecv: iface.packetsRecv || 0,
    timestamp: new Date()
  }))
})

const activeInterfacesCount = computed(() => {
  if (!networkMetrics.value?.interfaces) return 0
  return networkMetrics.value.interfaces.filter((iface: any) => iface.isUp).length
})

const currentBandwidth = computed(() => {
  if (!bandwidthData.value || bandwidthData.value.length === 0) return 0
  const latest = bandwidthData.value[bandwidthData.value.length - 1]
  return (latest.upload + latest.download)
})

const activeConnectionsCount = computed(() => {
  if (!networkConnections.value) return 0
  return networkConnections.value.filter(conn => conn.state === 'ESTABLISHED').length
})

const activeAlertsCount = computed(() => {
  if (!networkAlerts.value) return 0
  return networkAlerts.value.filter(alert => !alert.resolved).length
})

// Trend calculations (simplified)
const interfaceChange = ref(0)
const bandwidthTrend = ref(0)
const connectionsChange = ref(0)
const alertsChange = ref(0)

// Methods
const formatBandwidth = (bytes: number): string => {
  if (bytes === 0) return '0 Mbps'
  const mbps = (bytes * 8) / 1024 / 1024
  return `${mbps.toFixed(1)} Mbps`
}

const formatTime = (date: Date): string => {
  if (!date) return 'Never'
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).format(date)
}

const fetchNetworkData = async () => {
  try {
    loading.value = true

    // Try store first, fallback to direct API if store fails
    try {
      await Promise.all([
        networkStore.fetchMetrics(),
        networkStore.fetchConnections(),
        networkStore.fetchAlerts(),
        networkStore.fetchStatistics()
      ])
    } catch (storeError) {
      console.warn('Store fetch failed, using direct API:', storeError)

      // Direct API fallback
      const [metrics, connections, alerts, statistics, config] = await Promise.all([
        serviceCoordinator.GetNetworkMetrics(),
        serviceCoordinator.GetNetworkConnections(),
        serviceCoordinator.GetNetworkAlerts(),
        serviceCoordinator.GetNetworkStatistics(),
        serviceCoordinator.GetNetworkConfig()
      ])

      // Update store with direct API results
      networkStore.setMetrics(metrics)
      networkStore.setConnections(connections || [])
      networkStore.setAlerts(alerts || [])
      networkStore.setStatistics(statistics)
      networkStore.setConfig(config)
    }

  } catch (error) {
    console.error('Failed to fetch network data:', error)
    toast.error(`Failed to fetch network data: ${error.message}`)
  } finally {
    loading.value = false
  }
}

const toggleMonitoring = async () => {
  try {
    loading.value = true
    if (isMonitoring.value) {
      await networkStore.stopMonitoring()
      toast.success('Network monitoring stopped')
    } else {
      await networkStore.startMonitoring()
      toast.success('Network monitoring started')
    }
  } catch (error) {
    console.error('Failed to toggle monitoring:', error)
    toast.error(`Failed to toggle monitoring: ${error.message}`)
  } finally {
    loading.value = false
  }
}

const refreshData = async () => {
  if (loading.value) return

  try {
    loading.value = true
    await fetchNetworkData()
    toast.success('Network data refreshed')
  } catch (error) {
    toast.error(`Failed to refresh data: ${error.message}`)
  } finally {
    loading.value = false
  }
}

const onInterfaceSelected = (interfaceName: string) => {
  selectedInterface.value = interfaceName
  try {
    networkStore.fetchBandwidthData(interfaceName)
  } catch (error) {
    console.warn('Store bandwidth fetch failed, using direct data')
  }
}

const onConnectionSelected = (connection: any) => {
  selectedConnection.value = connection
}

const onAlertResolved = async (alertId: string) => {
  try {
    await networkStore.resolveAlert(alertId)
    toast.success('Alert resolved')
  } catch (error) {
    console.error('Failed to resolve alert:', error)
    toast.error(`Failed to resolve alert: ${error.message}`)
  }
}

const onAlertsCleared = async () => {
  try {
    await networkStore.clearAlerts()
    toast.success('Resolved alerts cleared')
  } catch (error) {
    console.error('Failed to clear alerts:', error)
    toast.error(`Failed to clear alerts: ${error.message}`)
  }
}

const onSettingsSave = async (config: any) => {
  try {
    await networkStore.updateConfig(config)
    showSettings.value = false
    toast.success('Network settings updated')
  } catch (error) {
    console.error('Failed to update settings:', error)
    toast.error(`Failed to update settings: ${error.message}`)
  }
}

const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  
  if (isMonitoring.value) {
    refreshTimer = setInterval(refreshData, updateInterval.value * 1000)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// Watchers
watch(isMonitoring, (newValue) => {
  if (newValue) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})

watch(updateInterval, () => {
  if (isMonitoring.value) {
    startAutoRefresh()
  }
})

// Lifecycle
onMounted(async () => {
  await fetchNetworkData()
  
  // Check monitoring status
  try {
    isMonitoring.value = await serviceCoordinator.IsNetworkMonitoring()
    if (isMonitoring.value) {
      startAutoRefresh()
    }
  } catch (error) {
    console.error('Failed to check monitoring status:', error)
  }
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.network-dashboard {
  padding: 1rem;
  background: var(--surface-primary);
  border-radius: 8px;
  min-height: 100vh;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-secondary);
}

.dashboard-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--accent-primary);
}

.dashboard-controls {
  display: flex;
  gap: 0.5rem;
}

.monitoring-btn,
.refresh-btn,
.settings-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  background: var(--surface-secondary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.875rem;
}

.monitoring-btn:hover,
.refresh-btn:hover,
.settings-btn:hover {
  background: var(--surface-tertiary);
  border-color: var(--border-primary);
}

.monitoring-btn.active {
  background: var(--accent-primary);
  color: white;
  border-color: var(--accent-primary);
}

.monitoring-btn:disabled,
.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.overview-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.overview-card {
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 8px;
  padding: 1.25rem;
  transition: all 0.2s ease;
}

.overview-card:hover {
  border-color: var(--border-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.card-header h3 {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.card-header svg {
  width: 20px;
  height: 20px;
  color: var(--accent-primary);
}

.card-content {
  text-align: center;
}

.metric-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.metric-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.metric-change {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.metric-change.positive {
  color: var(--success-primary);
}

.metric-change.negative {
  color: var(--danger-primary);
}

.metric-change svg {
  width: 12px;
  height: 12px;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  grid-auto-rows: minmax(300px, auto);
  gap: 1rem;
  margin-bottom: 1rem;
}

.grid-item {
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 8px;
  overflow: hidden;
}

.bandwidth-chart {
  grid-column: span 8;
  grid-row: span 2;
}

.interfaces-list {
  grid-column: span 4;
  grid-row: span 1;
}

.connections-list {
  grid-column: span 6;
  grid-row: span 2;
}

.alerts-panel {
  grid-column: span 3;
  grid-row: span 1;
}

.statistics-panel {
  grid-column: span 3;
  grid-row: span 1;
}

.connection-details {
  grid-column: span 3;
  grid-row: span 1;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: var(--surface-tertiary);
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  font-size: 0.875rem;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-secondary);
}

.status-item.status-error {
  color: var(--danger-primary);
}

.status-item svg {
  width: 14px;
  height: 14px;
}

/* Responsive design */
@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: repeat(6, 1fr);
  }
  
  .bandwidth-chart {
    grid-column: span 6;
  }
  
  .interfaces-list {
    grid-column: span 6;
  }
  
  .connections-list {
    grid-column: span 6;
  }
  
  .alerts-panel {
    grid-column: span 3;
  }
  
  .statistics-panel {
    grid-column: span 3;
  }
}

@media (max-width: 768px) {
  .dashboard-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }
  
  .dashboard-controls {
    justify-content: center;
  }
  
  .overview-cards {
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  }
  
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
  
  .grid-item {
    grid-column: span 1 !important;
  }
  
  .status-bar {
    flex-direction: column;
    gap: 0.5rem;
    align-items: stretch;
  }
}
</style>