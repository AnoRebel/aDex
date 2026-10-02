//go:build darwin

package models

import (
	"os"
	"syscall"
	"time"
)

// statTimes returns the access and change times recorded by the filesystem.
//
// macOS spells the timespec fields Atimespec/Ctimespec where Linux uses
// Atim/Ctim, which is why this is a separate file from the Linux version
// rather than a shared `!windows` one.
func statTimes(info os.FileInfo) (accessed, changed time.Time, ok bool) {
	sys, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	return time.Unix(sys.Atimespec.Sec, sys.Atimespec.Nsec),
		time.Unix(sys.Ctimespec.Sec, sys.Ctimespec.Nsec),
		true
}

// statOwner returns the numeric owner and group of the file.
func statOwner(info os.FileInfo) (uid, gid int, ok bool) {
	sys, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false
	}
	return int(sys.Uid), int(sys.Gid), true
}
