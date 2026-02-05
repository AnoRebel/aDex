import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import type { FontConfiguration, FontSettings, SystemFont, FontValidationResult } from '~/types/font'

// Import from bindings
import {
  GetFontConfigurations,
  GetFontConfiguration,
  CreateFontConfiguration,
  UpdateFontConfiguration,
  DeleteFontConfiguration,
  GetSystemFonts,
  GetMonospaceFonts,
  ScanSystemFonts,
  ImportFont,
  ValidateFont,
  GetFontMetrics,
  GetFontSettings,
  UpdateFontSettings,
  GetDefaultFontConfiguration,
  SetDefaultFontConfiguration
} from '~~/bindings'

// Wails bindings would be generated - for now we'll define the interface
interface FontService {
  GetConfigurations(): Promise<Record<string, FontConfiguration>>
  GetConfiguration(id: string): Promise<FontConfiguration | null>
  CreateConfiguration(config: FontConfiguration): Promise<void>
  UpdateConfiguration(config: FontConfiguration): Promise<void>
  DeleteConfiguration(id: string): Promise<void>
  GetSystemFonts(): Promise<Record<string, SystemFont>>
  GetMonospaceFonts(): Promise<SystemFont[]>
  ScanSystemFonts(): Promise<void>
  ImportFont(request: FontImportRequest): Promise<FontImportResult>
  ValidateFont(config: FontConfiguration): Promise<FontValidationResult>
  GetFontMetrics(family: string, size: number): Promise<FontMetrics>
  GetSettings(): Promise<FontSettings>
  UpdateSettings(settings: FontSettings): Promise<void>
  GetDefaultConfiguration(): Promise<FontConfiguration>
  SetDefaultConfiguration(id: string): Promise<void>
}

interface FontImportRequest {
  source: string
  name?: string
  family?: string
  weight?: string
  style?: string
  license?: string
  metadata?: Record<string, string>
  overwrite?: boolean
  validateOnly?: boolean
}

interface FontImportResult {
  success: boolean
  font?: SystemFont
  configuration?: FontConfiguration
  errors: string[]
  warnings: string[]
  importPath?: string
  checksum?: string
}

interface FontMetrics {
  family: string
  size: number
  weight: string
  ascent: number
  descent: number
  lineGap: number
  capHeight: number
  xHeight: number
  avgCharWidth: number
  maxCharWidth: number
  isMonospace: boolean
  characterCount?: number
  supportedChars?: string[]
}

// Mock Wails service - in real implementation this would be imported from bindings
const mockFontService: FontService = {
  GetConfigurations: async () => ({}),
  GetConfiguration: async () => null,
  CreateConfiguration: async () => {},
  UpdateConfiguration: async () => {},
  DeleteConfiguration: async () => {},
  GetSystemFonts: async () => ({}),
  GetMonospaceFonts: async () => [],
  ScanSystemFonts: async () => {},
  ImportFont: async () => ({ success: true, errors: [], warnings: [] }),
  ValidateFont: async () => ({ isValid: true, errors: [], warnings: [], suggestions: [], features: [], limitations: [] }),
  GetFontMetrics: async () => ({ family: '', size: 14, weight: '400', ascent: 11.2, descent: 2.8, lineGap: 2.8, capHeight: 9.8, xHeight: 7, avgCharWidth: 8.4, maxCharWidth: 12.6, isMonospace: true }),
  GetSettings: async () => ({
    defaultFontId: 'jetbrains-mono',
    fallbackFont: 'monospace',
    fontSize: 14,
    lineHeight: 1.4,
    letterSpacing: 0,
    enableLigatures: true,
    enableAntialias: true,
    hinting: 'slight' as const,
    fontFeatureSettings: {},
    allowCustomFonts: false,
    fontDirectories: ['/usr/share/fonts', '/usr/local/share/fonts', '~/.fonts'],
    autoDetectFonts: true,
    lastScanTime: new Date().toISOString()
  }),
  UpdateSettings: async () => {},
  GetDefaultConfiguration: async () => ({
    id: 'jetbrains-mono',
    name: 'JetBrains Mono',
    family: 'JetBrains Mono',
    size: 14,
    weight: '400',
    lineHeight: 1.4,
    letterSpacing: 0,
    ligatures: true,
    antialias: true,
    hinting: 'slight' as const,
    features: { liga: true, dlig: true },
    isBuiltIn: true,
    isDefault: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    metadata: {}
  }),
  SetDefaultConfiguration: async () => {}
}

