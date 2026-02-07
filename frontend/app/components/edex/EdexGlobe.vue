<template>
  <div class="mod-panel mod-globe">
    <!-- Header -->
    <div class="section-title">
      <span class="title-left">WORLD VIEW</span>
      <span class="title-right">PROJECTION</span>
    </div>

    <!-- Globe canvas container -->
    <div class="globe-container" ref="globeContainerRef">
      <canvas
        ref="canvasRef"
        class="globe-canvas"
        @mousemove="onMouseMove"
        @mouseleave="onMouseLeave"
      />

      <!-- Endpoint marker overlay -->
      <div
        v-if="endpointVisible"
        class="endpoint-marker"
        :style="endpointStyle"
      >
        <span class="marker-dot"></span>
        <span class="marker-label">{{ endpointLabel }}</span>
      </div>
    </div>

    <!-- Endpoint info -->
    <div class="globe-info">
      <div class="globe-info-row">
        <span class="globe-info-label">ENDPOINT</span>
        <span class="globe-info-value">{{ endpointCoords }}</span>
      </div>
      <div class="globe-info-row">
        <span class="globe-info-label">LOCATION</span>
        <span class="globe-info-value globe-info-truncated">{{ locationName }}</span>
      </div>
      <div class="globe-info-row">
        <span class="globe-info-label">CONNECTIONS</span>
        <span class="globe-info-value">{{ activeConnectionCount }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useNetworkStore } from '~/stores/network'

const networkStore = useNetworkStore()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const globeContainerRef = ref<HTMLElement | null>(null)
const mousePos = ref<{ x: number; y: number } | null>(null)

// Props for endpoint location (can be fed from parent or store)
const props = defineProps<{
  latitude?: number
  longitude?: number
  city?: string
  country?: string
}>()

// Globe rendering state
let animationFrame: number | null = null
let rotationAngle = 0
const ROTATION_SPEED = 0.002

// Computed
const endpointLat = computed(() => props.latitude ?? 0)
const endpointLon = computed(() => props.longitude ?? 0)

const endpointCoords = computed(() => {
  const lat = endpointLat.value
  const lon = endpointLon.value
  const latDir = lat >= 0 ? 'N' : 'S'
  const lonDir = lon >= 0 ? 'E' : 'W'
  return `${Math.abs(lat).toFixed(2)}${latDir} ${Math.abs(lon).toFixed(2)}${lonDir}`
})

const locationName = computed(() => {
  const parts = []
  if (props.city) parts.push(props.city)
  if (props.country) parts.push(props.country)
  return parts.length > 0 ? parts.join(', ') : 'Resolving...'
})

const activeConnectionCount = computed(() => {
  return networkStore.activeConnections?.length ?? 0
})

// Project lat/lon to 2D position on the canvas globe
const projectToCanvas = (lat: number, lon: number, cx: number, cy: number, radius: number) => {
  const phi = (90 - lat) * (Math.PI / 180)
  const theta = (lon + rotationAngle * (180 / Math.PI)) * (Math.PI / 180)

  const x3d = Math.sin(phi) * Math.cos(theta)
  const y3d = Math.cos(phi)
  const z3d = Math.sin(phi) * Math.sin(theta)

  // Only visible if on the front hemisphere
  if (z3d < -0.1) return null

  return {
    x: cx + x3d * radius,
    y: cy - y3d * radius,
    z: z3d
  }
}

const endpointVisible = ref(false)
const endpointStyle = ref<Record<string, string>>({})
const endpointLabel = computed(() => props.city || 'ENDPOINT')

// Update the CSS position of the endpoint marker
const updateEndpointMarker = (cx: number, cy: number, radius: number) => {
  const pos = projectToCanvas(endpointLat.value, endpointLon.value, cx, cy, radius)
  if (pos && pos.z > -0.1) {
    endpointVisible.value = true
    endpointStyle.value = {
      left: `${pos.x}px`,
      top: `${pos.y}px`,
      opacity: String(Math.max(0.3, (pos.z + 1) / 2))
    }
  } else {
    endpointVisible.value = false
  }
}

