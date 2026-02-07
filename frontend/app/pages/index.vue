<template>
  <!-- Boot / splash overlay -->
  <EdexBootScreen
    v-if="booting"
    @complete="onBootComplete"
  />

  <!-- Main application (hidden while booting) -->
  <div v-show="!booting" class="edex-app">

    <!-- Grid background pattern -->
    <div class="edex-background" :class="{ solidBackground: solidBg }" />

    <!-- TOP NAVIGATION BAR -->
    <nav class="edex-topbar" role="navigation" aria-label="Section navigation">
      <div
        class="topbar-section"
        :class="{ active: topSection === 'panel' }"
        @click="setTopSection('panel')"
      >PANEL</div>
      <div
        class="topbar-section"
        :class="{ active: topSection === 'system' }"
        @click="setTopSection('system')"
      >SYSTEM</div>
      <div
        class="topbar-section"
        :class="{ active: topSection === 'terminal' }"
        @click="setTopSection('terminal')"
      >TERMINAL</div>

      <div class="topbar-spacer" />

      <!-- Shell tab labels in top bar -->
      <div
        v-for="(tab, idx) in shellTabs"
        :key="tab.id"
        class="topbar-section"
        :class="{ active: activeTabIndex === idx }"
        @click="switchTab(idx)"
      >{{ tab.title || ('MAIN SHELL ' + (idx + 1)) }}</div>

      <!-- Empty tab placeholders -->
      <div
        v-for="n in (MAX_TABS - shellTabs.length)"
        :key="'empty-' + n"
        class="topbar-section"
        style="opacity: 0.3;"
      >EMPTY</div>

      <div class="topbar-spacer" />

      <div
        class="topbar-section"
        :class="{ active: topSection === 'rightPanel' }"
        @click="setTopSection('rightPanel')"
      >PANEL</div>
      <div
        class="topbar-section"
        :class="{ active: topSection === 'network' }"
        @click="setTopSection('network')"
      >NETWORK</div>
    </nav>

    <!-- MAIN THREE-COLUMN AREA -->
    <div class="edex-main">

      <!-- LEFT COLUMN -->
      <aside class="mod-column left">
        <EdexClock />
        <EdexSysinfo />
        <EdexHardware />
        <EdexCpuInfo />
        <EdexRamWatcher />
        <EdexToplist />
      </aside>

      <!-- CENTER: MAIN SHELL -->
      <section class="edex-center">
        <div class="main-shell">
          <!-- Tab strip inside the shell -->
          <div class="shell-tabs">
            <div
              v-for="(tab, idx) in shellTabs"
              :key="tab.id"
              class="shell-tab"
              :class="{ active: activeTabIndex === idx }"
              @click="switchTab(idx)"
            >
              <span>{{ tab.title || ('SHELL ' + (idx + 1)) }}</span>
            </div>
            <!-- Empty tab slots -->
            <div
              v-for="n in (MAX_TABS - shellTabs.length)"
              :key="'empty-tab-' + n"
              class="shell-tab"
              style="opacity: 0.3;"
              @click="addTab"
            >
              <span>EMPTY</span>
            </div>
          </div>

          <!-- Terminal emulator surface (one per tab, only active is visible) -->
          <div class="shell-terminal">
            <EdexTerminal
              v-for="(tab, idx) in shellTabs"
              v-show="activeTabIndex === idx"
              :key="tab.id"
              :ref="(el: any) => terminalRefs[tab.id] = el"
              :session-id="tab.sessionId"
              :active="activeTabIndex === idx"
            />
          </div>

          <!-- Status bar -->
          <div class="shell-status">
            <span>{{ shellStatusLeft }}</span>
            <span>{{ shellStatusRight }}</span>
          </div>
        </div>
      </section>

      <!-- RIGHT COLUMN -->
      <aside class="mod-column right">
        <EdexNetstat />
        <EdexGlobe />
        <EdexTraffic />
      </aside>
    </div>

    <!-- BOTTOM SECTION: FILESYSTEM + KEYBOARD -->
    <div class="edex-bottom">
      <EdexFilesystem @navigate="onFsNavigate" @open="onFsOpen" />
      <EdexKeyboard @key="onVirtualKey" />
    </div>

    <!-- Settings modal (toggled via keyboard shortcut or menu) -->
    <EdexSettingsModal
      v-model="showSettings"
      @settings-changed="onSettingsChanged"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useSystemStore } from '~/stores/system'
