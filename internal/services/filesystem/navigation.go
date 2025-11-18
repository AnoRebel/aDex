package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"
)

// FileNavigator provides file system navigation capabilities
type FileNavigator struct {
	eventBus     events.IEventBus
	history      map[string]*models.FileNavigationHistory
	mutex        sync.RWMutex
	config       *NavigationConfig
	bookmarkPath string
	bookmarks    map[string]*NavigationBookmark
	cache        map[string]*NavigationCacheEntry
	preferences  map[string]*NavigationPreferences
}

// BookmarkStore represents the persistent storage format for bookmarks
type BookmarkStore struct {
	Version   string                          `json:"version"`
	UpdatedAt time.Time                       `json:"updated_at"`
	Bookmarks map[string]*NavigationBookmark  `json:"bookmarks"`
}

// NavigationConfig contains configuration for file navigation
type NavigationConfig struct {
	MaxHistoryEntries    int           `json:"max_history_entries"`
	ShowHiddenFiles      bool          `json:"show_hidden_files"`
	BookmarkFile         string        `json:"bookmark_file"`
	AutoSaveBookmarks    bool          `json:"auto_save_bookmarks"`
	SortingMethod        SortingMethod `json:"sorting_method"`
	GroupDirectoriesFirst bool         `json:"group_directories_first"`
	CaseSensitiveSort    bool         `json:"case_sensitive_sort"`
	NaturalSorting       bool         `json:"natural_sorting"`
	CacheEnabled         bool         `json:"cache_enabled"`
	CacheTimeout         time.Duration `json:"cache_timeout"`
}

// SortingMethod defines how files and directories are sorted
type SortingMethod string

const (
	SortByName        SortingMethod = "name"
	SortBySize        SortingMethod = "size"
	SortByModified    SortingMethod = "modified"
	SortByType        SortingMethod = "type"
	SortByCreated     SortingMethod = "created"
	SortByAccessed    SortingMethod = "accessed"
)

// NavigationSession represents a navigation session
type NavigationSession struct {
	ID            string                           `json:"id"`
	CurrentPath   string                           `json:"current_path"`
	History       *models.FileNavigationHistory    `json:"history"`
	Bookmarks     map[string]*NavigationBookmark   `json:"bookmarks"`
	Preferences   *NavigationPreferences           `json:"preferences"`
	LastActivity  time.Time                        `json:"last_activity"`
	Cache         map[string]*NavigationCacheEntry `json:"cache,omitempty"`
}

// NavigationBookmark represents a bookmarked location
type NavigationBookmark struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
	UseCount    int       `json:"use_count"`
	Tags        []string  `json:"tags,omitempty"`
}

// NavigationPreferences contains user preferences for navigation
type NavigationPreferences struct {
	ShowHidden       bool           `json:"show_hidden"`
	SortingMethod    SortingMethod  `json:"sorting_method"`
	GroupDirsFirst   bool           `json:"group_dirs_first"`
	CaseSensitive    bool           `json:"case_sensitive"`
	NaturalSort      bool           `json:"natural_sort"`
	ViewMode         ViewMode       `json:"view_mode"`
	ItemsPerPage     int            `json:"items_per_page"`
	PreviewEnabled   bool           `json:"preview_enabled"`
	PreviewSize      int64          `json:"preview_size"`
}

// ViewMode defines how files are displayed
type ViewMode string

const (
	ViewModeList     ViewMode = "list"
	ViewModeGrid     ViewMode = "grid"
	ViewModeTree     ViewMode = "tree"
	ViewModeDetails  ViewMode = "details"
)

// NavigationCacheEntry represents a cached directory listing
type NavigationCacheEntry struct {
	Path      string                   `json:"path"`
	Entries   []*models.FileSystemEntry `json:"entries"`
	Timestamp time.Time                `json:"timestamp"`
	Checksum  string                   `json:"checksum"`
}

