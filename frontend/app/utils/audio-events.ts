/**
 * UI Event to Audio Event Mapping System
 *
 * This system maps UI events to appropriate audio effects,
 * handling event priorities, cooldowns, and conditions.
 */

export interface UIEventMapping {
  uiEvent: string
  audioEventId: string
  priority: number
  cooldownMs: number
  conditions?: string[]
  description?: string
}

export interface EventTriggerOptions {
  source?: string
  data?: any
  force?: boolean
  volume?: number
}

class AudioEventManager {
  private mappings = new Map<string, UIEventMapping[]>()
  private lastPlayed = new Map<string, number>()
  private isInitialized = false

  constructor() {
    this.initializeDefaultMappings()
  }

  /**
   * Initialize default UI event to audio event mappings
   */
  private initializeDefaultMappings(): void {
    const defaultMappings: UIEventMapping[] = [
      // System events
      {
        uiEvent: 'system:startup',
        audioEventId: 'system_startup',
        priority: 100,
        cooldownMs: 0,
        description: 'Application startup sound'
      },
      {
        uiEvent: 'system:shutdown',
        audioEventId: 'system_shutdown',
        priority: 100,
        cooldownMs: 0,
        description: 'Application shutdown sound'
      },
      {
        uiEvent: 'system:alert',
        audioEventId: 'system_alert',
        priority: 90,
        cooldownMs: 2000,
        description: 'Critical system alert'
      },

      // Terminal events
      {
        uiEvent: 'terminal:bell',
        audioEventId: 'terminal_bell',
        priority: 50,
        cooldownMs: 100,
        description: 'Terminal bell character'
      },
      {
        uiEvent: 'terminal:command_success',
        audioEventId: 'command_success',
        priority: 40,
        cooldownMs: 200,
        description: 'Command completed successfully'
      },
      {
        uiEvent: 'terminal:command_error',
        audioEventId: 'command_error',
        priority: 60,
        cooldownMs: 500,
        description: 'Command failed with error'
      },
      {
        uiEvent: 'terminal:command_start',
        audioEventId: 'command_start',
        priority: 30,
        cooldownMs: 100,
        description: 'Command execution started'
      },

      // UI Interaction events
      {
        uiEvent: 'ui:button_click',
        audioEventId: 'button_click',
        priority: 20,
        cooldownMs: 50,
        description: 'Button click sound'
      },
      {
        uiEvent: 'ui:menu_open',
        audioEventId: 'menu_open',
        priority: 25,
        cooldownMs: 100,
        description: 'Menu or dropdown opened'
      },
      {
        uiEvent: 'ui:menu_close',
        audioEventId: 'menu_close',
        priority: 20,
        cooldownMs: 50,
        description: 'Menu or dropdown closed'
      },
      {
        uiEvent: 'ui:tab_switch',
        audioEventId: 'tab_switch',
        priority: 15,
        cooldownMs: 100,
        description: 'Tab or view switched'
      },
      {
        uiEvent: 'ui:toggle_on',
        audioEventId: 'toggle_on',
        priority: 25,
        cooldownMs: 100,
        description: 'Toggle or switch activated'
      },
      {
        uiEvent: 'ui:toggle_off',
        audioEventId: 'toggle_off',
        priority: 20,
        cooldownMs: 100,
        description: 'Toggle or switch deactivated'
      },
      {
        uiEvent: 'ui:slider_change',
        audioEventId: 'slider_change',
        priority: 10,
        cooldownMs: 50,
        description: 'Slider value changed'
      },
      {
        uiEvent: 'ui:focus',
        audioEventId: 'focus',
        priority: 5,
        cooldownMs: 50,
        description: 'Element gained focus'
      },

      // Notification events
      {
        uiEvent: 'ui:notification',
        audioEventId: 'notification',
        priority: 70,
        cooldownMs: 300,
        description: 'General notification'
      },
      {
        uiEvent: 'ui:notification_success',
        audioEventId: 'notification_success',
        priority: 65,
        cooldownMs: 300,
        description: 'Success notification'
      },
      {
        uiEvent: 'ui:notification_warning',
        audioEventId: 'notification_warning',
        priority: 75,
        cooldownMs: 400,
        description: 'Warning notification'
      },
      {
        uiEvent: 'ui:notification_error',
        audioEventId: 'notification_error',
        priority: 85,
        cooldownMs: 500,
        description: 'Error notification'
      },

      // File system events
      {
        uiEvent: 'file:operation_start',
        audioEventId: 'file_operation_start',
        priority: 35,
        cooldownMs: 200,
        description: 'File operation started'
      },
      {
        uiEvent: 'file:operation_complete',
        audioEventId: 'file_complete',
        priority: 45,
        cooldownMs: 200,
        description: 'File operation completed successfully'
      },
      {
        uiEvent: 'file:operation_error',
        audioEventId: 'file_error',
        priority: 80,
        cooldownMs: 500,
        description: 'File operation failed'
      },
      {
        uiEvent: 'file:created',
        audioEventId: 'file_created',
        priority: 30,
        cooldownMs: 150,
        description: 'File or folder created'
      },
      {
        uiEvent: 'file:deleted',
        audioEventId: 'file_deleted',
        priority: 40,
        cooldownMs: 200,
        description: 'File or folder deleted'
      },
      {
        uiEvent: 'file:moved',
        audioEventId: 'file_moved',
        priority: 35,
        cooldownMs: 150,
        description: 'File or folder moved'
      },
      {
        uiEvent: 'file:copied',
        audioEventId: 'file_copied',
        priority: 35,
        cooldownMs: 150,
        description: 'File or folder copied'
      },

      // Network events
      {
        uiEvent: 'network:connected',
        audioEventId: 'network_connected',
        priority: 55,
        cooldownMs: 1000,
        description: 'Network connection established'
      },
      {
        uiEvent: 'network:disconnected',
        audioEventId: 'network_disconnected',
        priority: 65,
        cooldownMs: 1000,
        description: 'Network connection lost'
      },
      {
        uiEvent: 'network:data_received',
        audioEventId: 'network_data_received',
        priority: 15,
        cooldownMs: 100,
        description: 'Data received from network'
      },
      {
        uiEvent: 'network:data_sent',
        audioEventId: 'network_data_sent',
        priority: 10,
        cooldownMs: 100,
        description: 'Data sent over network'
      },

      // Keyboard events (for on-screen keyboard)
      {
        uiEvent: 'keyboard:keypress',
        audioEventId: 'keyboard_keypress',
        priority: 10,
        cooldownMs: 50,
        description: 'Keyboard key pressed'
      },
      {
        uiEvent: 'keyboard:enter',
        audioEventId: 'keyboard_enter',
        priority: 20,
        cooldownMs: 100,
        description: 'Enter key pressed'
      },
      {
        uiEvent: 'keyboard:backspace',
        audioEventId: 'keyboard_backspace',
        priority: 20,
        cooldownMs: 100,
        description: 'Backspace key pressed'
      },
      {
        uiEvent: 'keyboard:space',
        audioEventId: 'keyboard_space',
        priority: 15,
        cooldownMs: 80,
        description: 'Space key pressed'
      },

      // Search and filter events
      {
        uiEvent: 'search:start',
        audioEventId: 'search_start',
        priority: 25,
        cooldownMs: 200,
        description: 'Search initiated'
      },
      {
        uiEvent: 'search:complete',
        audioEventId: 'search_complete',
        priority: 30,
        cooldownMs: 200,
        description: 'Search completed'
      },
      {
        uiEvent: 'search:no_results',
        audioEventId: 'search_no_results',
        priority: 35,
        cooldownMs: 300,
        description: 'Search returned no results'
      },

      // Theme and appearance events
      {
        uiEvent: 'theme:changed',
        audioEventId: 'theme_changed',
        priority: 40,
        cooldownMs: 300,
        description: 'Theme or appearance changed'
      },
      {
        uiEvent: 'theme:animation',
        audioEventId: 'theme_animation',
        priority: 20,
        cooldownMs: 100,
        description: 'Theme animation or transition'
      },

      // Progress and loading events
      {
        uiEvent: 'progress:start',
        audioEventId: 'progress_start',
        priority: 30,
        cooldownMs: 200,
        description: 'Progress or loading started'
      },
      {
        uiEvent: 'progress:complete',
        audioEventId: 'progress_complete',
        priority: 40,
        cooldownMs: 200,
        description: 'Progress or loading completed'
      },

      // Window and focus events
      {
        uiEvent: 'window:minimize',
        audioEventId: 'window_minimize',
        priority: 25,
        cooldownMs: 200,
        description: 'Window minimized'
      },
      {
        uiEvent: 'window:maximize',
        audioEventId: 'window_maximize',
        priority: 25,
        cooldownMs: 200,
        description: 'Window maximized'
      },
      {
        uiEvent: 'window:focus_gained',
        audioEventId: 'window_focus_gained',
        priority: 15,
        cooldownMs: 100,
        description: 'Window gained focus'
      },
      {
        uiEvent: 'window:focus_lost',
        audioEventId: 'window_focus_lost',
        priority: 10,
        cooldownMs: 100,
        description: 'Window lost focus'
      }
    ]

    // Group mappings by UI event
    defaultMappings.forEach(mapping => {
      if (!this.mappings.has(mapping.uiEvent)) {
        this.mappings.set(mapping.uiEvent, [])
      }
      this.mappings.get(mapping.uiEvent)!.push(mapping)
    })

    // Sort mappings by priority (highest first)
    this.mappings.forEach(mappings => {
      mappings.sort((a, b) => b.priority - a.priority)
    })

    this.isInitialized = true
  }

