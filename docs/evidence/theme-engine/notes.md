# theme-engine

- OS:           linux 6.19.13-zen1-1-zen
- Theme:        all 21 (validation suite)
- Wails:        v2.12.0
- Date:         2026-05-05
- App SHA:      (uncommitted, branch main)
- Verifier:     automated — frontend/tests/composables/use-adex-theme.spec.ts

## What this proves

Validates the V2 theme engine end-to-end:
- the static asset library at `frontend/app/assets/data/themes/` contains
  21 well-formed theme JSONs that parse against the schema in
  `frontend/app/types/adex-theme.ts`
- each theme declares a layout preset from the typed enum
  (`default | disrupted | typeleft | fulltype | notype | colorfilter`),
  replacing the old injectCSS mechanism
- the runtime composable (`useAdexTheme`) applies CSS custom properties
  and the `data-layout` attribute within the 200 ms hot-swap budget
- localStorage persistence round-trips correctly
- unknown ids surface a typed error rather than silently falling back

## How to reproduce

```bash
bun scripts/convert-themes.mjs              # rebuild the asset tree
cd frontend
bunx vitest run tests/composables/use-adex-theme.spec.ts
```

Expected: `Test Files 1 passed (1) | Tests 7 passed (7)`.

## Observed

- 7/7 tests pass, total runtime ~240 ms in Vitest (happy-dom).
- See `load-report.md` in this directory for the full inventory + breakdown.

## UI screenshots

Per-theme rendered-UI captures (Tron / Tron-Disrupted / Blade / Horizon /
Matrix / Nord / Navy-Disrupted / Tron-Typeleft) are deferred until
`wails dev` is run on the dev machine. Use:

```bash
task evidence:capture -- theme-engine tron
task evidence:capture -- theme-engine tron-disrupted
# ...
```
