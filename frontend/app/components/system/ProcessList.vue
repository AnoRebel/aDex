<template>
  <div class="process-list">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <div class="w-3 h-3 bg-purple-500 rounded-full animate-pulse" />
        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
          Processes
        </h3>
        <span class="text-sm text-gray-500 dark:text-gray-400">
          {{ filteredProcesses.length }} of {{ totalProcesses }}
        </span>
      </div>

      <div class="flex items-center gap-3">
        <!-- Process counts -->
        <div class="flex gap-4 text-sm">
          <div class="text-center">
            <div class="text-xs text-gray-500 dark:text-gray-400">Running</div>
            <div class="font-semibold text-green-600 dark:text-green-400">
              {{ runningProcesses }}
            </div>
          </div>
          <div class="text-center">
            <div class="text-xs text-gray-500 dark:text-gray-400">Sleeping</div>
            <div class="font-semibold text-blue-600 dark:text-blue-400">
              {{ sleepingProcesses }}
            </div>
          </div>
        </div>

        <!-- Controls -->
        <button
          @click="toggleAutoRefresh"
          :class="[
            'px-2 py-1 text-xs rounded transition-colors',
            autoRefresh
              ? 'bg-purple-500 text-white'
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
          ]"
        >
          Auto
        </button>
      </div>
    </div>

    <!-- Search and filter -->
    <div class="mb-4 space-y-3">
      <!-- Search -->
      <div class="relative">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search processes..."
          class="w-full px-3 py-2 pl-9 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-purple-500"
        />
        <div class="absolute left-3 top-2.5 text-gray-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>

      <!-- Filters -->
      <div class="flex flex-wrap gap-2">
        <select
          v-model="selectedStatus"
          class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="">All Status</option>
          <option value="running">Running</option>
          <option value="sleeping">Sleeping</option>
          <option value="stopped">Stopped</option>
          <option value="zombie">Zombie</option>
        </select>

        <select
          v-model="sortBy"
          class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="cpu">Sort by CPU</option>
          <option value="memory">Sort by Memory</option>
          <option value="pid">Sort by PID</option>
          <option value="name">Sort by Name</option>
          <option value="threads">Sort by Threads</option>
        </select>

        <button
          @click="toggleSortOrder"
          class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          {{ sortOrder === 'desc' ? '↓' : '↑' }}
        </button>

        <button
          @click="clearFilters"
          class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Process list -->
    <div class="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th
                @click="setSortBy('pid')"
                class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                PID
              </th>
              <th
                @click="setSortBy('name')"
                class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                Name
              </th>
              <th
                @click="setSortBy('cpu')"
                class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                CPU
              </th>
              <th
                @click="setSortBy('memory')"
                class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                Memory
              </th>
              <th
                @click="setSortBy('threads')"
                class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                Threads
              </th>
              <th class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300">
                Status
              </th>
              <th class="px-3 py-2 text-left font-medium text-gray-700 dark:text-gray-300">
                User
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="process in paginatedProcesses"
              :key="process.pid"
              class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
              @click="selectProcess(process)"
            >
              <td class="px-3 py-2 font-mono text-xs text-gray-600 dark:text-gray-400">
                {{ process.pid }}
              </td>
              <td class="px-3 py-2">
                <div class="flex items-center gap-2">
                  <div class="font-medium text-gray-900 dark:text-gray-100">
                    {{ process.name }}
                  </div>
                  <div
                    v-if="process.status === 'running'"
                    class="w-2 h-2 bg-green-500 rounded-full"
                  />
                  <div
                    v-else-if="process.status === 'sleeping'"
                    class="w-2 h-2 bg-blue-500 rounded-full"
                  />
                  <div
                    v-else-if="process.status === 'zombie'"
                    class="w-2 h-2 bg-red-500 rounded-full"
                  />
                  <div
                    v-else
                    class="w-2 h-2 bg-gray-500 rounded-full"
                  />
                </div>
                <div class="text-xs text-gray-500 dark:text-gray-400 truncate max-w-xs">
                  {{ process.command }}
                </div>
              </td>
              <td class="px-3 py-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-2 min-w-0">
                    <div
                      class="h-2 bg-gradient-to-r from-purple-400 to-purple-600 rounded-full"
                      :style="{ width: `${Math.min(process.cpuPercent, 100)}%` }"
                    />
                  </div>
                  <div class="text-xs font-medium text-gray-900 dark:text-gray-100 min-w-12 text-right">
                    {{ formatPercentage(process.cpuPercent) }}
                  </div>
                </div>
              </td>
              <td class="px-3 py-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 bg-gray-200 dark:bg-gray-700 rounded-full h-2 min-w-0">
                    <div
                      class="h-2 bg-gradient-to-r from-blue-400 to-blue-600 rounded-full"
                      :style="{ width: `${Math.min(process.memoryPercent, 100)}%` }"
                    />
                  </div>
                  <div class="text-xs font-medium text-gray-900 dark:text-gray-100 min-w-12 text-right">
                    {{ formatPercentage(process.memoryPercent) }}
                  </div>
                </div>
              </td>
              <td class="px-3 py-2 text-xs text-gray-600 dark:text-gray-400">
                {{ process.numThreads }}
              </td>
              <td class="px-3 py-2">
                <span
                  :class="[
                    'px-2 py-1 text-xs rounded-full font-medium',
                    getStatusClass(process.status)
                  ]"
                >
                  {{ process.status }}
                </span>
              </td>
              <td class="px-3 py-2 text-xs text-gray-600 dark:text-gray-400">
                {{ process.user }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div class="flex items-center justify-between px-3 py-2 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
        <div class="text-xs text-gray-500 dark:text-gray-400">
          Showing {{ startIndex + 1 }}-{{ Math.min(endIndex, filteredProcesses.length) }} of {{ filteredProcesses.length }}
        </div>
        <div class="flex items-center gap-2">
          <button
            @click="previousPage"
            :disabled="currentPage === 1"
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Previous
          </button>
          <div class="px-2 py-1 text-xs text-gray-600 dark:text-gray-400">
            {{ currentPage }} / {{ totalPages }}
          </div>
          <button
            @click="nextPage"
            :disabled="currentPage === totalPages"
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- Process details modal -->
    <div
      v-if="selectedProcess"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click="closeProcessDetails"
    >
      <div
        class="bg-white dark:bg-gray-900 rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[80vh] overflow-y-auto"
        @click.stop
      >
        <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Process Details
          </h3>
          <button
            @click="closeProcessDetails"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="p-4 space-y-4">
          <!-- Basic info -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">PID</div>
              <div class="font-mono text-gray-900 dark:text-gray-100">{{ selectedProcess.pid }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Name</div>
              <div class="text-gray-900 dark:text-gray-100">{{ selectedProcess.name }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Status</div>
              <div>{{ selectedProcess.status }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">User</div>
              <div class="text-gray-900 dark:text-gray-100">{{ selectedProcess.user }}</div>
            </div>
          </div>

          <!-- Resource usage -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">CPU Usage</div>
              <div class="text-lg font-semibold text-purple-600 dark:text-purple-400">
                {{ formatPercentage(selectedProcess.cpuPercent) }}
              </div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Memory Usage</div>
              <div class="text-lg font-semibold text-blue-600 dark:text-blue-400">
                {{ formatPercentage(selectedProcess.memoryPercent) }}
              </div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Memory RSS</div>
              <div class="text-gray-900 dark:text-gray-100">{{ formatBytes(selectedProcess.memoryRSS) }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Memory VMS</div>
              <div class="text-gray-900 dark:text-gray-100">{{ formatBytes(selectedProcess.memoryVMS) }}</div>
            </div>
          </div>

          <!-- Additional info -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Threads</div>
              <div class="text-gray-900 dark:text-gray-100">{{ selectedProcess.numThreads }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">File Descriptors</div>
              <div class="text-gray-900 dark:text-gray-100">{{ selectedProcess.numFDs }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Parent PID</div>
              <div class="text-gray-900 dark:text-gray-100">{{ selectedProcess.ppid }}</div>
            </div>
            <div>
              <div class="text-sm text-gray-500 dark:text-gray-400">Started</div>
              <div class="text-gray-900 dark:text-gray-100">{{ formatDate(selectedProcess.createTime) }}</div>
            </div>
          </div>

          <!-- Command -->
          <div>
            <div class="text-sm text-gray-500 dark:text-gray-400 mb-1">Command</div>
            <div class="font-mono text-xs bg-gray-100 dark:bg-gray-800 p-2 rounded text-gray-900 dark:text-gray-100">
              {{ selectedProcess.command }}
            </div>
          </div>

          <!-- Working directory -->
          <div v-if="selectedProcess.cwd">
            <div class="text-sm text-gray-500 dark:text-gray-400 mb-1">Working Directory</div>
            <div class="font-mono text-xs bg-gray-100 dark:bg-gray-800 p-2 rounded text-gray-900 dark:text-gray-100">
              {{ selectedProcess.cwd }}
            </div>
          </div>

          <!-- Executable -->
          <div v-if="selectedProcess.executable">
            <div class="text-sm text-gray-500 dark:text-gray-400 mb-1">Executable</div>
            <div class="font-mono text-xs bg-gray-100 dark:bg-gray-800 p-2 rounded text-gray-900 dark:text-gray-100">
              {{ selectedProcess.executable }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { useSystemStore } from '~/stores/system'
import { useStorage } from '@vueuse/core'
import { useIntervalFn } from '@vueuse/core'
import type { ProcessInfo, ProcessSortField, ProcessSortOrder } from '~/types/system'

interface Props {
  processes?: ProcessInfo[]
  sortBy?: ProcessSortField
  sortOrder?: ProcessSortOrder
  filter?: {
    user?: string
    status?: string
    name?: string
    minCpu?: number
    minMemory?: number
    maxCpu?: number
    maxMemory?: number
  }
  maxItems?: number
  showDetails?: boolean
  autoRefresh?: boolean
  compact?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxItems: 50,
  showDetails: true,
  autoRefresh: true,
  compact: false
})

// Store
const systemStore = useSystemStore()

// Local state
const searchQuery = ref('')
const selectedStatus = ref('')
const sortBy = ref<ProcessSortField>(props.sortBy || 'cpu')
const sortOrder = ref<ProcessSortOrder>(props.sortOrder || 'desc')
const currentPage = ref(1)
const itemsPerPage = ref(props.maxItems)
const selectedProcess = ref<ProcessInfo | null>(null)
const autoRefresh = useStorage('process-list-auto-refresh', props.autoRefresh)

// Auto refresh interval
let { pause: pauseRefresh } = useIntervalFn(() => {
  if (autoRefresh.value) {
    systemStore.refreshMetrics()
  }
}, 2000)

// Computed properties
const processMetrics = computed(() => systemStore.processMetrics)
const processes = computed(() => processMetrics.value?.processes || [])
const totalProcesses = computed(() => processMetrics.value?.totalProcesses || 0)
const runningProcesses = computed(() => processMetrics.value?.runningProcesses || 0)
const sleepingProcesses = computed(() => processMetrics.value?.sleepingProcesses || 0)

// Filter and sort
const filteredProcesses = computed(() => {
  let filtered = [...processes.value]

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(p =>
      p.name.toLowerCase().includes(query) ||
      p.command.toLowerCase().includes(query) ||
      p.user.toLowerCase().includes(query) ||
      p.pid.toString().includes(query)
    )
  }

  // Status filter
  if (selectedStatus.value) {
    filtered = filtered.filter(p => p.status.toLowerCase() === selectedStatus.value.toLowerCase())
  }

  // Additional filters from props
  if (props.filter) {
    if (props.filter.user) {
      filtered = filtered.filter(p => p.user.includes(props.filter!.user))
    }
    if (props.filter.name) {
      filtered = filtered.filter(p => p.name.toLowerCase().includes(props.filter!.name.toLowerCase()))
    }
    if (props.filter.minCpu) {
      filtered = filtered.filter(p => p.cpuPercent >= props.filter!.minCpu!)
    }
    if (props.filter.maxCpu) {
      filtered = filtered.filter(p => p.cpuPercent <= props.filter!.maxCpu!)
    }
    if (props.filter.minMemory) {
      filtered = filtered.filter(p => p.memoryPercent >= props.filter!.minMemory!)
    }
    if (props.filter.maxMemory) {
      filtered = filtered.filter(p => p.memoryPercent <= props.filter!.maxMemory!)
    }
  }

  // Sort
  filtered.sort((a, b) => {
    let aValue: number | string
    let bValue: number | string

    switch (sortBy.value) {
      case 'cpu':
        aValue = a.cpuPercent
        bValue = b.cpuPercent
        break
      case 'memory':
        aValue = a.memoryPercent
        bValue = b.memoryPercent
        break
      case 'pid':
        aValue = a.pid
        bValue = b.pid
        break
      case 'name':
        aValue = a.name.toLowerCase()
        bValue = b.name.toLowerCase()
        break
      case 'status':
        aValue = a.status.toLowerCase()
        bValue = b.status.toLowerCase()
        break
      case 'threads':
        aValue = a.numThreads
        bValue = b.numThreads
        break
      case 'fds':
        aValue = a.numFDs
        bValue = b.numFDs
        break
      default:
        aValue = a.cpuPercent
        bValue = b.cpuPercent
    }

    if (typeof aValue === 'string' && typeof bValue === 'string') {
      return sortOrder.value === 'asc'
        ? aValue.localeCompare(bValue)
        : bValue.localeCompare(aValue)
    }

    return sortOrder.value === 'asc'
      ? (aValue as number) - (bValue as number)
      : (bValue as number) - (aValue as number)
  })

  return filtered
})

// Pagination
const totalPages = computed(() => Math.ceil(filteredProcesses.value.length / itemsPerPage.value))
const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage.value)
const endIndex = computed(() => startIndex.value + itemsPerPage.value)
const paginatedProcesses = computed(() => {
  return filteredProcesses.value.slice(startIndex.value, endIndex.value)
})

// Methods
const formatPercentage = (value: number): string => {
  return `${value.toFixed(1)}%`
}

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

const formatDate = (date: string | Date): string => {
  const d = new Date(date)
  return d.toLocaleString()
}

const getStatusClass = (status: string): string => {
  const lowerStatus = status.toLowerCase()
  if (lowerStatus === 'running') return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
  if (lowerStatus === 'sleeping') return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  if (lowerStatus === 'zombie') return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  if (lowerStatus === 'stopped') return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
}

const setSortBy = (field: ProcessSortField) => {
  if (sortBy.value === field) {
    toggleSortOrder()
  } else {
    sortBy.value = field
    sortOrder.value = 'desc'
  }
}

const toggleSortOrder = () => {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

const toggleAutoRefresh = () => {
  autoRefresh.value = !autoRefresh.value
}

const clearFilters = () => {
  searchQuery.value = ''
  selectedStatus.value = ''
  currentPage.value = 1
}

const selectProcess = (process: ProcessInfo) => {
  if (props.showDetails) {
    selectedProcess.value = process
  }
}

const closeProcessDetails = () => {
  selectedProcess.value = null
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

// Watch for changes
watch([searchQuery, selectedStatus, sortBy, sortOrder], () => {
  currentPage.value = 1
})

watch(() => props.maxItems, (newMax) => {
  itemsPerPage.value = newMax
  currentPage.value = 1
})

// Lifecycle
onUnmounted(() => {
  pauseRefresh()
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.process-list {
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

/* Custom scrollbar for table */
.overflow-x-auto::-webkit-scrollbar {
  height: 6px;
}

.overflow-x-auto::-webkit-scrollbar-track {
  background: #f1f5f9;
}

.overflow-x-auto::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

.overflow-x-auto::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

.dark .overflow-x-auto::-webkit-scrollbar-track {
  background: #374151;
}

.dark .overflow-x-auto::-webkit-scrollbar-thumb {
  background: #6b7280;
}

.dark .overflow-x-auto::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
</style>