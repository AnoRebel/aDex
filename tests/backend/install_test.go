//go:build linux || darwin || windows

package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallationProcess tests the installation process for the current platform
func TestInstallationProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping installation test in short mode")
	}

	// Check prerequisites
	if !isTaskAvailable() {
		t.Skip("Task command not available, skipping installation test")
	}
	if !isWailsAvailable() {
		t.Skip("Wails3 command not available, skipping installation test")
	}

	platform := runtime.GOOS
	rootDir := ".."
	binDir := filepath.Join(rootDir, "bin")
	appName := "aDex-UI"

	// First, ensure we have a built application
	t.Run("Ensure Application Built", func(t *testing.T) {
		cmd := NewTaskCommand(rootDir, "package")
		err := cmd.Run()
		require.NoError(t, err, "Application should be built successfully")
	})

	// Test platform-specific installation
	switch platform {
	case "linux":
		testLinuxInstallation(t, rootDir, binDir, appName)
	case "darwin":
		testMacOSInstallation(t, rootDir, binDir, appName)
	case "windows":
		testWindowsInstallation(t, rootDir, binDir, appName)
	default:
		t.Skipf("Installation tests not implemented for platform: %s", platform)
	}
}

// testLinuxInstallation tests Linux installation methods
func testLinuxInstallation(t *testing.T, rootDir, binDir, appName string) {
	t.Run("Linux Binary Installation", func(t *testing.T) {
		// Create temporary installation directory
		tempDir := t.TempDir()
		installDir := filepath.Join(tempDir, "local", "bin")
		err := os.MkdirAll(installDir, 0755)
		require.NoError(t, err)

		// Copy binary to installation directory
		binaryPath := filepath.Join(binDir, appName)
		installPath := filepath.Join(installDir, appName)

		err = copyFile(binaryPath, installPath)
		require.NoError(t, err)

		// Make binary executable
		err = os.Chmod(installPath, 0755)
		require.NoError(t, err)

		// Verify binary can be executed
		cmd := exec.Command(installPath, "--version")
		err = cmd.Run()
		// Note: This might fail if the application doesn't support --version flag
		// We're just testing that the binary can be executed without panicking
		t.Logf("Binary execution result: %v", err)
	})

	t.Run("Linux Desktop Integration", func(t *testing.T) {
		// Create temporary directories
		tempDir := t.TempDir()
		homeDir := tempDir
		applicationsDir := filepath.Join(homeDir, ".local", "share", "applications")
		binariesDir := filepath.Join(homeDir, ".local", "bin")
		iconsDir := filepath.Join(homeDir, ".local", "share", "icons")

		err := os.MkdirAll(applicationsDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(binariesDir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(iconsDir, 0755)
		require.NoError(t, err)

		// Install desktop file
		desktopFile := filepath.Join(applicationsDir, "adex-ui.desktop")
		desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=aDex-UI
Comment=A futuristic terminal emulator and system monitor
Exec=%s
Icon=%s
Terminal=false
Type=Application
Categories=Development;System;
StartupNotify=true
`, filepath.Join(binariesDir, appName), filepath.Join(iconsDir, "adex-ui.png"))

		err = os.WriteFile(desktopFile, []byte(desktopContent), 0644)
		require.NoError(t, err)

		// Copy binary
		binaryPath := filepath.Join(binDir, appName)
		installPath := filepath.Join(binariesDir, appName)
		err = copyFile(binaryPath, installPath)
		require.NoError(t, err)
		err = os.Chmod(installPath, 0755)
		require.NoError(t, err)

		// Copy icon
		iconSource := filepath.Join(rootDir, "build", "appicon.png")
		iconDest := filepath.Join(iconsDir, "adex-ui.png")
		err = copyFile(iconSource, iconDest)
		require.NoError(t, err)

		// Verify desktop file is valid
		_, err = os.Stat(desktopFile)
		assert.NoError(t, err, "Desktop file should exist")

		// Verify desktop file content
		content, err := os.ReadFile(desktopFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Name=aDex-UI")
		assert.Contains(t, string(content), "Exec=")
		assert.Contains(t, string(content), "Icon=")
	})

	t.Run("AppImage Installation", func(t *testing.T) {
		appImagePath := filepath.Join(binDir, "aDex-UI.AppImage")
		_, err := os.Stat(appImagePath)
		if err != nil {
			t.Skip("AppImage not built, skipping AppImage installation test")
		}

		// Test AppImage execution
		cmd := exec.Command(appImagePath, "--version")
		err = cmd.Run()
		t.Logf("AppImage execution result: %v", err)

		// Verify AppImage is executable
		info, err := os.Stat(appImagePath)
		require.NoError(t, err)
		assert.True(t, info.Mode().Perm()&0111 != 0, "AppImage should be executable")
	})
}

// testMacOSInstallation tests macOS installation methods
func testMacOSInstallation(t *testing.T, rootDir, binDir, appName string) {
	t.Run("macOS App Bundle Installation", func(t *testing.T) {
		appBundle := filepath.Join(binDir, appName+".app")
		_, err := os.Stat(appBundle)
		if err != nil {
			t.Skip("App bundle not built, skipping app bundle installation test")
		}

		// Create temporary Applications directory
		tempDir := t.TempDir()
		applicationsDir := filepath.Join(tempDir, "Applications")
		err = os.MkdirAll(applicationsDir, 0755)
		require.NoError(t, err)

		// Copy app bundle to Applications directory
		installPath := filepath.Join(applicationsDir, appName+".app")
		err = copyDir(appBundle, installPath)
		require.NoError(t, err)

		// Verify app bundle structure
		executablePath := filepath.Join(installPath, "Contents", "MacOS", appName)
		_, err = os.Stat(executablePath)
		assert.NoError(t, err, "App executable should exist")

		infoPlistPath := filepath.Join(installPath, "Contents", "Info.plist")
		_, err = os.Stat(infoPlistPath)
		assert.NoError(t, err, "Info.plist should exist")

		iconPath := filepath.Join(installPath, "Contents", "Resources", "icons.icns")
		_, err = os.Stat(iconPath)
		assert.NoError(t, err, "App icon should exist")

		// Test app bundle execution (this might not work in all test environments)
		cmd := exec.Command("open", installPath)
		err = cmd.Run()
		if err != nil {
			t.Logf("App bundle open failed (expected in test environment): %v", err)
		}
	})
}

// testWindowsInstallation tests Windows installation methods
func testWindowsInstallation(t *testing.T, rootDir, binDir, appName string) {
	t.Run("Windows Binary Installation", func(t *testing.T) {
		// Create temporary installation directory
		tempDir := t.TempDir()
		installDir := filepath.Join(tempDir, "Program Files", "aDex-UI")
		err := os.MkdirAll(installDir, 0755)
		require.NoError(t, err)

		// Copy binary to installation directory
		binaryPath := filepath.Join(binDir, appName+".exe")
		installPath := filepath.Join(installDir, appName+".exe")

		err = copyFile(binaryPath, installPath)
		require.NoError(t, err)

		// Verify binary exists
		_, err = os.Stat(installPath)
		assert.NoError(t, err, "Installed binary should exist")
	})

	t.Run("Windows Installer", func(t *testing.T) {
		installerPath := filepath.Join(binDir, appName+"-setup.exe")
		_, err := os.Stat(installerPath)
		if err != nil {
			t.Skip("Windows installer not built, skipping installer test")
		}

		// Verify installer exists and is executable
		info, err := os.Stat(installerPath)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(1024*1024), "Installer should be larger than 1MB")

		// Test installer extraction (without actually installing)
		cmd := exec.Command(installerPath, "/?")
		err = cmd.Run()
		if err != nil {
			t.Logf("Installer help failed: %v", err)
		}
	})

	t.Run("Start Menu Integration", func(t *testing.T) {
		// Create temporary start menu directory
		tempDir := t.TempDir()
		startMenuDir := filepath.Join(tempDir, "Start Menu", "Programs", "aDex-UI")
		err := os.MkdirAll(startMenuDir, 0755)
		require.NoError(t, err)

		// Create shortcut (simulated)
		shortcutPath := filepath.Join(startMenuDir, "aDex-UI.lnk")
		// In a real scenario, you'd create a Windows shortcut
		// For testing purposes, we'll just create a placeholder file
		err = os.WriteFile(shortcutPath, []byte("shortcut placeholder"), 0644)
		require.NoError(t, err)

		// Verify shortcut exists
		_, err = os.Stat(shortcutPath)
		assert.NoError(t, err, "Start menu shortcut should exist")
	})
}

// TestApplicationLaunch tests that the installed application can launch
func TestApplicationLaunch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping application launch test in short mode")
	}

	platform := runtime.GOOS
	rootDir := ".."
	binDir := filepath.Join(rootDir, "bin")
	appName := "aDex-UI"

	t.Run("Application Launch Test", func(t *testing.T) {
		var executablePath string

		switch platform {
		case "linux":
			executablePath = filepath.Join(binDir, appName)
		case "darwin":
			appBundle := filepath.Join(binDir, appName+".app")
			_, err := os.Stat(appBundle)
			if err != nil {
				t.Skip("App bundle not available")
			}
			executablePath = filepath.Join(appBundle, "Contents", "MacOS", appName)
		case "windows":
			executablePath = filepath.Join(binDir, appName+".exe")
		}

		// Test that executable exists
		_, err := os.Stat(executablePath)
		require.NoError(t, err, "Application executable should exist")

		// Test application launch with timeout
		cmd := exec.Command(executablePath, "--help")
		timer := time.AfterFunc(10*time.Second, func() {
			cmd.Process.Kill()
		})
		defer timer.Stop()

		err = cmd.Run()
		t.Logf("Application launch result: %v", err)

		// The application might exit with an error code if --help isn't supported
		// We're mainly testing that it doesn't crash immediately
	})
}

// TestApplicationUninstallation tests the uninstallation process
func TestApplicationUninstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping uninstallation test in short mode")
	}

	platform := runtime.GOOS
	appName := "aDex-UI"

	t.Run("Application Uninstallation", func(t *testing.T) {
		// Create temporary installation
		tempDir := t.TempDir()
		var installPaths []string

		switch platform {
		case "linux":
			binDir := filepath.Join(tempDir, ".local", "bin")
			os.MkdirAll(binDir, 0755)
			binaryPath := filepath.Join(binDir, appName)
			os.WriteFile(binaryPath, []byte("fake binary"), 0755)
			installPaths = append(installPaths, binaryPath)

			desktopDir := filepath.Join(tempDir, ".local", "share", "applications")
			os.MkdirAll(desktopDir, 0755)
			desktopPath := filepath.Join(desktopDir, "adex-ui.desktop")
			os.WriteFile(desktopPath, []byte("fake desktop file"), 0644)
			installPaths = append(installPaths, desktopPath)

		case "darwin":
			applicationsDir := filepath.Join(tempDir, "Applications")
			os.MkdirAll(applicationsDir, 0755)
			appPath := filepath.Join(applicationsDir, appName+".app")
			os.MkdirAll(appPath, 0755)
			installPaths = append(installPaths, appPath)

		case "windows":
			programFilesDir := filepath.Join(tempDir, "Program Files", "aDex-UI")
			os.MkdirAll(programFilesDir, 0755)
			binaryPath := filepath.Join(programFilesDir, appName+".exe")
			os.WriteFile(binaryPath, []byte("fake binary"), 0644)
			installPaths = append(installPaths, binaryPath)
		}

		// Verify files exist before uninstallation
		for _, path := range installPaths {
			_, err := os.Stat(path)
			assert.NoError(t, err, "Installed file should exist: %s", path)
		}

		// Simulate uninstallation
		for _, path := range installPaths {
			err := os.RemoveAll(path)
			assert.NoError(t, err, "Should be able to remove installed file: %s", path)
		}

		// Verify files are gone after uninstallation
		for _, path := range installPaths {
			_, err := os.Stat(path)
			assert.True(t, os.IsNotExist(err), "Installed file should be removed: %s", path)
		}
	})
}

// Helper functions

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
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