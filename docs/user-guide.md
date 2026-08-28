# aDex User Guide

aDex is a science-fiction desktop environment and terminal, inspired by
eDEX-UI. It puts a real shell, live system monitoring, a file browser and a
network view behind a single themed interface.

---

## Starting the application

Run the built binary:

```bash
./bin/aDex-UI
```

The window opens maximised. A boot sequence runs first, reporting each backend
service as it starts — configuration, colour schemes, fonts, network,
filesystem and the terminal service — before handing off to the interface.
If a service fails, the boot screen shows the failure and its detail rather
than exiting silently.

---

## The interface

The window is a top bar over three columns.

**Top bar** — section labels on the left (PANEL, SYSTEM, TERMINAL) and right
(PANEL, NETWORK), your shell tabs in the middle, and window controls at the
far right: settings (gear), fullscreen toggle, and quit.

**Left column** — clock, system information, CPU usage, memory and swap, and
the process list.

**Centre** — the terminal, with its tab strip above and a status line below.

**Right column** — network status, the world view globe with your approximate
location, and network traffic graphs.

**Bottom** — the file browser on the left and the on-screen keyboard on the
right.

---

## Terminal tabs

Up to **five** shells can run at once.

| Action | How |
|---|---|
| Switch tab | Click the tab, in the top bar or the shell tab strip |
| New shell | Click any `EMPTY` slot |
| Rename a tab | Double-click its label in the top bar, type, press Enter |
| Cancel a rename | Press Escape |
| Restart a dead shell | Click a tab marked `[DEAD]` |

Tab names persist across restarts. A shell that exits — because you typed
`exit`, or it crashed — is marked `[DEAD]` and can be restarted in place,
keeping its position and name.

---

## Themes and layouts

**21 themes** ship with the application, including `tron` (the default),
`matrix`, `nord`, `apollo`, `blade`, `cyborg`, `interstellar` and `red`.
Selecting a theme changes the accent colour, backgrounds, fonts, terminal
palette and globe colours immediately — no restart, no need to press Save.

**6 layout presets** rearrange the workspace: `default`, `disrupted`,
`typeleft`, `fulltype`, `notype` and `colorfilter`.

Every theme carries its own layout. If you have not chosen a layout
explicitly, the theme's layout is used — so picking `tron-disrupted` switches
you to the disrupted arrangement. Choosing a layout in Settings overrides
that for every theme; clearing it returns to following the theme.

If a stored theme or layout no longer exists — after an upgrade, say — the
application falls back to the default and repairs the stored value, rather
than starting unstyled.

### Custom layouts

Beyond the six built-in presets you can define your own arrangements in
`~/.config/aDex-UI/layouts.json`. Create the file with a list of layouts:

```json
[
  {
    "id": "minimal",
    "displayName": "Minimal",
    "regions": {
      "left":  ["clock", "cpu", "ram"],
      "right": ["netstat"]
    }
  }
]
```

Each layout needs an `id` (what gets stored as your layout choice) and a
`regions` map. Regions are `left`, `centre`, `right` and `bottom`; each holds
the panels to show there, in order.

Available panels:

| Panel | Shows |
|---|---|
| `clock` | Time and date |
| `sysinfo` | System information |
| `hardware` | Manufacturer, model, chassis |
| `cpu` | CPU usage and per-core graphs |
| `ram` | Memory and swap |
| `toplist` | Process list |
| `filesystem` | File browser |
| `netstat` | Network status |
| `globe` | World view |
| `traffic` | Network traffic graphs |
| `keyboard` | On-screen keyboard |

Omit a panel to hide it; reorder the list to reorder the column. Custom
layouts appear in Settings → Theme → Layout preset marked `(custom)`, and the
settings panel shows the file path.

Unknown panel or region names are skipped with a warning rather than breaking
the layout, and a malformed file leaves the built-in presets working. Restart
the application to pick up changes to the file.

---

## Settings

Open with the gear icon or **Ctrl+,**. Panels down the left, controls on the
right, and `RESET DEFAULTS` / `CANCEL` / `SAVE` along the bottom.

| Panel | What it controls |
|---|---|
| **Terminal** | Shell path, arguments, working directory |
| **Theme** | Theme and layout preset |
| **Keyboard** | On-screen keyboard layout |
| **Font** | Interface and terminal fonts, terminal font size |
| **Colour scheme** | Terminal colour palette |
| **Audio** | Master enable, volume, soundpack, per-category cues |
| **Network** | Ping target and refresh behaviour |
| **Security** | Security-related options |
| **System** | Startup directory and system behaviour |
| **Advanced** | Diagnostics and lower-level options |
| **Updates** | Update checking |

Theme, layout, keyboard, fonts and audio apply **as you change them**, so you
can judge the effect before committing. `CANCEL` reverts to the last saved
values; `SAVE` writes them to
`~/.config/aDex-UI/adex-ui-settings.json` and applies every group.

Changing the shell path affects **new** terminals; existing tabs keep the
shell they started with.

---

## The on-screen keyboard

The keyboard mirrors a physical one and sends keystrokes to the active
terminal. Modifiers (Shift, Ctrl, Alt, Fn) are momentary — they apply to the
next key and release automatically. CapsLock is a true toggle. Layouts are
selectable in Settings → Keyboard.

---

## Audio

Cues accompany the boot and shutdown sequences, typing, terminal output, tab
switches, panel opening, theme changes and destructive actions such as
terminating a process.

Cues fall into four categories you can switch off independently:

| Category | Covers |
|---|---|
| **Keyboard** | Typing, terminal input and output |
| **Destructive** | Process termination, denied actions, alarms |
| **Interface** | Clicks, panels, theme changes |
| **System** | Boot and shutdown sequences |

Master mute and volume affect everything. Two soundpacks are available:
`adex` (recorded samples) and `synth` (generated tones). If a sound file
cannot load, the application falls back to a synthesized tone rather than
failing.

Cues are played by the Go backend rather than the embedded browser. The
browser engine refuses to start audio until you interact with the window,
which used to make the whole boot sequence silent; playing through the
system audio stack has no such restriction, so the splash is audible from
the first frame.

---

## File browser

Lists the current directory, with `Show disks` for mounted volumes. Clicking
a directory navigates into it and changes the active terminal's directory to
match.

---

## Quitting

Use the **×** button in the top bar or **Ctrl+Q**, then confirm. A shutdown
sequence runs while backend services are released: terminal sessions and
their child processes are terminated, and audio devices are closed, before
the process exits.

---

## Keyboard shortcuts

| Shortcut | Action |
|---|---|
| `Ctrl+,` | Open settings |
| `Ctrl+Q` | Quit |
| `Ctrl+\`` | Toggle terminal |
| `Ctrl+Shift+F` | Toggle file browser |
| `Ctrl+Shift+S` | Toggle system monitor |

---

## Troubleshooting

**The interface looks unstyled.** A stored theme naming something that no
longer exists is repaired automatically on next start. If it persists, delete
`~/.config/aDex-UI/adex-ui-settings.json` to return to defaults.

**No sound.** Check Settings → Audio: master enable, volume, and the category
covering the cue you expect. Sounds are also suppressed when the window is in
the background if "mute in background" is on.

**A terminal tab shows `[DEAD]`.** The shell exited. Click the tab to start a
fresh one in the same slot.

**The window will not shrink far enough.** The window has a minimum of
640×480. The interface reflows down to that size.

---

## See also

- [Resource usage](resource-usage.md) — measured memory and CPU
- [Development](development.md) — building and contributing
