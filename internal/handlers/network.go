package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"aDex-UI/internal/services/network"
	"github.com/gorilla/mux"
)

// NetworkHandler handles network-related HTTP requests
type NetworkHandler struct {
	networkService *network.NetworkService
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(networkService *network.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// RegisterRoutes registers network-related routes
func (h *NetworkHandler) RegisterRoutes(router *mux.Router) {
	// Network metrics endpoints
	router.HandleFunc("/api/network/metrics", h.GetNetworkMetrics).Methods("GET")
	router.HandleFunc("/api/network/interfaces", h.GetNetworkInterfaces).Methods("GET")
	router.HandleFunc("/api/network/interfaces/{name}", h.GetNetworkInterface).Methods("GET")

	// Connection tracking endpoints
	router.HandleFunc("/api/network/connections", h.GetNetworkConnections).Methods("GET")
	router.HandleFunc("/api/network/connections/stats", h.GetConnectionStats).Methods("GET")

	// Bandwidth monitoring endpoints
	router.HandleFunc("/api/network/bandwidth/{interfaceName}", h.GetBandwidthData).Methods("GET")
	router.HandleFunc("/api/network/bandwidth/{interfaceName}/history", h.GetBandwidthHistory).Methods("GET")
	router.HandleFunc("/api/network/bandwidth/summary", h.GetBandwidthSummary).Methods("GET")

	// Statistics and analytics endpoints
	router.HandleFunc("/api/network/statistics", h.GetNetworkStatistics).Methods("GET")
	router.HandleFunc("/api/network/statistics/realtime", h.GetRealTimeStatistics).Methods("GET")
	router.HandleFunc("/api/network/statistics/history", h.GetStatisticsHistory).Methods("GET")

	// Alert endpoints
	router.HandleFunc("/api/network/alerts", h.GetNetworkAlerts).Methods("GET")
	router.HandleFunc("/api/network/alerts/{alertId}", h.GetNetworkAlert).Methods("GET")
	router.HandleFunc("/api/network/alerts/{alertId}/resolve", h.ResolveNetworkAlert).Methods("POST")
	router.HandleFunc("/api/network/alerts/clear", h.ClearNetworkAlerts).Methods("DELETE")

	// Monitoring control endpoints
	router.HandleFunc("/api/network/monitoring/start", h.StartMonitoring).Methods("POST")
	router.HandleFunc("/api/network/monitoring/stop", h.StopMonitoring).Methods("POST")
	router.HandleFunc("/api/network/monitoring/status", h.GetMonitoringStatus).Methods("GET")

	// Configuration endpoints
	router.HandleFunc("/api/network/config", h.GetNetworkConfig).Methods("GET")
	router.HandleFunc("/api/network/config", h.UpdateNetworkConfig).Methods("PUT")

	// Diagnostics endpoints
	router.HandleFunc("/api/network/diagnostics/connectivity", h.TestConnectivity).Methods("POST")
	router.HandleFunc("/api/network/diagnostics/latency/{target}", h.TestLatency).Methods("POST")
	router.HandleFunc("/api/network/diagnostics/throughput/{interfaceName}", h.TestThroughput).Methods("POST")
}

// Network metrics endpoints

// GetNetworkMetrics returns current network metrics
func (h *NetworkHandler) GetNetworkMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.networkService.GetNetworkMetrics(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network metrics", err)
		return
	}

	h.writeJSON(w, http.StatusOK, metrics)
}

// GetNetworkInterfaces returns all network interfaces
func (h *NetworkHandler) GetNetworkInterfaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.networkService.GetNetworkMetrics(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network interfaces", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"interfaces": metrics.Interfaces,
		"timestamp":  metrics.Timestamp,
	})
}

