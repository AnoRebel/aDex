#!/usr/bin/env bun
// scripts/verify-bindings.mjs
// Run via: bun scripts/verify-bindings.mjs
//
// Lint the frontend's Wails binding usage. For every import from a file under
// `frontend/app/lib/wailsjs/...`, verify each named import is actually
// exported by the resolved file. Catches the "rename a Go method, forget the
// frontend import" class of regression that the prior milestone leaked.

import { readFileSync, existsSync } from "node:fs";
import { readdir } from "node:fs/promises";
import { join, resolve, relative, dirname } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);
const FRONTEND = join(ROOT, "frontend", "app");
const WAILSJS_ROOT = join(FRONTEND, "lib", "wailsjs");

const errors = [];
const warnings = [];

function readExports(file) {
  if (!existsSync(file)) return null;
  const src = readFileSync(file, "utf8");
  const names = new Set();
  for (const m of src.matchAll(/export\s+(?:const|function|class|let|var)\s+([A-Za-z_][\w]*)/g)) {
    names.add(m[1]);
  }
  for (const m of src.matchAll(/export\s*\{([^}]+)\}/g)) {
    for (const part of m[1].split(",")) {
      const seg = part.trim();
      if (!seg) continue;
      const renamed = seg.match(/(\S+)\s+as\s+(\S+)/);
      names.add(renamed ? renamed[2] : seg);
    }
  }
  return names;
}

function resolveImport(specifier) {
  // Strip alias `~` → `frontend/app`
  let p;
  if (specifier.startsWith("~/")) {
    p = join(FRONTEND, specifier.slice(2));
  } else {
    return null;
  }
  for (const ext of [".ts", ".d.ts", ".js", ".tsx", "/index.ts", "/index.d.ts", "/index.js"]) {
    if (existsSync(p + ext)) return p + ext;
  }
  if (existsSync(p)) return p;
  return null;
}

async function* walk(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    if (entry.name === "node_modules" || entry.name === ".output" || entry.name === ".nuxt") continue;
    const p = join(dir, entry.name);
    if (entry.isDirectory()) {
      yield* walk(p);
    } else if (/\.(ts|tsx|vue|mjs|js)$/.test(entry.name)) {
      yield p;
    }
  }
}

const importRe =
  /import\s*(?:type\s*)?\{([^}]+)\}\s*from\s*['"]([^'"]+lib\/wailsjs[^'"]*)['"]/g;

let filesScanned = 0;
let importsScanned = 0;

const exportsCache = new Map();

for await (const file of walk(FRONTEND)) {
  if (file.startsWith(WAILSJS_ROOT)) continue;
  const src = readFileSync(file, "utf8");
  filesScanned += 1;
  for (const m of src.matchAll(importRe)) {
    importsScanned += 1;
    const specifier = m[2];
    const resolved = resolveImport(specifier);
    if (!resolved) {
      errors.push(`${relative(ROOT, file)}: cannot resolve '${specifier}'`);
      continue;
    }
    let exports = exportsCache.get(resolved);
    if (!exports) {
      exports = readExports(resolved) ?? new Set();
      exportsCache.set(resolved, exports);
    }
    if (exports.size === 0) {
      warnings.push(`${relative(ROOT, file)}: '${specifier}' resolved to ${relative(ROOT, resolved)} which exports nothing`);
      continue;
    }
    const names = m[1]
      .split(",")
      .map((s) => s.trim())
      .map((s) => s.replace(/\s+as\s+\S+/, ""))
      .filter(Boolean);
    for (const name of names) {
      if (!exports.has(name)) {
        errors.push(
          `${relative(ROOT, file)} imports '${name}' from '${specifier}' — not exported by ${relative(ROOT, resolved)}`,
        );
      }
    }
  }
}

// Flag the wrong-global-path bug — Wails v2 binds packages by their Go
// package name, so `*coordinator.ServiceCoordinator` resolves at
// `window.go.coordinator.ServiceCoordinator`, NOT `window.go.main.*`.
// The shim may keep `main` as a fallback for legacy test fixtures, but the
// canonical path must be present.
const SHIM = join(WAILSJS_ROOT, "coordinator.ts");
if (existsSync(SHIM)) {
  const src = readFileSync(SHIM, "utf8");
  const hasCanonical = /window\.go\??\.?\.?coordinator\??\.?\.?ServiceCoordinator/.test(src);
  const hasLegacyOnly = /window\.go\??\.?\.?main\??\.?\.?ServiceCoordinator/.test(src) && !hasCanonical;
  if (hasLegacyOnly) {
    errors.push(
      `${relative(ROOT, SHIM)} resolves bindings via 'window.go.main.ServiceCoordinator' only. Wails v2 binds packages by their Go package name; the coordinator's package is 'coordinator', so the canonical path 'window.go.coordinator.ServiceCoordinator' must be the primary resolver.`,
    );
  }
}

const summary = `Scanned ${filesScanned} files, ${importsScanned} binding imports across ${exportsCache.size} resolved targets`;

if (warnings.length > 0) {
  console.log("warnings:");
  for (const w of warnings) console.log(`  ! ${w}`);
}
if (errors.length > 0) {
  console.error("errors:");
  for (const e of errors) console.error(`  x ${e}`);
  console.error(summary);
  process.exit(1);
}
console.log(`ok ${summary}`);
