<template>
  <div ref="containerRef" class="adex-terminal-instance">
    <div ref="xtermRef" class="xterm-container" />
  </div>
</template>

<script setup lang="ts">
/**
 * AdexTerminal -- renders a single xterm.js terminal for one shell tab.
 *
 * Props:
 *   sessionId  – the backend terminal session id
 *   active     – whether this terminal is the currently visible one
 *
 * Exposes:
 *   write(data)  – write raw data to the terminal
 *   focus()      – focus the xterm element
 *   fit()        – refit the terminal to its container
 */
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { useResizeObserver, useTimeoutFn, useDebounceFn } from '@vueuse/core'
import { useTerminalStore } from '~/stores/terminal'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { pulseKey } from '~/composables/useKeyboardPulse'
// Static imports for xterm + addons.
//
// NOTE: We previously used dynamic `import('@xterm/xterm')` to defer the
// chunk, but Wails' webview serves assets via `wails://wails.localhost/...`,
// and dynamic-import URL resolution against that scheme intermittently
// fails — the fetch returns text/html (the SPA shell) instead of the JS
// chunk, the import rejects, and the user sees the "failed to load
// xterm.js" placeholder. Static imports get inlined into the main chunk
// at build time, which sidesteps the URL resolver entirely.
//
// SSR is already off (`ssr: false` in nuxt.config.ts), so importing xterm
// at module top-level is safe — the file only loads in the browser.
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebglAddon } from '@xterm/addon-webgl'
import { Events } from '~/lib/wailsjs/runtime'
// NOTE: We deliberately don't import @xterm/addon-canvas here.
//   1. The package only ships CJS (no `module` field), and Vite's
//      named-import-from-CJS interop fails for the UMD self-assigning
//      `module.exports = CanvasAddon` shape it uses, throwing
//      "Importing binding name 'CanvasAddon' is not found" at runtime.
//   2. Every modern Wails webview (WebView2 / WebKitGTK / WKWebView)
//      supports WebGL, so a CanvasAddon fallback never fires in practice.
//   3. If WebGL ever fails, xterm's built-in DOM renderer auto-activates —
//      we don't lose anything by skipping the canvas tier.

const props = defineProps<{
  sessionId: string
  active: boolean
}>()

// Emitted when this session's PTY exits. Parent marks the tab DEAD
// and surfaces a restart click target. The event is scoped to our
// sessionId so unmounting one tab doesn't tear down listeners on
// the others.
const emit = defineEmits<{
  (e: 'exited', sessionId: string): void
}>()

const terminalStore = useTerminalStore()
const containerRef = ref<HTMLElement | null>(null)
const xtermRef = ref<HTMLElement | null>(null)

let term: InstanceType<typeof XTerm> | null = null
let fitAddon: FitAddon | null = null
let wailsEventCleanup: (() => void) | null = null

