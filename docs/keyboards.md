# Keyboard Layouts

aDex ships **19 on-screen keyboard layouts** under
`frontend/app/assets/data/kb_layouts/`, indexed by
`frontend/app/assets/data/kb_layouts-index.json`.

The on-screen keyboard is a *visual* layout + a key→byte-sequence map.
Physical keyboard events are passed through to the terminal unchanged;
the layout only affects what the on-screen keys send when clicked and
which glyphs they display.

## Switching layouts

Settings → **KEYBOARD** tab. Selecting a layout applies it immediately
via `useAdexKeyboard().setLayout(id)` and persists to
`localStorage`. The `AdexKeyboard` component watches the engine's
`activeId` so the on-screen render flips without a restart.

Programmatically: `useAdexKeyboard().setLayout(id)`.

## Bundled layouts

| id          | Name |
|-------------|------|
| `da-DK`     | Danish (Denmark) |
| `de-DE`     | German (Germany) |
| `en-COLEMAK`| English (Colemak) |
| `en-DVORAK` | English (Dvorak) |
| `en-GB`     | English (UK) |
| `en-NORMAN` | English (Norman) |
| `en-US`     | English (US) |
| `en-WORKMAN`| English (Workman) |
| `es-ES`     | Spanish (Spain) |
| `es-LAT`    | Spanish (Latin America) |
| `fr-BEPO`   | French (Bépo) |
| `fr-FR`     | French (France) |
| `hu-HU`     | Hungarian (Hungary) |
| `it-IT`     | Italian (Italy) |
| `nl-BE`     | Dutch (Belgium) |
| `pt-BR`     | Portuguese (Brazil) |
| `sv-SE`     | Swedish (Sweden) |
| `tr-TR-F`   | Turkish (F) |
| `tr-TR-Q`   | Turkish (Q) |

## Layout schema

Index entry (`kb_layouts-index.json`):

```json
{
  "id": "en-US",
  "displayName": "English (US)",
  "file": "kb_layouts/en-US.json"
}
```

Layout file (`kb_layouts/<id>.json`) — keyed by row name
(`row_numbers`, `row_top`, `row_home`, `row_bottom`, `row_modifiers`,
etc.). Each entry is a key:

```json
{
  "name": "1",                  // glyph shown on the key
  "cmd": "1",                   // bytes sent when clicked
  "shift_name": "!",            // glyph when Shift is held
  "shift_cmd": "!",             // bytes sent with Shift
  "alt_cmd": "~~~CTRLSEQ1~~~1", // bytes sent with Alt
  "ctrl_cmd": "~~~CTRLSEQ1~~~", // bytes sent with Ctrl
  "fn_name": "F1",              // glyph in Fn mode
  "fn_cmd": "~~~CTRLSEQ1~~~OP"  // bytes sent in Fn mode
}
```

### Escape-sequence sentinels

`~~~CTRLSEQ1~~~` is a placeholder the keyboard engine expands to the
ASCII escape byte (`\x1b` / ESC, decimal 27) at send time. This keeps
the JSON free of raw control characters that would corrupt the file
or be stripped by editors. Other sentinels follow the same
`~~~NAME~~~` convention; the expansion table lives in
`useAdexKeyboard`.

A key with only `name` + `cmd` is a plain printable key. Special keys
(ESC, TAB, modifiers) carry `type` hints used for sizing/styling, not
for sequence resolution.

## key → sequence reference (common control keys)

| Key       | Sends (decoded) |
|-----------|-----------------|
| Enter     | `\r` (CR) |
| Tab       | `\t` |
| Backspace | `\x7f` (DEL) |
| Escape    | `\x1b` |
| ↑ ↓ → ←   | `\x1b[A` / `\x1b[B` / `\x1b[C` / `\x1b[D` |
| Home/End  | `\x1b[H` / `\x1b[F` |
| PgUp/PgDn | `\x1b[5~` / `\x1b[6~` |
| Delete    | `\x1b[3~` |
| F1–F12    | `\x1bOP`-style sequences (see `fn_cmd` per layout) |

The terminal-side decode of these (for the on-screen-keyboard pulse
highlight) is implemented in
`frontend/app/composables/useKeyboardPulse.ts`.

## Adding a custom layout

1. Create `frontend/app/assets/data/kb_layouts/<your-id>.json` with
   the row structure above.
2. Append an index entry to `kb_layouts-index.json`.
3. Rebuild. It appears in the KEYBOARD tab automatically — the
   selector is driven entirely off the index.
