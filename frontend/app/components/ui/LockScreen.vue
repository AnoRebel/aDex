<template>
  <div v-if="locked" class="lock-overlay" role="dialog" aria-modal="true" aria-label="Session locked">
    <div class="lock-panel">
      <div class="lock-title">
        <span class="lock-bracket">[</span> SESSION LOCKED <span class="lock-bracket">]</span>
      </div>
      <p class="lock-hint">Enter your passphrase to resume.</p>

      <input
        ref="inputRef"
        v-model="passphrase"
        type="password"
        class="lock-input"
        placeholder="passphrase"
        autocomplete="off"
        spellcheck="false"
        @keydown.enter.prevent="attemptUnlock"
      />

      <p v-if="error" class="lock-error">{{ error }}</p>

      <button type="button" class="lock-button" :disabled="busy" @click="attemptUnlock">
        {{ busy ? 'CHECKING…' : 'UNLOCK' }}
      </button>

      <p class="lock-note">
        This protects an unattended session. It is not a substitute for
        locking your operating-system session.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Session lock overlay.
 *
 * The overlay is the visible half; the backend refuses filesystem, terminal
 * and process-control calls while locked, so covering the UI is not what makes
 * this work — dismissing the overlay by hand gains nothing.
 *
 * Terminals keep running while locked: locking hides the session, it does not
 * destroy the user's work.
 */
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useIdle, useEventListener } from '@vueuse/core'
import {
  IsSessionLocked,
  UnlockSession,
  LockSession,
  IsLockConfigured,
  GetLockIdleTimeout,
} from '~/lib/wailsjs/coordinator'

const locked = ref(false)
const passphrase = ref('')
const error = ref('')
const busy = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

const idleSeconds = ref(0)
const configured = ref(false)

/** Poll the backend for the authoritative state. The lock can also be engaged
 *  from elsewhere (a keybind, or the idle timer), so the overlay follows the
 *  backend rather than owning the state itself. */
async function refresh() {
  try {
    locked.value = await IsSessionLocked()
    configured.value = await IsLockConfigured()
    idleSeconds.value = await GetLockIdleTimeout()
  } catch {
    // Backend unreachable (browser preview): stay unlocked rather than
    // trapping the user behind an overlay they cannot dismiss.
    locked.value = false
  }
}

async function attemptUnlock() {
  if (busy.value || !passphrase.value) return
  busy.value = true
  error.value = ''
  try {
    await UnlockSession(passphrase.value)
    passphrase.value = ''
    await refresh()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Incorrect passphrase'
    passphrase.value = ''
  } finally {
    busy.value = false
  }
}

/** Engage the lock. Exposed so a keybind and the settings panel can call it. */
async function lock() {
  try {
    await LockSession()
    await refresh()
  } catch (err) {
    // Refused when no passphrase is configured — that guard exists so a user
    // cannot lock themselves out with no way back in.
    console.warn('[lock] could not lock:', err)
  }
}
defineExpose({ lock, refresh })

// Auto-lock on idle. `useIdle` needs a fixed timeout, so re-create the watcher
// when the configured value changes.
const { idle } = useIdle(60_000)
watch([idle, idleSeconds], async ([isIdle, secs]) => {
  if (isIdle && secs > 0 && configured.value && !locked.value) {
    await lock()
  }
})

// Focus the field as soon as the overlay appears, so the user can just type.
watch(locked, async (isLocked) => {
  if (isLocked) {
    await nextTick()
    inputRef.value?.focus()
  }
})

// While locked, keep focus in the overlay so keystrokes cannot reach the
// terminal behind it.
useEventListener(document, 'focusin', (e) => {
  if (!locked.value) return
  const target = e.target as Node | null
  if (target && inputRef.value && !inputRef.value.contains(target)) {
    inputRef.value.focus()
  }
})

let poll: ReturnType<typeof setInterval> | null = null
onMounted(async () => {
  await refresh()
  poll = setInterval(refresh, 3000)
})
onUnmounted(() => {
  if (poll) clearInterval(poll)
})
</script>

<style scoped>
.lock-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color_black, #000);
  /* Opaque: the point is that the session's contents are not visible. */
}

.lock-panel {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  padding: 2rem 2.5rem;
  min-width: min(28rem, 90vw);
  border: var(--border_width, 1px) solid var(--color_accent, rgb(170, 207, 209));
  background: var(--color_light_black, #05080d);
  text-align: center;
}

.lock-title {
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 1.1rem;
  letter-spacing: 0.14em;
}

.lock-bracket { opacity: 0.5; }

.lock-hint,
.lock-note {
  margin: 0;
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 0.75rem;
  opacity: 0.7;
}

.lock-note { opacity: 0.45; font-size: 0.68rem; }

.lock-input {
  padding: 0.5em 0.8em;
  background: var(--color_black, #000);
  border: var(--border_width, 1px) solid var(--color_accent_dimmed, rgba(170, 207, 209, 0.4));
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 0.9rem;
  text-align: center;
  outline: none;
}

.lock-input:focus {
  border-color: var(--color_accent, rgb(170, 207, 209));
  box-shadow: 0 0 0 2px var(--color_accent_glow, rgba(170, 207, 209, 0.15));
}

.lock-error {
  margin: 0;
  color: #ff6b6b;
  font-family: var(--font_main, monospace);
  font-size: 0.75rem;
}

.lock-button {
  padding: 0.5em 1.2em;
  background: transparent;
  border: var(--border_width, 1px) solid var(--color_accent, rgb(170, 207, 209));
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 0.8rem;
  letter-spacing: 0.1em;
  cursor: pointer;
}

.lock-button:hover:not(:disabled) {
  background: var(--color_accent_glow, rgba(170, 207, 209, 0.15));
}

.lock-button:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
