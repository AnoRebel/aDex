<template>
  <div class="color-scheme-selector">
    <!-- Header -->
    <div class="selector-header">
      <h3 class="selector-title">Color Schemes</h3>
      <div class="selector-actions">
        <button
          class="action-button"
          @click="refreshSchemes"
          :disabled="isLoading"
          title="Refresh color schemes"
        >
          <Icon name="carbon:refresh" :class="{ 'animate-spin': isLoading }" />
        </button>
        <button
          class="action-button"
          @click="showImportDialog = true"
          title="Import color scheme"
        >
          <Icon name="carbon:download" />
        </button>
        <button
          class="action-button"
          @click="showCreateDialog = true"
          title="Create new color scheme"
        >
          <Icon name="carbon:add" />
        </button>
      </div>
    </div>

    <!-- Search and Filter -->
    <div class="selector-controls">
      <div class="search-box">
        <Icon name="carbon:search" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search color schemes..."
          class="search-input"
        />
      </div>
      <div class="filter-controls">
        <select v-model="selectedCategory" class="filter-select">
          <option value="all">All Schemes</option>
          <option value="dark">Dark Themes</option>
          <option value="light">Light Themes</option>
          <option value="builtin">Built-in</option>
          <option value="custom">Custom</option>
        </select>
        <label class="checkbox-label">
          <input
            v-model="showOnlyEnabled"
            type="checkbox"
            class="checkbox-input"
          />
          <span>Enabled only</span>
        </label>
      </div>
    </div>

    <!-- Color Scheme Grid -->
    <div class="schemes-container">
      <div v-if="isLoading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading color schemes...</p>
      </div>

      <div v-else-if="error" class="error-state">
        <Icon name="carbon:warning-alt" class="error-icon" />
        <p class="error-message">{{ error }}</p>
        <button @click="refreshSchemes" class="retry-button">
          <Icon name="carbon:reset" />
          Retry
        </button>
      </div>

      <div v-else-if="filteredSchemes.length === 0" class="empty-state">
        <Icon name="carbon:color-palette" class="empty-icon" />
        <p>No color schemes found</p>
        <button @click="showCreateDialog = true" class="create-button">
          <Icon name="carbon:add" />
          Create Color Scheme
        </button>
      </div>

      <div v-else class="schemes-grid">
        <div
          v-for="scheme in filteredSchemes"
          :key="scheme.id"
          class="scheme-card"
          :class="{
            'scheme-card--selected': selectedSchemeId === scheme.id,
            'scheme-card--default': defaultScheme?.id === scheme.id,
            'scheme-card--builtin': scheme.isBuiltIn
          }"
          @click="selectScheme(scheme.id)"
        >
          <!-- Scheme Preview -->
          <div class="scheme-preview">
            <div
              class="preview-terminal"
              :style="getPreviewStyles(scheme)"
            >
              <div class="preview-header">
                <span class="preview-title">{{ scheme.displayName || scheme.name }}</span>
                <div class="preview-controls">
                  <div class="preview-dot" style="background: #ff5f56;"></div>
                  <div class="preview-dot" style="background: #ffbd2e;"></div>
                  <div class="preview-dot" style="background: #27ca3f;"></div>
                </div>
              </div>
              <div class="preview-content">
                <div class="preview-line">
                  <span :style="{ color: scheme.colors.green }">$</span>
                  <span :style="{ color: scheme.colors.foreground }"> echo "Hello World"</span>
                </div>
                <div class="preview-line">
                  <span :style="{ color: scheme.colors.cyan }">Hello World</span>
                </div>
                <div class="preview-line">
                  <span :style="{ color: scheme.colors.green }">$</span>
                  <span :style="{ color: scheme.colors.foreground }"> </span>
                  <span :style="{ color: scheme.colors.blue }">ls</span>
                  <span :style="{ color: scheme.colors.foreground }"> -la</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Scheme Info -->
          <div class="scheme-info">
            <div class="scheme-header">
              <h4 class="scheme-name">{{ scheme.displayName || scheme.name }}</h4>
              <div class="scheme-badges">
                <span v-if="scheme.isBuiltIn" class="badge badge--builtin">Built-in</span>
                <span v-else class="badge badge--custom">Custom</span>
                <span v-if="scheme.isDark" class="badge badge--dark">Dark</span>
                <span v-else class="badge badge--light">Light</span>
                <span v-if="defaultScheme?.id === scheme.id" class="badge badge--default">Default</span>
              </div>
            </div>
            <p class="scheme-description">{{ scheme.description || 'No description available' }}</p>
            <div class="scheme-meta">
              <span v-if="scheme.author" class="meta-item">
                <Icon name="carbon:user" />
                {{ scheme.author }}
              </span>
              <span v-if="scheme.version" class="meta-item">
                <Icon name="carbon:tag" />
                v{{ scheme.version }}
              </span>
            </div>
          </div>

          <!-- Scheme Actions -->
          <div class="scheme-actions" @click.stop>
            <button
              class="action-button action-button--small"
              @click="previewScheme(scheme)"
              title="Preview scheme"
            >
              <Icon name="carbon:view" />
            </button>
            <button
              class="action-button action-button--small"
              @click="setAsDefault(scheme.id)"
              :disabled="defaultScheme?.id === scheme.id"
              title="Set as default"
            >
              <Icon name="carbon:star" :class="{ 'text-yellow-500': defaultScheme?.id === scheme.id }" />
            </button>
            <div class="dropdown" v-if="!scheme.isBuiltIn">
              <button
                class="action-button action-button--small"
                @click="toggleDropdown(scheme.id)"
                title="More options"
              >
                <Icon name="carbon:overflow-menu-horizontal" />
              </button>
              <div v-if="activeDropdown === scheme.id" class="dropdown-menu">
                <button @click="editScheme(scheme)" class="dropdown-item">
                  <Icon name="carbon:edit" />
                  Edit
                </button>
                <button @click="duplicateScheme(scheme)" class="dropdown-item">
                  <Icon name="carbon:copy" />
                  Duplicate
                </button>
                <button @click="exportScheme(scheme)" class="dropdown-item">
                  <Icon name="carbon:download" />
                  Export
                </button>
                <div class="dropdown-divider"></div>
                <button @click="deleteScheme(scheme)" class="dropdown-item dropdown-item--danger">
                  <Icon name="carbon:trash-can" />
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Import Dialog -->
    <ColorSchemeImportDialog
      v-if="showImportDialog"
      @close="showImportDialog = false"
      @imported="handleImported"
    />

    <!-- Create Dialog -->
    <ColorSchemeCreateDialog
      v-if="showCreateDialog"
      @close="showCreateDialog = false"
      @created="handleCreated"
    />

    <!-- Edit Dialog -->
    <ColorSchemeEditDialog
      v-if="editingScheme"
      :scheme="editingScheme"
      @close="editingScheme = null"
      @updated="handleUpdated"
    />

    <!-- Preview Dialog -->
    <ColorSchemePreviewDialog
      v-if="previewingScheme"
      :scheme="previewingScheme"
      @close="previewingScheme = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useColorScheme } from '~/composables/useColorScheme'
