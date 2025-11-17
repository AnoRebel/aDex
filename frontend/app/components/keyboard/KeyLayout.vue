<template>
  <div
    :class="layoutClasses"
    :style="layoutStyles"
    role="application"
    :aria-label="`Keyboard layout: ${layoutData?.displayName}`"
  >
    <!-- Function row (F1-F12) -->
    <div v-if="showFunctionRow" class="keyboard-row keyboard-row--function">
      <KeyButton
        v-for="keyData in functionRow"
        :key="keyData.key"
        :key-data="keyData"
        :is-shift-active="isShiftActive"
        :is-caps-lock-active="isCapsLockActive"
        :size="size"
        :theme="theme"
        :disabled="disabled"
        @key-press="handleKeyPress"
        @key-release="handleKeyRelease"
        @long-press="handleLongPress"
      />
    </div>

    <!-- Main keyboard rows -->
    <div
      v-for="(row, rowIndex) in layoutRows"
      :key="`row-${rowIndex}`"
      class="keyboard-row"
      :class="`keyboard-row--${rowIndex}`"
    >
      <KeyButton
        v-for="keyData in row"
        :key="`${keyData.key}-${rowIndex}`"
        :key-data="keyData"
        :is-shift-active="isShiftActive"
        :is-caps-lock-active="isCapsLockActive"
        :size="size"
        :theme="theme"
        :disabled="disabled"
        @key-press="handleKeyPress"
        @key-release="handleKeyRelease"
        @long-press="handleLongPress"
      />
    </div>

    <!-- Number pad (optional) -->
    <div v-if="showNumpad" class="keyboard-numpad">
      <div
        v-for="(row, rowIndex) in numpadRows"
        :key="`numpad-row-${rowIndex}`"
        class="keyboard-row keyboard-row--numpad"
      >
        <KeyButton
          v-for="keyData in row"
          :key="`numpad-${keyData.key}-${rowIndex}`"
          :key-data="keyData"
          :is-shift-active="isShiftActive"
          :is-caps-lock-active="isCapsLockActive"
          :size="size"
          :theme="theme"
          :disabled="disabled"
          @key-press="handleKeyPress"
          @key-release="handleKeyRelease"
          @long-press="handleLongPress"
        />
      </div>
    </div>

    <!-- Layout selector -->
    <div v-if="showLayoutSelector" class="keyboard-controls">
      <select
        v-model="selectedLayout"
        :aria-label="Select keyboard layout"
        class="layout-selector"
        :disabled="disabled"
      >
        <option
          v-for="layout in availableLayouts"
          :key="layout.name"
          :value="layout.name"
        >
          {{ layout.displayName }}
        </option>
      </select>

      <!-- Theme selector -->
      <select
        v-model="selectedTheme"
        :aria-label="Select keyboard theme"
        class="theme-selector"
        :disabled="disabled"
      >
        <option value="dark">Dark</option>
        <option value="light">Light</option>
        <option value="cyberpunk">Cyberpunk</option>
      </select>

      <!-- Size selector -->
      <select
        v-model="selectedSize"
        :aria-label="Select keyboard size"
        class="size-selector"
        :disabled="disabled"
      >
        <option value="small">Small</option>
        <option value="medium">Medium</option>
        <option value="large">Large</option>
      </select>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useKeyboard } from '~/composables/useKeyboard'
import KeyButton from './KeyButton.vue'
import type { KeyData } from '~/composables/useKeyboard'

interface Props {
  layoutName?: string
  showFunctionRow?: boolean
  showNumpad?: boolean
  showLayoutSelector?: boolean
  size?: 'small' | 'medium' | 'large'
  theme?: 'dark' | 'light' | 'cyberpunk'
  disabled?: boolean
  opacity?: number
  maximizeWidth?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showFunctionRow: false,
  showNumpad: false,
  showLayoutSelector: false,
  size: 'medium',
  theme: 'dark',
  disabled: false,
  opacity: 0.9,
  maximizeWidth: false
})

const emit = defineEmits<{
  keyPress: [key: string, modifiers: { shift: boolean; ctrl: boolean; alt: boolean; caps: boolean }]
  keyRelease: [key: string]
  longPress: [key: string]
  layoutChange: [layoutName: string]
}>()

// Use keyboard composable
const {
  currentLayout,
  currentLayoutData,
  availableLayouts,
  settings,
  isShiftActive,
  isCapsLockActive,
  isCtrlActive,
  isAltActive,
  setLayout,
  handleModifierPress,
  handleModifierRelease,
  triggerHapticFeedback
} = useKeyboard()

