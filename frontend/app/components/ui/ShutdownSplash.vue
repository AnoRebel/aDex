<template>
  <Teleport to="body">
    <Transition name="ss-fade" appear>
      <div
        v-if="visible"
        class="ss-screen"
        role="alertdialog"
        aria-live="assertive"
        aria-label="System shutting down"
      >
        <!-- Scanline overlay for the CRT vibe -->
        <div class="ss-scanlines" aria-hidden="true" />

        <!-- Corner brackets — same eDex chrome as the boot splash -->
        <div class="ss-corner ss-corner-tl" aria-hidden="true" />
        <div class="ss-corner ss-corner-tr" aria-hidden="true" />
        <div class="ss-corner ss-corner-bl" aria-hidden="true" />
        <div class="ss-corner ss-corner-br" aria-hidden="true" />

        <div class="ss-content">
          <!-- Glitching headline -->
          <h1 class="ss-title">
            <span
              v-for="(letter, i) in title"
              :key="i"
              class="ss-letter"
              :style="{ animationDelay: `${i * 0.04}s` }"
            >{{ letter === ' ' ? ' ' : letter }}</span>
          </h1>

          <div class="ss-sub">SYSTEM TERMINATION SEQUENCE</div>

          <!-- Reactive step log. Each `step.state` is a reactive `ref`
               so the row's class flips the instant the state changes,
               which is what triggers the per-step cue watcher below. -->
          <ul class="ss-steps" role="list">
            <li
              v-for="step in steps"
              :key="step.id"
              class="ss-step"
              :class="`ss-step-${step.state.value}`"
            >
              <span class="ss-step-tag">[{{ step.tag }}]</span>
              <span class="ss-step-text">{{ step.text }}</span>
              <span class="ss-step-status">{{ statusLabel(step.state.value) }}</span>
            </li>
          </ul>

          <!-- Countdown bar -->
          <div class="ss-bar">
            <div class="ss-bar-fill" :style="{ width: progress + '%' }" />
            <div class="ss-bar-shimmer" />
          </div>

          <div class="ss-counter">
            <span class="ss-counter-label">T-MINUS</span>
            <span class="ss-counter-value">{{ countdownLabel }}</span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
/*
 * ShutdownSplash — full-screen animated overlay shown between the user
 * confirming Quit and Wails actually destroying the window.
 *
 * Reactivity model (mirrors AdexBootScreen):
 *   - Each step has its own reactive `state` ref ('pending' | 'active'
 *     | 'done'). The template binds class to that ref so DOM updates
 *     are a pure consequence of state change — no manual class toggling.
 *   - A `watch` on each step's state fires the per-step audio cue when
 *     the state flips to 'active'. Audio is therefore a reaction to a
 *     reactive change, not a side effect of the timer tick — meaning
 *     it'd also fire correctly if some external code flipped a step
 *     state directly (e.g. a future "skip to next step" debug toggle).
 *   - The timer's only job is advancing state. Audio + visuals come
 *     out of reactive watchers downstream.
 *
 * The parent component mounts this via v-if and listens to `complete`.
 * The parent owns the actual QuitApp() call so this stays a pure
 * presentation component reusable for reboot / sign-out flows later.
 */
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { nowMs } from '~/utils/now'

type StepState = 'pending' | 'active' | 'done'

interface ShutdownStep {
  id: string
  tag: string
  text: string
  /** Soundpack event ID fired when this step's state flips to 'active'. */
  cue: string
  /** Reactive state — drives both the visual class and the audio watcher. */
  state: ReturnType<typeof ref<StepState>>
}

interface Props {
  /** Total time the splash stays up before emitting `complete`. */
  durationMs?: number
  /** Headline text. Each character is rendered as its own animated span. */
  title?: string
  /**
   * Master audio toggle. When false, all cue playback is suppressed at
   * the source — the per-step watchers still fire but `playCue` becomes
   * a no-op. Mirrors AdexBootScreen's `enableAudio` prop so the parent
   * page can wire the same persisted `adex-settings.audio.enabled` flag
   * into both splashes and get consistent behavior.
   */
  enableAudio?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  durationMs: 2400,
  title: 'SHUTTING DOWN',
  enableAudio: true,
})

const emit = defineEmits<{ complete: [] }>()

