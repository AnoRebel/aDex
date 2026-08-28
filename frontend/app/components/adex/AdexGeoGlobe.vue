<template>
  <div class="mod-panel mod-globe mod-geo-globe">
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
 * Country-outline globe.
 *
 * The alternative to the stylised dot-grid globe: an orthographic projection
 * of real country boundaries, so the endpoint marker sits on a recognisable
 * landmass. Selectable in Settings → Theme.
 *
 * Drawn on a 2D canvas rather than WebGL: the GPU path has crashed the
 * renderer on this platform, and a slowly rotating outline map does not
 * need it. The projection is computed inline instead of through d3-geo's
 * path generator — see loadWorld for the measurements behind that.
 *
 * Geometry is the 110m (low-resolution) world atlas: ~105 KB, which is the
 * right trade for a panel this size. The 50m and 10m files are 7x and 34x
 * larger for detail that would be invisible here.
 */
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useDocumentVisibility, useResizeObserver, useRafFn, useDevicePixelRatio } from '@vueuse/core'
import { feature } from 'topojson-client'
import { nowMs } from '~/utils/now'
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


// Tilt of the globe's axis towards the viewer, in radians. Matches the
// -15 degree pitch the d3 version used.
const TILT = -15 * Math.PI / 180

/** Current rotation about the polar axis, in degrees. Driven by the clock
 *  in the animation loop below, and read by draw(). */
let rotation = 0

/**
 * Load the world outline once and flatten it into typed arrays.
 *
 * We deliberately do NOT keep the GeoJSON around to hand to d3's geoPath on
 * every frame. Measured on this data (177 countries / 286 rings / 10,587
 * points), geoPath costs ~21ms per redraw while a raw pass over the same
 * coordinates costs ~0.14ms — the difference is d3's generic per-point
 * stream (clipping, adaptive resampling, transform plumbing), which an
 * orthographic globe at this size does not need. Projecting the points
 * directly is ~69x faster and is what keeps this panel off the main thread's
 * critical path.
 *
 * Longitude/latitude are pre-converted to radians, and sin/cos of latitude
 * are precomputed, because those are loop-invariant across frames.
 */
let ringOffsets: Int32Array | null = null
let cosLatCosLon: Float64Array | null = null
let cosLatSinLon: Float64Array | null = null
let sinLatArr: Float64Array | null = null

async function loadWorld() {
  if (ringOffsets) return

  const topo = (await import('world-atlas/countries-110m.json')).default as any
  const land = feature(topo, topo.objects.countries) as any

  const xs: number[] = []
  const ys: number[] = []
  const zs: number[] = []
  const offsets: number[] = []

  const DEG = Math.PI / 180
  for (const f of land.features) {
    const g = f.geometry
    if (!g) continue
    const polys =
      g.type === 'Polygon' ? [g.coordinates]
      : g.type === 'MultiPolygon' ? g.coordinates
      : []
    for (const poly of polys) {
      for (const ring of poly) {
        offsets.push(xs.length)
        for (const c of ring) {
          // Store each point as a unit vector on the sphere; rotating a
          // vector per frame is a handful of multiplies.
          const lon = c[0] * DEG
          const lat = c[1] * DEG
          const cosLat = Math.cos(lat)
          xs.push(cosLat * Math.cos(lon))
          ys.push(cosLat * Math.sin(lon))
          zs.push(Math.sin(lat))
        }
      }
    }
  }
  offsets.push(xs.length)

  cosLatCosLon = Float64Array.from(xs)
  cosLatSinLon = Float64Array.from(ys)
  sinLatArr = Float64Array.from(zs)
  ringOffsets = Int32Array.from(offsets)
}

function accent(): string {
  const s = getComputedStyle(document.documentElement)
  const rgb = s.getPropertyValue('--accent-rgb').trim() || '170, 207, 209'
  return `rgb(${rgb})`
}

