<template>
  <!-- Splash Screen -->
  <SplashScreen 
    v-if="showSplash" 
    :min-duration="2500"
    @complete="onSplashComplete"
  />

  <div v-show="!showSplash" class="adex-desktop" :class="desktopClasses">
    <!-- Background Effects -->
    <div class="background-effects">
      <div v-if="themeSettings.particles" class="particle-container">
        <div
          v-for="i in 50"
          :key="i"
          class="particle"
          :style="getParticleStyle(i)"
        ></div>
      </div>
      <div v-if="themeSettings.scanlines" class="scanlines"></div>
      <div class="background-grid"></div>
    </div>

    <!-- Main Desktop Layout -->
    <div class="desktop-layout">
      <!-- Top Bar -->
      <header class="top-bar">
        <div class="top-bar-left">
          <div class="app-logo">
            <span class="logo-text glitch">aDex-UI</span>
            <span class="logo-version">v2.0</span>
          </div>
          <div class="system-info">
            <span class="system-time">{{ currentTime }}</span>
            <span class="system-uptime">Uptime: {{ formatUptime(systemData.uptime) }}</span>
          </div>
        </div>

        <div class="top-bar-center">
          <div class="window-controls">
            <button @click="minimizeApp" class="window-btn minimize">─</button>
            <button @click="toggleFullscreen" class="window-btn maximize">□</button>
            <button @click="closeApp" class="window-btn close">✕</button>
          </div>
        </div>

        <div class="top-bar-right">
          <div class="quick-actions">
            <button @click="showSettingsModal" class="action-btn" title="Settings">
              <span class="btn-icon">⚙️</span>
            </button>
            <button @click="showThemeManager" class="action-btn" title="Themes">
              <span class="btn-icon">🎨</span>
            </button>
            <button @click="toggleSoundPanel" class="action-btn" title="Audio">
              <span class="btn-icon">{{ soundEnabled ? '🔊' : '🔇' }}</span>
            </button>
            <button @click="toggleVirtualKeyboard" class="action-btn" title="Keyboard">
              <span class="btn-icon">⌨️</span>
            </button>
          </div>
          <div class="system-status">
            <div class="status-indicator cpu" :class="getCpuStatusClass()">
              <span class="status-label">CPU</span>
              <span class="status-value">{{ Math.round(systemData.cpu?.usage || 0) }}%</span>
            </div>
            <div class="status-indicator memory" :class="getMemoryStatusClass()">
              <span class="status-label">MEM</span>
              <span class="status-value">{{ Math.round(systemData.memory?.usagePercent || 0) }}%</span>
            </div>
            <div class="status-indicator network" :class="getNetworkStatusClass()">
              <span class="status-label">NET</span>
              <span class="status-value">{{ formatBytes(networkData.speed || 0) }}/s</span>
            </div>
          </div>
        </div>
      </header>

      <!-- Main Content Area -->
      <main class="main-content">
        <!-- Left Sidebar -->
        <aside class="sidebar left-sidebar" :class="{ collapsed: sidebarCollapsed }">
          <div class="sidebar-header">
            <button @click="toggleSidebar" class="sidebar-toggle">
              <span class="toggle-icon">{{ sidebarCollapsed ? '▶' : '◀' }}</span>
            </button>
            <span v-if="!sidebarCollapsed" class="sidebar-title">SYSTEM</span>
          </div>

          <div v-if="!sidebarCollapsed" class="sidebar-content">
            <!-- System Monitor Mini -->
            <div class="sidebar-section">
              <h4 class="section-title">MONITOR</h4>
              <div class="mini-monitor">
                <div class="monitor-item">
                  <span class="monitor-label">CPU</span>
                  <div class="monitor-bar">
                    <div class="monitor-fill cpu" :style="{ width: `${systemData.cpu?.usage || 0}%` }"></div>
                  </div>
                </div>
                <div class="monitor-item">
                  <span class="monitor-label">RAM</span>
                  <div class="monitor-bar">
                    <div class="monitor-fill memory" :style="{ width: `${systemData.memory?.usagePercent || 0}%` }"></div>
                  </div>
                </div>
                <div class="monitor-item">
                  <span class="monitor-label">SWAP</span>
                  <div class="monitor-bar">
                    <div class="monitor-fill swap" :style="{ width: `${systemData.swap?.usagePercent || 0}%` }"></div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Quick Launch -->
            <div class="sidebar-section">
              <h4 class="section-title">QUICK LAUNCH</h4>
              <div class="quick-launch">
                <button @click="openTerminal" class="launch-btn">
                  <span class="launch-icon">💻</span>
                  <span class="launch-label">Terminal</span>
                </button>
                <button @click="openFileManager" class="launch-btn">
                  <span class="launch-icon">📁</span>
                  <span class="launch-label">Files</span>
                </button>
                <button @click="openNetworkMonitor" class="launch-btn">
                  <span class="launch-icon">🌐</span>
                  <span class="launch-label">Network</span>
                </button>
                <button @click="openSystemMonitor" class="launch-btn">
                  <span class="launch-icon">📊</span>
                  <span class="launch-label">Monitor</span>
                </button>
              </div>
            </div>

            <!-- Processes -->
            <div class="sidebar-section">
              <h4 class="section-title">PROCESSES</h4>
              <div class="process-list">
                <div
                  v-for="(process, index) in topProcesses.slice(0, 5)"
                  :key="index"
                  class="process-item"
                >
                  <span class="process-name">{{ process.name }}</span>
                  <span class="process-cpu">{{ process.cpu }}%</span>
                </div>
              </div>
            </div>
          </div>
        </aside>

        <!-- Central Workspace -->
        <div class="workspace">
          <!-- Window Management -->
          <div class="windows-container">
            <!-- Terminal Window -->
            <div
              v-if="windows.terminal.visible"
              class="desktop-window terminal-window"
              :class="{ active: activeWindow === 'terminal', maximized: windows.terminal.maximized }"
              :style="getWindowStyle('terminal')"
              @mousedown="setActiveWindow('terminal')"
            >
              <div class="window-header" @mousedown="startDrag('terminal', $event)">
                <div class="window-title">
                  <span class="window-icon">💻</span>
                  <span class="window-text">Terminal</span>
                </div>
                <div class="window-controls">
                  <button @click="minimizeWindow('terminal')" class="window-control minimize">─</button>
                  <button @click="maximizeWindow('terminal')" class="window-control maximize">□</button>
                  <button @click="closeWindow('terminal')" class="window-control close">✕</button>
                </div>
              </div>
              <div class="window-content">
                <TerminalEmulator />
              </div>
            </div>

            <!-- File Browser Window -->
            <div
              v-if="windows.fileBrowser.visible"
              class="desktop-window file-browser-window"
              :class="{ active: activeWindow === 'fileBrowser', maximized: windows.fileBrowser.maximized }"
              :style="getWindowStyle('fileBrowser')"
              @mousedown="setActiveWindow('fileBrowser')"
            >
              <div class="window-header" @mousedown="startDrag('fileBrowser', $event)">
                <div class="window-title">
                  <span class="window-icon">📁</span>
                  <span class="window-text">File System</span>
                </div>
                <div class="window-controls">
                  <button @click="minimizeWindow('fileBrowser')" class="window-control minimize">─</button>
                  <button @click="maximizeWindow('fileBrowser')" class="window-control maximize">□</button>
                  <button @click="closeWindow('fileBrowser')" class="window-control close">✕</button>
                </div>
              </div>
              <div class="window-content">
                <FileBrowser />
              </div>
            </div>

            <!-- System Monitor Window -->
            <div
              v-if="windows.systemMonitor.visible"
              class="desktop-window system-monitor-window"
              :class="{ active: activeWindow === 'systemMonitor', maximized: windows.systemMonitor.maximized }"
              :style="getWindowStyle('systemMonitor')"
              @mousedown="setActiveWindow('systemMonitor')"
            >
              <div class="window-header" @mousedown="startDrag('systemMonitor', $event)">
                <div class="window-title">
                  <span class="window-icon">📊</span>
                  <span class="window-text">System Monitor</span>
                </div>
                <div class="window-controls">
                  <button @click="minimizeWindow('systemMonitor')" class="window-control minimize">─</button>
                  <button @click="maximizeWindow('systemMonitor')" class="window-control maximize">□</button>
                  <button @click="closeWindow('systemMonitor')" class="window-control close">✕</button>
                </div>
              </div>
              <div class="window-content">
                <SystemMonitor />
              </div>
            </div>

            <!-- Network Monitor Window -->
            <div
              v-if="windows.networkMonitor.visible"
              class="desktop-window network-monitor-window"
              :class="{ active: activeWindow === 'networkMonitor', maximized: windows.networkMonitor.maximized }"
              :style="getWindowStyle('networkMonitor')"
              @mousedown="setActiveWindow('networkMonitor')"
            >
              <div class="window-header" @mousedown="startDrag('networkMonitor', $event)">
                <div class="window-title">
                  <span class="window-icon">🌐</span>
                  <span class="window-text">Network Monitor</span>
                </div>
                <div class="window-controls">
                  <button @click="minimizeWindow('networkMonitor')" class="window-control minimize">─</button>
                  <button @click="maximizeWindow('networkMonitor')" class="window-control maximize">□</button>
                  <button @click="closeWindow('networkMonitor')" class="window-control close">✕</button>
                </div>
              </div>
              <div class="window-content">
                <NetworkMonitor />
              </div>
            </div>

            <!-- Default desktop view when no windows are open -->
            <div v-if="!hasOpenWindows" class="desktop-welcome">
              <div class="welcome-content">
                <h1 class="welcome-title glitch">aDex-UI</h1>
                <p class="welcome-subtitle">Advanced Desktop Environment</p>
                <div class="welcome-stats">
                  <div class="stat-item">
                    <span class="stat-label">System Ready</span>
                    <span class="stat-value">✓</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Services Active</span>
                    <span class="stat-value">{{ activeServices }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Theme</span>
                    <span class="stat-value">{{ currentTheme }}</span>
                  </div>
                </div>
                <div class="welcome-actions">
                  <button @click="openTerminal" class="welcome-btn primary">
                    <span class="btn-icon">💻</span>
                    Open Terminal
                  </button>
                  <button @click="openSystemMonitor" class="welcome-btn secondary">
                    <span class="btn-icon">📊</span>
                    System Monitor
                  </button>
                  <button @click="showSettingsModal" class="welcome-btn secondary">
                    <span class="btn-icon">⚙️</span>
                    Settings
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Sidebar -->
        <aside class="sidebar right-sidebar">
          <div class="sidebar-header">
            <span class="sidebar-title">ACTIVITY</span>
          </div>

          <div class="sidebar-content">
            <!-- Notifications -->
            <div class="sidebar-section">
              <h4 class="section-title">NOTIFICATIONS</h4>
              <div class="notification-list">
                <div
                  v-for="(notification, index) in recentNotifications"
                  :key="index"
                  class="notification-item"
                  :class="notification.type"
                >
                  <span class="notification-icon">{{ getNotificationIcon(notification.type) }}</span>
                  <span class="notification-message">{{ notification.message }}</span>
                  <span class="notification-time">{{ formatTime(notification.timestamp) }}</span>
                </div>
              </div>
            </div>

            <!-- Quick Stats -->
            <div class="sidebar-section">
              <h4 class="section-title">STATISTICS</h4>
              <div class="stats-grid">
                <div class="stat-card">
                  <span class="stat-value">{{ systemData.processes?.length || 0 }}</span>
                  <span class="stat-label">Processes</span>
                </div>
                <div class="stat-card">
                  <span class="stat-value">{{ formatBytes(systemData.memory?.used || 0) }}</span>
                  <span class="stat-label">Memory Used</span>
                </div>
                <div class="stat-card">
                  <span class="stat-value">{{ networkData.connections || 0 }}</span>
                  <span class="stat-label">Connections</span>
                </div>
                <div class="stat-card">
                  <span class="stat-value">{{ systemData.uptimeHours || 0 }}h</span>
                  <span class="stat-label">Uptime</span>
                </div>
              </div>
            </div>
          </div>
        </aside>
      </main>

      <!-- Bottom Status Bar -->
      <footer class="status-bar">
        <div class="status-left">
          <span class="status-item">Ready</span>
          <span class="status-item">{{ activeConnections }} connections</span>
          <span class="status-item">{{ systemData.loadAvg || '0.00' }} load</span>
        </div>

        <div class="status-center">
          <div class="status-indicators">
            <div class="indicator" :class="{ active: soundEnabled }" title="Audio">
              <span class="indicator-icon">{{ soundEnabled ? '🔊' : '🔇' }}</span>
            </div>
            <div class="indicator" :class="{ active: networkStatus === 'connected' }" title="Network">
              <span class="indicator-icon">🌐</span>
            </div>
            <div class="indicator" :class="{ active: true }" title="Power">
              <span class="indicator-icon">⚡</span>
            </div>
          </div>
        </div>

        <div class="status-right">
          <span class="status-item">{{ systemData.os || 'Unknown OS' }}</span>
          <span class="status-item">{{ systemData.kernel || '' }}</span>
          <span class="status-item">{{ systemData.architecture || '' }}</span>
        </div>
      </footer>
    </div>

    <!-- Global Components -->
    <Teleport to="body">
      <!-- Settings Modal -->
      <SettingsModal 
        v-if="showSettings"
        ref="settingsModalRef" 
        @settings-changed="handleSettingsChange"
        @close="hideSettingsModal"
      />

      <!-- Theme Manager -->
      <ThemeManager 
        v-if="showThemes"
        ref="themeManagerRef"
        @close="hideThemeManager"
      />

      <!-- Sound System -->
      <SoundSystem ref="soundSystemRef" />

      <!-- Virtual Keyboard -->
      <VirtualKeyboard 
        v-if="showKeyboard"
        ref="virtualKeyboardRef"
        @close="toggleVirtualKeyboard"
      />
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useDateFormat } from '@vueuse/core'
import TerminalEmulator from '~/components/terminal/TerminalEmulator.vue'
import FileBrowser from '~/components/filesystem/FileBrowser.vue'
import SystemMonitor from '~/components/system/SystemMonitor.vue'
import NetworkMonitor from '~/components/system/NetworkMonitor.vue'
import SettingsModal from '~/components/ui/SettingsModal.vue'
import ThemeManager from '~/components/theme/ThemeManager.vue'
import SoundSystem from '~/components/audio/SoundSystem.vue'
import VirtualKeyboard from '~/components/ui/VirtualKeyboard.vue'
import SplashScreen from '~/components/ui/SplashScreen.vue'
import { useSystemStore } from '~/stores/system'
import { useNetworkStore } from '~/stores/network'
import { useFilesystemStore } from '~/stores/filesystem'
import { useTerminalStore } from '~/stores/terminal'
import { useThemeStore } from '~/stores/theme'
import { useAppStore } from '~/stores/app'

