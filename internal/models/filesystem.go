package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileSystemEntry represents a file or directory in the file system
type FileSystemEntry struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	IsDir        bool              `json:"is_dir"`
	Size         int64             `json:"size"`
	Mode         os.FileMode       `json:"mode"`
	ModTime      time.Time         `json:"mod_time"`
	Permissions  string            `json:"permissions"`
	Owner        string            `json:"owner"`
	Group        string            `json:"group"`
	IsHidden     bool              `json:"is_hidden"`
	IsExecutable bool              `json:"is_executable"`
	Children     []FileSystemEntry `json:"children,omitempty"`
	// Enhanced permission information
	PermissionsInfo *PermissionsInfo `json:"permissions_info,omitempty"`
	// Additional metadata
	ExtendedAttrs map[string]string `json:"extended_attrs,omitempty"`
	Thumbnail     string            `json:"thumbnail,omitempty,omitempty"`
	Preview       string            `json:"preview,omitempty,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	CreatedAt     time.Time         `json:"created_at,omitempty"`
	AccessedAt    time.Time         `json:"accessed_at,omitempty"`
	ChangedAt     time.Time         `json:"changed_at,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	Encoding      string            `json:"encoding,omitempty"`
}

// PermissionsInfo provides detailed permission information
type PermissionsInfo struct {
	Octal       string `json:"octal"`       // e.g., "755"
	Symbolic    string `json:"symbolic"`    // e.g., "rwxr-xr-x"
	User        string `json:"user"`        // e.g., "rwx"
	Group       string `json:"group"`       // e.g., "r-x"
	Other       string `json:"other"`       // e.g., "r-x"
	UserID      int    `json:"user_id"`     // UID
	GroupID     int    `json:"group_id"`    // GID
	IsOwner     bool   `json:"is_owner"`    // Current user is owner
	IsGroup     bool   `json:"is_group"`    // Current user is in group
	CanRead     bool   `json:"can_read"`    // Current user can read
	CanWrite    bool   `json:"can_write"`   // Current user can write
	CanExecute  bool   `json:"can_execute"` // Current user can execute
	Permissions uint32 `json:"permissions"` // Raw permission bits
}

// FileSystemEvent represents a file system change event
type FileSystemEvent struct {
	Path      string           `json:"path"`
	Operation string           `json:"operation"` // "create", "remove", "rename", "write", "chmod", "move"
	Entry     *FileSystemEntry `json:"entry,omitempty"`
	Timestamp int64            `json:"timestamp"`
	SessionID string           `json:"session_id,omitempty"`
	Error     string           `json:"error,omitempty"`
}

// FileSystemStats provides statistics about file system operations
type FileSystemStats struct {
	TotalFiles       int              `json:"total_files"`
	TotalDirectories int              `json:"total_directories"`
	TotalSize        int64            `json:"total_size"`
	LargestFile      *FileSystemEntry `json:"largest_file,omitempty"`
	MostRecent       *FileSystemEntry `json:"most_recent,omitempty"`
	HiddenFiles      int              `json:"hidden_files"`
	ExecutableFiles  int              `json:"executable_files"`
	LastScanTime     time.Time        `json:"last_scan_time"`
}

// FileSearchOptions defines search parameters for file system searches
type FileSearchOptions struct {
	Pattern       string     `json:"pattern"`
	CaseSensitive bool       `json:"case_sensitive"`
	IncludeHidden bool       `json:"include_hidden"`
	MaxResults    int        `json:"max_results"`
	SearchType    string     `json:"search_type"` // "name", "content", "path"
	FileTypes     []string   `json:"file_types"`
	ModifiedDate  *time.Time `json:"modified_date"`
	SizeRange     *SizeRange `json:"size_range,omitempty"`
}

// SizeRange defines a range for file size filtering
type SizeRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// FileOperationResult represents the result of a file system operation
type FileOperationResult struct {
	Success   bool             `json:"success"`
	Path      string           `json:"path"`
	Operation string           `json:"operation"`
	Error     string           `json:"error,omitempty"`
	Timestamp int64            `json:"timestamp"`
	Entry     *FileSystemEntry `json:"entry,omitempty"`
}

