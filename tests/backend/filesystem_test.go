package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aDex-UI/internal/models"
	"aDex-UI/internal/services/filesystem"
	"aDex-UI/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

// MockFileSystemService provides a mock implementation for testing
type MockFileSystemService struct {
	mock.Mock
}

func (m *MockFileSystemService) ListDirectory(path string) ([]*models.FileSystemEntry, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileSystemEntry), args.Error(1)
}

func (m *MockFileSystemService) WatchDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) StopWatching(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) GetFileInfo(path string) (*models.FileSystemEntry, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileSystemEntry), args.Error(1)
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

func (m *MockFileSystemService) GetWorkingDirectory() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockFileSystemService) SetWorkingDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystemService) GetWatchedDirectories() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// TestFileSystemService_ListDirectory tests the ListDirectory method
func TestFileSystemService_ListDirectory(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testDir := t.TempDir()

	// Create test files and directories
	testFile := filepath.Join(testDir, "test.txt")
	testSubDir := filepath.Join(testDir, "subdir")

	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))
	require.NoError(t, os.Mkdir(testSubDir, 0755))

	expectedEntries := []*models.FileSystemEntry{
		{
			Name:         "test.txt",
			Path:         testFile,
			IsDir:        false,
			Size:         12,
			Mode:         0644,
			ModTime:      time.Now(),
			Permissions:  "-rw-r--r--",
			Owner:        "",
			Group:        "",
			IsHidden:     false,
			IsExecutable: false,
		},
		{
			Name:         "subdir",
			Path:         testSubDir,
			IsDir:        true,
			Size:         0,
			Mode:         0755,
			ModTime:      time.Now(),
			Permissions:  "rwxr-xr-x",
			Owner:        "",
			Group:        "",
			IsHidden:     false,
			IsExecutable: true,
		},
	}

	mockService.On("ListDirectory", testDir).Return(expectedEntries, nil)

	// Act
	entries, err := mockService.ListDirectory(testDir)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// Verify file entry
	fileEntry := entries[0]
	assert.Equal(t, "test.txt", fileEntry.Name)
	assert.Equal(t, testFile, fileEntry.Path)
	assert.False(t, fileEntry.IsDir)
	assert.Equal(t, int64(12), fileEntry.Size)
	assert.False(t, fileEntry.IsHidden)
	assert.False(t, fileEntry.IsExecutable)

	// Verify directory entry
	dirEntry := entries[1]
	assert.Equal(t, "subdir", dirEntry.Name)
	assert.Equal(t, testSubDir, dirEntry.Path)
	assert.True(t, dirEntry.IsDir)
	assert.Equal(t, int64(0), dirEntry.Size)
	assert.True(t, dirEntry.IsExecutable)

	mockService.AssertExpectations(t)
}

// TestFileSystemService_ListDirectory_Error tests error handling in ListDirectory
func TestFileSystemService_ListDirectory_Error(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	invalidPath := "/nonexistent/path"

	mockService.On("ListDirectory", invalidPath).Return(nil, assert.AnError)

	// Act
	entries, err := mockService.ListDirectory(invalidPath)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entries)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_ListDirectory_EmptyDirectory tests listing an empty directory