// NavigationResult contains the result of a navigation operation
type NavigationResult struct {
	Path         string                   `json:"path"`
	Entries      []*models.FileSystemEntry `json:"entries"`
	ParentPath   string                   `json:"parent_path,omitempty"`
	HasParent    bool                     `json:"has_parent"`
	SessionID    string                   `json:"session_id"`
	CacheHit     bool                     `json:"cache_hit"`
	TotalCount   int                      `json:"total_count"`
	Directories  int                      `json:"directories"`
	Files        int                      `json:"files"`
	HiddenCount  int                      `json:"hidden_count"`
}

// NewFileNavigator creates a new file navigator instance
func NewFileNavigator(eventBus events.IEventBus) *FileNavigator {
	config := &NavigationConfig{
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
	}

	nav := &FileNavigator{
		eventBus:     eventBus,
		history:      make(map[string]*models.FileNavigationHistory),
		config:       config,
		bookmarkPath: filepath.Join(os.Getenv("HOME"), config.BookmarkFile),
		bookmarks:    make(map[string]*NavigationBookmark),
		cache:        make(map[string]*NavigationCacheEntry),
		preferences:  make(map[string]*NavigationPreferences),
	}

	// Load bookmarks if they exist
	if config.AutoSaveBookmarks {
		if err := nav.loadBookmarks(); err != nil {
			// Log error but continue - bookmarks are not critical
			if nav.eventBus != nil {
				ctx := context.Background()
				nav.eventBus.Publish(ctx, "navigation.bookmark_load_error", map[string]interface{}{
					"error": err.Error(),
				}, "file-navigator")
			}
		}
	}

	return nav
}

// CreateSession creates a new navigation session
func (fn *FileNavigator) CreateSession(sessionID string, startPath string) (*NavigationSession, error) {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	// Clean and validate start path
	cleanPath, err := fn.validatePath(startPath)
	if err != nil {
		return nil, fmt.Errorf("invalid start path: %w", err)
	}

	// Create navigation history
	history := &models.FileNavigationHistory{
		Current:    cleanPath,
		History:    []string{cleanPath},
		MaxHistory: fn.config.MaxHistoryEntries,
		Index:      0,
	}

	// Create session preferences
	preferences := &NavigationPreferences{
		ShowHidden:     fn.config.ShowHiddenFiles,
		SortingMethod:  fn.config.SortingMethod,
		GroupDirsFirst: fn.config.GroupDirectoriesFirst,
		CaseSensitive:  fn.config.CaseSensitiveSort,
		NaturalSort:    fn.config.NaturalSorting,
		ViewMode:       ViewModeList,
		ItemsPerPage:   100,
		PreviewEnabled: true,
		PreviewSize:    1024 * 1024, // 1MB
	}

	// Create session
	session := &NavigationSession{
		ID:           sessionID,
		CurrentPath:  cleanPath,
		History:      history,
		Bookmarks:    make(map[string]*NavigationBookmark),
		Preferences:  preferences,
		LastActivity: time.Now(),
		Cache:        make(map[string]*NavigationCacheEntry),
	}

	// Store history for session
	fn.history[sessionID] = history

	// Publish session created event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.session_created", map[string]interface{}{
			"session_id":   sessionID,
			"current_path": cleanPath,
		}, "file-navigator")
	}

	return session, nil
}

// NavigateTo navigates to a specific directory
func (fn *FileNavigator) NavigateTo(sessionID string, path string) (*NavigationResult, error) {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	// Get session history
	history, exists := fn.history[sessionID]
	if !exists {
		return nil, fmt.Errorf("navigation session %s not found", sessionID)
	}

	// Clean and validate path
	cleanPath, err := fn.validatePath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid navigation path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	// Check if it's a directory
	if info, err := os.Stat(cleanPath); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", cleanPath)
	}

	// Update history
	fn.updateHistory(history, cleanPath)

	// Get directory entries
	entries, err := fn.getDirectoryEntries(cleanPath, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Create navigation result
	result := fn.createNavigationResult(cleanPath, entries, sessionID)

	// Publish navigation event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.navigated", map[string]interface{}{
			"session_id":   sessionID,
			"path":         cleanPath,
			"entry_count":  len(entries),
			"parent_path":  result.ParentPath,
		}, "file-navigator")
	}

	return result, nil
}

