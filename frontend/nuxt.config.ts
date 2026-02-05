// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

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

  // CSS configuration for sci-fi theme
  css: ['~/assets/css/main.css'],

  // Build configuration for Wails integration
  build: {
    transpile: ['xterm-addon-webgl', 'xterm-addon-canvas']
  },

  // Vite configuration
  vite: {
    plugins: [
      tailwindcss(),
    ],
    define: {
      __WAILS_RUNTIME__: JSON.stringify(true)
    },
    optimizeDeps: {
      exclude: ['xterm', 'xterm-addon-fit', 'xterm-addon-webgl']
    }
  },

  // Dev server configuration
  devServer: {
    host: '127.0.0.1',
    port: 9245
  },

  nitro: {
    experimental: {
      wasm: true
    }
  },

  experimental: {
    typescriptPlugin: true
  }
})
