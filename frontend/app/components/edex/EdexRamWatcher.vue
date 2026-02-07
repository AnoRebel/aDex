<template>
  <div class="mod-panel mod-ram">
    <!-- Header -->
    <div class="section-title">
      <span class="title-left">MEMORY</span>
      <span class="title-right">USING {{ usedGiB }} OUT OF {{ totalGiB }} GiB</span>
    </div>

    <!-- 440-cell grid (40 columns x 11 rows) -->
    <div class="ram-grid">
      <div
        v-for="(state, index) in gridCells"
        :key="index"
        class="ram-cell"
        :class="state"
      />
    </div>

    <!-- Swap bar -->
    <div class="swap-bar">
      <span>SWAP</span>
      <div class="bar-track">
        <div class="bar-fill" :style="{ width: swapPercent + '%' }" />
      </div>
      <span>{{ swapPercent.toFixed(0) }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useSystemStore } from '~/stores/system'

const systemStore = useSystemStore()

const TOTAL_CELLS = 440 // 40 columns x 11 rows

// A shuffled index map so cells light up in a visually scattered pattern
const shuffledIndices = ref<number[]>([])

let refreshTimer: ReturnType<typeof setInterval> | null = null

// ---- Computed memory values ----

const memTotal = computed(() => {
  const metrics = systemStore.memoryMetrics
  if (metrics?.total) return metrics.total
  const stats = systemStore.systemStats as any
  return stats?.memory?.total ?? 0
})

const memUsed = computed(() => {
  const metrics = systemStore.memoryMetrics
  if (metrics?.used) return metrics.used
  const stats = systemStore.systemStats as any
  return stats?.memory?.used ?? 0
})

const memAvailable = computed(() => {
  const metrics = systemStore.memoryMetrics
  if (metrics?.available) return metrics.available
  // Derive available from total - used if not provided
  return memTotal.value - memUsed.value
})

const memFree = computed(() => {
  const metrics = systemStore.memoryMetrics
  if (metrics?.free) return metrics.free
  const stats = systemStore.systemStats as any
  return stats?.memory?.free ?? 0
})

const swapPercent = computed(() => {
  const metrics = systemStore.memoryMetrics
  if (metrics?.swapPercent != null) return metrics.swapPercent
  if (metrics?.swapTotal && metrics.swapTotal > 0) {
    return ((metrics.swapUsed ?? 0) / metrics.swapTotal) * 100
  }
  return 0
})

// Display helpers
const usedGiB = computed(() => bytesToGiB(memUsed.value))
const totalGiB = computed(() => bytesToGiB(memTotal.value))

function bytesToGiB(bytes: number): string {
  if (bytes <= 0) return '0.0'
  return (bytes / (1024 * 1024 * 1024)).toFixed(1)
}

// ---- Grid cell state ----

const gridCells = computed<string[]>(() => {
  const total = memTotal.value || 1
  const usedFraction = memUsed.value / total
  const availableFraction = memAvailable.value / total
  // free = whatever is left after used + available (could overlap; clamp)
  const freeFraction = Math.max(0, 1 - usedFraction - availableFraction)

  const usedCells = Math.round(usedFraction * TOTAL_CELLS)
  const availableCells = Math.round(availableFraction * TOTAL_CELLS)
  // Remainder goes to free
  const freeCells = TOTAL_CELLS - usedCells - availableCells

  // Build an ordered state array then scatter it via the shuffled indices
  const ordered: string[] = new Array(TOTAL_CELLS)
  let idx = 0
  for (let i = 0; i < usedCells && idx < TOTAL_CELLS; i++, idx++) {
    ordered[idx] = 'used'
  }
  for (let i = 0; i < availableCells && idx < TOTAL_CELLS; i++, idx++) {
    ordered[idx] = 'available'
  }
  while (idx < TOTAL_CELLS) {
    ordered[idx] = 'free'
    idx++
  }

  // Scatter: position each value at the shuffled position
  const scattered: string[] = new Array(TOTAL_CELLS)
  for (let i = 0; i < TOTAL_CELLS; i++) {
    scattered[shuffledIndices.value[i]] = ordered[i]
  }
  return scattered
})

// ---- Fisher-Yates shuffle ----

function fisherYatesShuffle(): number[] {
  const arr = Array.from({ length: TOTAL_CELLS }, (_, i) => i)
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = arr[i]
    arr[i] = arr[j]
    arr[j] = tmp
  }
  return arr
}

// ---- Lifecycle ----

onMounted(() => {
  // Generate the shuffle map once on mount (keeps visual distribution stable)
  shuffledIndices.value = fisherYatesShuffle()

  // Re-read store data every 1500ms (the store itself refreshes from backend)
  // The computed properties are reactive, but we force a re-shuffle periodically
  // to give the grid a subtle animation effect.
  refreshTimer = setInterval(() => {
    // Re-shuffle to animate cell distribution
    shuffledIndices.value = fisherYatesShuffle()
  }, 1500)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>
