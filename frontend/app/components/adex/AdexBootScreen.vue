<template>
  <Transition name="boot-overlay">
    <div v-if="visible" class="boot-overlay" @click.prevent>
      <!-- Scanline effect -->
      <div class="boot-scanline" />

      <!-- Boot content -->
      <div class="boot-content">
        <!-- Title with glitch effect -->
        <div class="boot-title-wrapper" :class="{ show: titleVisible }">
          <h1
            class="boot-title"
            :class="{ glitching: glitchActive }"
            data-text="aDex-UI"
          >
            aDex-UI
          </h1>
          <div class="boot-subtitle" :class="{ show: subtitleVisible }">
            ADVANCED DESKTOP EXPERIENCE
          </div>
        </div>

        <!-- Boot messages — signale-style: level icon + [tag] + text. -->
        <div class="boot-messages" :class="{ show: messagesVisible }">
          <div
            v-for="(msg, index) in visibleMessages"
            :key="index"
            class="boot-message"
            :class="[`boot-message-${msg.level}`, { done: msg.done, current: msg.current }]"
          >
            <span class="boot-message-icon">{{ levelIcon(msg.level) }}</span>
            <span class="boot-message-tag">[{{ msg.tag }}]</span>
            <span class="boot-message-text">{{ msg.text }}</span>
          </div>
        </div>

        <!-- Progress bar -->
        <div class="boot-progress-wrapper" :class="{ show: progressVisible }">
          <div class="boot-progress-track">
            <div
              class="boot-progress-fill"
              :style="{ width: progressPercent + '%' }"
            />
          </div>
          <div class="boot-progress-label">
            {{ progressPercent }}%
          </div>
        </div>

        <!-- Version info -->
        <div class="boot-version" :class="{ show: versionVisible }">
          v2.0.0
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { onKeyStroke } from '@vueuse/core'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { Events } from '~/lib/wailsjs/runtime'

// ---- Types ----

/**
 * BootMessage drives a single signale-style log line:
 *
 *   ⏵ [config]      Settings loaded from ~/.config/aDex-UI
 *   ⠿ [theme]       Loading theme engine...
 *   ✔ [terminal]    PTY service ready
 *   ✗ [filesystem]  fsnotify watcher init failed
 *
 * `level` controls icon + color (matches signale's conventions); `tag`
 * is the lowercase service name shown in [brackets].
 */
type BootLevel = 'info' | 'wait' | 'success' | 'warn' | 'error'

interface BootMessage {
  level: BootLevel
  tag: string
  text: string
  done: boolean
  current: boolean
  delay: number
}

const BOOT_LEVEL_ICON: Record<BootLevel, string> = {
  info:    '›',
  wait:    '⠿',
  success: '✔',
  warn:    '⚠',
  error:   '✗',
}

// ---- Props / Emits ----

const props = withDefaults(
  defineProps<{
    minDuration?: number
    enableAudio?: boolean
  }>(),
  {
    minDuration: 2500,
    enableAudio: false,
  }
)

const emit = defineEmits<{
  (e: 'complete'): void
  (e: 'progress', percent: number): void
}>()

// ---- State ----

const visible = ref(true)
const titleVisible = ref(false)
const subtitleVisible = ref(false)
const messagesVisible = ref(false)
const progressVisible = ref(false)
const versionVisible = ref(false)
const glitchActive = ref(false)
const progressPercent = ref(0)

