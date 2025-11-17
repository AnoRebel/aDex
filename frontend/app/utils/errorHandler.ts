import { useAppStore } from '~/stores/app'

export interface ErrorInfo {
  id: string
  type: 'network' | 'system' | 'user' | 'component' | 'backend' | 'filesystem' | 'terminal'
  severity: 'low' | 'medium' | 'high' | 'critical'
  title: string
  message: string
  details?: string
  timestamp: Date
  component?: string
  action?: string
  stack?: string
  context?: Record<string, any>
  retryable?: boolean
  reported?: boolean
}

export interface ErrorReport {
  error: ErrorInfo
  userAgent: string
  url: string
  appVersion: string
  systemInfo?: {
    os?: string
    browser?: string
    language?: string
    screenResolution?: string
  }
}

class ErrorHandler {
  private errors: ErrorInfo[] = []
  private maxErrors = 100
  private reportingEndpoint = '/api/errors/report'
  private isReporting = false

  constructor() {
    this.setupGlobalErrorHandlers()
  }

  private setupGlobalErrorHandlers(): void {
    // Handle unhandled JavaScript errors
    if (typeof window !== 'undefined') {
      window.addEventListener('error', (event) => {
        this.handleError({
          type: 'component',
          severity: 'high',
          title: 'JavaScript Error',
          message: event.message,
          details: `Line ${event.lineno}, Column ${event.colno} in ${event.filename}`,
          stack: event.error?.stack,
          context: {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno,
          },
        })
      })

      // Handle unhandled promise rejections
      window.addEventListener('unhandledrejection', (event) => {
        this.handleError({
          type: 'component',
          severity: 'high',
          title: 'Unhandled Promise Rejection',
          message: event.reason?.message || 'Promise rejected without error',
          details: String(event.reason),
          stack: event.reason?.stack,
          context: {
            reason: event.reason,
          },
        })
      })
    }
  }

  public handleError(error: Partial<ErrorInfo>): string {
    const errorInfo: ErrorInfo = {
      id: this.generateErrorId(),
      type: error.type || 'component',
      severity: error.severity || 'medium',
      title: error.title || 'Unknown Error',
      message: error.message || 'An error occurred',
      timestamp: new Date(),
      retryable: error.retryable ?? true,
      reported: false,
      ...error,
    }

    // Add to error log
    this.errors.push(errorInfo)

    // Keep error list manageable
    if (this.errors.length > this.maxErrors) {
      this.errors.shift()
    }

    // Log to console
    this.logError(errorInfo)

    // Show user notification
    this.showErrorNotification(errorInfo)

    // Auto-report critical errors
    if (errorInfo.severity === 'critical') {
      this.reportError(errorInfo)
    }

    return errorInfo.id
  }

  public handleNetworkError(error: Error, context?: Record<string, any>): string {
    return this.handleError({
      type: 'network',
      severity: 'medium',
      title: 'Network Error',
      message: error.message,
      details: `Failed to connect to backend service`,
      stack: error.stack,
      context,
      retryable: true,
    })
  }

  public handleBackendError(statusCode: number, message: string, context?: Record<string, any>): string {
    const severity = statusCode >= 500 ? 'high' : 'medium'
    return this.handleError({
      type: 'backend',
      severity,
      title: `Backend Error (${statusCode})`,
      message,
      details: `HTTP ${statusCode}: ${message}`,
      context: { statusCode, ...context },
      retryable: statusCode < 500,
    })
  }

  public handleFileSystemError(operation: string, path: string, error: Error): string {
    return this.handleError({
      type: 'filesystem',
      severity: 'medium',
      title: 'File System Error',
      message: `Failed to ${operation}: ${error.message}`,
      details: `Path: ${path}`,
      stack: error.stack,
      context: { operation, path },
      retryable: true,
    })
  }

  public handleTerminalError(command: string, error: Error): string {
    return this.handleError({
      type: 'terminal',
      severity: 'medium',
      title: 'Terminal Error',
      message: `Command failed: ${command}`,
      details: error.message,
      stack: error.stack,
      context: { command },
      retryable: false,
    })
  }

  public handleComponentError(componentName: string, error: Error, context?: Record<string, any>): string {
    return this.handleError({
      type: 'component',
      severity: 'high',
      title: `Component Error: ${componentName}`,
      message: error.message,
      details: `Error in ${componentName} component`,
      stack: error.stack,
      component: componentName,
      context,
      retryable: true,
    })
  }