// DirectoryWatcherConfig contains configuration for directory watching
type DirectoryWatcherConfig struct {
	Enabled        bool     `json:"enabled"`
	WatchHidden    bool     `json:"watch_hidden"`
	DebounceMs     int      `json:"debounce_ms"`
	MaxEvents      int      `json:"max_events"`
	BufferSize     int      `json:"buffer_size"`
	IgnorePatterns []string `json:"ignore_patterns"`
}

// FileNavigationHistory maintains navigation history
type FileNavigationHistory struct {
	Current    string   `json:"current"`
	History    []string `json:"history"`
	MaxHistory int      `json:"max_history"`
	Index      int      `json:"index"`
}

// FileOperation represents a file system operation type
type FileOperation string

const (
	OperationCreate   FileOperation = "create"
	OperationRemove   FileOperation = "remove"
	OperationRename   FileOperation = "rename"
	OperationWrite    FileOperation = "write"
	OperationChmod    FileOperation = "chmod"
	OperationMove     FileOperation = "move"
	OperationDelete   FileOperation = "delete"
	OperationCopy     FileOperation = "copy"
	OperationSymlink  FileOperation = "symlink"
	OperationHardlink FileOperation = "hardlink"
)

// Helper methods for FileSystemEntry

// IsRoot returns true if the entry is the root directory
func (f *FileSystemEntry) IsRoot() bool {
	return f.Path == "/" || (len(f.Path) == 3 && f.Path[1:3] == ":\\") // Windows drive root
}

// GetExtension returns the file extension
func (f *FileSystemEntry) GetExtension() string {
	if f.IsDir {
		return ""
	}
	return getFileExtension(f.Name)
}

// GetParent returns the parent directory path
func (f *FileSystemEntry) GetParent() string {
	if f.IsRoot() {
		return f.Path
	}
	return filepath.Dir(f.Path)
}

// GetRelativePath returns the path relative to a base directory
func (f *FileSystemEntry) GetRelativePath(base string) string {
	rel, err := filepath.Rel(base, f.Path)
	if err != nil {
		return f.Path
	}
	return rel
}

// GetSizeFormatted returns a human-readable file size
func (f *FileSystemEntry) GetSizeFormatted() string {
	return formatBytes(f.Size)
}

// GetPermissionsFormatted returns formatted permissions string
func (f *FileSystemEntry) GetPermissionsFormatted() string {
	return f.Mode.String()
}

// IsTextFile returns true if the file is likely a text file
func (f *FileSystemEntry) IsTextFile() bool {
	if f.IsDir {
		return false
	}

	// Check by extension
	textExtensions := map[string]bool{
		".txt": true, ".md": true, ".json": true, ".xml": true, ".yaml": true, ".yml": true,
		".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".html": true, ".css": true, ".scss": true, ".sass": true, ".less": true,
		".sh": true, ".bash": true, ".zsh": true, ".fish": true, ".ps1": true,
		".log": true, ".conf": true, ".config": true, ".ini": true, ".env": true,
		".sql": true, ".gitignore": true, ".dockerfile": true, "Dockerfile": true,
		".go": true, ".rs": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true,
		".java": true, ".class": true, ".kt": true, ".scala": true, ".php": true,
		".rb": true, ".pl": true, ".r": true,
	}

	ext := strings.ToLower(f.GetExtension())
	return textExtensions[ext]
}

// IsImageFile returns true if the file is an image
func (f *FileSystemEntry) IsImageFile() bool {
	if f.IsDir {
		return false
	}

	imageExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".svg": true, ".webp": true, ".ico": true, ".tiff": true, ".tif": true,
		".psd": true, ".ai": true, ".eps": true, ".raw": true, ".heic": true,
	}

	ext := strings.ToLower(f.GetExtension())
	return imageExtensions[ext]
}

// IsVideoFile returns true if the file is a video
func (f *FileSystemEntry) IsVideoFile() bool {
	if f.IsDir {
		return false
	}

	videoExtensions := map[string]bool{
		".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true,
		".flv": true, ".webm": true, ".m4v": true, ".mpg": true, ".mpeg": true,
		".3gp": true, ".ogv": true, ".ts": true, ".mts": true, ".vob": true,
	}

	ext := strings.ToLower(f.GetExtension())
	return videoExtensions[ext]
}

