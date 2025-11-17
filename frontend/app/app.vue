<template>
  <UApp>
    <div id="app" :class="themeClasses">
      <div class="app-container">
        <AppHeader />
        <main class="main-content">
          <NuxtPage />
        </main>
        <AppFooter />
      </div>

      <!-- Global modals and overlays -->
      <Teleport to="body">
        <AppModal v-if="modalStore.isOpen" />
        <AppNotification />
      </Teleport>
    </div>
  </UApp>
</template>

<script setup lang="ts">
import "@wailsio/runtime";
// App-level setup
import { computed } from 'vue'
import { useThemeStore } from '~/stores/theme'
import { useModalStore } from '~/stores/modal'

// Initialize stores
const themeStore = useThemeStore()
const modalStore = useModalStore()

// Compute theme classes for the app
const themeClasses = computed(() => ({
  'theme-cyberpunk': themeStore.currentTheme === 'cyberpunk',
  'theme-matrix': themeStore.currentTheme === 'matrix',
  'theme-neon': themeStore.currentTheme === 'neon',
  'theme-dark': themeStore.isDark,
  'theme-light': !themeStore.isDark,
}))

// Initialize theme on app start
onMounted(async () => {
  await themeStore.initializeTheme()
})

// SEO and meta
useHead({
  title: 'aDex-UI - Advanced Desktop Experience',
  meta: [
    { name: 'description', content: 'A futuristic desktop environment built with Nuxt v4 and Wails' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { name: 'theme-color', content: '#00ff00' }
  ],
  link: [
    { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
  ]
})
</script>

<style>
/* Global app styles */
#app {
  min-height: 100vh;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  padding: 1rem;
  overflow-y: auto;
}

/* Theme-specific styles */
.theme-cyberpunk {
  --bg-primary: #0a0a0a;
  --bg-secondary: #1a1a1a;
  --bg-tertiary: #2a2a2a;
  --text-primary: #00ff00;
  --text-secondary: #00cc00;
  --accent-primary: #ff00ff;
  --accent-secondary: #00ffff;
  --border-color: #00ff00;
  --shadow-color: rgba(0, 255, 0, 0.3);
}

.theme-matrix {
  --bg-primary: #000000;
  --bg-secondary: #0a0a0a;
  --bg-tertiary: #1a1a1a;
  --text-primary: #00ff00;
  --text-secondary: #00dd00;
  --accent-primary: #00ff00;
  --accent-secondary: #00aa00;
  --border-color: #00ff00;
  --shadow-color: rgba(0, 255, 0, 0.5);
}

.theme-neon {
  --bg-primary: #0d0221;
  --bg-secondary: #1a0b3d;
  --bg-tertiary: #2a1a5d;
  --text-primary: #ff00ff;
  --text-secondary: #cc00cc;
  --accent-primary: #00ffff;
  --accent-secondary: #ffff00;
  --border-color: #ff00ff;
  --shadow-color: rgba(255, 0, 255, 0.4);
}

.theme-dark {
  --bg-primary: #1a1a1a;
  --bg-secondary: #2a2a2a;
  --bg-tertiary: #3a3a3a;
  --text-primary: #ffffff;
  --text-secondary: #cccccc;
  --accent-primary: #00ff00;
  --accent-secondary: #0099ff;
  --border-color: #444444;
  --shadow-color: rgba(0, 0, 0, 0.3);
}

.theme-light {
  --bg-primary: #ffffff;
  --bg-secondary: #f5f5f5;
  --bg-tertiary: #e0e0e0;
  --text-primary: #000000;
  --text-secondary: #333333;
  --accent-primary: #0066cc;
  --accent-secondary: #00aa44;
  --border-color: #cccccc;
  --shadow-color: rgba(0, 0, 0, 0.1);
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: var(--bg-secondary);
}

::-webkit-scrollbar-thumb {
  background: var(--accent-primary);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--accent-secondary);
}

/* Global animations */
@keyframes glow {
  0% { box-shadow: 0 0 5px var(--shadow-color); }
  50% { box-shadow: 0 0 20px var(--shadow-color), 0 0 30px var(--shadow-color); }
  100% { box-shadow: 0 0 5px var(--shadow-color); }
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.7; }
  100% { opacity: 1; }
}

@keyframes slideIn {
  from { transform: translateX(-100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}

/* Utility classes */
.glow-effect {
  animation: glow 2s infinite;
}

.pulse-effect {
  animation: pulse 1.5s infinite;
}

.slide-in {
  animation: slideIn 0.3s ease-out;
}

.text-terminal {
  font-family: 'Courier New', monospace;
  letter-spacing: 0.05em;
}

.border-glow {
  border: 1px solid var(--border-color);
  box-shadow: 0 0 10px var(--shadow-color);
}
</style>
