<template>
  <div
    class="file-icon"
    :class="iconClasses"
    :style="iconStyles"
    :title="title"
  >
    <!-- Preview thumbnail using Nuxt Image with global config -->
    <NuxtImg
      v-if="shouldUseImageOptimization"
      v-bind="imageAttributes"
      class="thumbnail-image"
      @load="handleThumbnailLoad"
      @error="handleThumbnailError"
      @loadstart="handleLoadStart"
      @loadend="handleLoadEnd"
    />

    <!-- SVG icon -->
    <svg
      v-else-if="useSvgIcon"
      :width="size"
      :height="size"
      viewBox="0 0 24 24"
      class="svg-icon"
      :fill="iconColor"
    >
      <component :is="svgIconComponent" />
    </svg>

    <!-- Emoji/icon font fallback -->
    <div
      v-else
      class="emoji-icon"
      :style="{ fontSize: `${fontSize}px` }"
    >
      {{ emojiIcon }}
    </div>

    <!-- Status indicators -->
    <div v-if="showStatus" class="status-indicators">
      <!-- Loading indicator -->
      <div v-if="isLoading" class="loading-indicator">
        <div class="loading-spinner"></div>
      </div>

      <!-- Error indicator -->
      <div v-if="hasError" class="error-indicator">
        <svg width="12" height="12" viewBox="0 0 12 12">
          <path fill="currentColor" d="M6 0C2.686 0 0 2.686 0 6s2.686 6 6 6 6-2.686 6-6S9.314 0 6 0zm-1 4h2v4H5V4zm0 6h2v2H5v-2z"/>
        </svg>
      </div>

      <!-- Symlink indicator -->
      <div v-if="file.is_symlink" class="symlink-indicator">
        <svg width="10" height="10" viewBox="0 0 10 10">
          <path fill="currentColor" d="M7.5 1H2.5C1.673 1 1 1.673 1 2.5v5C1 8.327 1.673 9 2.5 9h5C8.327 9 9 8.327 9 7.5v-5C9 1.673 8.327 1 7.5 1zM2.5 2h5C7.776 2 8 2.224 8 2.5v5C8 7.776 7.776 8 7.5 8h-5C2.224 8 2 7.776 2 7.5v-5C2 2.224 2.224 2 2.5 2zm4.5 6L6 8.5 4 6.5H3L5.5 9 8 6.5H7z"/>
        </svg>
      </div>
    </div>

    <!-- File extension badge -->
    <div v-if="showExtension && fileExtension" class="extension-badge">
      {{ fileExtension }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FileSystemEntry } from '~/types/filesystem'
import { nuxtImageConfig } from '~/composables/useNuxtImageConfig'

// Props
interface Props {
  file: FileSystemEntry
  size?: number
  showPreview?: boolean
  showStatus?: boolean
  showExtension?: boolean
  color?: string
  theme?: 'light' | 'dark' | 'auto'
  style?: 'emoji' | 'svg' | 'mixed'
  // Whether the icon is above the fold (affects loading strategy)
  aboveFold?: boolean
  // Override global image config if needed
  overrideImageConfig?: {
    format?: 'auto' | 'avif' | 'webp' | 'jpg' | 'png'
    quality?: number
    loading?: 'lazy' | 'eager'
    placeholder?: boolean
  }
}

const props = withDefaults(defineProps<Props>(), {
  size: 24,
  showPreview: true,
  showStatus: true,
  showExtension: false,
  color: '',
  theme: 'auto',
  style: 'mixed',
  aboveFold: false
})

// Reactive state
const isLoading = ref(false)
const hasError = ref(false)
const thumbnailUrl = ref<string>('')

// Computed properties
const iconClasses = computed(() => {
  return [
    `file-icon--${props.style}`,
    `file-icon--${props.file.is_dir ? 'directory' : 'file'}`,
    {
      'file-icon--loading': isLoading.value,
      'file-icon--error': hasError.value,
      'file-icon--selected': props.file.is_selected,
      'file-icon--hidden': props.file.is_hidden,
      'file-icon--executable': props.file.is_executable,
      'file-icon--symlink': props.file.is_symlink,
      'file-icon--preview': thumbnailUrl.value,
    }
  ]
})

const iconStyles = computed(() => {
  const styles: Record<string, string> = {}

  if (props.color) {
    styles.color = props.color
  }

  return styles
})

const fontSize = computed(() => {
  return Math.round(props.size * 0.7)
})

const iconColor = computed(() => {
  if (props.color) return props.color

  // Return color based on file type
  if (props.file.is_dir) return '#FFB84D'
  if (props.file.is_executable) return '#4CAF50'
  if (props.file.is_hidden) return '#9E9E9E'

  return getFileTypeColor()
})

