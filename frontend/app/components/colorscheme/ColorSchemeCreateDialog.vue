<template>
  <div class="dialog-overlay" @click="close">
    <div class="dialog-container" @click.stop>
      <div class="dialog-header">
        <h2 class="dialog-title">Create Color Scheme</h2>
        <button class="close-button" @click="close" title="Close">
          <Icon name="carbon:close" />
        </button>
      </div>

      <div class="dialog-content">
        <p class="dialog-description">
          Create a new custom color scheme by starting with an existing scheme or creating from scratch.
        </p>

        <div class="create-options">
          <button class="create-option" @click="createFromScratch">
            <Icon name="carbon:add-alt" class="option-icon" />
            <div class="option-content">
              <h3 class="option-title">Create from Scratch</h3>
              <p class="option-description">Start with a blank color scheme template</p>
            </div>
          </button>

          <button class="create-option" @click="createFromExisting">
            <Icon name="carbon:copy" class="option-icon" />
            <div class="option-content">
              <h3 class="option-title">Duplicate Existing</h3>
              <p class="option-description">Copy an existing color scheme and customize it</p>
            </div>
          </button>
        </div>
      </div>

      <div class="dialog-footer">
        <button class="button button--secondary" @click="close">Cancel</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useColorScheme } from '~/composables/useColorScheme'
import type { ColorScheme } from '~/stores/colorscheme'

// Emits
const emit = defineEmits<{
  close: []
  created: []
}>()

// Composables
const { defaultScheme, createScheme } = useColorScheme()

// Methods
const close = () => {
  emit('close')
}

const createFromScratch = async () => {
  const newScheme: Partial<ColorScheme> = {
    id: `user-${Date.now()}`,
    name: 'Custom Color Scheme',
    displayName: 'Custom Color Scheme',
    description: 'A custom color scheme created from scratch',
    isBuiltIn: false,
    isDark: true
  }

  try {
    await createScheme(newScheme)
    emit('created')
    emit('close')
  } catch (error) {
    console.error('Failed to create color scheme:', error)
    alert('Failed to create color scheme. Please try again.')
  }
}

const createFromExisting = async () => {
  if (!defaultScheme.value) return

  const newScheme: Partial<ColorScheme> = {
    ...defaultScheme.value,
    id: `user-${Date.now()}`,
    name: `${defaultScheme.value.name} (Copy)`,
    displayName: `${defaultScheme.value.displayName || defaultScheme.value.name} (Copy)`,
    description: `A copy of ${defaultScheme.value.displayName || defaultScheme.value.name}`,
    isBuiltIn: false
  }

  delete (newScheme as any).id

  try {
    await createScheme(newScheme)
    emit('created')
    emit('close')
  } catch (error) {
    console.error('Failed to create color scheme:', error)
    alert('Failed to create color scheme. Please try again.')
  }
}
</script>

<style scoped>
.dialog-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4;
}

.dialog-container {
  @apply bg-surface rounded-xl shadow-2xl max-w-lg w-full overflow-hidden flex flex-col;
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

.create-options {
  @apply space-y-4;
}

.create-option {
  @apply w-full p-4 border border-border rounded-lg hover:border-primary hover:bg-surface-hover transition-all text-left;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4;
}

.option-icon {
  @apply w-12 h-12 text-primary flex-shrink-0;
}

.option-content {
  @apply flex-1;
}

.option-title {
  @apply text-base font-medium text-text mb-1;
}

.option-description {
  @apply text-sm text-text-secondary;
}

.dialog-footer {
  @apply flex items-center justify-end gap-3 p-6 border-t border-border;
}

.button {
  @apply px-4 py-2 rounded-lg font-medium transition-colors;
  border: none;
  cursor: pointer;
}

.button--secondary {
  @apply bg-surface-hover text-text hover:bg-surface-alt;
}
</style>