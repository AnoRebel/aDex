package security

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestLock(t *testing.T) *LockService {
	t.Helper()
	return &LockService{file: filepath.Join(t.TempDir(), "lock.json")}
}

func TestLock_RequiresPassphraseBeforeLocking(t *testing.T) {
	s := newTestLock(t)
	// Locking with no passphrase set would strand the user with no way back in.
	if err := s.Lock(); err == nil {
		t.Fatal("expected Lock to refuse when no passphrase is configured")
	}
	if s.IsLocked() {
		t.Error("service must not be locked when locking was refused")
	}
}

func TestLock_UnlockRequiresCorrectPassphrase(t *testing.T) {
	s := newTestLock(t)
	if err := s.SetPassphrase("", "correct horse"); err != nil {
		t.Fatalf("set passphrase: %v", err)
	}
	if err := s.Lock(); err != nil {
		t.Fatalf("lock: %v", err)
	}

	if err := s.Unlock("wrong"); err == nil {
		t.Error("expected the wrong passphrase to be rejected")
	}
	if !s.IsLocked() {
		t.Fatal("a failed unlock must leave the session locked")
	}
	if err := s.Unlock("correct horse"); err != nil {
		t.Fatalf("correct passphrase rejected: %v", err)
	}
	if s.IsLocked() {
		t.Error("session should be unlocked")
	}
}

// The passphrase itself must never be written to disk.
func TestLock_PassphraseIsNotStored(t *testing.T) {
	s := newTestLock(t)
	const secret = "super-secret-value"
	if err := s.SetPassphrase("", secret); err != nil {
		t.Fatalf("set: %v", err)
	}
	data, err := readFile(s.file)
	if err != nil {
		t.Fatalf("read lock file: %v", err)
	}
	if contains(data, secret) {
		t.Fatal("the passphrase was written to disk in recoverable form")
	}
}

func TestLock_ChangingPassphraseRequiresTheCurrentOne(t *testing.T) {
	s := newTestLock(t)
	if err := s.SetPassphrase("", "first"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.SetPassphrase("wrong", "second"); err == nil {
		t.Error("changing the passphrase must require the current one")
	}
	if err := s.SetPassphrase("first", "second"); err != nil {
		t.Errorf("correct current passphrase rejected: %v", err)
	}
}

func TestLock_DisableRequiresPassphrase(t *testing.T) {
	s := newTestLock(t)
	if err := s.SetPassphrase("", "pass"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.Disable("wrong"); err == nil {
		t.Error("disabling the lock must require the passphrase")
	}
	if err := s.Disable("pass"); err != nil {
		t.Errorf("disable with the correct passphrase failed: %v", err)
	}
	if s.IsConfigured() {
		t.Error("lock should no longer be configured")
	}
}

func TestLock_SurvivesReload(t *testing.T) {
	s := newTestLock(t)
	if err := s.SetPassphrase("", "persisted"); err != nil {
		t.Fatalf("set: %v", err)
	}
	reloaded := &LockService{file: s.file}
	reloaded.load()
	if !reloaded.IsConfigured() {
		t.Fatal("passphrase did not survive a reload")
	}
	if err := reloaded.Lock(); err != nil {
		t.Fatalf("lock after reload: %v", err)
	}
	if err := reloaded.Unlock("persisted"); err != nil {
		t.Errorf("unlock after reload: %v", err)
	}
}

func readFile(p string) (string, error) {
	b, err := os.ReadFile(p)
	return string(b), err
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		func() bool {
			for i := 0; i+len(needle) <= len(haystack); i++ {
				if haystack[i:i+len(needle)] == needle {
					return true
				}
			}
			return false
		}()
}