// Store instances
const systemStore = useSystemStore()
const themeStore = useThemeStore()
const appStore = useAppStore()

// Template refs
const settingsModalRef = ref()
const themeManagerRef = ref()
const soundSystemRef = ref()
const virtualKeyboardRef = ref()

// Reactive data
const currentTime = ref<string>('')
const sidebarCollapsed = ref<boolean>(false)
const activeWindow = ref<string>('')
const soundEnabled = ref<boolean>(true)
const networkStatus = ref<string>('connected')
const activeServices = ref<number>(4)

// Splash screen state
const showSplash = ref<boolean>(true)

// Modal/Panel visibility state
const showSettings = ref<boolean>(false)
const showThemes = ref<boolean>(false)
const showKeyboard = ref<boolean>(false)

// Window state with position (positions relative to workspace, not full screen)
const windows = ref({
  terminal: { visible: false, maximized: false, x: 50, y: 20, width: 700, height: 450 },
  fileBrowser: { visible: false, maximized: false, x: 100, y: 40, width: 600, height: 400 },
  systemMonitor: { visible: false, maximized: false, x: 150, y: 60, width: 600, height: 450 },
  networkMonitor: { visible: false, maximized: false, x: 200, y: 80, width: 550, height: 350 }
})

// Drag state
const dragging = ref<string | null>(null)
const dragOffset = ref({ x: 0, y: 0 })

