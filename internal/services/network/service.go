package network

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"aDex/internal/events"
	"aDex/internal/models"
	"aDex/internal/services/geoip"
	networkLib "github.com/shirou/gopsutil/v3/net"
)

// NetworkService provides comprehensive network monitoring functionality
type NetworkService struct {
	eventBus      events.IEventBus
	config        *NetworkServiceConfig
	connections   map[string]*NetworkConnection
	bandwidthData map[string]*BandwidthData
	alerts        []NetworkAlert
	mutex         sync.RWMutex
	isMonitoring  bool
	lastUpdate    time.Time
	geoipService  *geoip.GeoIPService
}

// NetworkServiceConfig contains configuration for the network service
type NetworkServiceConfig struct {
	RefreshInterval           time.Duration    `json:"refreshInterval"`
	EnableConnectionTracking  bool             `json:"enableConnectionTracking"`
	EnableBandwidthMonitoring bool             `json:"enableBandwidthMonitoring"`
	EnableAlerts              bool             `json:"enableAlerts"`
	MaxConnections            int              `json:"maxConnections"`
	BandwidthHistorySize      int              `json:"bandwidthHistorySize"`
	AlertThresholds           *AlertThresholds `json:"alertThresholds"`
}

// AlertThresholds defines thresholds for network alerts
type AlertThresholds struct {
	BandwidthUsageMBps float64 `json:"bandwidthUsageMBps"` // MB/s
	ConnectionCount    int     `json:"connectionCount"`
	PacketLossPercent  float64 `json:"packetLossPercent"`
	ErrorRatePercent   float64 `json:"errorRatePercent"`
	HighLatencyMs      int     `json:"highLatencyMs"`
}

// NetworkConnection represents an active network connection
type NetworkConnection struct {
	LocalAddr    string    `json:"localAddr"`
	RemoteAddr   string    `json:"remoteAddr"`
	State        string    `json:"state"`
	PID          int       `json:"pid"`
	ProcessName  string    `json:"processName"`
	Protocol     string    `json:"protocol"`
	BytesSent    uint64    `json:"bytesSent"`
	BytesRecv    uint64    `json:"bytesRecv"`
	Established  time.Time `json:"established"`
	LastActivity time.Time `json:"lastActivity"`
}

// BandwidthData tracks bandwidth usage over time
type BandwidthData struct {
	InterfaceName string         `json:"interfaceName"`
	Timestamps    []time.Time    `json:"timestamps"`
	BytesSent     []uint64       `json:"bytesSent"`
	BytesRecv     []uint64       `json:"bytesRecv"`
	PacketsSent   []uint64       `json:"packetsSent"`
	PacketsRecv   []uint64       `json:"packetsRecv"`
	CurrentRate   *BandwidthRate `json:"currentRate"`
	AverageRate   *BandwidthRate `json:"averageRate"`
	PeakRate      *BandwidthRate `json:"peakRate"`
}

// BandwidthRate represents bandwidth usage rate
type BandwidthRate struct {
	UploadBps    float64 `json:"uploadBps"`    // Bytes per second
	DownloadBps  float64 `json:"downloadBps"`  // Bytes per second
	UploadMbps   float64 `json:"uploadMbps"`   // Megabits per second
	DownloadMbps float64 `json:"downloadMbps"` // Megabits per second
}

