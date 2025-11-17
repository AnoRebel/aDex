<template>
  <div class="dialog-overlay" @click="close">
    <div class="dialog-container" @click.stop>
      <div class="dialog-header">
        <h2 class="dialog-title">Edit Color Scheme</h2>
        <button class="close-button" @click="close" title="Close">
          <Icon name="carbon:close" />
        </button>
      </div>

      <div class="dialog-content">
        <p class="dialog-description">
          Color scheme editing functionality is not yet implemented. This would open a comprehensive color editor.
        </p>

        <div class="coming-soon">
          <Icon name="carbon:tools" class="coming-soon-icon" />
          <h3 class="coming-soon-title">Coming Soon</h3>
          <p class="coming-soon-description">
            Advanced color scheme editing with real-time preview, color picker, and palette management will be available in a future update.
          </p>
        </div>
      </div>

      <div class="dialog-footer">
        <button class="button button--secondary" @click="close">Close</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ColorScheme } from '~/stores/colorscheme'

// Props
interface Props {
  scheme: ColorScheme
}

defineProps<Props>()

// Emits
const emit = defineEmits<{
  close: []
  updated: []
}>()

// Methods
const close = () => {
  emit('close')
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

.coming-soon {
  @apply text-center py-8;
}

.coming-soon-icon {
  @apply w-16 h-16 text-text-secondary mx-auto mb-4;
}

.coming-soon-title {
  @apply text-lg font-medium text-text mb-2;
}

.coming-soon-description {
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