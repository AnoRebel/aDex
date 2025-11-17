<template>
  <div class="connections-list">
    <div class="list-header">
      <h3>Active Connections</h3>
      <div class="header-controls">
        <select v-model="filterState" class="filter-select">
          <option value="all">All States</option>
          <option value="ESTABLISHED">Established</option>
          <option value="LISTEN">Listening</option>
          <option value="TIME_WAIT">Time Wait</option>
        </select>
        <select v-model="sortBy" class="sort-select">
          <option value="bandwidth">By Bandwidth</option>
          <option value="process">By Process</option>
          <option value="time">By Duration</option>
        </select>
      </div>
    </div>
    
    <div class="list-content">
      <div v-if="loading" class="list-loading">
        <div class="loading-spinner"></div>
        <p>Loading connections...</p>
      </div>
      
      <div v-else-if="filteredConnections.length === 0" class="list-empty">
        <div class="empty-icon">🔗</div>
        <p>No active connections found</p>
      </div>
      
      <div v-else class="connections-table">
        <div class="table-header">
          <div class="header-cell local">Local Address</div>
          <div class="header-cell remote">Remote Address</div>
          <div class="header-cell protocol">Protocol</div>
          <div class="header-cell process">Process</div>
          <div class="header-cell bandwidth">Bandwidth</div>
        </div>
        
        <div class="table-body">
          <div
            v-for="connection in paginatedConnections"
            :key="getConnectionKey(connection)"
            class="table-row"
            @click="selectConnection(connection)"
          >
            <div class="cell local">
              <div class="address">{{ connection.localAddr }}</div>
              <div class="state" :class="getStateClass(connection.state)">{{ connection.state }}</div>
            </div>
            <div class="cell remote">
              <div class="address">{{ connection.remoteAddr }}</div>
            </div>
            <div class="cell protocol">
              <span class="protocol-badge" :class="connection.protocol.toLowerCase()">
                {{ connection.protocol }}
              </span>
            </div>
            <div class="cell process">
              <div class="process-name">{{ connection.processName || 'Unknown' }}</div>
              <div class="process-pid" v-if="connection.pid">PID: {{ connection.pid }}</div>
            </div>
            <div class="cell bandwidth">
              <div class="bandwidth-bars">
                <div class="bandwidth-bar upload">
                  <div class="bar-fill upload-fill" :style="{ width: getUploadPercentage(connection) + '%' }"></div>
                </div>
                <div class="bandwidth-bar download">
                  <div class="bar-fill download-fill" :style="{ width: getDownloadPercentage(connection) + '%' }"></div>
                </div>
              </div>
              <div class="bandwidth-text">
                ↑ {{ formatBytes(connection.bytesSent) }}
                <br>
                ↓ {{ formatBytes(connection.bytesRecv) }}
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Pagination -->
      <div class="pagination" v-if="totalPages > 1">
        <button
          @click="currentPage = Math.max(1, currentPage - 1)"
          :disabled="currentPage === 1"
          class="pagination-btn"
        >
          Previous
        </button>
        <span class="page-info">
          Page {{ currentPage }} of {{ totalPages }}
        </span>
        <button
          @click="currentPage = Math.min(totalPages, currentPage + 1)"
          :disabled="currentPage === totalPages"
          class="pagination-btn"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'

const props = defineProps<{
  connections: any[]
  loading?: boolean
}>()

const emit = defineEmits<{
  'connection-selected': [connection: any]
}>()

// Reactive data
const filterState = ref('all')
const sortBy = ref('bandwidth')
const currentPage = ref(1)
const itemsPerPage = ref(10)

// Computed properties
const filteredConnections = computed(() => {
  let filtered = props.connections || []
  
  // Filter by state
  if (filterState.value !== 'all') {
    filtered = filtered.filter(conn => conn.state === filterState.value)
  }
  
  // Sort
  filtered.sort((a, b) => {
    switch (sortBy.value) {
      case 'bandwidth':
        const totalA = (a.bytesSent || 0) + (a.bytesRecv || 0)
        const totalB = (b.bytesSent || 0) + (b.bytesRecv || 0)
        return totalB - totalA
      case 'process':
        return (a.processName || '').localeCompare(b.processName || '')
      case 'time':
        return (b.established?.getTime() || 0) - (a.established?.getTime() || 0)
      default:
        return 0
    }
  })
  
  return filtered
})

const totalPages = computed(() => {
  return Math.ceil(filteredConnections.value.length / itemsPerPage.value)
})

const paginatedConnections = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredConnections.value.slice(start, end)
})

const maxBandwidth = computed(() => {
  if (!props.connections || props.connections.length === 0) return 1
  return Math.max(...props.connections.map(conn => 
    (conn.bytesSent || 0) + (conn.bytesRecv || 0)
  ))
})