// Nuxt Image optimization computed properties using global config
const imageAttributes = computed(() => {
  if (!thumbnailUrl.value || !isImageFile(props.file)) return null

  return nuxtImageConfig.getImageAttributes(
    {
      name: props.file.name,
      path: thumbnailUrl.value,
      isImage: true
    },
    props.size,
    props.aboveFold
  )
})

const shouldUseImageOptimization = computed(() => {
  return props.showPreview &&
         thumbnailUrl.value &&
         isImageFile(props.file) &&
         !!imageAttributes.value
})

const isImageFile = (file: FileSystemEntry): boolean => {
  const ext = file.name.split('.').pop()?.toLowerCase()
  return ['png', 'jpg', 'jpeg', 'gif', 'bmp', 'webp', 'svg', 'ico', 'tiff', 'tif', 'avif', 'heic', 'heif'].includes(ext || '')
}

const useSvgIcon = computed(() => {
  return props.style === 'svg' || (props.style === 'mixed' && hasSvgIcon.value)
})

const emojiIcon = computed(() => {
  return getFileEmoji()
})

const svgIconComponent = computed(() => {
  return getFileSvgIcon()
})

const fileExtension = computed(() => {
  if (props.file.is_dir) return ''
  const ext = props.file.name.split('.').pop()
  return ext ? ext.toUpperCase() : ''
})

const title = computed(() => {
  let title = props.file.name

  if (props.file.is_symlink) {
    title += ' (Symbolic link)'
  }

  if (props.file.is_hidden) {
    title += ' (Hidden)'
  }

  if (props.file.is_executable) {
    title += ' (Executable)'
  }

  return title
})

const hasSvgIcon = computed(() => {
  return !!getFileSvgIcon()
})

// Methods
const getFileEmoji = (): string => {
  // Directories
  if (props.file.is_dir) {
    if (props.file.name === '.') return '🏠'
    if (props.file.name === '..') return '⬆️'
    if (props.file.name.toLowerCase().includes('desktop')) return '🖥️'
    if (props.file.name.toLowerCase().includes('documents')) return '📄'
    if (props.file.name.toLowerCase().includes('downloads')) return '⬇️'
    if (props.file.name.toLowerCase().includes('pictures') || props.file.name.toLowerCase().includes('images')) return '🖼️'
    if (props.file.name.toLowerCase().includes('music')) return '🎵'
    if (props.file.name.toLowerCase().includes('videos')) return '🎬'
    if (props.file.name.toLowerCase().includes('projects')) return '📁'
    if (props.file.name.toLowerCase().includes('applications')) return '📦'
    if (props.file.name.toLowerCase().includes('system')) return '⚙️'
    if (props.file.name.toLowerCase().includes('temp')) return '🗑️'
    return '📁'
  }

  // Files by extension
  const ext = props.file.name.split('.').pop()?.toLowerCase()

  // Text and documents
  if (['txt', 'md', 'markdown', 'rst', 'log'].includes(ext || '')) return '📄'
  if (['pdf', 'epub', 'mobi'].includes(ext || '')) return '📕'
  if (['doc', 'docx', 'rtf', 'odt'].includes(ext || '')) return '📘'
  if (['xls', 'xlsx', 'ods', 'csv'].includes(ext || '')) return '📗'
  if (['ppt', 'pptx', 'odp'].includes(ext || '')) return '📙'

  // Code files
  if (['js', 'jsx', 'ts', 'tsx', 'vue', 'svelte', 'html', 'htm'].includes(ext || '')) return '🌐'
  if (['css', 'scss', 'sass', 'less', 'styl'].includes(ext || '')) return '🎨'
  if (['json', 'xml', 'yaml', 'yml', 'toml', 'ini', 'conf', 'config'].includes(ext || '')) return '📋'
  if (['py', 'pyw', 'pyc', 'pyo'].includes(ext || '')) return '🐍'
  if (['java', 'class', 'jar'].includes(ext || '')) return '☕'
  if (['c', 'cpp', 'cc', 'cxx', 'h', 'hpp', 'hxx'].includes(ext || '')) return '🔧'
  if (['cs', 'vb', 'fs'].includes(ext || '')) return '🔷'
  if (['php', 'phtml', 'php3', 'php4', 'php5'].includes(ext || '')) return '🐘'
  if (['rb', 'rbw', 'gem'].includes(ext || '')) return '💎'
  if (['go'].includes(ext || '')) return '🐹'
  if (['rs', 'rlib'].includes(ext || '')) return '🦀'
  if (['swift', 'm'].includes(ext || '')) return '🍎'
  if (['kt', 'kts'].includes(ext || '')) return '🎯'
  if (['scala', 'sc'].includes(ext || '')) return '🔷'
  if (['sql', 'db', 'sqlite', 'sqlite3'].includes(ext || '')) return '🗄️'
  if (['sh', 'bash', 'zsh', 'fish', 'ps1', 'bat', 'cmd'].includes(ext || '')) return '⌨️'

  // Images
  if (['png', 'jpg', 'jpeg', 'gif', 'bmp', 'svg', 'webp', 'ico', 'tiff', 'tif'].includes(ext || '')) return '🖼️'

  // Audio
  if (['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma', 'm4a', 'opus', 'aiff', 'au'].includes(ext || '')) return '🎵'

  // Video
  if (['mp4', 'avi', 'mkv', 'mov', 'wmv', 'flv', 'webm', 'm4v', '3gp', 'ogv', 'ts', 'vob'].includes(ext || '')) return '🎬'

  // Archives
  if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz', 'lzma', 'z', 'arj', 'cab', 'lzh', 'ace', 'tar', 'gz'].includes(ext || '')) return '📦'
  if (['deb', 'rpm', 'pkg', 'dmg', 'iso', 'img'].includes(ext || '')) return '💿'

  // Executables
  if (['exe', 'msi', 'app', 'deb', 'rpm', 'apk', 'ipa'].includes(ext || '')) return '⚙️'

  // Fonts
  if (['ttf', 'otf', 'woff', 'woff2', 'eot'].includes(ext || '')) return '🔤'

  // Special files
  if (['lock', 'pid'].includes(ext || '')) return '🔒'
  if (['key', 'pem', 'crt', 'cert', 'p12', 'pfx'].includes(ext || '')) return '🔐'
  if (['db', 'sqlite', 'mdb'].includes(ext || '')) return '🗄️'

  return '📄'
}

