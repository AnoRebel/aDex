package system

import (
	"fmt"
	"os"
	"strconv"
)

// ProcessSignal is a typed signal name accepted by SignalProcess. We keep
// this string-based (not syscall.Signal) so the Wails binding can shuttle
// it through JSON without platform-specific numeric values leaking to
// the frontend.
type ProcessSignal string

const (
	SignalTerm ProcessSignal = "SIGTERM"
	SignalKill ProcessSignal = "SIGKILL"
	SignalStop ProcessSignal = "SIGSTOP"
	SignalCont ProcessSignal = "SIGCONT"
	SignalHup  ProcessSignal = "SIGHUP"
	SignalInt  ProcessSignal = "SIGINT"
	SignalUsr1 ProcessSignal = "SIGUSR1"
	SignalUsr2 ProcessSignal = "SIGUSR2"
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

	// Signal delivery is platform-specific — see proc_unix.go and
	// proc_windows.go. The constants differ enough that a runtime check is
	// not sufficient: SIGSTOP and friends do not exist at all on Windows, so
	// referencing them would fail the build rather than the call.
	return sendSignal(p, sig)
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
