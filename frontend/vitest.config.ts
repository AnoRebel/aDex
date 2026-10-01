import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./tests/setup.ts'],
    coverage: {
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'tests/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/.nuxt/**',
        '**/.output/**'
      ]
    },
    globals: true,
    include: [
      'tests/**/*.{test,spec}.{js,mjs,cjs,ts,jsx,tsx}'
    ],
    exclude: [
      'node_modules/',
      '.nuxt/',
      'dist/',
      '.output/',
      // Playwright specs — they import @playwright/test and drive a real
      // browser, so Vitest can only fail to collect them. There is no
      // Playwright config in the repo yet; when one lands it runs these.
      'tests/e2e/**'
    ]
  },
  resolve: {
    alias: {
      // @xterm/addon-ligatures ships a broken `main`: it points at
      // lib/addon-ligatures.js, but the package only contains the .mjs build.
      // Vite's resolver falls back to it during a real build; Vitest's does
      // not, so anything importing useXTerm fails to load. Point at the file
      // that actually exists.
      '@xterm/addon-ligatures': resolve(
        __dirname,
        'node_modules/@xterm/addon-ligatures/lib/addon-ligatures.mjs',
      ),

      // Nuxt resolves `~` to `app/`, so every `~/<dir>` used in source needs
      // an entry here — the bare `~` fallback below points at the project
      // root and silently misses app/. A missing one fails the whole file at
      // import time ("Failed to resolve import"), not as a test failure.
      '~/components': resolve(__dirname, 'app/components'),
      '~/stores': resolve(__dirname, 'app/stores'),
      '~/composables': resolve(__dirname, 'app/composables'),
      '~/utils': resolve(__dirname, 'app/utils'),
      '~/types': resolve(__dirname, 'app/types'),
      '~/assets': resolve(__dirname, 'app/assets'),
      '~/lib': resolve(__dirname, 'app/lib'),
      '~~': resolve(__dirname, '.'),
      '@': resolve(__dirname, '.'),
      '~': resolve(__dirname, '.')
    }
  }
})
