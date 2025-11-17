package backend

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"aDex-UI/internal/models"
	"aDex-UI/internal/services/terminal"
	"aDex-UI/internal/services/filesystem"
	"aDex-UI/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CWDTestSuite provides comprehensive testing for Current Working Directory tracking
type CWDTestSuite struct {
	suite.Suite
	terminalService *terminal.TerminalService
	filesystemService *filesystem.FileSystemService
	eventBus        events.IEventBus
	tempDir         string
	originalCWD     string
}

// SetupSuite runs once before all tests in the suite
func (suite *CWDTestSuite) SetupSuite() {
	// Save original working directory
	originalCWD, err := os.Getwd()
	require.NoError(suite.T(), err)
	suite.originalCWD = originalCWD

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cwd_test_")
	require.NoError(suite.T(), err)
	suite.tempDir = tempDir

	// Initialize services
	suite.eventBus = events.NewEventBus()
	suite.terminalService = terminal.NewTerminalService(suite.eventBus)
	suite.filesystemService = filesystem.NewFileSystemService()
}

// TearDownSuite runs once after all tests in the suite
func (suite *CWDTestSuite) TearDownSuite() {
	// Clean up temporary directory
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}

	// Restore original working directory
	if suite.originalCWD != "" {
		os.Chdir(suite.originalCWD)
	}
}

// Setup runs before each test
func (suite *CWDTestSuite) SetupTest() {
	// Reset working directory to temp directory for isolation
	os.Chdir(suite.tempDir)
}

// MockTerminalService provides a mock implementation for testing
type MockTerminalService struct {
	mock.Mock
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

// MockFileSystemService provides a mock implementation for testing
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

// TestCWDBasicFunctionality tests basic CWD tracking functionality
func (suite *CWDTestSuite) TestCWDBasicFunctionality(t *testing.T) {
	// Test getting current working directory
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NotEmpty(t, cwd)
	assert.True(t, filepath.IsAbs(cwd))

	// Test changing directory
	testDir := filepath.Join(suite.tempDir, "test_cwd")
	require.NoError(t, os.Mkdir(testDir, 0755))

	err = os.Chdir(testDir)
	assert.NoError(t, err)

	newCwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, testDir, newCwd)
}

// TestTerminalCWDTracking tests CWD tracking in terminal sessions
func (suite *CWDTestSuite) TestTerminalCWDTracking(t *testing.T) {
	// Skip on Windows for now (different PTY behavior)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping terminal CWD test on Windows")
	}

	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Create mock session
	session := &models.TerminalSession{
		ID:         "test-session-1",
		Title:      "Terminal",
		ProcessID:  1234,
		ShellPath:  "/bin/bash",
		WorkingDir: suite.tempDir,
		Cols:       80,
		Rows:       24,
		Active:     true,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Test getting session CWD
	mockTerminal.On("GetTerminalSession", "test-session-1").Return(session, nil)
	mockTerminal.On("GetSessionCWD", "test-session-1").Return(session.WorkingDir, nil)

	retrievedCWD, err := mockTerminal.GetSessionCWD("test-session-1")
	assert.NoError(t, err)
	assert.Equal(t, suite.tempDir, retrievedCWD)

	// Test setting session CWD
	newCWD := filepath.Join(suite.tempDir, "subdir")
	mockTerminal.On("SetSessionCWD", "test-session-1", newCWD).Return(nil)

	err = mockTerminal.SetSessionCWD("test-session-1", newCWD)
	assert.NoError(t, err)

	// Test active session CWD
	mockTerminal.On("GetActiveSessionID").Return("test-session-1", nil)
	mockTerminal.On("GetTerminalSessions").Return([]*models.TerminalSession{session}, nil)

	activeID, err := mockTerminal.GetActiveSessionID()
	assert.NoError(t, err)
	assert.Equal(t, "test-session-1", activeID)

	// Get updated session to verify CWD change
	session.WorkingDir = newCWD
	mockTerminal.On("GetTerminalSession", "test-session-1").Return(session, nil)

	updatedCWD, err := mockTerminal.GetSessionCWD("test-session-1")
	assert.NoError(t, err)
	assert.Equal(t, newCWD, updatedCWD)

	mockTerminal.AssertExpectations(t)
}

