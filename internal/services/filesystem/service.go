package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"
)

// FileSystemService is the main service that provides file system operations
type FileSystemService struct {
	eventBus         events.IEventBus
	watcher          *DirectoryWatcher
	navigator        *FileNavigator
	config           *FileSystemServiceConfig
	mutex            sync.RWMutex
	isInitialized    bool
	watchedDirs      map[string]bool
	sessions         map[string]*ServiceSession
	operationQueue   chan *FileOperation
	stats            *FileSystemServiceStats
}

// FileSystemServiceConfig contains configuration for the file system service
type FileSystemServiceConfig struct {
	// General settings
	Enabled               bool          `json:"enabled"`
	MaxConcurrentOps      int           `json:"max_concurrent_ops"`
	DefaultPermissions    os.FileMode   `json:"default_permissions"`
	TempDirPattern        string        `json:"temp_dir_pattern"`

	// Directory watching
	WatchingEnabled       bool          `json:"watching_enabled"`
	DefaultWatchConfig    *models.DirectoryWatcherConfig `json:"default_watch_config"`

	// Navigation
	NavigationEnabled     bool          `json:"navigation_enabled"`
	DefaultNavigationConfig *NavigationConfig `json:"default_navigation_config"`

	// File operations
	CopyEnabled           bool          `json:"copy_enabled"`
	MoveEnabled           bool          `json:"move_enabled"`
	DeleteEnabled         bool          `json:"delete_enabled"`
	RenameEnabled         bool          `json:"rename_enabled"`
	CreateDirEnabled      bool          `json:"create_dir_enabled"`
	CreateFileEnabled     bool          `json:"create_file_enabled"`

	// Performance settings
	CacheEnabled          bool          `json:"cache_enabled"`
	CacheTimeout          time.Duration `json:"cache_timeout"`
	MaxCacheSize          int           `json:"max_cache_size"`
	ChunkSize             int64         `json:"chunk_size"`

	// Security settings
	AllowSymlinkFollowing bool          `json:"allow_symlink_following"`
	MaxFileSize           int64         `json:"max_file_size"`
	AllowedExtensions     []string      `json:"allowed_extensions"`
	BlockedExtensions     []string      `json:"blocked_extensions"`
}

// ServiceSession represents a file system service session
type ServiceSession struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	LastActivity  time.Time `json:"last_activity"`
	UserID        string    `json:"user_id,omitempty"`
	WorkingDir    string    `json:"working_dir"`
	Preferences   *SessionPreferences `json:"preferences"`
	OperationCount int      `json:"operation_count"`
}

// SessionPreferences contains session-specific preferences
type SessionPreferences struct {
	ShowHidden      bool   `json:"show_hidden"`
	AutoFollowSymlinks bool `json:"auto_follow_symlinks"`
	ConfirmDeletes  bool   `json:"confirm_deletes"`
	BackupOnMove    bool   `json:"backup_on_move"`
}

