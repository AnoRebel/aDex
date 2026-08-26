import { ref, computed, onMounted, onUnmounted, nextTick, watch, readonly } from 'vue'
import { Terminal } from '@xterm/xterm'
import { WebglAddon } from '@xterm/addon-webgl'
import { FitAddon } from '@xterm/addon-fit'
import { CanvasAddon } from '@xterm/addon-canvas'
import { LigaturesAddon } from '@xterm/addon-ligatures'
import type { TerminalTheme, TerminalSettings } from '~/types/terminal'
import { useColorScheme } from './useColorScheme'
import { useFontConfiguration } from './useFontConfiguration'
import type { FontConfiguration } from '~/types/font'

export interface XTermConfig {
  theme?: Partial<TerminalTheme>
  settings?: Partial<TerminalSettings>
  fontFamily?: string
  fontSize?: number
  lineHeight?: number
  letterSpacing?: number
  cursorStyle?: 'block' | 'underline' | 'bar'
  cursorBlink?: boolean
  scrollback?: number
  enableWebGL?: boolean
  enableCanvas?: boolean
  enableLigatures?: boolean
  cols?: number
  rows?: number
  allowTransparency?: boolean
  convertEol?: boolean
  termName?: string
  colorSchemeId?: string
  fontConfiguration?: FontConfiguration
  useSystemFontSettings?: boolean
}

export interface XTermOptions {
  container: HTMLElement | string
  onData?: (data: string) => void
  onTitleChange?: (title: string) => void
  onResize?: (cols: number, rows: number) => void
  onFocus?: () => void
  onBlur?: () => void
  onSelectionChange?: () => void
  onKey?: (event: { key: string; domEvent: KeyboardEvent }) => void
  config?: XTermConfig
}

