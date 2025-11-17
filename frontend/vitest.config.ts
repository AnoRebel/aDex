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
    globals: {
      // Mock browser APIs that are not available in jsdom
      document: true,
      window: true,
      navigator: true,
      console: true,
      CustomEvent: true,
      Event: true,
      HTMLElement: true,
      MouseEvent: true,
      DragEvent: true
    },
    include: [
      'tests/**/*.{test,spec}.{js,mjs,cjs,ts,jsx,tsx}'
    ],
    exclude: [
      'node_modules/',
      '.nuxt/',
      'dist/',
      '.output/'
    ]
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, '.'),
      '~': resolve(__dirname, '.')
    }
  }
})
