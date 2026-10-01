// tests/stores/filesystem.spec.ts
//
// Section 5.D boundary test for the Pinia filesystem store. The store
// goes through `useWails().filesystem.readDirectory(path)`, which itself
// calls the canonical `~/lib/wailsjs/coordinator` shim. We mock the shim
// and assert the store reduces real DirectoryEntry payloads correctly —
// using the camelCase field names the backend now ships (after the
// section 5.D JSON-tag fix in `backend/services/filesystem/service.go`).

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";

const ReadDirectory = vi.fn();

vi.mock("~/lib/wailsjs/coordinator", () => ({
  ReadDirectory,
}));

describe("filesystem store — section 5.D", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    ReadDirectory.mockReset();
  });
  afterEach(() => {
    vi.resetModules();
  });

  it("fetchDirectory splits backend payload into files vs directories", async () => {
    ReadDirectory.mockResolvedValueOnce([
      { name: "subdir", path: "/test/subdir", isDir: true, size: 0, modTime: "2026-05-06T00:00:00Z" },
      { name: "file.txt", path: "/test/file.txt", isDir: false, size: 100, modTime: "2026-05-06T00:00:00Z" },
      { name: "image.png", path: "/test/image.png", isDir: false, size: 2048, modTime: "2026-05-06T00:00:00Z" },
    ]);

    const { useFilesystemStore } = await import("../../app/stores/filesystem");
    const store = useFilesystemStore();
    await store.fetchDirectory("/test");

    expect(ReadDirectory).toHaveBeenCalledWith("/test");
    expect(store.directories.length).toBe(1);
    expect(store.directories[0].name).toBe("subdir");
    expect(store.files.length).toBe(2);
    expect(store.files.find((f) => f.name === "image.png")?.size).toBe(2048);
    expect(store.currentPath).toBe("/test");
  });

  it("empty backend payload yields empty arrays, not zeros", async () => {
    ReadDirectory.mockResolvedValueOnce([]);

    const { useFilesystemStore } = await import("../../app/stores/filesystem");
    const store = useFilesystemStore();
    await store.fetchDirectory("/empty");

    expect(store.files.length).toBe(0);
    expect(store.directories.length).toBe(0);
    expect(store.currentPath).toBe("/empty");
  });

  it("backend errors do not crash the store (graceful degradation)", async () => {
    ReadDirectory.mockRejectedValueOnce(new Error("permission denied"));

    const { useFilesystemStore } = await import("../../app/stores/filesystem");
    const store = useFilesystemStore();

    // The composable layer (`useWails().filesystem.readDirectory`) catches
    // binding errors and returns []. The store then proceeds normally:
    // empty files/directories, no thrown exception, no alert. This is the
    // documented contract — alerts are emitted only when the *store* itself
    // catches an exception, which happens for non-binding code paths
    // (createDirectory, deleteFile, etc.).
    await expect(store.fetchDirectory("/forbidden")).resolves.toBeUndefined();
    expect(store.files.length).toBe(0);
    expect(store.directories.length).toBe(0);
    // currentPath still updates because the call resolved (even if to []).
    expect(store.currentPath).toBe("/forbidden");
  });
});
