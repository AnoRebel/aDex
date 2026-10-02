//go:build windows

package models

import (
	"os"
	"time"
)

// statTimes reports no extended timestamps on Windows.
//
// Windows does expose creation and access times through
// syscall.Win32FileAttributeData, but they do not map onto the POSIX
// atime/ctime pair this type models — there is no change-time equivalent at
// all. Rather than invent one, report nothing and let the caller fall back to
// ModTime, which is accurate on every platform.
func statTimes(info os.FileInfo) (accessed, changed time.Time, ok bool) {
	return time.Time{}, time.Time{}, false
}

// statOwner reports no numeric owner on Windows.
//
// Windows identifies owners by SID, not by the small integer uid/gid pair
// this type carries, so there is nothing meaningful to return here.
func statOwner(info os.FileInfo) (uid, gid int, ok bool) {
	return 0, 0, false
}
