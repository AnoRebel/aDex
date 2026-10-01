<template>
  <Teleport to="body">
    <Transition name="ua-fade" appear>
      <div
        v-if="modelValue && release"
        class="ua-overlay"
        role="dialog"
        :aria-labelledby="`ua-title-${id}`"
        aria-modal="true"
        @click.self="onLater"
        @keydown.esc="onLater"
      >
        <div class="ua-card mod-panel">
          <div class="ua-header">
            <div :id="`ua-title-${id}`" class="ua-title">
              <span class="ua-bracket">[</span>
              UPDATE AVAILABLE
              <span class="ua-bracket">]</span>
            </div>
            <button
              class="ua-close"
              type="button"
              aria-label="Close update notification"
              @click="onLater"
            >X</button>
          </div>

          <div class="ua-meta">
            <div class="ua-version-row">
              <span class="ua-version-label">CURRENT</span>
              <span class="ua-version-value">{{ currentVersion }}</span>
            </div>
            <div class="ua-version-row">
              <span class="ua-version-label">LATEST</span>
              <span class="ua-version-value ua-version-new">
                {{ release.tag_name }}
                <span v-if="release.prerelease" class="ua-tag-pre">(pre-release)</span>
              </span>
            </div>
            <div v-if="releasePublishedDisplay" class="ua-version-row">
              <span class="ua-version-label">PUBLISHED</span>
              <span class="ua-version-value">{{ releasePublishedDisplay }}</span>
            </div>
          </div>

          <div v-if="release.name && release.name !== release.tag_name" class="ua-name">
            {{ release.name }}
          </div>

          <!-- Release notes. We render as preformatted text rather
               than parsed markdown to avoid pulling a markdown lib
               just for this — the eDex aesthetic suits monospace
               release notes anyway. -->
          <div class="ua-notes-label">RELEASE NOTES</div>
          <pre class="ua-notes" tabindex="0">{{ releaseBody || '(no release notes provided)' }}</pre>

          <div class="ua-actions">
            <button class="ua-btn ua-btn-secondary" type="button" @click="onSkip">
              SKIP THIS VERSION
            </button>
            <button class="ua-btn ua-btn-secondary" type="button" @click="onLater">
              REMIND ME LATER
            </button>
            <button
              class="ua-btn ua-btn-primary"
              type="button"
              :disabled="!release.html_url"
              @click="onDownload"
            >
              DOWNLOAD UPDATE
            </button>
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
 * Non-blocking modal shown when useUpdateChecker reports a newer
 * release than the local build. Three exits:
 *
 *   - SKIP THIS VERSION → adds the tag to settings.skippedVersions so
 *     the modal won't reappear for this specific release. The user
 *     can clear the skip list from Settings → Updates.
 *   - REMIND ME LATER → closes the modal but leaves the cached
 *     release intact so the next check (next launch, next interval)
 *     can re-surface it.
 *   - DOWNLOAD UPDATE → opens the release page in the OS browser
 *     (via Wails BrowserOpenURL when available, window.open
 *     otherwise). We deliberately don't auto-download the asset —
 *     update binaries vary by platform and the user should review
 *     the release page first.
 *
 * Reads everything from useUpdateChecker so the modal stays a pure
 * presentation component; parent just toggles modelValue.
 */
import { computed, useId } from 'vue'
import { format as formatDateFn, isValid as isValidDate } from 'date-fns'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { useUpdateChecker, type GitHubRelease } from '~/composables/useUpdateChecker'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const id = useId()
const { state, currentVersion, skipVersion, openReleasePage } = useUpdateChecker()
const audio = useAdexAudio()

const release = computed<GitHubRelease | null>(() => state.value.latest)

const releasePublishedDisplay = computed<string>(() => {
  const iso = release.value?.published_at
  if (!iso) return ''
  const d = new Date(iso)
  if (!isValidDate(d)) return ''
  // Short locale-neutral form; eDex aesthetic prefers ISO-ish stamps.
  return formatDateFn(d, 'yyyy-MM-dd HH:mm')
})

// GitHub releases sometimes include very long bodies. We don't cap
// the height here (the .ua-notes scrollbar handles overflow), but we
// DO trim leading/trailing whitespace so the preformatted block
// doesn't start with a blank line.
const releaseBody = computed<string>(() => (release.value?.body ?? '').trim())

function close() {
  emit('update:modelValue', false)
}

function onLater() {
  // Routine UI click — not destructive, just dismissing the modal
  // for now. The release stays cached so the next interval fire
  // re-shows it.
  try { audio.playCue('click') } catch { /* non-fatal */ }
  close()
}

function onSkip() {
  // Destructive-ish: this version will never be shown again until
  // the user clears the skip list in Settings → Updates. We fire
  // the destructive cue (handgun-click) to signal that this is a
  // commit-and-forget action, not a casual dismiss.
  try { audio.playCue('destructive') } catch { /* non-fatal */ }
  if (release.value?.tag_name) {
    skipVersion(release.value.tag_name)
  }
  close()
}

function onDownload() {
  try { audio.playCue('click') } catch { /* non-fatal */ }
  openReleasePage()
  // Leave the modal open so the user can see what they're getting
  // alongside the browser tab — they can dismiss explicitly. The
  // alternative (auto-close on download click) hides the version
  // info the moment they need to read it.
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

.ua-tag-pre {
  opacity: 0.6;
  font-size: 0.85em;
  margin-left: 0.5em;
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
  flex: 1 1 0;
  min-height: 0;
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

.ua-fade-enter-active { transition: opacity 180ms ease-out; }
.ua-fade-leave-active { transition: opacity 220ms ease-in; }
.ua-fade-enter-from,
.ua-fade-leave-to { opacity: 0; }
</style>
