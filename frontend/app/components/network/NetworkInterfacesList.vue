<template>
  <div class="interfaces-list">
    <div class="list-header">
      <h3>Network Interfaces</h3>
      <button @click="$emit('refresh')" class="refresh-btn" :disabled="loading">
        <Icon name="refresh" :class="{ spinning: loading }" />
      </button>
    </div>
    
    <div class="list-content">
      <div v-if="loading" class="list-loading">
        <div class="loading-spinner"></div>
        <p>Loading interfaces...</p>
      </div>
      
      <div v-else-if="interfaces.length === 0" class="list-empty">
        <div class="empty-icon">🌐</div>
        <p>No network interfaces found</p>
      </div>
      
      <div v-else class="interfaces-grid">
        <div
          v-for="iface in interfaces"
          :key="iface.name"
          :class="['interface-card', { active: iface.isUp, selected: selectedInterface === iface.name }]"
          @click="selectInterface(iface.name)"
        >
          <div class="interface-header">
            <div class="interface-name">{{ iface.name }}</div>
            <div :class="['interface-status', { active: iface.isUp }]">
              {{ iface.isUp ? 'Active' : 'Inactive' }}
            </div>
          </div>
          
          <div class="interface-stats">
            <div class="stat-row">
              <span class="stat-label">Sent:</span>
              <span class="stat-value">{{ formatBytes(iface.bytesSent) }}</span>
            </div>
            <div class="stat-row">
              <span class="stat-label">Received:</span>
              <span class="stat-value">{{ formatBytes(iface.bytesRecv) }}</span>
            </div>
            <div class="stat-row">
              <span class="stat-label">Speed:</span>
              <span class="stat-value">{{ formatSpeed(iface.speed) }}</span>
            </div>
            <div class="stat-row" v-if="iface.ipAddresses && iface.ipAddresses.length > 0">
              <span class="stat-label">IP:</span>
              <span class="stat-value">{{ iface.ipAddresses[0] }}</span>
            </div>
          </div>
          
          <div class="interface-activity">
            <div class="activity-indicator" :class="{ active: iface.isUp }"></div>
            <span class="activity-text">{{ iface.isUp ? 'Connected' : 'Disconnected' }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  interfaces: any[]
  loading?: boolean
  selectedInterface?: string
}>()

const emit = defineEmits<{
  'interface-selected': [interfaceName: string]
  'refresh': []
}>()

// Methods
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

const formatSpeed = (bits: number): string => {
  if (!bits || bits === 0) return 'Unknown'
  const mbps = bits / 1000000
  if (mbps < 1000) {
    return `${mbps.toFixed(0)} Mbps`
  }
  const gbps = mbps / 1000
  return `${gbps.toFixed(1)} Gbps`
}

const selectInterface = (interfaceName: string) => {
  emit('interface-selected', interfaceName)
}
</script>

<style scoped>
.interfaces-list {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-secondary);
}

.list-header h3 {
  margin: 0;
  font-size: 1.125rem;
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

.list-content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

.list-loading,
.list-empty {
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

.interfaces-grid {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.interface-card {
  padding: 1rem;
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  background: var(--surface-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.interface-card:hover {
  border-color: var(--border-primary);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.interface-card.active {
  border-color: var(--success-primary);
  background: rgba(16, 185, 129, 0.05);
}

.interface-card.selected {
  border-color: var(--accent-primary);
  background: rgba(59, 130, 246, 0.1);
}

.interface-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.interface-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.interface-status {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-weight: 500;
}

.interface-status.active {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success-primary);
}

.interface-status:not(.active) {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger-primary);
}

.interface-stats {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
}

.stat-label {
  color: var(--text-secondary);
  font-weight: 500;
}

.stat-value {
  color: var(--text-primary);
  font-weight: 500;
  font-family: 'Courier New', monospace;
}

.interface-activity {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border-secondary);
}

.activity-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-tertiary);
  transition: all 0.2s ease;
}

.activity-indicator.active {
  background: var(--success-primary);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  }
  70% {
    box-shadow: 0 0 0 6px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

.activity-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.interface-card.selected .activity-text {
  color: var(--accent-primary);
}
</style>