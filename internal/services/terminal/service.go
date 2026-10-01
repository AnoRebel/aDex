package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"aDex/internal/utils"
	"aDex/internal/events"
	"github.com/creack/pty"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service handles terminal operations and emulation
type Service struct {
	platform  *utils.FeatureDetection
	terminals map[string]*Terminal
	termLock  sync.RWMutex
	nextID    int
	eventBus  events.IEventBus

	// shellOverride, when non-empty, takes priority over the env-var /
	// platform-default chain in getShellCommand(). Set via the public
	// SetShellCommand from Settings → Terminal → "Shell Path".
	shellOverride   string
	shellOverrideMu sync.RWMutex
}

// NewService creates a new terminal service instance
func NewService() *Service {
	return &Service{
		platform:  utils.DetectPlatform(),
		terminals: make(map[string]*Terminal),
		nextID:    1,
	}
}

// SetEventBus sets the event bus for the service
func (s *Service) SetEventBus(eventBus events.IEventBus) {
	s.eventBus = eventBus
}


// Terminal represents a terminal session.
//
// JSON tags are critical: this struct is returned by Wails-bound methods
// (`CreateTerminal`, `GetTerminalInfo`) and Wails marshals every binding
// return value through encoding/json. Unmarshalable fields (channels,
// *exec.Cmd, *os.File, sync.Once, sync.RWMutex) MUST be tagged `-` or
// the frontend launches with `FAT | json: unsupported type: func() error`.
//
// Done MUST be closed exactly once via signalDone(); both the reader goroutine
// (when PTY EOFs) and CloseTerminal race to it, and double-close panics.
type Terminal struct {
	ID         string        `json:"id"`
	Width      int           `json:"width"`
	Height     int           `json:"height"`
	WorkingDir string        `json:"workingDir"`
	Command    *exec.Cmd     `json:"-"`
	PTY        *os.File      `json:"-"`
	Output     chan []byte   `json:"-"`
	Input      chan []byte   `json:"-"`
	Done       chan struct{} `json:"-"`
	doneOnce   sync.Once
	mu         sync.RWMutex
}

// signalDone closes Done exactly once.
func (t *Terminal) signalDone() {
	t.doneOnce.Do(func() { close(t.Done) })
}

// CreateTerminal creates a new terminal session in the user's home
// directory. Equivalent to CreateTerminalIn(ctx, width, height, "").
func (s *Service) CreateTerminal(ctx context.Context, width, height int) (*Terminal, error) {
	return s.CreateTerminalIn(ctx, width, height, "")
}

// CreateTerminalIn creates a new terminal session whose shell starts in
// the given working directory. Empty `cwd` falls back to the user's home
// directory; an absolute path is used verbatim. Used by the frontend's
// Settings → System "Open in" preference (cwd vs home).
func (s *Service) CreateTerminalIn(ctx context.Context, width, height int, cwd string) (*Terminal, error) {
	if !s.platform.HasFeature("pty") {
		return nil, fmt.Errorf("PTY not supported on this platform: %s", s.platform.Platform.OS)
	}

	s.termLock.Lock()
	defer s.termLock.Unlock()

	id := fmt.Sprintf("term-%d", s.nextID)
	s.nextID++

	// Determine shell command
	shell := s.getShellCommand()

	// Pick the start directory. We accept anything the caller hands us
	// without revalidating — the OS will reject a bad path and `cmd.Dir`
	// will keep the user-visible error attached to the spawn failure
	// rather than masking it as a generic PTY error.
	startDir := cwd
	if startDir == "" {
		startDir = s.platform.GetHomeDirectory()
	}

	// Create command
	cmd := exec.CommandContext(ctx, shell[0], shell[1:]...)
	cmd.Dir = startDir

	// Set up environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TERM=xterm-256color")

	// Start PTY
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	// Set terminal size
	err = pty.Setsize(ptyFile, &pty.Winsize{
		Cols: uint16(width),
		Rows: uint16(height),
	})
	if err != nil {
		ptyFile.Close()
		return nil, fmt.Errorf("failed to set terminal size: %w", err)
	}

	// Create terminal instance
	terminal := &Terminal{
		ID:         id,
		Width:      width,
		Height:     height,
		Command:    cmd,
		PTY:        ptyFile,
		Output:     make(chan []byte, 1024),
		Input:      make(chan []byte, 1024),
		Done:       make(chan struct{}),
		WorkingDir: startDir,
	}

	// Store terminal
	s.terminals[id] = terminal

	// Start output reader
	go s.readTerminalOutput(terminal)

	// Start input writer
	go s.writeTerminalInput(terminal)

	return terminal, nil
}