function draw() {
  const canvas = canvasRef.value
  const container = containerRef.value
  if (!canvas || !container || !ringOffsets || !cosLatCosLon || !cosLatSinLon || !sinLatArr) return

  const rect = container.getBoundingClientRect()
  if (rect.width < 8 || rect.height < 8) return

  const dpr = pixelRatio.value || 1
  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr
  canvas.style.width = `${rect.width}px`
  canvas.style.height = `${rect.height}px`

  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, rect.width, rect.height)

  const cx = rect.width / 2
  const cy = rect.height / 2
  const radius = Math.min(rect.width, rect.height) / 2 - 4
  if (radius <= 0) return

  const colour = accent()

  // Orthographic projection, computed inline.
  //
  // Each point is already a unit vector [cosφcosλ, cosφsinλ, sinφ] with +z
  // at the north pole. A frame is: spin about z by `lambda`, tilt about y by
  // TILT, drop the far hemisphere, and scale to pixels. The viewer looks
  // down +x. See loadWorld for why this is not delegated to d3's geoPath.
  const lambda = rotation * (Math.PI / 180)
  const sinL = Math.sin(lambda)
  const cosL = Math.cos(lambda)
  const sinT = Math.sin(TILT)
  const cosT = Math.cos(TILT)

  // Ocean disc.
  ctx.beginPath()
  ctx.arc(cx, cy, radius, 0, Math.PI * 2)
  ctx.strokeStyle = colour
  ctx.globalAlpha = 0.35
  ctx.lineWidth = 1
  ctx.stroke()

  // Graticule, drawn analytically rather than as projected geometry.
  ctx.globalAlpha = 0.12
  drawGraticule(ctx, cx, cy, radius, sinL, cosL, sinT, cosT)

  // Country outlines.
  ctx.beginPath()
  const offsets = ringOffsets
  const px = cosLatCosLon
  const py = cosLatSinLon
  const pz = sinLatArr

  for (let r = 0; r < offsets.length - 1; r++) {
    let penDown = false
    for (let i = offsets[r]!; i < offsets[r + 1]!; i++) {
      const x0 = px[i]!
      const y0 = py[i]!
      const z0 = pz[i]!

      // 1. Spin about the polar (z) axis.
      const x1 = x0 * cosL + y0 * sinL
      const y1 = y0 * cosL - x0 * sinL

      // 2. Tilt about the SCREEN-HORIZONTAL (y) axis, which mixes x and z.
      //    Rotating x with z here is what keeps the pole at the top of the
      //    disc; mixing the wrong pair pitched the globe over and showed it
      //    pole-on.
      const xv = x1 * cosT + z0 * sinT
      const zv = z0 * cosT - x1 * sinT

      // 3. The viewer looks down +x, so xv is depth: negative is the far side.
      if (xv < 0) {
        // Behind the globe: break the stroke rather than drawing a chord
        // straight across the visible face.
        penDown = false
        continue
      }

      // Screen right is y, screen up is z.
      const sx = cx + radius * y1
      const sy = cy - radius * zv
      if (penDown) {
        ctx.lineTo(sx, sy)
      } else {
        ctx.moveTo(sx, sy)
        penDown = true
      }
    }
  }
  ctx.globalAlpha = 0.75
  ctx.lineWidth = 0.6
  ctx.strokeStyle = colour
  ctx.stroke()

  // Endpoint marker, drawn only when it is on the visible hemisphere.
  // An orthographic projection maps both hemispheres onto the same disc, so
  // without this test a marker in East Africa was painted over the
  // Caribbean.
  const { lat, lon } = geo.value
  if (lat !== null && lon !== null) {
    const DEG = Math.PI / 180
    const cl = Math.cos(lat * DEG)
    const x0 = cl * Math.cos(lon * DEG)
    const y0 = cl * Math.sin(lon * DEG)
    const z0 = Math.sin(lat * DEG)

    const x1 = x0 * cosL + y0 * sinL
    const y1 = y0 * cosL - x0 * sinL
    const xv = x1 * cosT + z0 * sinT
    const zv = z0 * cosT - x1 * sinT

    if (xv >= 0) {
      const mx = cx + radius * y1
      const my = cy - radius * zv
      ctx.globalAlpha = 1
      ctx.fillStyle = colour
      ctx.beginPath(); ctx.arc(mx, my, 3, 0, Math.PI * 2); ctx.fill()
      ctx.beginPath(); ctx.arc(mx, my, 7, 0, Math.PI * 2)
      ctx.globalAlpha = 0.4; ctx.strokeStyle = colour; ctx.stroke()
    }
  }

  ctx.globalAlpha = 1
}

/**
 * Graticule: meridians and parallels at 20-degree spacing.
 *
 * Drawn by stepping along each line and projecting, same as the coastlines,
 * so it costs a few hundred points rather than a geoPath traversal.
 */
