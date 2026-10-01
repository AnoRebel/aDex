// tests/components/window-controls.spec.ts
//
// Boundary tests for the window controls in pages/index.vue. The full page
// pulls in every store plus xterm, so these import the runtime facade
// directly and assert it calls into Wails.
//
// These were written against Wails v2, where the runtime was a global
// (`window.runtime.Quit`, `window.runtime.WindowToggleMaximise`). v3 has no
// such global: `app/lib/wailsjs/runtime.ts` imports `@wailsio/runtime` and
// re-exports a thin facade, so the module is what has to be mocked. The old
// spec also exercised a `ToggleMaximise()` helper that has never existed on
// the v3 facade — maximise/unmaximise are separate calls, and pages/index.vue
// chooses between them after awaiting IsMaximised().

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const WailsWindow = {
  Maximise: vi.fn(),
  UnMaximise: vi.fn(),
  IsMaximised: vi.fn().mockResolvedValue(false),
  Center: vi.fn(),
  Fullscreen: vi.fn(),
  UnFullscreen: vi.fn(),
  IsFullscreen: vi.fn().mockResolvedValue(false),
  SetFrameless: vi.fn(),
};

const WailsApplication = { Quit: vi.fn() };

vi.mock("@wailsio/runtime", () => ({
  Window: WailsWindow,
  Application: WailsApplication,
  Events: { On: vi.fn(), Off: vi.fn(), Emit: vi.fn() },
}));

describe("WindowRuntime facade", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetModules();
  });

  it("Quit() forwards to Application.Quit", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");
    WindowRuntime.Quit();
    expect(WailsApplication.Quit).toHaveBeenCalledOnce();
  });

  it("Maximise() and UnMaximise() forward to the window", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");

    WindowRuntime.Maximise();
    expect(WailsWindow.Maximise).toHaveBeenCalledOnce();

    WindowRuntime.UnMaximise();
    expect(WailsWindow.UnMaximise).toHaveBeenCalledOnce();
  });

  it("IsMaximised() resolves the window's state", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");
    await expect(WindowRuntime.IsMaximised()).resolves.toBe(false);
    expect(WailsWindow.IsMaximised).toHaveBeenCalledOnce();
  });

  it("Fullscreen() and UnFullscreen() forward to the window", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");

    WindowRuntime.Fullscreen();
    expect(WailsWindow.Fullscreen).toHaveBeenCalledOnce();

    WindowRuntime.UnFullscreen();
    expect(WailsWindow.UnFullscreen).toHaveBeenCalledOnce();
  });

  it("SetFrameless() passes the flag through", async () => {
    const { WindowRuntime } = await import("../../app/lib/wailsjs/runtime");
    WindowRuntime.SetFrameless(true);
    expect(WailsWindow.SetFrameless).toHaveBeenCalledWith(true);
  });
});