// FileOperation represents a queued file system operation
type FileOperation struct {
	ID          string                   `json:"id"`
	Type        models.FileOperation     `json:"type"`
	SessionID   string                   `json:"session_id"`
	Source      string                   `json:"source"`
	Destination string                   `json:"destination,omitempty"`
	Options     map[string]interface{}   `json:"options,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	StartedAt   *time.Time               `json:"started_at,omitempty"`
	CompletedAt *time.Time               `json:"completed_at,omitempty"`
	Status      OperationStatus          `json:"status"`
	Result      *models.FileOperationResult `json:"result,omitempty"`
	Error       string                   `json:"error,omitempty"`
	Progress    float64                  `json:"progress"`
	Context     context.Context          `json:"-"`
}

// OperationStatus represents the status of a file operation
type OperationStatus string

const (
	StatusQueued     OperationStatus = "queued"
	StatusRunning    OperationStatus = "running"
	StatusCompleted  OperationStatus = "completed"
	StatusFailed     OperationStatus = "failed"
	StatusCancelled  OperationStatus = "cancelled"
)

// FileSystemServiceStats contains statistics about the file system service
type FileSystemServiceStats struct {
	SessionsActive       int                           `json:"sessions_active"`
	SessionsTotal        int                           `json:"sessions_total"`
	OperationsQueued     int                           `json:"operations_queued"`
	OperationsRunning    int                           `json:"operations_running"`
	OperationsCompleted  int                           `json:"operations_completed"`
	OperationsFailed     int                           `json:"operations_failed"`
	WatchedDirectories   int                           `json:"watched_directories"`
	TotalFilesAccessed   int64                         `json:"total_files_accessed"`
	TotalBytesTransferred int64                        `json:"total_bytes_transferred"`
	AverageOperationTime time.Duration                 `json:"average_operation_time"`
	StartTime            time.Time                     `json:"start_time"`
	LastActivity         time.Time                     `json:"last_activity"`
	OperationStats       map[string]*OperationStats    `json:"operation_stats"`
}

// OperationStats contains statistics for a specific operation type
type OperationStats struct {
	Count        int           `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AverageTime  time.Duration `json:"average_time"`
	SuccessRate  float64       `json:"success_rate"`
	LastExecuted time.Time     `json:"last_executed"`
}

// NewFileSystemService creates a new file system service instance
func NewFileSystemService(eventBus events.IEventBus) (*FileSystemService, error) {
	// Create default configuration
	config := &FileSystemServiceConfig{
		Enabled:            true,
		MaxConcurrentOps:   10,
		DefaultPermissions: 0644,
		TempDirPattern:     "dex-ui-*",

		WatchingEnabled:    true,
		DefaultWatchConfig: &models.DirectoryWatcherConfig{
			Enabled:       false,
			WatchHidden:   false,
			DebounceMs:    100,
			MaxEvents:     1000,
			BufferSize:    512,
			IgnorePatterns: []string{"*.tmp", "*.swp", ".git/*"},
		},

		NavigationEnabled: true,
		DefaultNavigationConfig: &NavigationConfig{
			MaxHistoryEntries:    100,
			ShowHiddenFiles:      false,
			BookmarkFile:         ".file_bookmarks.json",
			AutoSaveBookmarks:    true,
			SortingMethod:        SortByName,
			GroupDirectoriesFirst: true,
			CaseSensitiveSort:    false,
			NaturalSorting:       true,
			CacheEnabled:         true,
			CacheTimeout:         5 * time.Minute,
		},

		CopyEnabled:       true,
		MoveEnabled:       true,
		DeleteEnabled:     true,
		RenameEnabled:     true,
		CreateDirEnabled:  true,
		CreateFileEnabled: true,

		CacheEnabled:          true,
		CacheTimeout:          10 * time.Minute,
		MaxCacheSize:          1000,
		ChunkSize:             64 * 1024, // 64KB

		AllowSymlinkFollowing: false,
		MaxFileSize:           1024 * 1024 * 1024, // 1GB
		AllowedExtensions:     []string{},
		BlockedExtensions:     []string{".exe", ".bat", ".cmd", ".scr"},
	}

	// Create watcher
	watcher, err := NewDirectoryWatcher(eventBus)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory watcher: %w", err)
	}

	// Create navigator
	navigator := NewFileNavigator(eventBus)

	// Create service
	service := &FileSystemService{
		eventBus:       eventBus,
		watcher:        watcher,
		navigator:      navigator,
		config:         config,
		watchedDirs:    make(map[string]bool),
		sessions:       make(map[string]*ServiceSession),
		operationQueue: make(chan *FileOperation, 100),
		stats: &FileSystemServiceStats{
			StartTime:      time.Now(),
			LastActivity:   time.Now(),
			OperationStats: make(map[string]*OperationStats),
		},
	}

	return service, nil
}

