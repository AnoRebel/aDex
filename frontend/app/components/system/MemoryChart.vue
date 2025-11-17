<template>
  <div class="memory-chart">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-green-500 rounded-full animate-pulse" />
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          Memory Usage
        </h3>
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ formatBytes(memoryMetrics?.total || 0) }} Total
        </span>
      </div>

      <div class="flex items-center gap-3">
        <!-- Current usage indicator -->
        <div class="text-right">
          <div class="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {{ formatPercentage(memoryUsage) }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            {{ formatBytes(memoryMetrics?.used || 0) }} used
          </div>
        </div>

        <!-- Controls -->
        <div class="flex items-center gap-2">
          <button
            @click="toggleSwap"
            :class="[
              'px-2 py-1 text-xs rounded transition-colors',
              showSwap
                ? 'bg-green-500 text-white'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
            ]"
          >
            Swap
          </button>
          <button
            @click="toggleDetails"
            :class="[
              'px-2 py-1 text-xs rounded transition-colors',
              showDetails
                ? 'bg-green-500 text-white'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
            ]"
          >
            Details
          </button>
        </div>
      </div>
    </div>

    <!-- Main memory visualization -->
    <div class="mb-6">
      <div class="relative">
        <!-- Memory bar chart -->
        <div class="space-y-3">
          <!-- Used memory -->
          <div>
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-700 dark:text-gray-300">Used</span>
              <span class="text-gray-900 dark:text-gray-100 font-medium">
                {{ formatBytes(memoryMetrics?.used || 0) }} ({{ formatPercentage(memoryUsage) }})
              </span>
            </div>
            <div class="relative h-6 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="absolute top-0 left-0 h-full bg-gradient-to-r from-green-400 to-green-600 transition-all duration-300"
                :style="{ width: `${memoryUsage}%` }"
              />
              <div class="absolute inset-0 flex items-center justify-center">
                <span class="text-xs font-medium text-white mix-blend-difference">
                  {{ formatPercentage(memoryUsage) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Available memory -->
          <div>
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-700 dark:text-gray-300">Available</span>
              <span class="text-gray-900 dark:text-gray-100 font-medium">
                {{ formatBytes(memoryMetrics?.available || 0) }}
              </span>
            </div>
            <div class="relative h-4 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div
                class="absolute top-0 left-0 h-full bg-gradient-to-r from-blue-400 to-blue-500 transition-all duration-300"
                :style="{ width: `${availablePercentage}%` }"
              />
            </div>
          </div>

          <!-- Cache (if available) -->
          <div v-if="memoryMetrics?.cached">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-700 dark:text-gray-300">Cache</span>
              <span class="text-gray-900 dark:text-gray-100 font-medium">
                {{ formatBytes(memoryMetrics.cached) }}
              </span>
            </div>
            <div class="relative h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div
                class="absolute top-0 left-0 h-full bg-gradient-to-r from-yellow-400 to-yellow-500 transition-all duration-300"
                :style="{ width: `${cachePercentage}%` }"
              />
            </div>
          </div>

          <!-- Buffers (if available) -->
          <div v-if="memoryMetrics?.buffers">
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-700 dark:text-gray-300">Buffers</span>
              <span class="text-gray-900 dark:text-gray-100 font-medium">
                {{ formatBytes(memoryMetrics.buffers) }}
              </span>
            </div>
            <div class="relative h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div
                class="absolute top-0 left-0 h-full bg-gradient-to-r from-purple-400 to-purple-500 transition-all duration-300"
                :style="{ width: `${buffersPercentage}%` }"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Swap memory -->
    <div v-if="showSwap && memoryMetrics?.swapTotal > 0" class="mb-6">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Swap Memory
      </h4>
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
        <div class="space-y-3">
          <!-- Swap usage -->
          <div>
            <div class="flex justify-between text-sm mb-1">
              <span class="text-gray-700 dark:text-gray-300">Used</span>
              <span class="text-gray-900 dark:text-gray-100 font-medium">
                {{ formatBytes(memoryMetrics.swapUsed) }} ({{ formatPercentage(memoryMetrics.swapPercent) }})
              </span>
            </div>
            <div class="relative h-4 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
              <div
                class="absolute top-0 left-0 h-full transition-all duration-300"
                :class="swapColorClass"
                :style="{ width: `${memoryMetrics.swapPercent}%` }"
              />
            </div>
          </div>

          <!-- Swap total indicator -->
          <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400">
            <span>Total: {{ formatBytes(memoryMetrics.swapTotal) }}</span>
            <span>Free: {{ formatBytes(memoryMetrics.swapFree) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Memory history using Chart.js -->
    <div v-if="memoryHistory.length > 0" class="mb-6">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Usage History ({{ historyTimeframe }}s)
      </h4>
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
        <canvas ref="historyCanvas" width="400" height="100" />
      </div>
    </div>

    <!-- Detailed breakdown -->
    <div v-if="showDetails && memoryMetrics" class="mb-6">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Memory Breakdown
      </h4>
      <div class="grid grid-cols-2 gap-3">
        <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Total</div>
          <div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {{ formatBytes(memoryMetrics.total) }}
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Used</div>
          <div class="text-lg font-semibold" :class="memoryUsageColor">
            {{ formatBytes(memoryMetrics.used) }}
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Free</div>
          <div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {{ formatBytes(memoryMetrics.free) }}
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Available</div>
          <div class="text-lg font-semibold text-blue-600 dark:text-blue-400">
            {{ formatBytes(memoryMetrics.available) }}
          </div>
        </div>
        <div v-if="memoryMetrics.cached" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Cached</div>
          <div class="text-lg font-semibold text-yellow-600 dark:text-yellow-400">
            {{ formatBytes(memoryMetrics.cached) }}
          </div>
        </div>
        <div v-if="memoryMetrics.buffers" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Buffers</div>
          <div class="text-lg font-semibold text-purple-600 dark:text-purple-400">
            {{ formatBytes(memoryMetrics.buffers) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Memory pressure indicator -->
    <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div
            class="w-2 h-2 rounded-full"
            :class="memoryPressureColor"
          />
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            Memory Pressure
          </span>
        </div>
        <span class="text-sm font-semibold" :class="memoryPressureTextColor">
          {{ memoryPressureStatus }}
        </span>
      </div>
      <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
        {{ memoryPressureMessage }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useSystemStore } from '~/stores/system'
import { useStorage } from '@vueuse/core'
import { Chart, type ChartConfiguration, type LineControllerChartOptions } from 'chart.js'
import type { MemoryMetrics } from '~/types/system'

interface Props {
  data?: {
    used: Array<{ value: number; timestamp: Date }>
    free: Array<{ value: number; timestamp: Date }>
    cached?: Array<{ value: number; timestamp: Date }>
    buffers?: Array<{ value: number; timestamp: Date }>
    swap?: Array<{ value: number; timestamp: Date }>
  }
  showSwap?: boolean
  showBuffers?: boolean
  showCached?: boolean
  height?: number
  width?: number
  theme?: 'light' | 'dark'
}

const props = withDefaults(defineProps<Props>(), {
  showSwap: true,
  showBuffers: true,
  showCached: true,
  height: 300,
  width: 400,
  theme: 'light'
})

// Store
const systemStore = useSystemStore()

// Local state
const showSwap = useStorage('memory-chart-show-swap', props.showSwap)
const showDetails = useStorage('memory-chart-show-details', true)

// Chart references
const historyCanvas = ref<HTMLCanvasElement>()

// Chart instances
let historyChart: Chart | null = null

// Computed properties
const memoryMetrics = computed(() => systemStore.memoryMetrics)
const memoryUsage = computed(() => systemStore.memoryUsage)
const memoryHistory = computed(() => systemStore.memoryHistory)
const historyTimestamps = computed(() => systemStore.historyTimestamps)

// Memory calculations
const availablePercentage = computed(() => {
  if (!memoryMetrics.value?.total) return 0
  return (memoryMetrics.value.available / memoryMetrics.value.total) * 100
})

const cachePercentage = computed(() => {
  if (!memoryMetrics.value?.cached || !memoryMetrics.value?.total) return 0
  return (memoryMetrics.value.cached / memoryMetrics.value.total) * 100
})

const buffersPercentage = computed(() => {
  if (!memoryMetrics.value?.buffers || !memoryMetrics.value?.total) return 0
  return (memoryMetrics.value.buffers / memoryMetrics.value.total) * 100
})

// Memory pressure
const memoryPressureStatus = computed(() => {
  if (memoryUsage.value >= 90) return 'Critical'
  if (memoryUsage.value >= 75) return 'High'
  if (memoryUsage.value >= 50) return 'Medium'
  return 'Low'
})

const memoryPressureColor = computed(() => {
  if (memoryUsage.value >= 90) return 'bg-red-500'
  if (memoryUsage.value >= 75) return 'bg-yellow-500'
  if (memoryUsage.value >= 50) return 'bg-blue-500'
  return 'bg-green-500'
})

const memoryPressureTextColor = computed(() => {
  if (memoryUsage.value >= 90) return 'text-red-500'
  if (memoryUsage.value >= 75) return 'text-yellow-500'
  if (memoryUsage.value >= 50) return 'text-blue-500'
  return 'text-green-500'
})

const memoryPressureMessage = computed(() => {
  if (memoryUsage.value >= 90) return 'System is running out of memory. Consider closing applications.'
  if (memoryUsage.value >= 75) return 'High memory usage detected. Monitor system performance.'
  if (memoryUsage.value >= 50) return 'Moderate memory usage. System is running normally.'
  return 'Low memory usage. Plenty of memory available.'
})

// Swap visualization
const swapColorClass = computed(() => {
  if (!memoryMetrics.value) return 'bg-green-400'
  const swapPercent = memoryMetrics.value.swapPercent
  if (swapPercent >= 50) return 'bg-gradient-to-r from-red-400 to-red-600'
  if (swapPercent >= 25) return 'bg-gradient-to-r from-yellow-400 to-yellow-600'
  return 'bg-gradient-to-r from-green-400 to-green-600'
})

const memoryUsageColor = computed(() => {
  if (memoryUsage.value >= 90) return 'text-red-500'
  if (memoryUsage.value >= 75) return 'text-yellow-500'
  return 'text-green-500'
})

// History calculations
const historyTimeframe = computed(() => {
  if (historyTimestamps.value.length < 2) return 0
  const latest = historyTimestamps.value[historyTimestamps.value.length - 1].getTime()
  const earliest = historyTimestamps.value[0].getTime()
  return Math.floor((latest - earliest) / 1000)
})

// Chart initialization
const initHistoryChart = () => {
  if (!historyCanvas.value || memoryHistory.value.length === 0) return

  const ctx = historyCanvas.value.getContext('2d')
  if (!ctx) return

  const labels = memoryHistory.value.map((_, index) => {
    if (index % Math.ceil(memoryHistory.value.length / 8) === 0) {
      return `${historyTimeframe.value - Math.floor((index / memoryHistory.value.length) * historyTimeframe.value)}s`
    }
    return ''
  })

  const config: ChartConfiguration = {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label: 'Memory Usage',
        data: memoryHistory.value,
        borderColor: '#22c55e',
        backgroundColor: 'rgba(34, 197, 94, 0.1)',
        borderWidth: 2,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 4
      }]
    },
    options: {
      responsive: false,
      maintainAspectRatio: false,
      scales: {
        x: {
          display: true,
          grid: { display: false },
          ticks: {
            color: props.theme === 'dark' ? '#9ca3af' : '#6b7280',
            font: { size: 10 }
          }
        },
        y: {
          display: true,
          min: 0,
          max: 100,
          grid: {
            color: props.theme === 'dark' ? '#374151' : '#e5e7eb'
          },
          ticks: {
            color: props.theme === 'dark' ? '#9ca3af' : '#6b7280',
            font: { size: 10 },
            callback: (value) => `${value}%`
          }
        }
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: (context) => `Memory: ${Math.round(context.parsed.y)}%`
          }
        }
      }
    } as LineControllerChartOptions
  }

  if (historyChart) {
    historyChart.destroy()
  }
  historyChart = new Chart(ctx, config)
}

// Methods
const formatBytes = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`
}

const formatPercentage = (value: number): string => {
  return `${Math.round(value)}%`
}

const toggleSwap = () => {
  showSwap.value = !showSwap.value
}

const toggleDetails = () => {
  showDetails.value = !showDetails.value
}

// Update chart when data changes
watch(memoryHistory, () => {
  nextTick(() => {
    initHistoryChart()
  })
}, { deep: true })

// Initialize chart on mount
onMounted(() => {
  nextTick(() => {
    initHistoryChart()
  })
})

// Cleanup on unmount
onUnmounted(() => {
  if (historyChart) historyChart.destroy()
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.memory-chart {
  @apply bg-white dark:bg-gray-900 rounded-lg p-4 shadow-sm border border-gray-200 dark:border-gray-700;
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Custom transitions for memory bars */
.transition-all {
  transition-property: all;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 300ms;
}

/* Mix blend mode for better text visibility on colored backgrounds */
.mix-blend-difference {
  mix-blend-mode: difference;
}
</style>