// NavigateUp navigates to the parent directory
func (fn *FileNavigator) NavigateUp(sessionID string) (*NavigationResult, error) {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	history, exists := fn.history[sessionID]
	if !exists {
		return nil, fmt.Errorf("navigation session %s not found", sessionID)
	}

	currentPath := history.Current
	parentPath := filepath.Dir(currentPath)

	// Check if we're already at root
	if parentPath == currentPath {
		return nil, fmt.Errorf("already at root directory")
	}

	return fn.NavigateTo(sessionID, parentPath)
}

// NavigateBack navigates to the previous directory in history
func (fn *FileNavigator) NavigateBack(sessionID string) (*NavigationResult, error) {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	history, exists := fn.history[sessionID]
	if !exists {
		return nil, fmt.Errorf("navigation session %s not found", sessionID)
	}

	// Check if we can go back
	if history.Index <= 0 {
		return nil, fmt.Errorf("no previous directory in history")
	}

	// Move back in history
	history.Index--
	previousPath := history.History[history.Index]
	history.Current = previousPath

	// Get directory entries
	entries, err := fn.getDirectoryEntries(previousPath, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Create navigation result
	result := fn.createNavigationResult(previousPath, entries, sessionID)

	// Publish navigation event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.navigated_back", map[string]interface{}{
			"session_id":   sessionID,
			"path":         previousPath,
			"history_index": history.Index,
		}, "file-navigator")
	}

	return result, nil
}

// NavigateForward navigates to the next directory in history
func (fn *FileNavigator) NavigateForward(sessionID string) (*NavigationResult, error) {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	history, exists := fn.history[sessionID]
	if !exists {
		return nil, fmt.Errorf("navigation session %s not found", sessionID)
	}

	// Check if we can go forward
	if history.Index >= len(history.History)-1 {
		return nil, fmt.Errorf("no next directory in history")
	}

	// Move forward in history
	history.Index++
	nextPath := history.History[history.Index]
	history.Current = nextPath

	// Get directory entries
	entries, err := fn.getDirectoryEntries(nextPath, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Create navigation result
	result := fn.createNavigationResult(nextPath, entries, sessionID)

	// Publish navigation event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.navigated_forward", map[string]interface{}{
			"session_id":    sessionID,
			"path":          nextPath,
			"history_index": history.Index,
		}, "file-navigator")
	}

	return result, nil
}

// GetHistory returns the navigation history for a session
func (fn *FileNavigator) GetHistory(sessionID string) (*models.FileNavigationHistory, error) {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	history, exists := fn.history[sessionID]
	if !exists {
		return nil, fmt.Errorf("navigation session %s not found", sessionID)
	}

	// Return a copy
	historyCopy := *history
	historyCopy.History = make([]string, len(history.History))
	copy(historyCopy.History, history.History)

	return &historyCopy, nil
}

// ClearHistory clears the navigation history for a session
func (fn *FileNavigator) ClearHistory(sessionID string) error {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	history, exists := fn.history[sessionID]
	if !exists {
		return fmt.Errorf("navigation session %s not found", sessionID)
	}

	// Reset history to current path only
	currentPath := history.Current
	history.History = []string{currentPath}
	history.Index = 0

	// Publish event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.history_cleared", map[string]interface{}{
			"session_id": sessionID,
			"current_path": currentPath,
		}, "file-navigator")
	}

	return nil
}

