//go:build linux

package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// resolvePidCWD reads /proc/<pid>/cwd, the symlink to the live working dir.
// Matches edex-ui's Linux path: fs.readlink(`/proc/${pid}/cwd`).
func resolvePidCWD(pid int) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("invalid pid %d", pid)
	}
	link := filepath.Join("/proc", strconv.Itoa(pid), "cwd")
	dest, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	return dest, nil
}