// Default font service - uses imported bindings
const fontService: FontService = {
  GetConfigurations: async () => {
    try {
      const result = await GetFontConfigurations()
      return result as Record<string, FontConfiguration>
    } catch {
      return mockFontService.GetConfigurations()
    }
  },

  GetConfiguration: async (id: string) => {
    try {
      const result = await GetFontConfiguration(id)
      return result as FontConfiguration | null
    } catch {
      return mockFontService.GetConfiguration(id)
    }
  },

  CreateConfiguration: async (config: FontConfiguration) => {
    try {
      await CreateFontConfiguration(config)
    } catch {
      return mockFontService.CreateConfiguration(config)
    }
  },

  UpdateConfiguration: async (config: FontConfiguration) => {
    try {
      await UpdateFontConfiguration(config)
    } catch {
      return mockFontService.UpdateConfiguration(config)
    }
  },

  DeleteConfiguration: async (id: string) => {
    try {
      await DeleteFontConfiguration(id)
    } catch {
      return mockFontService.DeleteConfiguration(id)
    }
  },

  GetSystemFonts: async () => {
    try {
      const result = await GetSystemFonts()
      return result as Record<string, SystemFont>
    } catch {
      return mockFontService.GetSystemFonts()
    }
  },

  GetMonospaceFonts: async () => {
    try {
      const result = await GetMonospaceFonts()
      return result as SystemFont[]
    } catch {
      return mockFontService.GetMonospaceFonts()
    }
  },

  ScanSystemFonts: async () => {
    try {
      await ScanSystemFonts()
    } catch {
      return mockFontService.ScanSystemFonts()
    }
  },

  ImportFont: async (request: FontImportRequest) => {
    try {
      const result = await ImportFont(request)
      return result as FontImportResult
    } catch {
      return mockFontService.ImportFont(request)
    }
  },

  ValidateFont: async (config: FontConfiguration) => {
    try {
      const result = await ValidateFont(config)
      return result as FontValidationResult
    } catch {
      return mockFontService.ValidateFont(config)
    }
  },

  GetFontMetrics: async (family: string, size: number) => {
    try {
      const result = await GetFontMetrics(family, size)
      return result as FontMetrics
    } catch {
      return mockFontService.GetFontMetrics(family, size)
    }
  },

  GetSettings: async () => {
    try {
      const result = await GetFontSettings()
      return result as FontSettings
    } catch {
      return mockFontService.GetSettings()
    }
  },

  UpdateSettings: async (newSettings: FontSettings) => {
    try {
      await UpdateFontSettings(newSettings)
    } catch {
      return mockFontService.UpdateSettings(newSettings)
    }
  },

  GetDefaultConfiguration: async () => {
    try {
      const result = await GetDefaultFontConfiguration()
      return result as FontConfiguration
    } catch {
      return mockFontService.GetDefaultConfiguration()
    }
  },

  SetDefaultConfiguration: async (id: string) => {
    try {
      await SetDefaultFontConfiguration(id)
    } catch {
      return mockFontService.SetDefaultConfiguration(id)
    }
  }
}

