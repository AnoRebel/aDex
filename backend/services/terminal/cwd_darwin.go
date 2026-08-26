//go:build darwin

package terminal

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// resolvePidCWD shells out to lsof to read the live working directory. Mirrors
// edex-ui's Darwin path:
//
//	lsof -a -d cwd -p <pid> | tail -1 | awk '{ for (i=9; i<=NF; i++) printf "%s ", $i }'
//
// We do the field extraction in Go rather than chaining awk so we don't depend
// on a shell. The lsof output line for `-d cwd` looks like:
//
//	bash 12345 user  cwd  DIR  1,15   544  12345 /Users/user/code/dir with spaces
//
// where field 9 onward is the path (which may contain spaces).
func resolvePidCWD(pid int) (string, error) {
	if pid <= 0 {
		return "", fmt.Errorf("invalid pid %d", pid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	out, err := exec.CommandContext(ctx, "lsof", "-a", "-d", "cwd", "-p", strconv.Itoa(pid), "-Fn").Output()
	if err != nil {
		return "", fmt.Errorf("lsof: %w", err)
	}
	// `-Fn` produces lines like:
	//   p12345
	//   n/Users/user/code
	// The line beginning with 'n' is the path.
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.HasPrefix(line, "n") {
			return strings.TrimSpace(line[1:]), nil
		}
	}
	return "", fmt.Errorf("lsof returned no path for pid %d", pid)
}
