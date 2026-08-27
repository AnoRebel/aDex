<template>
  <div class="mod-panel mod-netstat">
    <!-- Header -->
    <div class="section-title">
      <span class="title-left">NETWORK STATUS</span>
      <span class="title-right">{{ primaryInterface }}</span>
    </div>

    <!-- Status rows -->
    <div class="netstat-body">
      <div class="netstat-row">
        <span class="netstat-label">STATE</span>
        <span class="netstat-value" :class="isOnline ? 'state-online' : 'state-offline'">
          {{ isOnline ? 'Online' : 'Offline' }}
        </span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">IPv4</span>
        <span class="netstat-value" :title="ipv4Address">{{ ipv4Address }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">IPv6</span>
        <span class="netstat-value netstat-value-truncated" :title="ipv6Address">{{ ipv6Address }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">PING</span>
        <span class="netstat-value">{{ pingLatency }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">TYPE</span>
        <span class="netstat-value">{{ interfaceType }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">SIGNAL</span>
        <span class="netstat-value">{{ signalStrength }}</span>
      </div>
    </div>

    <!-- The GLOBAL NETWORK MAP block (endpoint lat/lon/location/isp)
         used to live here AND inside AdexGlobe.vue, duplicating the
         data. The geo block is now consolidated into AdexGlobe so the
         user sees one canonical world-view block. -->
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useStorage, useIntervalFn } from '@vueuse/core'
import { useNetworkStore } from '~/stores/network'
import { MeasureLatency, GetSelfGeoIP } from '~/lib/wailsjs/coordinator'

const networkStore = useNetworkStore()

const pingMs = ref<number | null>(null)

// GeoIP data fetched locally
const geoData = ref<{
  lat: number | null
  lon: number | null
  city: string
  region: string
  country: string
  isp: string
}>({
  lat: null,
  lon: null,
  city: '',
  region: '',
  country: '',
  isp: ''
})

// Computed properties derived from the network store
const isOnline = computed(() => {
  if (networkStore.metrics?.interfaces) {
    return networkStore.metrics.interfaces.some((iface: any) => iface.isUp)
  }
  return navigator.onLine
})

// Single source of truth for the panel's "active" interface. All the
// derived fields (IP addresses, interface type, signal strength) hang
// off this so a one-line change in pickInterface() (e.g. picking the
// preferred adapter from settings) flows everywhere.
const activeInterface = computed(() => pickInterface())

const primaryInterface = computed(() => activeInterface.value?.name || 'eth0')

// Strip CIDR suffix (gopsutil returns '192.168.0.155/24'). Accept both
// `ipAddresses` (the live model field) and `addresses` (older shape)
// — different backend revs have used both.
function ifaceAddrs(iface: any): string[] {
  if (!iface) return []
  const list = iface.ipAddresses ?? iface.addresses ?? []
  return Array.isArray(list)
    ? list.map((a: string) => (a ? a.split('/')[0] : '')).filter(Boolean)
    : []
}

const ipv4Address = computed(() => {
  const iface = activeInterface.value
  const addrs = ifaceAddrs(iface)
  const ipv4 = addrs.find((a) => !a.includes(':'))
  if (ipv4) return ipv4
  return iface?.ipv4 || '0.0.0.0'
})

const ipv6Address = computed(() => {
  const iface = activeInterface.value
  const addrs = ifaceAddrs(iface)
  const ipv6 = addrs.find((a) => a.includes(':'))
  if (ipv6) return ipv6
  return iface?.ipv6 || '::1'
})

const interfaceType = computed(() => {
  const name = primaryInterface.value.toLowerCase()
  if (name.startsWith('wl') || name.startsWith('wifi') || name.startsWith('wlan')) return 'Wi-Fi'
  if (name.startsWith('eth') || name.startsWith('en')) return 'Ethernet'
  if (name.startsWith('lo')) return 'Loopback'
  if (name.startsWith('tun') || name.startsWith('wg')) return 'VPN'
  return 'Unknown'
})

const signalStrength = computed(() => {
  const iface = activeInterface.value
  if (iface?.signalStrength != null) return `${iface.signalStrength} dBm`
  if (iface?.speed) return `${iface.speed} Mbps`
  return isOnline.value ? 'Good' : 'N/A'
})

const pingLatency = computed(() => {
  // Coerce rather than trusting the declared type. A backend call can resolve
  // to a non-number when the IPC peer is absent (for example when the frontend
  // is opened in a plain browser for testing), and `.toFixed` on that value
  // throws inside the computed — which breaks this component's reactivity, not
  // just the ping readout.
  const ms = typeof pingMs.value === 'number' ? pingMs.value : Number(pingMs.value)
  if (pingMs.value !== null && Number.isFinite(ms)) {
    return `${ms.toFixed(1)} ms`
  }
  return isOnline.value ? 'Measuring...' : 'N/A'
})

const endpointLat = computed(() => {
  if (geoData.value.lat !== null) return geoData.value.lat.toFixed(4)
  return '0.0000'
})

const endpointLon = computed(() => {
  if (geoData.value.lon !== null) return geoData.value.lon.toFixed(4)
  return '0.0000'
})

const endpointLocation = computed(() => {
  const parts = []
  if (geoData.value.city) parts.push(geoData.value.city)
  if (geoData.value.region) parts.push(geoData.value.region)
  if (geoData.value.country) parts.push(geoData.value.country)
  return parts.length > 0 ? parts.join(', ') : 'Unknown'
})

const endpointISP = computed(() => {
  return geoData.value.isp || 'Unknown'
})

// Reactive settings reader for the network panel. Everything lives on
// the shared `adex-settings` useStorage key (so Settings → Network is
// the single source of truth, and the panel re-renders the moment the
// user changes any field).
const adexSettings = useStorage<{
  network?: {
    pingTarget?: string
    adapter?: string
    pingIntervalMs?: number
    geoipEnabled?: boolean
    geoipEndpoint?: string
  }
}>('adex-settings', {})

const pingTarget = computed(() =>
  (adexSettings.value?.network?.pingTarget ?? '').trim()
)

// Poll cadence — clamped to a sane range so a corrupt storage value
// (negative, zero, NaN) doesn't lock useIntervalFn into a tight loop.
const pingIntervalMs = computed(() => {
  const raw = Number(adexSettings.value?.network?.pingIntervalMs ?? 2000)
  if (!Number.isFinite(raw)) return 2000
  return Math.min(30_000, Math.max(1_000, Math.floor(raw)))
})

const geoipEnabled = computed(() =>
  adexSettings.value?.network?.geoipEnabled ?? true,
)

const geoipEndpoint = computed(() =>
  (adexSettings.value?.network?.geoipEndpoint ?? '').trim(),
)

// User-selected adapter name. Empty → auto (first up, non-loopback).
// When the saved adapter has disappeared (wifi unplugged), we fall
// back to auto so the panel keeps working instead of pinning a dead
// device.
const preferredAdapter = computed(() =>
  (adexSettings.value?.network?.adapter ?? '').trim()
)

function pickInterface(): any | null {
  const interfaces = networkStore.metrics?.interfaces
  if (!interfaces || interfaces.length === 0) return null
  const pref = preferredAdapter.value
  if (pref) {
    const match = interfaces.find((i: any) => i.name === pref)
    if (match) return match
  }
  return (
    interfaces.find((i: any) => i.isUp && i.name !== 'lo') ??
    interfaces[0]
  )
}

// Backend-driven ping: TCP handshake to the configured target. Falls
// back to -1 on timeout/error and we render 'N/A'. Replaces the
// previous /favicon.ico image-trick which couldn't resolve under the
// `wails://` scheme — that's why PING was stuck on "Measuring..."
// forever in earlier builds.
const measurePing = async () => {
  try {
    const raw = await MeasureLatency(pingTarget.value)
    // Guard the shape: a non-numeric result means the call did not reach the
    // backend, which should read as "no measurement" rather than poisoning the
    // computed above.
    const ms = typeof raw === 'number' ? raw : Number(raw)
    pingMs.value = Number.isFinite(ms) && ms >= 0 ? ms : null
  } catch {
    pingMs.value = null
  }
}

// Fetch GeoIP. First try the backend (kamero → ipify+iplocate
// failover). If THAT also returns nothing — possible if the Wails
// HTTP client is blocked by sandbox/firewall — fall back to direct
// frontend calls from the WebView (which has its own networking).
const fetchGeoData = async () => {
  // User opted out via Settings → Network → Enable GeoIP lookup. Bail
  // before any IPC/HTTP so we keep the network surface quiet (matters
  // for air-gapped systems or anyone who'd rather not phone home).
  if (!geoipEnabled.value) return
  if (geoData.value.lat !== null && geoData.value.lat !== 0) return
  // 1. Backend path — preferred because it caches and tries multiple
  //    providers without us reimplementing that logic.
  try {
    const g = await GetSelfGeoIP()
    if (g && (g.latitude != null || g.city)) {
      geoData.value = {
        lat: typeof g.latitude === 'number' ? g.latitude : null,
        lon: typeof g.longitude === 'number' ? g.longitude : null,
        city: g.city ?? '',
        region: g.region ?? '',
        country: g.country ?? '',
        isp: g.isp ?? '',
      }
      return
    }
    console.warn('[netstat] backend GetSelfGeoIP returned empty, falling back to WebView fetch')
  } catch (err) {
    console.warn('[netstat] backend GetSelfGeoIP threw, falling back:', err)
  }
  // 2. Frontend fallback — WebView fetch direct to the same providers.
  //    Respects the user's custom endpoint if they configured one.
  await fetchGeoFromWebview()
}

// Direct WebView fetch as a last resort. Tries kamero first, then
// ipify→iplocate. Keeps the same shape conversion the backend does so
// the rest of the panel sees consistent data.
async function fetchGeoFromWebview() {
  // Honor a user-supplied custom endpoint if they set one in Settings →
  // Network → GeoIP endpoint. Empty falls through to our default.
  const customUrl = geoipEndpoint.value
  const url = customUrl || 'https://geo.kamero.ai/api/geo'
  try {
    const res = await fetch(url, {
      headers: { Accept: 'application/json' },
    })
    if (res.ok) {
      const j = await res.json()
      if (j && (j.latitude != null || j.city)) {
        const lat = parseFloat(String(j.latitude))
        const lon = parseFloat(String(j.longitude))
        geoData.value = {
          lat: Number.isFinite(lat) ? lat : null,
          lon: Number.isFinite(lon) ? lon : null,
          city: j.city ?? '',
          region: j.countryRegion ?? '',
          country: j.country ?? '',
          isp: '',
        }
        return
      }
    }
  } catch { /* try ipify chain */ }
  try {
    const ipRes = await fetch('https://api.ipify.org?format=json')
    if (!ipRes.ok) return
    const ipJ = await ipRes.json()
    if (!ipJ?.ip) return
    const geoRes = await fetch(`https://iplocate.io/api/lookup/${ipJ.ip}`, {
      headers: { Accept: 'application/json' },
    })
    if (!geoRes.ok) return
    const j = await geoRes.json()
    geoData.value = {
      lat: typeof j.latitude === 'number' ? j.latitude : null,
      lon: typeof j.longitude === 'number' ? j.longitude : null,
      city: j.city ?? '',
      region: j.subdivision ?? '',
      country: j.country_code ?? '',
      isp: j?.asn?.name ?? '',
    }
  } catch (err) {
    console.warn('[netstat] WebView geo fallback failed:', err)
  }
}

// Polling function. We intentionally fetch connections from the same
// poll because the globe mod displays the count and otherwise nothing
// would populate it (CONNECTIONS would stay at 0 forever).
const poll = async () => {
  try {
    await networkStore.fetchMetrics()
    measurePing()
    await fetchGeoData()
    networkStore.fetchConnections?.().catch(() => undefined)
  } catch {
    // Silently continue polling
  }
}

// Poll cadence is user-configurable via Settings → Network → Ping
// cadence. useIntervalFn re-arms automatically when we pass a reactive
// ref as the interval — so changing the slider in settings flips the
// cadence live without remount. Clamped to 1-30s upstream so we can't
// hit a tight loop.
useIntervalFn(poll, pingIntervalMs, { immediate: true, immediateCallback: true })
</script>

<style scoped>
.mod-netstat {
  font-size: 1.1vh;
}

.netstat-body {
  padding: 0.2vh 0;
}

.netstat-row {
  display: flex;
  justify-content: space-between;
  padding: 0.15vh 0;
}

.netstat-label {
  text-transform: uppercase;
  opacity: 0.6;
  letter-spacing: 0.05em;
}

.netstat-value {
  text-align: right;
  font-family: var(--font_main);
}

.netstat-value-truncated {
  max-width: 10vw;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.state-online {
  color: var(--success, #10b981);
}

.state-offline {
  color: var(--error, #ef4444);
}

.netstat-section-label {
  font-size: 1vh;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  opacity: 0.5;
  padding: 0.6vh 0 0.2vh 0;
  border-top: var(--border_width) solid rgba(var(--color_r), var(--color_g), var(--color_b), 0.15);
  margin-top: 0.4vh;
}
</style>
