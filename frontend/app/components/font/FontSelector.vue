<template>
  <div class="font-selector">
    <!-- Search Bar -->
    <div class="font-selector__search">
      <div class="relative">
        <Icon name="carbon:search" class="absolute left-3 top-1/2 transform -translate-y-1/2 text-text-secondary w-4 h-4" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search fonts..."
          class="font-selector__search-input"
          @input="handleSearch"
        />
        <button
          v-if="searchQuery"
          @click="clearSearch"
          class="absolute right-3 top-1/2 transform -translate-y-1/2 text-text-secondary hover:text-text transition-colors"
        >
          <Icon name="carbon:close" class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Filter Tabs -->
    <div class="font-selector__tabs" v-if="showTabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        @click="activeTab = tab.key"
        :class="[
          'font-selector__tab',
          { 'font-selector__tab--active': activeTab === tab.key }
        ]"
      >
        <Icon :name="tab.icon" class="w-4 h-4" />
        {{ tab.label }}
        <span v-if="tab.count !== undefined" class="font-selector__tab-count">
          {{ tab.count }}
        </span>
      </button>
    </div>

    <!-- Font List -->
    <div class="font-selector__list" :style="{ maxHeight: maxHeight + 'px' }">
      <!-- Loading State -->
      <div v-if="isLoading" class="font-selector__loading">
        <div class="flex items-center justify-center py-8">
          <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
          <span class="ml-2 text-text-secondary">Loading fonts...</span>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="font-selector__error">
        <div class="flex items-center justify-center py-8 text-center">
          <Icon name="carbon:warning" class="w-6 h-6 text-error mb-2" />
          <p class="text-error mb-2">{{ error }}</p>
          <button @click="retry" class="text-primary hover:text-primary-dark text-sm">
            Retry
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredFonts.length === 0" class="font-selector__empty">
        <div class="text-center py-8">
          <Icon name="carbon:text-font" class="w-12 h-12 text-text-secondary mx-auto mb-4" />
          <p class="text-text-secondary mb-2">
            {{ searchQuery ? 'No fonts found matching your search' : 'No fonts available' }}
          </p>
          <button
            v-if="allowImport && !searchQuery"
            @click="$emit('import-font')"
            class="text-primary hover:text-primary-dark text-sm"
          >
            Import Font
          </button>
        </div>
      </div>

      <!-- Font Items -->
      <div v-else class="font-selector__items">
        <div
          v-for="font in paginatedFonts"
          :key="font.id || font.family"
          :class="[
            'font-selector__item',
            { 'font-selector__item--selected': isSelected(font) }
          ]"
          @click="selectFont(font)"
          @mouseenter="hoveredFont = font"
          @mouseleave="hoveredFont = null"
        >
          <!-- Font Preview -->
          <div
            class="font-selector__preview"
            :style="getPreviewStyle(font)"
          >
            {{ getPreviewText(font) }}
          </div>

          <!-- Font Info -->
          <div class="font-selector__info">
            <div class="font-selector__name">
              {{ font.name || font.displayName }}
              <span v-if="font.isBuiltIn" class="font-selector__builtin-badge">
                Built-in
              </span>
              <span v-if="font.isDefault" class="font-selector__default-badge">
                Default
              </span>
            </div>
            <div class="font-selector__details">
              {{ font.family }}
              <span v-if="font.weight && font.weight !== '400'" class="text-text-secondary">
                • {{ font.weight }}
              </span>
              <span v-if="font.style && font.style !== 'normal'" class="text-text-secondary">
                • {{ font.style }}
              </span>
            </div>
            <div v-if="font.description" class="font-selector__description">
              {{ font.description }}
            </div>
          </div>

          <!-- Font Actions -->
          <div class="font-selector__actions">
            <button
              v-if="allowDuplicate && !font.isBuiltIn"
              @click.stop="duplicateFont(font)"
              class="font-selector__action"
              title="Duplicate font"
            >
              <Icon name="carbon:copy" class="w-4 h-4" />
            </button>
            <button
              v-if="allowEdit && !font.isBuiltIn"
              @click.stop="editFont(font)"
              class="font-selector__action"
              title="Edit font"
            >
              <Icon name="carbon:edit" class="w-4 h-4" />
            </button>
            <button
              v-if="allowDelete && !font.isBuiltIn && !isDefaultFont(font)"
              @click.stop="deleteFont(font)"
              class="font-selector__action"
              title="Delete font"
            >
              <Icon name="carbon:trash-can" class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="showPagination && totalPages > 1" class="font-selector__pagination">
      <button
        @click="previousPage"
        :disabled="currentPage === 1"
        class="font-selector__pagination-btn"
      >
        <Icon name="carbon:chevron-left" class="w-4 h-4" />
      </button>
      <span class="font-selector__pagination-info">
        {{ currentPage }} / {{ totalPages }}
      </span>
      <button
        @click="nextPage"
        :disabled="currentPage === totalPages"
        class="font-selector__pagination-btn"
      >
        <Icon name="carbon:chevron-right" class="w-4 h-4" />
      </button>
    </div>

    <!-- Action Buttons -->
    <div v-if="showActions" class="font-selector__actions-bar">
      <button
        v-if="allowImport"
        @click="$emit('import-font')"
        class="font-selector__action-btn font-selector__action-btn--primary"
      >
        <Icon name="carbon:download" class="w-4 h-4" />
        Import Font
      </button>
      <button
        v-if="allowCreate"
        @click="$emit('create-font')"
        class="font-selector__action-btn"
      >
        <Icon name="carbon:add" class="w-4 h-4" />
        Create Font
      </button>
      <button
        v-if="allowScan"
        @click="scanFonts"
        :disabled="isScanning"
        class="font-selector__action-btn"
      >
        <Icon :name="isScanning ? 'carbon:loading' : 'carbon:scan'" class="w-4 h-4" />
        {{ isScanning ? 'Scanning...' : 'Scan System Fonts' }}
      </button>
    </div>

    <!-- Hover Preview (Floating) -->
    <Teleport to="body">
      <div
        v-if="showHoverPreview && hoveredFont"
        ref="hoverPreview"
        class="font-selector__hover-preview"
        :style="hoverPreviewStyle"
      >
        <div class="font-selector__hover-preview-content">
          <div
            class="font-selector__hover-preview-text"
            :style="getPreviewStyle(hoveredFont, true)"
          >
            {{ getExtendedPreviewText(hoveredFont) }}
          </div>
          <div class="font-selector__hover-preview-info">
            <div class="font-mono text-xs text-text-secondary">
              {{ hoveredFont.family }} • {{ hoveredFont.weight || '400' }}
            </div>
            <div v-if="hoveredFont.description" class="text-xs text-text-secondary mt-1">
              {{ hoveredFont.description }}
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useFontConfiguration } from '~/composables/useFontConfiguration'
import type { FontConfiguration, SystemFont } from '~/types/font'

