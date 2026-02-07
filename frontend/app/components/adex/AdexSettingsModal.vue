<template>
  <Transition name="settings-modal">
    <div v-if="modelValue" class="settings-overlay" @click.self="handleOverlayClick">
      <div class="settings-modal">
        <!-- Header -->
        <div class="settings-header">
          <div class="settings-header-title">
            <span class="settings-header-bracket">[</span>
            SYSTEM CONFIGURATION
            <span class="settings-header-bracket">]</span>
          </div>
          <button class="settings-close-btn" @click="cancel" title="Close settings">
            X
          </button>
        </div>

        <!-- Body: sidebar + content -->
        <div class="settings-body">
          <!-- Category Sidebar -->
          <div class="settings-sidebar">
            <div
              v-for="cat in categories"
              :key="cat.id"
              class="settings-sidebar-item"
              :class="{ active: activeCategory === cat.id }"
              @click="activeCategory = cat.id"
            >
              <span class="sidebar-marker">{{ activeCategory === cat.id ? '>' : ' ' }}</span>
              {{ cat.label }}
            </div>
          </div>

          <!-- Settings Content -->
          <div class="settings-content">
            <!-- SHELL -->
            <div v-show="activeCategory === 'shell'" class="settings-section">
              <div class="settings-section-title">SHELL CONFIGURATION</div>

              <div class="settings-field">
                <label class="settings-label">Shell Path</label>
                <input
                  v-model="localSettings.shell.path"
                  type="text"
                  class="settings-input"
                  placeholder="/bin/bash"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Shell Arguments</label>
                <input
                  v-model="localSettings.shell.args"
                  type="text"
                  class="settings-input"
                  placeholder="--login"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Working Directory</label>
                <input
                  v-model="localSettings.shell.workingDirectory"
                  type="text"
                  class="settings-input"
                  placeholder="/"
                />
              </div>
            </div>

            <!-- DISPLAY -->
            <div v-show="activeCategory === 'display'" class="settings-section">
              <div class="settings-section-title">DISPLAY SETTINGS</div>

              <div class="settings-field">
                <label class="settings-label">Theme</label>
                <select v-model="localSettings.display.theme" class="settings-select">
                  <option value="default">Default</option>
                  <option value="tron">Tron</option>
                  <option value="tron-disrupted">Tron Disrupted</option>
                  <option value="blade">Blade</option>
                  <option value="aDex-UI">aDex-UI Classic</option>
                  <option value="interstellar">Interstellar</option>
                  <option value="red-alert">Red Alert</option>
                  <option value="forest">Forest</option>
                  <option value="custom">Custom</option>
                </select>
              </div>

              <div class="settings-field">
                <label class="settings-label">Keyboard Layout</label>
                <select v-model="localSettings.display.keyboardLayout" class="settings-select">
                  <option value="en-US">US English (QWERTY)</option>
                  <option value="en-GB">UK English</option>
                  <option value="de-DE">German (QWERTZ)</option>
                  <option value="fr-FR">French (AZERTY)</option>
                  <option value="es-ES">Spanish</option>
                  <option value="pt-BR">Portuguese (Brazil)</option>
                  <option value="ja-JP">Japanese</option>
                  <option value="ko-KR">Korean</option>
                  <option value="zh-CN">Chinese (Simplified)</option>
                </select>
              </div>

              <div class="settings-field">
                <label class="settings-label">Terminal Font Size ({{ localSettings.display.terminalFontSize }}px)</label>
                <input
                  v-model.number="localSettings.display.terminalFontSize"
                  type="range"
                  class="settings-slider"
                  min="8"
                  max="24"
                  step="1"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Terminal Font Family</label>
                <select v-model="localSettings.display.fontFamily" class="settings-select">
                  <option value="'Fira Code', monospace">Fira Code</option>
                  <option value="'JetBrains Mono', monospace">JetBrains Mono</option>
                  <option value="'Cascadia Code', monospace">Cascadia Code</option>
                  <option value="'Source Code Pro', monospace">Source Code Pro</option>
                  <option value="'Consolas', monospace">Consolas</option>
                  <option value="'Monaco', monospace">Monaco</option>
                  <option value="monospace">System Monospace</option>
                </select>
              </div>
            </div>

            <!-- AUDIO -->
            <div v-show="activeCategory === 'audio'" class="settings-section">
              <div class="settings-section-title">AUDIO SETTINGS</div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Enable Audio</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.audio.enabled }"
                  @click="localSettings.audio.enabled = !localSettings.audio.enabled"
                >
                  {{ localSettings.audio.enabled ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field">
                <label class="settings-label">Master Volume ({{ localSettings.audio.volume }}%)</label>
                <input
                  v-model.number="localSettings.audio.volume"
                  type="range"
                  class="settings-slider"
                  min="0"
                  max="100"
                  step="1"
                  :disabled="!localSettings.audio.enabled"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Soundpack</label>
                <select
                  v-model="localSettings.audio.soundpack"
                  class="settings-select"
                  :disabled="!localSettings.audio.enabled"
                >
                  <option value="default">Default</option>
                  <option value="mechanical">Mechanical</option>
                  <option value="sci-fi">Sci-Fi</option>
                  <option value="minimal">Minimal</option>
                  <option value="none">None</option>
                </select>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Mute in Background</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.audio.muteInBackground }"
                  :disabled="!localSettings.audio.enabled"
                  @click="localSettings.audio.muteInBackground = !localSettings.audio.muteInBackground"
                >
                  {{ localSettings.audio.muteInBackground ? 'ON' : 'OFF' }}
                </button>
              </div>
            </div>

            <!-- SYSTEM -->
            <div v-show="activeCategory === 'system'" class="settings-section">
              <div class="settings-section-title">SYSTEM SETTINGS</div>

              <div class="settings-field">
                <label class="settings-label">Clock Format</label>
                <select v-model="localSettings.system.clockFormat" class="settings-select">
                  <option value="24h">24-Hour</option>
                  <option value="12h">12-Hour (AM/PM)</option>
                </select>
              </div>

              <div class="settings-field">
                <label class="settings-label">Ping Address</label>
                <input
                  v-model="localSettings.system.pingAddress"
                  type="text"
                  class="settings-input"
                  placeholder="1.1.1.1"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Boot Animation</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.system.bootAnimation }"
                  @click="localSettings.system.bootAnimation = !localSettings.system.bootAnimation"
                >
                  {{ localSettings.system.bootAnimation ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Show Grid Background</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.system.gridBackground }"
                  @click="localSettings.system.gridBackground = !localSettings.system.gridBackground"
                >
                  {{ localSettings.system.gridBackground ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Performance Mode</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.system.performanceMode }"
                  @click="localSettings.system.performanceMode = !localSettings.system.performanceMode"
                >
                  {{ localSettings.system.performanceMode ? 'ON' : 'OFF' }}
                </button>
              </div>
            </div>

            <!-- ADVANCED -->
            <div v-show="activeCategory === 'advanced'" class="settings-section">
              <div class="settings-section-title">ADVANCED SETTINGS</div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Allow Windowed Mode</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.advanced.allowWindowedMode }"
                  @click="localSettings.advanced.allowWindowedMode = !localSettings.advanced.allowWindowedMode"
                >
                  {{ localSettings.advanced.allowWindowedMode ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Experimental Features</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.advanced.experimentalFeatures }"
                  @click="localSettings.advanced.experimentalFeatures = !localSettings.advanced.experimentalFeatures"
                >
                  {{ localSettings.advanced.experimentalFeatures ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">WebGL Renderer</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.advanced.webglRenderer }"
                  @click="localSettings.advanced.webglRenderer = !localSettings.advanced.webglRenderer"
                >
                  {{ localSettings.advanced.webglRenderer ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Debug Mode</label>
                <button
                  class="settings-toggle"
                  :class="{ on: localSettings.advanced.debugMode }"
                  @click="localSettings.advanced.debugMode = !localSettings.advanced.debugMode"
                >
                  {{ localSettings.advanced.debugMode ? 'ON' : 'OFF' }}
                </button>
              </div>

              <div class="settings-field">
                <label class="settings-label">Terminal Scrollback ({{ localSettings.advanced.scrollback }} lines)</label>
                <input
                  v-model.number="localSettings.advanced.scrollback"
                  type="range"
                  class="settings-slider"
                  min="500"
                  max="50000"
                  step="500"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Footer: action buttons -->
        <div class="settings-footer">
          <button class="settings-btn settings-btn-secondary" @click="resetToDefaults">
            RESET DEFAULTS
          </button>
          <div class="settings-footer-spacer" />
          <button class="settings-btn settings-btn-secondary" @click="cancel">
            CANCEL
          </button>
          <button class="settings-btn settings-btn-primary" @click="save">
            SAVE
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'

// ---- Types ----

interface SettingsData {
  shell: {
    path: string
    args: string
    workingDirectory: string
  }
  display: {
    theme: string
    keyboardLayout: string
    terminalFontSize: number
    fontFamily: string
  }
  audio: {
    enabled: boolean
    volume: number
    soundpack: string
    muteInBackground: boolean
  }
  system: {
    clockFormat: '12h' | '24h'
    pingAddress: string
    bootAnimation: boolean
    gridBackground: boolean
    performanceMode: boolean
  }
  advanced: {
    allowWindowedMode: boolean
    experimentalFeatures: boolean
    webglRenderer: boolean
    debugMode: boolean
    scrollback: number
  }
}

interface Category {
  id: string
  label: string
}

// ---- Props / Emits ----

const props = withDefaults(
  defineProps<{
    modelValue: boolean
  }>(),
  {
    modelValue: false,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
  (e: 'settings-changed', settings: SettingsData): void
}>()

// ---- Constants ----

const STORAGE_KEY = 'adex-settings'

const categories: Category[] = [
  { id: 'shell', label: 'SHELL' },
  { id: 'display', label: 'DISPLAY' },
  { id: 'audio', label: 'AUDIO' },
  { id: 'system', label: 'SYSTEM' },
  { id: 'advanced', label: 'ADVANCED' },
]

// ---- State ----

const activeCategory = ref('shell')

function getDefaultSettings(): SettingsData {
  return {
    shell: {
      path: '/bin/bash',
      args: '--login',
      workingDirectory: '/',
    },
    display: {
      theme: 'default',
      keyboardLayout: 'en-US',
      terminalFontSize: 14,
      fontFamily: "'Fira Code', monospace",
    },
    audio: {
      enabled: false,
      volume: 50,
      soundpack: 'default',
      muteInBackground: false,
    },
    system: {
      clockFormat: '24h',
      pingAddress: '1.1.1.1',
      bootAnimation: true,
      gridBackground: true,
      performanceMode: false,
    },
    advanced: {
      allowWindowedMode: false,
      experimentalFeatures: false,
      webglRenderer: true,
      debugMode: false,
      scrollback: 5000,
    },
  }
}

const localSettings = reactive<SettingsData>(getDefaultSettings())

// ---- Persistence ----

function loadFromStorage(): SettingsData {
  if (typeof localStorage === 'undefined') return getDefaultSettings()
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const parsed = JSON.parse(stored) as Partial<SettingsData>
      const defaults = getDefaultSettings()
      return {
        shell: { ...defaults.shell, ...parsed.shell },
        display: { ...defaults.display, ...parsed.display },
        audio: { ...defaults.audio, ...parsed.audio },
        system: { ...defaults.system, ...parsed.system },
        advanced: { ...defaults.advanced, ...parsed.advanced },
      }
    }
  } catch (err) {
    console.error('Failed to load settings from localStorage:', err)
  }
  return getDefaultSettings()
}

function saveToStorage(settings: SettingsData) {
  if (typeof localStorage === 'undefined') return
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
  } catch (err) {
    console.error('Failed to save settings to localStorage:', err)
  }
}

// ---- Actions ----

function save() {
  const snapshot: SettingsData = JSON.parse(JSON.stringify(localSettings))
  saveToStorage(snapshot)
  emit('settings-changed', snapshot)
  closeModal()
}

function cancel() {
  // Revert to stored settings
  const stored = loadFromStorage()
  Object.assign(localSettings, stored)
  closeModal()
}

function resetToDefaults() {
  const defaults = getDefaultSettings()
  Object.assign(localSettings, defaults)
}

function closeModal() {
  emit('update:modelValue', false)
  emit('close')
}

function handleOverlayClick() {
  cancel()
}

// ---- Load on open ----

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      const stored = loadFromStorage()
      Object.assign(localSettings, stored)
      activeCategory.value = 'shell'
    }
  }
)

// ---- Init ----

onMounted(() => {
  const stored = loadFromStorage()
  Object.assign(localSettings, stored)
})
</script>

<style scoped>
/* Overlay */
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.75);
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Modal container */
.settings-modal {
  width: 58vw;
  max-height: 78vh;
  background: var(--color_light_black, #05080d);
  border: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  /* Sci-fi corner cuts */
  clip-path: polygon(
    0% 1.5vh,
    1.5vh 0%,
    calc(100% - 3vh) 0%,
    100% 3vh,
    100% calc(100% - 1.5vh),
    calc(100% - 1.5vh) 100%,
    3vh 100%,
    0% calc(100% - 3vh)
  );
}

/* Outer glow border effect */
.settings-modal::before {
  content: '';
  position: absolute;
  top: -0.15vh;
  left: -0.15vh;
  right: -0.15vh;
  bottom: -0.15vh;
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  clip-path: polygon(
    0% 1.5vh,
    1.5vh 0%,
    calc(100% - 3vh) 0%,
    100% 3vh,
    100% calc(100% - 1.5vh),
    calc(100% - 1.5vh) 100%,
    3vh 100%,
    0% calc(100% - 3vh)
  );
  pointer-events: none;
  z-index: -1;
}

/* Header */
.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 4vh;
  padding: 0 2vw;
  border-bottom: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  flex-shrink: 0;
}