// Start dragging a window
const startDrag = (windowId: string, event: MouseEvent) => {
  // Ignore if clicking on window controls
  if ((event.target as HTMLElement).closest('.window-controls')) return
  
  const win = windows.value[windowId as keyof typeof windows.value]
  if (win.maximized) return
  
  // Prevent text selection during drag
  event.preventDefault()
  
  dragging.value = windowId
  dragOffset.value = {
    x: event.clientX - win.x,
    y: event.clientY - win.y
  }
  setActiveWindow(windowId)
  
  // Add dragging class to body for cursor styling
  document.body.classList.add('window-dragging')
  
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

// Handle dragging
const onDrag = (event: MouseEvent) => {
  if (!dragging.value) return
  const win = windows.value[dragging.value as keyof typeof windows.value]
  win.x = Math.max(0, event.clientX - dragOffset.value.x)
  win.y = Math.max(0, event.clientY - dragOffset.value.y)
}

// Stop dragging
const stopDrag = () => {
  dragging.value = null
  document.body.classList.remove('window-dragging')
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

// Get window style
const getWindowStyle = (windowId: string) => {
  const win = windows.value[windowId as keyof typeof windows.value]
  if (win.maximized) return {}
  return {
    left: `${win.x}px`,
    top: `${win.y}px`,
    width: `${win.width}px`,
    height: `${win.height}px`
  }
}

// System data from stores
const systemData = computed(() => systemStore.systemData)
const networkData = computed(() => systemStore.networkData)
const topProcesses = computed(() => systemStore.topProcesses)
const recentNotifications = computed(() => appStore.unacknowledgedAlerts.slice(0, 5))

// Theme settings
const currentTheme = computed(() => themeStore.currentTheme?.name || 'Default')
const themeSettings = computed(() => themeStore.settings)

// Computed properties
const desktopClasses = computed(() => ({
  'sidebar-collapsed': sidebarCollapsed.value,
  'sound-enabled': soundEnabled.value,
  'particles-enabled': themeSettings.value.particles,
  'scanlines-enabled': themeSettings.value.scanlines,
  'high-contrast': themeSettings.value.highContrast,
  'reduce-motion': themeSettings.value.reduceMotion
}))

const hasOpenWindows = computed(() => {
  return Object.values(windows.value).some(window => window.visible)
})

const activeConnections = computed(() => {
  return networkData.value.connections || 0
})

// Methods
const updateTime = () => {
  currentTime.value = useDateFormat(new Date(), 'HH:mm:ss').value
}

const formatUptime = (seconds: number): string => {
  if (!seconds) return '0m'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const formatTime = (timestamp: number): string => {
  return useDateFormat(new Date(timestamp), 'HH:mm').value
}

const getParticleStyle = (index: number) => {
  return {
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    animationDelay: `${Math.random() * 10}s`,
    animationDuration: `${10 + Math.random() * 20}s`
  }
}

const getCpuStatusClass = () => {
  const usage = systemData.value.cpu?.usage || 0
  if (usage > 80) return 'critical'
  if (usage > 60) return 'warning'
  return 'normal'
}

const getMemoryStatusClass = () => {
  const usage = systemData.value.memory?.usagePercent || 0
  if (usage > 80) return 'critical'
  if (usage > 60) return 'warning'
  return 'normal'
}

const getNetworkStatusClass = () => {
  return networkStatus.value === 'connected' ? 'normal' : 'warning'
}

const getNotificationIcon = (type: string): string => {
  const icons = {
    info: 'ℹ️',
    success: '✅',
    warning: '⚠️',
    error: '❌'
  }
  return icons[type as keyof typeof icons] || 'ℹ️'
}

// Window management
const openTerminal = () => {
  windows.value.terminal.visible = true
  windows.value.terminal.maximized = false
  activeWindow.value = 'terminal'
  soundSystemRef.value?.playUISound('click')
}

const openFileManager = () => {
  windows.value.fileBrowser.visible = true
  windows.value.fileBrowser.maximized = false
  activeWindow.value = 'fileBrowser'
  soundSystemRef.value?.playUISound('click')
}

const openSystemMonitor = () => {
  windows.value.systemMonitor.visible = true
  windows.value.systemMonitor.maximized = false
  activeWindow.value = 'systemMonitor'
  soundSystemRef.value?.playUISound('click')
}

const openNetworkMonitor = () => {
  windows.value.networkMonitor.visible = true
  windows.value.networkMonitor.maximized = false
  activeWindow.value = 'networkMonitor'
  soundSystemRef.value?.playUISound('click')
}

const setActiveWindow = (windowId: string) => {
  activeWindow.value = windowId
}

const minimizeWindow = (windowId: string) => {
  if (windows.value[windowId as keyof typeof windows.value]) {
    windows.value[windowId as keyof typeof windows.value].visible = false
  }
  soundSystemRef.value?.playUISound('click')
}

const maximizeWindow = (windowId: string) => {
  if (windows.value[windowId as keyof typeof windows.value]) {
    windows.value[windowId as keyof typeof windows.value].maximized =
      !windows.value[windowId as keyof typeof windows.value].maximized
  }
  soundSystemRef.value?.playUISound('click')
}

const closeWindow = (windowId: string) => {
  if (windows.value[windowId as keyof typeof windows.value]) {
    windows.value[windowId as keyof typeof windows.value].visible = false
    windows.value[windowId as keyof typeof windows.value].maximized = false
  }
  soundSystemRef.value?.playUISound('click')
}

// UI controls
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
  soundSystemRef.value?.playUISound('click')
}

const showSettingsModal = () => {
  showSettings.value = true
  console.log('Settings modal opened')
}

const hideSettingsModal = () => {
  showSettings.value = false
}

const showThemeManager = () => {
  showThemes.value = true
  console.log('Theme manager opened')
}

const hideThemeManager = () => {
  showThemes.value = false
}

const toggleSoundPanel = () => {
  soundEnabled.value = !soundEnabled.value
  console.log('Sound toggled:', soundEnabled.value)
}

const toggleVirtualKeyboard = () => {
  showKeyboard.value = !showKeyboard.value
  console.log('Keyboard toggled:', showKeyboard.value)
}

const handleSettingsChange = (settings: any) => {
  console.log('Settings changed:', settings)
  // Apply settings changes
}

// Splash screen completion handler
const onSplashComplete = () => {
  showSplash.value = false
  console.log('Splash screen complete, app ready')
}

// Application controls
const minimizeApp = () => {
  // Use Wails v2 runtime to minimize the window
  try {
    const runtime = (window as any).runtime
    if (runtime?.WindowMinimise) {
      runtime.WindowMinimise()
    }
  } catch (e) {
    console.error('Failed to minimize app:', e)
  }
  soundSystemRef.value?.playUISound('click')
}

const toggleFullscreen = () => {
  try {
    const runtime = (window as any).runtime
    if (runtime?.WindowToggleMaximise) {
      runtime.WindowToggleMaximise()
    } else if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen()
    } else {
      document.exitFullscreen()
    }
  } catch (e) {
    console.error('Failed to toggle fullscreen:', e)
  }
  soundSystemRef.value?.playUISound('click')
}

const closeApp = () => {
  if (confirm('Are you sure you want to close aDex-UI?')) {
    // Use Wails v2 runtime to quit (window.runtime.Quit)
    try {
      const runtime = (window as any).runtime
      if (runtime?.Quit) {
        runtime.Quit()
      } else {
        // Fallback for browser testing
        window.close()
      }
    } catch (e) {
      console.error('Failed to close app:', e)
      window.close()
    }
  }
}

const handleQuit = () => {
  closeApp()
}

const toggleWindow = (windowId: string) => {
  const win = windows.value[windowId as keyof typeof windows.value]
  if (win) {
    win.visible = !win.visible
    if (win.visible) {
      activeWindow.value = windowId
    }
  }
  soundSystemRef.value?.playUISound('click')
}

// Event listeners for component communication
const setupEventListeners = () => {
  // Terminal toggle
  window.addEventListener('toggle-terminal', () => {
    toggleWindow('terminal')
  })

  // File browser toggle
  window.addEventListener('toggle-filebrowser', () => {
    toggleWindow('fileBrowser')
  })

  // System monitor toggle
  window.addEventListener('toggle-systemmonitor', () => {
    toggleWindow('systemMonitor')
  })

  // Settings modal
  window.addEventListener('open-settings', () => {
    showSettingsModal()
  })

  // Network monitor toggle
  window.addEventListener('toggle-networkmonitor', () => {
    toggleWindow('networkMonitor')
  })

  // Sound system toggle
  window.addEventListener('toggle-sound', () => {
    soundEnabled.value = !soundEnabled.value
    if (soundSystemRef.value) {
      soundSystemRef.value.setEnabled(soundEnabled.value)
    }
  })

  // Keyboard shortcuts
  window.addEventListener('keydown', (e) => {
    // Ctrl + \` for terminal
    if (e.ctrlKey && e.key === '`') {
      e.preventDefault()
      toggleWindow('terminal')
    }

    // Ctrl + Shift + F for file browser
    if (e.ctrlKey && e.shiftKey && e.key === 'F') {
      e.preventDefault()
      toggleWindow('fileBrowser')
    }

    // Ctrl + Shift + S for system monitor
    if (e.ctrlKey && e.shiftKey && e.key === 'S') {
      e.preventDefault()
      toggleWindow('systemMonitor')
    }

    // Ctrl + , for settings
    if (e.ctrlKey && e.key === ',') {
      e.preventDefault()
      showSettingsModal()
    }

    // Ctrl + Q to quit
    if (e.ctrlKey && e.key === 'q') {
      e.preventDefault()
      handleQuit()
    }
  })
}

// Lifecycle
let timeInterval: NodeJS.Timeout

onMounted(async () => {
  try {
    // Initialize time
    updateTime()
    timeInterval = setInterval(updateTime, 1000)

    // Initialize app store first
    const appStore = useAppStore()
    await appStore.initialize()

    // Initialize all stores in parallel
    await Promise.all([
      systemStore.initialize(),
      themeStore.initialize(),
      useNetworkStore().initialize(),
      useFilesystemStore().initialize(),
      useTerminalStore().initialize()
    ])

    // Setup event listeners for component communication
    setupEventListeners()

    // Show welcome notification
    appStore.addAlert({
      type: 'success',
      title: 'Welcome',
      message: 'aDex-UI initialized successfully',
      persistent: false
    })

    // Auto-open terminal on first load
    setTimeout(() => {
      openTerminal()
    }, 1000)

  } catch (error) {
    console.error('Failed to initialize aDex-UI:', error)

    // Show error notification
    appStore.addAlert({
      type: 'error',
      title: 'Initialization Error',
      message: 'Initialization failed. Some features may not work properly.',
      persistent: true
    })
  }
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
  systemStore.stopMonitoring()
})

// SEO
useHead({
  title: 'aDex-UI - Advanced Desktop Environment',
  meta: [
    { name: 'description', content: 'A futuristic desktop environment with terminal emulation, system monitoring, and file management' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { name: 'theme-color', content: '#00ff00' }
  ],
  link: [
    { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
  ]
})
</script>

<style scoped>
/* Import base styles and animations */
@import '~/assets/css/main.css';
@import '~/assets/css/animations.css';

.adex-desktop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--background);
  color: var(--text-primary);
  font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
  overflow: hidden;
  z-index: 1;
}

/* Background Effects */
.background-effects {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 0;
}

.particle-container {
  position: absolute;
  width: 100%;
  height: 100%;
}

.particle {
  position: absolute;
  width: 2px;
  height: 2px;
  background: var(--primary-400);
  border-radius: 50%;
  opacity: 0.6;
  animation: float linear infinite;
}

.scanlines {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(
    transparent 50%,
    rgba(0, 255, 65, 0.03) 51%
  );
  background-size: 100% 4px;
  animation: scanlines 8s linear infinite;
}

.background-grid {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image:
    linear-gradient(rgba(0, 255, 255, 0.15) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 255, 255, 0.15) 1px, transparent 1px);
  background-size: 40px 40px;
  animation: grid-move 20s linear infinite;
  pointer-events: none;
}

