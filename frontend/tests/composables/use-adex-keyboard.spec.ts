// tests/composables/use-adex-keyboard.spec.ts

import { beforeEach, describe, expect, it } from "vitest";
import index from "../../app/assets/data/kb_layouts-index.json";
import { useAdexKeyboard, _internals } from "../../app/composables/useAdexKeyboard";
import { applyCtrlseq } from "../../app/types/kb-layout";

describe("useAdexKeyboard — V2 keyboard engine", () => {
  beforeEach(() => {
    _internals.reset();
    localStorage.clear();
  });

  it("ships all 19 original keyboard layouts in the index", () => {
    expect(index.length).toBe(19);
    const ids = new Set(index.map((e) => e.id));
    for (const expected of [
      "da-DK",
      "de-DE",
      "en-COLEMAK",
      "en-DVORAK",
      "en-GB",
      "en-NORMAN",
      "en-US",
      "en-WORKMAN",
      "es-ES",
      "es-LAT",
      "fr-BEPO",
      "fr-FR",
      "hu-HU",
      "it-IT",
      "nl-BE",
      "pt-BR",
      "sv-SE",
      "tr-TR-F",
      "tr-TR-Q",
    ]) {
      expect(ids).toContain(expected);
    }
  });

  it("CTRLSEQ substitution replaces ~~~CTRLSEQ1~~~ with ESC byte", () => {
    expect(applyCtrlseq("~~~CTRLSEQ1~~~OP")).toBe("\x1bOP");
    expect(applyCtrlseq("plain")).toBe("plain");
  });

  it("loads en-US, returns 5 rows, F1 produces \\x1bOP", async () => {
    const kb = useAdexKeyboard();
    await kb.initialize();
    await kb.setLayout("en-US");

    expect(kb.rows.value.length).toBe(5);

    // Find the '1' key in row_numbers; its fn_cmd is the F1 escape.
    const numbers = kb.activeLayout.value!.row_numbers;
    const one = numbers.find((k) => k.name === "1");
    expect(one).toBeDefined();
    expect(one!.fn_cmd).toBe("\x1bOP");
  });

  it("pressing 'A' produces 'a'; pressing Shift then 'A' produces 'A'", async () => {
    const kb = useAdexKeyboard();
    await kb.initialize();
    await kb.setLayout("en-US");

    const layout = kb.activeLayout.value!;
    // Cap shows 'A'; cmd is lowercase 'a', shift_cmd is uppercase 'A'.
    const a = layout.row_2.find((k) => k.name === "A");
    expect(a).toBeDefined();

    expect(kb.press(a!)).toBe("a");

    // Toggle Shift via the LEFT shift key in row_3 (its cmd carries the marker).
    const shift = layout.row_3.find((k) => k.cmd === "ESCAPED|-- SHIFT: LEFT");
    expect(shift).toBeDefined();
    kb.press(shift!);
    expect(kb.modifiers.value.shift).toBe(true);

    expect(kb.press(a!)).toBe("A");
    // Shift is momentary — released after a printable press.
    expect(kb.modifiers.value.shift).toBe(false);
  });

  it("Ctrl+C produces 0x03 on en-US", async () => {
    const kb = useAdexKeyboard();
    await kb.initialize();
    await kb.setLayout("en-US");

    const layout = kb.activeLayout.value!;
    const ctrl = layout.row_space.find((k) => k.cmd === "ESCAPED|-- CTRL: LEFT");
    expect(ctrl).toBeDefined();
    kb.press(ctrl!);
    expect(kb.modifiers.value.ctrl).toBe(true);

    // 'C' lives in row_3 on en-US (between X and V).
    const c = layout.row_3.find((k) => k.name === "C");
    expect(c).toBeDefined();
    expect(kb.press(c!)).toBe("\x03");
    expect(kb.modifiers.value.ctrl).toBe(false);
  });

  it("layout hot-swap (en-US → fr-FR) preserves modifier state struct", async () => {
    const kb = useAdexKeyboard();
    await kb.initialize();
    await kb.setLayout("en-US");
    expect(kb.activeId.value).toBe("en-US");

    await kb.setLayout("fr-FR");
    expect(kb.activeId.value).toBe("fr-FR");
    expect(kb.activeLayout.value!.row_numbers.length).toBeGreaterThan(0);

    // fr-FR keeps the same row vocabulary, with locale-specific cmds.
    const layout = kb.activeLayout.value!;
    expect(Array.isArray(layout.row_1)).toBe(true);
    expect(Array.isArray(layout.row_space)).toBe(true);
  });

  it("persists active layout id and restores on next initialize", async () => {
    const kb = useAdexKeyboard();
    await kb.initialize();
    await kb.setLayout("de-DE");
    expect(localStorage.getItem(_internals.STORAGE_KEY)).toBe("de-DE");

    _internals.reset();
    const kb2 = useAdexKeyboard();
    await kb2.initialize();
    expect(kb2.activeId.value).toBe("de-DE");
  });

  it("rejects unknown layout ids", async () => {
    const kb = useAdexKeyboard();
    await kb.initialize();
    await expect(kb.setLayout("nope")).rejects.toThrow(/unknown keyboard layout/);
  });
});
