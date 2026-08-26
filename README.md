# aDex-UI - Modern Science Fiction Desktop Environment

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Vue.js-3.x+-4FC08D?style=for-the-badge&logo=vue.js&logoColor=white" alt="Vue.js">
  <img src="https://img.shields.io/badge/Nuxt-4.x+-00DC82?style=for-the-badge&logo=nuxt.js&logoColor=white" alt="Nuxt">
  <img src="https://img.shields.io/badge/Wails-v2.12-000000?style=for-the-badge&logo=wails&logoColor=white" alt="Wails">
  <img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="License">
</p>

A modern terminal emulator with system monitoring, built with Wails v2 and Nuxt 4. This project migrates the classic eDEX-UI experience to a modern, high-performance desktop application.

## ✨ Features

🖥️ **Terminal Emulator**
- Multi-tab terminal with xterm.js and WebGL rendering
- Full curses application support (htop, vim, tmux, etc.)
- Real-time current working directory tracking
- Customizable color schemes and fonts
- Copy/paste with system clipboard integration

📊 **System Monitoring**
- Real-time CPU, memory, disk, and process monitoring
- Professional animated charts with Chart.js
- <3% CPU usage with throttled updates
- Process sorting and filtering
- Temperature and sensor monitoring (where available)

🌐 **Network Monitoring**
- Interface status and configuration
- Real-time transfer rates and bandwidth usage
- Active connection tracking
- IP address information
- GeoIP support (optional, lazy loaded)

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

- **Backend**: Go 1.21+ with Wails v2.12.0
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

## Getting Started

### Prerequisites

- Go 1.25+ (latest stable)
- Bun 1.1+ or Node.js 20+ LTS
- Wails v2.12.0: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`

### Development

1. Install frontend dependencies:
   ```bash
   cd frontend && bun install
   ```

2. Generate Wails bindings:
   ```bash
   wails generate module
   ```

3. Run in development mode:
   ```bash
   wails dev
   ```

4. Or use Taskfile for structured commands:
   ```bash
   task dev  # Runs both frontend and backend
   ```

### Building

Build for current platform:
```bash
wails build
```

Build for all platforms:
```bash
task build:all
```

## Project Structure

```
aDex-UI/
├── main.go              # Wails application entry point
├── Taskfile.yml          # Root build configuration
├── app.go               # Go application context
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

## Documentation

- Development guide: `docs/development.md`
- Troubleshooting: `docs/troubleshooting.md`
- Theme documentation: `docs/themes/`

## Contributing

1. Fork the repository
2. Create a feature branch from main
3. Implement changes with proper testing
4. Submit a pull request with description

## License

Copyright © 2025 HackEAC
