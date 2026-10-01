package network

import (
	"context"
	"fmt"
	"time"

	"aDex/internal/events"
)

// Network-specific event types
const (
	// Network Connection Events
	NetworkConnectionEstablished = "network.connection.established"
	NetworkConnectionClosed      = "network.connection.closed"
	NetworkConnectionError       = "network.connection.error"

	// Network Interface Events
	NetworkInterfaceUp           = "network.interface.up"
	NetworkInterfaceDown         = "network.interface.down"
	NetworkInterfaceChanged      = "network.interface.changed"

	// Network Monitoring Events
	NetworkMonitoringStarted     = "network.monitoring.started"
	NetworkMonitoringStopped     = "network.monitoring.stopped"
	NetworkStatisticsUpdated     = "network.statistics.updated"
	NetworkAlertTriggered        = "network.alert.triggered"
	NetworkAlertResolved         = "network.alert.resolved"

	// Bandwidth Events
	BandwidthThresholdExceeded   = "network.bandwidth.threshold.exceeded"
	BandwidthUsageNormal         = "network.bandwidth.usage.normal"
	BandwidthPeakDetected        = "network.bandwidth.peak.detected"

	// Connectivity Events
	NetworkConnectivityLost      = "network.connectivity.lost"
	NetworkConnectivityRestored  = "network.connectivity.restored"
	NetworkLatencyHigh           = "network.latency.high"
	NetworkLatencyNormal         = "network.latency.normal"
)

