// Wails v3 Service Coordinator bindings.
//
// This module is a thin re-export layer over the generated bindings in
// `frontend/bindings/`, which `wails3 generate bindings` produces by static
// analysis of the Go source. It exists so that the ~19 components and stores
// importing `~/lib/wailsjs/coordinator` keep a stable import path, and so
// that the handful of historical short aliases (GetSchemes -> GetColorSchemes)
// continue to resolve.
//
// Do not add hand-written IPC here. Under Wails v2 this file wrapped
// `window.go.coordinator.*` by hand; v3 calls generated, typed stubs instead,
// so anything new belongs in the Go service and comes across on regeneration.

export * from '../../../bindings/aDex-UI/internal/services/coordinator/servicecoordinator.js'

import {
  GetColorSchemes,
  GetColorScheme,
  CreateColorScheme,
  UpdateColorScheme,
  DeleteColorScheme,
  ValidateColorScheme,
  GetColorSchemePreview,
  GetDefaultColorScheme,
  SetDefaultColorScheme,
  GetColorSchemeConfig,
  UpdateColorSchemeConfig,
  IsStarted,
  Initialize,
  StartMonitoring,
  Shutdown,
} from '../../../bindings/aDex-UI/internal/services/coordinator/servicecoordinator.js'

// Historical short aliases. The Go methods are named *ColorScheme*; earlier
// frontend code used the shorter forms, so both spellings stay valid.
export const GetSchemes = GetColorSchemes
export const GetScheme = GetColorScheme
export const CreateScheme = CreateColorScheme
export const UpdateScheme = UpdateColorScheme
export const DeleteScheme = DeleteColorScheme
export const ValidateScheme = ValidateColorScheme
export const GetSchemePreview = GetColorSchemePreview
export const GetDefaultScheme = GetDefaultColorScheme
export const SetDefaultScheme = SetDefaultColorScheme
export const GetConfig = GetColorSchemeConfig
export const UpdateConfig = UpdateColorSchemeConfig

// Coordinator lifecycle aliases, kept for call sites that used the
// disambiguated names when the coordinator was one of several bound structs.
export const IsCoordinatorStarted = IsStarted
export const InitializeCoordinator = Initialize
export const StartCoordinatorMonitoring = StartMonitoring
export const ShutdownCoordinator = Shutdown
