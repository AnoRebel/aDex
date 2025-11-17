// =============================================================================
// THEME SYSTEM TYPES
// =============================================================================
// This file contains comprehensive type definitions for the aDex-UI theme system
// It supports both the new backend-integrated architecture and legacy compatibility
// =============================================================================

// =============================================================================
// CORE THEME INTERFACES
// =============================================================================

/**
 * Main Theme interface representing a complete theme configuration
 */
export interface Theme {
  // Identification
  id: string
  name: string
  displayName?: string
  description: string
  author: string
  version: string

  // Visual Configuration
  colors: ThemeColors
  fonts: ThemeFonts
  effects: ThemeEffects
  settings: ThemeSettings

  // Extended Properties
  customProperties?: Record<string, any>
  metadata?: ThemeMetadata

  // Legacy Support
  isDark?: boolean
  isCustom?: boolean

  // Timestamps
  createdAt?: string
  updatedAt?: string
}

/**
 * Comprehensive color system with semantic groupings
 */
export interface ThemeColors {
  // Background colors
  background: {
    primary: string    // Main background
    secondary: string  // Secondary background
    tertiary: string   // Tertiary background
  }

  // Foreground/text colors
  foreground: {
    primary: string    // Main text color
    secondary: string  // Secondary text
    tertiary: string   // Tertiary text
  }

  // Accent colors for highlights and interactions
  accent: {
    primary: string    // Primary accent
    secondary: string  // Secondary accent
  }

  // Status colors for different states
  status: {
    success: string
    warning: string
    error: string
    info: string
  }

  // Terminal color palette (16 colors)
  terminal: string[]

  // UI component-specific colors
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

/**
 * Font configuration with families, sizes, weights, and spacing
 */
export interface ThemeFonts {
  // Font families
  families: {
    primary: string
    secondary: string
    monospace: string
  }

  // Font sizes
  sizes: {
    extraSmall: string      // xs
    small: string           // sm
    base: string            // md
    large: string           // lg
    extraLarge: string      // xl
    doubleExtraLarge: string // 2xl
  }

  // Font weights
  weights: {
    light: string     // 300
    normal: string    // 400
    medium: string    // 500
    semiBold: string   // 600
    bold: string       // 700
  }

  // Line heights
  lineHeights: {
    tight: string     // 1.25
    normal: string    // 1.5
    relaxed: string   // 1.75
  }

  // Letter spacing
  letterSpacing: {
    tight: string     // -0.025em
    normal: string    // 0
    wide: string      // 0.025em
  }
}

/**
 * Visual effects configuration
 */
export interface ThemeEffects {
  // Border radius values
  borderRadius: {
    small: string      // 0.25rem
    medium: string     // 0.5rem
    large: string      // 1rem
    extraLarge: string  // 1.5rem
    full: string       // 9999px
  }

  // Shadow definitions
  shadows: string[]

  // Gradient definitions
  gradients: string[]

  // Animation configurations
  animations: {
    duration: {
      fast: string      // 150ms
      normal: string    // 300ms
      slow: string      // 500ms
    }
    easing: {
      linear: string
      easeIn: string
      easeOut: string
      easeInOut: string
    }
  }

  // Blur effects
  blur: {
    small: string      // 4px
    medium: string     // 8px
    large: string      // 16px
    extraLarge: string  // 24px
  }
}

/**
 * Theme settings for behavior and interactions
 */
export interface ThemeSettings {
  // Opacity values
  disabledOpacity: number
  hoverOpacity: number
  activeOpacity: number

  // Transition durations
  transitionDuration: {
    fast: string
    normal: string
    slow: string
  }

