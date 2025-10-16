# aDex-UI - Advanced Desktop Experience UI

A modern desktop environment built with Wails v3 and Nuxt v4, providing a comprehensive desktop experience with terminal services, system monitoring, file management, theming, and audio control.

## Git Workflow

This project follows a Git Flow-inspired workflow:

### Branch Structure
- **main**: Production-ready code (protected branch)
- **develop**: Integration branch for all features
- **feature/***: Individual feature development branches

### Feature Branches
- `feature/terminal-service`: Terminal implementation and management
- `feature/system-monitoring`: System resource monitoring and visualization
- `feature/filesystem`: File browser and management system
- `feature/theme-system`: Theme management and customization
- `feature/audio-system`: Audio device management and control

### Development Process
1. Create feature branches from `develop`
2. Develop and test features in isolation
3. Merge completed features back to `develop`
4. Release candidates are merged from `develop` to `main`
5. Tags are created on `main` for releases

### Branch Protection
- `main` branch is protected and should only receive merges from `develop`
- All feature branches should be deleted after successful merge
- Use descriptive commit messages following conventional commits

## Project Structure

This project uses:
- **Backend**: Wails v3 (Go + Web technologies)
- **Frontend**: Nuxt v4 (Vue 3 ecosystem)
- **Styling**: Modern CSS with theme support
- **Build**: Cross-platform desktop application

## Getting Started

1. Clone the repository
2. Install dependencies (see respective documentation)
3. Run development server
4. Follow the feature-specific documentation for each module

## Development

See the steering documents in `/steering-docs/` for detailed implementation plans and architectural decisions.
