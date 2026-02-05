// Wails v2 Bindings for aDex-UI
// Re-export all coordinator bindings

export * from '~/lib/wailsjs/coordinator'

// Also export as default for compatibility
import * as ServiceCoordinator from '~/lib/wailsjs/coordinator'
export { ServiceCoordinator }
export default ServiceCoordinator
