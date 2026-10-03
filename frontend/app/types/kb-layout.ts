// types/kb-layout.ts
//
// Keyboard layout schema. Verbatim port of the upstream eDEX-UI shape
// because that schema's `cmd` / `shift_cmd` / `ctrl_cmd` / `alt_cmd` /
// `altshift_cmd` / `fn_cmd` / `capslck_cmd` fields are required to render
// the alt-graph chars, dead keys, and per-locale F-key escapes correctly
// across all 19 ported layouts.

/** A single key on the on-screen keyboard. Field semantics match the
 *  original `keyboard.class.js`. */
export interface KbKey {
  /** Glyph or label rendered on the cap. May contain ESCAPED|-- ICON:... or
   *  ESCAPED|-- SHIFT: LEFT-style markers for non-printable modifiers. */
  name: string;
  /** Bytes injected into the terminal when pressed (no modifiers). */
  cmd: string;

  /** Glyph rendered when Shift is held. */
  shift_name?: string;
  /** Bytes injected with Shift. Defaults to cmd if omitted. */
  shift_cmd?: string;

  /** Bytes injected with Ctrl. */
  ctrl_cmd?: string;
  /** Bytes injected with Alt (alt-graph). */
  alt_cmd?: string;
  /** Bytes injected with Alt+Shift. */
  altshift_cmd?: string;

  /** Glyph and bytes when Fn is active. */
  fn_name?: string;
  fn_cmd?: string;

  /** Some locales remap the caps-lock state (most don't — defaults to shift_cmd). */
  capslck_cmd?: string;
}

/** A layout's row vocabulary matches the upstream files exactly. */
export interface KbLayout {
  row_numbers: KbKey[];
  row_1: KbKey[];
  row_2: KbKey[];
  row_3: KbKey[];
  row_space: KbKey[];
}

export interface KbLayoutIndexEntry {
  id: string;
  displayName: string;
  /** Path relative to `assets/data/`. */
  file: string;
}
export type KbLayoutIndex = KbLayoutIndexEntry[];

/* Modifier state surfaced to consumers (mirrors the original
 * `data-isShiftOn`, `data-isCtrlOn`, etc. attributes). */
export interface KbModifiers {
  shift: boolean;
  capslck: boolean;
  ctrl: boolean;
  alt: boolean;
  fn: boolean;
}

export const EMPTY_MODIFIERS: KbModifiers = {
  shift: false,
  capslck: false,
  ctrl: false,
  alt: false,
  fn: false,
};

/* CTRLSEQ substitution table ---------------------------------------------
 *
 * The upstream JSON uses `~~~CTRLSEQ<n>~~~` placeholders so layout files
 * stay independent of the host's escape-sequence vocabulary. The original
 * runtime left the table empty (edex-ui/src/classes/keyboard.class.js:6),
 * which silently broke Ctrl+letter and the F-key sequences. We populate
 * it here with the standard xterm ESC + ASCII control-character mapping
 * derived from `en-US.json`'s usage:
 *
 *   1 → ESC (used by F-keys, arrows, and ALT-prefixed chars)
 *   2 → FS  (Ctrl+\)
 *   3 → GS  (Ctrl+])
 *   4..5 unused in en-US — kept empty for forward compatibility
 *   6 → Ctrl+Q  ... 21 → Ctrl+B
 *
 * Each Ctrl+<letter> is `<letter>.toUpperCase().charCodeAt(0) - 64`, which
 * yields the correct ASCII control byte (Ctrl+A → 0x01, Ctrl+C → 0x03).
 */
export const CTRLSEQ: ReadonlyArray<string> = Object.freeze([
  "",      // 0  unused
  "\x1b",  // 1  ESC
  "\x1c",  // 2  Ctrl+\  (FS)
  "\x1d",  // 3  Ctrl+]  (GS)
  "",      // 4
  "",      // 5
  "\x11",  // 6  Ctrl+Q  (DC1)
  "\x17",  // 7  Ctrl+W  (ETB)
  "\x12",  // 8  Ctrl+R  (DC2)
  "\x14",  // 9  Ctrl+T  (DC4)
  "\x19",  // 10 Ctrl+Y  (EM)
  "\x15",  // 11 Ctrl+U  (NAK)
  "\x10",  // 12 Ctrl+P  (DLE)
  "\x01",  // 13 Ctrl+A  (SOH)
  "\x13",  // 14 Ctrl+S  (DC3)
  "\x04",  // 15 Ctrl+D  (EOT)
  "\x06",  // 16 Ctrl+F  (ACK)
  "\x1a",  // 17 Ctrl+Z  (SUB)
  "\x18",  // 18 Ctrl+X  (CAN)
  "\x03",  // 19 Ctrl+C  (ETX) — interrupt
  "\x16",  // 20 Ctrl+V  (SYN)
  "\x02",  // 21 Ctrl+B  (STX)
]);

