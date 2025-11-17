<template>
  <button
    :class="keyClasses"
    :style="keyStyles"
    :aria-label="ariaLabel"
    :aria-pressed="isPressed"
    role="button"
    tabindex="0"
    @mousedown="handleMouseDown"
    @mouseup="handleMouseUp"
    @mouseleave="handleMouseLeave"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    @touchcancel="handleTouchCancel"
    @keydown="handleKeyDown"
    @keyup="handleKeyUp"
  >
    <span class="key-primary">{{ displayKey }}</span>
    <span v-if="hasShift" class="key-shift">{{ shiftKey }}</span>
    <span v-if="showCapsLock" class="caps-indicator">⇪</span>
  </button>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useKeyRepeat } from '~/composables/useKeyboard'

interface KeyData {
  key: string
  shift?: string
  width?: number
  type?: 'standard' | 'function' | 'special' | 'modifier' | 'space' | 'number' | 'operator'
  side?: 'left' | 'right'
}

interface Props {
  keyData: KeyData
  isShiftActive?: boolean
  isCapsLockActive?: boolean
  size?: 'small' | 'medium' | 'large'
  theme?: 'dark' | 'light' | 'cyberpunk'
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isShiftActive: false,
  isCapsLockActive: false,
  size: 'medium',
  theme: 'dark',
  disabled: false
})

const emit = defineEmits<{
  keyPress: [key: string, shiftActive: boolean]
  keyRelease: [key: string]
  longPress: [key: string]
}>()

// State
const isPressed = ref(false)
const isLongPressed = ref(false)
const repeatTimer = ref<NodeJS.Timeout | null>(null)
const longPressTimer = ref<NodeJS.Timeout | null>(null)

// Computed properties
const displayKey = computed(() => {
  if (props.keyData.type === 'space') return '␣'
  if (props.keyData.type === 'modifier' && props.keyData.side) {
    return props.keyData.key.charAt(0).toUpperCase() + props.keyData.key.slice(1)
  }
  if (props.keyData.type === 'special') {
    const displayMap: Record<string, string> = {
      'Backspace': '⌫',
      'Tab': '⇥',
      'Enter': '↵',
      'CapsLock': '⇪',
      'Escape': 'Esc',
      'Delete': 'Del',
      'Insert': 'Ins',
      'Home': '⇱',
      'End': '⇲',
      'PageUp': '⇞',
      'PageDown': '⇟',
      'ArrowUp': '↑',
      'ArrowDown': '↓',
      'ArrowLeft': '←',
      'ArrowRight': '→'
    }
    return displayMap[props.keyData.key] || props.keyData.key
  }
  if (props.keyData.type === 'function') {
    return props.keyData.key
  }

  // Handle shift and caps lock
  const useShift = props.isShiftActive !== props.isCapsLockActive
  if (useShift && props.keyData.shift) {
    return props.keyData.shift
  }

  return props.keyData.key
})

const shiftKey = computed(() => props.keyData.shift)
const hasShift = computed(() => !!props.keyData.shift)
const showCapsLock = computed(() =>
  props.isCapsLockActive &&
  props.keyData.key.match(/^[a-z]$/i) &&
  props.keyData.key === props.keyData.key.toLowerCase()
)

const ariaLabel = computed(() => {
  if (props.keyData.type === 'space') return 'Space bar'
  if (props.keyData.type === 'modifier') {
    return `${props.keyData.key} modifier key (${props.keyData.side})`
  }
  if (hasShift.value) {
    return `${props.keyData.key} with shift ${props.keyData.shift}`
  }
  return props.keyData.key
})

const keyClasses = computed(() => [
  'key-button',
  `key-button--${props.size}`,
  `key-button--${props.theme}`,
  `key-button--${props.keyData.type || 'standard'}`,
  {
    'key-button--pressed': isPressed.value,
    'key-button--long-pressed': isLongPressed.value,
    'key-button--disabled': props.disabled,
    'key-button--modifier': props.keyData.type === 'modifier',
    'key-button--function': props.keyData.type === 'function',
    'key-button--space': props.keyData.type === 'space',
    'key-button--shift-active': props.isShiftActive,
    'key-button--caps-lock': props.isCapsLockActive
  }
])

const keyStyles = computed(() => {
  const baseWidth = props.keyData.width || 1
  const sizeMultipliers = {
    small: 2,
    medium: 2.5,
    large: 3
  }

  const unitSize = sizeMultipliers[props.size]
  const width = baseWidth * unitSize

  return {
    '--key-width': `${width}rem`,
    '--key-height': `${unitSize}rem`
  }
})

// Composable for key repeat functionality
const { startRepeat, stopRepeat } = useKeyRepeat()

// Methods
const getActualKey = () => {
  if (props.keyData.type === 'space') return ' '
  if (props.keyData.type === 'modifier') return props.keyData.key

  const useShift = props.isShiftActive !== props.isCapsLockActive
  if (useShift && props.keyData.shift) {
    return props.keyData.shift
  }

  return props.keyData.key
}

const handleKeyPress = () => {
  if (props.disabled) return

  const actualKey = getActualKey()
  isPressed.value = true

  emit('keyPress', actualKey, props.isShiftActive)

  // Start repeat for non-modifier keys (except Enter and Tab)
  if (props.keyData.type !== 'modifier' &&
      props.keyData.key !== 'Enter' &&
      props.keyData.key !== 'Tab') {
    startRepeat(() => {
      emit('keyPress', actualKey, props.isShiftActive)
    })

    // Start long press timer
    longPressTimer.value = setTimeout(() => {
      isLongPressed.value = true
      emit('longPress', actualKey)
    }, 500)
  }

  // Haptic feedback if available
  if ('vibrate' in navigator) {
    navigator.vibrate(10)
  }
}

