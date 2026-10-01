// composables/useNotifications.ts
//
// Desktop notifications for long-running operations.
//
// The point is reach: a message posted to the platform's notification centre
// arrives even when aDex is behind another window, which an in-window toast
// cannot do. Short operations must never notify — a notification for something
// the user just watched finish is noise, and noise trains people to ignore
// them.

import { useStorage } from '@vueuse/core'

/** Operations faster than this are not worth a notification: the user was
 *  almost certainly still looking at the window. */
const MIN_DURATION_MS = 5_000

const settings = useStorage<{ advanced?: { notifications?: boolean } }>('adex-settings', {})

function enabled(): boolean {
  // Default on: the toggle exists to turn them off, not to discover them.
  return settings.value?.advanced?.notifications !== false
}

export function useNotifications() {
  /** Post a notification, if the user has them enabled. */
  async function notify(title: string, body: string): Promise<void> {
    if (!enabled()) return
    try {
      const { SendNotification } = await import('~/lib/wailsjs/coordinator')
      await SendNotification(title, body)
    } catch {
      // Never let a missing notification break the caller.
    }
  }

  /**
   * Run an operation and notify only if it took long enough to be worth it.
   *
   * Returns whatever the operation returns, and rethrows its error after
   * notifying — the caller's own error handling still runs.
   */
  async function notifyOnCompletion<T>(
    label: string,
    op: () => Promise<T>,
  ): Promise<T> {
    const started = Date.now()
    try {
      const result = await op()
      if (Date.now() - started >= MIN_DURATION_MS) {
        void notify(`${label} finished`, `Completed in ${formatDuration(Date.now() - started)}.`)
      }
      return result
    } catch (err) {
      if (Date.now() - started >= MIN_DURATION_MS) {
        void notify(
          `${label} failed`,
          err instanceof Error ? err.message : String(err),
        )
      }
      throw err
    }
  }

  return { notify, notifyOnCompletion, enabled }
}

function formatDuration(ms: number): string {
  const s = Math.round(ms / 1000)
  if (s < 60) return `${s}s`
  const m = Math.floor(s / 60)
  return `${m}m ${s % 60}s`
}
