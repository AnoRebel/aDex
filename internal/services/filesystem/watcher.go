package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"
	"github.com/fsnotify/fsnotify"
)

// DirectoryWatcher provides file system monitoring capabilities
type DirectoryWatcher struct {
	eventBus     events.IEventBus
	config       *models.DirectoryWatcherConfig
	watcher      *fsnotify.Watcher
	watchedPaths map[string]*WatchContext
	mutex        sync.RWMutex
	isRunning    bool
	eventChan    chan *models.FileSystemEvent
	stopChan     chan struct{}
	debouncer    *Debouncer
}

// WatchContext contains information about a watched directory
type WatchContext struct {
	Path        string
	Config      *models.DirectoryWatcherConfig
	SessionID   string
	LastEvent   time.Time
	EventCount  int
	IgnoreList  []string
	Recursive   bool
}

// WatchEventType maps fsnotify events to our event types
type WatchEventType string

const (
	EventCreate   WatchEventType = "create"
	EventRemove   WatchEventType = "remove"
	EventRename   WatchEventType = "rename"
	EventWrite    WatchEventType = "write"
	EventChmod    WatchEventType = "chmod"
	EventMove     WatchEventType = "move"
)

// NewDirectoryWatcher creates a new directory watcher service
func NewDirectoryWatcher(eventBus events.IEventBus) (*DirectoryWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	watcher := &DirectoryWatcher{
		eventBus:     eventBus,
		watcher:      fsWatcher,
		watchedPaths: make(map[string]*WatchContext),
		config: &models.DirectoryWatcherConfig{
			Enabled:       false,
			WatchHidden:   false,
			DebounceMs:    300, // Default 300ms debounce delay as specified
			MaxEvents:     1000,
			BufferSize:    512,
			IgnorePatterns: []string{
				"*.tmp", "*.swp", "*.swo", ".git/*", ".svn/*",
				"node_modules/*", ".DS_Store", "Thumbs.db",
			},
		},
		eventChan: make(chan *models.FileSystemEvent, 512),
		stopChan:  make(chan struct{}),
	}

	// Initialize debouncer
	debounceConfig := &DebounceConfig{
		Delay:       time.Duration(watcher.config.DebounceMs) * time.Millisecond,
		MaxPending:  watcher.config.MaxEvents,
		BatchSize:   50,
		GroupByPath: true,
		GroupByType: false,
	}
	watcher.debouncer = NewDebouncer(debounceConfig)

	return watcher, nil
}

// SetConfig updates the watcher configuration
func (dw *DirectoryWatcher) SetConfig(config *models.DirectoryWatcherConfig) error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	dw.config = config

	// If watcher is running, restart with new config
	if dw.isRunning {
		if err := dw.restart(); err != nil {
			return fmt.Errorf("failed to restart watcher with new config: %w", err)
		}
	}

	return nil
}

// GetConfig returns the current watcher configuration
func (dw *DirectoryWatcher) GetConfig() *models.DirectoryWatcherConfig {
	dw.mutex.RLock()
	defer dw.mutex.RUnlock()

	// Return a copy to prevent modification
	configCopy := *dw.config
	return &configCopy
}

// Start begins the directory watching service
func (dw *DirectoryWatcher) Start() error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	if dw.isRunning {
		return fmt.Errorf("directory watcher is already running")
	}

	if !dw.config.Enabled {
		return fmt.Errorf("directory watcher is disabled in configuration")
	}

	dw.isRunning = true

	// Initialize debouncer if not already done
	if dw.debouncer == nil {
		debounceConfig := &DebounceConfig{
			Delay:       time.Duration(dw.config.DebounceMs) * time.Millisecond,
			MaxPending:  dw.config.MaxEvents,
			BatchSize:   50,
			GroupByPath: true,
			GroupByType: false,
		}
		dw.debouncer = NewDebouncer(debounceConfig)
	}

	// Start debounced event processing goroutine
	go dw.processDebouncedEvents()

	// Start fsnotify event processing goroutine
	go dw.processFSNotifyEvents()

	return nil
}