import ColorSchemeImportDialog from './ColorSchemeImportDialog.vue'
import ColorSchemeCreateDialog from './ColorSchemeCreateDialog.vue'
import ColorSchemeEditDialog from './ColorSchemeEditDialog.vue'
import ColorSchemePreviewDialog from './ColorSchemePreviewDialog.vue'
import type { ColorScheme } from '~/stores/colorscheme'

// Props
interface Props {
  modelValue?: string
  showBuiltIn?: boolean
  showCustom?: boolean
  maxHeight?: string
}

const props = withDefaults(defineProps<Props>(), {
  showBuiltIn: true,
  showCustom: true,
  maxHeight: '600px'
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: string]
  'scheme-selected': [scheme: ColorScheme]
  'scheme-changed': [schemeId: string]
}>()

// Composables
const {
  isLoading,
  error,
  schemes,
  defaultScheme,
  selectedSchemeId,
  fetchSchemes,
  getScheme,
  setDefaultScheme,
  deleteScheme,
  exportScheme,
  selectScheme,
  createScheme,
  updateScheme
} = useColorScheme()

// Local state
const searchQuery = ref('')
const selectedCategory = ref('all')
const showOnlyEnabled = ref(false)
const showImportDialog = ref(false)
const showCreateDialog = ref(false)
const editingScheme = ref<ColorScheme | null>(null)
const previewingScheme = ref<ColorScheme | null>(null)
const activeDropdown = ref<string | null>(null)

