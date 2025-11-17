<template>
  <div class="virtual-keyboard" :class="{ visible: isVisible, minimized: isMinimized }">
    <div class="keyboard-header">
      <div class="keyboard-controls">
        <button @click="toggleMinimize" class="control-btn">
          <span class="btn-icon">{{ isMinimized ? '▲' : '▼' }}</span>
        </button>
        <button @click="toggleKeyboard" class="control-btn close-btn">
          <span class="btn-icon">✕</span>
        </button>
      </div>
      <div class="keyboard-title">
        <span class="glitch">VIRTUAL KEYBOARD</span>
      </div>
      <div class="keyboard-settings">
        <button @click="toggleSettings" class="control-btn">
          <span class="btn-icon">⚙</span>
        </button>
      </div>
    </div>

    <!-- Settings Panel -->
    <div v-if="showSettings" class="keyboard-settings-panel">
      <div class="settings-group">
        <label class="setting-label">
          <span>Keyboard Layout</span>
          <select v-model="selectedLayout" @change="changeLayout" class="setting-select">
            <option v-for="layout in availableLayouts" :key="layout.id" :value="layout.id">
              {{ layout.name }}
            </option>
          </select>
        </label>
      </div>
      <div class="settings-group">
        <label class="setting-label">
          <input v-model="autoHide" type="checkbox" />
          <span>Auto-hide when not typing</span>
        </label>
      </div>
      <div class="settings-group">
        <label class="setting-label">
          <input v-model="soundEnabled" type="checkbox" />
          <span>Key sounds</span>
        </label>
      </div>
      <div class="settings-group">
        <label class="setting-label">
          <span>Key size</span>
          <input
            v-model="keySize"
            type="range"
            min="30"
            max="60"
            class="setting-range"
          />
          <span class="size-value">{{ keySize }}px</span>
        </label>
      </div>
    </div>

    <!-- Main Keyboard Layout -->
    <div v-if="!isMinimized" class="keyboard-layout">
      <!-- Number Row -->
      <div class="keyboard-row">
        <button
          v-for="key in currentLayout.numberRow"
          :key="key.value"
          class="key key-number"
          :class="getKeyClass(key)"
          @click="handleKeyPress(key)"
          @touchstart="handleTouchStart(key)"
          @touchend="handleTouchEnd(key)"
        >
          <span class="key-main">{{ getMainLabel(key) }}</span>
          <span v-if="key.shift" class="key-shift">{{ key.shift }}</span>
        </button>
        <button
          key="backspace"
          class="key key-special key-backspace"
          @click="handleSpecialKey('backspace')"
          @touchstart="handleTouchStart('backspace')"
          @touchend="handleTouchEnd('backspace')"
        >
          <span class="key-icon">⌫</span>
          <span class="key-label">BACKSPACE</span>
        </button>
      </div>

      <!-- QWERTY Row -->
      <div class="keyboard-row">
        <button
          key="tab"
          class="key key-special key-tab"
          @click="handleSpecialKey('tab')"
          @touchstart="handleTouchStart('tab')"
          @touchend="handleTouchEnd('tab')"
        >
          <span class="key-icon">⇥</span>
          <span class="key-label">TAB</span>
        </button>
        <button
          v-for="key in currentLayout.qwertyRow"
          :key="key.value"
          class="key key-letter"
          :class="getKeyClass(key)"
          @click="handleKeyPress(key)"
          @touchstart="handleTouchStart(key)"
          @touchend="handleTouchEnd(key)"
        >
          <span class="key-main">{{ getMainLabel(key) }}</span>
          <span v-if="key.shift" class="key-shift">{{ key.shift }}</span>
        </button>
      </div>

      <!-- ASDF Row -->
      <div class="keyboard-row">
        <button
          key="caps"
          class="key key-special key-caps"
          :class="{ active: capsLock }"
          @click="handleCapsLock"
          @touchstart="handleTouchStart('caps')"
          @touchend="handleTouchEnd('caps')"
        >
          <span class="key-icon">⇪</span>
          <span class="key-label">CAPS</span>
        </button>
        <button
          v-for="key in currentLayout.asdfRow"
          :key="key.value"
          class="key key-letter"
          :class="getKeyClass(key)"
          @click="handleKeyPress(key)"
          @touchstart="handleTouchStart(key)"
          @touchend="handleTouchEnd(key)"
        >
          <span class="key-main">{{ getMainLabel(key) }}</span>
          <span v-if="key.shift" class="key-shift">{{ key.shift }}</span>
        </button>
        <button
          key="enter"
          class="key key-special key-enter"
          @click="handleSpecialKey('enter')"
          @touchstart="handleTouchStart('enter')"
          @touchend="handleTouchEnd('enter')"
        >
          <span class="key-icon">↵</span>
          <span class="key-label">ENTER</span>
        </button>
      </div>

      <!-- ZXCV Row -->
      <div class="keyboard-row">
        <button
          key="shift"
          class="key key-special key-shift"
          :class="{ active: shiftPressed }"
          @click="handleShift"
          @touchstart="handleTouchStart('shift')"
          @touchend="handleTouchEnd('shift')"
        >
          <span class="key-icon">⇧</span>
          <span class="key-label">SHIFT</span>
        </button>
        <button
          v-for="key in currentLayout.zxcvRow"
          :key="key.value"
          class="key key-letter"
          :class="getKeyClass(key)"
          @click="handleKeyPress(key)"
          @touchstart="handleTouchStart(key)"
          @touchend="handleTouchEnd(key)"
        >
          <span class="key-main">{{ getMainLabel(key) }}</span>
          <span v-if="key.shift" class="key-shift">{{ key.shift }}</span>
        </button>
      </div>

      <!-- Space Row -->
      <div class="keyboard-row">
        <button
          v-for="key in currentLayout.spaceRow"
          :key="key.value"
          class="key"
          :class="`key-${key.type}`"
          @click="handleSpecialKey(key.value)"
          @touchstart="handleTouchStart(key.value)"
          @touchend="handleTouchEnd(key.value)"
        >
          <span class="key-icon">{{ key.icon }}</span>
          <span class="key-label">{{ key.label }}</span>
        </button>
        <button
          key="space"
          class="key key-space"
          @click="handleSpecialKey('space')"
          @touchstart="handleTouchStart('space')"
          @touchend="handleTouchEnd('space')"
        >
          <span class="key-label">SPACE</span>
        </button>
      </div>
    </div>

    <!-- Visual feedback for pressed keys -->
    <div
      v-if="pressedKey"
      class="key-feedback"
      :style="{
        left: pressedKeyPosition.x + 'px',
        top: pressedKeyPosition.y + 'px'
      }"
    >
      {{ getMainLabel(pressedKey) }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'

interface Key {
  value: string
  shift?: string
  alt?: string
}

interface Layout {
  id: string
  name: string
  numberRow: Key[]
  qwertyRow: Key[]
  asdfRow: Key[]
  zxcvRow: Key[]
  spaceRow: Key[]
}

interface KeyPosition {
  x: number
  y: number
}

// Props
const emit = defineEmits<{
  keypress: [key: string]
  specialKey: [key: string]
}>()

// Reactive data
const isVisible = ref<boolean>(false)
const isMinimized = ref<boolean>(false)
const showSettings = ref<boolean>(false)
const selectedLayout = ref<string>('qwerty')
const shiftPressed = ref<boolean>(false)
const capsLock = ref<boolean>(false)
const autoHide = ref<boolean>(true)
const soundEnabled = ref<boolean>(true)
const keySize = ref<number>(45)
const pressedKey = ref<Key | string | null>(null)
const pressedKeyPosition = ref<KeyPosition>({ x: 0, y: 0 })

// Keyboard layouts
const availableLayouts: Layout[] = [
  {
    id: 'qwerty',
    name: 'QWERTY',
    numberRow: [
      { value: '`', shift: '~' },
      { value: '1', shift: '!' },
      { value: '2', shift: '@' },
      { value: '3', shift: '#' },
      { value: '4', shift: '$' },
      { value: '5', shift: '%' },
      { value: '6', shift: '^' },
      { value: '7', shift: '&' },
      { value: '8', shift: '*' },
      { value: '9', shift: '(' },
      { value: '0', shift: ')' },
      { value: '-', shift: '_' },
      { value: '=', shift: '+' }
    ],
    qwertyRow: [
      { value: 'q', shift: 'Q' },
      { value: 'w', shift: 'W' },
      { value: 'e', shift: 'E' },
      { value: 'r', shift: 'R' },
      { value: 't', shift: 'T' },
      { value: 'y', shift: 'Y' },
      { value: 'u', shift: 'U' },
      { value: 'i', shift: 'I' },
      { value: 'o', shift: 'O' },
      { value: 'p', shift: 'P' },
      { value: '[', shift: '{' },
      { value: ']', shift: '}' },
      { value: '\\', shift: '|' }
    ],
    asdfRow: [
      { value: 'a', shift: 'A' },
      { value: 's', shift: 'S' },
      { value: 'd', shift: 'D' },
      { value: 'f', shift: 'F' },
      { value: 'g', shift: 'G' },
      { value: 'h', shift: 'H' },
      { value: 'j', shift: 'J' },
      { value: 'k', shift: 'K' },
      { value: 'l', shift: 'L' },
      { value: ';', shift: ':' },
      { value: "'", shift: '"' }
    ],
    zxcvRow: [
      { value: 'z', shift: 'Z' },
      { value: 'x', shift: 'X' },
      { value: 'c', shift: 'C' },
      { value: 'v', shift: 'V' },
      { value: 'b', shift: 'B' },
      { value: 'n', shift: 'N' },
      { value: 'm', shift: 'M' },
      { value: ',', shift: '<' },
      { value: '.', shift: '>' },
      { value: '/', shift: '?' }
    ],
    spaceRow: [
      { value: 'ctrl', label: 'CTRL', type: 'modifier', icon: '⌃' },
      { value: 'alt', label: 'ALT', type: 'modifier', icon: '⌥' },
      { value: 'meta', label: 'CMD', type: 'modifier', icon: '⌘' }
    ]
  },
  {
    id: 'dvorak',
    name: 'Dvorak',
    numberRow: [
      { value: '`', shift: '~' },
      { value: '1', shift: '!' },
      { value: '2', shift: '@' },
      { value: '3', shift: '#' },
      { value: '4', shift: '$' },
      { value: '5', shift: '%' },
      { value: '6', shift: '^' },
      { value: '7', shift: '&' },
      { value: '8', shift: '*' },
      { value: '9', shift: '(' },
      { value: '0', shift: ')' },
      { value: '[', shift: '{' },
      { value: ']', shift: '}' }
    ],
    qwertyRow: [
      { value: "'", shift: '"' },
      { value: ',', shift: '<' },
      { value: '.', shift: '>' },
      { value: 'p', shift: 'P' },
      { value: 'y', shift: 'Y' },
      { value: 'f', shift: 'F' },
      { value: 'g', shift: 'G' },
      { value: 'c', shift: 'C' },
      { value: 'r', shift: 'R' },
      { value: 'l', shift: 'L' },
      { value: '/', shift: '?' },
      { value: '=', shift: '+' },
      { value: '\\', shift: '|' }
    ],
    asdfRow: [
      { value: 'a', shift: 'A' },
      { value: 'o', shift: 'O' },
      { value: 'e', shift: 'E' },
      { value: 'u', shift: 'U' },
      { value: 'i', shift: 'I' },
      { value: 'd', shift: 'D' },
      { value: 'h', shift: 'H' },
      { value: 't', shift: 'T' },
      { value: 'n', shift: 'N' },
      { value: 's', shift: 'S' },
      { value: '-', shift: '_' }
    ],
    zxcvRow: [
      { value: ';', shift: ':' },
      { value: 'q', shift: 'Q' },
      { value: 'j', shift: 'J' },
      { value: 'k', shift: 'K' },
      { value: 'x', shift: 'X' },
      { value: 'b', shift: 'B' },
      { value: 'm', shift: 'M' },
      { value: 'w', shift: 'W' },
      { value: 'v', shift: 'V' },
      { value: 'z', shift: 'Z' }
    ],
    spaceRow: [
      { value: 'ctrl', label: 'CTRL', type: 'modifier', icon: '⌃' },
      { value: 'alt', label: 'ALT', type: 'modifier', icon: '⌥' },
      { value: 'meta', label: 'CMD', type: 'modifier', icon: '⌘' }
    ]
  }
]

// Computed properties
const currentLayout = computed(() => {
  return availableLayouts.find(layout => layout.id === selectedLayout.value) || availableLayouts[0]
})

// Methods
const getMainLabel = (key: Key | string): string => {
  if (typeof key === 'string') {
    return key.toUpperCase()
  }

  if (shiftPressed.value || capsLock.value) {
    return key.shift || key.value.toUpperCase()
  }

  return key.value
}

const getKeyClass = (key: Key): string => {
  const classes = []

  if (shiftPressed.value && key.shift) {
    classes.push('shifted')
  }

  if (capsLock.value && key.value.length === 1 && /[a-zA-Z]/.test(key.value)) {
    classes.push('caps-active')
  }

  return classes.join(' ')
}

const handleKeyPress = (key: Key) => {
  const label = getMainLabel(key)
  emit('keypress', label)
  playKeySound()
  showKeyFeedback(key, event)
}

const handleSpecialKey = (key: string) => {
  emit('specialKey', key)
  playKeySound()

  if (key === 'shift') {
    shiftPressed.value = !shiftPressed.value
  }
}

const handleCapsLock = () => {
  capsLock.value = !capsLock.value
  playKeySound()
}

const handleShift = () => {
  shiftPressed.value = !shiftPressed.value
  playKeySound()
}

const handleTouchStart = (key: Key | string) => {
  if (typeof key !== 'string') {
    showKeyFeedback(key, event)
  }
}

const handleTouchEnd = () => {
  pressedKey.value = null
}

const showKeyFeedback = (key: Key, event: Event) => {
  if (!event.target) return

  const element = event.target as HTMLElement
  const rect = element.getBoundingClientRect()

  pressedKey.value = key
  pressedKeyPosition.value = {
    x: rect.left + rect.width / 2,
    y: rect.top - 40
  }

  setTimeout(() => {
    pressedKey.value = null
  }, 200)
}

const playKeySound = () => {
  if (!soundEnabled.value) return

  // Create a simple beep sound using Web Audio API
  const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
  const oscillator = audioContext.createOscillator()
  const gainNode = audioContext.createGain()

  oscillator.connect(gainNode)
  gainNode.connect(audioContext.destination)

  oscillator.frequency.value = 800 + Math.random() * 400
  oscillator.type = 'sine'

  gainNode.gain.setValueAtTime(0.1, audioContext.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.1)

  oscillator.start(audioContext.currentTime)
  oscillator.stop(audioContext.currentTime + 0.1)
}

const changeLayout = () => {
  // Layout change logic
  console.log('Layout changed to:', selectedLayout.value)
}

const toggleKeyboard = () => {
  isVisible.value = !isVisible.value
}

const toggleMinimize = () => {
  isMinimized.value = !isMinimized.value
}

const toggleSettings = () => {
  showSettings.value = !showSettings.value
}

const show = () => {
  isVisible.value = true
  isMinimized.value = false
}

const hide = () => {
  isVisible.value = false
}

// Auto-hide functionality
let autoHideTimeout: NodeJS.Timeout

const resetAutoHideTimer = () => {
  if (!autoHide.value) return

  clearTimeout(autoHideTimeout)
  autoHideTimeout = setTimeout(() => {
    if (isVisible.value && !isMinimized.value) {
      isMinimized.value = true
    }
  }, 5000) // Auto-hide after 5 seconds of inactivity
}

// Public methods
defineExpose({
  show,
  hide,
  toggle: toggleKeyboard
})

// Lifecycle
onMounted(() => {
  // Listen for focus events on input elements
  document.addEventListener('focusin', (event) => {
    const target = event.target as HTMLElement
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.contentEditable === 'true') {
      show()
    }
  })

  // Listen for keyboard events to keep keyboard visible
  document.addEventListener('keydown', resetAutoHideTimer)
})

