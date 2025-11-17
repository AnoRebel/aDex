package backend

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerminalSessionCreation tests the core terminal session creation functionality
func TestTerminalSessionCreation(t *testing.T) {
	// Skip on Windows for now as PTY support is different
	if runtime.GOOS == "windows" {
		t.Skip("PTY tests skipped on Windows - different implementation needed")
	}

	ctx := context.Background()

	// Test 1: Create a terminal service
	termService := terminal.NewService()
	require.NotNil(t, termService, "Terminal service should be created")

	// Test 2: Initialize the service
	err := termService.Initialize(ctx)
	assert.NoError(t, err, "Terminal service should initialize without errors")

	// Test 3: Create a basic terminal session
	options := &terminal.TerminalOptions{
		Shell: getDefaultShell(),
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	session, err := termService.CreateSession(ctx, options)
	require.NoError(t, err, "Terminal session should be created successfully")
	require.NotNil(t, session, "Session should not be nil")
	require.NotEmpty(t, session.ID, "Session should have an ID")
	assert.True(t, session.Active, "Session should be active")
	assert.Equal(t, options.Shell, session.Shell, "Session should use specified shell")
	assert.Equal(t, options.CWD, session.CWD, "Session should use specified directory")

	// Test 4: Verify session state
	sessions := termService.GetSessions()
	assert.Len(t, sessions, 1, "Should have exactly one session")
	assert.Equal(t, session.ID, sessions[0].ID, "Retrieved session should match created session")

	// Test 5: Write data to terminal
	testData := []byte("echo 'Hello from aDex-UI'\n")
	err = termService.WriteToSession(ctx, session.ID, testData)
	assert.NoError(t, err, "Should be able to write data to terminal")

	// Test 6: Read output from terminal (with timeout)
	outputChan := make(chan []byte, 10)

	// Subscribe to terminal output
	termService.SubscribeToOutput(session.ID, func(data []byte) {
		select {
		case outputChan <- data:
		default:
			// Don't block if channel is full
		}
	})

	// Wait for output with timeout
	select {
	case output := <-outputChan:
		assert.NotEmpty(t, output, "Should receive output from terminal")
		t.Logf("Terminal output: %s", string(output))
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for terminal output")
	}

	// Test 7: Resize terminal
	newSize := &terminal.TerminalSize{
		Rows: 40,
		Cols: 120,
	}

	err = termService.ResizeSession(ctx, session.ID, newSize.Rows, newSize.Cols)
	assert.NoError(t, err, "Should be able to resize terminal")

	// Test 8: Close session
	err = termService.CloseSession(ctx, session.ID)
	assert.NoError(t, err, "Should be able to close session")

	// Verify session is no longer active
	sessions = termService.GetSessions()
	assert.Empty(t, sessions, "Should have no active sessions")

	// Cleanup
	termService.Shutdown(ctx)
}

// TestTerminalMultipleSessions tests multiple concurrent terminal sessions
func TestTerminalMultipleSessions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY tests skipped on Windows")
	}

	ctx := context.Background()
	termService := terminal.NewService()

	err := termService.Initialize(ctx)
	require.NoError(t, err)
	defer termService.Shutdown(ctx)

	// Create multiple sessions
	const numSessions = 3
	sessions := make([]*terminal.TerminalSession, numSessions)

	for i := 0; i < numSessions; i++ {
		options := &terminal.TerminalOptions{
			Shell: getDefaultShell(),
			CWD:   "/tmp",
			Rows:  24,
			Cols:  80,
		}

		session, err := termService.CreateSession(ctx, options)
		require.NoError(t, err)
		require.NotNil(t, session)
		sessions[i] = session

		// Verify each session has unique ID
		for j := 0; j < i; j++ {
			assert.NotEqual(t, sessions[j].ID, session.ID, "Each session should have unique ID")
		}
	}

	// Verify all sessions are active
	activeSessions := termService.GetSessions()
	assert.Len(t, activeSessions, numSessions, "Should have correct number of active sessions")

	// Write to each session
	for i, session := range sessions {
		testData := []byte(fmt.Sprintf("echo 'Session %d'\n", i+1))
		err = termService.WriteToSession(ctx, session.ID, testData)
		assert.NoError(t, err, "Should be able to write to each session")
	}

	// Close all sessions
	for _, session := range sessions {
		err = termService.CloseSession(ctx, session.ID)
		assert.NoError(t, err)
	}

	// Verify all sessions are closed
	activeSessions = termService.GetSessions()
	assert.Empty(t, activeSessions, "All sessions should be closed")
}