// Initialize initializes the file system service
func (fss *FileSystemService) Initialize(ctx context.Context) error {
	fss.mutex.Lock()
	defer fss.mutex.Unlock()

	if fss.isInitialized {
		return fmt.Errorf("file system service is already initialized")
	}

	// Start directory watcher if enabled
	if fss.config.WatchingEnabled && fss.config.Enabled {
		if err := fss.watcher.Start(); err != nil {
			return fmt.Errorf("failed to start directory watcher: %w", err)
		}
	}

	// Start operation processor
	go fss.processOperations(ctx)

	// Update stats
	fss.stats.StartTime = time.Now()
	fss.stats.LastActivity = time.Now()
	fss.isInitialized = true

	// Publish initialization event
	if fss.eventBus != nil {
		fss.eventBus.Publish(ctx, "filesystem.service.initialized", map[string]interface{}{
			"config": fss.config,
		}, "filesystem-service")
	}

	return nil
}

// Shutdown shuts down the file system service
func (fss *FileSystemService) Shutdown(ctx context.Context) error {
	fss.mutex.Lock()
	defer fss.mutex.Unlock()

	if !fss.isInitialized {
		return nil
	}

	// Stop watcher
	if err := fss.watcher.Stop(); err != nil {
		return fmt.Errorf("failed to stop directory watcher: %w", err)
	}

	// Close operation queue
	close(fss.operationQueue)

	// Cleanup navigator
	if err := fss.navigator.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup navigator: %w", err)
	}

	// Cleanup watcher
	if err := fss.watcher.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup watcher: %w", err)
	}

	fss.isInitialized = false

	// Publish shutdown event
	if fss.eventBus != nil {
		fss.eventBus.Publish(ctx, "filesystem.service.shutdown", map[string]interface{}{
			"stats": fss.stats,
		}, "filesystem-service")
	}

	return nil
}

// IsInitialized returns whether the service is initialized
func (fss *FileSystemService) IsInitialized() bool {
	fss.mutex.RLock()
	defer fss.mutex.RUnlock()
	return fss.isInitialized
}

// SetConfig updates the service configuration
func (fss *FileSystemService) SetConfig(config *FileSystemServiceConfig) error {
	fss.mutex.Lock()
	defer fss.mutex.Unlock()

	fss.config = config

	// Update watcher config
	if fss.watcher != nil {
		watcherConfig := config.DefaultWatchConfig
		if watcherConfig == nil {
			watcherConfig = &models.DirectoryWatcherConfig{
				Enabled:    false,
				WatchHidden: false,
				DebounceMs: 100,
			}
		}
		if err := fss.watcher.SetConfig(watcherConfig); err != nil {
			return fmt.Errorf("failed to update watcher config: %w", err)
		}
	}

	return nil
}

// GetConfig returns the current service configuration
func (fss *FileSystemService) GetConfig() *FileSystemServiceConfig {
	fss.mutex.RLock()
	defer fss.mutex.RUnlock()

	// Return a copy
	configCopy := *fss.config
	return &configCopy
}

// CreateSession creates a new service session
func (fss *FileSystemService) CreateSession(sessionID string, userID string) (*ServiceSession, error) {
	fss.mutex.Lock()
	defer fss.mutex.Unlock()

	if _, exists := fss.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = os.Getenv("HOME")
		if workingDir == "" {
			workingDir = "/"
		}
	}

	// Create session preferences
	preferences := &SessionPreferences{
		ShowHidden:        false,
		AutoFollowSymlinks: false,
		ConfirmDeletes:    true,
		BackupOnMove:      false,
	}

	// Create session
	session := &ServiceSession{
		ID:            sessionID,
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
		UserID:        userID,
		WorkingDir:    workingDir,
		Preferences:   preferences,
		OperationCount: 0,
	}

	fss.sessions[sessionID] = session
	fss.stats.SessionsActive++
	fss.stats.SessionsTotal++

	return session, nil
}

// GetSession returns a service session
func (fss *FileSystemService) GetSession(sessionID string) (*ServiceSession, error) {
	fss.mutex.RLock()
	defer fss.mutex.RUnlock()

	session, exists := fss.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Return a copy
	sessionCopy := *session
	return &sessionCopy, nil
}

