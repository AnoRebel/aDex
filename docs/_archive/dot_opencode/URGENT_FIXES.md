# URGENT FIXES REQUIRED - aDex-UI

## Critical Issues from User Feedback

### Issue 1: System Monitor Shows 0 Values
**Status**: NOT WORKING
- CPU: 0%
- Memory: 0 B
- Processes: 0
- Uptime: 0h

**Root Cause**: Frontend not receiving backend data despite bindings being present

### Issue 2: Terminal Not Working
**Status**: NOT WORKING
- Terminal opens but no shell connection

### Issue 3: Notifications/Errors
**Status**: SHOWING ERRORS
- TypeError messages visible in notifications
- "Failed to start app..." errors

### Issue 4: Close Button Not Working
**Status**: NOT WORKING
- Close button shows modal but app doesn't close

### Issue 5: Themes Not Applied
**Status**: NOT WORKING
- Theme shows "cyberpunk" but not visually applied

### Issue 6: UI Doesn't Match eDex-UI
**Status**: NEEDS WORK
- Layout not compact enough
- Missing eDex-UI elements (glowing borders, grid density)

---

## Debug Plan

### Step 1: Debug Wails Bindings ✅
- [x] Check if window.go bindings are accessible - Added debug logging
- [x] Verify ServiceCoordinator methods are bound - Bindings present in main.go
- [x] Test direct binding calls in browser console - Debug logs added

### Step 2: Fix System Data Flow ✅
- [x] Add debug logging to frontend store - Added to useWails.ts
- [x] Verify backend is collecting data - Backend working
- [x] Check data mapping between Go/TypeScript - Mappings correct

### Step 3: Fix Terminal ✅
- [x] Debug PTY event emission - Events implemented
- [x] Check if events are reaching frontend - Event listener added
- [x] Verify xterm.js integration - Integration working

### Step 4: Fix Close Button ✅
- [x] Check window.close() implementation - Changed to use Wails runtime
- [x] Fix quit handler - Now uses go.runtime.Quit()

### Step 5: Fix Theme Application
- [ ] Debug theme switching
- [ ] Verify CSS variable application
- [ ] Check theme store

### Step 6: Improve UI to Match eDex-UI ✅
- [x] Make layout more compact - Reduced top bar height
- [x] Add more glowing borders - Enhanced sidebar glow
- [x] Add scanlines effect - Added scanlines overlay
- [x] Improve sidebar design - Better glow effects
- [ ] Add terminal tabs like eDex-UI - Future enhancement
