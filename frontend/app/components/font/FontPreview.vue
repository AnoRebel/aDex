<template>
  <div class="font-preview">
    <!-- Preview Header -->
    <div class="font-preview__header">
      <div class="font-preview__title">
        <h3 class="text-lg font-medium text-text">Font Preview</h3>
        <p v-if="configuration" class="text-sm text-text-secondary">
          {{ configuration.name }} • {{ configuration.family }}
        </p>
      </div>
      <div class="font-preview__controls">
        <!-- Size Control -->
        <div class="font-preview__control">
          <label class="text-xs text-text-secondary">Size</label>
          <div class="flex items-center gap-1">
            <button
              @click="decreaseSize"
              :disabled="previewSettings.fontSize <= 6"
              class="font-preview__control-btn"
            >
              <Icon name="carbon:subtract" class="w-3 h-3" />
            </button>
            <input
              v-model.number="previewSettings.fontSize"
              type="number"
              min="6"
              max="72"
              class="font-preview__control-input w-12"
            />
            <button
              @click="increaseSize"
              :disabled="previewSettings.fontSize >= 72"
              class="font-preview__control-btn"
            >
              <Icon name="carbon:add" class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Line Height Control -->
        <div class="font-preview__control">
          <label class="text-xs text-text-secondary">Line Height</label>
          <div class="flex items-center gap-1">
            <button
              @click="decreaseLineHeight"
              :disabled="previewSettings.lineHeight <= 0.8"
              class="font-preview__control-btn"
            >
              <Icon name="carbon:subtract" class="w-3 h-3" />
            </button>
            <input
              v-model.number="previewSettings.lineHeight"
              type="number"
              min="0.8"
              max="3"
              step="0.1"
              class="font-preview__control-input w-14"
            />
            <button
              @click="increaseLineHeight"
              :disabled="previewSettings.lineHeight >= 3"
              class="font-preview__control-btn"
            >
              <Icon name="carbon:add" class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Letter Spacing Control -->
        <div class="font-preview__control">
          <label class="text-xs text-text-secondary">Letter Spacing</label>
          <div class="flex items-center gap-1">
            <button
              @click="decreaseLetterSpacing"
              :disabled="previewSettings.letterSpacing <= -2"
              class="font-preview__control-btn"
            >
              <Icon name="carbon:subtract" class="w-3 h-3" />
            </button>
            <input
              v-model.number="previewSettings.letterSpacing"
              type="number"
              min="-2"
              max="10"
              step="0.1"
              class="font-preview__control-input w-12"
            />
            <button
              @click="increaseLetterSpacing"
              :disabled="previewSettings.letterSpacing >= 10"
              class="font-preview__control-btn"
            >
              <Icon name="carbon:add" class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Toggle Controls -->
        <div class="flex items-center gap-2">
          <label class="flex items-center gap-1 text-sm text-text-secondary cursor-pointer">
            <input
              v-model="previewSettings.showLineNumbers"
              type="checkbox"
              class="rounded"
            />
            Line Numbers
          </label>
          <label class="flex items-center gap-1 text-sm text-text-secondary cursor-pointer">
            <input
              v-model="previewSettings.highlightSyntax"
              type="checkbox"
              class="rounded"
            />
            Syntax Highlight
          </label>
        </div>
      </div>
    </div>

    <!-- Sample Text Tabs -->
    <div class="font-preview__tabs">
      <button
        v-for="tab in previewTabs"
        :key="tab.key"
        @click="activeTab = tab.key"
        :class="[
          'font-preview__tab',
          { 'font-preview__tab--active': activeTab === tab.key }
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Preview Content -->
    <div class="font-preview__content" :style="contentStyle">
      <!-- Loading State -->
      <div v-if="isLoading" class="font-preview__loading">
        <div class="flex items-center justify-center py-8">
          <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
          <span class="ml-2 text-text-secondary">Loading preview...</span>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="font-preview__error">
        <div class="flex items-center justify-center py-8 text-center">
          <Icon name="carbon:warning" class="w-6 h-6 text-error mb-2" />
          <p class="text-error mb-2">{{ error }}</p>
        </div>
      </div>

      <!-- Preview Display -->
      <div v-else class="font-preview__display">
        <div
          v-if="previewSettings.showLineNumbers"
          class="font-preview__line-numbers"
          :style="lineNumbersStyle"
        >
          <div
            v-for="n in lineCount"
            :key="n"
            class="font-preview__line-number"
          >
            {{ n }}
          </div>
        </div>

        <div
          class="font-preview__text"
          :class="{
            'font-preview__text--syntax': previewSettings.highlightSyntax,
            'font-preview__text--line-numbers': previewSettings.showLineNumbers
          }"
          :style="textStyle"
        >
          <template v-if="activeTab === 'alphabet'">
            <div class="font-preview__section">
              <h4 class="font-preview__section-title">Uppercase</h4>
              <div class="font-preview__alphabet">ABCDEFGHIJKLMNOPQRSTUVWXYZ</div>
            </div>
            <div class="font-preview__section">
              <h4 class="font-preview__section-title">Lowercase</h4>
              <div class="font-preview__alphabet">abcdefghijklmnopqrstuvwxyz</div>
            </div>
            <div class="font-preview__section">
              <h4 class="font-preview__section-title">Numbers</h4>
              <div class="font-preview__alphabet">0123456789</div>
            </div>
            <div class="font-preview__section">
              <h4 class="font-preview__section-title">Symbols</h4>
              <div class="font-preview__alphabet">!@#$%^&*()_+-=[]{}|;:'",./<>?</div>
            </div>
          </template>

          <template v-else-if="activeTab === 'sample'">
            <pre class="font-preview__code"><code>{{ getSampleText() }}</code></pre>
          </template>

          <template v-else-if="activeTab === 'programming'">
            <pre class="font-preview__code"><code>{{ getProgrammingSample() }}</code></pre>
          </template>

          <template v-else-if="activeTab === 'comparison'">
            <div class="font-preview__comparison">
              <div class="font-preview__comparison-column">
                <h4 class="font-preview__comparison-title">Regular Text</h4>
                <p class="font-preview__comparison-text">
                  The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs.
                </p>
              </div>
              <div class="font-preview__comparison-column">
                <h4 class="font-preview__comparison-title">Mixed Content</h4>
                <p class="font-preview__comparison-text">
                  function helloWorld() { console.log("Hello, 世界! 🌍"); return 42; }
                </p>
              </div>
              <div class="font-preview__comparison-column">
                <h4 class="font-preview__comparison-title">Technical</h4>
                <p class="font-preview__comparison-text">
                  λf.(λx.f(x x))(λx.f(x x)) → Y combinator → ∀x∃y[P(x,y) → Q(x)]
                </p>
              </div>
            </div>
          </template>

          <template v-else-if="activeTab === 'custom'">
            <div class="font-preview__custom">
              <textarea
                v-model="customText"
                placeholder="Enter your custom preview text here..."
                class="font-preview__custom-input"
                :style="textStyle"
              ></textarea>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Font Metrics -->
    <div v-if="showMetrics && metrics" class="font-preview__metrics">
      <h4 class="font-preview__metrics-title">Font Metrics</h4>
      <div class="font-preview__metrics-grid">
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Family:</span>
          <span class="font-preview__metric-value">{{ metrics.family }}</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Size:</span>
          <span class="font-preview__metric-value">{{ metrics.size }}px</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Weight:</span>
          <span class="font-preview__metric-value">{{ metrics.weight }}</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Monospace:</span>
          <span class="font-preview__metric-value">{{ metrics.isMonospace ? 'Yes' : 'No' }}</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Ascent:</span>
          <span class="font-preview__metric-value">{{ metrics.ascent.toFixed(1) }}px</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Descent:</span>
          <span class="font-preview__metric-value">{{ metrics.descent.toFixed(1) }}px</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Line Gap:</span>
          <span class="font-preview__metric-value">{{ metrics.lineGap.toFixed(1) }}px</span>
        </div>
        <div class="font-preview__metric">
          <span class="font-preview__metric-label">Avg Width:</span>
          <span class="font-preview__metric-value">{{ metrics.avgCharWidth.toFixed(1) }}px</span>
        </div>
      </div>
    </div>

    <!-- Validation Results -->
    <div v-if="showValidation && validationResult" class="font-preview__validation">
      <h4 class="font-preview__validation-title">Validation Results</h4>

      <!-- Errors -->
      <div v-if="validationResult.errors.length > 0" class="font-preview__validation-errors">
        <div class="flex items-center gap-2 mb-2">
          <Icon name="carbon:warning-alt" class="w-4 h-4 text-error" />
          <span class="text-sm font-medium text-error">Errors ({{ validationResult.errors.length }})</span>
        </div>
        <ul class="space-y-1">
          <li v-for="error in validationResult.errors" :key="error" class="text-sm text-error ml-6">
            • {{ error }}
          </li>
        </ul>
      </div>

      <!-- Warnings -->
      <div v-if="validationResult.warnings.length > 0" class="font-preview__validation-warnings">
        <div class="flex items-center gap-2 mb-2">
          <Icon name="carbon:warning" class="w-4 h-4 text-warning" />
          <span class="text-sm font-medium text-warning">Warnings ({{ validationResult.warnings.length }})</span>
        </div>
        <ul class="space-y-1">
          <li v-for="warning in validationResult.warnings" :key="warning" class="text-sm text-warning ml-6">
            • {{ warning }}
          </li>
        </ul>
      </div>

      <!-- Features -->
      <div v-if="validationResult.features.length > 0" class="font-preview__validation-features">
        <div class="flex items-center gap-2 mb-2">
          <Icon name="carbon:checkmark" class="w-4 h-4 text-success" />
          <span class="text-sm font-medium text-success">Features ({{ validationResult.features.length }})</span>
        </div>
        <div class="flex flex-wrap gap-1 ml-6">
          <span
            v-for="feature in validationResult.features"
            :key="feature"
            class="text-xs bg-success bg-opacity-10 text-success px-2 py-1 rounded"
          >
            {{ feature }}
          </span>
        </div>
      </div>

      <!-- Suggestions -->
      <div v-if="validationResult.suggestions.length > 0" class="font-preview__validation-suggestions">
        <div class="flex items-center gap-2 mb-2">
          <Icon name="carbon:idea" class="w-4 h-4 text-info" />
          <span class="text-sm font-medium text-info">Suggestions ({{ validationResult.suggestions.length }})</span>
        </div>
        <ul class="space-y-1">
          <li v-for="suggestion in validationResult.suggestions" :key="suggestion" class="text-sm text-info ml-6">
            • {{ suggestion }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useFontConfiguration } from '~/composables/useFontConfiguration'
