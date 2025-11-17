package system

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"aDex-UI/backend/utils"
)

// Service handles system monitoring and operations
type Service struct {
	platform    *utils.FeatureDetection
	isMonitoring bool
	stopChan    chan struct{}
}

// NewService creates a new system service instance
func NewService() *Service {
	return &Service{
		platform:  utils.DetectPlatform(),
		stopChan:  make(chan struct{}),
	}
}

// SystemInfo represents system information
type SystemInfo struct {
	OS           string
	Architecture string
	Hostname     string
	Uptime       time.Duration
	CPUUsage     float64
	MemoryUsage  MemoryInfo
	DiskUsage    []DiskInfo
	NetworkInfo  NetworkInfo
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Available uint64
	Percent   float64
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Mountpoint string
	Total      uint64
	Used       uint64
	Free       uint64
	Percent    float64
}

// NetworkInfo represents network information
type NetworkInfo struct {
	Interfaces []NetworkInterface
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name      string
	IPAddress string
	IsUp      bool
}

// GetSystemInfo retrieves current system information
func (s *Service) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	uptime := time.Duration(hostInfo.Uptime) * time.Second

	// Get network interfaces
	netInterfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var interfaces []NetworkInterface
	for _, iface := range netInterfaces {
		addrs := iface.Addrs
		var ipAddress string
		if len(addrs) > 0 {
			ipAddress = addrs[0].Addr
		}

		// Check if interface is up by checking if it has addresses
		isUp := len(addrs) > 0

		interfaces = append(interfaces, NetworkInterface{
			Name:      iface.Name,
			IPAddress: ipAddress,
			IsUp:      isUp,
		})
	}

	return &SystemInfo{
		OS:           hostInfo.OS,
		Architecture: hostInfo.KernelArch,
		Hostname:     hostInfo.Hostname,
		Uptime:       uptime,
		NetworkInfo: NetworkInfo{
			Interfaces: interfaces,
		},
	}, nil
}

// GetCPUUsage retrieves current CPU usage
func (s *Service) GetCPUUsage(ctx context.Context) (float64, error) {
	percent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return 0.0, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	if len(percent) > 0 {
		return percent[0], nil
	}
	return 0.0, nil
}

// GetMemoryUsage retrieves current memory usage
func (s *Service) GetMemoryUsage(ctx context.Context) (*MemoryInfo, error) {
	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}

	return &MemoryInfo{
		Total:     virtualMem.Total,
		Used:      virtualMem.Used,
		Available: virtualMem.Available,
		Percent:   virtualMem.UsedPercent,
	}, nil
}

// GetDiskUsage retrieves disk usage information
func (s *Service) GetDiskUsage(ctx context.Context) ([]DiskInfo, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	var diskInfos []DiskInfo
	for _, partition := range partitions {
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			continue // Skip partitions we can't get usage for
		}

		diskInfos = append(diskInfos, DiskInfo{
			Mountpoint: partition.Mountpoint,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			Percent:    usage.UsedPercent,
		})
	}

	return diskInfos, nil
}

// GetNetworkInfo retrieves network information
func (s *Service) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	netInterfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var interfaces []NetworkInterface
	for _, iface := range netInterfaces {
		addrs := iface.Addrs
		var ipAddress string
		if len(addrs) > 0 {
			ipAddress = addrs[0].Addr
		}

		// Check if interface is up by checking if it has addresses
		isUp := len(addrs) > 0

		interfaces = append(interfaces, NetworkInterface{
			Name:      iface.Name,
			IPAddress: ipAddress,
			IsUp:      isUp,
		})
	}

	return &NetworkInfo{
		Interfaces: interfaces,
	}, nil
}

// StartMonitoring starts system monitoring
func (s *Service) StartMonitoring(ctx context.Context, interval time.Duration) error {
	if s.isMonitoring {
		return fmt.Errorf("monitoring is already running")
	}

	s.isMonitoring = true
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		defer func() { s.isMonitoring = false }()

		for {
			select {
			case <-ticker.C:
				// This is where you would emit monitoring events
				// For now, we just collect data internally
				s.collectMetrics(ctx)

			case <-s.stopChan:
				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// StopMonitoring stops system monitoring
func (s *Service) StopMonitoring() error {
	if !s.isMonitoring {
		return nil
	}

	close(s.stopChan)
	s.stopChan = make(chan struct{})
	s.isMonitoring = false

	return nil
}

// collectMetrics collects current system metrics
func (s *Service) collectMetrics(ctx context.Context) {
	// Collect CPU, memory, disk usage
	// This could emit events or store data for later retrieval
	// For now, this is a placeholder for monitoring logic
	_, _ = s.GetCPUUsage(ctx)
	_, _ = s.GetMemoryUsage(ctx)
	_, _ = s.GetDiskUsage(ctx)
}
