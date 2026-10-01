#!/usr/bin/env bun
// scripts/copy-audio-cues.mjs
// Run via: bun scripts/copy-audio-cues.mjs
//
// Copies all 13 WAV cues from upstream eDEX-UI into:
//   frontend/app/assets/audio/edex/<cue>.wav
// and registers an `edex` soundpack inside the existing
//   frontend/app/assets/audio/soundpacks.json
// (preserving the other packs — default/minimal/retro/cyberpunk).
//
// Per the project decision: the 13 cues coexist with the existing 12-event
// vocabulary (`system_startup`, `terminal_bell`, ...). Legacy event names
// continue to work via the alias map registered by `useAdexAudio.ts`.

import { readdirSync, readFileSync, writeFileSync, copyFileSync, mkdirSync, existsSync } from "node:fs";
import { resolve, join, basename } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);
const SOURCE_DIR = resolve(ROOT, "..", "edex-ui", "src", "assets", "audio");
const OUT_DIR = join(ROOT, "frontend", "app", "assets", "audio", "edex");
const PACKS_FILE = join(ROOT, "frontend", "app", "assets", "audio", "soundpacks.json");

if (!existsSync(SOURCE_DIR)) {
  console.error(`source dir not found: ${SOURCE_DIR}`);
  process.exit(2);
}
mkdirSync(OUT_DIR, { recursive: true });

const cues = readdirSync(SOURCE_DIR).filter((f) => f.endsWith(".wav")).sort();

if (cues.length !== 13) {
  console.warn(`[warn] expected 13 WAV cues upstream, found ${cues.length}`);
}

// Cue metadata — categories and descriptions for the settings UI.
const CUE_META = {
  alarm:    { category: "alert",        description: "System threshold breach" },
  denied:   { category: "feedback",     description: "Action rejected" },
  error:    { category: "error",        description: "Error notification" },
  expand:   { category: "interaction",  description: "Panel/list expand" },
  folder:   { category: "interaction",  description: "Folder navigation" },
  granted:  { category: "feedback",     description: "Action accepted" },
  info:     { category: "notification", description: "Informational notice" },
  keyboard: { category: "interaction",  description: "On-screen keystroke" },
  panels:   { category: "interaction",  description: "Panel / modal / tab switch" },
  scan:     { category: "system",       description: "Scan or fuzzy-find start" },
  stdin:    { category: "terminal",     description: "Terminal stdin burst (rate-limited 4 Hz)" },
  stdout:   { category: "terminal",     description: "Terminal stdout burst (rate-limited 4 Hz)" },
  theme:    { category: "system",       description: "Theme applied" },
};

// Estimated playback duration per cue (ms). Used by the settings preview;
// ballpark only — Howler reads the real duration from the file at runtime.
const CUE_DURATION_MS = {
  alarm: 1200, denied: 200, error: 600, expand: 200, folder: 250, granted: 350,
  info: 500, keyboard: 80, panels: 250, scan: 1200, stdin: 60, stdout: 60, theme: 500,
};

const events = {};
for (const file of cues) {
  const id = basename(file, ".wav");
  copyFileSync(join(SOURCE_DIR, file), join(OUT_DIR, file));
  const meta = CUE_META[id] ?? { category: "system", description: id };
  events[id] = {
    id,
    name: id[0].toUpperCase() + id.slice(1),
    description: meta.description,
    category: meta.category,
    filePath: `/audio/edex/${file}`,
    duration: CUE_DURATION_MS[id] ?? 300,
    volume: 60,
    enabled: true,
  };
}

// Merge into the existing soundpacks.json without disturbing other packs.
let packs = {};
if (existsSync(PACKS_FILE)) {
  packs = JSON.parse(readFileSync(PACKS_FILE, "utf8"));
}
packs.edex = {
  id: "edex",
  name: "eDEX",
  displayName: "eDEX (original)",
  description: "13 original eDEX-UI cues — keyboard, panels, theme, stdin/stdout, granted, denied, error, info, alarm, expand, folder, scan",
  version: "1.0.0",
  author: "Gabriel Saillard / aDex-UI port",
  isBuiltIn: true,
  enabled: true,
  events,
};
writeFileSync(PACKS_FILE, JSON.stringify(packs, null, 2) + "\n");

console.log(`copied ${cues.length} WAV cues → frontend/app/assets/audio/edex/`);
console.log(`registered 'edex' pack in frontend/app/assets/audio/soundpacks.json`);
