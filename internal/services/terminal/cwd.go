package terminal

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// CWDStats summarizes per-session CWD-tracking state for the frontend.
// Mirrors the legacy /api/terminal/cwd/stats response shape so existing
// clients can adopt the new binding without reshaping.
type CWDStats struct {
	Enabled        bool                  `json:"enabled"`
	Platform       string                `json:"platform"`
	Sessions       int                   `json:"sessions"`
	Resolutions    int64                 `json:"resolutions"`
	LastResolvedAt int64                 `json:"lastResolvedAt"`
	Strategy       string                `json:"strategy"`
	PerSession     map[string]CWDSession `json:"perSession"`
}

// CWDSession is a single session's last-known CWD.
type CWDSession struct {
	CWD       string `json:"cwd"`
	UpdatedAt int64  `json:"updatedAt"`
}

// cwdResolutions counts successful CWD lookups so verify:evidence can prove
// the tracker is actually running.
var cwdResolutions int64

// GetCurrentCWD returns the live working directory of the shell process
// backing the given terminal. Strategy mirrors the original eDEX-UI:
//
//	Linux:   readlink /proc/<pid>/cwd
//	Darwin:  lsof -a -d cwd -p <pid> | tail -1 | awk '...'
//	Windows: not supported by upstream eDEX-UI; we fall back to the
//	         recorded WorkingDir (updated on path injection from the UI).
//
// Source: edex-ui/src/classes/terminal.class.js  _getTtyCWD()
func (s *Service) GetCurrentCWD(terminalID string) (string, error) {
	s.termLock.RLock()
	t, ok := s.terminals[terminalID]
	s.termLock.RUnlock()
	if !ok {
		return "", fmt.Errorf("terminal %q not found", terminalID)
	}
	if t.Command == nil || t.Command.Process == nil {
		return t.WorkingDir, nil
	}

	pid := t.Command.Process.Pid
	cwd, err := resolvePidCWD(pid)
	if err != nil || cwd == "" {
		// Fall back to the recorded value rather than fail outright.
		t.mu.Lock()
		recorded := t.WorkingDir
		t.mu.Unlock()
		if recorded != "" {
			return recorded, nil
		}
		return "", err
	}

	t.mu.Lock()
	t.WorkingDir = cwd
	t.mu.Unlock()
	atomic.AddInt64(&cwdResolutions, 1)
	return cwd, nil
}

// GetCWDStats returns aggregate tracker state.
func (s *Service) GetCWDStats() CWDStats {
	s.termLock.RLock()
	defer s.termLock.RUnlock()

	per := make(map[string]CWDSession, len(s.terminals))
	for id, t := range s.terminals {
		t.mu.RLock()
		per[id] = CWDSession{CWD: t.WorkingDir, UpdatedAt: time.Now().Unix()}
		t.mu.RUnlock()
	}

	return CWDStats{
		Enabled:        cwdStrategy() != "recorded",
		Platform:       runtime.GOOS,
		Sessions:       len(per),
		Resolutions:    atomic.LoadInt64(&cwdResolutions),
		LastResolvedAt: time.Now().Unix(),
		Strategy:       cwdStrategy(),
		PerSession:     per,
	}
}

// cwdStrategy reports the platform-specific resolution strategy in use.
func cwdStrategy() string {
	switch runtime.GOOS {
	case "linux":
		return "procfs"
	case "darwin":
		return "lsof"
	default:
		return "recorded"
	}
}