// AddBookmark adds a bookmark for the current directory
func (fn *FileNavigator) AddBookmark(sessionID string, name string, description string, tags []string) error {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	history, exists := fn.history[sessionID]
	if !exists {
		return fmt.Errorf("navigation session %s not found", sessionID)
	}

	bookmark := &NavigationBookmark{
		Name:        name,
		Path:        history.Current,
		Description: description,
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
		UseCount:    0,
		Tags:        tags,
	}

	// Store bookmark using path as key
	fn.bookmarks[history.Current] = bookmark

	// Auto-save if configured
	if fn.config.AutoSaveBookmarks {
		if err := fn.saveBookmarks(); err != nil {
			return fmt.Errorf("failed to save bookmarks: %w", err)
		}
	}

	// Publish event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.bookmark_added", map[string]interface{}{
			"session_id":  sessionID,
			"bookmark":    bookmark,
			"path":        history.Current,
		}, "file-navigator")
	}

	return nil
}

// RemoveBookmark removes a bookmark by path
func (fn *FileNavigator) RemoveBookmark(path string) error {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	if _, exists := fn.bookmarks[path]; !exists {
		return fmt.Errorf("bookmark not found for path: %s", path)
	}

	delete(fn.bookmarks, path)

	// Auto-save if configured
	if fn.config.AutoSaveBookmarks {
		if err := fn.saveBookmarks(); err != nil {
			return fmt.Errorf("failed to save bookmarks: %w", err)
		}
	}

	// Publish event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.bookmark_removed", map[string]interface{}{
			"path": path,
		}, "file-navigator")
	}

	return nil
}

// UpdateBookmark updates an existing bookmark
func (fn *FileNavigator) UpdateBookmark(path string, name string, description string, tags []string) error {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	bookmark, exists := fn.bookmarks[path]
	if !exists {
		return fmt.Errorf("bookmark not found for path: %s", path)
	}

	// Update bookmark fields
	if name != "" {
		bookmark.Name = name
	}
	bookmark.Description = description
	if tags != nil {
		bookmark.Tags = tags
	}

	// Auto-save if configured
	if fn.config.AutoSaveBookmarks {
		if err := fn.saveBookmarks(); err != nil {
			return fmt.Errorf("failed to save bookmarks: %w", err)
		}
	}

	return nil
}

// NavigateToBookmark navigates to a bookmarked location
func (fn *FileNavigator) NavigateToBookmark(sessionID string, path string) (*NavigationResult, error) {
	fn.mutex.Lock()

	bookmark, exists := fn.bookmarks[path]
	if !exists {
		fn.mutex.Unlock()
		return nil, fmt.Errorf("bookmark not found for path: %s", path)
	}

	// Update usage statistics
	bookmark.LastUsed = time.Now()
	bookmark.UseCount++

	// Auto-save if configured
	if fn.config.AutoSaveBookmarks {
		if err := fn.saveBookmarks(); err != nil {
			fn.mutex.Unlock()
			return nil, fmt.Errorf("failed to save bookmarks: %w", err)
		}
	}

	fn.mutex.Unlock()

	// Navigate to the bookmarked path
	return fn.NavigateTo(sessionID, path)
}

