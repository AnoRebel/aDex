<template>
  <div class="mod-panel mod-hardware">
    <div class="section-title">
      <span class="title-left">HARDWARE INSPECTOR</span>
    </div>

    <div class="hw-row">
      <span class="hw-label">MANUFACTURER</span>
      <span class="hw-value">{{ manufacturer }}</span>
    </div>
    <div class="hw-row">
      <span class="hw-label">MODEL</span>
      <span class="hw-value">{{ model }}</span>
    </div>
    <div class="hw-row">
      <span class="hw-label">CHASSIS</span>
      <span class="hw-value">{{ chassis }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSystemStore } from '~/stores/system'

const systemStore = useSystemStore()

const manufacturer = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return 'Unknown'
  // The Go backend may provide manufacturer info in extended fields.
  // Fall back through possible field names from gopsutil host.Info().
  return info.manufacturer
    || info.Manufacturer
    || info.vendorId
    || info.VendorID
    || 'Unknown'
})

const model = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return 'Unknown'
  return info.model
    || info.Model
    || info.productName
    || info.ProductName
    || info.hostname
    || info.Hostname
    || 'Unknown'
})

const chassis = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return 'Unknown'
  // gopsutil may expose virtualizationSystem or similar fields
  if (info.chassisType || info.ChassisType) {
    return info.chassisType || info.ChassisType
  }
  if (info.virtualizationSystem || info.VirtualizationSystem) {
    return `Virtual (${info.virtualizationSystem || info.VirtualizationSystem})`
  }
  // Derive a sensible default from platform
  const platform = info.platform || info.Platform || info.os || ''
  if (platform.toLowerCase().includes('darwin')) return 'Laptop'
  if (platform.toLowerCase().includes('linux')) return 'Desktop'
  if (platform.toLowerCase().includes('windows')) return 'Desktop'
  return 'Unknown'
})
</script>