// Internal state
const selectedLayout = ref(props.layoutName || currentLayout.value)
const selectedTheme = ref(props.theme)
const selectedSize = ref(props.size)
const numpadLayout = ref<any>(null)

// Load numpad layout
const loadNumpadLayout = async () => {
  try {
    const response = await fetch('/app/assets/data/keyboard-layouts.json')
    if (response.ok) {
      const data = await response.json()
      numpadLayout.value = data.layouts?.numpad || null
    }
  } catch (error) {
    console.warn('Could not load numpad layout:', error)
  }
}

// Computed properties
const layoutData = computed(() => {
  if (selectedLayout.value === currentLayout.value) {
    return currentLayoutData.value
  }

  // Try to find layout in available layouts
  const found = availableLayouts.value.find(l => l.name === selectedLayout.value)
  return found || null
})

const layoutRows = computed(() => {
  if (!layoutData.value?.rows) return []

  // Skip function row if it's shown separately
  const startIndex = props.showFunctionRow && layoutData.value?.rows?.[0]?.[0]?.type === 'function' ? 1 : 0
  return layoutData.value.rows.slice(startIndex)
})

const functionRow = computed(() => {
  if (!props.showFunctionRow || !layoutData.value?.rows) return []

  const firstRow = layoutData.value.rows[0]
  return firstRow?.[0]?.type === 'function' ? firstRow : []
})

const numpadRows = computed(() => {
  return numpadLayout.value?.rows || []
})

const layoutClasses = computed(() => [
  'keyboard-layout',
  `keyboard-layout--${selectedSize.value}`,
  `keyboard-layout--${selectedTheme.value}`,
  {
    'keyboard-layout--disabled': props.disabled,
    'keyboard-layout--max-width': props.maximizeWidth,
    'keyboard-layout--with-numpad': props.showNumpad,
    'keyboard-layout--with-function-row': props.showFunctionRow,
    'keyboard-layout--with-controls': props.showLayoutSelector
  }
])

const layoutStyles = computed(() => ({
  '--keyboard-opacity': props.opacity,
  '--keyboard-scale': selectedSize.value === 'small' ? 0.85 : selectedSize.value === 'large' ? 1.15 : 1
}))

// Methods
const handleKeyPress = (key: string, shiftActive: boolean) => {
  const keyData = getKeyData(key)

  if (keyData?.type === 'modifier') {
    handleModifierPress(key)
    triggerHapticFeedback()
  } else {
    // Calculate current modifier state
    const modifiers = {
      shift: shiftActive || isShiftActive.value,
      ctrl: isCtrlActive.value,
      alt: isAltActive.value,
      caps: isCapsLockActive.value
    }

    emit('keyPress', key, modifiers)
    triggerHapticFeedback(3) // Short vibration for regular keys
  }
}

const handleKeyRelease = (key: string) => {
  const keyData = getKeyData(key)

  if (keyData?.type === 'modifier') {
    handleModifierRelease(key)
  }

  emit('keyRelease', key)
}

const handleLongPress = (key: string) => {
  emit('longPress', key)
  triggerHapticFeedback([10, 50, 10]) // Double vibration pattern
}

const getKeyData = (key: string): KeyData | null => {
  if (!layoutData.value?.rows) return null

  for (const row of layoutData.value.rows) {
    const keyData = row.find(k => k.key === key)
    if (keyData) return keyData
  }

  return null
}

// Watchers
watch(selectedLayout, (newLayout) => {
  setLayout(newLayout)
  emit('layoutChange', newLayout)
})

watch(() => props.layoutName, (newLayout) => {
  if (newLayout && newLayout !== selectedLayout.value) {
    selectedLayout.value = newLayout
  }
})

watch(() => props.theme, (newTheme) => {
  selectedTheme.value = newTheme
})

watch(() => props.size, (newSize) => {
  selectedSize.value = newSize
})

// Lifecycle
onMounted(() => {
  loadNumpadLayout()
})
</script>

<style scoped>
.keyboard-layout {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.5rem;
  background: var(--dex-theme-background-primary, #0a0a0a);
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-large, 1rem);
  backdrop-filter: blur(10px);
  opacity: var(--keyboard-opacity);
  transform: scale(var(--keyboard-scale));
  transition: all 0.3s ease;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  max-width: 100%;
  margin: 0 auto;
}

.keyboard-layout--small {
  gap: 0.125rem;
  padding: 0.25rem;
}

.keyboard-layout--large {
  gap: 0.375rem;
  padding: 0.75rem;
}

.keyboard-layout--max-width {
  max-width: 800px;
}

.keyboard-layout--disabled {
  pointer-events: none;
  opacity: 0.5;
  filter: grayscale(50%);
}

