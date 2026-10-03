# wails-v3-migration

- OS:           linux 7.2 (Hyprland, Wayland)
- Theme:        tron
- Wails:        v3 (beta)
- Date:         2026-08-27
- App SHA:      bc59f03
- Verifier:     manual — screenshots of the running app

## What this proves

After moving from Wails v2 to v3 the app boots, renders the main interface and runs a terminal session through the v3 bindings. service-consolidation-verdicts.md records which duplicated backend services were kept.

## How to reproduce

1. `task build` and run `bin/adex`.
2. Let the boot sequence finish, then open a terminal tab.

## Observed

See the screenshots in this directory.