// IsAudioFile returns true if the file is an audio file
func (f *FileSystemEntry) IsAudioFile() bool {
	if f.IsDir {
		return false
	}

	audioExtensions := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true,
		".wma": true, ".m4a": true, ".opus": true, ".aiff": true, ".au": true,
		".ra": true, ".wv": true, ".mpc": true,
	}

	ext := strings.ToLower(f.GetExtension())
	return audioExtensions[ext]
}

// IsArchive returns true if the file is an archive
func (f *FileSystemEntry) IsArchive() bool {
	if f.IsDir {
		return false
	}

	archiveExtensions := map[string]bool{
		".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
		".7z": true, ".rar": true, ".iso": true, ".dmg": true, ".pkg": true,
		".deb": true, ".rpm": true, ".apk": true, ".exe": true, ".msi": true,
	}

	ext := strings.ToLower(f.GetExtension())
	return archiveExtensions[ext]
}

// UpdateTimestamps updates all timestamp fields to current time
func (f *FileSystemEntry) UpdateTimestamps() {
	now := time.Now()
	f.ModTime = now
	f.AccessedAt = now
	f.ChangedAt = now
}

// UpdateStats updates file statistics from OS info
func (f *FileSystemEntry) UpdateStats() error {
	info, err := os.Stat(f.Path)
	if err != nil {
		return WrapFileSystemError(err, f.Path, "stat")
	}

	f.Size = info.Size()
	f.Mode = info.Mode()
	f.ModTime = info.ModTime()

	// Update permission information
	if err := f.UpdatePermissionsInfo(); err != nil {
		// Don't fail the whole operation if permission update fails
		// just log the error and continue
		fmt.Printf("Warning: failed to update permission info for %s: %v\n", f.Path, err)
	}

	// Try to get extended time information if available
	if accessed, changed, ok := statTimes(info); ok {
		f.AccessedAt = accessed
		f.ChangedAt = changed
		// Use ModTime for creation time (birthtime not available on all systems)
		f.CreatedAt = info.ModTime()
	} else {
		// Fallback: use ModTime for other timestamps
		f.AccessedAt = info.ModTime()
		f.ChangedAt = info.ModTime()
		f.CreatedAt = info.ModTime()
	}

	return nil
}