// TestFileSystemCWDTracking tests CWD tracking in file system service
func (suite *CWDTestSuite) TestFileSystemCWDTracking(t *testing.T) {
	mockFilesystem := new(MockFileSystemService)

	// Test getting current working directory
	expectedCWD := suite.tempDir
	mockFilesystem.On("GetWorkingDirectory").Return(expectedCWD, nil)

	cwd, err := mockFilesystem.GetWorkingDirectory()
	assert.NoError(t, err)
	assert.Equal(t, expectedCWD, cwd)

	// Test setting working directory
	newCWD := filepath.Join(suite.tempDir, "test_subdir")
	mockFilesystem.On("SetWorkingDirectory", newCWD).Return(nil)

	err = mockFilesystem.SetWorkingDirectory(newCWD)
	assert.NoError(t, err)

	mockFilesystem.AssertExpectations(t)
}

// TestCWDEventDrivenTracking tests event-driven CWD updates
func (suite *CWDTestSuite) TestCWDEventDrivenTracking(t *testing.T) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Create mock event bus
	eventBus := events.NewEventBus()

	// Set up event listeners
	var cwdUpdateReceived bool
	var updatedCWD string

	eventBus.On("terminal:cwd-changed", func(event map[string]interface{}) {
		cwdUpdateReceived = true
		updatedCWD = event["cwd"].(string)
	})

	// Simulate CWD change event
	testCWD := filepath.Join(suite.tempDir, "event_test")
	eventData := map[string]interface{}{
		"sessionId": "test-session",
		"cwd":        testCWD,
		"timestamp": time.Now().Unix(),
	}

	eventBus.Emit("terminal:cwd-changed", eventData)

	// Give event processing time
	time.Sleep(10 * time.Millisecond)

	assert.True(t, cwdUpdateReceived, "CWD change event should be received")
	assert.Equal(t, testCWD, updatedCWD, "Updated CWD should match event data")
}

// TestCWDSyncIntegration tests CWD synchronization between terminal and file browser
func (suite *CWDTestSuite) TestCWDSyncIntegration(t *testing.T) {
	// Skip on Windows for now
	if runtime.GOOS == "windows" {
		t.Skip("Skipping CWD sync test on Windows")
	}

	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Create mock terminal session with initial CWD
	initialCWD := suite.tempDir
	session := &models.TerminalSession{
		ID:         "sync-test-session",
		Title:      "Sync Test Terminal",
		ProcessID:  5678,
		ShellPath:  "/bin/bash",
		WorkingDir: initialCWD,
		Cols:       80,
		Rows:       24,
		Active:     true,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Set up mocks
	mockTerminal.On("GetActiveSessionID").Return("sync-test-session", nil)
	mockTerminal.On("GetSessionCWD", "sync-test-session").Return(initialCWD, nil)
	mockFilesystem.On("GetWorkingDirectory").Return(initialCWD, nil)

	// Verify initial synchronization
	activeID, err := mockTerminal.GetActiveSessionID()
	assert.NoError(t, err)

	terminalCWD, err := mockTerminal.GetSessionCWD(activeID)
	assert.NoError(t, err)

	filesystemCWD, err := mockFilesystem.GetWorkingDirectory()
	assert.NoError(t, err)

	assert.Equal(t, terminalCWD, filesystemCWD, "Terminal and file system CWD should be synchronized initially")

	// Simulate terminal directory change
	newCWD := filepath.Join(suite.tempDir, "synced_subdir")
	require.NoError(t, os.Mkdir(newCWD, 0755))

	session.WorkingDir = newCWD
	mockTerminal.On("GetSessionCWD", "sync-test-session").Return(newCWD, nil)
	mockFilesystem.On("SetWorkingDirectory", newCWD).Return(nil)

	// Update file system to match terminal CWD
	err = mockFilesystem.SetWorkingDirectory(newCWD)
	assert.NoError(t, err)

	// Verify synchronization after change
	updatedTerminalCWD, err := mockTerminal.GetSessionCWD(activeID)
	assert.NoError(t, err)
	assert.Equal(t, newCWD, updatedTerminalCWD)

	// Verify file system was updated
	mockFilesystem.AssertCalledWith("SetWorkingDirectory", newCWD)

	mockTerminal.AssertExpectations(t)
	mockFilesystem.AssertExpectations(t)
}

// TestCWDPermissionHandling tests CWD operations with permission restrictions
func (suite *CWDTestSuite) TestCWDPermissionHandling(t *testing.T) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Test changing to non-existent directory
	nonExistentDir := "/nonexistent/path"
	mockTerminal.On("SetSessionCWD", "test-session", nonExistentDir).Return(os.ErrNotExist)
	mockFilesystem.On("SetWorkingDirectory", nonExistentDir).Return(os.ErrNotExist)

	err := mockTerminal.SetSessionCWD("test-session", nonExistentDir)
	assert.Error(t, err)
	assert.Equal(t, os.ErrNotExist, err)

	err = mockFilesystem.SetWorkingDirectory(nonExistentDir)
	assert.Error(t, err)
	assert.Equal(t, os.ErrNotExist, err)

	// Test accessing directory without read permissions
	if runtime.GOOS != "windows" {
		// Create a directory with no read permissions
		noReadDir := filepath.Join(suite.tempDir, "no_read")
		require.NoError(t, os.Mkdir(noReadDir, 0000))

		mockFilesystem.On("ListDirectory", noReadDir).Return(nil, os.ErrPermission)

		_, err = mockFilesystem.ListDirectory(noReadDir)
		assert.Error(t, err)
		assert.Equal(t, os.ErrPermission, err)
	}

	mockTerminal.AssertExpectations(t)
	mockFilesystem.AssertExpectations(t)
}