const getFileSvgIcon = () => {
  // Return SVG component name based on file type
  if (props.file.is_dir) return 'FolderIcon'

  const ext = props.file.name.split('.').pop()?.toLowerCase()

  // Map extensions to SVG components
  const iconMap: Record<string, string> = {
    // Text files
    'txt': 'TextFileIcon',
    'md': 'MarkdownFileIcon',
    'pdf': 'PDFFileIcon',

    // Code files
    'js': 'JavaScriptFileIcon',
    'ts': 'TypeScriptFileIcon',
    'vue': 'VueFileIcon',
    'html': 'HTMLFileIcon',
    'css': 'CSSFileIcon',
    'json': 'JSONFileIcon',
    'py': 'PythonFileIcon',
    'go': 'GoFileIcon',
    'rs': 'RustFileIcon',

    // Images
    'png': 'ImageFileIcon',
    'jpg': 'ImageFileIcon',
    'svg': 'SVGFileIcon',

    // Archives
    'zip': 'ArchiveFileIcon',
    'tar': 'ArchiveFileIcon',
    'gz': 'ArchiveFileIcon',
  }

  return iconMap[ext || ''] || 'DefaultFileIcon'
}

const getFileTypeColor = (): string => {
  const ext = props.file.name.split('.').pop()?.toLowerCase()

  // Color coding by file type
  if (props.file.is_dir) return '#FFB84D' // Orange for directories

  // Documents
  if (['txt', 'md', 'doc', 'docx', 'pdf'].includes(ext || '')) return '#4285F4' // Blue

  // Code
  if (['js', 'ts', 'vue', 'html', 'css', 'py', 'go', 'rs', 'java'].includes(ext || '')) return '#34A853' // Green

  // Images
  if (['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp'].includes(ext || '')) return '#FBBC04' // Yellow

  // Audio
  if (['mp3', 'wav', 'flac', 'aac', 'ogg'].includes(ext || '')) return '#EA4335' // Red

  // Video
  if (['mp4', 'avi', 'mkv', 'mov', 'webm'].includes(ext || '')) return '#9C27B0' // Purple

  // Archives
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext || '')) return '#607D8B' // Blue grey

  return '#757575' // Default grey
}

const loadThumbnail = async () => {
  if (!props.showPreview || !props.file.is_dir) return

  isLoading.value = true
  hasError.value = false

  try {
    // This would integrate with your thumbnail generation service
    // For now, we'll simulate with a placeholder
    const response = await fetch(`/api/filesystem/thumbnail?path=${encodeURIComponent(props.file.path)}&size=${props.size}`)

    if (response.ok) {
      const blob = await response.blob()
      thumbnailUrl.value = URL.createObjectURL(blob)
    } else {
      throw new Error('Thumbnail not available')
    }
  } catch (error) {
    console.warn('Failed to load thumbnail:', error)
    hasError.value = true
  } finally {
    isLoading.value = false
  }
}

const handleThumbnailLoad = () => {
  isLoading.value = false
  hasError.value = false
}