  /**
   * Get all mappings for a UI event
   */
  getMappings(uiEvent: string): UIEventMapping[] {
    return this.mappings.get(uiEvent) || []
  }

  /**
   * Get the highest priority mapping for a UI event
   */
  getBestMapping(uiEvent: string): UIEventMapping | null {
    const mappings = this.getMappings(uiEvent)
    return mappings.length > 0 ? mappings[0] : null
  }

  /**
   * Check if an event can be played based on cooldown
   */
  canPlayEvent(uiEvent: string, audioEventId: string): boolean {
    const mapping = this.getBestMapping(uiEvent)

    if (!mapping || mapping.audioEventId !== audioEventId) {
      return false
    }

    if (mapping.cooldownMs <= 0) {
      return true
    }

    const lastPlayed = this.lastPlayed.get(`${uiEvent}:${audioEventId}`) || 0
    const now = Date.now()

    return (now - lastPlayed) >= mapping.cooldownMs
  }

  /**
   * Mark an event as played
   */
  markEventPlayed(uiEvent: string, audioEventId: string): void {
    const key = `${uiEvent}:${audioEventId}`
    this.lastPlayed.set(key, Date.now())
  }

  /**
   * Add or update a mapping
   */
  addMapping(mapping: UIEventMapping): void {
    if (!this.mappings.has(mapping.uiEvent)) {
      this.mappings.set(mapping.uiEvent, [])
    }

    const mappings = this.mappings.get(mapping.uiEvent)!
    const existingIndex = mappings.findIndex(m =>
      m.uiEvent === mapping.uiEvent && m.audioEventId === mapping.audioEventId
    )

    if (existingIndex >= 0) {
      mappings[existingIndex] = mapping
    } else {
      mappings.push(mapping)
    }

    // Re-sort by priority
    mappings.sort((a, b) => b.priority - a.priority)
  }

