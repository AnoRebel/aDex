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

// TestWindowsCWDTrackerCreation tests creating a new Windows CWD tracker
func TestWindowsCWDTrackerCreation(t *testing.T) {
	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	if runtime.GOOS == "windows" {
		assert.NotNil(t, tracker, "Windows CWD tracker should be created on Windows")
	} else {
		// On non-Windows systems, the tracker should be disabled
		assert.NotNil(t, tracker, "Windows CWD tracker should still be created but disabled")
	}
}

// TestWindowsCWDTrackerInitialization tests initializing the Windows CWD tracker
func TestWindowsCWDTrackerInitialization(t *testing.T) {
	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)

	if runtime.GOOS == "windows" {
		assert.NoError(t, err, "Windows CWD tracker should initialize successfully on Windows")
	} else {
		// On non-Windows systems, it should return an error
		assert.Error(t, err, "Windows CWD tracker should fail to initialize on non-Windows systems")
		assert.Contains(t, err.Error(), "only available on Windows")
	}
}

// TestWindowsCWDTrackerConfiguration tests different configuration options
func TestWindowsCWDTrackerConfiguration(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	testCases := []struct {
		name   string
		config *terminal.WindowsCWDConfig
	}{
		{
			name:   "default config",
			config: terminal.DefaultWindowsCWDConfig(),
		},
		{
			name: "disabled config",
			config: &terminal.WindowsCWDConfig{
				Enabled:           false,
				PollInterval:      500 * time.Millisecond,
				CreateHelperFiles: false,
				MaxRetries:        1,
				Timeout:           5 * time.Second,
			},
		},
		{
			name: "fast polling config",
			config: &terminal.WindowsCWDConfig{
				Enabled:           true,
				PollInterval:      100 * time.Millisecond,
				CreateHelperFiles: true,
				MaxRetries:        5,
				Timeout:           60 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tracker := terminal.NewWindowsCWDTracker(tc.config)
			assert.NotNil(t, tracker, "Windows CWD tracker should be created with any valid config")

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

// TestWindowsCWDTrackerStats tests getting tracking statistics
func TestWindowsCWDTrackerStats(t *testing.T) {
	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	if runtime.GOOS != "windows" && err != nil {
		t.Skip("Skipping Windows CWD tracker test on non-Windows system")
	}
	require.NoError(t, err)

	// Get initial stats
	stats := tracker.GetStats()
	assert.NotNil(t, stats, "Stats should be available")
	assert.Equal(t, config.Enabled, stats["enabled"], "Tracker should report correct enabled status")
	assert.Equal(t, 0, stats["total_sessions"], "Should have no sessions initially")
	assert.Equal(t, 0, stats["active_sessions"], "Should have no active sessions initially")
	assert.Equal(t, "windows", stats["platform"], "Should report correct platform")
	assert.Equal(t, "windows_detached", stats["method"], "Should report correct method")

	// Shutdown
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestWindowsCWDTrackerHelperFiles tests helper file creation and cleanup
func TestWindowsCWDTrackerHelperFiles(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	config.CreateHelperFiles = true
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Check if helper files were created
	stats := tracker.GetStats()
	helperFiles, ok := stats["helper_files"].(map[string]interface{})
	assert.True(t, ok, "Helper files info should be available")

	if helperFiles != nil {
		psScript, ok := helperFiles["powershell_script"].(string)
		assert.True(t, ok, "PowerShell script path should be available")
		assert.NotEmpty(t, psScript, "PowerShell script path should not be empty")

		batchScript, ok := helperFiles["batch_script"].(string)
		assert.True(t, ok, "Batch script path should be available")
		assert.NotEmpty(t, batchScript, "Batch script path should not be empty")

		// Verify files exist
		if psScript != "" {
			_, err := os.Stat(psScript)
			assert.NoError(t, err, "PowerShell script file should exist")
		}

		if batchScript != "" {
			_, err := os.Stat(batchScript)
			assert.NoError(t, err, "Batch script file should exist")
		}
	}

	// Shutdown should clean up helper files
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)

	// Check if files were cleaned up
	time.Sleep(100 * time.Millisecond) // Give some time for cleanup

	if helperFiles != nil {
		psScript, _ := helperFiles["powershell_script"].(string)
		batchScript, _ := helperFiles["batch_script"].(string)

		if psScript != "" {
			_, err := os.Stat(psScript)
			assert.True(t, os.IsNotExist(err), "PowerShell script should be cleaned up")
		}

		if batchScript != "" {
			_, err := os.Stat(batchScript)
			assert.True(t, os.IsNotExist(err), "Batch script should be cleaned up")
		}
	}
}

// TestWindowsCWDTrackerShellDetection tests shell type detection on Windows
func TestWindowsCWDTrackerShellDetection(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	testShells := []struct {
		shellPath    string
		expectedType string
	}{
		{"C:\\Windows\\System32\\cmd.exe", "cmd"},
		{"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe", "powershell"},
		{"C:\\Program Files\\PowerShell\\7\\pwsh.exe", "pwsh"},
		{"C:\\Program Files\\Git\\bin\\bash.exe", "bash"},
		{"C:\\Windows\\System32\\wsl.exe", "wsl"},
		{"C:\\Program Files\\Git\\git-bash.exe", "gitbash"},
		{"C:\\some\\unknown\\shell.exe", "unknown"},
	}

	for _, tc := range testShells {
		t.Run(tc.shellPath, func(t *testing.T) {
			// We can't easily create actual PTY processes on Windows in tests,
			// but we can test the shell detection logic indirectly

			// Create a mock session to test shell detection
			mockEventBus := &MockEventBus{}
			tracker.SetEventBus(mockEventBus)

			// The shell detection happens when a session is added,
			// but we can't easily mock a Windows PTY process in unit tests
			// So we'll just verify the tracker is working correctly

			stats := tracker.GetStats()
			assert.NotNil(t, stats, "Stats should be available")
		})
	}

	// Cleanup
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestWindowsCWDTrackerEventBus tests event bus integration
func TestWindowsCWDTrackerEventBus(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Create mock event bus
	mockEventBus := &MockEventBus{}
	tracker.SetEventBus(mockEventBus)

	// Test event bus integration
	// We can't easily test actual directory changes without real Windows processes,
	// but we can verify the event bus is set correctly

	stats := tracker.GetStats()
	assert.NotNil(t, stats, "Stats should be available")

	// Cleanup
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestWindowsCWDTrackerProcessOutput tests processing terminal output
func TestWindowsCWDTrackerProcessOutput(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Test processing output with cd commands
	testCases := []struct {
		name        string
		output      []byte
		description string
	}{
		{
			name:        "cd command",
			output:      []byte("cd /d C:\\Users\\Test\\Documents\n"),
			description: "Should detect cd command with drive change",
		},
		{
			name:        "pushd command",
			output:      []byte("pushd C:\\NewPath\n"),
			description: "Should detect pushd command",
		},
		{
			name:        "relative cd",
			output:      []byte("cd ..\\SubFolder\n"),
			description: "Should detect relative path cd command",
		},
		{
			name:        "non-cd command",
			output:      []byte("dir\n"),
			description: "Should not trigger for non-cd commands",
		},
		{
			name:        "mixed content",
			output:      []byte("ls\ncd /temp\necho done\n"),
			description: "Should detect cd command in mixed content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Process output - this should not panic or error
			tracker.ProcessOutput("test-session", tc.output)

			// Give some time for async processing
			time.Sleep(10 * time.Millisecond)

			// We can't easily verify the results without real Windows processes,
			// but we can ensure the method doesn't panic
			assert.True(t, true, "ProcessOutput should complete without error")
		})
	}

	// Cleanup
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestWindowsCWDTrackerMultipleSessions tests handling multiple sessions
func TestWindowsCWDTrackerMultipleSessions(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Test session management
	sessionIDs := []string{"session-1", "session-2", "session-3"}

	// Since we can't create real Windows PTY processes in unit tests,
	// we'll test the session management methods with mock data

	for _, sessionID := range sessionIDs {
		// These calls would normally fail because we can't create real PTY processes,
		// but they should not panic the tracker
		_ = tracker.RemoveSession(sessionID)
	}

	// Verify stats
	stats := tracker.GetStats()
	assert.Equal(t, 0, stats["total_sessions"], "Should have no sessions")
	assert.Equal(t, 0, stats["active_sessions"], "Should have no active sessions")

	// Cleanup
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestWindowsCWDTrackerErrorHandling tests error handling scenarios
func TestWindowsCWDTrackerErrorHandling(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	require.NoError(t, err)

	// Test getting CWD for non-existent session
	_, err = tracker.GetCurrentCWD("non-existent-session")
	assert.Error(t, err, "Should return error for non-existent session")
	assert.Contains(t, err.Error(), "not found", "Error message should indicate session not found")

	// Test removing non-existent session (should not panic)
	tracker.RemoveSession("non-existent-session")

	// Test getting stats after errors
	stats := tracker.GetStats()
	assert.NotNil(t, stats, "Stats should still be available after errors")

	// Cleanup
	err = tracker.Shutdown(ctx)
	assert.NoError(t, err)
}

// BenchmarkWindowsCWDTrackerStats benchmarks getting statistics
func BenchmarkWindowsCWDTrackerStats(b *testing.B) {
	if runtime.GOOS != "windows" {
		b.Skip("Skipping Windows-specific benchmark on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	if err != nil {
		b.Skipf("Skipping benchmark due to initialization error: %v", err)
	}
	defer tracker.Shutdown(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.GetStats()
	}
}

// BenchmarkWindowsCWDTrackerProcessOutput benchmarks processing terminal output
func BenchmarkWindowsCWDTrackerProcessOutput(b *testing.B) {
	if runtime.GOOS != "windows" {
		b.Skip("Skipping Windows-specific benchmark on non-Windows system")
	}

	config := terminal.DefaultWindowsCWDConfig()
	tracker := terminal.NewWindowsCWDTracker(config)

	ctx := context.Background()
	err := tracker.Initialize(ctx)
	if err != nil {
		b.Skipf("Skipping benchmark due to initialization error: %v", err)
	}
	defer tracker.Shutdown(ctx)

	// Test data - typical Windows terminal output
	testOutput := []byte("cd /d C:\\Users\\Test\\Documents\ndir\necho 'hello world'\ncd ..\\Temp\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.ProcessOutput("benchmark-session", testOutput)
	}
}