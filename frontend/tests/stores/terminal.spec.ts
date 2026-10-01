// tests/stores/terminal.spec.ts
//
// Section 5.A boundary test for the Pinia terminal store. Mocks the Wails
// binding layer at the `~~/bindings` re-export point and asserts the store
// reduces backend payloads correctly across create / write / close.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";

// Mock the Wails binding surface used by stores/terminal.ts.
const CreateTerminal = vi.fn();
const CreateTerminalIn = vi.fn();
const WriteToTerminal = vi.fn();
const ResizeTerminal = vi.fn();
const CloseTerminal = vi.fn();

vi.mock("~/lib/wailsjs/coordinator", () => ({
  CreateTerminal: (cols: number, rows: number) => CreateTerminal(cols, rows),
  CreateTerminalIn: (cols: number, rows: number, cwd: string) =>
    CreateTerminalIn(cols, rows, cwd),
  WriteToTerminal: (id: string, data: unknown) => WriteToTerminal(id, data),
  ResizeTerminal: (id: string, cols: number, rows: number) => ResizeTerminal(id, cols, rows),
  CloseTerminal: (id: string) => CloseTerminal(id),
}));

// Stub the startup-cwd composable so the store falls back to an empty
// resolved path; the store will then use the legacy `CreateTerminal`
// path which the existing test assertions expect.
vi.mock("~/composables/useStartupCwd", () => ({
  useStartupCwd: () => ({
    resolve: async () => "",
    mode: { value: "home" },
    loadPaths: async () => ({ home: "", cwd: "" }),
  }),
}));

describe("terminal store — section 5.A", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    CreateTerminal.mockReset();
    CreateTerminalIn.mockReset();
    WriteToTerminal.mockReset();
    ResizeTerminal.mockReset();
    CloseTerminal.mockReset();
  });
  afterEach(() => {
    vi.resetModules();
  });

  it("createSession registers a new session keyed by the backend id", async () => {
    CreateTerminal.mockResolvedValueOnce({ id: "term-1" });

    const { useTerminalStore } = await import("../../app/stores/terminal");
    const store = useTerminalStore();

    const id = await store.createSession("MAIN");
    expect(id).toBe("term-1");
    expect(CreateTerminal).toHaveBeenCalledOnce();
    expect(store.allSessions.length).toBe(1);
    expect(store.allSessions[0].id).toBe("term-1");
    expect(store.allSessions[0].title).toBe("MAIN");
  });

  it("createSession surfaces backend failure as null + error state", async () => {
    CreateTerminal.mockRejectedValueOnce(new Error("PTY denied"));

    const { useTerminalStore } = await import("../../app/stores/terminal");
    const store = useTerminalStore();

    const id = await store.createSession("Boom");
    expect(id).toBeNull();
    // Error message surfaces on the store as `lastError`.
    expect(store.lastError).toBeTruthy();
    expect(String(store.lastError)).toContain("PTY denied");
  });

  it("sendInput forwards bytes to the active session", async () => {
    CreateTerminal.mockResolvedValueOnce({ id: "term-2" });
    WriteToTerminal.mockResolvedValueOnce(undefined);

    const { useTerminalStore } = await import("../../app/stores/terminal");
    const store = useTerminalStore();
    const id = await store.createSession("MAIN");
    expect(id).toBe("term-2");

    store.setActiveSession("term-2");
    await store.sendInput("ls -la\n");

    expect(WriteToTerminal).toHaveBeenCalledWith("term-2", "ls -la\n");
  });

  it("closeSession removes the session and forwards to backend", async () => {
    CreateTerminal.mockResolvedValueOnce({ id: "term-3" });
    CloseTerminal.mockResolvedValueOnce(undefined);

    const { useTerminalStore } = await import("../../app/stores/terminal");
    const store = useTerminalStore();
    await store.createSession("MAIN");

    const ok = await store.closeSession("term-3");
    expect(ok).toBe(true);
    expect(CloseTerminal).toHaveBeenCalledWith("term-3");
    expect(store.allSessions.length).toBe(0);
  });
});