// DeleteSession deletes a service session
func (fss *FileSystemService) DeleteSession(sessionID string) error {
	fss.mutex.Lock()
	defer fss.mutex.Unlock()

	session, exists := fss.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(fss.sessions, sessionID)
	fss.stats.SessionsActive--

	// Publish session deleted event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.session.deleted", map[string]interface{}{
			"session_id": sessionID,
			"operation_count": session.OperationCount,
		}, "filesystem-service")
	}

	return nil
}

// ListDirectory lists the contents of a directory
func (fss *FileSystemService) ListDirectory(path string) ([]*models.FileSystemEntry, error) {
	if !fss.config.Enabled {
		return nil, models.NewFileSystemError(models.ErrorConfiguration, path, "list",
			"file system service is disabled", nil)
	}

	// Validate and clean path
	cleanPath := filepath.Clean(path)
	if cleanPath == "" {
		return nil, models.NewFileSystemError(models.ErrorInvalidPath, path, "list",
			"invalid path: empty", nil)
	}

	// Create entry for validation
	entry := &models.FileSystemEntry{
		Name: filepath.Base(cleanPath),
		Path: cleanPath,
		IsDir: true,
	}

	// Validate path and check permissions
	if err := entry.ValidateForOperation("list"); err != nil {
		return nil, err
	}

	// Check if path exists and is a directory
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, models.WrapFileSystemError(err, cleanPath, "stat")
	}
	if !info.IsDir() {
		return nil, models.NewFileSystemError(models.ErrorNotDirectory, cleanPath, "list",
			"path is not a directory", nil)
	}

	// Read directory
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, models.WrapFileSystemError(err, cleanPath, "readdir")
	}

	// Convert to FileSystemEntry models
	fileEntries := make([]*models.FileSystemEntry, 0, len(entries))
	for _, dirEntry := range entries {
		fullPath := filepath.Join(cleanPath, dirEntry.Name())

		// Create file system entry
		fileEntry, err := models.NewFileSystemEntry(fullPath)
		if err != nil {
			// Create minimal entry if full creation fails
			info, _ := dirEntry.Info()
			fileEntry = &models.FileSystemEntry{
				Name:  dirEntry.Name(),
				Path:  fullPath,
				IsDir: dirEntry.IsDir(),
				Mode:  info.Mode(),
				Size:  info.Size(),
			}
		}

		fileEntries = append(fileEntries, fileEntry)
	}

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.directory.listed", map[string]interface{}{
			"path":  cleanPath,
			"count": len(fileEntries),
		}, "filesystem-service")
	}

	return fileEntries, nil
}

// GetFileInfo returns information about a file or directory
func (fss *FileSystemService) GetFileInfo(path string) (*models.FileSystemEntry, error) {
	if !fss.config.Enabled {
		return nil, models.NewFileSystemError(models.ErrorConfiguration, path, "stat",
			"file system service is disabled", nil)
	}

	// Validate and clean path
	cleanPath := filepath.Clean(path)
	if cleanPath == "" {
		return nil, models.NewFileSystemError(models.ErrorInvalidPath, path, "stat",
			"invalid path: empty", nil)
	}

	// Get file info
	entry, err := models.NewFileSystemEntry(cleanPath)
	if err != nil {
		return nil, models.WrapFileSystemError(err, cleanPath, "stat")
	}

	// Update stats
	fss.updateStats()

	return entry, nil
}

