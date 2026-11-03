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

## In-App Updates

The app checks the update manifest attached to the latest stable release
(`releases/latest/download/update-manifest.json`) and installs only artifacts
whose ed25519 signature verifies against `build/adex.key.pub`, which is
embedded in the binary.

### What updates in place

| Platform | In-app artifact | Updates in place when |
|----------|-----------------|-----------------------|
| macOS | `adex-darwin-<arch>.app.zip` | running from a `.app` bundle whose folder is writable |
| Windows | `adex-windows-<arch>.exe` | the install folder is writable (the NSIS installer installs per user, to `%LOCALAPPDATA%\Programs\aDex`) |
| Linux | `*.AppImage` | running from an AppImage in a writable folder |

Everything else (`.deb`, `.rpm`, pacman packages, the raw Linux binary, and
Linux architectures without an AppImage) is told about new releases and gets
a link to the release page. The NSIS installer is never used for in-app
updates.

### Release signing

The `release` job in `.github/workflows/build.yml` runs
`scripts/release-manifest.sh`, which:

1. Picks one in-app artifact per platform/arch (see the table above; Linux
   arches without an AppImage get the raw binary so their users are still
   notified).
2. Runs `wails3 updater manifest -version <tag without v> -key <key>
   -url-prefix https://github.com/AnoRebel/aDex/releases/download/<tag>/
   -output update-manifest.json` over them.
3. Runs `wails3 updater verify -manifest update-manifest.json -publickey
   build/adex.key.pub`, and fails if any built platform/arch is missing from
   the manifest or listed twice.
4. Writes `SHA256SUMS` over every asset, including the manifest.

The private key comes from the `ADEX_UPDATER_KEY` repository secret. The job
fails if it is missing, writes it to a `0600` temp file, and deletes it when
the step ends. The key itself is never committed (`adex.key` is gitignored).

To test the script locally, run it against a directory of release assets with
a throwaway key:

```bash
wails3 updater genkey -out /tmp/test.key
scripts/release-manifest.sh v1.2.3 ./assets /tmp/test.key /tmp/test.key.pub
```

### Trying an update locally

`ADEX_UPDATE_MANIFEST` reads the manifest from another address. It changes
only where the manifest comes from: what it offers must still be signed with
the key the app carries.

1. Set `info.version` in `build/config.yml` lower (say `0.9.0`), run
   `wails3 task linux:create:appimage`, and keep the AppImage somewhere
   writable. Restore the version and build again.
2. Sign a manifest for the second build and serve it:

   ```bash
   cp bin/adex-x86_64.AppImage serve/adex-linux-amd64.AppImage
   wails3 updater manifest -version 1.0.0 -name "aDex 1.0.0" -key adex.key \
     -url-prefix http://127.0.0.1:8491/ -output serve/update-manifest.json \
     serve/adex-linux-amd64.AppImage
   python3 -m http.server 8491 -d serve
   ```

3. Run the first AppImage with
   `ADEX_UPDATE_MANIFEST=http://127.0.0.1:8491/update-manifest.json`.
   Settings → Updates → CHECK NOW then finds, downloads, verifies, installs
   and restarts into it.

### Key rotation

Every shipped copy trusts only the key it was built with, so losing
`adex.key` means no existing install can verify another update. Keep an
offline backup. To rotate:

1. Generate a new pair: `wails3 updater genkey -out adex-new.key`.
2. Rename the current `build/adex.key.pub` to `build/adex.key.prev.pub` and
   commit the new public key as `build/adex.key.pub`. Release once with
   `ADEX_UPDATER_KEY` unchanged: that release is signed with the old key, so
   existing installs accept it, and it embeds the new key. While
   `adex.key.prev.pub` exists, CI verifies the manifest against it.
3. After that release, replace the secret with the new private key
   (`gh secret set ADEX_UPDATER_KEY < adex-new.key`) and delete
   `build/adex.key.prev.pub`.

Copies that skip the bridging release need one manual download.

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
- **Package Signing**: Enhanced security features
- **Cloud Build**: Remote build service integration
- **More Formats**: Additional package formats
- **Optimization**: Build size and performance optimization

This implementation provides a robust, comprehensive packaging solution for aDex across all major desktop platforms with automated builds, testing, and deployment workflows.