onUnmounted(() => {
  clearTimeout(autoHideTimeout)
})
</script>

<style scoped>
.virtual-keyboard {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(10, 10, 10, 0.95);
  border: 1px solid var(--surface-border);
  border-top: 2px solid var(--primary-500);
  backdrop-filter: blur(10px);
  transform: translateY(100%);
  transition: transform 0.3s ease;
  z-index: 1000;
  font-family: 'Fira Code', monospace;
}

.virtual-keyboard.visible {
  transform: translateY(0);
}

.virtual-keyboard.minimized {
  transform: translateY(calc(100% - 60px));
}

.keyboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  background: var(--surface);
  border-bottom: 1px solid var(--surface-border);
  cursor: pointer;
}

.keyboard-controls,
.keyboard-settings {
  display: flex;
  gap: 8px;
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.control-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.control-btn.close-btn:hover {
  border-color: var(--error);
  color: var(--error);
}

.btn-icon {
  font-size: 14px;
}

.keyboard-title {
  font-size: 12px;
  font-weight: bold;
  color: var(--primary-400);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.keyboard-settings-panel {
  padding: 15px;
  background: var(--surface-elevated);
  border-bottom: 1px solid var(--surface-border);
}

.settings-group {
  margin-bottom: 12px;
}

.settings-group:last-child {
  margin-bottom: 0;
}

.setting-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-primary);
  cursor: pointer;
}