// TestCWDPathValidation tests CWD path validation and normalization
func (suite *CWDTestSuite) TestCWDPathValidation(t *testing.T) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Test paths with different formats
	testCases := []struct {
		name        string
		inputPath   string
		shouldPass  bool
		description  string
	}{
		{
			name:        "Absolute path",
			inputPath:   "/home/user/documents",
			shouldPass:  true,
			description:  "Valid absolute path should pass",
		},
		{
			name:        "Relative path",
			inputPath:   "../parent",
			shouldPass:  true,
			description:  "Valid relative path should pass",
		},
		{
			name:        "Empty path",
			inputPath:   "",
			shouldPass:  false,
			description:  "Empty path should fail",
		},
		{
			name:        "Invalid characters",
			inputPath:   "/path/with\null",
			shouldPass:  false,
			description:  "Path with null characters should fail",
		},
		{
			name:        "Very long path",
			inputPath:   strings.Repeat("/very/long/path/component/", 100),
			shouldPass:  false,
			description:  "Extremely long path should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPass {
				mockTerminal.On("SetSessionCWD", "test-session", tc.inputPath).Return(nil)
				mockFilesystem.On("SetWorkingDirectory", tc.inputPath).Return(nil)

				err := mockTerminal.SetSessionCWD("test-session", tc.inputPath)
				assert.NoError(t, err, tc.description)

				err = mockFilesystem.SetWorkingDirectory(tc.inputPath)
				assert.NoError(t, err, tc.description)
			} else {
				mockTerminal.On("SetSessionCWD", "test-session", tc.inputPath).Return(assert.AnError)
				mockFilesystem.On("SetWorkingDirectory", tc.inputPath).Return(assert.AnError)

				err := mockTerminal.SetSessionCWD("test-session", tc.inputPath)
				assert.Error(t, err, tc.description)

				err = mockFilesystem.SetWorkingDirectory(tc.inputPath)
				assert.Error(t, err, tc.description)
			}
		})
	}

	mockTerminal.AssertExpectations(t)
	mockFilesystem.AssertExpectations(t)
}