// getShellCommand returns the appropriate shell command for the platform.
//
// Resolution order:
//  1. Explicit override set via SetShellCommand (Settings → Terminal →
//     "Shell Path"). Honored across all platforms.
//  2. Environment variable: $SHELL on Linux/macOS, $ComSpec on Windows.
//     This is the user's actual login shell — picking anything else
//     surprises them.
//  3. Platform default chain (zsh/bash/sh on *nix, PowerShell/cmd on
//     Windows).
//
// Each candidate is validated with exec.LookPath before returning so we
// never hand a bogus path to PTY.Start.
func (s *Service) getShellCommand() []string {
	// 1. Explicit settings override
	s.shellOverrideMu.RLock()
	override := s.shellOverride
	s.shellOverrideMu.RUnlock()
	if override != "" {
		if parts := s.parseShellCommand(override); parts != nil {
			return parts
		}
	}

	// 2. Environment-variable preferred shell
	if env := s.envPreferredShell(); env != nil {
		return env
	}

	// 3. Platform default chain
	switch s.platform.Platform.OS {
	case "windows":
		// PowerShell is the modern default; -NoLogo keeps the prompt clean.
		if path, err := exec.LookPath("pwsh.exe"); err == nil {
			return []string{path, "-NoLogo"}
		}
		if path, err := exec.LookPath("powershell.exe"); err == nil {
			return []string{path, "-NoLogo"}
		}
		// cmd.exe is the universal fallback — guaranteed to exist on
		// Windows. Look it up via $ComSpec first (handles non-default
		// installs); fall back to the literal name.
		if comspec := os.Getenv("ComSpec"); comspec != "" {
			if _, err := exec.LookPath(comspec); err == nil {
				return []string{comspec}
			}
		}
		return []string{"cmd.exe"}
	case "darwin":
		// macOS Catalina+ defaults to zsh; honor that, fall back to bash,
		// then sh which is always present.
		if path, err := exec.LookPath("zsh"); err == nil {
			return []string{path, "-l"}
		}
		if path, err := exec.LookPath("bash"); err == nil {
			return []string{path, "-l"}
		}
		return []string{"/bin/sh"}
	default:
		// Linux + other Unix: prefer zsh if installed, then bash, then sh.
		// (The original code looked up bash first, which clobbered users
		// whose $SHELL was zsh — that's the bug we're fixing here.)
		for _, candidate := range []string{"zsh", "bash"} {
			if path, err := exec.LookPath(candidate); err == nil {
				return []string{path, "-l"}
			}
		}
		return []string{"/bin/sh"}
	}
}

// envPreferredShell returns the user's preferred shell from the
// environment, or nil if it can't be resolved. Linux/macOS use $SHELL,
// Windows uses $ComSpec.
func (s *Service) envPreferredShell() []string {
	if s.platform.Platform.OS == "windows" {
		// $ComSpec is the canonical "what shell did the user launch this
		// process under" on Windows. We don't pass any args because cmd
		// has no equivalent of -l, and PowerShell's invocation flags
		// belong on the explicit-override path instead.
		if cs := os.Getenv("ComSpec"); cs != "" {
			if _, err := exec.LookPath(cs); err == nil {
				return []string{cs}
			}
		}
		return nil
	}
	// $SHELL is set by login shells on Linux/macOS. We add `-l` so the
	// shell sources its login dotfiles (.zprofile, .bash_profile) — the
	// user's PATH and aliases come along.
	if sh := os.Getenv("SHELL"); sh != "" {
		if _, err := exec.LookPath(sh); err == nil {
			return []string{sh, "-l"}
		}
	}
	return nil
}

// parseShellCommand parses a shell command string into command and args.
// Returns nil if the command can't be located on PATH.
func (s *Service) parseShellCommand(shell string) []string {
	parts := strings.Fields(shell)
	if len(parts) == 0 {
		return nil
	}
	if _, err := exec.LookPath(parts[0]); err != nil {
		return nil
	}
	return parts
}

