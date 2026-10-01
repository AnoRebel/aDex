# aDex — Modern Science Fiction Desktop Environment

<p align="center">
  <a href="https://github.com/AnoRebel/aDex/actions/workflows/build.yml"><img src="https://img.shields.io/github/actions/workflow/status/AnoRebel/aDex/build.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white" alt="Build"></a>
  <a href="https://github.com/AnoRebel/aDex/releases/latest"><img src="https://img.shields.io/github/v/release/AnoRebel/aDex?style=for-the-badge&logo=github" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/AnoRebel/aDex?style=for-the-badge" alt="License"></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Wails-v3_beta-D32A2D?style=for-the-badge&logo=wails&logoColor=white" alt="Wails">
  <img src="https://img.shields.io/badge/Nuxt-4.x-00DC82?style=for-the-badge&logo=nuxt&logoColor=white" alt="Nuxt">
  <img src="https://img.shields.io/badge/Vue-3.x-4FC08D?style=for-the-badge&logo=vue.js&logoColor=white" alt="Vue">
  <img src="https://img.shields.io/badge/TypeScript-6.x-3178C6?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript">
</p>

A modern terminal emulator with system monitoring, built with Wails v3 and Nuxt 4. This project migrates the classic eDEX-UI experience to a modern, high-performance desktop application.

## ✨ Features

🖥️ **Terminal Emulator**
- Multi-tab terminal (up to 5) with xterm.js; optional WebGL renderer
- Full curses application support (htop, vim, tmux, etc.)
- Real-time current working directory tracking
- Customizable color schemes and fonts
- Copy/paste with system clipboard integration

📊 **System Monitoring**
- Real-time CPU, memory, disk, and process monitoring
- Live charts built on TanStack Charts
- Throttled updates; see [resource usage](docs/resource-usage.md) for measured figures
- Process sorting and filtering
- Temperature and sensor monitoring (where available)

🌐 **Network Monitoring**
- Interface status and configuration
- Real-time transfer rates and bandwidth usage
- Active connection tracking
- IP address information
- GeoIP support (optional, lazy loaded)

🎨 **Theming & Layout**
- 21 themes, applied instantly and persisted
- 7 layout presets, plus user-defined layouts in `layouts.json`
- Every panel individually toggleable and reorderable
- Movable, resizable settings panel

🔒 **Session Lock**
- Passphrase-protected lock screen with idle auto-lock (`Ctrl+Shift+L`)
- Enforced in the backend, not just the UI — file, terminal and process
  access is refused while locked
- Terminals keep running while locked

🔊 **Audio**
- Cues for boot, shutdown, typing, tab changes and destructive actions
- Played by the Go backend, so the splash is audible from the first frame
- Per-category mutes, master volume, two soundpacks

📁 **File Browser**
- Visual file management with icon support
- Follows terminal current working directory
- Path injection to terminal
- Keyboard navigation support
- File permissions and size information

🎨 **Theme System**
- Modern CSS variable-based theming
- Legacy eDEX-UI theme import and conversion
- Live theme switching without restart
- Support for custom themes
- Multiple built-in themes (Cyberpunk, Retro, Minimal)

⌨️ **On-Screen Keyboard**
- Touch-friendly keyboard with haptic feedback
- Multiple layouts: QWERTY, QWERTZ, AZERTY, Dvorak, Colemak
- Key repeat functionality with configurable delays
- Accessibility support with ARIA labels
- Responsive design for tablets and touch devices

🔊 **Audio Effects System**
- Optional sound effects (disabled by default)
- Multiple soundpacks (Default, Minimal, Retro, Cyberpunk)
- Configurable volume controls (master, effects, notifications)
- Event mapping system for UI interactions
- Web Audio API with HTML5 fallback
- Cooldown system to prevent audio spam

## Technology Stack

- **Backend**: Go 1.25+ with Wails v3 (beta)
- **Frontend**: Nuxt v4 + Vue 3 + TypeScript + Pinia
- **Terminal**: xterm.js v5 with WebGL addon
- **Monitoring**: gopsutil for cross-platform system metrics
- **Build**: Taskfile-based build system for cross-platform deployment

