<template>
  <div class="mod-panel mod-cpu">
    <div class="section-title">
      <span class="title-left">{{ cpuModelName }}</span>
    </div>

    <!-- Chart group 1: cores 1 .. N/2 -->
    <div class="cpu-chart-container">
      <canvas ref="chartTopRef" />
    </div>

    <!-- Chart group 2: cores N/2+1 .. N -->
    <div class="cpu-chart-container">
      <canvas ref="chartBottomRef" />
    </div>

    <!-- Stats row -->
    <div class="cpu-stats">
      <div>
        <span class="stat-label">TEMP </span>
        <span>{{ temperatureDisplay }}</span>
      </div>
      <div>
        <span class="stat-label">FREQ </span>
        <span>{{ frequencyDisplay }}</span>
      </div>
      <div>
        <span class="stat-label">TASKS </span>
        <span>{{ taskCount }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useSystemStore } from '~/stores/system'

const systemStore = useSystemStore()

// Canvas refs
const chartTopRef = ref<HTMLCanvasElement | null>(null)
const chartBottomRef = ref<HTMLCanvasElement | null>(null)

// Chart history - each core stores its own rolling buffer of values
const HISTORY_LENGTH = 60
const coreHistories = ref<number[][]>([])

// Timers
let loadTimer: ReturnType<typeof setInterval> | null = null
let tempTimer: ReturnType<typeof setInterval> | null = null
let animFrameId: number | null = null

// Temperature cache (refreshed less often)
const cachedTemperature = ref<number | null>(null)

// ---- Computed data from store ----

const perCoreUsage = computed<number[]>(() => {
  // Prefer detailed cpuMetrics, fall back to systemStats cores
  const metrics = systemStore.cpuMetrics
  if (metrics?.perCoreUsage && metrics.perCoreUsage.length > 0) {
    return metrics.perCoreUsage
  }
  const stats = systemStore.systemStats
  if (stats?.cpu?.cores && stats.cpu.cores.length > 0) {
    return stats.cpu.cores.map((c: any) => c.usage ?? c.Usage ?? 0)
  }
  // Fallback: single overall usage replicated to 2 pseudo-cores
  const overall = stats?.cpu?.usage ?? metrics?.usagePercent ?? 0
  return [overall, overall]
})

const coreCount = computed(() => perCoreUsage.value.length || 2)

const cpuModelName = computed(() => {
  const metrics = systemStore.cpuMetrics
  if (metrics?.model) return metrics.model
  const stats = systemStore.systemStats as any
  if (stats?.cpu?.modelName) return stats.cpu.modelName
  return 'CPU'
})

const temperatureDisplay = computed(() => {
  const tempMetrics = systemStore.temperatureMetrics
  if (tempMetrics?.sensors && tempMetrics.sensors.length > 0) {
    const cpuSensor = tempMetrics.sensors.find(
      (s: any) => s.name.toLowerCase().includes('cpu') || s.name.toLowerCase().includes('core')
    ) || tempMetrics.sensors[0]
    if (cpuSensor?.temperature != null) {
      return `${Math.round(cpuSensor.temperature)}C`
    }
  }
  if (cachedTemperature.value !== null) {
    return `${cachedTemperature.value.toFixed(0)}C`
  }
  return '--C'
})

const frequencyDisplay = computed(() => {
  const metrics = systemStore.cpuMetrics
  const stats = systemStore.systemStats as any
  const freq = metrics?.frequency || stats?.cpu?.frequency || 0
  const freqMax = metrics?.frequencyMax || 0

  if (freq <= 0 && freqMax <= 0) return '-- GHz'

  // frequency values from gopsutil are in MHz
  const freqGHz = (freq / 1000).toFixed(2)
  if (freqMax > 0) {
    const maxGHz = (freqMax / 1000).toFixed(2)
    return `${freqGHz}/${maxGHz} GHz`
  }
  return `${freqGHz} GHz`
})

const taskCount = computed(() => {
  const pm = systemStore.processMetrics
  if (pm?.totalProcesses) return String(pm.totalProcesses)
  const procs = systemStore.processes
  if (procs && procs.length > 0) return String(procs.length)
  return '--'
})

// ---- Split cores into two groups ----

const halfIndex = computed(() => Math.ceil(coreCount.value / 2))

// ---- Chart drawing ----

function getAccentColor(): { r: number; g: number; b: number } {
  if (typeof document === 'undefined') return { r: 170, g: 207, b: 209 }
  const root = getComputedStyle(document.documentElement)
  const r = parseInt(root.getPropertyValue('--color_r').trim()) || 170
  const g = parseInt(root.getPropertyValue('--color_g').trim()) || 207
  const b = parseInt(root.getPropertyValue('--color_b').trim()) || 209
  return { r, g, b }
}

