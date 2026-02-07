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
        <span class="netstat-value">{{ ipv4Address }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">IPv6</span>
        <span class="netstat-value netstat-value-truncated">{{ ipv6Address }}</span>
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

    <!-- Global Network Map label -->
    <div class="netstat-section-label">GLOBAL NETWORK MAP</div>

    <!-- Endpoint coordinates -->
    <div class="netstat-body">
      <div class="netstat-row">
        <span class="netstat-label">ENDPOINT LAT</span>
        <span class="netstat-value">{{ endpointLat }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">ENDPOINT LON</span>
        <span class="netstat-value">{{ endpointLon }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">LOCATION</span>
        <span class="netstat-value netstat-value-truncated">{{ endpointLocation }}</span>
      </div>
      <div class="netstat-row">
        <span class="netstat-label">ISP</span>
        <span class="netstat-value netstat-value-truncated">{{ endpointISP }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useNetworkStore } from '~/stores/network'

const networkStore = useNetworkStore()

// Local polling state
let pollTimer: ReturnType<typeof setInterval> | null = null
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

const primaryInterface = computed(() => {
  const interfaces = networkStore.metrics?.interfaces
  if (interfaces && interfaces.length > 0) {
    const active = interfaces.find((iface: any) => iface.isUp && iface.name !== 'lo')
    return active?.name || interfaces[0].name || 'eth0'
  }
  return 'eth0'
})

const ipv4Address = computed(() => {
  const interfaces = networkStore.metrics?.interfaces
  if (interfaces) {
    const active = interfaces.find((iface: any) => iface.isUp && iface.name !== 'lo')
    if (active?.addresses) {
      const ipv4 = active.addresses.find((addr: string) => addr && !addr.includes(':'))
      if (ipv4) return ipv4
    }
    if (active?.ipv4) return active.ipv4
  }
  return '0.0.0.0'
})

const ipv6Address = computed(() => {
  const interfaces = networkStore.metrics?.interfaces
  if (interfaces) {
    const active = interfaces.find((iface: any) => iface.isUp && iface.name !== 'lo')
    if (active?.addresses) {
      const ipv6 = active.addresses.find((addr: string) => addr && addr.includes(':'))
      if (ipv6) return ipv6
    }
    if (active?.ipv6) return active.ipv6
  }
  return '::1'
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
  const interfaces = networkStore.metrics?.interfaces
  if (interfaces) {
    const active = interfaces.find((iface: any) => iface.isUp && iface.name !== 'lo')
    if (active?.signalStrength != null) return `${active.signalStrength} dBm`
    if (active?.speed) return `${active.speed} Mbps`
  }
  return isOnline.value ? 'Good' : 'N/A'
})

const pingLatency = computed(() => {
  if (pingMs.value !== null) {
    return `${pingMs.value.toFixed(1)} ms`
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

// Simulate ping measurement (actual ping would come from Go backend)
const measurePing = () => {
  const start = performance.now()
  const img = new Image()
  img.onload = () => {
    pingMs.value = performance.now() - start
  }
  img.onerror = () => {
    // Still measure the round-trip even on error
    const elapsed = performance.now() - start
    if (elapsed < 5000) {
      pingMs.value = elapsed
    } else {
      pingMs.value = null
    }
  }
  // Use a tiny request to measure latency
  img.src = `/favicon.ico?t=${Date.now()}`
}

// Fetch GeoIP data once
const fetchGeoData = async () => {
  try {
    // Try the backend service first via the network store metrics
    const metrics = networkStore.metrics
    if (metrics?.geoip) {
      geoData.value = {
        lat: metrics.geoip.latitude || null,
        lon: metrics.geoip.longitude || null,
        city: metrics.geoip.city || '',
        region: metrics.geoip.region || '',
        country: metrics.geoip.country || '',
        isp: metrics.geoip.isp || metrics.geoip.org || ''
      }
    }
  } catch (err) {
    console.error('Failed to fetch GeoIP data:', err)
  }
}

// Polling function
const poll = async () => {
  try {
    await networkStore.fetchMetrics()
    measurePing()
    await fetchGeoData()
  } catch (err) {
    // Silently continue polling
  }
}

onMounted(() => {
  poll()
  pollTimer = setInterval(poll, 2000)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
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