// ---- Init ----
function initTerminal() {
  if (!xtermRef.value) {
    console.error('[AdexTerminal] xtermRef is null at init — DOM not ready')
    return
  }

  // Read CSS vars for theme colors
  const style = getComputedStyle(document.documentElement)
  const bg = style.getPropertyValue('--terminal_bg').trim() || '#05080d'
  const fg = style.getPropertyValue('--terminal_fg').trim() || '#aacfd1'
  const cursor = style.getPropertyValue('--terminal_cursor').trim() || fg
  const selection = style.getPropertyValue('--terminal_selection').trim() || 'rgba(170,207,209,0.3)'

  // Font stack with Nerd Font support. The browser tries each name in
  // order — if a glyph (e.g. shell-prompt powerline arrows, dev icons)
  // is missing from the primary font, it falls through to the next.
  // Most modern dev setups have at least one of these installed; if
  // none are, xterm falls back to the OS monospace and Nerd glyphs
  // render as boxes — but the prompt text itself stays readable.
  const NERD_FONT_FALLBACK = [
    "'JetBrainsMono Nerd Font'",
    "'FiraCode Nerd Font'",
    "'Hack Nerd Font'",
    "'MesloLGS NF'",
    "'CaskaydiaCove Nerd Font'",
    "'Symbols Nerd Font'",
    "'Fira Code'",
    'monospace',
  ].join(', ')
  const userFont = style.getPropertyValue('--terminal_font').trim()
  // Append Nerd Font fallbacks AFTER the user's choice so prompt
  // glyphs still resolve even when the user picked a non-NF primary.
  const fontFamily = userFont
    ? `${userFont}, ${NERD_FONT_FALLBACK}`
    : NERD_FONT_FALLBACK

  term = new XTerm({
    fontFamily,
    fontSize: 14,
    lineHeight: 1.2,
    cursorStyle: 'block' as const,
    cursorBlink: true,
    scrollback: 1500,
    allowProposedApi: true,
    theme: {
      background: bg,
      foreground: fg,
      cursor,
      selection,
      black: '#000000',
      red: '#e06c75',
      green: '#98c379',
      yellow: '#d19a66',
      blue: '#61afef',
      magenta: '#c678dd',
      cyan: '#56b6c2',
      white: '#abb2bf',
      brightBlack: '#5c6370',
      brightRed: '#e06c75',
      brightGreen: '#98c379',
      brightYellow: '#d19a66',
      brightBlue: '#61afef',
      brightMagenta: '#c678dd',
      brightCyan: '#56b6c2',
      brightWhite: '#ffffff',
    },
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)

  // WebGL renderer for performance. If it ever fails (rare on modern
  // Wails webviews), xterm's built-in DOM renderer kicks in
  // automatically — no manual canvas fallback needed.
  try {
    term.loadAddon(new WebglAddon())
  } catch {
    // Fall through to xterm's default DOM renderer.
  }

  // Open and fit. We defer the first fit() via VueUse's useTimeoutFn
  // (which auto-cleans on unmount) so the container element has measured
  // layout — calling fit() before xterm sees a non-zero clientWidth
  // raises in some Wails webview builds.
  term.open(xtermRef.value)
  useTimeoutFn(() => {
    try {
      fitAddon?.fit()
    } catch {
      // ignore fit errors during init
    }
  }, 0)

  // Send user input to backend.
  //
  // Diagnostic log: if you type and nothing echoes, check DevTools
  // console for `[term.onData]` lines. Their presence proves xterm
  // has focus and the keystroke reached it; their ABSENCE means
  // focus is elsewhere (the AdexKeyboard buttons, the file manager,
  // or the topbar — anything that isn't a contenteditable / textarea
  // / xterm helper-textarea).
  // Typewriter feedback. Fire the `keyboard` cue on every keystroke
  // (data event from xterm) so the user gets the same tactile audio
  // feedback as a hardware terminal. The cue itself is rate-limited
  // internally and falls through to the synth when the WAV isn't
  // cached yet, so spamming doesn't queue infinite clicks.
  //
  // Note: term.onData fires for EACH byte the user typed — including
  // escape sequences (arrow keys → 3 bytes). We click once per onData
  // event, not per byte, by gating on data length so paste of a 200B
  // string emits one cue, not 200.
  const audio = useAdexAudio()
  term.onData((data: string) => {
    // eslint-disable-next-line no-console
    console.debug('[term.onData]', JSON.stringify(data), 'session', props.sessionId)
    if (data.length > 0) {
      // 'keypress' is the new short mechanical-keyboard click — shorter
      // and crisper than the legacy 'keyboard' cue (which had a longer
      // typewriter sample that sounded sluggish per keystroke).
      try { audio.playCue('keypress') } catch { /* non-fatal */ }
      // Flash the on-screen keyboard. xterm captures keystrokes at the
      // textarea level and calls preventDefault, so the keyboard's own
      // document-level keydown handler never sees terminal input. We
      // bridge through this shared module instead.
      try { pulseKey(data) } catch { /* non-fatal */ }
    }
    terminalStore.sendInput(data, props.sessionId).catch((err) => {
      console.warn('[term.onData] sendInput rejected:', err)
    })
  })

  // Notify backend of size changes
  term.onResize(({ cols, rows }: { cols: number; rows: number }) => {
    terminalStore.resizeSession(cols, rows, props.sessionId)
  })

  // Listen for output from the Wails backend
  setupWailsEvents()

  // Refit on container resize. Two safety nets here:
  //
  //   1. **Debounce** — without this, a continuous WM resize fires
  //      ResizeObserver dozens of times per second, each call asks
  //      xterm to refit, xterm rewrites the canvas, ResizeObserver
  //      sees the canvas change, fires again. Visible flicker.
  //   2. **Size diff guard** — `fitAddon.fit()` can sometimes leave
  //      the canvas one pixel different even when nothing changed,
  //      kicking off the same loop. Skip when the container's measured
  //      size is identical to last fit.
  let lastFitW = 0
  let lastFitH = 0
  const debouncedFit = useDebounceFn(() => {
    if (!props.active) return
    const el = containerRef.value
    if (!el) return
    const rect = el.getBoundingClientRect()
    if (rect.width === lastFitW && rect.height === lastFitH) return
    lastFitW = rect.width
    lastFitH = rect.height
    try {
      fitAddon?.fit()
    } catch {
      // Ignore; xterm fit can throw before the canvas is laid out.
    }
  }, 80)
  useResizeObserver(containerRef, () => debouncedFit())

  // If this terminal is active, focus it
  if (props.active) {
    term.focus()
  }

  // Write welcome message for the first terminal
  term.writeln(`\x1b[36mWelcome to aDex Terminal\x1b[0m`)
  term.writeln(`\x1b[90mSession: ${props.sessionId}\x1b[0m`)
  term.writeln('')
}

function setupWailsEvents() {
  // Wails v2 dispatches events with the emitted payload as the FIRST
  // argument to the listener. The backend `terminal.output` payload is
  // shaped:
  //   { terminalId: string, data: number[] }
  //
  // (Note `terminalId`, NOT `sessionId` — naming was inconsistent with
  // the frontend's prop and used to silently no-op every output write.)
  //
  // `data` arrives as a plain number array because Wails JSON-marshals
  // Go's `[]byte` that way; xterm's `term.write` accepts string |
  // Uint8Array but NOT a plain array, so we must convert before writing.
  if (typeof window === 'undefined') {
    console.warn('[AdexTerminal] Wails runtime not available — terminal output will not flow')
    return
  }

  // Decode a Wails-marshalled `[]byte` payload back to bytes.
  //
  // Go's encoding/json serialises `[]byte` as a STANDARD-base64 string,
  // not a number array — that's why an earlier version of this handler
  // (which assumed `[]byte` → `number[]`) ended up writing literal
  // characters like `NGg=DQ0bW...` straight into xterm. Decode via the
  // browser's atob, then build a Uint8Array byte-by-byte. Returns null
  // if the string isn't valid base64 (so callers can fall back to
  // writing the raw string for legacy paths).
  const base64ToBytes = (s: string): Uint8Array | null => {
    try {
      const bin = atob(s)
      const out = new Uint8Array(bin.length)
      for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i)
      return out
    } catch {
      return null
    }
  }

  const handler = (payload: unknown) => {
    if (!term) return
    if (typeof payload === 'string') {
      term.write(payload)
      return
    }
    if (typeof payload === 'object' && payload !== null) {
      const obj = payload as { terminalId?: string; sessionId?: string; data?: unknown }
      // Accept both terminalId (current backend) and sessionId (legacy
      // shape) so a future rename doesn't break us.
      const id = obj.terminalId ?? obj.sessionId
      if (id && id !== props.sessionId) return
      const raw = obj.data
      if (typeof raw === 'string') {
        // Wails marshals Go []byte as base64 — decode first.
        const decoded = base64ToBytes(raw)
        if (decoded) {
          term.write(decoded)
        } else {
          term.write(raw)
        }
      } else if (Array.isArray(raw)) {
        term.write(new Uint8Array(raw))
      } else if (raw instanceof Uint8Array) {
        term.write(raw)
      }
    }
  }

  // Session-scoped output channel ONLY. The backend emits per-session
  // `terminal.output.<id>` so each pane owns its own stream.
  //
  // Wails v3 returns an unsubscribe function from Events.On, so cleanup
  // is precise: calling it removes THIS listener and nothing else. (In
  // v2, EventsOff removed every listener registered for an event name,
  // which is why subscribing to a shared global channel meant one tab
  // unmounting silently disconnected output for all the others.)
  const outputEventName = `terminal.output.${props.sessionId}`
  const offOutput = Events.On(outputEventName, handler)

  // Per-session exit event. Backend emits `terminal.exited.<id>` when
  // the PTY closes.
  const exitedEventName = `terminal.exited.${props.sessionId}`
  const exitedHandler = () => {
    if (term) {
      term.write('\r\n\x1b[33m[process exited — click tab to restart]\x1b[0m\r\n')
    }
    emit('exited', props.sessionId)
  }
  const offExited = Events.On(exitedEventName, exitedHandler)

  wailsEventCleanup = () => {
    offOutput()
    offExited()
  }
}

