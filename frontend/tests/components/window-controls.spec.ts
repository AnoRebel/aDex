// tests/components/window-controls.spec.ts
// Boundary tests for the close + maximize controls in pages/index.vue.
// We don't mount the full page (it pulls every store + xterm); instead we
// import the runtime helpers directly and assert they call into Wails.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { installWailsMocks, uninstallWailsMocks } from "../_helpers/wails";

describe("WindowRuntime helpers", () => {
  beforeEach(() => {
    installWailsMocks();
  });
  afterEach(() => {
    uninstallWailsMocks();
    vi.resetModules();
  });

  it("Quit() forwards to window.runtime.Quit", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");
    WindowRuntime.Quit();
    const w = globalThis as unknown as { window?: { runtime?: { Quit?: ReturnType<typeof vi.fn> } } };
    const Quit = w.window?.runtime?.Quit;
    expect(Quit).toBeDefined();
    expect(Quit).toHaveBeenCalledOnce();
  });

  it("ToggleMaximise() forwards to window.runtime.WindowToggleMaximise", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");
    WindowRuntime.ToggleMaximise();
    const w = globalThis as unknown as {
      window?: { runtime?: { WindowToggleMaximise?: ReturnType<typeof vi.fn> } };
    };
    const T = w.window?.runtime?.WindowToggleMaximise;
    expect(T).toBeDefined();
    expect(T).toHaveBeenCalledOnce();
  });

  it("is a no-op when the runtime is missing", async () => {
    uninstallWailsMocks();
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");
    expect(() => WindowRuntime.Quit()).not.toThrow();
    expect(() => WindowRuntime.ToggleMaximise()).not.toThrow();
  });
});
