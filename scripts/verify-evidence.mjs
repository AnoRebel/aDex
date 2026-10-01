#!/usr/bin/env bun
// scripts/verify-evidence.mjs
// Run via: bun scripts/verify-evidence.mjs
//
// Per `specs/feature-evidence-protocol/spec.md`, every feature claimed
// completed in tasks.md must have at least one screenshot + notes.md under
// docs/evidence/<feature>/. This script walks the active openspec change's
// tasks.md, identifies feature names referenced by checked-off tasks, and
// fails if their evidence folder is missing or empty.
//
// Heuristic: features are derived from `docs/evidence/<feature>/` paths
// mentioned in checked-off tasks.

import { readFileSync, readdirSync, existsSync, statSync } from "node:fs";
import { resolve, join } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);
const CHANGE_DIR = join(ROOT, "openspec", "changes", "edex-parity-and-uplift");
const TASKS = join(CHANGE_DIR, "tasks.md");
const EVIDENCE = join(ROOT, "docs", "evidence");

if (!existsSync(TASKS)) {
  console.error(`No tasks file at ${TASKS}`);
  process.exit(2);
}

const src = readFileSync(TASKS, "utf8");
const checkedRe = /^- \[x\][^\n]*?docs\/evidence\/([\w._-]+)\//gim;

const required = new Set();
for (const m of src.matchAll(checkedRe)) {
  required.add(m[1]);
}

const errors = [];
for (const feature of required) {
  if (feature.startsWith("_")) continue; // baseline / template / protocol
  const dir = join(EVIDENCE, feature);
  if (!existsSync(dir)) {
    errors.push(`missing dir: docs/evidence/${feature}/`);
    continue;
  }
  const entries = readdirSync(dir).filter((n) => !n.startsWith("."));
  const hasNotes = entries.includes("notes.md");
  const hasArtifact = entries.some((n) => /\.(png|jpe?g|gif|webp|webm|mp4|svg|md|log)$/i.test(n) && n !== "notes.md");
  if (!hasNotes) errors.push(`docs/evidence/${feature}/notes.md missing`);
  if (!hasArtifact) errors.push(`docs/evidence/${feature}/ has no screenshot/log artifact`);
}

if (errors.length === 0) {
  console.log(`ok evidence present for ${required.size} feature(s)`);
  process.exit(0);
}

console.error("evidence verification failed:");
for (const e of errors) console.error(`  x ${e}`);
process.exit(1);