// Stop stops the directory watching service
func (dw *DirectoryWatcher) Stop() error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	if !dw.isRunning {
		return nil
	}

	dw.isRunning = false

	// Signal stop to goroutines
	close(dw.stopChan)
	dw.stopChan = make(chan struct{})

	// Stop the debouncer
	if dw.debouncer != nil {
		dw.debouncer.Stop()
	}

	// Close the underlying watcher
	if dw.watcher != nil {
		dw.watcher.Close()
	}

	// Clear watched paths
	dw.watchedPaths = make(map[string]*WatchContext)

	return nil
}

// IsRunning returns whether the watcher service is currently running
func (dw *DirectoryWatcher) IsRunning() bool {
	dw.mutex.RLock()
	defer dw.mutex.RUnlock()
	return dw.isRunning
}

// WatchDirectory starts watching a directory for changes
func (dw *DirectoryWatcher) WatchDirectory(path string, recursive bool) (string, error) {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	// Clean and validate path
	cleanPath := filepath.Clean(path)
	if cleanPath == "" {
		return "", fmt.Errorf("invalid path: empty string")
	}

	// Check if already watching
	if _, exists := dw.watchedPaths[cleanPath]; exists {
		return "", fmt.Errorf("directory %s is already being watched", cleanPath)
	}

	// Start the service if not running
	if !dw.isRunning {
		if err := dw.Start(); err != nil {
			return "", fmt.Errorf("failed to start watcher service: %w", err)
		}
	}

	// Add path to fsnotify watcher
	if err := dw.watcher.Add(cleanPath); err != nil {
		return "", fmt.Errorf("failed to add path %s to watcher: %w", cleanPath, err)
	}

	// Create watch context
	sessionID := generateSessionID()
	watchCtx := &WatchContext{
		Path:       cleanPath,
		Config:     dw.config,
		SessionID:  sessionID,
		LastEvent:  time.Now(),
		EventCount: 0,
		IgnoreList: dw.config.IgnorePatterns,
		Recursive:  recursive,
	}

	dw.watchedPaths[cleanPath] = watchCtx

	// If recursive, add subdirectories
	if recursive {
		if err := dw.addRecursiveWatches(cleanPath, watchCtx); err != nil {
			// Cleanup on error
			delete(dw.watchedPaths, cleanPath)
			dw.watcher.Remove(cleanPath)
			return "", fmt.Errorf("failed to add recursive watches: %w", err)
		}
	}

	return sessionID, nil
}

// StopWatching stops watching a directory
func (dw *DirectoryWatcher) StopWatching(path string) error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	cleanPath := filepath.Clean(path)
	watchCtx, exists := dw.watchedPaths[cleanPath]
	if !exists {
		return fmt.Errorf("directory %s is not being watched", cleanPath)
	}

	// Remove from fsnotify watcher
	if err := dw.watcher.Remove(cleanPath); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to remove path %s from watcher: %v\n", cleanPath, err)
	}

	// If recursive, remove subdirectory watches
	if watchCtx.Recursive {
		dw.removeRecursiveWatches(cleanPath)
	}

	// Remove from tracked paths
	delete(dw.watchedPaths, cleanPath)

	return nil
}

// GetWatchedDirectories returns a list of currently watched directories
func (dw *DirectoryWatcher) GetWatchedDirectories() []string {
	dw.mutex.RLock()
	defer dw.mutex.RUnlock()

	dirs := make([]string, 0, len(dw.watchedPaths))
	for path := range dw.watchedPaths {
		dirs = append(dirs, path)
	}

	return dirs
}

// GetWatchContext returns the watch context for a directory
func (dw *DirectoryWatcher) GetWatchContext(path string) (*WatchContext, error) {
	dw.mutex.RLock()
	defer dw.mutex.RUnlock()

	cleanPath := filepath.Clean(path)
	watchCtx, exists := dw.watchedPaths[cleanPath]
	if !exists {
		return nil, fmt.Errorf("directory %s is not being watched", cleanPath)
	}

	// Return a copy
	ctxCopy := *watchCtx
	return &ctxCopy, nil
}

// processDebouncedEvents processes debounced file system events
func (dw *DirectoryWatcher) processDebouncedEvents() {
	if dw.debouncer == nil {
		return
	}

	for {
		select {
		case event := <-dw.debouncer.Events():
			dw.handleFileSystemEvent(event)
		case <-dw.stopChan:
			return
		}
	}
}

