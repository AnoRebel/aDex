/**
 * Country-outline globe renderer.
 *
 * Runs the entire render loop off the main thread against an OffscreenCanvas
 * transferred from the component. The main thread never touches the canvas
 * again — it only posts state changes (size, colour, endpoint, visibility),
 * so a redraw cannot compete with the terminal, the charts or input handling.
 *
 * The worker has no DOM: it cannot read element sizes or CSS variables, so
 * the component measures those and sends them over. That is the documented
 * trade for OffscreenCanvas and the reason `resize` is an explicit message.
 */

/** Tilt of the globe's axis towards the viewer, in radians. */
const TILT = -15 * Math.PI / 180

/** Degrees of longitude per second. */
const DEGREES_PER_SECOND = 6

/**
 * Redraw cadence.
 *
 * A globe turning 6°/second does not need 60 updates per second to look
 * smooth — at 20fps each frame advances it 0.3°, which reads as continuous.
 * The angle is derived from the clock, so a dropped frame skips ahead rather
 * than slowing the rotation down.
 */
const TARGET_FPS = 20
const MIN_FRAME_INTERVAL_MS = 1000 / TARGET_FPS

/** Graticule spacing, and the step used when walking along each line. */
const GRATICULE_SPACING_DEG = 20
const GRATICULE_STEP = 6 * Math.PI / 180

let canvas: OffscreenCanvas | null = null
let ctx: OffscreenCanvasRenderingContext2D | null = null

// Geometry as unit vectors on the sphere, flattened into typed arrays with
// ring boundaries alongside. Rotating a unit vector is a few multiplies, so
// a whole frame stays in the fractions-of-a-millisecond range.
let vx: Float64Array | null = null
let vy: Float64Array | null = null
let vz: Float64Array | null = null
let ringOffsets: Int32Array | null = null

let cssWidth = 0
let cssHeight = 0
let colour = 'rgb(170, 207, 209)'

// Endpoint marker, in radians, or null while the lookup is pending.
let endpoint: { x: number; y: number; z: number } | null = null

let rotation = 0
let startedAt: number | null = null
let lastDrawAt = 0
let running = false
let rafId: number | null = null

