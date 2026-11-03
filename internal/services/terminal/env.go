package terminal

import "strings"

// appImageVars are set by an AppImage's AppRun and GTK hook for the app
// itself. GDK_BACKEND and GTK_THEME are forced there (x11, Adwaita), so they
// carry no choice of the user's.
var appImageVars = map[string]bool{
	"APPDIR": true, "APPIMAGE": true, "ARGV0": true, "OWD": true,
	"GDK_BACKEND": true, "GTK_THEME": true,
}

// shellEnv returns env without what an AppImage launch added. Shells must
// see the user's environment, not one whose PATH and LD_LIBRARY_PATH point
// into the AppImage's mount: commands would load the bundled libraries, and
// tools such as mise misread ARGV0. Outside an AppImage env is returned as is.
func shellEnv(env []string) []string {
	var appDir string
	for _, kv := range env {
		if v, ok := strings.CutPrefix(kv, "APPDIR="); ok {
			appDir = v
		}
	}
	if appDir == "" {
		return env
	}
	out := make([]string, 0, len(env))
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		if appImageVars[k] {
			continue
		}
		if strings.Contains(v, appDir) {
			kept := []string{}
			for _, part := range strings.Split(v, ":") {
				if !strings.HasPrefix(part, appDir) {
					kept = append(kept, part)
				}
			}
			if v = strings.Join(kept, ":"); v == "" {
				continue
			}
		}
		out = append(out, k+"="+v)
	}
	return out
}
