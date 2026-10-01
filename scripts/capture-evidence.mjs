#!/usr/bin/env bun
// scripts/capture-evidence.mjs
//
// Capture an evidence screenshot of the running aDex-UI app. Wraps the
// host-OS screenshotters documented in `docs/evidence/_protocol.md`, so an
// implementer can run `bun scripts/capture-evidence.mjs <feature> <theme>`
// while `wails dev` is running and end up with a properly-named PNG +
// notes.md skeleton in `docs/evidence/<feature>/`.
//
// Usage:
//   bun scripts/capture-evidence.mjs <feature> [theme]
//
// Examples:
//   bun scripts/capture-evidence.mjs system-monitor tron
//   bun scripts/capture-evidence.mjs terminal blade

import { spawnSync, execSync } from "node:child_process";
import { existsSync, mkdirSync, writeFileSync, readFileSync } from "node:fs";
import { resolve, join } from "node:path";

const ROOT = resolve(new URL("..", import.meta.url).pathname);
const EVIDENCE = join(ROOT, "docs", "evidence");
const TEMPLATE = join(EVIDENCE, "_template", "notes.md");

const [, , feature, themeArg] = process.argv;
if (!feature) {
  console.error("usage: bun scripts/capture-evidence.mjs <feature> [theme]");
  process.exit(2);
}
const theme = themeArg ?? "tron";

const dir = join(EVIDENCE, feature);
mkdirSync(dir, { recursive: true });

const stamp = new Date().toISOString().replace(/[:.]/g, "-");
const file = join(dir, `${theme}-${stamp}.png`);

function captureLinux() {
  // Prefer the GNOME screenshotter when present; fall back to grim (wayland)
  // then ImageMagick's `import`. We deliberately avoid `gnome-screenshot -i`
  // (interactive) so the script can be scripted in CI.
  const tools = [
    { bin: "gnome-screenshot", args: ["-w", "-f", file] },
    { bin: "grim",            args: [file] },
    { bin: "import",          args: ["-window", "root", file] },
  ];
  for (const t of tools) {
    try {
      execSync(`command -v ${t.bin}`, { stdio: "ignore" });
      const r = spawnSync(t.bin, t.args, { stdio: "inherit" });
      if (r.status === 0) return t.bin;
    } catch {
      // tool not present; try next
    }
  }
  throw new Error("no screenshotter found (tried gnome-screenshot, grim, import)");
}

function captureMac() {
  const r = spawnSync("screencapture", ["-W", file], { stdio: "inherit" });
  if (r.status !== 0) throw new Error(`screencapture exited ${r.status}`);
  return "screencapture";
}

function captureWin() {
  // PowerShell snippet that BitBlts the active aDex-UI window to disk.
  const ps = `
Add-Type -AssemblyName System.Windows.Forms,System.Drawing
$proc = Get-Process aDex-UI -ErrorAction Stop |
        Where-Object MainWindowHandle -ne 0 |
        Select-Object -First 1
[System.Runtime.InteropServices.Marshal]::ZeroFreeGlobalAllocAnsi(0) | Out-Null
$rect = New-Object System.Drawing.Rectangle 0,0,1920,1080
$bmp  = New-Object System.Drawing.Bitmap $rect.Width,$rect.Height
$g    = [System.Drawing.Graphics]::FromImage($bmp)
$g.CopyFromScreen($rect.Location, [System.Drawing.Point]::Empty, $rect.Size)
$bmp.Save("${file.replace(/\\/g, "\\\\")}", [System.Drawing.Imaging.ImageFormat]::Png)
`.trim();
  const r = spawnSync("powershell.exe", ["-NoProfile", "-Command", ps], {
    stdio: "inherit",
  });
  if (r.status !== 0) throw new Error(`powershell exited ${r.status}`);
  return "powershell";
}

let tool;
try {
  tool =
    process.platform === "darwin"
      ? captureMac()
      : process.platform === "win32"
        ? captureWin()
        : captureLinux();
} catch (err) {
  console.error(`screenshot failed: ${err.message}`);
  console.error("Capture manually per docs/evidence/_protocol.md and rerun.");
  process.exit(1);
}

console.log(`captured: ${file} (via ${tool})`);

// Seed notes.md if absent
const notes = join(dir, "notes.md");
if (!existsSync(notes) && existsSync(TEMPLATE)) {
  let body = readFileSync(TEMPLATE, "utf8");
  let sha = "unknown";
  try {
    sha = execSync("git rev-parse --short HEAD", { encoding: "utf8" }).trim();
  } catch {
    // not a git repo / no head
  }
  body = body
    .replace("<feature>", feature)
    .replace("<linux|darwin|windows> <version>", `${process.platform}`)
    .replace("<theme-id>", theme)
    .replace("<ISO 8601>", new Date().toISOString())
    .replace("<git rev-parse HEAD>", sha)
    .replace("<name>", process.env.USER ?? process.env.USERNAME ?? "anon");
  writeFileSync(notes, body);
  console.log(`seeded: ${notes}`);
}
