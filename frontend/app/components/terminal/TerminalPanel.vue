<template>
  <div
    class="terminal-panel"
    :class="{
      'terminal-panel--focused': isFocused,
      'terminal-panel--inactive': !session?.active
    }"
    @click="handleContainerClick"
    @contextmenu.prevent="handleContextMenu"
  >
    <!-- Terminal Header -->
    <div class="terminal-header" v-if="!hideHeader">
      <div class="terminal-info">
        <div class="terminal-title">{{ tab?.title || 'Terminal' }}</div>
        <div class="terminal-path" v-if="session?.cwd">{{ session.cwd }}</div>
      </div>
      <div class="terminal-controls">
        <!-- Bell indicator -->
        <div class="terminal-bell-indicator" v-if="bellNotifications.length > 0">
          <Icon name="carbon:notification" />
          <span class="terminal-bell-badge">{{ bellNotifications.length }}</span>
        </div>

        <!-- Font Configuration Selector -->
        <div class="font-selector-wrapper" v-if="!hideHeader">
          <button
            class="terminal-control"
            @click="showFontSelector = !showFontSelector"
            title="Change font"
          >
            <Icon name="carbon:text-font" />
          </button>

          <!-- Dropdown Font Selector -->
          <div v-if="showFontSelector" class="font-selector-dropdown">
            <div class="dropdown-header">
              <h4>Font Configuration</h4>
              <button @click="showFontSelector = false" class="dropdown-close">
                <Icon name="carbon:close" />
              </button>
            </div>
            <div class="font-options-list">
              <button
                v-for="config in availableFontConfigs"
                :key="config.id"
                class="font-option"
                :class="{ active: currentFontConfig?.id === config.id }"
                @click="selectFontConfiguration(config.id)"
              >
                <div class="font-option-info">
                  <span class="font-option-name">{{ config.name }}</span>
                  <span class="font-option-family">{{ config.family }} • {{ config.size }}px</span>
                </div>
                <Icon v-if="currentFontConfig?.id === config.id" name="carbon:checkmark" class="font-option-selected" />
              </button>
            </div>
            <div class="font-selector-footer">
              <button @click="openFontSettings" class="font-selector-action">
                <Icon name="carbon:settings" class="w-4 h-4" />
                Font Settings
              </button>
            </div>
          </div>
        </div>

        <!-- Color Scheme Selector -->
        <div class="color-scheme-selector-wrapper" v-if="!hideHeader">
          <button
            class="terminal-control"
            @click="showColorSchemeSelector = !showColorSchemeSelector"
            title="Change color scheme"
          >
            <Icon name="carbon:palette" />
          </button>

          <!-- Dropdown Color Scheme Selector -->
          <div v-if="showColorSchemeSelector" class="color-scheme-dropdown">
            <div class="dropdown-header">
              <h4>Color Schemes</h4>
              <button @click="showColorSchemeSelector = false" class="dropdown-close">
                <Icon name="carbon:close" />
              </button>
            </div>
            <div class="schemes-list">
              <button
                v-for="scheme in availableSchemes"
                :key="scheme.id"
                class="scheme-option"
                :class="{ active: currentScheme?.id === scheme.id }"
                @click="selectColorScheme(scheme.id)"
              >
                <div class="scheme-preview-colors">
                  <div class="color-dot" :style="{ backgroundColor: scheme.colors.background }"></div>
                  <div class="color-dot" :style="{ backgroundColor: scheme.colors.foreground }"></div>
                  <div class="color-dot" :style="{ backgroundColor: scheme.colors.red }"></div>
                  <div class="color-dot" :style="{ backgroundColor: scheme.colors.green }"></div>
                  <div class="color-dot" :style="{ backgroundColor: scheme.colors.blue }"></div>
                </div>
                <div class="scheme-info">
                  <span class="scheme-name">{{ scheme.displayName || scheme.name }}</span>
                  <span class="scheme-author" v-if="scheme.author">{{ scheme.author }}</span>
                </div>
                <Icon v-if="currentScheme?.id === scheme.id" name="carbon:checkmark" class="scheme-selected-icon" />
              </button>
            </div>
          </div>
        </div>

        <button
          class="terminal-control"
          @click="copyContent"
          title="Copy content"
          :disabled="!hasContent"
        >
          <Icon name="carbon:copy" />
        </button>
        <button
          class="terminal-control"
          @click="clearContent"
          title="Clear content"
          :disabled="!hasContent"
        >
          <Icon name="carbon:trash-can" />
        </button>
        <button
          class="terminal-control"
          @click="toggleFullscreen"
          :class="{ active: isFullscreen }"
          title="Toggle fullscreen"
        >
          <Icon :name="isFullscreen ? 'carbon:view-off' : 'carbon:view'" />
        </button>
      </div>
    </div>

    <!-- Terminal Content -->
    <div
      ref="terminalContent"
      class="terminal-content"
      :style="terminalStyle"
      @scroll="handleScroll"
    >
      <!-- xterm.js terminal will be mounted here -->
      <div ref="xtermContainer" class="xterm-container"></div>
    </div>

    <!-- Terminal Footer (resize handle) -->
    <div class="terminal-footer" v-if="!hideFooter">
      <div class="terminal-stats">
        <span class="terminal-stat" v-if="session">
          <Icon name="carbon:application" />
          PID: {{ session.pid || 'N/A' }}
        </span>
        <span class="terminal-stat" v-if="outputLength > 0">
          <Icon name="carbon:document" />
          {{ outputLength }} lines
        </span>
        <span class="terminal-stat">
          <Icon name="carbon:time" />
          {{ formatTime(session?.lastSeen) }}
        </span>
      </div>
      <div class="resize-handle" @mousedown="startResize">
        <Icon name="carbon:panel-expand" />
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="isLoading" class="terminal-loading">
      <div class="loading-spinner"></div>
      <div class="loading-text">Initializing terminal...</div>
    </div>

    <!-- Error Overlay -->
    <div v-if="error" class="terminal-error">
      <Icon name="carbon:warning-alt" />
      <div class="error-message">{{ error }}</div>
      <button @click="retryInitialization" class="error-retry">
        <Icon name="carbon:reset" />
        Retry
      </button>
    </div>

    <!-- Context Menu -->
    <TerminalContextMenu
      :visible="showContextMenu"
      :x="contextMenuPosition.x"
      :y="contextMenuPosition.y"
      :has-selection="hasActiveSelection"
      :has-output="hasOutput"
      :readonly="props.readonly"
      :terminal="getTerminal()"
      :output="terminalTab.output"
      :is-fullscreen="isFullscreen"
      @close="closeContextMenu"
      @copy-selection="handleContextMenuCopySelection"
      @copy-all="handleContextMenuCopyAll"
      @copy-output="handleContextMenuCopyOutput"
      @paste="handleContextMenuPaste"
      @clear-selection="handleContextMenuClearSelection"
      @clear-all="handleContextMenuClearAll"
      @find="handleContextMenuFind"
      @save-as="handleContextMenuSaveAs"
      @print="handleContextMenuPrint"
      @fullscreen="handleContextMenuFullscreen"
      @settings="handleContextMenuSettings"
    />

    <!-- Font Settings Dialog -->
    <div v-if="showFontSettingsDialog" class="font-settings-dialog-overlay" @click.self="showFontSettingsDialog = false">
      <div class="font-settings-dialog">
        <div class="font-settings-dialog-header">
          <h3 class="text-lg font-medium text-text">Font Settings</h3>
          <button @click="showFontSettingsDialog = false" class="text-text-secondary hover:text-text">
            <Icon name="carbon:close" class="w-5 h-5" />
          </button>
        </div>
        <div class="font-settings-dialog-content">
          <FontSettings
            :settings="fontSettings.value"
            @settings-changed="handleFontSettingsChanged"
            @apply-settings="handleFontSettingsApplied"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useTerminalTab } from '~/composables/useTerminal'
