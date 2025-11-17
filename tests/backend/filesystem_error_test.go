package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystemErrorCreation(t *testing.T) {
	tests := []struct {
		name        string
		errType     FileSystemErrorType
		path        string
		operation   string
		message     string
		originalErr error
	}{
		{
			name:      "permission error",
			errType:   ErrorPermission,
			path:      "/test/file.txt",
			operation: "read",
			message:   "access denied",
		},
		{
			name:        "not found error with original",
			errType:     ErrorNotFound,
			path:        "/test/missing.txt",
			operation:   "stat",
			message:     "file not found",
			originalErr: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewFileSystemError(tt.errType, tt.path, tt.operation, tt.message, tt.originalErr)

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
		name        string
		errType     FileSystemErrorType
		expected    func(*testing.T, *FileSystemError)
	}{
		{
			name:    "retryable errors",
			errType: ErrorTimeout,
			expected: func(t *testing.T, err *FileSystemError) {
				assert.True(t, err.IsRetryable())
			},
		},
		{
			name:    "non-retryable errors",
			errType: ErrorPermission,
			expected: func(t *testing.T, err *FileSystemError) {
				assert.False(t, err.IsRetryable())
			},
		},
		{
			name:    "permission related errors",
			errType: ErrorReadOnly,
			expected: func(t *testing.T, err *FileSystemError) {
				assert.True(t, err.IsPermissionRelated())
			},
		},
		{
			name:    "non-permission related errors",
			errType: ErrorNotFound,
			expected: func(t *testing.T, err *FileSystemError) {
				assert.False(t, err.IsPermissionRelated())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewFileSystemError(tt.errType, "/test/file.txt", "test", "test message", nil)
			tt.expected(t, err)
		})
	}
}

func TestWrapError(t *testing.T) {
	// Test wrapping nil error
	assert.Nil(t, WrapError(nil, "/test", "operation"))

	// Test wrapping os error
	err := WrapError(os.ErrPermission, "/test/file.txt", "read")
	assert.NotNil(t, err)
	assert.Equal(t, ErrorPermission, err.Type)
	assert.Equal(t, "/test/file.txt", err.Path)
	assert.Equal(t, "read", err.Operation)

	// Test wrapping existing FileSystemError
	originalErr := NewFileSystemError(ErrorNotFound, "/test/file.txt", "stat", "not found", nil)
	wrappedErr := WrapError(originalErr, "/test/file.txt", "stat")
	assert.Equal(t, originalErr, wrappedErr)
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected FileSystemErrorType
	}{
		{"nil error", nil, ErrorUnknown},
		{"permission error", os.ErrPermission, ErrorPermission},
		{"exist error", os.ErrExist, ErrorExists},
		{"not exist error", os.ErrNotExist, ErrorNotFound},
		{"custom permission error", &os.PathError{Err: os.ErrPermission}, ErrorPermission},
		{"generic error with permission text", assert.AnError, ErrorUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		errType  FileSystemErrorType
		path     string
		operation string
		expected string
	}{
		{
			name:      "permission error",
			errType:   ErrorPermission,
			path:      "/home/user/file.txt",
			operation: "read",
			expected:  "You don't have permission to read 'file.txt'",
		},
		{
			name:      "not found error",
			errType:   ErrorNotFound,
			path:      "/home/user/missing.txt",
			operation: "stat",
			expected:  "The file or folder 'missing.txt' could not be found",
		},
		{
			name:      "space error",
			errType:   ErrorSpace,
			path:      "/tmp",
			operation: "write",
			expected:  "There is not enough disk space to complete this operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUserMessage(tt.errType, tt.path, tt.operation)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestFileSystemEntryPermissionValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "dex-perm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test valid entry
	entry, err := NewFileSystemEntry(testFile)
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
		errorType   FileSystemErrorType
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
			errorType:   ErrorInvalidPath,
		},
		{
			name:        "path traversal",
			path:        "/home/user/../etc/passwd",
			expectError: true,
			errorType:   ErrorSecurity,
		},
		{
			name:        "path with null byte",
			path:        "/home/user\x00/file.txt",
			expectError: true,
			errorType:   ErrorInvalidPath,
		},
		{
			name:        "very long path",
			path:        strings.Repeat("/very/long/path", 100),
			expectError: true,
			errorType:   ErrorInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &FileSystemEntry{
				Name: filepath.Base(tt.path),
				Path: tt.path,
			}

			err := entry.ValidatePath()
			if tt.expectError {
				assert.Error(t, err)
				if fsErr, ok := err.(*FileSystemError); ok {
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
	defer os.RemoveAll(tempDir)

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
		{
			name:        "create on existing file",
			path:        testFile,
			operation:   "create",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &FileSystemEntry{
				Name: filepath.Base(tt.path),
				Path: tt.path,
				IsDir: false,
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
	defer os.RemoveAll(tempDir)

	// Create a test file with known permissions
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	entry, err := NewFileSystemEntry(testFile)
	require.NoError(t, err)

	// Test permission info update
	err = entry.UpdatePermissionsInfo()
	require.NoError(t, err)

	// Test effective permissions
	perms := entry.GetEffectivePermissions()
	assert.Len(t, perms, 3) // rwx format
	assert.Contains(t, perms, 'r')
	assert.Contains(t, perms, 'w')
	assert.NotContains(t, perms, 'x') // 0644 doesn't have execute

	// Test accessibility
	assert.True(t, entry.IsAccessible())
	assert.True(t, entry.IsWritable())
}

func TestFormatPermissionBits(t *testing.T) {
	tests := []struct {
		name       string
		permissions uint32
		highBit     int
		lowBit      int
		expected    string
	}{
		{
			name:        "read only",
			permissions: 0o444,
			highBit:     6,
			lowBit:      3,
			expected:    "r--",
		},
		{
			name:        "read write",
			permissions: 0o666,
			highBit:     6,
			lowBit:      3,
			expected:    "rw-",
		},
		{
			name:        "full permissions",
			permissions: 0o777,
			highBit:     6,
			lowBit:      3,
			expected:    "rwx",
		},
		{
			name:        "no permissions",
			permissions: 0o000,
			highBit:     6,
			lowBit:      3,
			expected:    "---",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPermissionBits(tt.permissions, tt.highBit, tt.lowBit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorCodeMapping(t *testing.T) {
	tests := []struct {
		name         string
		errType      FileSystemErrorType
		expectedCode int
	}{
		{"permission", ErrorPermission, 13},  // EACCES
		{"not found", ErrorNotFound, 2},      // ENOENT
		{"exists", ErrorExists, 17},         // EEXIST
		{"invalid path", ErrorInvalidPath, 22}, // EINVAL
		{"io error", ErrorIO, 5},            // EIO
		{"space", ErrorSpace, 28},           // ENOSPC
		{"read only", ErrorReadOnly, 30},     // EROFS
		{"busy", ErrorBusy, 16},             // EBUSY
		{"timeout", ErrorTimeout, 110},      // ETIMEDOUT
		{"unknown", ErrorUnknown, 1},        // EPERM
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewFileSystemError(tt.errType, "/test/file.txt", "test", "test", nil)
			assert.Equal(t, tt.expectedCode, err.GetErrorCode())
		})
	}
}

// Benchmark tests
func BenchmarkNewFileSystemError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFileSystemError(ErrorPermission, "/test/file.txt", "read", "test message", os.ErrPermission)
	}
}

func BenchmarkWrapError(b *testing.B) {
	err := os.ErrPermission
	for i := 0; i < b.N; i++ {
		WrapError(err, "/test/file.txt", "read")
	}
}

func BenchmarkUpdatePermissionsInfo(b *testing.B) {
	// Create a test file
	tempDir, err := os.MkdirTemp("", "dex-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "bench.txt")
	err = os.WriteFile(testFile, []byte("bench"), 0644)
	if err != nil {
		b.Fatal(err)
	}

	entry, err := NewFileSystemEntry(testFile)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.UpdatePermissionsInfo()
	}
}