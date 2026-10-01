// Wails v3 runtime facade.
//
// Wraps `@wailsio/runtime` so the existing call sites keep working across the
// v2 -> v3 move. Two differences are bridged here rather than at ~19 call
// sites:
//
//  1. v2 exposed the runtime on `window.runtime.*`. v3 ships a real module.
//  2. v2's `EventsOff(name)` unsubscribed by event NAME. v3's `Events.On`
//     instead returns an unsubscribe function and has no by-name removal.
//     `Off` below is therefore implemented by tracking the unsubscribe
//     functions this module hands out, keyed by name.
//
// Per-session terminal events (`terminal.output.<id>`) rely on this: each
// pane subscribes and unsubscribes under its own event name, so one pane
// closing must not disturb another's stream. Prefer calling the function
// returned by `Events.On` directly in new code — it is precise and needs no
// bookkeeping.

import { Events as WailsEvents, Window as WailsWindow, Application as WailsApplication } from '@wailsio/runtime'

export interface WailsEvent {
  name: string
  data?: any
}

// name -> set of unsubscribe callbacks handed out for that name.
const subscriptions = new Map<string, Set<() => void>>()

function track(eventName: string, off: () => void): () => void {
  let set = subscriptions.get(eventName)
  if (!set) {
    set = new Set()
    subscriptions.set(eventName, set)
  }
  set.add(off)

  // Wrap so a direct call also drops the bookkeeping entry.
  return () => {
    try {
      off()
    } finally {
      set?.delete(off)
      if (set && set.size === 0) subscriptions.delete(eventName)
    }
  }
}

export const Events = {
  /**
   * Subscribe to a backend event. Returns an unsubscribe function; calling it
   * removes only this listener, leaving other subscribers to the same event
   * untouched.
   */
  On: (eventName: string, callback: (...data: any) => void): (() => void) => {
    const off = WailsEvents.On(eventName, (event: any) => {
      // v3 delivers a single event object; v2 handlers were written to take
      // the payload directly, so unwrap `data` and keep the shape they expect.
      callback(event?.data)
    })
    return track(eventName, off)
  },

  /** Subscribe for a single delivery, then unsubscribe automatically. */
  Once: (eventName: string, callback: (...data: any) => void): (() => void) => {
    const off = WailsEvents.Once(eventName, (event: any) => {
      callback(event?.data)
    })
    return track(eventName, off)
  },

  /**
   * Remove every listener this module registered for the given event
   * name(s) — the v2 `EventsOff` contract, reimplemented on top of v3's
   * unsubscribe functions.
   */
  Off: (eventName: string, ...additionalEventNames: string[]) => {
    for (const name of [eventName, ...additionalEventNames]) {
      const set = subscriptions.get(name)
      if (!set) continue
      // Copy first: each off() mutates the set via the tracked wrapper.
      for (const off of [...set]) off()
      subscriptions.delete(name)
    }
  },

  /** Emit an event to the Go backend and any other listening windows. */
  Emit: (eventName: string, ...data: any) => {
    return WailsEvents.Emit(eventName, data.length <= 1 ? data[0] : data)
  },
}

export const Log = {
  Debug: (message: string) => console.debug(`[Wails] ${message}`),
  Info: (message: string) => console.log(`[Wails] ${message}`),
  Warning: (message: string) => console.warn(`[Wails] ${message}`),
  Error: (message: string) => console.error(`[Wails] ${message}`),
}

// Window / application controls used by `pages/index.vue`.
//
// Under v2 these went through `window.runtime.WindowMaximise` / the bound
// `QuitApp` method. v3 exposes them directly: `Window` acts on the calling
// window, and `Application.Quit` begins the same teardown the Go-side quit
// triggers — ShouldQuit, then service shutdown in reverse registration order.
export const WindowRuntime = {
  Maximise: () => WailsWindow.Maximise(),
  UnMaximise: () => WailsWindow.UnMaximise(),
  IsMaximised: () => WailsWindow.IsMaximised(),
  Center: () => WailsWindow.Center(),
  Fullscreen: () => WailsWindow.Fullscreen(),
  UnFullscreen: () => WailsWindow.UnFullscreen(),
  IsFullscreen: () => WailsWindow.IsFullscreen(),
  /** Toggle window decorations at runtime — no restart needed. */
  SetFrameless: (frameless: boolean) => WailsWindow.SetFrameless(frameless),
  Quit: () => WailsApplication.Quit(),
}