import { useXTerm } from '~/composables/useXTerm'
import { useClipboard } from '~/composables/useClipboard'
import { useBellNotifications } from '~/composables/useBellNotifications'
import { useTerminalStore } from '~/stores/terminal'
import { useUIStore } from '~/stores/ui'
import { useColorScheme } from '~/composables/useColorScheme'
import { useFontConfiguration } from '~/composables/useFontConfiguration'
import FontSettings from '~/components/font/FontSettings.vue'
import type { TerminalTheme } from '~/types/terminal'

// Props
interface Props {
  sessionId?: string
  hideHeader?: boolean
  hideFooter?: boolean
  readonly?: boolean
  autofocus?: boolean
  theme?: string
  fontFamily?: string
  fontSize?: number
  lineHeight?: number
  cursorStyle?: 'block' | 'underline' | 'bar'
  cursorBlink?: boolean
  opacity?: number
}

const props = withDefaults(defineProps<Props>(), {
  hideHeader: false,
  hideFooter: false,
  readonly: false,
  autofocus: true,
  theme: 'default-dark',
  fontFamily: 'JetBrains Mono',
  fontSize: 14,
  lineHeight: 1.5,
  cursorStyle: 'block',
  cursorBlink: false,
  opacity: 100
})

// Emits
const emit = defineEmits<{
  focus: []
  blur: []
  resize: [size: { width: number; height: number }]
  titleChange: [title: string]
  workingDirChange: [cwd: string]
  sessionCreated: [session: TerminalSession]
  sessionClosed: [sessionId: string]
}>()

