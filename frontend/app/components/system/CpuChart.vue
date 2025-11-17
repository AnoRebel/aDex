<template>
  <div class="cpu-chart">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-blue-500 rounded-full animate-pulse" />
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          CPU Usage
        </h3>
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ cpuMetrics?.model || 'Unknown' }}
        </span>
      </div>

      <div class="flex items-center gap-3">
        <!-- Current usage indicator -->
        <div class="text-right">
          <div class="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {{ formatPercentage(cpuUsage) }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            {{ cpuMetrics?.cores || 0 }} cores
          </div>
        </div>

        <!-- Controls -->
        <div class="flex items-center gap-2">
          <button
            @click="togglePerCore"
            :class="[
              'px-2 py-1 text-xs rounded transition-colors',
              showPerCore
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
            ]"
          >
            Per Core
          </button>
          <button
            @click="toggleHistory"
            :class="[
              'px-2 py-1 text-xs rounded transition-colors',
              showHistory
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
            ]"
          >
            History
          </button>
        </div>
      </div>
    </div>

    <!-- Main gauge using Chart.js -->
    <div class="relative mb-6">
      <div class="flex items-center justify-center">
        <canvas ref="gaugeCanvas" :width="gaugeSize" :height="gaugeSize" />
      </div>

      <!-- Frequency indicator -->
      <div v-if="cpuMetrics" class="absolute bottom-0 left-0 right-0 text-center">
        <div class="text-xs text-gray-500 dark:text-gray-400">
          {{ formatFrequency(cpuMetrics.frequency) }} / {{ formatFrequency(cpuMetrics.frequencyMax) }}
        </div>
      </div>
    </div>

    <!-- Per-core usage -->
    <div v-if="showPerCore && cpuMetrics?.perCoreUsage" class="mb-6">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Per Core Usage
      </h4>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
        <div
          v-for="(usage, index) in cpuMetrics.perCoreUsage"
          :key="index"
          class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2"
        >
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">
            Core {{ index }}
          </div>
          <div class="relative h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
            <div
              class="absolute top-0 left-0 h-full bg-gradient-to-r from-blue-400 to-blue-600 transition-all duration-300"
              :style="{ width: `${usage}%` }"
            />
          </div>
          <div class="text-xs text-gray-700 dark:text-gray-300 mt-1">
            {{ formatPercentage(usage) }}
          </div>
        </div>
      </div>
    </div>

    <!-- History chart using Chart.js -->
    <div v-if="showHistory && cpuHistory.length > 0" class="mb-6">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Usage History ({{ historyTimeframe }}s)
      </h4>
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
        <canvas ref="historyCanvas" width="400" height="100" />
      </div>
    </div>

    <!-- Load average (if available) -->
    <div v-if="cpuMetrics?.loadAverage && cpuMetrics.loadAverage.length > 0" class="mb-6">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Load Average
      </h4>
      <div class="grid grid-cols-3 gap-4">
        <div class="text-center">
          <div class="text-xs text-gray-500 dark:text-gray-400">1 min</div>
          <div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {{ cpuMetrics.loadAverage[0]?.toFixed(2) || 'N/A' }}
          </div>
        </div>
        <div class="text-center">
          <div class="text-xs text-gray-500 dark:text-gray-400">5 min</div>
          <div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {{ cpuMetrics.loadAverage[1]?.toFixed(2) || 'N/A' }}
          </div>
        </div>
        <div class="text-center">
          <div class="text-xs text-gray-500 dark:text-gray-400">15 min</div>
          <div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {{ cpuMetrics.loadAverage[2]?.toFixed(2) || 'N/A' }}
          </div>
        </div>
      </div>
    </div>

    <!-- Quick stats -->
    <div class="grid grid-cols-2 gap-4">
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
        <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Current Usage</div>
        <div class="text-lg font-semibold" :class="currentUsageColor">
          {{ formatPercentage(cpuUsage) }}
        </div>
      </div>
      <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
        <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Average</div>
        <div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {{ formatPercentage(averageUsage) }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useSystemStore } from '~/stores/system'
import { useStorage } from '@vueuse/core'
import { Chart, type ChartConfiguration, type DoughnutControllerChartOptions, type LineControllerChartOptions } from 'chart.js'
import type { CPUMetrics } from '~/types/system'

interface Props {
  data?: {
    overall: Array<{ value: number; timestamp: Date }>
    perCore?: Record<number, Array<{ value: number; timestamp: Date }>>
  }
  showPerCore?: boolean
  showLoadAverage?: boolean
  height?: number
  width?: number
  theme?: 'light' | 'dark'
}

const props = withDefaults(defineProps<Props>(), {
  showPerCore: true,
  showLoadAverage: true,
  height: 300,
  width: 400,
  theme: 'light'
})