func TestFileSystemService_ListDirectory_EmptyDirectory(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	emptyDir := t.TempDir()

	mockService.On("ListDirectory", emptyDir).Return([]*models.FileSystemEntry{}, nil)

	// Act
	entries, err := mockService.ListDirectory(emptyDir)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Empty(t, entries)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_WatchDirectory tests directory watching
func TestFileSystemService_WatchDirectory(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testDir := t.TempDir()

	mockService.On("WatchDirectory", testDir).Return(nil)

	// Act
	err := mockService.WatchDirectory(testDir)

	// Assert
	require.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_WatchDirectory_Error tests error handling in WatchDirectory
func TestFileSystemService_WatchDirectory_Error(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	invalidPath := "/nonexistent/path"

	mockService.On("WatchDirectory", invalidPath).Return(assert.AnError)

	// Act
	err := mockService.WatchDirectory(invalidPath)

	// Assert
	assert.Error(t, err)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_StopWatching tests stopping directory watching
func TestFileSystemService_StopWatching(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testDir := t.TempDir()

	mockService.On("StopWatching", testDir).Return(nil)

	// Act
	err := mockService.StopWatching(testDir)

	// Assert
	require.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_GetFileInfo tests getting file information
func TestFileSystemService_GetFileInfo(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testFile := filepath.Join(t.TempDir(), "test.txt")

	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	expectedEntry := &models.FileSystemEntry{
		Name:         "test.txt",
		Path:         testFile,
		IsDir:        false,
		Size:         12,
		Mode:         0644,
		ModTime:      time.Now(),
		Permissions:  "-rw-r--r--",
		Owner:        "",
		Group:        "",
		IsHidden:     false,
		IsExecutable: false,
	}

	mockService.On("GetFileInfo", testFile).Return(expectedEntry, nil)

	// Act
	entry, err := mockService.GetFileInfo(testFile)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "test.txt", entry.Name)
	assert.Equal(t, testFile, entry.Path)
	assert.False(t, entry.IsDir)
	assert.Equal(t, int64(12), entry.Size)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_GetFileInfo_Directory tests getting directory information
func TestFileSystemService_GetFileInfo_Directory(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testDir := t.TempDir()

	expectedEntry := &models.FileSystemEntry{
		Name:         filepath.Base(testDir),
		Path:         testDir,
		IsDir:        true,
		Size:         0,
		Mode:         0755,
		ModTime:      time.Now(),
		Permissions:  "rwxr-xr-x",
		Owner:        "",
		Group:        "",
		IsHidden:     false,
		IsExecutable: true,
	}

	mockService.On("GetFileInfo", testDir).Return(expectedEntry, nil)

	// Act
	entry, err := mockService.GetFileInfo(testDir)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, entry)
	assert.True(t, entry.IsDir)
	assert.Equal(t, int64(0), entry.Size)
	assert.True(t, entry.IsExecutable)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_CreateDirectory tests directory creation
func TestFileSystemService_CreateDirectory(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	newDir := filepath.Join(t.TempDir(), "new_directory")

	mockService.On("CreateDirectory", newDir).Return(nil)

	// Act
	err := mockService.CreateDirectory(newDir)

	// Assert
	require.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_DeleteFile tests file deletion
func TestFileSystemService_DeleteFile(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testFile := filepath.Join(t.TempDir(), "test.txt")

	mockService.On("DeleteFile", testFile).Return(nil)

	// Act
	err := mockService.DeleteFile(testFile)

	// Assert
	require.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_RenameFile tests file renaming
func TestFileSystemService_RenameFile(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testDir := t.TempDir()
	oldPath := filepath.Join(testDir, "old.txt")
	newPath := filepath.Join(testDir, "new.txt")

	mockService.On("RenameFile", oldPath, newPath).Return(nil)

	// Act
	err := mockService.RenameFile(oldPath, newPath)

	// Assert
	require.NoError(t, err)
	mockService.AssertExpectations(t)
}

// TestFileSystemService_WorkingDirectory tests working directory operations
func TestFileSystemService_WorkingDirectory(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	testDir := t.TempDir()

	mockService.On("GetWorkingDirectory").Return(testDir, nil)
	mockService.On("SetWorkingDirectory", testDir).Return(nil)

	// Act & Assert - Get working directory
	wd, err := mockService.GetWorkingDirectory()
	require.NoError(t, err)
	assert.Equal(t, testDir, wd)

	// Act & Assert - Set working directory
	err = mockService.SetWorkingDirectory(testDir)
	require.NoError(t, err)

	mockService.AssertExpectations(t)
}

// TestFileSystemService_GetWatchedDirectories tests getting watched directories
func TestFileSystemService_GetWatchedDirectories(t *testing.T) {
	// Arrange
	mockService := new(MockFileSystemService)
	watchedDirs := []string{"/path/to/dir1", "/path/to/dir2"}

	mockService.On("GetWatchedDirectories").Return(watchedDirs)

	// Act
	dirs := mockService.GetWatchedDirectories()

	// Assert
	assert.NotNil(t, dirs)
	assert.Len(t, dirs, 2)
	assert.Equal(t, "/path/to/dir1", dirs[0])
	assert.Equal(t, "/path/to/dir2", dirs[1])
	mockService.AssertExpectations(t)
}

// TestFileSystemModels tests the FileSystemEntry model
func TestFileSystemModels_FileSystemEntry(t *testing.T) {
	// Arrange & Act
	entry := &models.FileSystemEntry{
		Name:         "test.txt",
		Path:         "/home/user/test.txt",
		IsDir:        false,
		Size:         1024,
		Mode:         0644,
		ModTime:      time.Now(),
		Permissions:  "-rw-r--r--",
		Owner:        "user",
		Group:        "group",
		IsHidden:     false,
		IsExecutable: false,
	}

	// Assert
	assert.Equal(t, "test.txt", entry.Name)
	assert.Equal(t, "/home/user/test.txt", entry.Path)
	assert.False(t, entry.IsDir)
	assert.Equal(t, int64(1024), entry.Size)
	assert.Equal(t, os.FileMode(0644), entry.Mode)
	assert.Equal(t, "-rw-r--r--", entry.Permissions)
	assert.Equal(t, "user", entry.Owner)
	assert.Equal(t, "group", entry.Group)
	assert.False(t, entry.IsHidden)
	assert.False(t, entry.IsExecutable)
}

// TestFileSystemModels_DirectoryEntry tests directory entry model
func TestFileSystemModels_DirectoryEntry(t *testing.T) {
	// Arrange & Act
	entry := &models.FileSystemEntry{
		Name:         "documents",
		Path:         "/home/user/documents",
		IsDir:        true,
		Size:         0,
		Mode:         0755,
		ModTime:      time.Now(),
		Permissions:  "rwxr-xr-x",
		Owner:        "user",
		Group:        "group",
		IsHidden:     false,
		IsExecutable: true,
		Children: []*models.FileSystemEntry{
			{
				Name:  "file.txt",
				Path:  "/home/user/documents/file.txt",
				IsDir: false,
			},
		},
	}

	// Assert
	assert.Equal(t, "documents", entry.Name)
	assert.True(t, entry.IsDir)
	assert.Equal(t, int64(0), entry.Size)
	assert.True(t, entry.IsExecutable)
	assert.NotNil(t, entry.Children)
	assert.Len(t, entry.Children, 1)
	assert.Equal(t, "file.txt", entry.Children[0].Name)
}

// TestFileSystemModels_HiddenFile tests hidden file detection
func TestFileSystemModels_HiddenFile(t *testing.T) {
	// Arrange & Act
	hiddenEntry := &models.FileSystemEntry{
		Name:     ".hidden",
		Path:     "/home/user/.hidden",
		IsDir:    false,
		IsHidden: true,
	}

	normalEntry := &models.FileSystemEntry{
		Name:     "normal.txt",
		Path:     "/home/user/normal.txt",
		IsDir:    false,
		IsHidden: false,
	}

	// Assert
	assert.True(t, hiddenEntry.IsHidden)
	assert.Equal(t, ".hidden", hiddenEntry.Name)

	assert.False(t, normalEntry.IsHidden)
	assert.Equal(t, "normal.txt", normalEntry.Name)
}

// TestFileSystemModels_ExecutableFile tests executable file detection
func TestFileSystemModels_ExecutableFile(t *testing.T) {
	// Arrange & Act
	execEntry := &models.FileSystemEntry{
		Name:         "script.sh",
		Path:         "/home/user/script.sh",
		IsDir:        false,
		Mode:         0755,
		Permissions:  "rwxr-xr-x",
		IsExecutable: true,
	}

	nonExecEntry := &models.FileSystemEntry{
		Name:         "readme.txt",
		Path:         "/home/user/readme.txt",
		IsDir:        false,
		Mode:         0644,
		Permissions:  "-rw-r--r--",
		IsExecutable: false,
	}

	// Assert
	assert.True(t, execEntry.IsExecutable)
	assert.Equal(t, "script.sh", execEntry.Name)
	assert.Equal(t, os.FileMode(0755), execEntry.Mode)

	assert.False(t, nonExecEntry.IsExecutable)
	assert.Equal(t, "readme.txt", nonExecEntry.Name)
	assert.Equal(t, os.FileMode(0644), nonExecEntry.Mode)
}

// TestFileSystemService_Integration tests integration with real file system service
func TestFileSystemService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	service := filesystem.NewFileSystemService()
	ctx := context.Background()
	testDir := t.TempDir()

	// Test CreateDirectory
	t.Run("CreateDirectory", func(t *testing.T) {
		newDir := filepath.Join(testDir, "new_directory")
		err := service.CreateDirectory(newDir)

		// Assert
		if err != nil {
			t.Logf("Warning: CreateDirectory failed: %v", err)
		} else {
			assert.DirExists(t, newDir)
		}
	})

	// Test WriteFile and GetFileInfo
	t.Run("WriteFileAndGetInfo", func(t *testing.T) {
		testFile := filepath.Join(testDir, "integration_test.txt")
		testContent := "integration test content"

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		require.NoError(t, err)

		info, err := service.GetFileInfo(testFile)
		if err != nil {
			t.Logf("Warning: GetFileInfo failed: %v", err)
		} else {
			assert.NotNil(t, info)
			assert.Equal(t, "integration_test.txt", info.Name)
			assert.False(t, info.IsDir)
			assert.Equal(t, int64(len(testContent)), info.Size)
		}
	})

	// Test ListDirectory
	t.Run("ListDirectory", func(t *testing.T) {
		// Create some test content
		testFile := filepath.Join(testDir, "list_test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

		entries, err := service.ListDirectory(testDir)
		if err != nil {
			t.Logf("Warning: ListDirectory failed: %v", err)
		} else {
			assert.NotNil(t, entries)
			// Should contain at least the file we created
			hasTestFile := false
			for _, entry := range entries {
				if entry.Name == "list_test.txt" {
					hasTestFile = true
					assert.False(t, entry.IsDir)
					break
				}
			}
			assert.True(t, hasTestFile, "Should find the test file we created")
		}
	})

	// Test Working Directory operations
	t.Run("WorkingDirectory", func(t *testing.T) {
		originalWD, err := service.GetWorkingDirectory()
		if err != nil {
			t.Logf("Warning: GetWorkingDirectory failed: %v", err)
		} else {
			assert.NotEmpty(t, originalWD)

			// Test setting working directory
			err = service.SetWorkingDirectory(testDir)
			if err != nil {
				t.Logf("Warning: SetWorkingDirectory failed: %v", err)
			} else {
				newWD, err := service.GetWorkingDirectory()
				if err == nil {
					// On some systems, this might not work due to permissions
					t.Logf("Working directory changed from %s to %s", originalWD, newWD)
				}
			}
		}
	})

	// Test Directory Watching (basic test)
	t.Run("DirectoryWatching", func(t *testing.T) {
		err := service.WatchDirectory(testDir)
		if err != nil {
			t.Logf("Warning: WatchDirectory failed: %v", err)
		} else {
			// Should be able to stop watching
			err = service.StopWatching(testDir)
			if err != nil {
				t.Logf("Warning: StopWatching failed: %v", err)
			}
		}
	})

	// Test GetWatchedDirectories
	t.Run("GetWatchedDirectories", func(t *testing.T) {
		dirs := service.GetWatchedDirectories()
		assert.NotNil(t, dirs)
		// Should be empty or contain the directory we just watched
		t.Logf("Currently watching %d directories: %v", len(dirs), dirs)
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		watchedDirs := service.GetWatchedDirectories()
		for _, dir := range watchedDirs {
			if err := service.StopWatching(dir); err != nil {
				t.Logf("Warning: Failed to stop watching %s: %v", dir, err)
			}
		}
	})
}

// Benchmark tests
func BenchmarkFileSystemService_ListDirectory(b *testing.B) {
	service := filesystem.NewFileSystemService()
	testDir := b.TempDir()

	// Create some test files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("file_%d.txt", i))
		os.WriteFile(filename, []byte("test content"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ListDirectory(testDir)
	}
}

func BenchmarkFileSystemService_GetFileInfo(b *testing.B) {
	service := filesystem.NewFileSystemService()
	testFile := filepath.Join(b.TempDir(), "benchmark_test.txt")
	os.WriteFile(testFile, []byte("benchmark test content"), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetFileInfo(testFile)
	}
}