/** Replace every `~~~CTRLSEQ<n>~~~` token in `s` with CTRLSEQ[n]. */
export function applyCtrlseq(s: string): string {
  let out = s;
  for (let i = 1; i < CTRLSEQ.length; i++) {
    out = out.replaceAll(`~~~CTRLSEQ${i}~~~`, CTRLSEQ[i]);
  }
  return out;
}

/** Walk every cmd-bearing field of a layout, substituting CTRLSEQ tokens
 *  in place. Returns a new object; the input is not mutated. */
export function preprocessLayout(layout: KbLayout): KbLayout {
  const fix = (k: KbKey): KbKey => {
    const out: KbKey = { name: k.name, cmd: applyCtrlseq(k.cmd) };
    if (k.shift_name !== undefined) out.shift_name = k.shift_name;
    if (k.shift_cmd !== undefined) out.shift_cmd = applyCtrlseq(k.shift_cmd);
    if (k.ctrl_cmd !== undefined) out.ctrl_cmd = applyCtrlseq(k.ctrl_cmd);
    if (k.alt_cmd !== undefined) out.alt_cmd = applyCtrlseq(k.alt_cmd);
    if (k.altshift_cmd !== undefined) out.altshift_cmd = applyCtrlseq(k.altshift_cmd);
    if (k.fn_name !== undefined) out.fn_name = k.fn_name;
    if (k.fn_cmd !== undefined) out.fn_cmd = applyCtrlseq(k.fn_cmd);
    if (k.capslck_cmd !== undefined) out.capslck_cmd = applyCtrlseq(k.capslck_cmd);
    return out;
  };
  return {
    row_numbers: layout.row_numbers.map(fix),
    row_1: layout.row_1.map(fix),
    row_2: layout.row_2.map(fix),
    row_3: layout.row_3.map(fix),
    row_space: layout.row_space.map(fix),
  };
}

/** Pick the cmd to send for a key given the active modifier state.
 *  Mirrors the dispatch order in `keyboard.class.js:399-405`. */
export function resolveCmd(key: KbKey, m: KbModifiers): string {
  // alt+shift > alt > ctrl > capslck > shift > fn > base
  if (m.alt && m.shift && key.altshift_cmd !== undefined) return key.altshift_cmd;
  if (m.alt && key.alt_cmd !== undefined) return key.alt_cmd;
  if (m.ctrl && key.ctrl_cmd !== undefined) return key.ctrl_cmd;
  if (m.capslck && key.capslck_cmd !== undefined) return key.capslck_cmd;
  if ((m.shift || m.capslck) && key.shift_cmd !== undefined) return key.shift_cmd;
  if (m.fn && key.fn_cmd !== undefined) return key.fn_cmd;
  return key.cmd;
}

/** Pick the visible label for a key given modifier state. */
export function resolveName(key: KbKey, m: KbModifiers): string {
  if (m.fn && key.fn_name !== undefined) return key.fn_name;
  if ((m.shift || m.capslck) && key.shift_name !== undefined) return key.shift_name;
  return key.name;
}

/** Sticky-modifier markers in the cmd field, e.g. `ESCAPED|-- SHIFT: LEFT`. */
export const ESCAPED_PREFIX = "ESCAPED|-- ";

export function isModifierToggle(cmd: string): keyof KbModifiers | null {
  if (!cmd.startsWith(ESCAPED_PREFIX)) return null;
  const tag = cmd.slice(ESCAPED_PREFIX.length);
  if (tag.startsWith("CAPSLCK")) return "capslck";
  if (tag.startsWith("SHIFT")) return "shift";
  if (tag.startsWith("CTRL")) return "ctrl";
  if (tag.startsWith("ALT")) return "alt";
  if (tag.startsWith("FN")) return "fn";
  return null;
}
