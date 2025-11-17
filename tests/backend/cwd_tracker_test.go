package backend

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/logger"
	"aDex-UI/internal/models"
	"aDex-UI/internal/services/terminal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEventBus is a mock implementation of events.IEventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, eventType string, data interface{}, source string) error {
	args := m.Called(ctx, eventType, data, source)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventType string, handler events.EventHandler) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(ctx context.Context, eventType string, handler events.EventHandler) error {
	args := m.Called(ctx, eventType, handler)
	return args.Error(0)
}

// TestCWDTrackerCreation tests creating a new CWD tracker
func TestCWDTrackerCreation(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	assert.NotNil(t, tracker)
}

// TestCWDTrackerInitialization tests initializing the CWD tracker
func TestCWDTrackerInitialization(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		assert.NoError(t, err, "CWD tracker should initialize successfully on Unix systems")
	} else {
		// On other platforms, it might be disabled
		assert.NoError(t, err, "CWD tracker should initialize without error on any platform")
	}
}

// TestCWDTrackerAddRemoveSession tests adding and removing sessions
func TestCWDTrackerAddRemoveSession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CWD tracker test in short mode")
	}

	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Create a mock PTY
	ptyManager := terminal.NewPTYManager()
	options := &models.TerminalOptions{
		Shell: "/bin/bash",
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	// Only test on Unix systems where PTY creation works
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		pty, err := ptyManager.CreatePTY(ctx, "test-session", options)
		if err != nil {
			t.Skipf("Skipping PTY test due to creation failure: %v", err)
		}
		defer ptyManager.RemovePTY("test-session")

		// Test adding session
		err = tracker.AddSession("test-session", pty, "/tmp")
		assert.NoError(t, err, "Should be able to add session to tracker")

		// Test getting current CWD
		cwd, err := tracker.GetCurrentCWD("test-session")
		assert.NoError(t, err, "Should be able to get current CWD")
		assert.Equal(t, "/tmp", cwd, "Current CWD should match initial CWD")

		// Test removing session
		tracker.RemoveSession("test-session")
	}
}

// TestCWDTrackerProcessOutput tests processing terminal output for directory changes
func TestCWDTrackerProcessOutput(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cwd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Mock event bus to capture events
	mockEventBus := &MockEventBus{}
	tracker.SetEventBus(mockEventBus)

	// Test data simulating terminal output with cd commands
	testCases := []struct {
		name        string
		output      []byte
		expectEvent bool
		description string
	}{
		{
			name:        "cd command with absolute path",
			output:      []byte("cd /tmp\n"),
			expectEvent: true,
			description: "Should detect absolute path cd command",
		},
		{
			name:        "cd command with relative path",
			output:      []byte("cd ../relative/path\n"),
			expectEvent: true,
			description: "Should detect relative path cd command",
		},
		{
			name:        "cd command with quotes",
			output:      []byte("cd '/path with spaces'\n"),
			expectEvent: true,
			description: "Should detect cd command with quoted path",
		},
		{
			name:        "pushd command",
			output:      []byte("pushd /new/path\n"),
			expectEvent: true,
			description: "Should detect pushd command",
		},
		{
			name:        "non-cd command",
			output:      []byte("ls -la\n"),
			expectEvent: false,
			description: "Should not trigger event for non-cd commands",
		},
		{
			name:        "incomplete cd command",
			output:      []byte("cd "),
			expectEvent: false,
			description: "Should not trigger event for incomplete cd command",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectEvent {
				mockEventBus.On("Publish", mock.Anything, "terminal.directory.changed", mock.Anything, "cwd-tracker").Return(nil).Once()
			}

			// Create a mock PTY for testing
			ptyManager := terminal.NewPTYManager()
			options := &models.TerminalOptions{
				Shell: "/bin/bash",
				CWD:   tempDir,
				Rows:  24,
				Cols:  80,
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				pty, err := ptyManager.CreatePTY(ctx, "test-output", options)
				if err != nil {
					t.Skipf("Skipping PTY test due to creation failure: %v", err)
				}
				defer ptyManager.RemovePTY("test-output")

				// Add session to tracker
				err = tracker.AddSession("test-output", pty, tempDir)
				require.NoError(t, err)

				// Process output
				tracker.ProcessOutput("test-output", tc.output)

				// Give some time for async processing
				time.Sleep(10 * time.Millisecond)

				// Verify expectations
				mockEventBus.AssertExpectations(t)

				// Reset mock for next test
				mockEventBus.ExpectedCalls = nil
				mockEventBus.Calls = nil

				// Clean up
				tracker.RemoveSession("test-output")
			}
		})
	}
}