import { useTerminalStore } from '~/stores/terminal'
import { useFilesystemStore } from '~/stores/filesystem'
import { useNetworkStore } from '~/stores/network'
import { useAudioStore } from '~/stores/audio'
import { useThemeStore } from '~/stores/theme'
import { useAppStore } from '~/stores/app'

// Stores
const systemStore = useSystemStore()
const terminalStore = useTerminalStore()
const filesystemStore = useFilesystemStore()
const networkStore = useNetworkStore()
const audioStore = useAudioStore()
const themeStore = useThemeStore()
const appStore = useAppStore()

// Constants
const MAX_TABS = 5

// Template refs for terminal instances
const terminalRefs = ref<Record<string, any>>({})

// Boot state
const booting = ref(true)
const solidBg = ref(false)
const showSettings = ref(false)

function onBootComplete() {
  booting.value = false
  playSound('system_startup')
}

// Top bar active section (cosmetic)
const topSection = ref('terminal')
function setTopSection(section: string) {
  topSection.value = section
}

// Shell tabs
interface ShellTab {
  id: string
  sessionId: string
  title: string
}

const shellTabs = ref<ShellTab[]>([])
const activeTabIndex = ref(0)

async function addTab() {
  if (shellTabs.value.length >= MAX_TABS) return
  const num = shellTabs.value.length + 1
  const sessionId = await terminalStore.createSession(`Shell ${num}`)
  if (sessionId) {
    shellTabs.value.push({
      id: `tab-${Date.now()}`,
      sessionId,
      title: num === 1 ? 'MAIN SHELL' : `SHELL ${num}`,
    })
    activeTabIndex.value = shellTabs.value.length - 1
    terminalStore.setActiveSession(sessionId)
  }
}

function switchTab(idx: number) {
  if (idx < 0 || idx >= shellTabs.value.length) return
  activeTabIndex.value = idx
  terminalStore.setActiveSession(shellTabs.value[idx].sessionId)
  // Focus the terminal after switching
  nextTick(() => {
    const tab = shellTabs.value[idx]
    if (tab && terminalRefs.value[tab.id]) {
      terminalRefs.value[tab.id].focus?.()
    }
  })
}

async function closeCurrentTab() {
  if (shellTabs.value.length <= 1) return
  const tab = shellTabs.value[activeTabIndex.value]
  if (!tab) return
  await terminalStore.closeSession(tab.sessionId)
  delete terminalRefs.value[tab.id]
  shellTabs.value.splice(activeTabIndex.value, 1)
  if (activeTabIndex.value >= shellTabs.value.length) {
    activeTabIndex.value = Math.max(0, shellTabs.value.length - 1)
  }
  if (shellTabs.value.length > 0) {
    terminalStore.setActiveSession(shellTabs.value[activeTabIndex.value].sessionId)
  }
}

// Shell status bar
const shellStatusLeft = computed(() => {
  const tab = shellTabs.value[activeTabIndex.value]
  if (!tab) return ''
  const session = terminalStore.activeSession
  return session
    ? `${session.shell ?? '/bin/bash'}  --  ${session.workingDirectory ?? '~'}`
    : 'No active session'
})

const shellStatusRight = computed(() => {
  const d = new Date()
  const time = [d.getHours(), d.getMinutes(), d.getSeconds()]
    .map((v) => String(v).padStart(2, '0'))
    .join(':')
  const session = terminalStore.activeSession
  const dims = session ? `${session.columns ?? 80}x${session.rows ?? 24}` : ''
  return `${dims}  ${time}`
})

// Filesystem events
function onFsNavigate(path: string) {
  filesystemStore.fetchDirectory(path)
}