function drawChart(canvas: HTMLCanvasElement, coreIndices: number[]) {
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const dpr = window.devicePixelRatio || 1
  const rect = canvas.getBoundingClientRect()

  // Size the canvas to its CSS pixel size times device pixel ratio
  const w = rect.width
  const h = rect.height
  if (w === 0 || h === 0) return

  canvas.width = w * dpr
  canvas.height = h * dpr
  ctx.scale(dpr, dpr)

  // Clear
  ctx.clearRect(0, 0, w, h)

  const { r, g, b } = getAccentColor()

  // Draw grid lines (faint horizontal lines at 25%, 50%, 75%)
  ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, 0.07)`
  ctx.lineWidth = 0.5
  for (let pct = 0.25; pct < 1; pct += 0.25) {
    const y = h * (1 - pct)
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(w, y)
    ctx.stroke()
  }

  // Draw each core as a separate line
  const numCores = coreIndices.length
  if (numCores === 0) return

  for (let ci = 0; ci < numCores; ci++) {
    const coreIdx = coreIndices[ci]
    const history = coreHistories.value[coreIdx]
    if (!history || history.length < 2) continue

    const alpha = 0.3 + (0.7 * (ci / Math.max(numCores - 1, 1)))
    ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, ${alpha.toFixed(2)})`
    ctx.lineWidth = 1

    ctx.beginPath()
    const step = w / (HISTORY_LENGTH - 1)

    // Offset to right-align the data if we have fewer than HISTORY_LENGTH points
    const offset = (HISTORY_LENGTH - history.length) * step

    for (let i = 0; i < history.length; i++) {
      const x = offset + i * step
      const y = h - (history[i] / 100) * h
      if (i === 0) {
        ctx.moveTo(x, y)
      } else {
        // Simple line-to (smoothie-style would use bezier, but this is clean enough)
        ctx.lineTo(x, y)
      }
    }
    ctx.stroke()

    // Faint fill under the line
    ctx.lineTo(offset + (history.length - 1) * step, h)
    ctx.lineTo(offset, h)
    ctx.closePath()
    ctx.fillStyle = `rgba(${r}, ${g}, ${b}, ${(alpha * 0.08).toFixed(3)})`
    ctx.fill()
  }
}

function renderCharts() {
  if (chartTopRef.value) {
    const indices = Array.from({ length: halfIndex.value }, (_, i) => i)
    drawChart(chartTopRef.value, indices)
  }
  if (chartBottomRef.value) {
    const indices = Array.from(
      { length: coreCount.value - halfIndex.value },
      (_, i) => halfIndex.value + i
    )
    drawChart(chartBottomRef.value, indices)
  }
}

// ---- Data sampling ----

function sampleLoad() {
  const usage = perCoreUsage.value
  // Ensure histories array is sized to core count
  while (coreHistories.value.length < usage.length) {
    coreHistories.value.push([])
  }
  // Trim if core count decreased
  if (coreHistories.value.length > usage.length) {
    coreHistories.value.length = usage.length
  }

  for (let i = 0; i < usage.length; i++) {
    coreHistories.value[i].push(usage[i])
    if (coreHistories.value[i].length > HISTORY_LENGTH) {
      coreHistories.value[i].shift()
    }
  }

  renderCharts()
}

function sampleTemperature() {
  // Refresh cached temperature from store metrics
  const tempMetrics = systemStore.temperatureMetrics
  if (tempMetrics?.sensors && tempMetrics.sensors.length > 0) {
    const cpuSensor = tempMetrics.sensors.find(
      (s: any) => s.name.toLowerCase().includes('cpu') || s.name.toLowerCase().includes('core')
    ) || tempMetrics.sensors[0]
    if (cpuSensor?.temperature != null) {
      cachedTemperature.value = cpuSensor.temperature
    }
  }
}

// ---- Lifecycle ----

onMounted(() => {
  // Initial sample
  sampleLoad()
  sampleTemperature()

  // 500ms load polling
  loadTimer = setInterval(sampleLoad, 500)

  // 2000ms temperature polling
  tempTimer = setInterval(sampleTemperature, 2000)
})

onBeforeUnmount(() => {
  if (loadTimer) {
    clearInterval(loadTimer)
    loadTimer = null
  }
  if (tempTimer) {
    clearInterval(tempTimer)
    tempTimer = null
  }
  if (animFrameId !== null) {
    cancelAnimationFrame(animFrameId)
    animFrameId = null
  }
})
</script>