// Props
interface Props {
  modelValue?: string | FontConfiguration | null
  maxHeight?: number
  showTabs?: boolean
  showBuiltin?: boolean
  showCustom?: boolean
  showSystemFonts?: boolean
  allowImport?: boolean
  allowCreate?: boolean
  allowEdit?: boolean
  allowDelete?: boolean
  allowDuplicate?: boolean
  allowScan?: boolean
  filterMonospace?: boolean
  showPreview?: boolean
  showPagination?: boolean
  showActions?: boolean
  pageSize?: number
  placeholder?: string
  previewText?: string
}

const props = withDefaults(defineProps<Props>(), {
  maxHeight: 400,
  showTabs: true,
  showBuiltin: true,
  showCustom: true,
  showSystemFonts: false,
  allowImport: false,
  allowCreate: false,
  allowEdit: false,
  allowDelete: false,
  allowDuplicate: false,
  allowScan: false,
  filterMonospace: true,
  showPreview: true,
  showPagination: true,
  showActions: true,
  pageSize: 20,
  placeholder: 'Search fonts...',
  previewText: 'The quick brown fox jumps over the lazy dog 123'
})

// Emits
const emit = defineEmits<{
  'update:modelValue': [value: string | FontConfiguration | null]
  'font-selected': [font: FontConfiguration | SystemFont]
  'font-duplicate': [font: FontConfiguration | SystemFont]
  'font-edit': [font: FontConfiguration | SystemFont]
  'font-delete': [font: FontConfiguration | SystemFont]
  'import-font': []
  'create-font': []
  'scan-fonts': []
}>()

// Composables
const {
  configurations,
  systemFonts,
  monospaceFonts,
  currentConfiguration,
  isLoading,
  error,
  scanSystemFonts
} = useFontConfiguration()

// State
const searchQuery = ref('')
const activeTab = ref('all')
const currentPage = ref(1)
const hoveredFont = ref<FontConfiguration | SystemFont | null>(null)
const isScanning = ref(false)
const hoverPreview = ref<HTMLElement>()
const showHoverPreview = ref(false)
const hoverPreviewPosition = ref({ x: 0, y: 0 })