## Performance Targets

- Cold start: <2 seconds
- Idle CPU usage: <3%
- 60fps animations with 4ms frame budget
- Memory usage: <200MB at idle
- Cross-platform: Windows 10+, macOS 11+, Linux

## Documentation

**Using aDex**
- **[User guide](docs/user-guide.md)** — interface, terminal tabs, themes and layouts, settings, keyboard, audio and the session lock
- [Troubleshooting](docs/troubleshooting.md) — common problems and fixes
- [Resource usage](docs/resource-usage.md) — measured memory and CPU figures

**Extending aDex**
- [Themes](docs/themes.md) — writing a theme, the colour and layout schema
- [Custom layouts](docs/user-guide.md#custom-layouts) — arranging panels via `layouts.json`
- [Keyboard layouts](docs/keyboards.md) — adding an on-screen keyboard layout
- [Audio cues](docs/audio.md) — the cue vocabulary and soundpacks

**Developing aDex**
- [Development](docs/development.md) — building, the Wails v3 workflow, project layout
- [Contributing](CONTRIBUTING.md) — workflow, conventions, what a good change looks like
- [Cross-platform packaging](docs/cross-platform-packaging.md) — building for other platforms

## Getting Started

### Prerequisites

- Go 1.25+ (latest stable)
- Bun 1.1+ or Node.js 20+ LTS
- Wails v3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`

  Wails v3 is in beta and ships fixes frequently, so this project tracks the
  latest beta rather than pinning. Verified against **v3.0.0-beta.14**
  (2026-08-26). Run `wails3 doctor` to check your environment; on Linux it
  needs GTK4 and WebKitGTK 6.0.

### Development

1. Install frontend dependencies:
   ```bash
   cd frontend && bun install
   ```

2. Generate Wails bindings (also done automatically by the build tasks):
   ```bash
   wails3 generate bindings -ts
   ```

3. Run in development mode:
   ```bash
   wails3 dev
   ```

4. Or use Taskfile for structured commands:
   ```bash
   task dev  # Runs both frontend and backend
   ```

### Building

Build for current platform:
```bash
wails3 task build      # or: task build
```

The build compiles the Nuxt frontend to `frontend/dist` and embeds it in the
binary. Because that directory is generated (and gitignored), a clean checkout
must build the frontend before `go build` will succeed on its own.

Package for the current platform:
```bash
task package
```

Releases for Linux, macOS and Windows (amd64 and arm64) are built by CI from
a `v*` tag — see [`.github/workflows/build.yml`](.github/workflows/build.yml).

## Project Structure

```
aDex/
├── main.go              # Wails application entry point
├── tray.go              # System tray icon, menu and quick actions
├── Taskfile.yml         # Root build configuration
├── internal/
│   ├── services/        # Go backend services
│   │   ├── terminal/   # Terminal PTY management
│   │   ├── system/     # System metrics collection
│   │   ├── network/    # Network monitoring
│   │   ├── filesystem/ # File system operations
│   │   ├── settings/   # Configuration management
│   │   └── theme/      # Theme management
│   ├── models/          # Data models
│   └── events/          # Event handling
├── frontend/
│   ├── nuxt.config.ts    # Nuxt configuration
│   ├── components/      # Vue components
│   ├── composables/     # Vue composables
│   ├── stores/          # Pinia state management
│   └── bindings/        # Generated Wails bindings
└── build/               # Build configuration and assets
```

## Development Workflow

1. **Setup**: Configure development environment and install dependencies
2. **Backend**: Implement Go services with proper error handling
3. **Frontend**: Create Vue components with TypeScript and Pinia stores
4. **Integration**: Connect backend services to frontend via Wails bindings
5. **Testing**: Use Go testing + Vitest + Playwright
6. **Build**: Create platform-specific installers

## Configuration

- **Wails Config**: `build/config.yml`
- **Nuxt Config**: `frontend/nuxt.config.ts`
- **Build Tasks**: `Taskfile.yml` and `build/Taskfile.yml`

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for the workflow, conventions and
what to check before opening a pull request.

## License

Copyright © 2026 Ano Rebel