.settings-header-title {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.3vh;
  color: var(--color_accent, rgb(170, 207, 209));
  text-transform: uppercase;
  letter-spacing: 0.3em;
}

.settings-header-bracket {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
}

.settings-close-btn {
  background: none;
  border: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  width: 2.5vh;
  height: 2.5vh;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.settings-close-btn:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Body */
.settings-body {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* Sidebar */
.settings-sidebar {
  width: 12vw;
  border-right: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  padding: 1vh 0;
  flex-shrink: 0;
  overflow-y: auto;
}

.settings-sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.5vw;
  padding: 0.8vh 1.2vw;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
  text-transform: uppercase;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  user-select: none;
}

.settings-sidebar-item:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.08);
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-sidebar-item.active {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  color: var(--color_accent, rgb(170, 207, 209));
}

.sidebar-marker {
  font-size: 1.2vh;
  width: 1vw;
  flex-shrink: 0;
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Content */
.settings-content {
  flex: 1;
  padding: 1.5vh 2vw;
  overflow-y: auto;
  min-height: 0;
}

.settings-section-title {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.2vh;
  color: var(--color_accent, rgb(170, 207, 209));
  text-transform: uppercase;
  letter-spacing: 0.2em;
  padding-bottom: 0.8vh;
  margin-bottom: 1.5vh;
  border-bottom: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
}

/* Fields */
.settings-field {
  margin-bottom: 1.5vh;
}

.settings-field-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.settings-label {
  display: block;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.05vh;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.7);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 0.5vh;
}