// ---- Exposed methods ----
function write(data: string) {
  term?.write(data)
}

function focus() {
  term?.focus()
}

function fit() {
  try {
    fitAddon?.fit()
  } catch {
    // ignore
  }
}

defineExpose({ write, focus, fit })

// ---- Watchers ----
watch(
  () => props.active,
  (isActive) => {
    if (isActive) {
      nextTick(() => {
        fit()
        focus()
      })
    }
  }
)

// ---- Lifecycle ----
onMounted(() => {
  initTerminal()
})

onBeforeUnmount(() => {
  // useResizeObserver registered above is cleaned up automatically.
  wailsEventCleanup?.()
  term?.dispose()
  term = null
  fitAddon = null
})
</script>

<style scoped>
.adex-terminal-instance {
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;
}

.xterm-container {
  width: 100%;
  height: 100%;
}

/* xterm.js overrides for aDex-UI look */
.xterm-container :deep(.xterm) {
  height: 100%;
  padding: 0.3vh 0.3vw;
}

.xterm-container :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

.xterm-container :deep(.xterm-viewport::-webkit-scrollbar) {
  width: 0.3vw;
}

.xterm-container :deep(.xterm-viewport::-webkit-scrollbar-track) {
  background: transparent;
}

.xterm-container :deep(.xterm-viewport::-webkit-scrollbar-thumb) {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
}
</style>
