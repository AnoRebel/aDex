package main

import (
	"context"
	"embed"
	"log"

	"aDex-UI/backend/services/coordinator"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/.output/public folder will be embedded into the binary and
// made available to the frontend.
//
//go:embed all:frontend/.output/public
var assets embed.FS

// App represents the Wails application with all services
type App struct {
	ctx         context.Context
	coordinator *coordinator.ServiceCoordinator
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		coordinator: coordinator.NewServiceCoordinator(),
	}
}

// OnStartup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Initialize service coordinator
	if err := a.coordinator.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize coordinator:", err)
	}

	// Start monitoring
	if err := a.coordinator.StartMonitoring(ctx); err != nil {
		log.Println("Warning: Failed to start monitoring:", err)
	}
}

// OnShutdown is called when the app is shutting down
func (a *App) OnShutdown(ctx context.Context) {
	if a.coordinator != nil {
		a.coordinator.Shutdown()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return "Hello " + name + "!"
}

// main function serves as the application's entry point.
func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:      "aDex-UI",
		Width:      1920,
		Height:     1080,
		Fullscreen: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnShutdown:       app.OnShutdown,
		Bind: []interface{}{
			app,
			app.coordinator, // This exposes all coordinator methods to frontend
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
			},
			Appearance:           mac.DefaultAppearance,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
