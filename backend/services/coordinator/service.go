package coordinator

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"aDex-UI/internal/colorscheme"
	"aDex-UI/internal/events"
	"aDex-UI/internal/font"
	"aDex-UI/internal/network"
	"aDex-UI/backend/services/audio"
	"aDex-UI/backend/services/config"
	"aDex-UI/backend/services/filesystem"
	"aDex-UI/backend/services/system"
	"aDex-UI/backend/services/terminal"
	"aDex-UI/backend/services/theme"
	"aDex-UI/backend/utils"
)

// ServiceCoordinator manages all application services
type ServiceCoordinator struct {
	platform      *utils.FeatureDetection
	eventBus     *events.EventBus

	// Services
	filesystem   *filesystem.Service
	system       *system.Service
	terminal     *terminal.Service
	audio        *audio.Service
	network      *network.NetworkService
	config       *config.Service
	theme        *theme.Service
	colorScheme  *colorscheme.Service
	font         *font.Service

	// Service state
	isStarted    bool
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewServiceCoordinator creates a new service coordinator
func NewServiceCoordinator() *ServiceCoordinator {
	platform := utils.DetectPlatform()

	return &ServiceCoordinator{
		platform:  platform,
		eventBus:  events.GetEventBus(),
	}
}

// Initialize initializes all services
func (sc *ServiceCoordinator) Initialize(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.isStarted {
		return fmt.Errorf("service coordinator already started")
	}

	// Create context with cancellation
	sc.ctx, sc.cancel = context.WithCancel(ctx)

	// Get data directory for services
	dataDir := filepath.Join(".", "data")
	configPath := filepath.Join(dataDir, "colorscheme-config.json")
	schemesDir := filepath.Join(dataDir, "colorschemes")
	fontConfigPath := filepath.Join(dataDir, "font-config.json")

	// Initialize services in dependency order
	sc.filesystem = filesystem.NewService()
	sc.system = system.NewService()
	sc.terminal = terminal.NewService()
	sc.audio = audio.NewService()
	sc.network = network.NewNetworkService()
	sc.config = config.NewService()
	sc.theme = theme.NewService()
	sc.colorScheme = colorscheme.NewService(configPath, schemesDir)
	sc.font = font.NewService(fontConfigPath)

	// Initialize color scheme service
	if err := sc.colorScheme.Initialize(sc.ctx); err != nil {
		return fmt.Errorf("failed to initialize color scheme service: %w", err)
	}

	// Initialize font service
	if err := sc.font.Initialize(sc.ctx); err != nil {
		return fmt.Errorf("failed to initialize font service: %w", err)
	}

	// Set event bus for network service
	sc.network.SetEventBus(sc.eventBus)

	// Set up event listeners
	if err := sc.setupEventListeners(); err != nil {
		return fmt.Errorf("failed to setup event listeners: %w", err)
	}

	sc.isStarted = true

	// Publish startup event
	sc.eventBus.Publish(sc.ctx, events.AppStarted, map[string]interface{}{
		"services": []string{"filesystem", "system", "terminal", "audio", "network", "config", "theme", "colorscheme", "font"},
	}, "coordinator")

	return nil
}

// setupEventListeners sets up event listeners for service coordination
func (sc *ServiceCoordinator) setupEventListeners() error {
	// System monitoring events
	sc.eventBus.Subscribe(sc.ctx, []string{events.SystemInfoUpdated, events.SystemAlert}, sc.handleSystemEvent)

	// Terminal events
	sc.eventBus.Subscribe(sc.ctx, []string{events.TerminalCreated, events.TerminalClosed}, sc.handleTerminalEvent)

	// Filesystem events
	sc.eventBus.Subscribe(sc.ctx, []string{events.FileCreated, events.FileDeleted, events.FileModified}, sc.handleFilesystemEvent)

	// Config events
	sc.eventBus.Subscribe(sc.ctx, []string{events.ConfigChanged, events.ThemeChanged}, sc.handleConfigEvent)

	// Error events
	sc.eventBus.Subscribe(sc.ctx, []string{events.ErrorOccurred, events.PanicOccurred}, sc.handleErrorEvent)

	return nil
}

// handleSystemEvent handles system-related events
func (sc *ServiceCoordinator) handleSystemEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.SystemInfoUpdated:
		// Could trigger UI updates
		if data, ok := event.Data.(events.SystemInfoData); ok {
			// Check for alerts
			if data.CPUUsage > 90 {
				alertData := events.SystemAlertData{
					Type:       "cpu",
					Resource:   "CPU",
					Threshold:  90,
					Current:    data.CPUUsage,
					Message:    "High CPU usage detected",
					Severity:   "warning",
				}
				sc.eventBus.Publish(ctx, events.SystemAlert, alertData, "system-monitor")
			}
		}
	case events.SystemAlert:
		// Could trigger notifications or UI alerts
		_ = event.Data // Handle alert data
	}
	return nil
}

