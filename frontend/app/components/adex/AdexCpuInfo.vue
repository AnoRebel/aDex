<template>
  <div class="mod-panel mod-cpu">
    <div class="section-title">
      <span class="title-left">{{ cpuModelName }}</span>
    </div>

    <!-- Chart group 1: cores 1 .. N/2 -->
    <div class="cpu-chart-container">
      <AdexMultiSparkline
        :series="topSeries"
        :accent="accentRgb"
        :window-size="HISTORY_LENGTH"
        aria-label="CPU usage, first half of cores"
      />
    </div>

    <!-- Chart group 2: cores N/2+1 .. N -->
    <div class="cpu-chart-container">
      <AdexMultiSparkline
        :series="bottomSeries"
        :accent="accentRgb"
        :window-size="HISTORY_LENGTH"
        aria-label="CPU usage, second half of cores"
      />
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
import { useIntervalFn } from '@vueuse/core'
import { useSystemStore } from '~/stores/system'
import AdexMultiSparkline from '~/components/adex/AdexMultiSparkline.vue'
import { GetTemperatures } from '~/lib/wailsjs/coordinator'

const systemStore = useSystemStore()

// Chart history - each core stores its own rolling buffer of values
const HISTORY_LENGTH = 60
const coreHistories = ref<number[][]>([])

// Animation frame is the only manually-managed handle now — VueUse's
// useIntervalFn handles the load/temp polling cadence with built-in
// onUnmount cleanup, so we don't need explicit timer refs.
let animFrameId: number | null = null

// Temperature cache (refreshed less often)
const cachedTemperature = ref<number | null>(null)

// ---- Computed data from store ----

const perCoreUsage = computed<number[]>(() => {
  // Prefer detailed cpuMetrics, fall back to systemStats cores.
  const metrics = systemStore.cpuMetrics
  if (metrics?.perCoreUsage && metrics.perCoreUsage.length > 0) {
    return metrics.perCoreUsage
  }
  const stats = systemStore.systemStats
  // The coordinator's GetCPUUsage returns `cores: []float64` (raw
  // gopsutil cpu.PercentWithContext output), so cores comes across the
  // Wails bridge as a plain number array — NOT an array of `{usage}`
  // objects like the legacy SystemStats type implies. Accept either
  // shape so this stays correct if the backend ever changes again.
  const cores = stats?.cpu?.cores
  if (Array.isArray(cores) && cores.length > 0) {
    return cores.map((c: any) =>
      typeof c === 'number'
        ? c
        : (c?.usage ?? c?.Usage ?? 0)
    )
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

// ---- Chart data ----

// Theme accent as an `r, g, b` triple. The charts cannot resolve CSS custom
// properties themselves, so it is read here and passed down, and re-read when
// the theme swaps those properties on <html>.
const accentRgb = ref('170, 207, 209')

function readAccent() {
  if (typeof document === 'undefined') return
  const root = getComputedStyle(document.documentElement)
  const r = root.getPropertyValue('--color_r').trim() || '170'
  const g = root.getPropertyValue('--color_g').trim() || '207'
  const b = root.getPropertyValue('--color_b').trim() || '209'
  accentRgb.value = `${r}, ${g}, ${b}`
}

let themeObserver: MutationObserver | null = null

onMounted(() => {
  readAccent()
  themeObserver = new MutationObserver(readAccent)
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['style', 'class', 'data-theme'],
  })
})

// The cores are split across two stacked charts so a high core count stays
// legible; each chart gets its own half of the buffers.
const topSeries = computed(() => coreHistories.value.slice(0, halfIndex.value))
const bottomSeries = computed(() => coreHistories.value.slice(halfIndex.value))

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
}

async function sampleTemperature() {
  // Try the store metrics path first (cheap; reactive). If the store
  // hasn't been populated yet (the GetSystemMetrics binding doesn't
  // exist in this build), fall back to a direct GetTemperatures call.
  const tempMetrics = systemStore.temperatureMetrics
  if (tempMetrics?.sensors && tempMetrics.sensors.length > 0) {
    const cpuSensor = pickCpuSensor(tempMetrics.sensors)
    if (cpuSensor?.temperature != null) {
      cachedTemperature.value = cpuSensor.temperature
      return
    }
  }
  try {
    const sensors = await GetTemperatures()
    if (sensors.length === 0) return
    const cpuSensor = pickCpuSensor(sensors)
    if (cpuSensor?.temperature != null) {
      cachedTemperature.value = cpuSensor.temperature
    }
  } catch {
    // Sensor APIs frequently fail in containers/Wayland — leave the
    // cached value alone rather than spamming console errors.
  }
}

// Pick the most representative CPU sensor: prefer Intel's
// `coretemp_package_id_0`, then any 'cpu'/'core' sensor, then the
// hottest reading as last resort. Mirrors internal/services/system/cpu.go's
// scoring.
function pickCpuSensor(sensors: Array<{ name: string; temperature: number }>) {
  const lower = (s: string) => s.toLowerCase()
  const byName = (needle: string) =>
    sensors.find((s) => lower(s.name).includes(needle))
  return (
    byName('package_id_0') ??
    byName('package id 0') ??
    byName('tdie') ??
    byName('tctl') ??
    byName('cpu') ??
    byName('core') ??
    [...sensors].sort((a, b) => b.temperature - a.temperature)[0]
  )
}

// ---- Lifecycle ----

// Polling cadences. useIntervalFn auto-cleans on unmount and gives us
// pause/resume handles for free — we use `immediate: true` so the
// first tick fires on mount, replacing the manual prime call.
useIntervalFn(sampleLoad, 500, { immediate: true, immediateCallback: true })
useIntervalFn(sampleTemperature, 2000, { immediate: true, immediateCallback: true })

onBeforeUnmount(() => {
  if (animFrameId !== null) {
    cancelAnimationFrame(animFrameId)
    animFrameId = null
  }
  themeObserver?.disconnect()
  themeObserver = null
})
</script>
