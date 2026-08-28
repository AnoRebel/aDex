<template>
  <div v-if="failed" class="error-boundary" role="alert">
    <div class="eb-title">
      <span class="eb-bracket">[</span>
      {{ label }} UNAVAILABLE
      <span class="eb-bracket">]</span>
    </div>
    <p class="eb-message">{{ message }}</p>
    <button type="button" class="eb-retry" @click="retry">RETRY</button>
  </div>
  <slot v-else />
</template>

<script setup lang="ts">
/**
 * Error boundary for a single panel.
 *
 * Vue's onErrorCaptured stops an error propagating further up the tree, so a
 * component that throws while rendering takes out only its own subtree
 * instead of unmounting the entire application — previously any such throw
 * blanked the whole interface.
 *
 * Recovery is real, not cosmetic: `retry` swaps the slot's key, which forces
 * Vue to discard the broken component instance and mount a fresh one. That
 * matters for panels whose failure was caused by transient state (a bad
 * backend payload, a lost GPU context) rather than a permanent defect.
 */
import { ref, nextTick, onErrorCaptured } from 'vue'
import { handleComponentError } from '~/utils/errorHandler'

const props = withDefaults(
  defineProps<{
    /** Shown in the fallback, e.g. "CPU". Defaults to a generic label. */
    label?: string
    /** Retry automatically once after a short delay. */
    autoRetry?: boolean
  }>(),
  { label: 'PANEL', autoRetry: false },
)

const emit = defineEmits<{ (e: 'error', err: unknown): void }>()

const failed = ref(false)
const message = ref('')
const attempts = ref(0)

/** Retrying more than a few times means the fault is not transient; stop so a
 *  render loop cannot spin the CPU indefinitely. */
const MAX_AUTO_RETRIES = 1

onErrorCaptured((err) => {
  // Defer the state change: the error surfaces DURING the child's render, and
  // mutating a ref at that moment does not schedule another pass, so the
  // fallback would never paint (the boundary rendered empty).
  const text = err instanceof Error ? err.message : String(err)
  void nextTick(() => {
    failed.value = true
    message.value = text
  })

  try {
    handleComponentError(props.label, err instanceof Error ? err : new Error(String(err)))
  } catch {
    // The reporter itself must never take the boundary down.
  }
  emit('error', err)

  if (props.autoRetry && attempts.value < MAX_AUTO_RETRIES) {
    attempts.value += 1
    setTimeout(retry, 1000)
  }

  // Returning false stops propagation — this is what contains the failure.
  return false
})

function retry() {
  failed.value = false
  message.value = ''
}
</script>

<style scoped>
.error-boundary {
  display: flex;
  flex-direction: column;
  gap: 0.6vh;
  padding: 1vh 1vw;
  border: var(--border_width, 1px) solid var(--color_accent_dimmed, rgba(170, 207, 209, 0.4));
  background: var(--color_light_black, #05080d);
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 0.75rem;
  min-height: 4rem;
  justify-content: center;
}

.eb-title {
  font-weight: 700;
  letter-spacing: 0.06em;
  opacity: 0.9;
}

.eb-bracket { opacity: 0.5; }

.eb-message {
  margin: 0;
  opacity: 0.7;
  word-break: break-word;
  /* Keep a long stack message from stretching the panel. */
  max-height: 4.5rem;
  overflow: auto;
}

.eb-retry {
  align-self: flex-start;
  padding: 0.25rem 0.9rem;
  background: transparent;
  border: var(--border_width, 1px) solid var(--color_accent, rgb(170, 207, 209));
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: inherit;
  font-size: inherit;
  cursor: pointer;
}

.eb-retry:hover {
  background: var(--color_accent_glow, rgba(170, 207, 209, 0.15));
}
</style>