// handleTerminalEvent handles terminal-related events
func (sc *ServiceCoordinator) handleTerminalEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.TerminalCreated:
		// Log terminal creation, update terminal list
		_ = event.Data
	case events.TerminalClosed:
		// Clean up terminal resources
		_ = event.Data
	}
	return nil
}

// handleFilesystemEvent handles filesystem-related events
func (sc *ServiceCoordinator) handleFilesystemEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.FileCreated, events.FileDeleted, events.FileModified:
		// Could trigger UI updates or refreshes
		_ = event.Data
	}
	return nil
}

// handleConfigEvent handles configuration-related events
func (sc *ServiceCoordinator) handleConfigEvent(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.ConfigChanged:
		// Could trigger service reconfiguration
		if data, ok := event.Data.(events.ConfigData); ok {
			// Apply configuration changes to relevant services
			sc.applyConfigChange(data)
		}
	case events.ThemeChanged:
		// Update theme service
		_ = event.Data
	}
	return nil
}

// handleErrorEvent handles error-related events
func (sc *ServiceCoordinator) handleErrorEvent(ctx context.Context, event events.Event) error {
	// Log errors and potentially trigger recovery actions
	_ = event.Data
	return nil
}

// applyConfigChange applies configuration changes to relevant services
func (sc *ServiceCoordinator) applyConfigChange(data events.ConfigData) {
	switch data.Section {
	case "terminal":
		// Reconfigure terminal service if needed
		if sc.terminal != nil {
			// Update terminal configuration
		}
	case "audio":
		// Reconfigure audio service if needed
		if sc.audio != nil {
			// Update audio configuration
		}
	case "theme":
		// Update theme service
		if sc.theme != nil {
			// Apply theme changes
		}
	}
}

// StartMonitoring starts monitoring for all services that support it
func (sc *ServiceCoordinator) StartMonitoring(ctx context.Context) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted {
		return fmt.Errorf("service coordinator not started")
	}

	// Start system monitoring
	if sc.system != nil {
		if err := sc.system.StartMonitoring(ctx, 5*time.Second); err != nil {
			return fmt.Errorf("failed to start system monitoring: %w", err)
		}
	}

	// Start audio monitoring
	if sc.audio != nil {
		if err := sc.audio.StartMonitoring(ctx); err != nil {
			// Audio monitoring failure is not critical
			_ = err
		}
	}

	// Start network monitoring
	if sc.network != nil {
		if err := sc.network.StartMonitoring(ctx); err != nil {
			// Network monitoring failure is not critical
			_ = err
		}
	}

	return nil
}

// StopMonitoring stops monitoring for all services
func (sc *ServiceCoordinator) StopMonitoring() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted {
		return nil
	}

	// Stop system monitoring
	if sc.system != nil {
		if err := sc.system.StopMonitoring(); err != nil {
			return fmt.Errorf("failed to stop system monitoring: %w", err)
		}
	}

	// Stop audio monitoring
	if sc.audio != nil {
		if err := sc.audio.StopMonitoring(); err != nil {
			// Audio monitoring stop failure is not critical
			_ = err
		}
	}

	// Stop network monitoring
	if sc.network != nil {
		if err := sc.network.StopMonitoring(); err != nil {
			// Network monitoring stop failure is not critical
			_ = err
		}
	}

	return nil
}

