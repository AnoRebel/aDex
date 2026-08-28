package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aDex-UI/internal/services/audio"
	"aDex-UI/internal/services/config"
	"aDex-UI/internal/services/filesystem"
	"aDex-UI/internal/services/system"
	"aDex-UI/internal/services/terminal"
	"aDex-UI/internal/services/theme"
	"aDex-UI/internal/utils"
	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
	"aDex-UI/internal/services/colorscheme"
	"aDex-UI/internal/services/font"
	"aDex-UI/internal/services/network"
	"aDex-UI/internal/services/security"
	"aDex-UI/internal/services/settings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// convertToStruct converts an interface{} to a specific struct using JSON marshaling
// This is used for converting data from frontend (which comes as interface{}) to proper Go structs
func convertToStruct(source interface{}, dest interface{}) error {
	// Marshal the source to JSON
	jsonData, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}

	// Unmarshal JSON to destination struct
	if err := json.Unmarshal(jsonData, dest); err != nil {
		return fmt.Errorf("failed to unmarshal to destination: %w", err)
	}

	return nil
}

// ServiceCoordinator manages all application services
type ServiceCoordinator struct {
	platform *utils.FeatureDetection
	eventBus *events.EventBus

	// Services
	filesystem  *filesystem.Service
	system      *system.Service
	terminal    *terminal.Service
	audio       *audio.Service
	network     *network.NetworkService
	config      *config.Service
	theme       *theme.Service
	colorScheme *colorscheme.Service
	font        *font.Service
	// uiSettings is the file-backed source of truth for the frontend's
	// `adex-settings`. localStorage on the frontend is only a reactive
	// cache seeded from / flushed to this store.
	uiSettings *settings.UIStore

	// Session lock. Nil only if construction failed; every accessor guards.
	lock *security.LockService

	// Service state
	isStarted bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	// Cached self-GeoIP — geo lookups rate-limit hard, and the user's
	// public location rarely changes within a 30-min session. Cleared
	// implicitly on app restart. Guarded by its OWN mutex (geoIPMu),
	// NOT the global `mu`, because the cache update was holding `mu`
	// in write mode across a multi-second HTTP fetch which deadlocked
	// every other coordinator method (CreateTerminal, GetCPUUsage, etc).
	selfGeoIP   map[string]interface{}
	selfGeoIPAt time.Time
	geoIPMu     sync.Mutex
	// inflight prevents multiple concurrent network fetches when many
	// frontend components hit GetSelfGeoIP simultaneously at boot.
	geoIPInflight bool
	geoIPDone     chan struct{}
}

