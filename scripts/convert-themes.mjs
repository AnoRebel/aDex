#!/usr/bin/env bun
// scripts/convert-themes.mjs
// Run via: bun scripts/convert-themes.mjs
//
// Reads every theme JSON from the upstream eDEX-UI source tree
// (`/home/ano/Code/Dex-UI/edex-ui/src/assets/themes/`) and writes a
// port-compatible theme JSON for each one to:
//
//   frontend/app/assets/data/themes/<id>.json
//
// Plus an index manifest at:
//
//   frontend/app/assets/data/themes-index.json
//
// Layout classification: themes whose original had `injectCSS` matching the
// known eDEX-UI layout variants are mapped to a typed `layout` preset
// (`disrupted | typeleft | fulltype | notype | colorfilter`). Everything
// else gets `default`. We do NOT carry the injectCSS string into the new
// themes — it is replaced by stylesheets under `assets/css/layouts/`.

import { readdirSync, readFileSync, writeFileSync, mkdirSync, existsSync } from "node:fs";
import { resolve, join, basename } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);

/**
 * Terminal selection colour, always translucent.
 *
 * A few upstream themes (cyborg, cyborg-focus, matrix) give `selection` an
 * opaque hex value. xterm paints the selection OVER the glyphs, so an opaque
 * colour hides the selected text entirely — you cannot read what you have
 * just selected. Every other theme uses rgba at 0.3, so convert hex to the
 * same alpha rather than passing it through.
 */
function normaliseSelection(selection, colors) {
  const fallback = `rgba(${colors.r}, ${colors.g}, ${colors.b}, 0.3)`;
  if (!selection) return fallback;

  const hex = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(selection.trim());
  if (!hex) return selection; // already rgba()/rgb(), leave it alone

  const [r, g, b] = hex.slice(1).map((h) => parseInt(h, 16));
  return `rgba(${r}, ${g}, ${b}, 0.3)`;
}

const SOURCE_DIR = resolve(ROOT, "..", "edex-ui", "src", "assets", "themes");
const OUT_DIR = join(ROOT, "frontend", "app", "assets", "data", "themes");
const INDEX_FILE = join(ROOT, "frontend", "app", "assets", "data", "themes-index.json");
const LEGACY_BUNDLE = join(ROOT, "frontend", "app", "assets", "data", "themes.json");

if (!existsSync(SOURCE_DIR)) {
  console.error(`source dir not found: ${SOURCE_DIR}`);
  console.error("Adjust the SOURCE_DIR constant if the upstream repo lives elsewhere.");
  process.exit(2);
}
mkdirSync(OUT_DIR, { recursive: true });

