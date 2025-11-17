<template>
  <div class="theme-preview" :class="{ compact, 'show-labels': showLabels }">
    <!-- Preview Header -->
    <div class="preview-header" :style="headerStyles">
      <div v-if="showLabels" class="header-content">
        <div class="header-title">{{ theme.displayName || theme.name }}</div>
        <div class="header-controls">
          <div class="control-dot" :style="{ backgroundColor: theme.colors.accent.primary }"></div>
          <div class="control-dot" :style="{ backgroundColor: theme.colors.status.warning }"></div>
          <div class="control-dot" :style="{ backgroundColor: theme.colors.status.error }"></div>
        </div>
      </div>
    </div>

    <!-- Preview Content -->
    <div class="preview-content" :style="contentStyles">
      <!-- Terminal Section -->
      <div class="terminal-section" :style="terminalStyles">
        <div v-if="showLabels" class="terminal-header">
          <span class="terminal-title">TERMINAL</span>
          <div class="terminal-controls">
            <span class="terminal-button minimize">_</span>
            <span class="terminal-button maximize">□</span>
            <span class="terminal-button close">×</span>
          </div>
        </div>

        <div class="terminal-content">
          <div class="terminal-line">
            <span class="prompt" :style="{ color: theme.colors.accent.primary }">$</span>
            <span class="command">echo "Hello, World!"</span>
          </div>
          <div class="terminal-output" :style="{ color: theme.colors.foreground.secondary }">
            Hello, World!
          </div>
          <div class="terminal-line">
            <span class="prompt" :style="{ color: theme.colors.accent.primary }">$</span>
            <span class="command">ls -la</span>
          </div>
          <div class="terminal-output" :style="{ color: theme.colors.foreground.secondary }">
            <div>drwxr-xr-x 2 user user 4096 Dec 1 10:00 .</div>
            <div>drwxr-xr-x 3 user user 4096 Dec 1 10:01 ..</div>
            <div>-rw-r--r-- 1 user user  512 Dec 1 10:02 file.txt</div>
          </div>
        </div>
      </div>

      <!-- Sidebar Section -->
      <div v-if="!compact" class="sidebar-section" :style="sidebarStyles">
        <div v-if="showLabels" class="sidebar-header">
          <span class="sidebar-title">NAVIGATION</span>
        </div>

        <div class="sidebar-content">
          <div class="sidebar-item" :style="{ color: theme.colors.foreground.secondary }">
            <Icon name="home" />
            <span>Dashboard</span>
          </div>
          <div class="sidebar-item active" :style="{
            color: theme.colors.accent.primary,
            backgroundColor: theme.colors.accent.primary + '20'
          }">
            <Icon name="monitor" />
            <span>System</span>
          </div>
          <div class="sidebar-item" :style="{ color: theme.colors.foreground.secondary }">
            <Icon name="network" />
            <span>Network</span>
          </div>
          <div class="sidebar-item" :style="{ color: theme.colors.foreground.secondary }">
            <Icon name="folder" />
            <span>Files</span>
          </div>
        </div>
      </div>

      <!-- Status Bar -->
      <div class="status-bar" :style="statusBarStyles">
        <div class="status-left">
          <span class="status-item" :style="{ color: theme.colors.status.success }">
            <Icon name="check-circle" />
            READY
          </span>
          <span class="status-item" :style="{ color: theme.colors.foreground.secondary }">
            {{ theme.metadata?.category?.toUpperCase() || 'UNKNOWN' }}
          </span>
        </div>
        <div class="status-right">
          <span class="status-item" :style="{ color: theme.colors.accent.secondary }">
            CPU: 12%
          </span>
          <span class="status-item" :style="{ color: theme.colors.accent.primary }">
            MEM: 4.2GB
          </span>
        </div>
      </div>
    </div>

    <!-- Interactive Elements -->
    <div v-if="!compact && showInteractive" class="interactive-elements">
      <button class="demo-button" :style="buttonStyles" @click="handleClick">
        <Icon name="play" />
        Run Command
      </button>

      <div class="input-group" :style="inputGroupStyles">
        <input
          :value="inputValue"
          @input="inputValue = $event.target.value"
          class="demo-input"
          :style="inputStyles"
          placeholder="Enter command..."
        />
        <button class="demo-submit" :style="buttonStyles" @click="handleSubmit">
          <Icon name="arrow-right" />
        </button>
      </div>
    </div>

    <!-- Theme Info Overlay -->
    <div v-if="showInfo" class="theme-info-overlay">
      <div class="theme-info-content">
        <h4>{{ theme.displayName || theme.name }}</h4>
        <p>{{ theme.description }}</p>
        <div class="theme-meta">
          <span class="meta-item">by {{ theme.author }}</span>
          <span class="meta-item">v{{ theme.version }}</span>
        </div>
        <div v-if="theme.metadata?.tags?.length" class="theme-tags">
          <span
            v-for="tag in theme.metadata.tags"
            :key="tag"
            class="theme-tag"
            :style="{ backgroundColor: theme.colors.accent.primary + '20', color: theme.colors.accent.primary }"
          >
            {{ tag }}
          </span>
        </div>
      </div>
    </div>

    <!-- Hover Effect -->
    <div class="hover-overlay" :style="hoverOverlayStyles"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Theme } from '~/composables/useTheme'

