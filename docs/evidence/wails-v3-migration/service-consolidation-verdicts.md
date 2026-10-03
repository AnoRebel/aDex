# Duplicated service consolidation — verdicts

Date: 2026-08-27. Part of the Wails v2 → v3 migration.

Five services existed in both `backend/services/` (wired into the running app)
and `internal/services/` (no Go importers). The dead copies were 3–10x larger,
so they were evaluated rather than deleted unread: size could have meant a
more complete implementation worth adopting.

## Method to decide

For each pair, the test was whether the dead `internal/` implementation
satisfies the contract the coordinator actually depends on — the exported
methods it calls — and how each fits the Wails v3 Services model.

## Result

| Service | backend (live) | internal (dead) | Coordinator methods missing from `internal/` | Verdict |
|---|---|---|---|---|
| audio | 769 loc | 1,077 loc | `StartMonitoring`, `StopMonitoring` | **Keep backend** |
| filesystem | 305 loc | 3,154 loc | `CopyFile`, `MoveFile`, `ReadDirectory`, `WriteFile` | **Keep backend** |
| system | 720 loc | 2,378 loc | 9 of 13, incl. `GetCPUUsage`, `GetMemoryUsage`, `GetDiskUsage` | **Keep backend** |
| terminal | 738 loc | 3,180 loc | 10 of 12, incl. `CreateTerminal`, `WriteToTerminal`, `ResizeTerminal` | **Keep backend** |
| theme | 470 loc | 1,722 loc | (coordinator calls none directly) | **Keep backend** |

Every dead implementation fails the contract test: none provides the methods
the coordinator calls. They are a different, incomplete design — an earlier or
abandoned direction — not a fuller version of the live code. Larger line count
reflected unused surface area, not additional working capability.

## Consequences

- No verdict adopts a dead implementation, so **the consolidation introduces no
  behaviour change**.
- The five `internal/` duplicates are deleted; the live `backend/` packages move
  to `internal/` unchanged.
- Anything genuinely wanted from the deleted code remains recoverable from git
  history.

`internal/services/{colorscheme,error,font,geoip,network,performance,security,settings}`
are not duplicates — they are live and stay as they are.