  /**
   * Remove a mapping
   */
  removeMapping(uiEvent: string, audioEventId: string): boolean {
    const mappings = this.mappings.get(uiEvent)
    if (!mappings) {
      return false
    }

    const index = mappings.findIndex(m =>
      m.uiEvent === uiEvent && m.audioEventId === audioEventId
    )

    if (index >= 0) {
      mappings.splice(index, 1)
      if (mappings.length === 0) {
        this.mappings.delete(uiEvent)
      }
      return true
    }

    return false
  }

  /**
   * Get all UI events that have mappings
   */
  getAllUIEvents(): string[] {
    return Array.from(this.mappings.keys())
  }

  /**
   * Get all mappings
   */
  getAllMappings(): UIEventMapping[] {
    const allMappings: UIEventMapping[] = []
    this.mappings.forEach(mappings => {
      allMappings.push(...mappings)
    })
    return allMappings
  }

  /**
   * Clear all mappings
   */
  clearMappings(): void {
    this.mappings.clear()
    this.lastPlayed.clear()
  }

  /**
   * Reset cooldowns
   */
  resetCooldowns(): void {
    this.lastPlayed.clear()
  }

  /**
   * Get statistics about event usage
   */
  getStats() {
    const totalEvents = this.mappings.size
    let totalMappings = 0
    const eventsWithCooldowns = 0

    this.mappings.forEach(mappings => {
      totalMappings += mappings.length
    })

    return {
      totalEvents,
      totalMappings,
      averageMappingsPerEvent: totalEvents > 0 ? totalMappings / totalEvents : 0,
      cooldownsActive: this.lastPlayed.size
    }
  }