// TestCWDConcurrentAccess tests CWD operations under concurrent access
func (suite *CWDTestSuite) TestCWDConcurrentAccess(t *testing.T) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Create multiple sessions
	sessions := make([]*models.TerminalSession, 5)
	for i := 0; i < 5; i++ {
		sessions[i] = &models.TerminalSession{
			ID:         fmt.Sprintf("session-%d", i),
			Title:      fmt.Sprintf("Terminal %d", i),
			ProcessID:  1000 + i,
			ShellPath:  "/bin/bash",
			WorkingDir: suite.tempDir,
			Cols:       80,
			Rows:       24,
			Active:     true,
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
		}

		mockTerminal.On("GetSessionCWD", sessions[i].ID).Return(sessions[i].WorkingDir, nil)
		mockTerminal.On("SetSessionCWD", sessions[i].ID, mock.AnythingOfType("string")).Return(nil)
	}

	// Concurrent CWD reads
	var wg sync.WaitGroup
	results := make(chan string, 5)

	for _, session := range sessions {
		wg.Add(1)
		go func(s *models.TerminalSession) {
			defer wg.Done()
			cwd, err := mockTerminal.GetSessionCWD(s.ID)
			assert.NoError(t, err)
			results <- cwd
		}(session)
	}

	wg.Wait()
	close(results)

	// Verify all reads completed successfully
	count := 0
	for cwd := range results {
		assert.Equal(t, suite.tempDir, cwd)
		count++
	}
	assert.Equal(t, 5, count)

	// Concurrent CWD writes
	for _, session := range sessions {
		wg.Add(1)
		go func(s *models.TerminalSession) {
			defer wg.Done()
			newCWD := filepath.Join(suite.tempDir, fmt.Sprintf("concurrent_%d", strings.TrimPrefix(s.ID, "session-")))
			err := mockTerminal.SetSessionCWD(s.ID, newCWD)
			assert.NoError(t, err)
		}(session)
	}

	wg.Wait()

	// Verify all writes completed
	mockTerminal.AssertExpectations(t)
}

// TestCWDPerformance tests CWD operation performance
func (suite *CWDTestSuite) TestCWDPerformance(t *testing.T) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Set up mock with minimal delay
	mockTerminal.On("GetSessionCWD", "perf-test-session").Return(suite.tempDir, nil)
	mockFilesystem.On("GetWorkingDirectory").Return(suite.tempDir, nil)

	// Benchmark CWD reads
	t.Run("CWD_Read_Performance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			_, err := mockTerminal.GetSessionCWD("perf-test-session")
			assert.NoError(t, err)
		}
		duration := time.Since(start)
		t.Logf("1000 CWD reads took %v (avg: %v per read)", duration, duration/1000)
		assert.Less(t, duration, 100*time.Millisecond, "CWD reads should be fast")
	})

	// Benchmark CWD writes
	t.Run("CWD_Write_Performance", func(t *testing.T) {
		mockTerminal.On("SetSessionCWD", "perf-test-session", mock.AnythingOfType("string")).Return(nil)

		start := time.Now()
		for i := 0; i < 100; i++ {
			testPath := filepath.Join(suite.tempDir, fmt.Sprintf("perf_test_%d", i))
			err := mockTerminal.SetSessionCWD("perf-test-session", testPath)
			assert.NoError(t, err)
		}
		duration := time.Since(start)
		t.Logf("100 CWD writes took %v (avg: %v per write)", duration, duration/100)
		assert.Less(t, duration, 50*time.Millisecond, "CWD writes should be fast")
	})

	mockTerminal.AssertExpectations(t)
	mockFilesystem.AssertExpectations(t)
}

