package updater

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CanSelfUpdate reports whether the running copy can replace itself in place,
// and when it can't, why. Installs that can't (distro packages, the raw Linux
// binary, read-only locations) are still told about updates but only get a
// link to the release page.
func CanSelfUpdate() (bool, string) {
	exe, err := os.Executable()
	if err != nil {
		return false, "cannot locate the running executable"
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return Eligibility(runtime.GOOS, exe, os.Getenv("APPIMAGE"))
}

// Eligibility is CanSelfUpdate with its inputs passed in: the OS, the
// executable path and the value of $APPIMAGE.
//
// The updater stages the download in a directory beside the file it replaces
// and then renames it into place, so the deciding question on every platform
// is whether that directory is writable.
func Eligibility(goos, exe, appImage string) (bool, string) {
	var target string
	switch goos {
	case "linux":
		if appImage == "" {
			return false, "not running from an AppImage"
		}
		target = appImage
	case "darwin":
		bundle, ok := appBundle(exe)
		if !ok {
			return false, "not running from an .app bundle"
		}
		target = bundle
	case "windows":
		target = exe
	default:
		return false, "in-app updates are not supported on " + goos
	}
	dir := filepath.Dir(target)
	if !writable(dir) {
		return false, "install location is not writable: " + dir
	}
	return true, ""
}

// appBundle returns the outermost .app directory exe lives in.
func appBundle(exe string) (string, bool) {
	clean := filepath.Clean(exe)
	i := strings.Index(clean, ".app"+string(filepath.Separator))
	if i < 0 {
		return "", false
	}
	return clean[:i+len(".app")], true
}

// writable probes dir the same way the updater will use it: by creating and
// removing a file there. Permission bits alone miss ACLs, read-only mounts
// and sandboxing.
func writable(dir string) bool {
	f, err := os.CreateTemp(dir, ".adex-write-probe-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}