// Composables
const terminalTab = useTerminalTab(props.sessionId || '')
const terminalStore = useTerminalStore()
const uiStore = useUIStore()
const clipboard = useClipboard({
  enableNotifications: true
})
const { currentScheme } = useColorScheme()
const { currentConfiguration: currentFontConfig, selectConfiguration, settings: fontSettings, updateSettings } = useFontConfiguration()

const {
  notifications: bellNotifications,
  ringBell,
  clearBellNotifications,
  setActive: setBellActive
} = useBellNotifications(computed(() => terminalTab.session))

// Refs
const terminalContent = ref<HTMLElement>()
const xtermContainer = ref<HTMLElement>()

// XTerm composable
const {
  isInitialized,
  isFocused: xtermFocused,
  currentTheme,
  write: writeToTerminal,
  clear: clearTerminal,
  resize: resizeTerminal,
  fit: fitTerminal,
  focus: focusTerminal,
  blur: blurTerminal,
  hasSelection,
  getSelection,
  selectAll,
  clearSelection,
  paste,
  dispose,
  updateTheme,
  toggleFullscreen: toggleXTermFullscreen
} = useXTerm({
  container: xtermContainer,
  onData: handleTerminalData,
  onTitleChange: handleTerminalTitleChange,
  onResize: handleTerminalResize,
  onFocus: handleTerminalFocus,
  onBlur: handleTerminalBlur,
  onSelectionChange: handleTerminalSelectionChange,
  config: {
    // Use font configuration if available, otherwise use props
    fontConfiguration: currentFontConfig.value,
    useSystemFontSettings: true,
    // Fallback to props for legacy support
    fontFamily: props.fontFamily,
    fontSize: props.fontSize,
    lineHeight: props.lineHeight,
    cursorStyle: props.cursorStyle,
    cursorBlink: props.cursorBlink,
    enableWebGL: true,
    enableLigatures: true,
    allowTransparency: props.opacity < 100,
    scrollback: 10000
  }
})

// State
const isFocused = computed(() => xtermFocused.value)
const isFullscreen = ref(false)
const isLoading = ref(false)
const error = ref<string | null>(null)
const isResizing = ref(false)
const showColorSchemeSelector = ref(false)
const showFontSelector = ref(false)
const showFontSettingsDialog = ref(false)

// Color scheme integration
const availableSchemes = ref<any[]>([])
const { schemes, selectScheme } = useColorScheme()

