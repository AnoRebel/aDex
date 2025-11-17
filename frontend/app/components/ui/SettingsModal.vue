<template>
  <div v-if="isVisible" class="settings-overlay" @click="closeModal">
    <div class="settings-modal" @click.stop>
      <div class="modal-header">
        <h2 class="modal-title">
          <span class="glitch">SYSTEM CONFIGURATION</span>
        </h2>
        <button @click="closeModal" class="close-btn">
          <span class="close-icon">✕</span>
        </button>
      </div>

      <div class="modal-content">
        <!-- Settings Navigation -->
        <div class="settings-nav">
          <div class="nav-tabs">
            <button
              v-for="category in settingsCategories"
              :key="category.id"
              @click="activeCategory = category.id"
              class="nav-tab"
              :class="{ active: activeCategory === category.id }"
            >
              <span class="nav-icon">{{ category.icon }}</span>
              <span class="nav-label">{{ category.name }}</span>
            </button>
          </div>
        </div>

        <!-- Settings Content -->
        <div class="settings-content">
          <!-- General Settings -->
          <div v-if="activeCategory === 'general'" class="settings-section">
            <h3 class="section-title">General Settings</h3>

            <div class="setting-group">
              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Application Language</span>
                  <span class="setting-description">Select your preferred language</span>
                </div>
                <select v-model="settings.language" @change="saveSettings" class="setting-select">
                  <option value="en">English</option>
                  <option value="es">Español</option>
                  <option value="fr">Français</option>
                  <option value="de">Deutsch</option>
                  <option value="ja">日本語</option>
                  <option value="zh">中文</option>
                </select>
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Startup Behavior</span>
                  <span class="setting-description">What happens when the application starts</span>
                </div>
                <select v-model="settings.startupBehavior" @change="saveSettings" class="setting-select">
                  <option value="fullscreen">Start in fullscreen</option>
                  <option value="windowed">Start in windowed mode</option>
                  <option value="last">Remember last state</option>
                </select>
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Auto-save Interval</span>
                  <span class="setting-description">How often to save application state</span>
                </div>
                <select v-model="settings.autoSaveInterval" @change="saveSettings" class="setting-select">
                  <option value="disabled">Disabled</option>
                  <option value="1">Every 1 minute</option>
                  <option value="5">Every 5 minutes</option>
                  <option value="10">Every 10 minutes</option>
                  <option value="30">Every 30 minutes</option>
                </select>
              </label>
            </div>
          </div>

          <!-- Terminal Settings -->
          <div v-if="activeCategory === 'terminal'" class="settings-section">
            <h3 class="section-title">Terminal Configuration</h3>

            <div class="setting-group">
              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Default Shell</span>
                  <span class="setting-description">Shell to use for new terminals</span>
                </div>
                <select v-model="settings.shell" @change="saveSettings" class="setting-select">
                  <option value="bash">Bash</option>
                  <option value="zsh">Zsh</option>
                  <option value="fish">Fish</option>
                  <option value="powershell">PowerShell</option>
                  <option value="cmd">Command Prompt</option>
                </select>
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Font Size</span>
                  <span class="setting-description">Terminal font size in pixels</span>
                </div>
                <div class="setting-control">
                  <input
                    v-model="settings.fontSize"
                    type="range"
                    min="8"
                    max="24"
                    @input="saveSettings"
                    class="setting-slider"
                  />
                  <span class="setting-value">{{ settings.fontSize }}px</span>
                </div>
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Terminal Opacity</span>
                  <span class="setting-description">Background transparency level</span>
                </div>
                <div class="setting-control">
                  <input
                    v-model="settings.terminalOpacity"
                    type="range"
                    min="0.1"
                    max="1"
                    step="0.1"
                    @input="saveSettings"
                    class="setting-slider"
                  />
                  <span class="setting-value">{{ Math.round(settings.terminalOpacity * 100) }}%</span>
                </div>
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Blink Cursor</span>
                  <span class="setting-description">Enable cursor blinking animation</span>
                </div>
                <input
                  v-model="settings.blinkCursor"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Bell Sound</span>
                  <span class="setting-description">Play sound on bell character</span>
                </div>
                <input
                  v-model="settings.bellSound"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>
            </div>
          </div>

          <!-- Network Settings -->
          <div v-if="activeCategory === 'network'" class="settings-section">
            <h3 class="section-title">Network Configuration</h3>

            <div class="setting-group">
              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Ping Address</span>
                  <span class="setting-description">Address for latency testing</span>
                </div>
                <input
                  v-model="settings.pingAddress"
                  type="text"
                  @change="saveSettings"
                  placeholder="1.1.1.1"
                  class="setting-input"
                />
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Update Interval</span>
                  <span class="setting-description">Network monitoring update frequency</span>
                </div>
                <select v-model="settings.networkUpdateInterval" @change="saveSettings" class="setting-select">
                  <option value="1000">1 second</option>
                  <option value="2000">2 seconds</option>
                  <option value="5000">5 seconds</option>
                  <option value="10000">10 seconds</option>
                </select>
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Show GeoIP</span>
                  <span class="setting-description">Display geographical location information</span>
                </div>
                <input
                  v-model="settings.showGeoIP"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Track Bandwidth</span>
                  <span class="setting-description">Monitor network bandwidth usage</span>
                </div>
                <input
                  v-model="settings.trackBandwidth"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>
            </div>
          </div>

          <!-- Performance Settings -->
          <div v-if="activeCategory === 'performance'" class="settings-section">
            <h3 class="section-title">Performance Optimization</h3>

            <div class="setting-group">
              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Rendering Mode</span>
                  <span class="setting-description">Choose performance vs quality balance</span>
                </div>
                <select v-model="settings.renderingMode" @change="saveSettings" class="setting-select">
                  <option value="performance">Performance (60fps)</option>
                  <option value="balanced">Balanced (30fps)</option>
                  <option value="quality">Quality (uncapped)</option>
                </select>
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Hardware Acceleration</span>
                  <span class="setting-description">Use GPU for rendering acceleration</span>
                </div>
                <input
                  v-model="settings.hardwareAcceleration"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Reduce Motion</span>
                  <span class="setting-description">Disable animations for better performance</span>
                </div>
                <input
                  v-model="settings.reduceMotion"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Particle Effects</span>
                  <span class="setting-description">Enable background particle animations</span>
                </div>
                <input
                  v-model="settings.particleEffects"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Max Process History</span>
                  <span class="setting-description">Maximum number of processes to track</span>
                </div>
                <div class="setting-control">
                  <input
                    v-model="settings.maxProcessHistory"
                    type="range"
                    min="50"
                    max="500"
                    step="10"
                    @input="saveSettings"
                    class="setting-slider"
                  />
                  <span class="setting-value">{{ settings.maxProcessHistory }}</span>
                </div>
              </label>
            </div>
          </div>

          <!-- Accessibility Settings -->
          <div v-if="activeCategory === 'accessibility'" class="settings-section">
            <h3 class="section-title">Accessibility Options</h3>

            <div class="setting-group">
              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">UI Scale</span>
                  <span class="setting-description">Scale factor for user interface</span>
                </div>
                <div class="setting-control">
                  <input
                    v-model="settings.uiScale"
                    type="range"
                    min="0.8"
                    max="1.5"
                    step="0.1"
                    @input="saveSettings"
                    class="setting-slider"
                  />
                  <span class="setting-value">{{ Math.round(settings.uiScale * 100) }}%</span>
                </div>
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">High Contrast</span>
                  <span class="setting-description">Increase contrast for better visibility</span>
                </div>
                <input
                  v-model="settings.highContrast"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Screen Reader Support</span>
                  <span class="setting-description">Optimize for screen reader usage</span>
                </div>
                <input
                  v-model="settings.screenReader"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Focus Indicators</span>
                  <span class="setting-description">Show enhanced focus indicators</span>
                </div>
                <input
                  v-model="settings.focusIndicators"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>
            </div>
          </div>

          <!-- Advanced Settings -->
          <div v-if="activeCategory === 'advanced'" class="settings-section">
            <h3 class="section-title">Advanced Configuration</h3>

            <div class="setting-group">
              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Log Level</span>
                  <span class="setting-description">Minimum severity for log messages</span>
                </div>
                <select v-model="settings.logLevel" @change="saveSettings" class="setting-select">
                  <option value="error">Error</option>
                  <option value="warn">Warning</option>
                  <option value="info">Info</option>
                  <option value="debug">Debug</option>
                  <option value="trace">Trace</option>
                </select>
              </label>

              <label class="setting-item">
                <div class="setting-info">
                  <span class="setting-name">Cache Directory</span>
                  <span class="setting-description">Directory for temporary files</span>
                </div>
                <input
                  v-model="settings.cacheDirectory"
                  type="text"
                  @change="saveSettings"
                  class="setting-input"
                />
              </label>

              <label class="setting-item toggle">
                <div class="setting-info">
                  <span class="setting-name">Enable Debug Mode</span>
                  <span class="setting-description">Show additional debug information</span>
                </div>
                <input
                  v-model="settings.debugMode"
                  type="checkbox"
                  @change="saveSettings"
                  class="setting-toggle"
                />
              </label>

              <div class="setting-actions">
                <button @click="exportSettings" class="action-btn secondary">
                  <span class="btn-icon">📤</span>
                  Export Settings
                </button>
                <button @click="importSettings" class="action-btn secondary">
                  <span class="btn-icon">📥</span>
                  Import Settings
                </button>
                <button @click="resetSettings" class="action-btn danger">
                  <span class="btn-icon">🔄</span>
                  Reset to Defaults
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <div class="footer-info">
          <span class="version-info">Version {{ appVersion }}</span>
          <span class="settings-status">Settings {{ settingsSaved ? 'saved' : 'modified' }}</span>
        </div>
        <div class="footer-actions">
          <button @click="closeModal" class="action-btn secondary">
            Cancel
          </button>
          <button @click="applySettings" class="action-btn primary">
            Apply Changes
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Settings {
  // General
  language: string
  startupBehavior: string
  autoSaveInterval: string

  // Terminal
  shell: string
  fontSize: number
  terminalOpacity: number
  blinkCursor: boolean
  bellSound: boolean

  // Network
  pingAddress: string
  networkUpdateInterval: number
  showGeoIP: boolean
  trackBandwidth: boolean

  // Performance
  renderingMode: string
  hardwareAcceleration: boolean
  reduceMotion: boolean
  particleEffects: boolean
  maxProcessHistory: number

  // Accessibility
  uiScale: number
  highContrast: boolean
  screenReader: boolean
  focusIndicators: boolean

  // Advanced
  logLevel: string
  cacheDirectory: string
  debugMode: boolean
}