function onFsOpen(item: { name: string; path: string; type?: string }) {
  if (item.type === 'directory') {
    filesystemStore.fetchDirectory(item.path)
    const session = terminalStore.activeSession
    if (session) {
      terminalStore.sendInput(`cd "${item.path}"\n`, session.id)
    }
  }
}

// Virtual keyboard events
function onVirtualKey(key: string) {
  const session = terminalStore.activeSession
  if (session) {
    terminalStore.sendInput(key, session.id)
  }
}

// Settings
function onSettingsChanged(settings: any) {
  // Apply settings changes (theme, audio, etc.)
  if (settings.theme) {
    themeStore.setCurrentTheme?.(settings.theme)
  }
}

// Audio helper
function playSound(_eventId: string) {
  try {
    const store = audioStore as any
    if (store.settings?.enabled || store.isEnabled) {
      store.playEvent?.(_eventId) ?? store.playSound?.(_eventId)
    }
  } catch {
    // Non-critical
  }
}

// Keyboard shortcuts
function handleKeyDown(e: KeyboardEvent) {
  // Ctrl+Shift+1..5 => switch shell tab
  if (e.ctrlKey && e.shiftKey && !e.altKey) {
    const num = parseInt(e.key, 10)
    if (num >= 1 && num <= MAX_TABS && num <= shellTabs.value.length) {
      e.preventDefault()
      switchTab(num - 1)
      return
    }
  }
  // Ctrl+Shift+T => new tab
  if (e.ctrlKey && e.shiftKey && (e.key === 't' || e.key === 'T')) {
    e.preventDefault()
    addTab()
    return
  }
  // Ctrl+Shift+W => close current tab
  if (e.ctrlKey && e.shiftKey && (e.key === 'w' || e.key === 'W') && shellTabs.value.length > 1) {
    e.preventDefault()
    closeCurrentTab()
    return
  }
  // Ctrl+, => settings
  if (e.ctrlKey && e.key === ',') {
    e.preventDefault()
    showSettings.value = !showSettings.value
    return
  }
}

// Lifecycle
onMounted(async () => {
  // Initialise all stores in parallel (errors are non-fatal)
  const safeInit = (fn: () => Promise<any>) =>
    fn().catch((e: unknown) => console.warn('Store init:', e))

  await Promise.all([
    safeInit(() => appStore.initialize()),
    safeInit(() => systemStore.initialize()),
    safeInit(() => themeStore.initialize()),
    safeInit(() => terminalStore.initialize()),
    safeInit(async () => {
      // Filesystem store uses options API - call fetchDirectory instead of initialize
      try {
        if (typeof filesystemStore.initialize === 'function') {
          await filesystemStore.initialize()
        } else {
          await filesystemStore.fetchDirectory('/')
        }
      } catch {
        // non-fatal
      }
    }),
    safeInit(async () => {
      try {
        if (typeof networkStore.startMonitoring === 'function') {
          await networkStore.startMonitoring()
        } else if (typeof (networkStore as any).initialize === 'function') {
          await (networkStore as any).initialize()
        }
      } catch {
        // Network monitoring is optional
      }
    }),
  ])

  // Audio initialisation (non-blocking, defensive)
  try {
    const store = audioStore as any
    if (typeof store.initialize === 'function') {
      store.initialize()
    }
  } catch {
    // non-critical
  }

  // Seed the first shell tab from whatever session the terminal store created
  await nextTick()
  const sessions = terminalStore.allSessions
  if (sessions.length > 0) {
    shellTabs.value = sessions.map((s, idx) => ({
      id: `tab-${s.id}`,
      sessionId: s.id,
      title: idx === 0 ? 'MAIN SHELL' : (s.title ?? `SHELL ${idx + 1}`),
    }))
    activeTabIndex.value = 0
    terminalStore.setActiveSession(sessions[0].id)
  } else {
    await addTab()
  }

  // Register keyboard shortcuts
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
  systemStore.stopMonitoring()
})
</script>

<style scoped>
.edex-main {
  min-height: 0;
}

.mod-column {
  overflow-y: auto;
  overflow-x: hidden;
}

.edex-app {
  animation: fadeIn 0.4s cubic-bezier(0.19, 1, 0.22, 1) forwards;
}
</style>
