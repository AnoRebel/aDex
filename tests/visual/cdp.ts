/**
 * Minimal Chrome DevTools Protocol client for visual verification.
 *
 * Attaches to an already-running Chromium-family browser rather than
 * spawning one. bunwright (the project's browser-automation dependency)
 * always launches its own browser via Bun.WebView and exposes no
 * attach-to-existing-session option, so for driving the developer's live
 * Helium instance we speak CDP directly — which is what those libraries
 * wrap in any case.
 *
 * Helium specifics:
 *  - The HTTP discovery endpoints (/json/version, /json/list) are disabled,
 *    so the WebSocket URL cannot be looked up over HTTP.
 *  - Helium writes `DevToolsActivePort` into its user-data directory: line 1
 *    is the port, line 2 the browser WebSocket path. Both change on every
 *    restart, so this reads the file fresh on each run and never caches a
 *    ws:// URL.
 *  - Connecting a new debugging session may raise an in-browser permission
 *    prompt.
 */

import { readFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

/** Candidate user-data directories, per platform. */
const USER_DATA_DIRS = [
  join(homedir(), '.config/net.imput.helium'), // Linux
  join(homedir(), '.config/helium'), // Linux (alternate)
  join(homedir(), 'Library/Application Support/net.imput.helium'), // macOS
  join(homedir(), 'AppData/Local/imput/Helium/User Data'), // Windows
]

export function readDevToolsEndpoint(userDataDir?: string): string {
  const dirs = userDataDir ? [userDataDir] : USER_DATA_DIRS
  for (const dir of dirs) {
    try {
      const raw = readFileSync(join(dir, 'DevToolsActivePort'), 'utf8')
      const [port, path] = raw.trim().split('\n').map(s => s.trim())
      if (port && path) return `ws://127.0.0.1:${port}${path}`
    } catch {
      // Try the next candidate directory.
    }
  }
  throw new Error(
    `No DevToolsActivePort found. Looked in:\n  ${dirs.join('\n  ')}\n` +
      'Is the browser running with remote debugging enabled?',
  )
}

type Pending = { resolve: (v: any) => void; reject: (e: Error) => void }

/** A CDP session against one target (or the browser endpoint itself). */
export class CDPSession {
  #ws: WebSocket
  #id = 0
  #pending = new Map<number, Pending>()
  #listeners = new Map<string, Array<(params: any) => void>>()

  private constructor(ws: WebSocket) {
    this.#ws = ws
    ws.onmessage = (ev: MessageEvent) => {
      const msg = JSON.parse(String(ev.data))
      if (msg.id !== undefined) {
        const p = this.#pending.get(msg.id)
        if (!p) return
        this.#pending.delete(msg.id)
        msg.error ? p.reject(new Error(JSON.stringify(msg.error))) : p.resolve(msg.result)
      } else if (msg.method) {
        for (const fn of this.#listeners.get(msg.method) ?? []) fn(msg.params)
      }
    }
  }

  static connect(url: string, timeoutMs = 15_000): Promise<CDPSession> {
    return new Promise((resolve, reject) => {
      const ws = new WebSocket(url)
      const timer = setTimeout(
        () => reject(new Error(`CDP connect timed out after ${timeoutMs}ms: ${url}`)),
        timeoutMs,
      )
      ws.onopen = () => {
        clearTimeout(timer)
        resolve(new CDPSession(ws))
      }
      ws.onerror = () => {
        clearTimeout(timer)
        reject(new Error(`CDP connection failed: ${url}`))
      }
    })
  }

  send<T = any>(method: string, params: Record<string, unknown> = {}, sessionId?: string): Promise<T> {
    const id = ++this.#id
    return new Promise<T>((resolve, reject) => {
      this.#pending.set(id, { resolve, reject })
      this.#ws.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }))
      setTimeout(() => {
        if (this.#pending.delete(id)) reject(new Error(`CDP call timed out: ${method}`))
      }, 30_000)
    })
  }

  on(method: string, fn: (params: any) => void) {
    const list = this.#listeners.get(method) ?? []
    list.push(fn)
    this.#listeners.set(method, list)
  }

  close() {
    this.#ws.close()
  }
}