interface SettingsCategory {
  id: string
  name: string
  icon: string
}

// Props
const emit = defineEmits<{
  settingsChanged: [settings: Settings]
}>()

// Reactive data
const isVisible = ref<boolean>(false)
const activeCategory = ref<string>('general')
const settingsSaved = ref<boolean>(true)
const appVersion = ref<string>('2.0.0')

const settings = ref<Settings>({
  // General
  language: 'en',
  startupBehavior: 'fullscreen',
  autoSaveInterval: '5',

  // Terminal
  shell: 'bash',
  fontSize: 14,
  terminalOpacity: 0.8,
  blinkCursor: true,
  bellSound: true,

  // Network
  pingAddress: '1.1.1.1',
  networkUpdateInterval: 2000,
  showGeoIP: true,
  trackBandwidth: true,

  // Performance
  renderingMode: 'balanced',
  hardwareAcceleration: true,
  reduceMotion: false,
  particleEffects: true,
  maxProcessHistory: 200,

  // Accessibility
  uiScale: 1.0,
  highContrast: false,
  screenReader: false,
  focusIndicators: true,

  // Advanced
  logLevel: 'info',
  cacheDirectory: '/tmp/adex-ui',
  debugMode: false
})

const settingsCategories: SettingsCategory[] = [
  { id: 'general', name: 'General', icon: '⚙️' },
  { id: 'terminal', name: 'Terminal', icon: '💻' },
  { id: 'network', name: 'Network', icon: '🌐' },
  { id: 'performance', name: 'Performance', icon: '⚡' },
  { id: 'accessibility', name: 'Accessibility', icon: '♿' },
  { id: 'advanced', name: 'Advanced', icon: '🔧' }
]

