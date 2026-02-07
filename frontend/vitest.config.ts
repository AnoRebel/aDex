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
      '.output/'
    ]
  },
  resolve: {
    alias: {
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
