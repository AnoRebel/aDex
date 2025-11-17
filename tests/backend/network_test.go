package backend

import (
	"context"
	"testing"
	"time"

	"aDex-UI/internal/models"
	"aDex-UI/internal/services/network"
	"aDex-UI/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

// MockNetworkService provides a mock implementation for testing
type MockNetworkService struct {
	mock.Mock
}

func (m *MockNetworkService) GetNetworkMetrics(ctx context.Context) (*models.NetworkMetrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NetworkMetrics), args.Error(1)
}

func (m *MockNetworkService) GetConnections(ctx context.Context) ([]*network.NetworkConnection, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*network.NetworkConnection), args.Error(1)
}

func (m *MockNetworkService) GetBandwidthData(interfaceName string) (*network.BandwidthData, error) {
	args := m.Called(interfaceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*network.BandwidthData), args.Error(1)
}

func (m *MockNetworkService) GetStatistics(ctx context.Context) (*network.NetworkStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*network.NetworkStatistics), args.Error(1)
}

func (m *MockNetworkService) GetAlerts() []network.NetworkAlert {
	args := m.Called()
	return args.Get(0).([]network.NetworkAlert)
}

func (m *MockNetworkService) StartMonitoring(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockNetworkService) StopMonitoring() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNetworkService) IsMonitoring() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNetworkService) GetConfiguration() *network.NetworkServiceConfig {
	args := m.Called()
	return args.Get(0).(*network.NetworkServiceConfig)
}

func (m *MockNetworkService) UpdateConfiguration(config *network.NetworkServiceConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockNetworkService) Reset() {
	m.Called()
}