  private generateErrorId(): string {
    return `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private logError(error: ErrorInfo): void {
    const logMethod = {
      low: console.log,
      medium: console.warn,
      high: console.error,
      critical: console.error,
    }[error.severity]

    logMethod(`[${error.severity.toUpperCase()}] ${error.title}`, {
      message: error.message,
      details: error.details,
      component: error.component,
      context: error.context,
      stack: error.stack,
    })
  }

  private showErrorNotification(error: ErrorInfo): void {
    const appStore = useAppStore()

    // Don't show notifications for low severity errors
    if (error.severity === 'low') return

    const notificationType = {
      low: 'info',
      medium: 'warning',
      high: 'warning',
      critical: 'error',
    }[error.severity] as 'success' | 'warning' | 'error' | 'info'

    appStore.addAlert({
      type: notificationType,
      title: error.title,
      message: error.message,
      persistent: error.severity === 'critical',
      actions: error.retryable ? [
        {
          label: 'Retry',
          action: () => this.retryError(error.id),
          primary: true,
        },
        {
          label: 'Dismiss',
          action: () => {},
        },
      ] : [
        {
          label: 'Dismiss',
          action: () => {},
        },
      ],
    })
  }

  private async retryError(errorId: string): Promise<void> {
    const error = this.errors.find(e => e.id === errorId)
    if (!error || !error.retryable) return

    // Remove the error from the list
    this.removeError(errorId)

    // Retry based on error type
    switch (error.type) {
      case 'network':
        // Retry network request
        if (error.context?.url) {
          try {
            await fetch(error.context.url)
          } catch (retryError) {
            this.handleNetworkError(retryError as Error, error.context)
          }
        }
        break

      case 'filesystem':
        // Retry file operation
        if (error.context?.operation && error.context?.path) {
          // The specific retry logic would be handled by the component
          window.dispatchEvent(new CustomEvent('retry-filesystem-operation', {
            detail: error.context
          }))
        }
        break

      case 'terminal':
        // Retry terminal command
        if (error.context?.command) {
          window.dispatchEvent(new CustomEvent('retry-terminal-command', {
            detail: error.context.command
          }))
        }
        break

      case 'backend':
        // Retry backend request
        window.dispatchEvent(new CustomEvent('retry-backend-request', {
          detail: error.context
        }))
        break

      default:
        console.log(`Cannot retry error type: ${error.type}`)
    }
  }

  public async reportError(error: ErrorInfo): Promise<boolean> {
    if (this.isReporting || error.reported) return false

    try {
      this.isReporting = true

      const errorReport: ErrorReport = {
        error: {
          ...error,
          details: error.details || '',
          context: error.context || {},
        },
        userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'Unknown',
        url: typeof window !== 'undefined' ? window.location.href : 'Unknown',
        appVersion: this.getAppVersion(),
        systemInfo: this.getSystemInfo(),
      }

      const response = await fetch(this.reportingEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(errorReport),
      })

      if (response.ok) {
        error.reported = true
        return true
      } else {
        console.warn('Failed to report error:', response.status)
        return false
      }

    } catch (reportError) {
      console.warn('Error reporting failed:', reportError)
      return false
    } finally {
      this.isReporting = false
    }
  }

  private getAppVersion(): string {
    const appStore = useAppStore()
    return appStore.version || '2.0.0'
  }

  private getSystemInfo(): Record<string, string> {
    if (typeof navigator === 'undefined') return {}

    return {
      os: navigator.platform || 'Unknown',
      browser: `${navigator.appName} ${navigator.appVersion}`,
      language: navigator.language || 'Unknown',
      screenResolution: `${screen.width}x${screen.height}`,
    }
  }

  public getErrors(filter?: {
    type?: ErrorInfo['type']
    severity?: ErrorInfo['severity']
    component?: string
    limit?: number
  }): ErrorInfo[] {
    let filteredErrors = [...this.errors]

    if (filter?.type) {
      filteredErrors = filteredErrors.filter(e => e.type === filter.type)
    }

    if (filter?.severity) {
      filteredErrors = filteredErrors.filter(e => e.severity === filter.severity)
    }

    if (filter?.component) {
      filteredErrors = filteredErrors.filter(e => e.component === filter.component)
    }

    if (filter?.limit) {
      filteredErrors = filteredErrors.slice(-filter.limit)
    }

    return filteredErrors.reverse() // Most recent first
  }

  public removeError(errorId: string): void {
    this.errors = this.errors.filter(e => e.id !== errorId)
  }

  public clearErrors(): void {
    this.errors = []
  }

  public getErrorStats(): {
    total: number
    byType: Record<string, number>
    bySeverity: Record<string, number>
    recent: number
  } {
    const byType: Record<string, number> = {}
    const bySeverity: Record<string, number> = {}

    this.errors.forEach(error => {
      byType[error.type] = (byType[error.type] || 0) + 1
      bySeverity[error.severity] = (bySeverity[error.severity] || 0) + 1
    })

    const oneHourAgo = Date.now() - 60 * 60 * 1000
    const recent = this.errors.filter(e => e.timestamp.getTime() > oneHourAgo).length

    return {
      total: this.errors.length,
      byType,
      bySeverity,
      recent,
    }
  }

  public exportErrors(): string {
    return JSON.stringify({
      exportedAt: new Date().toISOString(),
      errors: this.errors,
      stats: this.getErrorStats(),
    }, null, 2)
  }
}

// Singleton instance
export const errorHandler = new ErrorHandler()

// Export convenience functions
export const handleError = (error: Partial<ErrorInfo>) => errorHandler.handleError(error)
export const handleNetworkError = (error: Error, context?: Record<string, any>) =>
  errorHandler.handleNetworkError(error, context)
export const handleBackendError = (statusCode: number, message: string, context?: Record<string, any>) =>
  errorHandler.handleBackendError(statusCode, message, context)
export const handleFileSystemError = (operation: string, path: string, error: Error) =>
  errorHandler.handleFileSystemError(operation, path, error)
export const handleTerminalError = (command: string, error: Error) =>
  errorHandler.handleTerminalError(command, error)
export const handleComponentError = (componentName: string, error: Error, context?: Record<string, any>) =>
  errorHandler.handleComponentError(componentName, error, context)

export default errorHandler