.settings-field-toggle .settings-label {
  margin-bottom: 0;
}

/* Input */
.settings-input {
  width: 100%;
  height: 3vh;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  padding: 0 0.8vw;
  outline: none;
  transition: border-color 0.15s, background 0.15s;
}

.settings-input:focus {
  border-color: var(--color_accent, rgb(170, 207, 209));
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.08);
}

.settings-input::placeholder {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.25);
}

/* Select */
.settings-select {
  width: 100%;
  height: 3vh;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  padding: 0 0.8vw;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23aacfd1' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.6vw center;
  background-size: 1vh;
  transition: border-color 0.15s;
}

.settings-select:focus {
  border-color: var(--color_accent, rgb(170, 207, 209));
}

.settings-select:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.settings-select option {
  background: var(--color_light_black, #05080d);
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Slider */
.settings-slider {
  width: 100%;
  height: 0.5vh;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  outline: none;
  border-radius: 0;
  margin-top: 0.5vh;
}

.settings-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 1.2vh;
  height: 2vh;
  background: var(--color_accent, rgb(170, 207, 209));
  cursor: pointer;
  border: none;
}

.settings-slider::-moz-range-thumb {
  width: 1.2vh;
  height: 2vh;
  background: var(--color_accent, rgb(170, 207, 209));
  cursor: pointer;
  border: none;
  border-radius: 0;
}

.settings-slider:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.settings-slider:disabled::-webkit-slider-thumb {
  cursor: not-allowed;
}

/* Toggle */
.settings-toggle {
  min-width: 5vw;
  height: 2.5vh;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1vh;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: all 0.15s;
  flex-shrink: 0;
}

.settings-toggle.on {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  border-color: var(--color_accent, rgb(170, 207, 209));
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-toggle:hover:not(:disabled) {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.12);
}

.settings-toggle:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

/* Footer */
.settings-footer {
  display: flex;
  align-items: center;
  gap: 1vw;
  padding: 1vh 2vw;
  border-top: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.03);
  flex-shrink: 0;
}

