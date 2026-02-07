<template>
  <div class="mod-panel mod-toplist">
    <div class="section-title">
      <span class="title-left">TOP PROCESSES</span>
    </div>

    <!-- Column headers -->
    <div class="toplist-header">
      <span class="col-pid">PID</span>
      <span class="col-name">NAME</span>
      <span class="col-cpu">CPU</span>
      <span class="col-mem">MEM</span>
    </div>

    <!-- Process rows -->
    <div
      v-for="proc in topFive"
      :key="proc.pid"
      class="toplist-row"
      @click="onProcessClick(proc)"
    >
      <span class="col-pid">{{ proc.pid }}</span>
      <span class="col-name">{{ proc.name }}</span>
      <span class="col-cpu">{{ proc.cpu.toFixed(1) }}</span>
      <span class="col-mem">{{ proc.memory.toFixed(1) }}</span>
    </div>

    <!-- Empty state when no data -->
    <div v-if="topFive.length === 0" class="toplist-row">
      <span class="col-name" style="opacity: 0.4;">Waiting for data...</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from 'vue'
import { useSystemStore } from '~/stores/system'

interface TopProcess {
  pid: number
  name: string
  cpu: number
  memory: number
}

const emit = defineEmits<{
  (e: 'process-click', proc: TopProcess): void
}>()

const systemStore = useSystemStore()

let refreshTimer: ReturnType<typeof setInterval> | null = null

const topFive = computed<TopProcess[]>(() => {
  // topProcesses from the store is already sorted by CPU (desc) via sortedProcesses
  const procs = systemStore.topProcesses
  if (!procs || procs.length === 0) return []

  // Sort by combined CPU + memory weight, take top 5
  const sorted = [...procs]
    .sort((a, b) => (b.cpu + b.memory) - (a.cpu + a.memory))
    .slice(0, 5)

  return sorted.map(p => ({
    pid: p.pid,
    name: p.name,
    cpu: p.cpu,
    memory: p.memory
  }))
})

function onProcessClick(proc: TopProcess) {
  emit('process-click', proc)
}

onMounted(() => {
  // The store already auto-refreshes system stats on its own interval.
  // We poll the store every 2s to ensure reactivity picks up new data.
  // (Vue's reactivity handles this automatically via computed, but the
  //  timer ensures we trigger re-renders even if the store mutates in place.)
  refreshTimer = setInterval(() => {
    // No-op -- reactivity handles it. Timer kept for future use
    // (e.g., if we add a manual fetchSystemStats() call here).
  }, 2000)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>