// Map an original theme id (basename without .json) to a layout preset.
// Drives both the field on the resulting theme and the `data-layout`
// attribute the shell sets at runtime.
function classifyLayout(id, originalInjectCSS) {
  // 1) Filename-based pattern (the canonical signal).
  if (id.endsWith("-disrupted") || /-disrupted/.test(id)) return "disrupted";
  if (id.endsWith("-typeleft")) return "typeleft";
  if (id.endsWith("-fulltype")) return "fulltype";
  if (id.endsWith("-notype")) return "notype";
  if (id.endsWith("-colorfilter")) return "colorfilter";

  // 2) Heuristic on the injectCSS payload (covers any community theme that
  //    doesn't follow the naming convention).
  if (originalInjectCSS) {
    const css = String(originalInjectCSS);
    if (/section#main_shell\s*\{[^}]*left:/.test(css)) return "disrupted";
    if (/section#filesystem\s*\{[^}]*left:\s*55vw/.test(css)) return "typeleft";
    if (/section#keyboard\s*\{[^}]*width:\s*100vw/.test(css)) return "fulltype";
    if (/section#filesystem\s*\{[^}]*width:\s*100vw/.test(css) &&
        /section#keyboard\s*\{\s*display:\s*none/.test(css)) return "notype";
  }
  return "default";
}

// Pretty display name from id: "tron-disrupted" -> "Tron Disrupted"
function displayName(id) {
  return id
    .split("-")
    .map((p) => p[0].toUpperCase() + p.slice(1))
    .join(" ");
}

const sourceFiles = readdirSync(SOURCE_DIR).filter((f) => f.endsWith(".json"));
const indexEntries = [];
let warnings = 0;

for (const file of sourceFiles) {
  const id = basename(file, ".json");
  const src = JSON.parse(readFileSync(join(SOURCE_DIR, file), "utf8"));

  // Sanity-check required fields. If a theme file is malformed upstream we
  // log + skip rather than write garbage.
  if (!src.colors || typeof src.colors.r !== "number") {
    console.warn(`[warn] ${id}: missing colors.r — skipping`);
    warnings++;
    continue;
  }

  const layout = classifyLayout(id, src.injectCSS);
  const out = {
    id,
    displayName: displayName(id),
    layout,
    colors: {
      r: src.colors.r,
      g: src.colors.g,
      b: src.colors.b,
      black: src.colors.black ?? "#000000",
      light_black: src.colors.light_black ?? "#05080d",
      grey: src.colors.grey ?? "#262828",
    },
    cssvars: {
      font_main: src.cssvars?.font_main ?? "United Sans Medium, Fira Code, monospace",
      font_main_light: src.cssvars?.font_main_light ?? "United Sans Light, Fira Code, monospace",
    },
    terminal: {
      fontFamily: src.terminal?.fontFamily ?? "Fira Mono, Fira Code, monospace",
      cursorStyle: src.terminal?.cursorStyle ?? "block",
      foreground: src.terminal?.foreground ?? `rgb(${src.colors.r}, ${src.colors.g}, ${src.colors.b})`,
      background: src.terminal?.background ?? src.colors.light_black ?? "#05080d",
      cursor: src.terminal?.cursor ?? `rgb(${src.colors.r}, ${src.colors.g}, ${src.colors.b})`,
      cursorAccent: src.terminal?.cursorAccent ?? src.terminal?.cursor ?? `rgb(${src.colors.r}, ${src.colors.g}, ${src.colors.b})`,
      selection: normaliseSelection(src.terminal?.selection, src.colors),
    },
    globe: {
      base: src.globe?.base ?? src.colors.black ?? "#000000",
      marker: src.globe?.marker ?? `rgb(${src.colors.r}, ${src.colors.g}, ${src.colors.b})`,
      pin: src.globe?.pin ?? `rgb(${src.colors.r}, ${src.colors.g}, ${src.colors.b})`,
      satellite: src.globe?.satellite ?? `rgb(${src.colors.r}, ${src.colors.g}, ${src.colors.b})`,
    },
  };

  writeFileSync(join(OUT_DIR, `${id}.json`), JSON.stringify(out, null, 2) + "\n");

  indexEntries.push({
    id,
    displayName: out.displayName,
    layout,
    primary: `rgb(${out.colors.r}, ${out.colors.g}, ${out.colors.b})`,
    file: `themes/${id}.json`,
  });
}

// Sort the index for stable output.
indexEntries.sort((a, b) => a.id.localeCompare(b.id));
writeFileSync(INDEX_FILE, JSON.stringify(indexEntries, null, 2) + "\n");

// Also rewrite the legacy `themes.json` bundle so existing components that
// import it directly (ThemeManager.vue, theme store loadAdexThemes()) see
// all 21 themes without code changes. We strip injectCSS to honor the V2
// "no injection" decision (D2) and add the `layout` field so legacy code
// can opt into reading it.
const bundle = {
  themes: indexEntries.map((entry) => {
    const t = JSON.parse(readFileSync(join(OUT_DIR, `${entry.id}.json`), "utf8"));
    return {
      id: t.id,
      name: t.displayName,
      layout: t.layout,
      colors: t.colors,
      cssvars: t.cssvars,
      terminal: t.terminal,
      globe: t.globe,
      injectCSS: "",
    };
  }),
};
writeFileSync(LEGACY_BUNDLE, JSON.stringify(bundle, null, 2) + "\n");

console.log(
  `converted ${indexEntries.length} themes → frontend/app/assets/data/themes/ (${warnings} warnings)`,
);
console.log(`wrote index:  frontend/app/assets/data/themes-index.json`);
console.log(`wrote bundle: frontend/app/assets/data/themes.json (legacy)`);

// Summary by layout preset for quick verification.
const byLayout = indexEntries.reduce((acc, t) => {
  acc[t.layout] = (acc[t.layout] ?? 0) + 1;
  return acc;
}, {});
console.log(`layout breakdown:`, byLayout);
