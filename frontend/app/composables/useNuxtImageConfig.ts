/**
 * Composable for managing Nuxt Image configurations
 * specifically optimized for file system thumbnails
 */

export interface ImageConfig {
  format: 'auto' | 'avif' | 'webp' | 'jpg' | 'png'
  quality: number
  loading: 'lazy' | 'eager'
  placeholder: boolean
  cdnUrl: string
  enableWebP: boolean
  enableAVIF: boolean
  optimizeForSize: boolean
}

export interface SizeConfig {
  small: { size: number; quality: number; blur: number }
  medium: { size: number; quality: number; blur: number }
  large: { size: number; quality: number; blur: number }
  xlarge: { size: number; quality: number; blur: number }
}

export const useNuxtImageConfig = () => {
  // Default configuration
  const defaultConfig: ImageConfig = {
    format: 'auto',
    quality: 80,
    loading: 'lazy',
    placeholder: true,
    cdnUrl: '',
    enableWebP: true,
    enableAVIF: true,
    optimizeForSize: true
  }

  // Size-specific optimizations
  const sizeConfig: SizeConfig = {
    small: { size: 16, quality: 60, blur: 2 },
    medium: { size: 24, quality: 70, blur: 3 },
    large: { size: 48, quality: 80, blur: 4 },
    xlarge: { size: 64, quality: 85, blur: 6 }
  }

  // Reactive configuration
  const config = ref<ImageConfig>({ ...defaultConfig })

  // Get optimal format based on browser support and preferences
  const getOptimalFormat = (): ImageConfig['format'] => {
    if (config.value.format !== 'auto') return config.value.format

    // Check browser support for modern formats
    if (typeof window !== 'undefined') {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')

      if (ctx && config.value.enableAVIF) {
        // Check AVIF support (simplified check)
        if (canvas.toDataURL('image/avif').indexOf('data:image/avif') === 0) {
          return 'avif'
        }
      }

      if (ctx && config.value.enableWebP) {
        // Check WebP support
        if (canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0) {
          return 'webp'
        }
      }
    }

    return 'jpg' // Fallback to JPEG for maximum compatibility
  }

  // Get size-specific configuration
  const getSizeConfig = (size: number): SizeConfig[keyof SizeConfig] => {
    if (!config.value.optimizeForSize) return sizeConfig.medium

    if (size <= 20) return sizeConfig.small
    if (size <= 32) return sizeConfig.medium
    if (size <= 56) return sizeConfig.large
    return sizeConfig.xlarge
  }

  // Get image quality based on file type and size
  const getOptimalQuality = (fileType: string, size: number): number => {
    const baseConfig = getSizeConfig(size)

    // Adjust quality based on file type
    switch (fileType.toLowerCase()) {
      case 'png':
      case 'svg':
        return Math.min(baseConfig.quality + 10, 100) // Higher quality for graphics
      case 'jpg':
      case 'jpeg':
        return baseConfig.quality // Standard quality for photos
      case 'webp':
        return baseConfig.quality + 5 // WebP can handle higher quality
      case 'avif':
        return baseConfig.quality + 10 // AVIF is more efficient
      default:
        return baseConfig.quality
    }
  }

  // Get placeholder configuration
  const getPlaceholderConfig = (size: number) => {
    const sizeConfig = getSizeConfig(size)
    return {
      blur: sizeConfig.blur,
      quality: Math.max(sizeConfig.quality - 20, 30)
    }
  }

  // Determine if image should use lazy loading
  const shouldUseLazyLoading = (size: number, isAboveFold = false): boolean => {
    // Small images or above-fold images use eager loading
    if (size <= 24 || isAboveFold) return false

    // Use configuration preference
    return config.value.loading === 'lazy'
  }

  // Generate CDN URL if configured
  const getCdnUrl = (path: string): string => {
    if (!config.value.cdnUrl) return path

    const cdnBase = config.value.cdnUrl.replace(/\/$/, '')
    const cleanPath = path.replace(/^\//, '')

    return `${cdnBase}/${cleanPath}`
  }

  // Get responsive sizes for srcset
  const getResponsiveSizes = (baseSize: number): string => {
    const sizes = []

    // Generate sizes for different pixel densities
    sizes.push(`${baseSize}px`) // 1x
    sizes.push(`${baseSize * 1.5}px`) // 1.5x
    sizes.push(`${baseSize * 2}px`) // 2x

    return sizes.join(' ')
  }

  // Update configuration
  const updateConfig = (newConfig: Partial<ImageConfig>) => {
    config.value = { ...config.value, ...newConfig }
  }

  // Reset to defaults
  const resetConfig = () => {
    config.value = { ...defaultConfig }
  }

  // Performance monitoring
  const performanceStats = ref({
    imagesLoaded: 0,
    imagesError: 0,
    totalLoadTime: 0,
    averageLoadTime: 0
  })

  const recordImageLoad = (loadTime: number) => {
    performanceStats.value.imagesLoaded++
    performanceStats.value.totalLoadTime += loadTime
    performanceStats.value.averageLoadTime =
      performanceStats.value.totalLoadTime / performanceStats.value.imagesLoaded
  }

  const recordImageError = () => {
    performanceStats.value.imagesError++
  }

  // Get optimal image attributes for a file
  const getImageAttributes = (file: {
    name: string
    path: string
    size?: number
    isImage?: boolean
  }, iconSize: number, isAboveFold = false) => {
    const isImageFile = file.isImage || isImageFileByExtension(file.name)
    const fileType = getFileExtension(file.name)
    const sizeConfig = getSizeConfig(iconSize)

    return {
      src: getCdnUrl(file.path),
      width: iconSize,
      height: iconSize,
      sizes: getResponsiveSizes(iconSize),
      format: isImageFile ? getOptimalFormat() : undefined,
      quality: isImageFile ? getOptimalQuality(fileType, iconSize) : undefined,
      loading: isImageFile ? (shouldUseLazyLoading(iconSize, isAboveFold) ? 'lazy' : 'eager') : 'eager',
      placeholder: isImageFile && config.value.placeholder ? getPlaceholderConfig(iconSize) : undefined,
      preload: !shouldUseLazyLoading(iconSize, isAboveFold),
      alt: file.name
    }
  }

  // Helper function to check if file is an image by extension
  const isImageFileByExtension = (filename: string): boolean => {
    const ext = filename.split('.').pop()?.toLowerCase()
    const imageExtensions = [
      'jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp', 'ico',
      'tiff', 'tif', 'avif', 'heic', 'heif'
    ]
    return imageExtensions.includes(ext || '')
  }

  // Helper function to get file extension
  const getFileExtension = (filename: string): string => {
    return filename.split('.').pop()?.toLowerCase() || ''
  }

  // Computed properties
  const isWebPSupported = computed(() => {
    if (typeof window === 'undefined') return false
    return config.value.enableWebP
  })

  const isAVIFSupported = computed(() => {
    if (typeof window === 'undefined') return false
    return config.value.enableAVIF
  })

  const performanceReport = computed(() => {
    const total = performanceStats.value.imagesLoaded + performanceStats.value.imagesError
    return {
      ...performanceStats.value,
      errorRate: total > 0 ? (performanceStats.value.imagesError / total) * 100 : 0,
      successRate: total > 0 ? (performanceStats.value.imagesLoaded / total) * 100 : 100
    }
  })

  return {
    // Configuration
    config: readonly(config),
    updateConfig,
    resetConfig,

    // Size and format optimization
    getOptimalFormat,
    getSizeConfig,
    getOptimalQuality,
    getPlaceholderConfig,
    shouldUseLazyLoading,

    // URL and CDN handling
    getCdnUrl,
    getResponsiveSizes,

    // Main API
    getImageAttributes,

    // Performance
    performanceStats: readonly(performanceStats),
    performanceReport: readonly(performanceReport),
    recordImageLoad,
    recordImageError,

    // Support detection
    isWebPSupported: readonly(isWebPSupported),
    isAVIFSupported: readonly(isAVIFSupported),

    // Constants
    sizeConfig: readonly(sizeConfig)
  }
}

// Export singleton instance
export const nuxtImageConfig = useNuxtImageConfig()