// processFSNotifyEvents processes raw fsnotify events
func (dw *DirectoryWatcher) processFSNotifyEvents() {
	for {
		select {
		case fsEvent, ok := <-dw.watcher.Events:
			if !ok {
				return
			}
			dw.handleFSNotifyEvent(fsEvent)
		case err, ok := <-dw.watcher.Errors:
			if !ok {
				return
			}
			dw.handleWatcherError(err)
		case <-dw.stopChan:
			return
		}
	}
}

// handleFSNotifyEvent processes a single fsnotify event
func (dw *DirectoryWatcher) handleFSNotifyEvent(fsEvent fsnotify.Event) {
	// Find which watch context this event belongs to
	dw.mutex.RLock()
	watchCtx, exists := dw.watchedPaths[filepath.Dir(fsEvent.Name)]
	if !exists {
		// Check if the exact path is being watched
		watchCtx, exists = dw.watchedPaths[fsEvent.Name]
		if !exists {
			dw.mutex.RUnlock()
			return
		}
	}
	dw.mutex.RUnlock()

	// Skip if event should be ignored
	if dw.shouldIgnoreEvent(fsEvent, watchCtx) {
		return
	}

	// Convert fsnotify event to our event type
	eventType := dw.mapFSNotifyEvent(fsEvent)
	if eventType == "" {
		return
	}

	// Get file entry for the event
	entry, err := models.NewFileSystemEntry(fsEvent.Name)
	if err != nil {
		// Create a minimal entry if we can't get full info
		entry = &models.FileSystemEntry{
			Name:  filepath.Base(fsEvent.Name),
			Path:  fsEvent.Name,
			IsDir: fsEvent.Op&fsnotify.Create != 0,
		}
	}

	// Create file system event
	event := &models.FileSystemEvent{
		Path:      fsEvent.Name,
		Operation: string(eventType),
		Entry:     entry,
		Timestamp: time.Now().Unix(),
		SessionID: watchCtx.SessionID,
	}

	// Send event to debouncer for processing
	if dw.debouncer != nil {
		dw.debouncer.AddEvent(event)
		watchCtx.EventCount++
		watchCtx.LastEvent = time.Now()
	} else {
		// Fallback to direct event channel if debouncer not available
		select {
		case dw.eventChan <- event:
			watchCtx.EventCount++
			watchCtx.LastEvent = time.Now()
		default:
			// Event channel is full, drop the event
			fmt.Printf("Warning: event channel full, dropping event for %s\n", fsEvent.Name)
		}
	}
}

// handleFileSystemEvent processes a file system event
func (dw *DirectoryWatcher) handleFileSystemEvent(event *models.FileSystemEvent) {
	// Apply debouncing
	if dw.config.DebounceMs > 0 {
		dw.debounceEvent(event)
		return
	}

	// Publish event immediately
	dw.publishEvent(event)
}

// shouldIgnoreEvent determines if an event should be ignored
func (dw *DirectoryWatcher) shouldIgnoreEvent(fsEvent fsnotify.Event, watchCtx *WatchContext) bool {
	// Skip hidden files if not configured to watch them
	if !dw.config.WatchHidden {
		baseName := filepath.Base(fsEvent.Name)
		if strings.HasPrefix(baseName, ".") {
			return true
		}
	}

	// Check ignore patterns
	for _, pattern := range watchCtx.IgnoreList {
		if matched, _ := filepath.Match(pattern, filepath.Base(fsEvent.Name)); matched {
			return true
		}
		// Check directory patterns
		if matched, _ := filepath.Match(pattern, fsEvent.Name); matched {
			return true
		}
	}

	// Skip temporary files
	ext := filepath.Ext(fsEvent.Name)
	if ext == ".tmp" || ext == ".swp" || ext == ".swo" {
		return true
	}

	return false
}

// mapFSNotifyEvent maps fsnotify operation to our event type
func (dw *DirectoryWatcher) mapFSNotifyEvent(fsEvent fsnotify.Event) WatchEventType {
	if fsEvent.Op&fsnotify.Create == fsnotify.Create {
		return EventCreate
	}
	if fsEvent.Op&fsnotify.Remove == fsnotify.Remove {
		return EventRemove
	}
	if fsEvent.Op&fsnotify.Rename == fsnotify.Rename {
		return EventRename
	}
	if fsEvent.Op&fsnotify.Write == fsnotify.Write {
		return EventWrite
	}
	if fsEvent.Op&fsnotify.Chmod == fsnotify.Chmod {
		return EventChmod
	}

	return ""
}