/** Flatten GeoJSON rings into unit vectors, once, at startup. */
function ingest(features: any[]) {
  const xs: number[] = []
  const ys: number[] = []
  const zs: number[] = []
  const offsets: number[] = []
  const DEG = Math.PI / 180

  for (const f of features) {
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

  vx = Float64Array.from(xs)
  vy = Float64Array.from(ys)
  vz = Float64Array.from(zs)
  ringOffsets = Int32Array.from(offsets)
}

function draw() {
  if (!canvas || !ctx || !vx || !vy || !vz || !ringOffsets) return
  if (cssWidth < 8 || cssHeight < 8) return

  ctx.clearRect(0, 0, cssWidth, cssHeight)

  const cx = cssWidth / 2
  const cy = cssHeight / 2
  const radius = Math.min(cssWidth, cssHeight) / 2 - 4
  if (radius <= 0) return

  // Each point is a unit vector [cosφcosλ, cosφsinλ, sinφ] with +z at the
  // north pole. A frame spins about z by `lambda`, tilts about y by TILT,
  // drops the far hemisphere, and scales to pixels. The viewer looks down
  // +x, so screen right is y and screen up is z.
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

  // Graticule.
  ctx.globalAlpha = 0.12
  drawGraticule(ctx, cx, cy, radius, sinL, cosL, sinT, cosT)

  // Country outlines. Every ring goes into ONE path before a single stroke:
  // batching this way is markedly faster than stroking per ring.
  ctx.beginPath()
  for (let r = 0; r < ringOffsets.length - 1; r++) {
    let penDown = false
    for (let i = ringOffsets[r]!; i < ringOffsets[r + 1]!; i++) {
      const x1 = vx[i]! * cosL + vy[i]! * sinL
      const y1 = vy[i]! * cosL - vx[i]! * sinL
      const z0 = vz[i]!

      const xv = x1 * cosT + z0 * sinT
      if (xv < 0) {
        // Behind the globe: break the stroke rather than drawing a chord
        // across the visible face.
        penDown = false
        continue
      }
      const zv = z0 * cosT - x1 * sinT

      const sx = cx + radius * y1
      const sy = cy - radius * zv
      if (penDown) ctx.lineTo(sx, sy)
      else { ctx.moveTo(sx, sy); penDown = true }
    }
  }
  ctx.globalAlpha = 0.75
  ctx.lineWidth = 0.6
  ctx.strokeStyle = colour
  ctx.stroke()

  // Endpoint marker, only when it is on the visible hemisphere — an
  // orthographic projection maps both hemispheres onto the same disc.
  if (endpoint) {
    const x1 = endpoint.x * cosL + endpoint.y * sinL
    const y1 = endpoint.y * cosL - endpoint.x * sinL
    const xv = x1 * cosT + endpoint.z * sinT
    if (xv >= 0) {
      const zv = endpoint.z * cosT - x1 * sinT
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

function drawGraticule(
  ctx: OffscreenCanvasRenderingContext2D,
  cx: number, cy: number, radius: number,
  sinL: number, cosL: number, sinT: number, cosT: number,
) {
  ctx.beginPath()
  const DEG = Math.PI / 180

  const project = (lon: number, lat: number): [number, number] | null => {
    const cl = Math.cos(lat)
    const x0 = cl * Math.cos(lon)
    const y0 = cl * Math.sin(lon)
    const z0 = Math.sin(lat)
    const x1 = x0 * cosL + y0 * sinL
    const y1 = y0 * cosL - x0 * sinL
    const xv = x1 * cosT + z0 * sinT
    if (xv < 0) return null
    const zv = z0 * cosT - x1 * sinT
    return [cx + radius * y1, cy - radius * zv]
  }

  for (let lonDeg = -180; lonDeg < 180; lonDeg += GRATICULE_SPACING_DEG) {
    const lon = lonDeg * DEG
    let pen = false
    for (let lat = -Math.PI / 2; lat <= Math.PI / 2 + 1e-9; lat += GRATICULE_STEP) {
      const pt = project(lon, lat)
      if (!pt) { pen = false; continue }
      if (pen) ctx.lineTo(pt[0], pt[1]); else { ctx.moveTo(pt[0], pt[1]); pen = true }
    }
  }

  for (let latDeg = -80; latDeg <= 80; latDeg += GRATICULE_SPACING_DEG) {
    const lat = latDeg * DEG
    let pen = false
    for (let lon = -Math.PI; lon <= Math.PI + 1e-9; lon += GRATICULE_STEP) {
      const pt = project(lon, lat)
      if (!pt) { pen = false; continue }
      if (pen) ctx.lineTo(pt[0], pt[1]); else { ctx.moveTo(pt[0], pt[1]); pen = true }
    }
  }

  ctx.stroke()
}

function tick() {
  if (!running) return
  const t = performance.now()
  if (startedAt === null) startedAt = t

  if (t - lastDrawAt >= MIN_FRAME_INTERVAL_MS) {
    lastDrawAt = t
    rotation = ((t - startedAt) / 1000 * DEGREES_PER_SECOND) % 360
    draw()
  }
  rafId = requestAnimationFrame(tick)
}

function start() {
  if (running) return
  running = true
  // Rebase so the globe resumes where it stopped rather than jumping forward
  // by however long it was suspended.
  startedAt = performance.now() - (rotation / DEGREES_PER_SECOND) * 1000
  rafId = requestAnimationFrame(tick)
}

function stop() {
  running = false
  if (rafId !== null) { cancelAnimationFrame(rafId); rafId = null }
}

self.onmessage = (e: MessageEvent) => {
  const msg = e.data
  switch (msg?.type) {
    case 'init': {
      canvas = msg.canvas as OffscreenCanvas
      ctx = canvas.getContext('2d')
      ingest(msg.features)
      applySize(msg.width, msg.height, msg.dpr)
      colour = msg.colour ?? colour
      draw()
      start()
      break
    }
    case 'resize':
      applySize(msg.width, msg.height, msg.dpr)
      draw()
      break
    case 'colour':
      colour = msg.colour
      draw()
      break
    case 'endpoint':
      endpoint = msg.endpoint
      draw()
      break
    case 'visibility':
      if (msg.visible) start(); else stop()
      break
    case 'dispose':
      stop()
      self.close()
      break
  }
}

/**
 * The worker cannot measure the DOM, so the component sends CSS pixel size
 * and devicePixelRatio. Backing-store size and the DPR transform are applied
 * here to keep the drawing code in CSS pixels.
 */
function applySize(width: number, height: number, dpr: number) {
  if (!canvas || !ctx) return
  cssWidth = width
  cssHeight = height
  canvas.width = Math.max(1, Math.round(width * dpr))
  canvas.height = Math.max(1, Math.round(height * dpr))
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
}

export {}