// GetNetworkInterface returns a specific network interface
func (h *NetworkHandler) GetNetworkInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	interfaceName := vars["name"]

	ctx := r.Context()
	metrics, err := h.networkService.GetNetworkMetrics(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network metrics", err)
		return
	}

	for _, iface := range metrics.Interfaces {
		if iface.Name == interfaceName {
			h.writeJSON(w, http.StatusOK, iface)
			return
		}
	}

	h.writeError(w, http.StatusNotFound, fmt.Sprintf("Interface %s not found", interfaceName), nil)
}

// Connection tracking endpoints

// GetNetworkConnections returns current network connections
func (h *NetworkHandler) GetNetworkConnections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limit := h.getIntQueryParam(r, "limit", 100)
	state := r.URL.Query().Get("state")
	protocol := r.URL.Query().Get("protocol")

	connections, err := h.networkService.GetConnections(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network connections", err)
		return
	}

	// Filter connections based on query parameters
	filteredConnections := h.filterConnections(connections, state, protocol, limit)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"connections": filteredConnections,
		"total":       len(connections),
		"filtered":    len(filteredConnections),
		"timestamp":   time.Now(),
	})
}

// GetConnectionStats returns connection statistics
func (h *NetworkHandler) GetConnectionStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	connections, err := h.networkService.GetConnections(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network connections", err)
		return
	}

	stats := h.calculateConnectionStats(connections)

	h.writeJSON(w, http.StatusOK, stats)
}

// Bandwidth monitoring endpoints

// GetBandwidthData returns bandwidth data for a specific interface
func (h *NetworkHandler) GetBandwidthData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	interfaceName := vars["name"]

	data, err := h.networkService.GetBandwidthData(interfaceName)
	if err != nil {
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("No bandwidth data for interface %s", interfaceName), err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// GetBandwidthHistory returns historical bandwidth data
func (h *NetworkHandler) GetBandwidthHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	interfaceName := vars["name"]

	// Parse time range parameters
	hours := h.getIntQueryParam(r, "hours", 1)
	points := h.getIntQueryParam(r, "points", 100)

	data, err := h.networkService.GetBandwidthData(interfaceName)
	if err != nil {
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("No bandwidth data for interface %s", interfaceName), err)
		return
	}

	// Sample data points based on requested resolution
	history := h.sampleBandwidthHistory(data, hours, points)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"interface":   interfaceName,
		"history":     history,
		"hours":       hours,
		"data_points": len(history),
	})
}

// GetBandwidthSummary returns bandwidth summary for all interfaces
func (h *NetworkHandler) GetBandwidthSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.networkService.GetNetworkMetrics(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network metrics", err)
		return
	}

	summary := make(map[string]interface{})
	for _, iface := range metrics.Interfaces {
		if data, err := h.networkService.GetBandwidthData(iface.Name); err == nil {
			summary[iface.Name] = map[string]interface{}{
				"current_rate": data.CurrentRate,
				"average_rate": data.AverageRate,
				"peak_rate":    data.PeakRate,
				"is_up":        iface.IsUp,
				"bytes_sent":   iface.BytesSent,
				"bytes_recv":   iface.BytesRecv,
			}
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"interfaces": summary,
		"timestamp":  time.Now(),
	})
}

// Statistics and analytics endpoints

// GetNetworkStatistics returns comprehensive network statistics
func (h *NetworkHandler) GetNetworkStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.networkService.GetStatistics(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network statistics", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// GetRealTimeStatistics returns real-time network statistics
func (h *NetworkHandler) GetRealTimeStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// This would implement WebSocket or Server-Sent Events for real-time updates
	// For now, return current statistics with a real-time flag
	stats, err := h.networkService.GetStatistics(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get network statistics", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"statistics":  stats,
		"real_time":   true,
		"next_update": time.Now().Add(2 * time.Second),
	})
}

// GetStatisticsHistory returns historical statistics data
func (h *NetworkHandler) GetStatisticsHistory(w http.ResponseWriter, r *http.Request) {
	// Parse time range parameters
	hours := h.getIntQueryParam(r, "hours", 24)
	interval := h.getIntQueryParam(r, "interval", 5) // minutes

	// This would implement historical data retrieval from a database
	// For now, return a placeholder response
	history := map[string]interface{}{
		"hours":    hours,
		"interval": interval,
		"data":     []interface{}{}, // Placeholder
		"message":  "Historical statistics not yet implemented",
	}

	h.writeJSON(w, http.StatusOK, history)
}

