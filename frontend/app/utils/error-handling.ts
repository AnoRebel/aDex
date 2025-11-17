/**
 * Error handling utilities for frontend error management
 */

import { useToast, useNotification } from '@vueuse/core'
import { useStorage } from '@vueuse/core'
import type { Ref } from 'vue'

export interface AppError {
  id: string
  code: string
  message: string
  userMessage: string
  severity: 'error' | 'warning' | 'info'
  category: 'system' | 'user' | 'network' | 'file'
  retryable: boolean
  recovered: boolean
  timestamp: Date
  context?: ErrorContext
  stackTrace?: StackFrame[]
}

export interface ErrorContext {
  operation: string
  component: string
  userId?: string
  sessionId?: string
  requestId?: string
  metadata?: Record<string, any>
  userAgent?: string
  timestamp: Date
}

export interface StackFrame {
  function: string
  file: string
  line: number
  package: string
}

export interface ErrorTemplate {
  code: string
  title: string
  description: string
  cause: string
  solution: string
  severity: string
  category: string
  retryable: boolean
  recoverySteps: string[]
  translations: Record<string, string>
}

export interface ErrorHandlerConfig {
  showStackTrace: boolean
  logLevel: 'debug' | 'info' | 'warn' | 'error'
  userFriendlyErrors: boolean
  errorReporting: boolean
  maxErrorsInMemory: number
  enableRecovery: boolean
  autoRetryAttempts: number
  retryDelay: number
  localizationEnabled: boolean
  defaultLanguage: string
}

export class ErrorHandler {
  private static instance: ErrorHandler
  private config: ErrorHandlerConfig
  private templates: Map<string, ErrorTemplate> = new Map()
  private errorHistory: Ref<AppError[]> = useStorage('error-history', [])
  private toast = useToast()
  private notification = useNotification()

  private constructor() {
    this.config = this.getDefaultConfig()
    this.loadDefaultTemplates()
  }

  static getInstance(): ErrorHandler {
    if (!ErrorHandler.instance) {
      ErrorHandler.instance = new ErrorHandler()
    }
    return ErrorHandler.instance
  }

  /**
   * Creates a new application error
   */
  createError(code: string, message: string, cause?: Error): AppError {
    const error: AppError = {
      id: this.generateErrorID(),
      code,
      message,
      timestamp: new Date(),
      severity: 'error',
      category: 'system',
      retryable: false,
      recovered: false
    }

    // Find template for this error code
    const template = this.templates.get(code)
    if (template) {
      error.userMessage = template.description
      error.severity = template.severity as any
      error.category = template.category as any
      error.retryable = template.retryable
    } else {
      error.userMessage = message
    }

    // Generate user-friendly message if enabled
    if (this.config.userFriendlyErrors) {
      error.userMessage = this.generateUserFriendlyMessage(error)
    }

    // Capture context
    error.context = this.captureErrorContext()

    // Capture stack trace if enabled
    if (this.config.showStackTrace) {
      error.stackTrace = this.captureStackTrace()
    }

    return error
  }

  /**
   * Handles an error with optional context
   */
  handleError(error: Error | AppError, context?: Partial<ErrorContext>): AppError {
    let appError: AppError

    if (this.isAppError(error)) {
      appError = error
    } else {
      appError = this.createError('UNKNOWN', error.message, error)
    }

    // Merge context if provided
    if (context) {
      appError.context = { ...appError.context, ...context }
    }

    // Add to error history
    this.addToHistory(appError)

    // Log the error
    this.logError(appError)

    // Show user notification
    this.showErrorNotification(appError)

    // Attempt recovery if enabled
    if (this.config.enableRecovery && appError.retryable) {
      appError.recovered = this.attemptRecovery(appError)
    }

    return appError
  }

  /**
   * Handles promises and catches errors
   */
  async handleAsync<T>(
    promise: Promise<T>,
    context?: Partial<ErrorContext>
  ): Promise<{ success: boolean; data?: T; error?: AppError }> {
    try {
      const data = await promise
      return { success: true, data }
    } catch (error) {
      const appError = this.handleError(error as Error, context)
      return { success: false, error: appError }
    }
  }

  /**
   * Wraps a function with error handling
   */
  wrapFunction<T extends any[], R>(
    fn: (...args: T) => R,
    context?: Partial<ErrorContext>
  ): (...args: T) => R | never {
    return (...args: T): R | never => {
      try {
        return fn(...args)
      } catch (error) {
        const appError = this.handleError(error as Error, {
          ...context,
          operation: `${fn.name}(${args.map(() => '...').join(', ')})`
        })
        throw appError
      }
    }
  }

  /**
   * Gets the appropriate error message for the user
   */
  getErrorMessage(error: AppError, language = this.config.defaultLanguage): string {
    const template = this.templates.get(error.code)
    if (template?.translations[language]) {
      return template.translations[language]
    }
    return error.userMessage
  }

  /**
   * Gets recovery steps for an error
   */
  getRecoverySteps(error: AppError): string[] {
    const template = this.templates.get(error.code)
    return template?.recoverySteps || [
      'Please try the operation again.',
      'If the problem persists, restart the application.',
      'Contact support if the issue continues.'
    ]
  }

