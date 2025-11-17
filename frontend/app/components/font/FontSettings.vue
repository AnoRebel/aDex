<template>
  <div class="font-settings">
    <!-- Settings Header -->
    <div class="font-settings__header">
      <h3 class="text-lg font-medium text-text">Font Settings</h3>
      <div class="flex items-center gap-2">
        <button
          @click="resetToDefaults"
          class="text-sm text-text-secondary hover:text-text transition-colors"
        >
          Reset to Defaults
        </button>
        <button
          @click="exportSettings"
          class="text-sm text-text-secondary hover:text-text transition-colors"
        >
          Export
        </button>
        <button
          @click="importSettings"
          class="text-sm text-primary hover:text-primary-dark transition-colors"
        >
          Import
        </button>
      </div>
    </div>

    <!-- Settings Content -->
    <div class="font-settings__content">
      <!-- Default Font Selection -->
      <div class="font-settings__section">
        <h4 class="font-settings__section-title">Default Font</h4>
        <div class="space-y-3">
          <div class="font-settings__field">
            <label class="font-settings__label">Font Configuration</label>
            <select
              v-model="localSettings.defaultFontId"
              @change="handleSettingChange"
              class="font-settings__select"
            >
              <option value="">Select a font...</option>
              <optgroup v-if="builtinConfigurations.length > 0" label="Built-in Fonts">
                <option
                  v-for="config in builtinConfigurations"
                  :key="config.id"
                  :value="config.id"
                >
                  {{ config.name }} ({{ config.family }})
                </option>
              </optgroup>
              <optgroup v-if="customConfigurations.length > 0" label="Custom Fonts">
                <option
                  v-for="config in customConfigurations"
                  :key="config.id"
                  :value="config.id"
                >
                  {{ config.name }} ({{ config.family }})
                </option>
              </optgroup>
            </select>
          </div>

          <div class="font-settings__field">
            <label class="font-settings__label">Fallback Font</label>
            <input
              v-model="localSettings.fallbackFont"
              @input="handleSettingChange"
              type="text"
              placeholder="monospace"
              class="font-settings__input"
            />
          </div>
        </div>
      </div>

      <!-- Basic Settings -->
      <div class="font-settings__section">
        <h4 class="font-settings__section-title">Basic Settings</h4>
        <div class="space-y-3">
          <div class="font-settings__field">
            <label class="font-settings__label">Font Size (px)</label>
            <div class="flex items-center gap-3">
              <input
                v-model.number="localSettings.fontSize"
                @input="handleSettingChange"
                type="range"
                min="6"
                max="72"
                class="font-settings__range flex-1"
              />
              <input
                v-model.number="localSettings.fontSize"
                @input="handleSettingChange"
                type="number"
                min="6"
                max="72"
                class="font-settings__number w-16"
              />
            </div>
            <div class="text-xs text-text-secondary mt-1">
              {{ getFontSizeDescription(localSettings.fontSize) }}
            </div>
          </div>

          <div class="font-settings__field">
            <label class="font-settings__label">Line Height</label>
            <div class="flex items-center gap-3">
              <input
                v-model.number="localSettings.lineHeight"
                @input="handleSettingChange"
                type="range"
                min="0.8"
                max="3"
                step="0.1"
                class="font-settings__range flex-1"
              />
              <input
                v-model.number="localSettings.lineHeight"
                @input="handleSettingChange"
                type="number"
                min="0.8"
                max="3"
                step="0.1"
                class="font-settings__number w-16"
              />
            </div>
            <div class="text-xs text-text-secondary mt-1">
              {{ getLineHeightDescription(localSettings.lineHeight) }}
            </div>
          </div>

          <div class="font-settings__field">
            <label class="font-settings__label">Letter Spacing (px)</label>
            <div class="flex items-center gap-3">
              <input
                v-model.number="localSettings.letterSpacing"
                @input="handleSettingChange"
                type="range"
                min="-2"
                max="10"
                step="0.1"
                class="font-settings__range flex-1"
              />
              <input
                v-model.number="localSettings.letterSpacing"
                @input="handleSettingChange"
                type="number"
                min="-2"
                max="10"
                step="0.1"
                class="font-settings__number w-16"
              />
            </div>
            <div class="text-xs text-text-secondary mt-1">
              {{ getLetterSpacingDescription(localSettings.letterSpacing) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Rendering Settings -->
      <div class="font-settings__section">
        <h4 class="font-settings__section-title">Rendering</h4>
        <div class="space-y-3">
          <div class="font-settings__field">
            <label class="font-settings__label">Font Hinting</label>
            <select
              v-model="localSettings.hinting"
              @change="handleSettingChange"
              class="font-settings__select"
            >
              <option value="none">None</option>
              <option value="slight">Slight</option>
              <option value="medium">Medium</option>
              <option value="full">Full</option>
            </select>
            <div class="text-xs text-text-secondary mt-1">
              {{ getHintingDescription(localSettings.hinting) }}
            </div>
          </div>

          <div class="font-settings__checkboxes">
            <label class="font-settings__checkbox">
              <input
                v-model="localSettings.enableLigatures"
                @change="handleSettingChange"
                type="checkbox"
              />
              <span class="font-settings__checkbox-text">Enable Ligatures</span>
            </label>
            <div class="text-xs text-text-secondary ml-6">
              Combine character pairs for better typography (e.g., fi, fl)
            </div>

            <label class="font-settings__checkbox">
              <input
                v-model="localSettings.enableAntialias"
                @change="handleSettingChange"
                type="checkbox"
              />
              <span class="font-settings__checkbox-text">Enable Antialiasing</span>
            </label>
            <div class="text-xs text-text-secondary ml-6">
              Smooth font edges for better readability
            </div>
          </div>
        </div>
      </div>

      <!-- Font Features -->
      <div class="font-settings__section">
        <h4 class="font-settings__section-title">Font Features</h4>
        <div class="space-y-2">
          <label
            v-for="(enabled, feature) in commonFontFeatures"
            :key="feature"
            class="font-settings__checkbox"
          >
            <input
              :checked="enabled"
              @change="toggleFontFeature(feature, $event)"
              type="checkbox"
            />
            <span class="font-settings__checkbox-text">
              {{ getFontFeatureName(feature) }}
            </span>
            <div class="text-xs text-text-secondary ml-6">
              {{ getFontFeatureDescription(feature) }}
            </div>
          </label>
        </div>
      </div>

      <!-- Advanced Settings -->
      <div class="font-settings__section">
        <h4 class="font-settings__section-title">
          Advanced
          <button
            @click="showAdvanced = !showAdvanced"
            class="ml-2 text-sm text-text-secondary hover:text-text"
          >
            <Icon
              :name="showAdvanced ? 'carbon:chevron-up' : 'carbon:chevron-down'"
              class="w-4 h-4"
            />
          </button>
        </h4>

        <div v-if="showAdvanced" class="space-y-3">
          <div class="font-settings__checkboxes">
            <label class="font-settings__checkbox">
              <input
                v-model="localSettings.allowCustomFonts"
                @change="handleSettingChange"
                type="checkbox"
              />
              <span class="font-settings__checkbox-text">Allow Custom Fonts</span>
            </label>

            <label class="font-settings__checkbox">
              <input
                v-model="localSettings.autoDetectFonts"
                @change="handleSettingChange"
                type="checkbox"
              />
              <span class="font-settings__checkbox-text">Auto-detect System Fonts</span>
            </label>
          </div>

          <div class="font-settings__field">
            <label class="font-settings__label">Font Directories</label>
            <div class="space-y-2">
              <div
                v-for="(dir, index) in localSettings.fontDirectories"
                :key="index"
                class="flex items-center gap-2"
              >
                <input
                  v-model="localSettings.fontDirectories[index]"
                  @input="handleSettingChange"
                  type="text"
                  placeholder="/path/to/fonts"
                  class="font-settings__input flex-1"
                />
                <button
                  @click="removeFontDirectory(index)"
                  class="text-text-secondary hover:text-error transition-colors"
                >
                  <Icon name="carbon:trash-can" class="w-4 h-4" />
                </button>
              </div>
              <button
                @click="addFontDirectory"
                class="text-sm text-primary hover:text-primary-dark transition-colors"
              >
                + Add Directory
              </button>
            </div>
          </div>

          <div v-if="localSettings.lastScanTime" class="font-settings__field">
            <label class="font-settings__label">Last Font Scan</label>
            <div class="text-sm text-text-secondary">
              {{ formatDate(localSettings.lastScanTime) }}
            </div>
            <button
              @click="scanFonts"
              :disabled="isScanning"
              class="text-sm text-primary hover:text-primary-dark transition-colors mt-1"
            >
              {{ isScanning ? 'Scanning...' : 'Scan Now' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Preview Section -->
      <div class="font-settings__section">
        <h4 class="font-settings__section-title">Preview</h4>
        <FontPreview
          :configuration="currentConfiguration"
          :show-metrics="false"
          :show-validation="false"
          :max-height="200"
        />
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="font-settings__actions">
      <button
        @click="applySettings"
        :disabled="!hasChanges || isApplying"
        class="font-settings__action-btn font-settings__action-btn--primary"
      >
        <Icon v-if="isApplying" name="carbon:loading" class="w-4 h-4 animate-spin" />
        <Icon v-else name="carbon:checkmark" class="w-4 h-4" />
        {{ isApplying ? 'Applying...' : 'Apply Changes' }}
      </button>
      <button
        @click="cancelChanges"
        :disabled="!hasChanges"
        class="font-settings__action-btn"
      >
        Cancel
      </button>
    </div>

    <!-- Hidden file input for import/export -->
    <input
      ref="fileInput"
      type="file"
      accept=".json"
      style="display: none"
      @change="handleFileImport"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useFontConfiguration } from '~/composables/useFontConfiguration'
import FontPreview from './FontPreview.vue'
import type { FontConfiguration, FontSettings } from '~/types/font'

// Props
interface Props {
  settings?: FontSettings | null
}

const props = defineProps<Props>()

// Emits
const emit = defineEmits<{
  'settings-changed': [settings: FontSettings]
  'apply-settings': [settings: FontSettings]
}>()

// Composables
const {
  configurations,
  currentConfiguration,
  updateSettings,
  scanSystemFonts,
  resetToDefaults
} = useFontConfiguration()

// State
const localSettings = ref<FontSettings>({
  defaultFontId: 'jetbrains-mono',
  fallbackFont: 'monospace',
  fontSize: 14,
  lineHeight: 1.4,
  letterSpacing: 0,
  enableLigatures: true,
  enableAntialias: true,
  hinting: 'slight',
  fontFeatureSettings: {},
  allowCustomFonts: false,
  fontDirectories: ['/usr/share/fonts', '/usr/local/share/fonts', '~/.fonts'],
  autoDetectFonts: true,
  lastScanTime: ''
})

const showAdvanced = ref(false)
const hasChanges = ref(false)
const isApplying = ref(false)
const isScanning = ref(false)
const fileInput = ref<HTMLInputElement>()

// Common font features
const commonFontFeatures = computed(() => ({
  'liga': localSettings.value.fontFeatureSettings.liga || false,
  'dlig': localSettings.value.fontFeatureSettings.dlig || false,
  'zero': localSettings.value.fontFeatureSettings.zero || false,
  'onum': localSettings.value.fontFeatureSettings.onum || false,
  'kern': localSettings.value.fontFeatureSettings.kern || false
}))

// Computed
const builtinConfigurations = computed(() => {
  return Object.values(configurations.value).filter(config => config.isBuiltIn)
})

const customConfigurations = computed(() => {
  return Object.values(configurations.value).filter(config => !config.isBuiltIn)
})

// Methods
const handleSettingChange = () => {
  hasChanges.value = true
}

const applySettings = async () => {
  if (!hasChanges.value) return

  isApplying.value = true
  try {
    await updateSettings(localSettings.value)
    emit('apply-settings', localSettings.value)
    hasChanges.value = false
  } catch (err) {
    console.error('Failed to apply font settings:', err)
  } finally {
    isApplying.value = false
  }
}

const cancelChanges = () => {
  if (props.settings) {
    localSettings.value = { ...props.settings }
  }
  hasChanges.value = false
}

const resetToDefaults = async () => {
  try {
    const defaults = await resetToDefaults()
    localSettings.value = { ...defaults }
    hasChanges.value = true
  } catch (err) {
    console.error('Failed to reset to defaults:', err)
  }
}

const exportSettings = () => {
  const data = JSON.stringify(localSettings.value, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'font-settings.json'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

const importSettings = () => {
  fileInput.value?.click()
}

const handleFileImport = (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const imported = JSON.parse(e.target?.result as string) as FontSettings
      localSettings.value = { ...imported }
      hasChanges.value = true
    } catch (err) {
      console.error('Failed to import font settings:', err)
    }
  }
  reader.readAsText(file)
}

const toggleFontFeature = (feature: string, event: Event) => {
  const enabled = (event.target as HTMLInputElement).checked
  localSettings.value.fontFeatureSettings[feature] = enabled
  handleSettingChange()
}

const addFontDirectory = () => {
  localSettings.value.fontDirectories.push('')
  handleSettingChange()
}

const removeFontDirectory = (index: number) => {
  localSettings.value.fontDirectories.splice(index, 1)
  handleSettingChange()
}

const scanFonts = async () => {
  isScanning.value = true
  try {
    await scanSystemFonts()
  } catch (err) {
    console.error('Failed to scan fonts:', err)
  } finally {
    isScanning.value = false
  }
}

// Description helpers
const getFontSizeDescription = (size: number) => {
  if (size < 10) return 'Very small - may be difficult to read'
  if (size < 12) return 'Small - good for compact displays'
  if (size < 16) return 'Standard - comfortable for most users'
  if (size < 20) return 'Large - good for accessibility'
  return 'Very large - uses significant screen space'
}

const getLineHeightDescription = (lineHeight: number) => {
  if (lineHeight < 1.0) return 'Compact - lines may overlap'
  if (lineHeight < 1.2) return 'Tight - reduced spacing'
  if (lineHeight < 1.6) return 'Normal - comfortable reading'
  return 'Loose - increased spacing'
}

const getLetterSpacingDescription = (spacing: number) => {
  if (spacing < -0.5) return 'Negative - characters may overlap'
  if (spacing < 0) return 'Slightly negative - condensed appearance'
  if (spacing < 1) return 'Normal - standard spacing'
  if (spacing < 3) return 'Expanded - increased readability'
  return 'Very expanded - uses more horizontal space'
}

const getHintingDescription = (hinting: string) => {
  switch (hinting) {
    case 'none':
      return 'No hinting - smoothest appearance but may be blurry'
    case 'slight':
      return 'Slight hinting - balanced appearance and clarity'
    case 'medium':
      return 'Medium hinting - sharper appearance'
    case 'full':
      return 'Full hinting - sharpest but may appear distorted'
    default:
      return ''
  }
}

const getFontFeatureName = (feature: string) => {
  const names: Record<string, string> = {
    'liga': 'Standard Ligatures',
    'dlig': 'Discretionary Ligatures',
    'zero': 'Slashed Zero',
    'onum': 'Old-style Numerals',
    'kern': 'Kerning'
  }
  return names[feature] || feature
}

const getFontFeatureDescription = (feature: string) => {
  const descriptions: Record<string, string> = {
    'liga': 'Combine character pairs like fi, fl, ff',
    'dlig': 'Optional decorative ligatures',
    'zero': 'Use slashed zero to distinguish from letter O',
    'onum': 'Use old-style numerals with varying heights',
    'kern': 'Adjust spacing between character pairs'
  }
  return descriptions[feature] || ''
}

const formatDate = (dateString: string) => {
  try {
    return new Date(dateString).toLocaleString()
  } catch {
    return 'Unknown'
  }
}

// Watchers
watch(() => props.settings, (newSettings) => {
  if (newSettings) {
    localSettings.value = { ...newSettings }
    hasChanges.value = false
  }
}, { immediate: true })

watch(() => localSettings.value.defaultFontId, () => {
  handleSettingChange()
}, { deep: true })
</script>

<style scoped>
@reference "../../assets/css/main.css";

.font-settings {
  @apply bg-surface border border-border rounded-lg overflow-hidden;
}

.font-settings__header {
  @apply flex items-center justify-between p-4 border-b border-border bg-surface;
}

.font-settings__header h3 {
  @apply m-0;
}

.font-settings__content {
  @apply p-4 space-y-6 max-h-96 overflow-y-auto;
}

.font-settings__section {
  @apply space-y-3;
}

.font-settings__section-title {
  @apply text-base font-medium text-text mb-3 flex items-center;
}

.font-settings__field {
  @apply space-y-1;
}

.font-settings__label {
  @apply block text-sm font-medium text-text;
}

.font-settings__input,
.font-settings__select,
.font-settings__number {
  @apply w-full px-3 py-2 bg-surface-hover border border-border rounded-md text-text placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent;
}

.font-settings__range {
  @apply w-full;
}

.font-settings__checkboxes {
  @apply space-y-3;
}

.font-settings__checkbox {
  @apply flex items-start gap-2 cursor-pointer;
}

.font-settings__checkbox input[type="checkbox"] {
  @apply mt-1 rounded border-border text-primary focus:ring-primary;
}

.font-settings__checkbox-text {
  @apply text-sm text-text;
}

.font-settings__actions {
  @apply flex items-center justify-end gap-2 p-4 border-t border-border bg-surface;
}

.font-settings__action-btn {
  @apply flex items-center gap-2 px-4 py-2 bg-surface hover:bg-surface-hover text-text border border-border rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}

.font-settings__action-btn--primary {
  @apply bg-primary hover:bg-primary-dark text-white border-primary;
}

/* Range input styling */
.font-settings__range::-webkit-slider-track {
  @apply bg-surface-hover h-2 rounded-full;
}

.font-settings__range::-webkit-slider-thumb {
  @apply appearance-none w-4 h-4 bg-primary rounded-full cursor-pointer border-2 border-surface;
}

.font-settings__range::-moz-range-track {
  @apply bg-surface-hover h-2 rounded-full;
}

.font-settings__range::-moz-range-thumb {
  @apply w-4 h-4 bg-primary rounded-full cursor-pointer border-2 border-surface;
}

.font-settings__range::-webkit-slider-thumb:hover {
  @apply bg-primary-dark;
}

.font-settings__range::-moz-range-thumb:hover {
  @apply bg-primary-dark;
}
</style>