<template>
  <div class="theme-manager">
    <div class="theme-header">
      <h3 class="glitch">THEME MANAGER</h3>
      <div class="theme-controls">
        <button @click="toggleThemePanel" class="control-btn">
          <span class="btn-icon">🎨</span>
        </button>
      </div>
    </div>

    <!-- Theme Selection Panel -->
    <div v-if="showThemePanel" class="theme-panel">
      <div class="theme-categories">
        <div class="category-tabs">
          <button
            v-for="category in themeCategories"
            :key="category.id"
            @click="selectedCategory = category.id"
            class="category-tab"
            :class="{ active: selectedCategory === category.id }"
          >
            <span class="category-icon">{{ category.icon }}</span>
            <span class="category-name">{{ category.name }}</span>
          </button>
        </div>
      </div>

      <div class="theme-grid">
        <div
          v-for="theme in filteredThemes"
          :key="theme.id"
          class="theme-card"
          :class="{ active: currentTheme?.id === theme.id }"
          @click="selectTheme(theme)"
        >
          <div class="theme-preview">
            <div class="preview-header" :style="getPreviewStyle(theme, 'header')">
              <div class="preview-title">{{ theme.name }}</div>
            </div>
            <div class="preview-content" :style="getPreviewStyle(theme, 'content')">
              <div class="preview-terminal" :style="getPreviewStyle(theme, 'terminal')">
                <div class="preview-line">
                  <span class="prompt">$</span>
                  <span class="command">echo "Hello World"</span>
                </div>
                <div class="preview-output">Hello World</div>
              </div>
              <div class="preview-sidebar" :style="getPreviewStyle(theme, 'sidebar')">
                <div class="preview-item">System</div>
                <div class="preview-item">Network</div>
                <div class="preview-item">Files</div>
              </div>
            </div>
          </div>
          <div class="theme-info">
            <div class="theme-name">{{ theme.name }}</div>
            <div class="theme-author">by {{ theme.author }}</div>
            <div class="theme-tags">
              <span
                v-for="tag in theme.tags"
                :key="tag"
                class="theme-tag"
              >
                {{ tag }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Theme Customization -->
      <div v-if="showCustomization" class="theme-customization">
        <h4 class="customization-title">CUSTOMIZE THEME</h4>
        <div class="customization-grid">
          <div class="customization-group">
            <label class="custom-label">Primary Color</label>
            <input
              v-model="customColors.primary"
              type="color"
              @change="updateCustomTheme"
              class="color-input"
            />
          </div>
          <div class="customization-group">
            <label class="custom-label">Accent Color</label>
            <input
              v-model="customColors.accent"
              type="color"
              @change="updateCustomTheme"
              class="color-input"
            />
          </div>
          <div class="customization-group">
            <label class="custom-label">Background</label>
            <input
              v-model="customColors.background"
              type="color"
              @change="updateCustomTheme"
              class="color-input"
            />
          </div>
          <div class="customization-group">
            <label class="custom-label">Terminal Color</label>
            <input
              v-model="customColors.terminal"
              type="color"
              @change="updateCustomTheme"
              class="color-input"
            />
          </div>
        </div>
        <div class="customization-actions">
          <button @click="saveCustomTheme" class="action-btn primary">
            <span class="btn-icon">💾</span>
            Save Theme
          </button>
          <button @click="resetCustomTheme" class="action-btn secondary">
            <span class="btn-icon">↺</span>
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Quick Theme Switcher -->
    <div class="quick-switcher">
      <div class="switcher-header">
        <span class="switcher-title">QUICK SWITCH</span>
      </div>
      <div class="theme-dots">
        <div
          v-for="theme in popularThemes"
          :key="theme.id"
          class="theme-dot"
          :class="{ active: currentTheme?.id === theme.id }"
          :style="{ backgroundColor: theme.colors.primary }"
          @click="selectTheme(theme)"
          :title="theme.name"
        ></div>
      </div>
    </div>

    <!-- Theme Effects -->
    <div class="theme-effects">
      <div class="effects-header">
        <span class="effects-title">EFFECTS</span>
      </div>
      <div class="effects-controls">
        <label class="effect-toggle">
          <input
            v-model="themeSettings.glowEffect"
            type="checkbox"
            @change="updateThemeSettings"
          />
          <span class="effect-label">Glow Effects</span>
        </label>
        <label class="effect-toggle">
          <input
            v-model="themeSettings.animations"
            type="checkbox"
            @change="updateThemeSettings"
          />
          <span class="effect-label">Animations</span>
        </label>
        <label class="effect-toggle">
          <input
            v-model="themeSettings.particles"
            type="checkbox"
            @change="updateThemeSettings"
          />
          <span class="effect-label">Particles</span>
        </label>
        <label class="effect-toggle">
          <input
            v-model="themeSettings.scanlines"
            type="checkbox"
            @change="updateThemeSettings"
          />
          <span class="effect-label">Scanlines</span>
        </label>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'

interface ThemeColors {
  primary: string
  accent: string
  background: string
  surface: string
  text: string
  terminal: string
  terminalText: string
}

interface Theme {
  id: string
  name: string
  author: string
  description: string
  colors: ThemeColors
  tags: string[]
  category: string
  isCustom?: boolean
}

interface ThemeSettings {
  glowEffect: boolean
  animations: boolean
  particles: boolean
  scanlines: boolean
}

// Reactive data
const showThemePanel = ref<boolean>(false)
const selectedCategory = ref<string>('tron')
const currentTheme = ref<Theme | null>(null)
const showCustomization = ref<boolean>(false)
const customColors = ref<ThemeColors>({
  primary: '#0ea5e9',
  accent: '#f59e0b',
  background: '#000000',
  surface: '#0a0a0a',
  text: '#ffffff',
  terminal: '#000000',
  terminalText: '#00ff41'
})

const themeSettings = ref<ThemeSettings>({
  glowEffect: true,
  animations: true,
  particles: false,
  scanlines: false
})

// Theme categories
const themeCategories = [
  { id: 'tron', name: 'TRON', icon: '🟦' },
  { id: 'cyberpunk', name: 'CYBERPUNK', icon: '🌆' },
  { id: 'matrix', name: 'MATRIX', icon: '💚' },
  { id: 'minimal', name: 'MINIMAL', icon: '◻️' },
  { id: 'custom', name: 'CUSTOM', icon: '🎨' }
]

// Available themes
const availableThemes: Theme[] = [
  // TRON Themes
  {
    id: 'tron-classic',
    name: 'TRON Classic',
    author: 'Gabriel Saillard',
    description: 'Original TRON Legacy inspired theme',
    colors: {
      primary: '#0ea5e9',
      accent: '#f59e0b',
      background: '#000000',
      surface: '#0a0a0a',
      text: '#ffffff',
      terminal: 'rgba(0, 0, 0, 0.8)',
      terminalText: '#00ff41'
    },
    tags: ['classic', 'original', 'blue'],
    category: 'tron'
  },
  {
    id: 'tron-orange',
    name: 'TRON Orange',
    author: 'Theme Designer',
    description: 'Orange variant of the TRON theme',
    colors: {
      primary: '#f97316',
      accent: '#0ea5e9',
      background: '#000000',
      surface: '#1a0a00',
      text: '#ffffff',
      terminal: 'rgba(0, 0, 0, 0.8)',
      terminalText: '#fb923c'
    },
    tags: ['orange', 'variant'],
    category: 'tron'
  },
  {
    id: 'tron-disrupted',
    name: 'TRON Disrupted',
    author: 'Experimental',
    description: 'Glitch effect variant',
    colors: {
      primary: '#dc2626',
      accent: '#7c3aed',
      background: '#000000',
      surface: '#1a0000',
      text: '#ffffff',
      terminal: 'rgba(0, 0, 0, 0.9)',
      terminalText: '#ef4444'
    },
    tags: ['experimental', 'glitch', 'red'],
    category: 'tron'
  },
  {
    id: 'tron-horizon',
    name: 'Horizon Full',
    author: 'Community',
    description: 'Beautiful gradient theme',
    colors: {
      primary: '#6366f1',
      accent: '#ec4899',
      background: 'linear-gradient(135deg, #1e1b4b 0%, #831843 100%)',
      surface: 'rgba(99, 102, 241, 0.1)',
      text: '#ffffff',
      terminal: 'rgba(30, 27, 75, 0.8)',
      terminalText: '#a78bfa'
    },
    tags: ['gradient', 'purple', 'beautiful'],
    category: 'tron'
  },

  // Cyberpunk Themes
  {
    id: 'cyberpunk-neon',
    name: 'Neon Dreams',
    author: 'Designer',
    description: 'Cyberpunk neon aesthetic',
    colors: {
      primary: '#ff00ff',
      accent: '#00ffff',
      background: '#0a0014',
      surface: '#1a0033',
      text: '#ffffff',
      terminal: 'rgba(10, 0, 20, 0.8)',
      terminalText: '#ff00ff'
    },
    tags: ['neon', 'cyberpunk', 'purple'],
    category: 'cyberpunk'
  },
  {
    id: 'cyberpunk-synthwave',
    name: 'Synthwave',
    author: 'Retro Designer',
    description: '80s synthwave aesthetic',
    colors: {
      primary: '#ff00ff',
      accent: '#ff6b6b',
      background: '#1a0033',
      surface: '#2d1b69',
      text: '#ffffff',
      terminal: 'rgba(26, 0, 51, 0.8)',
      terminalText: '#ff6b6b'
    },
    tags: ['retro', '80s', 'synthwave'],
    category: 'cyberpunk'
  },

  // Matrix Themes
  {
    id: 'matrix-green',
    name: 'Matrix Rain',
    author: 'Inspired',
    description: 'Classic Matrix green theme',
    colors: {
      primary: '#00ff41',
      accent: '#00ff41',
      background: '#000000',
      surface: '#0d1f0d',
      text: '#00ff41',
      terminal: '#000000',
      terminalText: '#00ff41'
    },
    tags: ['matrix', 'green', 'classic'],
    category: 'matrix'
  },

  // Minimal Themes
  {
    id: 'minimal-dark',
    name: 'Minimal Dark',
    author: 'Clean Designer',
    description: 'Clean minimal dark theme',
    colors: {
      primary: '#6b7280',
      accent: '#3b82f6',
      background: '#000000',
      surface: '#111827',
      text: '#f3f4f6',
      terminal: '#000000',
      terminalText: '#f3f4f6'
    },
    tags: ['minimal', 'clean', 'gray'],
    category: 'minimal'
  },
  {
    id: 'minimal-light',
    name: 'Minimal Light',
    author: 'Light Designer',
    description: 'Clean minimal light theme',
    colors: {
      primary: '#3b82f6',
      accent: '#ef4444',
      background: '#ffffff',
      surface: '#f3f4f6',
      text: '#111827',
      terminal: '#ffffff',
      terminalText: '#111827'
    },
    tags: ['minimal', 'light', 'clean'],
    category: 'minimal'
  }
]

// Computed properties
const filteredThemes = computed(() => {
  if (selectedCategory.value === 'custom') {
    return availableThemes.filter(theme => theme.isCustom)
  }
  return availableThemes.filter(theme => theme.category === selectedCategory.value)
})

const popularThemes = computed(() => {
  return availableThemes.filter(theme =>
    ['tron-classic', 'tron-orange', 'matrix-green', 'cyberpunk-neon'].includes(theme.id)
  )
})

// Methods
const toggleThemePanel = () => {
  showThemePanel.value = !showThemePanel.value
}

const selectTheme = (theme: Theme) => {
  currentTheme.value = theme
  applyTheme(theme)
}

const applyTheme = (theme: Theme) => {
  const root = document.documentElement

  // Apply CSS custom properties
  Object.entries(theme.colors).forEach(([key, value]) => {
    const cssVar = `--theme-${key}`
    root.style.setProperty(cssVar, value)
  })

  // Apply theme class
  document.body.className = `theme-${theme.id}`

  // Apply effects
  applyThemeEffects()

  // Save to localStorage
  localStorage.setItem('selectedTheme', theme.id)
  localStorage.setItem('themeSettings', JSON.stringify(themeSettings.value))
}

const applyThemeEffects = () => {
  const root = document.documentElement

  root.style.setProperty('--glow-effect', themeSettings.value.glowEffect ? '1' : '0')
  root.style.setProperty('--animations-enabled', themeSettings.value.animations ? '1' : '0')
  root.style.setProperty('--particles-enabled', themeSettings.value.particles ? '1' : '0')
  root.style.setProperty('--scanlines-enabled', themeSettings.value.scanlines ? '1' : '0')

  // Toggle effect classes
  document.body.classList.toggle('glow-enabled', themeSettings.value.glowEffect)
  document.body.classList.toggle('animations-enabled', themeSettings.value.animations)
  document.body.classList.toggle('particles-enabled', themeSettings.value.particles)
  document.body.classList.toggle('scanlines-enabled', themeSettings.value.scanlines)
}

const getPreviewStyle = (theme: Theme, element: string) => {
  const styles: Record<string, string> = {}

  switch (element) {
    case 'header':
      styles.background = theme.colors.surface
      styles.borderBottom = `1px solid ${theme.colors.primary}40`
      break
    case 'content':
      styles.background = theme.colors.background
      break
    case 'terminal':
      styles.background = theme.colors.terminal
      styles.border = `1px solid ${theme.colors.primary}40`
      styles.color = theme.colors.terminalText
      break
    case 'sidebar':
      styles.background = theme.colors.surface + '40'
      styles.border = `1px solid ${theme.colors.primary}20`
      break
  }

  return styles
}

const updateCustomTheme = () => {
  if (showCustomization.value) {
    const customTheme: Theme = {
      id: 'custom-temp',
      name: 'Custom Theme',
      author: 'You',
      description: 'Your custom theme',
      colors: { ...customColors.value },
      tags: ['custom'],
      category: 'custom',
      isCustom: true
    }
    applyTheme(customTheme)
  }
}

const saveCustomTheme = () => {
  const customTheme: Theme = {
    id: `custom-${Date.now()}`,
    name: `Custom Theme ${availableThemes.filter(t => t.isCustom).length + 1}`,
    author: 'You',
    description: 'Custom user theme',
    colors: { ...customColors.value },
    tags: ['custom'],
    category: 'custom',
    isCustom: true
  }

  availableThemes.push(customTheme)
  selectTheme(customTheme)
  showCustomization.value = false

  // Save to localStorage
  localStorage.setItem('customThemes', JSON.stringify(availableThemes.filter(t => t.isCustom)))
}

const resetCustomTheme = () => {
  customColors.value = {
    primary: '#0ea5e9',
    accent: '#f59e0b',
    background: '#000000',
    surface: '#0a0a0a',
    text: '#ffffff',
    terminal: 'rgba(0, 0, 0, 0.8)',
    terminalText: '#00ff41'
  }
  updateCustomTheme()
}

const updateThemeSettings = () => {
  applyThemeEffects()
  localStorage.setItem('themeSettings', JSON.stringify(themeSettings.value))
}

const loadSavedTheme = () => {
  const savedThemeId = localStorage.getItem('selectedTheme')
  const savedSettings = localStorage.getItem('themeSettings')
  const customThemes = localStorage.getItem('customThemes')

  // Load custom themes
  if (customThemes) {
    try {
      const custom = JSON.parse(customThemes)
      availableThemes.push(...custom)
    } catch (error) {
      console.error('Failed to load custom themes:', error)
    }
  }

  // Load settings
  if (savedSettings) {
    try {
      themeSettings.value = { ...themeSettings.value, ...JSON.parse(savedSettings) }
    } catch (error) {
      console.error('Failed to load theme settings:', error)
    }
  }

  // Load theme
  if (savedThemeId) {
    const theme = availableThemes.find(t => t.id === savedThemeId)
    if (theme) {
      selectTheme(theme)
      return
    }
  }

  // Default theme
  selectTheme(availableThemes[0])
}

// Watch for category changes
watch(selectedCategory, (newCategory) => {
  if (newCategory === 'custom') {
    showCustomization.value = true
  } else {
    showCustomization.value = false
  }
})

// Lifecycle
onMounted(() => {
  loadSavedTheme()
})
</script>

<style scoped>
.theme-manager {
  background: rgba(10, 10, 10, 0.95);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  padding: 20px;
  backdrop-filter: blur(10px);
  font-family: 'Fira Code', monospace;
  color: var(--text-primary);
}

.theme-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid var(--surface-border);
}

