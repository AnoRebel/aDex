package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"aDex-UI/backend/utils"
	"aDex-UI/internal/events"
	"github.com/creack/pty"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Service handles terminal operations and emulation
type Service struct {
	platform  *utils.FeatureDetection
	terminals map[string]*Terminal
	termLock  sync.RWMutex
	nextID    int
	eventBus  events.IEventBus
	wailsCtx  context.Context // Wails context for emitting frontend events
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

// SetWailsContext sets the Wails context for emitting frontend events
func (s *Service) SetWailsContext(ctx context.Context) {
	s.wailsCtx = ctx
}

// Terminal represents a terminal session
type Terminal struct {
	ID         string
	Width      int
	Height     int
	Command    *exec.Cmd
	PTY        *os.File
	Output     chan []byte
	Input      chan []byte
	Done       chan struct{}
	WorkingDir string
	mu         sync.RWMutex
}

// CreateTerminal creates a new terminal session
func (s *Service) CreateTerminal(ctx context.Context, width, height int) (*Terminal, error) {
	if !s.platform.HasFeature("pty") {
		return nil, fmt.Errorf("PTY not supported on this platform: %s", s.platform.Platform.OS)
	}

	s.termLock.Lock()
	defer s.termLock.Unlock()

	id := fmt.Sprintf("term-%d", s.nextID)
	s.nextID++

	// Determine shell command
	shell := s.getShellCommand()

	// Create command
	cmd := exec.CommandContext(ctx, shell[0], shell[1:]...)

	// Set up environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=xterm-256color"))

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
		WorkingDir: s.platform.GetHomeDirectory(),
	}

	// Store terminal
	s.terminals[id] = terminal

	// Start output reader
	go s.readTerminalOutput(terminal)

	// Start input writer
	go s.writeTerminalInput(terminal)

	return terminal, nil
}

// getShellCommand returns the appropriate shell command for the platform
func (s *Service) getShellCommand() []string {
	// Try to get shell from settings first (this would be integrated with config service)
	if shell := s.getShellFromSettings(); shell != "" {
		return s.parseShellCommand(shell)
	}

	// Fall back to platform defaults
	switch s.platform.Platform.OS {
	case "windows":
		// On Windows, try to find PowerShell or cmd
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			return []string{"powershell.exe", "-NoExit", "-Command", "-"}
		}
		return []string{"cmd.exe"}
	case "darwin":
		// On macOS, use zsh if available, otherwise bash
		if _, err := exec.LookPath("zsh"); err == nil {
			return []string{"zsh", "-l"}
		}
		return []string{"bash", "-l"}
	default:
		// On Linux and other Unix-like systems
		if _, err := exec.LookPath("bash"); err == nil {
			return []string{"bash", "-l"}
		}
		if shell := os.Getenv("SHELL"); shell != "" {
			return []string{shell, "-l"}
		}
		return []string{"sh"}
	}
}

// getShellFromSettings retrieves shell preference from settings
func (s *Service) getShellFromSettings() string {
	// This would integrate with the config service
	// For now, return empty to use defaults
	// TODO: Integrate with config service when implemented
	return ""
}

// parseShellCommand parses a shell command string into command and args
func (s *Service) parseShellCommand(shell string) []string {
	// Simple parsing - can be enhanced for complex shell commands
	parts := strings.Fields(shell)
	if len(parts) == 0 {
		return nil
	}

	// Validate that the shell exists
	if _, err := exec.LookPath(parts[0]); err != nil {
		return nil
	}

	return parts
}

// SetShellCommand sets a custom shell command (to be called from settings)
func (s *Service) SetShellCommand(shell string) error {
	// This would save to settings/config
	// For now, this is a placeholder
	return nil
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
	defer close(terminal.Done)
	defer close(terminal.Output)

	buf := make([]byte, 1024)
	for {
		n, err := terminal.PTY.Read(buf)
		if err != nil {
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])

			// Emit Wails event directly to frontend FIRST (non-blocking)
			if s.wailsCtx != nil {
				wailsRuntime.EventsEmit(s.wailsCtx, "terminal.output", map[string]interface{}{
					"terminalId": terminal.ID,
					"data":       data,
				})
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

// CloseTerminal closes a terminal session
func (s *Service) CloseTerminal(ctx context.Context, terminalID string) error {
	s.termLock.Lock()
	defer s.termLock.Unlock()

	terminal, exists := s.terminals[terminalID]
	if !exists {
		return fmt.Errorf("terminal not found: %s", terminalID)
	}

	// Signal goroutines to stop
	close(terminal.Done)

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