// SetShellCommand stores a user-supplied shell-path override. Empty
// string clears the override and falls back to env / platform default.
// Cross-platform: accepts "/usr/bin/zsh -l", "C:\\Windows\\System32\\
// cmd.exe", "pwsh -NoLogo -NoProfile", etc. Validation happens at
// terminal-spawn time, not here.
func (s *Service) SetShellCommand(shell string) error {
	s.shellOverrideMu.Lock()
	s.shellOverride = strings.TrimSpace(shell)
	s.shellOverrideMu.Unlock()
	return nil
}

// GetActiveShell returns the shell command that getShellCommand() would
// pick right now — used by the frontend to render the status bar
// (`/bin/zsh -- /home/ano`) honestly instead of guessing '/bin/bash'.
// Returns just the executable path (the first arg), not the full argv.
func (s *Service) GetActiveShell() string {
	parts := s.getShellCommand()
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// GetAvailableShells returns a list of available shells on the system
func (s *Service) GetAvailableShells(ctx context.Context) ([]string, error) {
	var shells []string

	switch s.platform.Platform.OS {
	case "windows":
		// Check for Windows shells
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			shells = append(shells, "powershell.exe")
		}
		if _, err := exec.LookPath("cmd.exe"); err == nil {
			shells = append(shells, "cmd.exe")
		}
		if _, err := exec.LookPath("wsl.exe"); err == nil {
			shells = append(shells, "wsl.exe")
		}
	default:
		// Check for Unix-like shells
		unixShells := []string{"bash", "zsh", "fish", "sh", "dash", "ksh", "tcsh", "csh"}
		for _, shell := range unixShells {
			if _, err := exec.LookPath(shell); err == nil {
				shells = append(shells, shell)
			}
		}
	}

	return shells, nil
}

// readTerminalOutput reads output from the PTY
func (s *Service) readTerminalOutput(terminal *Terminal) {
	defer terminal.signalDone()
	// Output is intentionally not closed here: Close races with this goroutine
	// and consumers (Wails event listeners + ReadFromTerminal channel readers)
	// must rely on Done as the cancellation signal, not a closed Output chan.

	// Emit a PER-SESSION exit event when the PTY closes (user typed
	// `exit`, shell crashed, or backend Close fired). We use the
	// session-id-scoped event name (`terminal.exited.<id>`) rather
	// than a global `terminal.exited` so each AdexTerminal can
	// EventsOff its OWN listener without accidentally removing the
	// listeners other tabs registered (the previous global path
	// silently disconnected output for ALL tabs whenever any one
	// component cleaned up).
	defer func() {
		if app := application.Get(); app != nil {
			app.Event.Emit(
				"terminal.exited."+terminal.ID,
				map[string]interface{}{"terminalId": terminal.ID},
			)
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := terminal.PTY.Read(buf)
		if err != nil {
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])

			// Emit Wails event to frontend. We emit ONLY the per-session
			// event name (`terminal.output.<id>`) so each AdexTerminal
			// can EventsOff its own listener at unmount without
			// disconnecting other tabs — Wails' EventsOff is keyed by
			// event NAME, so a global `terminal.output` listener
			// shared across N tabs gets nuked the first time any tab
			// cleans up. Wails v3 emission needs no caller context;
			// application.Get() returns nil before the app is up
			// (e.g. under unit tests), and the emit is skipped.
			if app := application.Get(); app != nil {
				app.Event.Emit(
					"terminal.output."+terminal.ID,
					map[string]interface{}{
						"terminalId": terminal.ID,
						"data":       data,
					},
				)
			}

			// Send to Go channel (may block if channel is full)
			select {
			case terminal.Output <- data:
			case <-terminal.Done:
				return
			default:
				// Channel full, skip Go-side delivery but Wails event was already sent
			}
		}
	}
}

// writeTerminalInput writes input to the PTY
func (s *Service) writeTerminalInput(terminal *Terminal) {
	for {
		select {
		case data := <-terminal.Input:
			if terminal.PTY != nil {
				terminal.PTY.Write(data)
			}
		case <-terminal.Done:
			return
		}
	}
}