.theme-controls {
  display: flex;
  gap: 8px;
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.control-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.btn-icon {
  font-size: 16px;
}

.theme-panel {
  margin-bottom: 20px;
}

.theme-categories {
  margin-bottom: 20px;
}

.category-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 10px;
}

.category-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 20px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  font-size: 11px;
}

.category-tab:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.category-tab.active {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

.category-icon {
  font-size: 14px;
}

.category-name {
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.theme-card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
}

.theme-card:hover {
  border-color: var(--primary-500);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(14, 165, 233, 0.15);
}

.theme-card.active {
  border-color: var(--primary-500);
  box-shadow: 0 0 20px rgba(14, 165, 233, 0.3);
}

.theme-preview {
  height: 120px;
  overflow: hidden;
  position: relative;
}

.preview-header {
  height: 30px;
  display: flex;
  align-items: center;
  padding: 0 10px;
  font-size: 10px;
  font-weight: bold;
}

.preview-content {
  height: 90px;
  display: flex;
  padding: 8px;
  gap: 8px;
}

.preview-terminal {
  flex: 1;
  padding: 6px;
  border-radius: 4px;
  font-size: 9px;
  font-family: monospace;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.preview-line {
  display: flex;
  gap: 4px;
}

.prompt {
  color: var(--accent-500);
}

.command {
  color: inherit;
}

.preview-output {
  color: var(--text-secondary);
}

.preview-sidebar {
  width: 40px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 8px;
  padding: 4px;
  border-radius: 4px;
}

.preview-item {
  color: var(--text-secondary);
  text-align: center;
}

.theme-info {
  padding: 12px;
}

.theme-name {
  font-weight: bold;
  color: var(--text-primary);
  margin-bottom: 4px;
  font-size: 12px;
}

.theme-author {
  color: var(--text-secondary);
  font-size: 10px;
  margin-bottom: 8px;
}

.theme-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.theme-tag {
  padding: 2px 6px;
  background: var(--surface-elevated);
  border-radius: 10px;
  font-size: 8px;
  color: var(--text-secondary);
}

.theme-customization {
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 20px;
}

.customization-title {
  color: var(--primary-400);
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 15px;
}

.customization-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 15px;
  margin-bottom: 15px;
}

.customization-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.custom-label {
  font-size: 10px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.color-input {
  width: 100%;
  height: 40px;
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  cursor: pointer;
  background: var(--surface);
}

.customization-actions {
  display: flex;
  gap: 10px;
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

.quick-switcher {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 15px;
}

.switcher-header {
  margin-bottom: 10px;
}

.switcher-title {
  color: var(--primary-400);
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.theme-dots {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.theme-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 2px solid transparent;
}

.theme-dot:hover {
  transform: scale(1.1);
  border-color: var(--text-primary);
}

.theme-dot.active {
  border-color: var(--text-primary);
  box-shadow: 0 0 15px currentColor;
}

.theme-effects {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
}

.effects-header {
  margin-bottom: 10px;
}

.effects-title {
  color: var(--primary-400);
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.effects-controls {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 10px;
}

.effect-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 10px;
  color: var(--text-primary);
}

.effect-toggle input[type="checkbox"] {
  accent-color: var(--primary-500);
}

.effect-label {
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Glitch effect for title */
.glitch {
  position: relative;
  color: var(--primary-400);
  font-size: 14px;
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
  .theme-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 10px;
  }

  .customization-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .effects-controls {
    grid-template-columns: repeat(2, 1fr);
  }

  .category-tabs {
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .theme-grid {
    grid-template-columns: 1fr;
  }

  .customization-grid,
  .effects-controls {
    grid-template-columns: 1fr;
  }

  .customization-actions {
    flex-direction: column;
  }
}
</style>