export function useXTerm(options: XTermOptions) {
  const {
    container,
    onData,
    onTitleChange,
    onResize,
    onFocus,
    onBlur,
    onSelectionChange,
    onKey,
    config = {}
  } = options

  // Color scheme integration
  const { getXtermTheme, currentScheme } = useColorScheme()

  // Font configuration integration
  const { currentConfiguration: currentFontConfig, settings: fontSettings } = useFontConfiguration()

  // Refs
  const terminalContainer = ref<HTMLElement>()
  const terminal = ref<Terminal>()
  const fitAddon = ref<FitAddon>()
  const webglAddon = ref<WebglAddon>()
  const canvasAddon = ref<CanvasAddon>()
  const ligaturesAddon = ref<LigaturesAddon>()

  const isInitialized = ref(false)
  const isFocused = ref(false)
  const isResizing = ref(false)

  const currentTheme = ref<Partial<TerminalTheme>>(config.theme || {})
  const currentSettings = ref<Partial<TerminalSettings>>(config.settings || {})

  // Computed
  const theme = computed(() => {
    // Get color scheme theme if available
    const colorSchemeTheme = currentScheme.value ? getXtermTheme(currentScheme.value) : null

    // Use color scheme theme or fallback to default
    if (colorSchemeTheme && Object.keys(colorSchemeTheme).length > 0) {
      return colorSchemeTheme
    }

    // Fallback default colors
    const defaultColors = {
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

    return defaultColors
  })

  // Computed font configuration
  const effectiveFontConfig = computed(() => {
    // Priority: explicit config > provided font configuration > system font settings > defaults
    if (config.fontConfiguration) {
      return config.fontConfiguration
    }

    if (config.useSystemFontSettings && currentFontConfig.value) {
      return currentFontConfig.value
    }

    if (config.fontFamily) {
      // Create a basic config from legacy props
      return {
        id: 'legacy',
        name: 'Legacy Font',
        family: config.fontFamily,
        size: config.fontSize ?? 14,
        weight: '400',
        lineHeight: config.lineHeight ?? 1.5,
        letterSpacing: config.letterSpacing ?? 0,
        ligatures: config.enableLigatures ?? true,
        antialias: true,
        hinting: 'slight',
        features: {},
        isBuiltIn: false,
        isDefault: false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      } as FontConfiguration
    }

    return null
  })

  const terminalOptions = computed(() => {
    const fontConfig = effectiveFontConfig.value
    const settings = fontSettings.value

    // Get font settings from font configuration or fallback
    const fontFamily = fontConfig?.family ||
                     config.fontFamily ||
                     currentSettings.value.fontFamily ||
                     'JetBrains Mono, Fira Code, Consolas, monospace'

    const fontSize = fontConfig?.size ||
                    config.fontSize ||
                    currentSettings.value.fontSize ||
                    14

    const fontWeight = fontConfig?.weight ||
                      currentTheme.value?.font?.weight ||
                      'normal'

    const letterSpacing = fontConfig?.letterSpacing ||
                         config.letterSpacing ||
                         currentSettings.value.letterSpacing ||
                         0

    const lineHeight = fontConfig?.lineHeight ||
                      config.lineHeight ||
                      currentSettings.value.lineHeight ||
                      1.5

    const enableLigatures = fontConfig?.ligatures !== undefined ?
                           fontConfig.ligatures :
                           (config.enableLigatures ?? true)

    // Build font feature settings string
    const fontFeatures = fontConfig?.features || {}
    const fontFeatureSettings = Object.entries(fontFeatures)
      .filter(([_, enabled]) => enabled)
      .map(([feature, _]) => `"${feature}"`)
      .join(', ') || 'normal'

    return {
      allowTransparency: config.allowTransparency ?? true,
      cursorBlink: config.cursorBlink ?? currentSettings.value.cursorBlink ?? false,
      cursorStyle: config.cursorStyle ?? currentSettings.value.cursorStyle ?? 'block',
      cursorWidth: currentTheme.value?.cursor?.width,
      fontFamily,
      fontSize,
      fontWeight,
      fontWeightBold: 'bold',
      letterSpacing,
      lineHeight,
      scrollback: config.scrollback ?? currentSettings.value.scrollbackSize ?? 10000,
      theme: theme.value,
      cols: config.cols,
      rows: config.rows,
      convertEol: config.convertEol ?? currentSettings.value.convertEol ?? false,
      termName: config.termName ?? 'xterm-256color',
      fastScrollModifier: currentSettings.value.fastScrollModifier || 'alt',
      wordSeparator: currentSettings.value.wordSeparator || ' ()[]{}',
      altClickMovesCursor: true,
      rightClickSelectsWord: true,
      rendererType: 'dom' as const, // Will be overridden by addons
      screenReaderMode: currentSettings.value.screenReaderMode ?? false,
      macOptionIsMeta: currentSettings.value.altGrIsMeta ?? false,
      macOptionClickForcesSelection: false,
      windowsPty: {
        backend: 'conpty',
        buildNumber: 22621
      },
      fontFeatureSettings: fontFeatureSettings
    }
  })

  // Methods
  const initialize = async (): Promise<void> => {
    try {
      // Get container element
      const containerElement = typeof container === 'string'
        ? document.getElementById(container)
        : container

      if (!containerElement) {
        throw new Error('Terminal container not found')
      }

      terminalContainer.value = containerElement

      // Create terminal instance
      terminal.value = new Terminal(terminalOptions.value)

      // Create and load addons
      await loadAddons()

      // Open terminal in container
      terminal.value.open(containerElement)

      // Setup event handlers
      setupEventHandlers()

      // Focus terminal if configured
      if (currentSettings.value.autoFocus) {
        await nextTick()
        terminal.value?.focus()
      }

      // Fit terminal to container
      if (fitAddon.value) {
        await nextTick()
        fitAddon.value.fit()
      }

      isInitialized.value = true

    } catch (error) {
      console.error('Failed to initialize xterm:', error)
      throw error
    }
  }

  const loadAddons = async (): Promise<void> => {
    if (!terminal.value) return

    try {
      // Load FitAddon (always required)
      fitAddon.value = new FitAddon()
      terminal.value.loadAddon(fitAddon.value)

      // Load WebGL addon if enabled and supported
      if (config.enableWebGL !== false) {
        try {
          webglAddon.value = new WebglAddon()
          terminal.value.loadAddon(webglAddon.value)
          console.debug('WebGL renderer enabled')
        } catch (error) {
          console.warn('WebGL renderer not supported, falling back to Canvas:', error)
          // Fallback to Canvas renderer
          loadCanvasAddon()
        }
      } else if (config.enableCanvas) {
        loadCanvasAddon()
      }

      // Load Ligatures addon if enabled
      if (config.enableLigatures && currentSettings.value.enableLigatures) {
        try {
          ligaturesAddon.value = new LigaturesAddon({})
          terminal.value.loadAddon(ligaturesAddon.value)
          console.debug('Ligatures enabled')
        } catch (error) {
          console.warn('Failed to enable ligatures:', error)
        }
      }

    } catch (error) {
      console.error('Failed to load terminal addons:', error)
      throw error
    }
  }

  const loadCanvasAddon = (): void => {
    if (!terminal.value) return

    try {
      canvasAddon.value = new CanvasAddon()
      terminal.value.loadAddon(canvasAddon.value)
      console.debug('Canvas renderer enabled')
    } catch (error) {
      console.warn('Canvas renderer not supported:', error)
    }
  }

  const setupEventHandlers = (): void => {
    if (!terminal.value) return

    // Data input handler
    if (onData) {
      terminal.value.onData(onData)
    }

    // Title change handler
    if (onTitleChange) {
      terminal.value.onTitleChange(onTitleChange)
    }

    // Resize handler
    if (onResize && fitAddon.value) {
      const handleResize = () => {
        if (fitAddon.value && terminal.value && !isResizing.value) {
          const { cols, rows } = fitAddon.value
          onResize(cols, rows)
        }
      }

      // Setup resize observer
      const resizeObserver = new ResizeObserver(() => {
        if (!isResizing.value) {
          handleResize()
        }
      })

      resizeObserver.observe(terminalContainer.value!)
    }

    // Focus handlers
    terminal.value.onFocus(() => {
      isFocused.value = true
      onFocus?.()
    })

    terminal.value.onBlur(() => {
      isFocused.value = false
      onBlur?.()
    })

    // Selection handler
    if (onSelectionChange) {
      terminal.value.onSelectionChange(onSelectionChange)
    }

    // Key handler
    if (onKey) {
      terminal.value.onKey(onKey)
    }

    // Bell handler
    terminal.value.onBell(() => {
      // Handle bell event
      document.dispatchEvent(new CustomEvent('terminal.bell', {
        detail: { sessionId: 'current' }
      }))
    })
  }

  const write = (data: string): void => {
    terminal.value?.write(data)
  }

  const writeln = (data: string): void => {
    terminal.value?.writeln(data)
  }

  const clear = (): void => {
    terminal.value?.clear()
  }

  const reset = (): void => {
    terminal.value?.reset()
  }

  const focus = (): void => {
    terminal.value?.focus()
  }

  const blur = (): void => {
    terminal.value?.blur()
  }

  const hasSelection = (): boolean => {
    return terminal.value?.hasSelection() ?? false
  }

  const getSelection = (): string => {
    return terminal.value?.getSelection() ?? ''
  }

  const selectAll = (): void => {
    terminal.value?.selectAll()
  }

  const clearSelection = (): void => {
    terminal.value?.clearSelection()
  }

  const resize = (cols: number, rows: number): void => {
    if (terminal.value && cols > 0 && rows > 0) {
      isResizing.value = true
      try {
        terminal.value.resize(cols, rows)
      } catch (error) {
        console.warn('Failed to resize terminal:', error)
      } finally {
        isResizing.value = false
      }
    }
  }

  const fit = (): void => {
    if (fitAddon.value) {
      fitAddon.value.fit()
    }
  }

  const scrollToTop = (): void => {
    terminal.value?.scrollToTop()
  }

  const scrollToBottom = (): void => {
    terminal.value?.scrollToBottom()
  }

  const scrollToLine = (line: number): void => {
    terminal.value?.scrollToLine(line)
  }

  const paste = (text: string): void => {
    terminal.value?.paste(text)
  }

  const dispose = (): void => {
    if (terminal.value) {
      terminal.value.dispose()
      terminal.value = undefined
    }

    // Clean up addons
    webglAddon.value = undefined
    canvasAddon.value = undefined
    fitAddon.value = undefined
    ligaturesAddon.value = undefined

    isInitialized.value = false
    isFocused.value = false
  }

  const updateTheme = (theme: Partial<TerminalTheme>): void => {
    currentTheme.value = theme
    if (terminal.value) {
      terminal.value.options.theme = theme.value
    }
  }

  const applyColorScheme = (schemeId: string): void => {
    // This will be handled by the color scheme composable automatically
    // through the computed theme property
    console.log('Applying color scheme:', schemeId)
  }

  const updateFontConfiguration = (fontConfig: FontConfiguration): void => {
    if (terminal.value) {
      const options = {
        fontFamily: `'${fontConfig.family}', monospace`,
        fontSize: fontConfig.size,
        fontWeight: fontConfig.weight,
        letterSpacing: fontConfig.letterSpacing,
        lineHeight: fontConfig.lineHeight,
        fontFeatureSettings: Object.entries(fontConfig.features || {})
          .filter(([_, enabled]) => enabled)
          .map(([feature, _]) => `"${feature}"`)
          .join(', ') || 'normal'
      }

      Object.assign(terminal.value.options, options)
    }
  }

  const applyFontConfiguration = (fontConfig: FontConfiguration | null): void => {
    if (fontConfig) {
      updateFontConfiguration(fontConfig)
    }
  }

  const getCurrentFontConfiguration = (): FontConfiguration | null => {
    return effectiveFontConfig.value
  }

  const updateSettings = (settings: Partial<TerminalSettings>): void => {
    currentSettings.value = { ...currentSettings.value, ...settings }

    if (terminal.value) {
      const options: Partial<TerminalOptions> = {}

      if (settings.cursorBlink !== undefined) {
        options.cursorBlink = settings.cursorBlink
      }
      if (settings.cursorStyle !== undefined) {
        options.cursorStyle = settings.cursorStyle
      }
      if (settings.fontFamily !== undefined) {
        options.fontFamily = settings.fontFamily
      }
      if (settings.fontSize !== undefined) {
        options.fontSize = settings.fontSize
      }
      if (settings.lineHeight !== undefined) {
        options.lineHeight = settings.lineHeight
      }
      if (settings.letterSpacing !== undefined) {
        options.letterSpacing = settings.letterSpacing
      }
      if (settings.scrollbackSize !== undefined) {
        options.scrollback = settings.scrollbackSize
      }
      if (settings.convertEol !== undefined) {
        options.convertEol = settings.convertEol
      }
      if (settings.screenReaderMode !== undefined) {
        options.screenReaderMode = settings.screenReaderMode
      }
      if (settings.altGrIsMeta !== undefined) {
        options.macOptionIsMeta = settings.altGrIsMeta
      }

      Object.assign(terminal.value.options, options)
    }
  }

  const toggleFullscreen = (): void => {
    if (!terminalContainer.value) return

    if (!document.fullscreenElement) {
      terminalContainer.value.requestFullscreen()
    } else {
      document.exitFullscreen()
    }
  }

  const getTerminal = (): Terminal | undefined => {
    return terminal.value
  }

  const getFitAddon = (): FitAddon | undefined => {
    return fitAddon.value
  }

  const getWebglAddon = (): WebglAddon | undefined => {
    return webglAddon.value
  }

  const getCanvasAddon = (): CanvasAddon | undefined => {
    return canvasAddon.value
  }

  const getLigaturesAddon = (): LigaturesAddon | undefined => {
    return ligaturesAddon.value
  }

  // Watchers
  watch(
    () => currentTheme.value,
    (newTheme) => {
      if (terminal.value) {
        terminal.value.options.theme = theme.value
      }
    },
    { deep: true }
  )

  watch(
    () => currentSettings.value,
    () => {
      updateSettings(currentSettings.value)
    },
    { deep: true }
  )

  // Watch font configuration changes
  watch(
    effectiveFontConfig,
    (newFontConfig) => {
      if (terminal.value && newFontConfig) {
        updateFontConfiguration(newFontConfig)
      }
    },
    { deep: true }
  )

  // Watch current font configuration from composable
  watch(
    currentFontConfig,
    (newFontConfig) => {
      if (terminal.value && newFontConfig && config.useSystemFontSettings) {
        updateFontConfiguration(newFontConfig)
      }
    },
    { deep: true }
  )

  // Initialize immediately instead of waiting for mounted
  // This allows the composable to be used programmatically
  const initPromise = nextTick().then(initialize)

  // Lifecycle
  onMounted(async () => {
    await initPromise
  })

  onUnmounted(() => {
    dispose()
  })

  return {
    // State
    isInitialized,
    isFocused,
    isResizing,
    currentTheme,
    currentSettings,

    // Terminal instance
    terminal: readonly(terminal),
    terminalContainer: readonly(terminalContainer),

    // Addons
    fitAddon: readonly(fitAddon),
    webglAddon: readonly(webglAddon),
    canvasAddon: readonly(canvasAddon),
    ligaturesAddon: readonly(ligaturesAddon),

    // Methods
    initialize,
    write,
    writeln,
    clear,
    reset,
    focus,
    blur,
    hasSelection,
    getSelection,
    selectAll,
    clearSelection,
    resize,
    fit,
    scrollToTop,
    scrollToBottom,
    scrollToLine,
    paste,
    dispose,
    updateTheme,
    updateSettings,
    toggleFullscreen,
    applyColorScheme,
    updateFontConfiguration,
    applyFontConfiguration,

    // Getters
    getTerminal,
    getFitAddon,
    getWebglAddon,
    getCanvasAddon,
    getLigaturesAddon,
    getCurrentFontConfiguration,
    effectiveFontConfig: readonly(effectiveFontConfig)
  }
}

export default useXTerm