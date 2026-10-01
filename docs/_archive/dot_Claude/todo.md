# Mission: Complete aDex-UI Application Fixes - Phase 2

## M1: UI/UX Fixes | status: completed
- [x] T1.1: Add missing `toggleWindow` function in index.vue
- [x] T1.2: Add missing `handleQuit` function in index.vue
- [x] T1.3: Fix TypeScript build errors
- [x] T1.4: Build verification (frontend + backend)

## M2: Component Fixes | status: in_progress

### T2.1: Fix SystemMonitor Component | agent:Worker | status: in_progress
- [x] S2.1.1: Simplify SystemMonitor to work without child components | size:M
- [x] S2.1.2: Remove problematic CpuChart, MemoryChart, ProcessList dependencies | size:S
- [x] S2.1.3: Connect to systemStore data properly | size:S

### T2.2: Fix NetworkMonitor Component | agent:Worker | status: pending
- [x] S2.2.1: Verify NetworkMonitor renders with demo data | size:S
- [x] S2.2.2: Add missing CSS variable fallbacks | size:S

### T2.3: Fix FileBrowser Component | agent:Worker | status: pending
- [x] S2.3.1: Verify FileBrowser renders with mock data | size:S
- [x] S2.3.2: Add global declaration for window._keyboardNavInterval | size:S

### T2.4: Fix Settings Modal | agent:Worker | status: pending
- [x] S2.4.1: Debug Settings button click handler | size:S
- [x] S2.4.2: Add Theme tab to Settings Modal | size:M
- [x] S2.4.3: Connect theme selection to ThemeStore | size:S

### T2.5: Theme System Integration | agent:Worker | status: pending
- [x] S2.5.1: Fix CSS variable names in ThemeManager | size:M
- [x] S2.5.2: Add global theme application mechanism | size:S
- [x] S2.5.3: Add theme persistence | size:S

## M3: Final Verification | agent:Reviewer | status: pending
- [ ] T3.1: Run `bun run build` to verify no build errors
- [ ] T3.2: Run `wails dev` to verify functionality
- [ ] T3.3: Test all windows open and display data

## Summary
Phase 2 focuses on making all component windows functional:
1. SystemMonitor - needs simplified implementation
2. NetworkMonitor - already works with demo data  
3. FileBrowser - already works with mock data
4. Settings Modal - needs visibility fix and theme integration
5. Theme System - needs CSS variable mapping
