package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"aDex-UI/internal/colorscheme"
	"aDex-UI/internal/events"
	"aDex-UI/internal/font"
	"aDex-UI/internal/logger"
)

// App represents the Wails application context
type App struct {
	application    *application.App
	eventBus       *events.EventBus
	logger         *logger.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	colorSchemeSvc *colorscheme.Service
	fontSvc        *font.Service
}

// NewApp creates a new application context
func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Level:  "info",
		Format: "text",
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Initialize event bus
	eventBus := events.GetEventBus()

	return &App{
		ctx:      ctx,
		cancel:   cancel,
		logger:   appLogger,
		eventBus: eventBus,
	}
}

// Initialize sets up the application with the given Wails app instance
func (a *App) Initialize(app *application.App) error {
	a.application = app
	a.logger.Info("Application context initialized")

	// Initialize services here
	return a.initializeServices()
}

// GetContext returns the application context
func (a *App) GetContext() context.Context {
	return a.ctx
}

// GetEventBus returns the application event bus
func (a *App) GetEventBus() *events.EventBus {
	return a.eventBus
}

// GetLogger returns the application logger
func (a *App) GetLogger() *logger.Logger {
	return a.logger
}

// GetColorSchemeService returns the color scheme service
func (a *App) GetColorSchemeService() *colorscheme.Service {
	return a.colorSchemeSvc
}

// GetFontService returns the font service
func (a *App) GetFontService() *font.Service {
	return a.fontSvc
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	a.logger.Info("Shutting down application...")

	// Shutdown color scheme service
	if a.colorSchemeSvc != nil {
		if err := a.colorSchemeSvc.Shutdown(a.ctx); err != nil {
			a.logger.Error("Failed to shutdown color scheme service", err)
		}
	}

	// Shutdown font service
	if a.fontSvc != nil {
		if err := a.fontSvc.Shutdown(a.ctx); err != nil {
			a.logger.Error("Failed to shutdown font service", err)
		}
	}

	// Cancel context to stop all background operations
	a.cancel()

	// Wait for all goroutines to finish
	a.wg.Wait()

	a.logger.Info("Application shutdown complete")
	return nil
}

// initializeServices initializes all core services
func (a *App) initializeServices() error {
	// Initialize color scheme service
	dataDir := a.application.GetContext().GetDataDirectory()
	configPath := filepath.Join(dataDir, "colorscheme-config.json")
	schemesDir := filepath.Join(dataDir, "colorschemes")

	a.colorSchemeSvc = colorscheme.NewService(configPath, schemesDir)
	a.colorSchemeSvc.SetEventBus(a.eventBus)

	if err := a.colorSchemeSvc.Initialize(a.ctx); err != nil {
		a.logger.Error("Failed to initialize color scheme service", err)
		return err
	}

	// Initialize font service
	fontConfigPath := filepath.Join(dataDir, "font-config.json")

	a.fontSvc = font.NewService(fontConfigPath)
	a.fontSvc.SetEventBus(a.eventBus)

	if err := a.fontSvc.Initialize(a.ctx); err != nil {
		a.logger.Error("Failed to initialize font service", err)
		return err
	}

	a.logger.Info("Core services initialized")
	return nil
}