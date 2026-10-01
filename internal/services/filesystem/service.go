package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"aDex-UI/internal/utils"
)

// Service handles filesystem operations
type Service struct {
	watcher     *fsnotify.Watcher
	watchers    map[string]*fsnotify.Watcher
	watcherLock sync.RWMutex
	platform    *utils.FeatureDetection
}

// NewService creates a new filesystem service instance
func NewService() *Service {
	watcher, _ := fsnotify.NewWatcher()
	return &Service{
		watcher:  watcher,
		watchers: make(map[string]*fsnotify.Watcher),
		platform: utils.DetectPlatform(),
	}
}

// FileInfo represents file information.
//
// JSON tags use camelCase to match the frontend's `useWails`/`FileBrowser`
// expectations and align with DirectoryEntry's shape (which already worked
// in the frontend). Without explicit tags, encoding/json emits PascalCase
// and the frontend's `entry.isDir` reads return undefined.
type FileInfo struct {
	Name        string      `json:"name"`
	Path        string      `json:"path"`
	Size        int64       `json:"size"`
	IsDirectory bool        `json:"isDir"`
	Mode        os.FileMode `json:"mode"`
	ModTime     time.Time   `json:"modTime"`
	Permissions string      `json:"permissions"`
	Owner       string      `json:"owner"`
	Group       string      `json:"group"`
}

// DirectoryEntry represents a directory entry.
type DirectoryEntry struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"isDir"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	Extension string    `json:"extension"`
}

// ReadDirectory reads directory contents
func (s *Service) ReadDirectory(ctx context.Context, path string) ([]DirectoryEntry, error) {
	normalizedPath := s.platform.NormalizePath(path)

	entries, err := os.ReadDir(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", normalizedPath, err)
	}

	var directoryEntries []DirectoryEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't get info for
		}

		dirEntry := DirectoryEntry{
			Name:      entry.Name(),
			Path:      filepath.Join(normalizedPath, entry.Name()),
			IsDir:     entry.IsDir(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Extension: filepath.Ext(entry.Name()),
		}

		directoryEntries = append(directoryEntries, dirEntry)
	}

	return directoryEntries, nil
}

// GetFileInfo retrieves file information
func (s *Service) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	normalizedPath := s.platform.NormalizePath(path)

	info, err := os.Stat(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", normalizedPath, err)
	}

	fileInfo := &FileInfo{
		Name:        info.Name(),
		Path:        normalizedPath,
		Size:        info.Size(),
		IsDirectory: info.IsDir(),
		Mode:        info.Mode(),
		ModTime:     info.ModTime(),
		Permissions: info.Mode().String(),
	}

	return fileInfo, nil
}

// CreateDirectory creates a new directory
func (s *Service) CreateDirectory(ctx context.Context, path string, mode os.FileMode) error {
	normalizedPath := s.platform.NormalizePath(path)
	return os.MkdirAll(normalizedPath, mode)
}

// DeleteFile deletes a file or directory
func (s *Service) DeleteFile(ctx context.Context, path string) error {
	normalizedPath := s.platform.NormalizePath(path)
	return os.RemoveAll(normalizedPath)
}

// CopyFile copies a file from source to destination
func (s *Service) CopyFile(ctx context.Context, src, dst string) error {
	srcPath := s.platform.NormalizePath(src)
	dstPath := s.platform.NormalizePath(dst)

	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(srcPath)
	if err == nil {
		os.Chmod(dstPath, sourceInfo.Mode())
	}

	return nil
}

// MoveFile moves/renames a file
func (s *Service) MoveFile(ctx context.Context, src, dst string) error {
	srcPath := s.platform.NormalizePath(src)
	dstPath := s.platform.NormalizePath(dst)

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	return os.Rename(srcPath, dstPath)
}

// ReadFile reads file contents
func (s *Service) ReadFile(ctx context.Context, path string) ([]byte, error) {
	normalizedPath := s.platform.NormalizePath(path)
	return os.ReadFile(normalizedPath)
}

// WriteFile writes content to a file
func (s *Service) WriteFile(ctx context.Context, path string, data []byte, mode os.FileMode) error {
	normalizedPath := s.platform.NormalizePath(path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(normalizedPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(normalizedPath, data, mode)
}

// SearchFiles searches for files matching a pattern
func (s *Service) SearchFiles(ctx context.Context, root, pattern string) ([]string, error) {
	normalizedRoot := s.platform.NormalizePath(root)
	var matches []string

	err := filepath.Walk(normalizedRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/dirs we can't access
		}

		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search directory %s: %w", normalizedRoot, err)
	}

	return matches, nil
}

// WatchDirectory watches for directory changes
func (s *Service) WatchDirectory(ctx context.Context, path string) (<-chan string, error) {
	if !s.platform.HasFeature("inotify") {
		return nil, fmt.Errorf("file watching not supported on this platform")
	}

	normalizedPath := s.platform.NormalizePath(path)

	s.watcherLock.Lock()
	defer s.watcherLock.Unlock()

	// Create new watcher for this directory
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	// Add the directory to the watcher
	err = watcher.Add(normalizedPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	// Store the watcher
	s.watchers[normalizedPath] = watcher

	// Create channel for events
	eventChan := make(chan string, 100)

	// Start goroutine to handle events
	go func() {
		defer watcher.Close()
		defer close(eventChan)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				eventChan <- event.Name

			case <-watcher.Errors:
				return

			case <-ctx.Done():
				return
			}
		}
	}()

	return eventChan, nil
}

// StopWatching stops watching a directory
func (s *Service) StopWatching(path string) error {
	normalizedPath := s.platform.NormalizePath(path)

	s.watcherLock.Lock()
	defer s.watcherLock.Unlock()

	if watcher, exists := s.watchers[normalizedPath]; exists {
		delete(s.watchers, normalizedPath)
		return watcher.Close()
	}

	return nil
}

// Close cleans up the filesystem service
func (s *Service) Close() error {
	s.watcherLock.Lock()
	defer s.watcherLock.Unlock()

	// Close all watchers
	for path, watcher := range s.watchers {
		watcher.Close()
		delete(s.watchers, path)
	}

	// Close main watcher
	if s.watcher != nil {
		return s.watcher.Close()
	}

	return nil
}
