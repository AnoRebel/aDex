// Type definitions for Wails v2 Runtime
// Wails v2 injects `window.runtime` with flat functions like:
//   window.runtime.EventsOn(name, callback)
//   window.runtime.EventsEmit(name, ...data)
//   window.runtime.EventsOff(name)
//   window.runtime.Quit()
//   window.runtime.WindowMinimise()
//   window.runtime.LogInfo(message)
//   etc.

export interface WailsEvent {
  name: string
  data?: any
}

// Declare the Wails v2 runtime shape on window
declare global {
  interface Window {
    runtime?: {
      // Events (flat functions, NOT nested under .Events)
      EventsOn(eventName: string, callback: (...data: any) => void): () => void
      EventsOnce(eventName: string, callback: (...data: any) => void): () => void
      EventsOnMultiple(eventName: string, callback: (...data: any) => void, maxCallbacks: number): () => void
      EventsEmit(eventName: string, ...data: any): void
      EventsOff(eventName: string, ...additionalEventNames: string[]): void
      EventsOffAll(): void

      // Logging
      LogDebug(message: string): void
      LogInfo(message: string): void
      LogWarning(message: string): void
      LogError(message: string): void
      LogPrint(message: string): void
      LogTrace(message: string): void
      LogFatal(message: string): void

      // Window management
      Quit(): void
      Hide(): void
      Show(): void
      WindowMinimise(): void
      WindowUnminimise(): void
      WindowMaximise(): void
      WindowUnmaximise(): void
      WindowToggleMaximise(): void
      WindowFullscreen(): void
      WindowUnfullscreen(): void
      WindowCenter(): void
      WindowSetTitle(title: string): void
      WindowSetSize(width: number, height: number): void
      WindowGetSize(): { w: number; h: number }
      WindowSetPosition(x: number, y: number): void
      WindowGetPosition(): { x: number; y: number }
      WindowHide(): void
      WindowShow(): void
      WindowReload(): void
      WindowReloadApp(): void
      WindowSetAlwaysOnTop(b: boolean): void
      WindowSetBackgroundColour(R: number, G: number, B: number, A: number): void
      WindowIsFullscreen(): boolean
      WindowIsMaximised(): boolean
      WindowIsMinimised(): boolean
      WindowIsNormal(): boolean

      // Browser
      BrowserOpenURL(url: string): void

      // Clipboard
      ClipboardGetText(): string
      ClipboardSetText(text: string): void

      // Environment
      Environment(): any

      // Screen
      ScreenGetAll(): any[]
    }
  }
}

// Events API - wraps Wails v2 runtime event functions
export const Events = {
  On: (eventName: string, callback: (...data: any) => void): (() => void) | void => {
    if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
      return window.runtime.EventsOn(eventName, callback)
    } else {
      console.log(`[Events.On] ${eventName} - Wails runtime not available`)
    }
  },
  Once: (eventName: string, callback: (...data: any) => void): (() => void) | void => {
    if (typeof window !== 'undefined' && window.runtime?.EventsOnce) {
      return window.runtime.EventsOnce(eventName, callback)
    }
  },
  Off: (eventName: string, ...additionalEventNames: string[]) => {
    if (typeof window !== 'undefined' && window.runtime?.EventsOff) {
      window.runtime.EventsOff(eventName, ...additionalEventNames)
    }
  },
  Emit: (eventName: string, ...data: any) => {
    if (typeof window !== 'undefined' && window.runtime?.EventsEmit) {
      window.runtime.EventsEmit(eventName, ...data)
    } else {
      console.log(`[Events.Emit] ${eventName}:`, data)
    }
  }
}

export const Log = {
  Debug: (message: string) => {
    if (typeof window !== 'undefined' && window.runtime?.LogDebug) {
      window.runtime.LogDebug(message)
    } else {
      console.debug(`[Wails] ${message}`)
    }
  },
  Info: (message: string) => {
    if (typeof window !== 'undefined' && window.runtime?.LogInfo) {
      window.runtime.LogInfo(message)
    } else {
      console.log(`[Wails] ${message}`)
    }
  },
  Warning: (message: string) => {
    if (typeof window !== 'undefined' && window.runtime?.LogWarning) {
      window.runtime.LogWarning(message)
    } else {
      console.warn(`[Wails] ${message}`)
    }
  },
  Error: (message: string) => {
    if (typeof window !== 'undefined' && window.runtime?.LogError) {
      window.runtime.LogError(message)
    } else {
      console.error(`[Wails] ${message}`)
    }
  }
}

// Window management helpers
export const WindowRuntime = {
  Quit: () => {
    if (typeof window !== 'undefined' && window.runtime?.Quit) {
      window.runtime.Quit()
    }
  },
  Minimise: () => {
    if (typeof window !== 'undefined' && window.runtime?.WindowMinimise) {
      window.runtime.WindowMinimise()
    }
  },
  Maximise: () => {
    if (typeof window !== 'undefined' && window.runtime?.WindowMaximise) {
      window.runtime.WindowMaximise()
    }
  },
  ToggleMaximise: () => {
    if (typeof window !== 'undefined' && window.runtime?.WindowToggleMaximise) {
      window.runtime.WindowToggleMaximise()
    }
  },
  Fullscreen: () => {
    if (typeof window !== 'undefined' && window.runtime?.WindowFullscreen) {
      window.runtime.WindowFullscreen()
    }
  },
  Center: () => {
    if (typeof window !== 'undefined' && window.runtime?.WindowCenter) {
      window.runtime.WindowCenter()
    }
  },
  SetTitle: (title: string) => {
    if (typeof window !== 'undefined' && window.runtime?.WindowSetTitle) {
      window.runtime.WindowSetTitle(title)
    }
  }
}
