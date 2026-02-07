<template>
  <div class="mod-panel mod-sysinfo">
    <div class="section-title">
      <span class="title-left">SYSTEM INFO</span>
    </div>

    <div class="sysinfo-row">
      <span class="sysinfo-label">YEAR</span>
      <span class="sysinfo-value">{{ currentYear }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">DATE</span>
      <span class="sysinfo-value">{{ currentDate }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">UPTIME</span>
      <span class="sysinfo-value">{{ uptimeDisplay }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">OS</span>
      <span class="sysinfo-value">{{ osDisplay }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">POWER</span>
      <span class="sysinfo-value">{{ powerDisplay }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useSystemStore } from '~/stores/system'

const systemStore = useSystemStore()

const currentYear = ref('')
const currentDate = ref('')

let dateTimer: ReturnType<typeof setInterval> | null = null
let uptimeTimer: ReturnType<typeof setInterval> | null = null

// Local uptime offset so the display ticks between store refreshes
const localUptimeOffset = ref(0)

const uptimeDisplay = computed(() => {
  const baseUptime = systemStore.systemInfo?.uptime ?? 0
  const totalSeconds = baseUptime + localUptimeOffset.value
  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  return `${String(days).padStart(2, '0')}:${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
})

const osDisplay = computed(() => {
  const info = systemStore.systemInfo
  if (!info) return 'Unknown'
  // Use whichever fields are available from the store
  const os = (info as any).os || (info as any).platform || info.platform || 'Unknown'
  const arch = (info as any).arch || (info as any).architecture || info.architecture || ''
  const kernel = (info as any).kernel || (info as any).kernelVersion || info.kernelVersion || ''
  const parts = [os]
  if (arch) parts.push(arch)
  if (kernel) parts.push(kernel)
  return parts.join(' ')
})

const powerDisplay = computed(() => {
  // Battery / power information is not available from the current backend.
  // Show "AC" as default. A future backend extension can provide battery data.
  return 'AC Power'
})

function updateDate() {
  const now = new Date()
  currentYear.value = String(now.getFullYear())
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  currentDate.value = `${month}/${day}`
}

function msUntilMidnight(): number {
  const now = new Date()
  const midnight = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1)
  return midnight.getTime() - now.getTime()
}

function scheduleMidnightRefresh() {
  // Clear any existing timer
  if (dateTimer) {
    clearTimeout(dateTimer as any)
    dateTimer = null
  }
  const ms = msUntilMidnight()
  dateTimer = setTimeout(() => {
    updateDate()
    // After midnight, schedule the next midnight refresh
    scheduleMidnightRefresh()
  }, ms) as any
}

onMounted(() => {
  updateDate()
  scheduleMidnightRefresh()

  // Increment the local uptime offset every 60 seconds
  uptimeTimer = setInterval(() => {
    localUptimeOffset.value += 60
  }, 60000)
})

onBeforeUnmount(() => {
  if (dateTimer) {
    clearTimeout(dateTimer as any)
    dateTimer = null
  }
  if (uptimeTimer) {
    clearInterval(uptimeTimer)
    uptimeTimer = null
  }
})
</script>
