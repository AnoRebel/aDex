<template>
  <Teleport to="body">
    <Transition name="keyboard">
      <div
        v-if="isVisible"
        :class="keyboardClasses"
        :style="keyboardStyles"
        role="dialog"
        :aria-label="'On-screen keyboard'"
        aria-modal="true"
        @keydown.esc="handleEscapeKey"
        @mousedown="handleMouseDown"
      >
        <!-- Keyboard backdrop -->
        <div
          class="keyboard-backdrop"
          @click="handleBackdropClick"
          aria-hidden="true"
        />

        <!-- Keyboard container -->
        <div
          class="keyboard-container"
          role="application"
          :aria-label="`Virtual keyboard: ${currentLayoutData?.displayName || 'QWERTY'}`"
        >
          <!-- Keyboard header -->
          <div class="keyboard-header">
            <div class="keyboard-title">
              <h2 class="keyboard-title-text">Virtual Keyboard</h2>
              <p class="keyboard-subtitle">{{ currentLayoutData?.displayName || 'QWERTY' }} Layout</p>
            </div>

            <div class="keyboard-controls">
              <!-- Layout switcher -->
              <select
                v-model="selectedLayout"
                aria-label="Select keyboard layout"
                class="layout-switcher"
                title="Switch keyboard layout"
              >
                <option
                  v-for="layout in availableLayouts"
                  :key="layout.name"
                  :value="layout.name"
                >
                  {{ layout.displayName }}
                </option>
              </select>

              <!-- Settings toggle -->
              <button
                :class="['settings-button', { 'settings-button--active': showSettings }]"
                :aria-expanded="showSettings"
                aria-label="Keyboard settings"
                title="Keyboard settings"
                @click="toggleSettings"
              >
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M8 4.754a3.246 3.246 0 1 0 0 6.492 3.246 3.246 0 0 0 0-6.492zM5.754 8a2.246 2.246 0 1 1 4.492 0 2.246 2.246 0 0 1-4.492 0z"/>
                  <path d="M9.796 1.343c-5.864 0-10.713 4.85-10.713 10.713 0 5.864 4.85 10.713 10.713 10.713 5.864 0 10.713-4.85 10.713-10.713 0-5.864-4.85-10.713-10.713-10.713zm1.44 2.518c.095.368.198.747.307 1.135.448.018.883.056 1.304.112l.432-1.257c.08.018.158.038.236.058.346.089.69.189 1.03.299l.525-1.215a13.477 13.477 0 0 1 2.095.974l-.254 1.306c.436.272.847.584 1.23.938.381-.354.792-.666 1.228-.938l-.254-1.306c.336-.208.68-.408 1.03-.597.078-.02.156-.04.236-.058l.432 1.257c.42-.056.855-.094 1.303-.112.11-.388.212-.767.307-1.135l1.31-.197c.066-.59.102-1.191.102-1.802 0-.61-.036-1.211-.102-1.802l-1.31-.197c-.095-.368-.198-.747-.307-1.135-.448-.018-.883-.056-1.304-.112l-.432-1.257c-.08.018-.158.038-.236.058-.346.089-.69.189-1.03.299l-.525-1.215a13.477 13.477 0 0 0-2.095.974l.254 1.306c-.436.272-.847.584-1.23.938-.381-.354-.792-.666-1.228-.938l.254-1.306c-.336-.208-.68-.408-1.03-.597a6.36 6.36 0 0 0-.236-.058l-.432 1.257c-.42-.056-.855-.094-1.303-.112-.11-.388-.212-.767-.307-1.135l-1.31-.197C8.036 5.588 8 4.987 8 4.376c0-.61.036-1.211.102-1.802l1.31-.197z"/>
                </svg>
              </button>

              <!-- Hide button -->
              <button
                class="hide-button"
                aria-label="Hide keyboard"
                title="Hide keyboard"
                @click="hideKeyboard"
              >
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M7.646 4.646a.5.5 0 0 1 .708 0L10 6.293l1.646-1.647a.5.5 0 0 1 .708.708L10.707 7l1.647 1.646a.5.5 0 0 1-.708.708L10 7.707l-1.646 1.647a.5.5 0 0 1-.708-.708L8.293 7 6.646 5.354a.5.5 0 0 1 0-.708z"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Settings panel -->
          <Transition name="settings">
            <div v-if="showSettings" class="keyboard-settings">
              <div class="settings-row">
                <label class="settings-label">Theme:</label>
                <select v-model="selectedTheme" class="settings-select">
                  <option value="dark">Dark</option>
                  <option value="light">Light</option>
                  <option value="cyberpunk">Cyberpunk</option>
                </select>
              </div>

              <div class="settings-row">
                <label class="settings-label">Size:</label>
                <select v-model="selectedSize" class="settings-select">
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </div>

              <div class="settings-row">
                <label class="settings-label">Opacity:</label>
                <input
                  v-model.number="keyboardOpacity"
                  type="range"
                  min="0.1"
                  max="1"
                  step="0.1"
                  class="settings-slider"
                />
                <span class="settings-value">{{ (keyboardOpacity * 100).toFixed(0) }}%</span>
              </div>

              <div class="settings-row">
                <label class="settings-label">Key Repeat:</label>
                <input
                  v-model.number="repeatDelay"
                  type="range"
                  min="100"
                  max="1000"
                  step="100"
                  class="settings-slider"
                />
                <span class="settings-value">{{ repeatDelay }}ms</span>
              </div>

              <div class="settings-row">
                <label class="settings-checkbox">
                  <input
                    v-model="hapticFeedback"
                    type="checkbox"
                    class="settings-input"
                  />
                  <span>Haptic Feedback</span>
                </label>
              </div>

              <div class="settings-row">
                <label class="settings-checkbox">
                  <input
                    v-model="soundFeedback"
                    type="checkbox"
                    class="settings-input"
                  />
                  <span>Sound Feedback</span>
                </label>
              </div>

              <div class="settings-row">
                <label class="settings-checkbox">
                  <input
                    v-model="autoShow"
                    type="checkbox"
                    class="settings-input"
                  />
                  <span>Auto-show on input focus</span>
                </label>
              </div>
            </div>
          </Transition>

          <!-- Keyboard layout -->
          <KeyLayout
            :layout-name="selectedLayout"
            :show-function-row="showFunctionRow"
            :show-numpad="showNumpad"
            :show-layout-selector="false"
            :size="selectedSize"
            :theme="selectedTheme"
            :opacity="keyboardOpacity"
            :maximize-width="true"
            @key-press="handleKeyPress"
            @key-release="handleKeyRelease"
            @long-press="handleLongPress"
            @layout-change="handleLayoutChange"
          />

          <!-- Status bar -->
          <div class="keyboard-status">
            <div class="status-indicators">
              <span
                v-if="isShiftActive"
                class="status-indicator status-indicator--active"
                title="Shift active"
              >
                ⇧
              </span>
              <span
                v-if="isCapsLockActive"
                class="status-indicator status-indicator--active"
                title="Caps Lock active"
              >
                ⇪
              </span>
              <span
                v-if="isCtrlActive"
                class="status-indicator status-indicator--active"
                title="Ctrl active"
              >
                ^Ctrl
              </span>
              <span
                v-if="isAltActive"
                class="status-indicator status-indicator--active"
                title="Alt active"
              >
                ⎇Alt
              </span>
            </div>

            <div class="status-text">
              {{ keystrokeCount }} keystrokes
            </div>

            <div class="status-actions">
              <button
                class="status-button"
                @click="resetKeyboard"
                title="Reset modifiers"
              >
                Reset
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useKeyboard } from '~/composables/useKeyboard'
import { useMagicKeys } from '@vueuse/core'
import KeyLayout from './KeyLayout.vue'

