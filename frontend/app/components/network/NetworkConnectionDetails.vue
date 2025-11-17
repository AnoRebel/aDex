<template>
  <div class="connection-details">
    <div class="details-header">
      <h3>Connection Details</h3>
      <button @click="$emit('close')" class="close-btn">
        <Icon name="x" />
      </button>
    </div>
    
    <div class="details-content" v-if="connection">
      <!-- Connection Overview -->
      <div class="overview-section">
        <div class="connection-state" :class="getStateClass(connection.state)">
          <Icon :name="getStateIcon(connection.state)" />
          <span>{{ connection.state }}</span>
        </div>
        <div class="connection-protocol">
          <Icon :name="getProtocolIcon(connection.protocol)" />
          <span>{{ connection.protocol }}</span>
        </div>
        <div class="connection-duration">
          <Icon name="clock" />
          <span>{{ formatDuration(connection.established) }}</span>
        </div>
      </div>
      
      <!-- Address Information -->
      <div class="section">
        <h4 class="section-title">Address Information</h4>
        <div class="address-grid">
          <div class="address-item">
            <span class="address-label">Local:</span>
            <span class="address-value">{{ connection.localAddr }}</span>
          </div>
          <div class="address-item">
            <span class="address-label">Remote:</span>
            <span class="address-value">{{ connection.remoteAddr }}</span>
          </div>
        </div>
      </div>
      
      <!-- Process Information -->
      <div class="section">
        <h4 class="section-title">Process Information</h4>
        <div class="process-info">
          <div class="process-item" v-if="connection.processName">
            <span class="process-label">Name:</span>
            <span class="process-value">{{ connection.processName }}</span>
          </div>
          <div class="process-item" v-if="connection.pid">
            <span class="process-label">PID:</span>
            <span class="process-value">{{ connection.pid }}</span>
          </div>
          <div class="process-item" v-if="connection.executable">
            <span class="process-label">Executable:</span>
            <span class="process-value">{{ connection.executable }}</span>
          </div>
          <div class="process-item" v-if="connection.cwd">
            <span class="process-label">Working Directory:</span>
            <span class="process-value">{{ connection.cwd }}</span>
          </div>
        </div>
      </div>
      
      <!-- Traffic Information -->
      <div class="section">
        <h4 class="section-title">Traffic Information</h4>
        <div class="traffic-info">
          <div class="traffic-item">
            <div class="traffic-header sent">
              <Icon name="upload" />
              <span>Bytes Sent</span>
            </div>
            <div class="traffic-value">{{ formatBytes(connection.bytesSent) }}</div>
            <div class="traffic-bar">
              <div class="bar-fill sent-fill" :style="{ width: getUploadPercentage() + '%' }"></div>
            </div>
          </div>
          
          <div class="traffic-item">
            <div class="traffic-header received">
              <Icon name="download" />
              <span>Bytes Received</span>
            </div>
            <div class="traffic-value">{{ formatBytes(connection.bytesRecv) }}</div>
            <div class="traffic-bar">
              <div class="bar-fill received-fill" :style="{ width: getDownloadPercentage() + '%' }"></div>
            </div>
          </div>
          
          <div class="traffic-item total">
            <div class="traffic-header">
              <Icon name="activity" />
              <span>Total Traffic</span>
            </div>
            <div class="traffic-value">{{ formatBytes(connection.bytesSent + connection.bytesRecv) }}</div>
            <div class="traffic-bar">
              <div class="bar-fill total-fill" :style="{ width: getTotalPercentage() + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Packet Information -->
      <div class="section" v-if="connection.packetsSent !== undefined || connection.packetsRecv !== undefined">
        <h4 class="section-title">Packet Information</h4>
        <div class="packet-info">
          <div class="packet-item">
            <span class="packet-label">Packets Sent:</span>
            <span class="packet-value">{{ formatNumber(connection.packetsSent) }}</span>
          </div>
          <div class="packet-item">
            <span class="packet-label">Packets Received:</span>
            <span class="packet-value">{{ formatNumber(connection.packetsRecv) }}</span>
          </div>
          <div class="packet-item">
            <span class="packet-label">Total Packets:</span>
            <span class="packet-value">{{ formatNumber((connection.packetsSent || 0) + (connection.packetsRecv || 0)) }}</span>
          </div>
        </div>
      </div>
      
      <!-- Timestamp Information -->
      <div class="section">
        <h4 class="section-title">Timestamp Information</h4>
        <div class="timestamp-info">
          <div class="timestamp-item">
            <span class="timestamp-label">Established:</span>
            <span class="timestamp-value">{{ formatDateTime(connection.established) }}</span>
          </div>
          <div class="timestamp-item" v-if="connection.lastActivity">
            <span class="timestamp-label">Last Activity:</span>
            <span class="timestamp-value">{{ formatDateTime(connection.lastActivity) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  connection: any
}>()

const emit = defineEmits<{
  'close': []
}>()

// Computed properties
const maxTraffic = computed(() => {
  if (!props.connection) return 1
  return (props.connection.bytesSent || 0) + (props.connection.bytesRecv || 0)
})

// Methods
const getStateClass = (state: string): string => {
  switch (state) {
    case 'ESTABLISHED':
      return 'state-established'
    case 'LISTEN':
      return 'state-listening'
    case 'TIME_WAIT':
      return 'state-waiting'
    case 'CLOSE_WAIT':
      return 'state-closing'
    default:
      return 'state-other'
  }
}

const getStateIcon = (state: string): string => {
  switch (state) {
    case 'ESTABLISHED':
      return 'link'
    case 'LISTEN':
      return 'server'
    case 'TIME_WAIT':
      return 'clock'
    case 'CLOSE_WAIT':
      return 'x-circle'
    default:
      return 'help-circle'
  }
}

const getProtocolIcon = (protocol: string): string => {
  switch (protocol?.toLowerCase()) {
    case 'tcp':
      return 'layers'
    case 'udp':
      return 'zap'
    case 'icmp':
      return 'radio'
    default:
      return 'help-circle'
  }
}

const getUploadPercentage = (): number => {
  if (!props.connection) return 0
  return Math.min(((props.connection.bytesSent || 0) / maxTraffic.value) * 100, 100)
}

const getDownloadPercentage = (): number => {
  if (!props.connection) return 0
  return Math.min(((props.connection.bytesRecv || 0) / maxTraffic.value) * 100, 100)
}

const getTotalPercentage = (): number => {
  return Math.max(getUploadPercentage(), getDownloadPercentage())
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

const formatNumber = (num: number): string => {
  return num.toLocaleString()
}

const formatDuration = (established: Date): string => {
  if (!established) return 'Unknown'
  const now = new Date()
  const diff = now.getTime() - established.getTime()
  
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  
  if (days > 0) return `${days}d ${hours % 24}h`
  if (hours > 0) return `${hours}h ${minutes % 60}m`
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`
  return `${seconds}s`
}

const formatDateTime = (date: Date): string => {
  if (!date) return 'Unknown'
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).format(date)
}
</script>

<style scoped>
.connection-details {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.details-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-secondary);
}

.details-header h3 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
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

.close-btn:hover {
  background: var(--surface-secondary);
  border-color: var(--border-primary);
}

.details-content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.overview-section {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.connection-state,
.connection-protocol,
.connection-duration {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
}

.connection-state {
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
}

.connection-state.state-established {
  background: rgba(16, 185, 129, 0.1);
  border-color: var(--success-primary);
  color: var(--success-primary);
}

.connection-state.state-listening {
  background: rgba(59, 130, 246, 0.1);
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.connection-state.state-waiting {
  background: rgba(245, 158, 11, 0.1);
  border-color: var(--warning-primary);
  color: var(--warning-primary);
}

.connection-state.state-closing {
  background: rgba(156, 163, 175, 0.1);
  border-color: var(--text-secondary);
  color: var(--text-secondary);
}

.connection-protocol {
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  color: var(--text-primary);
}

.connection-duration {
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  color: var(--text-primary);
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

.address-grid {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.address-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background: var(--surface-secondary);
  border-radius: 4px;
}

.address-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.address-value {
  font-size: 0.875rem;
  font-family: 'Courier New', monospace;
  color: var(--text-primary);
  word-break: break-all;
}

.process-info,
.traffic-info,
.packet-info,
.timestamp-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.process-item,
.packet-item,
.timestamp-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background: var(--surface-secondary);
  border-radius: 4px;
}

.process-label,
.packet-label,
.timestamp-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.process-value,
.packet-value,
.timestamp-value {
  font-size: 0.875rem;
  color: var(--text-primary);
  text-align: right;
}

.traffic-item {
  padding: 0.75rem;
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
}

.traffic-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--text-primary);
}

.traffic-header.sent {
  color: var(--warning-primary);
}

.traffic-header.received {
  color: var(--accent-primary);
}

.traffic-value {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
  font-family: 'Courier New', monospace;
}

.traffic-bar {
  height: 8px;
  background: var(--surface-tertiary);
  border-radius: 4px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.sent-fill {
  background: linear-gradient(90deg, #f59e0b, #d97706);
}

.received-fill {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.total-fill {
  background: linear-gradient(90deg, #10b981, #059669);
}

.traffic-item.total {
  border-color: var(--success-primary);
  background: rgba(16, 185, 129, 0.05);
}

/* Responsive design */
@media (max-width: 768px) {
  .overview-section {
    flex-direction: column;
  }
  
  .address-item,
  .process-item,
  .packet-item,
  .timestamp-item {
    flex-direction: column;
    align-items: stretch;
    gap: 0.25rem;
    text-align: left;
  }
  
  .address-value,
  .process-value,
  .packet-value,
  .timestamp-value {
    text-align: left;
  }
  
  .traffic-header {
    justify-content: center;
  }
}
</style>