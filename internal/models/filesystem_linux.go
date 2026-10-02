//go:build linux

package models

import (
	"os"
	"syscall"
	"time"
)

// statTimes returns the access and change times recorded by the filesystem.
//
// The stat struct's field names differ between Unix flavours — Linux spells
// them Atim/Ctim, macOS and the BSDs Atimespec/Ctimespec — so each gets its
// own file rather than sharing one `!windows` build tag. Windows has no such
// struct in Go's syscall package at all.
//
// `ok` is false when the FileInfo carries no stat data, in which case the
// caller falls back to ModTime.
func statTimes(info os.FileInfo) (accessed, changed time.Time, ok bool) {
	sys, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	return time.Unix(sys.Atim.Sec, sys.Atim.Nsec),
		time.Unix(sys.Ctim.Sec, sys.Ctim.Nsec),
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