  /**
   * Clears error history
   */
  clearHistory(): void {
    this.errorHistory.value = []
  }

  /**
   * Gets error history
   */
  getErrorHistory(): AppError[] {
    return this.errorHistory.value
  }

  /**
   * Registers a custom error template
   */
  registerTemplate(template: ErrorTemplate): void {
    this.templates.set(template.code, template)
  }

  /**
   * Updates the error handling configuration
   */
  setConfig(config: Partial<ErrorHandlerConfig>): void {
    this.config = { ...this.config, ...config }
  }

  /**
   * Gets current configuration
   */
  getConfig(): ErrorHandlerConfig {
    return { ...this.config }
  }

  /**
   * Creates an error report
   */
  createErrorReport(): {
    id: string
    timestamp: Date
    version: string
    errors: AppError[]
    stats: {
      totalErrors: number
      errorsByCode: Record<string, number>
      errorsBySeverity: Record<string, number>
      errorsByCategory: Record<string, number>
      recoveredErrors: number
      retryableErrors: number
    }
  } {
    const errors = this.errorHistory.value

    const stats = {
      totalErrors: errors.length,
      errorsByCode: {} as Record<string, number>,
      errorsBySeverity: {} as Record<string, number>,
      errorsByCategory: {} as Record<string, number>,
      recoveredErrors: errors.filter(e => e.recovered).length,
      retryableErrors: errors.filter(e => e.retryable).length
    }

    // Calculate statistics
    errors.forEach(error => {
      stats.errorsByCode[error.code] = (stats.errorsByCode[error.code] || 0) + 1
      stats.errorsBySeverity[error.severity] = (stats.errorsBySeverity[error.severity] || 0) + 1
      stats.errorsByCategory[error.category] = (stats.errorsByCategory[error.category] || 0) + 1
    })

    return {
      id: this.generateReportID(),
      timestamp: new Date(),
      version: '1.0.0',
      errors,
      stats
    }
  }

  // Private helper methods

  private getDefaultConfig(): ErrorHandlerConfig {
    return {
      showStackTrace: false,
      logLevel: 'error',
      userFriendlyErrors: true,
      errorReporting: false,
      maxErrorsInMemory: 100,
      enableRecovery: true,
      autoRetryAttempts: 3,
      retryDelay: 2000,
      localizationEnabled: true,
      defaultLanguage: 'en'
    }
  }

  private loadDefaultTemplates(): void {
    const templates: ErrorTemplate[] = [
      {
        code: 'FILE_NOT_FOUND',
        title: 'File Not Found',
        description: 'The requested file could not be found.',
        cause: 'The file may have been moved, deleted, or never existed.',
        solution: 'Check the file path and ensure the file exists.',
        severity: 'error',
        category: 'file',
        retryable: false,
        recoverySteps: [
          'Verify the file path is correct',
          'Check if the file exists in the expected location',
          'Ensure you have permission to access the file'
        ],
        translations: {
          en: 'The requested file could not be found.',
          es: 'El archivo solicitado no pudo ser encontrado.',
          fr: 'Le fichier demandé n\'a pas pu être trouvé.'
        }
      },
      {
        code: 'NETWORK_ERROR',
        title: 'Network Connection Failed',
        description: 'Unable to connect to the network or server.',
        cause: 'Network connection may be unavailable or the server is not responding.',
        solution: 'Check your network connection and try again.',
        severity: 'error',
        category: 'network',
        retryable: true,
        recoverySteps: [
          'Check your internet connection',
          'Verify the server is accessible',
          'Try again in a few moments',
          'Contact your network administrator if issues persist'
        ],
        translations: {
          en: 'Unable to connect to the network or server.',
          es: 'No se puede conectar a la red o al servidor.',
          fr: 'Impossible de se connecter au réseau ou au serveur.'
        }
      },
      {
        code: 'VALIDATION_ERROR',
        title: 'Invalid Input',
        description: 'The provided input is not valid.',
        cause: 'The input may contain invalid characters or format.',
        solution: 'Please check your input and try again.',
        severity: 'warning',
        category: 'user',
        retryable: true,
        recoverySteps: [
          'Check your input for typos or errors',
          'Ensure all required fields are filled',
          'Follow the specified format requirements'
        ],
        translations: {
          en: 'The provided input is not valid.',
          es: 'La entrada proporcionada no es válida.',
          fr: 'L\'entrée fournie n\'est pas valide.'
        }
      }
    ]

    templates.forEach(template => {
      this.templates.set(template.code, template)
    })
  }

  private captureErrorContext(): ErrorContext {
    return {
      operation: '',
      component: '',
      timestamp: new Date(),
      userAgent: navigator.userAgent
    }
  }

  private captureStackTrace(): StackFrame[] {
    const stack = new Error().stack
    if (!stack) return []

    return stack
      .split('\n')
      .slice(2) // Skip current function and the captureStackTrace call
      .map(line => {
        const match = line.match(/at (.+) \((.+):(\d+):(\d+)\)/)
        if (match) {
          return {
            function: match[1],
            file: match[2],
            line: parseInt(match[3]),
            package: ''
          }
        }
        return null
      })
      .filter(Boolean) as StackFrame[]
  }