import type { FontConfiguration, FontValidationResult, FontMetrics } from '~/types/font'

// Props
interface Props {
  configuration?: FontConfiguration | null
  showMetrics?: boolean
  showValidation?: boolean
  maxHeight?: number
}

const props = withDefaults(defineProps<Props>(), {
  showMetrics: true,
  showValidation: true,
  maxHeight: 600
})

// Emits
const emit = defineEmits<{
  'settings-changed': [settings: any]
}>()

// Composables
const { validateConfiguration, getFontMetrics } = useFontConfiguration()

// State
const isLoading = ref(false)
const error = ref<string | null>(null)
const activeTab = ref('sample')
const customText = ref('')
const validationResult = ref<FontValidationResult | null>(null)
const metrics = ref<FontMetrics | null>(null)

const previewSettings = ref({
  fontSize: 14,
  lineHeight: 1.4,
  letterSpacing: 0,
  showLineNumbers: false,
  highlightSyntax: false
})

const previewTabs = [
  { key: 'alphabet', label: 'Alphabet' },
  { key: 'sample', label: 'Sample' },
  { key: 'programming', label: 'Programming' },
  { key: 'comparison', label: 'Comparison' },
  { key: 'custom', label: 'Custom' }
]

// Computed
const lineCount = computed(() => {
  const text = getCurrentText()
  return text.split('\n').length
})