// TestTerminalSessionErrors tests error handling in terminal sessions
func TestTerminalSessionErrors(t *testing.T) {
	ctx := context.Background()
	termService := terminal.NewService()

	err := termService.Initialize(ctx)
	require.NoError(t, err)
	defer termService.Shutdown(ctx)

	// Test writing to non-existent session
	err = termService.WriteToSession(ctx, "non-existent", []byte("test"))
	assert.Error(t, err, "Should return error when writing to non-existent session")

	// Test resizing non-existent session
	err = termService.ResizeSession(ctx, "non-existent", 24, 80)
	assert.Error(t, err, "Should return error when resizing non-existent session")

	// Test closing non-existent session
	err = termService.CloseSession(ctx, "non-existent")
	assert.Error(t, err, "Should return error when closing non-existent session")

	// Test creating session with invalid shell
	options := &terminal.TerminalOptions{
		Shell: "/non-existent-shell",
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	_, err = termService.CreateSession(ctx, options)
	assert.Error(t, err, "Should return error when creating session with invalid shell")
}

// TestTerminalSessionPersistence tests basic session persistence functionality
func TestTerminalSessionPersistence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY tests skipped on Windows")
	}

	ctx := context.Background()
	termService := terminal.NewService()

	err := termService.Initialize(ctx)
	require.NoError(t, err)
	defer termService.Shutdown(ctx)

	// Create a session with some commands
	options := &terminal.TerminalOptions{
		Shell: getDefaultShell(),
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	session, err := termService.CreateSession(ctx, options)
	require.NoError(t, err)
	require.NotNil(t, session)

	// Write some commands
	commands := []string{
		"export TEST_VAR='aDex-UI Test'\n",
		"echo $TEST_VAR\n",
		"pwd\n",
	}

	for _, cmd := range commands {
		err = termService.WriteToSession(ctx, session.ID, []byte(cmd))
		assert.NoError(t, err)
		time.Sleep(100 * time.Millisecond) // Small delay for command processing
	}

	// Get session history (if implemented)
	history := termService.GetSessionHistory(session.ID)
	if history != nil {
		assert.NotEmpty(t, history, "Session should have command history")
	}

	// Close session
	err = termService.CloseSession(ctx, session.ID)
	assert.NoError(t, err)
}

// TestTerminalEventEmission tests that terminal events are properly emitted
func TestTerminalEventEmission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("PTY tests skipped on Windows")
	}

	ctx := context.Background()
	termService := terminal.NewService()

	err := termService.Initialize(ctx)
	require.NoError(t, err)
	defer termService.Shutdown(ctx)

	// Get event bus
	eventBus := termService.GetEventBus()
	require.NotNil(t, eventBus, "Service should have event bus")

	// Track received events
	var receivedEvents []interface{}

	// Subscribe to terminal events
	subscription, err := eventBus.Subscribe(ctx, []string{
		models.TerminalCreated,
		models.TerminalOutput,
		models.TerminalClosed,
	}, func(ctx context.Context, event models.Event) error {
		receivedEvents = append(receivedEvents, event)
		return nil
	})
	require.NoError(t, err)
	defer eventBus.Unsubscribe(subscription.ID)

	// Create a session
	options := &terminal.TerminalOptions{
		Shell: getDefaultShell(),
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	session, err := termService.CreateSession(ctx, options)
	require.NoError(t, err)

	// Write some data to trigger output events
	err = termService.WriteToSession(ctx, session.ID, []byte("echo 'test'\n"))
	assert.NoError(t, err)

	// Wait for events
	time.Sleep(500 * time.Millisecond)

	// Verify we received events
	assert.Greater(t, len(receivedEvents), 0, "Should have received terminal events")

	// Close session
	err = termService.CloseSession(ctx, session.ID)
	assert.NoError(t, err)

	// Wait for close event
	time.Sleep(100 * time.Millisecond)
}

// Helper function to get default shell for the platform
func getDefaultShell() string {
	if runtime.GOOS == "windows" {
		return "cmd.exe"
	}

	// Try common shells in order of preference
	shells := []string{"/bin/bash", "/bin/zsh", "/bin/sh", "/usr/bin/bash"}

	for _, shell := range shells {
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}

	// Fallback to /bin/sh
	return "/bin/sh"
}

// BenchmarkTerminalSessionCreation benchmarks terminal session creation performance
func BenchmarkTerminalSessionCreation(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("PTY tests skipped on Windows")
	}

	ctx := context.Background()
	termService := terminal.NewService()

	err := termService.Initialize(ctx)
	if err != nil {
		b.Fatalf("Failed to initialize terminal service: %v", err)
	}
	defer termService.Shutdown(ctx)

	options := &terminal.TerminalOptions{
		Shell: getDefaultShell(),
		CWD:   "/tmp",
		Rows:  24,
		Cols:  80,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		session, err := termService.CreateSession(ctx, options)
		if err != nil {
			b.Fatalf("Failed to create session: %v", err)
		}

		err = termService.CloseSession(ctx, session.ID)
		if err != nil {
			b.Fatalf("Failed to close session: %v", err)
		}
	}
}