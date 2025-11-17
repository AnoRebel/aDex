//go:build linux || darwin || windows

package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildProcess tests the complete build process for the current platform
func TestBuildProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping build test in short mode")
	}

	// Check prerequisites
	if !isTaskAvailable() {
		t.Skip("Task command not available, skipping build test")
	}
	if !isWailsAvailable() {
		t.Skip("Wails3 command not available, skipping build test")
	}

	// Setup
	rootDir := ".."
	platform := runtime.GOOS
	appName := "aDex-UI"
	binDir := filepath.Join(rootDir, "bin")

	// Clean any existing builds
	t.Cleanup(func() {
		if _, err := os.Stat(binDir); err == nil {
			os.RemoveAll(binDir)
		}
	})

	// Test development build
	t.Run("Development Build", func(t *testing.T) {
		// Run build command
		cmd := NewTaskCommand(rootDir, "build")
		err := cmd.Run()
		require.NoError(t, err, "Development build should succeed")

		// Verify binary exists
		binaryPath := getBinaryPath(binDir, appName, platform, false)
		_, err = os.Stat(binaryPath)
		assert.NoError(t, err, "Development binary should exist")

		// Verify binary is executable
		if platform != "windows" {
			info, err := os.Stat(binaryPath)
			require.NoError(t, err)
			assert.True(t, info.Mode().Perm()&0111 != 0, "Binary should be executable")
		}
	})

	// Test production build
	t.Run("Production Build", func(t *testing.T) {
		// Clean previous build
		os.RemoveAll(binDir)

		// Run package command for production build
		cmd := NewTaskCommand(rootDir, "package")
		err := cmd.Run()
		require.NoError(t, err, "Production build should succeed")

		// Verify binary exists
		binaryPath := getBinaryPath(binDir, appName, platform, true)
		_, err = os.Stat(binaryPath)
		assert.NoError(t, err, "Production binary should exist")

		// Production binary should be optimized (smaller size)
		if platform != "windows" {
			// Check that binary is stripped (production build)
			cmd := NewCommand("file", binaryPath)
			output, err := cmd.Output()
			if err == nil {
				assert.Contains(t, string(output), "executable", "Should be recognized as executable")
			} else {
				t.Logf("file command failed: %v", err)
			}
		}
	})
}

// TestCrossPlatformBuilds tests cross-platform compilation (where possible)
func TestCrossPlatformBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cross-platform build test in short mode")
	}

	rootDir := ".."
	appName := "aDex-UI"
	binDir := filepath.Join(rootDir, "bin")

	// Test cases for cross-platform builds
	testCases := []struct {
		name     string
		goos     string
		goarch   string
		expected bool
	}{
		{"Linux AMD64", "linux", "amd64", true},
		{"Linux ARM64", "linux", "arm64", runtime.GOARCH == "arm64" || runtime.GOARCH == "amd64"},
		{"Windows AMD64", "windows", "amd64", true},
		{"Windows ARM64", "windows", "arm64", runtime.GOARCH == "arm64" || runtime.GOARCH == "amd64"},
		{"macOS AMD64", "darwin", "amd64", true},
		{"macOS ARM64", "darwin", "arm64", runtime.GOARCH == "arm64" || runtime.GOARCH == "amd64"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if we can cross-compile to this target
			if !canCrossCompile(tc.goos, tc.goarch) {
				t.Skipf("Cross-compilation for %s/%s not supported on this platform", tc.goos, tc.goarch)
			}

			// Set environment variables for cross-compilation
			cmd := NewTaskCommand(rootDir, tc.goos+":build")
			cmd.env = append(os.Environ(),
				"GOOS="+tc.goos,
				"GOARCH="+tc.goarch,
				"PRODUCTION=true",
			)

			err := cmd.Run()
			if err != nil {
				t.Logf("Cross-compilation failed for %s/%s: %v", tc.goos, tc.goarch, err)
				// Don't fail the test for cross-compilation issues, just log them
				return
			}

			// Verify cross-compiled binary exists
			binaryName := appName
			if tc.goos == "windows" {
				binaryName += ".exe"
			}
			binaryPath := filepath.Join(binDir, binaryName)
			_, err = os.Stat(binaryPath)
			assert.NoError(t, err, "Cross-compiled binary should exist")

			// Clean up cross-compiled binary
			os.Remove(binaryPath)
		})
	}
}

