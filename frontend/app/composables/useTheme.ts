import { ref, computed, watch, onMounted, onUnmounted, readonly } from 'vue'
import { useWails } from './useWails'

// Type definitions for theme system (temporary until types/theme.ts is created)
export interface ThemeColors {
  background: {
    primary: string
    secondary: string
    tertiary: string
  }
  foreground: {
    primary: string
    secondary: string
    tertiary: string
  }
  accent: {
    primary: string
    secondary: string
  }
  status: {
    success: string
    warning: string
    error: string
    info: string
  }
  terminal: string[]
  ui: {
    buttonBackground: string
    buttonForeground: string
    buttonHover: string
    buttonActive: string
    inputBackground: string
    inputForeground: string
    inputBorder: string
    inputFocus: string
    border: string
    shadow: string
  }
}

export interface ThemeFonts {
  families: {
    primary: string
    secondary: string
    monospace: string
  }
  sizes: {
    extraSmall: string
    small: string
    base: string
    large: string
    extraLarge: string
    doubleExtraLarge: string
  }
  weights: {
    light: string
    normal: string
    medium: string
    semiBold: string
    bold: string
  }
  lineHeights: {
    tight: string
    normal: string
    relaxed: string
  }
  letterSpacing: {
    tight: string
    normal: string
    wide: string
  }
}

export interface ThemeEffects {
  borderRadius: {
    small: string
    medium: string
    large: string
    extraLarge: string
    full: string
  }
  shadows: string[]
  gradients: string[]
  animations: {
    duration: {
      fast: string
      normal: string
      slow: string
    }
    easing: {
      linear: string
      easeIn: string
      easeOut: string
      easeInOut: string
    }
  }
  blur: {
    small: string
    medium: string
    large: string
    extraLarge: string
  }
}

export interface ThemeSettings {
  disabledOpacity: number
  hoverOpacity: number
  activeOpacity: number
  transitionDuration: {
    fast: string
    normal: string
    slow: string
  }
  zIndexScale: Record<string, number>
}

export interface Theme {
  id: string
  name: string
  displayName?: string
  description: string
  author: string
  version: string
  colors: ThemeColors
  fonts: ThemeFonts
  effects: ThemeEffects
  settings: ThemeSettings
  customProperties?: Record<string, any>
  metadata?: {
    createdAt?: string
    updatedAt?: string
    tags?: string[]
    category?: string
  }
  isDark?: boolean
  isCustom?: boolean
}

export interface CustomTheme {
  name: string
  theme: Theme
  createdAt: string
}

export interface GeneratedCSS {
  variables: string
  themeId: string
  generatedAt: string
  hash: string
  variableCount: number
  metadata: Record<string, string>
}

export interface ThemeServiceConfig {
  themeDirectory?: string
  defaultTheme?: string
  autoSave?: boolean
  autoReload?: boolean
  cacheEnabled?: boolean
  variablePrefix?: string
  minifyCSS?: boolean
  watchInterval?: number
  maxCacheSize?: number
  enableLegacyImport?: boolean
}

// Global state
const isInitialized = ref(false)
const isLoading = ref(false)
const currentTheme = ref<Theme | null>(null)
const availableThemes = ref<Theme[]>([])
const cssVariables = ref<string>('')
const error = ref<string | null>(null)
const config = ref<ThemeServiceConfig>({
  themeDirectory: 'themes',
  defaultTheme: 'default-dark',
  autoSave: true,
  autoReload: true,
  cacheEnabled: true,
  variablePrefix: '--dex-theme',
  minifyCSS: false,
  watchInterval: 5000,
  maxCacheSize: 100,
  enableLegacyImport: true
})

// Legacy state for backward compatibility
const customThemes = ref<CustomTheme[]>([])
const themeTransition = ref('fade') // 'fade', 'slide', 'zoom'
const autoSwitch = ref(false)
const switchSchedule = ref({
  from: '20:00', // 8 PM
  to: '08:00',   // 8 AM
})

