import { defineStore } from 'pinia'
import {
  GetSchemes,
  GetScheme,
  CreateScheme,
  UpdateScheme,
  DeleteScheme,
  SetDefaultScheme,
  GetDefaultScheme,
  GetConfig,
  UpdateConfig,
  GetSchemePreview,
  ValidateScheme
} from '~/lib/wailsjs/coordinator'

export interface ColorScheme {
  id: string
  name: string
  displayName?: string
  description?: string
  colors: {
    background: string
    foreground: string
    cursor: string
    cursorAccent?: string
    selection: string
    selectionForeground?: string
    black: string
    red: string
    green: string
    yellow: string
    blue: string
    magenta: string
    cyan: string
    white: string
    brightBlack: string
    brightRed: string
    brightGreen: string
    brightYellow: string
    brightBlue: string
    brightMagenta: string
    brightCyan: string
    brightWhite: string
    ansi?: Record<number, string>
  }
  font: {
    family: string
    size: number
    weight: string
    lineHeight: number
    letterSpacing?: number
    ligatures: boolean
    antialias: boolean
    hinting: string
  }
  cursor: {
    style: string
    blink: boolean
    blinkInterval?: number
    width?: number
    color?: string
    accent?: string
  }
  background: {
    type: string
    value?: string
    opacity?: number
    blur?: number
    size?: string
    position?: string
  }
  author?: string
  version?: string
  isBuiltIn?: boolean
  isDark?: boolean
  createdAt: string
  updatedAt: string
  extensions?: Record<string, any>
}

export interface ColorSchemeConfig {
  defaultScheme: string
  userSchemes: Record<string, ColorScheme>
  enabledSchemes: string[]
  autoSwitch: boolean
  importPath: string
  exportPath: string
}

export interface ColorSchemePreview {
  schemeId: string
  name: string
  colors: Record<string, string>
  preview: string
  sample: string
  createdAt: string
}

export interface ColorSchemeValidationResult {
  valid: boolean
  errors: Array<{
    field: string
    message: string
    code: string
  }>
  warnings: Array<{
    field: string
    message: string
    code: string
  }>
  scheme?: ColorScheme
}

