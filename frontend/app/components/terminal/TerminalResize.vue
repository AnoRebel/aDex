<template>
  <div
    ref="resizeContainer"
    class="terminal-resize"
    :class="{
      'terminal-resize--active': isResizing,
      'terminal-resize--horizontal': direction === 'horizontal',
      'terminal-resize--vertical': direction === 'vertical',
      'terminal-resize--both': direction === 'both'
    }"
    @mousedown="handleMouseDown"
    @touchstart="handleTouchStart"
    @contextmenu.prevent
  >
    <!-- Resize Handle -->
    <div
      class="terminal-resize__handle"
      :class="`terminal-resize__handle--${direction}`"
    >
      <div class="terminal-resize__handle-visual">
        <!-- Dots for resize handle visual -->
        <div
          v-for="i in handleDots"
          :key="i"
          class="terminal-resize__handle-dot"
        ></div>
      </div>
    </div>

    <!-- Resize Overlay (shown during resize) -->
    <div
      v-if="isResizing && showOverlay"
      class="terminal-resize__overlay"
      :style="overlayStyle"
    >
      <div class="terminal-resize__overlay-content">
        <div class="terminal-resize__overlay-size">
          {{ currentSize.width }} × {{ currentSize.height }}
        </div>
        <div class="terminal-resize__overlay-grid">
          {{ gridInfo.cols }} × {{ gridInfo.rows }}
        </div>
        <div class="terminal-resize__overlay-position">
          {{ Math.round(currentPosition.x) }}, {{ Math.round(currentPosition.y) }}
        </div>
      </div>
    </div>

    <!-- Resize Ghost (shows new size during resize) -->
    <div
      v-if="isResizing && showGhost"
      class="terminal-resize__ghost"
      :style="ghostStyle"
    ></div>

    <!-- Minimum Size Indicator -->
    <div
      v-if="showMinSizeIndicator && isNearMinimum"
      class="terminal-resize__min-size-indicator"
    >
      <Icon name="carbon:warning-alt" />
      <span>Minimum size: {{ minSize.width }}×{{ minSize.height }}</span>
    </div>

    <!-- Snap Indicators -->
    <div
      v-if="isResizing && snapIndicators.length > 0"
      class="terminal-resize__snap-indicators"
    >
      <div
        v-for="(indicator, index) in snapIndicators"
        :key="index"
        class="terminal-resize__snap-indicator"
        :style="indicator.style"
        :class="`terminal-resize__snap-indicator--${indicator.type}`"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'

// Types
interface ResizeSize {
  width: number
  height: number
}

interface ResizePosition {
  x: number
  y: number
}

interface GridInfo {
  cols: number
  rows: number
}

interface SnapIndicator {
  type: 'horizontal' | 'vertical'
  style: Record<string, string>
}

interface ResizeConstraints {
  minWidth?: number
  minHeight?: number
  maxWidth?: number
  maxHeight?: number
  aspectRatio?: number
  preserveAspectRatio?: boolean
}

// Props
interface Props {
  direction?: 'horizontal' | 'vertical' | 'both'
  initialSize?: ResizeSize
  minSize?: ResizeSize
  maxSize?: ResizeSize
  constraints?: ResizeConstraints
  step?: number
  snapToGrid?: boolean
  gridSize?: number
  showOverlay?: boolean
  showGhost?: boolean
  showMinSizeIndicator?: boolean
  enableKeyboard?: boolean
  debounceMs?: number
  onResizeStart?: () => void
  onResize?: (size: ResizeSize) => void
  onResizeEnd?: (size: ResizeSize) => void
}

const props = withDefaults(defineProps<Props>(), {
  direction: 'both',
  initialSize: () => ({ width: 600, height: 400 }),
  minSize: () => ({ width: 300, height: 200 }),
  maxSize: () => ({ width: 1920, height: 1080 }),
  step: 1,
  snapToGrid: false,
  gridSize: 10,
  showOverlay: true,
  showGhost: false,
  showMinSizeIndicator: true,
  enableKeyboard: true,
  debounceMs: 16
})

// Emits
const emit = defineEmits<{
  resizeStart: [event: MouseEvent | TouchEvent, size: ResizeSize]
  resize: [size: ResizeSize, position: ResizePosition]
  resizeEnd: [size: ResizeSize]
  snap: [position: ResizePosition]
  minSizeReached: [size: ResizeSize]
  maxSizeReached: [size: ResizeSize]
}>()

