<template>
  <div class="dialog-overlay" @click="close">
    <div class="dialog-container" @click.stop>
      <div class="dialog-header">
        <h2 class="dialog-title">Import Color Scheme</h2>
        <button class="close-button" @click="close" title="Close">
          <Icon name="carbon:close" />
        </button>
      </div>

      <div class="dialog-content">
        <p class="dialog-description">
          Import a color scheme from a JSON file. Select a file that contains a valid color scheme configuration.
        </p>

        <div class="import-area">
          <input
            ref="fileInput"
            type="file"
            accept=".json"
            @change="handleFileSelect"
            class="file-input"
          />
          <div class="import-drop-zone" @click="triggerFileSelect" @dragover.prevent @drop="handleFileDrop">
            <Icon name="carbon:upload" class="import-icon" />
            <p class="import-text">Click to select or drag and drop a JSON file</p>
            <p class="import-hint">Supported format: JSON</p>
          </div>
        </div>

        <div v-if="selectedFile" class="file-info">
          <Icon name="carbon:document" class="file-icon" />
          <span class="file-name">{{ selectedFile.name }}</span>
          <button @click="clearFile" class="clear-button">
            <Icon name="carbon:close" />
          </button>
        </div>
      </div>

      <div class="dialog-footer">
        <button class="button button--secondary" @click="close">Cancel</button>
        <button
          class="button button--primary"
          @click="importScheme"
          :disabled="!selectedFile || isImporting"
        >
          <Icon v-if="isImporting" name="carbon:loading" class="animate-spin" />
          <Icon v-else name="carbon:download" />
          {{ isImporting ? 'Importing...' : 'Import' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useColorScheme } from '~/composables/useColorScheme'

// Emits
const emit = defineEmits<{
  close: []
  imported: []
}>()

// Composables
const { importScheme: importSchemeHandler } = useColorScheme()

// State
const fileInput = ref<HTMLInputElement>()
const selectedFile = ref<File | null>(null)
const isImporting = ref(false)

// Methods
const close = () => {
  emit('close')
}

const triggerFileSelect = () => {
  fileInput.value?.click()
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    selectedFile.value = file
  }
}

const handleFileDrop = (event: DragEvent) => {
  event.preventDefault()
  const file = event.dataTransfer?.files[0]
  if (file && file.type === 'application/json') {
    selectedFile.value = file
  }
}

const clearFile = () => {
  selectedFile.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const importScheme = async () => {
  if (!selectedFile.value) return

  isImporting.value = true

  try {
    const text = await selectedFile.value.text()
    const data = JSON.parse(text)

    await importSchemeHandler(data, 'json')
    emit('imported')
    emit('close')
  } catch (error) {
    console.error('Failed to import color scheme:', error)
    alert('Failed to import color scheme. Please check the file format.')
  } finally {
    isImporting.value = false
  }
}
</script>

<style scoped>
.dialog-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4;
}

.dialog-container {
  @apply bg-surface rounded-xl shadow-2xl max-w-md w-full overflow-hidden flex flex-col;
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
  @apply p-6 space-y-6;
}

.dialog-description {
  @apply text-text-secondary;
}

.import-area {
  @apply space-y-4;
}

.file-input {
  @apply hidden;
}

.import-drop-zone {
  @apply border-2 border-dashed border-border rounded-lg p-8 text-center cursor-pointer hover:border-primary transition-colors;
}

.import-icon {
  @apply w-12 h-12 text-text-secondary mx-auto mb-4;
}

.import-text {
  @apply text-text font-medium mb-2;
}

.import-hint {
  @apply text-sm text-text-secondary;
}

.file-info {
  @apply flex items-center gap-3 p-3 bg-surface-hover rounded-lg;
}

.file-icon {
  @apply text-text-secondary;
}

.file-name {
  @apply flex-1 text-sm text-text;
}

.clear-button {
  @apply p-1 rounded hover:bg-surface text-text-secondary hover:text-text transition-colors;
  border: none;
  background: none;
  cursor: pointer;
}

.dialog-footer {
  @apply flex items-center justify-end gap-3 p-6 border-t border-border;
}

.button {
  @apply px-4 py-2 rounded-lg font-medium transition-colors flex items-center gap-2;
  border: none;
  cursor: pointer;
}

.button:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.button--secondary {
  @apply bg-surface-hover text-text hover:bg-surface-alt;
}

.button--primary {
  @apply bg-primary text-white hover:bg-primary-hover;
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