// Methods
const showModal = () => {
  isVisible.value = true
}

const closeModal = () => {
  isVisible.value = false
}

const saveSettings = () => {
  settingsSaved.value = false
  localStorage.setItem('adex-settings', JSON.stringify(settings.value))
  emit('settingsChanged', settings.value)
}

const applySettings = () => {
  saveSettings()
  settingsSaved.value = true

  // Apply UI scale
  document.documentElement.style.setProperty('--ui-scale', settings.value.uiScale.toString())

  // Apply high contrast
  document.body.classList.toggle('high-contrast', settings.value.highContrast)

  // Apply reduce motion
  document.body.classList.toggle('reduce-motion', settings.value.reduceMotion)

  // Apply screen reader optimizations
  document.body.classList.toggle('screen-reader', settings.value.screenReader)

  // Apply focus indicators
  document.body.classList.toggle('enhanced-focus', settings.value.focusIndicators)

  closeModal()
}

const exportSettings = () => {
  const settingsData = JSON.stringify(settings.value, null, 2)
  const blob = new Blob([settingsData], { type: 'application/json' })
  const url = URL.createObjectURL(blob)

  const a = document.createElement('a')
  a.href = url
  a.download = `adex-settings-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)

  URL.revokeObjectURL(url)
}

const importSettings = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'

  input.onchange = (event) => {
    const file = (event.target as HTMLInputElement).files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      try {
        const importedSettings = JSON.parse(e.target?.result as string)
        settings.value = { ...settings.value, ...importedSettings }
        saveSettings()
      } catch (error) {
        console.error('Failed to import settings:', error)
      }
    }
    reader.readAsText(file)
  }

  input.click()
}

const resetSettings = () => {
  if (confirm('Are you sure you want to reset all settings to defaults? This cannot be undone.')) {
    // Reset to default values
    settings.value = {
      // General
      language: 'en',
      startupBehavior: 'fullscreen',
      autoSaveInterval: '5',

      // Terminal
      shell: 'bash',
      fontSize: 14,
      terminalOpacity: 0.8,
      blinkCursor: true,
      bellSound: true,

      // Network
      pingAddress: '1.1.1.1',
      networkUpdateInterval: 2000,
      showGeoIP: true,
      trackBandwidth: true,

      // Performance
      renderingMode: 'balanced',
      hardwareAcceleration: true,
      reduceMotion: false,
      particleEffects: true,
      maxProcessHistory: 200,

      // Accessibility
      uiScale: 1.0,
      highContrast: false,
      screenReader: false,
      focusIndicators: true,

      // Advanced
      logLevel: 'info',
      cacheDirectory: '/tmp/adex-ui',
      debugMode: false
    }

    saveSettings()
    applySettings()
  }
}

const loadSettings = () => {
  const savedSettings = localStorage.getItem('adex-settings')
  if (savedSettings) {
    try {
      settings.value = { ...settings.value, ...JSON.parse(savedSettings) }
    } catch (error) {
      console.error('Failed to load settings:', error)
    }
  }

  // Apply loaded settings immediately
  applySettings()
}

// Public methods
defineExpose({
  showModal,
  closeModal
})

// Lifecycle
onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(5px);
}

.settings-modal {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  width: 90vw;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  font-family: 'Fira Code', monospace;
  color: var(--text-primary);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid var(--surface-border);
  background: var(--surface-elevated);
  border-radius: 12px 12px 0 0;
}

.modal-title {
  font-size: 16px;
  font-weight: bold;
  margin: 0;
}

.close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover {
  border-color: var(--error);
  color: var(--error);
}

.close-icon {
  font-size: 18px;
}

.modal-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.settings-nav {
  width: 200px;
  background: var(--surface);
  border-right: 1px solid var(--surface-border);
  padding: 20px 0;
}

.nav-tabs {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  font-size: 11px;
}

.nav-tab:hover {
  background: var(--surface-elevated);
  color: var(--text-primary);
}

.nav-tab.active {
  background: var(--primary-500);
  color: var(--text-primary);
  border-left: 3px solid var(--accent-500);
}

.nav-icon {
  font-size: 16px;
}

.nav-label {
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.settings-section {
  margin-bottom: 30px;
}

.section-title {
  font-size: 14px;
  font-weight: bold;
  color: var(--primary-400);
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--surface-border);
}

.setting-group {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  transition: all 0.2s ease;
}

.setting-item:hover {
  border-color: var(--primary-500);
}

.setting-item.toggle {
  cursor: pointer;
}

.setting-info {
  flex: 1;
}

.setting-name {
  font-size: 12px;
  font-weight: bold;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.setting-description {
  font-size: 10px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.setting-select,
.setting-input {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 11px;
  padding: 6px 10px;
  min-width: 150px;
  transition: all 0.2s ease;
}

.setting-select:focus,
.setting-input:focus {
  outline: none;
  border-color: var(--primary-500);
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.2);
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 150px;
}

.setting-slider {
  flex: 1;
  height: 4px;
  background: var(--surface);
  border-radius: 2px;
  outline: none;
  -webkit-appearance: none;
  appearance: none;
}

.setting-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 16px;
  height: 16px;
  background: var(--primary-500);
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.5);
}

.setting-slider::-moz-range-thumb {
  width: 16px;
  height: 16px;
  background: var(--primary-500);
  border-radius: 50%;
  cursor: pointer;
  border: none;
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.5);
}

.setting-value {
  font-size: 11px;
  font-weight: bold;
  color: var(--primary-400);
  min-width: 45px;
  text-align: right;
}

.setting-toggle {
  width: 20px;
  height: 20px;
  accent-color: var(--primary-500);
  cursor: pointer;
}

.setting-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--surface-border);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  font-size: 11px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn.primary {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

.action-btn.primary:hover {
  background: var(--primary-600);
}

.action-btn.secondary {
  background: var(--surface-elevated);
  color: var(--text-primary);
}

.action-btn.secondary:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.action-btn.danger {
  background: var(--error);
  border-color: var(--error);
  color: var(--text-primary);
}

.action-btn.danger:hover {
  background: #dc2626;
}

.btn-icon {
  font-size: 12px;
}

.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-top: 1px solid var(--surface-border);
  background: var(--surface-elevated);
  border-radius: 0 0 12px 12px;
}

.footer-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.version-info {
  font-size: 10px;
  color: var(--text-secondary);
}

.settings-status {
  font-size: 10px;
  font-weight: bold;
}

.settings-status:not(:empty) {
  color: var(--success);
}

.footer-actions {
  display: flex;
  gap: 10px;
}

/* Glitch effect for title */
.glitch {
  position: relative;
  color: var(--primary-400);
  font-size: 16px;
  font-weight: bold;
  text-transform: uppercase;
  text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  animation: glitch 2s infinite;
}

@keyframes glitch {
  0%, 90%, 100% {
    text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  }
  95% {
    text-shadow: -2px 2px 0 var(--accent-500), 2px -2px 0 var(--primary-600);
  }
}

/* Responsive design */
@media (max-width: 768px) {
  .settings-modal {
    width: 95vw;
    max-height: 95vh;
  }

  .modal-content {
    flex-direction: column;
  }

  .settings-nav {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--surface-border);
    padding: 10px 0;
  }

  .nav-tabs {
    flex-direction: row;
    overflow-x: auto;
    padding: 0 20px;
  }

  .nav-tab {
    flex-direction: column;
    gap: 4px;
    padding: 8px 12px;
    min-width: 80px;
    text-align: center;
  }

  .nav-label {
    font-size: 9px;
  }

  .setting-item {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .setting-select,
  .setting-input {
    width: 100%;
    min-width: auto;
  }

  .setting-control {
    min-width: auto;
  }

  .modal-footer {
    flex-direction: column;
    gap: 15px;
    text-align: center;
  }

  .footer-actions {
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .modal-header {
    padding: 15px;
  }

  .settings-content {
    padding: 15px;
  }

  .setting-item {
    padding: 12px;
  }

  .setting-actions {
    flex-direction: column;
  }
}
</style>