// NewServiceCoordinator creates a new service coordinator
func NewServiceCoordinator() *ServiceCoordinator {
	platform := utils.DetectPlatform()

	return &ServiceCoordinator{
		platform: platform,
		eventBus: events.GetEventBus(),
		lock:     security.NewLockService(),
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

	// Each `bootStep` pair below emits a `wait` line, runs init, then
	// emits `success` (or returns an error which main.go can surface).
	// The frontend boot screen waits for `boot.complete` before
	// fading out.

	emitBootStage("wait", "config", "Loading service registry...", "")
	sc.filesystem = filesystem.NewService()
	sc.system = system.NewService()
	sc.terminal = terminal.NewService()
	sc.audio = audio.NewService()
	sc.network = network.NewNetworkService()
	sc.config = config.NewService()
	sc.theme = theme.NewService()
	sc.colorScheme = colorscheme.NewService(configPath, schemesDir)
	sc.font = font.NewService(fontConfigPath)
	// UI settings store — file-backed source of truth for the
	// frontend's adex-settings. A failure here is non-fatal: the
	// frontend still works off its localStorage cache, it just
	// won't survive a cleared-cache / reinstall. So we log and
	// continue rather than aborting boot.
	if uiStore, err := settings.NewUIStore(logger.GetDefaultLogger()); err != nil {
		emitBootStage("error", "config", "UI settings store unavailable", err.Error())
	} else {
		sc.uiSettings = uiStore
	}
	emitBootStage("success", "config", "Service registry ready", "")

	// Initialize color scheme service
	emitBootStage("wait", "colorscheme", "Loading color schemes...", "")
	if err := sc.colorScheme.Initialize(sc.ctx); err != nil {
		emitBootStage("error", "colorscheme", "Color scheme service failed", err.Error())
		return fmt.Errorf("failed to initialize color scheme service: %w", err)
	}
	emitBootStage("success", "colorscheme", "Color schemes loaded", "")

	// Initialize font service
	emitBootStage("wait", "font", "Scanning system fonts...", "")
	if err := sc.font.Initialize(sc.ctx); err != nil {
		emitBootStage("error", "font", "Font service failed", err.Error())
		return fmt.Errorf("failed to initialize font service: %w", err)
	}
	emitBootStage("success", "font", "Font service ready", "")

	// Set event bus for network service
	emitBootStage("wait", "network", "Resolving network interfaces...", "")
	sc.network.SetEventBus(sc.eventBus)
	emitBootStage("success", "network", "Network service ready", "")

	// Wire the terminal service. Its PTY read loop emits
	// `terminal.output.<id>` events straight through the Wails v3 runtime,
	// so there is no context to hand it here.
	emitBootStage("wait", "terminal", "Spawning PTY service...", "")
	sc.terminal.SetEventBus(sc.eventBus)
	emitBootStage("success", "terminal", "Terminal service ready", "")

	// Set up event listeners
	emitBootStage("wait", "events", "Wiring event listeners...", "")
	if err := sc.setupEventListeners(); err != nil {
		emitBootStage("error", "events", "Event listener setup failed", err.Error())
		return fmt.Errorf("failed to setup event listeners: %w", err)
	}
	emitBootStage("success", "events", "Event bus ready", "")

	sc.isStarted = true

	// Publish startup event
	sc.eventBus.Publish(sc.ctx, events.AppStarted, map[string]interface{}{
		"services": []string{"filesystem", "system", "terminal", "audio", "network", "config", "theme", "colorscheme", "font"},
	}, "coordinator")

	// Final boot-complete signal to the frontend boot screen so it can
	// fade out as soon as services are actually ready (not after a wall-
	// clock minimum). The screen still respects its `minDuration` prop
	// for visual continuity, but it now stops waiting on real progress
	// rather than a fake setTimeout chain.
	if app := application.Get(); app != nil {
		app.Event.Emit("boot.complete", map[string]interface{}{
			"at": time.Now().UnixMilli(),
		})
	}

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
					Type:      "cpu",
					Resource:  "CPU",
					Threshold: 90,
					Current:   data.CPUUsage,
					Message:   "High CPU usage detected",
					Severity:  "warning",
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
// Shutdown gracefully tears every service down. Called from Wails's
// OnShutdown. Order matters:
//
//  1. Publish the shutdown event so subscribers can clean up.
//  2. Stop monitoring loops (StopMonitoring is idempotent).
//  3. Shutdown the audio service so its CGO PCM writer goroutine winds down
//     (without this, `go test` hangs at process exit on Linux ALSA).
//  4. Cancel the coordinator's context so any context-bound goroutines
//     return.
//  5. Close per-service resources.
//  6. Drain the event bus and wait for delivery goroutines to exit.
func (sc *ServiceCoordinator) Shutdown() error {
	sc.mu.Lock()
	if !sc.isStarted {
		sc.mu.Unlock()
		return nil
	}

	// Snapshot what we need under the lock, then release it before any
	// blocking work so we don't deadlock with services calling back into
	// the coordinator during their own teardown.
	bus := sc.eventBus
	ctx := sc.ctx
	cancel := sc.cancel
	audio := sc.audio
	term := sc.terminal
	fs := sc.filesystem
	sc.isStarted = false
	sc.mu.Unlock()

	if bus != nil {
		_ = bus.Publish(ctx, events.AppShutdown, nil, "coordinator")
	}

	_ = sc.StopMonitoring()

	// Kill every active PTY child BEFORE cancelling the parent ctx —
	// otherwise the read goroutines block on PTY.Read while we wait for
	// them to exit, and "wails dev" can't reap the binary because of
	// the lingering child processes. Each Close inside Shutdown gets a
	// 2s timeout so a wedged PTY can't hold up app teardown.
	if term != nil {
		term.Shutdown()
	}

	if audio != nil {
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 2*time.Second)
		_ = audio.Shutdown(shutdownCtx)
		cancelShutdown()
	}

	if cancel != nil {
		cancel()
	}

	if fs != nil {
		fs.Close()
	}

	if bus != nil {
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 2*time.Second)
		_ = bus.Shutdown(shutdownCtx)
		cancelShutdown()
	}

	return nil
}

// getServiceInternal is intentionally unexported so Wails does NOT bind
// it. Wails v2 binds every uppercase-leading method of the bound struct
// and JSON-marshals the return value at call time. Our service structs
// hold channels, contexts, mutexes, callbacks, and (for audio) a CGO
// oto context — none of which JSON-encode. Returning any of those
// triggers a launch fatal:
//
//   FAT | json: unsupported type: func() error
//
// The returned `interface{}` is non-nil for known types, nil otherwise;
// callers type-assert.
func (sc *ServiceCoordinator) getServiceInternal(serviceType string) interface{} {
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

// Bus accessor intentionally omitted.
//
// Wails v2 binds every exported (uppercase-leading) method of the struct
// passed to options.App.Bind to the frontend, and every return value is
// shipped through encoding/json. *events.EventBus contains
// context.CancelFunc fields (per-subscription) which JSON cannot encode,
// triggering the launch fatal:
//
//   FAT | json: unsupported type: func() error
//
// Therefore we DO NOT expose `GetEventBus()` on the coordinator.
// Internal Go callers (tests, adjacent services) that need the bus
// construct their own via `events.NewEventBus()` and inject it where
// required. The global singleton is reachable via `events.GetEventBus()`
// in the events package itself.

// GetPlatform returns platform information
func (sc *ServiceCoordinator) GetPlatform() *utils.FeatureDetection {
	return sc.platform
}

// StartupPaths returns the directories the frontend can offer the user as
// "open in this directory on launch" choices.
//
//	{ "home": "/home/ano", "cwd": "/var/www/aDex-UI" }
//
// Driven by a new initial-cwd setting in the Settings modal so the
// terminal + file manager respect a user-chosen default.
type StartupPaths struct {
	Home string `json:"home"`
	CWD  string `json:"cwd"`
}

// GetStartupPaths exposes the launch CWD and the user's home directory so
// the frontend can route the terminal + file manager to whichever the
// user picked in Settings → System → "Open in".
func (sc *ServiceCoordinator) GetStartupPaths() StartupPaths {
	home := ""
	if sc.platform != nil {
		home = sc.platform.GetHomeDirectory()
	}
	cwd, err := os.Getwd()
	if err != nil || cwd == "" {
		// On error fall back to home so we always return something
		// usable; an empty string would force the frontend into '/'.
		cwd = home
	}
	return StartupPaths{Home: home, CWD: cwd}
}

// IsStarted returns whether the coordinator is started
func (sc *ServiceCoordinator) IsStarted() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.isStarted
}