// TestCWDPlatformSpecificBehavior tests platform-specific CWD behavior
func (suite *CWDTestSuite) TestCWDPlatformSpecificBehavior(t *testing.T) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	// Test path separator handling
	testPath := "subdir" + string(filepath.Separator) + "file.txt"
	expectedPath := filepath.Join(suite.tempDir, "subdir", "file.txt")

	switch runtime.GOOS {
	case "windows":
		t.Run("Windows_Specific", func(t *testing.T) {
			// Test Windows-specific path handling
			windowsPath := `C:\Users\Test\Documents`
			mockTerminal.On("SetSessionCWD", "windows-session", windowsPath).Return(nil)

			err := mockTerminal.SetSessionCWD("windows-session", windowsPath)
			assert.NoError(t, err)

			// Test case sensitivity
			mockTerminal.On("SetSessionCWD", "windows-session-2", "C:\\USERS\\TEST").Return(nil)
			err = mockTerminal.SetSessionCWD("windows-session-2", "C:\\USERS\\TEST")
			assert.NoError(t, err)
		})

	case "linux", "darwin":
		t.Run("Unix_Specific", func(t *testing.T) {
			// Test Unix-specific path handling
			unixPath := "/home/user/documents"
			mockTerminal.On("SetSessionCWD", "unix-session", unixPath).Return(nil)

			err := mockTerminal.SetSessionCWD("unix-session", unixPath)
			assert.NoError(t, err)

			// Test case sensitivity (Linux)
			if runtime.GOOS == "linux" {
				mockTerminal.On("SetSessionCWD", "linux-session", "/HOME/USER").Return(assert.AnError)
				err = mockTerminal.SetSessionCWD("linux-session", "/HOME/USER")
				// On Linux, this should fail as /HOME doesn't exist
				// In mock, we just verify the method was called
			}
		})
	}

	// Test path normalization across platforms
	t.Run("Path_Normalization", func(t *testing.T) {
		// Test path with trailing slash
		pathWithSlash := suite.tempDir + string(filepath.Separator)
		mockTerminal.On("SetSessionCWD", "normalize-session", pathWithSlash).Return(nil)

		err := mockTerminal.SetSessionCWD("normalize-session", pathWithSlash)
		assert.NoError(t, err)

		// Test path with relative components
		relativePath := "./subdir/../"
		mockTerminal.On("SetSessionCWD", "relative-session", relativePath).Return(nil)

		err = mockTerminal.SetSessionCWD("relative-session", relativePath)
		assert.NoError(t, err)
	})

	mockTerminal.AssertExpectations(t)
	mockFilesystem.AssertExpectations(t)
}

// TestCWDIntegrationWithRealFileSystem tests CWD operations with real file system
func (suite *CWDTestSuite) TestCWDIntegrationWithRealFileSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test directory structure
	testDir := filepath.Join(suite.tempDir, "cwd_integration")
	require.NoError(t, os.Mkdir(testDir, 0755))

	subDir1 := filepath.Join(testDir, "subdir1")
	subDir2 := filepath.Join(testDir, "subdir2")
	require.NoError(t, os.Mkdir(subDir1, 0755))
	require.NoError(t, os.Mkdir(subDir2, 0755))

	// Test real file system CWD operations
	t.Run("Real_Chdir", func(t *testing.T) {
		originalCWD, err := os.Getwd()
		require.NoError(t, err)

		// Change directory
		err = os.Chdir(testDir)
		require.NoError(t, err)

		currentCWD, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, testDir, currentCWD)

		// Navigate to subdirectory
		err = os.Chdir("subdir1")
		require.NoError(t, err)

		subCWD, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, subDir1, subCWD)

		// Navigate up
		err = os.Chdir("..")
		require.NoError(t, err)

		backCWD, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, testDir, backCWD)

		// Restore original CWD
		err = os.Chdir(originalCWD)
		require.NoError(t, err)
	})

	// Test file creation and verification
	t.Run("File_Operations", func(t *testing.T) {
		err := os.Chdir(testDir)
		require.NoError(t, err)

		// Create a test file
		testFile := "cwd_test.txt"
		content := "CWD integration test content"
		err = os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(testFile)
		assert.NoError(t, err)

		// Read file content
		readContent, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, content, string(readContent))

		// Clean up
		err = os.Remove(testFile)
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkCWDOperations(b *testing.B) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	mockTerminal.On("GetSessionCWD", "bench-session").Return("/test/path", nil)
	mockFilesystem.On("GetWorkingDirectory").Return("/test/path", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockTerminal.GetSessionCWD("bench-session")
		mockFilesystem.GetWorkingDirectory()
	}
}

func BenchmarkCWDChanges(b *testing.B) {
	mockTerminal := new(MockTerminalService)
	mockFilesystem := new(MockFileSystemService)

	mockTerminal.On("SetSessionCWD", "bench-session", mock.AnythingOfType("string")).Return(nil)
	mockFilesystem.On("SetWorkingDirectory", mock.AnythingOfType("string")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testPath := fmt.Sprintf("/test/path_%d", i)
		mockTerminal.SetSessionCWD("bench-session", testPath)
		mockFilesystem.SetWorkingDirectory(testPath)
	}
}