# Contributing to aDex

Thanks for taking an interest. This document covers how to get set up, the
conventions the codebase follows, and the things worth checking before you
open a pull request.

## Getting set up

```bash
# One-time
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
wails3 doctor          # verifies the toolchain; Linux needs GTK4 + WebKitGTK 6.0
cd frontend && bun install

# Develop
wails3 dev             # live-reloading app
wails3 task build      # production binary at bin/adex
```

aDex tracks the latest Wails v3 **beta** rather than pinning, because v3 is
pre-stable and ships fixes frequently. An upstream release can break the build
between one day and the next; that is expected, not a defect in this codebase.
Re-run `go get github.com/wailsapp/wails/v3@latest`, reinstall the CLI so the
two match, and re-verify.

## Project layout

```
main.go                  Application entry point and window configuration
internal/
  appdir/                Config and data directory resolution (single source)
  services/
    coordinator/         Composes every service; the only Wails-registered one
    terminal/  system/   PTY, metrics, filesystem, audio, network, …
    security/            Input validation and the session lock
  models/                Types crossing the Go/JS boundary
frontend/app/
  pages/index.vue        The shell: columns, top bar, layout
  components/adex/       The panels ("mods")
  composables/           Theme, keyboard, audio, layout and module engines
  assets/css/layouts/    Layout presets, keyed on data-layout
tests/                   Go tests; frontend tests live in frontend/tests
```

All Go code lives under `internal/`. There is no `backend/` tree — it was
merged during the Wails v3 migration.

## Conventions

**Comments explain why, not what.** The codebase leans on this heavily: many
non-obvious lines exist because of a specific bug, and the comment names it.
If you change such a line, update the reasoning with it.

**Verify against the running app.** Several bugs here looked fixed but were
not, because a CSS rule lost a specificity contest or a setting wrote a
variable nothing read. Prefer measuring the real DOM or a real run over
reasoning about what should happen.

**Scoped styles win.** Component `<style scoped>` blocks compile to selectors
carrying a `data-v-*` attribute, so they beat a plain selector in `main.css`
even when the plain one looks more specific. Column and panel sizing is owned
by the scoped block in `pages/index.vue`; put changes there.

**Assets under `app/assets/` are not served over HTTP** — only
`frontend/public/` is. Load them with `import.meta.glob` so the bundler
resolves them, never with a runtime-computed `import()` or a fetch.

**Prefer explicit imports.** Nuxt auto-import has failed at runtime in this
project; components and composables are imported explicitly.

## Testing

```bash
go test ./...                      # Go
cd frontend && bunx vitest run     # frontend unit tests
cd frontend && bunx vue-tsc --noEmit
```

A browser can drive the UI (`bun run dev` plus `tests/visual/`), but **backend
calls do not work there**: with no Wails IPC peer, a binding call resolves to
the SPA's `index.html` with a 200 status rather than throwing. Verify anything
backend-dependent against the packaged binary.

## Adding things

- **A theme** — see [docs/themes.md](docs/themes.md). Add the JSON under
  `frontend/app/assets/data/themes/` and register it in `themes-index.json`.
- **A layout preset** — add a CSS file under `assets/css/layouts/` keyed on
  `[data-layout="your-id"]`, import it from `main.css`, and register the id in
  the settings list and `internal/services/settings/validate.go`. Note the
  shell is flexbox: absolute offsets ported from eDEX-UI do not reflow.
- **A panel (mod)** — create the component under `components/adex/`, add it to
  `PANEL_COMPONENTS` in `pages/index.vue`, and register it in
  `composables/useModules.ts` so it is toggleable and orderable. Panel ids are
  a compatibility surface: renaming one resets users' saved settings.
- **A backend method** — add it to `ServiceCoordinator`; bindings regenerate
  with `wails3 generate bindings -ts`. If it exposes files, terminals or
  processes, guard it with `sc.guardLocked()`.

## Pull requests

Before opening one:

- `go test ./...` and the frontend checks pass
- The packaged binary starts, and exits cleanly
- New behaviour is covered by a test, or you say why it is not
- Commit messages explain the reasoning, not just the change

Larger changes use [OpenSpec](openspec/) — a proposal, specs, design and task
breakdown before implementation. Run `openspec list` to see active changes.