// SetWailsContext propagates a Wails-issued lifecycle context to services
// that emit runtime events. MUST be called from main.go inside Wails'
// OnStartup hook — Wails' EventsEmit terminates the process via log.Fatalf
// if it receives a context it didn't issue (recover() can't catch it).
// Frontend code never calls this; it's exported only for main.go.
//
// Call BEFORE Initialize so Initialize() can publish boot.stage events
// during service bring-up. The terminal service (created inside
// Initialize) gets its copy via Initialize once it exists.
// ServiceStartup is the Wails v3 service lifecycle hook. Wails calls it
// during app.Run() before any window is shown, and aborts startup if it
// returns an error. Services are started in registration order, so
// anything this coordinator depends on must be registered before it.
//
// The ctx Wails supplies here is valid for the application's lifetime;
// Initialize derives its own cancellable child from it for the
// background monitoring goroutines.
func (sc *ServiceCoordinator) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	if err := sc.Initialize(ctx); err != nil {
		return err
	}
	return sc.StartMonitoring(ctx)
}

// ServiceShutdown is the Wails v3 service lifecycle hook fired during
// application teardown. Services shut down in reverse registration
// order. Shutdown is idempotent (guarded by isStarted), so repeated
// calls from other code paths remain safe.
func (sc *ServiceCoordinator) ServiceShutdown() error {
	return sc.Shutdown()
}

// emitBootStage emits a single boot-sequence event to the frontend.
// Shape:
//
//	{
//	  "stage":  "wait" | "success" | "warn" | "error",
//	  "tag":    "config" | "theme" | "terminal" | ...,
//	  "text":   "Loading settings...",
//	  "detail": "(optional error message or extra context)"
//	}
//
// Frontend `AdexBootScreen.vue` subscribes via the Wails v3 `Events.On`
// runtime API and renders one signale-style log line per event.
//
// Safe to call from any goroutine and at any point during startup:
// Wails v3 event emission takes no caller-supplied context, so there is
// no lock-ordering hazard here and no need for callers to pre-snapshot
// anything. Before the application exists (unit tests constructing the
// coordinator directly), application.Get() returns nil and the emit is
// skipped.
func emitBootStage(stage, tag, text, detail string) {
	app := application.Get()
	if app == nil {
		// Running under tests, or before the application is up.
		return
	}
	app.Event.Emit("boot.stage", map[string]interface{}{
		"stage":  stage,
		"tag":    tag,
		"text":   text,
		"detail": detail,
	})
}

// Terminal service methods

// CreateTerminal creates a new terminal session in the user's home
// directory. Kept for backwards compatibility with frontend code that
// pre-dates the Settings → System "Open in" preference.
func (sc *ServiceCoordinator) CreateTerminal(width, height int) (*terminal.Terminal, error) {
	if err := sc.guardLocked(); err != nil {
		return nil, err
	}

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

// CreateTerminalIn creates a new terminal session whose shell starts in
// the supplied working directory. Empty `cwd` falls back to home (matches
// the legacy CreateTerminal behavior).
func (sc *ServiceCoordinator) CreateTerminalIn(width, height int, cwd string) (*terminal.Terminal, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted {
		return nil, fmt.Errorf("service coordinator not started")
	}

	if sc.terminal == nil {
		return nil, fmt.Errorf("terminal service not available")
	}

	return sc.terminal.CreateTerminalIn(sc.ctx, width, height, cwd)
}

// WriteToTerminal writes data to a terminal session.
//
// Frontend-facing signature is `string` (not `[]byte`) because Wails
// bind marshals JS strings transparently for Go `string` params, while
// Go `[]byte` params expect a base64-encoded JSON string. Mixing those
// caused every keystroke to be re-decoded as base64 → garbage bytes
// → garbage echo. Keeping the wire type as string and converting
// internally is simpler and round-trips control codes (ESC, Ctrl-C,
// arrow-key escape sequences) correctly because xterm.js sends them
// as raw byte strings in the same encoding the PTY expects.
func (sc *ServiceCoordinator) WriteToTerminal(terminalID string, data string) error {
	if err := sc.guardLocked(); err != nil {
		return err
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return fmt.Errorf("terminal service not available")
	}

	return sc.terminal.WriteToTerminal(sc.ctx, terminalID, []byte(data))
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

// GetTerminalCWD returns the live current working directory of the shell
// process backing the given terminal. Replaces the previous HTTP endpoint at
// /api/terminal/session/:sessionId/cwd with a direct Wails binding.
func (sc *ServiceCoordinator) GetTerminalCWD(terminalID string) (string, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return "", fmt.Errorf("terminal service not available")
	}
	return sc.terminal.GetCurrentCWD(terminalID)
}

// GetCWDStats returns aggregate CWD-tracking stats across all terminals.
// Replaces the previous HTTP endpoint at /api/terminal/cwd/stats.
func (sc *ServiceCoordinator) GetCWDStats() (terminal.CWDStats, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return terminal.CWDStats{}, fmt.Errorf("terminal service not available")
	}
	return sc.terminal.GetCWDStats(), nil
}

