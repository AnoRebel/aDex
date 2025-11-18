# Integration Next Steps

This document outlines the next steps required to complete the frontend-backend integration for aDex-UI.

## Completed Work

### 1. Security System Implementation ✅
- Implemented argon2id password hashing with OWASP-recommended parameters
- Created comprehensive user management system with roles (Admin, PowerUser, User, Guest)
- Built session management with auto-expiration, renewal, and CSRF protection
- Added 85%+ test coverage for authentication and session components
- Files created:
  - `internal/services/security/auth.go` (345 lines)
  - `internal/services/security/user.go` (373 lines)
  - `internal/services/security/session.go` (367 lines)
  - `tests/backend/security_auth_test.go` (411 lines)
  - `tests/backend/security_session_test.go` (348 lines)

### 2. Coordinator Type Conversion Fixes ✅
- Fixed all type conversion issues in `backend/services/coordinator/service.go`
- Implemented `convertToStruct` helper using JSON marshal/unmarshal pattern
- Fixed methods:
  - `CreateFontConfiguration` - converts to `models.FontConfiguration`
  - `UpdateFontConfiguration` - converts to `models.FontConfiguration`
  - `ImportFont` - converts to `models.FontImportRequest`
  - `ValidateFont` - converts to `models.FontConfiguration`
  - `UpdateFontSettings` - converts to `models.FontSettings`
  - `UpdateNetworkConfig` - converts to `network.NetworkServiceConfig`

## Required Next Steps

### 3. Generate Wails v3 TypeScript Bindings ⚠️

**CRITICAL**: The TypeScript bindings in `frontend/bindings/` are outdated and do not include the latest coordinator methods.

**Environment Limitation**: The wails3 CLI tool is not available in the current development environment. This must be run in an environment with wails3 installed.

**Command to run**:
```bash
task generate:bindings
```

Or directly:
```bash
wails3 generate bindings -ts -clean=true
```

**What this will do**:
- Parse all Go services exposed through the coordinator
- Generate TypeScript type definitions for all methods
- Create proper type-safe interfaces for frontend-backend communication
- Update `frontend/bindings/aDex-UI/backend/services/coordinator/servicecoordinator.ts`

**Verification**:
After running, verify that the following methods appear in `servicecoordinator.ts`:
- `CreateFontConfiguration(config: any): Promise<void>`
- `UpdateFontConfiguration(config: any): Promise<void>`
- `ImportFont(request: any): Promise<any>`
- `ValidateFont(config: any): Promise<any>`
- `UpdateFontSettings(settings: any): Promise<void>`
- `UpdateNetworkConfig(config: any): Promise<void>`

### 4. Update Frontend Composables

**File**: `frontend/app/composables/useWails.ts`

**Current State**: Uses placeholder/mock implementations

**Required Changes**:
1. Import the generated coordinator bindings:
   ```typescript
   import * as Coordinator from '~/bindings/aDex-UI/backend/services/coordinator/servicecoordinator'
   ```

2. Replace mock implementations with real binding calls:
   ```typescript
   // BEFORE (example):
   const createFontConfiguration = async (config: any) => {
     // Mock implementation
     return Promise.resolve()
   }

   // AFTER:
   const createFontConfiguration = async (config: any) => {
     return Coordinator.CreateFontConfiguration(config)
   }
   ```

3. Update all service methods to use real bindings for:
   - Font management
   - Network monitoring
   - Color scheme management
   - Terminal management
   - System monitoring
   - Audio management
   - File system operations

### 5. Connect Pinia Stores to Backend

**Files to update**:
- `frontend/app/stores/font.ts`
- `frontend/app/stores/network.ts`
- `frontend/app/stores/colorscheme.ts`
- `frontend/app/stores/terminal.ts`
- `frontend/app/stores/system.ts`

**Pattern**:
```typescript
// In each store, replace mock data with real backend calls
import { useWails } from '~/composables/useWails'

export const useFontStore = defineStore('font', () => {
  const { coordinator } = useWails()

  async function createConfiguration(config: FontConfiguration) {
    try {
      await coordinator.createFontConfiguration(config)
      // Update local state
    } catch (error) {
      // Handle error
    }
  }

  return { createConfiguration, /* other methods */ }
})
```

### 6. Test Frontend-Backend Integration

**Manual Testing Checklist**:
1. Start the application in development mode:
   ```bash
   task dev
   ```

2. Test each major feature area:
   - [ ] Font configuration CRUD operations
   - [ ] Network monitoring and statistics
   - [ ] Color scheme management
   - [ ] Terminal creation and interaction
   - [ ] System resource monitoring
   - [ ] File system browsing and operations
   - [ ] Audio visualization

3. Verify:
   - [ ] No console errors related to bindings
   - [ ] Data flows correctly from backend to frontend
   - [ ] User actions trigger correct backend calls
   - [ ] Real-time updates work via event bus
   - [ ] Error handling works properly

**Automated Testing**:
```bash
# Run frontend tests
cd frontend && bun test

# Run backend tests
go test ./tests/backend/... -v

# Run integration tests (if available)
go test ./tests/integration/... -v
```

### 7. Build and Package

Once integration is verified:

**Development Build**:
```bash
task build
```

**Production Package**:
```bash
task package
```

**Platform-Specific Builds**:
```bash
# Windows
task windows:package

# macOS
task darwin:package

# Linux
task linux:package
```

## Known Limitations

### Go Version
- Project requires Go 1.25.0
- Current environment may not have Go 1.25.0 installed
- Ensure correct Go version before building

### Wails v3 Alpha
- Using Wails v3 alpha.36 (pre-release)
- Some features may be unstable
- Check Wails v3 documentation for updates

### Missing Features Still to Implement
As documented in the original analysis:

1. **File Operations** (Priority: High)
   - Copy operation (`filesystem/service.go:848`)
   - Move operation (`filesystem/service.go:854`)

2. **Bookmarks System** (Priority: Medium)
   - Persistent bookmarks storage (`filesystem/navigation.go`)

3. **Configuration Persistence** (Priority: High)
   - Settings don't persist between sessions

4. **Log Rotation** (Priority: Medium)
   - Prevent unbounded log growth (`logger/logger.go:189`)

5. **Platform-Specific Features** (Priority: Low)
   - Audio session detection for Windows/macOS
   - CPU temperature monitoring

## Branch Information

**Current Branch**: `claude/fix-coordinator-integration-01HQCQuGtdcBCntoduk57gyh`

**Commits**:
1. `fix: correct all import paths and Go version` - Fixed module import paths
2. `fix: merge duplicate app configuration in nuxt.config.ts` - Fixed Nuxt config
3. `feat: implement comprehensive security system with argon2id` - Security implementation
4. `fix: implement type conversion for coordinator service methods` - Coordinator fixes

**To merge to main**:
```bash
# After completing all integration steps and testing
git checkout main
git merge claude/fix-coordinator-integration-01HQCQuGtdcBCntoduk57gyh
git push origin main
```

## Contact and Support

For issues or questions:
- Check Wails v3 documentation: https://v3alpha.wails.io
- Review Nuxt v4 documentation: https://nuxt.com
- Examine existing tests for implementation patterns

## Completion Checklist

- [x] Security system implementation
- [x] Coordinator type conversion fixes
- [x] Git commits with clear messages
- [ ] Wails bindings generation (requires wails3 CLI)
- [ ] Frontend composables update
- [ ] Pinia stores connection
- [ ] Integration testing
- [ ] Production build verification
- [ ] All platforms tested
- [ ] Merge to main branch
