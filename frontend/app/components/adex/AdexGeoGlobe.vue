<template>
  <div class="mod-panel mod-geo-globe">
    <div class="section-title">
      <span>WORLD VIEW</span>
      <span class="section-title-right">COUNTRIES</span>
    </div>

    <div ref="containerRef" class="globe-container">
      <canvas ref="canvasRef" class="geo-globe-canvas" />
    </div>

    <!-- Same readouts as the classic globe: swapping the projection should
         not cost the user the ISP and connection count. `title` exposes the
         full value when CSS ellipsis-truncates the row. -->
    <div class="geo-globe-meta">
      <div class="geo-row">
        <span>ENDPOINT</span>
        <span :title="endpointLabel">{{ endpointLabel }}</span>
      </div>
      <div class="geo-row">
        <span>LOCATION</span>
        <span class="geo-truncate" :title="locationLabel">{{ locationLabel }}</span>
      </div>
      <div v-if="ispName" class="geo-row">
        <span>ISP</span>
        <span class="geo-truncate" :title="ispName">{{ ispName }}</span>
      </div>
      <div class="geo-row">
        <span>CONNECTIONS</span>
        <span>{{ activeConnectionCount }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Country-outline globe — the alternative to the stylised dot-grid globe,
 * selectable in Settings → Theme → Globe style.
 *
 * All rendering happens in a Web Worker against an OffscreenCanvas. The
 * canvas is transferred once at mount and the main thread never draws to it
 * again; this component only measures the DOM (size, theme colour) and posts
 * that state across, because a worker has no DOM of its own.
 *
 * That split is the point: the globe redraws real country geometry, and doing
 * it on the main thread put it in direct competition with the terminal, the
 * charts and input handling. Off-thread, a slow frame here cannot stutter
 * anything else.
 *
 * 2D canvas rather than WebGL: the GPU path has crashed the renderer on this
 * platform, and a slowly rotating outline map does not need it.
 *
 * Geometry is the 110m world atlas (~105 KB), loaded lazily so a user on the
 * classic globe never downloads it.
 */
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useDocumentVisibility, useResizeObserver, useDevicePixelRatio } from '@vueuse/core'
import { feature } from 'topojson-client'
import { useNetworkStore } from '~/stores/network'
import { GetSelfGeoIP } from '~/lib/wailsjs/coordinator'

const containerRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const networkStore = useNetworkStore()
const { pixelRatio } = useDevicePixelRatio()

const geo = ref<{
  lat: number | null
  lon: number | null
  city: string
  country: string
  isp: string
}>({
  lat: null, lon: null, city: '', country: '', isp: '',
})

const ispName = computed(() => geo.value.isp)

// Matches AdexGlobe: the count of connections the network store is tracking.
const activeConnectionCount = computed(() => networkStore.activeConnections?.length ?? 0)

const endpointLabel = computed(() => {
  const { lat, lon } = geo.value
  if (lat === null || lon === null) return '—'
  const ns = lat >= 0 ? 'N' : 'S'
  const ew = lon >= 0 ? 'E' : 'W'
  return `${Math.abs(lat).toFixed(2)}${ns} ${Math.abs(lon).toFixed(2)}${ew}`
})

const locationLabel = computed(() => {
  const { city, country } = geo.value
  // 'Resolving...' rather than a dash, matching the classic globe: the
  // lookup is asynchronous and usually does land.
  return [city, country].filter(Boolean).join(', ') || 'Resolving...'
})

let worker: Worker | null = null

/** Read the theme accent. Only the main thread can see CSS variables. */
function accent(): string {
  const s = getComputedStyle(document.documentElement)
  const rgb = s.getPropertyValue('--accent-rgb').trim() || '170, 207, 209'
  return `rgb(${rgb})`
}

function post(msg: Record<string, unknown>) {
  worker?.postMessage(msg)
}

/** Send the container's current CSS size; the worker owns the backing store. */
function pushSize() {
  const el = containerRef.value
  if (!el || !worker) return
  const rect = el.getBoundingClientRect()
  if (rect.width < 8 || rect.height < 8) return
  post({ type: 'resize', width: rect.width, height: rect.height, dpr: pixelRatio.value || 1 })
}

async function refreshGeo() {
  try {
    const raw: any = await GetSelfGeoIP()
    const lat = Number(raw?.latitude ?? raw?.lat)
    const lon = Number(raw?.longitude ?? raw?.lon)
    geo.value = {
      lat: Number.isFinite(lat) ? lat : null,
      lon: Number.isFinite(lon) ? lon : null,
      city: String(raw?.city ?? ''),
      country: String(raw?.country ?? raw?.country_name ?? ''),
      isp: String(raw?.isp ?? raw?.org ?? ''),
    }
  } catch {
    // Geo lookup is decorative; the globe still renders without it.
  }
}

// Hand the endpoint to the worker as a unit vector so it needs no trig setup.
watch(() => [geo.value.lat, geo.value.lon] as const, ([lat, lon]) => {
  if (lat === null || lon === null) {
    post({ type: 'endpoint', endpoint: null })
    return
  }
  const DEG = Math.PI / 180
  const cl = Math.cos(lat * DEG)
  post({
    type: 'endpoint',
    endpoint: { x: cl * Math.cos(lon * DEG), y: cl * Math.sin(lon * DEG), z: Math.sin(lat * DEG) },
  })
})

// Suspend while hidden — a render loop nobody can see is wasted work, and
// the worker cannot observe document visibility itself.
const visibility = useDocumentVisibility()
watch(visibility, v => post({ type: 'visibility', visible: v === 'visible' }))

useResizeObserver(containerRef, pushSize)
watch(pixelRatio, pushSize)

onMounted(async () => {
  await nextTick()
  const canvas = canvasRef.value
  const container = containerRef.value
  if (!canvas || !container) return

  const topo = (await import('world-atlas/countries-110m.json')).default as any
  const land = feature(topo, topo.objects.countries) as any

  worker = new Worker(new URL('~/workers/geoGlobe.worker.ts', import.meta.url), { type: 'module' })

  const rect = container.getBoundingClientRect()
  const offscreen = canvas.transferControlToOffscreen()
  worker.postMessage({
    type: 'init',
    canvas: offscreen,
    features: land.features,
    width: rect.width,
    height: rect.height,
    dpr: pixelRatio.value || 1,
    colour: accent(),
  }, [offscreen])

  void refreshGeo()
})

onUnmounted(() => {
  post({ type: 'dispose' })
  worker?.terminate()
  worker = null
})
</script>

<style scoped>
.geo-globe-canvas {
  display: block;
  width: 100%;
  height: 100%;
}

/* The canvas is sized by ASPECT RATIO, not by viewport height. Competing vh
   caps here and on the panel in pages/index.vue clipped either the bottom of
   the sphere or the readouts underneath it, depending on which cap lost. A
   square box derived from the column's own width has no such conflict — the
   globe is circular, so its natural box IS square. */
.globe-container {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  max-height: 17vh;
  flex: 0 0 auto;
  overflow: hidden;
  margin: 0.3vh 0;
}

.mod-geo-globe {
  display: flex;
  flex-direction: column;
}

/* The readouts are the panel's reason for existing alongside the map; never
   let the globe compress them away. */
.geo-globe-meta {
  flex: 0 0 auto;
  font-size: 1.1vh;
  color: var(--color_accent, rgb(170, 207, 209));
}

.geo-row {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  opacity: 0.85;
}

.geo-row > span:first-child { opacity: 0.6; flex: none; }

/* ISP and location routinely overflow a column this narrow; clip them rather
   than letting the row wrap and shift the panel height. */
.geo-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
</style>