// SetShellCommand stores a user-supplied shell-path override for new
// terminals. Empty string clears the override and falls back to the
// $SHELL / $ComSpec / platform-default chain. Called from
// Settings → Terminal → "Shell Path".
func (sc *ServiceCoordinator) SetShellCommand(shell string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return fmt.Errorf("terminal service not available")
	}
	return sc.terminal.SetShellCommand(shell)
}

// GetActiveShell returns the shell executable path the resolver would
// pick *right now* (override → $SHELL/$ComSpec → platform default).
// Lets the frontend status bar display the honest shell path instead
// of guessing '/bin/bash'.
func (sc *ServiceCoordinator) GetActiveShell() (string, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.terminal == nil {
		return "", fmt.Errorf("terminal service not available")
	}
	return sc.terminal.GetActiveShell(), nil
}

// Color Scheme service methods

// GetColorSchemes returns all available color schemes
func (sc *ServiceCoordinator) GetColorSchemes() (map[string]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return nil, fmt.Errorf("color scheme service not available")
	}

	// Convert map[string]*models.ColorScheme to map[string]interface{}
	schemes := sc.colorScheme.GetSchemes()
	result := make(map[string]interface{})
	for id, scheme := range schemes {
		result[id] = scheme
	}
	return result, nil
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

	// Convert interface{} to *models.ColorScheme
	var colorScheme models.ColorScheme
	if err := convertToStruct(scheme, &colorScheme); err != nil {
		return fmt.Errorf("invalid color scheme: %w", err)
	}

	return sc.colorScheme.CreateScheme(&colorScheme)
}

// UpdateColorScheme updates an existing color scheme
func (sc *ServiceCoordinator) UpdateColorScheme(scheme interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.colorScheme == nil {
		return fmt.Errorf("color scheme service not available")
	}

	// Convert interface{} to *models.ColorScheme
	var colorScheme models.ColorScheme
	if err := convertToStruct(scheme, &colorScheme); err != nil {
		return fmt.Errorf("invalid color scheme: %w", err)
	}

	return sc.colorScheme.UpdateScheme(&colorScheme)
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

	// Convert interface{} to *models.ColorSchemeConfig
	var schemeConfig models.ColorSchemeConfig
	if err := convertToStruct(config, &schemeConfig); err != nil {
		return fmt.Errorf("invalid color scheme config: %w", err)
	}

	return sc.colorScheme.UpdateConfig(&schemeConfig)
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

	// Convert interface{} to *models.ColorScheme
	var colorScheme models.ColorScheme
	if err := convertToStruct(scheme, &colorScheme); err != nil {
		return nil, fmt.Errorf("invalid color scheme: %w", err)
	}

	return sc.colorScheme.ValidateScheme(&colorScheme), nil
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
	var fontConfig models.FontConfiguration
	if err := convertToStruct(config, &fontConfig); err != nil {
		return fmt.Errorf("invalid font configuration: %w", err)
	}

	return sc.font.CreateConfiguration(&fontConfig)
}

// UpdateFontConfiguration updates an existing font configuration
func (sc *ServiceCoordinator) UpdateFontConfiguration(config interface{}) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return fmt.Errorf("font service not available")
	}

	// Convert interface{} to *models.FontConfiguration
	var fontConfig models.FontConfiguration
	if err := convertToStruct(config, &fontConfig); err != nil {
		return fmt.Errorf("invalid font configuration: %w", err)
	}

	return sc.font.UpdateConfiguration(&fontConfig)
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

	// Convert interface{} to *models.FontImportRequest
	var importRequest models.FontImportRequest
	if err := convertToStruct(request, &importRequest); err != nil {
		return nil, fmt.Errorf("invalid font import request: %w", err)
	}

	result, err := sc.font.ImportFont(&importRequest)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ValidateFont validates a font configuration
func (sc *ServiceCoordinator) ValidateFont(config interface{}) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.font == nil {
		return nil, fmt.Errorf("font service not available")
	}

	// Convert interface{} to *models.FontConfiguration
	var fontConfig models.FontConfiguration
	if err := convertToStruct(config, &fontConfig); err != nil {
		return nil, fmt.Errorf("invalid font configuration: %w", err)
	}

	result := sc.font.ValidateFont(&fontConfig)
	return result, nil
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

	// Convert interface{} to *models.FontSettings
	var fontSettings models.FontSettings
	if err := convertToStruct(settings, &fontSettings); err != nil {
		return fmt.Errorf("invalid font settings: %w", err)
	}

	return sc.font.UpdateSettings(&fontSettings)
}

// GetUISettings returns the file-persisted frontend settings
// (the `adex-settings` shape). The frontend calls this once on boot
// to hydrate its localStorage cache; the backend file is the source
// of truth, so on a conflict this value wins. Returns the typed
// *models.UISettings so the Wails-generated TS binding stays a
// checkable contract rather than an opaque blob.
func (sc *ServiceCoordinator) GetUISettings() (*models.UISettings, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.uiSettings == nil {
		return nil, fmt.Errorf("UI settings store not available")
	}

	return sc.uiSettings.Get(), nil
}

