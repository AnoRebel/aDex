<template>
  <div class="edex-keyboard" @mousedown.prevent>
    <div
      v-for="(row, rowIndex) in layoutRows"
      :key="rowIndex"
      class="kb-row"
    >
      <button
        v-for="(keyData, keyIndex) in row"
        :key="`${rowIndex}-${keyIndex}-${keyData.key}`"
        class="kb-key"
        :class="getKeyClasses(keyData, rowIndex, keyIndex)"
        :style="getKeyStyle(keyData)"
        @mousedown.prevent="onKeyDown(keyData)"
        @mouseup.prevent="onKeyUp(keyData)"
        @mouseleave="onKeyUp(keyData)"
        :title="getKeyTitle(keyData)"
      >
        <span class="kb-key-label">{{ getDisplayLabel(keyData) }}</span>
        <span
          v-if="keyData.shift && !isModifier(keyData) && !isSpecialKey(keyData)"
          class="kb-key-shift"
        >{{ keyData.shift }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useTerminalStore } from '~/stores/terminal'

// Types matching the keyboard layout JSON structure
interface KeyData {
  key: string
  shift?: string
  width?: number
  type?: 'standard' | 'function' | 'special' | 'modifier' | 'space' | 'number' | 'operator'
  side?: 'left' | 'right'
}

interface KeyboardLayout {
  name: string
  displayName: string
  type: string
  rows: KeyData[][]
}

const terminalStore = useTerminalStore()

// State
const layouts = ref<Record<string, KeyboardLayout>>({})
const currentLayoutName = ref('qwerty')
const pressedKeys = ref<Set<string>>(new Set())
const physicalPressedKeys = ref<Set<string>>(new Set())

// Modifier states
const shiftActive = ref(false)
const ctrlActive = ref(false)
const altActive = ref(false)
const capsLockActive = ref(false)
const fnActive = ref(false)

// Base key unit width in vw for proportional sizing
const KEY_UNIT_VW = 2.7

// Computed
const currentLayout = computed<KeyboardLayout | null>(() => {
  return layouts.value[currentLayoutName.value] || null
})

const layoutRows = computed<KeyData[][]>(() => {
  return currentLayout.value?.rows || []
})

// Helper: check if key is a modifier
const isModifier = (keyData: KeyData): boolean => {
  return keyData.type === 'modifier' ||
    ['Shift', 'Ctrl', 'Alt', 'Meta', 'CapsLock', 'Fn'].includes(keyData.key)
}

// Helper: check if key is a special (non-printable) key
const isSpecialKey = (keyData: KeyData): boolean => {
  return keyData.type === 'special' || keyData.type === 'function' ||
    ['Backspace', 'Tab', 'Enter', 'Escape', 'CapsLock', 'Delete',
     'ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight',
     'Home', 'End', 'PageUp', 'PageDown', 'Insert', 'NumLock'].includes(keyData.key)
}

// Get the display label for a key
const getDisplayLabel = (keyData: KeyData): string => {
  // If shift or capslock is active and it is a letter or has a shift mapping
  if (keyData.type === 'space') return ''
  if (isModifier(keyData)) return keyData.key
  if (isSpecialKey(keyData)) {
    // Abbreviations for special keys
    const abbreviations: Record<string, string> = {
      'Backspace': 'BKSP',
      'CapsLock': 'CAPS',
      'Escape': 'ESC',
      'Enter': 'RET',
      'Delete': 'DEL',
      'NumLock': 'NUM',
    }
    return abbreviations[keyData.key] || keyData.key.toUpperCase()
  }
  if (keyData.type === 'function') return keyData.key

  // Regular key: show the shifted version if shift/caps is active
  const useShift = shiftActive.value !== capsLockActive.value
  if (useShift && keyData.shift) return keyData.shift
  return keyData.key
}

// Get key title tooltip
const getKeyTitle = (keyData: KeyData): string => {
  if (keyData.shift) return `${keyData.key} / ${keyData.shift}`
  return keyData.key
}

// Get CSS classes for a key
const getKeyClasses = (keyData: KeyData, rowIndex: number, keyIndex: number): Record<string, boolean> => {
  const keyId = getKeyId(keyData, rowIndex, keyIndex)
  const isPressed = pressedKeys.value.has(keyId) || physicalPressedKeys.value.has(normalizeKeyName(keyData.key))

  return {
    'pressed': isPressed,
    'active': isPressed,
    'fn-key': keyData.type === 'function',
    'spacebar': keyData.type === 'space',
    'modifier-key': isModifier(keyData),
    'special-key': isSpecialKey(keyData),
    'modifier-active': isModifierActive(keyData),
    [`wide-${getWidthClass(keyData.width || 1)}`]: (keyData.width || 1) > 1,
  }
}