// Font configuration integration
const availableFontConfigs = computed(() => {
  return Object.values(configurations.value).sort((a, b) => {
    // Built-in configurations first, then alphabetical
    if (a.isBuiltIn && !b.isBuiltIn) return -1
    if (!a.isBuiltIn && b.isBuiltIn) return 1
    return a.name.localeCompare(b.name)
  })
})

// Context menu state
const showContextMenu = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })

// Computed
const session = computed(() => terminalTab.session)
const tab = computed(() => terminalTab.tab)
const outputLength = computed(() => terminalTab.output.length)
const hasContent = computed(() => outputLength.value > 0)
const hasActiveSelection = computed(() => isInitialized.value && hasSelection())
const hasOutput = computed(() => outputLength.value > 0)

const terminalStyle = computed(() => {
  const style: Record<string, string> = {
    fontFamily: props.fontFamily,
    fontSize: `${props.fontSize}px`,
    lineHeight: props.lineHeight.toString(),
    opacity: `${props.opacity / 100}`,
  }

  // Apply theme colors
  if (props.theme) {
    const themeColors = getThemeColors(props.theme)
    Object.assign(style, themeColors)
  }

  return style
})

// Terminal event handlers
const handleTerminalData = async (data: string) => {
  if (props.readonly) return

  try {
    await terminalTab.write(data)
  } catch (err) {
    console.error('Failed to write to terminal:', err)
    error.value = err instanceof Error ? err.message : 'Failed to write to terminal'
  }
}

const handleTerminalTitleChange = (title: string) => {
  if (tab.value && tab.value.title !== title) {
    terminalTab.setTitle(title)
    emit('titleChange', title)
  }
}

const handleTerminalResize = (cols: number, rows: number) => {
  emit('resize', { width: cols * 8, height: rows * 16 }) // Approximate character dimensions
}

const handleTerminalFocus = () => {
  setBellActive(true)
  emit('focus')
}

const handleTerminalBlur = () => {
  setBellActive(false)
  emit('blur')
}

const handleTerminalSelectionChange = () => {
  if (uiStore.settings.copyOnSelect && hasSelection()) {
    const selection = getSelection()
    if (selection) {
      copyToClipboard(selection)
    }
  }
}

// Methods
const initializeTerminal = async () => {
  if (!props.sessionId) return

  try {
    isLoading.value = true
    error.value = null

    // The useXTerm composable handles terminal initialization
    // We just need to wait for it to be ready
    await nextTick()

    // Focus if requested
    if (props.autofocus) {
      await nextTick()
      focusTerminal()
    }

    // Fit terminal to container
    await nextTick()
    fitTerminal()

    isLoading.value = false

    // Listen to output changes from the composable
    watch(() => terminalTab.output, (newOutput) => {
      if (isInitialized.value && newOutput.length > 0) {
        const lastOutput = newOutput[newOutput.length - 1]
        writeToTerminal(lastOutput)
      }
    }, { deep: false })

  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to initialize terminal'
    console.error('Terminal initialization error:', err)
    isLoading.value = false
  }
}

const handleContainerClick = () => {
  if (!isFocused.value) {
    focusTerminal()
  }
}

const handleContextMenu = (event: MouseEvent) => {
  event.preventDefault()

  showContextMenu.value = true
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
}

const copyContent = async () => {
  if (!isInitialized.value) return

  try {
    const selection = getSelection()
    if (!selection) {
      // If no selection, copy entire output buffer
      const output = terminalTab.output.join('')
      await clipboard.writeText(output, 'terminal')
    } else {
      await clipboard.writeText(selection, 'terminal')
    }
  } catch (err) {
    console.error('Failed to copy to clipboard:', err)
    error.value = 'Failed to copy to clipboard'
  }
}

const clearContent = () => {
  clearTerminal()
  terminalTab.clearOutput()
}

const toggleFullscreen = () => {
  toggleXTermFullscreen()
  isFullscreen.value = !isFullscreen.value
}

