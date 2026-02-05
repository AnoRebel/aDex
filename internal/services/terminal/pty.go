package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"aDex-UI/internal/models"
	"golang.org/x/crypto/ssh/terminal"
)

// PTY represents a pseudo-terminal pair
type PTY struct {
	master       *os.File
	slave        *os.File
	cmd          *exec.Cmd
	sessionID    string
	size         *models.TerminalSize
	active       bool
	created      time.Time
	lastActivity time.Time
}

// PTYManager manages PTY operations
type PTYManager struct {
	ptyMap map[string]*PTY
}

// NewPTYManager creates a new PTY manager
func NewPTYManager() *PTYManager {
	return &PTYManager{
		ptyMap: make(map[string]*PTY),
	}
}

// CreatePTY creates a new pseudo-terminal pair
func (pm *PTYManager) CreatePTY(ctx context.Context, sessionID string, options *models.TerminalOptions) (*PTY, error) {
	pty := &PTY{
		sessionID:    sessionID,
		size:         &models.TerminalSize{Rows: options.Rows, Cols: options.Cols},
		active:       false,
		created:      time.Now(),
		lastActivity: time.Now(),
	}

	var err error

	// Platform-specific PTY creation
	switch runtime.GOOS {
	case "windows":
		err = pty.createWindowsPTY(options)
	case "darwin", "linux":
		err = pty.createUnixPTY(options)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create PTY: %w", err)
	}

	// Store PTY in manager
	pm.ptyMap[sessionID] = pty
	pty.active = true

	return pty, nil
}

// Open opens a new PTY pair (master and slave)
func (pty *PTY) Open() (*os.File, *os.File, error) {
	// Open the master PTY device
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open /dev/ptmx: %w", err)
	}

	// Get the slave device name
	slaveName, err := ptsname(master)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("failed to get slave name: %w", err)
	}

	// Unlock the slave device
	if err := unlockpt(master); err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("failed to unlock slave: %w", err)
	}

	// Open the slave device
	slave, err := os.OpenFile(slaveName, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("failed to open slave %s: %w", slaveName, err)
	}

	return master, slave, nil
}

// ptsname returns the name of the slave PTY device
func ptsname(master *os.File) (string, error) {
	var n uint32
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n)))
	if errno != 0 {
		return "", errno
	}
	return fmt.Sprintf("/dev/pts/%d", n), nil
}

// unlockpt unlocks the slave PTY device
func unlockpt(master *os.File) error {
	var unlock int32
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&unlock)))
	if errno != 0 {
		return errno
	}
	return nil
}

// createUnixPTY creates a PTY on Unix-like systems (Linux, macOS)
func (pty *PTY) createUnixPTY(options *models.TerminalOptions) error {
	// Open master side of PTY
	master, slave, err := pty.Open()
	if err != nil {
		return fmt.Errorf("failed to open PTY: %w", err)
	}

	pty.master = master
	pty.slave = slave

	// Set terminal size
	if err := pty.SetSize(options.Rows, options.Cols); err != nil {
		return fmt.Errorf("failed to set terminal size: %w", err)
	}

	// Prepare command
	shell := options.Shell
	if shell == "" {
		shell = "/bin/bash"
	}

	pty.cmd = exec.Command(shell)
	pty.cmd.Stdin = slave
	pty.cmd.Stdout = slave
	pty.cmd.Stderr = slave

	// Set working directory
	if options.CWD != "" {
		pty.cmd.Dir = options.CWD
	}

	// Set environment variables
	if len(options.Env) > 0 {
		env := os.Environ()
		for k, v := range options.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		pty.cmd.Env = env
	}

	// Make the slave the controlling terminal
	pty.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0,
	}

	// Start the command
	if err := pty.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Close slave end (it's now owned by the child process)
	slave.Close()
	pty.slave = nil

	return nil
}

// createWindowsPTY creates a PTY on Windows using ConPTY
func (pty *PTY) createWindowsPTY(options *models.TerminalOptions) error {
	// Windows PTY implementation using ConPTY API
	// This is a simplified version - in production, you'd use proper ConPTY bindings

	shell := options.Shell
	if shell == "" {
		shell = "cmd.exe"
	}

	pty.cmd = exec.Command(shell)

	// For Windows, we'll use pipes instead of PTY for now
	// In a full implementation, you'd use Windows ConPTY API
	stdinPipe, err := pty.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdoutPipe, err := pty.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// We don't need stderr as a separate pipe for Windows PTY emulation
	// The stderr will be merged with stdout
	pty.cmd.Stderr = pty.cmd.Stdout

	// Create a pipe pair to act as our master/slave
	// For Windows, we'll create os.Pipe() for IO
	masterReader, masterWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create master pipe: %w", err)
	}

	// Store the master file for reading
	pty.master = masterReader

	// Start goroutine to copy stdout to master writer
	go func() {
		defer masterWriter.Close()
		io.Copy(masterWriter, stdoutPipe)
	}()

	// Store stdin pipe reference in a closure for the Write method
	// We'll use the stdinPipe directly in the command
	_ = stdinPipe // Used by the command

	// Set working directory
	if options.CWD != "" {
		pty.cmd.Dir = options.CWD
	}

	// Set environment variables
	if len(options.Env) > 0 {
		env := os.Environ()
		for k, v := range options.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		pty.cmd.Env = env
	}

	// Start the command
	if err := pty.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	return nil
}

