<template>
  <div class="theme-selector">
    <!-- Header -->
    <div class="selector-header">
      <h3 class="glitch-text">THEME SELECTOR</h3>
      <div class="header-controls">
        <button
          @click="togglePanel"
          class="control-btn"
          :class="{ active: isOpen }"
        >
          <Icon :name="isOpen ? 'x' : 'palette'" />
        </button>
      </div>
    </div>

    <!-- Main Panel -->
    <Transition name="slide-down" appear>
      <div v-if="isOpen" class="selector-panel">
        <!-- Category Tabs -->
        <div class="category-tabs">
          <button
            v-for="category in categories"
            :key="category.id"
            @click="selectedCategory = category.id"
            class="category-tab"
            :class="{ active: selectedCategory === category.id }"
          >
            <span class="category-icon">{{ category.icon }}</span>
            <span class="category-name">{{ category.name }}</span>
            <span class="category-count">{{ getCategoryCount(category.id) }}</span>
          </button>
        </div>

        <!-- Search Bar -->
        <div class="search-section">
          <div class="search-bar">
            <Icon name="search" class="search-icon" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search themes..."
              class="search-input"
            />
            <button
              v-if="searchQuery"
              @click="clearSearch"
              class="clear-search"
            >
              <Icon name="x" />
            </button>
          </div>
        </div>

        <!-- Theme Grid -->
        <div class="themes-section">
          <div v-if="isLoading" class="loading-state">
            <Icon name="loader-2" class="animate-spin" />
            <span>Loading themes...</span>
          </div>

          <div v-else-if="filteredThemes.length === 0" class="empty-state">
            <Icon name="palette" />
            <span>No themes found</span>
            <button @click="createNewTheme" class="create-theme-btn">
              <Icon name="plus" />
              Create Theme
            </button>
          </div>

          <div v-else class="theme-grid">
            <div
              v-for="theme in filteredThemes"
              :key="theme.id"
              class="theme-card"
              :class="{
                active: currentTheme?.id === theme.id,
                custom: theme.isCustom
              }"
              @click="selectTheme(theme)"
            >
              <!-- Theme Preview -->
              <div class="theme-preview">
                <ThemePreview
                  :theme="theme"
                  :compact="true"
                  :show-labels="false"
                />
              </div>

              <!-- Theme Info -->
              <div class="theme-info">
                <div class="theme-header-info">
                  <h4 class="theme-name">{{ theme.displayName || theme.name }}</h4>
                  <div class="theme-badges">
                    <span v-if="theme.isCustom" class="badge custom">CUSTOM</span>
                    <span v-if="isCurrentTheme(theme)" class="badge active">ACTIVE</span>
                  </div>
                </div>

                <p class="theme-description">{{ theme.description }}</p>

                <div class="theme-meta">
                  <span class="theme-author">by {{ theme.author }}</span>
                  <span class="theme-version">v{{ theme.version }}</span>
                </div>

                <!-- Theme Tags -->
                <div v-if="theme.metadata?.tags?.length" class="theme-tags">
                  <span
                    v-for="tag in theme.metadata.tags.slice(0, 3)"
                    :key="tag"
                    class="theme-tag"
                  >
                    {{ tag }}
                  </span>
                  <span
                    v-if="theme.metadata.tags.length > 3"
                    class="theme-tag more"
                  >
                    +{{ theme.metadata.tags.length - 3 }}
                  </span>
                </div>
              </div>

              <!-- Theme Actions -->
              <div class="theme-actions">
                <button
                  @click.stop="previewTheme(theme)"
                  class="action-btn preview"
                  title="Preview theme"
                >
                  <Icon name="eye" />
                </button>
                <button
                  v-if="theme.isCustom"
                  @click.stop="editTheme(theme)"
                  class="action-btn edit"
                  title="Edit theme"
                >
                  <Icon name="edit" />
                </button>
                <button
                  v-if="theme.isCustom && !isCurrentTheme(theme)"
                  @click.stop="deleteTheme(theme)"
                  class="action-btn delete"
                  title="Delete theme"
                >
                  <Icon name="trash-2" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="quick-actions">
          <button @click="createNewTheme" class="action-button primary">
            <Icon name="plus" />
            Create New Theme
          </button>
          <button @click="importTheme" class="action-button secondary">
            <Icon name="upload" />
            Import Theme
          </button>
          <button @click="exportThemes" class="action-button secondary">
            <Icon name="download" />
            Export Themes
          </button>
        </div>
      </div>
    </Transition>

    <!-- Error Display -->
    <div v-if="error" class="error-message">
      <Icon name="alert-circle" />
      <span>{{ error }}</span>
      <button @click="clearError" class="error-close">
        <Icon name="x" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useThemeStore } from '~/stores/theme'
import { useThemePreview } from '~/composables/useTheme'
import type { Theme } from '~/composables/useTheme'

// Props
interface Props {
  initialOpen?: boolean
  showCreateButton?: boolean
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  initialOpen: false,
  showCreateButton: true,
  compact: false
})

