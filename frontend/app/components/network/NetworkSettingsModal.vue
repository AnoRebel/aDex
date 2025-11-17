<template>
  <div class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>Network Settings</h3>
        <button @click="$emit('close')" class="close-btn">
          <Icon name="x" />
        </button>
      </div>
      
      <div class="modal-body">
        <!-- Monitoring Settings -->
        <div class="settings-section">
          <h4 class="section-title">Monitoring Settings</h4>
          
          <div class="setting-item">
            <label class="setting-label">
              <input
                type="checkbox"
                v-model="localConfig.enableConnectionTracking"
                class="setting-checkbox"
              />
              Enable Connection Tracking
            </label>
            <p class="setting-description">Track active network connections and their statistics</p>
          </div>
          
          <div class="setting-item">
            <label class="setting-label">
              <input
                type="checkbox"
                v-model="localConfig.enableBandwidthMonitoring"
                class="setting-checkbox"
              />
              Enable Bandwidth Monitoring
            </label>
            <p class="setting-description">Monitor bandwidth usage for all interfaces</p>
          </div>
          
          <div class="setting-item">
            <label class="setting-label">
              <input
                type="checkbox"
                v-model="localConfig.enableAlerts"
                class="setting-checkbox"
              />
              Enable Network Alerts
            </label>
            <p class="setting-description">Receive alerts for network issues</p>
          </div>
        </div>
        
        <!-- Performance Settings -->
        <div class="settings-section">
          <h4 class="section-title">Performance Settings</h4>
          
          <div class="setting-item">
            <label class="setting-label">
              Refresh Interval (seconds)
            </label>
            <input
              type="number"
              v-model.number="localConfig.refreshInterval"
              min="1"
              max="60"
              class="setting-input"
            />
            <p class="setting-description">How often to refresh network data</p>
          </div>
          
          <div class="setting-item">
            <label class="setting-label">
              Max Connections
            </label>
            <input
              type="number"
              v-model.number="localConfig.maxConnections"
              min="10"
              max="10000"
              class="setting-input"
            />
            <p class="setting-description">Maximum number of connections to track</p>
          </div>
          
          <div class="setting-item">
            <label class="setting-label">
              Bandwidth History Size
            </label>
            <input
              type="number"
              v-model.number="localConfig.bandwidthHistorySize"
              min="10"
              max="1000"
              class="setting-input"
            />
            <p class="setting-description">Number of data points to keep in history</p>
          </div>
        </div>
        
        <!-- Alert Thresholds -->
        <div class="settings-section">
          <h4 class="section-title">Alert Thresholds</h4>
          
          <div class="setting-item">
            <label class="setting-label">
              Bandwidth Usage (Mbps)
            </label>
            <input
              type="number"
              v-model.number="localConfig.alertThresholds.bandwidthUsageMBps"
              min="1"
              max="1000"
              step="0.1"
              class="setting-input"
            />
            <p class="setting-description">Alert when bandwidth exceeds this value</p>
          </div>
          
          <div class="setting-item">
            <label class="setting-label">
              Connection Count
            </label>
            <input
              type="number"
              v-model.number="localConfig.alertThresholds.connectionCount"
              min="10"
              max="10000"
              class="setting-input"
            />
            <p class="setting-description">Alert when connection count exceeds this value</p>
          </div>
          
          <div class="setting-item">
            <label class="setting-label">
              Packet Loss (%)
            </label>
            <input
              type="number"
              v-model.number="localConfig.alertThresholds.packetLossPercent"
              min="0.1"
              max="10"
              step="0.1"
              class="setting-input"
            />
            <p class="setting-description">Alert when packet loss exceeds this percentage</p>
          </div>
        </div>
      </div>
      
      <div class="modal-footer">
        <button @click="$emit('close')" class="btn btn-secondary">
          Cancel
        </button>
        <button @click="saveSettings" class="btn btn-primary" :disabled="loading">
          <Icon v-if="loading" name="loader" class="spinning" />
          {{ loading ? 'Saving...' : 'Save Settings' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  config: any
}>()

const emit = defineEmits<{
  'close': []
  'save': [config: any]
}>()

// Reactive data
const localConfig = ref<any>({})
const loading = ref(false)

// Initialize local config from props
watch(() => props.config, (newConfig) => {
  if (newConfig) {
    localConfig.value = { ...newConfig }
  }
}, { immediate: true, deep: true })

// Initialize with defaults if no config provided
if (!props.config) {
  localConfig.value = {
    enableConnectionTracking: true,
    enableBandwidthMonitoring: true,
    enableAlerts: true,
    refreshInterval: 2,
    maxConnections: 1000,
    bandwidthHistorySize: 300,
    alertThresholds: {
      bandwidthUsageMBps: 100,
      connectionCount: 500,
      packetLossPercent: 1.0,
      errorRatePercent: 0.1,
      highLatencyMs: 1000
    }
  }
}

// Methods
const saveSettings = async () => {
  loading.value = true
  
  try {
    // Simulate async save operation
    await new Promise(resolve => setTimeout(resolve, 500))
    
    emit('save', localConfig.value)
    emit('close')
  } catch (error) {
    console.error('Failed to save settings:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: var(--surface-primary);
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-secondary);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
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

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.settings-section {
  margin-bottom: 2rem;
}

.settings-section:last-child {
  margin-bottom: 0;
}

.section-title {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-secondary);
}

.setting-item {
  margin-bottom: 1.5rem;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
  cursor: pointer;
}

.setting-checkbox {
  accent-color: var(--accent-primary);
}

.setting-input {
  display: block;
  width: 100%;
  max-width: 200px;
  padding: 0.5rem;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-primary);
  font-size: 0.875rem;
}

.setting-input:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.setting-description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.4;
  margin: 0;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.5rem;
  border-top: 1px solid var(--border-secondary);
}

.btn {
  padding: 0.625rem 1.25rem;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-secondary {
  background: var(--surface-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-secondary);
}

.btn-secondary:hover {
  background: var(--surface-tertiary);
  border-color: var(--border-primary);
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-primary-hover);
}

.btn:disabled {
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

/* Responsive design */
@media (max-width: 640px) {
  .modal-content {
    margin: 1rem;
    max-height: 95vh;
  }
  
  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 1rem;
  }
  
  .setting-input {
    max-width: none;
  }
  
  .modal-footer {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>