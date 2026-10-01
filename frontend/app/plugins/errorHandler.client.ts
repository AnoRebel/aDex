/**
 * Last-resort error handling.
 *
 * Per-panel <ErrorBoundary> components contain component failures. This
 * catches everything they cannot: errors in async callbacks, timers, event
 * handlers and promise chains, which do not pass through Vue's
 * onErrorCaptured and would otherwise surface as an unhandled rejection or a
 * silently dead interface.
 *
 * The goal is that the application stays usable. Nothing here rethrows.
 */
import { handleComponentError } from '~/utils/errorHandler'

export default defineNuxtPlugin((nuxtApp) => {
  // Vue-level errors that escaped every boundary.
  nuxtApp.vueApp.config.errorHandler = (err, _instance, info) => {
    const error = err instanceof Error ? err : new Error(String(err))
    console.error('[vue] unhandled error:', error, info)
    try {
      handleComponentError('APP', error, { info })
    } catch {
      /* the reporter must never itself throw */
    }
  }

  if (typeof window === 'undefined') return

  // Rejected promises with no .catch — the most common way a background task
  // fails invisibly.
  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    const error = reason instanceof Error ? reason : new Error(String(reason))
    console.error('[app] unhandled promise rejection:', error)
    try {
      handleComponentError('ASYNC', error)
    } catch {
      /* ignore */
    }
    // Prevent the default console noise; the error is already recorded.
    event.preventDefault()
  })

  // Uncaught runtime errors from timers and native event handlers.
  window.addEventListener('error', (event) => {
    const error = event.error instanceof Error ? event.error : new Error(event.message)
    console.error('[app] uncaught error:', error)
    try {
      handleComponentError('RUNTIME', error, {
        source: event.filename,
        line: event.lineno,
      })
    } catch {
      /* ignore */
    }
  })
})
