<template>
  <div class="mod-panel mod-sysinfo">
    <div class="section-title">
      <span class="title-left">{{ t('panels.sysinfo.title') }}</span>
    </div>

    <div class="sysinfo-row">
      <span class="sysinfo-label">{{ t('panels.sysinfo.year') }}</span>
      <span class="sysinfo-value">{{ currentYear }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">{{ t('panels.sysinfo.date') }}</span>
      <span class="sysinfo-value">{{ currentDate }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">{{ t('panels.sysinfo.uptime') }}</span>
      <span class="sysinfo-value">{{ uptimeDisplay }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">OS</span>
      <span class="sysinfo-value sysinfo-value-truncated" :title="osDisplay">{{ osDisplay }}</span>
    </div>
    <div v-if="kernelDisplay" class="sysinfo-row">
      <span class="sysinfo-label">{{ t('panels.sysinfo.kernel') }}</span>
      <span class="sysinfo-value sysinfo-value-truncated" :title="kernelDisplay">{{ kernelDisplay }}</span>
    </div>
    <div class="sysinfo-row">
      <span class="sysinfo-label">{{ t('panels.sysinfo.power') }}</span>
      <span class="sysinfo-value" :title="powerDisplay">{{ powerDisplay }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { useNow, useDateFormat } from '@vueuse/core'
import { intervalToDuration } from 'date-fns'
import { useSystemStore } from '~/stores/system'
import { useI18n } from '~/composables/useI18n'
import { GetPowerInfo } from '~/lib/wailsjs/coordinator'

const { t } = useI18n()

const systemStore = useSystemStore()

// Live power state — polled every 5 s from the cross-platform
// distatus/battery binding. Replaces the previous hardcoded "AC Power"
// string so unplugging the laptop actually changes the display.
const power = ref<{ source: string; percent: number; onBattery: boolean; status?: string } | null>(null)

let powerTimer: ReturnType<typeof setInterval> | null = null

// `useNow({ interval: 1000 })` from VueUse gives us a reactive Date
// that updates every second — perfect for live YEAR / DATE / UPTIME
// displays without managing setInterval timers by hand.
const now = useNow({ interval: 1000 })

// YEAR / DATE formatted via VueUse's useDateFormat (which delegates
// to date-fns under the hood for tokenized formatting).
const currentYear = useDateFormat(now, 'YYYY')
const currentDate = useDateFormat(now, 'MM/DD')

// Pin the boot timestamp the moment we get the system uptime from
// the backend. Then live-difference against `now` so the display
// ticks every second WITHOUT polling the backend.
const bootEpochMs = computed(() => {
  const u = systemStore.systemInfo?.uptime ?? 0
  if (u <= 0) return null
  return Date.now() - u * 1000
})

const uptimeDisplay = computed(() => {
  const boot = bootEpochMs.value
  if (!boot) return '--'
  // intervalToDuration({ start, end }) returns
  // { years, months, days, hours, minutes, seconds }. For uptime we
  // collapse years/months into days for a "Nd HH:MM:SS" feel that
  // matches the eDex aesthetic but with explicit unit labels so the
  // user knows which field is which (the earlier 02:22:13 was
  // ambiguous between H:M:S and D:H:M).
  const d = intervalToDuration({ start: boot, end: now.value })
  const totalDays = (d.years ?? 0) * 365 + (d.months ?? 0) * 30 + (d.days ?? 0)
  const hh = String(d.hours ?? 0).padStart(2, '0')
  const mm = String(d.minutes ?? 0).padStart(2, '0')
  const ss = String(d.seconds ?? 0).padStart(2, '0')
  return totalDays > 0
    ? t('panels.sysinfo.uptimeDays', { days: totalDays, hh, mm, ss })
    : t('panels.sysinfo.uptimeTime', { hh, mm, ss })
})

const osDisplay = computed(() => {
  const info = systemStore.systemInfo
  if (!info) return t('panels.unknown')
  const os = (info as any).os || (info as any).platform || info.platform || t('panels.unknown')
  const arch = (info as any).arch || (info as any).architecture || info.architecture || ''
  return arch ? `${os} ${arch}` : os
})

// Kernel as its own row so it doesn't overflow alongside OS+arch.
const kernelDisplay = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return ''
  return info.kernel || info.kernelVersion || ''
})

// "AC" / "BATTERY 87%" / "BATTERY (Charging) 64%" — derived from the
// live power poll. Falls back to "AC Power" so desktop builds without
// a battery still read sensibly.
const powerDisplay = computed(() => {
  const p = power.value
  if (!p) return t('panels.sysinfo.acPower')
  if (p.source === 'Battery') {
    return p.percent >= 0
      ? t('panels.sysinfo.batteryPercent', { percent: p.percent })
      : t('panels.sysinfo.battery')
  }
  // On AC. If we have a percent + charging status, surface it; otherwise
  // just "AC Power".
  if (p.percent >= 0 && p.status === 'Charging') {
    return t('panels.sysinfo.acCharging', { percent: p.percent })
  }
  return t('panels.sysinfo.acPower')
})

async function refreshPower() {
  try {
    power.value = await GetPowerInfo()
  } catch {
    // Non-fatal — keep showing last good value.
  }
}

onMounted(() => {
  // YEAR / DATE / UPTIME all derive from `useNow` which ticks every
  // second — no manual setInterval / scheduleMidnightRefresh needed.
  // Power state still polls separately because GetPowerInfo is an
  // IPC round-trip that doesn't need to fire every second.
  refreshPower()
  powerTimer = setInterval(refreshPower, 5000)
})

onBeforeUnmount(() => {
  if (powerTimer) {
    clearInterval(powerTimer)
    powerTimer = null
  }
})
</script>
