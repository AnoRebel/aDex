//go:build windows

package system

import (
	"fmt"
	"os"
)

// sendSignal approximates POSIX signalling on Windows.
//
// Windows has no signal delivery for arbitrary processes: kill and terminate
// both become TerminateProcess, and there is no equivalent of SIGSTOP,
// SIGCONT or the user-defined signals. Those are reported as unsupported
// rather than silently doing nothing, so the UI can say so.
func sendSignal(p *os.Process, sig ProcessSignal) error {
	switch sig {
	case SignalKill, SignalTerm:
		// Both map to a hard kill; there is no graceful SIGTERM equivalent.
		return p.Kill()
	default:
		return fmt.Errorf("signal %s not supported on windows", sig)
	}
}