// Shutdown gracefully shuts down all services
func (sc *ServiceCoordinator) Shutdown() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.isStarted {
		return nil
	}

	// Publish shutdown event
	sc.eventBus.Publish(sc.ctx, events.AppShutdown, nil, "coordinator")

	// Stop monitoring
	sc.StopMonitoring()

	// Cancel context to stop all goroutines
	if sc.cancel != nil {
		sc.cancel()
	}

	// Close services
	if sc.filesystem != nil {
		sc.filesystem.Close()
	}

	sc.isStarted = false
	return nil
}

// GetService returns a service by type (interface{})
func (sc *ServiceCoordinator) GetService(serviceType string) interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	switch serviceType {
	case "filesystem":
		return sc.filesystem
	case "system":
		return sc.system
	case "terminal":
		return sc.terminal
	case "audio":
		return sc.audio
	case "network":
		return sc.network
	case "config":
		return sc.config
	case "theme":
		return sc.theme
	case "colorscheme":
		return sc.colorScheme
	case "font":
		return sc.font
	default:
		return nil
	}
}

// GetEventBus returns the event bus
func (sc *ServiceCoordinator) GetEventBus() *events.EventBus {
	return sc.eventBus
}

// GetPlatform returns platform information
func (sc *ServiceCoordinator) GetPlatform() *utils.FeatureDetection {
	return sc.platform
}

// IsStarted returns whether the coordinator is started
func (sc *ServiceCoordinator) IsStarted() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.isStarted
}

// Terminal service methods

// CreateTerminal creates a new terminal session
func (sc *ServiceCoordinator) CreateTerminal(width, height int) (*terminal.Terminal, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted {
		return nil, fmt.Errorf("service coordinator not started")
	}

	if sc.terminal == nil {
		return nil, fmt.Errorf("terminal service not available")
	}

	return sc.terminal.CreateTerminal(sc.ctx, width, height)
}

// WriteToTerminal writes data to a terminal session
func (sc *ServiceCoordinator) WriteToTerminal(terminalID string, data []byte) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return fmt.Errorf("terminal service not available")
	}

	return sc.terminal.WriteToTerminal(sc.ctx, terminalID, data)
}

// ResizeTerminal resizes a terminal session
func (sc *ServiceCoordinator) ResizeTerminal(terminalID string, width, height int) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return fmt.Errorf("terminal service not available")
	}

	return sc.terminal.ResizeTerminal(sc.ctx, terminalID, width, height)
}

// CloseTerminal closes a terminal session
func (sc *ServiceCoordinator) CloseTerminal(terminalID string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return fmt.Errorf("terminal service not available")
	}

	return sc.terminal.CloseTerminal(sc.ctx, terminalID)
}

// ListTerminals returns all active terminal sessions
func (sc *ServiceCoordinator) ListTerminals() ([]string, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return nil, fmt.Errorf("terminal service not available")
	}

	return sc.terminal.ListTerminals(sc.ctx)
}

// GetTerminalInfo returns information about a terminal session
func (sc *ServiceCoordinator) GetTerminalInfo(terminalID string) (*terminal.Terminal, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return nil, fmt.Errorf("terminal service not available")
	}

	return sc.terminal.GetTerminalInfo(sc.ctx, terminalID)
}

// Color Scheme service methods

// GetColorSchemes returns all available color schemes
func (sc *ServiceCoordinator) GetColorSchemes() (map[string]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.GetSchemes(), nil
}

// GetColorScheme returns a specific color scheme by ID
func (sc *ServiceCoordinator) GetColorScheme(id string) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	scheme := sc.colorScheme.GetScheme(id)
	if scheme == nil {
		return nil, fmt.Errorf("color scheme with ID '%s' not found", id)
	}

	return scheme, nil
}

// CreateColorScheme creates a new color scheme
func (sc *ServiceCoordinator) CreateColorScheme(scheme interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.CreateScheme(scheme)
}