// SaveUISettings persists the frontend's settings object. The
// frontend sends the `adex-settings` value verbatim as a JSON string
// (debounced on every change), and this writes it atomically to the
// source-of-truth file. Taking a raw string rather than interface{}
// avoids a lossy map round-trip and lets the store unmarshal onto a
// defaults base so a partial / older-shaped payload still persists
// completely.
func (sc *ServiceCoordinator) SaveUISettings(payload string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.uiSettings == nil {
		return fmt.Errorf("UI settings store not available")
	}

	return sc.uiSettings.Save(payload)
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

	// Convert interface{} to *network.NetworkServiceConfig
	var networkConfig network.NetworkServiceConfig
	if err := convertToStruct(config, &networkConfig); err != nil {
		return fmt.Errorf("invalid network configuration: %w", err)
	}

	return sc.network.UpdateConfiguration(&networkConfig)
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

// System service methods - exposed for frontend consumption

// GetSystemInfo returns comprehensive system information
func (sc *ServiceCoordinator) GetSystemInfo() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return nil, fmt.Errorf("system service not available")
	}

	info, err := sc.system.GetSystemInfo(sc.ctx)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// GetCPUUsage returns current CPU usage with per-core data
func (sc *ServiceCoordinator) GetCPUUsage() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return nil, fmt.Errorf("system service not available")
	}

	cpuInfo, err := sc.system.GetCPUInfo(sc.ctx)
	if err != nil {
		// Fallback to simple usage if detailed info fails
		usage, simpleErr := sc.system.GetCPUUsage(sc.ctx)
		if simpleErr != nil {
			return nil, err
		}
		return map[string]interface{}{
			"usage":     usage,
			"cores":     []float64{},
			"coreCount": 0,
			"modelName": "",
			"frequency": 0.0,
		}, nil
	}

	return map[string]interface{}{
		"usage":     cpuInfo.Usage,
		"cores":     cpuInfo.Cores,
		"coreCount": cpuInfo.CoreCount,
		"modelName": cpuInfo.ModelName,
		"frequency": cpuInfo.Frequency,
	}, nil
}

// GetSelfGeoIP returns the geolocation for the current public IP.
//
// Two-provider strategy with automatic failover:
//
//  1. PRIMARY: geo.kamero.ai — free, open source, no API key, no
//     hard rate limit, sub-50ms via Vercel Edge. One round trip.
//  2. FALLBACK: ipify (just the IP) → iplocate (geo for that IP).
//     Two round trips but bulletproof — both providers are widely
//     used and have generous free tiers.
//
// Both paths are tried before giving up; we cache the answer for 30
// minutes since public-IP geolocation rarely changes mid-session.
// Returns an empty map (not error) on total failure so the UI shows
// "Resolving..." and retries on the next poll cycle.
func (sc *ServiceCoordinator) GetSelfGeoIP() (map[string]interface{}, error) {
	// Read parent ctx under the global lock (cheap, no network), then
	// drop the lock immediately. Everything else uses geoIPMu only.
	sc.mu.RLock()
	parentCtx := sc.ctx
	sc.mu.RUnlock()

	// Fast path: warm cache. Hold geoIPMu only long enough to peek.
	sc.geoIPMu.Lock()
	if sc.selfGeoIP != nil && time.Since(sc.selfGeoIPAt) < 30*time.Minute {
		out := sc.selfGeoIP
		sc.geoIPMu.Unlock()
		return out, nil
	}

	// Coalesce concurrent callers: if a fetch is already in flight,
	// wait for it instead of starting a duplicate. AdexGlobe +
	// AdexNetstat both call this at mount, and previously they each
	// fired their own 8-second HTTP fetch.
	if sc.geoIPInflight {
		done := sc.geoIPDone
		sc.geoIPMu.Unlock()
		<-done
		sc.geoIPMu.Lock()
		out := sc.selfGeoIP
		sc.geoIPMu.Unlock()
		if out == nil {
			return map[string]interface{}{}, nil
		}
		return out, nil
	}
	sc.geoIPInflight = true
	sc.geoIPDone = make(chan struct{})
	sc.geoIPMu.Unlock()

	// Always close the channel when we're done so waiters wake up,
	// regardless of whether we got data or hit an error.
	defer func() {
		sc.geoIPMu.Lock()
		sc.geoIPInflight = false
		close(sc.geoIPDone)
		sc.geoIPDone = nil
		sc.geoIPMu.Unlock()
	}()

	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithTimeout(parentCtx, 8*time.Second)
	defer cancel()

	// iplocate is the PRIMARY now because it's the only provider in
	// our chain that returns an ISP (asn.name). kamero is a fast +
	// lightweight fallback when iplocate is unreachable.
	out := tryIpifyAndIplocate(ctx)
	if out == nil {
		out = tryKameroGeo(ctx)
	}
	if out == nil {
		return map[string]interface{}{}, nil
	}

	sc.geoIPMu.Lock()
	sc.selfGeoIP = out
	sc.selfGeoIPAt = time.Now()
	sc.geoIPMu.Unlock()

	return out, nil
}

