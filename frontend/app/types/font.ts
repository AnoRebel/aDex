/**
 * Font-related type definitions for the eDEX-UI frontend
 * Provides comprehensive TypeScript interfaces for font configuration
 */

// ============================================================================
// Core Font Types
// ============================================================================

export interface FontConfiguration {
  id: string
  name: string
  description?: string
  family: string
  size: number
  weight: string | number
  lineHeight: number
  letterSpacing: number
  ligatures: boolean
  antialias: boolean
  hinting: FontHinting
  features: Record<string, boolean>
  isBuiltIn: boolean
  isDefault: boolean
  author?: string
  version?: string
  createdAt: string
  updatedAt: string
  metadata?: Record<string, string>
}

export type FontHinting = 'none' | 'slight' | 'medium' | 'full'

export type FontWeight =
  | '100' | '200' | '300' | '400' | '500' | '600' | '700' | '800' | '900'
  | 'thin' | 'extra-light' | 'light' | 'normal' | 'medium'
  | 'semi-bold' | 'bold' | 'extra-bold' | 'black'

export interface SystemFont {
  family: string
  displayName: string
  style: string
  weight: string
  isMonospace: boolean
  isInstalled: boolean
  filePath?: string
  format: FontFormat
  size?: number
  checksum?: string
  metadata: FontMetadata
}

export type FontFormat = 'ttf' | 'otf' | 'woff' | 'woff2' | 'eot' | 'svg' | 'unknown'

export interface FontMetadata {
  psName?: string
  version?: string
  copyright?: string
  trademark?: string
  manufacturer?: string
  designer?: string
  description?: string
  vendorUrl?: string
  designerUrl?: string
  license?: string
  licenseUrl?: string
  panose?: number[]
  vendorId?: string
  cssFamily?: string
  cssWeight?: string
  cssStyle?: string
  unicodeRange?: string[]
  supportedScripts?: string[]
  custom?: Record<string, string>
}

// ============================================================================
// Font Preview Types
// ============================================================================

export interface FontPreview {
  text: string
  sampleText: string
  fontSize: number
  lineHeight: number
  letterSpacing: number
  colors: FontPreviewColors
  showLineNumbers: boolean
  showCharacterInfo: boolean
  renderingHints: string[]
}

export interface FontPreviewColors {
  background: string
  foreground: string
  selection: string
  lineNumbers: string
  cursor: string
  accent: string
}

// ============================================================================
// Font Settings Types
// ============================================================================

export interface FontSettings {
  defaultFontId: string
  fallbackFont: string
  fontSize: number
  lineHeight: number
  letterSpacing: number
  enableLigatures: boolean
  enableAntialias: boolean
  hinting: FontHinting
  fontFeatureSettings: Record<string, boolean>
  allowCustomFonts: boolean
  fontDirectories: string[]
  autoDetectFonts: boolean
  lastScanTime: string
}

// ============================================================================
// Font Validation Types
// ============================================================================

export interface FontValidationResult {
  isValid: boolean
  errors: string[]
  warnings: string[]
  suggestions: string[]
  features: string[]
  limitations: string[]
}

// ============================================================================
// Font Import Types
// ============================================================================

export interface FontImportRequest {
  source: string // file path, URL, or base64 data
  name?: string
  family?: string
  weight?: string
  style?: string
  license?: string
  metadata?: Record<string, string>
  overwrite?: boolean
  validateOnly?: boolean
}

export interface FontImportResult {
  success: boolean
  font?: SystemFont
  configuration?: FontConfiguration
  errors: string[]
  warnings: string[]
  importPath?: string
  checksum?: string
}

// ============================================================================
// Font Metrics Types
// ============================================================================