interface Props {
  modelValue?: boolean
  targetElement?: HTMLElement
  autoShow?: boolean
  showFunctionRow?: boolean
  showNumpad?: boolean
  layout?: string
  size?: 'small' | 'medium' | 'large'
  theme?: 'dark' | 'light' | 'cyberpunk'
  opacity?: number
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  autoShow: false,
  showFunctionRow: false,
  showNumpad: false,
  layout: '',
  size: 'medium',
  theme: 'dark',
  opacity: 0.9
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  keyPress: [key: string, modifiers: any]
  keyRelease: [key: string]
  longPress: [key: string]
  layoutChange: [layoutName: string]
}>()

// Use keyboard composable
const {
  isVisible,
  currentLayout,
  currentLayoutData,
  availableLayouts,
  settings,
  isShiftActive,
  isCapsLockActive,
  isCtrlActive,
  isAltActive,
  setLayout,
  showKeyboard,
  hideKeyboard,
  resetModifiers,
  updateSettings
} = useKeyboard()

// Internal state
const showSettings = ref(false)
const selectedLayout = ref(props.layout || currentLayout.value)
const selectedTheme = ref(props.theme)
const selectedSize = ref(props.size)
const keyboardOpacity = ref(props.opacity)
const keystrokeCount = ref(0)