// tryKameroGeo: single round trip to kamero. Returns nil on any
// failure so the caller can fall through to the next provider.
func tryKameroGeo(ctx context.Context) map[string]interface{} {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://geo.kamero.ai/api/geo", nil)
	if err != nil {
		return nil
	}
	// kamero rejects requests without a User-Agent with a connection
	// reset. Stable identifier so they can rate-limit per-app fairly.
	req.Header.Set("User-Agent", "aDex-UI/1.0 (https://github.com/AnoRebel/Dex-UI)")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}
	// kamero returns latitude/longitude as STRINGS — parse them so the
	// frontend can plot on the canvas without extra coercion.
	return map[string]interface{}{
		"latitude":  parseFloatField(raw["latitude"]),
		"longitude": parseFloatField(raw["longitude"]),
		"city":      raw["city"],
		"region":    raw["countryRegion"],
		"country":   raw["country"],
		"continent": raw["continent"],
		"timezone":  raw["timezone"],
		"ip":        raw["ip"],
	}
}

// tryIpifyAndIplocate: ipify → iplocate fallback. Two requests but
// reliable. ipify just gives us the public IP (the user's egress);
// iplocate gives the full geo record for that IP.
func tryIpifyAndIplocate(ctx context.Context) map[string]interface{} {
	// Step 1: get our public IP from ipify.
	ipReq, err := http.NewRequestWithContext(ctx, "GET", "https://api.ipify.org?format=json", nil)
	if err != nil {
		return nil
	}
	ipReq.Header.Set("User-Agent", "aDex-UI/1.0")
	ipResp, err := http.DefaultClient.Do(ipReq)
	if err != nil {
		return nil
	}
	defer ipResp.Body.Close()
	if ipResp.StatusCode != http.StatusOK {
		return nil
	}
	ipBody, err := io.ReadAll(ipResp.Body)
	if err != nil {
		return nil
	}
	var ipResult struct{ IP string `json:"ip"` }
	if err := json.Unmarshal(ipBody, &ipResult); err != nil || ipResult.IP == "" {
		return nil
	}

	// Step 2: enrich with iplocate.
	geoReq, err := http.NewRequestWithContext(ctx, "GET", "https://iplocate.io/api/lookup/"+ipResult.IP, nil)
	if err != nil {
		return nil
	}
	geoReq.Header.Set("User-Agent", "aDex-UI/1.0")
	geoReq.Header.Set("Accept", "application/json")
	geoResp, err := http.DefaultClient.Do(geoReq)
	if err != nil {
		return nil
	}
	defer geoResp.Body.Close()
	if geoResp.StatusCode != http.StatusOK {
		return nil
	}
	geoBody, err := io.ReadAll(geoResp.Body)
	if err != nil {
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(geoBody, &raw); err != nil {
		return nil
	}
	// iplocate returns lat/lon as numbers already; subdivision is the
	// region name. ASN.name is the closest analogue to "ISP" so we
	// surface it for the netstat panel.
	isp, _ := "", ""
	if asn, ok := raw["asn"].(map[string]interface{}); ok {
		if name, ok := asn["name"].(string); ok {
			isp = name
		}
	}
	return map[string]interface{}{
		"latitude":  parseFloatField(raw["latitude"]),
		"longitude": parseFloatField(raw["longitude"]),
		"city":      raw["city"],
		"region":    raw["subdivision"],
		"country":   raw["country_code"],
		"continent": raw["continent"],
		"timezone":  raw["time_zone"],
		"isp":       isp,
		"ip":        raw["ip"],
	}
}

// parseFloatField accepts either a JSON number or a numeric string and
// returns a float64. Used because providers disagree on the lat/lon
// type (kamero strings, iplocate numbers).
func parseFloatField(v interface{}) interface{} {
	switch x := v.(type) {
	case float64:
		return x
	case string:
		var f float64
		if _, err := fmt.Sscanf(x, "%f", &f); err == nil {
			return f
		}
	}
	return nil
}

// GetPowerInfo returns the current power source + battery percent
// using distatus/battery (cross-platform). The frontend SystemInfo
// panel polls this so POWER reflects "AC" vs "Battery" in real time
// — earlier the frontend hardcoded "AC Power".
func (sc *ServiceCoordinator) GetPowerInfo() (map[string]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return map[string]interface{}{
			"source":    "AC",
			"percent":   -1,
			"onBattery": false,
		}, nil
	}
	info, err := sc.system.GetPowerInfo(sc.ctx)
	if err != nil || info == nil {
		return map[string]interface{}{
			"source":    "AC",
			"percent":   -1,
			"onBattery": false,
		}, nil
	}
	return map[string]interface{}{
		"source":    info.Source,
		"percent":   info.Percent,
		"onBattery": info.OnBattery,
		"status":    info.Status,
	}, nil
}

// MeasureLatency does a TCP round-trip to the supplied host:port and
// returns elapsed milliseconds. Empty target → 8.8.8.8:53 (Google DNS).
// Returns -1 on failure rather than erroring so the UI shows "N/A"
// instead of an error toast.
//
// Frontend exposes this with two preset choices (8.8.8.8:53 and
// 1.1.1.1:53) plus a free-text override in Settings → Network. Picking
// TCP-DNS instead of ICMP avoids the CAP_NET_RAW / setuid dance that
// breaks ping on sandboxed builds (Snap/Flatpak/macOS app store).
func (sc *ServiceCoordinator) MeasureLatency(target string) (float64, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return -1, nil
	}
	return sc.system.MeasureLatency(sc.ctx, target)
}

