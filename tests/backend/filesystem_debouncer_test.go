package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"
	"aDex-UI/internal/services/filesystem"

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

func (m *MockEventBus) Broadcast(ctx context.Context, data interface{}, source string) error {
	args := m.Called(ctx, data, source)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventTypes []string, handler events.EventHandler) (*events.EventSubscription, error) {
	args := m.Called(ctx, eventTypes, handler)
	return args.Get(0).(*events.EventSubscription), args.Error(1)
}

func (m *MockEventBus) SubscribeOnce(ctx context.Context, eventType string, handler events.EventHandler) (*events.EventSubscription, error) {
	args := m.Called(ctx, eventType, handler)
	return args.Get(0).(*events.EventSubscription), args.Error(1)
}

func (m *MockEventBus) Unsubscribe(subscriptionID string) error {
	args := m.Called(subscriptionID)
	return args.Error(0)
}

func (m *MockEventBus) GetSubscribers(eventType string) int {
	args := m.Called(eventType)
	return args.Int(0)
}

// TestDebouncerCreation tests creating a new debouncer
func TestDebouncerCreation(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:           300 * time.Millisecond,
		MaxPending:      100,
		BatchSize:       50,
		GroupByPath:     true,
		GroupByType:     false,
	}

	debouncer := filesystem.NewDebouncer(config)
	assert.NotNil(t, debouncer, "Debouncer should be created")
	// Note: We can't access private fields directly, so we test via public methods
	stats := debouncer.GetStats()
	assert.Equal(t, config.Delay.String(), stats["delay"])

	// Cleanup
	debouncer.Stop()
}

// TestDebouncerDefaultConfig tests using default configuration
func TestDebouncerDefaultConfig(t *testing.T) {
	debouncer := filesystem.NewDebouncer(nil)
	assert.NotNil(t, debouncer, "Debouncer should be created with default config")

	// Test default delay is applied
	stats := debouncer.GetStats()
	assert.Equal(t, "300ms", stats["delay"], "Default delay should be 300ms")

	// Cleanup
	debouncer.Stop()
}