const bootMessages: BootMessage[] = [
  { level: 'wait',    tag: 'config',    text: 'Loading settings...',                       done: false, current: false, delay: 180 },
  { level: 'wait',    tag: 'theme',     text: 'Initializing theme engine...',              done: false, current: false, delay: 220 },
  { level: 'wait',    tag: 'keyboard',  text: 'Loading keyboard layout...',                done: false, current: false, delay: 160 },
  { level: 'wait',    tag: 'audio',     text: 'Initializing audio cue system...',          done: false, current: false, delay: 200 },
  { level: 'wait',    tag: 'system',    text: 'Probing CPU / memory / processes...',       done: false, current: false, delay: 240 },
  { level: 'wait',    tag: 'network',   text: 'Resolving network interfaces...',           done: false, current: false, delay: 280 },
  { level: 'wait',    tag: 'filesystem',text: 'Mounting filesystem service...',            done: false, current: false, delay: 200 },
  { level: 'wait',    tag: 'terminal',  text: 'Spawning PTY service...',                   done: false, current: false, delay: 260 },
  { level: 'success', tag: 'kernel',    text: 'aDex-UI ready — handing off to UI',         done: false, current: false, delay: 160 },
]

const visibleMessages = ref<BootMessage[]>([])

function levelIcon(level: BootLevel): string {
  return BOOT_LEVEL_ICON[level] ?? '›'
}

// ---- Audio ----
//
// All cues route through the shared useAdexAudio composable so the boot
// screen plays through the same Howler-backed pipeline as every other
// surface in the app (settings modal preview, terminal bell, shutdown
// splash, etc.). That composable already handles:
//   - AudioContext creation + autoplay-policy resume
//   - Howler unlock-listener arming (with WeakSet dedup so we don't
//     leak callbacks across thousands of plays)
//   - Synth fallback when WAV assets fail to load
//   - Explicit node disconnect in `onended` for the synth path
//   - Per-cue volume + global volume scaling
//
// The previous hand-rolled AudioContext + oscillator code duplicated
// most of that logic with its own bugs (no `onended` cleanup, no
// suspend retry, no shared volume preferences).

const audio = useAdexAudio()

function playCue(cue: string) {
  if (!props.enableAudio) return
  try {
    audio.playCue(cue)
  } catch {
    /* swallow — audio is decorative, never block boot */
  }
}

function playBootSound() {
  // 'scan' is the long sweep cue — perfect "system coming online" feel.
  playCue('scan')
}

function playMessageSound() {
  // 'stdin' is the short blip used for terminal stream activity; it
  // doubles nicely as a per-step boot tick (~30ms).
  playCue('stdin')
}

function playCompleteSound() {
  // 'granted' is the rising "access granted" chord — fits the
  // "boot complete, system ready" beat.
  playCue('granted')
}

// ---- Helpers ----

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

// ---- Boot Sequence ----

let bootAborted = false
let bootUnsubscribe: (() => void) | null = null
// Single-handoff latch. Both the Esc skip path and the natural
// sequence-completion path check + set this so `emit('complete')`
// fires exactly once even if a skip races the timed completion.
// Declared up here (not next to skipBoot) so runBootSequence's
// closure references a fully-initialized binding, not one in the
// temporal dead zone.
let handedOff = false

// Map a backend boot.stage payload onto our existing BootMessage shape so
// the template doesn't have to special-case backend-vs-fallback rows.
function backendToBootMessage(payload: {
  stage: 'wait' | 'success' | 'warn' | 'error'
  tag: string
  text: string
  detail?: string
}): BootMessage {
  const lvlMap: Record<typeof payload.stage, BootLevel> = {
    wait: 'wait',
    success: 'success',
    warn: 'warn',
    error: 'error',
  }
  return {
    level: lvlMap[payload.stage],
    tag: payload.tag,
    text: payload.detail ? `${payload.text} — ${payload.detail}` : payload.text,
    done: payload.stage !== 'wait',
    current: payload.stage === 'wait',
    delay: 0,
  }
}

/** Subscribe to backend boot events. Returns an unsubscribe fn or null
 *  when the Wails runtime isn't available (Storybook, browser preview,
 *  jsdom test) — callers fall back to the canned message list.
 *
 *  Wails v3 has no `window.runtime` global; events come from the runtime
 *  module, and `Events.On` returns the unsubscribe function directly
 *  (v2 required removing listeners by event name). The null-return
 *  contract is kept so non-desktop contexts still degrade to the canned
 *  sequence. */