// Map width values to CSS width classes
const getWidthClass = (width: number): string => {
  if (width <= 1) return '0'
  if (width <= 1.5) return '1'
  if (width <= 2) return '2'
  if (width <= 2.5) return '2'
  if (width <= 3) return '3'
  return '3'
}

// Get inline style for key width
const getKeyStyle = (keyData: KeyData): Record<string, string> => {
  const width = keyData.width || 1
  if (keyData.type === 'space') {
    return { width: `${width * KEY_UNIT_VW}vw` }
  }
  if (width !== 1) {
    return { width: `${width * KEY_UNIT_VW}vw` }
  }
  return {}
}

// Check if a modifier key is currently active
const isModifierActive = (keyData: KeyData): boolean => {
  const key = keyData.key.toLowerCase()
  if (key === 'shift') return shiftActive.value
  if (key === 'ctrl') return ctrlActive.value
  if (key === 'alt') return altActive.value
  if (key === 'capslock') return capsLockActive.value
  if (key === 'fn') return fnActive.value
  return false
}

// Generate a unique ID for each key button
const getKeyId = (keyData: KeyData, rowIndex: number, keyIndex: number): string => {
  const side = keyData.side ? `-${keyData.side}` : ''
  return `${keyData.key}${side}-${rowIndex}-${keyIndex}`
}

// Normalize a key name for matching physical keyboard events
const normalizeKeyName = (key: string): string => {
  const map: Record<string, string> = {
    'Escape': 'escape',
    'Backspace': 'backspace',
    'Tab': 'tab',
    'Enter': 'enter',
    'CapsLock': 'capslock',
    'Shift': 'shift',
    'Ctrl': 'control',
    'Alt': 'alt',
    'Space': ' ',
    'Delete': 'delete',
    'NumLock': 'numlock',
    'Meta': 'meta',
  }
  return map[key] || key.toLowerCase()
}

// Handle on-screen key press
const onKeyDown = (keyData: KeyData) => {
  // Handle modifier toggle
  if (isModifier(keyData)) {
    toggleModifier(keyData.key)
    return
  }

  // Build the character to send
  const char = resolveKeyChar(keyData)

  // Send to active terminal
  sendToTerminal(char, keyData)
}

const onKeyUp = (keyData: KeyData) => {
  // Reset shift after a non-modifier key press (one-shot shift)
  if (!isModifier(keyData) && shiftActive.value && !capsLockActive.value) {
    shiftActive.value = false
  }
}

// Resolve the character to send based on current modifier state
const resolveKeyChar = (keyData: KeyData): string => {
  if (keyData.type === 'space') return ' '

  // Special keys send escape sequences
  const specialMap: Record<string, string> = {
    'Escape': '\x1b',
    'Backspace': '\x7f',
    'Tab': '\t',
    'Enter': '\r',
    'Delete': '\x1b[3~',
    'ArrowUp': '\x1b[A',
    'ArrowDown': '\x1b[B',
    'ArrowRight': '\x1b[C',
    'ArrowLeft': '\x1b[D',
    'Home': '\x1b[H',
    'End': '\x1b[F',
    'PageUp': '\x1b[5~',
    'PageDown': '\x1b[6~',
  }

  if (specialMap[keyData.key]) return specialMap[keyData.key]

  // Function keys
  if (keyData.type === 'function') {
    const fNum = parseInt(keyData.key.replace('F', ''), 10)
    if (fNum >= 1 && fNum <= 4) return `\x1bO${String.fromCharCode(79 + fNum)}`
    if (fNum >= 5 && fNum <= 12) {
      const codes = [15, 17, 18, 19, 20, 21, 23, 24]
      return `\x1b[${codes[fNum - 5]}~`
    }
    return ''
  }

  // Regular characters
  const useShift = shiftActive.value !== capsLockActive.value
  let char = useShift && keyData.shift ? keyData.shift : keyData.key

  // Ctrl modifier: send control character
  if (ctrlActive.value && char.length === 1) {
    const code = char.toUpperCase().charCodeAt(0)
    if (code >= 65 && code <= 90) {
      // Ctrl+A through Ctrl+Z
      return String.fromCharCode(code - 64)
    }
    // Ctrl+[ = ESC, Ctrl+\ = 0x1c, etc.
    if (char === '[') return '\x1b'
    if (char === '\\') return '\x1c'
    if (char === ']') return '\x1d'
  }

  // Alt modifier: prefix with ESC
  if (altActive.value && char.length === 1) {
    return `\x1b${char}`
  }

  return char
}

