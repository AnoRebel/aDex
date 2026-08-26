//go:build !linux && !darwin

package terminal

import "fmt"

// resolvePidCWD is unsupported on this platform; callers fall back to the
// recorded WorkingDir, matching upstream eDEX-UI behavior. On Windows we
// could integrate the existing internal/services/terminal/WindowsCWDTracker
// in a follow-up change.
func resolvePidCWD(_ int) (string, error) {
	return "", fmt.Errorf("live cwd tracking not supported on this platform")
}
