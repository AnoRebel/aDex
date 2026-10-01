package system

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"syscall"
)

// ProcessSignal is a typed signal name accepted by SignalProcess. We keep
// this string-based (not syscall.Signal) so the Wails binding can shuttle
// it through JSON without platform-specific numeric values leaking to
// the frontend.
type ProcessSignal string

const (
	SignalTerm  ProcessSignal = "SIGTERM"
	SignalKill  ProcessSignal = "SIGKILL"
	SignalStop  ProcessSignal = "SIGSTOP"
	SignalCont  ProcessSignal = "SIGCONT"
	SignalHup   ProcessSignal = "SIGHUP"
	SignalInt   ProcessSignal = "SIGINT"
	SignalUsr1  ProcessSignal = "SIGUSR1"
	SignalUsr2  ProcessSignal = "SIGUSR2"
)

// SignalProcess sends the given signal to the process at pid.
// Implementation notes:
//
//   - Linux/macOS: uses os.Process.Signal with the matching syscall.Signal.
//   - Windows: only SIGKILL (mapped to TerminateProcess via os.Process.Kill)
//     and SIGINT (mapped to GenerateConsoleCtrlEvent via /usr/local equivalents)
//     are honored. Stop/Cont/Hup/Usr1/Usr2 return an error so the UI can
//     disable the menu entries on Windows.
//
// Source-of-truth for the signal table is internal to this file so the
// coordinator's binding surface stays stable across platforms.
func (s *Service) SignalProcess(pid int, sig ProcessSignal) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if runtime.GOOS == "windows" {
		switch sig {
		case SignalKill, SignalTerm:
			// Both map to a hard kill on Windows; there is no graceful
			// SIGTERM equivalent for arbitrary processes.
			return p.Kill()
		default:
			return fmt.Errorf("signal %s not supported on windows", sig)
		}
	}

	var ssig syscall.Signal
	switch sig {
	case SignalTerm:
		ssig = syscall.SIGTERM
	case SignalKill:
		ssig = syscall.SIGKILL
	case SignalStop:
		ssig = syscall.SIGSTOP
	case SignalCont:
		ssig = syscall.SIGCONT
	case SignalHup:
		ssig = syscall.SIGHUP
	case SignalInt:
		ssig = syscall.SIGINT
	case SignalUsr1:
		ssig = syscall.SIGUSR1
	case SignalUsr2:
		ssig = syscall.SIGUSR2
	default:
		return fmt.Errorf("unknown signal %s", sig)
	}
	return p.Signal(ssig)
}

// GetProcessInfo returns a richer view of a single process suitable for
// the UI's "Process Info" modal. The data comes from gopsutil — most of
// the heavy reads happen in /proc on Linux so this is a few-microsecond
// call per pid.
type ProcessInfoExtended struct {
	PID           int      `json:"pid"`
	PPID          int      `json:"ppid"`
	Name          string   `json:"name"`
	Command       string   `json:"command"`
	Executable    string   `json:"executable"`
	CWD           string   `json:"cwd"`
	User          string   `json:"user"`
	Status        string   `json:"status"`
	CPUPercent    float64  `json:"cpuPercent"`
	MemoryPercent float64  `json:"memoryPercent"`
	MemoryRSS     uint64   `json:"memoryRss"`
	NumThreads    int      `json:"numThreads"`
	NumFDs        int      `json:"numFds"`
	CreatedAt     int64    `json:"createdAt"`
	Environment   []string `json:"environment"`
}

// renicePid is platform-conditional and lives in proc_unix.go / proc_other.go.

// formatPid is used by error messages where we want the literal pid as a
// string; pulled out so signal helpers stay branchless.
func formatPid(pid int) string { return strconv.Itoa(pid) }
