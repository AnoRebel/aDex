#!/usr/bin/env bun
// scripts/verify-evidence.mjs
// Run via: bun scripts/verify-evidence.mjs
//
// Every feature folder under docs/evidence/ must hold a notes.md plus at least
// one screenshot or log artifact. Folders starting with "_" (baseline,
// template, protocol) are reference material and are skipped.

import { readdirSync, existsSync, statSync } from "node:fs";
import { resolve, join } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);
const EVIDENCE = join(ROOT, "docs", "evidence");

if (!existsSync(EVIDENCE)) {
  console.error("No docs/evidence directory");
  process.exit(2);
}

const features = readdirSync(EVIDENCE).filter(
  (n) => !n.startsWith("_") && !n.startsWith(".") && statSync(join(EVIDENCE, n)).isDirectory(),
);

const errors = [];
for (const feature of features) {
  const entries = readdirSync(join(EVIDENCE, feature)).filter((n) => !n.startsWith("."));
  const hasNotes = entries.includes("notes.md");
  const hasArtifact = entries.some((n) => /\.(png|jpe?g|gif|webp|webm|mp4|svg|md|log)$/i.test(n) && n !== "notes.md");
  if (!hasNotes) errors.push(`docs/evidence/${feature}/notes.md missing`);
  if (!hasArtifact) errors.push(`docs/evidence/${feature}/ has no screenshot/log artifact`);
}

if (errors.length === 0) {
  console.log(`ok evidence present for ${features.length} feature(s)`);
  process.exit(0);
}

console.error("evidence verification failed:");
for (const e of errors) console.error(`  x ${e}`);
process.exit(1);
