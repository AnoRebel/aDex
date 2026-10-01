/**
 * Monotonic-ish timestamp in milliseconds, safe in every runtime the app
 * runs in.
 *
 * Some webview contexts (including the Wails WebView on certain platforms)
 * expose a partial `performance` object without `now`. Code that guarded with
 * `typeof performance !== "undefined"` therefore passed the check and then
 * threw `performance.now is not a function` — which, in the audio path,
 * silently killed every rate-limited sound cue.
 *
 * Prefer `performance.now()` when it is genuinely callable (monotonic, immune
 * to wall-clock adjustments) and fall back to `Date.now()` otherwise.
 */
export function nowMs(): number {
  return typeof performance?.now === 'function' ? performance.now() : Date.now()
}
