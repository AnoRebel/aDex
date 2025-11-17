package models

import "time"

// FileSystemEntry represents a filesystem entry (file or directory)
type FileSystemEntry struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Type         string            `json:"type"` // file, directory, symlink, etc.
	Size         int64             `json:"size"`
	Permissions  string            `json:"permissions"`
	Owner        string            `json:"owner"`
	Group        string            `json:"group"`
	Modified     time.Time         `json:"modified"`
	Accessed     time.Time         `json:"accessed"`
	Created      time.Time         `json:"created"`
	IsHidden     bool              `json:"is_hidden"`
	IsExecutable bool              `json:"is_executable"`
	IsReadable   bool              `json:"is_readable"`
	IsWritable   bool              `json:"is_writable"`
	Extension    string            `json:"extension"`
	MimeType     string            `json:"mime_type"`
	Attributes   map[string]string `json:"attributes"`
	Thumbnail    string            `json:"thumbnail,omitempty"`
}

// Directory represents a directory with its contents
type Directory struct {
	Path     string             `json:"path"`
	Entries  []FileSystemEntry  `json:"entries"`
	Total    int                `json:"total"`
	Files    int                `json:"files"`
	Dirs     int                `json:"dirs"`
	Symlinks int                `json:"symlinks"`
	Size     int64              `json:"size"`
	Modified time.Time          `json:"modified"`
}

// FileOperation represents a file operation
type FileOperation struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // copy, move, delete, create, etc.
	Source      string      `json:"source"`
	Destination string      `json:"destination,omitempty"`
	Status      string      `json:"status"` // pending, running, completed, failed
	Progress    float64     `json:"progress"`
	BytesTotal  int64       `json:"bytes_total"`
	BytesDone   int64       `json:"bytes_done"`
	Speed       int64       `json:"speed"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// FileSystemStats represents filesystem statistics
type FileSystemStats struct {
	Path           string    `json:"path"`
	TotalFiles     int       `json:"total_files"`
	TotalDirs      int       `json:"total_dirs"`
	TotalSize      int64     `json:"total_size"`
	FreeSpace      int64     `json:"free_space"`
	UsedSpace      int64     `json:"used_space"`
	LargestFile    string    `json:"largest_file"`
	LargestFileSize int64    `json:"largest_file_size"`
	LastModified   time.Time `json:"last_modified"`
}

// Bookmark represents a filesystem bookmark
type Bookmark struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	LastUsed    time.Time `json:"last_used"`
	IsSystem    bool      `json:"is_system"`
}

// FileSearchResult represents a file search result
type FileSearchResult struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	Modified     time.Time `json:"modified"`
	Match        string    `json:"match"`
	Line         int       `json:"line,omitempty"`
	Context      string    `json:"context,omitempty"`
}

// FileWatcher represents a filesystem watcher
type FileWatcher struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Recursive   bool      `json:"recursive"`
	Filters     []string  `json:"filters"`
	Events      []string  `json:"events"`
	IsActive    bool      `json:"is_active"`
	Created     time.Time `json:"created"`
	LastEvent   time.Time `json:"last_event"`
}

// FileSystemEvent represents a filesystem event
type FileSystemEvent struct {
	Type      string    `json:"type"` // create, modify, delete, move, etc.
	Path      string    `json:"path"`
	OldPath   string    `json:"old_path,omitempty"`
	Size      int64     `json:"size,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	WatcherID string    `json:"watcher_id"`
}

// FilePreview represents a file preview
type FilePreview struct {
	Path        string      `json:"path"`
	Type        string      `json:"type"`
	Content     string      `json:"content"`
	Size        int         `json:"size"`
	IsTruncated bool        `json:"is_truncated"`
	LineCount   int         `json:"line_count"`
	Encoding    string      `json:"encoding"`
	Language    string      `json:"language"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

// MountPoint represents a filesystem mount point
type MountPoint struct {
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Filesystem string `json:"filesystem"`
	Options    string `json:"options"`
	Total      int64  `json:"total"`
	Used       int64  `json:"used"`
	Free       int64  `json:"free"`
	Percent    float64 `json:"percent"`
}

// FilePermission represents file permissions
type FilePermission struct {
	OwnerRead    bool `json:"owner_read"`
	OwnerWrite   bool `json:"owner_write"`
	OwnerExecute bool `json:"owner_execute"`
	GroupRead    bool `json:"group_read"`
	GroupWrite   bool `json:"group_write"`
	GroupExecute bool `json:"group_execute"`
	OtherRead    bool `json:"other_read"`
	OtherWrite   bool `json:"other_write"`
	OtherExecute bool `json:"other_execute"`
}

// FileSystemConfig represents filesystem configuration
type FileSystemConfig struct {
	ShowHiddenFiles    bool     `json:"show_hidden_files"`
	ShowSystemFiles    bool     `json:"show_system_files"`
	DefaultView        string   `json:"default_view"` // list, grid, tree
	SortBy             string   `json:"sort_by"`
	SortOrder          string   `json:"sort_order"`
	Bookmarks          []string `json:"bookmarks"`
	RecentFiles        []string `json:"recent_files"`
	MaxRecentFiles     int      `json:"max_recent_files"`
	EnablePreview      bool     `json:"enable_preview"`
	MaxPreviewSize     int64    `json:"max_preview_size"`
	DoubleClickAction  string   `json:"double_click_action"`
	ConfirmDelete      bool     `json:"confirm_delete"`
	ShowFileSizeFormat string   `json:"show_file_size_format"`
}

// FileSystemError represents filesystem operation errors
type FileSystemError struct {
	Operation string `json:"operation"`
	Path      string `json:"path"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