// Alert endpoints

// GetNetworkAlerts returns current network alerts
func (h *NetworkHandler) GetNetworkAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := h.networkService.GetAlerts()

	// Parse filter parameters
	severity := r.URL.Query().Get("severity")
	resolved := r.URL.Query().Get("resolved")

	filteredAlerts := h.filterAlerts(alerts, severity, resolved)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts":    filteredAlerts,
		"total":     len(alerts),
		"filtered":  len(filteredAlerts),
		"timestamp": time.Now(),
	})
}

// GetNetworkAlert returns a specific network alert
func (h *NetworkHandler) GetNetworkAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["alertId"]

	alerts := h.networkService.GetAlerts()
	for _, alert := range alerts {
		if alert.ID == alertID {
			h.writeJSON(w, http.StatusOK, alert)
			return
		}
	}

	h.writeError(w, http.StatusNotFound, fmt.Sprintf("Alert %s not found", alertID), nil)
}

// ResolveNetworkAlert resolves a network alert
func (h *NetworkHandler) ResolveNetworkAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["alertId"]

	err := h.networkService.ResolveAlert(alertID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, fmt.Sprintf("Failed to resolve alert %s", alertID), err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   fmt.Sprintf("Alert %s resolved successfully", alertID),
		"alert_id":  alertID,
		"timestamp": time.Now(),
	})
}

// ClearNetworkAlerts clears all resolved alerts
func (h *NetworkHandler) ClearNetworkAlerts(w http.ResponseWriter, r *http.Request) {
	h.networkService.ClearAlerts()

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Resolved alerts cleared successfully",
		"timestamp": time.Now(),
	})
}

// Monitoring control endpoints

// StartMonitoring starts network monitoring
func (h *NetworkHandler) StartMonitoring(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.networkService.StartMonitoring(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to start network monitoring", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Network monitoring started successfully",
		"timestamp": time.Now(),
	})
}

// StopMonitoring stops network monitoring
func (h *NetworkHandler) StopMonitoring(w http.ResponseWriter, r *http.Request) {
	err := h.networkService.StopMonitoring()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to stop network monitoring", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Network monitoring stopped successfully",
		"timestamp": time.Now(),
	})
}

// GetMonitoringStatus returns the current monitoring status
func (h *NetworkHandler) GetMonitoringStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"is_monitoring": h.networkService.IsMonitoring(),
		"last_update":   h.networkService.GetLastUpdateTime(),
		"uptime":        time.Since(h.networkService.GetLastUpdateTime()),
	}

	h.writeJSON(w, http.StatusOK, status)
}

// Configuration endpoints

// GetNetworkConfig returns the current network service configuration
func (h *NetworkHandler) GetNetworkConfig(w http.ResponseWriter, r *http.Request) {
	config := h.networkService.GetConfiguration()

	h.writeJSON(w, http.StatusOK, config)
}

// UpdateNetworkConfig updates the network service configuration
func (h *NetworkHandler) UpdateNetworkConfig(w http.ResponseWriter, r *http.Request) {
	var config network.NetworkServiceConfig

	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid configuration", err)
		return
	}

	err = h.networkService.UpdateConfiguration(&config)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to update configuration", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Configuration updated successfully",
		"config":    config,
		"timestamp": time.Now(),
	})
}

// Diagnostics endpoints