// GetBookmarks returns all bookmarks
func (fn *FileNavigator) GetBookmarks() ([]*NavigationBookmark, error) {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	bookmarks := make([]*NavigationBookmark, 0, len(fn.bookmarks))
	for _, bookmark := range fn.bookmarks {
		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

// GetBookmarkByPath returns a specific bookmark by path
func (fn *FileNavigator) GetBookmarkByPath(path string) (*NavigationBookmark, error) {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	bookmark, exists := fn.bookmarks[path]
	if !exists {
		return nil, fmt.Errorf("bookmark not found for path: %s", path)
	}

	return bookmark, nil
}

// IsBookmarked checks if a path is bookmarked
func (fn *FileNavigator) IsBookmarked(path string) bool {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	_, exists := fn.bookmarks[path]
	return exists
}

// SetPreferences updates navigation preferences for a session
func (fn *FileNavigator) SetPreferences(sessionID string, preferences *NavigationPreferences) error {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	if preferences == nil {
		return fmt.Errorf("preferences cannot be nil")
	}

	// Store preferences for the session
	fn.preferences[sessionID] = preferences

	// Publish event
	if fn.eventBus != nil {
		ctx := context.Background()
		fn.eventBus.Publish(ctx, "navigation.preferences_updated", map[string]interface{}{
			"session_id":  sessionID,
			"preferences": preferences,
		}, "file-navigator")
	}

	return nil
}

// GetPreferences returns navigation preferences for a session
func (fn *FileNavigator) GetPreferences(sessionID string) (*NavigationPreferences, error) {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	preferences, exists := fn.preferences[sessionID]
	if !exists {
		// Return default preferences if not set
		return &NavigationPreferences{
			ShowHidden:     fn.config.ShowHiddenFiles,
			SortingMethod:  fn.config.SortingMethod,
			GroupDirsFirst: fn.config.GroupDirectoriesFirst,
			CaseSensitive:  fn.config.CaseSensitiveSort,
			NaturalSort:    fn.config.NaturalSorting,
			ViewMode:       ViewModeList,
			ItemsPerPage:   100,
			PreviewEnabled: true,
			PreviewSize:    1024 * 1024, // 1MB
		}, nil
	}

	// Return a copy
	prefCopy := *preferences
	return &prefCopy, nil
}

// validatePath validates and cleans a file system path
func (fn *FileNavigator) validatePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Clean the path
	cleanPath := filepath.Clean(path)

	// Convert to absolute path
	if !filepath.IsAbs(cleanPath) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		cleanPath = filepath.Join(wd, cleanPath)
	}

	return cleanPath, nil
}

// updateHistory updates the navigation history
func (fn *FileNavigator) updateHistory(history *models.FileNavigationHistory, newPath string) {
	// Remove any entries after current index (forward history)
	if history.Index < len(history.History)-1 {
		history.History = history.History[:history.Index+1]
	}

	// Add new path
	history.History = append(history.History, newPath)
	history.Index = len(history.History) - 1

	// Trim history if it exceeds max size
	if len(history.History) > history.MaxHistory {
		excess := len(history.History) - history.MaxHistory
		history.History = history.History[excess:]
		history.Index -= excess
	}

	// Update current path
	history.Current = newPath
}

// getDirectoryEntries retrieves entries for a directory
func (fn *FileNavigator) getDirectoryEntries(path string, sessionID string) ([]*models.FileSystemEntry, error) {
	// Check cache first
	if fn.config.CacheEnabled {
		if cached := fn.getCachedEntries(path, sessionID); cached != nil {
			return cached, nil
		}
	}

	// Read directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Convert to FileSystemEntry models
	fileEntries := make([]*models.FileSystemEntry, 0, len(entries))
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		// Skip hidden files if not configured to show them
		if !fn.config.ShowHiddenFiles && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Create file system entry
		fileEntry, err := models.NewFileSystemEntry(fullPath)
		if err != nil {
			// Create minimal entry if full creation fails
			info, _ := entry.Info()
			fileEntry = &models.FileSystemEntry{
				Name:  entry.Name(),
				Path:  fullPath,
				IsDir: entry.IsDir(),
				Mode:  info.Mode(),
				Size:  info.Size(),
			}
		}

		fileEntries = append(fileEntries, fileEntry)
	}

	// Sort entries
	fn.sortEntries(fileEntries)

	// Cache the results
	if fn.config.CacheEnabled {
		fn.cacheEntries(path, fileEntries, sessionID)
	}

	return fileEntries, nil
}

// sortEntries sorts directory entries according to configuration
func (fn *FileNavigator) sortEntries(entries []*models.FileSystemEntry) {
	sort.Slice(entries, func(i, j int) bool {
		entryA, entryB := entries[i], entries[j]

		// Group directories first if configured
		if fn.config.GroupDirectoriesFirst {
			if entryA.IsDir && !entryB.IsDir {
				return true
			}
			if !entryA.IsDir && entryB.IsDir {
				return false
			}
		}

		// Sort based on configured method
		switch fn.config.SortingMethod {
		case SortByName:
			return fn.compareNames(entryA.Name, entryB.Name)
		case SortBySize:
			if entryA.Size == entryB.Size {
				return fn.compareNames(entryA.Name, entryB.Name)
			}
			return entryA.Size < entryB.Size
		case SortByModified:
			if entryA.ModTime.Equal(entryB.ModTime) {
				return fn.compareNames(entryA.Name, entryB.Name)
			}
			return entryA.ModTime.Before(entryB.ModTime)
		case SortByType:
			extA := entryA.GetExtension()
			extB := entryB.GetExtension()
			if extA == extB {
				return fn.compareNames(entryA.Name, entryB.Name)
			}
			return extA < extB
		default:
			return fn.compareNames(entryA.Name, entryB.Name)
		}
	})
}