// Computed
const filteredSchemes = computed(() => {
  let filtered = Object.values(schemes.value)

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(scheme =>
      scheme.name.toLowerCase().includes(query) ||
      scheme.displayName?.toLowerCase().includes(query) ||
      scheme.description?.toLowerCase().includes(query) ||
      scheme.author?.toLowerCase().includes(query)
    )
  }

  // Filter by category
  switch (selectedCategory.value) {
    case 'dark':
      filtered = filtered.filter(scheme => scheme.isDark)
      break
    case 'light':
      filtered = filtered.filter(scheme => !scheme.isDark)
      break
    case 'builtin':
      filtered = filtered.filter(scheme => scheme.isBuiltIn)
      break
    case 'custom':
      filtered = filtered.filter(scheme => !scheme.isBuiltIn)
      break
  }

  // Filter by built-in/custom visibility
  if (!props.showBuiltIn) {
    filtered = filtered.filter(scheme => !scheme.isBuiltIn)
  }
  if (!props.showCustom) {
    filtered = filtered.filter(scheme => scheme.isBuiltIn)
  }

  // Filter by enabled schemes
  if (showOnlyEnabled.value) {
    const enabledIds = defaultScheme.value?.id ? [defaultScheme.value.id] : []
    filtered = filtered.filter(scheme => enabledIds.includes(scheme.id))
  }

  return filtered
})

// Methods
const refreshSchemes = async () => {
  await fetchSchemes()
}

const selectSchemeHandler = async (schemeId: string) => {
  selectScheme(schemeId)
  emit('update:modelValue', schemeId)
  emit('scheme-changed', schemeId)

  const scheme = getScheme(schemeId)
  if (scheme) {
    emit('scheme-selected', scheme)
  }
}

const setAsDefault = async (schemeId: string) => {
  await setDefaultScheme(schemeId)
  await refreshSchemes()
}

const previewScheme = async (scheme: ColorScheme) => {
  previewingScheme.value = scheme
}

const editScheme = (scheme: ColorScheme) => {
  editingScheme.value = scheme
  activeDropdown.value = null
}

const duplicateScheme = async (scheme: ColorScheme) => {
  const duplicatedScheme: Partial<ColorScheme> = {
    ...scheme,
    id: `user-${Date.now()}`,
    name: `${scheme.name} (Copy)`,
    displayName: `${scheme.displayName || scheme.name} (Copy)`,
    isBuiltIn: false,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  }

  delete (duplicatedScheme as any).id
  await createScheme(duplicatedScheme)
  await refreshSchemes()
  activeDropdown.value = null
}

const exportSchemeHandler = async (scheme: ColorScheme) => {
  try {
    await exportScheme(scheme.id, 'json')
  } catch (error) {
    console.error('Failed to export scheme:', error)
  }
  activeDropdown.value = null
}

const deleteSchemeHandler = async (scheme: ColorScheme) => {
  if (confirm(`Are you sure you want to delete the color scheme "${scheme.displayName || scheme.name}"?`)) {
    try {
      await deleteScheme(scheme.id)
      await refreshSchemes()
    } catch (error) {
      console.error('Failed to delete scheme:', error)
    }
  }
  activeDropdown.value = null
}

const toggleDropdown = (schemeId: string) => {
  activeDropdown.value = activeDropdown.value === schemeId ? null : schemeId
}

const getPreviewStyles = (scheme: ColorScheme) => {
  return {
    backgroundColor: scheme.colors.background,
    color: scheme.colors.foreground,
    fontFamily: scheme.font.family,
    fontSize: '12px',
    borderRadius: '6px',
    border: `1px solid ${scheme.colors.selection}`
  }
}

const handleImported = () => {
  showImportDialog.value = false
  refreshSchemes()
}

const handleCreated = () => {
  showCreateDialog.value = false
  refreshSchemes()
}

const handleUpdated = () => {
  editingScheme.value = null
  refreshSchemes()
}

const handleClickOutside = (event: MouseEvent) => {
  if (activeDropdown.value && !(event.target as HTMLElement).closest('.dropdown')) {
    activeDropdown.value = null
  }
}

// Lifecycle
onMounted(async () => {
  await refreshSchemes()
  if (props.modelValue) {
    selectSchemeHandler(props.modelValue)
  }

  // Add click outside listener
  document.addEventListener('click', handleClickOutside)
})

// Watch for external value changes
watch(() => props.modelValue, (newValue) => {
  if (newValue && newValue !== selectedSchemeId.value) {
    selectSchemeHandler(newValue)
  }
})
</script>

<style scoped>
.color-scheme-selector {
  @apply flex flex-col h-full;
}