// Props
interface Props {
  theme: Theme
  compact?: boolean
  showLabels?: boolean
  showInteractive?: boolean
  showInfo?: boolean
  size?: 'small' | 'medium' | 'large'
}

const props = withDefaults(defineProps<Props>(), {
  compact: false,
  showLabels: true,
  showInteractive: false,
  showInfo: false,
  size: 'medium'
})

// Reactive state
const inputValue = ref('')

// Computed styles based on theme
const headerStyles = computed(() => ({
  backgroundColor: props.theme.colors.background.secondary,
  borderBottom: `1px solid ${props.theme.colors.ui.border}`,
  color: props.theme.colors.foreground.primary
}))

const contentStyles = computed(() => ({
  backgroundColor: props.theme.colors.background.primary,
  color: props.theme.colors.foreground.primary
}))

const terminalStyles = computed(() => ({
  backgroundColor: props.theme.colors.ui.inputBackground,
  border: `1px solid ${props.theme.colors.ui.inputBorder}`,
  color: props.theme.colors.ui.inputForeground,
  fontFamily: props.theme.fonts.families.monospace,
  fontSize: props.compact ? '10px' : '11px'
}))

const sidebarStyles = computed(() => ({
  backgroundColor: props.theme.colors.background.secondary + '40',
  borderLeft: `1px solid ${props.theme.colors.ui.border}`,
  color: props.theme.colors.foreground.secondary
}))

const statusBarStyles = computed(() => ({
  backgroundColor: props.theme.colors.background.tertiary,
  borderTop: `1px solid ${props.theme.colors.ui.border}`,
  color: props.theme.colors.foreground.secondary,
  fontSize: '10px'
}))

const buttonStyles = computed(() => ({
  backgroundColor: props.theme.colors.ui.buttonBackground,
  color: props.theme.colors.ui.buttonForeground,
  border: `1px solid ${props.theme.colors.ui.inputBorder}`,
  fontFamily: props.theme.fonts.families.primary,
  fontSize: '11px',
  borderRadius: props.theme.effects.borderRadius.small
}))

const inputStyles = computed(() => ({
  backgroundColor: props.theme.colors.ui.inputBackground,
  color: props.theme.colors.ui.inputForeground,
  border: `1px solid ${props.theme.colors.ui.inputBorder}`,
  fontFamily: props.theme.fonts.families.monospace,
  fontSize: '11px'
}))

const inputGroupStyles = computed(() => ({
  border: `1px solid ${props.theme.colors.ui.inputBorder}`,
  borderRadius: props.theme.effects.borderRadius.small
}))

const hoverOverlayStyles = computed(() => ({
  background: `linear-gradient(135deg, ${props.theme.colors.accent.primary}20 0%, ${props.theme.colors.accent.secondary}20 100%)`,
  opacity: 0
}))

// Methods
const handleClick = () => {
  console.log('Button clicked in theme preview')
}

