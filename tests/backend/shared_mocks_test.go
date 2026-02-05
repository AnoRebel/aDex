package tests

import (
	"context"

	"aDex-UI/internal/events"
	"aDex-UI/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockEventBus is the shared mock implementation of events.IEventBus
// All test files should use this mock to avoid redeclaration errors
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*events.EventSubscription), args.Error(1)
}

func (m *MockEventBus) SubscribeOnce(ctx context.Context, eventType string, handler events.EventHandler) (*events.EventSubscription, error) {
	args := m.Called(ctx, eventType, handler)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

// MockFileSystemService is the shared mock implementation for filesystem operations
type MockFileSystemService struct {
	mock.Mock
}

func (m *MockFileSystemService) GetWorkingDirectory() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockFileSystemService) SetWorkingDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) ListDirectory(path string) ([]*models.FileSystemEntry, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileSystemEntry), args.Error(1)
}

func (m *MockFileSystemService) GetFileInfo(path string) (*models.FileSystemEntry, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileSystemEntry), args.Error(1)
}

func (m *MockFileSystemService) WatchDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) StopWatching(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) CreateDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) RenameFile(oldPath, newPath string) error {
	args := m.Called(oldPath, newPath)
	return args.Error(0)
}

func (m *MockFileSystemService) GetWatchedDirectories() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

// MockTerminalService is the shared mock implementation for terminal operations
type MockTerminalService struct {
	mock.Mock
}

func (m *MockTerminalService) GetSessionCWD(sessionID string) (string, error) {
	args := m.Called(sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockTerminalService) SetSessionCWD(sessionID string, cwd string) error {
	args := m.Called(sessionID, cwd)
	return args.Error(0)
}

func (m *MockTerminalService) GetActiveSessionID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTerminalService) GetTerminalSessions() ([]*models.TerminalSession, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TerminalSession), args.Error(1)
}

func (m *MockTerminalService) GetTerminalSession(sessionID string) (*models.TerminalSession, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TerminalSession), args.Error(1)
}

func (m *MockTerminalService) CreateTerminalSession(shellPath string, cols int, rows int) (*models.TerminalSession, error) {
	args := m.Called(shellPath, cols, rows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TerminalSession), args.Error(1)
}

// Utility functions for tests
// Note: Commented out since they are unused but may be needed for future tests.
/*
// isTaskAvailable checks if a task runner is available
func isTaskAvailable() bool {
	return false
}

// isWailsAvailable checks if wails is available
func isWailsAvailable() bool {
	return false
}
*/
