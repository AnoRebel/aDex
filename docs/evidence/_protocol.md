# Evidence Capture Protocol

Every feature claimed complete requires three evidence layers:

1. **Go boundary test** — `tests/backend/<service>_boundary_test.go` (happy + error path).
2. **Vitest store/composable test** — `tests/frontend/<store>.spec.ts` mocking the binding.
3. **UI evidence** — at least one screenshot at `docs/evidence/<feature>/<theme>.png` plus `notes.md`.

## Capturing screenshots

Wails v2 doesn't ship a built-in screenshotter that's CLI-friendly, so use the host OS:

### Linux (X11 / Wayland)

```bash
# X11
import -window "$(xdotool search --name 'aDex-UI' | head -1)" docs/evidence/<feature>/<theme>.png

# Wayland (sway / Hyprland)
grim -g "$(slurp -w "$(swaymsg -t get_tree | jq -r '.. | select(.name? == "aDex-UI") | .rect | "\(.x),\(.y) \(.width)x\(.height)"')")" docs/evidence/<feature>/<theme>.png

# Generic GNOME
gnome-screenshot -w -f docs/evidence/<feature>/<theme>.png
```

### macOS

```bash
# Active window (Cmd+Shift+4 then Space then click)
screencapture -W docs/evidence/<feature>/<theme>.png
```

### Windows (PowerShell)

```powershell
Add-Type -AssemblyName System.Windows.Forms,System.Drawing
$proc = Get-Process aDex-UI | Where-Object MainWindowHandle -ne 0 | Select-Object -First 1
# ... use the handle to BitBlt; or use Snipping Tool's CLI
```

## notes.md template

Copy `docs/evidence/_template/notes.md` to the feature directory and fill it in:

```
# <feature>

- OS:           <linux|darwin|windows> <version>
- Theme:        <theme-id>
- Wails:        v2.12.0
- Date:         <ISO 8601>
- App SHA:      <git rev-parse HEAD>
- Verifier:     <name>

## What this proves

<one paragraph>

## How to reproduce

1. <step>
2. <step>

## Observed

<bullets>
```

## Wails native helper (for the future)

When time permits, wire `runtime.WindowSetSize` + a screenshotter binary into `bun run evidence:capture <feature> <theme>` so the CI loop can produce evidence without manual capture.
