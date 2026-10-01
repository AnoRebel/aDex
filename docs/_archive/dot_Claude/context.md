# Project Context

## Environment
- Language: TypeScript + Go
- Runtime: Node.js (Bun) + Go 1.21+
- Build: Wails v2.11.0
- Test: Not yet configured
- Package Manager: Bun (frontend), Go modules (backend)

## Project Type
- [x] Application (Desktop)
- Wails v2 desktop app with Go backend and Nuxt 4 frontend

## Infrastructure
- Container: None
- Orchestration: None
- CI/CD: GitHub Actions (workflows exist but not reviewed)
- Cloud: None

## Structure
- Source: `/frontend/app/` (Vue/Nuxt), `/backend/` (Go)
- Tests: `/tests/` (exists but minimal)
- Docs: `/docs/`
- Entry: `main.go`

## Key Files
- `main.go` - Wails app entry, ServiceCoordinator binding
- `frontend/app/pages/index.vue` - Main desktop UI
- `frontend/app/components/` - All UI components
- `frontend/app/stores/` - Pinia stores

## Conventions
- Naming: camelCase (TS), PascalCase (Vue components)
- Imports: Nuxt auto-imports, `~` alias
- Error handling: try/catch with console.error

## Commands
```bash
# Development
wails dev

# Build
go build -o bin/aDex-UI
cd frontend && bun run generate
```

## Notes
- Wails v2 (not v3) - uses different runtime API
- Frontend at http://localhost:3000 during dev
- ServiceCoordinator provides system/terminal/file APIs
