// tests/stores/network.spec.ts
//
// Section 5.C boundary test for the Pinia network store. The store was
// rewritten in section 5.0 to use dedicated coordinator bindings
// (`GetNetworkMetrics`, ...) instead of the broken `GetService('network')`
// pattern that caused the launch crash. Asserts the new wiring works.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";

const GetNetworkMetrics = vi.fn();
const GetNetworkConnections = vi.fn();
const GetNetworkAlerts = vi.fn();
const GetNetworkStatistics = vi.fn();
const GetNetworkConfig = vi.fn();
const UpdateNetworkConfig = vi.fn();
const GetBandwidthData = vi.fn();
const ResetNetworkService = vi.fn();
const IsNetworkMonitoring = vi.fn();
const StartNetworkMonitoring = vi.fn();
const StopNetworkMonitoring = vi.fn();
const ResolveNetworkAlert = vi.fn();
const ClearNetworkAlerts = vi.fn();

vi.mock("~/lib/wailsjs/coordinator", () => ({
  GetNetworkMetrics,
  GetNetworkConnections,
  GetNetworkAlerts,
  GetNetworkStatistics,
  GetNetworkConfig,
  UpdateNetworkConfig,
  GetBandwidthData,
  ResetNetworkService,
  IsNetworkMonitoring,
  StartNetworkMonitoring,
  StopNetworkMonitoring,
  ResolveNetworkAlert,
  ClearNetworkAlerts,
}));

describe("network store — section 5.C", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    [
      GetNetworkMetrics,
      GetNetworkConnections,
      GetNetworkAlerts,
      GetNetworkStatistics,
      GetNetworkConfig,
      UpdateNetworkConfig,
      GetBandwidthData,
      ResetNetworkService,
      IsNetworkMonitoring,
      StartNetworkMonitoring,
      StopNetworkMonitoring,
      ResolveNetworkAlert,
      ClearNetworkAlerts,
    ].forEach((m) => m.mockReset());
  });
  afterEach(() => {
    vi.resetModules();
  });

  it("fetchMetrics populates from the dedicated GetNetworkMetrics binding", async () => {
    GetNetworkMetrics.mockResolvedValueOnce({
      interfaces: [{ name: "eth0", isUp: true }],
      totalBytesSent: 100,
      totalBytesRecv: 200,
    });

    const { useNetworkStore } = await import("../../app/stores/network");
    const store = useNetworkStore();
    await store.fetchMetrics();

    expect(GetNetworkMetrics).toHaveBeenCalledOnce();
    expect(store.metrics).not.toBeNull();
    expect(store.metrics.totalBytesSent).toBe(100);
    expect(store.activeInterfaces.length).toBe(1);
  });

  it("startMonitoring uses the network-specific binding (not coordinator-level)", async () => {
    StartNetworkMonitoring.mockResolvedValueOnce(undefined);

    const { useNetworkStore } = await import("../../app/stores/network");
    const store = useNetworkStore();
    await store.startMonitoring();

    expect(StartNetworkMonitoring).toHaveBeenCalledOnce();
    expect(store.isMonitoring).toBe(true);
  });

  it("resolveAlert calls the backend AND mutates local alert state", async () => {
    ResolveNetworkAlert.mockResolvedValueOnce(undefined);

    const { useNetworkStore } = await import("../../app/stores/network");
    const store = useNetworkStore();
    store.alerts = [{ id: "a1", resolved: false }] as never;

    await store.resolveAlert("a1");
    expect(ResolveNetworkAlert).toHaveBeenCalledWith("a1");
    expect(store.alerts[0].resolved).toBe(true);
  });

  it("clearAlerts forwards to backend then strips resolved", async () => {
    ClearNetworkAlerts.mockResolvedValueOnce(undefined);

    const { useNetworkStore } = await import("../../app/stores/network");
    const store = useNetworkStore();
    store.alerts = [
      { id: "a", resolved: true },
      { id: "b", resolved: false },
    ] as never;

    await store.clearAlerts();
    expect(ClearNetworkAlerts).toHaveBeenCalledOnce();
    expect(store.alerts.length).toBe(1);
    expect(store.alerts[0].id).toBe("b");
  });
});