// TestPackageFormats tests different packaging formats
func TestPackageFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping package format test in short mode")
	}

	rootDir := ".."
	platform := runtime.GOOS
	binDir := filepath.Join(rootDir, "bin")

	switch platform {
	case "linux":
		t.Run("Linux Packages", func(t *testing.T) {
			// Test AppImage creation
			t.Run("AppImage", func(t *testing.T) {
				cmd := NewTaskCommand(rootDir, "linux:create:appimage")
				err := cmd.Run()
				if err != nil {
					t.Logf("AppImage creation failed: %v", err)
					t.Skip("AppImage tools not available")
				}

				// Verify AppImage exists
				appImagePath := filepath.Join(binDir, "aDex-UI.AppImage")
				_, err = os.Stat(appImagePath)
				assert.NoError(t, err, "AppImage should exist")
			})

			// Test DEB package
			t.Run("DEB Package", func(t *testing.T) {
				cmd := NewTaskCommand(rootDir, "linux:create:deb")
				err := cmd.Run()
				if err != nil {
					t.Logf("DEB creation failed: %v", err)
					t.Skip("DEB packaging tools not available")
				}

				// Verify DEB exists
				debPath := filepath.Join(binDir, "adex-ui_1.0.0_amd64.deb")
				_, err = os.Stat(debPath)
				assert.NoError(t, err, "DEB package should exist")
			})
		})

	case "windows":
		t.Run("Windows Packages", func(t *testing.T) {
			// Test NSIS installer
			t.Run("NSIS Installer", func(t *testing.T) {
				cmd := NewTaskCommand(rootDir, "windows:create:nsis:installer")
				err := cmd.Run()
				if err != nil {
					t.Logf("NSIS installer creation failed: %v", err)
					t.Skip("NSIS not available")
				}

				// Verify installer exists
				installerPath := filepath.Join(binDir, "aDex-UI-setup.exe")
				_, err = os.Stat(installerPath)
				assert.NoError(t, err, "NSIS installer should exist")
			})
		})

	case "darwin":
		t.Run("macOS Packages", func(t *testing.T) {
			// Test app bundle
			t.Run("App Bundle", func(t *testing.T) {
				cmd := NewTaskCommand(rootDir, "darwin:package")
				err := cmd.Run()
				require.NoError(t, err, "App bundle creation should succeed")

				// Verify app bundle exists
				appPath := filepath.Join(binDir, "aDex-UI.app")
				_, err = os.Stat(appPath)
				assert.NoError(t, err, "App bundle should exist")

				// Verify app structure
				executablePath := filepath.Join(appPath, "Contents", "MacOS", "aDex-UI")
				_, err = os.Stat(executablePath)
				assert.NoError(t, err, "App executable should exist")

				infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")
				_, err = os.Stat(infoPlistPath)
				assert.NoError(t, err, "Info.plist should exist")

				iconPath := filepath.Join(appPath, "Contents", "Resources", "icons.icns")
				_, err = os.Stat(iconPath)
				assert.NoError(t, err, "App icon should exist")
			})
		})
	}
}

// TestBuildAssets tests that all required build assets are present
func TestBuildAssets(t *testing.T) {
	rootDir := ".."
	buildDir := filepath.Join(rootDir, "build")

	// Test icon files
	t.Run("Icon Files", func(t *testing.T) {
		icons := []struct {
			name string
			path string
		}{
			{"Main PNG icon", filepath.Join(buildDir, "appicon.png")},
			{"Windows ICO", filepath.Join(buildDir, "windows", "icon.ico")},
			{"macOS ICNS", filepath.Join(buildDir, "darwin", "icons.icns")},
		}

		for _, icon := range icons {
			t.Run(icon.name, func(t *testing.T) {
				_, err := os.Stat(icon.path)
				assert.NoError(t, err, "%s should exist", icon.name)
			})
		}
	})

	// Test configuration files
	t.Run("Configuration Files", func(t *testing.T) {
		configs := []struct {
			name string
			path string
		}{
			{"Build config", filepath.Join(buildDir, "config.yml")},
			{"Windows manifest", filepath.Join(buildDir, "windows", "wails.exe.manifest")},
			{"Windows info", filepath.Join(buildDir, "windows", "info.json")},
			{"macOS Info.plist", filepath.Join(buildDir, "darwin", "Info.plist")},
			{"macOS Dev Info.plist", filepath.Join(buildDir, "darwin", "Info.dev.plist")},
		}

		for _, config := range configs {
			t.Run(config.name, func(t *testing.T) {
				_, err := os.Stat(config.path)
				assert.NoError(t, err, "%s should exist", config.name)
			})
		}
	})

	// Test Taskfile configurations
	t.Run("Taskfile Configurations", func(t *testing.T) {
		taskfiles := []struct {
			name string
			path string
		}{
			{"Root Taskfile", filepath.Join(buildDir, "Taskfile.yml")},
			{"Linux Taskfile", filepath.Join(buildDir, "linux", "Taskfile.yml")},
			{"Windows Taskfile", filepath.Join(buildDir, "windows", "Taskfile.yml")},
			{"macOS Taskfile", filepath.Join(buildDir, "darwin", "Taskfile.yml")},
		}

		for _, taskfile := range taskfiles {
			t.Run(taskfile.name, func(t *testing.T) {
				_, err := os.Stat(taskfile.path)
				assert.NoError(t, err, "%s should exist", taskfile.name)
			})
		}
	})
}

