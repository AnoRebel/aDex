// tests/stores/system.spec.ts
//
// Section 5.B boundary test for the Pinia system store. Mocks the Wails
// binding layer at `~/lib/wailsjs/coordinator` and asserts the store
// reduces backend payloads correctly. Specifically asserts the previously-
// observed "shows zeros" symptom doesn't return: when the binding returns
// real data, the store must surface it (not all-zero defaults).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";

const GetSystemInfo = vi.fn();
const GetCPUUsage = vi.fn();
const GetMemoryUsage = vi.fn();
const GetDiskUsage = vi.fn();
const GetTopProcesses = vi.fn();
const GetNetworkInfo = vi.fn();

vi.mock("~/lib/wailsjs/coordinator", () => ({
  GetSystemInfo,
  GetCPUUsage,
  GetMemoryUsage,
  GetDiskUsage,
  GetTopProcesses,
  GetNetworkInfo,
}));

describe("system store — section 5.B", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    GetSystemInfo.mockReset();
    GetCPUUsage.mockReset();
    GetMemoryUsage.mockReset();
    GetDiskUsage.mockReset();
    GetTopProcesses.mockReset();
    GetNetworkInfo.mockReset();
  });
  afterEach(() => {
    vi.resetModules();
  });

  it("fetchSystemInfo populates the store from coordinator binding", async () => {
    GetSystemInfo.mockResolvedValueOnce({
      Hostname: "testbox",
      OS: "linux",
      Architecture: "x86_64",
      Uptime: 86_400_000_000_000, // 1 day in ns
      KernelVersion: "6.1.0",
    });

    const { useSystemStore } = await import("../../app/stores/system");
    const store = useSystemStore();
    await store.fetchSystemInfo();

    expect(GetSystemInfo).toHaveBeenCalledOnce();
    expect(store.systemInfo).not.toBeNull();
    expect(store.systemInfo?.hostname).toBe("testbox");
    expect(store.systemInfo?.os).toBe("linux");
  });

  it("fetchSystemStats merges CPU + memory + disk + processes", async () => {
    GetCPUUsage.mockResolvedValueOnce({
      usage: 42.5,
      cores: [40, 45],
      coreCount: 2,
      modelName: "Test CPU",
      frequency: 3200,
    });
    GetMemoryUsage.mockResolvedValueOnce({
      total: 16_000_000_000,
      used: 8_000_000_000,
      free: 8_000_000_000,
      usage: 50,
    });
    GetDiskUsage.mockResolvedValueOnce([
      { mountpoint: "/", total: 500_000_000_000, used: 250_000_000_000, percent: 50 },
    ]);
    GetTopProcesses.mockResolvedValueOnce([
      { PID: 1, Name: "init", CPUPercent: 0.1, MemoryPercent: 0.2 },
      { PID: 100, Name: "node", CPUPercent: 5, MemoryPercent: 3 },
    ]);

    const { useSystemStore } = await import("../../app/stores/system");
    const store = useSystemStore();
    await store.fetchSystemStats();

    expect(store.systemStats).not.toBeNull();
    // The "shows zeros" regression: if useWails or the store dropped the
    // payload, every field would be 0. Assert the real values flow through.
    expect(store.systemStats?.cpu.usage).toBe(42.5);
    expect(store.systemStats?.cpu.coreCount).toBe(2);
    expect(store.systemStats?.memory.total).toBe(16_000_000_000);
    expect(store.systemStats?.memory.usage).toBe(50);
    expect(Array.isArray(store.systemStats?.disk)).toBe(true);
    expect(store.processes.length).toBe(2);
    expect(store.processes[0].name).toBe("init");
  });

  it("zero fallback when the binding throws", async () => {
    GetCPUUsage.mockRejectedValueOnce(new Error("backend down"));
    GetMemoryUsage.mockRejectedValueOnce(new Error("backend down"));
    GetDiskUsage.mockRejectedValueOnce(new Error("backend down"));
    GetTopProcesses.mockRejectedValueOnce(new Error("backend down"));

    const { useSystemStore } = await import("../../app/stores/system");
    const store = useSystemStore();

    // The composable layer (`useWails().system.*`) catches binding errors
    // and returns zero/empty payloads as a graceful-degradation contract.
    // The store should therefore complete without throwing and surface
    // zeros — exactly what we want when the backend is unreachable.
    await expect(store.fetchSystemStats()).resolves.toBeUndefined();
    if (store.systemStats) {
      expect(store.systemStats.cpu.usage).toBe(0);
      expect(store.systemStats.memory.total).toBe(0);
    }
  });
});