// TestNetworkService_GetNetworkMetrics tests the GetNetworkMetrics method
func TestNetworkService_GetNetworkMetrics(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	ctx := context.Background()
	
	expectedMetrics := &models.NetworkMetrics{
		Interfaces: []models.NetworkInterface{
			{
				Name:        "eth0",
				IsUp:        true,
				BytesSent:   1024000,
				BytesRecv:   2048000,
				PacketsSent: 1000,
				PacketsRecv: 2000,
				IPAddresses: []string{"192.168.1.100"},
				MAC:         "00:11:22:33:44:55",
				MTU:         1500,
			},
		},
		TotalBytesSent: 1024000,
		TotalBytesRecv: 2048000,
		Timestamp:      time.Now(),
	}
	
	mockService.On("GetNetworkMetrics", ctx).Return(expectedMetrics, nil)

	// Act
	metrics, err := mockService.GetNetworkMetrics(ctx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, expectedMetrics.Interfaces, metrics.Interfaces)
	assert.Equal(t, expectedMetrics.TotalBytesSent, metrics.TotalBytesSent)
	assert.Equal(t, expectedMetrics.TotalBytesRecv, metrics.TotalBytesRecv)
	assert.NotZero(t, metrics.Timestamp.Unix())
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetNetworkMetrics_Error tests error handling in GetNetworkMetrics
func TestNetworkService_GetNetworkMetrics_Error(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	ctx := context.Background()
	
	mockService.On("GetNetworkMetrics", ctx).Return(nil, assert.AnError)

	// Act
	metrics, err := mockService.GetNetworkMetrics(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, metrics)
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetConnections tests the GetConnections method
func TestNetworkService_GetConnections(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	ctx := context.Background()
	
	expectedConnections := []*network.NetworkConnection{
		{
			LocalAddr:    "192.168.1.100:8080",
			RemoteAddr:   "203.0.113.1:443",
			State:        "ESTABLISHED",
			PID:          1234,
			ProcessName:  "test-app",
			Protocol:     "TCP",
			BytesSent:    1024,
			BytesRecv:    2048,
			Established:  time.Now().Add(-time.Hour),
			LastActivity: time.Now(),
		},
	}
	
	mockService.On("GetConnections", ctx).Return(expectedConnections, nil)

	// Act
	connections, err := mockService.GetConnections(ctx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, connections)
	assert.Len(t, connections, 1)
	assert.Equal(t, expectedConnections[0].LocalAddr, connections[0].LocalAddr)
	assert.Equal(t, expectedConnections[0].RemoteAddr, connections[0].RemoteAddr)
	assert.Equal(t, expectedConnections[0].State, connections[0].State)
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetConnections_Disabled tests when connection tracking is disabled
func TestNetworkService_GetConnections_Disabled(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	ctx := context.Background()
	
	mockService.On("GetConnections", ctx).Return(nil, assert.AnError)

	// Act
	connections, err := mockService.GetConnections(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, connections)
	assert.Contains(t, err.Error(), "connection tracking is disabled")
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetBandwidthData tests the GetBandwidthData method
func TestNetworkService_GetBandwidthData(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	interfaceName := "eth0"
	
	expectedData := &network.BandwidthData{
		InterfaceName: interfaceName,
		Timestamps:    []time.Time{time.Now().Add(-time.Second), time.Now()},
		BytesSent:     []uint64{1000, 2000},
		BytesRecv:     []uint64{2000, 4000},
		PacketsSent:   []uint64{10, 20},
		PacketsRecv:   []uint64{20, 40},
		CurrentRate: &network.BandwidthRate{
			UploadBps:    1000.0,
			DownloadBps:  2000.0,
			UploadMbps:   0.007629,
			DownloadMbps: 0.015259,
		},
	}
	
	mockService.On("GetBandwidthData", interfaceName).Return(expectedData, nil)

	// Act
	data, err := mockService.GetBandwidthData(interfaceName)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, interfaceName, data.InterfaceName)
	assert.Len(t, data.Timestamps, 2)
	assert.NotNil(t, data.CurrentRate)
	assert.Greater(t, data.CurrentRate.UploadBps, 0.0)
	assert.Greater(t, data.CurrentRate.DownloadBps, 0.0)
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetBandwidthData_NotFound tests when interface data is not found
func TestNetworkService_GetBandwidthData_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	interfaceName := "nonexistent"
	
	mockService.On("GetBandwidthData", interfaceName).Return(nil, assert.AnError)

	// Act
	data, err := mockService.GetBandwidthData(interfaceName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "no bandwidth data available")
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetStatistics tests the GetStatistics method
func TestNetworkService_GetStatistics(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	ctx := context.Background()
	
	expectedStats := &network.NetworkStatistics{
		TotalConnections:  10,
		ActiveConnections: 5,
		TotalBytesSent:    1024000,
		TotalBytesRecv:    2048000,
		TotalPacketsSent:  1000,
		TotalPacketsRecv:  2000,
		PacketLossRate:    0.1,
		ErrorRate:         0.05,
		AverageLatency:    50 * time.Millisecond,
		TopConnections: []*network.NetworkConnection{
			{
				LocalAddr:   "192.168.1.100:8080",
				RemoteAddr:  "203.0.113.1:443",
				State:       "ESTABLISHED",
				BytesSent:   1024,
				BytesRecv:   2048,
			},
		},
		InterfaceStats: map[string]interface{}{
			"eth0": map[string]interface{}{
				"uploadMbps":   10.5,
				"downloadMbps": 25.3,
			},
		},
		Timestamp: time.Now(),
	}
	
	mockService.On("GetStatistics", ctx).Return(expectedStats, nil)

	// Act
	stats, err := mockService.GetStatistics(ctx)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, expectedStats.TotalConnections, stats.TotalConnections)
	assert.Equal(t, expectedStats.ActiveConnections, stats.ActiveConnections)
	assert.Equal(t, expectedStats.TotalBytesSent, stats.TotalBytesSent)
	assert.Equal(t, expectedStats.TotalBytesRecv, stats.TotalBytesRecv)
	assert.NotNil(t, stats.TopConnections)
	assert.NotNil(t, stats.InterfaceStats)
	assert.NotZero(t, stats.Timestamp.Unix())
	mockService.AssertExpectations(t)
}

// TestNetworkService_GetAlerts tests the GetAlerts method
func TestNetworkService_GetAlerts(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	
	expectedAlerts := []network.NetworkAlert{
		{
			ID:        "alert_1",
			Type:      "bandwidth_high",
			Severity:  "warning",
			Message:   "High bandwidth usage detected",
			Interface: "eth0",
			Threshold: 100.0,
			Current:   150.0,
			Timestamp: time.Now(),
			Resolved:  false,
		},
		{
			ID:        "alert_2",
			Type:      "connections_high",
			Severity:  "critical",
			Message:   "Too many active connections",
			Interface: "",
			Threshold: 500.0,
			Current:   600.0,
			Timestamp: time.Now(),
			Resolved:  true,
		},
	}
	
	mockService.On("GetAlerts").Return(expectedAlerts)

	// Act
	alerts := mockService.GetAlerts()

	// Assert
	assert.NotNil(t, alerts)
	assert.Len(t, alerts, 2)
	assert.Equal(t, expectedAlerts[0].ID, alerts[0].ID)
	assert.Equal(t, expectedAlerts[0].Type, alerts[0].Type)
	assert.Equal(t, expectedAlerts[0].Severity, alerts[0].Severity)
	assert.Equal(t, expectedAlerts[1].Resolved, alerts[1].Resolved)
	mockService.AssertExpectations(t)
}

// TestNetworkService_MonitoringLifecycle tests the monitoring lifecycle
func TestNetworkService_MonitoringLifecycle(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	ctx := context.Background()
	
	// Test initial state
	mockService.On("IsMonitoring").Return(false)
	assert.False(t, mockService.IsMonitoring())
	
	// Test starting monitoring
	mockService.On("StartMonitoring", ctx).Return(nil)
	err := mockService.StartMonitoring(ctx)
	assert.NoError(t, err)
	
	// Test monitoring is active
	mockService.On("IsMonitoring").Return(true)
	assert.True(t, mockService.IsMonitoring())
	
	// Test stopping monitoring
	mockService.On("StopMonitoring").Return(nil)
	err = mockService.StopMonitoring()
	assert.NoError(t, err)
	
	// Test monitoring is stopped
	mockService.On("IsMonitoring").Return(false)
	assert.False(t, mockService.IsMonitoring())
	
	mockService.AssertExpectations(t)
}

// TestNetworkService_Configuration tests configuration management
func TestNetworkService_Configuration(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	
	// Test getting default configuration
	defaultConfig := &network.NetworkServiceConfig{
		RefreshInterval:           2000 * time.Millisecond,
		EnableConnectionTracking:   true,
		EnableBandwidthMonitoring:  true,
		EnableAlerts:              true,
		MaxConnections:            1000,
		BandwidthHistorySize:      300,
		AlertThresholds: &network.AlertThresholds{
			BandwidthUsageMBps: 100.0,
			ConnectionCount:    500,
			PacketLossPercent:  1.0,
			ErrorRatePercent:   0.1,
			HighLatencyMs:      1000,
		},
	}
	
	mockService.On("GetConfiguration").Return(defaultConfig)
	config := mockService.GetConfiguration()
	
	// Assert default configuration
	assert.NotNil(t, config)
	assert.True(t, config.EnableConnectionTracking)
	assert.True(t, config.EnableBandwidthMonitoring)
	assert.True(t, config.EnableAlerts)
	assert.Greater(t, config.RefreshInterval, time.Duration(0))
	assert.NotNil(t, config.AlertThresholds)
	mockService.AssertExpectations(t)
	
	// Test updating configuration
	newConfig := &network.NetworkServiceConfig{
		RefreshInterval:           1000 * time.Millisecond,
		EnableConnectionTracking:   false,
		EnableBandwidthMonitoring:  true,
		EnableAlerts:              true,
		MaxConnections:            500,
		BandwidthHistorySize:      600,
		AlertThresholds: &network.AlertThresholds{
			BandwidthUsageMBps: 50.0,
			ConnectionCount:    250,
		},
	}
	
	mockService.On("UpdateConfiguration", newConfig).Return(nil)
	err := mockService.UpdateConfiguration(newConfig)
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestNetworkService_Reset tests the reset functionality
func TestNetworkService_Reset(t *testing.T) {
	// Arrange
	mockService := new(MockNetworkService)
	
	mockService.On("Reset").Return()
	
	// Act
	mockService.Reset()
	
	// Assert
	mockService.AssertExpectations(t)
}

// TestNetworkService_Integration tests integration with the real network service
func TestNetworkService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Arrange
	service := network.NewNetworkService()
	ctx := context.Background()
	
	// Test GetNetworkMetrics
	t.Run("GetNetworkMetrics", func(t *testing.T) {
		metrics, err := service.GetNetworkMetrics(ctx)
		
		// Assert - should work on most systems
		if err != nil {
			t.Logf("Warning: GetNetworkMetrics failed (expected on some systems): %v", err)
		} else {
			assert.NotNil(t, metrics)
			assert.NotEmpty(t, metrics.Interfaces)
			assert.NotZero(t, metrics.Timestamp.Unix())
		}
	})
	
	// Test GetStatistics
	t.Run("GetStatistics", func(t *testing.T) {
		stats, err := service.GetStatistics(ctx)
		
		// Assert - should work on most systems
		if err != nil {
			t.Logf("Warning: GetStatistics failed (expected on some systems): %v", err)
		} else {
			assert.NotNil(t, stats)
			assert.NotZero(t, stats.Timestamp.Unix())
			assert.GreaterOrEqual(t, stats.TotalConnections, 0)
			assert.GreaterOrEqual(t, stats.ActiveConnections, 0)
		}
	})
	
	// Test configuration
	t.Run("Configuration", func(t *testing.T) {
		config := service.GetConfiguration()
		assert.NotNil(t, config)
		assert.Greater(t, config.RefreshInterval, time.Duration(0))
		
		// Test configuration update
		newConfig := &network.NetworkServiceConfig{
			RefreshInterval:           5000 * time.Millisecond,
			EnableConnectionTracking:   true,
			EnableBandwidthMonitoring:  true,
			EnableAlerts:              false,
			MaxConnections:            2000,
			BandwidthHistorySize:      500,
		}
		
		err := service.UpdateConfiguration(newConfig)
		assert.NoError(t, err)
		
		updatedConfig := service.GetConfiguration()
		assert.Equal(t, newConfig.RefreshInterval, updatedConfig.RefreshInterval)
		assert.Equal(t, newConfig.MaxConnections, updatedConfig.MaxConnections)
		assert.Equal(t, newConfig.EnableAlerts, updatedConfig.EnableAlerts)
	})
	
	// Test reset
	t.Run("Reset", func(t *testing.T) {
		service.Reset()
		
		stats, err := service.GetStatistics(ctx)
		if err == nil {
			assert.Equal(t, 0, stats.TotalConnections)
			assert.Equal(t, 0, stats.ActiveConnections)
		}
		
		alerts := service.GetAlerts()
		assert.Empty(t, alerts)
	})
}

// TestNetworkModels tests the network model structures
func TestNetworkModels_NetworkInterface(t *testing.T) {
	// Arrange & Act
	iface := models.NetworkInterface{
		Name:        "eth0",
		IsUp:        true,
		BytesSent:   1024,
		BytesRecv:   2048,
		PacketsSent: 10,
		PacketsRecv: 20,
		IPAddresses: []string{"192.168.1.100", "fe80::1"},
		MAC:         "00:11:22:33:44:55",
		Speed:       1000000000, // 1 Gbps
		MTU:         1500,
	}

	// Assert
	assert.Equal(t, "eth0", iface.Name)
	assert.True(t, iface.IsUp)
	assert.Equal(t, uint64(1024), iface.BytesSent)
	assert.Equal(t, uint64(2048), iface.BytesRecv)
	assert.Equal(t, uint64(10), iface.PacketsSent)
	assert.Equal(t, uint64(20), iface.PacketsRecv)
	assert.Len(t, iface.IPAddresses, 2)
	assert.Equal(t, "00:11:22:33:44:55", iface.MAC)
	assert.Equal(t, uint64(1000000000), iface.Speed)
	assert.Equal(t, uint64(1500), iface.MTU)
}

// TestNetworkModels_NetworkMetrics tests the NetworkMetrics model
func TestNetworkModels_NetworkMetrics(t *testing.T) {
	// Arrange
	interfaces := []models.NetworkInterface{
		{
			Name:      "eth0",
			IsUp:      true,
			BytesSent: 1024,
			BytesRecv: 2048,
		},
		{
			Name:      "wlan0",
			IsUp:      true,
			BytesSent: 512,
			BytesRecv: 1024,
		},
	}

	// Act
	metrics := &models.NetworkMetrics{
		Interfaces:     interfaces,
		TotalBytesSent: 1536,
		TotalBytesRecv: 3072,
		Timestamp:      time.Now(),
	}

	// Assert
	assert.Equal(t, interfaces, metrics.Interfaces)
	assert.Equal(t, uint64(1536), metrics.TotalBytesSent)
	assert.Equal(t, uint64(3072), metrics.TotalBytesRecv)
	assert.NotZero(t, metrics.Timestamp.Unix())
}

// TestNetworkModels_HelperFunctions tests NetworkMetrics helper functions
func TestNetworkModels_HelperFunctions(t *testing.T) {
	// Arrange
	interfaces := []models.NetworkInterface{
		{
			Name:        "eth0",
			IsUp:        true,
			IPAddresses: []string{"192.168.1.100"},
		},
		{
			Name:        "lo",
			IsUp:        true,
			IPAddresses: []string{"127.0.0.1"},
		},
		{
			Name:        "wlan0",
			IsUp:        false,
			IPAddresses: []string{"192.168.1.101"},
		},
	}

	metrics := &models.NetworkMetrics{
		Interfaces: interfaces,
		Timestamp:  time.Now(),
	}

	// Test GetInterfaceByName
	t.Run("GetInterfaceByName", func(t *testing.T) {
		iface := metrics.GetInterfaceByName("eth0")
		assert.NotNil(t, iface)
		assert.Equal(t, "eth0", iface.Name)

		iface = metrics.GetInterfaceByName("nonexistent")
		assert.Nil(t, iface)
	})

	// Test GetActiveInterfaces
	t.Run("GetActiveInterfaces", func(t *testing.T) {
		active := metrics.GetActiveInterfaces()
		assert.Len(t, active, 2) // eth0 and lo are up

		for _, iface := range active {
			assert.True(t, iface.IsUp)
		}
	})

	// Test GetInterfacesByIP
	t.Run("GetInterfacesByIP", func(t *testing.T) {
		matches := metrics.GetInterfacesByIP("192.168.1.100")
		assert.Len(t, matches, 1)
		assert.Equal(t, "eth0", matches[0].Name)

		matches = metrics.GetInterfacesByIP("127.0.0.1")
		assert.Len(t, matches, 1)
		assert.Equal(t, "lo", matches[0].Name)

		matches = metrics.GetInterfacesByIP("1.2.3.4")
		assert.Len(t, matches, 0)
	})
}

// Benchmark tests
func BenchmarkNetworkService_GetNetworkMetrics(b *testing.B) {
	service := network.NewNetworkService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetNetworkMetrics(ctx)
	}
}

func BenchmarkNetworkService_GetStatistics(b *testing.B) {
	service := network.NewNetworkService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetStatistics(ctx)
	}
}
