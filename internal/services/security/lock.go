package security

import (
	"aDex-UI/internal/appdir"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/argon2"
)

// Lock state for the session.
//
// Scope, stated plainly: this gates an already-running local session against
// someone who walks up to an unattended machine. It is NOT authentication and
// NOT a substitute for OS-level security — the application runs as the
// invoking user and anyone with access to that account can open a terminal
// directly, without going through aDex at all.
//
// The lock is enforced in the backend rather than only drawn over the UI: a
// curtain in the frontend is trivially bypassed while the bound methods behind
// it still answer.

const lockFileName = "lock.json"

// Argon2id parameters. Deliberately not the cheapest available: the stored
// value protects against someone with filesystem access trying to recover the
// passphrase offline.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

type lockFile struct {
	// Base64 salt and Argon2id hash. The passphrase itself is never stored.
	Salt string `json:"salt"`
	Hash string `json:"hash"`
	// IdleTimeoutSeconds of 0 means "never lock automatically".
	IdleTimeoutSeconds int  `json:"idleTimeoutSeconds"`
	Enabled            bool `json:"enabled"`
}

// LockService owns the locked state and the passphrase.
type LockService struct {
	mu     sync.RWMutex
	locked bool
	file   string
	cfg    lockFile
}

func NewLockService() *LockService {
	path := filepath.Join(appdir.Config(), lockFileName)
	s := &LockService{file: path}
	s.load()
	return s
}

func (s *LockService) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return // No lock configured; that is the default state.
	}
	var cfg lockFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	s.cfg = cfg
}

func (s *LockService) persist() error {
	if err := os.MkdirAll(filepath.Dir(s.file), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	// 0600: the hash must not be world-readable.
	return os.WriteFile(s.file, data, 0o600)
}

func hashPassphrase(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

// SetPassphrase configures (or changes) the unlock passphrase and enables the
// lock. Changing it requires the current one once a passphrase is already set,
// so an unattended unlocked session cannot simply be re-keyed.
func (s *LockService) SetPassphrase(current, next string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cfg.Hash != "" {
		if !s.verifyLocked(current) {
			return errors.New("current passphrase is incorrect")
		}
	}
	if len(next) < 4 {
		return errors.New("passphrase must be at least 4 characters")
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	s.cfg.Salt = base64.StdEncoding.EncodeToString(salt)
	s.cfg.Hash = base64.StdEncoding.EncodeToString(hashPassphrase(next, salt))
	s.cfg.Enabled = true
	return s.persist()
}

// verifyLocked checks a passphrase. Caller must hold the lock.
func (s *LockService) verifyLocked(passphrase string) bool {
	if s.cfg.Hash == "" || s.cfg.Salt == "" {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(s.cfg.Salt)
	if err != nil {
		return false
	}
	want, err := base64.StdEncoding.DecodeString(s.cfg.Hash)
	if err != nil {
		return false
	}
	got := hashPassphrase(passphrase, salt)
	// Constant time: a length-independent comparison avoids leaking how much
	// of the hash matched via timing.
	return subtle.ConstantTimeCompare(got, want) == 1
}

// Lock engages the lock. A no-op when no passphrase is configured, so a user
// cannot lock themselves out with no way back in.
func (s *LockService) Lock() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.Hash == "" {
		return errors.New("no passphrase configured")
	}
	s.locked = true
	return nil
}

// Unlock clears the lock when the passphrase matches.
func (s *LockService) Unlock(passphrase string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.locked {
		return nil
	}
	if !s.verifyLocked(passphrase) {
		return errors.New("incorrect passphrase")
	}
	s.locked = false
	return nil
}

// IsLocked reports the current state. Bound methods consult this before
// serving, which is what makes the lock more than a UI curtain.
func (s *LockService) IsLocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.locked
}

// IsConfigured reports whether a passphrase has been set.
func (s *LockService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Hash != ""
}

// IdleTimeoutSeconds returns the auto-lock delay; 0 means never.
func (s *LockService) IdleTimeoutSeconds() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.IdleTimeoutSeconds
}

// SetIdleTimeout configures the auto-lock delay in seconds. 0 disables it.
func (s *LockService) SetIdleTimeout(seconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seconds < 0 {
		seconds = 0
	}
	s.cfg.IdleTimeoutSeconds = seconds
	return s.persist()
}

// Disable removes the lock entirely. Requires the current passphrase, so an
// unattended session cannot be un-secured by a passer-by.
func (s *LockService) Disable(passphrase string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cfg.Hash != "" && !s.verifyLocked(passphrase) {
		return errors.New("incorrect passphrase")
	}
	s.cfg = lockFile{}
	s.locked = false
	// Remove the file rather than leaving an empty one behind.
	if err := os.Remove(s.file); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ErrLocked is returned by bound methods that refuse to serve while locked.
var ErrLocked = errors.New("session is locked")