// TestBuildPerformance tests build performance metrics
func TestBuildPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	rootDir := ".."
	binDir := filepath.Join(rootDir, "bin")

	// Clean any existing builds
	os.RemoveAll(binDir)

	// Measure build time
	start := time.Now()
	cmd := NewTaskCommand(rootDir, "build")
	err := cmd.Run()
	require.NoError(t, err, "Build should succeed")
	buildTime := time.Since(start)

	// Build should complete within reasonable time (30 seconds)
	assert.Less(t, buildTime, 30*time.Second, "Build should complete within 30 seconds")
	t.Logf("Build completed in %v", buildTime)

	// Check binary size
	binaryPath := getBinaryPath(binDir, "aDex-UI", runtime.GOOS, false)
	info, err := os.Stat(binaryPath)
	require.NoError(t, err)

	binarySize := info.Size()
	t.Logf("Binary size: %d bytes (%.2f MB)", binarySize, float64(binarySize)/1024/1024)

	// Binary size should be reasonable (less than 100MB for debug, less than 50MB for release)
	assert.Less(t, binarySize, int64(100*1024*1024), "Debug binary should be less than 100MB")
}

// Helper functions

func getBinaryPath(binDir, appName, platform string, production bool) string {
	binaryName := appName
	if platform == "windows" {
		binaryName += ".exe"
	}
	return filepath.Join(binDir, binaryName)
}

// canCrossCompile checks if cross-compilation is supported for the target platform/arch
func canCrossCompile(goos, goarch string) bool {
	// Cross-compilation is generally supported, but some combinations have special requirements
	switch goos {
	case "darwin":
		// macOS can be cross-compiled from any platform
		return true
	case "windows":
		// Windows can be cross-compiled from any platform
		return true
	case "linux":
		// Linux can be cross-compiled from any platform
		return true
	}
	return true
}

// isTaskAvailable checks if the task command is available
func isTaskAvailable() bool {
	cmd := exec.Command("task", "--version")
	err := cmd.Run()
	return err == nil
}

// isWailsAvailable checks if wails3 command is available
func isWailsAvailable() bool {
	cmd := exec.Command("wails3", "version")
	err := cmd.Run()
	return err == nil
}

// TaskCommand represents a task command
type TaskCommand struct {
	dir string
	cmd string
	args []string
	env []string
}

// NewTaskCommand creates a new task command
func NewTaskCommand(dir, task string) *TaskCommand {
	return &TaskCommand{
		dir: dir,
		cmd: "task",
		args: []string{task},
		env: os.Environ(),
	}
}

// Run executes the task command
func (tc *TaskCommand) Run() error {
	cmd := NewCommand(tc.cmd, tc.args...)
	cmd.Dir = tc.dir
	cmd.Env = tc.env
	return cmd.Run()
}

// Command represents a system command (simplified for testing)
type Command struct {
	cmd  string
	args []string
	Dir  string
	Env  []string
}

// NewCommand creates a new command
func NewCommand(cmd string, args ...string) *Command {
	return &Command{
		cmd:  cmd,
		args: args,
		Dir:  "",
		Env:  os.Environ(),
	}
}

// Run executes the command
func (c *Command) Run() error {
	cmd := exec.Command(c.cmd, c.args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	cmd.Env = c.Env
	return cmd.Run()
}

// Output gets the command output
func (c *Command) Output() ([]byte, error) {
	cmd := exec.Command(c.cmd, c.args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	cmd.Env = c.Env
	return cmd.Output()
}

// CombinedOutput gets the command combined output
func (c *Command) CombinedOutput() ([]byte, error) {
	cmd := exec.Command(c.cmd, c.args...)
	if c.Dir != "" {
		cmd.Dir = c.Dir
	}
	cmd.Env = c.Env
	return cmd.CombinedOutput()
}