/* Additional sci-fi overlay effects */
.background-effects::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: 
    radial-gradient(ellipse at 20% 20%, rgba(0, 255, 255, 0.03) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 80%, rgba(0, 255, 128, 0.03) 0%, transparent 50%);
  pointer-events: none;
}

/* Scanlines effect like eDex-UI */
.background-effects::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: repeating-linear-gradient(
    0deg,
    rgba(0, 0, 0, 0.1),
    rgba(0, 0, 0, 0.1) 1px,
    transparent 1px,
    transparent 2px
  );
  pointer-events: none;
  opacity: 0.3;
}

@keyframes grid-move {
  0% { transform: translate(0, 0); }
  100% { transform: translate(50px, 50px); }
}

/* Desktop Layout */
.desktop-layout {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100vh;
  z-index: 1;
}

/* Top Bar */
.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 32px;
  background: rgba(5, 10, 15, 0.98);
  border-bottom: 1px solid rgba(0, 255, 255, 0.3);
  padding: 0 15px;
  z-index: 100;
  box-shadow: 0 2px 10px rgba(0, 255, 255, 0.1);
}

.top-bar-left,
.top-bar-center,
.top-bar-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.app-logo {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.logo-text {
  font-size: 16px;
  font-weight: bold;
  color: var(--primary-400);
}

.logo-version {
  font-size: 10px;
  color: var(--text-secondary);
}

.system-info {
  display: flex;
  gap: 15px;
  font-size: 11px;
  color: var(--text-secondary);
}

.window-controls {
  display: flex;
  gap: 2px;
}

.window-btn {
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 3px;
  transition: all 0.2s ease;
  font-size: 12px;
}

.window-btn:hover {
  background: var(--surface-elevated);
  color: var(--text-primary);
}

.window-btn.close:hover {
  background: var(--error);
  color: var(--text-primary);
}

.quick-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 32px;
  height: 32px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.action-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.system-status {
  display: flex;
  gap: 15px;
}

.status-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 9px;
  min-width: 50px;
}

