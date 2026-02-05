// Wails v2 Service Coordinator Bindings
// These bindings interface with the Go backend through Wails v2 runtime

// In Wails v2, bound methods are available on window.go.{Package}.{Struct}
// The coordinator is bound in main.go, so it's available at runtime

declare global {
  interface Window {
    go?: {
      main?: {
        ServiceCoordinator?: ServiceCoordinatorBindings
      }
    }
  }
}

// ServiceCoordinator bindings interface
interface ServiceCoordinatorBindings {
  // Lifecycle
  Initialize(): Promise<void>
  StartMonitoring(): Promise<void>
  Shutdown(): Promise<void>
  IsStarted(): Promise<boolean>
  
  // Terminal Service
  CreateTerminal(width: number, height: number): Promise<any>
  WriteToTerminal(terminalID: string, data: number[]): Promise<void>
  ResizeTerminal(terminalID: string, width: number, height: number): Promise<void>
  CloseTerminal(terminalID: string): Promise<void>
  ListTerminals(): Promise<string[]>
  GetTerminalInfo(terminalID: string): Promise<any>

  // Filesystem Service
  ReadDirectory(path: string): Promise<any>
  
  // Color Scheme Service
  GetColorSchemes(): Promise<Record<string, any>>
  GetColorScheme(id: string): Promise<any>
  CreateColorScheme(scheme: any): Promise<void>
  UpdateColorScheme(scheme: any): Promise<void>
  DeleteColorScheme(id: string): Promise<void>
  SetDefaultColorScheme(id: string): Promise<void>
  GetDefaultColorScheme(): Promise<any>
  GetColorSchemeConfig(): Promise<any>
  UpdateColorSchemeConfig(config: any): Promise<void>
  GetColorSchemePreview(id: string): Promise<any>
  ValidateColorScheme(scheme: any): Promise<any>
  
  // Font Service
  GetFontConfigurations(): Promise<Record<string, any>>
  GetFontConfiguration(id: string): Promise<any>
  CreateFontConfiguration(config: any): Promise<void>
  UpdateFontConfiguration(config: any): Promise<void>
  DeleteFontConfiguration(id: string): Promise<void>
  GetSystemFonts(): Promise<Record<string, any>>
  GetMonospaceFonts(): Promise<any[]>
  ScanSystemFonts(): Promise<void>
  ImportFont(request: any): Promise<any>
  ValidateFont(config: any): Promise<any>
  GetFontMetrics(family: string, size: number): Promise<any>
  GetFontSettings(): Promise<any>
  UpdateFontSettings(settings: any): Promise<void>
  GetDefaultFontConfiguration(): Promise<any>
  SetDefaultFontConfiguration(id: string): Promise<void>
  
  // Network Service
  GetNetworkMetrics(): Promise<any>
  GetNetworkConnections(): Promise<any>
  GetBandwidthData(interfaceName: string): Promise<any>
  GetNetworkStatistics(): Promise<any>
  GetNetworkAlerts(): Promise<any>
  ResolveNetworkAlert(alertID: string): Promise<void>
  ClearNetworkAlerts(): Promise<void>
  StartNetworkMonitoring(): Promise<void>
  StopNetworkMonitoring(): Promise<void>
  GetNetworkConfig(): Promise<any>
  UpdateNetworkConfig(config: any): Promise<void>
  ResetNetworkService(): Promise<void>
  IsNetworkMonitoring(): Promise<boolean>
  
  // System Service
  GetService(serviceType: string): Promise<any>
  GetPlatform(): Promise<any>
  GetEventBus(): Promise<any>
  
  // System Metrics (real data from gopsutil)
  GetSystemInfo(): Promise<any>
  GetCPUUsage(): Promise<any>
  GetMemoryUsage(): Promise<any>
  GetDiskUsage(): Promise<any>
  GetNetworkInfo(): Promise<any>
  
  // Process Management
  GetTopProcesses(metric: string, limit: number): Promise<any>
}

// Helper to get the coordinator bindings with retry support
function getCoordinator(): ServiceCoordinatorBindings {
  const coordinator = window.go?.main?.ServiceCoordinator
  if (!coordinator) {
    throw new Error('ServiceCoordinator not available. Make sure Wails bindings are initialized.')
  }
  return coordinator
}

// Wait for coordinator bindings to be ready (with timeout)
async function waitForCoordinator(timeoutMs: number = 5000): Promise<ServiceCoordinatorBindings> {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    const coordinator = window.go?.main?.ServiceCoordinator
    if (coordinator) return coordinator
    await new Promise(resolve => setTimeout(resolve, 100))
  }
  throw new Error('ServiceCoordinator not available after timeout.')
}

// Safe wrapper that waits for bindings if needed
async function safeGetCoordinator(): Promise<ServiceCoordinatorBindings> {
  try {
    return getCoordinator()
  } catch {
    return waitForCoordinator()
  }
}

// Terminal Service Methods
export const CreateTerminal = async (width: number, height: number): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.CreateTerminal(width, height)
}