// GetTemperatures returns hardware temperature sensors. Wraps the
// system service so the frontend's network-status / hardware mods can
// pull `{name, temperature, high, critical}` rows for display.
// Returns an empty array (not error) when sensors are unavailable.
func (sc *ServiceCoordinator) GetTemperatures() ([]map[string]interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return []map[string]interface{}{}, nil
	}
	sensors, err := sc.system.GetTemperatures(sc.ctx)
	if err != nil {
		return []map[string]interface{}{}, nil
	}
	out := make([]map[string]interface{}, 0, len(sensors))
	for _, s := range sensors {
		out = append(out, map[string]interface{}{
			"name":        s.Name,
			"temperature": s.Temperature,
			"high":        s.High,
			"critical":    s.Critical,
		})
	}
	return out, nil
}

// GetMemoryUsage returns current memory usage information
func (sc *ServiceCoordinator) GetMemoryUsage() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return nil, fmt.Errorf("system service not available")
	}

	memInfo, err := sc.system.GetMemoryUsage(sc.ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total": memInfo.Total,
		"used":  memInfo.Used,
		"free":  memInfo.Available,
		"usage": memInfo.Percent,
	}, nil
}

// GetDiskUsage returns disk usage information for all partitions
func (sc *ServiceCoordinator) GetDiskUsage() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return nil, fmt.Errorf("system service not available")
	}

	diskInfo, err := sc.system.GetDiskUsage(sc.ctx)
	if err != nil {
		return nil, err
	}

	// Convert to interface slice for JSON serialization
	result := make([]interface{}, len(diskInfo))
	for i, disk := range diskInfo {
		result[i] = map[string]interface{}{
			"mountpoint": disk.Mountpoint,
			"total":      disk.Total,
			"used":       disk.Used,
			"free":       disk.Free,
			"usage":      disk.Percent,
		}
	}

	return result, nil
}

// GetNetworkInfo returns network interface information
func (sc *ServiceCoordinator) GetNetworkInfo() (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return nil, fmt.Errorf("system service not available")
	}

	netInfo, err := sc.system.GetNetworkInfo(sc.ctx)
	if err != nil {
		return nil, err
	}
	return netInfo, nil
}

// GetTopProcesses returns top processes by CPU or memory usage
func (sc *ServiceCoordinator) GetTopProcesses(metric string, limit int) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return nil, fmt.Errorf("system service not available")
	}

	processes, err := sc.system.GetTopProcesses(sc.ctx, metric, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top processes: %w", err)
	}

	return processes, nil
}

// SignalProcess sends the given signal name (e.g. "SIGTERM", "SIGKILL")
// to the process at pid. Driven by the process-table context menu.
//
// We accept the signal as a string instead of an int so the Wails JSON
// boundary stays stable and platform-neutral; the system service maps
// the name onto a syscall.Signal internally.
func (sc *ServiceCoordinator) SignalProcess(pid int, signal string) error {
	if err := sc.guardLocked(); err != nil {
		return err
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.system == nil {
		return fmt.Errorf("system service not available")
	}
	return sc.system.SignalProcess(pid, system.ProcessSignal(signal))
}

// ReadDirectory reads the contents of a directory
func (sc *ServiceCoordinator) ReadDirectory(path string) (interface{}, error) {
	if err := sc.guardLocked(); err != nil {
		return nil, err
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return nil, fmt.Errorf("filesystem service not available")
	}

	entries, err := sc.filesystem.ReadDirectory(sc.ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	return entries, nil
}

// GetFileInfo returns metadata for a single path.
// Used by the file-manager Properties / Info modal.
func (sc *ServiceCoordinator) GetFileInfo(path string) (interface{}, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return nil, fmt.Errorf("filesystem service not available")
	}
	return sc.filesystem.GetFileInfo(sc.ctx, path)
}

// CreateDirectory creates a new directory at path with mode 0o755.
// Used by the file-manager "New Folder" action.
func (sc *ServiceCoordinator) CreateDirectory(path string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return fmt.Errorf("filesystem service not available")
	}
	return sc.filesystem.CreateDirectory(sc.ctx, path, 0o755)
}

// CreateFile creates an empty file at path.
// Used by the file-manager "New File" action.
func (sc *ServiceCoordinator) CreateFile(path string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return fmt.Errorf("filesystem service not available")
	}
	return sc.filesystem.WriteFile(sc.ctx, path, []byte{}, 0o644)
}

// DeleteFile permanently removes a file or empty directory.
// The frontend MUST confirm with the user before calling this.
func (sc *ServiceCoordinator) DeleteFile(path string) error {
	if err := sc.guardLocked(); err != nil {
		return err
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return fmt.Errorf("filesystem service not available")
	}
	return sc.filesystem.DeleteFile(sc.ctx, path)
}

// MoveFile moves/renames a file from src to dst (used for both rename and
// move-to-trash flows).
func (sc *ServiceCoordinator) MoveFile(src, dst string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return fmt.Errorf("filesystem service not available")
	}
	return sc.filesystem.MoveFile(sc.ctx, src, dst)
}

