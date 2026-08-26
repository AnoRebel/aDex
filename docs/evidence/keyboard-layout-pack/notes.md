# keyboard-layout-pack

- OS:           linux 6.19.13-zen1-1-zen
- Theme:        n/a
- Wails:        v2.12.0
- Date:         2026-05-05
- App SHA:      (uncommitted, branch main)
- Verifier:     automated — frontend/tests/composables/use-adex-keyboard.spec.ts

## What this proves

Validates the V2 keyboard system end-to-end:
- 19 layouts present in the index, matching the upstream eDEX-UI set
  (da-DK, de-DE, en-COLEMAK, en-DVORAK, en-GB, en-NORMAN, en-US,
  en-WORKMAN, es-ES, es-LAT, fr-BEPO, fr-FR, hu-HU, it-IT, nl-BE,
  pt-BR, sv-SE, tr-TR-F, tr-TR-Q)
- Schema preserved verbatim (row_numbers / row_1..3 / row_space with
  cmd / shift_cmd / ctrl_cmd / alt_cmd / fn_cmd / capslck_cmd) so
  alt-graph chars and dead keys keep working
- CTRLSEQ substitution fills the placeholders the upstream runtime left
  empty, so Ctrl+C now actually produces 0x03, F1 produces \x1bOP, etc.
- Modifier dispatch order matches the original: alt+shift > alt > ctrl
  > capslck > shift > fn > base
- Hot layout switch + persistence + unknown-id error all covered

## How to reproduce

```bash
bun scripts/copy-kb-layouts.mjs
cd frontend
bunx vitest run tests/composables/use-adex-keyboard.spec.ts
```

Expected: `Tests 8 passed (8)`.

## Observed

- 8/8 tests pass (~290 ms in jsdom)
- 1 upstream-source repair logged: `en-WORKMAN.json` was missing its
  closing `}` — the copier auto-repairs it. Tracked here for awareness;
  no manual intervention needed.

## UI screenshots

Per-layout rendered captures (en-US, fr-FR, de-DE, en-DVORAK at minimum)
are deferred until `wails dev` runs on the dev machine:

```bash
task evidence:capture -- keyboard-layout-pack en-US
task evidence:capture -- keyboard-layout-pack fr-FR
```