function subscribeBootEvents(onComplete: () => void): (() => void) | null {
  if (typeof window === 'undefined') return null

  const offStage = Events.On('boot.stage', (raw: unknown) => {
    const payload = raw as {
      stage: 'wait' | 'success' | 'warn' | 'error'
      tag: string
      text: string
      detail?: string
    }
    if (!payload || typeof payload.tag !== 'string') return

    // If we have an existing wait row for this tag, upgrade it in place
    // so we don't spam two lines per service.
    const existing = visibleMessages.value.find(
      (m) => m.tag === payload.tag && m.current,
    )
    if (existing) {
      existing.level = payload.stage === 'wait' ? 'wait' : (payload.stage === 'success' ? 'success' : payload.stage === 'warn' ? 'warn' : 'error')
      existing.done = payload.stage !== 'wait'
      existing.current = payload.stage === 'wait'
      if (payload.detail) existing.text = `${payload.text} — ${payload.detail}`
    } else {
      visibleMessages.value.push(backendToBootMessage(payload))
    }
    playMessageSound()
  })

  const offComplete = Events.On('boot.complete', () => {
    progressPercent.value = 100
    emit('progress', 100)
    onComplete()
  })

  return () => {
    try { offStage?.() } catch { /* ignore */ }
    try { offComplete?.() } catch { /* ignore */ }
  }
}

async function runBootSequence(): Promise<void> {
  const startTime = Date.now()

  // No explicit audio init — useAdexAudio lazy-creates its
  // AudioContext on the first playCue() call and self-handles the
  // autoplay-policy resume + Howler unlock pattern.

  // Phase 1: Title appears with glitch
  await sleep(200)
  if (bootAborted) return

  playBootSound()
  titleVisible.value = true
  glitchActive.value = true

  await sleep(600)
  if (bootAborted) return

  glitchActive.value = false
  subtitleVisible.value = true

  await sleep(400)
  if (bootAborted) return

  // Phase 2: Boot messages — driven by backend `boot.stage` events when
  // Wails runtime is available, fallback to canned timer-based reveal
  // when running in a plain browser preview / jsdom test.
  messagesVisible.value = true
  progressVisible.value = true

  let backendComplete = false
  const completed = new Promise<void>((resolve) => {
    const unsubscribe = subscribeBootEvents(() => {
      backendComplete = true
      resolve()
    })
    if (unsubscribe) bootUnsubscribe = unsubscribe
    // No runtime → resolve immediately so we fall through to the canned
    // sequence below.
    if (!unsubscribe) resolve()
  })

  // Wait at most 10s for backend boot.complete; if it doesn't arrive we
  // assume something is wrong and proceed with whatever we've got.
  await Promise.race([completed, sleep(10_000)])

  // Canned fallback when no backend events ever arrived (wails dev not
  // running, jsdom test, etc). Reveals the static list with the same
  // pacing as the V1 sequence.
  if (!backendComplete && visibleMessages.value.length === 0) {
    const totalMessages = bootMessages.length
    for (let i = 0; i < totalMessages; i++) {
      if (bootAborted) return

      const msg = { ...bootMessages[i], current: true }
      visibleMessages.value.push(msg)

      playMessageSound()

      const baseProgress = Math.round(((i + 0.5) / totalMessages) * 90)
      progressPercent.value = baseProgress
      emit('progress', baseProgress)

      await sleep(msg.delay)
      if (bootAborted) return

      visibleMessages.value[i].done = true
      visibleMessages.value[i].current = false

      const doneProgress = Math.round(((i + 1) / totalMessages) * 90)
      progressPercent.value = doneProgress
      emit('progress', doneProgress)

      // Brief pause between messages
      await sleep(80)
    }
  }

  if (bootAborted) return

  // Phase 3: Progress to 100%
  progressPercent.value = 95
  emit('progress', 95)
  await sleep(200)
  if (bootAborted) return

  progressPercent.value = 100
  emit('progress', 100)
  versionVisible.value = true

  // Glitch flash on complete
  glitchActive.value = true
  await sleep(150)
  glitchActive.value = false

  playCompleteSound()

  // Ensure minimum duration
  const elapsed = Date.now() - startTime
  const remaining = props.minDuration - elapsed
  if (remaining > 0) {
    await sleep(remaining)
  }

  if (bootAborted) return

  // Phase 4: Fade out
  await sleep(300)
  if (handedOff) return // user pressed Esc during the fade
  visible.value = false

  // No audio cleanup needed — useAdexAudio owns the AudioContext at
  // module scope and reuses it across mounts. Closing it here would
  // break any subsequent cue from any other surface.

  handedOff = true
  emit('complete')
}

