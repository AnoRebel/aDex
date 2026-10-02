//go:build !windows

package system

import (
	"fmt"
	"os"
	"syscall"
)

// sendSignal delivers a POSIX signal to the process.
//
// The signal constants live in a build-tagged file because most of them —
// SIGSTOP, SIGCONT, SIGUSR1, SIGUSR2 — simply do not exist in Go's syscall
// package on Windows, so referencing them unconditionally fails the Windows
// build at compile time rather than at run time.
func sendSignal(p *os.Process, sig ProcessSignal) error {
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
