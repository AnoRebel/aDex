<template>
  <div class="dialog-overlay" @click="close">
    <div class="dialog-container" @click.stop>
      <!-- Dialog Header -->
      <div class="dialog-header">
        <h2 class="dialog-title">Color Scheme Preview</h2>
        <button class="close-button" @click="close" title="Close">
          <Icon name="carbon:close" />
        </button>
      </div>

      <!-- Dialog Content -->
      <div class="dialog-content">
        <div class="preview-section">
          <h3 class="section-title">Terminal Preview</h3>
          <div
            class="preview-terminal"
            :style="getTerminalStyles()"
          >
            <div class="terminal-header">
              <div class="terminal-controls">
                <div class="control-dot" style="background: #ff5f56;"></div>
                <div class="control-dot" style="background: #ffbd2e;"></div>
                <div class="control-dot" style="background: #27ca3f;"></div>
              </div>
              <div class="terminal-title">{{ scheme.displayName || scheme.name }}</div>
            </div>
            <div class="terminal-content">
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.green }">$</span>
                <span :style="{ color: scheme.colors.foreground }"> echo "Welcome to {{ scheme.displayName || scheme.name }}"</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.cyan }">Welcome to {{ scheme.displayName || scheme.name }}</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.green }">$</span>
                <span :style="{ color: scheme.colors.foreground }"> </span>
                <span :style="{ color: scheme.colors.blue }">ls</span>
                <span :style="{ color: scheme.colors.foreground }"> -la /home/user</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.foreground }">total 48</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.brightBlack }">drwxr-xr-x</span>
                <span :style="{ color: scheme.colors.foreground }"> 12 user staff  384 Jan 15 10:30</span>
                <span :style="{ color: scheme.colors.blue, fontWeight: 'bold' }">Documents</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.brightBlack }">drwxr-xr-x</span>
                <span :style="{ color: scheme.colors.foreground }">  8 user staff  256 Jan 15 10:30</span>
                <span :style="{ color: scheme.colors.blue, fontWeight: 'bold' }">Downloads</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.brightBlack }">-rw-r--r--</span>
                <span :style="{ color: scheme.colors.foreground }">  1 user staff 1024 Jan 15 10:30</span>
                <span :style="{ color: scheme.colors.white }">readme.txt</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.brightBlack }">-rwxr-xr-x</span>
                <span :style="{ color: scheme.colors.foreground }">  1 user staff 2048 Jan 15 10:30</span>
                <span :style="{ color: scheme.colors.green }">script.sh</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.green }">$</span>
                <span :style="{ color: scheme.colors.foreground }"> </span>
                <span :style="{ color: scheme.colors.blue }">cat</span>
                <span :style="{ color: scheme.colors.foreground }"> config.json</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.brightRed }">{</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.foreground }">  </span>
                <span :style="{ color: scheme.colors.yellow }">"theme"</span>
                <span :style="{ color: scheme.colors.foreground }">: </span>
                <span :style="{ color: scheme.colors.cyan }">"{{ scheme.name }}"</span>
                <span :style="{ color: scheme.colors.foreground }">,</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.foreground }">  </span>
                <span :style="{ color: scheme.colors.yellow }">"version"</span>
                <span :style="{ color: scheme.colors.foreground }">: </span>
                <span :style="{ color: scheme.colors.cyan }">"{{ scheme.version || '1.0.0' }}"</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.brightRed }">}</span>
              </div>
              <div class="terminal-line">
                <span :style="{ color: scheme.colors.green }">$</span>
                <span :style="{ color: scheme.colors.foreground, cursor: scheme.cursor.style === 'block' ? 'block' : 'inherit' }">_</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Color Palette -->
        <div class="palette-section">
          <h3 class="section-title">Color Palette</h3>
          <div class="color-grid">
            <div class="color-group">
              <h4 class="color-group-title">Basic Colors</h4>
              <div class="color-list">
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.background }"></div>
                  <span class="color-name">Background</span>
                  <span class="color-value">{{ scheme.colors.background }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.foreground }"></div>
                  <span class="color-name">Foreground</span>
                  <span class="color-value">{{ scheme.colors.foreground }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.cursor }"></div>
                  <span class="color-name">Cursor</span>
                  <span class="color-value">{{ scheme.colors.cursor }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.selection }"></div>
                  <span class="color-name">Selection</span>
                  <span class="color-value">{{ scheme.colors.selection }}</span>
                </div>
              </div>
            </div>

            <div class="color-group">
              <h4 class="color-group-title">Normal Colors</h4>
              <div class="color-list">
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.black }"></div>
                  <span class="color-name">Black</span>
                  <span class="color-value">{{ scheme.colors.black }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.red }"></div>
                  <span class="color-name">Red</span>
                  <span class="color-value">{{ scheme.colors.red }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.green }"></div>
                  <span class="color-name">Green</span>
                  <span class="color-value">{{ scheme.colors.green }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.yellow }"></div>
                  <span class="color-name">Yellow</span>
                  <span class="color-value">{{ scheme.colors.yellow }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.blue }"></div>
                  <span class="color-name">Blue</span>
                  <span class="color-value">{{ scheme.colors.blue }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.magenta }"></div>
                  <span class="color-name">Magenta</span>
                  <span class="color-value">{{ scheme.colors.magenta }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.cyan }"></div>
                  <span class="color-name">Cyan</span>
                  <span class="color-value">{{ scheme.colors.cyan }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.white }"></div>
                  <span class="color-name">White</span>
                  <span class="color-value">{{ scheme.colors.white }}</span>
                </div>
              </div>
            </div>

            <div class="color-group">
              <h4 class="color-group-title">Bright Colors</h4>
              <div class="color-list">
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightBlack }"></div>
                  <span class="color-name">Bright Black</span>
                  <span class="color-value">{{ scheme.colors.brightBlack }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightRed }"></div>
                  <span class="color-name">Bright Red</span>
                  <span class="color-value">{{ scheme.colors.brightRed }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightGreen }"></div>
                  <span class="color-name">Bright Green</span>
                  <span class="color-value">{{ scheme.colors.brightGreen }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightYellow }"></div>
                  <span class="color-name">Bright Yellow</span>
                  <span class="color-value">{{ scheme.colors.brightYellow }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightBlue }"></div>
                  <span class="color-name">Bright Blue</span>
                  <span class="color-value">{{ scheme.colors.brightBlue }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightMagenta }"></div>
                  <span class="color-name">Bright Magenta</span>
                  <span class="color-value">{{ scheme.colors.brightMagenta }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightCyan }"></div>
                  <span class="color-name">Bright Cyan</span>
                  <span class="color-value">{{ scheme.colors.brightCyan }}</span>
                </div>
                <div class="color-item">
                  <div class="color-swatch" :style="{ backgroundColor: scheme.colors.brightWhite }"></div>
                  <span class="color-name">Bright White</span>
                  <span class="color-value">{{ scheme.colors.brightWhite }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Scheme Information -->
        <div class="info-section">
          <h3 class="section-title">Scheme Information</h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">Name:</span>
              <span class="info-value">{{ scheme.displayName || scheme.name }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">ID:</span>
              <span class="info-value">{{ scheme.id }}</span>
            </div>
            <div class="info-item" v-if="scheme.author">
              <span class="info-label">Author:</span>
              <span class="info-value">{{ scheme.author }}</span>
            </div>
            <div class="info-item" v-if="scheme.version">
              <span class="info-label">Version:</span>
              <span class="info-value">{{ scheme.version }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Type:</span>
              <span class="info-value">{{ scheme.isBuiltIn ? 'Built-in' : 'Custom' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Theme:</span>
              <span class="info-value">{{ scheme.isDark ? 'Dark' : 'Light' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Font:</span>
              <span class="info-value">{{ scheme.font.family }} {{ scheme.font.size }}px</span>
            </div>
            <div class="info-item">
              <span class="info-label">Cursor:</span>
              <span class="info-value">{{ scheme.cursor.style }} {{ scheme.cursor.blink ? '(blinking)' : '(static)' }}</span>
            </div>
            <div class="info-item" v-if="scheme.description">
              <span class="info-label">Description:</span>
              <span class="info-value">{{ scheme.description }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Created:</span>
              <span class="info-value">{{ formatDate(scheme.createdAt) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Updated:</span>
              <span class="info-value">{{ formatDate(scheme.updatedAt) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Dialog Footer -->
      <div class="dialog-footer">
        <button class="button button--secondary" @click="close">
          Close
        </button>
        <button v-if="!scheme.isBuiltIn" class="button button--primary" @click="setAsDefault">
          <Icon name="carbon:star" />
          Set as Default
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useColorScheme } from '~/composables/useColorScheme'
import type { ColorScheme } from '~/stores/colorscheme'

// Props
interface Props {
  scheme: ColorScheme
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  close: []
  'set-default': [scheme: ColorScheme]
}>()

// Composables
const { setDefaultScheme } = useColorScheme()

// Methods
const close = () => {
  emit('close')
}

const setAsDefault = async () => {
  try {
    await setDefaultScheme(props.scheme.id)
    emit('close')
  } catch (error) {
    console.error('Failed to set default scheme:', error)
  }
}

const getTerminalStyles = () => {
  return {
    backgroundColor: props.scheme.colors.background,
    color: props.scheme.colors.foreground,
    fontFamily: props.scheme.font.family,
    fontSize: '13px',
    lineHeight: '1.4',
    cursor: props.scheme.cursor.style === 'block' ? 'block' : 'text'
  }
}

const formatDate = (dateString: string) => {
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
  } catch {
    return 'Invalid date'
  }
}
</script>

<style scoped>
.dialog-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4;
}

.dialog-container {
  @apply bg-surface rounded-xl shadow-2xl max-w-6xl w-full max-h-[90vh] overflow-hidden flex flex-col;
}

.dialog-header {
  @apply flex items-center justify-between p-6 border-b border-border;
}

.dialog-title {
  @apply text-xl font-semibold text-text m-0;
}

.close-button {
  @apply p-2 rounded-lg hover:bg-surface-hover text-text-secondary hover:text-text transition-colors;
  border: none;
  background: none;
  cursor: pointer;
}

.dialog-content {
  @apply flex-1 overflow-auto p-6 space-y-8;
}

.section-title {
  @apply text-lg font-medium text-text mb-4;
}

.preview-terminal {
  @apply font-mono rounded-lg shadow-inner overflow-hidden;
  max-height: 400px;
}

.terminal-header {
  @apply flex items-center justify-between px-4 py-2 border-b border-border;
  background: rgba(0, 0, 0, 0.1);
}

.terminal-controls {
  @apply flex items-center gap-2;
}

.control-dot {
  @apply w-3 h-3 rounded-full;
}

.terminal-title {
  @apply text-sm font-medium;
}

.terminal-content {
  @apply p-4 space-y-1;
}

.terminal-line {
  @apply leading-tight;
  white-space: pre;
}

.palette-section {
  @apply space-y-6;
}

.color-grid {
  @apply grid grid-cols-1 lg:grid-cols-3 gap-6;
}

.color-group {
  @apply space-y-3;
}

.color-group-title {
  @apply text-sm font-medium text-text-secondary mb-3;
}

.color-list {
  @apply space-y-2;
}

.color-item {
  @apply flex items-center gap-3 p-2 rounded-lg hover:bg-surface-hover;
}

.color-swatch {
  @apply w-6 h-6 rounded border border-border;
  flex-shrink: 0;
}

.color-name {
  @apply text-sm font-medium text-text min-w-[80px];
}

.color-value {
  @apply text-xs text-text-secondary font-mono;
  flex: 1;
  text-align: right;
}

.info-section {
  @apply space-y-4;
}

.info-grid {
  @apply grid grid-cols-1 md:grid-cols-2 gap-4;
}

.info-item {
  @apply flex items-start gap-2;
}

.info-label {
  @apply text-sm font-medium text-text-secondary min-w-[80px];
}

.info-value {
  @apply text-sm text-text flex-1;
}

.dialog-footer {
  @apply flex items-center justify-end gap-3 p-6 border-t border-border;
}

.button {
  @apply px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2;
  border: none;
  cursor: pointer;
}

.button--secondary {
  @apply bg-surface-hover text-text hover:bg-surface-alt;
}

.button--primary {
  @apply bg-primary text-white hover:bg-primary-hover;
}
</style>