.setting-select {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  color: var(--text-primary);
  font-family: inherit;
  font-size: 10px;
  padding: 4px 8px;
}

.setting-range {
  width: 80px;
  accent-color: var(--primary-500);
}

.size-value {
  color: var(--primary-400);
  font-weight: bold;
  min-width: 35px;
  text-align: right;
}

.keyboard-layout {
  padding: 15px;
}

.keyboard-row {
  display: flex;
  justify-content: center;
  gap: 4px;
  margin-bottom: 4px;
}

.key {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: v-bind('keySize + "px"');
  min-width: v-bind('keySize + "px"');
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 10px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.1s ease;
  position: relative;
  overflow: hidden;
  user-select: none;
  -webkit-user-select: none;
}

.key:hover {
  background: var(--surface-elevated);
  border-color: var(--primary-500);
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(14, 165, 233, 0.3);
}

.key:active,
.key.pressed {
  background: var(--primary-500);
  color: var(--text-primary);
  transform: translateY(0);
  box-shadow: 0 0 20px rgba(14, 165, 233, 0.5);
}

.key.active {
  background: rgba(14, 165, 233, 0.3);
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.key-main {
  font-size: 14px;
  font-weight: bold;
}

.key-shift {
  position: absolute;
  top: 2px;
  right: 4px;
  font-size: 8px;
  color: var(--text-secondary);
}

.key.shifted .key-main {
  display: none;
}

.key.shifted .key-shift {
  position: static;
  font-size: 14px;
  color: var(--text-primary);
}

.key.caps-active .key-main {
  text-transform: uppercase;
}

/* Special keys */
.key-special {
  min-width: v-bind('(keySize * 1.5) + "px"');
  background: var(--surface-elevated);
}

.key-special:hover {
  background: linear-gradient(135deg, var(--surface-elevated), var(--primary-500));
}

.key-backspace {
  min-width: v-bind('(keySize * 2) + "px"');
}

.key-tab {
  min-width: v-bind('(keySize * 1.8) + "px"');
}

.key-caps {
  min-width: v-bind('(keySize * 2.2) + "px"');
}

.key-enter {
  min-width: v-bind('(keySize * 2.5) + "px"');
}

.key-shift {
  min-width: v-bind('(keySize * 2.8) + "px"');
}

.key-space {
  min-width: v-bind('(keySize * 6) + "px"');
}

.key-icon {
  font-size: 16px;
  margin-bottom: 2px;
}

.key-label {
  font-size: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-secondary);
}