const startResize = (event: MouseEvent) => {
  if (isResizing.value || props.readonly) return

  isResizing.value = true

  const startX = event.clientX
  const startY = event.clientY
  const startHeight = terminalContent.value?.offsetHeight || 400
  const startWidth = terminalContent.value?.offsetWidth || 600

  const handleMouseMove = (e: MouseEvent) => {
    if (!isResizing.value) return

    const deltaX = e.clientX - startX
    const deltaY = e.clientY - startY

    const newWidth = Math.max(300, startWidth + deltaX)
    const newHeight = Math.max(200, startHeight + deltaY)

    if (terminalContent.value) {
      terminalContent.value.style.width = `${newWidth}px`
      terminalContent.value.style.height = `${newHeight}px`
    }

    // Fit terminal to new container size
    setTimeout(() => {
      fitTerminal()
    }, 10)

    emit('resize', { width: newWidth, height: newHeight })
  }

  const handleMouseUp = () => {
    isResizing.value = false
    document.removeEventListener('mousemove', handleMouseMove)
    document.removeEventListener('mouseup', handleMouseUp)
  }

  document.addEventListener('mousemove', handleMouseMove)
  document.addEventListener('mouseup', handleMouseUp)
}

const handleScroll = () => {
  // Auto-scroll to bottom when user scrolls to bottom
  if (terminalContent.value) {
    const { scrollTop, scrollHeight, clientHeight } = terminalContent.value
    const isAtBottom = scrollTop + clientHeight >= scrollHeight - 10

    if (isAtBottom && isInitialized.value) {
      // Terminal will auto-scroll with new output
      // The useXTerm composable handles scrolling automatically
    }
  }
}

const retryInitialization = () => {
  initializeTerminal()
}

// Context menu handlers
const handleContextMenuCopySelection = async () => {
  if (!isInitialized.value) return

  const success = await clipboard.copyTerminalSelection({ getSelection })
  if (success) {
    showContextMenu.value = false
  }
}

const handleContextMenuCopyAll = async () => {
  const success = await clipboard.copyTerminalOutput(terminalTab.output)
  if (success) {
    showContextMenu.value = false
  }
}

const handleContextMenuCopyOutput = async () => {
  // Copy only the output without prompts
  const outputLines = terminalTab.output.filter(line => !line.includes('$') && !line.includes('>'))
  const success = await clipboard.copyTerminalOutput(outputLines)
  if (success) {
    showContextMenu.value = false
  }
}

const handleContextMenuPaste = async () => {
  if (!isInitialized.value || props.readonly) return

  const success = await clipboard.pasteToTerminal({ paste })
  if (success) {
    const text = await clipboard.readText()
    if (text) {
      // Handle paste event if needed
    }
    showContextMenu.value = false
  }
}

const handleContextMenuClearSelection = () => {
  if (isInitialized.value) {
    clearSelection()
  }
  showContextMenu.value = false
}

const handleContextMenuClearAll = () => {
  clearTerminal()
  terminalTab.clearOutput()
  showContextMenu.value = false
}

const handleContextMenuFind = () => {
  // Implement find functionality
  showContextMenu.value = false
}

const handleContextMenuSaveAs = () => {
  // Implement save as functionality
  showContextMenu.value = false
}

const handleContextMenuPrint = () => {
  // Implement print functionality
  showContextMenu.value = false
}

const handleContextMenuFullscreen = () => {
  toggleFullscreen()
  showContextMenu.value = false
}

const handleContextMenuSettings = () => {
  // Implement settings functionality
  showContextMenu.value = false
}

const closeContextMenu = () => {
  showContextMenu.value = false
}

// Color scheme methods
const selectColorScheme = async (schemeId: string) => {
  try {
    selectScheme(schemeId)
    showColorSchemeSelector.value = false

    // Apply the color scheme to the terminal
    if (terminal.value && schemes.value[schemeId]) {
      const scheme = schemes.value[schemeId]
      updateTheme({
        colors: scheme.colors,
        font: scheme.font,
        cursor: scheme.cursor
      })
    }
  } catch (error) {
    console.error('Failed to select color scheme:', error)
  }
}

