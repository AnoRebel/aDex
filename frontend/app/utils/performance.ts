/**
 * Performance Optimization Utilities for aDex-UI
 * Provides performance monitoring, optimization, and debugging tools
 */

// Performance monitoring interface
interface PerformanceMetrics {
  fps: number
  memoryUsage: number
  renderTime: number
  updateTime: number
  timestamp: number
}

interface PerformanceConfig {
  enableMonitoring: boolean
  enableOptimizations: boolean
  targetFPS: number
  maxMemoryUsage: number
  enableDebugMode: boolean
}

class PerformanceManager {
  private config: PerformanceConfig
  private metrics: PerformanceMetrics[]
  private observers: PerformanceObserver[]
  private animationFrameId: number | null = null
  private lastFrameTime: number = 0
  private frameCount: number = 0
  private fpsUpdateInterval: number = 1000 // Update FPS every second
  private lastFPSUpdate: number = 0

  constructor(config: Partial<PerformanceConfig> = {}) {
    this.config = {
      enableMonitoring: true,
      enableOptimizations: true,
      targetFPS: 60,
      maxMemoryUsage: 512 * 1024 * 1024, // 512MB
      enableDebugMode: false,
      ...config
    }

    this.metrics = []
    this.observers = []

    this.init()
  }

  /**
   * Initialize performance monitoring and optimizations
   */
  private init(): void {
    if (this.config.enableMonitoring) {
      this.setupPerformanceObservers()
      this.startFPSSampling()
    }

    if (this.config.enableOptimizations) {
      this.applyOptimizations()
    }

    // Handle visibility changes
    document.addEventListener('visibilitychange', this.handleVisibilityChange.bind(this))

    // Handle memory pressure
    if ('memory' in performance) {
      this.monitorMemoryUsage()
    }

    // Setup performance debugging
    if (this.config.enableDebugMode) {
      this.setupDebugMode()
    }
  }

  /**
   * Setup performance observers for various metrics
   */
  private setupPerformanceObservers(): void {
    // Measure navigation timing
    if ('PerformanceObserver' in window) {
      try {
        const navigationObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries()
          entries.forEach((entry) => {
            if (entry.entryType === 'navigation') {
              console.log('Navigation performance:', entry)
            }
          })
        })
        navigationObserver.observe({ entryTypes: ['navigation'] })
        this.observers.push(navigationObserver)
      } catch (error) {
        console.warn('Navigation timing observer not supported:', error)
      }