// Computed
const tabs = computed(() => [
  {
    key: 'all',
    label: 'All',
    icon: 'carbon:text-font',
    count: allFonts.value.length
  },
  {
    key: 'builtin',
    label: 'Built-in',
    icon: 'carbon:catalog',
    count: builtinConfigurations.value.length
  },
  {
    key: 'custom',
    label: 'Custom',
    icon: 'carbon:user',
    count: customConfigurations.value.length
  },
  {
    key: 'system',
    label: 'System',
    icon: 'carbon:desktop',
    count: systemFonts.value.length
  }
])

const allFonts = computed(() => {
  const fonts: (FontConfiguration | SystemFont)[] = []

  if (props.showBuiltin) {
    fonts.push(...builtinConfigurations.value)
  }

  if (props.showCustom) {
    fonts.push(...customConfigurations.value)
  }

  if (props.showSystemFonts) {
    fonts.push(...systemFontsArray.value)
  }

  return fonts
})

const builtinConfigurations = computed(() => {
  return Object.values(configurations.value).filter(config => config.isBuiltIn)
})

const customConfigurations = computed(() => {
  return Object.values(configurations.value).filter(config => !config.isBuiltIn)
})

const systemFontsArray = computed(() => {
  const fonts = Object.values(systemFonts.value)
  if (props.filterMonospace) {
    return fonts.filter(font => font.isMonospace)
  }
  return fonts
})

const filteredFonts = computed(() => {
  let fonts = allFonts.value

  // Filter by active tab
  if (activeTab.value !== 'all') {
    switch (activeTab.value) {
      case 'builtin':
        fonts = fonts.filter(font => 'isBuiltIn' in font && font.isBuiltIn)
        break
      case 'custom':
        fonts = fonts.filter(font => 'isBuiltIn' in font && !font.isBuiltIn)
        break
      case 'system':
        fonts = fonts.filter(font => 'isInstalled' in font && font.isInstalled)
        break
    }
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    fonts = fonts.filter(font => {
      const searchableFields = [
        font.name || font.displayName,
        font.family,
        font.description || '',
        ('weight' in font ? font.weight : ''),
        ('style' in font ? font.style : '')
      ].join(' ').toLowerCase()

      return searchableFields.includes(query)
    })
  }

  return fonts
})

const totalPages = computed(() => {
  return Math.ceil(filteredFonts.value.length / props.pageSize)
})

const paginatedFonts = computed(() => {
  const start = (currentPage.value - 1) * props.pageSize
  const end = start + props.pageSize
  return filteredFonts.value.slice(start, end)
})

const hoverPreviewStyle = computed(() => {
  return {
    position: 'fixed',
    left: hoverPreviewPosition.value.x + 'px',
    top: hoverPreviewPosition.value.y + 'px',
    zIndex: 1000,
    pointerEvents: 'none' as const
  }
})

// Methods
const handleSearch = () => {
  currentPage.value = 1
}

const clearSearch = () => {
  searchQuery.value = ''
  currentPage.value = 1
}

const selectFont = (font: FontConfiguration | SystemFont) => {
  const value = 'id' in font ? font.id : font.family
  emit('update:modelValue', value)
  emit('font-selected', font)
}

const isSelected = (font: FontConfiguration | SystemFont) => {
  if (!props.modelValue) return false

  const fontId = 'id' in font ? font.id : font.family
  const selectedId = typeof props.modelValue === 'string' ? props.modelValue : props.modelValue?.id

  return fontId === selectedId
}

const isDefaultFont = (font: FontConfiguration | SystemFont) => {
  return 'isDefault' in font && font.isDefault
}

const duplicateFont = (font: FontConfiguration | SystemFont) => {
  emit('font-duplicate', font)
}

const editFont = (font: FontConfiguration | SystemFont) => {
  emit('font-edit', font)
}

const deleteFont = (font: FontConfiguration | SystemFont) => {
  emit('font-delete', font)
}

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const retry = async () => {
  // Reload fonts
}

const scanFonts = async () => {
  isScanning.value = true
  try {
    await scanSystemFonts()
    emit('scan-fonts')
  } finally {
    isScanning.value = false
  }
}

const getPreviewStyle = (font: FontConfiguration | SystemFont, large = false) => {
  const size = large ? (font.size || 16) : (font.size || 14)
  const lineHeight = font.lineHeight || 1.4
  const letterSpacing = font.letterSpacing || 0
  const weight = font.weight || '400'

  return {
    fontFamily: `'${font.family}', monospace`,
    fontSize: `${size}px`,
    fontWeight: weight,
    lineHeight: lineHeight.toString(),
    letterSpacing: `${letterSpacing}px`,
    fontFeatureSettings: ('ligatures' in font && font.ligatures) ? '"liga", "dlig"' : 'normal'
  }
}

