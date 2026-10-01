// tests/frontend/_helpers.ts
//
// Shared scaffolding for Vitest store and composable tests.
// Mocks the Wails v2 globals (`window.go`, `window.runtime`) so Pinia stores
// and composables can be exercised without a live runtime.

import { vi } from "vitest";

/**
 * Install a minimal Wails v2 runtime + ServiceCoordinator on `window`.
 * Returns the mock bag so tests can assert calls.
 */
export interface CoordinatorMethods {
  [k: string]: ReturnType<typeof vi.fn>;
}

export interface InstalledMocks {
  coordinator: CoordinatorMethods;
  events: {
    On: ReturnType<typeof vi.fn>;
    Once: ReturnType<typeof vi.fn>;
    Off: ReturnType<typeof vi.fn>;
    Emit: ReturnType<typeof vi.fn>;
    listeners: Map<string, Array<(...args: unknown[]) => void>>;
    fire(eventName: string, ...data: unknown[]): void;
  };
  runtime: {
    Quit: ReturnType<typeof vi.fn>;
    WindowToggleMaximise: ReturnType<typeof vi.fn>;
  };
}

export function installWailsMocks(
  methods: Record<string, unknown> = {},
): InstalledMocks {
  const coordinator: CoordinatorMethods = {};
  for (const [name, impl] of Object.entries(methods)) {
    coordinator[name] = typeof impl === "function" ? vi.fn(impl as never) : vi.fn().mockResolvedValue(impl);
  }

  const listeners = new Map<string, Array<(...args: unknown[]) => void>>();
  const On = vi.fn((event: string, cb: (...a: unknown[]) => void) => {
    const arr = listeners.get(event) ?? [];
    arr.push(cb);
    listeners.set(event, arr);
    return () => {
      const idx = arr.indexOf(cb);
      if (idx >= 0) arr.splice(idx, 1);
    };
  });
  const Off = vi.fn((event: string) => listeners.delete(event));
  const Once = vi.fn((event: string, cb: (...a: unknown[]) => void) => {
    const wrap = (...a: unknown[]) => {
      cb(...a);
      Off(event);
    };
    return On(event, wrap);
  });
  const Emit = vi.fn();

  const Quit = vi.fn();
  const WindowToggleMaximise = vi.fn();

  // window.go.coordinator.ServiceCoordinator — the *correct* path.
  // Also mirror at window.go.main.ServiceCoordinator for the legacy shim.
  const w = globalThis as unknown as {
    window?: { go?: unknown; runtime?: unknown };
    go?: unknown;
    runtime?: unknown;
  };
  const targetWindow = (w.window ?? w) as Record<string, unknown>;

  const goRoot = {
    coordinator: { ServiceCoordinator: coordinator },
    main: { ServiceCoordinator: coordinator },
  };
  targetWindow.go = goRoot;
  targetWindow.runtime = {
    EventsOn: On,
    EventsOnce: Once,
    EventsOff: Off,
    EventsEmit: Emit,
    Quit,
    WindowToggleMaximise,
    WindowMaximise: vi.fn(),
    WindowMinimise: vi.fn(),
    LogInfo: vi.fn(),
    LogError: vi.fn(),
    LogWarning: vi.fn(),
    LogDebug: vi.fn(),
    ClipboardGetText: vi.fn(() => ""),
    ClipboardSetText: vi.fn(),
  };

  return {
    coordinator,
    events: {
      On,
      Once,
      Off,
      Emit,
      listeners,
      fire(eventName: string, ...data: unknown[]) {
        for (const cb of listeners.get(eventName) ?? []) cb(...data);
      },
    },
    runtime: { Quit, WindowToggleMaximise },
  };
}

export function uninstallWailsMocks() {
  const w = globalThis as unknown as { window?: Record<string, unknown> };
  if (w.window) {
    delete w.window.go;
    delete w.window.runtime;
  }
  // happy-dom case: globalThis === window
  delete (globalThis as Record<string, unknown>).go;
  delete (globalThis as Record<string, unknown>).runtime;
}
