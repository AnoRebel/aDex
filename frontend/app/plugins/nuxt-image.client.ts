/**
 * Nuxt Image plugin for file system optimizations
 * Configures global image settings for optimal performance
 */

export default defineNuxtPlugin({
  name: 'nuxt-image-config',
  enforce: 'post', // Run after other plugins
  setup() {
    // Get the composable - must be called inside plugin/setup context
    let imageConfig: ReturnType<typeof useNuxtImageConfig> | null = null
    
    try {
      imageConfig = useNuxtImageConfig()
    } catch (e) {
      console.warn('useNuxtImageConfig not available yet, skipping image config')
      return
    }
    
    if (!imageConfig) return

  // Detect user preferences and browser capabilities
  if (import.meta.client) {
    // Check for reduced motion preference
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    // Check for data saver mode
    const connection = (navigator as any).connection || (navigator as any).mozConnection || (navigator as any).webkitConnection
    const isDataSaver = (navigator as any).connection?.saveData || false

    // Check network quality
    const isSlowNetwork = connection?.effectiveType && ['slow-2g', '2g', '3g'].includes(connection.effectiveType)

    // Adjust configuration based on user preferences and network conditions
    if (isDataSaver || isSlowNetwork) {
      // Reduce quality for data saver or slow networks
      imageConfig.updateConfig({
        quality: 60,
        placeholder: true,
        loading: 'lazy'
      })
    }

    if (prefersReducedMotion) {
      // Disable animations that might affect performance
      document.documentElement.setAttribute('data-reduced-motion', 'true')
    }

    // Add event listeners for network changes
    if (connection) {
      const handleNetworkChange = () => {
        if (connection.saveData || ['slow-2g', '2g'].includes(connection.effectiveType)) {
          imageConfig.updateConfig({
            quality: 50,
            enableAVIF: false,
            enableWebP: connection.effectiveType !== '2g'
          })
        } else if (['3g'].includes(connection.effectiveType)) {
          imageConfig.updateConfig({
            quality: 70,
            enableAVIF: false,
            enableWebP: true
          })
        } else {
          // Good network - use optimal settings
          imageConfig.resetConfig()
        }
      }

      connection.addEventListener('change', handleNetworkChange)
    }

    // Monitor page visibility for lazy loading optimization
    const handleVisibilityChange = () => {
      if (document.hidden) {
        document.documentElement.setAttribute('data-page-hidden', 'true')
      } else {
        document.documentElement.removeAttribute('data-page-hidden')
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    // Monitor device memory for adaptive loading
    if ('deviceMemory' in navigator) {
      const memory = (navigator as any).deviceMemory as number

      if (memory < 4) {
        imageConfig.updateConfig({
          quality: 65,
          loading: 'lazy',
          enableAVIF: false
        })
      } else if (memory < 8) {
        imageConfig.updateConfig({
          quality: 75,
          loading: 'lazy'
        })
      }
    }

    // Add performance monitoring
    if ('PerformanceObserver' in window) {
      try {
        const observer = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.entryType === 'largest-contentful-paint') {
              console.debug('LCP:', entry.name, entry.startTime)
            }
          }
        })

        observer.observe({ entryTypes: ['largest-contentful-paint'] })
      } catch (e) {
        // PerformanceObserver not supported for these entry types
      }
    }

    // Provide global error handling for image loading
    window.addEventListener('error', (event) => {
      const target = event.target as HTMLImageElement
      if (target?.tagName === 'IMG') {
        console.warn('Image loading error:', {
          src: target.src,
          alt: target.alt
        })

        // Could implement fallback images here
        if (!target.dataset.fallback) {
          target.dataset.fallback = 'true'
          // Use a simple placeholder instead of external file
          target.style.backgroundColor = '#333'
        }
      }
    }, true)
  }

  // Set up CDN URL from environment variables if available
  const runtimeConfig = useRuntimeConfig()
  const cdnUrl = runtimeConfig.public?.cdnUrl || runtimeConfig.public?.imageCdnUrl
  if (cdnUrl) {
    imageConfig.updateConfig({ cdnUrl })
  }

  // Initialize with optimal defaults
  imageConfig.updateConfig({
    format: 'auto',
    quality: 80,
    loading: 'lazy',
    placeholder: true,
    enableWebP: true,
    enableAVIF: true,
    optimizeForSize: true
  })

  console.log('Nuxt Image plugin initialized')
  }
})