// ResizeTerminal resizes an existing terminal session
func (s *Service) ResizeTerminal(ctx context.Context, terminalID string, width, height int) error {
	s.termLock.RLock()
	terminal, exists := s.terminals[terminalID]
	s.termLock.RUnlock()

	if !exists {
		return fmt.Errorf("terminal not found: %s", terminalID)
	}

	terminal.mu.Lock()
	defer terminal.mu.Unlock()

	if terminal.PTY != nil {
		err := pty.Setsize(terminal.PTY, &pty.Winsize{
			Cols: uint16(width),
			Rows: uint16(height),
		})
		if err != nil {
			return fmt.Errorf("failed to resize terminal: %w", err)
		}

		terminal.Width = width
		terminal.Height = height
	}

	return nil
}

// WriteToTerminal writes input to the terminal
func (s *Service) WriteToTerminal(ctx context.Context, terminalID string, data []byte) error {
	s.termLock.RLock()
	terminal, exists := s.terminals[terminalID]
	s.termLock.RUnlock()

	if !exists {
		return fmt.Errorf("terminal not found: %s", terminalID)
	}

	select {
	case terminal.Input <- data:
		return nil
	case <-terminal.Done:
		return fmt.Errorf("terminal is closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReadFromTerminal reads output from the terminal
func (s *Service) ReadFromTerminal(ctx context.Context, terminalID string) (<-chan []byte, error) {
	s.termLock.RLock()
	terminal, exists := s.terminals[terminalID]
	s.termLock.RUnlock()

	if !exists {
		return nil, fmt.Errorf("terminal not found: %s", terminalID)
	}

	return terminal.Output, nil
}

// Shutdown closes every active terminal session and kills its child
// process. Called from coordinator.Shutdown() so the app exits cleanly
// without leaking PTY children — Linux/macOS will reparent orphans to
// init, but Windows handles them differently and Wails dev observed
// these as the reason the binary refused to die ("wails dev" stayed
// alive because the child shells held console handles).
func (s *Service) Shutdown() {
	s.termLock.Lock()
	ids := make([]string, 0, len(s.terminals))
	for id := range s.terminals {
		ids = append(ids, id)
	}
	s.termLock.Unlock()
	// Close each terminal outside the lock (CloseTerminal takes the
	// lock itself; doing it inline would deadlock).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, id := range ids {
		_ = s.CloseTerminal(ctx, id)
	}
}

// CloseTerminal closes a terminal session
func (s *Service) CloseTerminal(ctx context.Context, terminalID string) error {
	s.termLock.Lock()
	defer s.termLock.Unlock()

	terminal, exists := s.terminals[terminalID]
	if !exists {
		return fmt.Errorf("terminal not found: %s", terminalID)
	}

	// Signal goroutines to stop. Safe under double-close: signalDone uses
	// sync.Once internally, so the reader goroutine racing with us is fine.
	terminal.signalDone()

	// Close PTY
	if terminal.PTY != nil {
		terminal.PTY.Close()
	}

	// Kill the process if it's still running
	if terminal.Command != nil && terminal.Command.Process != nil {
		terminal.Command.Process.Kill()
	}

	// Remove from terminals map
	delete(s.terminals, terminalID)

	return nil
}

// ListTerminals returns a list of all active terminal IDs
func (s *Service) ListTerminals(ctx context.Context) ([]string, error) {
	s.termLock.RLock()
	defer s.termLock.RUnlock()

	var terminals []string
	for id := range s.terminals {
		terminals = append(terminals, id)
	}

	return terminals, nil
}

// GetTerminalInfo returns information about a specific terminal
func (s *Service) GetTerminalInfo(ctx context.Context, terminalID string) (*Terminal, error) {
	s.termLock.RLock()
	defer s.termLock.RUnlock()

	terminal, exists := s.terminals[terminalID]
	if !exists {
		return nil, fmt.Errorf("terminal not found: %s", terminalID)
	}

	// Return a copy to avoid external modification
	info := &Terminal{
		ID:         terminal.ID,
		Width:      terminal.Width,
		Height:     terminal.Height,
		WorkingDir: terminal.WorkingDir,
	}

	return info, nil
}