// Event listeners
const eventListeners = new Map<string, (theme?: Theme) => void>()

/**
 * useTheme composable provides theme management functionality
 */
export function useTheme() {
  const wails = useWails()

  // Computed properties
  const isDark = computed(() => {
    if (!currentTheme.value) return true
    if (currentTheme.value.isDark !== undefined) {
      return currentTheme.value.isDark
    }
    // Calculate from background color
    const bgColor = currentTheme.value.colors.background.primary
    const rgb = hexToRgb(bgColor)
    if (!rgb) return true
    const luminance = (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) / 255
    return luminance < 0.5
  })

  const isLightTheme = computed(() => !isDark.value)

  const primaryColor = computed(() => currentTheme.value?.colors.accent.primary || '#00ff41')
  const backgroundColor = computed(() => currentTheme.value?.colors.background.primary || '#0a0a0a')
  const foregroundColor = computed(() => currentTheme.value?.colors.foreground.primary || '#ffffff')

  const themeCategories = computed(() => {
    const categories = new Set<string>()
    availableThemes.value.forEach(theme => {
      if (theme.metadata?.category) {
        categories.add(theme.metadata.category)
      }
    })
    return Array.from(categories)
  })

  const themesByCategory = computed(() => {
    const grouped: Record<string, Theme[]> = {}
    availableThemes.value.forEach(theme => {
      const category = theme.metadata?.category || 'uncategorized'
      if (!grouped[category]) {
        grouped[category] = []
      }
      grouped[category].push(theme)
    })
    return grouped
  })

  // Legacy computed properties for backward compatibility
  const currentThemeData = computed(() => {
    const customTheme = customThemes.value.find(t => t.name === currentTheme.value?.id)
    return customTheme?.theme || currentTheme.value || createDefaultTheme()
  })

  const allAvailableThemes = computed(() => {
    const themes = [...availableThemes.value]
    const custom = customThemes.value.map(ct => ({
      ...ct.theme,
      isCustom: true,
    }))
    return [...themes, ...custom]
  })

  const darkThemes = computed(() =>
    allAvailableThemes.value.filter(theme => theme.isDark !== false)
  )

  const lightThemes = computed(() =>
    allAvailableThemes.value.filter(theme => theme.isDark === false)
  )

  // Helper function to create default theme
  const createDefaultTheme = (): Theme => ({
    id: 'cyberpunk',
    name: 'cyberpunk',
    displayName: 'Cyberpunk',
    description: 'Neon green on dark background',
    author: 'aDex-UI Team',
    version: '1.0.0',
    isDark: true,
    colors: {
      background: {
        primary: '#0a0a0a',
        secondary: '#1a1a1a',
        tertiary: '#2a2a2a'
      },
      foreground: {
        primary: '#00ff00',
        secondary: '#00cc00',
        tertiary: '#009900'
      },
      accent: {
        primary: '#ff00ff',
        secondary: '#cc00cc'
      },
      status: {
        success: '#00ff00',
        warning: '#ffaa00',
        error: '#ff0066',
        info: '#00ffff'
      },
      terminal: [
        '#000000', '#ff0000', '#00ff00', '#ffff00',
        '#0000ff', '#ff00ff', '#00ffff', '#ffffff',
        '#808080', '#ff8080', '#80ff80', '#ffff80',
        '#8080ff', '#ff80ff', '#80ffff', '#c0c0c0'
      ],
      ui: {
        buttonBackground: '#1a1a1a',
        buttonForeground: '#00ff00',
        buttonHover: '#2a2a2a',
        buttonActive: '#ff00ff',
        inputBackground: '#0a0a0a',
        inputForeground: '#00ff00',
        inputBorder: '#00ff00',
        inputFocus: '#ff00ff',
        border: '#333333',
        shadow: 'rgba(0, 255, 0, 0.2)'
      }
    },
    fonts: getDefaultFonts(),
    effects: getDefaultEffects(),
    settings: getDefaultSettings(),
    metadata: {
      category: 'built-in',
      tags: ['cyberpunk', 'neon', 'dark']
    }
  })

  // Default values
  const getDefaultFonts = (): ThemeFonts => ({
    families: {
      primary: '"JetBrains Mono", "Fira Code", monospace',
      secondary: '"Inter", system-ui, sans-serif',
      monospace: '"JetBrains Mono", "Fira Code", monospace'
    },
    sizes: {
      extraSmall: '0.75rem',
      small: '0.875rem',
      base: '1rem',
      large: '1.125rem',
      extraLarge: '1.25rem',
      doubleExtraLarge: '1.5rem'
    },
    weights: {
      light: '300',
      normal: '400',
      medium: '500',
      semiBold: '600',
      bold: '700'
    },
    lineHeights: {
      tight: '1.25',
      normal: '1.5',
      relaxed: '1.75'
    },
    letterSpacing: {
      tight: '-0.025em',
      normal: '0',
      wide: '0.025em'
    }
  })

  const getDefaultEffects = (): ThemeEffects => ({
    borderRadius: {
      small: '0.25rem',
      medium: '0.5rem',
      large: '1rem',
      extraLarge: '1.5rem',
      full: '9999px'
    },
    shadows: [
      '0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.24)',
      '0 4px 6px rgba(0, 0, 0, 0.16), 0 2px 4px rgba(0, 0, 0, 0.12)'
    ],
    gradients: [
      'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)'
    ],
    animations: {
      duration: {
        fast: '150ms',
        normal: '300ms',
        slow: '500ms'
      },
      easing: {
        linear: 'linear',
        easeIn: 'cubic-bezier(0.4, 0, 1, 1)',
        easeOut: 'cubic-bezier(0, 0, 0.2, 1)',
        easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)'
      }
    },
    blur: {
      small: '4px',
      medium: '8px',
      large: '16px',
      extraLarge: '24px'
    }
  })

  const getDefaultSettings = (): ThemeSettings => ({
    disabledOpacity: 0.5,
    hoverOpacity: 0.8,
    activeOpacity: 1,
    transitionDuration: {
      fast: '150ms',
      normal: '300ms',
      slow: '500ms'
    },
    zIndexScale: {
      dropdown: 1000,
      sticky: 1020,
      fixed: 1030,
      modalBackdrop: 1040,
      modal: 1050,
      popover: 1060,
      tooltip: 1070,
      toast: 1080
    }
  })

  // Predefined themes (backward compatibility)
  const predefinedThemes: Record<string, Theme> = {
    cyberpunk: createDefaultTheme(),
  }

  // Methods
  const initialize = async (): Promise<void> => {
    if (isInitialized.value) return

    try {
      isLoading.value = true
      error.value = null

      // Initialize theme service through Wails
      if (wails.isReady.value) {
        try {
          await wails.callService('theme', 'initialize', config.value)
          // Load available themes
          await loadAvailableThemes()
          // Get current theme from backend
          const currentThemeId = await wails.callService('theme', 'getCurrentThemeId')
          if (currentThemeId) {
            await setCurrentTheme(currentThemeId)
          } else {
            await setCurrentTheme(config.value.defaultTheme!)
          }
        } catch (backendError) {
          console.warn('Backend theme service not available, using fallback:', backendError)
          // Fallback to legacy behavior
          await initializeTheme()
        }
      } else {
        // Fallback to legacy behavior
        await initializeTheme()
      }

      isInitialized.value = true
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error during theme initialization'
      error.value = errorMessage
      console.error('Theme initialization failed:', err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const loadAvailableThemes = async (): Promise<void> => {
    try {
      const themes = await wails.callService<Theme[]>('theme', 'getAllThemes')
      availableThemes.value = themes || []
    } catch (err) {
      console.error('Failed to load available themes:', err)
      // Fallback to predefined themes
      availableThemes.value = Object.values(predefinedThemes)
    }
  }

  const setCurrentTheme = async (themeId: string): Promise<void> => {
    if (!themeId) {
      throw new Error('Theme ID cannot be empty')
    }

    try {
      isLoading.value = true

      // Set theme through Wails if available
      if (wails.isReady.value) {
        try {
          await wails.callService('theme', 'setCurrentTheme', themeId)
          // Get theme details from backend
          const theme = await wails.callService<Theme>('theme', 'getTheme', themeId)
          if (theme) {
            currentTheme.value = theme
            await generateCSSVariables(theme)
            applyThemeToDOM()
            triggerEvent('theme:changed', theme)
            return
          }
        } catch (backendError) {
          console.warn('Backend theme setting failed, using fallback:', backendError)
        }
      }

      // Fallback: find theme in available themes or predefined themes
      const theme = availableThemes.value.find(t => t.id === themeId) ||
                   predefinedThemes[themeId] ||
                   createDefaultTheme()

      currentTheme.value = theme
      await generateCSSVariables(theme)
      applyThemeToDOM()
      triggerEvent('theme:changed', theme)

      // Save to localStorage
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('adex-theme', themeId)
      }

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to set theme'
      error.value = errorMessage
      console.error('Failed to set theme:', err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const generateCSSVariables = async (theme: Theme): Promise<void> => {
    try {
      if (wails.isReady.value) {
        const css = await wails.callService<GeneratedCSS>('theme', 'generateCSSVariables', theme.id)
        if (css) {
          cssVariables.value = css.variables
          return
        }
      }
      // Fallback: generate CSS locally
      cssVariables.value = generateFallbackCSS(theme)
    } catch (err) {
      console.error('Failed to generate CSS variables:', err)
      // Generate fallback CSS
      cssVariables.value = generateFallbackCSS(theme)
    }
  }

  const generateFallbackCSS = (theme: Theme): string => {
    const prefix = config.value.variablePrefix!
    let css = `:root {\n`

    // Add color variables
    css += `  ${prefix}-background-primary: ${theme.colors.background.primary};\n`
    css += `  ${prefix}-background-secondary: ${theme.colors.background.secondary};\n`
    css += `  ${prefix}-foreground-primary: ${theme.colors.foreground.primary};\n`
    css += `  ${prefix}-foreground-secondary: ${theme.colors.foreground.secondary};\n`
    css += `  ${prefix}-accent-primary: ${theme.colors.accent.primary};\n`
    css += `  ${prefix}-accent-secondary: ${theme.colors.accent.secondary};\n`

    // Add font variables
    css += `  ${prefix}-font-family-primary: ${theme.fonts.families.primary};\n`
    css += `  ${prefix}-font-family-mono: ${theme.fonts.families.monospace};\n`

    css += `}\n`
    return css
  }

  const applyThemeToDOM = (): void => {
    if (typeof document === 'undefined') return

    // Remove existing theme style tag
    const existingStyle = document.getElementById('dex-theme-variables')
    if (existingStyle) {
      existingStyle.remove()
    }

    // Create new style tag
    const style = document.createElement('style')
    style.id = 'dex-theme-variables'
    style.textContent = cssVariables.value
    document.head.appendChild(style)

    // Update body class
    document.body.classList.remove('theme-dark', 'theme-light')
    document.body.classList.add(isDark.value ? 'theme-dark' : 'theme-light')

    // Update body data attribute
    document.body.setAttribute('data-theme', currentTheme.value?.id || '')

    // Update meta theme color
    const metaThemeColor = document.querySelector('meta[name="theme-color"]')
    if (metaThemeColor) {
      metaThemeColor.setAttribute('content', backgroundColor.value)
    }
  }

  const saveTheme = async (theme: Theme): Promise<void> => {
    try {
      if (wails.isReady.value) {
        await wails.callService('theme', 'saveTheme', theme)
      }

      // Update theme in available themes list
      const index = availableThemes.value.findIndex(t => t.id === theme.id)
      if (index >= 0) {
        availableThemes.value[index] = theme
      } else {
        availableThemes.value.push(theme)
      }

      // If this is the current theme, update it
      if (currentTheme.value?.id === theme.id) {
        currentTheme.value = theme
        await generateCSSVariables(theme)
        applyThemeToDOM()
      }

      triggerEvent('theme:saved', theme)
    } catch (err) {
      console.error('Failed to save theme:', err)
      throw err
    }
  }

  const createTheme = (themeData: Partial<Theme>): Theme => {
    const now = new Date().toISOString()
    const theme: Theme = {
      id: themeData.id || generateThemeId(themeData.name || 'custom'),
      name: themeData.name || 'Custom Theme',
      displayName: themeData.displayName || themeData.name || 'Custom Theme',
      description: themeData.description || '',
      author: themeData.author || 'User',
      version: themeData.version || '1.0.0',
      colors: themeData.colors || getDefaultColors(),
      fonts: themeData.fonts || getDefaultFonts(),
      effects: themeData.effects || getDefaultEffects(),
      settings: themeData.settings || getDefaultSettings(),
      customProperties: themeData.customProperties || {},
      metadata: {
        createdAt: now,
        updatedAt: now,
        tags: themeData.metadata?.tags || [],
        category: themeData.metadata?.category || 'custom'
      }
    }

    return theme
  }

  const deleteTheme = async (themeId: string): Promise<void> => {
    if (!themeId) {
      throw new Error('Theme ID cannot be empty')
    }

    if (themeId === config.value.defaultTheme) {
      throw new Error('Cannot delete default theme')
    }

    try {
      if (wails.isReady.value) {
        await wails.callService('theme', 'deleteTheme', themeId)
      }

      // Remove from available themes
      availableThemes.value = availableThemes.value.filter(t => t.id !== themeId)

      // If this was the current theme, switch to default
      if (currentTheme.value?.id === themeId) {
        await setCurrentTheme(config.value.defaultTheme!)
      }

      triggerEvent('theme:deleted', undefined)
    } catch (err) {
      console.error('Failed to delete theme:', err)
      throw err
    }
  }

  // Event handling
  const subscribe = (eventType: string, callback: (theme?: Theme) => void): void => {
    eventListeners.set(eventType, callback)
  }

  const unsubscribe = (eventType: string): void => {
    eventListeners.delete(eventType)
  }

  const triggerEvent = (eventType: string, theme?: Theme): void => {
    const callback = eventListeners.get(eventType)
    if (callback) {
      callback(theme)
    }
  }

  // Utility functions
  const generateThemeId = (name: string): string => {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]/g, '-')
      .replace(/-+/g, '-')
      .replace(/^-|-$/g, '')
  }

  const hexToRgb = (hex: string): { r: number; g: number; b: number } | null => {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
    return result ? {
      r: parseInt(result[1], 16),
      g: parseInt(result[2], 16),
      b: parseInt(result[3], 16)
    } : null
  }

  const getDefaultColors = (): ThemeColors => ({
    background: {
      primary: '#0a0a0a',
      secondary: '#1a1a1a',
      tertiary: '#2a2a2a'
    },
    foreground: {
      primary: '#ffffff',
      secondary: '#cccccc',
      tertiary: '#999999'
    },
    accent: {
      primary: '#00ff41',
      secondary: '#00cc33'
    },
    status: {
      success: '#00ff41',
      warning: '#ffaa00',
      error: '#ff3333',
      info: '#00aaff'
    },
    terminal: [
      '#000000', '#ff0000', '#00ff00', '#ffff00',
      '#0000ff', '#ff00ff', '#00ffff', '#ffffff',
      '#808080', '#ff8080', '#80ff80', '#ffff80',
      '#8080ff', '#ff80ff', '#80ffff', '#c0c0c0'
    ],
    ui: {
      buttonBackground: '#1a1a1a',
      buttonForeground: '#ffffff',
      buttonHover: '#2a2a2a',
      buttonActive: '#00ff41',
      inputBackground: '#0a0a0a',
      inputForeground: '#ffffff',
      inputBorder: '#333333',
      inputFocus: '#00ff41',
      border: '#333333',
      shadow: 'rgba(0, 0, 0, 0.5)'
    }
  })

  // Legacy methods for backward compatibility
  const switchTheme = async (themeName: string): Promise<boolean> => {
    try {
      await setCurrentTheme(themeName)
      return true
    } catch (error) {
      console.error('Failed to switch theme:', error)
      return false
    }
  }

  const toggleDarkMode = async (): Promise<void> => {
    const newDarkState = !isDark.value
    const availableThemeNames = newDarkState ? darkThemes.value : lightThemes.value

    if (availableThemeNames.length > 0) {
      const targetTheme = availableThemeNames.find(t => t.id === currentTheme.value?.id) || availableThemeNames[0]
      await setCurrentTheme(targetTheme.id)
    }
  }

  const createCustomTheme = async (themeData: CustomTheme): Promise<boolean> => {
    try {
      customThemes.value.push(themeData)

      // Save to backend if available
      if (wails.isReady.value) {
        await saveTheme(themeData.theme)
      }

      return true
    } catch (error) {
      console.error('Failed to create custom theme:', error)
      return false
    }
  }

  const deleteCustomTheme = async (themeName: string): Promise<boolean> => {
    try {
      const index = customThemes.value.findIndex(t => t.name === themeName)
      if (index > -1) {
        customThemes.value.splice(index, 1)

        // If current theme was deleted, switch to default
        if (currentTheme.value?.id === themeName) {
          await setCurrentTheme('cyberpunk')
        }

        return true
      }
    } catch (error) {
      console.error('Failed to delete custom theme:', error)
    }

    return false
  }

  const loadThemeFromStorage = (): string => {
    if (typeof localStorage !== 'undefined') {
      const savedTheme = localStorage.getItem('adex-theme')
      if (savedTheme) {
        return savedTheme
      }
    }
    return 'cyberpunk'
  }

  const initializeTheme = async (): Promise<void> => {
    try {
      isLoading.value = true

      // Load saved theme from storage
      const savedTheme = loadThemeFromStorage()

      // Apply initial theme
      await setCurrentTheme(savedTheme)

      // Setup auto-switch
      if (autoSwitch.value) {
        checkAutoSwitch()
        setInterval(checkAutoSwitch, 60000) // Check every minute
      }

    } catch (error) {
      console.error('Failed to initialize theme:', error)
    } finally {
      isLoading.value = false
    }
  }

  const checkAutoSwitch = (): void => {
    if (!autoSwitch.value) return

    const now = new Date()
    const currentTime = now.getHours() * 60 + now.getMinutes()

    const [fromHour, fromMin] = switchSchedule.value.from.split(':').map(Number)
    const [toHour, toMin] = switchSchedule.value.to.split(':').map(Number)

    const fromTime = fromHour * 60 + fromMin
    const toTime = toHour * 60 + toMin

    let shouldBeDark = false

    if (fromTime <= toTime) {
      // Normal day (e.g., 20:00 to 08:00 next day)
      shouldBeDark = currentTime >= fromTime || currentTime <= toTime
    } else {
      // Same day (e.g., 08:00 to 20:00)
      shouldBeDark = currentTime >= fromTime && currentTime <= toTime
    }

    if (shouldBeDark !== isDark.value) {
      toggleDarkMode()
    }
  }

  const generateAccentColor = (): string => {
    const colors = ['#ff00ff', '#00ffff', '#ffff00', '#ff00aa', '#00ffaa', '#aa00ff']
    return colors[Math.floor(Math.random() * colors.length)]
  }

  const applyGradientEffect = (element: HTMLElement, type: 'glow' | 'pulse' | 'slide'): void => {
    const animations = {
      glow: 'glow 2s ease-in-out infinite',
      pulse: 'pulse 1.5s ease-in-out infinite',
      slide: 'slideIn 0.3s ease-out',
    }

    element.style.animation = animations[type]
  }

  // Watch for theme changes
  watch(currentTheme, (newTheme) => {
    if (newTheme) {
      generateCSSVariables(newTheme)
      applyThemeToDOM()
    }
  })

  // Watch for auto-switch changes
  watch([autoSwitch, switchSchedule], () => {
    if (autoSwitch.value) {
      checkAutoSwitch()
    }
  })

  // Watch for Wails availability
  watch(() => wails.isReady.value, (isReady) => {
    if (isReady && !isInitialized.value) {
      initialize()
    }
  })

  // Auto-initialize on mount
  onMounted(async () => {
    if (!isInitialized.value) {
      await initialize()
    }
  })

  // Cleanup
  const cleanup = (): void => {
    eventListeners.clear()
    if (typeof document !== 'undefined') {
      const style = document.getElementById('dex-theme-variables')
      if (style) {
        style.remove()
      }
    }
  }

  onUnmounted(() => {
    cleanup()
  })

  return {
    // State (readonly for safety)
    isInitialized: readonly(isInitialized),
    isLoading: readonly(isLoading),
    currentTheme: readonly(currentTheme),
    availableThemes: readonly(availableThemes),
    cssVariables: readonly(cssVariables),
    error: readonly(error),
    config: readonly(config),

    // Legacy state (backward compatibility)
    customThemes: readonly(customThemes),
    themeTransition: readonly(themeTransition),
    autoSwitch,
    switchSchedule,

    // Computed
    isDark: readonly(isDark),
    isLightTheme,
    primaryColor,
    backgroundColor,
    foregroundColor,
    themeCategories,
    themesByCategory,
    currentThemeData,
    allAvailableThemes,
    darkThemes,
    lightThemes,

    // Core methods
    initialize,
    setCurrentTheme,
    saveTheme,
    deleteTheme,
    createTheme,
    generateCSSVariables,
    applyThemeToDOM,

    // Event methods
    subscribe,
    unsubscribe,
    triggerEvent,

    // Utility methods
    generateThemeId,
    hexToRgb,
    getDefaultColors,
    getDefaultFonts,
    getDefaultEffects,
    getDefaultSettings,

    // Legacy methods (backward compatibility)
    switchTheme,
    toggleDarkMode,
    createCustomTheme,
    deleteCustomTheme,
    generateAccentColor,
    applyGradientEffect,
    initializeTheme,
    loadThemeFromStorage,

    // Cleanup
    cleanup
  }
}

/**
 * useThemePreview composable for theme preview functionality
 */
export function useThemePreview() {
  const theme = useTheme()

  const previewTheme = (themeData: Partial<Theme>) => {
    const fullTheme = theme.createTheme(themeData)

    // Create temporary CSS variables
    const tempCSS = theme.generateFallbackCSS ?
      theme.generateFallbackCSS(fullTheme) :
      generatePreviewCSS(fullTheme)

    // Apply temporary styles
    if (typeof document !== 'undefined') {
      let previewStyle = document.getElementById('dex-theme-preview')
      if (!previewStyle) {
        previewStyle = document.createElement('style')
        previewStyle.id = 'dex-theme-preview'
        document.head.appendChild(previewStyle)
      }
      previewStyle.textContent = tempCSS
    }

    return fullTheme
  }

  const generatePreviewCSS = (theme: Theme): string => {
    const prefix = '--dex-theme-preview'
    let css = `:root {\n`

    // Add color variables
    css += `  ${prefix}-background-primary: ${theme.colors.background.primary};\n`
    css += `  ${prefix}-foreground-primary: ${theme.colors.foreground.primary};\n`
    css += `  ${prefix}-accent-primary: ${theme.colors.accent.primary};\n`

    css += `}\n`
    return css
  }

  const clearPreview = () => {
    if (typeof document !== 'undefined') {
      const previewStyle = document.getElementById('dex-theme-preview')
      if (previewStyle) {
        previewStyle.remove()
      }
    }
  }

  return {
    previewTheme,
    clearPreview
  }
}
