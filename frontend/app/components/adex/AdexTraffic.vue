<template>
  <div class="mod-panel mod-traffic">
    <!-- Header -->
    <div class="section-title">
      <span class="title-left">NETWORK TRAFFIC</span>
      <span class="title-right">UP / DOWN, MB/S</span>
    </div>

    <!-- Total transferred -->
    <div class="traffic-total">
      <span class="traffic-total-label">TOTAL</span>
      <span class="traffic-total-value">{{ totalTransferred }}</span>
    </div>

    <!-- Upload chart -->
    <div class="traffic-chart-section">
      <div class="traffic-chart-label">
        <span class="chart-direction">UPLOAD</span>
        <span class="chart-rate">{{ currentUploadRate }} MB/s</span>
      </div>
      <div class="traffic-chart" ref="uploadChartContainer">
        <canvas ref="uploadCanvas" class="traffic-canvas" />
        <div class="chart-scale">
          <span>{{ uploadScaleMax.toFixed(1) }}</span>
          <span>{{ (uploadScaleMax / 2).toFixed(1) }}</span>
          <span>0.0</span>
        </div>
      </div>
    </div>

    <!-- Download chart -->
    <div class="traffic-chart-section">
      <div class="traffic-chart-label">
        <span class="chart-direction">DOWNLOAD</span>
        <span class="chart-rate">{{ currentDownloadRate }} MB/s</span>
      </div>
      <div class="traffic-chart" ref="downloadChartContainer">
        <canvas ref="downloadCanvas" class="traffic-canvas" />
        <div class="chart-scale">
          <span>{{ downloadScaleMax.toFixed(1) }}</span>
          <span>{{ (downloadScaleMax / 2).toFixed(1) }}</span>
          <span>0.0</span>
        </div>
      </div>
    </div>

    <!-- Stats footer -->
    <div class="traffic-stats">
      <span>SENT: {{ formatBytes(totalBytesSent) }}</span>
      <span>RECV: {{ formatBytes(totalBytesRecv) }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useStorage, useIntervalFn } from '@vueuse/core'
import { useNetworkStore } from '~/stores/network'

const networkStore = useNetworkStore()

// Canvas refs
const uploadCanvas = ref<HTMLCanvasElement | null>(null)
const downloadCanvas = ref<HTMLCanvasElement | null>(null)
const uploadChartContainer = ref<HTMLElement | null>(null)
const downloadChartContainer = ref<HTMLElement | null>(null)

// Chart data - keep 60 data points (60 seconds of data at 1s intervals)
const MAX_POINTS = 60
const uploadHistory = ref<number[]>(new Array(MAX_POINTS).fill(0))
const downloadHistory = ref<number[]>(new Array(MAX_POINTS).fill(0))

// Previous values for calculating rate
let prevBytesSent = 0
let prevBytesRecv = 0
let prevTimestamp = Date.now()

// Scale maximums
const uploadScaleMax = ref(1.0)
const downloadScaleMax = ref(1.0)

// Total bytes
const totalBytesSent = ref(0)
const totalBytesRecv = ref(0)

// Current rates
const currentUploadRate = computed(() => {
  const last = uploadHistory.value[uploadHistory.value.length - 1]
  return last.toFixed(2)
})

const currentDownloadRate = computed(() => {
  const last = downloadHistory.value[downloadHistory.value.length - 1]
  return last.toFixed(2)
})

const totalTransferred = computed(() => {
  return formatBytes(totalBytesSent.value + totalBytesRecv.value)
})

// Utility: format bytes to human readable
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

// Read CSS accent color
const getAccentColor = (): { r: string; g: string; b: string } => {
  const style = getComputedStyle(document.documentElement)
  return {
    r: style.getPropertyValue('--color_r').trim() || '170',
    g: style.getPropertyValue('--color_g').trim() || '207',
    b: style.getPropertyValue('--color_b').trim() || '209'
  }
}

