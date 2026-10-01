package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PlatformInfo holds platform-specific information and capabilities
type PlatformInfo struct {
	OS               string
	Arch             string
	IsWindows        bool
	IsLinux          bool
	IsDarwin         bool
	IsUnix           bool
	HasPTY           bool
	HasInotify       bool
	HasAudioAPI      bool
	PathSeparator    string
	LineEnding       string
	ExecutableExt    string
	SupportsSymlinks bool
}

// FeatureDetection holds platform feature detection results
type FeatureDetection struct {
	Platform PlatformInfo
	Features map[string]bool
}

// DetectPlatform performs platform detection and feature availability
func DetectPlatform() *FeatureDetection {
	fd := &FeatureDetection{
		Platform: PlatformInfo{
			OS:               runtime.GOOS,
			Arch:             runtime.GOARCH,
			IsWindows:        runtime.GOOS == "windows",
			IsLinux:          runtime.GOOS == "linux",
			IsDarwin:         runtime.GOOS == "darwin",
			IsUnix:           runtime.GOOS != "windows",
			PathSeparator:    string(filepath.Separator),
			LineEnding:       "\n",
			ExecutableExt:    "",
			SupportsSymlinks: true,
		},
		Features: make(map[string]bool),
	}

	// Set platform-specific values
	if fd.Platform.IsWindows {
		fd.Platform.PathSeparator = "\\"
		fd.Platform.LineEnding = "\r\n"
		fd.Platform.ExecutableExt = ".exe"
		fd.Platform.SupportsSymlinks = false
	}

	// Detect features
	fd.detectFeatures()

	return fd
}

// detectFeatures checks for platform-specific feature availability
func (fd *FeatureDetection) detectFeatures() {
	// PTY support
	fd.Platform.HasPTY = !fd.Platform.IsWindows

	// File system notifications
	fd.Platform.HasInotify = fd.Platform.IsLinux

	// Audio API (basic detection)
	fd.Platform.HasAudioAPI = true // Will be refined in audio service

	// Additional feature detection can be added here
	fd.Features["pty"] = fd.Platform.HasPTY
	fd.Features["inotify"] = fd.Platform.HasInotify
	fd.Features["audio"] = fd.Platform.HasAudioAPI
	fd.Features["symlinks"] = fd.Platform.SupportsSymlinks
	fd.Features["unix_permissions"] = fd.Platform.IsUnix
}

// HasFeature checks if a specific feature is available
func (fd *FeatureDetection) HasFeature(feature string) bool {
	available, exists := fd.Features[feature]
	return exists && available
}

// GetHomeDirectory returns the user's home directory
func (fd *FeatureDetection) GetHomeDirectory() string {
	if fd.Platform.IsWindows {
		// Windows home directory logic
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home
		}
		if home := os.Getenv("HOME"); home != "" {
			return home
		}
		return "C:\\Users\\Default"
	}

	// Unix-like systems
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return "/"
}

// GetTempDirectory returns the system's temporary directory
func (fd *FeatureDetection) GetTempDirectory() string {
	if temp := os.Getenv("TMPDIR"); temp != "" {
		return temp
	}
	if temp := os.Getenv("TEMP"); temp != "" {
		return temp
	}
	if temp := os.Getenv("TMP"); temp != "" {
		return temp
	}

	if fd.Platform.IsWindows {
		return "C:\\Temp"
	}
	return "/tmp"
}

// IsHiddenFile checks if a file is hidden on the current platform
func (fd *FeatureDetection) IsHiddenFile(filename string) bool {
	if fd.Platform.IsWindows {
		// Windows: check for hidden attribute or files starting with '.'
		return strings.HasPrefix(filename, ".") || fd.hasWindowsHiddenAttribute(filename)
	}
	// Unix-like: files starting with '.'
	return strings.HasPrefix(filename, ".")
}

// hasWindowsHiddenAttribute checks if a file has the Windows hidden attribute
func (fd *FeatureDetection) hasWindowsHiddenAttribute(path string) bool {
	// This would require Windows-specific API calls
	// For now, just check if filename starts with '.'
	return false
}

// JoinPath joins path elements using platform-appropriate separator
func (fd *FeatureDetection) JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// NormalizePath normalizes a path for the current platform
func (fd *FeatureDetection) NormalizePath(path string) string {
	if fd.Platform.IsWindows {
		path = strings.ReplaceAll(path, "/", "\\")
	} else {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	return filepath.Clean(path)
}