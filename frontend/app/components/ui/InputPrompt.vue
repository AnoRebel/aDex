<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="ip-overlay"
      role="dialog"
      :aria-labelledby="`ip-title-${id}`"
      aria-modal="true"
      @click.self="cancel"
      @keydown.esc="cancel"
    >
      <div class="ip-card mod-panel">
        <div :id="`ip-title-${id}`" class="ip-title">
          <span class="ip-bracket">[</span>
          {{ title }}
          <span class="ip-bracket">]</span>
        </div>
        <p v-if="description" class="ip-desc">{{ description }}</p>
        <input
          ref="inputRef"
          v-model="value"
          type="text"
          class="ip-input"
          :placeholder="placeholder"
          @keydown.enter="confirm"
        />
        <div class="ip-actions">
          <button type="button" class="ip-btn" @click="cancel">CANCEL</button>
          <button type="button" class="ip-btn ip-btn-primary" @click="confirm">
            {{ confirmLabel }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * InputPrompt — small modal asking for a single string. Used by the file
 * manager's "New Folder", "New File", and "Rename" actions. Confirm-on-Enter,
 * Cancel-on-Esc/Backdrop. Trimmed empty input is treated as Cancel.
 */
import { ref, watch, nextTick, useId } from 'vue'

const props = defineProps<{
  modelValue: boolean
  title: string
  description?: string
  placeholder?: string
  initialValue?: string
  confirmLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm', value: string): void
  (e: 'cancel'): void
}>()

const id = useId()
const value = ref(props.initialValue ?? '')
const inputRef = ref<HTMLInputElement | null>(null)
const confirmLabel = computedConfirm()

function computedConfirm() {
  return props.confirmLabel ?? 'OK'
}

watch(
  () => props.modelValue,
  async (open) => {
    if (open) {
      value.value = props.initialValue ?? ''
      await nextTick()
      inputRef.value?.focus()
      inputRef.value?.select()
    }
  },
  { immediate: true },
)

function confirm() {
  const trimmed = value.value.trim()
  if (!trimmed) {
    cancel()
    return
  }
  emit('confirm', trimmed)
  emit('update:modelValue', false)
}

function cancel() {
  emit('cancel')
  emit('update:modelValue', false)
}
</script>

<style scoped>
.ip-overlay {
  position: fixed;
  inset: 0;
  z-index: 10001;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
}

.ip-card {
  --corner-cut: 1vh;
  width: min(38vw, 60vh);
  padding: 2vh 2vw;
  background: var(--bg-glass, rgba(5, 8, 13, 0.85));
  box-shadow: var(--glow-strong);
}

.ip-title {
  font-size: var(--t-xl, 1.85vh);
  letter-spacing: 0.2em;
  color: var(--accent);
  margin-bottom: 0.6vh;
}

.ip-bracket {
  opacity: 0.6;
}

.ip-desc {
  font-size: var(--t-md);
  opacity: 0.7;
  margin: 0 0 1.2vh 0;
}

.ip-input {
  width: 100%;
  padding: 0.8vh 1vw;
  background: rgba(0, 0, 0, 0.4);
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  font-family: var(--font_main);
  font-size: var(--t-md);
  margin-bottom: 1.4vh;
}

.ip-input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: var(--glow-mid);
}

.ip-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1vw;
}

.ip-btn {
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

.ip-btn:hover {
  background: var(--surface-2);
}

.ip-btn-primary {
  background: var(--surface-2);
  border-color: var(--accent);
}

.ip-btn-primary:hover {
  background: var(--surface-3);
}
</style>