// compareNames compares two file names for sorting
func (fn *FileNavigator) compareNames(nameA, nameB string) bool {
	if fn.config.NaturalSorting {
		return fn.naturalCompare(nameA, nameB)
	}

	if fn.config.CaseSensitiveSort {
		return nameA < nameB
	}

	return strings.ToLower(nameA) < strings.ToLower(nameB)
}

// naturalCompare performs natural (human-friendly) string comparison
func (fn *FileNavigator) naturalCompare(a, b string) bool {
	// Simple natural sort implementation
	aParts := fn.splitNumericParts(a)
	bParts := fn.splitNumericParts(b)

	minLen := len(aParts)
	if len(bParts) < minLen {
		minLen = len(bParts)
	}

	for i := 0; i < minLen; i++ {
		aPart, bPart := aParts[i], bParts[i]

		// Try numeric comparison
		if aNum, aOk := fn.isNumeric(aPart); aOk {
			if bNum, bOk := fn.isNumeric(bPart); bOk {
				if aNum != bNum {
					return aNum < bNum
				}
			}
		}

		// String comparison
		if aPart != bPart {
			if !fn.config.CaseSensitiveSort {
				return strings.ToLower(aPart) < strings.ToLower(bPart)
			}
			return aPart < bPart
		}
	}

	return len(aParts) < len(bParts)
}

