<template>
  <div class="statistics-panel">
    <div class="panel-header">
      <h3>Network Statistics</h3>
      <button @click="refreshStats" class="refresh-btn" :disabled="loading">
        <Icon name="refresh" :class="{ spinning: loading }" />
      </button>
    </div>
    
    <div class="panel-content">
      <div v-if="loading" class="panel-loading">
        <div class="loading-spinner"></div>
        <p>Loading statistics...</p>
      </div>
      
      <div v-else-if="!statistics" class="panel-empty">
        <div class="empty-icon">📊</div>
        <p>No statistics available</p>
      </div>
      
      <div v-else class="stats-content">
        <!-- Overview Stats -->
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon">
              <Icon name="link" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.totalConnections || 0 }}</div>
              <div class="stat-label">Total Connections</div>
            </div>
          </div>
          
          <div class="stat-card">
            <div class="stat-icon">
              <Icon name="activity" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatBytes(statistics.totalBytesSent + statistics.totalBytesRecv) }}</div>
              <div class="stat-label">Total Traffic</div>
            </div>
          </div>
          
          <div class="stat-card">
            <div class="stat-icon">
              <Icon name="upload" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatBytes(statistics.totalBytesSent) }}</div>
              <div class="stat-label">Bytes Sent</div>
            </div>
          </div>
          
          <div class="stat-card">
            <div class="stat-icon">
              <Icon name="download" />
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatBytes(statistics.totalBytesRecv) }}</div>
              <div class="stat-label">Bytes Received</div>
            </div>
          </div>
        </div>
        
        <!-- Top Connections -->
        <div class="section" v-if="statistics.topConnections && statistics.topConnections.length > 0">
          <h4 class="section-title">Top Connections by Bandwidth</h4>
          <div class="top-connections">
            <div
              v-for="(connection, index) in statistics.topConnections.slice(0, 5)"
              :key="getConnectionKey(connection)"
              class="connection-item"
            >
              <div class="connection-rank">{{ index + 1 }}</div>
              <div class="connection-info">
                <div class="connection-process">{{ connection.processName || 'Unknown' }}</div>
                <div class="connection-address">{{ connection.remoteAddr }}</div>
              </div>
              <div class="connection-traffic">
                <div class="traffic-value">{{ formatBytes(connection.bytesSent + connection.bytesRecv) }}</div>
                <div class="traffic-bars">
                  <div class="traffic-bar sent">
                    <div class="bar-fill sent-fill" :style="{ width: getTrafficPercentage(connection, 'sent') + '%' }"></div>
                  </div>
                  <div class="traffic-bar received">
                    <div class="bar-fill received-fill" :style="{ width: getTrafficPercentage(connection, 'received') + '%' }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Interface Stats -->
        <div class="section" v-if="statistics.interfaceStats">
          <h4 class="section-title">Interface Statistics</h4>
          <div class="interface-stats">
            <div
              v-for="(stats, interfaceName) in statistics.interfaceStats"
              :key="interfaceName"
              class="interface-item"
            >
              <div class="interface-name">{{ interfaceName }}</div>
              <div class="interface-metrics">
                <div class="metric">
                  <span class="metric-label">Upload:</span>
                  <span class="metric-value">{{ formatSpeed(stats.uploadMbps || 0) }}</span>
                </div>
                <div class="metric">
                  <span class="metric-label">Download:</span>
                  <span class="metric-value">{{ formatSpeed(stats.downloadMbps || 0) }}</span>
                </div>
                <div class="metric" v-if="stats.peakUpload">
                  <span class="metric-label">Peak:</span>
                  <span class="metric-value">{{ formatSpeed(stats.peakUpload) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Last Update -->
        <div class="section">
          <div class="last-update">
            <span class="update-label">Last updated:</span>
            <span class="update-time">{{ formatTime(statistics.timestamp) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  statistics?: any
  loading?: boolean
}>()

const emit = defineEmits<{
  'refresh': []
}>()

// Computed properties
const maxTraffic = computed(() => {
  if (!props.statistics?.topConnections) return 1
  return Math.max(...props.statistics.topConnections.map((conn: any) => 
    (conn.bytesSent || 0) + (conn.bytesRecv || 0)
  ))
})

// Methods
const getConnectionKey = (connection: any): string => {
  return `${connection.protocol}-${connection.localAddr}-${connection.remoteAddr}-${connection.pid}`
}

const getTrafficPercentage = (connection: any, type: 'sent' | 'received'): number => {
  const value = type === 'sent' ? (connection.bytesSent || 0) : (connection.bytesRecv || 0)
  const total = (connection.bytesSent || 0) + (connection.bytesRecv || 0)
  if (total === 0) return 0
  
  // Use overall max traffic as reference
  return Math.min((value / maxTraffic.value) * 100, 100)
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

const formatSpeed = (mbps: number): string => {
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

const formatTime = (timestamp: Date | string): string => {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : timestamp
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).format(date)
}

const refreshStats = () => {
  emit('refresh')
}
</script>

<style scoped>
.statistics-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-secondary);
}

.panel-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.refresh-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--surface-secondary);
  border-color: var(--border-primary);
}

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

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

.panel-loading,
.panel-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text-secondary);
  text-align: center;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-secondary);
  border-top: 3px solid var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.stats-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  transition: all 0.2s ease;
}

.stat-card:hover {
  border-color: var(--border-primary);
  transform: translateY(-1px);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: var(--accent-primary);
  color: white;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.2;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.section-title {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-secondary);
}

.top-connections {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.connection-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  font-size: 0.75rem;
}

.connection-rank {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--accent-primary);
  color: white;
  font-weight: 600;
  font-size: 0.625rem;
}

.connection-info {
  flex: 1;
}

.connection-process {
  font-weight: 500;
  color: var(--text-primary);
}

.connection-address {
  font-size: 0.625rem;
  color: var(--text-secondary);
  font-family: 'Courier New', monospace;
}

.connection-traffic {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}

.traffic-value {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.75rem;
}

.traffic-bars {
  display: flex;
  gap: 0.25rem;
}

.traffic-bar {
  width: 40px;
  height: 4px;
  background: var(--surface-tertiary);
  border-radius: 2px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

.sent-fill {
  background: linear-gradient(90deg, #f59e0b, #d97706);
}

.received-fill {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.interface-stats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.interface-item {
  padding: 0.75rem;
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
}

.interface-name {
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
}

.interface-metrics {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.metric {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
}

.metric-label {
  color: var(--text-secondary);
}

.metric-value {
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
  font-weight: 500;
}

.last-update {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  font-size: 0.75rem;
}

.update-label {
  color: var(--text-secondary);
}

.update-time {
  color: var(--text-primary);
  font-weight: 500;
}

/* Responsive design */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .connection-item {
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
  }
  
  .connection-traffic {
    align-items: stretch;
  }
  
  .traffic-bars {
    justify-content: center;
  }
  
  .interface-metrics {
    gap: 0.5rem;
  }
  
  .metric {
    flex-direction: column;
    gap: 0.125rem;
  }
  
  .last-update {
    flex-direction: column;
    gap: 0.25rem;
    text-align: center;
  }
}
</style>