.selector-header {
  @apply flex items-center justify-between p-4 border-b border-border;
}

.selector-title {
  @apply text-lg font-semibold text-text;
  margin: 0;
}

.selector-actions {
  @apply flex items-center gap-2;
}

.action-button {
  @apply p-2 rounded-lg hover:bg-surface-hover text-text-secondary hover:text-text transition-colors;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-button:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.action-button--small {
  @apply p-1.5;
}

.selector-controls {
  @apply flex flex-col gap-3 p-4 border-b border-border;
}

.search-box {
  @apply relative flex items-center;
}

.search-icon {
  @apply absolute left-3 text-text-secondary;
}

.search-input {
  @apply w-full pl-10 pr-4 py-2 bg-surface border border-border rounded-lg text-text placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary;
}

.filter-controls {
  @apply flex items-center justify-between gap-4;
}

.filter-select {
  @apply px-3 py-2 bg-surface border border-border rounded-lg text-text focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary;
}

.checkbox-label {
  @apply flex items-center gap-2 text-sm text-text-secondary cursor-pointer;
}

.checkbox-input {
  @apply rounded border-border text-primary focus:ring-primary focus:border-primary;
}

.schemes-container {
  @apply flex-1 overflow-auto p-4;
  max-height: v-bind(maxHeight);
}

.loading-state,
.error-state,
.empty-state {
  @apply flex flex-col items-center justify-center h-64 text-center;
}

.loading-spinner {
  @apply w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin mb-4;
}

.error-icon,
.empty-icon {
  @apply w-12 h-12 text-text-secondary mb-4;
}

.error-message {
  @apply text-text-secondary mb-4;
}

.retry-button,
.create-button {
  @apply px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-hover transition-colors flex items-center gap-2;
  border: none;
  cursor: pointer;
}

.schemes-grid {
  @apply grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4;
}

.scheme-card {
  @apply relative bg-surface border border-border rounded-lg overflow-hidden cursor-pointer transition-all hover:border-primary hover:shadow-lg;
}

.scheme-card--selected {
  @apply border-primary ring-2 ring-primary ring-opacity-20;
}

.scheme-card--default {
  @apply ring-2 ring-yellow-500 ring-opacity-20;
}

.scheme-preview {
  @apply p-3 bg-surface-alt;
}

.preview-terminal {
  @apply p-3 font-mono text-xs rounded shadow-inner;
  min-height: 80px;
}

.preview-header {
  @apply flex items-center justify-between mb-2;
}

.preview-title {
  @apply text-xs font-medium;
}

.preview-controls {
  @apply flex items-center gap-1;
}

.preview-dot {
  @apply w-2 h-2 rounded-full;
}

.preview-content {
  @apply space-y-1;
}

.preview-line {
  @apply leading-tight;
}

.scheme-info {
  @apply p-4;
}

.scheme-header {
  @apply flex items-start justify-between mb-2;
}

.scheme-name {
  @apply text-base font-medium text-text m-0;
}

.scheme-badges {
  @apply flex flex-wrap gap-1;
}

.badge {
  @apply px-2 py-1 text-xs rounded-full font-medium;
}

.badge--builtin {
  @apply bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300;
}

.badge--custom {
  @apply bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300;
}

.badge--dark {
  @apply bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300;
}

.badge--light {
  @apply bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300;
}

.badge--default {
  @apply bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300;
}

.scheme-description {
  @apply text-sm text-text-secondary mb-3 m-0;
}

.scheme-meta {
  @apply flex flex-wrap gap-3 text-xs text-text-secondary;
}

.meta-item {
  @apply flex items-center gap-1;
}

.scheme-actions {
  @apply absolute top-2 right-2 flex items-center gap-1;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.scheme-card:hover .scheme-actions {
  @apply opacity-100;
}

.dropdown {
  @apply relative;
}

.dropdown-menu {
  @apply absolute right-0 top-full mt-1 w-48 bg-surface border border-border rounded-lg shadow-lg z-50;
}

.dropdown-item {
  @apply w-full px-3 py-2 text-left text-sm text-text hover:bg-surface-hover transition-colors flex items-center gap-2;
  border: none;
  background: none;
  cursor: pointer;
}

.dropdown-item--danger {
  @apply text-error hover:bg-error hover:bg-opacity-10;
}

.dropdown-divider {
  @apply h-px bg-border my-1;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}
</style>