.settings-footer-spacer {
  flex: 1;
}

.settings-btn {
  height: 3vh;
  padding: 0 1.5vw;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.05vh;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: all 0.15s;
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  background: transparent;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.7);
}

.settings-btn:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.1);
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-btn-primary {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  border-color: var(--color_accent, rgb(170, 207, 209));
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-btn-primary:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
}

.settings-btn-secondary {
  border-color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
}

.settings-btn-secondary:hover {
  border-color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.8);
}

/* Modal transition */
.settings-modal-enter-active {
  transition: all 0.3s cubic-bezier(0.19, 1, 0.22, 1);
}

.settings-modal-leave-active {
  transition: all 0.2s ease-in;
}

.settings-modal-enter-from {
  opacity: 0;
}

.settings-modal-enter-from .settings-modal {
  transform: scale(0.95) translateY(1vh);
  opacity: 0;
}

.settings-modal-leave-to {
  opacity: 0;
}

.settings-modal-leave-to .settings-modal {
  transform: scale(0.95) translateY(1vh);
  opacity: 0;
}

/* Scrollbar in settings */
.settings-content::-webkit-scrollbar,
.settings-sidebar::-webkit-scrollbar {
  width: 0.3vw;
}

.settings-content::-webkit-scrollbar-track,
.settings-sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.settings-content::-webkit-scrollbar-thumb,
.settings-sidebar::-webkit-scrollbar-thumb {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
}

.settings-content::-webkit-scrollbar-thumb:hover,
.settings-sidebar::-webkit-scrollbar-thumb:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
}
</style>