// ---- Skip ----

function skipBoot() {
  if (handedOff) return
  handedOff = true
  // Halt the in-flight sequence's remaining awaits — every phase
  // checks `bootAborted` after each sleep, so this short-circuits
  // them. Then immediately hide + hand off.
  bootAborted = true
  visible.value = false
  emit('complete')
}

// Esc skips the boot intro for this launch only. (The permanent
// "never show the intro" preference is the Settings → Advanced →
// Skip boot intro toggle, read by pages/index.vue. This is the
// one-shot escape hatch — useful in dev when restarting often.)
// onKeyStroke auto-cleans on unmount, no manual listener teardown.
onKeyStroke('Escape', (e) => {
  e.preventDefault()
  skipBoot()
})

// Guard the natural-completion path too: runBootSequence's final
// `emit('complete')` must respect the same handedOff latch so a
// skip mid-fade-out doesn't race the timed completion.
watch(visible, (v) => {
  if (!v) handedOff = true
})

// ---- Lifecycle ----

onMounted(() => {
  runBootSequence()
})

onBeforeUnmount(() => {
  bootAborted = true
  if (bootUnsubscribe) {
    try { bootUnsubscribe() } catch { /* ignore */ }
    bootUnsubscribe = null
  }
  // Intentionally don't close the audio context here — useAdexAudio
  // owns it at module scope so it survives boot → main UI handoff
  // and stays usable for terminal bells, settings previews, etc.
})
</script>

<style scoped>
.boot-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: var(--color_black, #000000);
  z-index: 99999;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  user-select: none;
}

/* Scanline overlay */
.boot-scanline {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 0.2vh,
    rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.015) 0.2vh,
    rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.015) 0.4vh
  );
}

/* Content container */
.boot-content {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2vh;
  width: 60vw;
  max-width: 50vw;
}

/* Title */
.boot-title-wrapper {
  opacity: 0;
  transform: translateY(-1vh);
  transition: opacity 0.4s ease-out, transform 0.4s ease-out;
  text-align: center;
}

.boot-title-wrapper.show {
  opacity: 1;
  transform: translateY(0);
}

.boot-title {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 6vh;
  font-weight: 700;
  color: var(--color_accent, rgb(170, 207, 209));
  letter-spacing: 0.8vw;
  text-transform: uppercase;
  position: relative;
  display: inline-block;
  line-height: 1.2;
}

/* Glitch pseudo-elements */
.boot-title::before,
.boot-title::after {
  content: attr(data-text);
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.boot-title::before {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.7);
  z-index: -1;
}

.boot-title::after {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
  z-index: -2;
}

.boot-title.glitching::before {
  animation: boot-glitch-top 0.15s linear infinite;
}

.boot-title.glitching::after {
  animation: boot-glitch-bottom 0.15s linear infinite;
}

@keyframes boot-glitch-top {
  0% {
    clip-path: inset(0 0 60% 0);
    transform: translate(-0.3vw, -0.2vh);
  }
  25% {
    clip-path: inset(20% 0 40% 0);
    transform: translate(0.4vw, 0.1vh);
  }
  50% {
    clip-path: inset(40% 0 20% 0);
    transform: translate(-0.2vw, 0.3vh);
  }
  75% {
    clip-path: inset(10% 0 50% 0);
    transform: translate(0.3vw, -0.1vh);
  }
  100% {
    clip-path: inset(0 0 60% 0);
    transform: translate(-0.1vw, 0.2vh);
  }
}

