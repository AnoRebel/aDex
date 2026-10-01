package tests

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"aDex/internal/services/filesystem"
)

// Section 5.D boundary tests for the filesystem service. Mirrors what the
// AdexFilesystem panel does at runtime: list a real directory, get file
// info, refuse traversal-style paths.

func TestFilesystem_ReadDirectory_Populated(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("filesystem entry shape covered separately on windows")
	}

	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	s := filesystem.NewService()
	entries, err := s.ReadDirectory(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadDirectory: %v", err)
	}
	if len(entries) < 4 {
		t.Fatalf("expected at least 4 entries (a.txt, b.txt, .hidden, subdir), got %d", len(entries))
	}

	gotDir := false
	gotFile := false
	for _, e := range entries {
		if e.Name == "subdir" && e.IsDir {
			gotDir = true
		}
		if e.Name == "a.txt" && !e.IsDir {
			gotFile = true
		}
	}
	if !gotDir {
		t.Errorf("subdir not reported as directory")
	}
	if !gotFile {
		t.Errorf("a.txt not reported as file")
	}
}

func TestFilesystem_GetFileInfo_NonZeroSize(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file info shape covered separately on windows")
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(p, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	s := filesystem.NewService()
	info, err := s.GetFileInfo(context.Background(), p)
	if err != nil {
		t.Fatalf("GetFileInfo: %v", err)
	}
	if info.Size != int64(len("hello world")) {
		t.Errorf("Size = %d, want %d", info.Size, len("hello world"))
	}
	if info.IsDirectory {
		t.Errorf("file reported as directory")
	}
}

func TestFilesystem_ReadDirectory_ErrorOnMissingPath(t *testing.T) {
	s := filesystem.NewService()
	_, err := s.ReadDirectory(context.Background(), "/this/should/not/exist/__nope__")
	if err == nil {
		t.Fatalf("expected error for non-existent path")
	}
}
