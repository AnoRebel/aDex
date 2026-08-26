// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";
import fs from "node:fs";
import path from "node:path";

// Read the canonical app version from wails.json's info.productVersion
// so the frontend version-checker can compare against GitHub Releases
// without us hand-syncing a string in two places. Falls back to '0.0.0'
// if the file is missing or malformed — never crash the build over a
// non-critical injection.
const wailsConfigPath = path.resolve(__dirname, "..", "wails.json");
let APP_VERSION = "0.0.0";
try {
  const raw = fs.readFileSync(wailsConfigPath, "utf-8");
  const parsed = JSON.parse(raw) as { info?: { productVersion?: string } };
  if (parsed?.info?.productVersion) APP_VERSION = parsed.info.productVersion;
} catch {
  // Build still proceeds with the 0.0.0 fallback.
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