@keyframes boot-glitch-bottom {
  0% {
    clip-path: inset(60% 0 0 0);
    transform: translate(0.3vw, 0.2vh);
  }
  25% {
    clip-path: inset(40% 0 0 0);
    transform: translate(-0.4vw, -0.1vh);
  }
  50% {
    clip-path: inset(50% 0 0 0);
    transform: translate(0.2vw, -0.3vh);
  }
  75% {
    clip-path: inset(30% 0 0 0);
    transform: translate(-0.3vw, 0.1vh);
  }
  100% {
    clip-path: inset(60% 0 0 0);
    transform: translate(0.1vw, -0.2vh);
  }
}

/* Subtitle */
.boot-subtitle {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.4vh;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
  letter-spacing: 0.6vw;
  text-transform: uppercase;
  margin-top: 0.8vh;
  opacity: 0;
  transition: opacity 0.5s ease-out;
}

.boot-subtitle.show {
  opacity: 1;
}

/* Boot messages */
.boot-messages {
  width: 100%;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.2vh;
  line-height: 1.8;
  opacity: 0;
  transition: opacity 0.3s ease-out;
  max-height: 30vh;
  overflow: hidden;
}

.boot-messages.show {
  opacity: 1;
}

.boot-message {
  display: flex;
  align-items: baseline;
  gap: 0.6vw;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.6);
  animation: boot-msg-appear 0.15s ease-out forwards;
}

.boot-message.done {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.8);
}

.boot-message.current {
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Signale-style icon in a fixed-width gutter so tags + text align across
 * all rows regardless of the level glyph width. */
.boot-message-icon {
  flex-shrink: 0;
  width: 1.4vw;
  text-align: center;
  font-weight: bold;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.55);
}

.boot-message-tag {
  flex-shrink: 0;
  min-width: 7vw;
  text-align: right;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.45);
  letter-spacing: 0.08em;
}

.boot-message.current .boot-message-icon {
  color: var(--color_accent, rgb(170, 207, 209));
  animation: boot-blink 0.6s steps(1) infinite;
}

.boot-message-success .boot-message-icon { color: var(--success, var(--ok, #10b981)); }
.boot-message-warn    .boot-message-icon { color: var(--warning, var(--warn, #f59e0b)); }
.boot-message-error   .boot-message-icon { color: var(--error, var(--err, #ef4444)); }
.boot-message-error   .boot-message-text { color: var(--error, var(--err, #ef4444)); }

@keyframes boot-msg-appear {
  from {
    opacity: 0;
    transform: translateX(-0.5vw);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes boot-blink {
  0%, 49% {
    opacity: 1;
  }
  50%, 100% {
    opacity: 0.3;
  }
}

/* Progress bar */
.boot-progress-wrapper {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 1vw;
  opacity: 0;
  transition: opacity 0.3s ease-out;
}

.boot-progress-wrapper.show {
  opacity: 1;
}

.boot-progress-track {
  flex: 1;
  height: 0.4vh;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.1);
  position: relative;
  overflow: hidden;
}

.boot-progress-fill {
  height: 100%;
  background: var(--color_accent, rgb(170, 207, 209));
  transition: width 0.2s ease-out;
  position: relative;
}

.boot-progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 2vw;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4)
  );
}

.boot-progress-label {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  color: var(--color_accent, rgb(170, 207, 209));
  font-variant-numeric: tabular-nums;
  min-width: 3vw;
  text-align: right;
}

/* Version */
.boot-version {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1vh;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  letter-spacing: 0.2vw;
  opacity: 0;
  transition: opacity 0.4s ease-out;
}

.boot-version.show {
  opacity: 1;
}

/* Overlay transition */
.boot-overlay-leave-active {
  transition: opacity 0.5s ease-out;
}

.boot-overlay-leave-to {
  opacity: 0;
}
</style>
