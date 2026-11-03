package terminal

import (
	"slices"
	"testing"
)

func TestShellEnvStripsAppImageLaunch(t *testing.T) {
	const d = "/tmp/.mount_aDexX"
	got := shellEnv([]string{
		"APPDIR=" + d, "APPIMAGE=/home/u/aDex.AppImage", "ARGV0=aDex.AppImage", "OWD=/home/u",
		"GDK_BACKEND=x11", "GTK_THEME=Adwaita:dark",
		"PATH=" + d + "/usr/bin:/usr/local/bin:/usr/bin",
		"LD_LIBRARY_PATH=" + d + "/usr/lib:" + d + "/usr/lib64",
		"XDG_DATA_DIRS=" + d + "/usr/share:/usr/share:/usr/local/share",
		"GTK_PATH=" + d + "//usr/lib/gtk-4.0",
		"HOME=/home/u", "LANG=en_GB.UTF-8",
	})
	want := []string{
		"PATH=/usr/local/bin:/usr/bin",
		"XDG_DATA_DIRS=/usr/share:/usr/local/share",
		"HOME=/home/u", "LANG=en_GB.UTF-8",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("shellEnv:\n got %q\nwant %q", got, want)
	}
}

func TestShellEnvOutsideAppImageUnchanged(t *testing.T) {
	env := []string{"PATH=/usr/bin", "GDK_BACKEND=wayland", "HOME=/home/u"}
	if got := shellEnv(env); !slices.Equal(got, env) {
		t.Fatalf("shellEnv changed a normal environment: %q", got)
	}
}
