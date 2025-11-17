package network

import (
	"time"
)

// Extended network models for comprehensive monitoring

// NetworkProtocol represents different network protocols
type NetworkProtocol string

const (
	ProtocolTCP  NetworkProtocol = "tcp"
	ProtocolUDP  NetworkProtocol = "udp"
	ProtocolICMP NetworkProtocol = "icmp"
	ProtocolIPv4 NetworkProtocol = "ipv4"
	ProtocolIPv6 NetworkProtocol = "ipv6"
)

// ConnectionState represents the state of a network connection
type ConnectionState string

const (
	StateEstablished ConnectionState = "ESTABLISHED"
	StateSynSent    ConnectionState = "SYN_SENT"
	StateSynRecv    ConnectionState = "SYN_RECV"
	StateFinWait1   ConnectionState = "FIN_WAIT_1"
	StateFinWait2   ConnectionState = "FIN_WAIT_2"
	StateTimeWait   ConnectionState = "TIME_WAIT"
	StateClose      ConnectionState = "CLOSE"
	StateCloseWait  ConnectionState = "CLOSE_WAIT"
	StateLastAck    ConnectionState = "LAST_ACK"
	StateListen     ConnectionState = "LISTEN"
	StateClosing    ConnectionState = "CLOSING"
	StateUnknown    ConnectionState = "UNKNOWN"
)

// NetworkInterfaceType represents the type of network interface
type NetworkInterfaceType string

const (
	InterfaceTypeEthernet   NetworkInterfaceType = "ethernet"
	InterfaceTypeWiFi       NetworkInterfaceType = "wifi"
	InterfaceTypeLoopback   NetworkInterfaceType = "loopback"
	InterfaceTypeTunnel     NetworkInterfaceType = "tunnel"
	InterfaceTypeBridge     NetworkInterfaceType = "bridge"
	InterfaceTypeVirtual    NetworkInterfaceType = "virtual"
	InterfaceTypeDocker     NetworkInterfaceType = "docker"
	InterfaceTypeUnknown    NetworkInterfaceType = "unknown"
)

// ConnectionStats represents detailed connection statistics
type ConnectionStats struct {
	Protocol         NetworkProtocol `json:"protocol"`
	LocalAddress     string          `json:"localAddress"`
	LocalPort        int             `json:"localPort"`
	RemoteAddress    string          `json:"remoteAddress"`
	RemotePort       int             `json:"remotePort"`
	State            ConnectionState `json:"state"`
	PID              int             `json:"pid"`
	ProcessName      string          `json:"processName"`
	ProcessPath      string          `json:"processPath"`
	BytesSent        uint64          `json:"bytesSent"`
	BytesReceived    uint64          `json:"bytesReceived"`
	PacketsSent      uint64          `json:"packetsSent"`
	PacketsReceived  uint64          `json:"packetsReceived"`
	ConnectionTime   time.Time       `json:"connectionTime"`
	LastActivity     time.Time       `json:"lastActivity"`
	Duration         time.Duration   `json:"duration"`
}

// InterfaceDetails represents detailed interface information
type InterfaceDetails struct {
	Name             string                `json:"name"`
	Type             NetworkInterfaceType  `json:"type"`
	IsUp             bool                  `json:"isUp"`
	IsPromiscuous    bool                  `json:"isPromiscuous"`
	MACAddress       string                `json:"macAddress"`
	MTU              uint64                `json:"mtu"`
	Speed            uint64                `json:"speed"`
	Duplex           string                `json:"duplex"`
	IPAddresses      []InterfaceIP         `json:"ipAddresses"`
	IPv6Addresses    []InterfaceIP         `json:"ipv6Addresses"`
	Gateway          string                `json:"gateway"`
	DNSServers       []string              `json:"dnsServers"`
	Statistics       InterfaceStatistics   `json:"statistics"`
	TrafficHistory   *TrafficHistory       `json:"trafficHistory,omitempty"`
}

// InterfaceIP represents an IP address with additional information
type InterfaceIP struct {
	Address      string    `json:"address"`
	Netmask      string    `json:"netmask"`
	CIDR         string    `json:"cidr"`
	IsPrimary    bool      `json:"isPrimary"`
	IsTemporary  bool      `json:"isTemporary"`
	AssignedAt   time.Time `json:"assignedAt"`
	LeaseExpires *time.Time `json:"leaseExpires,omitempty"`
}

