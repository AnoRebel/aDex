// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";
import fs from "node:fs";
import path from "node:path";
import { parse as parseYaml } from "yaml";

// Read the canonical app version from build/config.yml's `info.version`.
// That file is the single source of truth: the Wails build tooling stamps the
// binary from it, and reading it here keeps the frontend's update checker
// comparing against the same string without a hand-synced duplicate.
//
// This runs at Nuxt config-evaluation time — plain Node, outside Vite's module
// graph — so no Vite YAML plugin is involved (Vite has no native YAML support;
// it handles JSON, CSS, WASM and workers, but YAML needs a plugin). The `yaml`
// package is already present in the dependency tree.
//
// A missing or malformed config is a hard failure rather than a silent
// "0.0.0" fallback: shipping a mis-versioned binary breaks update checks in a
// way that is invisible until users are already on the wrong version.
const wailsConfigPath = path.resolve(__dirname, "..", "build", "config.yml");
let APP_VERSION: string;
{
  let raw: string;
  try {
    raw = fs.readFileSync(wailsConfigPath, "utf-8");
  } catch (cause) {
    throw new Error(
      `Unable to read ${wailsConfigPath}. It is the source of truth for the app version.`,
      { cause },
    );
  }

  const parsed = parseYaml(raw) as { info?: { version?: unknown } } | null;
  const version = parsed?.info?.version;
  if (typeof version !== "string" || version.trim() === "") {
    throw new Error(
      `${wailsConfigPath} has no usable \`info.version\` (found: ${JSON.stringify(version)}).`,
    );
  }
  APP_VERSION = version;
}

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  ssr: false,

  // App configuration
  app: {
    baseURL: "/",
    buildAssetsDir: "/_nuxt/",
    head: {
      title: 'aDex-UI',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'A modern science fiction desktop environment terminal application' }
      ]
    }
  },

  modules: [
    '@nuxt/eslint',
    '@nuxt/image',
    '@nuxt/scripts',
    '@nuxt/ui',
    '@nuxt/test-utils/module',
    '@vueuse/nuxt',
    '@vueuse/motion/nuxt',
    '@pinia/nuxt',
    '@nuxt/icon'
  ],

  // Runtime config for Wails backend communication
  runtimeConfig: {
    // Private keys (only available on server-side)
    wailsBackendUrl: process.env.WAILS_BACKEND_URL || 'ws://localhost:34115',

    // Public keys (exposed to client-side)
    public: {
      appName: 'aDex-UI',
      appVersion: '1.0.0'
    }
  },

  // CSS configuration for sci-fi theme. xterm.css MUST be loaded — without
  // it the terminal viewport renders with broken sizing and the cell text
  // is invisible (which is exactly the "blank terminal" symptom we hit on
  // 2026-05-06). xterm 5 ships its CSS at xterm/css/xterm.css.
  css: ['~/assets/css/main.css', '@xterm/xterm/css/xterm.css'],

  // Build configuration for Wails integration
  build: {
    transpile: ['@xterm/addon-webgl'],
  },

  // Vite configuration
  vite: {
    plugins: [
      tailwindcss(),
    ],
    define: {
      __WAILS_RUNTIME__: JSON.stringify(true),
      // Surface the canonical app version to the JS bundle so
      // useUpdateChecker can compare against GitHub Releases without
      // a runtime IPC round-trip. JSON.stringify is required by Vite's
      // define API (raw values are inserted verbatim into source, so
      // a string needs to be a string-literal string).
      __APP_VERSION__: JSON.stringify(APP_VERSION),
    },
    // Pre-bundle xterm so Vite's ESM resolver doesn't fall through to the
    // legacy CJS path. We migrated from `xterm@5.3` (no `module`/`exports`
    // map, ESM resolution silently fails) to the scoped `@xterm/*` packages
    // shipped with proper `exports` since 2024.
    optimizeDeps: {
      include: [
        '@xterm/xterm',
        '@xterm/addon-fit',
        '@xterm/addon-webgl',
        // Note: @xterm/addon-canvas is intentionally NOT in this list.
        // The package only ships CJS; Vite's named-import-from-CJS interop
        // throws "Importing binding name 'CanvasAddon' is not found" at
        // runtime. We rely on xterm's built-in DOM renderer if WebGL ever
        // fails. The package is kept in package.json so xterm 6's optional
        // dep resolution stays happy, but no source file imports it.
      ],
    },
  },

  // Dev server configuration
  devServer: {
    host: '127.0.0.1',
    port: 9245
  },

  nitro: {
    // Emit the static bundle to `frontend/dist` — the location Wails v3's
    // build Taskfile, dev config, and `//go:embed` directive all assume by
    // default. Nuxt's own default is `.output/public`; pointing Nuxt at the
    // v3 convention is one line, whereas overriding the v3 tooling to look
    // at `.output/public` would mean re-applying that override every time
    // the generated build assets are refreshed during the beta period.
    output: {
      publicDir: 'dist'
    },
    experimental: {
      wasm: true
    }
  },

  experimental: {
    typescriptPlugin: true
  }
})
