import { defineStore } from 'pinia'

// Wails v3 API import
// import { ServiceCoordinator } from '~/wailsjs/go/backend/services/coordinator/ServiceCoordinator'

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
    error: null as string | null,
    serviceCoordinator: new ServiceCoordinator()
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
        this.metrics = await this.serviceCoordinator.GetNetworkMetrics()
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network metrics: ${error}`
        throw error
      }
    },

    async fetchConnections() {
      try {
        this.connections = await this.serviceCoordinator.GetNetworkConnections()
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network connections: ${error}`
        throw error
      }
    },

    async fetchAlerts() {
      try {
        this.alerts = await this.serviceCoordinator.GetNetworkAlerts()
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network alerts: ${error}`
        throw error
      }
    },

    async fetchStatistics() {
      try {
        this.statistics = await this.serviceCoordinator.GetNetworkStatistics()
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network statistics: ${error}`
        throw error
      }
    },

    async fetchConfig() {
      try {
        this.config = await this.serviceCoordinator.GetNetworkConfig()
      } catch (error) {
        this.error = `Failed to fetch network config: ${error}`
        throw error
      }
    },

    async fetchBandwidthData(interfaceName: string) {
      try {
        const data = await this.serviceCoordinator.GetBandwidthData(interfaceName)
        this.bandwidthHistory = data.bandwidthHistory || []
        this.lastUpdateTime = new Date()
        return data
      } catch (error) {
        this.error = `Failed to fetch bandwidth data: ${error}`
        throw error
      }
    },

    async startMonitoring() {
      try {
        await this.serviceCoordinator.StartNetworkMonitoring()
        this.isMonitoring = true
      } catch (error) {
        this.error = `Failed to start monitoring: ${error}`
        throw error
      }
    },

    async stopMonitoring() {
      try {
        await this.serviceCoordinator.StopNetworkMonitoring()
        this.isMonitoring = false
      } catch (error) {
        this.error = `Failed to stop monitoring: ${error}`
        throw error
      }
    },

    async resolveAlert(alertId: string) {
      try {
        await this.serviceCoordinator.ResolveNetworkAlert(alertId)
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
        await this.serviceCoordinator.ClearNetworkAlerts()
        // Remove resolved alerts from local state
        this.alerts = this.alerts.filter(alert => !alert.resolved)
      } catch (error) {
        this.error = `Failed to clear alerts: ${error}`
        throw error
      }
    },

    async updateConfig(config: any) {
      try {
        await this.serviceCoordinator.UpdateNetworkConfig(config)
        this.config = config
      } catch (error) {
        this.error = `Failed to update config: ${error}`
        throw error
      }
    },

    async resetService() {
      try {
        await this.serviceCoordinator.ResetNetworkService()
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
        this.isMonitoring = await this.serviceCoordinator.IsNetworkMonitoring()
      } catch (error) {
        this.error = `Failed to check monitoring status: ${error}`
      }
    }
  }
})