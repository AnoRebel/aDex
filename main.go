package main

import (
	"context"
	"embed"
	_ "embed"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"aDex-UI/backend/services/coordinator"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/.output/public
var assets embed.FS

// main function serves as the application's entry point. It initializes the service coordinator,
// creates a window, and starts the application with proper service management.
func main() {
	// Create application context for service initialization
	ctx := context.Background()

	// Create and initialize the service coordinator
	serviceCoordinator := coordinator.NewServiceCoordinator()
	if err := serviceCoordinator.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize service coordinator: %v", err)
	}

	// Start monitoring for system and audio services
	if err := serviceCoordinator.StartMonitoring(ctx); err != nil {
		log.Printf("Warning: Failed to start monitoring: %v", err)
	}

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Services' exposes the service coordinator to the frontend.
	// 'Mac' options tailor the application when running on macOS.
	app := application.New(application.Options{
		Name:        "aDex-UI",
		Description: "A modern science fiction desktop environment terminal application",
		Services: []application.Service{
			application.NewService(serviceCoordinator),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		})

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "aDex-UI",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Create a goroutine that emits a time-based event every second through the service coordinator's event bus.
	// The frontend can listen to this event and update the UI accordingly.
	go func() {
		eventBus := serviceCoordinator.GetEventBus()
		for {
			now := time.Now().Format(time.RFC1123)
			// Publish time event through the service coordinator's event bus
			eventBus.Publish(ctx, "time.updated", map[string]interface{}{
				"time": now,
				"unix": time.Now().Unix(),
			}, "system")
			time.Sleep(time.Second)
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