const handleThumbnailError = () => {
  isLoading.value = false
  hasError.value = true

  // Record error in performance metrics
  nuxtImageConfig.recordImageError()

  thumbnailUrl.value = ''
  loadStartTime.value = 0
}

// Nuxt Image specific event handlers with performance tracking
const loadStartTime = ref<number>(0)

const handleLoadStart = () => {
  isLoading.value = true
  hasError.value = false
  loadStartTime.value = Date.now()
}

const handleLoadEnd = () => {
  isLoading.value = false

  // Record performance metrics
  if (loadStartTime.value > 0) {
    const loadTime = Date.now() - loadStartTime.value
    nuxtImageConfig.recordImageLoad(loadTime)
    loadStartTime.value = 0
  }
}

// Watch for file changes
watch(() => props.file.path, () => {
  // Clear existing thumbnail
  if (thumbnailUrl.value) {
    URL.revokeObjectURL(thumbnailUrl.value)
    thumbnailUrl.value = ''
  }

  hasError.value = false
  loadThumbnail()
})

// Initialize
if (props.showPreview) {
  loadThumbnail()
}

// Cleanup
onUnmounted(() => {
  if (thumbnailUrl.value) {
    URL.revokeObjectURL(thumbnailUrl.value)
  }
})
</script>

<style scoped>
.file-icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: v-bind('`${size}px`');
  height: v-bind('`${size}px`');
  border-radius: 4px;
  transition: all 0.2s ease;
  overflow: hidden;
}

.file-icon--directory {
  color: #FFB84D;
}

.file-icon--file {
  color: v-bind('iconColor');
}

.file-icon--loading {
  opacity: 0.7;
}

.file-icon--error {
  opacity: 0.5;
}

.file-icon--selected {
  background: rgba(66, 133, 244, 0.1);
  border: 1px solid rgba(66, 133, 244, 0.3);
}

.file-icon--hidden {
  opacity: 0.6;
}

.file-icon--executable {
  background: rgba(52, 168, 83, 0.1);
}

.file-icon--symlink::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 0 6px 6px 0;
  border-color: transparent #9C27B0 transparent transparent;
}

.file-icon--preview {
  background: var(--surface);
  border: 1px solid var(--surface-border);
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 3px;
  transition: opacity 0.2s ease;
}

.thumbnail-image img {
  border-radius: 3px;
}

/* Nuxt Image specific optimizations */
.thumbnail-image[data-state="loading"] {
  opacity: 0.6;
}

.thumbnail-image[data-state="error"] {
  opacity: 0.3;
}

.thumbnail-image[data-state="loaded"] {
  opacity: 1;
}

/* Performance optimizations for different sizes */
.file-icon[data-size="small"] .thumbnail-image {
  image-rendering: pixelated;
  image-rendering: -moz-crisp-edges;
  image-rendering: crisp-edges;
}

.file-icon[data-size="large"] .thumbnail-image,
.file-icon[data-size="extra-large"] .thumbnail-image {
  image-rendering: auto;
}

.svg-icon {
  flex-shrink: 0;
}

.emoji-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.status-indicators {
  position: absolute;
  top: -2px;
  right: -2px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.loading-indicator,
.error-indicator,
.symlink-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--surface);
  border: 1px solid var(--surface-border);
}

.loading-indicator {
  background: var(--primary-500);
  border-color: var(--primary-600);
}

.loading-spinner {
  width: 10px;
  height: 10px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-top: 1px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.error-indicator {
  background: var(--error-500);
  border-color: var(--error-600);
  color: white;
}

.symlink-indicator {
  background: var(--secondary-500);
  border-color: var(--secondary-600);
  color: white;
}

.extension-badge {
  position: absolute;
  bottom: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.7);
  color: white;
  font-size: 8px;
  font-weight: 600;
  padding: 1px 3px;
  border-radius: 2px;
  text-transform: uppercase;
  line-height: 1;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Dark theme adjustments */
@media (prefers-color-scheme: dark) {
  .file-icon--preview {
    background: var(--surface-elevated);
    border-color: var(--surface-border);
  }

  .extension-badge {
    background: rgba(255, 255, 255, 0.9);
    color: var(--text-primary);
  }
}

/* Hover effects */
.file-icon:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.file-icon--selected:hover {
  transform: scale(1.02);
}

/* Focus styles for accessibility */
.file-icon:focus {
  outline: 2px solid var(--primary-500);
  outline-offset: 2px;
}

/* Animation */
@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* Size variations */
.file-icon[data-size="small"] {
  width: 16px;
  height: 16px;
}

.file-icon[data-size="large"] {
  width: 48px;
  height: 48px;
}

.file-icon[data-size="extra-large"] {
  width: 64px;
  height: 64px;
}
</style>