const loadAvailableSchemes = async () => {
  try {
    availableSchemes.value = Object.values(schemes.value)
  } catch (error) {
    console.error('Failed to load color schemes:', error)
  }
}

// Font configuration methods
const selectFontConfiguration = async (configId: string) => {
  try {
    await selectConfiguration(configId)
    showFontSelector.value = false

    // Apply the font configuration to the terminal
    if (terminal.value) {
      const fontConfig = currentFontConfig.value
      if (fontConfig) {
        updateTheme({
          colors: theme.value,
          font: {
            family: fontConfig.family,
            size: fontConfig.size,
            weight: fontConfig.weight,
            lineHeight: fontConfig.lineHeight,
            ligatures: fontConfig.ligatures
          },
          cursor: {
            style: props.cursorStyle,
            blink: props.cursorBlink
          }
        })
      }
    }
  } catch (error) {
    console.error('Failed to select font configuration:', error)
  }
}

const openFontSettings = () => {
  showFontSelector.value = false
  showFontSettingsDialog.value = true
}

const handleFontSettingsChanged = (newSettings: any) => {
  // Font settings changed, but not yet applied
}

const handleFontSettingsApplied = async (newSettings: any) => {
  try {
    await updateSettings(newSettings)
    showFontSettingsDialog.value = false
  } catch (error) {
    console.error('Failed to apply font settings:', error)
  }
}

// Terminal bell event handler
const handleTerminalBell = (event: CustomEvent) => {
  ringBell({
    visual: true,
    audible: true,
    title: 'Terminal Bell',
    message: 'Terminal activity detected',
    source: 'terminal'
  })
}

const getThemeColors = (themeName: string) => {
  // This would integrate with the theme system
  // For now, return default dark theme colors
  const defaultTheme = {
    background: '#1e1e1e',
    foreground: '#f0f0f0',
    cursor: '#ffffff',
    selection: 'rgba(255, 255, 255, 0.3)',
    black: '#000000',
    red: '#ff5555',
    green: '#50fa7b',
    yellow: '#f1fa8c',
    blue: '#bd93f9',
    magenta: '#ff79c6',
    cyan: '#8be9fd',
    white: '#f8f8f2',
    brightBlack: '#6272a4',
    brightRed: '#ff6e6e',
    brightGreen: '#69ff94',
    brightYellow: '#ffffa5',
    brightBlue: '#d6acff',
    brightMagenta: '#ff92df',
    brightCyan: '#a4ffff',
    brightWhite: '#ffffff'
  }

  return defaultTheme
}

const createTheme = (themeName: string) => {
  const colors = getThemeColors(themeName)

  return {
    background: colors.background,
    foreground: colors.foreground,
    cursor: colors.cursor,
    selection: colors.selection,
    black: colors.black,
    red: colors.red,
    green: colors.green,
    yellow: colors.yellow,
    blue: colors.blue,
    magenta: colors.magenta,
    cyan: colors.cyan,
    white: colors.white,
    brightBlack: colors.brightBlack,
    brightRed: colors.brightRed,
    brightGreen: colors.brightGreen,
    brightYellow: colors.brightYellow,
    brightBlue: colors.brightBlue,
    brightMagenta: colors.brightMagenta,
    brightCyan: colors.brightCyan,
    brightWhite: colors.brightWhite
  }
}

const formatTime = (timeString?: string) => {
  if (!timeString) return 'N/A'

  try {
    const date = new Date(timeString)
    return date.toLocaleTimeString()
  } catch {
    return 'N/A'
  }
}

// Click outside handler
const handleClickOutside = (event: MouseEvent) => {
  if (showContextMenu.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.terminal-context-menu')) {
      closeContextMenu()
    }
  }

  if (showColorSchemeSelector.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.color-scheme-selector-wrapper') && !target.closest('.color-scheme-dropdown')) {
      showColorSchemeSelector.value = false
    }
  }

  if (showFontSelector.value) {
    const target = event.target as HTMLElement
    if (!target.closest('.font-selector-wrapper') && !target.closest('.font-selector-dropdown')) {
      showFontSelector.value = false
    }
  }
}