// NetworkAlert represents a network-related alert
type NetworkAlert struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Message    string                 `json:"message"`
	Interface  string                 `json:"interface"`
	Threshold  float64                `json:"threshold"`
	Current    float64                `json:"current"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolvedAt,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NetworkStatistics represents comprehensive network statistics
type NetworkStatistics struct {
	TotalConnections  int                    `json:"totalConnections"`
	ActiveConnections int                    `json:"activeConnections"`
	TotalBytesSent    uint64                 `json:"totalBytesSent"`
	TotalBytesRecv    uint64                 `json:"totalBytesRecv"`
	TotalPacketsSent  uint64                 `json:"totalPacketsSent"`
	TotalPacketsRecv  uint64                 `json:"totalPacketsRecv"`
	PacketLossRate    float64                `json:"packetLossRate"`
	ErrorRate         float64                `json:"errorRate"`
	AverageLatency    time.Duration          `json:"averageLatency"`
	TopConnections    []*NetworkConnection   `json:"topConnections"`
	InterfaceStats    map[string]interface{} `json:"interfaceStats"`
	Timestamp         time.Time              `json:"timestamp"`
}

// DefaultNetworkServiceConfig returns default configuration
func DefaultNetworkServiceConfig() *NetworkServiceConfig {
	return &NetworkServiceConfig{
		RefreshInterval:           2000 * time.Millisecond,
		EnableConnectionTracking:  true,
		EnableBandwidthMonitoring: true,
		EnableAlerts:              true,
		MaxConnections:            1000,
		BandwidthHistorySize:      300, // 5 minutes at 2-second intervals
		AlertThresholds: &AlertThresholds{
			BandwidthUsageMBps: 100.0, // 100 MB/s
			ConnectionCount:    500,
			PacketLossPercent:  1.0,
			ErrorRatePercent:   0.1,
			HighLatencyMs:      1000,
		},
	}
}

// NewNetworkService creates a new network service instance
func NewNetworkService() *NetworkService {
	return NewNetworkServiceWithConfig(DefaultNetworkServiceConfig())
}

// NewNetworkServiceWithConfig creates a new network service with custom configuration
func NewNetworkServiceWithConfig(config *NetworkServiceConfig) *NetworkService {
	service := &NetworkService{
		config:        config,
		connections:   make(map[string]*NetworkConnection),
		bandwidthData: make(map[string]*BandwidthData),
		alerts:        make([]NetworkAlert, 0),
		lastUpdate:    time.Now(),
	}

	// Initialize GeoIP service (lazy loaded)
	service.geoipService = geoip.NewGeoIPService(nil)

	return service
}

// SetEventBus sets the event bus for the service
func (s *NetworkService) SetEventBus(eventBus events.IEventBus) {
	s.eventBus = eventBus
}

// GetEventBus returns the current event bus
func (s *NetworkService) GetEventBus() events.IEventBus {
	return s.eventBus
}

// GetNetworkMetrics returns current network interface metrics
func (s *NetworkService) GetNetworkMetrics(ctx context.Context) (*models.NetworkMetrics, error) {
	interfaces, err := networkLib.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Get IO counters for statistics
	ioCounters, err := networkLib.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network IO counters: %w", err)
	}

	// Build a map of IO counters by interface name
	ioCounterMap := make(map[string]networkLib.IOCountersStat)
	for _, counter := range ioCounters {
		ioCounterMap[counter.Name] = counter
	}

	// Carrier-detection map: gopsutil's "up" flag is just IFF_UP (admin
	// up), so an unplugged ethernet still reports up. The stdlib's
	// FlagRunning corresponds to IFF_RUNNING which the kernel only
	// asserts when the link has carrier. We use this to mark
	// link-active interfaces honestly so the "auto-pick first up"
	// fallback in the frontend stops landing on dead ethernet ports
	// when the user is actually on wifi.
	stdRunning := map[string]bool{}
	if stdIfs, stdErr := net.Interfaces(); stdErr == nil {
		for _, si := range stdIfs {
			if si.Flags&net.FlagRunning != 0 {
				stdRunning[si.Name] = true
			}
		}
	}

	var networkInterfaces []models.NetworkInterface
	var totalSent, totalRecv uint64

	for _, iface := range interfaces {
		// Skip loopback interfaces
		if iface.Name == "lo" || iface.Name == "lo0" {
			continue
		}

		// "Up" means BOTH admin-up AND carrier-present. Without the
		// carrier check, an unplugged ethernet (admin-up but no link)
		// would beat an active wifi adapter alphabetically and the
		// netstat panel would proudly show 0.0.0.0 for eno1.
		adminUp := false
		for _, flag := range iface.Flags {
			if strings.ToLower(flag) == "up" {
				adminUp = true
				break
			}
		}
		isUp := adminUp && stdRunning[iface.Name]

		// Get IO counters for this interface
		counter, hasCounter := ioCounterMap[iface.Name]

		// Convert addresses to string slice
		var ipAddresses []string
		for _, addr := range iface.Addrs {
			ipAddresses = append(ipAddresses, addr.Addr)
		}

		networkIface := models.NetworkInterface{
			Name:        iface.Name,
			IsUp:        isUp,
			MAC:         iface.HardwareAddr,
			MTU:         uint64(iface.MTU),
			IPAddresses: ipAddresses,
		}

		if hasCounter {
			networkIface.BytesSent = counter.BytesSent
			networkIface.BytesRecv = counter.BytesRecv
			networkIface.PacketsSent = counter.PacketsSent
			networkIface.PacketsRecv = counter.PacketsRecv
			networkIface.Errin = counter.Errin
			networkIface.Errout = counter.Errout
			networkIface.Dropin = counter.Dropin
			networkIface.Dropout = counter.Dropout
			totalSent += counter.BytesSent
			totalRecv += counter.BytesRecv
		}

		networkInterfaces = append(networkInterfaces, networkIface)

		// Update bandwidth data if enabled
		if s.config.EnableBandwidthMonitoring {
			s.updateBandwidthData(&networkIface)
		}
	}

	metrics := &models.NetworkMetrics{
		Interfaces:     networkInterfaces,
		TotalBytesSent: totalSent,
		TotalBytesRecv: totalRecv,
		Timestamp:      time.Now(),
	}

	s.lastUpdate = time.Now()

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.NetworkUpdated, map[string]interface{}{
			"metrics":   metrics,
			"timestamp": time.Now(),
		}, "network-service")
	}

	return metrics, nil
}

// GetConnections returns current network connections
func (s *NetworkService) GetConnections(ctx context.Context) ([]*NetworkConnection, error) {
	if !s.config.EnableConnectionTracking {
		return nil, fmt.Errorf("connection tracking is disabled")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	connections := make([]*NetworkConnection, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}

	// Sort by bytes sent + received (most active first)
	sort.Slice(connections, func(i, j int) bool {
		totalI := connections[i].BytesSent + connections[i].BytesRecv
		totalJ := connections[j].BytesSent + connections[j].BytesRecv
		return totalI > totalJ
	})

	// Limit to max connections
	if len(connections) > s.config.MaxConnections {
		connections = connections[:s.config.MaxConnections]
	}

	return connections, nil
}

// GetBandwidthData returns bandwidth usage data for an interface
func (s *NetworkService) GetBandwidthData(interfaceName string) (*BandwidthData, error) {
	if !s.config.EnableBandwidthMonitoring {
		return nil, fmt.Errorf("bandwidth monitoring is disabled")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if data, exists := s.bandwidthData[interfaceName]; exists {
		return data, nil
	}

	return nil, fmt.Errorf("no bandwidth data available for interface %s", interfaceName)
}

// GetStatistics returns comprehensive network statistics
func (s *NetworkService) GetStatistics(ctx context.Context) (*NetworkStatistics, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := &NetworkStatistics{
		Timestamp: time.Now(),
	}

	// Get current metrics
	metrics, err := s.GetNetworkMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network metrics: %w", err)
	}

	stats.TotalBytesSent = metrics.TotalBytesSent
	stats.TotalBytesRecv = metrics.TotalBytesRecv
	stats.TotalPacketsSent = 0
	stats.TotalPacketsRecv = 0

	for _, iface := range metrics.Interfaces {
		stats.TotalPacketsSent += iface.PacketsSent
		stats.TotalPacketsRecv += iface.PacketsRecv
	}

	// Connection statistics
	stats.TotalConnections = len(s.connections)
	stats.ActiveConnections = 0
	for _, conn := range s.connections {
		if conn.State == "ESTABLISHED" {
			stats.ActiveConnections++
		}
	}

	// Top connections by bandwidth usage
	stats.TopConnections = make([]*NetworkConnection, 0)
	for _, conn := range s.connections {
		stats.TopConnections = append(stats.TopConnections, conn)
	}
	sort.Slice(stats.TopConnections, func(i, j int) bool {
		totalI := stats.TopConnections[i].BytesSent + stats.TopConnections[i].BytesRecv
		totalJ := stats.TopConnections[j].BytesSent + stats.TopConnections[j].BytesRecv
		return totalI > totalJ
	})
	if len(stats.TopConnections) > 10 {
		stats.TopConnections = stats.TopConnections[:10]
	}

	// Interface statistics
	stats.InterfaceStats = make(map[string]interface{})
	for _, data := range s.bandwidthData {
		if data.CurrentRate != nil {
			stats.InterfaceStats[data.InterfaceName] = map[string]interface{}{
				"uploadMbps":   data.CurrentRate.UploadMbps,
				"downloadMbps": data.CurrentRate.DownloadMbps,
				"peakUpload":   data.PeakRate.UploadMbps,
				"peakDownload": data.PeakRate.DownloadMbps,
			}
		}
	}

	// Emit event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(ctx, "network.statistics.updated", map[string]interface{}{
			"statistics": stats,
			"timestamp":  time.Now(),
		}, "network-service")
	}

	return stats, nil
}

// GetAlerts returns current network alerts
func (s *NetworkService) GetAlerts() []NetworkAlert {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	alerts := make([]NetworkAlert, len(s.alerts))
	copy(alerts, s.alerts)
	return alerts
}

// StartMonitoring starts continuous network monitoring.
//
// The polling loop runs on its own goroutine so this returns immediately,
// matching the system and audio services. It previously ran the ticker loop
// synchronously and never returned; under Wails v2 that was masked because
// the call sat at the tail of OnStartup, after the window already existed,
// with its error only logged as a warning. Under Wails v3 the equivalent
// call happens in ServiceStartup, which gates window creation — so a
// blocking implementation here stops the app from ever showing a window.
//
// The loop owns ctx: cancelling it (coordinator Shutdown cancels the
// derived service context) stops the ticker and clears isMonitoring.
func (s *NetworkService) StartMonitoring(ctx context.Context) error {
	// Idempotent: the coordinator starts monitoring during ServiceStartup,
	// and the frontend's network store also asks for it when it initialises.
	// "Already running" is the desired end state, not a failure, so report
	// success rather than surfacing a spurious error to the caller.
	if s.isMonitoring {
		return nil
	}

	s.isMonitoring = true
	ticker := time.NewTicker(s.config.RefreshInterval)

	go func() {
		defer ticker.Stop()
		defer func() { s.isMonitoring = false }()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.monitor(ctx); err != nil {
					// Log error but continue monitoring
					fmt.Printf("Error during network monitoring: %v\n", err)
				}
			}
		}
	}()

	return nil
}

// StopMonitoring stops network monitoring
func (s *NetworkService) StopMonitoring() error {
	if !s.isMonitoring {
		return fmt.Errorf("monitoring is not started")
	}

	s.isMonitoring = false
	return nil
}

// monitor performs one monitoring cycle
func (s *NetworkService) monitor(ctx context.Context) error {
	// Update network metrics
	if _, err := s.GetNetworkMetrics(ctx); err != nil {
		return fmt.Errorf("failed to update network metrics: %w", err)
	}

	// Update connections if enabled
	if s.config.EnableConnectionTracking {
		if err := s.updateConnections(ctx); err != nil {
			return fmt.Errorf("failed to update connections: %w", err)
		}
	}

	// Check alerts if enabled
	if s.config.EnableAlerts {
		s.checkAlerts()
	}

	return nil
}

// updateBandwidthData updates bandwidth tracking data for an interface
func (s *NetworkService) updateBandwidthData(iface *models.NetworkInterface) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, exists := s.bandwidthData[iface.Name]
	if !exists {
		data = &BandwidthData{
			InterfaceName: iface.Name,
			Timestamps:    make([]time.Time, 0),
			BytesSent:     make([]uint64, 0),
			BytesRecv:     make([]uint64, 0),
			PacketsSent:   make([]uint64, 0),
			PacketsRecv:   make([]uint64, 0),
		}
		s.bandwidthData[iface.Name] = data
	}

	// Add current data
	now := time.Now()
	data.Timestamps = append(data.Timestamps, now)
	data.BytesSent = append(data.BytesSent, iface.BytesSent)
	data.BytesRecv = append(data.BytesRecv, iface.BytesRecv)
	data.PacketsSent = append(data.PacketsSent, iface.PacketsSent)
	data.PacketsRecv = append(data.PacketsRecv, iface.PacketsRecv)

	// Maintain history size
	if len(data.Timestamps) > s.config.BandwidthHistorySize {
		data.Timestamps = data.Timestamps[1:]
		data.BytesSent = data.BytesSent[1:]
		data.BytesRecv = data.BytesRecv[1:]
		data.PacketsSent = data.PacketsSent[1:]
		data.PacketsRecv = data.PacketsRecv[1:]
	}

	// Calculate rates if we have at least 2 data points
	if len(data.Timestamps) >= 2 {
		s.calculateBandwidthRates(data)
	}
}

// calculateBandwidthRates calculates current and average bandwidth rates
func (s *NetworkService) calculateBandwidthRates(data *BandwidthData) {
	if len(data.Timestamps) < 2 {
		return
	}

	// Get the last two data points
	n := len(data.Timestamps)
	timeDelta := data.Timestamps[n-1].Sub(data.Timestamps[n-2]).Seconds()

	if timeDelta <= 0 {
		return
	}

	// Calculate current rate
	bytesSentDelta := int64(data.BytesSent[n-1] - data.BytesSent[n-2])
	bytesRecvDelta := int64(data.BytesRecv[n-1] - data.BytesRecv[n-2])

	currentUploadBps := float64(bytesSentDelta) / timeDelta
	currentDownloadBps := float64(bytesRecvDelta) / timeDelta

	data.CurrentRate = &BandwidthRate{
		UploadBps:    currentUploadBps,
		DownloadBps:  currentDownloadBps,
		UploadMbps:   currentUploadBps * 8 / 1024 / 1024,
		DownloadMbps: currentDownloadBps * 8 / 1024 / 1024,
	}

	// Update peak rates
	if data.PeakRate == nil || currentUploadBps > data.PeakRate.UploadBps {
		if data.PeakRate == nil {
			data.PeakRate = &BandwidthRate{}
		}
		data.PeakRate.UploadBps = currentUploadBps
		data.PeakRate.UploadMbps = currentUploadBps * 8 / 1024 / 1024
	}

	if data.PeakRate == nil || currentDownloadBps > data.PeakRate.DownloadBps {
		if data.PeakRate == nil {
			data.PeakRate = &BandwidthRate{}
		}
		data.PeakRate.DownloadBps = currentDownloadBps
		data.PeakRate.DownloadMbps = currentDownloadBps * 8 / 1024 / 1024
	}

	// Calculate average rate over the entire history
	if len(data.Timestamps) >= 2 {
		totalTime := data.Timestamps[n-1].Sub(data.Timestamps[0]).Seconds()
		if totalTime > 0 {
			avgUploadBps := float64(data.BytesSent[n-1]-data.BytesSent[0]) / totalTime
			avgDownloadBps := float64(data.BytesRecv[n-1]-data.BytesRecv[0]) / totalTime

			data.AverageRate = &BandwidthRate{
				UploadBps:    avgUploadBps,
				DownloadBps:  avgDownloadBps,
				UploadMbps:   avgUploadBps * 8 / 1024 / 1024,
				DownloadMbps: avgDownloadBps * 8 / 1024 / 1024,
			}
		}
	}
}

// updateConnections updates the network connections list
func (s *NetworkService) updateConnections(ctx context.Context) error {
	// This is a placeholder - in a real implementation, you would use
	// platform-specific APIs to get actual connection information
	// For now, we'll maintain a simple connection tracking system

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clean up old connections (older than 5 minutes)
	cutoff := time.Now().Add(-5 * time.Minute)
	for key, conn := range s.connections {
		if conn.LastActivity.Before(cutoff) {
			delete(s.connections, key)
		}
	}

	return nil
}

// checkAlerts checks for network alert conditions
func (s *NetworkService) checkAlerts() {
	if !s.config.EnableAlerts || s.config.AlertThresholds == nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check bandwidth alerts
	for interfaceName, data := range s.bandwidthData {
		if data.CurrentRate != nil {
			uploadMBps := data.CurrentRate.UploadMbps
			downloadMBps := data.CurrentRate.DownloadMbps
			totalMBps := uploadMBps + downloadMBps

			if totalMBps > s.config.AlertThresholds.BandwidthUsageMBps {
				s.createAlert("bandwidth_high", "High bandwidth usage", "warning",
					fmt.Sprintf("Interface %s is using %.2f MB/s (threshold: %.2f MB/s)",
						interfaceName, totalMBps, s.config.AlertThresholds.BandwidthUsageMBps),
					interfaceName, totalMBps, s.config.AlertThresholds.BandwidthUsageMBps)
			}
		}
	}

	// Check connection count alerts
	if len(s.connections) > s.config.AlertThresholds.ConnectionCount {
		s.createAlert("connections_high", "High connection count", "warning",
			fmt.Sprintf("System has %d active connections (threshold: %d)",
				len(s.connections), s.config.AlertThresholds.ConnectionCount),
			"", float64(len(s.connections)), float64(s.config.AlertThresholds.ConnectionCount))
	}
}

// createAlert creates a new network alert
func (s *NetworkService) createAlert(alertType, message, severity, details, interfaceName string, current, threshold float64) {
	alert := NetworkAlert{
		ID:        fmt.Sprintf("%s_%d", alertType, time.Now().Unix()),
		Type:      alertType,
		Severity:  severity,
		Message:   details,
		Interface: interfaceName,
		Threshold: threshold,
		Current:   current,
		Timestamp: time.Now(),
		Resolved:  false,
		Metadata: map[string]interface{}{
			"message": message,
		},
	}

	s.alerts = append(s.alerts, alert)

	// Emit alert event if event bus is available
	if s.eventBus != nil {
		s.eventBus.Publish(context.Background(), events.SystemAlert, map[string]interface{}{
			"alert":     alert,
			"type":      "network",
			"timestamp": time.Now(),
		}, "network-service")
	}

	// Keep only last 100 alerts
	if len(s.alerts) > 100 {
		s.alerts = s.alerts[1:]
	}
}

// ResolveAlert marks an alert as resolved
func (s *NetworkService) ResolveAlert(alertID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i := range s.alerts {
		if s.alerts[i].ID == alertID {
			s.alerts[i].Resolved = true
			now := time.Now()
			s.alerts[i].ResolvedAt = &now

			// Emit resolved event if event bus is available
			if s.eventBus != nil {
				s.eventBus.Publish(context.Background(), "network.alert.resolved", map[string]interface{}{
					"alert":     s.alerts[i],
					"timestamp": now,
				}, "network-service")
			}

			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// ClearAlerts removes all resolved alerts
func (s *NetworkService) ClearAlerts() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var activeAlerts []NetworkAlert
	for _, alert := range s.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	s.alerts = activeAlerts
}

// GetConfiguration returns the current configuration
func (s *NetworkService) GetConfiguration() *NetworkServiceConfig {
	return s.config
}

// UpdateConfiguration updates the service configuration
func (s *NetworkService) UpdateConfiguration(config *NetworkServiceConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if config.RefreshInterval <= 0 {
		return fmt.Errorf("refresh interval must be positive")
	}

	if config.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}

	s.config = config
	return nil
}

// Reset resets all network monitoring data
func (s *NetworkService) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.connections = make(map[string]*NetworkConnection)
	s.bandwidthData = make(map[string]*BandwidthData)
	s.alerts = make([]NetworkAlert, 0)
	s.lastUpdate = time.Now()
}

// IsMonitoring returns whether monitoring is currently active
func (s *NetworkService) IsMonitoring() bool {
	return s.isMonitoring
}

// GetLastUpdateTime returns the timestamp of the last update
func (s *NetworkService) GetLastUpdateTime() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastUpdate
}

// GeoIP methods

// GetGeoIPService returns the GeoIP service instance
func (s *NetworkService) GetGeoIPService() *geoip.GeoIPService {
	return s.geoipService
}

// SetGeoIPConfig sets the GeoIP service configuration
func (s *NetworkService) SetGeoIPConfig(config *geoip.GeoIPConfig) error {
	if s.geoipService == nil {
		return fmt.Errorf("GeoIP service not initialized")
	}
	return s.geoipService.SetConfig(config)
}

// GetGeoIPConfig returns the current GeoIP configuration
func (s *NetworkService) GetGeoIPConfig() *geoip.GeoIPConfig {
	if s.geoipService == nil {
		return nil
	}
	return s.geoipService.GetConfig()
}

// EnableGeoIP enables the GeoIP service
func (s *NetworkService) EnableGeoIP() error {
	if s.geoipService == nil {
		return fmt.Errorf("GeoIP service not initialized")
	}
	return s.geoipService.Enable()
}

// DisableGeoIP disables the GeoIP service and clears cache
func (s *NetworkService) DisableGeoIP() error {
	if s.geoipService == nil {
		return fmt.Errorf("GeoIP service not initialized")
	}
	return s.geoipService.Disable()
}

// IsGeoIPEnabled returns whether GeoIP service is enabled
func (s *NetworkService) IsGeoIPEnabled() bool {
	if s.geoipService == nil {
		return false
	}
	return s.geoipService.IsEnabled()
}

// GetConnectionsWithGeoIP returns network connections with GeoIP information
func (s *NetworkService) GetConnectionsWithGeoIP(ctx context.Context) ([]*models.NetworkConnectionWithGeo, error) {
	connections, err := s.GetConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	if !s.IsGeoIPEnabled() {
		// Return connections without GeoIP data
		result := make([]*models.NetworkConnectionWithGeo, len(connections))
		for i, conn := range connections {
			result[i] = &models.NetworkConnectionWithGeo{
				LocalAddr:    conn.LocalAddr,
				RemoteAddr:   conn.RemoteAddr,
				State:        conn.State,
				PID:          conn.PID,
				ProcessName:  conn.ProcessName,
				Protocol:     conn.Protocol,
				BytesSent:    conn.BytesSent,
				BytesRecv:    conn.BytesRecv,
				Established:  conn.Established,
				LastActivity: conn.LastActivity,
			}
		}
		return result, nil
	}

	result := make([]*models.NetworkConnectionWithGeo, len(connections))
	for i, conn := range connections {
		connWithGeo := &models.NetworkConnectionWithGeo{
			LocalAddr:    conn.LocalAddr,
			RemoteAddr:   conn.RemoteAddr,
			State:        conn.State,
			PID:          conn.PID,
			ProcessName:  conn.ProcessName,
			Protocol:     conn.Protocol,
			BytesSent:    conn.BytesSent,
			BytesRecv:    conn.BytesRecv,
			Established:  conn.Established,
			LastActivity: conn.LastActivity,
		}

		// Extract IP from remote address
		ip := s.extractIPFromAddr(conn.RemoteAddr)
		if ip != "" {
			// Try to get GeoIP data (use cached data if available)
			if geoipData, err := s.geoipService.GetCachedGeoIPData(ip); err == nil {
				connWithGeo.RemoteGeoIP = &models.GeoIPInfo{
					IP:           geoipData.IP,
					Country:      geoipData.Country,
					CountryCode:  geoipData.CountryCode,
					Region:       geoipData.Region,
					RegionCode:   geoipData.RegionCode,
					City:         geoipData.City,
					Latitude:     geoipData.Latitude,
					Longitude:    geoipData.Longitude,
					ISP:          geoipData.ISP,
					Organization: geoipData.Organization,
					AS:           geoipData.AS,
					Timezone:     geoipData.Timezone,
					IsVPN:        geoipData.IsVPN,
					IsProxy:      geoipData.IsProxy,
					IsMobile:     geoipData.IsMobile,
					Timestamp:    geoipData.Timestamp,
				}
				connWithGeo.CountryFlag = connWithGeo.RemoteGeoIP.GetCountryFlag()
			} else {
				// Try to fetch fresh data (non-blocking)
				go func() {
					if freshData, err := s.geoipService.GetGeoIPData(context.Background(), ip); err == nil {
						// Cache is updated automatically by the GeoIP service
						// Emit event to notify frontend
						if s.eventBus != nil {
							s.eventBus.Publish(context.Background(), "network:geoip-updated", map[string]interface{}{
								"connection": conn,
								"geoip":      freshData,
								"timestamp":  time.Now(),
							}, "network-service")
						}
					}
				}()
			}
		}

		result[i] = connWithGeo
	}

	return result, nil
}

// GetGeoIPDataForIP returns GeoIP data for a specific IP address
func (s *NetworkService) GetGeoIPDataForIP(ctx context.Context, ip string) (*models.GeoIPInfo, error) {
	if !s.IsGeoIPEnabled() {
		return nil, fmt.Errorf("GeoIP service is disabled")
	}

	geoipData, err := s.geoipService.GetGeoIPData(ctx, ip)
	if err != nil {
		return nil, err
	}

	return &models.GeoIPInfo{
		IP:           geoipData.IP,
		Country:      geoipData.Country,
		CountryCode:  geoipData.CountryCode,
		Region:       geoipData.Region,
		RegionCode:   geoipData.RegionCode,
		City:         geoipData.City,
		Latitude:     geoipData.Latitude,
		Longitude:    geoipData.Longitude,
		ISP:          geoipData.ISP,
		Organization: geoipData.Organization,
		AS:           geoipData.AS,
		Timezone:     geoipData.Timezone,
		IsVPN:        geoipData.IsVPN,
		IsProxy:      geoipData.IsProxy,
		IsMobile:     geoipData.IsMobile,
		Timestamp:    geoipData.Timestamp,
	}, nil
}

// GetGeoIPCacheStats returns GeoIP cache statistics
func (s *NetworkService) GetGeoIPCacheStats() *models.GeoIPCacheStats {
	if s.geoipService == nil {
		return &models.GeoIPCacheStats{
			Enabled: false,
		}
	}

	stats := s.geoipService.GetCacheStats()
	return &models.GeoIPCacheStats{
		Size:       stats["size"].(int),
		MaxSize:    stats["maxSize"].(int),
		LastUpdate: stats["lastUpdate"].(time.Time),
		Enabled:    stats["enabled"].(bool),
	}
}

// ClearGeoIPCache clears the GeoIP cache
func (s *NetworkService) ClearGeoIPCache() error {
	if s.geoipService == nil {
		return fmt.Errorf("GeoIP service not initialized")
	}
	return s.geoipService.ClearCache()
}

// extractIPFromAddr extracts IP address from network address string
func (s *NetworkService) extractIPFromAddr(addr string) string {
	// Handle IPv6 addresses in brackets
	if len(addr) > 0 && addr[0] == '[' {
		if end := strings.Index(addr, "]"); end != -1 {
			addr = addr[1:end]
		}
	}

	// Split on colon to separate IP from port
	if colonIndex := strings.LastIndex(addr, ":"); colonIndex != -1 {
		addr = addr[:colonIndex]
	}

	// Validate IP
	if net.ParseIP(addr) != nil {
		return addr
	}

	return ""
}