// Direct cue path through the shared composable — same pipeline the
// boot screen uses, so both ends of the session lifecycle play through
// one Howler-backed engine with consistent volume/mute behavior.
const audio = useAdexAudio()
const visible = ref(true)

// ---- Sequence definition ------------------------------------------------
//
// Five steps that map to what the backend's OnShutdown actually does,
// in roughly the order it does them. The visual cadence is independent
// of the real teardown — backend Shutdown returns when it's done, and
// we don't have an incremental progress channel — but it's not lying,
// just front-running. Each step gets the same airtime.

function makeStep(
  id: string,
  tag: string,
  text: string,
  cue: string,
): ShutdownStep {
  return { id, tag, text, cue, state: ref<StepState>('pending') }
}

const steps: ShutdownStep[] = [
  // Cue names use the canonical EdexCue vocabulary directly so they
  // map 1:1 onto WAV assets / synth recipes without going through
  // the legacy alias table in useAdexAudio.
  makeStep('sessions', 'term',  'Flushing terminal sessions',     'granted'),
  makeStep('services', 'svc',   'Halting background services',    'panels'),
  makeStep('state',    'store', 'Persisting state to disk',       'folder'),
  makeStep('audio',    'audio', 'Closing audio context',          'denied'),
  makeStep('kernel',   'kern',  'Releasing runtime resources',    'alarm'),
]

// ---- Reactive driver state ---------------------------------------------

const startedAt = ref(0)
const progress = ref(0)
const remainingMs = ref(props.durationMs)

const countdownLabel = ref('0.0s')
function fmtCountdown(ms: number) {
  return `${Math.max(0, ms / 1000).toFixed(1)}s`
}

function statusLabel(state: StepState): string {
  switch (state) {
    case 'done':    return 'OK'
    case 'active':  return '...'
    case 'pending':
    default:        return '   '
  }
}

// ---- Audio --------------------------------------------------------------

/**
 * Play a cue via the shared useAdexAudio composable. The composable
 * already honors its own settings.muted flag (driven by the global
 * audio-enabled toggle), so we only need to gate on the parent's
 * enableAudio prop. Wrapped in try/catch because shutdown happens in
 * parallel with backend teardown — playback failures must never
 * block the visual sequence.
 */
function playCue(cue: string) {
  if (!props.enableAudio) return
  try {
    audio.playCue(cue)
  } catch {
    /* swallow — we're shutting down */
  }
}

// ---- Per-step cue watchers ---------------------------------------------
//
// Each step's state ref gets its own watcher. When state flips from
// 'pending' → 'active' we fire the cue. This is the same pattern the
// boot screen uses: audio is a reaction to a reactive state change,
// not bolted onto the timer callback. That means:
//   - A step that re-activates (theoretically possible if the parent
//     reset state) would re-fire its cue.
//   - We don't have to remember to call playCue() in the timer tick.

for (const step of steps) {
  watch(step.state, (now, prev) => {
    if (now === 'active' && prev !== 'active') {
      playCue(step.cue)
    }
  })
}

// ---- Sequence runner ---------------------------------------------------
//
// useIntervalFn (VueUse) ticks every 60ms. Each tick:
//   1) Recomputes elapsed → progress, countdown.
//   2) Figures out which step should currently be 'active' based on
//      elapsed / step-duration.
//   3) Flips state on any step that needs to change — those flips fan
//      out to the watchers above, which fire the cues.
//
// Self-cleans on unmount because we use useIntervalFn, not setInterval.

const stepDurationMs = Math.floor(props.durationMs / steps.length)
const TICK_MS = 60

const { pause: pauseTick, resume: resumeTick } = useIntervalFn(
  () => {
    const elapsed = nowMs() - startedAt.value
    progress.value = Math.min(100, (elapsed / props.durationMs) * 100)
    remainingMs.value = Math.max(0, props.durationMs - elapsed)
    countdownLabel.value = fmtCountdown(remainingMs.value)

    const targetIdx = Math.min(
      steps.length - 1,
      Math.floor(elapsed / stepDurationMs),
    )

    // Mark every prior step done, current one active, future pending.
    // Setting state on a ref that already holds that value is a no-op
    // for Vue reactivity (refs do an Object.is check before notifying)
    // so the per-step watchers only fire on actual transitions.
    for (let i = 0; i < steps.length; i++) {
      const desired: StepState =
        i < targetIdx ? 'done' : i === targetIdx ? 'active' : 'pending'
      if (steps[i].state.value !== desired) {
        steps[i].state.value = desired
      }
    }

    if (elapsed >= props.durationMs) {
      finish()
    }
  },
  TICK_MS,
  { immediate: false },
)

