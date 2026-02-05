import { defineStore } from 'pinia'
import { StartMonitoring, StopMonitoring } from '~~/bindings'

export const useNetworkStore = defineStore('network', {
  state: () => ({
    metrics: null as any,
    connections: [] as any[],
    alerts: [] as any[],
    statistics: null as any,
    config: null as any,
    bandwidthHistory: [] as any[],
    isMonitoring: false,
    lastUpdateTime: null as Date | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    activeInterfaces: (state) => {
      if (!state.metrics?.interfaces) return []
      return state.metrics.interfaces.filter((iface: any) => iface.isUp)
    },

    activeConnections: (state) => {
      return state.connections.filter(conn => conn.state === 'ESTABLISHED')
    },

    activeAlerts: (state) => {
      return state.alerts.filter(alert => !alert.resolved)
    },

    totalBandwidth: (state) => {
      if (!state.metrics) return 0
      return (state.metrics.totalBytesSent || 0) + (state.metrics.totalBytesRecv || 0)
    },

    criticalAlerts: (state) => {
      return state.alerts.filter(alert => !alert.resolved && alert.severity === 'error')
    }
  },

  actions: {
    async fetchMetrics() {
      try {
        const service = await GetService('network')
        this.metrics = service?.GetMetrics ? await service.GetMetrics() : null
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network metrics: ${error}`
        throw error
      }
    },

    async fetchConnections() {
      try {
        const service = await GetService('network')
        this.connections = service?.GetConnections ? await service.GetConnections() : []
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network connections: ${error}`
        throw error
      }
    },

    async fetchAlerts() {
      try {
        // Network alerts would be fetched from the network service
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network alerts: ${error}`
        throw error
      }
    },

    async fetchStatistics() {
      try {
        // Statistics would be fetched from the network service
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network statistics: ${error}`
        throw error
      }
    },

    async fetchConfig() {
      try {
        const service = await GetService('network')
        this.config = service?.GetConfiguration ? await service.GetConfiguration() : null
      } catch (error) {
        this.error = `Failed to fetch network config: ${error}`
        throw error
      }
    },

    async fetchBandwidthData(interfaceName: string) {
      try {
        const service = await GetService('network')
        const data = service?.GetBandwidthData ? await service.GetBandwidthData(interfaceName) : null
        this.bandwidthHistory = data?.bandwidthHistory || []
        this.lastUpdateTime = new Date()
        return data
      } catch (error) {
        this.error = `Failed to fetch bandwidth data: ${error}`
        throw error
      }
    },

    async startMonitoring() {
      try {
        await StartMonitoring()
        this.isMonitoring = true
      } catch (error) {
        this.error = `Failed to start monitoring: ${error}`
        throw error
      }
    },

    async stopMonitoring() {
      try {
        await StopMonitoring()
        this.isMonitoring = false
      } catch (error) {
        this.error = `Failed to stop monitoring: ${error}`
        throw error
      }
    },

    async resolveAlert(alertId: string) {
      try {
        // Update local alert state
        const alert = this.alerts.find(a => a.id === alertId)
        if (alert) {
          alert.resolved = true
          alert.resolvedAt = new Date()
        }
      } catch (error) {
        this.error = `Failed to resolve alert: ${error}`
        throw error
      }
    },

    async clearAlerts() {
      try {
        // Remove resolved alerts from local state
        this.alerts = this.alerts.filter(alert => !alert.resolved)
      } catch (error) {
        this.error = `Failed to clear alerts: ${error}`
        throw error
      }
    },

    async updateConfig(config: any) {
      try {
        const service = await GetService('network')
        if (service?.UpdateConfiguration) {
          await service.UpdateConfiguration(config)
        }
        this.config = config
      } catch (error) {
        this.error = `Failed to update config: ${error}`
        throw error
      }
    },

    async resetService() {
      try {
        const service = await GetService('network')
        if (service?.Reset) {
          await service.Reset()
        }
        // Reset local state
        this.metrics = null
        this.connections = []
        this.alerts = []
        this.statistics = null
        this.bandwidthHistory = []
        this.error = null
      } catch (error) {
        this.error = `Failed to reset service: ${error}`
        throw error
      }
    },

    async refreshAllData() {
      this.loading = true
      this.error = null
      
      try {
        await Promise.all([
          this.fetchMetrics(),
          this.fetchConnections(),
          this.fetchAlerts(),
          this.fetchStatistics(),
          this.fetchConfig()
        ])
      } catch (error) {
        this.error = `Failed to refresh data: ${error}`
      } finally {
        this.loading = false
      }
    },

    clearError() {
      this.error = null
    },

    // Update monitoring status
    async updateMonitoringStatus() {
      try {
        const service = await GetService('network')
        this.isMonitoring = service?.IsMonitoring ? await service.IsMonitoring() : false
      } catch (error) {
        this.error = `Failed to check monitoring status: ${error}`
      }
    }
  }
})