// NetworkEventData represents common network event data
type NetworkEventData struct {
	Interface   string                 `json:"interface"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Severity    string                 `json:"severity,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectionEventData represents connection-specific event data
type ConnectionEventData struct {
	NetworkEventData
	ConnectionID string `json:"connectionId"`
	LocalAddr    string `json:"localAddr"`
	RemoteAddr   string `json:"remoteAddr"`
	Protocol     string `json:"protocol"`
	PID          int    `json:"pid"`
	ProcessName  string `json:"processName"`
}

// InterfaceEventData represents interface-specific event data
type InterfaceEventData struct {
	NetworkEventData
	OldState string `json:"oldState,omitempty"`
	NewState string `json:"newState"`
	Reason   string `json:"reason,omitempty"`
}

// BandwidthEventData represents bandwidth-specific event data
type BandwidthEventData struct {
	NetworkEventData
	InterfaceName   string  `json:"interfaceName"`
	CurrentUpload   float64 `json:"currentUpload"`   // Mbps
	CurrentDownload float64 `json:"currentDownload"` // Mbps
	Threshold       float64 `json:"threshold"`       // Mbps
	AverageUpload   float64 `json:"averageUpload"`   // Mbps
	AverageDownload float64 `json:"averageDownload"` // Mbps
}

// AlertEventData represents alert-specific event data
type AlertEventData struct {
	NetworkEventData
	AlertID     string  `json:"alertId"`
	AlertType   string  `json:"alertType"`
	Message     string  `json:"message"`
	Threshold   float64 `json:"threshold"`
	CurrentValue float64 `json:"currentValue"`
	Resolved    bool    `json:"resolved"`
}

// ConnectivityEventData represents connectivity-specific event data
type ConnectivityEventData struct {
	NetworkEventData
	IsConnected   bool          `json:"isConnected"`
	Latency       time.Duration `json:"latency"`
	GatewayIP     string        `json:"gatewayIP,omitempty"`
	DNSServers    []string      `json:"dnsServers,omitempty"`
}

// EventPublisher provides methods for publishing network events
type EventPublisher struct {
	eventBus events.IEventBus
	source   string
}

// NewEventPublisher creates a new network event publisher
func NewEventPublisher(eventBus events.IEventBus, source string) *EventPublisher {
	return &EventPublisher{
		eventBus: eventBus,
		source:   source,
	}
}

// PublishConnectionEstablished publishes a connection established event
func (p *EventPublisher) PublishConnectionEstablished(ctx context.Context, conn *NetworkConnection) error {
	if p.eventBus == nil {
		return nil
	}

	data := ConnectionEventData{
		NetworkEventData: NetworkEventData{
			Interface: extractInterfaceFromAddress(conn.LocalAddr),
			Timestamp: time.Now(),
			Source:    p.source,
		},
		ConnectionID: generateConnectionID(conn),
		LocalAddr:    conn.LocalAddr,
		RemoteAddr:   conn.RemoteAddr,
		Protocol:     conn.Protocol,
		PID:          conn.PID,
		ProcessName:  conn.ProcessName,
	}

	return p.eventBus.Publish(ctx, NetworkConnectionEstablished, data, p.source)
}

// PublishConnectionClosed publishes a connection closed event
func (p *EventPublisher) PublishConnectionClosed(ctx context.Context, conn *NetworkConnection) error {
	if p.eventBus == nil {
		return nil
	}

	data := ConnectionEventData{
		NetworkEventData: NetworkEventData{
			Interface: extractInterfaceFromAddress(conn.LocalAddr),
			Timestamp: time.Now(),
			Source:    p.source,
		},
		ConnectionID: generateConnectionID(conn),
		LocalAddr:    conn.LocalAddr,
		RemoteAddr:   conn.RemoteAddr,
		Protocol:     conn.Protocol,
		PID:          conn.PID,
		ProcessName:  conn.ProcessName,
	}

	return p.eventBus.Publish(ctx, NetworkConnectionClosed, data, p.source)
}

// PublishInterfaceStateChanged publishes an interface state change event
func (p *EventPublisher) PublishInterfaceStateChanged(ctx context.Context, interfaceName, oldState, newState, reason string) error {
	if p.eventBus == nil {
		return nil
	}

	data := InterfaceEventData{
		NetworkEventData: NetworkEventData{
			Interface: interfaceName,
			Timestamp: time.Now(),
			Source:    p.source,
		},
		OldState: oldState,
		NewState: newState,
		Reason:   reason,
	}

	eventType := NetworkInterfaceChanged
	if newState == "up" {
		eventType = NetworkInterfaceUp
	} else if newState == "down" {
		eventType = NetworkInterfaceDown
	}

	return p.eventBus.Publish(ctx, eventType, data, p.source)
}

// PublishBandwidthAlert publishes a bandwidth alert event
func (p *EventPublisher) PublishBandwidthAlert(ctx context.Context, alert *NetworkAlert, bandwidthData *BandwidthData) error {
	if p.eventBus == nil {
		return nil
	}

	data := AlertEventData{
		NetworkEventData: NetworkEventData{
			Interface: alert.Interface,
			Timestamp: alert.Timestamp,
			Source:    p.source,
			Severity:  alert.Severity,
		},
		AlertID:      alert.ID,
		AlertType:    alert.Type,
		Message:      alert.Message,
		Threshold:    alert.Threshold,
		CurrentValue: alert.Current,
		Resolved:     alert.Resolved,
	}

	eventType := NetworkAlertTriggered
	if alert.Resolved {
		eventType = NetworkAlertResolved
	}

	return p.eventBus.Publish(ctx, eventType, data, p.source)
}

// PublishBandwidthThresholdExceeded publishes a bandwidth threshold exceeded event
func (p *EventPublisher) PublishBandwidthThresholdExceeded(ctx context.Context, interfaceName string, current, threshold, avgUpload, avgDownload float64) error {
	if p.eventBus == nil {
		return nil
	}

	data := BandwidthEventData{
		NetworkEventData: NetworkEventData{
			Interface: interfaceName,
			Timestamp: time.Now(),
			Source:    p.source,
			Severity:  "warning",
		},
		InterfaceName:   interfaceName,
		CurrentUpload:   avgUpload,
		CurrentDownload: avgDownload,
		Threshold:       threshold,
		AverageUpload:   avgUpload,
		AverageDownload: avgDownload,
	}

	return p.eventBus.Publish(ctx, BandwidthThresholdExceeded, data, p.source)
}

// PublishConnectivityChanged publishes a connectivity change event
func (p *EventPublisher) PublishConnectivityChanged(ctx context.Context, isConnected bool, latency time.Duration, gatewayIP string, dnsServers []string) error {
	if p.eventBus == nil {
		return nil
	}

	data := ConnectivityEventData{
		NetworkEventData: NetworkEventData{
			Timestamp: time.Now(),
			Source:    p.source,
			Severity:  "info",
		},
		IsConnected: isConnected,
		Latency:     latency,
		GatewayIP:   gatewayIP,
		DNSServers:  dnsServers,
	}

	eventType := NetworkConnectivityRestored
	if !isConnected {
		eventType = NetworkConnectivityLost
		data.Severity = "error"
	}

	return p.eventBus.Publish(ctx, eventType, data, p.source)
}

// PublishLatencyAlert publishes a latency alert event
func (p *EventPublisher) PublishLatencyAlert(ctx context.Context, interfaceName string, latency time.Duration, threshold time.Duration) error {
	if p.eventBus == nil {
		return nil
	}

	isHigh := latency > threshold
	eventType := NetworkLatencyNormal
	severity := "info"
	if isHigh {
		eventType = NetworkLatencyHigh
		severity = "warning"
	}

	data := NetworkEventData{
		Interface: interfaceName,
		Timestamp: time.Now(),
		Source:    p.source,
		Severity:  severity,
		Metadata: map[string]interface{}{
			"latency_ms":    latency.Milliseconds(),
			"threshold_ms":  threshold.Milliseconds(),
			"is_high":       isHigh,
		},
	}

	return p.eventBus.Publish(ctx, eventType, data, p.source)
}

// PublishMonitoringStarted publishes a monitoring started event
func (p *EventPublisher) PublishMonitoringStarted(ctx context.Context) error {
	if p.eventBus == nil {
		return nil
	}

	data := NetworkEventData{
		Timestamp: time.Now(),
		Source:    p.source,
		Severity:  "info",
		Metadata: map[string]interface{}{
			"service": "network",
		},
	}

	return p.eventBus.Publish(ctx, NetworkMonitoringStarted, data, p.source)
}

// PublishMonitoringStopped publishes a monitoring stopped event
func (p *EventPublisher) PublishMonitoringStopped(ctx context.Context) error {
	if p.eventBus == nil {
		return nil
	}

	data := NetworkEventData{
		Timestamp: time.Now(),
		Source:    p.source,
		Severity:  "info",
		Metadata: map[string]interface{}{
			"service": "network",
		},
	}

	return p.eventBus.Publish(ctx, NetworkMonitoringStopped, data, p.source)
}

// PublishStatisticsUpdated publishes a statistics updated event
func (p *EventPublisher) PublishStatisticsUpdated(ctx context.Context, stats *NetworkStatistics) error {
	if p.eventBus == nil {
		return nil
	}

	data := NetworkEventData{
		Timestamp: time.Now(),
		Source:    p.source,
		Severity:  "info",
		Metadata: map[string]interface{}{
			"statistics": stats,
		},
	}

	return p.eventBus.Publish(ctx, NetworkStatisticsUpdated, data, p.source)
}

// Helper functions

// generateConnectionID generates a unique connection ID
func generateConnectionID(conn *NetworkConnection) string {
	return fmt.Sprintf("%s_%s_%s_%d", 
		conn.Protocol, conn.LocalAddr, conn.RemoteAddr, conn.PID)
}

// extractInterfaceFromAddress extracts interface name from address
func extractInterfaceFromAddress(addr string) string {
	// This is a simplified implementation
	// In a real implementation, you would parse the address
	// and determine which interface it belongs to
	return "unknown"
}

// GetNetworkEventContracts returns all network event contracts
func GetNetworkEventContracts() map[string]*events.EventContract {
	return map[string]*events.EventContract{
		NetworkConnectionEstablished: {
			Type:    NetworkConnectionEstablished,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"connectionId": "string",
				"localAddr":    "string",
				"remoteAddr":   "string",
				"protocol":     "string",
				"pid":          "int",
				"processName":  "string",
			},
			Validator: validateConnectionEventData,
		},

		NetworkConnectionClosed: {
			Type:    NetworkConnectionClosed,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"connectionId": "string",
				"localAddr":    "string",
				"remoteAddr":   "string",
				"protocol":     "string",
			},
			Validator: validateConnectionEventData,
		},

		NetworkInterfaceUp: {
			Type:    NetworkInterfaceUp,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"interface": "string",
				"newState":  "string",
				"reason":    "string",
			},
			Validator: validateInterfaceEventData,
		},

		NetworkInterfaceDown: {
			Type:    NetworkInterfaceDown,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"interface": "string",
				"newState":  "string",
				"reason":    "string",
			},
			Validator: validateInterfaceEventData,
		},

		BandwidthThresholdExceeded: {
			Type:    BandwidthThresholdExceeded,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"interfaceName":   "string",
				"currentUpload":   "float64",
				"currentDownload": "float64",
				"threshold":       "float64",
			},
			Validator: validateBandwidthEventData,
		},

		NetworkConnectivityLost: {
			Type:    NetworkConnectivityLost,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"isConnected": "bool",
				"latency":     "duration",
				"gatewayIP":   "string",
			},
			Validator: validateConnectivityEventData,
		},

		NetworkConnectivityRestored: {
			Type:    NetworkConnectivityRestored,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"isConnected": "bool",
				"latency":     "duration",
				"gatewayIP":   "string",
			},
			Validator: validateConnectivityEventData,
		},

		NetworkAlertTriggered: {
			Type:    NetworkAlertTriggered,
			Version: "1.0.0",
			Schema: map[string]interface{}{
				"alertId":       "string",
				"alertType":     "string",
				"message":       "string",
				"threshold":     "float64",
				"currentValue":  "float64",
				"resolved":      "bool",
			},
			Validator: validateAlertEventData,
		},
	}
}

// Event validation functions

func validateConnectionEventData(event events.Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateInterfaceEventData(event events.Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateBandwidthEventData(event events.Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateConnectivityEventData(event events.Event) error {
	// Implementation would validate the structure and values
	return nil
}

func validateAlertEventData(event events.Event) error {
	// Implementation would validate the structure and values
	return nil
}