// InterfaceStatistics represents detailed interface statistics
type InterfaceStatistics struct {
	BytesReceived    uint64        `json:"bytesReceived"`
	BytesTransmitted uint64        `json:"bytesTransmitted"`
	PacketsReceived  uint64        `json:"packetsReceived"`
	PacketsTransmitted uint64       `json:"packetsTransmitted"`
	ErrorsIn         uint64        `json:"errorsIn"`
	ErrorsOut        uint64        `json:"errorsOut"`
	DroppedIn        uint64        `json:"droppedIn"`
	DroppedOut       uint64        `json:"droppedOut"`
	Collisions       uint64        `json:"collisions"`
	ReceiveErrors    uint64        `json:"receiveErrors"`
	TransmitErrors   uint64        `json:"transmitErrors"`
	ReceiveDropped   uint64        `json:"receiveDropped"`
	TransmitDropped  uint64        `json:"transmitDropped"`
	Multicast        uint64        `json:"multicast"`
	BytesReceivedRate    float64   `json:"bytesReceivedRate"`    // bytes/second
	BytesTransmittedRate float64   `json:"bytesTransmittedRate"` // bytes/second
	PacketsReceivedRate  float64   `json:"packetsReceivedRate"`  // packets/second
	PacketsTransmittedRate float64 `json:"packetsTransmittedRate"` // packets/second
	LastUpdate       time.Time     `json:"lastUpdate"`
}

// TrafficHistory represents historical traffic data
type TrafficHistory struct {
	InterfaceName string          `json:"interfaceName"`
	Timestamps    []time.Time     `json:"timestamps"`
	BytesIn       []uint64        `json:"bytesIn"`
	BytesOut      []uint64        `json:"bytesOut"`
	PacketsIn     []uint64        `json:"packetsIn"`
	PacketsOut    []uint64        `json:"packetsOut"`
	ErrorsIn      []uint64        `json:"errorsIn"`
	ErrorsOut     []uint64        `json:"errorsOut"`
	MaxDataPoints int             `json:"maxDataPoints"`
}