  /**
   * Export mappings to JSON
   */
  exportMappings(): string {
    const data = {
      version: '1.0.0',
      exportedAt: new Date().toISOString(),
      mappings: this.getAllMappings()
    }
    return JSON.stringify(data, null, 2)
  }

  /**
   * Import mappings from JSON
   */
  importMappings(jsonData: string): boolean {
    try {
      const data = JSON.parse(jsonData)

      if (!data.mappings || !Array.isArray(data.mappings)) {
        throw new Error('Invalid mapping data format')
      }

      this.clearMappings()
      data.mappings.forEach((mapping: any) => {
        this.addMapping(mapping)
      })

      return true
    } catch (error) {
      console.error('Failed to import mappings:', error)
      return false
    }
  }

  /**
   * Check if the manager is initialized
   */
  isReady(): boolean {
    return this.isInitialized
  }
}

// Singleton instance
export const audioEventManager = new AudioEventManager()

/**
 * Utility functions for triggering audio events
 */
export const triggerAudioEvent = async (
  uiEvent: string,
  options: EventTriggerOptions = {}
): Promise<string | null> => {
  const mapping = audioEventManager.getBestMapping(uiEvent)

  if (!mapping) {
    return null
  }

  // Check cooldown
  if (!options.force && !audioEventManager.canPlayEvent(uiEvent, mapping.audioEventId)) {
    return null
  }

  try {
    // This would integrate with the audio composable
    // For now, we'll just mark as played and return the event ID
    audioEventManager.markEventPlayed(uiEvent, mapping.audioEventId)

    // TODO: Integrate with useAudio composable
    // const audio = useAudio()
    // await audio.playEvent(mapping.audioEventId)

    return mapping.audioEventId
  } catch (error) {
    console.error(`Failed to trigger audio event ${mapping.audioEventId}:`, error)
    return null
  }
}

/**
 * Create a debounced version of triggerAudioEvent
 */
export const createDebouncedAudioTrigger = (
  delay: number
) => {
  const timeouts = new Map<string, NodeJS.Timeout>()

  return (uiEvent: string, options: EventTriggerOptions = {}) => {
    const key = `${uiEvent}_${options.source || 'default'}`

    // Clear existing timeout
    if (timeouts.has(key)) {
      clearTimeout(timeouts.get(key)!)
    }

    // Set new timeout
    const timeout = setTimeout(() => {
      triggerAudioEvent(uiEvent, options)
      timeouts.delete(key)
    }, delay)

    timeouts.set(key, timeout)
  }
}

// Export common debounced triggers
export const debouncedKeyPressTrigger = createDebouncedAudioTrigger(50)
export const debouncedSliderChangeTrigger = createDebouncedAudioTrigger(100)
export const debouncedScrollTrigger = createDebouncedAudioTrigger(150)

export type { UIEventMapping, EventTriggerOptions }
export { AudioEventManager }