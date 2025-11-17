/**
 * Audio Loader Utility
 *
 * This utility handles loading and caching of audio files.
 * Currently a placeholder implementation that will be extended
 * to use Web Audio API or Howler.js integration.
 */

export interface AudioFile {
  id: string
  name: string
  path: string
  duration: number
  volume: number
  category: string
}

export interface LoadedAudio {
  id: string
  audioElement: HTMLAudioElement | null
  isLoaded: boolean
  error: string | null
  loadPromise: Promise<void> | null
}

class AudioLoader {
  private cache = new Map<string, LoadedAudio>()
  private supportedFormats = ['wav', 'mp3', 'ogg', 'aac']
  private isWebAudioSupported = false
  private audioContext: AudioContext | null = null

  constructor() {
    this.initializeAudioContext()
  }

  /**
   * Initialize Web Audio API context if supported
   */
  private initializeAudioContext() {
    try {
      if (typeof window !== 'undefined' && 'AudioContext' in window) {
        this.audioContext = new AudioContext()
        this.isWebAudioSupported = true
        console.log('Web Audio API supported')
      } else {
        console.log('Web Audio API not supported, using HTML5 Audio fallback')
      }
    } catch (error) {
      console.warn('Failed to initialize Web Audio API:', error)
      this.isWebAudioSupported = false
    }
  }

  /**
   * Check if audio loading is supported
   */
  isSupported(): boolean {
    return typeof Audio !== 'undefined' || this.isWebAudioSupported
  }

  /**
   * Get supported audio formats
   */
  getSupportedFormats(): string[] {
    return [...this.supportedFormats]
  }

  /**
   * Load an audio file
   */
  async loadAudio(audioFile: AudioFile): Promise<LoadedAudio> {
    // Check cache first
    if (this.cache.has(audioFile.id)) {
      const cached = this.cache.get(audioFile.id)!
      if (cached.loadPromise) {
        await cached.loadPromise
      }
      return cached
    }

    // Create cache entry
    const loadedAudio: LoadedAudio = {
      id: audioFile.id,
      audioElement: null,
      isLoaded: false,
      error: null,
      loadPromise: null
    }

    this.cache.set(audioFile.id, loadedAudio)

    // Start loading
    loadedAudio.loadPromise = this.loadAudioElement(audioFile, loadedAudio)

    try {
      await loadedAudio.loadPromise
    } catch (error) {
      console.error(`Failed to load audio ${audioFile.id}:`, error)
    }

    return loadedAudio
  }

  /**
   * Load audio using HTML5 Audio element
   */
  private async loadAudioElement(audioFile: AudioFile, loadedAudio: LoadedAudio): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const audio = new Audio()

        audio.addEventListener('canplaythrough', () => {
          loadedAudio.isLoaded = true
          loadedAudio.audioElement = audio
          resolve()
        }, { once: true })

        audio.addEventListener('error', (e) => {
          const error = e as any
          loadedAudio.error = `Failed to load audio: ${error.message || 'Unknown error'}`
          reject(new Error(loadedAudio.error))
        }, { once: true })

        // Set audio properties
        audio.preload = 'auto'
        audio.volume = audioFile.volume / 100

