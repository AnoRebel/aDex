# aDex-UI Quick Start Guide

Get up and running with aDex-UI in under 5 minutes!

## 🚀 Quick Installation

### Option 1: Download Pre-built Binary

1. **Download for your platform**:
   - [Windows (x64)](https://github.com/adex-ui/adex-ui/releases/latest/download/aDex-UI_Windows_x64.exe)
   - [macOS (Intel)](https://github.com/adex-ui/adex-ui/releases/latest/download/aDex-UI_macOS_intel.dmg)
   - [macOS (Apple Silicon)](https://github.com/adex-ui/adex-ui/releases/latest/download/aDex-UI_macOS_arm64.dmg)
   - [Linux (x64)](https://github.com/adex-ui/adex-ui/releases/latest/download/aDex-UI_Linux_x64.tar.gz)

2. **Install and run**:
   - **Windows**: Run the installer or extract and run `aDex-UI.exe`
   - **macOS**: Open the DMG and drag aDex-UI to Applications
   - **Linux**: Extract the archive and run `./aDex-UI`

### Option 2: Build from Source

1. **Prerequisites**:
   ```bash
   # Install Go 1.21+
   # Install Bun or Node.js 18+
   # Install Task runner
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

2. **Clone and build**:
   ```bash
   git clone https://github.com/adex-ui/adex-ui.git
   cd adex-ui
   task build
   ```

## 🖥️ First Launch

### 1. Start aDex-UI
Double-click the application or run from terminal:
```bash
./aDex-UI
```

### 2. Create Your First Terminal Tab
- Click the **+** button in the terminal tab bar
- Or press `Ctrl+T` (Windows/Linux) or `Cmd+T` (macOS)

### 3. Try Some Commands
```bash
# System information
uname -a

# Directory listing
ls -la

# System monitoring (if installed)
htop

# File editor
vim ~/.bashrc

# Exit vim: :q! then Enter
```

## ⚙️ Basic Configuration

### Terminal Settings

1. **Open Settings**:
   - Click the gear icon ⚙️ in the top bar
   - Or use menu: `Settings → Terminal`

2. **Configure Your Shell**:
   ```
   Shell Path: /bin/bash        # Default shell
   Shell Path: /bin/zsh         # Z shell
   Shell Path: /bin/fish         # Fish shell
   ```

3. **Adjust Appearance**:
   - Font size: Increase for better readability
   - Theme: Choose a terminal color scheme
   - Transparency: Adjust window opacity

### System Monitoring

1. **Enable Monitoring**:
   - Check "Show System Monitor" in settings
   - Choose update interval (1-5 seconds recommended)

2. **Monitor Tabs**:
   - **CPU**: Real-time CPU usage and cores
   - **Memory**: RAM usage and processes
   - **Network**: Interface status and bandwidth
   - **Disk**: Storage usage and I/O

### File Browser

1. **Enable File Browser**:
   - Check "Show File Browser" in settings
   - Choose position (left, right, or bottom)

2. **Navigate Files**:
   - Double-click folders to open
   - Single-click files to select
   - Right-click for context menu
   - Drag files to terminal for path injection

## 🎨 Customizing Your Experience

### Choose a Theme

1. **Open Theme Selector**:
   - Settings → Themes
   - Or click the color palette icon 🎨

2. **Try Built-in Themes**:
   - **Default**: Clean blue theme
   - **Cyberpunk**: Neon sci-fi theme
   - **Retro**: Green phosphor terminal
   - **Minimal**: Clean, distraction-free

3. **Import Legacy Themes**:
   - Click "Import Theme"
   - Select eDEX-UI theme directory
   - Auto-converts to modern format

### On-Screen Keyboard (Tablets/Touch)

1. **Enable Keyboard**:
   - Settings → Keyboard
   - Check "Show on-screen keyboard"

2. **Choose Layout**:
   - QWERTY (default)
   - QWERTZ (German)
   - AZERTY (French)
   - Dvorak (optimized)
   - Colemak (ergonomic)

3. **Adjust Settings**:
   - Key repeat speed
   - Visual feedback
   - Haptic feedback

### Audio Effects

1. **Enable Audio** (Optional):
   - Settings → Audio
   - Check "Enable sound effects"
   - Start with low volume (20-30%)

2. **Choose Soundpack**:
   - **Default**: Clean, professional sounds
   - **Minimal**: Subtle, reduced noise
   - **Retro**: 8-bit computer sounds
   - **Cyberpunk**: Electronic sci-fi sounds

## 🔧 Common Tasks

### Multiple Terminal Sessions

```bash
# Create new tab
Ctrl+T or Cmd+T

# Switch tabs
Ctrl+Tab or Ctrl+1/2/3...

# Close current tab
Ctrl+W or Cmd+W

# Rename tab (double-click tab label)
```

### Copy and Paste

```bash
# Copy selection
Ctrl+Shift+C or Cmd+Shift+C

# Paste from clipboard
Ctrl+Shift+V or Cmd+Shift+V

# Or use right-click context menu
```

### File Operations

```bash
# Navigate to terminal directory
cd /path/to/folder

# Current working directory appears in file browser

# Drag file to terminal to inject path
# Right-click file for options
```

### System Monitoring

```bash
# Monitor specific process
# Process appears in process list with CPU/memory usage

# Check network activity
# Interface shows real-time bandwidth

# Disk usage monitoring
# See which directories use most space
```

## 📱 Keyboard Shortcuts

### Global Shortcuts
- `F11` or `Cmd+F11`: Toggle fullscreen
- `Ctrl+K` or `Cmd+K`: Quick command palette
- `Ctrl+P` or `Cmd+P`: Settings
- `Ctrl+,` or `Cmd+,`: Preferences

### Terminal Shortcuts
- `Ctrl+C`: Cancel current command
- `Ctrl+L`: Clear terminal
- `Ctrl+R`: Search command history
- `Ctrl+A`: Move to beginning of line
- `Ctrl+E`: Move to end of line
- `Ctrl+U`: Delete to beginning of line
- `Ctrl+K`: Delete to end of line

### Tab Management
- `Ctrl+T`: New terminal tab
- `Ctrl+W`: Close current tab
- `Ctrl+Tab`: Next tab
- `Ctrl+Shift+Tab`: Previous tab
- `Ctrl+1..9`: Switch to tab N

## 🆘 Getting Help

### Documentation
- **Full Documentation**: [docs/development.md](docs/development.md)
- **Troubleshooting**: [docs/troubleshooting.md](docs/troubleshooting.md)
- **Theme Guide**: [docs/themes/theme-guide.md](docs/themes/theme-guide.md)

### Common Issues

#### Application Won't Start
1. Check system requirements:
   - Windows 10+ (64-bit)
   - macOS 11+ (Intel or Apple Silicon)
   - Linux with GTK3 and WebKit2GTK

2. Try running from terminal:
   ```bash
   ./aDex-UI --debug
   ```

3. Check console output for error messages

#### Terminal Not Working
1. Verify shell is available:
   ```bash
   which bash
   which zsh
   ```

2. Check permissions in settings
3. Try different shell path

#### High CPU Usage
1. Reduce monitoring update frequency
2. Close unused terminal tabs
3. Disable visual effects in settings

### Community Support
- **GitHub Issues**: [Report bugs](https://github.com/adex-ui/adex-ui/issues)
- **Discussions**: [Q&A and discussions](https://github.com/adex-ui/adex-ui/discussions)
- **Wiki**: [Additional documentation](https://github.com/adex-ui/adex-ui/wiki)

## 🎉 Next Steps

1. **Explore Features**:
   - Try different terminal applications (vim, tmux, htop)
   - Experiment with themes and customization
   - Test audio effects and keyboard layouts

2. **Advanced Configuration**:
   - Create custom themes (see Theme Guide)
   - Configure startup commands
   - Set up workspace layouts

3. **Contribute**:
   - Report bugs and suggest features
   - Submit pull requests
   - Share custom themes

4. **Stay Updated**:
   - Watch for new releases
   - Follow project announcements
   - Join community discussions

---

**Enjoy your new terminal experience!** 🚀

*For detailed information, see the full [documentation](docs/).*