// Lifecycle
onMounted(async () => {
  await nextTick()

  // Load available color schemes
  await loadAvailableSchemes()

  if (props.sessionId) {
    await initializeTerminal()
  }

  // Add click outside listener
  document.addEventListener('click', handleClickOutside)

  // Add terminal bell event listener
  document.addEventListener('terminal.bell', handleTerminalBell)
})

onUnmounted(() => {
  // Cleanup terminal - handled by useXTerm composable
  dispose()

  // Remove click outside listener
  document.removeEventListener('click', handleClickOutside)

  // Remove terminal bell event listener
  document.removeEventListener('terminal.bell', handleTerminalBell)

  // Clear bell notifications
  clearBellNotifications()
})

// Watch for session changes
watch(() => props.sessionId, async (newSessionId) => {
  if (newSessionId) {
    // Initialize new terminal
    await initializeTerminal()
  }
})

// Watch for theme changes
watch(() => props.theme, (newTheme) => {
  if (newTheme) {
    const theme = getThemeColors(newTheme)
    updateTheme({
      colors: theme,
      font: {
        family: props.fontFamily,
        size: props.fontSize,
        weight: 'normal',
        lineHeight: props.lineHeight,
        ligatures: true
      },
      cursor: {
        style: props.cursorStyle,
        blink: props.cursorBlink
      }
    })
  }
})

// Watch for font changes
watch(() => props.fontFamily, (newFontFamily) => {
  updateTheme({
    font: {
      family: newFontFamily,
      size: props.fontSize,
      weight: 'normal',
      lineHeight: props.lineHeight,
      ligatures: true
    }
  })
})

watch(() => props.fontSize, (newFontSize) => {
  updateTheme({
    font: {
      family: props.fontFamily,
      size: newFontSize,
      weight: 'normal',
      lineHeight: props.lineHeight,
      ligatures: true
    }
  })
})
</script>

<style scoped>
@reference "../../assets/css/main.css";
.terminal-panel {
  @apply relative flex flex-col bg-surface border border-border rounded-lg overflow-hidden;
  min-height: 200px;
  height: 100%;
  transition: all 0.2s ease;
}

.terminal-panel--focused {
  @apply border-primary;
  box-shadow: 0 0 0 1px theme('colors.primary');
}

.terminal-panel--inactive {
  @apply opacity-75;
}

.terminal-header {
  @apply flex items-center justify-between px-3 py-2 bg-surface border-b border-border;
  flex-shrink: 0;
}

.terminal-info {
  @apply flex flex-col min-w-0 flex-1;
}

.terminal-title {
  @apply text-sm font-medium text-text truncate;
}

.terminal-path {
  @apply text-xs text-text-secondary truncate;
}

.terminal-controls {
  @apply flex items-center gap-1;
}