// GetMIMEType determines the MIME type of the file
func (f *FileSystemEntry) GetMIMEType() string {
	if f.IsDir {
		return "inode/directory"
	}

	// Use file extension to determine MIME type
	ext := strings.ToLower(f.GetExtension())
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// UpdatePermissionsInfo updates the permission information for the entry
func (f *FileSystemEntry) UpdatePermissionsInfo() error {
	info, err := os.Stat(f.Path)
	if err != nil {
		return WrapFileSystemError(err, f.Path, "stat")
	}

	// Get current user information for permission checks
	currentUID := os.Getuid()
	currentGID := os.Getgid()

	// Extract permission bits
	mode := info.Mode()
	permissions := uint32(mode.Perm())

	// Numeric ownership, where the platform has it. Windows identifies owners
	// by SID rather than uid/gid, so these stay zero there.
	uid, gid, _ := statOwner(info)

	// Create permission info
	f.PermissionsInfo = &PermissionsInfo{
		Octal:       fmt.Sprintf("%o", permissions),
		Symbolic:    mode.String(),
		User:        formatPermissionBits(permissions, 6, 9), // User bits (rwx)
		Group:       formatPermissionBits(permissions, 3, 6), // Group bits (rwx)
		Other:       formatPermissionBits(permissions, 0, 3), // Other bits (rwx)
		UserID:      uid,
		GroupID:     gid,
		Permissions: permissions,
	}

	// Check ownership and access rights
	f.PermissionsInfo.IsOwner = f.PermissionsInfo.UserID == currentUID
	f.PermissionsInfo.IsGroup = f.PermissionsInfo.GroupID == currentGID

	// Check read permissions
	if f.PermissionsInfo.IsOwner {
		f.PermissionsInfo.CanRead = permissions&0400 != 0
		f.PermissionsInfo.CanWrite = permissions&0200 != 0
		f.PermissionsInfo.CanExecute = permissions&0100 != 0
	} else if f.PermissionsInfo.IsGroup {
		f.PermissionsInfo.CanRead = permissions&0040 != 0
		f.PermissionsInfo.CanWrite = permissions&0020 != 0
		f.PermissionsInfo.CanExecute = permissions&0010 != 0
	} else {
		f.PermissionsInfo.CanRead = permissions&0004 != 0
		f.PermissionsInfo.CanWrite = permissions&0002 != 0
		f.PermissionsInfo.CanExecute = permissions&0001 != 0
	}

	return nil
}

// CheckPermissions verifies if the current operation is allowed
func (f *FileSystemEntry) CheckPermissions(operation string) error {
	if f.PermissionsInfo == nil {
		if err := f.UpdatePermissionsInfo(); err != nil {
			return WrapFileSystemError(err, f.Path, "check_permissions")
		}
	}

	switch strings.ToLower(operation) {
	case "read", "stat", "list":
		if !f.PermissionsInfo.CanRead {
			return NewFileSystemError(ErrorPermission, f.Path, operation,
				"read permission denied", nil)
		}
	case "write", "create", "modify":
		if !f.PermissionsInfo.CanWrite {
			return NewFileSystemError(ErrorPermission, f.Path, operation,
				"write permission denied", nil)
		}
	case "execute", "search":
		if !f.PermissionsInfo.CanExecute {
			return NewFileSystemError(ErrorPermission, f.Path, operation,
				"execute permission denied", nil)
		}
	default:
		// For unknown operations, require write access
		if !f.PermissionsInfo.CanWrite {
			return NewFileSystemError(ErrorPermission, f.Path, operation,
				"permission denied for operation", nil)
		}
	}

	return nil
}

// ValidatePath validates the file path for security and correctness
func (f *FileSystemEntry) ValidatePath() error {
	if f.Path == "" {
		return NewFileSystemError(ErrorInvalidPath, f.Path, "validate",
			"empty path", nil)
	}

	// Check for path traversal attempts
	if strings.Contains(f.Path, "..") {
		cleaned := filepath.Clean(f.Path)
		if cleaned != f.Path {
			return NewFileSystemError(ErrorSecurity, f.Path, "validate",
				"path traversal detected", nil)
		}
	}

	// Check for invalid characters
	if strings.ContainsAny(f.Path, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f") {
		return NewFileSystemError(ErrorInvalidPath, f.Path, "validate",
			"invalid characters in path", nil)
	}

	// Check path length (common limits)
	if len(f.Path) > 4096 {
		return NewFileSystemError(ErrorInvalidPath, f.Path, "validate",
			"path too long", nil)
	}

	return nil
}

// ValidateForOperation validates the entry for a specific operation
func (f *FileSystemEntry) ValidateForOperation(operation string) error {
	// First validate the path
	if err := f.ValidatePath(); err != nil {
		return err
	}

	// Check if the entry exists
	if _, err := os.Stat(f.Path); err != nil {
		if os.IsNotExist(err) && operation == "create" {
			// For create operations, the parent directory must exist
			parent := filepath.Dir(f.Path)
			if _, parentErr := os.Stat(parent); parentErr != nil {
				return WrapError(parentErr, parent, "validate_parent")
			}
			return nil
		}
		return WrapError(err, f.Path, "validate")
	}

	// Check permissions for the operation
	if err := f.CheckPermissions(operation); err != nil {
		return err
	}

	// Additional validation based on operation type
	switch strings.ToLower(operation) {
	case "create":
		if f.IsDir {
			return NewFileSystemError(ErrorExists, f.Path, operation,
				"directory already exists", nil)
		}
	case "delete":
		// Check if the path is a system directory
		systemPaths := []string{"/", "/bin", "/sbin", "/usr", "/etc", "/var", "/sys", "/proc", "/dev"}
		for _, sysPath := range systemPaths {
			if f.Path == sysPath || strings.HasPrefix(f.Path, sysPath+"/") {
				return NewFileSystemError(ErrorSecurity, f.Path, operation,
					"cannot delete system directory", nil)
			}
		}
	case "rename", "move":
		// Ensure the target is within allowed bounds
		if strings.HasPrefix(f.Path, "/") && !strings.HasPrefix(f.Path, "/tmp/") &&
			!strings.HasPrefix(f.Path, "/home/") && !strings.HasPrefix(f.Path, "/Users/") {
			// This is a system path - be more careful
			systemPaths := []string{"/", "/bin", "/sbin", "/usr", "/etc", "/var", "/sys", "/proc", "/dev"}
			for _, sysPath := range systemPaths {
				if f.Path == sysPath || strings.HasPrefix(f.Path, sysPath+"/") {
					return NewFileSystemError(ErrorSecurity, f.Path, operation,
						"cannot modify system path", nil)
				}
			}
		}
	}

	return nil
}

// GetEffectivePermissions returns the effective permissions for the current user
func (f *FileSystemEntry) GetEffectivePermissions() string {
	if f.PermissionsInfo == nil {
		if err := f.UpdatePermissionsInfo(); err != nil {
			return "???"
		}
	}

	var perms []rune

	if f.PermissionsInfo.CanRead {
		perms = append(perms, 'r')
	} else {
		perms = append(perms, '-')
	}

	if f.PermissionsInfo.CanWrite {
		perms = append(perms, 'w')
	} else {
		perms = append(perms, '-')
	}

	if f.PermissionsInfo.CanExecute {
		perms = append(perms, 'x')
	} else {
		perms = append(perms, '-')
	}

	return string(perms)
}

// IsAccessible returns true if the file is accessible by the current user
func (f *FileSystemEntry) IsAccessible() bool {
	if f.PermissionsInfo == nil {
		if err := f.UpdatePermissionsInfo(); err != nil {
			return false
		}
	}

	return f.PermissionsInfo.CanRead
}

// IsWritable returns true if the file is writable by the current user
func (f *FileSystemEntry) IsWritable() bool {
	if f.PermissionsInfo == nil {
		if err := f.UpdatePermissionsInfo(); err != nil {
			return false
		}
	}

	return f.PermissionsInfo.CanWrite
}

// IsValid returns true if the file system entry is valid
func (f *FileSystemEntry) IsValid() bool {
	return f.Name != "" && f.Path != "" && (f.IsDir || f.Size >= 0)
}

// Clone creates a deep copy of the FileSystemEntry
func (f *FileSystemEntry) Clone() *FileSystemEntry {
	clone := *f

	// Deep copy children
	if len(f.Children) > 0 {
		clone.Children = make([]FileSystemEntry, len(f.Children))
		copy(clone.Children, f.Children)
	}

	// Deep copy maps
	if f.ExtendedAttrs != nil {
		clone.ExtendedAttrs = make(map[string]string)
		for k, v := range f.ExtendedAttrs {
			clone.ExtendedAttrs[k] = v
		}
	}

	if f.Tags != nil {
		clone.Tags = make([]string, len(f.Tags))
		copy(clone.Tags, f.Tags)
	}

	return &clone
}

// FileSystemError represents different types of file system errors
type FileSystemError struct {
	Type        FileSystemErrorType `json:"type"`
	Path        string              `json:"path"`
	Operation   string              `json:"operation"`
	Message     string              `json:"message"`
	OriginalErr error               `json:"-"`
	UserMessage string              `json:"user_message"`
	Permissions *PermissionsInfo    `json:"permissions,omitempty"`
	ErrorCode   int                 `json:"error_code"`
	Timestamp   time.Time           `json:"timestamp"`
}

// FileSystemErrorType represents categories of file system errors
type FileSystemErrorType string

const (
	ErrorPermission    FileSystemErrorType = "permission_denied"
	ErrorNotFound      FileSystemErrorType = "not_found"
	ErrorExists        FileSystemErrorType = "already_exists"
	ErrorInvalidPath   FileSystemErrorType = "invalid_path"
	ErrorIO            FileSystemErrorType = "io_error"
	ErrorSpace         FileSystemErrorType = "no_space"
	ErrorQuota         FileSystemErrorType = "quota_exceeded"
	ErrorReadOnly      FileSystemErrorType = "read_only"
	ErrorBusy          FileSystemErrorType = "resource_busy"
	ErrorCorrupted     FileSystemErrorType = "corrupted"
	ErrorLocked        FileSystemErrorType = "locked"
	ErrorNetwork       FileSystemErrorType = "network_error"
	ErrorTimeout       FileSystemErrorType = "timeout"
	ErrorInvalidName   FileSystemErrorType = "invalid_name"
	ErrorDirectoryFull FileSystemErrorType = "directory_full"
	ErrorTooManyLinks  FileSystemErrorType = "too_many_links"
	ErrorNotDirectory  FileSystemErrorType = "not_directory"
	ErrorNotFile       FileSystemErrorType = "not_file"
	ErrorSymLoop       FileSystemErrorType = "symlink_loop"
	ErrorInvalidMode   FileSystemErrorType = "invalid_mode"
	ErrorSecurity      FileSystemErrorType = "security_violation"
	ErrorConfiguration FileSystemErrorType = "configuration_error"
	ErrorUnknown       FileSystemErrorType = "unknown"
)

// Error returns a formatted error string
func (e *FileSystemError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("[%s] %s: %s (path: %s, operation: %s)", e.Type, e.Message, e.OriginalErr.Error(), e.Path, e.Operation)
	}
	return fmt.Sprintf("[%s] %s (path: %s, operation: %s)", e.Type, e.Message, e.Path, e.Operation)
}

