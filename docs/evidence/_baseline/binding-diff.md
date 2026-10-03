# Wails Binding Diff — Baseline

Captured: 2026-05-04
Wails: v2.12.0
Go: 1.25.0

## Binding location

The coordinator is bound from `main.go` (line 82-84):

```go
Bind: []interface{}{
    app,
    app.coordinator, // exposes ServiceCoordinator methods
}
```

Wails v2 names the binding by its struct package + name. `app.coordinator` is a `*coordinator.ServiceCoordinator`, so the runtime path is:

- `window.go.coordinator.ServiceCoordinator.<Method>` (correct)

The hand-written shim at `frontend/app/lib/wailsjs/coordinator.ts` line 99–115 looks for it at:

- `window.go.main.ServiceCoordinator.<Method>` (**WRONG** — package is `coordinator`, not `main`)

This is the **root cause** of "System monitor shows zeros". Methods resolve to `undefined` and silently fail; some defensive code returns mock values.

## Coordinator methods exposed (Go)

Source: `backend/services/coordinator/service.go`

```
Initialize, StartMonitoring, Shutdown, IsStarted

# Terminal
CreateTerminal, WriteToTerminal, ResizeTerminal, CloseTerminal, ListTerminals, GetTerminalInfo

# Filesystem
ReadDirectory

# Color Scheme
GetColorSchemes, GetColorScheme, CreateColorScheme, UpdateColorScheme, DeleteColorScheme,
SetDefaultColorScheme, GetDefaultColorScheme, GetColorSchemeConfig, UpdateColorSchemeConfig,
GetColorSchemePreview, ValidateColorScheme

# Font
GetFontConfigurations, GetFontConfiguration, CreateFontConfiguration, UpdateFontConfiguration,
DeleteFontConfiguration, GetSystemFonts, GetMonospaceFonts, ScanSystemFonts, ImportFont,
ValidateFont, GetFontMetrics, GetFontSettings, UpdateFontSettings,
GetDefaultFontConfiguration, SetDefaultFontConfiguration

# Network
GetNetworkMetrics, GetNetworkConnections, GetBandwidthData, GetNetworkStatistics,
GetNetworkAlerts, ResolveNetworkAlert, ClearNetworkAlerts, StartNetworkMonitoring,
StopNetworkMonitoring, GetNetworkConfig, UpdateNetworkConfig, ResetNetworkService,
IsNetworkMonitoring

# System metrics
GetSystemInfo, GetCPUUsage, GetMemoryUsage, GetDiskUsage, GetNetworkInfo

# Process
GetTopProcesses
```

## Frontend binding shims (TypeScript)

Source: `frontend/app/lib/wailsjs/coordinator.ts`

The shim defines the same surface but resolves through the wrong global path. Wails v2 also generates real bindings under `frontend/app/lib/wailsjs/wailsjs/go/coordinator/ServiceCoordinator.{js,d.ts}` — those are correct. The hand-written shim should either be deleted or use the generated bindings.

## Methods defined in spec but not yet on coordinator

These are added:

- `GetTerminalCWD(id string) (string, error)`
- `GetCWDStats() (CWDStats, error)`

## Structural debt

There are two parallel service trees:
- `backend/services/{audio,coordinator,config,filesystem,system,terminal,theme}` — the active set (used by `app.coordinator`)
- `internal/services/{audio,colorscheme,error,filesystem,font,geoip,network,performance,security,settings,system,terminal,theme}` — older parallel implementations, some still used by the coordinator (`colorscheme`, `font`, `network`)

This is out of scope for this change but documented here so a future cleanup change can deduplicate.

## Action items derived from this diff

1. Replace the hand-written shim's global path: `window.go.main.ServiceCoordinator` → `window.go.coordinator.ServiceCoordinator`. Better: delete the shim and use the generated `wailsjs/go/coordinator/ServiceCoordinator` directly.
2. Add `GetTerminalCWD` + `GetCWDStats`.
3. Add a runtime probe at app boot that resolves `window.go.coordinator.ServiceCoordinator.IsStarted` and surfaces a fatal banner if missing — prevents the silent-mock failure mode.