const handleSubmit = () => {
  console.log('Form submitted with value:', inputValue.value)
  inputValue.value = ''
}
</script>

<style scoped>
.theme-preview {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 8px;
  overflow: hidden;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.theme-preview:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

.theme-preview:hover .hover-overlay {
  opacity: 1;
}

.theme-preview.compact {
  border-radius: 4px;
}

.preview-header {
  height: 30px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  font-size: 10px;
  font-weight: bold;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.header-controls {
  display: flex;
  gap: 6px;
}

.control-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  box-shadow: 0 0 4px currentColor;
}

.preview-content {
  height: calc(100% - 30px);
  display: flex;
  flex-direction: column;
}

.terminal-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  margin: 8px;
  border-radius: 4px;
  overflow: hidden;
}

.theme-preview.compact .terminal-section {
  margin: 4px;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 8px;
  font-size: 9px;
  border-bottom: 1px solid;
  opacity: 0.7;
}

.terminal-title {
  font-weight: bold;
}

.terminal-controls {
  display: flex;
  gap: 8px;
}

.terminal-button {
  font-size: 8px;
  opacity: 0.7;
  cursor: pointer;
}

.terminal-content {
  flex: 1;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-preview.compact .terminal-content {
  padding: 4px;
  gap: 2px;
}

.terminal-line {
  display: flex;
  gap: 6px;
  align-items: center;
}

.prompt {
  font-weight: bold;
  min-width: 12px;
}

.command {
  color: inherit;
}

.terminal-output {
  font-size: 0.9em;
  line-height: 1.2;
  padding-left: 18px;
}

.sidebar-section {
  width: 120px;
  border-left: 1px solid;
  margin: 0 8px 8px 0;
}

.sidebar-header {
  padding: 4px 8px;
  font-size: 9px;
  font-weight: bold;
  border-bottom: 1px solid;
  opacity: 0.7;
}

.sidebar-content {
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  border-radius: 3px;
  font-size: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.sidebar-item.active {
  font-weight: bold;
}

.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 12px;
  font-size: 9px;
}

.status-left,
.status-right {
  display: flex;
  gap: 12px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
}

.interactive-elements {
  position: absolute;
  bottom: 40px;
  left: 8px;
  right: 8px;
  display: flex;
  gap: 8px;
  z-index: 10;
}

.demo-button {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.demo-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.input-group {
  display: flex;
  flex: 1;
  border-radius: inherit;
  overflow: hidden;
}

.demo-input {
  flex: 1;
  padding: 6px 8px;
  border: none;
  outline: none;
  background: transparent;
}

.demo-input::placeholder {
  opacity: 0.6;
}

.demo-submit {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.theme-info-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.9);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 20;
}

.theme-preview:hover .theme-info-overlay {
  opacity: 1;
}

.theme-info-content {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 16px;
  max-width: 200px;
  text-align: center;
}

.theme-info-content h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: var(--text-primary);
}

.theme-info-content p {
  margin: 0 0 12px 0;
  font-size: 11px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.theme-meta {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-bottom: 12px;
}

.meta-item {
  font-size: 10px;
  color: var(--text-secondary);
}

.theme-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  justify-content: center;
}

.theme-tag {
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 8px;
  font-weight: 500;
}

.hover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;
  border-radius: inherit;
}

/* Size variations */
.theme-preview.small .terminal-section {
  font-size: 9px;
}

.theme-preview.small .sidebar-section {
  width: 80px;
}

.theme-preview.large .terminal-section {
  font-size: 12px;
}

.theme-preview.large .sidebar-section {
  width: 160px;
}

/* Responsive design */
@media (max-width: 768px) {
  .sidebar-section {
    width: 100px;
  }

  .status-left,
  .status-right {
    gap: 8px;
  }

  .interactive-elements {
    flex-direction: column;
    gap: 6px;
  }
}

@media (max-width: 480px) {
  .sidebar-section {
    display: none;
  }

  .status-bar {
    font-size: 8px;
  }

  .status-left,
  .status-right {
    gap: 6px;
  }
}
</style>