      // Measure resource loading
      try {
        const resourceObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries()
          entries.forEach((entry) => {
            if (entry.entryType === 'resource') {
              this.analyzeResourcePerformance(entry as PerformanceResourceTiming)
            }
          })
        })
        resourceObserver.observe({ entryTypes: ['resource'] })
        this.observers.push(resourceObserver)
      } catch (error) {
        console.warn('Resource timing observer not supported:', error)
      }

      // Measure paint timing
      try {
        const paintObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries()
          entries.forEach((entry) => {
            console.log('Paint timing:', entry.name, entry.startTime)
          })
        })
        paintObserver.observe({ entryTypes: ['paint'] })
        this.observers.push(paintObserver)
      } catch (error) {
        console.warn('Paint timing observer not supported:', error)
      }

      // Measure long tasks
      try {
        const longTaskObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries()
          entries.forEach((entry) => {
            console.warn('Long task detected:', entry.duration, 'ms')
            this.optimizeForLongTask(entry.duration)
          })
        })
        longTaskObserver.observe({ entryTypes: ['longtask'] })
        this.observers.push(longTaskObserver)
      } catch (error) {
        console.warn('Long task observer not supported:', error)
      }
    }
  }

  /**
   * Analyze resource loading performance
   */
  private analyzeResourcePerformance(entry: PerformanceResourceTiming): void {
    const loadTime = entry.responseEnd - entry.requestStart
    const size = entry.transferSize || 0

    if (loadTime > 1000) {
      console.warn(`Slow resource detected: ${entry.name} (${loadTime.toFixed(2)}ms)`)
    }

    if (size > 1024 * 1024) { // > 1MB
      console.warn(`Large resource detected: ${entry.name} (${(size / 1024 / 1024).toFixed(2)}MB)`)
    }
  }

  /**
   * Start FPS sampling
   */
  private startFPSSampling(): void {
    const measureFPS = (timestamp: number) => {
      if (this.lastFrameTime === 0) {
        this.lastFrameTime = timestamp
        this.lastFPSUpdate = timestamp
      }

      this.frameCount++

      // Update FPS every second
      if (timestamp - this.lastFPSUpdate >= this.fpsUpdateInterval) {
        const fps = Math.round((this.frameCount * 1000) / (timestamp - this.lastFPSUpdate))
        this.recordMetric('fps', fps)

        // Check if FPS is below target
        if (fps < this.config.targetFPS * 0.8) {
          this.optimizeForLowFPS(fps)
        }

        this.frameCount = 0
        this.lastFPSUpdate = timestamp
      }

      this.lastFrameTime = timestamp
      this.animationFrameId = requestAnimationFrame(measureFPS)
    }

    this.animationFrameId = requestAnimationFrame(measureFPS)
  }

  /**
   * Monitor memory usage
   */
  private monitorMemoryUsage(): void {
    const checkMemory = () => {
      if ('memory' in performance) {
        const memory = (performance as any).memory
        const usedMB = memory.usedJSHeapSize / 1024 / 1024

        this.recordMetric('memoryUsage', memory.usedJSHeapSize)

        if (memory.usedJSHeapSize > this.config.maxMemoryUsage) {
          console.warn(`High memory usage: ${usedMB.toFixed(2)}MB`)
          this.optimizeForMemoryPressure()
        }
      }
    }

    // Check memory every 5 seconds
    setInterval(checkMemory, 5000)
    checkMemory() // Initial check
  }

  /**
   * Handle visibility changes for performance optimization
   */
  private handleVisibilityChange(): void {
    if (document.hidden) {
      // Page is hidden, reduce activity
      this.pauseIntensiveTasks()
      this.reduceFPS()
    } else {
      // Page is visible, resume normal activity
      this.resumeIntensiveTasks()
      this.restoreFPS()
    }
  }

  /**
   * Apply performance optimizations
   */
  private applyOptimizations(): void {
    // Optimize images
    this.optimizeImages()

    // Optimize animations
    this.optimizeAnimations()

    // Optimize rendering
    this.optimizeRendering()

    // Setup lazy loading
    this.setupLazyLoading()

    // Optimize event listeners
    this.optimizeEventListeners()
  }

  /**
   * Optimize images
   */
  private optimizeImages(): void {
    const images = document.querySelectorAll('img')
    images.forEach((img) => {
      // Add loading="lazy" to images
      if (!img.hasAttribute('loading')) {
        img.setAttribute('loading', 'lazy')
      }

      // Add error handling
      img.addEventListener('error', () => {
        img.style.display = 'none'
      })
    })
  }

  /**
   * Optimize animations
   */
  private optimizeAnimations(): void {
    // Detect reduced motion preference
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)')

    if (prefersReducedMotion.matches) {
      document.body.classList.add('reduce-motion')
    }

    prefersReducedMotion.addEventListener('change', (e) => {
      if (e.matches) {
        document.body.classList.add('reduce-motion')
      } else {
        document.body.classList.remove('reduce-motion')
      }
    })
  }

  /**
   * Optimize rendering
   */
  private optimizeRendering(): void {
    // Use transform3d for hardware acceleration
    const style = document.createElement('style')
    style.textContent = `
      .gpu-accelerated {
        transform: translateZ(0);
        backface-visibility: hidden;
        perspective: 1000px;
      }
    `
    document.head.appendChild(style)

    // Add GPU acceleration class to interactive elements
    const interactiveElements = document.querySelectorAll('button, .card, .btn')
    interactiveElements.forEach((el) => {
      el.classList.add('gpu-accelerated')
    })
  }

  /**
   * Setup lazy loading for components
   */
  private setupLazyLoading(): void {
    // Intersection Observer for lazy loading
    if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('loaded')
            observer.unobserve(entry.target)
          }
        })
      }, {
        rootMargin: '50px'
      })

      // Observe lazy-load elements
      const lazyElements = document.querySelectorAll('.lazy-load')
      lazyElements.forEach((el) => observer.observe(el))
    }
  }

  /**
   * Optimize event listeners
   */
  private optimizeEventListeners(): void {
    // Use passive event listeners where possible
    const passiveOptions = { passive: true } as AddEventListenerOptions

    document.addEventListener('touchstart', () => {}, passiveOptions)
    document.addEventListener('wheel', () => {}, passiveOptions)
  }

  /**
   * Optimize for low FPS
   */
  private optimizeForLowFPS(currentFPS: number): void {
    console.warn(`Low FPS detected: ${currentFPS} (target: ${this.config.targetFPS})`)

    // Reduce animation complexity
    document.body.classList.add('low-performance')

    // Disable particle effects
    const particles = document.querySelectorAll('.particle-container')
    particles.forEach((p) => (p as HTMLElement).style.display = 'none')

    // Reduce update frequency for intensive components
    this.throttleUpdates()
  }

  /**
   * Optimize for memory pressure
   */
  private optimizeForMemoryPressure(): void {
    console.warn('Memory pressure detected, applying optimizations')

    // Clear caches
    if ('caches' in window) {
      caches.keys().then((names) => {
        names.forEach((name) => {
          caches.delete(name)
        })
      })
    }

    // Remove unused event listeners
    this.cleanupEventListeners()

    // Force garbage collection if available
    if ((window as any).gc) {
      (window as any).gc()
    }
  }

  /**
   * Optimize for long tasks
   */
  private optimizeForLongTask(duration: number): void {
    console.warn(`Long task detected: ${duration}ms`)

    // Consider breaking up the task
    if (duration > 100) {
      this.suggestTaskOptimization(duration)
    }
  }

  /**
   * Pause intensive tasks when page is hidden
   */
  private pauseIntensiveTasks(): void {
    // Pause animations
    document.body.style.animationPlayState = 'paused'

    // Stop interval-based updates
    this.throttleUpdates()
  }

  /**
   * Resume intensive tasks when page is visible
   */
  private resumeIntensiveTasks(): void {
    // Resume animations
    document.body.style.animationPlayState = 'running'

    // Restore normal update frequency
    this.restoreUpdateFrequency()
  }

  /**
   * Reduce FPS for better performance
   */
  private reduceFPS(): void {
    // Implementation would depend on the animation system
    console.log('Reducing FPS for performance')
  }

  /**
   * Restore normal FPS
   */
  private restoreFPS(): void {
    // Implementation would depend on the animation system
    console.log('Restoring normal FPS')
  }

  /**
   * Throttle updates for performance
   */
  private throttleUpdates(): void {
    // Implementation would throttle component updates
    console.log('Throttling updates for performance')
  }

  /**
   * Restore normal update frequency
   */
  private restoreUpdateFrequency(): void {
    // Implementation would restore normal update frequency
    console.log('Restoring normal update frequency')
  }

  /**
   * Cleanup event listeners
   */
  private cleanupEventListeners(): void {
    // Implementation would remove unused event listeners
    console.log('Cleaning up event listeners')
  }

  /**
   * Suggest task optimization
   */
  private suggestTaskOptimization(duration: number): void {
    console.log(`Consider breaking up ${duration}ms task into smaller chunks`)
  }

  /**
   * Setup debug mode
   */
  private setupDebugMode(): void {
    // Add performance debug overlay
    this.createDebugOverlay()

    // Add debug keyboard shortcuts
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey && e.shiftKey && e.key === 'P') {
        this.toggleDebugOverlay()
      }
    })
  }

  /**
   * Create debug overlay
   */
  private createDebugOverlay(): void {
    const overlay = document.createElement('div')
    overlay.id = 'performance-debug-overlay'
    overlay.style.cssText = `
      position: fixed;
      top: 10px;
      right: 10px;
      background: rgba(0, 0, 0, 0.9);
      color: #00ff41;
      font-family: monospace;
      font-size: 12px;
      padding: 10px;
      border-radius: 5px;
      z-index: 10000;
      min-width: 200px;
      display: none;
    `

    document.body.appendChild(overlay)
    this.updateDebugOverlay()
  }

  /**
   * Update debug overlay
   */
  private updateDebugOverlay(): void {
    const overlay = document.getElementById('performance-debug-overlay')
    if (!overlay) return

    const latestMetrics = this.getLatestMetrics()
    const memoryMB = latestMetrics ? (latestMetrics.memoryUsage / 1024 / 1024).toFixed(2) : 'N/A'

    overlay.innerHTML = `
      <div>FPS: ${latestMetrics?.fps || 'N/A'}</div>
      <div>Memory: ${memoryMB} MB</div>
      <div>Render: ${latestMetrics?.renderTime || 'N/A'} ms</div>
      <div>Update: ${latestMetrics?.updateTime || 'N/A'} ms</div>
    `

    if (this.config.enableDebugMode) {
      requestAnimationFrame(() => this.updateDebugOverlay())
    }
  }

  /**
   * Toggle debug overlay
   */
  private toggleDebugOverlay(): void {
    const overlay = document.getElementById('performance-debug-overlay')
    if (overlay) {
      overlay.style.display = overlay.style.display === 'none' ? 'block' : 'none'
    }
  }

  /**
   * Record a performance metric
   */
  private recordMetric(type: keyof Omit<PerformanceMetrics, 'timestamp'>, value: number): void {
    const metric: PerformanceMetrics = {
      fps: 0,
      memoryUsage: 0,
      renderTime: 0,
      updateTime: 0,
      timestamp: Date.now(),
      [type]: value
    }

    this.metrics.push(metric)

    // Keep only last 100 metrics
    if (this.metrics.length > 100) {
      this.metrics.shift()
    }
  }

  /**
   * Get latest performance metrics
   */
  public getLatestMetrics(): PerformanceMetrics | null {
    return this.metrics.length > 0 ? this.metrics[this.metrics.length - 1] : null
  }

  /**
   * Get average metrics over a time period
   */
  public getAverageMetrics(durationMs: number = 5000): PerformanceMetrics | null {
    const now = Date.now()
    const recentMetrics = this.metrics.filter(m => now - m.timestamp <= durationMs)

    if (recentMetrics.length === 0) return null

    const sum = recentMetrics.reduce((acc, metric) => ({
      fps: acc.fps + metric.fps,
      memoryUsage: acc.memoryUsage + metric.memoryUsage,
      renderTime: acc.renderTime + metric.renderTime,
      updateTime: acc.updateTime + metric.updateTime,
      timestamp: 0
    }), {
      fps: 0,
      memoryUsage: 0,
      renderTime: 0,
      updateTime: 0,
      timestamp: 0
    })

    const count = recentMetrics.length
    return {
      fps: Math.round(sum.fps / count),
      memoryUsage: Math.round(sum.memoryUsage / count),
      renderTime: Math.round(sum.renderTime / count),
      updateTime: Math.round(sum.updateTime / count),
      timestamp: now
    }
  }

  /**
   * Generate performance report
   */
  public generateReport(): string {
    const latest = this.getLatestMetrics()
    const average = this.getAverageMetrics(30000) // 30 seconds

    if (!latest || !average) {
      return 'Insufficient data for performance report'
    }

    return `
Performance Report
==================
Current FPS: ${latest.fps}
Average FPS: ${average.fps}
Current Memory: ${(latest.memoryUsage / 1024 / 1024).toFixed(2)} MB
Average Memory: ${(average.memoryUsage / 1024 / 1024).toFixed(2)} MB
Average Render Time: ${average.renderTime} ms
Average Update Time: ${average.updateTime} ms
Target FPS: ${this.config.targetFPS}
Max Memory: ${(this.config.maxMemoryUsage / 1024 / 1024).toFixed(2)} MB
    `
  }

  /**
   * Cleanup performance manager
   */
  public cleanup(): void {
    // Cancel animation frame
    if (this.animationFrameId) {
      cancelAnimationFrame(this.animationFrameId)
    }

    // Disconnect observers
    this.observers.forEach(observer => observer.disconnect())

    // Remove debug overlay
    const overlay = document.getElementById('performance-debug-overlay')
    if (overlay) {
      overlay.remove()
    }

    // Remove event listeners
    document.removeEventListener('visibilitychange', this.handleVisibilityChange.bind(this))
  }
}

// Export singleton instance
export const performanceManager = new PerformanceManager()

// Export utility functions
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout
  return (...args: Parameters<T>) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }
}

export const measurePerformance = <T extends (...args: any[]) => any>(
  name: string,
  func: T
): T => {
  return ((...args: Parameters<T>) => {
    const start = performance.now()
    const result = func(...args)
    const end = performance.now()
    console.log(`${name} took ${(end - start).toFixed(2)} ms`)
    return result
  }) as T
}

export default performanceManager