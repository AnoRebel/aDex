import { computed, ref, watch } from 'vue'
import { useColorSchemeStore } from '~/stores/colorscheme'

export interface ColorSchemeConfig {
  defaultScheme: string
  userSchemes: Record<string, any>
  enabledSchemes: string[]
  autoSwitch: boolean
  importPath: string
  exportPath: string
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
  scheme?: any
}

export interface ColorSchemePreview {
  schemeId: string
  name: string
  colors: Record<string, string>
  preview: string
  sample: string
  createdAt: string
}

export function useColorScheme() {
  const store = useColorSchemeStore()

  // Reactive state
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const selectedSchemeId = ref<string | null>(null)

  // Computed properties
  const schemes = computed(() => store.schemes)
  const defaultScheme = computed(() => store.defaultScheme)
  const config = computed(() => store.config)
  const userSchemes = computed(() => store.userSchemes)
  const builtInSchemes = computed(() => store.builtInSchemes)

  const currentScheme = computed(() => {
    if (selectedSchemeId.value) {
      return schemes.value[selectedSchemeId.value]
    }
    return defaultScheme.value
  })

  const enabledSchemes = computed(() => {
    const enabled = config.value?.enabledSchemes || []
    return enabled.map(id => schemes.value[id]).filter(Boolean)
  })

  // Methods
  const fetchSchemes = async () => {
    try {
      isLoading.value = true
      error.value = null
      await store.fetchSchemes()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch color schemes'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const getScheme = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      return await store.getScheme(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to get color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const createScheme = async (scheme: any) => {
    try {
      isLoading.value = true
      error.value = null
      const result = await store.createScheme(scheme)
      await fetchSchemes() // Refresh schemes
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const updateScheme = async (scheme: any) => {
    try {
      isLoading.value = true
      error.value = null
      const result = await store.updateScheme(scheme)
      await fetchSchemes() // Refresh schemes
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const deleteScheme = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      await store.deleteScheme(id)
      await fetchSchemes() // Refresh schemes

      // Clear selection if deleted scheme was selected
      if (selectedSchemeId.value === id) {
        selectedSchemeId.value = null
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const setDefaultScheme = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null
      await store.setDefaultScheme(id)
      await fetchSchemes() // Refresh schemes
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to set default color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const updateConfig = async (config: Partial<ColorSchemeConfig>) => {
    try {
      isLoading.value = true
      error.value = null
      await store.updateConfig(config)
      await fetchSchemes() // Refresh schemes
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update color scheme configuration'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const getSchemePreview = async (id: string): Promise<ColorSchemePreview | null> => {
    try {
      isLoading.value = true
      error.value = null
      return await store.getSchemePreview(id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to get color scheme preview'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const validateScheme = async (scheme: any): Promise<ColorSchemeValidationResult> => {
    try {
      isLoading.value = true
      error.value = null
      return await store.validateScheme(scheme)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to validate color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const selectScheme = (id: string | null) => {
    selectedSchemeId.value = id
  }

  const resetToDefaults = async () => {
    try {
      isLoading.value = true
      error.value = null
      await store.resetToDefaults()
      await fetchSchemes() // Refresh schemes
      selectedSchemeId.value = null
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to reset to defaults'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const exportScheme = async (id: string, format: string = 'json') => {
    try {
      isLoading.value = true
      error.value = null
      return await store.exportScheme(id, format)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to export color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  const importScheme = async (data: any, format: string = 'json') => {
    try {
      isLoading.value = true
      error.value = null
      const result = await store.importScheme(data, format)
      await fetchSchemes() // Refresh schemes
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to import color scheme'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Utility methods
  const isDarkScheme = (scheme: any) => {
    return scheme?.isDark ?? false
  }

  const isBuiltInScheme = (scheme: any) => {
    return scheme?.isBuiltIn ?? false
  }

  const getSchemeDisplayName = (scheme: any) => {
    return scheme?.displayName || scheme?.name || 'Unknown'
  }

  const getSchemeDescription = (scheme: any) => {
    return scheme?.description || ''
  }

  const getSchemeColors = (scheme: any) => {
    return scheme?.colors || {}
  }

  const getXtermTheme = (scheme: any) => {
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
  }

  // Auto-switch based on system theme
  const enableAutoSwitch = () => {
    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      const handleThemeChange = (e: MediaQueryListEvent) => {
        const preferredScheme = e.matches ? 'default-dark' : 'default-light'
        setDefaultScheme(preferredScheme)
      }

      mediaQuery.addEventListener('change', handleThemeChange)

      // Set initial theme
      const preferredScheme = mediaQuery.matches ? 'default-dark' : 'default-light'
      setDefaultScheme(preferredScheme)

      // Return cleanup function
      return () => {
        mediaQuery.removeEventListener('change', handleThemeChange)
      }
    }
    return () => {}
  }

  // Watch for config changes
  watch(() => config.value?.autoSwitch, (autoSwitch) => {
    if (autoSwitch) {
      enableAutoSwitch()
    }
  }, { immediate: true })

  // Initialize
  const init = async () => {
    await fetchSchemes()

    // Set initial selection to default scheme
    if (defaultScheme.value) {
      selectedSchemeId.value = defaultScheme.value.id
    }
  }

  return {
    // State
    isLoading,
    error,
    selectedSchemeId,

    // Computed
    schemes,
    defaultScheme,
    config,
    userSchemes,
    builtInSchemes,
    currentScheme,
    enabledSchemes,

    // Methods
    init,
    fetchSchemes,
    getScheme,
    createScheme,
    updateScheme,
    deleteScheme,
    setDefaultScheme,
    updateConfig,
    getSchemePreview,
    validateScheme,
    selectScheme,
    resetToDefaults,
    exportScheme,
    importScheme,

    // Utility methods
    isDarkScheme,
    isBuiltInScheme,
    getSchemeDisplayName,
    getSchemeDescription,
    getSchemeColors,
    getXtermTheme,
    enableAutoSwitch,
  }
}

export default useColorScheme