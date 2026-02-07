<template>
  <UApp>
    <div id="adex-app" :style="themeVars">
      <NuxtPage />
    </div>
  </UApp>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useThemeStore } from '~/stores/theme'

const themeStore = useThemeStore()

// Dynamically inject theme CSS custom properties from the theme store.
// These map to the --color_r / --color_g / --color_b variables consumed by
// main.css so the entire UI re-themes without a page reload.
const themeVars = computed(() => {
  const theme = themeStore.currentTheme
  if (!theme) return {}

  // Extract accent RGB components when the theme exposes full color objects
  const colors = theme.colors as Record<string, any> | undefined
  const accent =
    colors?.foreground?.primary ??
    colors?.accent?.primary ??
    colors?.primary ??
    null

  if (!accent) return {}

  // Parse hex or rgb string into r,g,b components
  const rgb = parseColor(accent)
  if (!rgb) return {}

  return {
    '--color_r': String(rgb.r),
    '--color_g': String(rgb.g),
    '--color_b': String(rgb.b),
  } as Record<string, string>
})

function parseColor(color: string): { r: number; g: number; b: number } | null {
  // Hex format: #rrggbb
  const hex = /^#?([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(color)
  if (hex && hex[1] && hex[2] && hex[3]) {
    return {
      r: parseInt(hex[1], 16),
      g: parseInt(hex[2], 16),
      b: parseInt(hex[3], 16),
    }
  }
  // rgb() format
  const rgbMatch = /rgb\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)/.exec(color)
  if (rgbMatch && rgbMatch[1] && rgbMatch[2] && rgbMatch[3]) {
    return {
      r: parseInt(rgbMatch[1]),
      g: parseInt(rgbMatch[2]),
      b: parseInt(rgbMatch[3]),
    }
  }
  return null
}

onMounted(async () => {
  try {
    await themeStore.initialize()
  } catch (e) {
    console.warn('Theme initialization warning:', e)
  }
})

useHead({
  title: 'aDex-UI',
  meta: [
    { name: 'description', content: 'A modern science fiction desktop environment' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
  ],
  link: [{ rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
})
</script>

<style>
/* The root container fills the viewport. All layout is handled by
   pages/index.vue and main.css -- no header, footer, or scrolling here. */
#adex-app {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}
</style>
