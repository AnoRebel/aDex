<template>
  <header class="app-header">
    <div class="header-content">
      <div class="logo-section">
        <h1 class="app-title">aDex</h1>
        <span class="version">v1.0.0</span>
      </div>
      
      <nav class="navigation">
        <NuxtLink to="/" class="nav-link">Home</NuxtLink>
        <NuxtLink to="/terminal" class="nav-link">Terminal</NuxtLink>
        <NuxtLink to="/filesystem" class="nav-link">Files</NuxtLink>
        <NuxtLink to="/system" class="nav-link">System</NuxtLink>
      </nav>
      
      <div class="header-actions">
        <button @click="toggleTheme" class="theme-toggle">
          {{ themeStore.isDark ? '🌙' : '☀️' }}
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useThemeStore } from '~/stores/theme'

const themeStore = useThemeStore()

const toggleTheme = () => {
  themeStore.toggleTheme()
}
</script>

<style scoped>
.app-header {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  padding: 0.5rem 1rem;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.app-title {
  font-size: 1.5rem;
  font-weight: bold;
  color: var(--text-primary);
  margin: 0;
}

.version {
  font-size: 0.875rem;
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
}

.navigation {
  display: flex;
  gap: 1rem;
}

.nav-link {
  color: var(--text-primary);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 0.25rem;
  transition: all 0.3s ease;
}

.nav-link:hover,
.nav-link.router-link-active {
  background: var(--bg-tertiary);
  color: var(--accent-primary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.theme-toggle {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 0.5rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.theme-toggle:hover {
  background: var(--bg-tertiary);
  border-color: var(--accent-primary);
}
</style>