export function useFontConfiguration() {
  // State
  const configurations = ref<Record<string, FontConfiguration>>({})
  const systemFonts = ref<Record<string, SystemFont>>({})
  const monospaceFonts = ref<SystemFont[]>([])
  const settings = ref<FontSettings | null>(null)
  const currentConfiguration = ref<FontConfiguration | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const validationResults = ref<Map<string, FontValidationResult>>(new Map())

  // Computed
  const availableConfigurations = computed(() => {
    return Object.values(configurations.value).sort((a, b) => {
      // Built-in configurations first, then alphabetical
      if (a.isBuiltIn && !b.isBuiltIn) return -1
      if (!a.isBuiltIn && b.isBuiltIn) return 1
      return a.name.localeCompare(b.name)
    })
  })

  const customConfigurations = computed(() => {
    return availableConfigurations.value.filter(config => !config.isBuiltIn)
  })

  const builtinConfigurations = computed(() => {
    return availableConfigurations.value.filter(config => config.isBuiltIn)
  })

  const defaultConfiguration = computed(() => {
    return availableConfigurations.value.find(config => config.isDefault) || null
  })

  const hasUnsavedChanges = ref(false)

  // Methods
  const loadConfigurations = async () => {
    try {
      isLoading.value = true
      error.value = null

      const configs = await fontService.GetConfigurations()
      configurations.value = configs

      // Load current/default configuration
      if (settings.value?.defaultFontId) {
        currentConfiguration.value = await fontService.GetConfiguration(settings.value.defaultFontId)
      } else {
        currentConfiguration.value = await fontService.GetDefaultConfiguration()
      }

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load font configurations'
      console.error('Failed to load font configurations:', err)
    } finally {
      isLoading.value = false
    }
  }

  const loadSystemFonts = async () => {
    try {
      isLoading.value = true
      error.value = null

      const [fonts, monoFonts] = await Promise.all([
        fontService.GetSystemFonts(),
        fontService.GetMonospaceFonts()
      ])

      systemFonts.value = fonts
      monospaceFonts.value = monoFonts

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load system fonts'
      console.error('Failed to load system fonts:', err)
    } finally {
      isLoading.value = false
    }
  }

  const loadSettings = async () => {
    try {
      settings.value = await fontService.GetSettings()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load font settings'
      console.error('Failed to load font settings:', err)
    }
  }

  const createConfiguration = async (config: FontConfiguration) => {
    try {
      isLoading.value = true
      error.value = null

      // Validate configuration
      const validation = await validateConfiguration(config)
      if (!validation.isValid) {
        throw new Error(`Invalid configuration: ${validation.errors.join(', ')}`)
      }

      await fontService.CreateConfiguration(config)
      await loadConfigurations()

      return true

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create font configuration'
      console.error('Failed to create font configuration:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const updateConfiguration = async (config: FontConfiguration) => {
    try {
      isLoading.value = true
      error.value = null

      // Validate configuration
      const validation = await validateConfiguration(config)
      if (!validation.isValid) {
        throw new Error(`Invalid configuration: ${validation.errors.join(', ')}`)
      }

      await fontService.UpdateConfiguration(config)
      await loadConfigurations()

      // Update current configuration if it's the one being updated
      if (currentConfiguration.value?.id === config.id) {
        currentConfiguration.value = { ...config }
      }

      return true

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update font configuration'
      console.error('Failed to update font configuration:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const deleteConfiguration = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null

      await fontService.DeleteConfiguration(id)
      await loadConfigurations()

      // Clear current configuration if it was deleted
      if (currentConfiguration.value?.id === id) {
        currentConfiguration.value = await fontService.GetDefaultConfiguration()
      }

      return true

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete font configuration'
      console.error('Failed to delete font configuration:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const validateConfiguration = async (config: FontConfiguration): Promise<FontValidationResult> => {
    try {
      const result = await fontService.ValidateFont(config)
      validationResults.value.set(config.id || 'preview', result)
      return result
    } catch (err) {
      const errorResult: FontValidationResult = {
        isValid: false,
        errors: [err instanceof Error ? err.message : 'Validation failed'],
        warnings: [],
        suggestions: [],
        features: [],
        limitations: []
      }
      validationResults.value.set(config.id || 'preview', errorResult)
      return errorResult
    }
  }

  const scanSystemFonts = async () => {
    try {
      isLoading.value = true
      error.value = null

      await fontService.ScanSystemFonts()
      await loadSystemFonts()

      return true

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to scan system fonts'
      console.error('Failed to scan system fonts:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const importFont = async (request: FontImportRequest): Promise<FontImportResult> => {
    try {
      isLoading.value = true
      error.value = null

      const result = await fontService.ImportFont(request)

      if (result.success) {
        await loadSystemFonts()
        await loadConfigurations()
      }

      return result

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to import font'
      console.error('Failed to import font:', err)
      return {
        success: false,
        errors: [err instanceof Error ? err.message : 'Import failed'],
        warnings: []
      }
    } finally {
      isLoading.value = false
    }
  }

  const getFontMetrics = async (family: string, size: number): Promise<FontMetrics | null> => {
    try {
      return await fontService.GetFontMetrics(family, size)
    } catch (err) {
      console.error('Failed to get font metrics:', err)
      return null
    }
  }

  const updateSettings = async (newSettings: FontSettings) => {
    try {
      isLoading.value = true
      error.value = null

      await fontService.UpdateSettings(newSettings)
      await loadSettings()

      return true

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update font settings'
      console.error('Failed to update font settings:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const setDefaultConfiguration = async (id: string) => {
    try {
      isLoading.value = true
      error.value = null

      await fontService.SetDefaultConfiguration(id)
      await loadConfigurations()
      await loadSettings()

      // Update current configuration
      currentConfiguration.value = await fontService.GetConfiguration(id)

      return true

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to set default font configuration'
      console.error('Failed to set default font configuration:', err)
      return false
    } finally {
      isLoading.value = false
    }
  }

  const selectConfiguration = async (id: string) => {
    try {
      const config = await fontService.GetConfiguration(id)
      if (config) {
        currentConfiguration.value = config
        hasUnsavedChanges.value = false
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to select font configuration'
      console.error('Failed to select font configuration:', err)
    }
  }

  const resetToDefaults = async () => {
    try {
      isLoading.value = true
      error.value = null

      const defaultConfig = await fontService.GetDefaultConfiguration()
      if (defaultConfig) {
        currentConfiguration.value = { ...defaultConfig }
        hasUnsavedChanges.value = false
      }

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to reset to defaults'
      console.error('Failed to reset to defaults:', err)
    } finally {
      isLoading.value = false
    }
  }

  const duplicateConfiguration = async (sourceId: string, newName: string): Promise<boolean> => {
    try {
      const source = await fontService.GetConfiguration(sourceId)
      if (!source) {
        throw new Error('Source configuration not found')
      }

      const duplicate: FontConfiguration = {
        ...source,
        id: `${source.id}-copy-${Date.now()}`,
        name: newName,
        isBuiltIn: false,
        isDefault: false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }

      return await createConfiguration(duplicate)

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to duplicate configuration'
      console.error('Failed to duplicate configuration:', err)
      return false
    }
  }

  const exportConfiguration = (config: FontConfiguration): string => {
    return JSON.stringify(config, null, 2)
  }

  const importConfigurationFromJson = async (jsonString: string): Promise<boolean> => {
    try {
      const config = JSON.parse(jsonString) as FontConfiguration

      // Generate new ID and reset timestamps
      config.id = `imported-${Date.now()}`
      config.createdAt = new Date().toISOString()
      config.updatedAt = new Date().toISOString()
      config.isBuiltIn = false
      config.isDefault = false

      return await createConfiguration(config)

    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to import configuration from JSON'
      console.error('Failed to import configuration from JSON:', err)
      return false
    }
  }

  // Watch for unsaved changes
  watch(currentConfiguration, () => {
    hasUnsavedChanges.value = true
  }, { deep: true })

  // Initialize
  const initialize = async () => {
    await Promise.all([
      loadConfigurations(),
      loadSettings(),
      loadSystemFonts()
    ])
  }

  // Lifecycle
  onMounted(() => {
    initialize()
  })

  return {
    // State
    configurations,
    systemFonts,
    monospaceFonts,
    settings,
    currentConfiguration,
    isLoading,
    error,
    validationResults,
    hasUnsavedChanges,

    // Computed
    availableConfigurations,
    customConfigurations,
    builtinConfigurations,
    defaultConfiguration,

    // Methods
    loadConfigurations,
    loadSystemFonts,
    loadSettings,
    createConfiguration,
    updateConfiguration,
    deleteConfiguration,
    validateConfiguration,
    scanSystemFonts,
    importFont,
    getFontMetrics,
    updateSettings,
    setDefaultConfiguration,
    selectConfiguration,
    resetToDefaults,
    duplicateConfiguration,
    exportConfiguration,
    importConfigurationFromJson,

    // Utility
    initialize
  }
}

export function useFontPreview() {
  const previewText = ref(`The quick brown fox jumps over the lazy dog
1234567890 !@#$%^&*()[]{}<>+=-_
~\`|\\:";'<>?,./

function helloWorld() {
  console.log("Hello, World!");
  return true;
}

const array = [1, 2, 3, 4, 5];
const object = { key: "value", nested: { prop: true } };

// Comments with special characters: /* */ // → ← ↑ ↓
if (condition) {
  // This is a conditional statement
  execute();
}

// Unicode test: αβγδεζηθικλμνξοπρστυφχψω ☃ ❤ ✈ ⚡`)

  const previewSettings = ref({
    fontSize: 14,
    lineHeight: 1.4,
    letterSpacing: 0,
    showLineNumbers: true,
    showCharacterInfo: false
  })

  const updatePreviewText = (text: string) => {
    previewText.value = text
  }

  const updatePreviewSettings = (settings: Partial<typeof previewSettings.value>) => {
    previewSettings.value = { ...previewSettings.value, ...settings }
  }

  return {
    previewText,
    previewSettings,
    updatePreviewText,
    updatePreviewSettings
  }
}

export default useFontConfiguration