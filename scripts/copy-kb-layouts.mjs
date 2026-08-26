#!/usr/bin/env bun
// scripts/copy-kb-layouts.mjs
// Run via: bun scripts/copy-kb-layouts.mjs
//
// Copies all 19 keyboard layouts from upstream eDEX-UI into:
//   frontend/app/assets/data/kb_layouts/<id>.json
// and writes:
//   frontend/app/assets/data/kb_layouts-index.json
//
// We preserve the upstream JSON shape verbatim (row_numbers / row_1 / row_2
// / row_3 / row_space with name/cmd/shift_cmd/ctrl_cmd/alt_cmd/fn_cmd
// fields) because that schema carries locale correctness — alt-graph
// chars, dead keys, per-locale F-key escapes — that a simpler shape would
// silently lose.
//
// The `~~~CTRLSEQ<n>~~~` placeholders in the source are NOT pre-substituted
// at copy time; they are substituted at load time in
// `composables/useKeyboard.ts`, mirroring the original
// keyboard.class.js behavior.

import { readdirSync, readFileSync, writeFileSync, mkdirSync, existsSync } from "node:fs";
import { resolve, join, basename } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);
const SOURCE_DIR = resolve(ROOT, "..", "edex-ui", "src", "assets", "kb_layouts");
const OUT_DIR = join(ROOT, "frontend", "app", "assets", "data", "kb_layouts");
const INDEX_FILE = join(ROOT, "frontend", "app", "assets", "data", "kb_layouts-index.json");

if (!existsSync(SOURCE_DIR)) {
  console.error(`source dir not found: ${SOURCE_DIR}`);
  process.exit(2);
}
mkdirSync(OUT_DIR, { recursive: true });

// id -> human-readable label. Pulled from common locale names so the
// settings UI doesn't show raw ids.
const DISPLAY_NAMES = {
  "da-DK": "Danish (Denmark)",
  "de-DE": "German (Germany)",
  "en-COLEMAK": "English (Colemak)",
  "en-DVORAK": "English (Dvorak)",
  "en-GB": "English (UK)",
  "en-NORMAN": "English (Norman)",
  "en-US": "English (US)",
  "en-WORKMAN": "English (Workman)",
  "es-ES": "Spanish (Spain)",
  "es-LAT": "Spanish (Latin America)",
  "fr-BEPO": "French (Bépo)",
  "fr-FR": "French (France)",
  "hu-HU": "Hungarian (Hungary)",
  "it-IT": "Italian (Italy)",
  "nl-BE": "Dutch (Belgium)",
  "pt-BR": "Portuguese (Brazil)",
  "sv-SE": "Swedish (Sweden)",
  "tr-TR-F": "Turkish (F)",
  "tr-TR-Q": "Turkish (Q)",
};

const REQUIRED_ROWS = ["row_numbers", "row_1", "row_2", "row_3", "row_space"];

const sourceFiles = readdirSync(SOURCE_DIR)
  .filter((f) => f.endsWith(".json"))
  .sort();

const indexEntries = [];
let warnings = 0;

for (const file of sourceFiles) {
  const id = basename(file, ".json");
  let raw = readFileSync(join(SOURCE_DIR, file), "utf8");
  let parsed;
  try {
    parsed = JSON.parse(raw);
  } catch (err) {
    // Upstream `en-WORKMAN.json` is missing its closing brace. Try a one-shot
    // fix-up: append `}` if the file otherwise looks complete (ends with `]`).
    const trimmed = raw.trimEnd();
    if (trimmed.endsWith("]")) {
      try {
        parsed = JSON.parse(trimmed + "\n}");
        console.warn(`[warn] ${id}: upstream missing closing brace — auto-repaired`);
      } catch {
        console.warn(`[warn] ${id}: invalid JSON — ${err.message}`);
        warnings++;
        continue;
      }
    } else {
      console.warn(`[warn] ${id}: invalid JSON — ${err.message}`);
      warnings++;
      continue;
    }
  }

  // Schema sanity check — every required row must exist and be a non-empty array.
  let ok = true;
  for (const row of REQUIRED_ROWS) {
    if (!Array.isArray(parsed[row]) || parsed[row].length === 0) {
      console.warn(`[warn] ${id}: missing or empty ${row}`);
      ok = false;
    }
  }
  if (!ok) {
    warnings++;
    continue;
  }

  writeFileSync(join(OUT_DIR, `${id}.json`), JSON.stringify(parsed, null, 2) + "\n");

  indexEntries.push({
    id,
    displayName: DISPLAY_NAMES[id] ?? id,
    file: `kb_layouts/${id}.json`,
  });
}

indexEntries.sort((a, b) => a.id.localeCompare(b.id));
writeFileSync(INDEX_FILE, JSON.stringify(indexEntries, null, 2) + "\n");

console.log(
  `copied ${indexEntries.length} keyboard layouts → frontend/app/assets/data/kb_layouts/ (${warnings} warnings)`,
);
console.log(`wrote index: frontend/app/assets/data/kb_layouts-index.json`);