.terminal-control {
  @apply p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-text transition-colors;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.terminal-control:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.terminal-control.active {
  @apply text-primary;
}

.terminal-bell-indicator {
  @apply relative flex items-center justify-center mr-2;
  color: var(--terminal-bell-color, #fbbf24);
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.terminal-bell-badge {
  @apply absolute -top-1 -right-1 bg-error text-white text-xs rounded-full h-4 w-4 flex items-center justify-center font-medium;
  min-width: 1rem;
  font-size: 0.75rem;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.terminal-content {
  @apply flex-1 overflow-auto relative;
  font-family: var(--font-family-mono);
  background-color: var(--terminal-bg);
  color: var(--terminal-fg);
}

.xterm-container {
  @apply h-full;
}

.terminal-footer {
  @apply flex items-center justify-between px-3 py-1 bg-surface border-t border-border;
  flex-shrink: 0;
}

.terminal-stats {
  @apply flex items-center gap-4 text-xs text-text-secondary;
}

.terminal-stat {
  @apply flex items-center gap-1;
}

.resize-handle {
  @apply p-1 cursor-ns-resize text-text-secondary hover:text-text;
}

.resize-handle:active {
  @apply text-primary;
}

.terminal-loading,
.terminal-error {
  @apply absolute inset-0 flex flex-col items-center justify-center bg-surface/90 backdrop-blur-sm;
}

.loading-spinner {
  @apply w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin;
}

.loading-text {
  @apply mt-2 text-sm text-text-secondary;
}

.error-message {
  @apply mt-2 text-sm text-error text-center max-w-xs;
}

.error-retry {
  @apply mt-3 px-3 py-1 bg-error text-white rounded text-sm flex items-center gap-1;
  border: none;
  background: none;
  cursor: pointer;
}

/* Color Scheme Selector Styles */
.color-scheme-selector-wrapper {
  @apply relative;
}

.color-scheme-dropdown {
  @apply absolute top-full right-0 mt-2 w-80 bg-surface border border-border rounded-lg shadow-lg z-50;
  max-height: 400px;
  overflow: hidden;
}

.dropdown-header {
  @apply flex items-center justify-between p-3 border-b border-border;
}

.dropdown-header h4 {
  @apply text-sm font-medium text-text m-0;
}

.dropdown-close {
  @apply p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-text transition-colors;
  border: none;
  background: none;
  cursor: pointer;
}

.schemes-list {
  @apply overflow-y-auto max-h-80;
}

.scheme-option {
  @apply w-full p-3 flex items-center gap-3 hover:bg-surface-hover transition-colors text-left border-0 bg-transparent cursor-pointer;
}

.scheme-option.active {
  @apply bg-primary bg-opacity-10 text-primary;
}

.scheme-preview-colors {
  @apply flex gap-1 flex-shrink-0;
}

.color-dot {
  @apply w-4 h-4 rounded-full border border-border;
}

.scheme-info {
  @apply flex-1 min-w-0;
}

.scheme-name {
  @apply block text-sm font-medium text-text truncate;
}

.scheme-author {
  @apply block text-xs text-text-secondary truncate;
}

.scheme-selected-icon {
  @apply text-primary flex-shrink-0;
}

/* Font Selector Styles */
.font-selector-wrapper {
  @apply relative;
}

.font-selector-dropdown {
  @apply absolute top-full right-0 mt-2 w-72 bg-surface border border-border rounded-lg shadow-lg z-50;
  max-height: 400px;
  overflow: hidden;
}

.font-options-list {
  @apply overflow-y-auto max-h-64;
}

.font-option {
  @apply w-full p-3 flex items-center gap-3 hover:bg-surface-hover transition-colors text-left border-0 bg-transparent cursor-pointer;
}

.font-option.active {
  @apply bg-primary bg-opacity-10 text-primary;
}

.font-option-info {
  @apply flex-1 min-w-0;
}

.font-option-name {
  @apply block text-sm font-medium text-text truncate;
}

.font-option-family {
  @apply block text-xs text-text-secondary truncate;
}

.font-option-selected {
  @apply text-primary flex-shrink-0;
}

.font-selector-footer {
  @apply p-3 border-t border-border;
}

.font-selector-action {
  @apply w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-text hover:bg-surface-hover rounded-md transition-colors;
}

/* Font Settings Dialog Styles */
.font-settings-dialog-overlay {
  @apply fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50;
}

.font-settings-dialog {
  @apply bg-surface border border-border rounded-lg shadow-xl max-w-4xl max-h-[80vh] w-full mx-4 flex flex-col;
}

.font-settings-dialog-header {
  @apply flex items-center justify-between p-4 border-b border-border;
}

.font-settings-dialog-header h3 {
  @apply m-0;
}

.font-settings-dialog-content {
  @apply flex-1 overflow-y-auto p-0;
}

/* Fullscreen styles */
.terminal-panel:fullscreen {
  @apply fixed inset-0 z-50 rounded-none;
}
</style>