// Draw a single line chart on a canvas
const drawChart = (
  canvas: HTMLCanvasElement | null,
  container: HTMLElement | null,
  data: number[],
  scaleMax: number
) => {
  if (!canvas || !container) return

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const rect = container.getBoundingClientRect()
  const dpr = window.devicePixelRatio || 1

  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr
  canvas.style.width = `${rect.width}px`
  canvas.style.height = `${rect.height}px`
  ctx.scale(dpr, dpr)

  const w = rect.width
  const h = rect.height
  const { r, g, b } = getAccentColor()

  // Clear
  ctx.clearRect(0, 0, w, h)

  // Draw horizontal grid lines
  ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, 0.07)`
  ctx.lineWidth = 0.5
  for (let i = 0; i <= 4; i++) {
    const y = (h / 4) * i
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(w, y)
    ctx.stroke()
  }

  // Draw vertical grid lines
  for (let i = 0; i <= 6; i++) {
    const x = (w / 6) * i
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, h)
    ctx.stroke()
  }

  if (data.length < 2) return

  // Calculate the scale leaving some headroom
  const max = scaleMax > 0 ? scaleMax : 1

  // Draw the fill area
  const gradient = ctx.createLinearGradient(0, 0, 0, h)
  gradient.addColorStop(0, `rgba(${r}, ${g}, ${b}, 0.3)`)
  gradient.addColorStop(1, `rgba(${r}, ${g}, ${b}, 0.02)`)

  ctx.fillStyle = gradient
  ctx.beginPath()
  ctx.moveTo(0, h)

  const pointSpacing = w / (MAX_POINTS - 1)
  for (let i = 0; i < data.length; i++) {
    const x = i * pointSpacing
    const y = h - (data[i] / max) * h
    if (i === 0) {
      ctx.lineTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  }

  ctx.lineTo((data.length - 1) * pointSpacing, h)
  ctx.closePath()
  ctx.fill()

  // Draw the line
  ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, 0.9)`
  ctx.lineWidth = 1.5
  ctx.lineJoin = 'round'
  ctx.lineCap = 'round'
  ctx.beginPath()

  for (let i = 0; i < data.length; i++) {
    const x = i * pointSpacing
    const y = h - (data[i] / max) * h
    if (i === 0) {
      ctx.moveTo(x, y)
    } else {
      ctx.lineTo(x, y)
    }
  }

  ctx.stroke()

  // Draw glow on the last point
  const lastX = (data.length - 1) * pointSpacing
  const lastY = h - (data[data.length - 1] / max) * h

  ctx.fillStyle = `rgba(${r}, ${g}, ${b}, 1)`
  ctx.beginPath()
  ctx.arc(lastX, lastY, 2, 0, Math.PI * 2)
  ctx.fill()

  // Glow ring
  ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, 0.4)`
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.arc(lastX, lastY, 4, 0, Math.PI * 2)
  ctx.stroke()
}

// Update charts
const redrawCharts = () => {
  drawChart(uploadCanvas.value, uploadChartContainer.value, uploadHistory.value, uploadScaleMax.value)
  drawChart(downloadCanvas.value, downloadChartContainer.value, downloadHistory.value, downloadScaleMax.value)
}

// Update scale to adapt to data
const updateScale = (history: number[], currentScale: { value: number }) => {
  const max = Math.max(...history, 0.01)
  // Add 20% headroom, round to nice number
  const target = max * 1.2
  // Smooth scaling
  if (target > currentScale.value) {
    currentScale.value = target
  } else {
    // Slowly decrease scale
    currentScale.value = currentScale.value * 0.98 + target * 0.02
  }
  // Minimum scale
  if (currentScale.value < 0.1) currentScale.value = 0.1
}

// Per-adapter byte totals when settings name a specific interface,
// otherwise the system-wide totals. Reading from useStorage keeps this
// reactive to settings changes without a remount.
const adexSettings = useStorage<{
  network?: { adapter?: string; trafficIntervalMs?: number }
}>('adex-settings', {})

// User-configurable sample cadence — clamped to keep useIntervalFn out
// of tight loops on corrupt storage values.
const trafficIntervalMs = computed(() => {
  const raw = Number(adexSettings.value?.network?.trafficIntervalMs ?? 1000)
  if (!Number.isFinite(raw)) return 1000
  return Math.min(5_000, Math.max(250, Math.floor(raw)))
})

// Switching adapters mid-session would otherwise produce a giant
// negative spike on the next tick (delta of two unrelated counters).
// Reset history + prev counters so the new adapter starts clean.
//
// Use the THIRD-arg-less form (no `immediate: true`) so the watcher
// doesn't fire on the initial useStorage hydration (`undefined → ''`),
// which used to wipe prev counters on mount and prevent the very
// first delta from ever being computed.
watch(
  () => (adexSettings.value?.network?.adapter ?? '').trim(),
  (next, prev) => {
    if (next === prev) return
    prevBytesSent = 0
    prevBytesRecv = 0
    prevTimestamp = Date.now()
    uploadHistory.value = new Array(MAX_POINTS).fill(0)
    downloadHistory.value = new Array(MAX_POINTS).fill(0)
  }
)

function resolveBytes(): { sent: number; recv: number } {
  const metrics = networkStore.metrics
  if (!metrics) return { sent: 0, recv: 0 }
  // Per-adapter when settings name a specific interface AND that
  // interface has live counters. Otherwise fall through to system
  // totals — better to graph "all traffic" than nothing.
  const pref = (adexSettings.value?.network?.adapter ?? '').trim()
  if (pref && Array.isArray(metrics.interfaces)) {
    const iface = metrics.interfaces.find((i: any) => i.name === pref)
    if (iface) {
      const sent = Number(iface.bytesSent ?? iface.totalBytesSent ?? 0)
      const recv = Number(iface.bytesRecv ?? iface.totalBytesRecv ?? 0)
      if (sent > 0 || recv > 0) return { sent, recv }
    }
  }
  return {
    sent: Number(metrics.totalBytesSent || 0),
    recv: Number(metrics.totalBytesRecv || 0),
  }
}

// Poll and update data
const poll = async () => {
  try {
    await networkStore.fetchMetrics()

    const metrics = networkStore.metrics
    const now = Date.now()
    const elapsed = (now - prevTimestamp) / 1000 // seconds

    if (metrics && elapsed > 0) {
      const { sent: bytesSent, recv: bytesRecv } = resolveBytes()

      totalBytesSent.value = bytesSent
      totalBytesRecv.value = bytesRecv

      // Calculate MB/s rates
      if (prevBytesSent > 0 || prevBytesRecv > 0) {
        const uploadRate = Math.max(0, (bytesSent - prevBytesSent) / elapsed / (1024 * 1024))
        const downloadRate = Math.max(0, (bytesRecv - prevBytesRecv) / elapsed / (1024 * 1024))

        // Push to history, shift old values
        uploadHistory.value.push(uploadRate)
        if (uploadHistory.value.length > MAX_POINTS) {
          uploadHistory.value.shift()
        }

        downloadHistory.value.push(downloadRate)
        if (downloadHistory.value.length > MAX_POINTS) {
          downloadHistory.value.shift()
        }

        // Update scales
        updateScale(uploadHistory.value, uploadScaleMax)
        updateScale(downloadHistory.value, downloadScaleMax)
      }

      prevBytesSent = bytesSent
      prevBytesRecv = bytesRecv
      prevTimestamp = now
    }
  } catch (err) {
    // Silently continue polling on error
  }

  redrawCharts()
}

onMounted(() => {
  nextTick(() => {
    // Initial draw — chart canvas needs DOM layout before measuring.
    redrawCharts()
  })
})

// Network sample cadence — user-configurable via Settings → Network →
// Traffic sample cadence. Passing the reactive ref makes useIntervalFn
// re-arm automatically on slider change.
useIntervalFn(poll, trafficIntervalMs, { immediate: true, immediateCallback: true })
</script>

<style scoped>
/* Two-stack layout: header / total / [UPLOAD section] / [DOWNLOAD section]
 * / footer. The two chart sections grow to share whatever vertical space
 * the panel was allotted by the column flex (see pages/index.vue) so
 * BOTH charts always render — the previous CSS used `height: 8vh` on
 * each chart which caused the second one to overflow and disappear on
 * sub-1080 viewports. */
.mod-traffic {
  font-size: 1vh;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.traffic-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.2vh 0;
  font-size: 1.1vh;
  flex-shrink: 0;
}

.traffic-total-label {
  text-transform: uppercase;
  opacity: 0.6;
  letter-spacing: 0.1em;
}

.traffic-total-value {
  font-family: var(--font_main);
  color: var(--color_accent);
}

.traffic-chart-section {
  margin-bottom: 0.3vh;
  display: flex;
  flex-direction: column;
  /* Both UP and DOWN sections claim equal share of remaining height. */
  flex: 1 1 0;
  min-height: 0;
}

.traffic-chart-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.15vh 0;
  font-size: 0.9vh;
  flex-shrink: 0;
}

.chart-direction {
  text-transform: uppercase;
  opacity: 0.6;
  letter-spacing: 0.1em;
}

.chart-rate {
  font-family: var(--font_main);
  color: var(--color_accent);
}

.traffic-chart {
  /* Mix grow + a fixed pixel min-height. Pure flex: 1 1 0 collapsed
   * the canvas to 1px on tight columns — graphs invisible despite the
   * data being live. min-height: 50px guarantees a renderable area. */
  flex: 1 1 0;
  min-height: 50px;
  position: relative;
  overflow: hidden;
  border: var(--border_width) solid rgba(var(--color_r), var(--color_g), var(--color_b), 0.1);
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.02);
}

.traffic-canvas {
  width: 100%;
  height: 100%;
  display: block;
}

.chart-scale {
  position: absolute;
  top: 0;
  right: 0.2vw;
  bottom: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  font-size: 0.8vh;
  opacity: 0.4;
  pointer-events: none;
  padding: 0.1vh 0;
  text-align: right;
  font-family: var(--font_main);
}

.traffic-stats {
  display: flex;
  justify-content: space-between;
  font-size: 0.9vh;
  opacity: 0.7;
  padding-top: 0.2vh;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  flex-shrink: 0;
}
</style>