        // Start loading
        audio.src = audioFile.path
      } catch (error) {
        loadedAudio.error = error instanceof Error ? error.message : 'Failed to create audio element'
        reject(new Error(loadedAudio.error))
      }
    })
  }

  /**
   * Play an audio file
   */
  async playAudio(audioId: string, volume?: number): Promise<boolean> {
    const loadedAudio = this.cache.get(audioId)

    if (!loadedAudio || !loadedAudio.isLoaded || !loadedAudio.audioElement) {
      console.warn(`Audio ${audioId} not loaded`)
      return false
    }

    try {
      const audio = loadedAudio.audioElement

      // Reset if already playing
      if (!audio.paused) {
        audio.currentTime = 0
      }

      // Set volume if provided
      if (volume !== undefined) {
        audio.volume = Math.max(0, Math.min(1, volume / 100))
      }

      // Play the audio
      await audio.play()
      return true
    } catch (error) {
      console.error(`Failed to play audio ${audioId}:`, error)
      return false
    }
  }

  /**
   * Stop playing an audio file
   */
  stopAudio(audioId: string): boolean {
    const loadedAudio = this.cache.get(audioId)

    if (!loadedAudio || !loadedAudio.audioElement) {
      return false
    }

    try {
      const audio = loadedAudio.audioElement
      audio.pause()
      audio.currentTime = 0
      return true
    } catch (error) {
      console.error(`Failed to stop audio ${audioId}:`, error)
      return false
    }
  }

  /**
   * Preload multiple audio files
   */
  async preloadMultiple(audioFiles: AudioFile[]): Promise<LoadedAudio[]> {
    const promises = audioFiles.map(file => this.loadAudio(file))
    return Promise.all(promises)
  }

  /**
   * Get cached audio
   */
  getCachedAudio(audioId: string): LoadedAudio | undefined {
    return this.cache.get(audioId)
  }

  /**
   * Check if audio is loaded
   */
  isAudioLoaded(audioId: string): boolean {
    const loadedAudio = this.cache.get(audioId)
    return loadedAudio?.isLoaded || false
  }

  /**
   * Get load progress for an audio file
   */
  getLoadProgress(audioId: string): number {
    const loadedAudio = this.cache.get(audioId)

    if (!loadedAudio || !loadedAudio.audioElement) {
      return 0
    }

    const audio = loadedAudio.audioElement

    if (loadedAudio.isLoaded) {
      return 1
    }

    // Estimate progress based on readyState
    switch (audio.readyState) {
      case audio.HAVE_NOTHING: return 0
      case audio.HAVE_METADATA: return 0.1
      case audio.HAVE_CURRENT_DATA: return 0.3
      case audio.HAVE_FUTURE_DATA: return 0.7
      case audio.HAVE_ENOUGH_DATA: return 1
      default: return 0
    }
  }

  /**
   * Clear audio cache
   */
  clearCache(): void {
    // Stop all playing audio
    this.cache.forEach(loadedAudio => {
      if (loadedAudio.audioElement && !loadedAudio.audioElement.paused) {
        loadedAudio.audioElement.pause()
      }
    })

    this.cache.clear()
  }

  /**
   * Remove specific audio from cache
   */
  removeFromCache(audioId: string): boolean {
    const loadedAudio = this.cache.get(audioId)

    if (loadedAudio) {
      if (loadedAudio.audioElement && !loadedAudio.audioElement.paused) {
        loadedAudio.audioElement.pause()
      }
      return this.cache.delete(audioId)
    }

    return false
  }

  /**
   * Get cache statistics
   */
  getCacheStats() {
    const total = this.cache.size
    const loaded = Array.from(this.cache.values()).filter(a => a.isLoaded).length
    const errors = Array.from(this.cache.values()).filter(a => a.error).length

    return {
      total,
      loaded,
      errors,
      progress: total > 0 ? loaded / total : 0
    }
  }

  /**
   * Cleanup resources
   */
  cleanup(): void {
    this.clearCache()

    if (this.audioContext) {
      try {
        this.audioContext.close()
      } catch (error) {
        console.warn('Failed to close audio context:', error)
      }
      this.audioContext = null
      this.isWebAudioSupported = false
    }
  }
}

// Singleton instance
export const audioLoader = new AudioLoader()

/**
 * Utility functions for common audio operations
 */
export const audioUtils = {
  /**
   * Create audio file objects from soundpack data
   */
  createAudioFiles(soundpack: any, baseUrl: string = '/audio'): AudioFile[] {
    const audioFiles: AudioFile[] = []

    if (soundpack.events) {
      Object.values(soundpack.events).forEach((event: any) => {
        if (event && event.id && event.filePath) {
          audioFiles.push({
            id: event.id,
            name: event.name || event.id,
            path: event.filePath.startsWith('http')
              ? event.filePath
              : `${baseUrl}${event.filePath}`,
            duration: event.duration || 0,
            volume: event.volume || 50,
            category: event.category || 'unknown'
          })
        }
      })
    }

    return audioFiles
  },

  /**
   * Format duration in milliseconds to human-readable string
   */
  formatDuration(ms: number): string {
    if (ms < 1000) {
      return `${ms}ms`
    }
    return `${(ms / 1000).toFixed(1)}s`
  },

  /**
   * Validate audio file path
   */
  isValidAudioPath(path: string): boolean {
    if (!path || typeof path !== 'string') {
      return false
    }

    const extension = path.split('.').pop()?.toLowerCase()
    return this.audioLoader.getSupportedFormats().includes(extension || '')
  }
}

// Export types and utilities
export type { AudioFile, LoadedAudio }
export { AudioLoader }