const contentStyle = computed(() => ({
  maxHeight: props.maxHeight + 'px'
}))

const textStyle = computed(() => {
  if (!props.configuration) return {}

  return {
    fontFamily: `'${props.configuration.family}', monospace`,
    fontSize: `${previewSettings.value.fontSize}px`,
    fontWeight: props.configuration.weight,
    lineHeight: previewSettings.value.lineHeight.toString(),
    letterSpacing: `${previewSettings.value.letterSpacing}px`,
    fontFeatureSettings: props.configuration.ligatures ? '"liga", "dlig"' : 'normal'
  }
})

const lineNumbersStyle = computed(() => ({
  fontFamily: `'${props.configuration?.family || 'monospace'}', monospace`,
  fontSize: `${previewSettings.value.fontSize}px`,
  lineHeight: previewSettings.value.lineHeight.toString(),
  letterSpacing: `${previewSettings.value.letterSpacing}px`
}))

// Methods
const decreaseSize = () => {
  if (previewSettings.value.fontSize > 6) {
    previewSettings.value.fontSize--
    emitChanges()
  }
}

const increaseSize = () => {
  if (previewSettings.value.fontSize < 72) {
    previewSettings.value.fontSize++
    emitChanges()
  }
}

const decreaseLineHeight = () => {
  if (previewSettings.value.lineHeight > 0.8) {
    previewSettings.value.lineHeight -= 0.1
    emitChanges()
  }
}

