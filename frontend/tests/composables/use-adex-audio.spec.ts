// tests/composables/use-adex-audio.spec.ts

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useAdexAudio, _internals, EDEX_CUES, resolveCue } from "../../app/composables/useAdexAudio";

// Mock Howler so we don't hit real audio in jsdom.
const playSpy = vi.fn();
const volumeSpy = vi.fn();
vi.mock("howler", () => ({
  Howl: class {
    constructor(_: unknown) {}
    play() { playSpy(); }
    volume(_v: number) { volumeSpy(_v); }
  },
}));

// Cue playback moved from the webview to Go: the webview refuses to start
// audio before a user gesture, which silently dropped every boot-splash cue.
// The backend is therefore the primary path and Howler only the fallback, so
// that is what these tests assert against.
const playCueSpy = vi.fn().mockResolvedValue(undefined);
vi.mock("~/lib/wailsjs/coordinator", () => ({
  // The composable probes GetAvailableCues first and only uses the backend
  // when it answers; without this the probe fails and playback falls through
  // to the webview path.
  GetAvailableCues: () => Promise.resolve([...EDEX_CUES]),
  PlayCue: (...args: unknown[]) => playCueSpy(...args),
}));

describe("useAdexAudio — V2 audio engine", () => {
  beforeEach(() => {
    _internals.reset();
    playCueSpy.mockClear();
    playSpy.mockClear();
    volumeSpy.mockClear();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("ships the 13 canonical eDEX cues plus the aDex additions", () => {
    // The legacy 13 must all survive; aDex then adds its own cues on top
    // (keypress, destructive, gunshot, click), so assert the canonical set is
    // present rather than pinning a total that grows.
    for (const expected of [
      "alarm", "denied", "error", "expand", "folder", "granted", "info",
      "keyboard", "panels", "scan", "stdin", "stdout", "theme",
      "keypress", "destructive", "gunshot", "click",
    ]) {
      expect(EDEX_CUES).toContain(expected);
    }
  });

  it("resolves legacy event names to canonical cues", () => {
    // button_click routes to the click-tone cue; `keyboard` is reserved for
    // the mechanical typewriter click.
    expect(resolveCue("button_click")).toBe("click");
    expect(resolveCue("system_alert")).toBe("alarm");
    expect(resolveCue("notification")).toBe("info");
    expect(resolveCue("command_success")).toBe("granted");
    expect(resolveCue("system_startup")).toBe("scan");
    expect(resolveCue("totally_unknown")).toBeNull();
    // Canonical names pass through.
    expect(resolveCue("keyboard")).toBe("keyboard");
  });

  it("playCue routes to the Go backend", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    audio.playCue("keyboard");
    // The binding is imported lazily, so let the microtask queue drain.
    await new Promise((r) => setTimeout(r, 20));
    expect(playCueSpy).toHaveBeenCalled();
    expect(playCueSpy.mock.calls[0][0]).toBe("keyboard");
  });

  it("global mute suppresses playback", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    audio.setMuted(true);
    audio.playCue("panels");
    await new Promise((r) => setTimeout(r, 5));
    expect(playSpy).not.toHaveBeenCalled();
    expect(audio.playCount.value).toBe(0);
  });

  it("backgroundMuted suppresses playback", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    audio.setBackgroundMuted(true);
    audio.playCue("info");
    await new Promise((r) => setTimeout(r, 5));
    expect(playSpy).not.toHaveBeenCalled();
  });

  it("rate-limits stdin/stdout to 4 Hz", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    audio.playCue("stdout");
    audio.playCue("stdout");
    audio.playCue("stdout");
    await new Promise((r) => setTimeout(r, 5));
    // Only the first should fire — the next two are inside the 250 ms window.
    expect(audio.playCount.value).toBe(1);
  });

  it("synth pack does not call Howler", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    audio.setPack("synth");
    audio.playCue("granted");
    await new Promise((r) => setTimeout(r, 5));
    expect(playSpy).not.toHaveBeenCalled();
    // playCount still increments — the synth recipe is invoked.
    expect(audio.playCount.value).toBe(1);
  });

  it("persists settings round-trip", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    audio.setGlobalVolume(0.25);
    audio.setPack("synth");
    expect(localStorage.getItem(_internals.STORAGE_KEY)).toMatch(/synth/);
    expect(localStorage.getItem(_internals.STORAGE_KEY)).toMatch(/0\.25/);
  });

  it("ignores unknown cue names without throwing", async () => {
    const audio = useAdexAudio();
    await audio.initialize();
    expect(() => audio.playCue("totally-fake")).not.toThrow();
    expect(audio.playCount.value).toBe(0);
  });
});