  private generateUserFriendlyMessage(error: AppError): string {
    const template = this.templates.get(error.code)
    if (template) {
      return template.description
    }

    switch (error.code) {
      case 'UNKNOWN':
        return 'An unexpected error occurred. Please try again.'
      default:
        return 'Something went wrong. Please check your input and try again.'
    }
  }

  private addToHistory(error: AppError): void {
    this.errorHistory.value.unshift(error)

    // Limit the number of errors in memory
    if (this.errorHistory.value.length > this.config.maxErrorsInMemory) {
      this.errorHistory.value = this.errorHistory.value.slice(0, this.config.maxErrorsInMemory)
    }
  }

  private logError(error: AppError): void {
    const logData = {
      id: error.id,
      code: error.code,
      message: error.message,
      component: error.context?.component,
      operation: error.context?.operation,
      severity: error.severity
    }

    switch (this.config.logLevel) {
      case 'debug':
        console.debug('Error occurred', logData)
        break
      case 'info':
        console.info('Error occurred', logData)
        break
      case 'warn':
        console.warn('Error occurred', logData)
        break
      case 'error':
      default:
        console.error('Error occurred', { ...logData, stack: error.stackTrace })
        break
    }
  }

  private showErrorNotification(error: AppError): void {
    const message = this.getErrorMessage(error)
    const title = this.templates.get(error.code)?.title || 'Error'

    if (error.severity === 'error') {
      this.toast.error(message, {
        description: error.code,
        timeout: 5000
      })
    } else if (error.severity === 'warning') {
      this.toast.warning(message, {
        description: error.code,
        timeout: 3000
      })
    } else {
      this.toast.info(message, {
        description: error.code,
        timeout: 2000
      })
    }
  }

  private attemptRecovery(error: AppError): boolean {
    console.log('Attempting recovery for error:', error.id, error.code)
    // This is a placeholder for recovery logic
    return false
  }

  private isAppError(error: any): error is AppError {
    return error && typeof error === 'object' && 'id' in error && 'code' in error
  }

  private generateErrorID(): string {
    return `ERR_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private generateReportID(): string {
    return `RPT_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }
}

/**
 * Vue composable for error handling
 */
export function useErrorHandler() {
  const errorHandler = ErrorHandler.getInstance()

  const handleError = (error: Error | AppError, context?: Partial<ErrorContext>) => {
    return errorHandler.handleError(error, context)
  }

  const handleAsync = async <T>(
    promise: Promise<T>,
    context?: Partial<ErrorContext>
  ) => {
    return errorHandler.handleAsync(promise, context)
  }

  const createError = (code: string, message: string, cause?: Error) => {
    return errorHandler.createError(code, message, cause)
  }

  const getRecoverySteps = (error: AppError) => {
    return errorHandler.getRecoverySteps(error)
  }

  const getErrorHistory = () => {
    return errorHandler.getErrorHistory()
  }

  const clearHistory = () => {
    return errorHandler.clearHistory()
  }

  const generateReport = () => {
    return errorHandler.createErrorReport()
  }

  return {
    handleError,
    handleAsync,
    createError,
    getRecoverySteps,
    getErrorHistory,
    clearHistory,
    generateReport,
    config: errorHandler.getConfig(),
    setConfig: errorHandler.setConfig.bind(errorHandler)
  }
}

/**
 * Global error handler for unhandled errors
 */
export function setupGlobalErrorHandler() {
  const errorHandler = ErrorHandler.getInstance()

  // Handle unhandled promise rejections
  window.addEventListener('unhandledrejection', (event) => {
    const error = errorHandler.handleError(
      event.reason instanceof Error ? event.reason : new Error(String(event.reason)),
      {
        operation: 'unhandled promise rejection',
        component: 'global'
      }
    )
    console.error('Unhandled promise rejection:', error)
  })

  // Handle uncaught errors
  window.addEventListener('error', (event) => {
    const error = errorHandler.handleError(
      event.error instanceof Error ? event.error : new Error(event.message),
      {
        operation: 'uncaught error',
        component: 'global',
        metadata: {
          filename: event.filename,
          lineno: event.lineno,
          colno: event.colno
        }
      }
    )
    console.error('Uncaught error:', error)
  })
}

/**
 * Error boundary component for Vue
 */
export const ErrorBoundary = {
  name: 'ErrorBoundary',
  props: {
    fallback: {
      type: Function,
      default: () => 'Something went wrong. Please refresh the page.'
    }
  },
  data() {
    return {
      hasError: false,
      error: null as AppError | null
    }
  },
  errorCaptured(error: Error, instance: any, info: string) {
    const errorHandler = ErrorHandler.getInstance()
    this.error = errorHandler.handleError(error, {
      operation: info,
      component: instance?.$options.name || 'Unknown'
    })
    this.hasError = true
    return false // Prevent error from propagating further
  },
  render() {
    if (this.hasError) {
      return this.fallback(this.error)
    }
    return this.$slots.default?.()
  }
}