// UpdateColorScheme updates an existing color scheme
func (sc *ServiceCoordinator) UpdateColorScheme(scheme interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.UpdateScheme(scheme)
}

// DeleteColorScheme deletes a color scheme
func (sc *ServiceCoordinator) DeleteColorScheme(id string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.DeleteScheme(id)
}

// SetDefaultColorScheme sets the default color scheme
func (sc *ServiceCoordinator) SetDefaultColorScheme(id string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.SetDefaultScheme(id)
}

// GetDefaultColorScheme returns the default color scheme
func (sc *ServiceCoordinator) GetDefaultColorScheme() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	scheme := sc.colorScheme.GetDefaultScheme()
	if scheme == nil {
		return nil, fmt.Errorf("no default color scheme found")
	}

	return scheme, nil
}

// GetColorSchemeConfig returns the color scheme configuration
func (sc *ServiceCoordinator) GetColorSchemeConfig() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.GetConfig(), nil
}

// UpdateColorSchemeConfig updates the color scheme configuration
func (sc *ServiceCoordinator) UpdateColorSchemeConfig(config interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.UpdateConfig(config)
}

// GetColorSchemePreview generates a preview for a color scheme
func (sc *ServiceCoordinator) GetColorSchemePreview(id string) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.GetSchemePreview(id)
}

// ValidateColorScheme validates a color scheme
func (sc *ServiceCoordinator) ValidateColorScheme(scheme interface{}) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	return sc.colorScheme.ValidateScheme(scheme), nil
}

// Font service methods

// GetFontConfigurations returns all font configurations
func (sc *ServiceCoordinator) GetFontConfigurations() (map[string]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	configs := sc.font.GetConfigurations()
	result := make(map[string]interface{})
	for id, config := range configs {
		result[id] = config
	}
	return result, nil
}

// GetFontConfiguration returns a specific font configuration by ID
func (sc *ServiceCoordinator) GetFontConfiguration(id string) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	config := sc.font.GetConfiguration(id)
	if config == nil {
		return nil, fmt.Errorf("font configuration with ID '%s' not found", id)
	}

	return config, nil
}

// CreateFontConfiguration creates a new font configuration
func (sc *ServiceCoordinator) CreateFontConfiguration(config interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	// Convert interface{} to *models.FontConfiguration
	// This will need proper type conversion based on how the data comes from frontend
	return fmt.Errorf("font configuration creation needs proper type conversion")
}

// UpdateFontConfiguration updates an existing font configuration
func (sc *ServiceCoordinator) UpdateFontConfiguration(config interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	return fmt.Errorf("font configuration update needs proper type conversion")
}

// DeleteFontConfiguration deletes a font configuration
func (sc *ServiceCoordinator) DeleteFontConfiguration(id string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	return sc.font.DeleteConfiguration(id)
}

// GetSystemFonts returns all detected system fonts
func (sc *ServiceCoordinator) GetSystemFonts() (map[string]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	fonts := sc.font.GetSystemFonts()
	result := make(map[string]interface{})
	for id, font := range fonts {
		result[id] = font
	}
	return result, nil
}

// GetMonospaceFonts returns monospace fonts suitable for terminal use
func (sc *ServiceCoordinator) GetMonospaceFonts() ([]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	fonts := sc.font.GetMonospaceFonts()
	result := make([]interface{}, len(fonts))
	for i, font := range fonts {
		result[i] = font
	}
	return result, nil
}

// ScanSystemFonts scans the system for available fonts
func (sc *ServiceCoordinator) ScanSystemFonts() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	return sc.font.ScanSystemFonts()
}

// ImportFont imports a font from a file or URL
func (sc *ServiceCoordinator) ImportFont(request interface{}) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	return fmt.Errorf("font import needs proper type conversion")
}

// ValidateFont validates a font configuration
func (sc *ServiceCoordinator) ValidateFont(config interface{}) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	return fmt.Errorf("font validation needs proper type conversion")
}

