# aDex-UI Troubleshooting Guide

This guide helps diagnose and resolve common issues with aDex-UI.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Build Problems](#build-problems)
- [Runtime Issues](#runtime-issues)
- [Terminal Problems](#terminal-problems)
- [System Monitoring Issues](#system-monitoring-issues)
- [Network Connectivity Problems](#network-connectivity-problems)
- [File Browser Issues](#file-browser-issues)
- [Theme and Display Problems](#theme-and-display-problems)
- [Audio Issues](#audio-issues)
- [Keyboard Issues](#keyboard-issues)
- [Performance Problems](#performance-problems)
- [Cross-Platform Issues](#cross-platform-issues)
- [Development Issues](#development-issues)

## Installation Issues

### Prerequisites Not Found

**Problem**: `go: command not found` or `node: command not found`

**Solution**:
1. **Install Go 1.21+**:
   ```bash
   # macOS
   brew install go

   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install golang-go

   # Windows
   # Download from https://golang.org/dl/
   ```

2. **Install Node.js 18+**:
   ```bash
   # macOS
   brew install node

   # Ubuntu/Debian
   curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
   sudo apt-get install -y nodejs

   # Windows
   # Download from https://nodejs.org/
   ```

3. **Install Task runner**:
   ```bash
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

### Dependency Installation Fails

**Problem**: `go mod download` or `bun install` fails

**Solution**:
1. **Check network connection**:
   ```bash
   curl -I https://proxy.golang.org
   curl -I https://registry.npmjs.org
   ```

2. **Clear Go module cache**:
   ```bash
   go clean -modcache
   go mod download
   ```

3. **Clear Node.js cache**:
   ```bash
   cd frontend
   rm -rf node_modules bun.lockb
   bun install
   ```

4. **Use alternative registry** (for Node.js):
   ```bash
   bun config set registry https://registry.npmjs.org/
   ```

## Build Problems

### Build Fails on macOS

**Problem**: `ld: library not found` or code signing errors

**Solution**:
1. **Install Xcode Command Line Tools**:
   ```bash
   xcode-select --install
   ```

2. **Set proper environment variables**:
   ```bash
   export CGO_CFLAGS="-mmacosx-version-min=10.13"
   export CGO_LDFLAGS="-mmacosx-version-min=10.13"
   ```

3. **Code signing issues**:
   ```bash
   # Create a self-signed certificate if needed
   security create-keychain -p "" build.keychain
   security default-keychain -s build.keychain
   ```

### Build Fails on Windows

**Problem**: `gcc: command not found` or MinGW issues

**Solution**:
1. **Install TDM-GCC or MinGW-w64**:
   - Download from https://jmeubank.github.io/tdm-gcc/
   - Add to system PATH

2. **Use Windows Subsystem for Linux (WSL)**:
   ```bash
   wsl --install
   ```

3. **Set Go environment for Windows**:
   ```bash
   set GOOS=windows
   set GOARCH=amd64
   set CGO_ENABLED=1
   ```

### Build Fails on Linux

**Problem**: Missing system libraries

**Solution**:
1. **Install build essentials**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

   # Fedora
   sudo dnf install gcc gcc-c++ gtk3-devel webkit2gtk4.0-devel

   # Arch Linux
   sudo pacman -S base-devel gtk3 webkit2gtk
   ```

2. **Install additional dependencies**:
   ```bash
   # For audio support
   sudo apt-get install libasound2-dev

   # For system monitoring
   sudo apt-get install libprocps-dev
   ```

## Runtime Issues

### Application Won't Start

**Problem**: Double-clicking the app does nothing

**Solution**:
1. **Check console output**:
   ```bash
   # macOS
   /Applications/aDex-UI.app/Contents/MacOS/aDex-UI

   # Windows
   ./aDex-UI.exe

   # Linux
   ./aDex-UI
   ```

2. **Check permissions**:
   ```bash
   # macOS
   sudo chmod +x /Applications/aDex-UI.app/Contents/MacOS/aDex-UI

   # Linux
   chmod +x ./aDex-UI
   ```

3. **Check antivirus software** (Windows):
   - Add aDex-UI to antivirus exclusions
   - Run as administrator if needed

### Application Crashes on Startup

**Problem**: App opens briefly then crashes

**Solution**:
1. **Check system requirements**:
   - macOS 10.15+ (Catalina)
   - Windows 10+ (64-bit)
   - Linux with GTK3 and WebKit2GTK

2. **Reset configuration**:
   ```bash
   # macOS
   rm -rf ~/Library/Application\ Support/aDex-UI

   # Windows
   rd /s "%APPDATA%\aDex-UI"

   # Linux
   rm -rf ~/.config/aDex-UI
   ```

3. **Check for conflicting applications**:
   - Close other terminal emulators
   - Disable other system monitoring tools

## Terminal Problems

### Terminal Not Responding

**Problem**: Terminal shows blank screen or freezes

**Solution**:
1. **Check shell availability**:
   ```bash
   which bash
   which zsh
   echo $SHELL
   ```

2. **Reset terminal settings**:
   - Delete terminal session files
   - Reset shell configuration

3. **Check PTY permissions**:
   ```bash
   # Linux
   ls -la /dev/pts/
   sudo chmod 666 /dev/pts/*
   ```

### Terminal Commands Not Working

**Problem**: Commands don't execute or show no output

**Solution**:
1. **Check PATH environment**:
   ```bash
   echo $PATH
   which ls
   ```

2. **Reset terminal environment**:
   ```bash
   export PATH=/usr/local/bin:/usr/bin:/bin
   export TERM=xterm-256color
   ```

3. **Check for command aliases**:
   ```bash
   alias
   unalias ls  # Reset problematic aliases
   ```

### Terminal Colors Not Displaying

**Problem**: Terminal shows only black and white text

**Solution**:
1. **Check TERM variable**:
   ```bash
   echo $TERM
   export TERM=xterm-256color
   ```

2. **Test color support**:
   ```bash
   echo $'\e[31mRed\e[0m'
   ```

3. **Reset color scheme**:
   - Use default terminal theme
   - Check color scheme settings in preferences

## System Monitoring Issues

### System Metrics Not Updating

**Problem**: CPU, memory, or disk usage shows 0% or old values

**Solution**:
1. **Check permissions**:
   ```bash
   # macOS
   osascript -e 'tell application "System Events" to get name of processes'

   # Linux
   cat /proc/meminfo
   cat /proc/stat
   ```

2. **Restart monitoring service**:
   - Close and reopen the application
   - Check for system monitoring conflicts

3. **Check system tools**:
   ```bash
   # macOS
   top -l 1

   # Linux
   top -bn1
   df -h
   free -h
   ```

### High CPU Usage

**Problem**: aDex-UI uses excessive CPU resources

**Solution**:
1. **Check update intervals**:
   - Increase monitoring update frequency
   - Reduce data collection frequency

2. **Profile CPU usage**:
   ```bash
   # macOS
   sample aDex-UI 10 -file cpu_profile.txt

   # Linux
   perf record -p $(pgrep aDex-UI) -- sleep 10
   perf report
   ```

3. **Close unused terminals**:
   - Each terminal session consumes resources
   - Limit concurrent terminal sessions

## Network Connectivity Problems

### Network Interface Not Detected

**Problem**: Network monitor shows no interfaces

**Solution**:
1. **Check network interfaces**:
   ```bash
   ip addr show
   ifconfig -a
   ```

2. **Check network permissions**:
   ```bash
   # Linux
   sudo chmod 644 /proc/net/dev

   # macOS
   sudo ifconfig
   ```

3. **Restart network service**:
   ```bash
   sudo systemctl restart network-manager
   ```

### Incorrect Network Statistics

**Problem**: Transfer rates seem wrong or don't update

**Solution**:
1. **Reset network counters**:
   ```bash
   sudo ip -s link show
   ```

2. **Check for multiple interfaces**:
   - Disable VPN or virtual network adapters
   - Select correct interface in settings

3. **Calibrate network monitoring**:
   - Compare with system tools
   - Adjust update intervals

## File Browser Issues

### File Browser Not Following Terminal CWD

**Problem**: File browser shows different directory than terminal

**Solution**:
1. **Check CWD tracking**:
   ```bash
   # Linux/macOS
   echo $PWD

   # Test with simple command
   cd /tmp
   echo $PWD
   ```

2. **Enable CWD tracking in settings**:
   - Check settings enable "Follow Terminal CWD"
   - Restart terminal session

3. **Manual refresh**:
   - Use refresh button in file browser
   - Press F5 to refresh

### File Operations Fail

**Problem**: Copy, move, or delete operations don't work

**Solution**:
1. **Check file permissions**:
   ```bash
   ls -la
   stat filename
   ```

2. **Check available space**:
   ```bash
   df -h
   ```

3. **Check for locked files**:
   ```bash
   lsof | grep filename
   ```

## Theme and Display Problems

### Theme Not Applying

**Problem**: Selected theme doesn't change appearance

**Solution**:
1. **Check theme format**:
   - Ensure theme file is valid JSON
   - Verify required fields are present

2. **Reset theme cache**:
   ```bash
   rm -rf ~/.config/aDex-UI/themes
   ```

3. **Check CSS variables**:
   - Open developer tools
   - Inspect CSS variables in :root

### Display Scaling Issues

**Problem**: UI elements too large or too small

**Solution**:
1. **Check system DPI settings**:
   - macOS: System Preferences → Display
   - Windows: Display settings → Scale
   - Linux: Display settings in desktop environment

2. **Adjust application scaling**:
   - Use application zoom controls
   - Check font size settings

3. **Reset display settings**:
   - Reset application preferences
   - Use default theme

## Audio Issues

### No Sound Effects

**Problem**: Audio effects don't play when enabled

**Solution**:
1. **Check audio system**:
   ```bash
   # macOS
   afplay /System/Library/Sounds/Ping.aiff

   # Linux
   aplay /usr/share/sounds/alsa/Front_Left.wav

   # Windows
   # Test in Sound settings
   ```

2. **Check application settings**:
   - Verify audio is enabled
   - Check master volume
   - Verify individual event settings

3. **Check audio files**:
   - Verify audio files exist in correct location
   - Check file permissions
   - Test audio file format compatibility

### Audio Distortion or Cracking

**Problem**: Audio sounds distorted or has noise

**Solution**:
1. **Check audio format**:
   - Use WAV or MP3 format
   - Verify audio bit rate and sample rate

2. **Adjust audio settings**:
   - Reduce volume levels
   - Check system audio settings

3. **Check for conflicts**:
   - Close other audio applications
   - Disable system sounds temporarily

## Keyboard Issues

### On-Screen Keyboard Not Appearing

**Problem**: Keyboard doesn't show when clicking input fields

**Solution**:
1. **Enable on-screen keyboard**:
   - Check keyboard settings in preferences
   - Verify "Show on-screen keyboard" is enabled

2. **Check input field focus**:
   - Click input field to focus
   - Tab into input field

3. **Check touch device support**:
   - Verify touch events are working
   - Check device compatibility

### Keyboard Input Not Working

**Problem**: Keys don't type when pressed

**Solution**:
1. **Check keyboard layout**:
   - Verify correct layout is selected
   - Try different keyboard layouts

2. **Check input field compatibility**:
   - Test with different input types
   - Verify field accepts text input

3. **Reset keyboard settings**:
   - Restore default keyboard settings
   - Check for conflicting key bindings

## Performance Problems

### Slow Application Startup

**Problem**: Application takes too long to open

**Solution**:
1. **Check system resources**:
   - Close other applications
   - Check available memory

2. **Disable heavy features**:
   - Disable audio effects
   - Reduce system monitoring frequency

3. **Check for corrupted data**:
   - Reset application settings
   - Clear cache files

### Memory Usage Too High

**Problem**: Application uses excessive memory

**Solution**:
1. **Monitor memory usage**:
   ```bash
   # macOS
   top -pid $(pgrep aDex-UI)

   # Linux
   ps aux | grep aDex-UI
   ```

2. **Reduce terminal sessions**:
   - Close unused terminal tabs
   - Limit terminal history

3. **Check for memory leaks**:
   - Monitor memory over time
   - Restart application periodically

## Cross-Platform Issues

### Features Not Working on Specific Platform

**Problem**: Feature works on one platform but not others

**Solution**:
1. **Check platform-specific code**:
   - Review build tags in Go code
   - Check platform-specific CSS

2. **Test with debug builds**:
   - Enable debug logging
   - Check console output

3. **Report platform-specific bugs**:
   - Include platform details in bug report
   - Provide system information

### Package Installation Fails

**Problem**: Application installer doesn't work

**Solution**:
1. **Check system requirements**:
   - Verify OS version compatibility
   - Check required system libraries

2. **Try manual installation**:
   - Download binary directly
   - Extract and run manually

3. **Check security settings**:
   - Allow applications from unknown developers
   - Disable Gatekeeper temporarily (macOS)

## Development Issues

### Hot Reload Not Working

**Problem**: Changes don't appear when developing

**Solution**:
1. **Check development server**:
   ```bash
   task dev:frontend
   task dev:backend
   ```

2. **Clear build cache**:
   ```bash
   rm -rf frontend/.nuxt
   rm -rf frontend/.output
   ```

3. **Check for syntax errors**:
   - Run TypeScript compiler
   - Check Go build errors

### Test Failures

**Problem**: Tests pass locally but fail in CI

**Solution**:
1. **Check test environment**:
   - Verify Node.js and Go versions
   - Check system dependencies

2. **Check test timing**:
   - Increase test timeouts
   - Add explicit waits

3. **Check test isolation**:
   - Ensure tests don't share state
   - Clean up test data properly

### Build Errors in CI

**Problem**: Build passes locally but fails in CI

**Solution**:
1. **Check build environment**:
   - Verify build dependencies
   - Check environment variables

2. **Check cross-compilation**:
   - Verify build targets
   - Check platform-specific dependencies

3. **Debug build process**:
   - Enable verbose logging
   - Check build logs

---

## Getting Additional Help

If you're still experiencing issues:

1. **Check existing issues**: Look through GitHub issues for similar problems
2. **Search documentation**: Review the full documentation for related information
3. **Enable debug mode**: Run with debug flags for more detailed output
4. **Create detailed bug report**: Include system information, error logs, and reproduction steps
5. **Join community discussions**: Ask questions in community forums or Discord

### Debug Mode

Enable debug mode for detailed logging:

```bash
# Environment variable
export ADEX_DEBUG=1
./aDex-UI

# Or via build flags
go run -tags debug .
```

### System Information Collection

For bug reports, include:

```bash
# System info
uname -a
go version
node --version
bun --version

# Application info
./aDex-UI --version

# Hardware info
lscpu
free -h
df -h
```

---

Remember to check the [development guide](development.md) for regular development workflow and best practices.