# Themes

aDex-UI ships **21 built-in themes**, each pairing a color palette with a
layout preset. Themes are pure data — JSON files under
`frontend/app/assets/data/themes/` indexed by
`frontend/app/assets/data/themes-index.json`.

## Switching themes

Settings → **THEME** tab. Selecting a theme applies it immediately
(no Save needed) and persists the choice to `localStorage`
(`adex-settings.display.theme`). It rehydrates on the next launch.

Programmatically: `useAdexTheme().setTheme(id)`.

## Layout vs theme

Every theme declares a `layout` field — one of the seven layout presets
(see [Layout presets](#layout-presets)). By default the active theme's
bundled layout drives the shell. The user can **override** the layout
independently via Settings → THEME → **Layout preset**; the override
wins over the theme's choice and is stored at
`adex-settings.display.layout` (empty string = follow the theme).

Priority (first non-empty wins):

1. `adex-settings.display.layout` — explicit user override
2. The active theme's `layout` field
3. `'default'`

## Layout presets

CSS lives in `frontend/app/assets/css/layouts/<preset>.css` and applies
via the `data-layout` attribute on `.adex-app`.

| Preset        | Description |
|---------------|-------------|
| `default`     | Classic eDEX — left mods column, center terminal + tabs, file manager bottom-left, on-screen keyboard bottom-right, right mods column. |
| `disrupted`   | Asymmetric grid — narrower left column, wider right column, terminal between them. |
| `typeleft`    | File manager and keyboard swap sides in the bottom row. |
| `fulltype`    | Full-**width on-screen keyboard**, file manager hidden. (Not a fullscreen terminal — see `terminal-focus`.) |
| `notype`      | On-screen keyboard hidden; the file manager takes the whole bottom row. |
| `colorfilter` | Monochrome pass over the interface; the terminal keeps its colours so program output stays readable. |
| `terminal-focus` | Terminal fills the window; side columns, keyboard and file manager hidden. |

Every panel (left mods, right mods, keyboard, file manager, terminal)
is a first-class layout slot. New panels MUST add their
`display:none` / repositioning rules to each preset file, and must
reflow on available width rather than assuming a fixed sibling
configuration.

## Theme schema

Index entry (`themes-index.json`):

```json
{
  "id": "tron",
  "displayName": "Tron",
  "layout": "default",
  "primary": "rgb(170, 207, 209)",
  "file": "themes/tron.json"
}
```

Theme file (`themes/<id>.json`) — the eDEX-derived shape:

```json
{
  "colors": { "r": 170, "g": 207, "b": 209,
              "black": "#000000", "light_black": "#05080d", "grey": "#262828" },
  "cssvars": { "font_main": "United Sans Medium",
               "font_main_light": "United Sans Light" },
  "terminal": { "fontFamily": "Fira Code", "cursorStyle": "block",
                "foreground": "#aacfd1", "background": "#05080d",
                "cursor": "#aacfd1", "selection": "rgba(170,207,209,0.3)" },
  "globe": { "base": "#000000", "marker": "#aacfd1",
             "pin": "#aacfd1", "satellite": "#aacfd1" }
}
```

## Adding a custom theme

1. Create `frontend/app/assets/data/themes/<your-id>.json` following
   the schema above.
2. Append an index entry to `themes-index.json` with `id`,
   `displayName`, `layout` (one of the six presets), `primary`, and
   `file`.
3. Rebuild (`bun run build`). The new theme appears in the THEME tab
   automatically — the selector is driven entirely off the index.

There is no separate registration step: the theme engine
(`useAdexTheme`) reads the index at runtime.

## Bundled themes

`apollo, apollo-notype, blade, chalkboard, chalkboard-ligatures,
chalkboard-notype, cyborg, cyborg-focus, interstellar, matrix, navy,
navy-disrupted, navy-notype, nord, red, tron, tron-colorfilter,
tron-disrupted, tron-fulltype, tron-notype, tron-typeleft`

The `-notype`, `-disrupted`, `-fulltype`, `-colorfilter`, `-typeleft`
suffixed variants are the same palette pre-paired with that layout
preset (so picking them sets both at once).
