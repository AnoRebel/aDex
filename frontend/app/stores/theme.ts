import { defineStore } from 'pinia'
import { useTheme } from '~/composables/useTheme'

// Import types from composable for now (until types/theme.ts is created)
import type {
  Theme,
  CustomTheme,
  ThemeSettings,
  ThemeColors,
  ThemeFonts,
  ThemeEffects,
  ThemeServiceConfig
} from '~/composables/useTheme'

interface ThemeState {
  // Current state
  currentThemeId: string
  currentTheme: Theme | null
  isDark: boolean
  isInitialized: boolean

  // Collections
  availableThemes: Theme[]
  customThemes: CustomTheme[]

  // UI state
  isLoading: boolean
  error: string | null

  // Settings
  settings: ThemeSettings
  config: ThemeServiceConfig

  // Legacy support
  predefinedThemes: Record<string, Theme>
}

export const useThemeStore = defineStore('theme', {
  state: (): ThemeState => ({
    // Current state
    currentThemeId: 'cyberpunk',
    currentTheme: null,
    isDark: true,
    isInitialized: false,

    // Collections
    availableThemes: [],
    customThemes: [],

    // UI state
    isLoading: false,
    error: null,

    // Settings
    settings: {
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
    },

    // Configuration
    config: {
      themeDirectory: 'themes',
      defaultTheme: 'cyberpunk',
      autoSave: true,
      autoReload: true,
      cacheEnabled: true,
      variablePrefix: '--dex-theme',
      minifyCSS: false,
      watchInterval: 5000,
      maxCacheSize: 100,
      enableLegacyImport: true
    },
    predefinedThemes: {
      cyberpunk: {
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
        fonts: {
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
        },
        effects: {
          borderRadius: {
            small: '0.25rem',
            medium: '0.5rem',
            large: '1rem',
            extraLarge: '1.5rem',
            full: '9999px'
          },
          shadows: [
            '0 1px 3px rgba(0, 255, 0, 0.12), 0 1px 2px rgba(0, 255, 0, 0.24)',
            '0 4px 6px rgba(255, 0, 255, 0.16), 0 2px 4px rgba(255, 0, 255, 0.12)'
          ],
          gradients: [
            'linear-gradient(135deg, #00ff00 0%, #ff00ff 100%)',
            'linear-gradient(135deg, #ff00ff 0%, #00ffff 100%)'
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
        },
        settings: {
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
        },
        metadata: {
          category: 'built-in',
          tags: ['cyberpunk', 'neon', 'dark']
        }
      },
      
      matrix: {
        name: 'matrix',
        displayName: 'Matrix',
        description: 'Classic Matrix green rain',
        isDark: true,
        colors: {
          primary: '#00ff00',
          secondary: '#00dd00',
          accent: '#00ff00',
          background: '#000000',
          surface: '#0a0a0a',
          text: '#00ff00',
          textSecondary: '#00dd00',
          border: '#00ff00',
          success: '#00ff00',
          warning: '#ffaa00',
          error: '#ff0066',
          info: '#00ffff',
        },
        typography: {
          fontFamily: '"Courier New", monospace',
          fontSize: {
            xs: '0.75rem',
            sm: '0.875rem',
            base: '1rem',
            lg: '1.125rem',
            xl: '1.25rem',
          },
        },
        effects: {
          glow: true,
          animation: true,
          shadows: true,
        },
      },
      
      neon: {
        name: 'neon',
        displayName: 'Neon Nights',
        description: 'Vibrant neon colors on dark',
        isDark: true,
        colors: {
          primary: '#ff00ff',
          secondary: '#cc00cc',
          accent: '#00ffff',
          background: '#0d0221',
          surface: '#1a0b3d',
          text: '#ff00ff',
          textSecondary: '#cc00cc',
          border: '#ff00ff',
          success: '#00ff00',
          warning: '#ffaa00',
          error: '#ff0066',
          info: '#00ffff',
        },
        typography: {
          fontFamily: '"JetBrains Mono", monospace',
          fontSize: {
            xs: '0.75rem',
            sm: '0.875rem',
            base: '1rem',
            lg: '1.125rem',
            xl: '1.25rem',
          },
        },
        effects: {
          glow: true,
          animation: true,
          shadows: true,
        },
      },
      
      dark: {
        name: 'dark',
        displayName: 'Dark Mode',
        description: 'Clean dark theme',
        isDark: true,
        colors: {
          primary: '#00ff00',
          secondary: '#00cc00',
          accent: '#0099ff',
          background: '#1a1a1a',
          surface: '#2a2a2a',
          text: '#ffffff',
          textSecondary: '#cccccc',
          border: '#444444',
          success: '#00ff00',
          warning: '#ffaa00',
          error: '#ff0066',
          info: '#0099ff',
        },
        typography: {
          fontFamily: '"Inter", sans-serif',
          fontSize: {
            xs: '0.75rem',
            sm: '0.875rem',
            base: '1rem',
            lg: '1.1125rem',
            xl: '1.25rem',
          },
        },
        effects: {
          glow: false,
          animation: false,
          shadows: true,
        },
      },
      
      light: {
        name: 'light',
        displayName: 'Light Mode',
        description: 'Clean light theme',
        isDark: false,
        colors: {
          primary: '#0066cc',
          secondary: '#0052a3',
          accent: '#00aa44',
          background: '#ffffff',
          surface: '#f5f5f5',
          text: '#000000',
          textSecondary: '#333333',
          border: '#cccccc',
          success: '#00aa44',
          warning: '#ff8800',
          error: '#cc0000',
          info: '#0066cc',
        },
        typography: {
          fontFamily: '"Inter", sans-serif',
          fontSize: {
            xs: '0.75rem',
            sm: '0.875rem',
            base: '1rem',
            lg: '1.1125rem',
            xl: '1.25rem',
          },
        },
        effects: {
          glow: false,
          animation: false,
          shadows: true,
        },
      },
    },
  }),

  getters: {
    // Get theme composable instance
    themeComposable: () => {
      return useTheme()
    },

    // Current theme data (backward compatibility)
    currentThemeData: (state) => {
      if (state.currentTheme) {
        return state.currentTheme
      }

      const customTheme = state.customThemes.find(t => t.name === state.currentThemeId)
      return customTheme?.theme || state.predefinedThemes[state.currentThemeId] || state.predefinedThemes.cyberpunk
    },

    // All available themes
    availableThemes: (state) => {
      if (state.availableThemes.length > 0) {
        return state.availableThemes
      }

      const themes = Object.values(state.predefinedThemes)
      const custom = state.customThemes.map(ct => ct.theme)
      return [...themes, ...custom]
    },

    // Dark themes
    darkThemes: (state) =>
      state.availableThemes.filter(theme => theme.isDark !== false),

    // Light themes
    lightThemes: (state) =>
      state.availableThemes.filter(theme => theme.isDark === false),

    // Theme categories
    themeCategories: (state) => {
      const categories = new Set<string>()
      state.availableThemes.forEach(theme => {
        if (theme.metadata?.category) {
          categories.add(theme.metadata.category)
        }
      })
      return Array.from(categories)
    },

    // Themes grouped by category
    themesByCategory: (state) => {
      const grouped: Record<string, Theme[]> = {}
      state.availableThemes.forEach(theme => {
        const category = theme.metadata?.category || 'uncategorized'
        if (!grouped[category]) {
          grouped[category] = []
        }
        grouped[category].push(theme)
      })
      return grouped
    },

    // Legacy getters for backward compatibility
    customThemeByName: (state) => {
      return (name: string) => state.customThemes.find(t => t.name === name)
    },

    isCustomTheme: (state) => {
      return (themeName: string) =>
        !!state.customThemes.find(t => t.name === themeName)
    },

    // Theme variables (generated from theme composable)
    cssVariables: (state) => {
      const theme = state.currentTheme || state.currentThemeData
      if (!theme) return ''

      const prefix = state.config.variablePrefix || '--dex-theme'
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
    },

    // Check if store is ready
    isReady: (state) => {
      return state.isInitialized && state.currentTheme !== null
    }
  },

  actions: {
    // Initialize store with theme composable
    async initialize() {
      try {
        this.isLoading = true
        this.error = null

        const themeComposable = this.themeComposable

        // Initialize theme composable
        await themeComposable.initialize()

        // Sync state with composable
        this.syncWithComposable()

        // Set up event listeners
        themeComposable.subscribe('theme:changed', (theme) => {
          this.syncWithComposable()
        })

        themeComposable.subscribe('theme:saved', (theme) => {
          if (theme) {
            const index = this.availableThemes.findIndex(t => t.id === theme.id)
            if (index >= 0) {
              this.availableThemes[index] = theme
            } else {
              this.availableThemes.push(theme)
            }
          }
        })

        this.isInitialized = true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to initialize theme store'
        console.error('Theme store initialization failed:', error)
      } finally {
        this.isLoading = false
      }
    },

    // Sync store state with composable
    syncWithComposable() {
      const themeComposable = this.themeComposable

      this.currentThemeId = themeComposable.currentTheme.value?.id || 'cyberpunk'
      this.currentTheme = themeComposable.currentTheme.value
      this.isDark = themeComposable.isDark.value
      this.availableThemes = [...themeComposable.availableThemes.value]
      this.error = themeComposable.error.value
    },

    // Set current theme (integrates with composable)
    async setCurrentTheme(themeId: string) {
      try {
        this.isLoading = true
        this.error = null

        const themeComposable = this.themeComposable
        await themeComposable.setCurrentTheme(themeId)

        // Sync state
        this.syncWithComposable()

        // Save to localStorage for backward compatibility
        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('adex-theme', themeId)
        }

        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to set theme'
        console.error('Failed to set theme:', error)
        return false
      } finally {
        this.isLoading = false
      }
    },

    // Toggle dark mode (integrates with composable)
    async toggleDarkMode() {
      try {
        const themeComposable = this.themeComposable
        await themeComposable.toggleDarkMode()

        // Sync state
        this.syncWithComposable()

        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to toggle dark mode'
        console.error('Failed to toggle dark mode:', error)
        return false
      }
    },

    // Add custom theme (integrates with composable)
    async addCustomTheme(customTheme: CustomTheme) {
      try {
        this.customThemes.push(customTheme)

        // Use composable to save theme
        const themeComposable = this.themeComposable
        await themeComposable.createCustomTheme(customTheme)

        // Sync state
        this.syncWithComposable()

        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to add custom theme'
        console.error('Failed to add custom theme:', error)
        return false
      }
    },

    // Remove custom theme (integrates with composable)
    async removeCustomTheme(themeName: string) {
      try {
        const index = this.customThemes.findIndex(t => t.name === themeName)
        if (index > -1) {
          this.customThemes.splice(index, 1)

          // If current theme was deleted, switch to default
          if (this.currentThemeId === themeName) {
            await this.setCurrentTheme(this.config.defaultTheme)
          }

          // Use composable to delete theme
          const themeComposable = this.themeComposable
          await themeComposable.deleteCustomTheme(themeName)

          // Sync state
          this.syncWithComposable()
        }

        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to remove custom theme'
        console.error('Failed to remove custom theme:', error)
        return false
      }
    },

    // Update custom theme
    async updateCustomTheme(themeName: string, updates: Partial<CustomTheme>) {
      try {
        const theme = this.customThemes.find(t => t.name === themeName)
        if (theme) {
          Object.assign(theme, updates)
          theme.modifiedAt = new Date()

          // Use composable to save updated theme
          const themeComposable = this.themeComposable
          await themeComposable.saveTheme(theme.theme)

          // Sync state
          this.syncWithComposable()

          // If this is the current theme, it will be automatically updated
          // by the theme:changed event listener
        }

        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to update custom theme'
        console.error('Failed to update custom theme:', error)
        return false
      }
    },

    // Apply theme (legacy method - now uses composable)
    applyTheme(theme: Theme) {
      const themeComposable = this.themeComposable
      // The composable will handle DOM updates
      return themeComposable.applyThemeToDOM()
    },

    // Settings management (backward compatibility)
    updateSettings(newSettings: Partial<ThemeState['settings']>) {
      this.settings = { ...this.settings, ...newSettings }
      this.saveSettings()
    },

    saveSettings() {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('adex-theme-settings', JSON.stringify(this.settings))
      }
    },

    loadSettings() {
      if (typeof localStorage !== 'undefined') {
        const saved = localStorage.getItem('adex-theme-settings')
        if (saved) {
          try {
            const settings = JSON.parse(saved)
            this.settings = { ...this.settings, ...settings }
          } catch (error) {
            console.error('Failed to load theme settings:', error)
          }
        }
      }
    },

    // Custom themes management (backward compatibility)
    saveCustomThemes() {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('adex-custom-themes', JSON.stringify(this.customThemes))
      }
    },

    loadCustomThemes() {
      if (typeof localStorage !== 'undefined') {
        const saved = localStorage.getItem('adex-custom-themes')
        if (saved) {
          try {
            const themes = JSON.parse(saved)
            this.customThemes = themes.map((theme: any) => ({
              ...theme,
              createdAt: new Date(theme.createdAt),
              modifiedAt: new Date(theme.modifiedAt),
            }))
          } catch (error) {
            console.error('Failed to load custom themes:', error)
          }
        }
      }
    },

    // Utility methods (backward compatibility)
    generateAccentColor(): string {
      const colors = ['#ff00ff', '#00ffff', '#ffff00', '#ff00aa', '#00ffaa', '#aa00ff']
      return colors[Math.floor(Math.random() * colors.length)]
    },

    exportThemes(): string {
      const exportData = {
        themes: this.availableThemes,
        customThemes: this.customThemes,
        settings: this.settings,
        version: '1.0.0',
        exportDate: new Date(),
      }

      return JSON.stringify(exportData, null, 2)
    },

    importThemes(themeData: string): boolean {
      try {
        const data = JSON.parse(themeData)

        if (data.customThemes && Array.isArray(data.customThemes)) {
          this.customThemes = [...this.customThemes, ...data.customThemes]
          this.saveCustomThemes()
        }

        if (data.settings) {
          this.updateSettings(data.settings)
        }

        return true
      } catch (error) {
        console.error('Failed to import themes:', error)
        return false
      }
    },

    // Loading state management
    setLoading(isLoading: boolean) {
      this.isLoading = isLoading
    },

    // Error handling
    clearError() {
      this.error = null
    },

    // Refresh themes from composable
    async refreshThemes() {
      try {
        const themeComposable = this.themeComposable
        await themeComposable.refreshThemes()
        this.syncWithComposable()
        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to refresh themes'
        console.error('Failed to refresh themes:', error)
        return false
      }
    },

    // Create new theme using composable
    async createTheme(themeData: Partial<Theme>) {
      try {
        const themeComposable = this.themeComposable
        const theme = themeComposable.createTheme(themeData)
        await themeComposable.saveTheme(theme)

        // Sync state
        this.syncWithComposable()

        return theme
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to create theme'
        console.error('Failed to create theme:', error)
        throw error
      }
    },

    // Delete theme using composable
    async deleteTheme(themeId: string) {
      try {
        const themeComposable = this.themeComposable
        await themeComposable.deleteTheme(themeId)

        // Sync state
        this.syncWithComposable()

        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to delete theme'
        console.error('Failed to delete theme:', error)
        return false
      }
    },

    // Legacy reset method
    reset() {
      this.currentThemeId = 'cyberpunk'
      this.currentTheme = null
      this.isDark = true
      this.customThemes = []
      this.error = null

      // Reinitialize with composable
      this.initialize()
    },

    // Cleanup
    cleanup() {
      const themeComposable = this.themeComposable
      themeComposable.cleanup()
    }
  },
})