export interface FontMetrics {
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

// ============================================================================
// Font Theme Integration Types
// ============================================================================

export interface FontThemeIntegration {
  configuration: FontConfiguration
  xtermTheme: XTermFontTheme
  cssVariables: Record<string, string>
  renderingConfig: RenderingConfig
}

export interface XTermFontTheme {
  fontFamily: string
  fontSize: number
  fontWeight: string | number
  lineHeight: number
  letterSpacing: number
}

export interface RenderingConfig {
  enableLigatures: boolean
  enableAntialias: boolean
  hinting: FontHinting
  features: Record<string, boolean>
  fallback: string[]
}

// ============================================================================
// Font UI Component Types
// ============================================================================

export interface FontSelectorOptions {
  showBuiltin?: boolean
  showCustom?: boolean
  showSystemFonts?: boolean
  allowImport?: boolean
  allowCreate?: boolean
  filterMonospace?: boolean
  showPreview?: boolean
  showMetrics?: boolean
  maxHeight?: number
  placeholder?: string
}

export interface FontEditorOptions {
  showAdvanced?: boolean
  showPreview?: boolean
  showValidation?: boolean
  allowUnsafeFeatures?: boolean
  showFeatureHints?: boolean
}

export interface FontPreviewOptions {
  sampleTexts?: string[]
  customSampleText?: string
  showLineNumbers?: boolean
  showCharacterInfo?: boolean
  highlightSyntax?: boolean
  backgroundColor?: string
  foregroundColor?: string
}

// ============================================================================
// Font Event Types
// ============================================================================

export interface FontEvent {
  type: string
  timestamp: string
  data: any
  source: string
}

export interface FontConfigurationChangedEvent extends FontEvent {
  type: 'font.configuration.changed'
  data: {
    configurationId: string
    name: string
    family: string
    changes: string[]
  }
}

export interface FontConfigurationCreatedEvent extends FontEvent {
  type: 'font.configuration.created'
  data: {
    configurationId: string
    name: string
    family: string
  }
}

export interface FontConfigurationDeletedEvent extends FontEvent {
  type: 'font.configuration.deleted'
  data: {
    configurationId: string
    name: string
  }
}

export interface FontSettingsChangedEvent extends FontEvent {
  type: 'font.settings.changed'
  data: {
    changes: string[]
    previousSettings: Partial<FontSettings>
    newSettings: FontSettings
  }
}

export interface FontScannedEvent extends FontEvent {
  type: 'font.system.scanned'
  data: {
    fontCount: number
    newFonts: string[]
    scanDuration: number
  }
}

export interface FontImportedEvent extends FontEvent {
  type: 'font.imported'
  data: {
    fontFamily: string
    displayName: string
    source: string
    configurationId?: string
  }
}

// ============================================================================
// Font Preset Types
// ============================================================================

export interface FontPreset {
  id: string
  name: string
  description: string
  category: FontPresetCategory
  configuration: FontConfiguration
  tags: string[]
  author?: string
  version?: string
  isRecommended?: boolean
}

export type FontPresetCategory =
  | 'programming'
  | 'accessibility'
  | 'performance'
  | 'aesthetic'
  | 'vintage'
  | 'modern'
  | 'custom'

export interface FontPresetCollection {
  id: string
  name: string
  description: string
  presets: FontPreset[]
  version: string
  author?: string
  lastUpdated: string
}

// ============================================================================
// Font Utility Types
// ============================================================================

export interface FontSearchOptions {
  query?: string
  families?: string[]
  weights?: string[]
  styles?: string[]
  formats?: FontFormat[]
  monospaceOnly?: boolean
  installedOnly?: boolean
  builtinOnly?: boolean
  customOnly?: boolean
  limit?: number
  offset?: number
}

export interface FontSearchResult {
  fonts: SystemFont[]
  configurations: FontConfiguration[]
  totalCount: number
  hasMore: boolean
  nextOffset?: number
}

export interface FontComparison {
  fonts: SystemFont[]
  metrics: FontMetrics[]
  differences: FontComparisonDifference[]
  recommendation?: string
}

export interface FontComparisonDifference {
  property: string
  font1: string | number
  font2: string | number
  significance: 'minor' | 'moderate' | 'major'
  impact: string
}

export interface FontRecommendation {
  fonts: RecommendedFont[]
  criteria: FontRecommendationCriteria
  confidence: number
  explanation: string
}

export interface RecommendedFont {
  font: SystemFont
  score: number
  reasons: string[]
  concerns: string[]
  compatibility: FontCompatibility
}

export interface FontRecommendationCriteria {
  useCase: FontUseCase
  preferences: FontPreferences
  constraints: FontConstraints
}

export type FontUseCase =
  | 'programming'
  | 'terminal'
  | 'editing'
  | 'presentation'
  | 'accessibility'
  | 'performance'

export interface FontPreferences {
  size: number
  weight: FontWeight
  lineHeight: number
  letterSpacing: number
  ligatures: boolean
  antialias: boolean
  hinting: FontHinting
}

export interface FontConstraints {
  maxSize?: number
  maxFileSize?: number
  supportedFormats?: FontFormat[]
  excludeFamilies?: string[]
  requireFeatures?: string[]
}

export interface FontCompatibility {
  platforms: string[]
  browsers: string[]
  renderers: string[]
  features: string[]
  limitations: string[]
  fallback: string[]
}

// ============================================================================
// Font Error Types
// ============================================================================

export interface FontError {
  code: string
  message: string
  details?: any
  configurationId?: string
  fontId?: string
  timestamp: string
  recoverable: boolean
  suggestions?: string[]
}

export interface FontValidationError extends FontError {
  field: string
  value: any
  constraint: string
}

export interface FontImportError extends FontError {
  source: string
  stage: 'validation' | 'parsing' | 'installation' | 'configuration'
  partialSuccess?: boolean
}

// ============================================================================
// Export Default Types
// ============================================================================

export type {
  // Core types
  FontConfiguration as Configuration,
  SystemFont as Font,
  FontSettings as Settings,

  // Preview types
  FontPreview as Preview,
  FontPreviewColors as PreviewColors,

  // Validation types
  FontValidationResult as ValidationResult,

  // Import types
  FontImportRequest as ImportRequest,
  FontImportResult as ImportResult,

  // Metrics types
  FontMetrics as Metrics,

  // UI types
  FontSelectorOptions as SelectorOptions,
  FontEditorOptions as EditorOptions,
  FontPreviewOptions as PreviewOptions,

  // Event types
  FontEvent as Event,
  FontConfigurationChangedEvent as ConfigurationChangedEvent,

  // Utility types
  FontSearchOptions as SearchOptions,
  FontSearchResult as SearchResult,
  FontComparison as Comparison,
  FontRecommendation as Recommendation,

  // Error types
  FontError as Error,
  FontValidationError as ValidationError
}

export default {
  // Core
  FontConfiguration,
  SystemFont,
  FontSettings,

  // Preview
  FontPreview,
  FontPreviewColors,

  // Validation
  FontValidationResult,

  // Import
  FontImportRequest,
  FontImportResult,

  // Metrics
  FontMetrics,

  // UI
  FontSelectorOptions,
  FontEditorOptions,
  FontPreviewOptions,

  // Events
  FontEvent,
  FontConfigurationChangedEvent,
  FontConfigurationCreatedEvent,
  FontConfigurationDeletedEvent,
  FontSettingsChangedEvent,
  FontScannedEvent,
  FontImportedEvent,

  // Presets
  FontPreset,
  FontPresetCategory,
  FontPresetCollection,

  // Utility
  FontSearchOptions,
  FontSearchResult,
  FontComparison,
  FontRecommendation,

  // Errors
  FontError,
  FontValidationError,
  FontImportError,

  // Enums
  FontHinting,
  FontWeight,
  FontFormat,
  FontPresetCategory,
  FontUseCase
}