function drawGraticule(
  ctx: CanvasRenderingContext2D,
  cx: number, cy: number, radius: number,
  sinL: number, cosL: number, sinT: number, cosT: number,
) {
  ctx.beginPath()
  const DEG = Math.PI / 180
  const STEP = 6 * DEG

  const project = (lon: number, lat: number): [number, number] | null => {
    const cl = Math.cos(lat)
    const x0 = cl * Math.cos(lon)
    const y0 = cl * Math.sin(lon)
    const z0 = Math.sin(lat)
    const x1 = x0 * cosL + y0 * sinL
    const y1 = y0 * cosL - x0 * sinL
    const xv = x1 * cosT + z0 * sinT
    const zv = z0 * cosT - x1 * sinT
    if (xv < 0) return null
    return [cx + radius * y1, cy - radius * zv]
  }

  // Meridians.
  for (let lonDeg = -180; lonDeg < 180; lonDeg += 20) {
    const lon = lonDeg * DEG
    let pen = false
    for (let lat = -Math.PI / 2; lat <= Math.PI / 2 + 1e-9; lat += STEP) {
      const pt = project(lon, lat)
      if (!pt) { pen = false; continue }
      if (pen) ctx.lineTo(pt[0], pt[1]); else { ctx.moveTo(pt[0], pt[1]); pen = true }
    }
  }

  // Parallels.
  for (let latDeg = -80; latDeg <= 80; latDeg += 20) {
    const lat = latDeg * DEG
    let pen = false
    for (let lon = -Math.PI; lon <= Math.PI + 1e-9; lon += STEP) {
      const pt = project(lon, lat)
      if (!pt) { pen = false; continue }
      if (pen) ctx.lineTo(pt[0], pt[1]); else { ctx.moveTo(pt[0], pt[1]); pen = true }
    }
  }

  ctx.stroke()
}

/**
 * Rotation is driven by elapsed TIME, not by frame count.
 *
 * Redrawing this globe costs ~13ms of pure geometry traversal per frame
 * (measured: 110m countries + graticule + sphere), so the loop cannot be
 * assumed to hit 60fps — it shares the main thread with the terminal and
 * the charts. A fixed per-frame step therefore makes the rotation speed a
 * function of machine load: it crawled to a near-standstill in practice.
 *
 * Deriving the angle from the clock keeps the globe turning at the same
 * visible rate no matter what frame rate we actually achieve; a slow
 * machine drops frames instead of slowing down.
 */
const DEGREES_PER_SECOND = 6

/**
 * Redraw cadence.
 *
 * Drawing this globe costs ~13-20ms of geometry traversal per frame (110m
 * country outlines + graticule), so redrawing on every animation frame
 * saturates the main thread and makes the whole interface feel laggy — the
 * terminal, the charts and the settings modal all share it.
 *
 * At 20fps each redraw advances it 0.3 degrees, which reads as continuous
 * motion at this size. The rotation angle comes from the clock, so frames
 * are dropped rather than the motion being slowed.
 */
const TARGET_FPS = 20
const MIN_FRAME_INTERVAL_MS = 1000 / TARGET_FPS

let startedAt: number | null = null
let lastDrawAt = 0

// useRafFn owns the loop lifecycle (including teardown on unmount), so the
// component no longer tracks a raw frame handle.
const raf = useRafFn(() => {
  const t = nowMs()
  if (startedAt === null) startedAt = t
  if (t - lastDrawAt < MIN_FRAME_INTERVAL_MS) return
  lastDrawAt = t

  rotation = ((t - startedAt) / 1000 * DEGREES_PER_SECOND) % 360
  draw()
}, { immediate: false })

function start() {
  if (raf.isActive.value) return
  // Rebase so the globe resumes from where it stopped rather than jumping
  // forward by however long the window was hidden.
  startedAt = nowMs() - (rotation / DEGREES_PER_SECOND) * 1000
  raf.resume()
}

function stop() {
  raf.pause()
}

// Suspend while hidden — a render loop nobody can see is the single largest
// idle cost in the interface.
const visibility = useDocumentVisibility()
watch(visibility, v => (v === 'visible' ? start() : stop()))

useResizeObserver(containerRef, () => draw())
watch(() => networkStore.metrics, () => draw())

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

onMounted(async () => {
  await loadWorld()
  await nextTick()
  draw()
  start()
  void refreshGeo()
})
</script>

<style scoped>
.geo-globe-canvas {
  display: block;
  width: 100%;
  height: 100%;
}

/* The canvas is sized by ASPECT RATIO, not by viewport height.
   Competing vh caps here and on the panel in pages/index.vue were the cause
   of a long back-and-forth: whichever cap lost clipped either the bottom of
   the sphere or the readouts underneath it. A square box derived from the
   column's own width has no such conflict — the globe is circular, so its
   natural box IS square, and the panel simply sizes to its content. */
.globe-container {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  /* Cap it for very wide columns so the globe cannot dominate the column,
     and floor it so it stays legible in narrow ones. */
  max-height: 17vh;
  flex: 0 0 auto;
  overflow: hidden;
  margin: 0.3vh 0;
}

.mod-geo-globe {
  display: flex;
  flex-direction: column;
}

/* The readouts are the panel's reason for existing alongside the map;
   never let the globe compress them away. */
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

/* ISP and location routinely overflow a column this narrow; clip them
   rather than letting the row wrap and shift the panel height. */
.geo-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
</style>