// Toggle modifier keys
const toggleModifier = (key: string) => {
  switch (key.toLowerCase()) {
    case 'shift':
      shiftActive.value = !shiftActive.value
      break
    case 'ctrl':
      ctrlActive.value = !ctrlActive.value
      break
    case 'alt':
      altActive.value = !altActive.value
      break
    case 'capslock':
      capsLockActive.value = !capsLockActive.value
      break
    case 'fn':
      fnActive.value = !fnActive.value
      break
  }
}

// Send the resolved character to the active terminal
const sendToTerminal = (char: string, keyData: KeyData) => {
  if (!char) return

  try {
    terminalStore.sendInput(char)
  } catch (err) {
    console.error('Failed to send key to terminal:', err)
  }

  // Reset one-shot modifiers (ctrl and alt reset after use, shift handled in onKeyUp)
  if (ctrlActive.value) ctrlActive.value = false
  if (altActive.value) altActive.value = false
}

// --- Physical keyboard event handling ---
// Highlight on-screen keys when physical keys are pressed

const handlePhysicalKeyDown = (e: KeyboardEvent) => {
  const key = e.key.toLowerCase()
  physicalPressedKeys.value.add(key)

  // Sync modifier state
  shiftActive.value = e.shiftKey
  ctrlActive.value = e.ctrlKey
  altActive.value = e.altKey

  if (e.key === 'CapsLock') {
    capsLockActive.value = !capsLockActive.value
  }
}

const handlePhysicalKeyUp = (e: KeyboardEvent) => {
  const key = e.key.toLowerCase()
  physicalPressedKeys.value.delete(key)

  // Sync modifier state
  if (!e.shiftKey) shiftActive.value = false
  if (!e.ctrlKey) ctrlActive.value = false
  if (!e.altKey) altActive.value = false
}

// --- Load keyboard layouts ---
const loadLayouts = async () => {
  try {
    // Import the keyboard layouts JSON as a static asset
    const data = await import('~/assets/data/keyboard-layouts.json')
    if (data.layouts) {
      layouts.value = data.layouts
    } else if (data.default?.layouts) {
      layouts.value = data.default.layouts
    }

    // Load saved layout preference
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('adex-keyboard-settings')
      if (saved) {
        try {
          const parsed = JSON.parse(saved)
          if (parsed.currentLayout && layouts.value[parsed.currentLayout]) {
            currentLayoutName.value = parsed.currentLayout
          }
        } catch {
          // ignore parse errors
        }
      }
    }
  } catch (err) {
    console.error('Failed to load keyboard layouts:', err)
    // Fallback minimal layout
    layouts.value = {
      qwerty: {
        name: 'QWERTY',
        displayName: 'QWERTY (US)',
        type: 'standard',
        rows: [
          [
            { key: 'Escape', width: 1, type: 'function' },
            { key: 'F1', width: 1, type: 'function' },
            { key: 'F2', width: 1, type: 'function' },
            { key: 'F3', width: 1, type: 'function' },
            { key: 'F4', width: 1, type: 'function' },
            { key: 'F5', width: 1, type: 'function' },
            { key: 'F6', width: 1, type: 'function' },
            { key: 'F7', width: 1, type: 'function' },
            { key: 'F8', width: 1, type: 'function' },
            { key: 'F9', width: 1, type: 'function' },
            { key: 'F10', width: 1, type: 'function' },
            { key: 'F11', width: 1, type: 'function' },
            { key: 'F12', width: 1, type: 'function' }
          ],
          [
            { key: '`', shift: '~', width: 1 },
            { key: '1', shift: '!', width: 1 },
            { key: '2', shift: '@', width: 1 },
            { key: '3', shift: '#', width: 1 },
            { key: '4', shift: '$', width: 1 },
            { key: '5', shift: '%', width: 1 },
            { key: '6', shift: '^', width: 1 },
            { key: '7', shift: '&', width: 1 },
            { key: '8', shift: '*', width: 1 },
            { key: '9', shift: '(', width: 1 },
            { key: '0', shift: ')', width: 1 },
            { key: '-', shift: '_', width: 1 },
            { key: '=', shift: '+', width: 1 },
            { key: 'Backspace', width: 2, type: 'special' }
          ],
          [
            { key: 'Tab', width: 1.5, type: 'special' },
            { key: 'q', shift: 'Q', width: 1 },
            { key: 'w', shift: 'W', width: 1 },
            { key: 'e', shift: 'E', width: 1 },
            { key: 'r', shift: 'R', width: 1 },
            { key: 't', shift: 'T', width: 1 },
            { key: 'y', shift: 'Y', width: 1 },
            { key: 'u', shift: 'U', width: 1 },
            { key: 'i', shift: 'I', width: 1 },
            { key: 'o', shift: 'O', width: 1 },
            { key: 'p', shift: 'P', width: 1 },
            { key: '[', shift: '{', width: 1 },
            { key: ']', shift: '}', width: 1 },
            { key: '\\', shift: '|', width: 1.5 }
          ],
          [
            { key: 'CapsLock', width: 1.75, type: 'special' },
            { key: 'a', shift: 'A', width: 1 },
            { key: 's', shift: 'S', width: 1 },
            { key: 'd', shift: 'D', width: 1 },
            { key: 'f', shift: 'F', width: 1 },
            { key: 'g', shift: 'G', width: 1 },
            { key: 'h', shift: 'H', width: 1 },
            { key: 'j', shift: 'J', width: 1 },
            { key: 'k', shift: 'K', width: 1 },
            { key: 'l', shift: 'L', width: 1 },
            { key: ';', shift: ':', width: 1 },
            { key: "'", shift: '"', width: 1 },
            { key: 'Enter', width: 2.25, type: 'special' }
          ],
          [
            { key: 'Shift', width: 2.25, type: 'modifier', side: 'left' },
            { key: 'z', shift: 'Z', width: 1 },
            { key: 'x', shift: 'X', width: 1 },
            { key: 'c', shift: 'C', width: 1 },
            { key: 'v', shift: 'V', width: 1 },
            { key: 'b', shift: 'B', width: 1 },
            { key: 'n', shift: 'N', width: 1 },
            { key: 'm', shift: 'M', width: 1 },
            { key: ',', shift: '<', width: 1 },
            { key: '.', shift: '>', width: 1 },
            { key: '/', shift: '?', width: 1 },
            { key: 'Shift', width: 2.75, type: 'modifier', side: 'right' }
          ],
          [
            { key: 'Ctrl', width: 1.5, type: 'modifier', side: 'left' },
            { key: 'Alt', width: 1.5, type: 'modifier', side: 'left' },
            { key: 'Space', width: 7, type: 'space' },
            { key: 'Alt', width: 1.5, type: 'modifier', side: 'right' },
            { key: 'Ctrl', width: 1.5, type: 'modifier', side: 'right' }
          ]
        ]
      }
    }
  }
}