const increaseLineHeight = () => {
  if (previewSettings.value.lineHeight < 3) {
    previewSettings.value.lineHeight += 0.1
    emitChanges()
  }
}

const decreaseLetterSpacing = () => {
  if (previewSettings.value.letterSpacing > -2) {
    previewSettings.value.letterSpacing -= 0.1
    emitChanges()
  }
}

const increaseLetterSpacing = () => {
  if (previewSettings.value.letterSpacing < 10) {
    previewSettings.value.letterSpacing += 0.1
    emitChanges()
  }
}

const getCurrentText = () => {
  switch (activeTab.value) {
    case 'alphabet':
      return [
        'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
        'abcdefghijklmnopqrstuvwxyz',
        '0123456789',
        '!@#$%^&*()_+-=[]{}|;:\'",./<>?'
      ].join('\n')
    case 'sample':
      return getSampleText()
    case 'programming':
      return getProgrammingSample()
    case 'comparison':
      return [
        'The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs.',
        'function helloWorld() { console.log("Hello, 世界! 🌍"); return 42; }',
        'λf.(λx.f(x x))(λx.f(x x)) → Y combinator → ∀x∃y[P(x,y) → Q(x)]'
      ].join('\n')
    case 'custom':
      return customText.value
    default:
      return ''
  }
}

const getSampleText = () => {
  return `The quick brown fox jumps over the lazy dog.
Pack my box with five dozen liquor jugs.

1234567890 !@#$%^&*()_+-=[]{}|;:'",./<>?

ABCDEFGHIJKLMNOPQRSTUVWXYZ
abcdefghijklmnopqrstuvwxyz

Special characters: é à ü ñ ç ö å æ œ ß
Unicode symbols: α β γ δ ε ζ η θ ≈ ≠ ≤ ≥ ∞ ∑ ∏
Emojis: 😊 🎉 🔥 🚀 💡 ⚡ 🌟

Programming constructs:
function helloWorld() {
  const message = "Hello, World! 👋";
  console.log(message);
  return true;
}

if (condition) {
  execute();
} else {
  handleError();
}

// This is a comment
/* Multi-line
   comment */
`
}

const getProgrammingSample = () => {
  return `import React, { useState, useEffect } from 'react';

interface User {
  id: number;
  name: string;
  email: string;
  avatar?: string;
}

const UserProfile: React.FC<{ userId: number }> = ({ userId }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        setLoading(true);
        const response = await fetch(\`/api/users/\${userId}\`);

        if (!response.ok) {
          throw new Error(\`HTTP error! status: \${response.status}\`);
        }

        const userData: User = await response.json();
        setUser(userData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, [userId]);

  if (loading) return <div>Loading profile...</div>;
  if (error) return <div className="error">Error: {error}</div>;
  if (!user) return <div>User not found</div>;

  return (
    <div className="user-profile">
      <img src={user.avatar || '/default-avatar.png'} alt={user.name} />
      <h2>{user.name}</h2>
      <p>{user.email}</p>
    </div>
  );
};

export default UserProfile;`
}

const emitChanges = () => {
  emit('settings-changed', { ...previewSettings.value })
}

const loadValidation = async () => {
  if (props.configuration && props.showValidation) {
    try {
      validationResult.value = await validateConfiguration(props.configuration)
    } catch (err) {
      console.error('Failed to validate font configuration:', err)
    }
  }
}

const loadMetrics = async () => {
  if (props.configuration && props.showMetrics) {
    try {
      metrics.value = await getFontMetrics(
        props.configuration.family,
        previewSettings.value.fontSize
      )
    } catch (err) {
      console.error('Failed to load font metrics:', err)
    }
  }
}

// Watchers
watch(() => props.configuration, () => {
  loadValidation()
  loadMetrics()
}, { immediate: true })