/* Theme variants */
.keyboard-layout--light {
  background: rgba(255, 255, 255, 0.95);
  border-color: #e0e0e0;
  color: #000000;
}

.keyboard-layout--cyberpunk {
  background: rgba(10, 10, 15, 0.95);
  border-color: #ff00ff;
  box-shadow: 0 0 30px rgba(255, 0, 255, 0.2), 0 4px 20px rgba(0, 0, 0, 0.5);
}

.keyboard-layout--cyberpunk::before {
  content: '';
  position: absolute;
  top: -1px;
  left: -1px;
  right: -1px;
  bottom: -1px;
  background: linear-gradient(45deg, #ff00ff, #00ffff, #ff00ff);
  border-radius: inherit;
  opacity: 0.1;
  z-index: -1;
  animation: cyberpunkGlow 3s ease-in-out infinite;
}

@keyframes cyberpunkGlow {
  0%, 100% { opacity: 0.1; }
  50% { opacity: 0.3; }
}

/* Keyboard rows */
.keyboard-row {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.125rem;
  flex-wrap: nowrap;
  width: 100%;
}

.keyboard-row--function {
  padding: 0 1rem;
}

.keyboard-row--numpad {
  flex-direction: column;
  gap: 0.25rem;
}

/* Special row adjustments */
.keyboard-row--0 {
  justify-content: center;
}

.keyboard-row--1 {
  justify-content: flex-start;
}

/* Controls section */
.keyboard-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border-top: 1px solid var(--dex-theme-border, #333);
  flex-wrap: wrap;
}

.layout-selector,
.theme-selector,
.size-selector {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--dex-theme-border, #333);
  border-radius: var(--dex-theme-border-radius-medium, 0.375rem);
  background: var(--dex-theme-input-background, #1a1a1a);
  color: var(--dex-theme-input-foreground, #ffffff);
  font-family: var(--dex-theme-font-family-secondary, 'Inter', sans-serif);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.layout-selector:hover,
.theme-selector:hover,
.size-selector:hover:not(:disabled) {
  background: var(--dex-theme-input-border, #2a2a2a);
  border-color: var(--dex-theme-input-focus, #00ff41);
}

.layout-selector:focus,
.theme-selector:focus,
.size-selector:focus {
  outline: 2px solid var(--dex-theme-input-focus, #00ff41);
  outline-offset: 2px;
}

.layout-selector:disabled,
.theme-selector:disabled,
.size-selector:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Numpad layout */
.keyboard-numpad {
  display: flex;
  margin-left: 1rem;
  padding-left: 1rem;
  border-left: 1px solid var(--dex-theme-border, #333);
}

/* Responsive adjustments */
@media (max-width: 1024px) {
  .keyboard-layout {
    padding: 0.375rem;
    gap: 0.1875rem;
  }

  .keyboard-layout--with-numpad {
    flex-direction: column;
  }

  .keyboard-numpad {
    margin-left: 0;
    margin-top: 0.5rem;
    padding-left: 0;
    border-left: none;
    border-top: 1px solid var(--dex-theme-border, #333);
    justify-content: center;
  }

  .keyboard-numpad .keyboard-row {
    flex-direction: row;
  }
}

@media (max-width: 768px) {
  .keyboard-layout {
    padding: 0.25rem;
    gap: 0.125rem;
    border-radius: var(--dex-theme-border-radius-medium, 0.375rem);
  }

  .keyboard-controls {
    padding: 0.375rem;
    gap: 0.5rem;
  }

  .layout-selector,
  .theme-selector,
  .size-selector {
    padding: 0.25rem 0.5rem;
    font-size: 0.625rem;
  }
}

@media (max-width: 480px) {
  .keyboard-layout {
    width: 100%;
    border-radius: var(--dex-theme-border-radius-medium, 0.375rem) var(--dex-theme-border-radius-medium, 0.375rem) 0 0;
    border-left: none;
    border-right: none;
    border-bottom: none;
  }

  .keyboard-row {
    gap: 0.0625rem;
  }
}

/* Accessibility */
.keyboard-layout:focus-within {
  outline: 2px solid var(--dex-theme-input-focus, #00ff41);
  outline-offset: 2px;
}

/* Animation for layout changes */
.keyboard-layout {
  animation: layoutSlideIn 0.3s ease-out;
}

@keyframes layoutSlideIn {
  from {
    transform: translateY(20px) scale(var(--keyboard-scale));
    opacity: 0;
  }
  to {
    transform: translateY(0) scale(var(--keyboard-scale));
    opacity: var(--keyboard-opacity);
  }
}

/* Print styles */
@media print {
  .keyboard-layout {
    display: none;
  }
}
</style>