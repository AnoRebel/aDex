<template>
  <div class="geoip-settings">
    <div class="settings-header">
      <h3>GeoIP Settings</h3>
      <div class="toggle-container">
        <label class="toggle">
          <input
            type="checkbox"
            v-model="localSettings.enabled"
            @change="toggleGeoIP"
            :disabled="isLoading"
          />
          <span class="slider"></span>
        </label>
        <span class="toggle-label">Enable GeoIP</span>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <Icon name="alert-triangle" />
      <span>{{ error }}</span>
    </div>

    <div v-if="localSettings.enabled" class="settings-content">
      <!-- Provider Selection -->
      <div class="setting-group">
        <label class="setting-label">Provider</label>
        <select
          v-model="localSettings.provider"
          @change="updateSettings"
          :disabled="isLoading"
          class="setting-select"
        >
          <option
            v-for="provider in availableProviders"
            :key="provider.name"
            :value="provider.name"
          >
            {{ provider.displayName }} ({{ provider.freeTier ? 'Free' : 'Paid' }})
          </option>
        </select>
        <p class="setting-description">
          {{ selectedProvider?.description }}
        </p>
      </div>

      <!-- Display Options -->
      <div class="setting-group">
        <label class="setting-label">Display Options</label>
        <div class="checkbox-group">
          <label class="checkbox-item">
            <input
              type="checkbox"
              v-model="localSettings.showFlags"
              @change="updateSettings"
              :disabled="isLoading"
            />
            <span class="checkmark"></span>
            Show country flags
          </label>
          <label class="checkbox-item">
            <input
              type="checkbox"
              v-model="localSettings.showCity"
              @change="updateSettings"
              :disabled="isLoading"
            />
            <span class="checkmark"></span>
            Show city information
          </label>
          <label class="checkbox-item">
            <input
              type="checkbox"
              v-model="localSettings.showISP"
              @change="updateSettings"
              :disabled="isLoading"
            />
            <span class="checkmark"></span>
            Show ISP information
          </label>
        </div>
      </div>

      <!-- Cache Settings -->
      <div class="setting-group">
        <label class="setting-label">Cache Settings</label>
        <div class="cache-info">
          <div class="cache-stats">
            <span class="stat-label">Cache size:</span>
            <span class="stat-value">
              {{ cacheStats?.size || 0 }} / {{ cacheStats?.maxSize || 1000 }}
            </span>
          </div>
          <div class="cache-stats">
            <span class="stat-label">Hit rate:</span>
            <span class="stat-value">{{ cacheHitRate.toFixed(1) }}%</span>
          </div>
          <div class="cache-stats">
            <span class="stat-label">Last update:</span>
            <span class="stat-value">
              {{ cacheStats?.lastUpdate ? formatTime(cacheStats.lastUpdate) : 'Never' }}
            </span>
          </div>
        </div>
        <div class="cache-actions">
          <button
            @click="refreshCacheStats"
            :disabled="isLoading"
            class="btn btn-secondary"
          >
            <Icon name="refresh" :class="{ spinning: isLoading }" />
            Refresh
          </button>
          <button
            @click="clearCache"
            :disabled="isLoading"
            class="btn btn-danger"
          >
            <Icon name="trash-2" />
            Clear Cache
          </button>
        </div>
      </div>

      <!-- Privacy Settings -->
      <div class="setting-group">
        <label class="setting-label">Privacy</label>
        <label class="checkbox-item">
          <input
            type="checkbox"
            v-model="localSettings.privacyMode"
            @change="updateSettings"
            :disabled="isLoading"
          />
          <span class="checkmark"></span>
          Privacy mode (reduced data collection)
        </label>
        <p class="setting-description">
          When enabled, limits the amount of information stored and displayed.
        </p>
      </div>

      <!-- Performance Settings -->
      <div class="setting-group">
        <label class="setting-label">Performance</label>
        <label class="checkbox-item">
          <input
            type="checkbox"
            v-model="localSettings.autoLookup"
            @change="updateSettings"
            :disabled="isLoading"
          />
          <span class="checkmark"></span>
          Automatic lookup for new connections
        </label>
        <p class="setting-description">
          Automatically fetches GeoIP data for new network connections.
        </p>
      </div>
    </div>

    <div v-if="!localSettings.enabled" class="disabled-message">
      <Icon name="info" />
      <p>
        Enable GeoIP to display geographical information for network connections.
        This feature requires internet access and may impact privacy.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useGeoIP, GeoIPUtils } from '~/app/composables/useGeoIP'
