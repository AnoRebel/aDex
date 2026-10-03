# settings-ux

- OS:           linux 7.2 (Hyprland, Wayland)
- Theme:        tron, apollo
- Wails:        v3 (beta)
- Date:         2026-08-28
- App SHA:      bc59f03
- Verifier:     manual — screenshots of the running app

## What this proves

The settings modal and side columns render with themed form controls (inputs, selects, toggles) rather than browser defaults, the theme panel applies changes live, and the side columns fit without clipping at small window sizes (800x600, 1366x768).

## How to reproduce

1. `task build` and run `bin/adex`.
2. Open Settings (Ctrl+,), switch themes and inspect each panel; resize the window to 800x600.

## Observed

See the screenshots in this directory.
