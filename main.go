package main

import (
	"embed"
	"log"

	"aDex-UI/backend/services/coordinator"
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
		Name:        "aDex-UI",
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
		Title:            "aDex-UI",
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
		},
		// v2 set WebviewIsTransparent/WindowIsTranslucent/DisableWindowIcon
		// to false on Windows and Linux. Those are the v3 defaults
		// (Windows translucency is now expressed as BackdropType, whose
		// zero value is opaque), so no explicit block is needed. Linux
		// keeps an explicit entry only to pin GPU policy — leaving
		// options.Linux nil makes Wails default WebviewGpuPolicy to
		// Never (wailsapp/wails#2977), which would cost us the WebGL
		// renderer the globe and terminal rely on.
		Linux: application.LinuxWindow{
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    application.WebviewGpuPolicyOnDemand,
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