// TestDebouncerEventGrouping tests event grouping functionality
func TestDebouncerEventGrouping(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:           100 * time.Millisecond,
		MaxPending:      10,
		BatchSize:       5,
		GroupByPath:     true,
		GroupByType:     false,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	// Create test events
	event1 := &models.FileSystemEvent{
		Path:      "/test/file1.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}

	event2 := &models.FileSystemEvent{
		Path:      "/test/file1.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}

	event3 := &models.FileSystemEvent{
		Path:      "/test/file2.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}

	// Add events
	debouncer.AddEvent(event1)
	debouncer.AddEvent(event2)
	debouncer.AddEvent(event3)

	// Check pending events
	stats := debouncer.GetStats()
	assert.Equal(t, 2, stats["pending_count"], "Should have 2 pending events (grouped by path)")

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	// Check events were processed
	stats = debouncer.GetStats()
	assert.Equal(t, 0, stats["pending_count"], "All events should be processed")
}

// TestDebouncerBatchProcessing tests batch processing of events
func TestDebouncerBatchProcessing(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:           50 * time.Millisecond,
		MaxPending:      20,
		BatchSize:       3,
		GroupByPath:     false,
		GroupByType:     false,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	// Create multiple test events
	events := make([]*models.FileSystemEvent, 10)
	for i := 0; i < 10; i++ {
		events[i] = &models.FileSystemEvent{
			Path:      filepath.Join("/test", "file", string(rune('a'+i))),
			Operation: "write",
			Timestamp: time.Now().Unix(),
		}
		debouncer.AddEvent(events[i])
	}

	// Wait for batch processing
	time.Sleep(200 * time.Millisecond)

	// Check all events were processed
	stats := debouncer.GetStats()
	assert.Equal(t, 0, stats["pending_count"], "All events should be processed")
}

// TestDebouncerStopAndCleanup tests stopping and cleanup
func TestDebouncerStopAndCleanup(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:       100 * time.Millisecond,
		MaxPending:  10,
		BatchSize:   5,
		GroupByPath: true,
	}

	debouncer := filesystem.NewDebouncer(config)
	assert.True(t, debouncer.GetStats()["running"].(bool), "Debouncer should be running initially")

	// Add some events
	event := &models.FileSystemEvent{
		Path:      "/test/file.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}
	debouncer.AddEvent(event)

	// Stop the debouncer
	debouncer.Stop()
	assert.False(t, debouncer.GetStats()["running"].(bool), "Debouncer should not be running after stop")

	// Try to add events after stopping
	debouncer.AddEvent(event) // Should not panic
	stats := debouncer.GetStats()
	assert.Equal(t, 0, stats["pending_count"], "No events should be pending after stop")
}

// TestDebouncerFlush tests manual flushing of events
func TestDebouncerFlush(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:       500 * time.Millisecond, // Long delay
		MaxPending:  10,
		BatchSize:   5,
		GroupByPath: true,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	// Add some events
	event := &models.FileSystemEvent{
		Path:      "/test/file.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}
	debouncer.AddEvent(event)

	// Check event is pending
	stats := debouncer.GetStats()
	assert.Equal(t, 1, stats["pending_count"], "Event should be pending")

	// Manually flush
	debouncer.Flush()

	// Wait a bit for flush to complete
	time.Sleep(50 * time.Millisecond)

	// Check event was processed
	stats = debouncer.GetStats()
	assert.Equal(t, 0, stats["pending_count"], "Event should be processed after flush")
}

// TestDebouncerClearPending tests clearing pending events
func TestDebouncerClearPending(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:       500 * time.Millisecond,
		MaxPending:  10,
		BatchSize:   5,
		GroupByPath: true,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	// Add some events
	event := &models.FileSystemEvent{
		Path:      "/test/file.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}
	debouncer.AddEvent(event)
	debouncer.AddEvent(event)

	// Check events are pending
	stats := debouncer.GetStats()
	assert.Equal(t, 1, stats["pending_count"], "Events should be pending (grouped)")

	// Clear pending events
	debouncer.ClearPending()

	// Check events were cleared
	stats = debouncer.GetStats()
	assert.Equal(t, 0, stats["pending_count"], "No events should be pending after clear")
}

// TestDebouncerEventGroups tests getting event group information
func TestDebouncerEventGroups(t *testing.T) {
	config := &filesystem.DebounceConfig{
		Delay:       500 * time.Millisecond,
		MaxPending:  10,
		BatchSize:   5,
		GroupByPath: true,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	// Add some events
	event := &models.FileSystemEvent{
		Path:      "/test/file.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}
	debouncer.AddEvent(event)
	debouncer.AddEvent(event) // Same path, should be grouped

	// Get event groups
	groups := debouncer.GetEventGroups()
	assert.Len(t, groups, 1, "Should have 1 event group")
	assert.Equal(t, "/test/file.txt", groups[0].Path)
	assert.Equal(t, "write", groups[0].Operation)
	assert.Equal(t, 2, groups[0].Count)
}

// TestDirectoryWatcherDebouncing tests the directory watcher with debouncing enabled
func TestDirectoryWatcherDebouncing(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dex-watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock event bus
	mockEventBus := &MockEventBus{}
	mockEventBus.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create watcher with debouncing enabled
	watcher, err := filesystem.NewDirectoryWatcher(mockEventBus)
	require.NoError(t, err)

	config := &models.DirectoryWatcherConfig{
		Enabled:       true,
		WatchHidden:   false,
		DebounceMs:    100, // 100ms debounce delay
		MaxEvents:     100,
		BufferSize:    512,
		IgnorePatterns: []string{},
	}

	err = watcher.SetConfig(config)
	require.NoError(t, err)

	// Start the watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Watch the temporary directory
	sessionID, err := watcher.WatchDirectory(tempDir, false)
	require.NoError(t, err)
	require.NotEmpty(t, sessionID)

	// Create a test file quickly (multiple rapid events)
	testFile := filepath.Join(tempDir, "test.txt")

	// Create file
	file, err := os.Create(testFile)
	require.NoError(t, err)

	// Write to file multiple times rapidly
	for i := 0; i < 5; i++ {
		file.WriteString("test content\n")
		file.Sync()
		time.Sleep(10 * time.Millisecond) // Rapid writes
	}
	file.Close()

	// Wait for debouncing
	time.Sleep(200 * time.Millisecond)

	// Check that events were debounced (should have fewer events than writes)
	stats := watcher.GetStats()
	assert.True(t, stats["debouncing"].(bool), "Debouncing should be enabled")

	debouncerStats, exists := stats["debouncer_stats"].(map[string]interface{})
	assert.True(t, exists, "Debouncer stats should be available")
	assert.NotNil(t, debouncerStats)

	// Clean up
	err = watcher.StopWatching(tempDir)
	require.NoError(t, err)
}

// TestDirectoryWatcherDebouncingDisabled tests directory watcher with debouncing disabled
func TestDirectoryWatcherDebouncingDisabled(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dex-watcher-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock event bus
	mockEventBus := &MockEventBus{}
	mockEventBus.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create watcher
	watcher, err := filesystem.NewDirectoryWatcher(mockEventBus)
	require.NoError(t, err)

	config := &models.DirectoryWatcherConfig{
		Enabled:       true,
		WatchHidden:   false,
		DebounceMs:    0, // Debouncing disabled
		MaxEvents:     100,
		BufferSize:    512,
		IgnorePatterns: []string{},
	}

	err = watcher.SetConfig(config)
	require.NoError(t, err)

	// Start the watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Watch the temporary directory
	sessionID, err := watcher.WatchDirectory(tempDir, false)
	require.NoError(t, err)
	require.NotEmpty(t, sessionID)

	// Check stats
	stats := watcher.GetStats()
	assert.False(t, stats["debouncing"].(bool), "Debouncing should be disabled")

	// Clean up
	err = watcher.StopWatching(tempDir)
	require.NoError(t, err)
}

// BenchmarkDebouncerAddEvent benchmarks adding events to the debouncer
func BenchmarkDebouncerAddEvent(b *testing.B) {
	config := &filesystem.DebounceConfig{
		Delay:       300 * time.Millisecond,
		MaxPending:  1000,
		BatchSize:   50,
		GroupByPath: true,
		GroupByType: false,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	event := &models.FileSystemEvent{
		Path:      "/test/file.txt",
		Operation: "write",
		Timestamp: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		debouncer.AddEvent(event)
	}
}

// BenchmarkDebouncerGetStats benchmarks getting debouncer statistics
func BenchmarkDebouncerGetStats(b *testing.B) {
	config := &filesystem.DebounceConfig{
		Delay:       300 * time.Millisecond,
		MaxPending:  100,
		BatchSize:   50,
		GroupByPath: true,
		GroupByType: false,
	}

	debouncer := filesystem.NewDebouncer(config)
	defer debouncer.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = debouncer.GetStats()
	}
}