// Draw wireframe globe with dot grid on canvas
const drawGlobe = () => {
  const canvas = canvasRef.value
  if (!canvas) return

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const container = globeContainerRef.value
  if (!container) return

  // Size the canvas to the container
  const rect = container.getBoundingClientRect()
  const dpr = window.devicePixelRatio || 1
  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr
  canvas.style.width = `${rect.width}px`
  canvas.style.height = `${rect.height}px`
  ctx.scale(dpr, dpr)

  const w = rect.width
  const h = rect.height
  const cx = w / 2
  const cy = h / 2
  const radius = Math.min(w, h) * 0.4

  // Clear
  ctx.clearRect(0, 0, w, h)

  // Read the accent color from CSS variables
  const style = getComputedStyle(document.documentElement)
  const r = style.getPropertyValue('--color_r').trim() || '170'
  const g = style.getPropertyValue('--color_g').trim() || '207'
  const b = style.getPropertyValue('--color_b').trim() || '209'

  // Draw globe outline
  ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, 0.3)`
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.arc(cx, cy, radius, 0, Math.PI * 2)
  ctx.stroke()

  // Draw latitude lines as dot grids
  const dotSize = 1.2
  for (let lat = -80; lat <= 80; lat += 20) {
    for (let lon = -180; lon < 180; lon += 10) {
      const pos = projectToCanvas(lat, lon, cx, cy, radius)
      if (pos && pos.z > -0.05) {
        const alpha = Math.max(0.1, (pos.z + 1) / 2) * 0.6
        ctx.fillStyle = `rgba(${r}, ${g}, ${b}, ${alpha})`
        ctx.beginPath()
        ctx.arc(pos.x, pos.y, dotSize, 0, Math.PI * 2)
        ctx.fill()
      }
    }
  }

  // Draw longitude arcs
  ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, 0.12)`
  ctx.lineWidth = 0.5
  for (let lon = -180; lon < 180; lon += 30) {
    ctx.beginPath()
    let started = false
    for (let lat = -90; lat <= 90; lat += 3) {
      const pos = projectToCanvas(lat, lon, cx, cy, radius)
      if (pos && pos.z > -0.05) {
        if (!started) {
          ctx.moveTo(pos.x, pos.y)
          started = true
        } else {
          ctx.lineTo(pos.x, pos.y)
        }
      } else {
        started = false
      }
    }
    ctx.stroke()
  }

  // Draw latitude arcs
  for (let lat = -60; lat <= 60; lat += 30) {
    ctx.beginPath()
    let started = false
    for (let lon = -180; lon < 180; lon += 3) {
      const pos = projectToCanvas(lat, lon, cx, cy, radius)
      if (pos && pos.z > -0.05) {
        if (!started) {
          ctx.moveTo(pos.x, pos.y)
          started = true
        } else {
          ctx.lineTo(pos.x, pos.y)
        }
      } else {
        started = false
      }
    }
    ctx.stroke()
  }

  // Draw endpoint as a pulsing dot
  const epPos = projectToCanvas(endpointLat.value, endpointLon.value, cx, cy, radius)
  if (epPos && epPos.z > -0.1) {
    const pulse = Math.sin(Date.now() / 300) * 0.3 + 0.7
    // Outer ring
    ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, ${pulse * 0.6})`
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.arc(epPos.x, epPos.y, 5 + pulse * 3, 0, Math.PI * 2)
    ctx.stroke()

    // Inner dot
    ctx.fillStyle = `rgba(${r}, ${g}, ${b}, ${Math.max(0.5, pulse)})`
    ctx.beginPath()
    ctx.arc(epPos.x, epPos.y, 2.5, 0, Math.PI * 2)
    ctx.fill()
  }

  // Update endpoint HTML marker
  updateEndpointMarker(cx, cy, radius)
}

const animate = () => {
  rotationAngle += ROTATION_SPEED
  drawGlobe()
  animationFrame = requestAnimationFrame(animate)
}

const onMouseMove = (e: MouseEvent) => {
  const rect = canvasRef.value?.getBoundingClientRect()
  if (rect) {
    mousePos.value = { x: e.clientX - rect.left, y: e.clientY - rect.top }
  }
}

const onMouseLeave = () => {
  mousePos.value = null
}

onMounted(() => {
  nextTick(() => {
    animate()
  })
})

onUnmounted(() => {
  if (animationFrame !== null) {
    cancelAnimationFrame(animationFrame)
    animationFrame = null
  }
})
</script>

<style scoped>
.mod-globe {
  font-size: 1.1vh;
}

.globe-container {
  position: relative;
  width: 100%;
  height: 14vh;
  overflow: hidden;
  margin: 0.3vh 0;
}

.globe-canvas {
  width: 100%;
  height: 100%;
  display: block;
}

.endpoint-marker {
  position: absolute;
  transform: translate(-50%, -100%);
  pointer-events: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: opacity 0.3s;
}

.marker-dot {
  display: block;
  width: 0.5vh;
  height: 0.5vh;
  border-radius: 50%;
  background: var(--color_accent);
  box-shadow: 0 0 0.4vh var(--color_accent);
}

.marker-label {
  font-size: 0.8vh;
  color: var(--color_accent);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  white-space: nowrap;
  margin-top: 0.1vh;
  opacity: 0.8;
}

.globe-info {
  padding: 0.2vh 0;
}

.globe-info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.1vh 0;
}

.globe-info-label {
  text-transform: uppercase;
  opacity: 0.6;
  letter-spacing: 0.05em;
}

.globe-info-value {
  text-align: right;
  font-family: var(--font_main);
}

.globe-info-truncated {
  max-width: 10vw;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
