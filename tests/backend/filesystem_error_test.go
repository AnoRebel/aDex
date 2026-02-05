package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aDex-UI/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemErrorCreation(t *testing.T) {
	tests := []struct {
		name        string
		errType     models.FileSystemErrorType
		path        string
		operation   string
		message     string
		originalErr error
	}{
		{
			name:      "permission error",
			errType:   models.ErrorPermission,
			path:      "/test/file.txt",
			operation: "read",
			message:   "access denied",
		},
		{
			name:        "not found error with original",
			errType:     models.ErrorNotFound,
			path:        "/test/missing.txt",
			operation:   "stat",
			message:     "file not found",
			originalErr: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.NewFileSystemError(tt.errType, tt.path, tt.operation, tt.message, tt.originalErr)

			assert.Equal(t, tt.errType, err.Type)
			assert.Equal(t, tt.path, err.Path)
			assert.Equal(t, tt.operation, err.Operation)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.originalErr, err.OriginalErr)
			assert.NotEmpty(t, err.UserMessage)
			assert.NotZero(t, err.Timestamp)
			assert.True(t, strings.Contains(err.Error(), string(tt.errType)))
		})
	}
}

func TestFileSystemErrorProperties(t *testing.T) {
	tests := []struct {
		name     string
		errType  models.FileSystemErrorType
		expected func(*testing.T, *models.FileSystemError)
	}{
		{
			name:    "retryable errors",
			errType: models.ErrorTimeout,
			expected: func(t *testing.T, err *models.FileSystemError) {
				assert.True(t, err.IsRetryable())
			},
		},
		{
			name:    "non-retryable errors",
			errType: models.ErrorPermission,
			expected: func(t *testing.T, err *models.FileSystemError) {
				assert.False(t, err.IsRetryable())
			},
		},
		{
			name:    "permission related errors",
			errType: models.ErrorReadOnly,
			expected: func(t *testing.T, err *models.FileSystemError) {
				assert.True(t, err.IsPermissionRelated())
			},
		},
		{
			name:    "non-permission related errors",
			errType: models.ErrorNotFound,
			expected: func(t *testing.T, err *models.FileSystemError) {
				assert.False(t, err.IsPermissionRelated())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.NewFileSystemError(tt.errType, "/test/file.txt", "test", "test message", nil)
			tt.expected(t, err)
		})
	}
}

func TestWrapFileSystemError(t *testing.T) {
	// Test wrapping nil error
	assert.Nil(t, models.WrapFileSystemError(nil, "/test", "operation"))

	// Test wrapping os error
	err := models.WrapFileSystemError(os.ErrPermission, "/test/file.txt", "read")
	assert.NotNil(t, err)
	assert.Equal(t, models.ErrorPermission, err.Type)
	assert.Equal(t, "/test/file.txt", err.Path)
	assert.Equal(t, "read", err.Operation)

	// Test wrapping existing FileSystemError
	originalErr := models.NewFileSystemError(models.ErrorNotFound, "/test/file.txt", "stat", "not found", nil)
	wrappedErr := models.WrapFileSystemError(originalErr, "/test/file.txt", "stat")
	assert.Equal(t, originalErr, wrappedErr)
}

func TestFileSystemEntryPermissionValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dex-perm-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test valid entry
	entry, err := models.NewFileSystemEntry(testFile)
	require.NoError(t, err)

	// Test permission validation
	err = entry.CheckPermissions("read")
	assert.NoError(t, err, "Should have read permission")

	err = entry.CheckPermissions("write")
	assert.NoError(t, err, "Should have write permission")

	// Test permission info update
	err = entry.UpdatePermissionsInfo()
	assert.NoError(t, err)
	assert.NotNil(t, entry.PermissionsInfo)
	assert.NotEmpty(t, entry.PermissionsInfo.Octal)
	assert.NotEmpty(t, entry.PermissionsInfo.Symbolic)
}

