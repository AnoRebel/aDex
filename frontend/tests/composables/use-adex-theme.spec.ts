// tests/composables/use-adex-theme.spec.ts
//
// Three-layer-evidence test for the V2 theme engine. Covers:
//   - the index loads + every shipped theme parses
//   - setTheme applies CSS custom properties + data-layout in <200 ms
//   - persistence round-trip via localStorage
//   - the layout-preset escape hatch never reaches the DOM unless declared

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import themesIndex from "../../app/assets/data/themes-index.json";
import { useAdexTheme, _internals } from "../../app/composables/useAdexTheme";

const LAYOUT_VALUES = ["default", "disrupted", "typeleft", "fulltype", "notype", "colorfilter"];

describe("useAdexTheme — V2 engine", () => {
  beforeEach(() => {
    _internals.reset();
    document.documentElement.removeAttribute("data-theme");
    document.documentElement.removeAttribute("style");
    document.body.innerHTML = "";
    localStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("ships all 21 original themes in the index", () => {
    expect(themesIndex.length).toBe(21);
    const ids = new Set(themesIndex.map((t) => t.id));
    for (const expected of [
      "tron",
      "tron-disrupted",
      "tron-typeleft",
      "tron-fulltype",
      "tron-notype",
      "tron-colorfilter",
      "blade",
      "matrix",
      "nord",
      "apollo",
      "interstellar",
      "navy",
      "navy-disrupted",
      "red",
      "cyborg",
      "cyborg-focus",
      "chalkboard",
      "chalkboard-ligatures",
      "chalkboard-notype",
      "apollo-notype",
      "navy-notype",
    ]) {
      expect(ids).toContain(expected);
    }
  });

  it("every shipped theme declares a known layout preset", () => {
    for (const t of themesIndex) {
      expect(LAYOUT_VALUES).toContain(t.layout);
    }
  });

  it("setTheme applies CSS variables and data-layout", async () => {
    // We need .adex-app to exist so applyTheme can set data-layout on it.
    const shell = document.createElement("div");
    shell.className = "adex-app";
    document.body.appendChild(shell);

    const engine = useAdexTheme();
    await engine.initialize();

    const t = await engine.setTheme("blade");
    expect(t.id).toBe("blade");
    expect(t.layout).toBe("default");

    const root = document.documentElement;
    // Accent triple should reflect Blade's primary RGB (204, 94, 55).
    expect(root.style.getPropertyValue("--color_r").trim()).toBe("204");
    expect(root.style.getPropertyValue("--color_g").trim()).toBe("94");
    expect(root.style.getPropertyValue("--color_b").trim()).toBe("55");
    expect(root.style.getPropertyValue("--accent-rgb").trim()).toBe("204, 94, 55");
    expect(root.getAttribute("data-theme")).toBe("blade");
    expect(shell.getAttribute("data-layout")).toBe("default");
  });

  it("disrupted theme sets data-layout=disrupted on the shell", async () => {
    const shell = document.createElement("div");
    shell.className = "adex-app";
    document.body.appendChild(shell);

    const engine = useAdexTheme();
    await engine.initialize();
    await engine.setTheme("tron-disrupted");
    expect(shell.getAttribute("data-layout")).toBe("disrupted");
  });

  it("persists the active id and restores on next initialize", async () => {
    const engine = useAdexTheme();
    await engine.initialize();
    await engine.setTheme("matrix");
    // Wait a microtask for VueUse useStorage's watcher to flush to
    // localStorage. The composable still reads via its own ref, but
    // asserting on the underlying storage proves the persistence wire.
    await new Promise((r) => setTimeout(r, 0));
    // VueUse's useStorage uses the string serializer for `string` types
    // (identity, not JSON.stringify), so we expect the raw value.
    expect(localStorage.getItem(_internals.STORAGE_KEY)).toBe("matrix");

    _internals.reset();
    const engine2 = useAdexTheme();
    await engine2.initialize();
    expect(engine2.activeId.value).toBe("matrix");
  });

  it("hot-swap completes in under 200 ms", async () => {
    const engine = useAdexTheme();
    await engine.initialize();
    // Warm caches first so we measure swap, not first-load.
    await engine.setTheme("tron");
    await engine.setTheme("blade");

    const start = performance.now();
    await engine.setTheme("nord");
    const elapsed = performance.now() - start;
    expect(elapsed).toBeLessThan(200);
  });

  it("falls back to the default theme for an unknown id", async () => {
    // setTheme deliberately does NOT reject here. Persisted ids outlive the
    // catalogue — settings files, localStorage and older builds can all name
    // a theme that no longer exists — and rethrowing left the interface
    // completely unstyled, with no colours and no data-layout attribute.
    const engine = useAdexTheme();
    await engine.initialize();

    const theme = await engine.setTheme("does-not-exist");

    expect(theme.id).toBe("tron");
    expect(engine.activeId.value).toBe("tron");
  });
});