// CreateDirectory creates a new directory
func (fss *FileSystemService) CreateDirectory(path string) error {
	if !fss.config.Enabled || !fss.config.CreateDirEnabled {
		return models.NewFileSystemError(models.ErrorConfiguration, path, "create",
			"directory creation is disabled", nil)
	}

	// Validate and clean path
	cleanPath := filepath.Clean(path)
	if cleanPath == "" {
		return models.NewFileSystemError(models.ErrorInvalidPath, path, "create",
			"invalid path: empty", nil)
	}

	// Validate path and check permissions for parent directory
	parentPath := filepath.Dir(cleanPath)
	parentEntry := &models.FileSystemEntry{
		Name: filepath.Base(parentPath),
		Path: parentPath,
		IsDir: true,
	}

	if err := parentEntry.ValidateForOperation("create"); err != nil {
		return models.WrapFileSystemError(err, cleanPath, "create_parent_validation")
	}

	// Check if directory already exists
	if _, err := os.Stat(cleanPath); err == nil {
		return models.NewFileSystemError(models.ErrorExists, cleanPath, "create",
			"directory already exists", nil)
	}

	// Create directory
	err := os.MkdirAll(cleanPath, 0755)
	if err != nil {
		return models.WrapFileSystemError(err, cleanPath, "mkdir")
	}

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.directory.created", map[string]interface{}{
			"path": cleanPath,
		}, "filesystem-service")
	}

	return nil
}

// DeleteFile deletes a file or directory
func (fss *FileSystemService) DeleteFile(path string) error {
	if !fss.config.Enabled || !fss.config.DeleteEnabled {
		return models.NewFileSystemError(models.ErrorConfiguration, path, "delete",
			"file deletion is disabled", nil)
	}

	// Validate and clean path
	cleanPath := filepath.Clean(path)
	if cleanPath == "" {
		return models.NewFileSystemError(models.ErrorInvalidPath, path, "delete",
			"invalid path: empty", nil)
	}

	// Create entry for validation
	entry, err := models.NewFileSystemEntry(cleanPath)
	if err != nil {
		return models.WrapFileSystemError(err, cleanPath, "stat")
	}

	// Validate path and check permissions
	if err := entry.ValidateForOperation("delete"); err != nil {
		return err
	}

	// Check if path exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		return models.WrapFileSystemError(err, cleanPath, "stat")
	}

	// Additional check for directory deletion
	if info.IsDir() {
		// Check if directory is empty (optional safety check)
		entries, readErr := os.ReadDir(cleanPath)
		if readErr == nil && len(entries) > 0 {
			// Directory is not empty - this is a safety measure
			return models.NewFileSystemError(models.ErrorDirectoryFull, cleanPath, "delete",
				"directory is not empty", nil)
		}
	}

	// Delete file or directory
	if info.IsDir() {
		err = os.RemoveAll(cleanPath)
	} else {
		err = os.Remove(cleanPath)
	}
	if err != nil {
		return models.WrapFileSystemError(err, cleanPath, "delete")
	}

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.file.deleted", map[string]interface{}{
			"path": cleanPath,
			"is_directory": info.IsDir(),
		}, "filesystem-service")
	}

	return nil
}

// RenameFile renames a file or directory
func (fss *FileSystemService) RenameFile(oldPath, newPath string) error {
	if !fss.config.Enabled || !fss.config.RenameEnabled {
		return fmt.Errorf("file renaming is disabled")
	}

	// Validate and clean paths
	cleanOldPath := filepath.Clean(oldPath)
	cleanNewPath := filepath.Clean(newPath)
	if cleanOldPath == "" || cleanNewPath == "" {
		return fmt.Errorf("invalid path: empty")
	}

	// Rename file
	err := os.Rename(cleanOldPath, cleanNewPath)
	if err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.file.renamed", map[string]interface{}{
			"old_path": cleanOldPath,
			"new_path": cleanNewPath,
		}, "filesystem-service")
	}

	return nil
}

// GetWorkingDirectory returns the current working directory
func (fss *FileSystemService) GetWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}

// SetWorkingDirectory changes the current working directory
func (fss *FileSystemService) SetWorkingDirectory(path string) error {
	// Validate and clean path
	cleanPath := filepath.Clean(path)
	if cleanPath == "" {
		return fmt.Errorf("invalid path: empty")
	}

	// Change directory
	err := os.Chdir(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.working_directory.changed", map[string]interface{}{
			"new_path": cleanPath,
		}, "filesystem-service")
	}

	return nil
}