// Refs
const resizeContainer = ref<HTMLElement>()
const isResizing = ref(false)
const isDragging = ref(false)

const startPosition = ref<ResizePosition>({ x: 0, y: 0 })
const currentPosition = ref<ResizePosition>({ x: 0, y: 0 })
const startSize = ref<ResizeSize>({ width: 0, height: 0 })
const currentSize = ref<ResizeSize>({ width: 0, height: 0 })

const mousePosition = ref<ResizePosition>({ x: 0, y: 0 })
const snapIndicators = ref<SnapIndicator[]>([])

const resizeTimeout = ref<number>()

// Computed
const handleDots = computed(() => {
  switch (props.direction) {
    case 'horizontal':
      return 2
    case 'vertical':
      return 3
    case 'both':
    default:
      return 3
  }
})

const overlayStyle = computed(() => ({
  left: `${Math.min(startPosition.value.x, currentPosition.value.x)}px`,
  top: `${Math.min(startPosition.value.y, currentPosition.value.y)}px`,
  width: `${Math.abs(currentPosition.value.x - startPosition.value.x)}px`,
  height: `${Math.abs(currentPosition.value.y - startPosition.value.y)}px`
}))

const ghostStyle = computed(() => ({
  width: `${currentSize.value.width}px`,
  height: `${currentSize.value.height}px`
}))

const gridInfo = computed((): GridInfo => {
  // Approximate character grid size (assuming monospace font)
  const charWidth = 8
  const charHeight = 16
  return {
    cols: Math.floor(currentSize.value.width / charWidth),
    rows: Math.floor(currentSize.value.height / charHeight)
  }
})

const isNearMinimum = computed(() => {
  return (
    currentSize.value.width <= props.minSize.width + 20 ||
    currentSize.value.height <= props.minSize.height + 20
  )
})

const effectiveMinSize = computed(() => {
  return {
    width: props.constraints?.minWidth || props.minSize.width,
    height: props.constraints?.minHeight || props.minSize.height
  }
})

const effectiveMaxSize = computed(() => {
  return {
    width: props.constraints?.maxWidth || props.maxSize.width,
    height: props.constraints?.maxHeight || props.maxSize.height
  }
})

// Methods
const handleMouseDown = (event: MouseEvent) => {
  if (event.button !== 0) return // Only left click

  startResize(event, {
    x: event.clientX,
    y: event.clientY
  })
}

const handleTouchStart = (event: TouchEvent) => {
  if (event.touches.length !== 1) return

  const touch = event.touches[0]
  startResize(event, {
    x: touch.clientX,
    y: touch.clientY
  })
}

const startResize = (event: MouseEvent | TouchEvent, position: ResizePosition) => {
  if (!resizeContainer.value) return

  isResizing.value = true
  isDragging.value = true

  startPosition.value = { ...position }
  currentPosition.value = { ...position }
  startSize.value = { ...currentSize.value }

  // Emit resize start
  emit('resizeStart', event, { ...currentSize.value })
  props.onResizeStart?.()

  // Add global event listeners
  document.addEventListener('mousemove', handleMouseMove, { passive: false })
  document.addEventListener('mouseup', handleMouseUp, { passive: false })
  document.addEventListener('touchmove', handleTouchMove, { passive: false })
  document.addEventListener('touchend', handleTouchEnd, { passive: false })

  // Prevent default text selection during resize
  document.body.style.userSelect = 'none'
  document.body.style.cursor = getCursorForDirection()

  event.preventDefault()
}

const handleMouseMove = (event: MouseEvent) => {
  if (!isResizing.value) return

  updateResize({
    x: event.clientX,
    y: event.clientY
  })

  event.preventDefault()
}

const handleTouchMove = (event: TouchEvent) => {
  if (!isResizing.value || event.touches.length !== 1) return

  const touch = event.touches[0]
  updateResize({
    x: touch.clientX,
    y: touch.clientY
  })

  event.preventDefault()
}

const updateResize = (position: ResizePosition) => {
  mousePosition.value = { ...position }
  currentPosition.value = { ...position }

  let newSize = calculateNewSize(position)

  // Apply constraints
  newSize = applyConstraints(newSize)

  // Snap to grid if enabled
  if (props.snapToGrid) {
    newSize = snapToGrid(newSize)
    checkSnapIndicators(newSize)
  }

  currentSize.value = newSize

  // Debounce resize events
  if (props.debounceMs > 0) {
    if (resizeTimeout.value) {
      clearTimeout(resizeTimeout.value)
    }
    resizeTimeout.value = window.setTimeout(() => {
      emitResize(newSize)
    }, props.debounceMs)
  } else {
    emitResize(newSize)
  }

  // Check size limits
  checkSizeLimits(newSize)
}