export const WriteToTerminal = async (terminalID: string, data: string | Uint8Array): Promise<void> => {
  const bytes = typeof data === 'string' ? new TextEncoder().encode(data) : data
  const c = await safeGetCoordinator()
  return c.WriteToTerminal(terminalID, Array.from(bytes))
}

export const ResizeTerminal = async (terminalID: string, width: number, height: number): Promise<void> => {
  const c = await safeGetCoordinator()
  return c.ResizeTerminal(terminalID, width, height)
}

export const CloseTerminal = async (terminalID: string): Promise<void> => {
  const c = await safeGetCoordinator()
  return c.CloseTerminal(terminalID)
}

export const ListTerminals = async (): Promise<string[]> => {
  const c = await safeGetCoordinator()
  return c.ListTerminals()
}

export const GetTerminalInfo = async (terminalID: string): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetTerminalInfo(terminalID)
}

// Filesystem Service Methods
export const ReadDirectory = async (path: string): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.ReadDirectory(path)
}

// Color Scheme Service Methods
export const GetColorSchemes = async (): Promise<Record<string, any>> => {
  return getCoordinator().GetColorSchemes()
}

export const GetColorScheme = async (id: string): Promise<any> => {
  return getCoordinator().GetColorScheme(id)
}

export const CreateColorScheme = async (scheme: any): Promise<void> => {
  return getCoordinator().CreateColorScheme(scheme)
}

export const UpdateColorScheme = async (scheme: any): Promise<void> => {
  return getCoordinator().UpdateColorScheme(scheme)
}

export const DeleteColorScheme = async (id: string): Promise<void> => {
  return getCoordinator().DeleteColorScheme(id)
}

export const SetDefaultColorScheme = async (id: string): Promise<void> => {
  return getCoordinator().SetDefaultColorScheme(id)
}

export const GetDefaultColorScheme = async (): Promise<any> => {
  return getCoordinator().GetDefaultColorScheme()
}

export const GetColorSchemeConfig = async (): Promise<any> => {
  return getCoordinator().GetColorSchemeConfig()
}

export const UpdateColorSchemeConfig = async (config: any): Promise<void> => {
  return getCoordinator().UpdateColorSchemeConfig(config)
}

export const GetColorSchemePreview = async (id: string): Promise<any> => {
  return getCoordinator().GetColorSchemePreview(id)
}

export const ValidateColorScheme = async (scheme: any): Promise<any> => {
  return getCoordinator().ValidateColorScheme(scheme)
}

// Aliases for colorscheme store compatibility
export const GetSchemes = GetColorSchemes
export const GetScheme = GetColorScheme
export const CreateScheme = CreateColorScheme
export const UpdateScheme = UpdateColorScheme
export const DeleteScheme = DeleteColorScheme
export const SetDefaultScheme = SetDefaultColorScheme
export const GetDefaultScheme = GetDefaultColorScheme
export const GetConfig = GetColorSchemeConfig
export const UpdateConfig = UpdateColorSchemeConfig
export const GetSchemePreview = GetColorSchemePreview
export const ValidateScheme = ValidateColorScheme

// Font Service Methods
export const GetFontConfigurations = async (): Promise<Record<string, any>> => {
  return getCoordinator().GetFontConfigurations()
}

export const GetFontConfiguration = async (id: string): Promise<any> => {
  return getCoordinator().GetFontConfiguration(id)
}

export const CreateFontConfiguration = async (config: any): Promise<void> => {
  return getCoordinator().CreateFontConfiguration(config)
}

export const UpdateFontConfiguration = async (config: any): Promise<void> => {
  return getCoordinator().UpdateFontConfiguration(config)
}

export const DeleteFontConfiguration = async (id: string): Promise<void> => {
  return getCoordinator().DeleteFontConfiguration(id)
}

export const GetSystemFonts = async (): Promise<Record<string, any>> => {
  return getCoordinator().GetSystemFonts()
}

export const GetMonospaceFonts = async (): Promise<any[]> => {
  return getCoordinator().GetMonospaceFonts()
}

export const ScanSystemFonts = async (): Promise<void> => {
  return getCoordinator().ScanSystemFonts()
}

export const ImportFont = async (request: any): Promise<any> => {
  return getCoordinator().ImportFont(request)
}

export const ValidateFont = async (config: any): Promise<any> => {
  return getCoordinator().ValidateFont(config)
}

export const GetFontMetrics = async (family: string, size: number): Promise<any> => {
  return getCoordinator().GetFontMetrics(family, size)
}

export const GetFontSettings = async (): Promise<any> => {
  return getCoordinator().GetFontSettings()
}

export const UpdateFontSettings = async (settings: any): Promise<void> => {
  return getCoordinator().UpdateFontSettings(settings)
}

export const GetDefaultFontConfiguration = async (): Promise<any> => {
  return getCoordinator().GetDefaultFontConfiguration()
}

export const SetDefaultFontConfiguration = async (id: string): Promise<void> => {
  return getCoordinator().SetDefaultFontConfiguration(id)
}

// Network Service Methods
export const GetNetworkMetrics = async (): Promise<any> => {
  return getCoordinator().GetNetworkMetrics()
}

