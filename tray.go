package main

import (
	"runtime"

	"aDex-UI/internal/services/security"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// setupSystemTray installs the tray icon, its click behaviour and its menu.
//
// Behaviour, as asked for:
//   - Left click toggles the window between shown and hidden.
//   - Right click opens a menu of quick actions.
//
// The window is deliberately NOT attached with SystemTray.AttachWindow. That
// helper repositions the window next to the tray icon and sizes it like a
// popover, which suits a small utility panel but not aDex: the window opens
// maximised and is a full desktop environment, so being yanked to a corner on
// every toggle would be wrong. Toggling visibility by hand keeps the window
// wherever the user left it.
func setupSystemTray(app *application.App, window *application.WebviewWindow) {
	tray := app.SystemTray.New()

	// The label must be set BEFORE Run: on Linux it becomes the
	// StatusNotifierItem's Id and Title, which default to "Wails" when left
	// empty. Panels show Title on hover and key their icon lookup off Id, so
	// leaving it unset made the tray read "Wails" and fall back to a generic
	// placeholder icon instead of using the pixmap we supply.
	tray.SetLabel("aDex")

	tray.SetTooltip("aDex — click to show or hide")
	applyTrayIcon(tray)

	// Left click: show or hide, restoring focus when showing. Without the
	// explicit Focus the window can come back behind whatever the user was
	// looking at, which reads as the click having done nothing.
	tray.OnClick(func() {
		toggleWindow(window)
	})

	menu := buildTrayMenu(app, window)
	tray.SetMenu(menu)

	// Right click: open the quick actions. SetMenu alone is enough on most
	// platforms (applySmartDefaults wires ShowMenu when a menu is present),
	// but stating it here keeps the two buttons' behaviour in one place and
	// independent of that default.
	tray.OnRightClick(func() {
		tray.OpenMenu()
	})

	tray.Run()
}

// applyTrayIcon picks the artwork that suits the platform's tray.
//
// macOS wants a template icon so the system can recolour it for light and
// dark menu bars.
//
// Linux has no light/dark distinction to offer: its setDarkModeIcon simply
// calls setIcon, so whichever is set LAST wins regardless of the desktop
// theme. Setting the light-ink glyph there made the icon invisible on a light
// panel. Most Linux panels are dark, so the light-ink variant is the single
// icon used, and SetDarkModeIcon is not called at all.
//
// Windows genuinely honours both, so it gets both.
func applyTrayIcon(tray *application.SystemTray) {
	switch runtime.GOOS {
	case "darwin":
		tray.SetTemplateIcon(trayIconDark)
	case "linux":
		tray.SetIcon(trayIconLight)
	default:
		tray.SetIcon(trayIconDark)
		tray.SetDarkModeIcon(trayIconLight)
	}
}

// toggleWindow flips the window between visible and hidden.
func toggleWindow(window *application.WebviewWindow) {
	if window == nil {
		return
	}
	if window.IsVisible() {
		window.Hide()
		return
	}
	window.Show()
	window.Focus()
}

// buildTrayMenu assembles the right-click quick actions.
//
// These are deliberately the things worth doing WITHOUT the window in front
// of you — show/hide, jump straight to a new terminal, lock the session,
// quit. Anything needing the interface already open belongs in the app, not
// here.
func buildTrayMenu(app *application.App, window *application.WebviewWindow) *application.Menu {
	menu := app.NewMenu()

	menu.Add("Show / Hide").OnClick(func(_ *application.Context) {
		toggleWindow(window)
	})

	menu.AddSeparator()

	// Emitted to the frontend rather than handled here: creating a terminal
	// is the shell's job, and index.vue already owns the tab lifecycle.
	menu.Add("New Terminal").OnClick(func(_ *application.Context) {
		if window != nil {
			window.Show()
			window.Focus()
		}
		app.Event.Emit("tray:new-terminal")
	})

	menu.Add("Settings").OnClick(func(_ *application.Context) {
		if window != nil {
			window.Show()
			window.Focus()
		}
		app.Event.Emit("tray:open-settings")
	})

	menu.AddSeparator()

	// Only offered when a passphrase is actually configured — locking with no
	// way to unlock would strand the user. This reads the same on-disk config
	// the coordinator's lock service does; the menu is built once at startup,
	// which is the point at which that state is known.
	if security.NewLockService().IsConfigured() {
		menu.Add("Lock Session").OnClick(func(_ *application.Context) {
			if window != nil {
				window.Show()
				window.Focus()
			}
			app.Event.Emit("tray:lock")
		})
		menu.AddSeparator()
	}

	menu.Add("Quit aDex").OnClick(func(_ *application.Context) {
		app.Quit()
	})

	return menu
}
