import { ref, computed, type Ref } from 'vue'

export interface LoadingState {
  id: string
  name: string
  isLoading: boolean
  progress?: number
  message?: string
  error?: string
  startTime: Date
  timeout?: number
  cancellable?: boolean
  onCancel?: () => void
}

export interface LoadingOptions {
  message?: string
  progress?: number
  timeout?: number
  cancellable?: boolean
  onCancel?: () => void
}

class LoadingManager {
  private loadingStates: Map<string, LoadingState> = new Map()
  private globalLoading = ref(false)
  private activeLoaders = ref(0)

  public createLoading(name: string, options: LoadingOptions = {}): string {
    const id = `loading_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`

    const loadingState: LoadingState = {
      id,
      name,
      isLoading: true,
      progress: options.progress,
      message: options.message,
      startTime: new Date(),
      timeout: options.timeout,
      cancellable: options.cancellable ?? true,
      onCancel: options.onCancel,
    }

    this.loadingStates.set(id, loadingState)
    this.activeLoaders.value++

    // Update global loading state
    this.globalLoading.value = this.activeLoaders.value > 0

    // Set timeout if specified
    if (options.timeout) {
      setTimeout(() => {
        this.updateLoading(id, { error: 'Operation timed out' })
        this.completeLoading(id)
      }, options.timeout)
    }

    return id
  }

  public updateLoading(id: string, updates: Partial<LoadingState>): void {
    const loadingState = this.loadingStates.get(id)
    if (loadingState) {
      Object.assign(loadingState, updates)
    }
  }

  public completeLoading(id: string, error?: string): void {
    const loadingState = this.loadingStates.get(id)
    if (loadingState) {
      loadingState.isLoading = false
      loadingState.error = error
      this.activeLoaders.value = Math.max(0, this.activeLoaders.value - 1)

      // Update global loading state
      this.globalLoading.value = this.activeLoaders.value > 0

      // Remove completed loader after a short delay
      setTimeout(() => {
        this.loadingStates.delete(id)
      }, 1000)
    }
  }

  public cancelLoading(id: string): boolean {
    const loadingState = this.loadingStates.get(id)
    if (loadingState && loadingState.cancellable) {
      if (loadingState.onCancel) {
        loadingState.onCancel()
      }
      this.completeLoading(id, 'Cancelled by user')
      return true
    }
    return false
  }

  public getLoading(id: string): LoadingState | undefined {
    return this.loadingStates.get(id)
  }

  public isLoading(id: string): boolean {
    const loadingState = this.loadingStates.get(id)
    return loadingState?.isLoading ?? false
  }

  public isGlobalLoading(): Ref<boolean> {
    return this.globalLoading
  }

  public getActiveLoaders(): LoadingState[] {
    return Array.from(this.loadingStates.values()).filter(state => state.isLoading)
  }

  public getLoadingByName(name: string): LoadingState[] {
    return Array.from(this.loadingStates.values()).filter(state => state.name === name)
  }

  public hasActiveLoaders(): boolean {
    return this.activeLoaders.value > 0
  }

  public getLoadingStats(): {
    total: number
    active: number
    completed: number
    withErrors: number
  } {
    const states = Array.from(this.loadingStates.values())
    return {
      total: states.length,
      active: states.filter(s => s.isLoading).length,
      completed: states.filter(s => !s.isLoading).length,
      withErrors: states.filter(s => s.error).length,
    }
  }

  public clearAll(): void {
    this.loadingStates.clear()
    this.activeLoaders.value = 0
    this.globalLoading.value = false
  }

  public createAsyncWrapper<T>(
    name: string,
    asyncFunction: () => Promise<T>,
    options: LoadingOptions = {}
  ): Promise<T> {
    const loadingId = this.createLoading(name, options)

    return asyncFunction()
      .then(result => {
        this.completeLoading(loadingId)
        return result
      })
      .catch(error => {
        this.completeLoading(loadingId, error.message)
        throw error
      })
  }
}

// Singleton instance
export const loadingManager = new LoadingManager()

// Composable for using loading states in components
export function useLoading() {
  const isLoading = computed(() => loadingManager.isGlobalLoading().value)
  const activeLoaders = computed(() => loadingManager.getActiveLoaders())

  const createLoading = (name: string, options?: LoadingOptions) => {
    return loadingManager.createLoading(name, options)
  }

  const updateLoading = (id: string, updates: Partial<LoadingState>) => {
    loadingManager.updateLoading(id, updates)
  }

  const completeLoading = (id: string, error?: string) => {
    loadingManager.completeLoading(id, error)
  }

  const cancelLoading = (id: string) => {
    return loadingManager.cancelLoading(id)
  }

  const wrapAsync = <T>(
    name: string,
    asyncFunction: () => Promise<T>,
    options?: LoadingOptions
  ): Promise<T> => {
    return loadingManager.createAsyncWrapper(name, asyncFunction, options)
  }

  return {
    isLoading,
    activeLoaders,
    createLoading,
    updateLoading,
    completeLoading,
    cancelLoading,
    wrapAsync,
    // Convenience methods
    isGlobalLoading: loadingManager.isGlobalLoading(),
    getActiveLoaders: () => loadingManager.getActiveLoaders(),
    getLoadingStats: () => loadingManager.getLoadingStats(),
  }
}

// Global loading overlay component
export const LoadingOverlay = {
  name: 'LoadingOverlay',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    message: {
      type: String,
      default: 'Loading...',
    },
    progress: {
      type: Number,
      default: undefined,
    },
    cancellable: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['cancel'],
  setup(props: any, { emit }: any) {
    const { activeLoaders } = useLoading()

    const displayLoaders = computed(() => {
      if (props.show) {
        return [{
          id: 'manual',
          name: 'Manual Loading',
          isLoading: true,
          message: props.message,
          progress: props.progress,
          startTime: new Date(),
          cancellable: props.cancellable,
        }]
      }
      return activeLoaders.value
    })

    const handleCancel = (loader: LoadingState) => {
      if (loader.id === 'manual') {
        emit('cancel')
      } else if (loader.cancellable) {
        loadingManager.cancelLoading(loader.id)
      }
    }

    return () => {
      if (displayLoaders.value.length === 0) return null

      return h('div', { class: 'loading-overlay' }, [
        h('div', { class: 'loading-backdrop' }),
        h('div', { class: 'loading-content' }, [
          // Loading spinner
          h('div', { class: 'loading-spinner' }),

          // Loading messages
          h('div', { class: 'loading-messages' },
            displayLoaders.value.map(loader =>
              h('div', { key: loader.id, class: 'loading-message' }, [
                h('div', { class: 'loading-text' }, loader.message || 'Loading...'),

                // Progress bar if available
                loader.progress !== undefined &&
                  h('div', { class: 'loading-progress-bar' }, [
                    h('div', {
                      class: 'loading-progress-fill',
                      style: { width: `${loader.progress}%` }
                    })
                  ]),

                // Cancel button if cancellable
                loader.cancellable &&
                  h('button', {
                    class: 'loading-cancel-btn',
                    onClick: () => handleCancel(loader)
                  }, 'Cancel')
              ])
            )
          ),
        ]),
      ])
    }
  }
}

// Export convenience functions
export const createLoading = (name: string, options?: LoadingOptions) =>
  loadingManager.createLoading(name, options)

export const completeLoading = (id: string, error?: string) =>
  loadingManager.completeLoading(id, error)

export const wrapAsync = <T>(
  name: string,
  asyncFunction: () => Promise<T>,
  options?: LoadingOptions
): Promise<T> =>
  loadingManager.createAsyncWrapper(name, asyncFunction, options)

export default loadingManager