// WatchDirectory starts watching a directory for changes
func (fss *FileSystemService) WatchDirectory(path string) error {
	if !fss.config.Enabled || !fss.config.WatchingEnabled {
		return fmt.Errorf("directory watching is disabled")
	}

	sessionID, err := fss.watcher.WatchDirectory(path, true)
	if err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Track watched directory
	fss.watchedDirs[path] = true

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.directory.watched", map[string]interface{}{
			"path":       path,
			"session_id": sessionID,
		}, "filesystem-service")
	}

	return nil
}

// StopWatching stops watching a directory
func (fss *FileSystemService) StopWatching(path string) error {
	err := fss.watcher.StopWatching(path)
	if err != nil {
		return fmt.Errorf("failed to stop watching directory: %w", err)
	}

	// Remove from tracked directories
	delete(fss.watchedDirs, path)

	// Update stats
	fss.updateStats()

	// Publish event
	if fss.eventBus != nil {
		ctx := context.Background()
		fss.eventBus.Publish(ctx, "filesystem.directory.unwatched", map[string]interface{}{
			"path": path,
		}, "filesystem-service")
	}

	return nil
}

// GetWatchedDirectories returns a list of watched directories
func (fss *FileSystemService) GetWatchedDirectories() []string {
	dirs := fss.watcher.GetWatchedDirectories()

	// Update stats
	fss.stats.WatchedDirectories = len(dirs)

	return dirs
}

// GetStats returns service statistics
func (fss *FileSystemService) GetStats() *FileSystemServiceStats {
	fss.mutex.RLock()
	defer fss.mutex.RUnlock()

	// Update dynamic stats
	fss.stats.SessionsActive = len(fss.sessions)
	fss.stats.WatchedDirectories = len(fss.watchedDirs)
	fss.stats.OperationsQueued = len(fss.operationQueue)

	// Return a copy
	statsCopy := *fss.stats
	return &statsCopy
}

// processOperations processes queued file operations
func (fss *FileSystemService) processOperations(ctx context.Context) {
	for {
		select {
		case operation := <-fss.operationQueue:
			if operation != nil {
				fss.executeOperation(ctx, operation)
			}
		case <-ctx.Done():
			return
		}
	}
}

// executeOperation executes a single file operation
func (fss *FileSystemService) executeOperation(ctx context.Context, operation *FileOperation) {
	// Update operation status
	now := time.Now()
	operation.StartedAt = &now
	operation.Status = StatusRunning

	// Update stats
	fss.stats.OperationsRunning++
	fss.stats.OperationsQueued--

	// Execute operation based on type
	var err error
	switch operation.Type {
	case models.OperationCopy:
		err = fss.executeCopy(ctx, operation)
	case models.OperationMove:
		err = fss.executeMove(ctx, operation)
	case models.OperationDelete:
		err = fss.executeDelete(ctx, operation)
	default:
		err = fmt.Errorf("unsupported operation type: %s", operation.Type)
	}

	// Update operation result
	completedAt := time.Now()
	operation.CompletedAt = &completedAt

	if err != nil {
		operation.Status = StatusFailed
		operation.Error = err.Error()
		fss.stats.OperationsFailed++
	} else {
		operation.Status = StatusCompleted
		fss.stats.OperationsCompleted++
	}

	fss.stats.OperationsRunning--
	fss.updateOperationStats(operation)
}

// executeCopy executes a copy operation
func (fss *FileSystemService) executeCopy(ctx context.Context, operation *FileOperation) error {
	if !fss.config.CopyEnabled {
		return fmt.Errorf("copy operation is disabled")
	}

	src := filepath.Clean(operation.Source)
	dst := filepath.Clean(operation.Destination)

	if src == "" || dst == "" {
		return fmt.Errorf("invalid source or destination path")
	}

	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", src, err)
	}

	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if srcInfo.IsDir() {
		return fss.copyDir(ctx, src, dst, operation)
	}
	return fss.copyFile(ctx, src, dst, srcInfo.Mode(), operation)
}