// Emits
const emit = defineEmits<{
  themeSelected: [theme: Theme]
  themeCreated: [theme: Theme]
  themeDeleted: [themeId: string]
  togglePanel: [isOpen: boolean]
}>()

// Store and composables
const themeStore = useThemeStore()
const themePreview = useThemePreview()

// Reactive state
const isOpen = ref(props.initialOpen)
const selectedCategory = ref('all')
const searchQuery = ref('')

// Categories
const categories = [
  { id: 'all', name: 'ALL', icon: '🌈' },
  { id: 'built-in', name: 'BUILT-IN', icon: '⚙️' },
  { id: 'custom', name: 'CUSTOM', icon: '🎨' },
  { id: 'dark', name: 'DARK', icon: '🌙' },
  { id: 'light', name: 'LIGHT', icon: '☀️' }
]

// Computed properties
const currentTheme = computed(() => themeStore.currentTheme)
const availableThemes = computed(() => themeStore.availableThemes)
const isLoading = computed(() => themeStore.isLoading)
const error = computed(() => themeStore.error)

const filteredThemes = computed(() => {
  let themes = availableThemes.value

  // Filter by category
  if (selectedCategory.value !== 'all') {
    if (selectedCategory.value === 'built-in') {
      themes = themes.filter(t => !t.isCustom)
    } else if (selectedCategory.value === 'custom') {
      themes = themes.filter(t => t.isCustom)
    } else if (selectedCategory.value === 'dark') {
      themes = themes.filter(t => t.isDark !== false)
    } else if (selectedCategory.value === 'light') {
      themes = themes.filter(t => t.isDark === false)
    }
  }

  // Filter by search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    themes = themes.filter(theme =>
      theme.name.toLowerCase().includes(query) ||
      theme.displayName?.toLowerCase().includes(query) ||
      theme.description.toLowerCase().includes(query) ||
      theme.author.toLowerCase().includes(query) ||
      theme.metadata?.tags?.some(tag => tag.toLowerCase().includes(query))
    )
  }

  return themes
})

// Methods
const togglePanel = () => {
  isOpen.value = !isOpen.value
  emit('togglePanel', isOpen.value)
}

const selectTheme = async (theme: Theme) => {
  try {
    await themeStore.setCurrentTheme(theme.id)
    emit('themeSelected', theme)
  } catch (error) {
    console.error('Failed to select theme:', error)
  }
}

const previewTheme = (theme: Theme) => {
  themePreview.previewTheme(theme)
}

const editTheme = (theme: Theme) => {
  // Navigate to theme editor or open edit modal
  console.log('Edit theme:', theme.id)
  // Implementation would depend on your routing/modal system
}

const deleteTheme = async (theme: Theme) => {
  if (!confirm(`Are you sure you want to delete the theme "${theme.name}"?`)) {
    return
  }

  try {
    await themeStore.deleteTheme(theme.id)
    emit('themeDeleted', theme.id)
  } catch (error) {
    console.error('Failed to delete theme:', error)
  }
}

const createNewTheme = () => {
  // Navigate to theme creator or open creation modal
  console.log('Create new theme')
  // Implementation would depend on your routing/modal system
}

const importTheme = () => {
  // Open file picker or import dialog
  console.log('Import theme')
  // Implementation would depend on your file handling system
}