const handleKeyRelease = () => {
  if (props.disabled) return

  isPressed.value = false
  isLongPressed.value = false

  stopRepeat()

  if (longPressTimer.value) {
    clearTimeout(longPressTimer.value)
    longPressTimer.value = null
  }

  emit('keyRelease', props.keyData.key)
}

// Mouse event handlers
const handleMouseDown = (event: MouseEvent) => {
  event.preventDefault()
  handleKeyPress()
}

const handleMouseUp = (event: MouseEvent) => {
  event.preventDefault()
  handleKeyRelease()
}

const handleMouseLeave = (event: MouseEvent) => {
  if (isPressed.value) {
    handleKeyRelease()
  }
}

// Touch event handlers
const handleTouchStart = (event: TouchEvent) => {
  event.preventDefault()
  handleKeyPress()
}

const handleTouchEnd = (event: TouchEvent) => {
  event.preventDefault()
  handleKeyRelease()
}

const handleTouchCancel = (event: TouchEvent) => {
  if (isPressed.value) {
    handleKeyRelease()
  }
}

// Keyboard event handlers for accessibility
const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    handleKeyPress()
  }
}

const handleKeyUp = (event: KeyboardEvent) => {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    handleKeyRelease()
  }
}

// Cleanup
onUnmounted(() => {
  stopRepeat()
  if (repeatTimer.value) {
    clearTimeout(repeatTimer.value)
  }
  if (longPressTimer.value) {
    clearTimeout(longPressTimer.value)
  }
})
</script>

<style scoped>
.key-button {
  position: relative;
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-medium, 0.375rem);
  background: var(--dex-theme-input-background, #1a1a1a);
  color: var(--dex-theme-input-foreground, #ffffff);
  font-family: var(--dex-theme-font-family-mono, monospace);
  font-weight: 500;
  cursor: pointer;
  user-select: none;
  transition: all 0.15s ease-out;
  min-height: var(--key-height);
  width: var(--key-width);
  padding: 0.25rem;
  margin: 0.125rem;
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  will-change: transform, box-shadow;
}

.key-button:hover:not(.key-button--disabled) {
  background: var(--dex-theme-input-border, #2a2a2a);
  transform: translateY(-1px);
  box-shadow:
    0 4px 8px rgba(0, 0, 0, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.key-button--pressed {
  transform: translateY(1px);
  box-shadow:
    0 1px 2px rgba(0, 0, 0, 0.2),
    inset 0 2px 4px rgba(0, 0, 0, 0.3);
  background: var(--dex-theme-button-active, #00ff41);
}

.key-button--long-pressed {
  animation: keyPulse 0.5s infinite alternate;
}

.key-button--disabled {
  opacity: 0.5;
  cursor: not-allowed;
  filter: grayscale(50%);
}

/* Key type specific styles */
.key-button--modifier {
  background: var(--dex-theme-accent-primary, #00ff41);
  color: var(--dex-theme-background-primary, #000000);
  font-weight: 600;
}

.key-button--function {
  background: var(--dex-theme-background-secondary, #2a2a2a);
  font-size: 0.75rem;
}

.key-button--space {
  background: var(--dex-theme-background-tertiary, #3a3a3a);
  border-radius: var(--dex-theme-border-radius-large, 0.5rem);
}

.key-button--number,
.key-button--operator {
  background: var(--dex-theme-accent-secondary, #00cc33);
  color: var(--dex-theme-background-primary, #000000);
}

/* Size variants */
.key-button--small {
  font-size: 0.75rem;
}

.key-button--medium {
  font-size: 0.875rem;
}

.key-button--large {
  font-size: 1rem;
}

/* Theme variants */
.key-button--light {
  --key-bg: #ffffff;
  --key-text: #000000;
  --key-border: #cccccc;
  --key-hover: #f5f5f5;
}

.key-button--cyberpunk {
  border-color: #ff00ff;
  background: #0a0a0f;
  color: #00ffff;
  box-shadow:
    0 0 10px rgba(255, 0, 255, 0.3),
    inset 0 1px 0 rgba(0, 255, 255, 0.2);
}

.key-button--cyberpunk:hover {
  box-shadow:
    0 0 20px rgba(255, 0, 255, 0.5),
    inset 0 1px 0 rgba(0, 255, 255, 0.4);
}

.key-button--cyberpunk.key-button--pressed {
  background: #ff0080;
  box-shadow:
    0 0 30px rgba(255, 0, 128, 0.7),
    inset 0 2px 4px rgba(0, 0, 0, 0.5);
}

/* Key elements */
.key-primary {
  font-weight: 500;
  line-height: 1.2;
}

.key-shift {
  position: absolute;
  top: 2px;
  right: 4px;
  font-size: 0.625rem;
  opacity: 0.7;
  font-weight: 400;
}

.caps-indicator {
  position: absolute;
  top: 2px;
  left: 4px;
  font-size: 0.5rem;
  opacity: 0.8;
}

/* Accessibility */
.key-button:focus {
  outline: 2px solid var(--dex-theme-input-focus, #00ff41);
  outline-offset: 2px;
}

/* Animations */
@keyframes keyPulse {
  0% {
    transform: scale(1);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 15px var(--dex-theme-accent-primary, #00ff41);
  }
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .key-button {
    margin: 0.0625rem;
    font-size: 0.75rem;
  }

  .key-button--small {
    font-size: 0.625rem;
  }

  .key-button--large {
    font-size: 0.875rem;
  }
}
</style>