.status-indicator.normal {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.status-indicator.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.status-indicator.critical {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.status-label {
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-value {
  font-weight: bold;
}

/* Main Content */
.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Sidebars */
.sidebar {
  background: rgba(5, 10, 15, 0.98);
  border-right: 1px solid rgba(0, 255, 255, 0.3);
  display: flex;
  flex-direction: column;
  transition: all 0.3s ease;
  z-index: 50;
  box-shadow: 
    inset -5px 0 15px rgba(0, 255, 255, 0.05),
    0 0 20px rgba(0, 255, 255, 0.1);
}

.left-sidebar {
  width: 250px;
}

.left-sidebar.collapsed {
  width: 50px;
}

.right-sidebar {
  width: 200px;
  border-left: 1px solid var(--surface-border);
  border-right: none;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  border-bottom: 1px solid var(--surface-border);
  min-height: 50px;
}

.sidebar-toggle {
  background: transparent;
  border: none;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 12px;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.sidebar-toggle:hover {
  background: var(--surface-elevated);
}

.sidebar-title {
  font-size: 11px;
  font-weight: bold;
  color: var(--primary-400);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
}

.sidebar-section {
  margin-bottom: 25px;
}

.section-title {
  font-size: 10px;
  font-weight: bold;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 10px;
}

/* Mini Monitor */
.mini-monitor {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.monitor-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.monitor-label {
  font-size: 9px;
  color: var(--text-secondary);
  min-width: 35px;
}

.monitor-bar {
  flex: 1;
  height: 4px;
  background: var(--surface-elevated);
  border-radius: 2px;
  overflow: hidden;
}

.monitor-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

.monitor-fill.cpu {
  background: linear-gradient(90deg, #0ea5e9, #38bdf8);
}

.monitor-fill.memory {
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

.monitor-fill.swap {
  background: linear-gradient(90deg, #10b981, #34d399);
}

/* Quick Launch */
.quick-launch {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.launch-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 9px;
}

.launch-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
  transform: translateY(-2px);
}

.launch-icon {
  font-size: 16px;
}

/* Process List */
.process-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.process-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 9px;
  padding: 2px 0;
}

.process-name {
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 120px;
}

.process-cpu {
  color: var(--accent-400);
  font-weight: bold;
}

/* Workspace */
.workspace {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--background);
  position: relative;
  overflow: hidden;
}

.windows-container {
  flex: 1;
  position: relative;
  padding: 20px;
}

/* Desktop Windows */
.desktop-window {
  position: absolute;
  background: rgba(10, 15, 20, 0.95);
  border: 1px solid rgba(0, 255, 255, 0.3);
  border-radius: 4px;
  box-shadow: 
    0 0 20px rgba(0, 255, 255, 0.1),
    inset 0 0 20px rgba(0, 255, 255, 0.02);
  overflow: hidden;
  /* Only transition non-position properties to allow smooth dragging */
  transition: border-color 0.3s ease, box-shadow 0.3s ease, width 0.2s ease, height 0.2s ease;
  min-width: 400px;
  min-height: 300px;
}

/* Dragging state */
body.window-dragging {
  cursor: grabbing !important;
  user-select: none !important;
}

body.window-dragging * {
  cursor: grabbing !important;
  user-select: none !important;
}

.desktop-window.active {
  border-color: rgba(0, 255, 255, 0.6);
  box-shadow: 
    0 0 30px rgba(0, 255, 255, 0.2),
    0 0 60px rgba(0, 255, 255, 0.1),
    inset 0 0 30px rgba(0, 255, 255, 0.03);
  z-index: 10;
}

.desktop-window.maximized {
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  bottom: 0 !important;
  width: 100% !important;
  height: 100% !important;
  border-radius: 0;
}

.window-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 35px;
  background: var(--surface-elevated);
  border-bottom: 1px solid var(--surface-border);
  padding: 0 15px;
  cursor: grab;
  user-select: none;
}

.window-header:active {
  cursor: grabbing;
}

.window-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.window-icon {
  font-size: 14px;
}

.window-text {
  font-size: 12px;
  font-weight: bold;
  color: var(--text-primary);
}

.window-controls {
  display: flex;
  gap: 2px;
}

.window-control {
  width: 20px;
  height: 20px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 3px;
  font-size: 10px;
  transition: all 0.2s ease;
}

.window-control:hover {
  background: var(--surface);
  color: var(--text-primary);
}

.window-control.close:hover {
  background: var(--error);
  color: var(--text-primary);
}

.window-content {
  height: calc(100% - 35px);
  overflow: hidden;
}

/* Window positioning */
.terminal-window {
  top: 20px;
  left: 20px;
  width: 800px;
  height: 500px;
}

.file-browser-window {
  top: 80px;
  left: 300px;
  width: 700px;
  height: 600px;
}

.system-monitor-window {
  top: 40px;
  right: 20px;
  width: 600px;
  height: 400px;
}

.network-monitor-window {
  bottom: 20px;
  right: 20px;
  width: 700px;
  height: 450px;
}

/* Welcome Screen */
.desktop-welcome {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 40px;
}

.welcome-content {
  text-align: center;
  max-width: 500px;
}

.welcome-title {
  font-size: 48px;
  font-weight: bold;
  color: var(--primary-400);
  margin-bottom: 10px;
}

.welcome-subtitle {
  font-size: 16px;
  color: var(--text-secondary);
  margin-bottom: 30px;
}

.welcome-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 40px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.stat-label {
  font-size: 11px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.stat-value {
  font-size: 18px;
  font-weight: bold;
  color: var(--accent-400);
}

.welcome-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
}

.welcome-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 12px;
  font-weight: bold;
}

.welcome-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
  transform: translateY(-2px);
}

.welcome-btn.primary {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

.welcome-btn.primary:hover {
  background: var(--primary-600);
}

/* Notifications */
.notification-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px;
  background: var(--surface-elevated);
  border-radius: 4px;
  font-size: 9px;
}

.notification-item.success {
  border-left: 2px solid var(--success);
}

.notification-item.warning {
  border-left: 2px solid var(--warning);
}

.notification-item.error {
  border-left: 2px solid var(--error);
}

.notification-icon {
  font-size: 10px;
}

.notification-message {
  flex: 1;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-time {
  color: var(--text-secondary);
  font-size: 8px;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px;
  background: var(--surface-elevated);
  border-radius: 6px;
}

.stat-card .stat-value {
  font-size: 16px;
  font-weight: bold;
  color: var(--primary-400);
}

.stat-card .stat-label {
  font-size: 8px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Status Bar */
.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 25px;
  background: var(--surface);
  border-top: 1px solid var(--surface-border);
  padding: 0 20px;
  font-size: 10px;
  z-index: 100;
}

.status-left,
.status-center,
.status-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.status-item {
  color: var(--text-secondary);
}

.status-indicators {
  display: flex;
  gap: 8px;
}

.indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--surface-elevated);
  transition: all 0.2s ease;
}

.indicator.active {
  background: var(--primary-500);
  box-shadow: 0 0 10px var(--primary-500);
}

.indicator-icon {
  font-size: 10px;
}

/* Responsive Design */
@media (max-width: 1024px) {
  .left-sidebar {
    width: 200px;
  }

  .right-sidebar {
    width: 150px;
  }

  .desktop-window {
    min-width: 300px;
  }

  .terminal-window {
    width: 600px;
    height: 400px;
  }
}

@media (max-width: 768px) {
  .top-bar {
    padding: 0 10px;
  }

  .system-info {
    display: none;
  }

  .quick-actions {
    gap: 4px;
  }

  .action-btn {
    width: 28px;
    height: 28px;
    font-size: 12px;
  }

  .left-sidebar {
    width: 60px;
  }

  .left-sidebar.collapsed {
    width: 50px;
  }

  .sidebar-title,
  .section-title {
    display: none;
  }

  .quick-launch {
    grid-template-columns: 1fr;
  }

  .welcome-actions {
    flex-direction: column;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }
}

/* Performance Optimizations */
.reduce-motion * {
  animation-duration: 0.01ms !important;
  animation-iteration-count: 1 !important;
  transition-duration: 0.01ms !important;
}

.high-contrast {
  --surface-border: #ffffff;
  --text-secondary: #ffffff;
}

/* Scrollbar Styling */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: var(--surface);
}

::-webkit-scrollbar-thumb {
  background: var(--surface-border);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--primary-500);
}

/* Glitch effect for title */
.glitch {
  position: relative;
  color: var(--primary-400);
  font-size: 16px;
  font-weight: bold;
  text-transform: uppercase;
  text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  animation: glitch 2s infinite;
}

@keyframes glitch {
  0%, 90%, 100% {
    text-shadow:
      2px 2px 0 var(--accent-500),
      -2px -2px 0 var(--primary-600);
  }
  92% {
    text-shadow:
      -2px 2px 0 var(--accent-500),
      2px -2px 0 var(--primary-600);
  }
  94% {
    text-shadow:
      2px -2px 0 var(--accent-500),
      -2px 2px 0 var(--primary-600);
  }
  96% {
    text-shadow:
      -2px -2px 0 var(--accent-500),
      2px 2px 0 var(--primary-600);
  }
}
</style>