// Settings
const repeatDelay = ref(settings.value.keyRepeatDelay)
const hapticFeedback = ref(settings.value.hapticFeedback)
const soundFeedback = ref(settings.value.soundFeedback)
const autoShow = ref(props.autoShow || settings.value.autoShow)
const showFunctionRow = ref(props.showFunctionRow)
const showNumpad = ref(props.showNumpad)

// Computed properties
const keyboardClasses = computed(() => [
  'onscreen-keyboard',
  `onscreen-keyboard--${selectedTheme.value}`,
  {
    'onscreen-keyboard--visible': isVisible.value,
    'onscreen-keyboard--with-settings': showSettings.value
  }
])

const keyboardStyles = computed(() => {
  let styles: Record<string, string> = {}

  if (props.targetElement && typeof window !== 'undefined') {
    const rect = props.targetElement.getBoundingClientRect()
    const viewportHeight = window.innerHeight

    styles.bottom = `${viewportHeight - rect.bottom}px`
  } else {
    styles.bottom = '0'
  }

  return styles
})

// Methods
const handleKeyPress = (key: string, modifiers: any) => {
  keystrokeCount.value++

  // Emit to parent
  emit('keyPress', key, modifiers)

  // Sound feedback (if implemented)
  if (soundFeedback.value && typeof window !== 'undefined') {
    playKeySound(key)
  }
}

const handleKeyRelease = (key: string) => {
  emit('keyRelease', key)
}

const handleLongPress = (key: string) => {
  emit('longPress', key)
}

const handleLayoutChange = (layoutName: string) => {
  emit('layoutChange', layoutName)
}

const toggleSettings = () => {
  showSettings.value = !showSettings.value
}

const resetKeyboard = () => {
  resetModifiers()
  keystrokeCount.value = 0
}

const handleEscapeKey = () => {
  hideKeyboard()
}

const handleBackdropClick = () => {
  hideKeyboard()
}

const handleMouseDown = (event: MouseEvent) => {
  // Prevent clicks from propagating to underlying elements
  event.preventDefault()
  event.stopPropagation()
}

// Audio feedback (placeholder implementation)
const playKeySound = (key: string) => {
  try {
    const context = new (window.AudioContext || (window as any).webkitAudioContext)()
    const oscillator = context.createOscillator()
    const gainNode = context.createGain()

    oscillator.connect(gainNode)
    gainNode.connect(context.destination)

    // Different frequencies for different key types
    let frequency = 800
    if (key === ' ') frequency = 600
    else if (key === 'Enter') frequency = 1000
    else if (key.length === 1 && key.match(/[a-z]/i)) frequency = 900

    oscillator.frequency.setValueAtTime(frequency, context.currentTime)
    oscillator.type = 'sine'

    gainNode.gain.setValueAtTime(0.1, context.currentTime)
    gainNode.gain.exponentialRampToValueAtTime(0.01, context.currentTime + 0.05)

    oscillator.start(context.currentTime)
    oscillator.stop(context.currentTime + 0.05)
  } catch (error) {
    // Ignore audio errors
  }
}