// splitNumericParts splits a string into numeric and non-numeric parts
func (fn *FileNavigator) splitNumericParts(s string) []string {
	var parts []string
	var current strings.Builder

	for _, r := range s {
		isDigit := r >= '0' && r <= '9'
		hasContent := current.Len() > 0

		if hasContent {
			lastChar := current.String()[current.Len()-1]
			lastIsDigit := lastChar >= '0' && lastChar <= '9'

			if isDigit != lastIsDigit {
				parts = append(parts, current.String())
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// isNumeric checks if a string represents a number
func (fn *FileNavigator) isNumeric(s string) (int, bool) {
	var num int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		num = num*10 + int(r-'0')
	}
	return num, true
}

// createNavigationResult creates a navigation result from entries
func (fn *FileNavigator) createNavigationResult(path string, entries []*models.FileSystemEntry, sessionID string) *NavigationResult {
	parentPath := filepath.Dir(path)
	hasParent := parentPath != path

	// Count files, directories, and hidden files
	fileCount := 0
	dirCount := 0
	hiddenCount := 0

	for _, entry := range entries {
		if entry.IsDir {
			dirCount++
		} else {
			fileCount++
		}
		if strings.HasPrefix(entry.Name, ".") {
			hiddenCount++
		}
	}

	return &NavigationResult{
		Path:        path,
		Entries:     entries,
		ParentPath:  parentPath,
		HasParent:   hasParent,
		SessionID:   sessionID,
		TotalCount:  len(entries),
		Directories: dirCount,
		Files:       fileCount,
		HiddenCount: hiddenCount,
	}
}

// getCachedEntries retrieves cached directory entries
func (fn *FileNavigator) getCachedEntries(path string, sessionID string) []*models.FileSystemEntry {
	cacheKey := fmt.Sprintf("%s:%s", sessionID, path)
	entry, exists := fn.cache[cacheKey]
	if !exists {
		return nil
	}

	// Check if cache has expired
	if time.Since(entry.Timestamp) > fn.config.CacheTimeout {
		delete(fn.cache, cacheKey)
		return nil
	}

	return entry.Entries
}

// cacheEntries caches directory entries
func (fn *FileNavigator) cacheEntries(path string, entries []*models.FileSystemEntry, sessionID string) {
	cacheKey := fmt.Sprintf("%s:%s", sessionID, path)
	fn.cache[cacheKey] = &NavigationCacheEntry{
		Path:      path,
		Entries:   entries,
		Timestamp: time.Now(),
	}
}

// InvalidateCache invalidates cache for a specific path or all caches
func (fn *FileNavigator) InvalidateCache(path string) {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	if path == "" {
		// Clear all cache
		fn.cache = make(map[string]*NavigationCacheEntry)
		return
	}

	// Clear cache entries containing this path
	for key := range fn.cache {
		if strings.Contains(key, path) {
			delete(fn.cache, key)
		}
	}
}

// loadBookmarks loads bookmarks from file
func (fn *FileNavigator) loadBookmarks() error {
	// Check if bookmark file exists
	if _, err := os.Stat(fn.bookmarkPath); os.IsNotExist(err) {
		// No bookmark file yet, this is fine
		return nil
	}

	// Read bookmark file
	data, err := os.ReadFile(fn.bookmarkPath)
	if err != nil {
		return fmt.Errorf("failed to read bookmarks file: %w", err)
	}

	// Handle empty file
	if len(data) == 0 {
		return nil
	}

	// Parse bookmark store
	var store BookmarkStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse bookmarks file: %w", err)
	}

	// Load bookmarks into memory
	if store.Bookmarks != nil {
		fn.bookmarks = store.Bookmarks
	}

	return nil
}

// saveBookmarks saves bookmarks to file
func (fn *FileNavigator) saveBookmarks() error {
	// Create bookmark store
	store := BookmarkStore{
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
		Bookmarks: fn.bookmarks,
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(fn.bookmarkPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create bookmarks directory: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(fn.bookmarkPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write bookmarks file: %w", err)
	}

	return nil
}

// ExportBookmarks exports bookmarks to a specified file
func (fn *FileNavigator) ExportBookmarks(exportPath string) error {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	store := BookmarkStore{
		Version:   "1.0.0",
		UpdatedAt: time.Now(),
		Bookmarks: fn.bookmarks,
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	if err := os.WriteFile(exportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportBookmarks imports bookmarks from a file
func (fn *FileNavigator) ImportBookmarks(importPath string, merge bool) error {
	// Read import file
	data, err := os.ReadFile(importPath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Parse bookmark store
	var store BookmarkStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	if merge {
		// Merge with existing bookmarks
		for path, bookmark := range store.Bookmarks {
			fn.bookmarks[path] = bookmark
		}
	} else {
		// Replace all bookmarks
		fn.bookmarks = store.Bookmarks
	}

	// Auto-save if configured
	if fn.config.AutoSaveBookmarks {
		if err := fn.saveBookmarks(); err != nil {
			return fmt.Errorf("failed to save merged bookmarks: %w", err)
		}
	}

	return nil
}

// GetStats returns navigation statistics
func (fn *FileNavigator) GetStats() map[string]interface{} {
	fn.mutex.RLock()
	defer fn.mutex.RUnlock()

	return map[string]interface{}{
		"active_sessions":  len(fn.history),
		"bookmark_count":   len(fn.bookmarks),
		"cache_entries":    len(fn.cache),
		"preferences_count": len(fn.preferences),
		"config":           fn.config,
		"bookmark_path":    fn.bookmarkPath,
	}
}

// Cleanup cleans up resources
func (fn *FileNavigator) Cleanup() error {
	fn.mutex.Lock()
	defer fn.mutex.Unlock()

	// Save bookmarks
	if fn.config.AutoSaveBookmarks {
		if err := fn.saveBookmarks(); err != nil {
			return fmt.Errorf("failed to save bookmarks: %w", err)
		}
	}

	// Clear all sessions
	fn.history = make(map[string]*models.FileNavigationHistory)

	return nil
}