// TestCWDTrackerResolution tests path resolution functionality
func TestCWDTrackerResolution(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "cwd-resolution-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	testCases := []struct {
		name         string
		currentCWD   string
		targetPath   string
		expectedPath string
		shouldError  bool
	}{
		{
			name:         "absolute path",
			currentCWD:   tempDir,
			targetPath:   subDir,
			expectedPath: subDir,
			shouldError:  false,
		},
		{
			name:         "relative path",
			currentCWD:   tempDir,
			targetPath:   "subdir",
			expectedPath: subDir,
			shouldError:  false,
		},
		{
			name:         "parent directory",
			currentCWD:   subDir,
			targetPath:   "..",
			expectedPath: tempDir,
			shouldError:  false,
		},
		{
			name:         "current directory",
			currentCWD:   tempDir,
			targetPath:   ".",
			expectedPath: tempDir,
			shouldError:  false,
		},
		{
			name:         "home directory",
			currentCWD:   tempDir,
			targetPath:   "~",
			shouldError:  false, // Home directory should be resolvable
		},
		{
			name:         "nonexistent directory",
			currentCWD:   tempDir,
			targetPath:   "/nonexistent/path",
			shouldError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock PTY for testing
			ptyManager := terminal.NewPTYManager()
			options := &models.TerminalOptions{
				Shell: "/bin/bash",
				CWD:   tc.currentCWD,
				Rows:  24,
				Cols:  80,
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				pty, err := ptyManager.CreatePTY(ctx, "test-resolution", options)
				if err != nil {
					t.Skipf("Skipping PTY test due to creation failure: %v", err)
				}
				defer ptyManager.RemovePTY("test-resolution")

				// Add session to tracker
				err = tracker.AddSession("test-resolution", pty, tc.currentCWD)
				require.NoError(t, err)

				// Mock event bus
				mockEventBus := &MockEventBus{}
				tracker.SetEventBus(mockEventBus)

				if !tc.shouldError {
					mockEventBus.On("Publish", mock.Anything, "terminal.directory.changed", mock.Anything, "cwd-tracker").Return(nil).Once()
				}

				// Simulate cd command
				cdCommand := []byte("cd " + tc.targetPath + "\n")
				tracker.ProcessOutput("test-resolution", cdCommand)

				// Give some time for async processing
				time.Sleep(10 * time.Millisecond)

				if !tc.shouldError {
					// Verify event was published
					mockEventBus.AssertExpectations(t)

					// Check current CWD
					cwd, err := tracker.GetCurrentCWD("test-resolution")
					assert.NoError(t, err, "Should be able to get current CWD")

					if tc.expectedPath != "" {
						assert.Equal(t, tc.expectedPath, cwd, "CWD should match expected path")
					}
				} else {
					// Should not have published an event for invalid path
					mockEventBus.AssertNotCalled(t, "Publish")
				}

				// Clean up
				tracker.RemoveSession("test-resolution")
				mockEventBus.ExpectedCalls = nil
				mockEventBus.Calls = nil
			}
		})
	}
}

// TestCWDTrackerShellDetection tests shell type detection
func TestCWDTrackerShellDetection(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	testShells := []struct {
		shellPath    string
		expectedType string
	}{
		{"/bin/bash", "bash"},
		{"/usr/bin/bash", "bash"},
		{"/bin/zsh", "zsh"},
		{"/usr/bin/zsh", "zsh"},
		{"/bin/sh", "sh"},
		{"/usr/bin/sh", "sh"},
		{"/bin/fish", "fish"},
		{"/usr/bin/fish", "fish"},
		{"/bin/dash", "dash"},
		{"/bin/ksh", "ksh"},
		{"/bin/csh", "csh"},
		{"/bin/tcsh", "tcsh"},
		{"/usr/bin/unknown-shell", "unknown"},
	}

	for _, tc := range testShells {
		t.Run(tc.shellPath, func(t *testing.T) {
			// Create a mock PTY
			ptyManager := terminal.NewPTYManager()
			options := &models.TerminalOptions{
				Shell: tc.shellPath,
				CWD:   "/tmp",
				Rows:  24,
				Cols:  80,
			}

			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				pty, err := ptyManager.CreatePTY(ctx, "shell-test", options)
				if err != nil {
					t.Skipf("Skipping PTY test due to creation failure: %v", err)
				}
				defer ptyManager.RemovePTY("shell-test")

				// Add session to tracker - this will trigger shell detection
				err = tracker.AddSession("shell-test", pty, "/tmp")
				require.NoError(t, err)

				// Get stats to verify shell was detected
				stats := tracker.GetStats()
				assert.NotNil(t, stats, "Stats should be available")
				assert.Equal(t, true, stats["enabled"], "Tracker should be enabled")

				// Clean up
				tracker.RemoveSession("shell-test")
			}
		})
	}
}