const calculateNewSize = (position: ResizePosition): ResizeSize => {
  const deltaX = position.x - startPosition.value.x
  const deltaY = position.y - startPosition.value.y

  let newWidth = startSize.value.width
  let newHeight = startSize.value.height

  switch (props.direction) {
    case 'horizontal':
      newWidth = startSize.value.width + deltaX
      break
    case 'vertical':
      newHeight = startSize.value.height + deltaY
      break
    case 'both':
    default:
      newWidth = startSize.value.width + deltaX
      newHeight = startSize.value.height + deltaY
      break
  }

  // Apply step
  if (props.step > 1) {
    newWidth = Math.round(newWidth / props.step) * props.step
    newHeight = Math.round(newHeight / props.step) * props.step
  }

  return { width: newWidth, height: newHeight }
}

const applyConstraints = (size: ResizeSize): ResizeSize => {
  let { width, height } = size

  // Apply min/max constraints
  width = Math.max(effectiveMinSize.value.width, width)
  height = Math.max(effectiveMinSize.value.height, height)
  width = Math.min(effectiveMaxSize.value.width, width)
  height = Math.min(effectiveMaxSize.value.height, height)

  // Apply aspect ratio constraint if specified
  if (props.constraints?.preserveAspectRatio && props.constraints.aspectRatio) {
    const aspectRatio = props.constraints.aspectRatio
    const currentRatio = width / height

    if (currentRatio > aspectRatio) {
      width = height * aspectRatio
    } else {
      height = width / aspectRatio
    }
  }

  return { width, height }
}

const snapToGrid = (size: ResizeSize): ResizeSize => {
  if (!props.snapToGrid || props.gridSize <= 0) return size

  return {
    width: Math.round(size.width / props.gridSize) * props.gridSize,
    height: Math.round(size.height / props.gridSize) * props.gridSize
  }
}

const checkSnapIndicators = (size: ResizeSize) => {
  snapIndicators.value = []

  if (!props.snapToGrid || props.gridSize <= 0) return

  // Check if near grid lines
  const xNearGrid = Math.abs(size.width % props.gridSize) < props.gridSize / 4
  const yNearGrid = Math.abs(size.height % props.gridSize) < props.gridSize / 4

  if (xNearGrid) {
    snapIndicators.value.push({
      type: 'vertical',
      style: {
        left: `${size.width}px`,
        top: '0',
        height: '100%'
      }
    })
  }

  if (yNearGrid) {
    snapIndicators.value.push({
      type: 'horizontal',
      style: {
        left: '0',
        top: `${size.height}px`,
        width: '100%'
      }
    })
  }
}

const checkSizeLimits = (size: ResizeSize) => {
  // Check minimum size
  if (
    size.width <= effectiveMinSize.value.width + 5 ||
    size.height <= effectiveMinSize.value.height + 5
  ) {
    emit('minSizeReached', size)
  }

  // Check maximum size
  if (
    size.width >= effectiveMaxSize.value.width - 5 ||
    size.height >= effectiveMaxSize.value.height - 5
  ) {
    emit('maxSizeReached', size)
  }
}

const emitResize = (size: ResizeSize) => {
  emit('resize', size, currentPosition.value)
  props.onResize?.(size)
}

const handleMouseUp = (event: MouseEvent) => {
  endResize()
}

const handleTouchEnd = (event: TouchEvent) => {
  endResize()
}

const endResize = () => {
  if (!isResizing.value) return

  isResizing.value = false
  isDragging.value = false
  snapIndicators.value = []

  // Remove global event listeners
  document.removeEventListener('mousemove', handleMouseMove)
  document.removeEventListener('mouseup', handleMouseUp)
  document.removeEventListener('touchmove', handleTouchMove)
  document.removeEventListener('touchend', handleTouchEnd)

  // Restore body styles
  document.body.style.userSelect = ''
  document.body.style.cursor = ''

  // Clear resize timeout
  if (resizeTimeout.value) {
    clearTimeout(resizeTimeout.value)
  }

  // Emit final resize
  emit('resizeEnd', { ...currentSize.value })
  props.onResizeEnd?.({ ...currentSize.value })
}

