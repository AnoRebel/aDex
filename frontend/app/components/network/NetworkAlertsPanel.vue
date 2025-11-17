<template>
  <div class="alerts-panel">
    <div class="panel-header">
      <h3>Network Alerts</h3>
      <div class="header-controls">
        <span class="alert-count" :class="{ 'has-alerts': activeAlerts.length > 0 }">
          {{ activeAlerts.length }}
        </span>
        <button
          v-if="resolvedAlerts.length > 0"
          @click="clearResolvedAlerts"
          class="clear-btn"
          :disabled="loading"
        >
          Clear
        </button>
      </div>
    </div>
    
    <div class="panel-content">
      <div v-if="loading" class="panel-loading">
        <div class="loading-spinner"></div>
        <p>Loading alerts...</p>
      </div>
      
      <div v-else-if="alerts.length === 0" class="panel-empty">
        <div class="empty-icon">✅</div>
        <p>No network alerts</p>
      </div>
      
      <div v-else class="alerts-list">
        <div
          v-for="alert in alerts"
          :key="alert.id"
          :class="['alert-item', getAlertClass(alert)]"
        >
          <div class="alert-header">
            <div class="alert-icon">
              <Icon :name="getAlertIcon(alert.severity)" />
            </div>
            <div class="alert-title">{{ alert.type }}</div>
            <div class="alert-actions">
              <button
                v-if="!alert.resolved"
                @click="resolveAlert(alert.id)"
                class="resolve-btn"
                :disabled="loading"
                title="Resolve alert"
              >
                <Icon name="check" />
              </button>
            </div>
          </div>
          
          <div class="alert-message">{{ alert.message }}</div>
          
          <div class="alert-details">
            <div class="detail-item" v-if="alert.interface">
              <span class="detail-label">Interface:</span>
              <span class="detail-value">{{ alert.interface }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Threshold:</span>
              <span class="detail-value">{{ formatAlertValue(alert.threshold) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Current:</span>
              <span class="detail-value">{{ formatAlertValue(alert.current) }}</span>
            </div>
          </div>
          
          <div class="alert-footer">
            <span class="alert-time">{{ formatTime(alert.timestamp) }}</span>
            <span class="alert-status" :class="{ resolved: alert.resolved }">
              {{ alert.resolved ? 'Resolved' : 'Active' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  alerts: any[]
  loading?: boolean
}>()

const emit = defineEmits<{
  'alert-resolved': [alertId: string]
  'alerts-cleared': []
}>()

// Computed properties
const activeAlerts = computed(() => {
  return props.alerts.filter(alert => !alert.resolved)
})

const resolvedAlerts = computed(() => {
  return props.alerts.filter(alert => alert.resolved)
})

// Methods
const getAlertClass = (alert: any): string => {
  if (alert.resolved) return 'alert-resolved'
  
  switch (alert.severity) {
    case 'error':
      return 'alert-error'
    case 'warning':
      return 'alert-warning'
    case 'info':
      return 'alert-info'
    default:
      return 'alert-default'
  }
}

const getAlertIcon = (severity: string): string => {
  switch (severity) {
    case 'error':
      return 'alert-circle'
    case 'warning':
      return 'alert-triangle'
    case 'info':
      return 'info'
    default:
      return 'bell'
  }
}

const formatAlertValue = (value: number): string => {
  if (value < 1000) {
    return `${value.toFixed(1)}`
  }
  if (value < 1000000) {
    return `${(value / 1000).toFixed(1)}K`
  }
  return `${(value / 1000000).toFixed(1)}M`
}

const formatTime = (timestamp: Date): string => {
  const now = new Date()
  const diff = now.getTime() - timestamp.getTime()
  const minutes = Math.floor(diff / 60000)
  
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

const resolveAlert = async (alertId: string) => {
  emit('alert-resolved', alertId)
}

const clearResolvedAlerts = () => {
  emit('alerts-cleared')
}
</script>

<style scoped>
.alerts-panel {
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

.header-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.alert-count {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 24px;
  height: 24px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  background: var(--surface-secondary);
  color: var(--text-secondary);
}

.alert-count.has-alerts {
  background: var(--danger-primary);
  color: white;
}

.clear-btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 0.75rem;
  transition: all 0.2s ease;
}

.clear-btn:hover:not(:disabled) {
  background: var(--surface-secondary);
  border-color: var(--border-primary);
  color: var(--text-primary);
}

.clear-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.panel-loading,
.panel-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 150px;
  color: var(--text-secondary);
  text-align: center;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-secondary);
  border-top: 2px solid var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 0.75rem;
}

.empty-icon {
  font-size: 2rem;
  margin-bottom: 0.75rem;
  opacity: 0.5;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.alerts-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.alert-item {
  padding: 0.75rem;
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  background: var(--surface-primary);
  transition: all 0.2s ease;
}

.alert-item:hover {
  border-color: var(--border-primary);
  transform: translateY(-1px);
}

.alert-error {
  border-left: 4px solid var(--danger-primary);
  background: rgba(239, 68, 68, 0.05);
}

.alert-warning {
  border-left: 4px solid var(--warning-primary);
  background: rgba(245, 158, 11, 0.05);
}

.alert-info {
  border-left: 4px solid var(--accent-primary);
  background: rgba(59, 130, 246, 0.05);
}

.alert-default {
  border-left: 4px solid var(--text-secondary);
}

.alert-resolved {
  opacity: 0.6;
  border-left-color: var(--success-primary);
}

.alert-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.alert-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
}

.alert-error .alert-icon {
  color: var(--danger-primary);
}

.alert-warning .alert-icon {
  color: var(--warning-primary);
}

.alert-info .alert-icon {
  color: var(--accent-primary);
}

.alert-default .alert-icon {
  color: var(--text-secondary);
}

.alert-title {
  flex: 1;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.alert-actions {
  display: flex;
  gap: 0.25rem;
}

.resolve-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.resolve-btn:hover:not(:disabled) {
  background: var(--success-primary);
  border-color: var(--success-primary);
  color: white;
}

.resolve-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.alert-message {
  font-size: 0.875rem;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
  line-height: 1.4;
}

.alert-details {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.5rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
}

.detail-label {
  color: var(--text-secondary);
  font-weight: 500;
}

.detail-value {
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
}

.alert-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
}

.alert-time {
  color: var(--text-secondary);
}

.alert-status {
  padding: 0.125rem 0.375rem;
  border-radius: 10px;
  font-weight: 500;
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger-primary);
}

.alert-status.resolved {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success-primary);
}

/* Responsive design */
@media (max-width: 480px) {
  .panel-header {
    flex-direction: column;
    gap: 0.75rem;
    align-items: stretch;
  }
  
  .alert-details {
    font-size: 0.7rem;
  }
  
  .detail-item {
    flex-direction: column;
    gap: 0.125rem;
  }
}
</style>