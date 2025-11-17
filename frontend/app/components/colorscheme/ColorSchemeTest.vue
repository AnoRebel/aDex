<template>
  <div class="color-scheme-test">
    <h2>Color Scheme Test</h2>

    <div class="test-controls">
      <button @click="testColorSchemeFetching" :disabled="isLoading">
        Test Fetch Schemes
      </button>
      <button @click="testColorSchemeApplication" :disabled="isLoading || !selectedScheme">
        Test Apply Scheme
      </button>
      <button @click="testColorSchemePersistence" :disabled="isLoading">
        Test Persistence
      </button>
    </div>

    <div class="test-status">
      <div v-if="isLoading" class="status-loading">
        <Icon name="carbon:loading" class="animate-spin" />
        Loading...
      </div>
      <div v-if="error" class="status-error">
        <Icon name="carbon:warning-alt" />
        {{ error }}
      </div>
      <div v-if="success" class="status-success">
        <Icon name="carbon:checkmark-filled" />
        {{ success }}
      </div>
    </div>

    <div class="test-results">
      <h3>Test Results</h3>
      <div class="result-item">
        <strong>Total Schemes:</strong> {{ Object.keys(schemes).length }}
      </div>
      <div class="result-item">
        <strong>Current Scheme:</strong> {{ selectedScheme?.displayName || selectedScheme?.name || 'None' }}
      </div>
      <div class="result-item">
        <strong>Default Scheme:</strong> {{ defaultScheme?.displayName || defaultScheme?.name || 'None' }}
      </div>
    </div>

    <!-- Color Scheme Preview -->
    <div v-if="selectedScheme" class="scheme-preview">
      <h3>Current Scheme Preview</h3>
      <div class="preview-grid">
        <div class="color-item">
          <div class="color-swatch" :style="{ backgroundColor: selectedScheme.colors.background }"></div>
          <span>Background</span>
          <code>{{ selectedScheme.colors.background }}</code>
        </div>
        <div class="color-item">
          <div class="color-swatch" :style="{ backgroundColor: selectedScheme.colors.foreground }"></div>
          <span>Foreground</span>
          <code>{{ selectedScheme.colors.foreground }}</code>
        </div>
        <div class="color-item">
          <div class="color-swatch" :style="{ backgroundColor: selectedScheme.colors.cursor }"></div>
          <span>Cursor</span>
          <code>{{ selectedScheme.colors.cursor }}</code>
        </div>
        <div class="color-item">
          <div class="color-swatch" :style="{ backgroundColor: selectedScheme.colors.red }"></div>
          <span>Red</span>
          <code>{{ selectedScheme.colors.red }}</code>
        </div>
        <div class="color-item">
          <div class="color-swatch" :style="{ backgroundColor: selectedScheme.colors.green }"></div>
          <span>Green</span>
          <code>{{ selectedScheme.colors.green }}</code>
        </div>
        <div class="color-item">
          <div class="color-swatch" :style="{ backgroundColor: selectedScheme.colors.blue }"></div>
          <span>Blue</span>
          <code>{{ selectedScheme.colors.blue }}</code>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useColorScheme } from '~/composables/useColorScheme'

// Composables
const {
  isLoading,
  error,
  schemes,
  selectedScheme,
  defaultScheme,
  fetchSchemes,
  selectScheme,
  setDefaultScheme,
  getXtermTheme
} = useColorScheme()

// Local state
const success = ref('')

// Test methods
const testColorSchemeFetching = async () => {
  try {
    success.value = ''
    error.value = null

    await fetchSchemes()

    const schemeCount = Object.keys(schemes.value).length
    success.value = `Successfully fetched ${schemeCount} color schemes`
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch color schemes'
  }
}

const testColorSchemeApplication = async () => {
  try {
    success.value = ''
    error.value = null

    if (!selectedScheme.value) {
      // Select first available scheme
      const schemeIds = Object.keys(schemes.value)
      if (schemeIds.length === 0) {
        throw new Error('No color schemes available')
      }
      selectScheme(schemeIds[0])
    }

    const theme = getXtermTheme(selectedScheme.value)

    if (theme && Object.keys(theme).length > 0) {
      success.value = `Successfully applied "${selectedScheme.value.displayName || selectedScheme.value.name}" theme with ${Object.keys(theme).length} colors`
    } else {
      throw new Error('Failed to generate xterm theme')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to apply color scheme'
  }
}

const testColorSchemePersistence = async () => {
  try {
    success.value = ''
    error.value = null

    if (!selectedScheme.value) {
      throw new Error('No color scheme selected')
    }

    // Test setting as default
    await setDefaultScheme(selectedScheme.value.id)

    // Simulate page reload by waiting a bit
    await new Promise(resolve => setTimeout(resolve, 1000))

    // Check if default persisted
    if (defaultScheme.value?.id === selectedScheme.value.id) {
      success.value = `Successfully persisted "${selectedScheme.value.displayName || selectedScheme.value.name}" as default scheme`
    } else {
      throw new Error('Default scheme was not persisted')
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to test color scheme persistence'
  }
}
</script>

<style scoped>
.color-scheme-test {
  @apply p-6 max-w-4xl mx-auto;
}

h2, h3 {
  @apply text-xl font-semibold text-text mb-4;
}

h3 {
  @apply text-lg;
}

.test-controls {
  @apply flex gap-4 mb-6;
}

.test-controls button {
  @apply px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors;
}

.test-status {
  @apply mb-6 space-y-2;
}

.status-loading,
.status-error,
.status-success {
  @apply flex items-center gap-2 p-3 rounded-lg;
}

.status-loading {
  @apply bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300;
}

.status-error {
  @apply bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300;
}

.status-success {
  @apply bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300;
}

.test-results {
  @apply p-4 bg-surface border border-border rounded-lg mb-6;
}

.result-item {
  @apply py-1 text-sm;
}

.scheme-preview {
  @apply p-4 bg-surface border border-border rounded-lg;
}

.preview-grid {
  @apply grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4;
}

.color-item {
  @apply flex items-center gap-3 p-2 bg-surface-hover rounded-lg;
}

.color-swatch {
  @apply w-8 h-8 rounded border border-border;
}

.color-item span {
  @apply flex-1 text-sm font-medium;
}

.color-item code {
  @apply text-xs font-mono text-text-secondary;
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