// Watchers
watch(() => props.modelValue, (newValue) => {
  if (newValue !== isVisible.value) {
    if (newValue) {
      showKeyboard()
    } else {
      hideKeyboard()
    }
  }
})

watch(isVisible, (newValue) => {
  emit('update:modelValue', newValue)
})

watch(selectedLayout, (newLayout) => {
  setLayout(newLayout)
})

watch([selectedTheme, selectedSize, keyboardOpacity, repeatDelay, hapticFeedback, soundFeedback, autoShow], () => {
  updateSettings({
    theme: selectedTheme.value,
    size: selectedSize.value,
    opacity: keyboardOpacity.value,
    keyRepeatDelay: repeatDelay.value,
    hapticFeedback: hapticFeedback.value,
    soundFeedback: soundFeedback.value,
    autoShow: autoShow.value
  })
})

watch(() => props.layout, (newLayout) => {
  if (newLayout && newLayout !== selectedLayout.value) {
    selectedLayout.value = newLayout
  }
})

watch(() => props.theme, (newTheme) => {
  if (newTheme !== selectedTheme.value) {
    selectedTheme.value = newTheme
  }
})

watch(() => props.size, (newSize) => {
  if (newSize !== selectedSize.value) {
    selectedSize.value = newSize
  }
})

watch(() => props.opacity, (newOpacity) => {
  if (newOpacity !== keyboardOpacity.value) {
    keyboardOpacity.value = newOpacity
  }
})

// Magic keys for hardware keyboard integration
const { escape } = useMagicKeys()
watch(escape, (pressed) => {
  if (pressed && isVisible.value) {
    hideKeyboard()
  }
})

// Auto-show functionality
const setupAutoShow = () => {
  if (!autoShow.value || typeof window === 'undefined') return

  const handleInputFocus = (event: FocusEvent) => {
    const target = event.target as HTMLElement
    if (target && (
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.contentEditable === 'true'
    )) {
      // Small delay to ensure the keyboard appears after the focus
      setTimeout(() => {
        if (!isVisible.value) {
          showKeyboard()
        }
      }, 100)
    }
  }

  const handleInputBlur = (event: FocusEvent) => {
    const target = event.target as HTMLElement
    if (target && (
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.contentEditable === 'true'
    )) {
      // Small delay to allow focus to move to another input
      setTimeout(() => {
        const activeElement = document.activeElement
        if (!activeElement || (
          activeElement.tagName !== 'INPUT' &&
          activeElement.tagName !== 'TEXTAREA' &&
          activeElement.contentEditable !== 'true'
        )) {
          hideKeyboard()
        }
      }, 200)
    }
  }

  document.addEventListener('focusin', handleInputFocus)
  document.addEventListener('focusout', handleInputBlur)

  onUnmounted(() => {
    document.removeEventListener('focusin', handleInputFocus)
    document.removeEventListener('focusout', handleInputBlur)
  })
}

// Lifecycle
onMounted(() => {
  setupAutoShow()
})
</script>

<style scoped>
.onscreen-keyboard {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  font-family: var(--dex-theme-font-family-secondary, 'Inter', sans-serif);
}

.onscreen-keyboard--visible {
  pointer-events: auto;
}

.keyboard-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(2px);
  z-index: 1;
}

.keyboard-container {
  position: relative;
  z-index: 2;
  background: var(--dex-theme-background-primary, #0a0a0a);
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-large, 1rem) var(--dex-theme-border-radius-large, 1rem) 0 0;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(10px);
  max-width: 100vw;
  margin: 0 auto;
}

/* Theme variants */
.onscreen-keyboard--light .keyboard-container {
  background: rgba(255, 255, 255, 0.95);
  border-color: #e0e0e0;
  color: #000000;
}

.onscreen-keyboard--cyberpunk .keyboard-container {
  background: rgba(10, 10, 15, 0.95);
  border-color: #ff00ff;
  box-shadow: 0 -4px 30px rgba(255, 0, 255, 0.3), 0 -4px 20px rgba(0, 0, 0, 0.8);
}

/* Header */
.keyboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--dex-theme-border, #333);
  gap: 1rem;
}

.keyboard-title {
  flex: 1;
}

