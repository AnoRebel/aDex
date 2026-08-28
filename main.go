package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"aDex-UI/internal/appdir"
	"aDex-UI/internal/services/coordinator"
	"aDex-UI/internal/services/settings"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// The Nuxt static bundle is emitted to frontend/dist (see the
// `nitro.output.publicDir` setting in frontend/nuxt.config.ts), which is the
// location the Wails v3 build tooling assumes by default.
//
// NOTE: frontend/dist is gitignored, so a clean checkout must build the
// frontend (`wails3 task build`, or `bun run generate` in frontend/) before
// `go build` will succeed — embed refuses to compile against an empty
// directory.
//
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Native crashes (heap corruption inside the webview or an audio backend)
	// abort the process without a Go panic, so nothing normally reaches the
	// log. Route glibc's diagnostics and Go's own crash output to a file the
	// user can hand over after a crash.
	//
	// MALLOC_CHECK_=2 makes glibc abort AT the offending free rather than
	// later when the heap is already inconsistent, which is the difference
	// between a usable report and "corrupted double-linked list" with no
	// context. It is cheap enough to leave on.
	if os.Getenv("MALLOC_CHECK_") == "" {
		_ = os.Setenv("MALLOC_CHECK_", "2")
	}
	if f, err := os.OpenFile(crashLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
		// Go writes panics and fatal runtime errors to fd 2.
		_ = debugSetCrashOutput(f)
	}

	// The coordinator owns every backend service. Registering it as a Wails v3
	// service hands its lifecycle to Wails: ServiceStartup runs during
	// app.Run() before any window is shown — returning an error there aborts
	// startup — and ServiceShutdown runs during teardown. Services start in
	// registration order and shut down in reverse, which is what gives the
	// coordinator its dependency ordering for free.
	//
	// The terminal service is constructed and torn down by the coordinator
	// rather than registered separately, so it stays inside that ordering.
	app := application.New(application.Options{
		Name:        "aDex",
		Description: "A modern science fiction desktop environment terminal application",
		Services: []application.Service{
			application.NewService(coordinator.NewServiceCoordinator()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Window sizing strategy:
	//   - Width/Height are the fallback size, used when the window manager
	//     cannot honour StartState.
	//   - StartState: WindowStateMaximised opens the window covering the
	//     active monitor's work area — the eDEX-UI signature look. This
	//     replaces the v2 workaround of calling WindowMaximise from
	//     OnDomReady, which existed because v2's start-state hint raced the
	//     compositor on Wayland/X11.
	//   - MinWidth/MinHeight are a hard floor enforced by the window manager,
	//     not a styling hint: below them the window simply cannot be resized,
	//     however well the CSS reflows. v2 set 900x600, which stopped the app
	//     fitting on smaller or scaled displays even though the stylesheets
	//     carry breakpoints down to 480px. They are set to 640x480 here so the
	//     OS-level floor stops fighting the responsive layout — the reflow
	//     rules, not the window manager, decide how narrow is usable.
	//
	// v2 additionally needed explicit MaxWidth/MaxHeight of 16384 to defeat a
	// GTK quirk that clamped the window to the current monitor's geometry when
	// the max dimensions were left at zero. v3 does not reimpose that clamp,
	// so the workaround is deliberately not carried over — leaving them unset
	// means "no maximum", which is the intended behaviour on multi-monitor
	// setups.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "aDex",
		Width:            1600,
		Height:           1000,
		MinWidth:         640,
		MinHeight:        480,
		StartState:       application.WindowStateMaximised,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		Mac: application.MacWindow{
			TitleBar:                application.MacTitleBarDefault,
			InvisibleTitleBarHeight: 0,
			// NOTE on splash audio: the webview requires a user gesture before
			// it will play sound, so boot-splash cues only become audible once
			// the user first interacts. Wails v3 exposes
			// MacWebviewPreferences.EnableAutoplayWithoutUserAction to lift
			// this, but (a) it is macOS/iOS only — there is no Linux
			// equivalent, since Wails never sets WebKitGTK's
			// media-playback-requires-user-gesture — and (b) its type comes
			// from wails/v3/internal/optional, which application code cannot
			// import. So it cannot be set from here on any platform.
		},
		// GPU acceleration is ON by default, and configurable.
		//
		// Three WebKitWebProcess coredumps here point into the Mesa stack (two
		// SIGABRT heap corruptions in libgallium, one SIGSEGV in dri_gbm during
		// exit). Turning acceleration off avoids that path — but it makes the
		// interface noticeably laggy, which is too high a price to impose by
		// default for a driver bug that may not affect a given machine.
		//
		// So it stays on, and Settings -> Advanced exposes it with a warning.
		// Users hitting the crash can turn it off and trade smoothness for
		// stability.
		//
		// Note WebviewGpuPolicyOnDemand is NOT a middle setting here: WebKitGTK
		// 6.0 removed ON_DEMAND and Wails maps it to ALWAYS.
		Linux: application.LinuxWindow{
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linuxGpuPolicy(),
		},
	})

	// Run blocks until the application exits. Quit is initiated from the
	// frontend via the runtime's Application.Quit(), which triggers the same
	// teardown path: ShouldQuit, then each service's ServiceShutdown in
	// reverse registration order. The coordinator's Shutdown is idempotent
	// (guarded by isStarted), so a repeated call is harmless.
	// Record how the run ended.
	//
	// Three of the crashes so far left NO Go panic and no core dump for this
	// process: the WebKit renderer died and Wails unwound the parent normally,
	// so there was nothing to find afterwards. Writing the outcome here means
	// the next unexplained exit at least says whether app.Run returned an
	// error, returned cleanly, or never returned at all.
	logSessionEvent("run: start")
	err := app.Run()
	if err != nil {
		logSessionEvent("run: exited with error: " + err.Error())
		log.Fatal(err)
	}
	logSessionEvent("run: exited cleanly")
}

// logSessionEvent appends a timestamped line to the crash log. Best-effort:
// diagnostics must never themselves break the application.
func logSessionEvent(msg string) {
	f, err := os.OpenFile(crashLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(time.Now().Format(time.RFC3339) + " " + msg + "\n")
}

// crashLogPath returns the file native and runtime crash output is appended to.
// It sits beside the settings file so a user reporting a crash has one obvious
// place to look.
func crashLogPath() string {
	return filepath.Join(appdir.Config(), "crash.log")
}

// debugSetCrashOutput wraps debug.SetCrashOutput so the call site stays
// readable; it duplicates Go's crash output to the given file.
func debugSetCrashOutput(f *os.File) error {
	return debug.SetCrashOutput(f, debug.CrashOptions{})
}

// linuxGpuPolicy decides whether the webview may use GPU acceleration.
//
// On by default: software rendering is stable but visibly laggy. The setting
// is read from the UI settings file rather than a flag, so the choice persists
// and is changeable from Settings -> Advanced. ADEX_GPU overrides it for a
// single run (1 = on, 0 = off), which is useful when the crash makes the app
// hard to reach.
func linuxGpuPolicy() application.WebviewGpuPolicy {
	switch os.Getenv("ADEX_GPU") {
	case "1":
		return application.WebviewGpuPolicyAlways
	case "0":
		return application.WebviewGpuPolicyNever
	}
	if settings.GPUAccelerationDisabled() {
		return application.WebviewGpuPolicyNever
	}
	return application.WebviewGpuPolicyAlways
}
