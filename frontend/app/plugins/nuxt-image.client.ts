/**
 * Nuxt Image plugin for file system optimizations
 * Configures global image settings for optimal performance
 */

export default defineNuxtPlugin(() => {
  const { nuxtImageConfig } = useNuxtImageConfig()

  // Detect user preferences and browser capabilities
  if (typeof window !== 'undefined') {
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
      nuxtImageConfig.updateConfig({
        quality: 60,
        placeholder: true,
        loading: 'lazy'
      })
    }

    if (prefersReducedMotion) {
      // Disable animations that might affect performance
      // This can be used by components to adjust behavior
      document.documentElement.setAttribute('data-reduced-motion', 'true')
    }

    // Add event listeners for network changes
    if (connection) {
      const handleNetworkChange = () => {
        if (connection.saveData || ['slow-2g', '2g'].includes(connection.effectiveType)) {
          nuxtImageConfig.updateConfig({
            quality: 50,
            enableAVIF: false, // Disable AVIF on very slow networks
            enableWebP: connection.effectiveType !== '2g'
          })
        } else if (['3g'].includes(connection.effectiveType)) {
          nuxtImageConfig.updateConfig({
            quality: 70,
            enableAVIF: false,
            enableWebP: true
          })
        } else {
          // Good network - use optimal settings
          nuxtImageConfig.resetConfig()
        }
      }

      connection.addEventListener('change', handleNetworkChange)
    }

    // Monitor page visibility for lazy loading optimization
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Page is hidden - can pause certain operations
        document.documentElement.setAttribute('data-page-hidden', 'true')
      } else {
        // Page is visible - resume normal operations
        document.documentElement.removeAttribute('data-page-hidden')
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    // Monitor device memory for adaptive loading
    if ('deviceMemory' in navigator) {
      const memory = (navigator as any).deviceMemory as number

      if (memory < 4) {
        // Low memory device
        nuxtImageConfig.updateConfig({
          quality: 65,
          loading: 'lazy',
          enableAVIF: false
        })
      } else if (memory < 8) {
        // Medium memory device
        nuxtImageConfig.updateConfig({
          quality: 75,
          loading: 'lazy'
        })
      }
    }

    // Add performance monitoring
    if ('performance' in window && 'measure' in performance) {
      // Monitor Core Web Vitals related to image loading
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'largest-contentful-paint') {
            const lcpEntry = entry as PerformanceEntry
            console.debug('LCP:', lcpEntry.name, lcpEntry.startTime)
          } else if (entry.entryType === 'layout-shift') {
            const clsEntry = entry as any
            if (!clsEntry.hadRecentInput) {
              console.debug('CLS:', clsEntry.value)
            }
          } else if (entry.entryType === 'first-input') {
            const fidEntry = entry as PerformanceEntry
            console.debug('FID:', fidEntry.processingStart - fidEntry.startTime)
          }
        }
      })

      observer.observe({ entryTypes: ['largest-contentful-paint', 'layout-shift', 'first-input'] })
    }
  }

  // Set up CDN URL from environment variables if available
  const cdnUrl = useRuntimeConfig().public.cdnUrl || useRuntimeConfig().public.imageCdnUrl
  if (cdnUrl) {
    nuxtImageConfig.updateConfig({ cdnUrl })
  }

  // Initialize performance monitoring
  nuxtImageConfig.updateConfig({
    format: 'auto',
    quality: 80,
    loading: 'lazy',
    placeholder: true,
    enableWebP: true,
    enableAVIF: true,
    optimizeForSize: true
  })

  // Provide global error handling for image loading
  window.addEventListener('error', (event) => {
    const target = event.target as HTMLImageElement
    if (target?.tagName === 'IMG') {
      // Log image loading errors for debugging
      console.warn('Image loading error:', {
        src: target.src,
        alt: target.alt,
        error: event.error
      })

      // Could implement fallback images here
      if (!target.dataset.fallback) {
        // Try to load a fallback placeholder
        target.dataset.fallback = 'true'
        target.src = '/images/placeholder-error.jpg'
      }
    }
  }, true)

  console.log('Nuxt Image plugin initialized with file system optimizations')
})