const exportThemes = () => {
  try {
    const exportData = themeStore.exportThemes()
    const blob = new Blob([exportData], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `themes-${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Failed to export themes:', error)
  }
}

const clearSearch = () => {
  searchQuery.value = ''
}

const clearError = () => {
  themeStore.clearError()
}

const getCategoryCount = (categoryId: string) => {
  switch (categoryId) {
    case 'all':
      return availableThemes.value.length
    case 'built-in':
      return availableThemes.value.filter(t => !t.isCustom).length
    case 'custom':
      return availableThemes.value.filter(t => t.isCustom).length
    case 'dark':
      return availableThemes.value.filter(t => t.isDark !== false).length
    case 'light':
      return availableThemes.value.filter(t => t.isDark === false).length
    default:
      return 0
  }
}

const isCurrentTheme = (theme: Theme) => {
  return currentTheme.value?.id === theme.id
}

// Initialize store on mount
onMounted(async () => {
  try {
    if (!themeStore.isReady) {
      await themeStore.initialize()
    }
  } catch (error) {
    console.error('Failed to initialize theme store:', error)
  }
})

// Cleanup on unmount
onUnmounted(() => {
  themePreview.clearPreview()
})

// Watch for external changes
watch(() => props.initialOpen, (newValue) => {
  isOpen.value = newValue
})
</script>

<style scoped>
.theme-selector {
  background: rgba(10, 10, 10, 0.95);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  color: var(--text-primary);
  overflow: hidden;
}

.selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: var(--surface);
  border-bottom: 1px solid var(--surface-border);
}

.glitch-text {
  color: var(--accent-primary);
  font-size: 14px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 2px;
  animation: glitch 2s infinite;
}

@keyframes glitch {
  0%, 90%, 100% {
    text-shadow: 2px 2px 0 var(--accent-secondary), -2px -2px 0 var(--accent-primary);
  }
  95% {
    text-shadow: -2px 2px 0 var(--accent-secondary), 2px -2px 0 var(--accent-primary);
  }
}

.header-controls {
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
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.control-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--background-primary);
}

/* Slide transition */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.slide-down-enter-from {
  max-height: 0;
  opacity: 0;
}

.slide-down-leave-to {
  max-height: 0;
  opacity: 0;
}

.slide-down-enter-to,
.slide-down-leave-from {
  max-height: 1000px;
  opacity: 1;
}

.selector-panel {
  padding: 20px;
}

/* Category Tabs */
.category-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  overflow-x: auto;
  padding-bottom: 4px;
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
  font-weight: 500;
}

.category-tab:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.category-tab.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--background-primary);
}

.category-icon {
  font-size: 14px;
}

.category-name {
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.category-count {
  background: var(--surface-elevated);
  border-radius: 10px;
  padding: 2px 6px;
  font-size: 9px;
  min-width: 16px;
  text-align: center;
}

/* Search Section */
.search-section {
  margin-bottom: 20px;
}

.search-bar {
  display: flex;
  align-items: center;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 0 12px;
}

.search-icon {
  color: var(--text-secondary);
  margin-right: 8px;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 13px;
  padding: 10px 0;
  outline: none;
}

.search-input::placeholder {
  color: var(--text-secondary);
}

.clear-search {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: color 0.2s ease;
}

.clear-search:hover {
  color: var(--text-primary);
}

/* Loading and Empty States */
.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
  color: var(--text-secondary);
}

.loading-state .animate-spin {
  animation: spin 1s linear infinite;
  margin-bottom: 12px;
  font-size: 24px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.create-theme-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 16px;
  padding: 8px 16px;
  background: var(--accent-primary);
  border: none;
  border-radius: 6px;
  color: var(--background-primary);
  cursor: pointer;
  font-family: inherit;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.create-theme-btn:hover {
  background: var(--accent-secondary);
}

/* Theme Grid */
.theme-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.theme-card {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
  position: relative;
}

.theme-card:hover {
  border-color: var(--accent-primary);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 255, 65, 0.15);
}

.theme-card.active {
  border-color: var(--accent-primary);
  box-shadow: 0 0 20px rgba(0, 255, 65, 0.3);
}

.theme-card.custom {
  background: linear-gradient(135deg, var(--surface) 0%, var(--surface-elevated) 100%);
}

.theme-preview {
  height: 120px;
  overflow: hidden;
}

.theme-info {
  padding: 16px;
}

.theme-header-info {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.theme-name {
  font-size: 14px;
  font-weight: bold;
  color: var(--text-primary);
  margin: 0;
}

.theme-badges {
  display: flex;
  gap: 4px;
}

.badge {
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 8px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.badge.custom {
  background: var(--accent-primary);
  color: var(--background-primary);
}

.badge.active {
  background: var(--success-color);
  color: var(--background-primary);
}

.theme-description {
  color: var(--text-secondary);
  font-size: 11px;
  line-height: 1.4;
  margin: 0 0 8px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.theme-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 10px;
  color: var(--text-secondary);
}

.theme-author {
  font-weight: 500;
}

.theme-version {
  opacity: 0.7;
}

.theme-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.theme-tag {
  padding: 2px 6px;
  background: var(--surface-elevated);
  border-radius: 8px;
  font-size: 9px;
  color: var(--text-secondary);
}

.theme-tag.more {
  background: var(--accent-primary);
  color: var(--background-primary);
}

.theme-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  padding: 8px 16px 16px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.action-btn.delete:hover {
  border-color: var(--error-color);
  color: var(--error-color);
}

/* Quick Actions */
.quick-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid var(--surface-border);
}

.action-button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  background: var(--surface);
  color: var(--text-primary);
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.action-button:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.action-button.primary {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--background-primary);
}

.action-button.primary:hover {
  background: var(--accent-secondary);
  border-color: var(--accent-secondary);
}

/* Error Message */
.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(255, 51, 51, 0.1);
  border: 1px solid var(--error-color);
  border-radius: 6px;
  color: var(--error-color);
  font-size: 12px;
  margin: 16px 20px;
}

.error-close {
  margin-left: auto;
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 2px;
}

/* Responsive Design */
@media (max-width: 768px) {
  .theme-grid {
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 12px;
  }

  .category-tabs {
    justify-content: center;
  }

  .quick-actions {
    flex-direction: column;
  }

  .action-button {
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .theme-grid {
    grid-template-columns: 1fr;
  }

  .selector-header,
  .selector-panel {
    padding: 12px 16px;
  }

  .category-tabs {
    gap: 4px;
  }

  .category-tab {
    padding: 6px 10px;
    font-size: 10px;
  }
}
</style>