export const useColorSchemeStore = defineStore('colorscheme', {
  state: () => ({
    schemes: {} as Record<string, ColorScheme>,
    defaultScheme: null as ColorScheme | null,
    config: null as ColorSchemeConfig | null,
    isLoading: false,
    error: null as string | null,
  }),

  getters: {
    userSchemes: (state) => {
      const userSchemes: Record<string, ColorScheme> = {}
      for (const [id, scheme] of Object.entries(state.schemes)) {
        if (!scheme.isBuiltIn) {
          userSchemes[id] = scheme
        }
      }
      return userSchemes
    },

    builtInSchemes: (state) => {
      const builtInSchemes: Record<string, ColorScheme> = {}
      for (const [id, scheme] of Object.entries(state.schemes)) {
        if (scheme.isBuiltIn) {
          builtInSchemes[id] = scheme
        }
      }
      return builtInSchemes
    },

    enabledSchemes: (state) => {
      if (!state.config) return []
      return state.config.enabledSchemes
        .map(id => state.schemes[id])
        .filter(Boolean) as ColorScheme[]
    },

    darkSchemes: (state) => {
      const darkSchemes: Record<string, ColorScheme> = {}
      for (const [id, scheme] of Object.entries(state.schemes)) {
        if (scheme.isDark) {
          darkSchemes[id] = scheme
        }
      }
      return darkSchemes
    },

    lightSchemes: (state) => {
      const lightSchemes: Record<string, ColorScheme> = {}
      for (const [id, scheme] of Object.entries(state.schemes)) {
        if (!scheme.isDark) {
          lightSchemes[id] = scheme
        }
      }
      return lightSchemes
    },
  },

  actions: {
    async fetchSchemes() {
      this.isLoading = true
      this.error = null

      try {
        // Call the Go backend
        const schemes = await GetSchemes()

        this.schemes = schemes

        // Get default scheme
        const defaultScheme = await this.getDefaultScheme()
        this.defaultScheme = defaultScheme

        // Get config
        const config = await this.getConfig()
        this.config = config

      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to fetch color schemes'
        throw error
      } finally {
        this.isLoading = false
      }
    },

    async getScheme(id: string) {
      try {
        return await GetScheme(id)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to get color scheme'
        throw error
      }
    },

    async createScheme(scheme: ColorScheme) {
      try {
        return await CreateScheme(scheme)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to create color scheme'
        throw error
      }
    },

    async updateScheme(scheme: ColorScheme) {
      try {
        return await UpdateScheme(scheme)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to update color scheme'
        throw error
      }
    },

    async deleteScheme(id: string) {
      try {
        await DeleteScheme(id)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to delete color scheme'
        throw error
      }
    },

    async setDefaultScheme(id: string) {
      try {
        await SetDefaultScheme(id)

        // Update local state
        const scheme = this.schemes[id]
        if (scheme) {
          this.defaultScheme = scheme
        }
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to set default color scheme'
        throw error
      }
    },

    async getDefaultScheme() {
      try {
        return await GetDefaultScheme()
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to get default color scheme'
        throw error
      }
    },

    async getConfig() {
      try {
        return await GetConfig()
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to get color scheme configuration'
        throw error
      }
    },

    async updateConfig(config: Partial<ColorSchemeConfig>) {
      try {
        const newConfig = { ...this.config, ...config } as ColorSchemeConfig
        await UpdateConfig(newConfig)
        this.config = newConfig
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to update color scheme configuration'
        throw error
      }
    },

    async getSchemePreview(id: string) {
      try {
        return await GetSchemePreview(id)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to get color scheme preview'
        throw error
      }
    },

    async validateScheme(scheme: ColorScheme) {
      try {
        return await ValidateScheme(scheme)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to validate color scheme'
        throw error
      }
    },

    async resetToDefaults() {
      try {
        // Reset to default dark scheme
        await this.setDefaultScheme('default-dark')

        // Update config to defaults
        const defaultConfig: Partial<ColorSchemeConfig> = {
          defaultScheme: 'default-dark',
          enabledSchemes: ['default-dark', 'default-light', 'solarized-dark', 'solarized-light'],
          autoSwitch: false,
        }
        await this.updateConfig(defaultConfig)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to reset to defaults'
        throw error
      }
    },

    async exportScheme(id: string, format: string = 'json') {
      try {
        const scheme = this.schemes[id]
        if (!scheme) {
          throw new Error(`Color scheme with ID '${id}' not found`)
        }

        const data = JSON.stringify(scheme, null, 2)
        const blob = new Blob([data], { type: 'application/json' })
        const url = URL.createObjectURL(blob)

        const link = document.createElement('a')
        link.href = url
        link.download = `${scheme.name}.json`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)

        URL.revokeObjectURL(url)

        return data
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to export color scheme'
        throw error
      }
    },

    async importScheme(data: any, format: string = 'json') {
      try {
        let scheme: ColorScheme

        if (format === 'json') {
          if (typeof data === 'string') {
            scheme = JSON.parse(data)
          } else {
            scheme = data
          }
        } else {
          throw new Error(`Unsupported import format: ${format}`)
        }

        // Validate the scheme
        const validation = await this.validateScheme(scheme)
        if (!validation.valid) {
          throw new Error(`Invalid color scheme: ${validation.errors.map(e => e.message).join(', ')}`)
        }

        // Generate a unique ID if not provided
        if (!scheme.id) {
          scheme.id = `user-${Date.now()}`
        }

        // Ensure it's marked as user scheme
        scheme.isBuiltIn = false
        scheme.createdAt = new Date().toISOString()
        scheme.updatedAt = new Date().toISOString()

        // Create the scheme
        return await this.createScheme(scheme)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Failed to import color scheme'
        throw error
      }
    },

    // Utility methods
    getXtermTheme(scheme: ColorScheme) {
      if (!scheme || !scheme.colors) return {}

      const colors = scheme.colors
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
        brightWhite: colors.brightWhite,
      }
    },

    clearError() {
      this.error = null
    },
  },
})