func TestFileSystemEntryPathValidation(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errorType   models.FileSystemErrorType
	}{
		{
			name:        "valid path",
			path:        "/home/user/file.txt",
			expectError: false,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
			errorType:   models.ErrorInvalidPath,
		},
		{
			name:        "path traversal",
			path:        "/home/user/../etc/passwd",
			expectError: true,
			errorType:   models.ErrorSecurity,
		},
		{
			name:        "path with null byte",
			path:        "/home/user\x00/file.txt",
			expectError: true,
			errorType:   models.ErrorInvalidPath,
		},
		{
			name:        "very long path",
			path:        strings.Repeat("/very/long/path", 300), // Needs to exceed 4096 chars
			expectError: true,
			errorType:   models.ErrorInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &models.FileSystemEntry{
				Name: filepath.Base(tt.path),
				Path: tt.path,
			}

			err := entry.ValidatePath()
			if tt.expectError {
				assert.Error(t, err)
				if fsErr, ok := err.(*models.FileSystemError); ok {
					assert.Equal(t, tt.errorType, fsErr.Type)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileSystemEntryOperationValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dex-op-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		operation   string
		expectError bool
	}{
		{
			name:        "valid read operation",
			path:        testFile,
			operation:   "read",
			expectError: false,
		},
		{
			name:        "valid write operation",
			path:        testFile,
			operation:   "write",
			expectError: false,
		},
		{
			name:        "invalid operation on system path",
			path:        "/etc/passwd",
			operation:   "delete",
			expectError: true,
		},
		// Note: The ValidateForOperation method only checks for directory existence,
		// not file existence when operation is "create". If we want to test the
		// create-on-existing behavior, we would need to set IsDir=true
		{
			name:        "create on existing directory",
			path:        tempDir, // Use a directory that exists
			operation:   "create",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Determine if it's a directory based on the path
			isDir := tt.path == tempDir || tt.path == "/etc/passwd"
			entry := &models.FileSystemEntry{
				Name:  filepath.Base(tt.path),
				Path:  tt.path,
				IsDir: isDir,
			}

			err := entry.ValidateForOperation(tt.operation)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileSystemEntryPermissionHelpers(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dex-helper-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test file with known permissions
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	entry, err := models.NewFileSystemEntry(testFile)
	require.NoError(t, err)

	// Test permission info update
	err = entry.UpdatePermissionsInfo()
	require.NoError(t, err)

	// Test effective permissions
	perms := entry.GetEffectivePermissions()
	assert.Len(t, perms, 3) // rwx format
	// Check that 'r' is in the first position (read) and 'w' in second (write)
	assert.Equal(t, byte('r'), perms[0])
	assert.Equal(t, byte('w'), perms[1])
	assert.Equal(t, byte('-'), perms[2]) // 0644 doesn't have execute for owner

	// Test accessibility
	assert.True(t, entry.IsAccessible())
	assert.True(t, entry.IsWritable())
}

func TestErrorCodeMapping(t *testing.T) {
	tests := []struct {
		name         string
		errType      models.FileSystemErrorType
		expectedCode int
	}{
		{"permission", models.ErrorPermission, 13},    // EACCES
		{"not found", models.ErrorNotFound, 2},        // ENOENT
		{"exists", models.ErrorExists, 17},            // EEXIST
		{"invalid path", models.ErrorInvalidPath, 22}, // EINVAL
		{"io error", models.ErrorIO, 5},               // EIO
		{"space", models.ErrorSpace, 28},              // ENOSPC
		{"read only", models.ErrorReadOnly, 30},       // EROFS
		{"busy", models.ErrorBusy, 16},                // EBUSY
		{"timeout", models.ErrorTimeout, 110},         // ETIMEDOUT
		{"unknown", models.ErrorUnknown, 1},           // EPERM
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.NewFileSystemError(tt.errType, "/test/file.txt", "test", "test", nil)
			assert.Equal(t, tt.expectedCode, err.GetErrorCode())
		})
	}
}

// Benchmark tests
func BenchmarkNewFileSystemError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = models.NewFileSystemError(models.ErrorPermission, "/test/file.txt", "read", "test message", os.ErrPermission)
	}
}

func BenchmarkWrapFileSystemError(b *testing.B) {
	err := os.ErrPermission
	for i := 0; i < b.N; i++ {
		_ = models.WrapFileSystemError(err, "/test/file.txt", "read")
	}
}

func BenchmarkUpdatePermissionsInfo(b *testing.B) {
	// Create a test file
	tempDir, err := os.MkdirTemp("", "dex-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	testFile := filepath.Join(tempDir, "bench.txt")
	err = os.WriteFile(testFile, []byte("bench"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	entry, err := models.NewFileSystemEntry(testFile)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entry.UpdatePermissionsInfo()
	}
}
