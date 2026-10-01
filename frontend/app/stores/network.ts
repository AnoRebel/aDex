import { defineStore } from 'pinia'
import {
  GetNetworkMetrics,
  GetNetworkConnections,
  GetNetworkAlerts,
  GetNetworkStatistics,
  GetNetworkConfig,
  UpdateNetworkConfig,
  GetBandwidthData,
  ResetNetworkService,
  IsNetworkMonitoring,
  StartNetworkMonitoring,
  StopNetworkMonitoring,
  ResolveNetworkAlert,
  ClearNetworkAlerts,
} from '~/lib/wailsjs/coordinator'

// Pre-V2 this store called `GetService('network')` and then invoked methods
// on the returned struct. Wails marshals every binding return through JSON,
// so the "service" arrived on the frontend as a plain DTO with no methods —
// every subsequent call was a no-op (or worse, a launch crash:
// `json: unsupported type: func() error`). Migrated to the dedicated
// coordinator bindings, which return only JSON-safe payloads.

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
  }),

  getters: {
    activeInterfaces: (state) => {
      if (!state.metrics?.interfaces) return []
      return state.metrics.interfaces.filter((iface: any) => iface.isUp)
    },
    activeConnections: (state) =>
      state.connections.filter((c) => c.state === 'ESTABLISHED'),
    activeAlerts: (state) => state.alerts.filter((a) => !a.resolved),
    totalBandwidth: (state) => {
      if (!state.metrics) return 0
      return (state.metrics.totalBytesSent || 0) + (state.metrics.totalBytesRecv || 0)
    },
    criticalAlerts: (state) =>
      state.alerts.filter((a) => !a.resolved && a.severity === 'error'),
  },

  actions: {
    async fetchMetrics() {
      try {
        this.metrics = await GetNetworkMetrics()
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network metrics: ${error}`
        throw error
      }
    },

    async fetchConnections() {
      try {
        const data = await GetNetworkConnections()
        this.connections = Array.isArray(data) ? data : (data?.connections ?? [])
        this.lastUpdateTime = new Date()
      } catch (error) {
        this.error = `Failed to fetch network connections: ${error}`
        throw error
      }
    },

    async fetchAlerts() {
      try {
        const data = await GetNetworkAlerts()
        this.alerts = Array.isArray(data) ? data : (data?.alerts ?? [])
        this.lastUpdateTime = new Date()
      } catch {
        // Non-critical: alerts may not be available on all backends.
        this.lastUpdateTime = new Date()
      }
    },

    async fetchStatistics() {
      try {
        this.statistics = await GetNetworkStatistics()
        this.lastUpdateTime = new Date()
      } catch {
        this.lastUpdateTime = new Date()
      }
    },

    async fetchConfig() {
      try {
        this.config = await GetNetworkConfig()
      } catch (error) {
        this.error = `Failed to fetch network config: ${error}`
        throw error
      }
    },

    async fetchBandwidthData(interfaceName: string) {
      try {
        const data = await GetBandwidthData(interfaceName)
        this.bandwidthHistory = data?.bandwidthHistory ?? []
        this.lastUpdateTime = new Date()
        return data
      } catch (error) {
        this.error = `Failed to fetch bandwidth data: ${error}`
        throw error
      }
    },

    async startMonitoring() {
      try {
        await StartNetworkMonitoring()
        this.isMonitoring = true
      } catch (error) {
        this.error = `Failed to start monitoring: ${error}`
        throw error
      }
    },

    async stopMonitoring() {
      try {
        await StopNetworkMonitoring()
        this.isMonitoring = false
      } catch (error) {
        this.error = `Failed to stop monitoring: ${error}`
        throw error
      }
    },

    async resolveAlert(alertId: string) {
      try {
        await ResolveNetworkAlert(alertId)
        const alert = this.alerts.find((a) => a.id === alertId)
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
        await ClearNetworkAlerts()
        this.alerts = this.alerts.filter((a) => !a.resolved)
      } catch (error) {
        this.error = `Failed to clear alerts: ${error}`
        throw error
      }
    },

    async updateConfig(config: any) {
      try {
        await UpdateNetworkConfig(config)
        this.config = config
      } catch (error) {
        this.error = `Failed to update config: ${error}`
        throw error
      }
    },

    async resetService() {
      try {
        await ResetNetworkService()
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
          this.fetchConfig(),
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

    async updateMonitoringStatus() {
      try {
        this.isMonitoring = await IsNetworkMonitoring()
      } catch (error) {
        this.error = `Failed to check monitoring status: ${error}`
      }
    },

    async initialize() {
      try {
        this.loading = true
        this.error = null
        await this.startMonitoring()
        await this.fetchMetrics()
      } catch (error) {
        this.error = `Failed to initialize network store: ${error}`
        console.error('Network store initialization failed:', error)
      } finally {
        this.loading = false
      }
    },
  },
})
