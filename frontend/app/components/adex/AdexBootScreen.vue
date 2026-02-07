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

        <!-- Boot messages -->
        <div class="boot-messages" :class="{ show: messagesVisible }">
          <div
            v-for="(msg, index) in visibleMessages"
            :key="index"
            class="boot-message"
            :class="{ done: msg.done, current: msg.current }"
          >
            <span class="boot-message-prefix">[{{ msg.done ? 'OK' : '..' }}]</span>
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

// ---- Types ----

interface BootMessage {
  text: string
  done: boolean
  current: boolean
  delay: number
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
  { text: 'Initializing aDex-UI v2.0...', done: false, current: false, delay: 200 },
  { text: 'Loading kernel modules...', done: false, current: false, delay: 300 },
  { text: 'Starting system services...', done: false, current: false, delay: 350 },
  { text: 'Initializing terminal subsystem...', done: false, current: false, delay: 280 },
  { text: 'Loading theme engine...', done: false, current: false, delay: 220 },
  { text: 'Connecting to backend services...', done: false, current: false, delay: 400 },
  { text: 'System ready.', done: false, current: false, delay: 150 },
]

const visibleMessages = ref<BootMessage[]>([])

// ---- Audio ----

let audioContext: AudioContext | null = null

function initAudio(): boolean {
  if (!props.enableAudio) return false
  try {
    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
    return true
  } catch {
    return false
  }
}

function playBootTone(frequency: number, duration: number, startTime?: number) {
  if (!audioContext) return
  try {
    const oscillator = audioContext.createOscillator()
    const gainNode = audioContext.createGain()

    oscillator.connect(gainNode)
    gainNode.connect(audioContext.destination)

    oscillator.type = 'sine'
    oscillator.frequency.setValueAtTime(frequency, audioContext.currentTime)

    const start = startTime || audioContext.currentTime
    gainNode.gain.setValueAtTime(0, start)
    gainNode.gain.linearRampToValueAtTime(0.05, start + 0.01)
    gainNode.gain.exponentialRampToValueAtTime(0.001, start + duration)

    oscillator.start(start)
    oscillator.stop(start + duration)
  } catch {
    // Audio playback failed silently
  }
}

function playBootSound() {
  if (!audioContext) return
  // Short ascending tones for boot sequence feel
  playBootTone(220, 0.08, audioContext.currentTime)
  playBootTone(330, 0.08, audioContext.currentTime + 0.06)
  playBootTone(440, 0.12, audioContext.currentTime + 0.12)
}

function playMessageSound() {
  if (!audioContext) return
  playBootTone(800, 0.03)
}

function playCompleteSound() {
  if (!audioContext) return
  const now = audioContext.currentTime
  playBootTone(440, 0.1, now)
  playBootTone(554, 0.1, now + 0.08)
  playBootTone(659, 0.15, now + 0.16)
}

// ---- Helpers ----

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

// ---- Boot Sequence ----

let bootAborted = false

async function runBootSequence(): Promise<void> {
  const startTime = Date.now()

  // Initialize audio
  initAudio()

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

  // Phase 2: Boot messages appear one by one
  messagesVisible.value = true
  progressVisible.value = true

  const totalMessages = bootMessages.length
  for (let i = 0; i < totalMessages; i++) {
    if (bootAborted) return

    const msg = { ...bootMessages[i], current: true }
    visibleMessages.value.push(msg)

    playMessageSound()

    // Update progress
    const baseProgress = Math.round(((i + 0.5) / totalMessages) * 90)
    progressPercent.value = baseProgress
    emit('progress', baseProgress)

    await sleep(msg.delay)
    if (bootAborted) return

    // Mark as done
    visibleMessages.value[i].done = true
    visibleMessages.value[i].current = false

    const doneProgress = Math.round(((i + 1) / totalMessages) * 90)
    progressPercent.value = doneProgress
    emit('progress', doneProgress)

    // Brief pause between messages
    await sleep(80)
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
  visible.value = false

  // Cleanup audio
  if (audioContext) {
    try {
      audioContext.close()
    } catch {
      // Ignore
    }
    audioContext = null
  }

  emit('complete')
}

// ---- Lifecycle ----

onMounted(() => {
  runBootSequence()
})

onBeforeUnmount(() => {
  bootAborted = true
  if (audioContext) {
    try {
      audioContext.close()
    } catch {
      // Ignore
    }
    audioContext = null
  }
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

.boot-message-prefix {
  flex-shrink: 0;
  width: 3vw;
  text-align: right;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
}

.boot-message.done .boot-message-prefix {
  color: var(--success, #10b981);
}

.boot-message.current .boot-message-prefix {
  color: var(--color_accent, rgb(170, 207, 209));
  animation: boot-blink 0.6s steps(1) infinite;
}

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