// GetFontMetrics returns metrics for a font
func (sc *ServiceCoordinator) GetFontMetrics(family string, size int) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	metrics, err := sc.font.GetFontMetrics(family, size)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// GetFontSettings returns current font settings
func (sc *ServiceCoordinator) GetFontSettings() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	return sc.font.GetSettings(), nil
}

// UpdateFontSettings updates font settings
func (sc *ServiceCoordinator) UpdateFontSettings(settings interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	return fmt.Errorf("font settings update needs proper type conversion")
}

// GetDefaultFontConfiguration returns the default font configuration
func (sc *ServiceCoordinator) GetDefaultFontConfiguration() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	config := sc.font.GetDefaultConfiguration()
	if config == nil {
		return nil, fmt.Errorf("no default font configuration found")
	}

	return config, nil
}

// SetDefaultFontConfiguration sets the default font configuration
func (sc *ServiceCoordinator) SetDefaultFontConfiguration(id string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	return sc.font.SetDefaultConfiguration(id)
}

// Network service methods

// GetNetworkMetrics returns current network metrics
func (sc *ServiceCoordinator) GetNetworkMetrics() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return nil, fmt.Errorf("network service not available")
	}

	metrics, err := sc.network.GetNetworkMetrics(sc.ctx)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// GetNetworkConnections returns current network connections
func (sc *ServiceCoordinator) GetNetworkConnections() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return nil, fmt.Errorf("network service not available")
	}

	connections, err := sc.network.GetConnections(sc.ctx)
	if err != nil {
		return nil, err
	}
	return connections, nil
}

// GetBandwidthData returns bandwidth data for a specific interface
func (sc *ServiceCoordinator) GetBandwidthData(interfaceName string) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return nil, fmt.Errorf("network service not available")
	}

	data, err := sc.network.GetBandwidthData(interfaceName)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetNetworkStatistics returns comprehensive network statistics
func (sc *ServiceCoordinator) GetNetworkStatistics() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return nil, fmt.Errorf("network service not available")
	}

	stats, err := sc.network.GetStatistics(sc.ctx)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetNetworkAlerts returns current network alerts
func (sc *ServiceCoordinator) GetNetworkAlerts() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return nil, fmt.Errorf("network service not available")
	}

	alerts := sc.network.GetAlerts()
	return alerts, nil
}

// ResolveNetworkAlert resolves a network alert
func (sc *ServiceCoordinator) ResolveNetworkAlert(alertID string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return fmt.Errorf("network service not available")
	}

	return sc.network.ResolveAlert(alertID)
}

// ClearNetworkAlerts clears all resolved network alerts
func (sc *ServiceCoordinator) ClearNetworkAlerts() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return fmt.Errorf("network service not available")
	}

	sc.network.ClearAlerts()
	return nil
}

// StartNetworkMonitoring starts network monitoring
func (sc *ServiceCoordinator) StartNetworkMonitoring() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return fmt.Errorf("network service not available")
	}

	return sc.network.StartMonitoring(sc.ctx)
}

// StopNetworkMonitoring stops network monitoring
func (sc *ServiceCoordinator) StopNetworkMonitoring() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return fmt.Errorf("network service not available")
	}

	return sc.network.StopMonitoring()
}

// GetNetworkConfig returns the current network service configuration
func (sc *ServiceCoordinator) GetNetworkConfig() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return nil, fmt.Errorf("network service not available")
	}

	config := sc.network.GetConfiguration()
	return config, nil
}

// UpdateNetworkConfig updates the network service configuration
func (sc *ServiceCoordinator) UpdateNetworkConfig(config interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return fmt.Errorf("network service not available")
	}

	// This would need proper type conversion
	return fmt.Errorf("network config update needs proper type conversion")
}

// ResetNetworkService resets the network service
func (sc *ServiceCoordinator) ResetNetworkService() error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return fmt.Errorf("network service not available")
	}

	sc.network.Reset()
	return nil
}

// IsNetworkMonitoring returns whether network monitoring is currently active
func (sc *ServiceCoordinator) IsNetworkMonitoring() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.network == nil {
		return false
	}

	return sc.network.IsMonitoring()
}