export const GetNetworkConnections = async (): Promise<any> => {
  return getCoordinator().GetNetworkConnections()
}

export const GetBandwidthData = async (interfaceName: string): Promise<any> => {
  return getCoordinator().GetBandwidthData(interfaceName)
}

export const GetNetworkStatistics = async (): Promise<any> => {
  return getCoordinator().GetNetworkStatistics()
}

export const GetNetworkAlerts = async (): Promise<any> => {
  return getCoordinator().GetNetworkAlerts()
}

export const ResolveNetworkAlert = async (alertID: string): Promise<void> => {
  return getCoordinator().ResolveNetworkAlert(alertID)
}

export const ClearNetworkAlerts = async (): Promise<void> => {
  return getCoordinator().ClearNetworkAlerts()
}

export const StartNetworkMonitoring = async (): Promise<void> => {
  return getCoordinator().StartNetworkMonitoring()
}

export const StopNetworkMonitoring = async (): Promise<void> => {
  return getCoordinator().StopNetworkMonitoring()
}

export const GetNetworkConfig = async (): Promise<any> => {
  return getCoordinator().GetNetworkConfig()
}

export const UpdateNetworkConfig = async (config: any): Promise<void> => {
  return getCoordinator().UpdateNetworkConfig(config)
}

export const ResetNetworkService = async (): Promise<void> => {
  return getCoordinator().ResetNetworkService()
}

export const IsNetworkMonitoring = async (): Promise<boolean> => {
  return getCoordinator().IsNetworkMonitoring()
}

// System/Platform Methods
export const GetPlatform = async (): Promise<any> => {
  return getCoordinator().GetPlatform()
}

export const GetService = async (serviceType: string): Promise<any> => {
  return getCoordinator().GetService(serviceType)
}

// System Metrics Methods (real data from gopsutil)
// These use safeGetCoordinator() to wait for bindings to be ready
export const GetSystemInfo = async (): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetSystemInfo()
}

export const GetCPUUsage = async (): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetCPUUsage()
}

export const GetMemoryUsage = async (): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetMemoryUsage()
}

export const GetDiskUsage = async (): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetDiskUsage()
}

export const GetNetworkInfo = async (): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetNetworkInfo()
}

// Process Management Methods
export const GetTopProcesses = async (metric: string = 'cpu', limit: number = 10): Promise<any> => {
  const c = await safeGetCoordinator()
  return c.GetTopProcesses(metric, limit)
}

// Lifecycle Methods
export const InitializeCoordinator = async (): Promise<void> => {
  return getCoordinator().Initialize()
}

export const StartCoordinatorMonitoring = async (): Promise<void> => {
  return getCoordinator().StartMonitoring()
}

export const ShutdownCoordinator = async (): Promise<void> => {
  return getCoordinator().Shutdown()
}

export const IsCoordinatorStarted = async (): Promise<boolean> => {
  return getCoordinator().IsStarted()
}

// Aliases for store compatibility
export const StartMonitoring = StartNetworkMonitoring
export const StopMonitoring = StopNetworkMonitoring

// Re-export all as a namespace for compatibility
export const ServiceCoordinator = {
  // Terminal
  CreateTerminal,
  WriteToTerminal,
  ResizeTerminal,
  CloseTerminal,
  ListTerminals,
  GetTerminalInfo,
  
  // Color Scheme
  GetColorSchemes,
  GetColorScheme,
  CreateColorScheme,
  UpdateColorScheme,
  DeleteColorScheme,
  SetDefaultColorScheme,
  GetDefaultColorScheme,
  GetColorSchemeConfig,
  UpdateColorSchemeConfig,
  GetColorSchemePreview,
  ValidateColorScheme,
  
  // Font
  GetFontConfigurations,
  GetFontConfiguration,
  CreateFontConfiguration,
  UpdateFontConfiguration,
  DeleteFontConfiguration,
  GetSystemFonts,
  GetMonospaceFonts,
  ScanSystemFonts,
  ImportFont,
  ValidateFont,
  GetFontMetrics,
  GetFontSettings,
  UpdateFontSettings,
  GetDefaultFontConfiguration,
  SetDefaultFontConfiguration,
  
  // Network
  GetNetworkMetrics,
  GetNetworkConnections,
  GetBandwidthData,
  GetNetworkStatistics,
  GetNetworkAlerts,
  ResolveNetworkAlert,
  ClearNetworkAlerts,
  StartNetworkMonitoring,
  StopNetworkMonitoring,
  GetNetworkConfig,
  UpdateNetworkConfig,
  ResetNetworkService,
  IsNetworkMonitoring,
  
  // System
  GetPlatform,
  GetService,
  
  // Filesystem
  ReadDirectory,

  // System Metrics
  GetSystemInfo,
  GetCPUUsage,
  GetMemoryUsage,
  GetDiskUsage,
  GetNetworkInfo,
  
  // Process Management
  GetTopProcesses,
  
  // Lifecycle
  Initialize: InitializeCoordinator,
  StartMonitoring: StartCoordinatorMonitoring,
  Shutdown: ShutdownCoordinator,
  IsStarted: IsCoordinatorStarted,
}

export default ServiceCoordinator
