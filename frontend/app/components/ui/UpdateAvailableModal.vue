<template>
  <Teleport to="body">
    <Transition name="ua-fade" appear>
      <div
        v-if="modelValue && release"
        class="ua-overlay"
        role="dialog"
        :aria-labelledby="`ua-title-${id}`"
        aria-modal="true"
        :data-phase="view"
        @click.self="onLater"
        @keydown.esc="onLater"
      >
        <div ref="cardRef" class="ua-card mod-panel">
          <div class="ua-header">
            <div :id="`ua-title-${id}`" class="ua-title">
              <span class="ua-bracket">[</span>
              {{ title }}
              <span class="ua-bracket">]</span>
            </div>
            <button
              class="ua-close"
              type="button"
              :aria-label="t('updates.closeLabel')"
              @click="onLater"
            >X</button>
          </div>

          <div class="ua-meta">
            <div class="ua-version-row">
              <span class="ua-version-label">{{ t('updates.current') }}</span>
              <span class="ua-version-value">{{ currentVersion }}</span>
            </div>
            <div class="ua-version-row">
              <span class="ua-version-label">{{ t('updates.latest') }}</span>
              <span class="ua-version-value ua-version-new">{{ release.version }}</span>
            </div>
            <div v-if="releasePublishedDisplay" class="ua-version-row">
              <span class="ua-version-label">{{ t('updates.published') }}</span>
              <span class="ua-version-value">{{ releasePublishedDisplay }}</span>
            </div>
          </div>

          <template v-if="view === 'available'">
            <div v-if="release.name && release.name !== release.version" class="ua-name">
              {{ release.name }}
            </div>
            <!-- Monospace notes suit the eDex look and avoid a markdown lib. -->
            <div class="ua-notes-label">{{ t('updates.releaseNotes') }}</div>
            <pre class="ua-notes" tabindex="0">{{ releaseBody || t('updates.noNotes') }}</pre>
            <p v-if="!canSelfUpdate" class="ua-status ua-hint" data-testid="ua-notify-only">
              {{ selfUpdateReason ? t('updates.cannotSelfUpdateReason', { reason: selfUpdateReason }) : t('updates.cannotSelfUpdate') }}
              {{ t('updates.downloadFromPage') }}
            </p>
          </template>

          <div v-else-if="view === 'downloading'" class="ua-progress-block">
            <div class="ua-status">{{ t('updates.downloading') }}</div>
            <div
              class="ua-progress"
              role="progressbar"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-valuenow="percent ?? undefined"
            >
              <!-- No size yet (or none sent): a sweeping bar, so it never
                   sits at 0% looking stuck. -->
              <div
                class="ua-progress-fill"
                :class="{ 'ua-indeterminate': percent === null }"
                :style="percent === null ? undefined : { width: `${percent}%` }"
              />
            </div>
            <div class="ua-progress-text">{{ progressText }}</div>
          </div>

          <div
            v-else-if="view === 'verifying' || view === 'installing'"
            class="ua-status ua-busy"
            role="status"
            aria-live="polite"
          >
            <span class="ua-spinner" aria-hidden="true" />
            {{ view === 'verifying' ? t('updates.verifying') : t('updates.installing') }}
          </div>

          <div v-else-if="view === 'ready'" class="ua-status ua-ok">
            {{ t('updates.installed', { version: release.version }) }}
          </div>

          <div v-else-if="view === 'error'" class="ua-status ua-error" role="alert">
            <span class="ua-version-label">{{ t('updates.stageFailed', { stage: stageLabel }) }}</span>
            <span>{{ error?.message }}</span>
          </div>

          <div class="ua-actions">
            <template v-if="view === 'available'">
              <button class="ua-btn ua-btn-secondary" type="button" @click="onSkip">{{ t('updates.skip') }}</button>
              <button class="ua-btn ua-btn-secondary" type="button" @click="onLater">{{ t('updates.remindLater') }}</button>
              <button v-if="canSelfUpdate" class="ua-btn ua-btn-primary" type="button" @click="onInstall">{{ t('updates.install') }}</button>
              <button v-else class="ua-btn ua-btn-primary" type="button" @click="onOpenPage">{{ t('updates.openPage') }}</button>
            </template>
            <template v-else-if="view === 'ready'">
              <button class="ua-btn ua-btn-secondary" type="button" @click="onLater">{{ t('updates.later') }}</button>
              <button class="ua-btn ua-btn-primary" type="button" @click="onRestart">{{ t('updates.restartApply') }}</button>
            </template>
            <template v-else-if="view === 'error'">
              <button class="ua-btn ua-btn-secondary" type="button" @click="onLater">{{ t('updates.close') }}</button>
              <button class="ua-btn ua-btn-primary" type="button" @click="onRetry">{{ t('updates.retry') }}</button>
            </template>
            <button v-else class="ua-btn ua-btn-secondary" type="button" @click="onLater">{{ t('updates.hide') }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * UpdateAvailableModal
 *
 * Presents the update flow driven by useUpdateChecker:
 *   available → downloading → verifying → installing → ready, or error.
 * Installs that can't replace themselves only get OPEN RELEASE PAGE.
 * Closing never cancels anything: a running download carries on, and a
 * ready update applies only when the user chooses RESTART & APPLY.
 */
import { computed, nextTick, ref, useId, watch } from 'vue'
import { format as formatDateFn, isValid as isValidDate } from 'date-fns'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { useUpdateChecker } from '~/composables/useUpdateChecker'
import { useI18n } from '~/composables/useI18n'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const id = useId()
const updates = useUpdateChecker()
const { release, phase, progress, error, currentVersion, canSelfUpdate } = updates
const audio = useAdexAudio()
const { t } = useI18n()

type View = 'available' | 'downloading' | 'verifying' | 'installing' | 'ready' | 'error'

const view = computed<View>(() => {
  switch (phase.value) {
    case 'downloading':
    case 'verifying':
    case 'installing':
    case 'ready':
    case 'error':
      return phase.value
    default:
      return 'available'
  }
})

const title = computed(() => t({
  available: 'updates.titleAvailable',
  downloading: 'updates.titleUpdating',
  verifying: 'updates.titleUpdating',
  installing: 'updates.titleUpdating',
  ready: 'updates.titleReady',
  error: 'updates.titleFailed',
}[view.value]))

const KNOWN_STAGES = ['check', 'download', 'verify', 'install', 'update']
/** Known stages are translated; anything else is shown upper-cased as sent. */
const stageLabel = computed(() => {
  const stage = error.value?.stage || 'update'
  return KNOWN_STAGES.includes(stage) ? t(`updates.stages.${stage}`) : stage.toUpperCase()
})

// Move focus into the dialog on open and on each step, to its main action
// (or the close button while nothing can be pressed), so Enter and Esc act
// on it rather than on whatever had focus behind the overlay.
const cardRef = ref<HTMLElement | null>(null)
watch([() => props.modelValue, view], async ([open]) => {
  if (!open) return
  await nextTick()
  const card = cardRef.value
  ;(card?.querySelector<HTMLElement>('.ua-btn-primary') ?? card?.querySelector<HTMLElement>('.ua-close'))?.focus()
}, { immediate: true })

const selfUpdateReason = computed(() => updates.info.value?.selfUpdateReason ?? '')

const releasePublishedDisplay = computed<string>(() => {
  const iso = release.value?.publishedAt
  if (!iso) return ''
  const d = new Date(iso)
  // Go's zero time marshals as year 1; treat it as "unknown".
  if (!isValidDate(d) || d.getFullYear() < 2000) return ''
  return formatDateFn(d, 'yyyy-MM-dd HH:mm')
})

const releaseBody = computed<string>(() => (release.value?.notes ?? '').trim())

const percent = computed<number | null>(() => {
  const p = progress.value
  if (!p || p.total <= 0) return null
  return Math.min(100, Math.round((p.written / p.total) * 100))
})

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 ** 2) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / 1024 ** 2).toFixed(1)} MB`
}

const progressText = computed<string>(() => {
  const p = progress.value
  if (!p) return t('updates.starting')
  const done = formatBytes(p.written)
  const total = formatBytes(p.total)
  const rate = formatBytes(p.rate)
  if (p.total > 0) {
    return p.rate > 0
      ? t('updates.progressRate', { done, total, rate })
      : t('updates.progress', { done, total })
  }
  return p.rate > 0 ? t('updates.progressUnknownRate', { done, rate }) : done
})

function cue(name: 'click' | 'destructive') {
  try { audio.playCue(name) } catch { /* non-fatal */ }
}

function close() {
  emit('update:modelValue', false)
}

function onLater() {
  cue('click')
  close()
}

function onSkip() {
  // Commit-and-forget until the skip list is cleared in Settings.
  cue('destructive')
  if (release.value) updates.skipVersion(release.value.version)
  close()
}

function onInstall() {
  cue('click')
  void updates.install()
}

function onOpenPage() {
  cue('click')
  updates.openReleasePage()
}

function onRestart() {
  cue('click')
  void updates.restart()
}

function onRetry() {
  cue('click')
  void (error.value?.stage === 'check' ? updates.check() : updates.install())
}
</script>

<style scoped>
.ua-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(2px);
  padding: clamp(0.5rem, 2vh, 2rem) clamp(0.5rem, 2vw, 2rem);
}

.ua-card {
  --corner-cut: 1.2vh;
  width: min(640px, 92vw);
  max-height: 86dvh;
  display: flex;
  flex-direction: column;
  padding: 1.6vh 1.6vw;
  background: var(--bg-glass, rgba(5, 8, 13, 0.92));
  box-shadow: var(--glow-strong, 0 0 30px rgba(0, 255, 255, 0.2));
  overflow: hidden;
}

.ua-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.2vh;
}

.ua-title {
  font-size: clamp(0.95rem, 1.5vh, 1.2rem);
  letter-spacing: 0.2em;
  color: var(--accent, var(--color_accent, #aacfd1));
}

.ua-bracket { opacity: 0.55; }

.ua-close {
  background: transparent;
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color, rgba(170, 207, 209, 0.3)));
  color: var(--accent, var(--color_accent, #aacfd1));
  padding: 0.4vh 1vw;
  cursor: pointer;
  font-family: inherit;
  font-size: clamp(0.7rem, 1.05vh, 1rem);
  letter-spacing: 0.2em;
}

.ua-close:hover {
  background: var(--surface-2, rgba(255, 255, 255, 0.06));
}

.ua-meta {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.4vh 1.2vw;
  margin-bottom: 1vh;
  font-size: clamp(0.7rem, 1.05vh, 0.9rem);
  letter-spacing: 0.1em;
}

.ua-version-row {
  display: contents;
}

.ua-version-label {
  opacity: 0.55;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  font-size: 0.85em;
  align-self: center;
}

.ua-version-value {
  font-family: var(--font_main, 'Fira Code', monospace);
  color: var(--accent, var(--color_accent, #aacfd1));
}

.ua-version-new {
  color: var(--ok, #10b981);
}


.ua-name {
  font-size: clamp(0.85rem, 1.3vh, 1.05rem);
  letter-spacing: 0.05em;
  margin-bottom: 1vh;
  opacity: 0.9;
}

.ua-notes-label {
  font-size: clamp(0.65rem, 0.95vh, 0.8rem);
  letter-spacing: 0.2em;
  opacity: 0.5;
  text-transform: uppercase;
  margin-bottom: 0.4vh;
}

.ua-notes {
  /* Sized by its content; shrinks (and scrolls) only once the card reaches
     its max height. A zero basis collapsed it to a clipped sliver. */
  flex: 0 1 auto;
  min-height: 4.5em;
  margin: 0 0 1.4vh 0;
  padding: 1vh 1vw;
  background: rgba(0, 0, 0, 0.4);
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color, rgba(170, 207, 209, 0.2)));
  color: var(--accent, var(--color_accent, #aacfd1));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: clamp(0.7rem, 1vh, 0.9rem);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-y: auto;
  scrollbar-width: thin;
}

.ua-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5vw;
  justify-content: flex-end;
  flex: 0 0 auto;
}

.ua-btn {
  padding: 0.7vh 1.4vw;
  background: transparent;
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color, rgba(170, 207, 209, 0.3)));
  color: var(--accent, var(--color_accent, #aacfd1));
  font-family: inherit;
  font-size: clamp(0.7rem, 1.05vh, 0.95rem);
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 120ms ease, border-color 120ms ease;
}

.ua-btn:hover:not(:disabled) {
  background: var(--surface-2, rgba(170, 207, 209, 0.08));
  border-color: var(--accent, var(--color_accent, #aacfd1));
}

.ua-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.ua-btn-primary {
  background: rgba(16, 185, 129, 0.15);
  border-color: var(--ok, #10b981);
  color: var(--ok, #10b981);
}

.ua-btn-primary:hover:not(:disabled) {
  background: rgba(16, 185, 129, 0.3);
}

.ua-status {
  display: flex;
  flex-direction: column;
  gap: 0.4vh;
  margin: 1vh 0 1.4vh;
  font-size: clamp(0.75rem, 1.1vh, 0.95rem);
  letter-spacing: 0.1em;
  color: var(--accent, var(--color_accent, #aacfd1));
}

.ua-hint { opacity: 0.75; margin-top: 0; }
.ua-ok { color: var(--ok, #10b981); }
.ua-error { color: var(--danger, #ef4444); }

.ua-progress-block { margin: 1vh 0 1.4vh; }

.ua-progress {
  height: 0.8vh;
  min-height: 6px;
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color, rgba(170, 207, 209, 0.3)));
  background: rgba(0, 0, 0, 0.4);
  overflow: hidden;
}

.ua-progress-fill {
  height: 100%;
  background: var(--accent, var(--color_accent, #aacfd1));
  transition: width 120ms linear;
}

.ua-indeterminate {
  width: 30%;
  animation: ua-sweep 1.2s ease-in-out infinite;
}

@keyframes ua-sweep {
  from { transform: translateX(-100%); }
  to { transform: translateX(333%); }
}

.ua-busy {
  flex-direction: row;
  align-items: center;
  gap: 0.8em;
}

.ua-spinner {
  width: 1em;
  height: 1em;
  flex: none;
  border: 2px solid currentColor;
  border-right-color: transparent;
  animation: ua-spin 0.8s linear infinite;
}

@keyframes ua-spin { to { transform: rotate(360deg); } }

@media (prefers-reduced-motion: reduce) {
  .ua-indeterminate, .ua-spinner { animation-duration: 3s; }
}

.ua-progress-text {
  margin-top: 0.5vh;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: clamp(0.65rem, 0.95vh, 0.85rem);
  opacity: 0.75;
}

.ua-fade-enter-active { transition: opacity 180ms ease-out; }
.ua-fade-leave-active { transition: opacity 220ms ease-in; }
.ua-fade-enter-from,
.ua-fade-leave-to { opacity: 0; }
</style>