// Unwrap returns the original error
func (e *FileSystemError) Unwrap() error {
	return e.OriginalErr
}

// Is checks if the error matches the target
func (e *FileSystemError) Is(target error) bool {
	if t, ok := target.(*FileSystemError); ok {
		return e.Type == t.Type
	}
	return false
}

// IsRetryable returns whether the operation can be retried
func (e *FileSystemError) IsRetryable() bool {
	switch e.Type {
	case ErrorTimeout, ErrorNetwork, ErrorBusy:
		return true
	case ErrorIO, ErrorSpace:
		return true
	default:
		return false
	}
}

// IsPermissionRelated returns whether the error is permission-related
func (e *FileSystemError) IsPermissionRelated() bool {
	switch e.Type {
	case ErrorPermission, ErrorReadOnly, ErrorSecurity:
		return true
	default:
		return false
	}
}

// GetErrorCode returns a platform-specific error code
func (e *FileSystemError) GetErrorCode() int {
	if e.ErrorCode != 0 {
		return e.ErrorCode
	}

	// Map error types to common error codes
	switch e.Type {
	case ErrorPermission:
		return 13 // EACCES
	case ErrorNotFound:
		return 2 // ENOENT
	case ErrorExists:
		return 17 // EEXIST
	case ErrorInvalidPath:
		return 22 // EINVAL
	case ErrorIO:
		return 5 // EIO
	case ErrorSpace:
		return 28 // ENOSPC
	case ErrorReadOnly:
		return 30 // EROFS
	case ErrorBusy:
		return 16 // EBUSY
	case ErrorTimeout:
		return 110 // ETIMEDOUT
	default:
		return 1 // EPERM
	}
}