const getPreviewText = (font: FontConfiguration | SystemFont) => {
  return props.previewText
}

const getExtendedPreviewText = (font: FontConfiguration | SystemFont) => {
  return `The quick brown fox jumps over the lazy dog
1234567890 !@#$%^&*()[]{}<>

function hello() {
  console.log("Hello, World!");
}

if (condition) {
  execute();
}`
}

const updateHoverPreviewPosition = (event: MouseEvent) => {
  if (hoveredFont.value) {
    hoverPreviewPosition.value = {
      x: event.clientX + 15,
      y: event.clientY - 50
    }
    showHoverPreview.value = true
  }
}

const hideHoverPreview = () => {
  showHoverPreview.value = false
  hoveredFont.value = null
}

// Watchers
watch(searchQuery, () => {
  currentPage.value = 1
})

watch(activeTab, () => {
  currentPage.value = 1
})

// Lifecycle
onMounted(() => {
  document.addEventListener('mousemove', updateHoverPreviewPosition)
  document.addEventListener('mouseleave', hideHoverPreview)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', updateHoverPreviewPosition)
  document.removeEventListener('mouseleave', hideHoverPreview)
})
</script>

<style scoped>
@reference "../../assets/css/main.css";

.font-selector {
  @apply bg-surface border border-border rounded-lg overflow-hidden;
}

.font-selector__search {
  @apply p-3 border-b border-border;
}

.font-selector__search-input {
  @apply w-full pl-10 pr-10 py-2 bg-surface-hover border border-border rounded-md text-text placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent;
}

.font-selector__tabs {
  @apply flex border-b border-border bg-surface;
}

.font-selector__tab {
  @apply flex items-center gap-2 px-4 py-2 text-sm font-medium text-text-secondary hover:text-text transition-colors border-b-2 border-transparent;
}

.font-selector__tab--active {
  @apply text-primary border-primary;
}

.font-selector__tab-count {
  @apply text-xs bg-surface-hover px-1.5 py-0.5 rounded;
}

.font-selector__list {
  @apply overflow-y-auto;
}

.font-selector__loading,
.font-selector__error,
.font-selector__empty {
  @apply px-4 py-8;
}

.font-selector__items {
  @apply divide-y divide-border;
}

.font-selector__item {
  @apply flex items-center gap-3 p-4 hover:bg-surface-hover cursor-pointer transition-colors;
}

.font-selector__item--selected {
  @apply bg-primary bg-opacity-10 border-l-4 border-primary;
}

.font-selector__preview {
  @apply flex-1 font-mono text-sm text-text truncate;
  min-width: 0;
}

.font-selector__info {
  @apply flex-1 min-w-0;
}

.font-selector__name {
  @apply font-medium text-text flex items-center gap-2;
}

.font-selector__builtin-badge {
  @apply text-xs bg-blue-100 text-blue-800 px-1.5 py-0.5 rounded;
}

.font-selector__default-badge {
  @apply text-xs bg-green-100 text-green-800 px-1.5 py-0.5 rounded;
}

.font-selector__details {
  @apply text-sm text-text-secondary mt-1;
}

.font-selector__description {
  @apply text-xs text-text-secondary mt-1 truncate;
}

.font-selector__actions {
  @apply flex items-center gap-1 opacity-0 transition-opacity;
}

.font-selector__item:hover .font-selector__actions {
  @apply opacity-100;
}

.font-selector__action {
  @apply p-1 rounded hover:bg-surface text-text-secondary hover:text-text transition-colors;
}

.font-selector__pagination {
  @apply flex items-center justify-center gap-2 p-3 border-t border-border bg-surface;
}

.font-selector__pagination-btn {
  @apply p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-text transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}

.font-selector__pagination-info {
  @apply text-sm text-text-secondary px-2;
}

.font-selector__actions-bar {
  @apply flex items-center gap-2 p-3 border-t border-border bg-surface;
}

.font-selector__action-btn {
  @apply flex items-center gap-2 px-3 py-2 bg-surface hover:bg-surface-hover text-text border border-border rounded-md transition-colors;
}

.font-selector__action-btn--primary {
  @apply bg-primary hover:bg-primary-dark text-white border-primary;
}

.font-selector__hover-preview {
  @apply bg-surface border border-border rounded-lg shadow-lg p-4 max-w-sm;
}

.font-selector__hover-preview-content {
  @apply space-y-2;
}

.font-selector__hover-preview-text {
  @apply font-mono text-sm leading-relaxed break-words;
}

.font-selector__hover-preview-info {
  @apply space-y-1;
}
</style>