// copyFile copies a single file from src to dst
func (fss *FileSystemService) copyFile(ctx context.Context, src, dst string, mode os.FileMode, operation *FileOperation) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Get file size for progress tracking
	srcInfo, _ := srcFile.Stat()
	totalSize := srcInfo.Size()
	var copied int64

	// Copy in chunks with progress tracking
	buf := make([]byte, fss.config.ChunkSize)
	if fss.config.ChunkSize == 0 {
		buf = make([]byte, 32*1024) // Default 32KB chunks
	}

	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := srcFile.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read source file: %w", err)
		}
		if n == 0 {
			break
		}

		if _, err := dstFile.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write to destination file: %w", err)
		}

		copied += int64(n)
		if totalSize > 0 && operation != nil {
			operation.Progress = float64(copied) / float64(totalSize)
		}
	}

	return nil
}

// copyDir recursively copies a directory from src to dst
func (fss *FileSystemService) copyDir(ctx context.Context, src, dst string, operation *FileOperation) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := fss.copyDir(ctx, srcPath, dstPath, operation); err != nil {
				return err
			}
		} else {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}
			if err := fss.copyFile(ctx, srcPath, dstPath, info.Mode(), operation); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeMove executes a move operation
func (fss *FileSystemService) executeMove(ctx context.Context, operation *FileOperation) error {
	if !fss.config.MoveEnabled {
		return fmt.Errorf("move operation is disabled")
	}

	src := filepath.Clean(operation.Source)
	dst := filepath.Clean(operation.Destination)

	if src == "" || dst == "" {
		return fmt.Errorf("invalid source or destination path")
	}

	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Try atomic rename first (works on same filesystem)
	err := os.Rename(src, dst)
	if err == nil {
		operation.Progress = 1.0
		return nil
	}

	// Check if it's a cross-device error - if so, copy then delete
	// This handles moving files across different filesystems/partitions
	if linkErr, ok := err.(*os.LinkError); ok {
		// Cross-device link error typically indicates different filesystems
		if linkErr.Err.Error() == "invalid cross-device link" ||
		   linkErr.Err.Error() == "cross-device link" {
			// Fall back to copy + delete
			if err := fss.executeCopy(ctx, operation); err != nil {
				return fmt.Errorf("failed to copy during cross-device move: %w", err)
			}

			// Delete source after successful copy
			srcInfo, err := os.Stat(src)
			if err != nil {
				return fmt.Errorf("failed to stat source after copy: %w", err)
			}

			if srcInfo.IsDir() {
				if err := os.RemoveAll(src); err != nil {
					return fmt.Errorf("failed to remove source directory after move: %w", err)
				}
			} else {
				if err := os.Remove(src); err != nil {
					return fmt.Errorf("failed to remove source file after move: %w", err)
				}
			}

			return nil
		}
	}

	// Return the original rename error for other cases
	return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
}

// executeDelete executes a delete operation
func (fss *FileSystemService) executeDelete(ctx context.Context, operation *FileOperation) error {
	return fss.DeleteFile(operation.Source)
}

// updateStats updates service statistics
func (fss *FileSystemService) updateStats() {
	fss.stats.LastActivity = time.Now()
}

// updateOperationStats updates operation statistics
func (fss *FileSystemService) updateOperationStats(operation *FileOperation) {
	opType := string(operation.Type)
	stats, exists := fss.stats.OperationStats[opType]
	if !exists {
		stats = &OperationStats{}
		fss.stats.OperationStats[opType] = stats
	}

	stats.Count++
	stats.LastExecuted = *operation.CompletedAt

	if operation.StartedAt != nil && operation.CompletedAt != nil {
		duration := operation.CompletedAt.Sub(*operation.StartedAt)
		stats.TotalTime += duration
		stats.AverageTime = stats.TotalTime / time.Duration(stats.Count)
	}

	if operation.Status == StatusCompleted {
		successCount := 0
		if stats.SuccessRate > 0 {
			successCount = int(stats.SuccessRate * float64(stats.Count-1))
		}
		successCount++
		stats.SuccessRate = float64(successCount) / float64(stats.Count)
	}
}