// NewFileSystemError creates a new file system error
func NewFileSystemError(errType FileSystemErrorType, path, operation, message string, originalErr error) *FileSystemError {
	userMessage := generateUserMessage(errType, path, operation)

	errorCode := 0
	if originalErr != nil {
		if osErr, ok := originalErr.(*os.PathError); ok {
			errorCode = getErrorCodeFromOsError(osErr.Err)
		}
	}

	return &FileSystemError{
		Type:        errType,
		Path:        path,
		Operation:   operation,
		Message:     message,
		OriginalErr: originalErr,
		UserMessage: userMessage,
		ErrorCode:   errorCode,
		Timestamp:   time.Now(),
	}
}

// WrapFileSystemError wraps an existing error as a FileSystemError
func WrapFileSystemError(err error, path, operation string) *FileSystemError {
	if err == nil {
		return nil
	}

	// Check if it's already a FileSystemError
	if fsErr, ok := err.(*FileSystemError); ok {
		return fsErr
	}

	// Determine error type from error
	errType := classifyError(err)
	message := err.Error()

	return NewFileSystemError(errType, path, operation, message, err)
}

// classifyError determines the FileSystemErrorType from a standard error
func classifyError(err error) FileSystemErrorType {
	if err == nil {
		return ErrorUnknown
	}

	errStr := err.Error()

	// Check for specific OS errors
	if os.IsPermission(err) {
		return ErrorPermission
	}
	if os.IsExist(err) {
		return ErrorExists
	}
	if os.IsNotExist(err) {
		return ErrorNotFound
	}

	// Check error messages for patterns
	switch {
	case containsIgnoreCase(errStr, "permission denied"), containsIgnoreCase(errStr, "access denied"):
		return ErrorPermission
	case containsIgnoreCase(errStr, "file not found"), containsIgnoreCase(errStr, "no such file"):
		return ErrorNotFound
	case containsIgnoreCase(errStr, "already exists"), containsIgnoreCase(errStr, "file exists"):
		return ErrorExists
	case containsIgnoreCase(errStr, "invalid path"), containsIgnoreCase(errStr, "illegal character"):
		return ErrorInvalidPath
	case containsIgnoreCase(errStr, "read-only"), containsIgnoreCase(errStr, "read only"):
		return ErrorReadOnly
	case containsIgnoreCase(errStr, "no space"), containsIgnoreCase(errStr, "disk full"):
		return ErrorSpace
	case containsIgnoreCase(errStr, "resource busy"), containsIgnoreCase(errStr, "device busy"):
		return ErrorBusy
	case containsIgnoreCase(errStr, "timeout"), containsIgnoreCase(errStr, "timed out"):
		return ErrorTimeout
	case containsIgnoreCase(errStr, "network"), containsIgnoreCase(errStr, "connection"):
		return ErrorNetwork
	case containsIgnoreCase(errStr, "corrupted"), containsIgnoreCase(errStr, "checksum"):
		return ErrorCorrupted
	case containsIgnoreCase(errStr, "locked"), containsIgnoreCase(errStr, "in use"):
		return ErrorLocked
	case containsIgnoreCase(errStr, "too many links"), containsIgnoreCase(errStr, "link count"):
		return ErrorTooManyLinks
	case containsIgnoreCase(errStr, "not a directory"), containsIgnoreCase(errStr, "not directory"):
		return ErrorNotDirectory
	case containsIgnoreCase(errStr, "not a file"), containsIgnoreCase(errStr, "is a directory"):
		return ErrorNotFile
	case containsIgnoreCase(errStr, "symbolic link loop"), containsIgnoreCase(errStr, "symlink loop"):
		return ErrorSymLoop
	case containsIgnoreCase(errStr, "quota"), containsIgnoreCase(errStr, "limit exceeded"):
		return ErrorQuota
	default:
		return ErrorUnknown
	}
}