// Store
const systemStore = useSystemStore()

// Local state
const showPerCore = useStorage('cpu-chart-show-per-core', props.showPerCore)
const showHistory = useStorage('cpu-chart-show-history', true)

// Chart references
const gaugeCanvas = ref<HTMLCanvasElement>()
const historyCanvas = ref<HTMLCanvasElement>()

// Chart instances
let gaugeChart: Chart | null = null
let historyChart: Chart | null = null

// Chart dimensions
const gaugeSize = computed(() => Math.min(props.width, 120))

// Computed properties
const cpuMetrics = computed(() => systemStore.cpuMetrics)
const cpuUsage = computed(() => systemStore.cpuUsage)
const cpuHistory = computed(() => systemStore.cpuHistory)
const historyTimestamps = computed(() => systemStore.historyTimestamps)

// History calculations
const historyTimeframe = computed(() => {
  if (historyTimestamps.value.length < 2) return 0
  const latest = historyTimestamps.value[historyTimestamps.value.length - 1].getTime()
  const earliest = historyTimestamps.value[0].getTime()
  return Math.floor((latest - earliest) / 1000)
})

const averageUsage = computed(() => {
  if (cpuHistory.value.length === 0) return 0
  const sum = cpuHistory.value.reduce((acc, val) => acc + val, 0)
  return sum / cpuHistory.value.length
})

const currentUsageColor = computed(() => {
  if (cpuUsage.value >= 90) return 'text-red-500'
  if (cpuUsage.value >= 75) return 'text-amber-500'
  return 'text-green-500'
})

// Chart initialization
const initGaugeChart = () => {
  if (!gaugeCanvas.value) return

  const ctx = gaugeCanvas.value.getContext('2d')
  if (!ctx) return

  const percentage = cpuUsage.value
  const color = percentage >= 90 ? '#ef4444' : percentage >= 75 ? '#f59e0b' : '#3b82f6'

  const config: ChartConfiguration = {
    type: 'doughnut',
    data: {
      datasets: [{
        data: [percentage, 100 - percentage],
        backgroundColor: [color, '#e5e7eb'],
        borderWidth: 0,
        circumference: 270,
        rotation: 225
      }]
    },
    options: {
      responsive: false,
      maintainAspectRatio: false,
      cutout: '70%',
      plugins: {
        legend: { display: false },
        tooltip: { enabled: false },
        title: {
          display: true,
          text: `${Math.round(percentage)}%`,
          position: 'center',
          font: { size: 20, weight: 'bold' },
          color: props.theme === 'dark' ? '#f9fafb' : '#111827'
        }
      }
    } as DoughnutControllerChartOptions
  }

  if (gaugeChart) {
    gaugeChart.destroy()
  }
  gaugeChart = new Chart(ctx, config)
}

const initHistoryChart = () => {
  if (!historyCanvas.value || cpuHistory.value.length === 0) return

  const ctx = historyCanvas.value.getContext('2d')
  if (!ctx) return

  const labels = cpuHistory.value.map((_, index) => {
    if (index % Math.ceil(cpuHistory.value.length / 8) === 0) {
      return `${historyTimeframe.value - Math.floor((index / cpuHistory.value.length) * historyTimeframe.value)}s`
    }
    return ''
  })

  const config: ChartConfiguration = {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label: 'CPU Usage',
        data: cpuHistory.value,
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
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
            label: (context) => `CPU: ${Math.round(context.parsed.y)}%`
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
const formatPercentage = (value: number): string => {
  return `${Math.round(value)}%`
}

const formatFrequency = (mhz: number): string => {
  if (mhz < 1000) {
    return `${mhz.toFixed(0)} MHz`
  }
  return `${(mhz / 1000).toFixed(2)} GHz`
}

const togglePerCore = () => {
  showPerCore.value = !showPerCore.value
}

const toggleHistory = () => {
  showHistory.value = !showHistory.value
  if (showHistory.value) {
    nextTick(() => initHistoryChart())
  }
}

// Update charts when data changes
watch([cpuUsage, cpuHistory], () => {
  nextTick(() => {
    initGaugeChart()
    if (showHistory.value) {
      initHistoryChart()
    }
  })
}, { deep: true })

// Initialize charts on mount
onMounted(() => {
  nextTick(() => {
    initGaugeChart()
    if (showHistory.value) {
      initHistoryChart()
    }
  })
})

// Cleanup on unmount
onUnmounted(() => {
  if (gaugeChart) gaugeChart.destroy()
  if (historyChart) historyChart.destroy()
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.cpu-chart {
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

/* Custom transitions for gauge */
.transition-all {
  transition-property: all;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 300ms;
}
</style>