const getCursorForDirection = (): string => {
  switch (props.direction) {
    case 'horizontal':
      return 'ew-resize'
    case 'vertical':
      return 'ns-resize'
    case 'both':
    default:
      return 'nwse-resize'
  }
}

const setSize = (size: ResizeSize) => {
  currentSize.value = { ...size }
}

const reset = () => {
  currentSize.value = { ...props.initialSize }
}

// Keyboard shortcuts
const handleKeyDown = (event: KeyboardEvent) => {
  if (!props.enableKeyboard || !isResizing.value) return

  let deltaSize = { width: 0, height: 0 }
  const step = event.shiftKey ? props.gridSize * 5 : props.gridSize

  switch (event.key) {
    case 'ArrowUp':
      deltaSize.height = -step
      break
    case 'ArrowDown':
      deltaSize.height = step
      break
    case 'ArrowLeft':
      deltaSize.width = -step
      break
    case 'ArrowRight':
      deltaSize.width = step
      break
    case 'Enter':
    case 'Escape':
      endResize()
      return
    default:
      return
  }

  event.preventDefault()

  const newPosition = {
    x: currentPosition.value.x + deltaSize.width,
    y: currentPosition.value.y + deltaSize.height
  }

  updateResize(newPosition)
}

// Lifecycle
onMounted(() => {
  currentSize.value = { ...props.initialSize }

  if (props.enableKeyboard) {
    document.addEventListener('keydown', handleKeyDown)
  }
})

onUnmounted(() => {
  endResize()

  if (props.enableKeyboard) {
    document.removeEventListener('keydown', handleKeyDown)
  }
})

// Expose methods for parent components
defineExpose({
  setSize,
  reset,
  isResizing: readonly(isResizing),
  currentSize: readonly(currentSize)
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.terminal-resize {
  @apply relative;
}

.terminal-resize--active {
  z-index: 1000;
}

.terminal-resize__handle {
  @apply flex items-center justify-center transition-all duration-200;
}

.terminal-resize__handle--horizontal {
  @apply cursor-ew-resize w-full h-2;
}

.terminal-resize__handle--vertical {
  @apply cursor-ns-resize w-2 h-full;
}

.terminal-resize__handle--both {
  @apply cursor-nwse-resize w-4 h-4;
}

.terminal-resize__handle-visual {
  @apply flex items-center justify-center gap-0.5;
}

.terminal-resize__handle--horizontal .terminal-resize__handle-visual {
  @apply flex-row;
}

.terminal-resize__handle--vertical .terminal-resize__handle-visual {
  @apply flex-col;
}

.terminal-resize__handle--both .terminal-resize__handle-visual {
  @apply flex-col;
}

.terminal-resize__handle-dot {
  @apply w-1 h-1 bg-text-secondary rounded-full opacity-50 transition-opacity duration-200;
}

.terminal-resize__handle:hover .terminal-resize__handle-dot {
  @apply opacity-100;
  background-color: var(--terminal-resize-handle-color, var(--color-primary));
}

.terminal-resize__overlay {
  @apply fixed border-2 border-primary bg-primary/10 pointer-events-none;
  z-index: 9999;
  border-style: dashed;
}

.terminal-resize__overlay-content {
  @apply absolute top-2 left-2 bg-surface border border-border rounded px-2 py-1 text-xs;
}

.terminal-resize__overlay-size {
  @apply font-mono font-medium text-primary;
}

.terminal-resize__overlay-grid,
.terminal-resize__overlay-position {
  @apply font-mono text-text-secondary;
}

.terminal-resize__ghost {
  @apply absolute top-0 left-0 border-2 border-primary bg-surface/50 pointer-events-none;
  border-style: solid;
}

.terminal-resize__min-size-indicator {
  @apply absolute -top-8 left-0 flex items-center gap-1 bg-warning text-warning-foreground text-xs px-2 py-1 rounded;
  z-index: 1001;
}

.terminal-resize__snap-indicator {
  @apply absolute bg-primary pointer-events-none;
}

.terminal-resize__snap-indicator--horizontal {
  @apply h-0.5 w-full left-0;
}

.terminal-resize__snap-indicator--vertical {
  @apply w-0.5 h-full top-0;
}

/* Global styles during resize */
:global(.body--resizing) {
  user-select: none !important;
}

:global(.cursor-ew-resize) {
  cursor: ew-resize !important;
}

:global(.cursor-ns-resize) {
  cursor: ns-resize !important;
}

:global(.cursor-nwse-resize) {
  cursor: nwse-resize !important;
}
</style>