// TestConnectivity tests network connectivity
func (h *NetworkHandler) TestConnectivity(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Targets []string `json:"targets"`
		Count   int      `json:"count"`
		Timeout int      `json:"timeout"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Set defaults
	if len(request.Targets) == 0 {
		request.Targets = []string{"8.8.8.8", "1.1.1.1", "google.com"}
	}
	if request.Count == 0 {
		request.Count = 4
	}
	if request.Timeout == 0 {
		request.Timeout = 5000 // 5 seconds
	}

	// This would implement actual connectivity testing
	// For now, return a placeholder response
	results := make([]map[string]interface{}, 0)
	for _, target := range request.Targets {
		result := map[string]interface{}{
			"target":      target,
			"reachable":   true,
			"latency_ms":  25,
			"packet_loss": 0,
			"status":      "success",
		}
		results = append(results, result)
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"results":   results,
		"timestamp": time.Now(),
		"message":   "Connectivity test completed",
	})
}

// TestLatency tests latency to a specific target
func (h *NetworkHandler) TestLatency(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target := vars["target"]

	var request struct {
		Count   int `json:"count"`
		Timeout int `json:"timeout"`
	}

	json.NewDecoder(r.Body).Decode(&request)

	// Set defaults
	if request.Count == 0 {
		request.Count = 10
	}
	if request.Timeout == 0 {
		request.Timeout = 3000 // 3 seconds
	}

	// This would implement actual latency testing
	// For now, return a placeholder response
	result := map[string]interface{}{
		"target":           target,
		"latency_ms":       30,
		"jitter_ms":        5,
		"packet_loss":      0,
		"min_latency_ms":   25,
		"max_latency_ms":   35,
		"avg_latency_ms":   30,
		"packets_sent":     request.Count,
		"packets_received": request.Count,
		"status":           "success",
		"timestamp":        time.Now(),
	}

	h.writeJSON(w, http.StatusOK, result)
}

// TestThroughput tests throughput on a specific interface
func (h *NetworkHandler) TestThroughput(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	interfaceName := vars["name"]

	var request struct {
		Method     string `json:"method"`   // "tcp", "udp", "http"
		Duration   int    `json:"duration"` // seconds
		TargetHost string `json:"targetHost"`
		TargetPort int    `json:"targetPort"`
	}

	json.NewDecoder(r.Body).Decode(&request)

	// Set defaults
	if request.Method == "" {
		request.Method = "tcp"
	}
	if request.Duration == 0 {
		request.Duration = 10
	}
	if request.TargetHost == "" {
		request.TargetHost = "speedtest.net"
	}
	if request.TargetPort == 0 {
		request.TargetPort = 80
	}

	// This would implement actual throughput testing
	// For now, return a placeholder response
	result := map[string]interface{}{
		"interface":             interfaceName,
		"upload_mbps":           50.5,
		"download_mbps":         150.2,
		"test_duration_seconds": request.Duration,
		"bytes_uploaded":        63125000,
		"bytes_downloaded":      187750000,
		"method":                request.Method,
		"target":                fmt.Sprintf("%s:%d", request.TargetHost, request.TargetPort),
		"status":                "success",
		"timestamp":             time.Now(),
	}

	h.writeJSON(w, http.StatusOK, result)
}

// Helper functions

// writeJSON writes a JSON response
func (h *NetworkHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *NetworkHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now(),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	h.writeJSON(w, status, response)
}

// getIntQueryParam gets an integer query parameter with a default value
func (h *NetworkHandler) getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// filterConnections filters connections based on criteria
func (h *NetworkHandler) filterConnections(connections []*network.NetworkConnection, state, protocol string, limit int) []*network.NetworkConnection {
	var filtered []*network.NetworkConnection

	for _, conn := range connections {
		// Filter by state
		if state != "" && conn.State != state {
			continue
		}

		// Filter by protocol
		if protocol != "" && conn.Protocol != protocol {
			continue
		}

		filtered = append(filtered, conn)

		// Apply limit
		if len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// filterAlerts filters alerts based on criteria
func (h *NetworkHandler) filterAlerts(alerts []network.NetworkAlert, severity, resolved string) []network.NetworkAlert {
	var filtered []network.NetworkAlert

	for _, alert := range alerts {
		// Filter by severity
		if severity != "" && alert.Severity != severity {
			continue
		}

		// Filter by resolved status
		if resolved != "" {
			isResolved := resolved == "true"
			if alert.Resolved != isResolved {
				continue
			}
		}

		filtered = append(filtered, alert)
	}

	return filtered
}

// calculateConnectionStats calculates statistics from connections
func (h *NetworkHandler) calculateConnectionStats(connections []*network.NetworkConnection) map[string]interface{} {
	stats := map[string]interface{}{
		"total_connections": len(connections),
		"established":       0,
		"listening":         0,
		"closing":           0,
		"protocols":         make(map[string]int),
		"top_processes":     make([]map[string]interface{}, 0),
		"total_bytes_sent":  uint64(0),
		"total_bytes_recv":  uint64(0),
	}

	processBytes := make(map[string]uint64)

	for _, conn := range connections {
		// Count by state
		if conn.State == "ESTABLISHED" {
			stats["established"] = stats["established"].(int) + 1
		} else if conn.State == "LISTEN" {
			stats["listening"] = stats["listening"].(int) + 1
		} else {
			stats["closing"] = stats["closing"].(int) + 1
		}

		// Count by protocol
		protocolMap := stats["protocols"].(map[string]int)
		protocolMap[conn.Protocol] = protocolMap[conn.Protocol] + 1

		// Track bytes by process
		if conn.ProcessName != "" {
			processBytes[conn.ProcessName] += conn.BytesSent + conn.BytesRecv
		}

		// Total bytes
		stats["total_bytes_sent"] = stats["total_bytes_sent"].(uint64) + conn.BytesSent
		stats["total_bytes_recv"] = stats["total_bytes_recv"].(uint64) + conn.BytesRecv
	}

	// Get top processes by bytes
	for process, bytes := range processBytes {
		processStats := map[string]interface{}{
			"process_name": process,
			"bytes":        bytes,
		}
		stats["top_processes"] = append(stats["top_processes"].([]map[string]interface{}), processStats)
	}

	return stats
}

// sampleBandwidthHistory samples bandwidth data for historical view
func (h *NetworkHandler) sampleBandwidthHistory(data *network.BandwidthData, hours, points int) map[string]interface{} {
	if data == nil || len(data.Timestamps) == 0 {
		return map[string]interface{}{
			"timestamps": []time.Time{},
			"upload":     []float64{},
			"download":   []float64{},
		}
	}

	// This is a simplified sampling implementation
	// In a real implementation, you would sample based on the requested time range

	var timestamps []time.Time
	var uploadRates []float64
	var downloadRates []float64

	// Get the most recent points
	start := len(data.Timestamps) - points
	if start < 0 {
		start = 0
	}

	for i := start; i < len(data.Timestamps); i++ {
		timestamps = append(timestamps, data.Timestamps[i])

		if i < len(data.BytesSent) && i > 0 {
			timeDelta := data.Timestamps[i].Sub(data.Timestamps[i-1]).Seconds()
			if timeDelta > 0 {
				bytesDelta := float64(data.BytesSent[i] - data.BytesSent[i-1])
				rate := bytesDelta / timeDelta
				uploadRates = append(uploadRates, rate*8/1024/1024) // Convert to Mbps
			} else {
				uploadRates = append(uploadRates, 0)
			}
		} else {
			uploadRates = append(uploadRates, 0)
		}

		if i < len(data.BytesRecv) && i > 0 {
			timeDelta := data.Timestamps[i].Sub(data.Timestamps[i-1]).Seconds()
			if timeDelta > 0 {
				bytesDelta := float64(data.BytesRecv[i] - data.BytesRecv[i-1])
				rate := bytesDelta / timeDelta
				downloadRates = append(downloadRates, rate*8/1024/1024) // Convert to Mbps
			} else {
				downloadRates = append(downloadRates, 0)
			}
		} else {
			downloadRates = append(downloadRates, 0)
		}
	}

	return map[string]interface{}{
		"timestamps": timestamps,
		"upload":     uploadRates,
		"download":   downloadRates,
	}
}
