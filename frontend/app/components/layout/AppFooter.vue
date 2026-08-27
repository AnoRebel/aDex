<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="system-info">
        <span class="status-indicator" :class="{ 'online': isOnline }"></span>
        <span class="status-text">{{ isOnline ? 'Connected' : 'Offline' }}</span>
      </div>
      
      <div class="app-info">
        <span>&copy; 2025 aDex</span>
        <span class="separator">|</span>
        <span>Advanced Desktop Experience</span>
      </div>
      
      <div class="performance-metrics">
        <span class="metric">CPU: {{ cpuUsage }}%</span>
        <span class="separator">|</span>
        <span class="metric">MEM: {{ memoryUsage }}%</span>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isOnline = ref(navigator.onLine)
const cpuUsage = ref(0)
const memoryUsage = ref(0)

const updateNetworkStatus = () => {
  isOnline.value = navigator.onLine
}

const updatePerformanceMetrics = () => {
  // Simulate performance metrics
  cpuUsage.value = Math.floor(Math.random() * 30) + 10
  memoryUsage.value = Math.floor(Math.random() * 20) + 40
}

onMounted(() => {
  window.addEventListener('online', updateNetworkStatus)
  window.addEventListener('offline', updateNetworkStatus)
  
  // Update metrics every 5 seconds
  const metricsInterval = setInterval(updatePerformanceMetrics, 5000)
  updatePerformanceMetrics()
  
  onUnmounted(() => {
    window.removeEventListener('online', updateNetworkStatus)
    window.removeEventListener('offline', updateNetworkStatus)
    clearInterval(metricsInterval)
  })
})
</script>

<style scoped>
.app-footer {
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
  padding: 0.5rem 1rem;
  margin-top: auto;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.system-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-secondary);
  transition: all 0.3s ease;
}

.status-indicator.online {
  background: #00ff00;
  box-shadow: 0 0 5px #00ff00;
}

.app-info,
.performance-metrics {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.separator {
  color: var(--border-color);
}

.metric {
  font-family: 'JetBrains Mono', monospace;
}

@media (max-width: 768px) {
  .footer-content {
    flex-direction: column;
    gap: 0.5rem;
    text-align: center;
  }
}
</style>
