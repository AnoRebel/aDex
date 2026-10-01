# Cross-Platform Application Packaging Implementation

This document describes the comprehensive cross-platform packaging implementation for aDex, enabling native application packages for Windows, macOS, and Linux.

## Overview

The implementation provides complete packaging support using go-task (Task) with Wails v3, supporting multiple distribution formats per platform with automated build and deployment workflows.

## Platform Support

### Windows
- **Executable**: `aDex.exe` with GUI subsystem
- **NSIS Installer**: Traditional Windows installer
- **MSIX Package**: Modern Windows Store package
- **Portable ZIP**: Standalone portable version
- **Code Signing**: Support for Authenticode signing

### macOS
- **App Bundle**: `.app` package with proper structure
- **Universal Binary**: Combined ARM64 + AMD64 support
- **DMG Installer**: Disk image with license support
- **Code Signing**: Developer ID signing support
- **Notarization**: Apple notarization workflow

### Linux
- **AppImage**: Portable AppImage distribution
- **DEB Package**: Debian/Ubuntu package
- **RPM Package**: Red Hat/Fedora/CentOS package
- **AUR Package**: Arch Linux package
- **Snap Package**: Universal Linux package
- **Tarball**: Generic tarball with installer

## Build System

### Task Structure
```
Taskfile.yml (root)
├── build/Taskfile.yml (common tasks)
├── build/darwin/Taskfile.yml (macOS specific)
├── build/windows/Taskfile.yml (Windows specific)
└── build/linux/Taskfile.yml (Linux specific)
```

### Key Commands
```bash
# Build for current platform
task build

# Package for current platform
task package

# Platform-specific builds
task darwin:build
task windows:build
task linux:build

# Cross-platform packaging
task darwin:package:universal
task windows:create:msix
task linux:create:appimage

# Clean artifacts
task clean
```

## Build Configuration

### Environment Variables
- `VERSION`: Application version (default: 1.0.0)
- `PRODUCTION`: Enable production optimizations (default: false)
- `BUILD_DATE`: Build timestamp
- `COMMIT`: Git commit SHA
- `COMPANY_NAME`: Company name (default: HackEAC)
- `DESCRIPTION`: Application description

### Code Signing (Optional)
- **macOS**: `CODESIGN_IDENTITY`, `APPLE_ID`, `APPLE_PASSWORD`, `APPLE_TEAM_ID`
- **Windows**: `CERT_PATH`, `CERT_PASSWORD`

## Testing

### Build Tests
- Cross-platform compilation verification
- Binary validation and execution tests
- Performance benchmarking
- Asset integrity checks

### Installation Tests
- Platform-specific installation verification
- Desktop integration testing
- Uninstallation validation
- Launch and execution verification

### Test Commands
```bash
# Run all tests
go test ./tests/backend/...

# Run specific test suites
go test ./tests/backend/build_test.go
go test ./tests/backend/install_test.go

# Run with coverage
go test -cover ./tests/backend/...
```

## Continuous Integration

### GitHub Actions Workflow
The `.github/workflows/build.yml` provides:

- **Multi-platform builds** on Ubuntu, Windows, and macOS
- **Cross-compilation** for all supported architectures
- **Automated testing** and security scanning
- **Artifact management** and release automation
- **Docker image** building and publishing

### Workflow Triggers
- Push to main/develop branches
- Pull requests
- Tagged releases (v*)
- Manual workflow dispatch

## Build Assets

### Icons and Metadata
- **Source Icon**: `build/appicon.png` (512x512 PNG)
- **Windows Icon**: `build/windows/icon.ico` (auto-generated)
- **macOS Icon**: `build/darwin/icons.icns` (auto-generated)
- **Desktop Entry**: `build/linux/aDex.desktop`

### Configuration Files
- **Build Config**: `build/config.yml` (Wails configuration)
- **Windows Info**: `build/windows/info.json` (executable metadata)
- **Windows Manifest**: `build/windows/wails.exe.manifest`
- **macOS Info**: `build/darwin/Info.plist` (app bundle metadata)
- **Linux Package**: `build/linux/nfpm/nfpm.yaml` (package metadata)

## Distribution Packages

### File Naming Convention
```
aDex-{VERSION}-{PLATFORM}-{ARCH}.{FORMAT}
Examples:
- aDex-1.0.0-windows-amd64.exe
- aDex-1.0.0-darwin-universal.dmg
- aDex-1.0.0-linux-amd64.AppImage
- aDex-1.0.0-linux-amd64.deb
```

### Package Contents
- **Windows**: Executable, icon, uninstaller, dependencies
- **macOS**: App bundle with executable, resources, Info.plist
- **Linux**: Binary, desktop file, icon, dependencies, installation scripts

## Installation

### Windows
```bash
# NSIS Installer
aDex-1.0.0-windows-amd64-setup.exe

# MSIX Package (Windows 10/11)
aDex-1.0.0-windows-amd64.msix

# Portable ZIP
aDex-1.0.0-windows-amd64-portable.zip
```

### macOS
```bash
# App Bundle
aDex-1.0.0-darwin-universal.app

# DMG Installer
aDex-1.0.0-darwin-universal.dmg
```

### Linux
```bash
# AppImage (Universal)
./aDex-1.0.0-linux-amd64.AppImage

# DEB Package (Debian/Ubuntu)
sudo dpkg -i aDex_1.0.0_amd64.deb

# RPM Package (Red Hat/Fedora)
sudo rpm -i aDex-1.0.0-1.x86_64.rpm

# Snap Package (Universal)
sudo snap install aDex_1.0.0_amd64.snap

# Tarball (Generic)
tar -xzf aDex-1.0.0-linux-x86_64.tar.gz
cd aDex-1.0.0
sudo ./install.sh
```

## Dependencies

### System Dependencies
- **Go**: 1.23+
- **Node.js**: 22+
- **Task**: go-task latest
- **Wails3**: v3.0.0-alpha.36+

### Build Dependencies
- **Linux**: libgtk-3-dev, libwebkit2gtk-4.1-dev
- **macOS**: Xcode Command Line Tools
- **Windows**: Visual Studio Build Tools

### Runtime Dependencies
- **Linux**: libgtk-3-0, libwebkit2gtk-4.1-0
- **macOS**: macOS 10.15+
- **Windows**: Windows 10+ with WebView2

## Troubleshooting

### Common Issues

1. **Build Failures**
   - Ensure all dependencies are installed
   - Check Go and Node.js versions
   - Verify Wails3 installation

2. **Code Signing Errors**
   - Verify certificate paths and passwords
   - Check certificate validity
   - Ensure proper code signing tools are installed

3. **Package Installation Issues**
   - Check system compatibility
   - Verify dependencies are installed
   - Check file permissions

### Debug Commands
```bash
# Enable verbose logging
task build --verbose

# Check build environment
task --list-all

# Test individual components
go test -v ./tests/backend/build_test.go
```

## Maintenance

### Version Updates
1. Update version in `build/config.yml`
2. Update version in package metadata files
3. Tag release with semantic versioning
4. Run full test suite before release

### Platform Support
- Regularly test on all supported platforms
- Update dependencies to latest stable versions
- Monitor for platform-specific build issues
- Update CI workflows as needed

## Future Enhancements

### Potential Improvements
- **Auto-updater**: Built-in update mechanism
- **Package Signing**: Enhanced security features
- **Cloud Build**: Remote build service integration
- **More Formats**: Additional package formats
- **Optimization**: Build size and performance optimization

This implementation provides a robust, comprehensive packaging solution for aDex across all major desktop platforms with automated builds, testing, and deployment workflows.