// Write writes data to the PTY
func (pty *PTY) Write(data []byte) error {
	if !pty.active || pty.master == nil {
		return fmt.Errorf("PTY is not active")
	}

	pty.lastActivity = time.Now()

	_, err := pty.master.Write(data)
	if err != nil {
		// Check if the process is still running
		if pty.cmd.Process != nil && pty.cmd.ProcessState != nil {
			if pty.cmd.ProcessState.Exited() {
				pty.active = false
				return fmt.Errorf("terminal process has exited")
			}
		}
		return fmt.Errorf("failed to write to PTY: %w", err)
	}

	return nil
}

// Read reads data from the PTY
func (pty *PTY) Read(buf []byte) (int, error) {
	if !pty.active || pty.master == nil {
		return 0, fmt.Errorf("PTY is not active")
	}

	pty.lastActivity = time.Now()

	n, err := pty.master.Read(buf)
	if err != nil {
		if err == io.EOF {
			pty.active = false
			return n, io.EOF
		}
		return n, fmt.Errorf("failed to read from PTY: %w", err)
	}

	return n, nil
}

// Resize changes the terminal size
func (pty *PTY) SetSize(rows, cols uint16) error {
	if pty.master == nil {
		return fmt.Errorf("PTY master is not available")
	}

	pty.size.Rows = rows
	pty.size.Cols = cols

	switch runtime.GOOS {
	case "linux", "darwin":
		return pty.resizeUnix(rows, cols)
	case "windows":
		// Windows resize would use ConPTY API
		return nil
	default:
		return fmt.Errorf("resize not supported on platform: %s", runtime.GOOS)
	}
}

// resizeUnix resizes the terminal on Unix-like systems
func (pty *PTY) resizeUnix(rows, cols uint16) error {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{
		Row:    rows,
		Col:    cols,
		Xpixel: 0,
		Ypixel: 0,
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		pty.master.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(ws)),
	)

	if errno != 0 {
		return fmt.Errorf("failed to set window size: %v", errno)
	}

	return nil
}

// Close closes the PTY and terminates the associated process
func (pty *PTY) Close() error {
	pty.active = false

	var errs []error

	// Close master file
	if pty.master != nil {
		if err := pty.master.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close master: %w", err))
		}
		pty.master = nil
	}

	// Close slave file if still open
	if pty.slave != nil {
		if err := pty.slave.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close slave: %w", err))
		}
		pty.slave = nil
	}

	// Terminate the process
	if pty.cmd != nil && pty.cmd.Process != nil {
		if err := pty.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			errs = append(errs, fmt.Errorf("failed to send SIGTERM: %w", err))
		}

		// Wait for process to exit or force kill after timeout
		done := make(chan error, 1)
		go func() {
			err := pty.cmd.Wait()
			done <- err
		}()

		select {
		case <-done:
			// Process exited normally
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't exit
			if err := pty.cmd.Process.Kill(); err != nil {
				errs = append(errs, fmt.Errorf("failed to kill process: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing PTY: %v", errs)
	}

	return nil
}

// IsActive returns whether the PTY is active
func (pty *PTY) IsActive() bool {
	if !pty.active {
		return false
	}

	// Check if process is still running
	if pty.cmd != nil && pty.cmd.Process != nil && pty.cmd.ProcessState != nil {
		if pty.cmd.ProcessState.Exited() {
			pty.active = false
			return false
		}
	}

	return true
}

// GetSize returns the current terminal size
func (pty *PTY) GetSize() *models.TerminalSize {
	if pty.size == nil {
		return &models.TerminalSize{Rows: 24, Cols: 80}
	}
	return pty.size
}

// GetProcess returns the process information
func (pty *PTY) GetProcess() *os.Process {
	if pty.cmd != nil {
		return pty.cmd.Process
	}
	return nil
}

// GetLastActivity returns the last activity time
func (pty *PTY) GetLastActivity() time.Time {
	return pty.lastActivity
}

// RemovePTY removes a PTY from the manager
func (pm *PTYManager) RemovePTY(sessionID string) {
	if pty, exists := pm.ptyMap[sessionID]; exists {
		pty.Close()
		delete(pm.ptyMap, sessionID)
	}
}

// GetPTY retrieves a PTY by session ID
func (pm *PTYManager) GetPTY(sessionID string) *PTY {
	return pm.ptyMap[sessionID]
}

// GetAllPTYs returns all active PTYs
func (pm *PTYManager) GetAllPTYs() map[string]*PTY {
	result := make(map[string]*PTY)
	for id, pty := range pm.ptyMap {
		if pty.IsActive() {
			result[id] = pty
		} else {
			// Clean up inactive PTYs
			delete(pm.ptyMap, id)
		}
	}
	return result
}

// Cleanup closes all PTYs
func (pm *PTYManager) Cleanup() {
	for sessionID, pty := range pm.ptyMap {
		pty.Close()
		delete(pm.ptyMap, sessionID)
	}
}

// Platform-specific PTY operations
var (
	// Import platform-specific packages
	_ = terminal.IsTerminal
)