// NetworkRoute represents a network route
type NetworkRoute struct {
	Destination   string    `json:"destination"`
	Gateway       string    `json:"gateway"`
	Genmask       string    `json:"genmask"`
	Flags         string    `json:"flags"`
	Metric        int       `json:"metric"`
	Ref           int       `json:"ref"`
	Use           int       `json:"use"`
	Iface         string    `json:"iface"`
	MSS           int       `json:"mss"`
	Window        int       `json:"window"`
	IRTT          int       `json:"irtt"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

// NetworkArpEntry represents an ARP table entry
type NetworkArpEntry struct {
	IPAddress   string    `json:"ipAddress"`
	MACAddress  string    `json:"macAddress"`
	Interface   string    `json:"interface"`
	Type        string    `json:"type"` // "ether", "incomplete", etc.
	Flags       string    `json:"flags"`
	Mask        string    `json:"mask,omitempty"`
	LastSeen    time.Time `json:"lastSeen"`
	IsPermanent bool      `json:"isPermanent"`
}

// NetworkDnsEntry represents a DNS cache entry
type NetworkDnsEntry struct {
	DomainName   string    `json:"domainName"`
	IPAddresses  []string  `json:"ipAddresses"`
	Type         string    `json:"type"` // A, AAAA, CNAME, etc.
	TTL          int       `json:"ttl"`
	CachedAt     time.Time `json:"cachedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// NetworkLatencyMeasurement represents a latency measurement
type NetworkLatencyMeasurement struct {
	Target       string        `json:"target"`
	Interface    string        `json:"interface"`
	Latency      time.Duration `json:"latency"`
	Jitter       time.Duration `json:"jitter"`
	PacketLoss   float64       `json:"packetLoss"`
	SuccessRate  float64       `json:"successRate"`
	ProbesSent   int           `json:"probesSent"`
	ProbesReceived int         `json:"probesReceived"`
	MinLatency   time.Duration `json:"minLatency"`
	MaxLatency   time.Duration `json:"maxLatency"`
	AvgLatency   time.Duration `json:"avgLatency"`
	MeasuredAt   time.Time     `json:"measuredAt"`
}

// NetworkThroughputMeasurement represents a throughput measurement
type NetworkThroughputMeasurement struct {
	Interface     string        `json:"interface"`
	UploadBps     float64       `json:"uploadBps"`     // bytes per second
	DownloadBps   float64       `json:"downloadBps"`   // bytes per second
	UploadMbps    float64       `json:"uploadMbps"`    // megabits per second
	DownloadMbps  float64       `json:"downloadMbps"`  // megabits per second
	TestDuration  time.Duration `json:"testDuration"`
	BytesUploaded uint64        `json:"bytesUploaded"`
	BytesDownloaded uint64      `json:"bytesDownloaded"`
	MeasuredAt    time.Time     `json:"measuredAt"`
	TestMethod    string        `json:"testMethod"` // "tcp", "udp", "http", etc.
}

// NetworkTopology represents network topology information
type NetworkTopology struct {
	LocalNetworks []NetworkSegment `json:"localNetworks"`
	Gateways      []NetworkGateway `json:"gateways"`
	DNSServers    []NetworkServer  `json:"dnsServers"`
	ExternalIP    string           `json:"externalIP"`
	ISPs          []NetworkISP     `json:"isps"`
	LastUpdated   time.Time        `json:"lastUpdated"`
}

// NetworkSegment represents a network segment/subnet
type NetworkSegment struct {
	NetworkAddress string   `json:"networkAddress"`
	Netmask        string   `json:"netmask"`
	CIDR           string   `json:"cidr"`
	Broadcast      string   `json:"broadcast"`
	Interface      string   `json:"interface"`
	Gateway        string   `json:"gateway"`
	Domain         string   `json:"domain"`
	Hosts          []string `json:"hosts"`
	IsLocal        bool     `json:"isLocal"`
}

// NetworkGateway represents a network gateway
type NetworkGateway struct {
	IPAddress string `json:"ipAddress"`
	MACAddress string `json:"macAddress"`
	Interface string `json:"interface"`
	Metric    int    `json:"metric"`
	IsDefault bool   `json:"isDefault"`
	IsActive  bool   `json:"isActive"`
	Reachable bool   `json:"reachable"`
}

// NetworkServer represents a network server (DNS, NTP, etc.)
type NetworkServer struct {
	IPAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Type      string `json:"type"` // "dns", "ntp", etc.
	IsActive  bool   `json:"isActive"`
	ResponseTime time.Duration `json:"responseTime"`
	LastChecked time.Time `json:"lastChecked"`
}

// NetworkISP represents Internet Service Provider information
type NetworkISP struct {
	Name     string   `json:"name"`
	ASNumber string   `json:"asNumber"`
	Country  string   `json:"country"`
	IPRange  string   `json:"ipRange"`
	Gateways []string `json:"gateways"`
}

// NetworkSecurity represents network security information
type NetworkSecurity struct {
	FirewallRules  []FirewallRule `json:"firewallRules"`
	BlockedIPs     []string       `json:"blockedIps"`
	AllowedIPs     []string       `json:"allowedIps"`
	ActiveScans    []SecurityScan `json:"activeScans"`
	ThreatLevel    string         `json:"threatLevel"`
	LastScanned    time.Time      `json:"lastScanned"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          string   `json:"id"`
	Action      string   `json:"action"` // "allow", "deny", "log"
	Direction   string   `json:"direction"` // "in", "out", "both"
	Protocol    string   `json:"protocol"`
	SourceIP    string   `json:"sourceIp"`
	SourcePort  int      `json:"sourcePort"`
	DestIP      string   `json:"destIp"`
	DestPort    int      `json:"destPort"`
	Interface   string   `json:"interface"`
	Priority    int      `json:"priority"`
	CreatedAt   time.Time `json:"createdAt"`
	IsActive    bool     `json:"isActive"`
	Hits        int      `json:"hits"`
	LastHit     time.Time `json:"lastHit"`
}

// SecurityScan represents a security scan
type SecurityScan struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // "port", "vulnerability", etc."
	Target      string    `json:"target"`
	Status      string    `json:"status"` // "running", "completed", "failed"`
	Progress    float64   `json:"progress"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Results     []string  `json:"results"`
	Threats     []Threat  `json:"threats"`
}

// Threat represents a security threat
type Threat struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
	Count       int       `json:"count"`
	IsBlocked   bool      `json:"isBlocked"`
}

// Helper functions for network models

// IsStateActive checks if a connection state indicates active data transfer
func (s ConnectionState) IsStateActive() bool {
	return s == StateEstablished || s == StateFinWait1 || s == StateFinWait2
}

// IsStateClosing checks if a connection state indicates connection termination
func (s ConnectionState) IsStateClosing() bool {
	return s == StateFinWait1 || s == StateFinWait2 || s == StateTimeWait || 
		   s == StateClose || s == StateCloseWait || s == StateLastAck || 
		   s == StateClosing
}

// IsStateListening checks if a connection state indicates listening for connections
func (s ConnectionState) IsStateListening() bool {
	return s == StateListen
}

// GetConnectionDuration returns the duration of a connection
func (c *ConnectionStats) GetConnectionDuration() time.Duration {
	if c.ConnectionTime.IsZero() {
		return 0
	}
	if c.LastActivity.IsZero() {
		return time.Since(c.ConnectionTime)
	}
	return c.LastActivity.Sub(c.ConnectionTime)
}

// IsExpired checks if a connection is considered expired (no activity for too long)
func (c *ConnectionStats) IsExpired(timeout time.Duration) bool {
	if c.LastActivity.IsZero() {
		return time.Since(c.ConnectionTime) > timeout
	}
	return time.Since(c.LastActivity) > timeout
}

// GetBandwidthUtilization calculates bandwidth utilization percentage
func (i *InterfaceDetails) GetBandwidthUtilization() float64 {
	if i.Speed == 0 {
		return 0
	}
	
	currentRate := i.Statistics.BytesReceivedRate + i.Statistics.BytesTransmittedRate
	utilization := (currentRate * 8) / float64(i.Speed) // Convert to bits
	
	if utilization > 1.0 {
		utilization = 1.0 // Cap at 100%
	}
	
	return utilization * 100
}

// IsHealthy checks if the interface appears healthy based on error rates
func (i *InterfaceDetails) IsHealthy() bool {
	totalPackets := i.Statistics.PacketsReceived + i.Statistics.PacketsTransmitted
	if totalPackets == 0 {
		return true // No traffic, can't determine health
	}
	
	totalErrors := i.Statistics.ErrorsIn + i.Statistics.ErrorsOut
	errorRate := float64(totalErrors) / float64(totalPackets)
	
	// Consider unhealthy if error rate exceeds 1%
	return errorRate <= 0.01
}

// AddDataPoint adds a new data point to traffic history
func (t *TrafficHistory) AddDataPoint(timestamp time.Time, bytesIn, bytesOut, packetsIn, packetsOut, errorsIn, errorsOut uint64) {
	t.Timestamps = append(t.Timestamps, timestamp)
	t.BytesIn = append(t.BytesIn, bytesIn)
	t.BytesOut = append(t.BytesOut, bytesOut)
	t.PacketsIn = append(t.PacketsIn, packetsIn)
	t.PacketsOut = append(t.PacketsOut, packetsOut)
	t.ErrorsIn = append(t.ErrorsIn, errorsIn)
	t.ErrorsOut = append(t.ErrorsOut, errorsOut)
	
	// Maintain maximum data points
	if len(t.Timestamps) > t.MaxDataPoints {
		t.Timestamps = t.Timestamps[1:]
		t.BytesIn = t.BytesIn[1:]
		t.BytesOut = t.BytesOut[1:]
		t.PacketsIn = t.PacketsIn[1:]
		t.PacketsOut = t.PacketsOut[1:]
		t.ErrorsIn = t.ErrorsIn[1:]
		t.ErrorsOut = t.ErrorsOut[1:]
	}
}

// GetAverageRates calculates average rates over the history
func (t *TrafficHistory) GetAverageRates() (avgBytesIn, avgBytesOut float64) {
	if len(t.Timestamps) < 2 {
		return 0, 0
	}
	
	duration := t.Timestamps[len(t.Timestamps)-1].Sub(t.Timestamps[0]).Seconds()
	if duration <= 0 {
		return 0, 0
	}
	
	totalBytesIn := t.BytesIn[len(t.BytesIn)-1] - t.BytesIn[0]
	totalBytesOut := t.BytesOut[len(t.BytesOut)-1] - t.BytesOut[0]
	
	avgBytesIn = float64(totalBytesIn) / duration
	avgBytesOut = float64(totalBytesOut) / duration
	
	return avgBytesIn, avgBytesOut
}

// IsExpired checks if a DNS entry has expired
func (d *NetworkDnsEntry) IsExpired() bool {
	return time.Now().After(d.ExpiresAt)
}

// GetSuccessRate calculates the success rate of a latency measurement
func (l *NetworkLatencyMeasurement) GetSuccessRate() float64 {
	if l.ProbesSent == 0 {
		return 0
	}
	return float64(l.ProbesReceived) / float64(l.ProbesSent) * 100
}

// IsHighLatency checks if latency is considered high
func (l *NetworkLatencyMeasurement) IsHighLatency(threshold time.Duration) bool {
	return l.AvgLatency > threshold
}

// IsHighPacketLoss checks if packet loss is considered high
func (l *NetworkLatencyMeasurement) IsHighPacketLoss(threshold float64) bool {
	return l.PacketLoss > threshold
}