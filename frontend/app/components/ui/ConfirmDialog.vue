<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="cd-overlay"
      role="alertdialog"
      :aria-labelledby="`cd-title-${id}`"
      :aria-describedby="`cd-body-${id}`"
      aria-modal="true"
      @click.self="cancel"
      @keydown.esc="cancel"
    >
      <div class="cd-card mod-panel" :class="{ 'cd-card-danger': danger }">
        <div :id="`cd-title-${id}`" class="cd-title">
          <span class="cd-bracket">[</span>
          {{ title }}
          <span class="cd-bracket">]</span>
        </div>
        <p :id="`cd-body-${id}`" class="cd-body">{{ message }}</p>
        <div class="cd-actions">
          <button ref="cancelBtn" type="button" class="cd-btn" @click="cancel">
            {{ cancelLabel ?? 'CANCEL' }}
          </button>
          <button
            type="button"
            class="cd-btn"
            :class="danger ? 'cd-btn-danger' : 'cd-btn-primary'"
            @click="confirm"
          >
            {{ confirmLabel ?? 'OK' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * ConfirmDialog — minimal dialog for destructive or significant actions.
 * Default focus is on the cancel button so Enter never accidentally
 * confirms a destructive operation. The user must explicitly Tab+Enter
 * or click the danger button.
 */
import { ref, watch, nextTick, useId } from 'vue'

const props = defineProps<{
  modelValue: boolean
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const id = useId()
const cancelBtn = ref<HTMLButtonElement | null>(null)

function confirm() {
  emit('confirm')
  emit('update:modelValue', false)
}
function cancel() {
  emit('cancel')
  emit('update:modelValue', false)
}

// Move focus to the (safe) cancel button whenever the dialog opens, so a
// stray Enter cannot trigger the destructive action.
watch(
  () => props.modelValue,
  async (open) => {
    if (open) {
      await nextTick()
      cancelBtn.value?.focus()
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.cd-overlay {
  position: fixed;
  inset: 0;
  z-index: 10002;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
}

.cd-card {
  --corner-cut: 1vh;
  width: min(38vw, 60vh);
  padding: 2vh 2vw;
  background: var(--bg-glass, rgba(5, 8, 13, 0.85));
  box-shadow: var(--glow-strong);
}

.cd-card-danger {
  border-color: var(--err);
  box-shadow: 0 0 0.6vh rgba(239, 68, 68, 0.4),
              0 0 2vh rgba(239, 68, 68, 0.2);
}

.cd-title {
  font-size: var(--t-xl, 1.85vh);
  letter-spacing: 0.2em;
  color: var(--accent);
  margin-bottom: 0.8vh;
}

.cd-card-danger .cd-title {
  color: var(--err);
}

.cd-bracket {
  opacity: 0.6;
}

.cd-body {
  font-size: var(--t-md);
  opacity: 0.85;
  margin: 0 0 1.6vh 0;
  line-height: 1.5;
}

.cd-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1vw;
}

.cd-btn {
  padding: 0.7vh 2vw;
  background: transparent;
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  font-family: inherit;
  font-size: var(--t-md);
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
}

.cd-btn:hover {
  background: var(--surface-2);
}

.cd-btn-primary {
  background: var(--surface-2);
  border-color: var(--accent);
}

.cd-btn-danger {
  background: rgba(239, 68, 68, 0.15);
  border-color: var(--err);
  color: var(--err);
}

.cd-btn-danger:hover {
  background: rgba(239, 68, 68, 0.3);
}
</style>