.keyboard-title-text {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--dex-theme-foreground-primary, #ffffff);
}

.keyboard-subtitle {
  margin: 0;
  font-size: 0.75rem;
  color: var(--dex-theme-foreground-secondary, #cccccc);
  opacity: 0.8;
}

.keyboard-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.layout-switcher {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-medium, 0.375rem);
  background: var(--dex-theme-input-background, #1a1a1a);
  color: var(--dex-theme-input-foreground, #ffffff);
  font-size: 0.75rem;
  cursor: pointer;
}

.settings-button,
.hide-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-medium, 0.375rem);
  background: var(--dex-theme-input-background, #1a1a1a);
  color: var(--dex-theme-input-foreground, #ffffff);
  cursor: pointer;
  transition: all 0.2s ease;
}

.settings-button--active {
  background: var(--dex-theme-accent-primary, #00ff41);
  color: var(--dex-theme-background-primary, #000000);
  border-color: var(--dex-theme-accent-primary, #00ff41);
}

/* Settings panel */
.keyboard-settings {
  padding: 1rem;
  border-bottom: 1px solid var(--dex-theme-border, #333);
  background: var(--dex-theme-background-secondary, #1a1a1a);
}

.settings-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
  gap: 1rem;
}

.settings-row:last-child {
  margin-bottom: 0;
}

.settings-label {
  font-size: 0.875rem;
  color: var(--dex-theme-foreground-primary, #ffffff);
  font-weight: 500;
}

.settings-select,
.settings-slider {
  flex: 1;
  max-width: 200px;
}

.settings-value {
  min-width: 3rem;
  text-align: right;
  font-size: 0.75rem;
  color: var(--dex-theme-foreground-secondary, #cccccc);
}

.settings-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--dex-theme-foreground-primary, #ffffff);
}

.settings-input {
  margin: 0;
}

/* Status bar */
.keyboard-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  border-top: 1px solid var(--dex-theme-border, #333);
  background: var(--dex-theme-background-secondary, #1a1a1a);
  font-size: 0.75rem;
}

.status-indicators {
  display: flex;
  gap: 0.25rem;
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  padding: 0.125rem 0.25rem;
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-small, 0.25rem);
  background: var(--dex-theme-background-tertiary, #2a2a2a);
  color: var(--dex-theme-foreground-secondary, #cccccc);
  font-size: 0.625rem;
  opacity: 0.5;
}

.status-indicator--active {
  opacity: 1;
  background: var(--dex-theme-accent-primary, #00ff41);
  color: var(--dex-theme-background-primary, #000000);
  border-color: var(--dex-theme-accent-primary, #00ff41);
}

.status-text {
  color: var(--dex-theme-foreground-secondary, #cccccc);
  opacity: 0.8;
}

.status-button {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-small, 0.25rem);
  background: transparent;
  color: var(--dex-theme-foreground-secondary, #cccccc);
  font-size: 0.625rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.status-button:hover {
  background: var(--dex-theme-accent-primary, #00ff41);
  color: var(--dex-theme-background-primary, #000000);
  border-color: var(--dex-theme-accent-primary, #00ff41);
}

/* Animations */
.keyboard-enter-active,
.keyboard-leave-active {
  transition: all 0.3s ease;
}

.keyboard-enter-from,
.keyboard-leave-to {
  transform: translateY(100%);
}

.settings-enter-active,
.settings-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}

.settings-enter-from,
.settings-leave-to {
  max-height: 0;
  opacity: 0;
}

.settings-enter-to,
.settings-leave-from {
  max-height: 300px;
  opacity: 1;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .keyboard-header {
    padding: 0.5rem;
  }

  .keyboard-title-text {
    font-size: 0.75rem;
  }

  .keyboard-subtitle {
    font-size: 0.625rem;
  }

  .settings-row {
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
  }

  .keyboard-status {
    padding: 0.375rem 0.5rem;
  }

  .status-indicator {
    min-width: 1.5rem;
    font-size: 0.5rem;
  }
}

/* Print styles */
@media print {
  .onscreen-keyboard {
    display: none;
  }
}
</style>