watch(() => previewSettings.value.fontSize, () => {
  loadMetrics()
})

watch(() => activeTab, () => {
  if (activeTab.value === 'custom' && !customText.value) {
    customText.value = getSampleText()
  }
})

// Lifecycle
onMounted(() => {
  loadValidation()
  loadMetrics()
})
</script>

<style scoped>
@reference "../../assets/css/main.css";

.font-preview {
  @apply bg-surface border border-border rounded-lg overflow-hidden;
}

.font-preview__header {
  @apply flex items-center justify-between p-4 border-b border-border bg-surface;
}

.font-preview__title h3 {
  @apply m-0;
}

.font-preview__controls {
  @apply flex items-center gap-4;
}

.font-preview__control {
  @apply flex flex-col items-center gap-1;
}

.font-preview__control label {
  @apply text-xs text-text-secondary;
}

.font-preview__control-btn {
  @apply p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-text transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}

.font-preview__control-input {
  @apply w-full text-center bg-surface border border-border rounded px-1 py-0.5 text-sm text-text focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary;
}

.font-preview__tabs {
  @apply flex border-b border-border bg-surface;
}

.font-preview__tab {
  @apply px-4 py-2 text-sm font-medium text-text-secondary hover:text-text transition-colors border-b-2 border-transparent;
}

.font-preview__tab--active {
  @apply text-primary border-primary;
}

.font-preview__content {
  @apply overflow-y-auto;
}

.font-preview__loading,
.font-preview__error {
  @apply flex items-center justify-center py-8;
}

.font-preview__display {
  @apply flex relative;
}

.font-preview__line-numbers {
  @apply text-right text-text-secondary pr-3 select-none border-r border-border;
  flex-shrink: 0;
}

.font-preview__line-number {
  @apply leading-relaxed;
}

.font-preview__text {
  @apply flex-1 overflow-x-auto;
}

.font-preview__text--line-numbers {
  @apply ml-3;
}

.font-preview__section {
  @apply mb-6;
}

.font-preview__section:last-child {
  @apply mb-0;
}

.font-preview__section-title {
  @apply text-sm font-medium text-text mb-2;
}

.font-preview__alphabet {
  @apply text-text leading-relaxed break-all;
}

.font-preview__code {
  @apply m-0 p-4 bg-surface-hover rounded text-sm leading-relaxed overflow-x-auto;
}

.font-preview__comparison {
  @apply space-y-6 p-4;
}

.font-preview__comparison-column {
  @apply space-y-2;
}

.font-preview__comparison-title {
  @apply text-sm font-medium text-text mb-2;
}

.font-preview__comparison-text {
  @apply text-text leading-relaxed break-words;
}

.font-preview__custom {
  @apply p-4;
}

.font-preview__custom-input {
  @apply w-full h-64 p-3 bg-surface-hover border border-border rounded resize-none focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent;
}

.font-preview__metrics {
  @apply p-4 border-t border-border bg-surface;
}

.font-preview__metrics-title {
  @apply text-sm font-medium text-text mb-3;
}

.font-preview__metrics-grid {
  @apply grid grid-cols-2 md:grid-cols-4 gap-3;
}

.font-preview__metric {
  @apply flex justify-between items-center text-sm;
}

.font-preview__metric-label {
  @apply text-text-secondary;
}

.font-preview__metric-value {
  @apply text-text font-medium;
}

.font-preview__validation {
  @apply p-4 border-t border-border bg-surface space-y-4;
}

.font-preview__validation-title {
  @apply text-sm font-medium text-text mb-3;
}

.font-preview__validation-errors,
.font-preview__validation-warnings,
.font-preview__validation-features,
.font-preview__validation-suggestions {
  @apply space-y-2;
}

/* Syntax highlighting classes (simplified) */
.font-preview__text--syntax :deep(.keyword) {
  @apply text-purple-600 dark:text-purple-400;
}

.font-preview__text--syntax :deep(.string) {
  @apply text-green-600 dark:text-green-400;
}

.font-preview__text--syntax :deep(.number) {
  @apply text-blue-600 dark:text-blue-400;
}

.font-preview__text--syntax :deep(.comment) {
  @apply text-gray-500 dark:text-gray-400 italic;
}

.font-preview__text--syntax :deep(.function) {
  @apply text-blue-700 dark:text-blue-300 font-medium;
}
</style>