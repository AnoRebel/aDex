// types/adex-theme.ts
//
// V2 theme schema — port-compatible with eDEX-UI but with the unsafe
// `injectCSS` field replaced by a typed `layout` enum. See:
//   frontend/app/assets/css/layouts/<layout>.css
//
// Themes live as individual JSON files at:
//   frontend/app/assets/data/themes/<id>.json
// indexed by:
//   frontend/app/assets/data/themes-index.json

export type LayoutPreset =
  | "default"
  | "disrupted"
  | "typeleft"
  | "fulltype"
  | "notype"
  | "colorfilter";

export interface AdexThemeColors {
  /** Primary accent — RGB component triple (0..255). Composed into
   *  `--accent-rgb`, `--accent`, `--accent-dim`, etc. */
  r: number;
  g: number;
  b: number;
  /** Pure black (used as `--bg-deep`). */
  black: string;
  /** Slightly-lifted black for terminal background and panel surfaces. */
  light_black: string;
  /** Mid-grey reserved for inactive controls. */
  grey: string;
}

export interface AdexThemeCssVars {
  font_main: string;
  font_main_light: string;
}

export type TerminalCursorStyle = "block" | "underline" | "bar";

export interface AdexThemeTerminal {
  fontFamily: string;
  cursorStyle: TerminalCursorStyle;
  foreground: string;
  background: string;
  cursor: string;
  cursorAccent: string;
  /** rgba() string */
  selection: string;
}

export interface AdexThemeGlobe {
  base: string;
  marker: string;
  pin: string;
  satellite: string;
}

export interface AdexTheme {
  id: string;
  displayName: string;
  /** Drives `data-layout` on the shell root + selects a layout-preset CSS. */
  layout: LayoutPreset;
  colors: AdexThemeColors;
  cssvars: AdexThemeCssVars;
  terminal: AdexThemeTerminal;
  globe: AdexThemeGlobe;
}

export interface AdexThemeIndexEntry {
  id: string;
  displayName: string;
  layout: LayoutPreset;
  /** Pre-composed `rgb(r, g, b)` string — convenient for previews. */
  primary: string;
  /** Path relative to `assets/data/`. */
  file: string;
}

export type AdexThemeIndex = AdexThemeIndexEntry[];

/* Runtime guards ---------------------------------------------------------- */

const LAYOUT_VALUES: ReadonlySet<LayoutPreset> = new Set([
  "default",
  "disrupted",
  "typeleft",
  "fulltype",
  "notype",
  "colorfilter",
]);

export function isLayoutPreset(value: unknown): value is LayoutPreset {
  return typeof value === "string" && LAYOUT_VALUES.has(value as LayoutPreset);
}

/** Parse + validate a JSON object loaded from `themes/<id>.json`.
 *  Throws on malformed input so callers can surface a typed error. */
export function parseAdexTheme(raw: unknown): AdexTheme {
  if (!raw || typeof raw !== "object") {
    throw new TypeError("theme: expected object");
  }
  const t = raw as Record<string, unknown>;
  const colors = (t.colors ?? {}) as Record<string, unknown>;
  const cssvars = (t.cssvars ?? {}) as Record<string, unknown>;
  const terminal = (t.terminal ?? {}) as Record<string, unknown>;
  const globe = (t.globe ?? {}) as Record<string, unknown>;

  const need = (cond: boolean, msg: string) => {
    if (!cond) throw new TypeError(`theme: ${msg}`);
  };
  need(typeof t.id === "string", "id must be a string");
  need(typeof t.displayName === "string", "displayName must be a string");
  need(isLayoutPreset(t.layout), `layout must be one of ${[...LAYOUT_VALUES].join(", ")}`);
  need(
    typeof colors.r === "number" && typeof colors.g === "number" && typeof colors.b === "number",
    "colors.r/g/b must be numbers",
  );
  return {
    id: t.id as string,
    displayName: t.displayName as string,
    layout: t.layout as LayoutPreset,
    colors: {
      r: colors.r as number,
      g: colors.g as number,
      b: colors.b as number,
      black: String(colors.black ?? "#000000"),
      light_black: String(colors.light_black ?? "#05080d"),
      grey: String(colors.grey ?? "#262828"),
    },
    cssvars: {
      font_main: String(cssvars.font_main ?? "Fira Code, monospace"),
      font_main_light: String(cssvars.font_main_light ?? "Fira Code, monospace"),
    },
    terminal: {
      fontFamily: String(terminal.fontFamily ?? "Fira Code, monospace"),
      cursorStyle: (terminal.cursorStyle as TerminalCursorStyle) ?? "block",
      foreground: String(terminal.foreground ?? "#ffffff"),
      background: String(terminal.background ?? "#000000"),
      cursor: String(terminal.cursor ?? "#ffffff"),
      cursorAccent: String(terminal.cursorAccent ?? "#ffffff"),
      selection: String(terminal.selection ?? "rgba(255,255,255,0.3)"),
    },
    globe: {
      base: String(globe.base ?? "#000000"),
      marker: String(globe.marker ?? "#ffffff"),
      pin: String(globe.pin ?? "#ffffff"),
      satellite: String(globe.satellite ?? "#ffffff"),
    },
  };
}