// generateUserMessage creates a user-friendly error message
func generateUserMessage(errType FileSystemErrorType, path, operation string) string {
	fileName := filepath.Base(path)

	switch errType {
	case ErrorPermission:
		return fmt.Sprintf("You don't have permission to %s '%s'. Check file permissions or try running with elevated privileges.", operation, fileName)
	case ErrorNotFound:
		return fmt.Sprintf("The file or folder '%s' could not be found. It may have been moved or deleted.", fileName)
	case ErrorExists:
		return fmt.Sprintf("A file or folder named '%s' already exists. Choose a different name or delete the existing one.", fileName)
	case ErrorInvalidPath:
		return fmt.Sprintf("The file path '%s' contains invalid characters or is not properly formatted.", path)
	case ErrorIO:
		return fmt.Sprintf("There was a problem reading or writing '%s'. The file may be corrupted or the drive may have issues.", fileName)
	case ErrorSpace:
		return fmt.Sprintf("There is not enough disk space to complete this operation.")
	case ErrorQuota:
		return fmt.Sprintf("You have exceeded your disk quota. Free up space or contact your administrator.")
	case ErrorReadOnly:
		return fmt.Sprintf("The file '%s' is read-only and cannot be modified.", fileName)
	case ErrorBusy:
		return fmt.Sprintf("The file '%s' is currently in use by another program. Try again later.", fileName)
	case ErrorCorrupted:
		return fmt.Sprintf("The file '%s' appears to be corrupted and cannot be accessed.", fileName)
	case ErrorLocked:
		return fmt.Sprintf("The file '%s' is locked. Unlock it or try again later.", fileName)
	case ErrorNetwork:
		return fmt.Sprintf("There was a network error accessing '%s'. Check your connection.", path)
	case ErrorTimeout:
		return fmt.Sprintf("The operation timed out. Try again or check if the system is responsive.")
	case ErrorInvalidName:
		return fmt.Sprintf("The file name '%s' contains invalid characters.", fileName)
	case ErrorDirectoryFull:
		return fmt.Sprintf("The directory cannot contain more files. The limit has been reached.")
	case ErrorTooManyLinks:
		return fmt.Sprintf("The file '%s' has too many symbolic links and cannot be accessed.", fileName)
	case ErrorNotDirectory:
		return fmt.Sprintf("'%s' is not a directory.", path)
	case ErrorNotFile:
		return fmt.Sprintf("'%s' is not a file.", path)
	case ErrorSymLoop:
		return fmt.Sprintf("There is a circular reference in the symbolic links for '%s'.", path)
	case ErrorSecurity:
		return fmt.Sprintf("Access to '%s' was blocked for security reasons.", path)
	default:
		return fmt.Sprintf("An unexpected error occurred while accessing '%s'.", fileName)
	}
}