// Methods
const getConnectionKey = (connection: any): string => {
  return `${connection.protocol}-${connection.localAddr}-${connection.remoteAddr}-${connection.pid}`
}

const getStateClass = (state: string): string => {
  switch (state) {
    case 'ESTABLISHED':
      return 'state-established'
    case 'LISTEN':
      return 'state-listening'
    case 'TIME_WAIT':
      return 'state-waiting'
    default:
      return 'state-other'
  }
}

const getUploadPercentage = (connection: any): number => {
  return Math.min(((connection.bytesSent || 0) / maxBandwidth.value) * 100, 100)
}

const getDownloadPercentage = (connection: any): number => {
  return Math.min(((connection.bytesRecv || 0) / maxBandwidth.value) * 100, 100)
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

const selectConnection = (connection: any) => {
  emit('connection-selected', connection)
}

// Watchers
watch(filterState, () => {
  currentPage.value = 1
})

watch(sortBy, () => {
  currentPage.value = 1
})
</script>

<style scoped>
.connections-list {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-secondary);
}

.list-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.header-controls {
  display: flex;
  gap: 0.5rem;
}

.filter-select,
.sort-select {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-primary);
  font-size: 0.75rem;
}

.list-content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

.list-loading,
.list-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--text-secondary);
  text-align: center;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-secondary);
  border-top: 3px solid var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.connections-table {
  width: 100%;
  font-size: 0.75rem;
}

.table-header {
  display: grid;
  grid-template-columns: 2fr 2fr 1fr 2fr 2fr;
  gap: 0.5rem;
  padding: 0.75rem;
  background: var(--surface-secondary);
  border: 1px solid var(--border-secondary);
  border-radius: 4px 4px 0 0;
  font-weight: 600;
  color: var(--text-secondary);
}

.header-cell {
  display: flex;
  align-items: center;
}

.table-body {
  border: 1px solid var(--border-secondary);
  border-top: none;
  border-radius: 0 0 4px 4px;
  max-height: 400px;
  overflow-y: auto;
}

.table-row {
  display: grid;
  grid-template-columns: 2fr 2fr 1fr 2fr 2fr;
  gap: 0.5rem;
  padding: 0.75rem;
  border-bottom: 1px solid var(--border-secondary);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.table-row:hover {
  background: var(--surface-secondary);
}

.table-row:last-child {
  border-bottom: none;
}

.cell {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.address {
  font-family: 'Courier New', monospace;
  font-size: 0.7rem;
  color: var(--text-primary);
  word-break: break-all;
}

.state {
  font-size: 0.65rem;
  padding: 0.125rem 0.375rem;
  border-radius: 10px;
  font-weight: 500;
  margin-top: 0.25rem;
  width: fit-content;
}

.state-established {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success-primary);
}

.state-listening {
  background: rgba(59, 130, 246, 0.1);
  color: var(--accent-primary);
}

.state-waiting {
  background: rgba(245, 158, 11, 0.1);
  color: var(--warning-primary);
}

.state-other {
  background: rgba(156, 163, 175, 0.1);
  color: var(--text-secondary);
}

.protocol-badge {
  font-size: 0.65rem;
  padding: 0.125rem 0.375rem;
  border-radius: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.protocol-badge.tcp {
  background: rgba(59, 130, 246, 0.1);
  color: var(--accent-primary);
}

.protocol-badge.udp {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success-primary);
}

.protocol-badge.icmp {
  background: rgba(245, 158, 11, 0.1);
  color: var(--warning-primary);
}

.process-name {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.7rem;
}

.process-pid {
  font-size: 0.6rem;
  color: var(--text-secondary);
  margin-top: 0.125rem;
}

.bandwidth-bars {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.5rem;
}

.bandwidth-bar {
  height: 4px;
  background: var(--surface-tertiary);
  border-radius: 2px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

.upload-fill {
  background: linear-gradient(90deg, #f59e0b, #d97706);
}

.download-fill {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.bandwidth-text {
  font-size: 0.6rem;
  color: var(--text-secondary);
  line-height: 1.2;
  font-family: 'Courier New', monospace;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border-top: 1px solid var(--border-secondary);
  margin-top: 1rem;
}

.pagination-btn {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--border-secondary);
  border-radius: 4px;
  background: var(--surface-primary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.pagination-btn:hover:not(:disabled) {
  background: var(--surface-secondary);
  border-color: var(--border-primary);
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

/* Responsive design */
@media (max-width: 768px) {
  .list-header {
    flex-direction: column;
    gap: 0.75rem;
    align-items: stretch;
  }
  
  .header-controls {
    justify-content: center;
  }
  
  .table-header,
  .table-row {
    grid-template-columns: 1.5fr 1.5fr 1fr 1.5fr;
  }
  
  .cell.bandwidth {
    grid-column: span 1;
  }
}
</style>