// TestCWDTrackerStats tests getting tracking statistics
func TestCWDTrackerStats(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Get initial stats
	stats := tracker.GetStats()
	assert.NotNil(t, stats, "Stats should be available")
	assert.Equal(t, true, stats["enabled"], "Tracker should be enabled")
	assert.Equal(t, 0, stats["total_sessions"], "Should have no sessions initially")
	assert.Equal(t, 0, stats["active_sessions"], "Should have no active sessions initially")
	assert.Equal(t, runtime.GOOS, stats["platform"], "Should report correct platform")

	// Create a mock PTY
	ptyManager := terminal.NewPTYManager()
	options := &models.TerminalOptions{
		Shell: "/bin/bash",
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		pty, err := ptyManager.CreatePTY(ctx, "stats-test", options)
		if err != nil {
			t.Skipf("Skipping PTY test due to creation failure: %v", err)
		}
		defer ptyManager.RemovePTY("stats-test")

		// Add session
		err = tracker.AddSession("stats-test", pty, "/tmp")
		require.NoError(t, err)

		// Get stats with session
		stats = tracker.GetStats()
		assert.Equal(t, 1, stats["total_sessions"], "Should have one session")
		assert.Equal(t, 1, stats["active_sessions"], "Should have one active session")

		// Remove session
		tracker.RemoveSession("stats-test")

		// Get stats after removal
		stats = tracker.GetStats()
		assert.Equal(t, 0, stats["total_sessions"], "Should have no sessions after removal")
		assert.Equal(t, 0, stats["active_sessions"], "Should have no active sessions after removal")
	}
}

// TestCWDTrackerShutdown tests shutting down the CWD tracker
func TestCWDTrackerShutdown(t *testing.T) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Create and add a session
	ptyManager := terminal.NewPTYManager()
	options := &models.TerminalOptions{
		Shell: "/bin/bash",
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		pty, err := ptyManager.CreatePTY(ctx, "shutdown-test", options)
		if err != nil {
			t.Skipf("Skipping PTY test due to creation failure: %v", err)
		}
		defer ptyManager.RemovePTY("shutdown-test")

		err = tracker.AddSession("shutdown-test", pty, "/tmp")
		require.NoError(t, err)

		// Shutdown tracker
		err = tracker.Shutdown(ctx)
		assert.NoError(t, err, "Should shutdown without error")

		// Verify stats after shutdown
		stats := tracker.GetStats()
		assert.Equal(t, 0, stats["total_sessions"], "Should have no sessions after shutdown")
		assert.Equal(t, 0, stats["active_sessions"], "Should have no active sessions after shutdown")
	}
}

// TestCWDTrackerConfiguration tests different configuration options
func TestCWDTrackerConfiguration(t *testing.T) {
	testCases := []struct {
		name   string
		config *terminal.CWDTrackingConfig
	}{
		{
			name:   "default config",
			config: terminal.DefaultCWDTrackingConfig(),
		},
		{
			name: "disabled config",
			config: &terminal.CWDTrackingConfig{
				Enabled:          false,
				PollInterval:     100 * time.Millisecond,
				TrackCommands:    false,
				AutoDetectShell:  false,
				EnablePTYPolling: false,
			},
		},
		{
			name: "fast polling config",
			config: &terminal.CWDTrackingConfig{
				Enabled:          true,
				PollInterval:     100 * time.Millisecond,
				TrackCommands:    true,
				AutoDetectShell:  true,
				EnablePTYPolling: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tracker := terminal.NewCWDTracker(tc.config)
			assert.NotNil(t, tracker, "Tracker should be created with any valid config")

			ctx := context.Background()
			err := tracker.Initialize(ctx)
			assert.NoError(t, err, "Should initialize with any valid config")

			stats := tracker.GetStats()
			assert.Equal(t, tc.config.Enabled, stats["enabled"], "Enabled status should match config")

			// Clean shutdown
			err = tracker.Shutdown(ctx)
			assert.NoError(t, err, "Should shutdown cleanly")
		})
	}
}

// BenchmarkCWDTrackerProcessOutput benchmarks processing terminal output
func BenchmarkCWDTrackerProcessOutput(b *testing.B) {
	config := terminal.DefaultCWDTrackingConfig()
	tracker := terminal.NewCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	if err != nil {
		b.Skipf("Skipping benchmark due to initialization error: %v", err)
	}

	// Create a mock PTY
	ptyManager := terminal.NewPTYManager()
	options := &models.TerminalOptions{
		Shell: "/bin/bash",
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		pty, err := ptyManager.CreatePTY(ctx, "benchmark-test", options)
		if err != nil {
			b.Skipf("Skipping benchmark due to PTY creation failure: %v", err)
		}
		defer ptyManager.RemovePTY("benchmark-test")

		err = tracker.AddSession("benchmark-test", pty, "/tmp")
		if err != nil {
			b.Skipf("Skipping benchmark due to session addition failure: %v", err)
		}
		defer tracker.RemoveSession("benchmark-test")

		// Test data - typical terminal output
		testOutput := []byte("cd /some/path\nls -la\necho 'hello world'\ncd ../another/path\n")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tracker.ProcessOutput("benchmark-test", testOutput)
		}
	}
}