// debounceEvent applies debouncing to events (legacy method - now handled by Debouncer)
func (dw *DirectoryWatcher) debounceEvent(event *models.FileSystemEvent) {
	// This method is deprecated - events are now handled by the Debouncer
	// Keeping this method for backward compatibility
	dw.publishEvent(event)
}

// publishEvent publishes a file system event
func (dw *DirectoryWatcher) publishEvent(event *models.FileSystemEvent) {
	if dw.eventBus != nil {
		ctx := context.Background()
		eventType := fmt.Sprintf("filesystem.%s", event.Operation)

		if err := dw.eventBus.Publish(ctx, eventType, event, "filesystem-watcher"); err != nil {
			fmt.Printf("Warning: failed to publish filesystem event: %v\n", err)
		}
	}
}

// handleWatcherError handles errors from the fsnotify watcher
func (dw *DirectoryWatcher) handleWatcherError(err error) {
	// Create error event
	event := &models.FileSystemEvent{
		Path:      "",
		Operation: "error",
		Timestamp: time.Now().Unix(),
		Error:     err.Error(),
	}

	// Publish error event
	if dw.eventBus != nil {
		ctx := context.Background()
		if err := dw.eventBus.Publish(ctx, "filesystem.error", event, "filesystem-watcher"); err != nil {
			fmt.Printf("Warning: failed to publish filesystem error event: %v\n", err)
		}
	}
}

// addRecursiveWatches adds watches for all subdirectories
func (dw *DirectoryWatcher) addRecursiveWatches(basePath string, watchCtx *WatchContext) error {
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Don't re-add the base path
			if path != basePath {
				if watchErr := dw.watcher.Add(path); watchErr != nil {
					fmt.Printf("Warning: failed to watch subdirectory %s: %v\n", path, watchErr)
				}
			}
		}

		return nil
	})
}

// removeRecursiveWatches removes watches for all subdirectories
func (dw *DirectoryWatcher) removeRecursiveWatches(basePath string) {
	// Note: In a more sophisticated implementation, we would track
	// all recursively added paths and remove them explicitly.
	// For now, we rely on the watcher.Close() to clean up.
}

// restart restarts the watcher service
func (dw *DirectoryWatcher) restart() error {
	// Stop current watcher
	dw.watcher.Close()

	// Create new watcher
	newWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create new watcher: %w", err)
	}

	dw.watcher = newWatcher

	// Re-add all watched paths
	for path, _ := range dw.watchedPaths {
		if err := dw.watcher.Add(path); err != nil {
			fmt.Printf("Warning: failed to re-add watch for %s: %v\n", path, err)
		}
	}

	return nil
}

// generateSessionID generates a unique session ID for watching
func generateSessionID() string {
	return fmt.Sprintf("fs-watch-%d", time.Now().UnixNano())
}

// GetStats returns statistics about the watcher
func (dw *DirectoryWatcher) GetStats() map[string]interface{} {
	dw.mutex.RLock()
	defer dw.mutex.RUnlock()

	stats := map[string]interface{}{
		"is_running":        dw.isRunning,
		"watched_paths":     len(dw.watchedPaths),
		"event_buffer_size": len(dw.eventChan),
		"max_buffer_size":   cap(dw.eventChan),
		"config":           dw.config,
		"debouncing":       dw.config.DebounceMs > 0,
	}

	// Add debouncer statistics if available
	if dw.debouncer != nil {
		stats["debouncer_stats"] = dw.debouncer.GetStats()
	}

	// Add per-path statistics
	pathStats := make(map[string]interface{})
	for path, ctx := range dw.watchedPaths {
		pathStats[path] = map[string]interface{}{
			"session_id":  ctx.SessionID,
			"event_count": ctx.EventCount,
			"last_event":  ctx.LastEvent,
			"recursive":   ctx.Recursive,
		}
	}
	stats["paths"] = pathStats

	return stats
}

// Cleanup performs cleanup of resources
func (dw *DirectoryWatcher) Cleanup() error {
	if err := dw.Stop(); err != nil {
		return fmt.Errorf("failed to stop watcher: %w", err)
	}

	// Close event channel
	close(dw.eventChan)

	return nil
}