  // Z-index scale for consistent layering
  zIndexScale: Record<string, number>
}

/**
 * Theme metadata for categorization and search
 */
export interface ThemeMetadata {
  createdAt?: string
  updatedAt?: string
  tags?: string[]
  category?: string
  license?: string
  source?: string
}

// =============================================================================
// LEGACY COMPATIBILITY INTERFACES
// =============================================================================

/**
 * Legacy theme interface for backward compatibility
 * @deprecated Use the new Theme interface instead
 */
export interface LegacyTheme {
  name: string
  displayName: string
  description: string
  isDark: boolean
  colors: {
    primary: string
    secondary: string
    accent: string
    background: string
    surface: string
    text: string
    textSecondary: string
    border: string
    success: string
    warning: string
    error: string
    info: string
  }
  typography: {
    fontFamily: string
    fontSize: {
      xs: string
      sm: string
      base: string
      lg: string
      xl: string
    }
    fontWeight?: {
      light: number
      normal: number
      medium: number
      bold: number
    }
    lineHeight?: {
      tight: number
      normal: number
      relaxed: number
    }
  }
  effects: {
    glow: boolean
    animation: boolean
    shadows: boolean
    transitions?: boolean
    gradients?: boolean
    backdrop?: boolean
  }
  isCustom?: boolean
}

// =============================================================================
// CUSTOM THEME INTERFACES
// =============================================================================

/**
 * Custom theme wrapper with additional metadata
 */
export interface CustomTheme {
  name: string
  theme: Theme
  createdAt: string
  modifiedAt?: string
  tags?: string[]
  preview?: string
  author?: string
  license?: string
}

/**
 * Theme service configuration
 */
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

/**
 * Generated CSS variables with metadata
 */
export interface GeneratedCSS {
  variables: string
  themeId: string
  generatedAt: string
  hash: string
  variableCount: number
  metadata: Record<string, string>
}

// =============================================================================
// THEME MANAGEMENT INTERFACES
// =============================================================================

/**
 * Theme validation result
 */
export interface ThemeValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

/**
 * Theme export/import data structure
 */
export interface ThemeExportData {
  themes: Theme[]
  customThemes?: CustomTheme[]
  version: string
  exportDate: string
  metadata?: {
    author?: string
    description?: string
    tags?: string[]
  }
}

/**
 * Theme import result with conflict resolution
 */
export interface ThemeImportResult {
  success: boolean
  importedThemes: Theme[]
  conflicts: ThemeConflict[]
  errors: string[]
}

/**
 * Theme conflict information
 */
export interface ThemeConflict {
  type: 'duplicate' | 'version' | 'incompatible'
  existing: Theme
  incoming: Theme
  resolution?: 'overwrite' | 'merge' | 'skip' | 'rename'
}

// =============================================================================
// UI COMPONENT INTERFACES
// =============================================================================

/**
 * Theme preview configuration
 */
export interface ThemePreview {
  themeName: string
  screenshot?: string
  colors: {
    primary: string
    secondary: string
    background: string
    text: string
    accent: string
  }
  components: {
    button: string
    card: string
    input: string
    terminal: string
  }
  size?: 'small' | 'medium' | 'large'
}

/**
 * Theme selector configuration
 */
export interface ThemeSelectorConfig {
  showCreateButton?: boolean
  showImportButton?: boolean
  showExportButton?: boolean
  compact?: boolean
  searchable?: boolean
  categorizable?: boolean
  previewEnabled?: boolean
}

/**
 * Theme editor state and configuration
 */
export interface ThemeEditorState {
  currentTheme: Theme
  selectedCategory: 'colors' | 'fonts' | 'effects' | 'settings' | 'custom'
  selectedField?: string
  previewMode: 'light' | 'dark' | 'auto'
  isPreviewing: boolean
  hasUnsavedChanges: boolean
  validationErrors: ThemeValidationError[]
}

/**
 * Theme validation error
 */
export interface ThemeValidationError {
  field: string
  message: string
  severity: 'error' | 'warning' | 'info'
  path?: string[]
}

// =============================================================================
// ANIMATION AND TRANSITION INTERFACES
// =============================================================================

/**
 * Theme transition configuration
 */
export interface ThemeTransition {
  type: 'fade' | 'slide' | 'zoom' | 'flip' | 'none'
  duration: number
  easing: string
  delay?: number
}

/**
 * Theme animation keyframes
 */
export interface ThemeAnimation {
  name: string
  keyframes: Record<string, Record<string, string>>
  duration: number
  easing: string
  iteration: number | 'infinite'
  direction?: 'normal' | 'reverse' | 'alternate'
}

// =============================================================================
// UTILITY INTERFACES
// =============================================================================

/**
 * Theme color palette generation
 */
export interface ColorPalette {
  name: string
  colors: string[]
  type: 'monochromatic' | 'analogous' | 'complementary' | 'triadic' | 'tetradic'
  base: string
  harmony?: string
}

/**
 * Theme variable definition
 */
export interface ThemeVariable {
  name: string
  value: string
  type: 'color' | 'size' | 'font' | 'animation' | 'other'
  category: string
  description?: string
  cssProperty?: string
}

/**
 * Theme preset for quick customization
 */
export interface ThemePreset {
  name: string
  description?: string
  colors?: Partial<ThemeColors>
  fonts?: Partial<ThemeFonts>
  effects?: Partial<ThemeEffects>
  settings?: Partial<ThemeSettings>
}

/**
 * Component-specific theming
 */
export interface ComponentTheme {
  name: string
  base: Record<string, string>
  variants: Record<string, Record<string, string>>
  sizes: Record<string, Record<string, string>>
  states?: Record<string, Record<string, string>>
}

// =============================================================================
// LEGACY SETTINGS INTERFACES
// =============================================================================

/**
 * Legacy theme settings for backward compatibility
 * @deprecated Use ThemeSettings instead
 */
export interface LegacyThemeSettings {
  currentTheme: string
  autoSwitch: boolean
  switchSchedule: {
    from: string
    to: string
  }
  transition: ThemeTransition
  customCSS: string
  enableAnimations: boolean
  reduceMotion: boolean
  highContrast: boolean
  fontSize: 'small' | 'medium' | 'large' | 'extra-large'
}

// =============================================================================
// TYPE GUARDS AND HELPERS
// =============================================================================

/**
 * Type guard for Theme interface
 */
export function isTheme(obj: any): obj is Theme {
  return obj &&
    typeof obj.id === 'string' &&
    typeof obj.name === 'string' &&
    obj.colors !== undefined &&
    obj.fonts !== undefined &&
    obj.effects !== undefined &&
    obj.settings !== undefined
}

/**
 * Type guard for LegacyTheme interface
 */
export function isLegacyTheme(obj: any): obj is LegacyTheme {
  return obj &&
    typeof obj.name === 'string' &&
    obj.colors !== undefined &&
    obj.typography !== undefined &&
    obj.effects !== undefined
}

/**
 * Type guard for CustomTheme interface
 */
export function isCustomTheme(obj: any): obj is CustomTheme {
  return obj &&
    typeof obj.name === 'string' &&
    obj.theme !== undefined &&
    typeof obj.createdAt === 'string'
}

// =============================================================================
// DEFAULT VALUES AND CONSTANTS
// =============================================================================

/**
 * Default theme colors
 */
export const DEFAULT_THEME_COLORS: ThemeColors = {
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
} as const

/**
 * Default theme fonts
 */
export const DEFAULT_THEME_FONTS: ThemeFonts = {
  families: {
    primary: '"Inter", system-ui, sans-serif',
    secondary: '"Roboto", Arial, sans-serif',
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
} as const

/**
 * Default theme effects
 */
export const DEFAULT_THEME_EFFECTS: ThemeEffects = {
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
} as const

/**
 * Default theme settings
 */
export const DEFAULT_THEME_SETTINGS: ThemeSettings = {
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
} as const

/**
 * Default theme service configuration
 */
export const DEFAULT_THEME_CONFIG: ThemeServiceConfig = {
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
} as const