function finish() {
  pauseTick()
  progress.value = 100
  countdownLabel.value = '0.0s'
  // Mark the trailing step done so its row turns green for the final
  // frame the user sees before the window vanishes.
  for (const step of steps) step.state.value = 'done'
  // Final capstone cue — independent of any step. Gives the sequence
  // an audible end-cap rather than just trailing off.
  playCue('alarm')
  setTimeout(() => {
    visible.value = false
    emit('complete')
  }, 400)
}

onMounted(() => {
  // Lead with the opening cue so audio arrives a hair before the
  // title finishes painting in. 'scan' is the long sweep — perfect
  // "system winding down" feel paired with the title glitch.
  playCue('scan')

  // Defer one frame so the first paint has progress=0 — without this
  // the bar would start at ~2% from the very first tick.
  requestAnimationFrame(() => {
    startedAt.value = nowMs()
    resumeTick()
  })
})

onBeforeUnmount(() => {
  pauseTick()
})
</script>

<style scoped>
/* ─── Full-screen overlay ───────────────────────────────────────────── */
.ss-screen {
  position: fixed;
  inset: 0;
  z-index: 99999;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(ellipse at center,
      rgba(239, 68, 68, 0.08) 0%,
      rgba(0, 0, 0, 0.96) 70%);
  color: var(--accent, #aacfd1);
  font-family: var(--font_main, 'JetBrains Mono', 'Fira Code', monospace);
  overflow: hidden;
  animation: ss-vignette 1.5s ease-out;
}

@keyframes ss-vignette {
  from { filter: contrast(0.7) brightness(1.2); }
  to   { filter: contrast(1.05) brightness(1); }
}

/* CRT scanlines */
.ss-scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: repeating-linear-gradient(
    to bottom,
    rgba(255, 255, 255, 0.04) 0px,
    rgba(255, 255, 255, 0.04) 1px,
    transparent 1px,
    transparent 3px
  );
  mix-blend-mode: overlay;
  animation: ss-scan-roll 6s linear infinite;
  opacity: 0.55;
}

@keyframes ss-scan-roll {
  from { background-position: 0 0; }
  to   { background-position: 0 100dvh; }
}

/* Corner brackets — same as boot screen */
.ss-corner {
  position: absolute;
  width: clamp(28px, 5vh, 64px);
  height: clamp(28px, 5vh, 64px);
  border: 0 solid rgba(239, 68, 68, 0.55);
  filter: drop-shadow(0 0 6px rgba(239, 68, 68, 0.6));
}

.ss-corner-tl { top: 3vh; left: 3vw;   border-top-width: 2px; border-left-width: 2px; }
.ss-corner-tr { top: 3vh; right: 3vw;  border-top-width: 2px; border-right-width: 2px; }
.ss-corner-bl { bottom: 3vh; left: 3vw;  border-bottom-width: 2px; border-left-width: 2px; }
.ss-corner-br { bottom: 3vh; right: 3vw; border-bottom-width: 2px; border-right-width: 2px; }

/* ─── Content layout ────────────────────────────────────────────────── */
.ss-content {
  position: relative;
  z-index: 2;
  width: min(90vw, 800px);
  text-align: center;
  padding: 2vh 2vw;
}

/* Glitching headline */
.ss-title {
  font-size: clamp(2rem, 7vh, 5rem);
  letter-spacing: 0.25em;
  margin: 0 0 1.2vh 0;
  color: #ef4444;
  text-shadow:
    0 0 10px rgba(239, 68, 68, 0.85),
    0 0 22px rgba(239, 68, 68, 0.55),
    0 0 40px rgba(239, 68, 68, 0.35);
  position: relative;
}

.ss-letter {
  display: inline-block;
  animation: ss-letter-drop 0.45s ease-out backwards,
             ss-glitch 1.8s steps(2, end) infinite;
}