// Helper functions for error handling
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func getErrorCodeFromOsError(osErr error) int {
	// Map common OS errors to error codes
	if osErr == os.ErrPermission {
		return 13 // EACCES
	}
	if osErr == os.ErrExist {
		return 17 // EEXIST
	}
	if osErr == os.ErrNotExist {
		return 2 // ENOENT
	}
	if osErr == os.ErrClosed {
		return 9 // EBADF
	}
	if osErr == os.ErrInvalid {
		return 22 // EINVAL
	}
	return 1 // EPERM default
}

// Helper functions

// formatPermissionBits formats permission bits for user, group, or other
func formatPermissionBits(permissions uint32, highBit, lowBit int) string {
	var result []rune

	// Extract bits for read, write, execute
	readBit := (permissions >> uint(highBit-1)) & 1
	writeBit := (permissions >> uint(highBit-2)) & 1
	execBit := (permissions >> uint(highBit-3)) & 1

	if readBit == 1 {
		result = append(result, 'r')
	} else {
		result = append(result, '-')
	}

	if writeBit == 1 {
		result = append(result, 'w')
	} else {
		result = append(result, '-')
	}

	if execBit == 1 {
		result = append(result, 'x')
	} else {
		result = append(result, '-')
	}

	return string(result)
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return ext
	}

	// Handle filenames with multiple dots
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}

	return ""
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), int64(0)
	for n := bytes; n >= unit && exp < 3; exp++ {
		n /= unit
		div *= unit
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPEZY"[exp])
}

// NewFileSystemEntry creates a new FileSystemEntry from OS file info
func NewFileSystemEntry(path string) (*FileSystemEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Get extended time information if available
	var createdAt, accessedAt, changedAt time.Time
	if a, c, ok := statTimes(info); ok {
		accessedAt = a
		changedAt = c
		// Use ModTime for creation time (birthtime not available on all systems)
		createdAt = info.ModTime()
	} else {
		// Fallback: use ModTime for other timestamps
		createdAt = info.ModTime()
		accessedAt = info.ModTime()
		changedAt = info.ModTime()
	}

	entry := &FileSystemEntry{
		Name:          filepath.Base(path),
		Path:          path,
		IsDir:         info.IsDir(),
		Size:          info.Size(),
		Mode:          info.Mode(),
		ModTime:       info.ModTime(),
		Permissions:   info.Mode().String(),
		Owner:         "", // Would need additional OS-specific calls
		Group:         "", // Would need additional OS-specific calls
		IsHidden:      strings.HasPrefix(filepath.Base(path), "."),
		IsExecutable:  info.Mode().Perm()&0111 != 0, // Execute bits for user, group, other
		Children:      nil,
		ExtendedAttrs: make(map[string]string),
		Tags:          []string{},
		CreatedAt:     createdAt,
		AccessedAt:    accessedAt,
		ChangedAt:     changedAt,
		ContentType:   "",
		Encoding:      "",
	}

	return entry, nil
}

// NewFileSystemEntryFromInfo creates a new FileSystemEntry from existing os.FileInfo
func NewFileSystemEntryFromInfo(path string, info os.FileInfo) *FileSystemEntry {
	entry := &FileSystemEntry{
		Name:          filepath.Base(path),
		Path:          path,
		IsDir:         info.IsDir(),
		Size:          info.Size(),
		Mode:          info.Mode(),
		ModTime:       info.ModTime(),
		Permissions:   info.Mode().String(),
		Owner:         "", // Would need additional OS-specific calls
		Group:         "", // Would need additional OS-specific calls
		IsHidden:      strings.HasPrefix(filepath.Base(path), "."),
		IsExecutable:  info.Mode().Perm()&0111 != 0,
		Children:      nil,
		ExtendedAttrs: make(map[string]string),
		Tags:          []string{},
		CreatedAt:     time.Time{}, // Initialize with zero value
		AccessedAt:    time.Time{}, // Initialize with zero value
		ChangedAt:     time.Time{}, // Initialize with zero value
		ContentType:   "",
		Encoding:      "",
	}

	return entry
}