import type { GeoIPProvider } from '~/app/types/geoip'

const {
  isEnabled,
  isLoading,
  error,
  cacheStats,
  settings,
  isGeoIPEnabled,
  cacheHitRate,
  enableGeoIP,
  disableGeoIP,
  updateSettings: updateGeoIPSettings,
  refreshCacheStats,
  clearCache
} = useGeoIP()

const localSettings = ref({ ...settings.value })
const availableProviders = ref<GeoIPProvider[]>([])

// Computed
const selectedProvider = computed(() => {
  return availableProviders.value.find(p => p.name === localSettings.value.provider)
})

// Methods
const toggleGeoIP = async () => {
  if (localSettings.value.enabled) {
    await enableGeoIP()
  } else {
    await disableGeoIP()
  }
}

const updateSettings = async () => {
  await updateGeoIPSettings(localSettings.value)
}

const formatTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleString()
}

// Watch for settings changes from external sources
watch(settings, (newSettings) => {
  localSettings.value = { ...newSettings }
}, { deep: true })

// Initialize providers
onMounted(() => {
  availableProviders.value = GeoIPUtils.getAvailableProviders()
})
</script>

<style scoped>
.geoip-settings {
  padding: 1rem;
  background: var(--surface-primary);
  border-radius: 8px;
  border: 1px solid var(--border-secondary);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-secondary);
}

.settings-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.toggle-container {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.toggle {
  position: relative;
  display: inline-block;
  width: 48px;
  height: 24px;
}

.toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--border-secondary);
  transition: 0.3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--accent-primary);
}

input:checked + .slider:before {
  transform: translateX(24px);
}

.toggle-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  margin-bottom: 1rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--danger-primary);
  border-radius: 4px;
  color: var(--danger-primary);
  font-size: 0.875rem;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.setting-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.setting-label {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.setting-select {
  padding: 0.5rem;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-primary);
  font-size: 0.875rem;
}

.setting-select:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.setting-description {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.4;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.checkbox-item input[type="checkbox"] {
  display: none;
}

.checkmark {
  width: 16px;
  height: 16px;
  border: 2px solid var(--border-secondary);
  border-radius: 3px;
  position: relative;
  transition: all 0.2s ease;
}

.checkbox-item input[type="checkbox"]:checked + .checkmark {
  background-color: var(--accent-primary);
  border-color: var(--accent-primary);
}

.checkbox-item input[type="checkbox"]:checked + .checkmark:after {
  content: "";
  position: absolute;
  left: 4px;
  top: 1px;
  width: 4px;
  height: 8px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.cache-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.cache-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.stat-label {
  color: var(--text-secondary);
  font-weight: 500;
}

.stat-value {
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
}

.cache-actions {
  display: flex;
  gap: 0.5rem;
}

.btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary {
  background: var(--surface-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-secondary);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--surface-tertiary);
  border-color: var(--border-primary);
}

.btn-danger {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger-primary);
  border: 1px solid var(--danger-primary);
}

.btn-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.2);
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

.disabled-message {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--surface-secondary);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  line-height: 1.5;
}

.disabled-message svg {
  flex-shrink: 0;
  margin-top: 0.125rem;
  color: var(--text-tertiary);
}

.disabled-message p {
  margin: 0;
}
</style>