@keyframes ss-letter-drop {
  from { opacity: 0; transform: translateY(-18px) scaleY(0.6); }
  to   { opacity: 1; transform: translateY(0) scaleY(1); }
}

@keyframes ss-glitch {
  0%, 80%, 100% { transform: translate(0, 0); filter: none; }
  82% { transform: translate(-1px, 1px); filter: hue-rotate(20deg); }
  84% { transform: translate(2px, -1px); filter: hue-rotate(-10deg); }
  86% { transform: translate(0, 0); }
}

.ss-sub {
  font-size: clamp(0.7rem, 1.2vh, 1rem);
  letter-spacing: 0.5em;
  opacity: 0.7;
  margin-bottom: 3vh;
  text-transform: uppercase;
}

/* ─── Step log ──────────────────────────────────────────────────────── */
.ss-steps {
  list-style: none;
  padding: 0;
  margin: 0 auto 2.5vh auto;
  text-align: left;
  width: min(80vw, 540px);
  font-size: clamp(0.8rem, 1.4vh, 1.05rem);
  letter-spacing: 0.08em;
}

.ss-step {
  display: grid;
  grid-template-columns: max-content 1fr max-content;
  gap: 0.8vw;
  padding: 0.4vh 0;
  align-items: baseline;
  transition: opacity 220ms ease, color 220ms ease;
}

.ss-step-tag {
  font-family: var(--font_main);
  text-transform: lowercase;
  opacity: 0.7;
}

.ss-step-status {
  font-family: var(--font_main);
  min-width: 3ch;
  text-align: right;
}

.ss-step-pending { opacity: 0.25; }

.ss-step-active {
  opacity: 1;
  color: #ef4444;
  text-shadow: 0 0 6px rgba(239, 68, 68, 0.55);
  animation: ss-row-flash 0.45s ease-out;
}

.ss-step-active .ss-step-status::after {
  content: '';
  animation: ss-dots 0.6s steps(4, end) infinite;
}

.ss-step-done {
  opacity: 0.65;
  color: #10b981;
}

@keyframes ss-row-flash {
  from { background: rgba(239, 68, 68, 0.18); }
  to   { background: transparent; }
}

@keyframes ss-dots {
  0%, 100% { content: ''; }
  25%      { content: '.'; }
  50%      { content: '..'; }
  75%      { content: '...'; }
}

/* ─── Progress bar ──────────────────────────────────────────────────── */
.ss-bar {
  width: min(80vw, 540px);
  height: 4px;
  margin: 0 auto 1.5vh auto;
  background: rgba(255, 255, 255, 0.06);
  position: relative;
  overflow: hidden;
}

.ss-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #ef4444 0%, #f59e0b 60%, #ef4444 100%);
  transition: width 100ms linear;
  box-shadow: 0 0 12px rgba(239, 68, 68, 0.8);
}

.ss-bar-shimmer {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(255, 255, 255, 0.18) 50%,
    transparent 100%
  );
  animation: ss-shimmer 1.2s linear infinite;
}

@keyframes ss-shimmer {
  from { transform: translateX(-100%); }
  to   { transform: translateX(100%); }
}

/* ─── Countdown ─────────────────────────────────────────────────────── */
.ss-counter {
  font-size: clamp(0.8rem, 1.4vh, 1.05rem);
  letter-spacing: 0.4em;
  display: flex;
  gap: 1vw;
  justify-content: center;
  align-items: baseline;
}

.ss-counter-label { opacity: 0.55; }
.ss-counter-value {
  font-family: var(--font_main);
  color: #ef4444;
  text-shadow: 0 0 8px rgba(239, 68, 68, 0.7);
  min-width: 4ch;
  text-align: left;
}

/* ─── Transitions ───────────────────────────────────────────────────── */
.ss-fade-enter-active { transition: opacity 180ms ease-out; }
.ss-fade-leave-active { transition: opacity 320ms ease-in; }
.ss-fade-enter-from,
.ss-fade-leave-to { opacity: 0; }

/* ─── Responsive ────────────────────────────────────────────────────── */
@media (max-width: 640px) {
  .ss-title { letter-spacing: 0.15em; }
  .ss-sub { letter-spacing: 0.3em; }
  .ss-corner { display: none; }
  .ss-steps { width: 92vw; font-size: 0.85rem; }
}
</style>