// CopyFile copies src to dst.
func (sc *ServiceCoordinator) CopyFile(src, dst string) error {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if !sc.isStarted || sc.filesystem == nil {
		return fmt.Errorf("filesystem service not available")
	}
	return sc.filesystem.CopyFile(sc.ctx, src, dst)
}

// GetCustomLayouts returns the user's custom panel arrangements, read from
// layouts.json in the config directory. A missing file yields an empty list —
// that is the normal case for a user who has defined none.
//
// Entries are returned as generic JSON: the frontend owns the panel and region
// vocabulary and validates against it, so parsing into a typed struct here
// would duplicate that vocabulary in a second place.
func (sc *ServiceCoordinator) GetCustomLayouts() ([]interface{}, error) {
	return settings.LoadCustomLayouts()
}

// GetCustomLayoutsPath returns where custom layouts are read from, so the
// settings panel can tell the user which file to create or edit.
func (sc *ServiceCoordinator) GetCustomLayoutsPath() string {
	return settings.CustomLayoutsPath()
}

// PlayCue plays an interface sound from the Go side.
//
// Audio moved out of the webview because the webview will not start playback
// before a user gesture (WebKitGTK's media-playback-requires-user-gesture,
// which Wails exposes no Linux setting for). That silently dropped every boot
// splash cue until the user first clicked or typed. Playing through the OS
// audio stack has no such precondition, so splash audio works from the first
// frame.
//
// Volume is the already-resolved value for this cue (0..1): the frontend still
// owns policy — per-cue volume, category mutes, rate limiting — and this call
// performs playback only.
//
// Returns nil when audio is unavailable. Cues are decorative and must never
// turn a missing sound device into a caller-visible failure.
func (sc *ServiceCoordinator) PlayCue(id string, volume float64) error {
	sc.mu.RLock()
	svc := sc.audio
	started := sc.isStarted
	sc.mu.RUnlock()

	if !started || svc == nil {
		return nil
	}
	return svc.PlayCue(id, volume)
}

// GetAvailableCues lists the cue ids the backend can play, so the frontend can
// tell whether to use the Go path or fall back to its own webview playback.
func (sc *ServiceCoordinator) GetAvailableCues() []string {
	sc.mu.RLock()
	svc := sc.audio
	sc.mu.RUnlock()

	if svc == nil {
		return []string{}
	}
	return svc.AvailableCues()
}

/* Session lock -----------------------------------------------------------
 *
 * Scope: this gates an already-running session against someone who walks up
 * to an unattended machine. It is NOT authentication. The application runs as
 * the invoking OS user, so anyone with access to that account can open a
 * terminal directly without going through aDex.
 *
 * Enforcement lives here rather than only in the UI: `guardLocked` is applied
 * to the methods that expose the filesystem, terminals and process control, so
 * a locked session refuses to serve even if the overlay is bypassed. */

// guardLocked returns ErrLocked when the session is locked.
func (sc *ServiceCoordinator) guardLocked() error {
	if sc.lock != nil && sc.lock.IsLocked() {
		return security.ErrLocked
	}
	return nil
}

// LockSession engages the lock. Fails when no passphrase is configured, so a
// user cannot lock themselves out with no way back in.
func (sc *ServiceCoordinator) LockSession() error {
	if sc.lock == nil {
		return fmt.Errorf("lock unavailable")
	}
	return sc.lock.Lock()
}

// UnlockSession clears the lock when the passphrase matches.
func (sc *ServiceCoordinator) UnlockSession(passphrase string) error {
	if sc.lock == nil {
		return fmt.Errorf("lock unavailable")
	}
	return sc.lock.Unlock(passphrase)
}

// IsSessionLocked reports the current lock state.
func (sc *ServiceCoordinator) IsSessionLocked() bool {
	return sc.lock != nil && sc.lock.IsLocked()
}

// IsLockConfigured reports whether a passphrase has been set.
func (sc *ServiceCoordinator) IsLockConfigured() bool {
	return sc.lock != nil && sc.lock.IsConfigured()
}

// SetLockPassphrase sets or changes the passphrase. Changing an existing one
// requires the current value.
func (sc *ServiceCoordinator) SetLockPassphrase(current, next string) error {
	if sc.lock == nil {
		return fmt.Errorf("lock unavailable")
	}
	return sc.lock.SetPassphrase(current, next)
}

// DisableLock removes the lock; requires the current passphrase.
func (sc *ServiceCoordinator) DisableLock(passphrase string) error {
	if sc.lock == nil {
		return fmt.Errorf("lock unavailable")
	}
	return sc.lock.Disable(passphrase)
}

// GetLockIdleTimeout returns the auto-lock delay in seconds; 0 means never.
func (sc *ServiceCoordinator) GetLockIdleTimeout() int {
	if sc.lock == nil {
		return 0
	}
	return sc.lock.IdleTimeoutSeconds()
}

// SetLockIdleTimeout sets the auto-lock delay in seconds; 0 disables it.
func (sc *ServiceCoordinator) SetLockIdleTimeout(seconds int) error {
	if sc.lock == nil {
		return fmt.Errorf("lock unavailable")
	}
	return sc.lock.SetIdleTimeout(seconds)
}