// Lifecycle
onMounted(() => {
  loadLayouts()
  document.addEventListener('keydown', handlePhysicalKeyDown)
  document.addEventListener('keyup', handlePhysicalKeyUp)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handlePhysicalKeyDown)
  document.removeEventListener('keyup', handlePhysicalKeyUp)
})
</script>

<style scoped>
/* The parent CSS classes (edex-keyboard, kb-row, kb-key) are defined in main.css.
   Below are component-specific additions. */

.edex-keyboard {
  user-select: none;
  -webkit-user-select: none;
}

.kb-key {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0;
  overflow: hidden;
}

/* Key press animation: fill background with accent color */
.kb-key::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.6);
  transform: scaleY(0);
  transform-origin: bottom;
  transition: transform 0.08s ease-out;
  z-index: 0;
}

.kb-key.pressed::before,
.kb-key.active::before {
  transform: scaleY(1);
}

.kb-key.pressed,
.kb-key.active {
  color: var(--color_black);
}

/* Modifier active state (stays lit) */
.kb-key.modifier-active {
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.35);
  border-color: var(--color_accent);
}

.kb-key-label {
  position: relative;
  z-index: 1;
  font-size: 1.2vh;
  line-height: 1;
}

.kb-key-shift {
  position: relative;
  z-index: 1;
  font-size: 0.8vh;
  opacity: 0.4;
  line-height: 1;
  margin-top: 0.1vh;
}

/* Show shift char more prominently when shift/caps is active */
.kb-key:hover .kb-key-shift,
.kb-key.modifier-active .kb-key-shift {
  opacity: 0.8;
}

/* Function keys styling */
.kb-key.fn-key {
  font-size: 0.9vh;
}

.kb-key.fn-key .kb-key-label {
  font-size: 0.9vh;
}

/* Spacebar */
.kb-key.spacebar {
  min-width: 18vw;
}

.kb-key.spacebar .kb-key-label {
  display: none;
}

/* Special keys */
.kb-key.special-key .kb-key-label {
  font-size: 1vh;
  letter-spacing: 0.05em;
}

/* Modifier keys */
.kb-key.modifier-key .kb-key-label {
  font-size: 1vh;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

/* Width overrides via inline style handle most sizing,
   but provide fallback classes for common widths */
.kb-key.wide-1 { min-width: 3.8vw; }
.kb-key.wide-2 { min-width: 5vw; }
.kb-key.wide-3 { min-width: 7vw; }
</style>
