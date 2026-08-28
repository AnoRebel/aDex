package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"aDex-UI/internal/appdir"
	"aDex-UI/internal/services/coordinator"
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
		// GPU acceleration is OFF by default on Linux.
		//
		// Three WebKitWebProcess coredumps here all point into the Mesa stack:
		// two SIGABRT heap corruptions inside libgallium mid-session, and a
		// SIGSEGV inside dri_gbm/libgbm during process exit. The renderer dying
		// takes the window with it, and none of it is reachable from Go — which
		// is why the Go crash log stayed empty throughout.
		//
		// Note that WebviewGpuPolicyOnDemand is NOT a middle setting on this
		// platform: WebKitGTK 6.0 removed ON_DEMAND, and Wails maps the value
		// to ALWAYS, so the previous setting was forcing acceleration on
		// permanently rather than leaving it to the engine.
		//
		// Software rendering costs some compositing performance but keeps the
		// application alive. ADEX_GPU=1 opts back in for anyone on a driver
		// where this is not a problem.
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
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
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
// Defaults to Never because the Mesa driver on this platform has been observed
// crashing the web process both mid-session and at exit. Set ADEX_GPU=1 to
// re-enable it.
func linuxGpuPolicy() application.WebviewGpuPolicy {
	if os.Getenv("ADEX_GPU") == "1" {
		return application.WebviewGpuPolicyAlways
	}
	return application.WebviewGpuPolicyNever
}