/* Key feedback animation */
.key-feedback {
  position: fixed;
  background: var(--primary-500);
  color: var(--text-primary);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  pointer-events: none;
  z-index: 2000;
  animation: keyFeedback 0.3s ease-out;
  box-shadow: 0 0 20px rgba(14, 165, 233, 0.8);
}

@keyframes keyFeedback {
  0% {
    transform: scale(0.8) translateY(0);
    opacity: 0;
  }
  50% {
    transform: scale(1.1) translateY(-5px);
    opacity: 1;
  }
  100% {
    transform: scale(1) translateY(-10px);
    opacity: 0;
  }
}

/* Glitch effect for title */
.glitch {
  position: relative;
  color: var(--primary-400);
  font-size: 14px;
  font-weight: bold;
  text-transform: uppercase;
  text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  animation: glitch 2s infinite;
}

@keyframes glitch {
  0%, 90%, 100% {
    text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  }
  95% {
    text-shadow: -2px 2px 0 var(--accent-500), 2px -2px 0 var(--primary-600);
  }
}

/* Responsive design */
@media (max-width: 768px) {
  .keyboard-layout {
    padding: 10px;
  }

  .keyboard-row {
    gap: 2px;
    margin-bottom: 2px;
  }

  .key {
    min-height: 35px;
    min-width: 35px;
    font-size: 9px;
  }

  .key-main {
    font-size: 12px;
  }

  .key-icon {
    font-size: 14px;
  }

  .key-label {
    font-size: 7px;
  }

  .key-backspace,
  .key-tab,
  .key-caps,
  .key-enter,
  .key-shift,
  .key-space {
    min-width: auto;
  }

  .key-space {
    min-width: 150px;
  }
}

@media (max-width: 480px) {
  .key {
    min-height: 30px;
    min-width: 30px;
    font-size: 8px;
  }

  .key-main {
    font-size: 10px;
  }

  .key-space {
    min-width: 120px;
  }
}

/* Touch-specific optimizations */
@media (pointer: coarse) {
  .key {
    min-height: v-bind('(keySize * 1.2) + "px"');
    min-width: v-bind('(keySize * 1.2) + "px"');
  }

  .keyboard-row {
    gap: 6px;
    margin-bottom: 6px;
  }
}
</style>