<template>
  <div class="mod-panel mod-hardware">
    <div class="section-title">
      <span class="title-left">{{ t('panels.hardware.title') }}</span>
    </div>

    <div class="hw-row">
      <span class="hw-label">{{ t('panels.hardware.manufacturer') }}</span>
      <span class="hw-value">{{ manufacturer }}</span>
    </div>
    <div class="hw-row">
      <span class="hw-label">{{ t('panels.hardware.model') }}</span>
      <span class="hw-value">{{ model }}</span>
    </div>
    <div class="hw-row">
      <span class="hw-label">{{ t('panels.hardware.chassis') }}</span>
      <span class="hw-value">{{ chassis }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSystemStore } from '~/stores/system'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()

const systemStore = useSystemStore()

const manufacturer = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return t('panels.unknown')
  // The Go backend may provide manufacturer info in extended fields.
  // Fall back through possible field names from gopsutil host.Info().
  return info.manufacturer
    || info.Manufacturer
    || info.vendorId
    || info.VendorID
    || t('panels.unknown')
})

const model = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return t('panels.unknown')
  return info.model
    || info.Model
    || info.productName
    || info.ProductName
    || info.hostname
    || info.Hostname
    || t('panels.unknown')
})

const chassis = computed(() => {
  const info = systemStore.systemInfo as any
  if (!info) return t('panels.unknown')
  // gopsutil may expose virtualizationSystem or similar fields
  if (info.chassisType || info.ChassisType) {
    return info.chassisType || info.ChassisType
  }
  if (info.virtualizationSystem || info.VirtualizationSystem) {
    return t('panels.hardware.virtual', { system: info.virtualizationSystem || info.VirtualizationSystem })
  }
  // Derive a sensible default from platform
  const platform = info.platform || info.Platform || info.os || ''
  if (platform.toLowerCase().includes('darwin')) return t('panels.hardware.laptop')
  if (platform.toLowerCase().includes('linux')) return t('panels.hardware.desktop')
  if (platform.toLowerCase().includes('windows')) return t('panels.hardware.desktop')
  return t('panels.unknown')
})
</script>
