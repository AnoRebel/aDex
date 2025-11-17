/**
 * Vue Audio Plugin
 *
 * Provides Vue plugin and directives for easy audio event triggering
 * from components and templates.
 */

import type { App, Directive } from 'vue'
import { triggerAudioEvent, createDebouncedAudioTrigger } from '~/utils/audio-events'

// Plugin options
export interface AudioPluginOptions {
  enableDebug?: boolean
  respectReducedMotion?: boolean
  respectFocusVisible?: boolean
}

// Audio plugin instance
export const createAudioPlugin = (options: AudioPluginOptions = {}) => {
  const {
    enableDebug = false,
    respectReducedMotion = true,
    respectFocusVisible = true
  } = options

  return {
    install(app: App) {
      // Global properties
      app.config.globalProperties.$audio = {
        trigger: triggerAudioEvent,
        play: triggerAudioEvent,
        debug: enableDebug
      }

      // Provide to composition API
      app.provide('audio', {
        trigger: triggerAudioEvent,
        play: triggerAudioEvent,
        options: {
          enableDebug,
          respectReducedMotion,
          respectFocusVisible
        }
      })

      // Register directives
      app.directive('audio-sound', createAudioSoundDirective())
      app.directive('audio-hover', createAudioHoverDirective())
      app.directive('audio-click', createAudioClickDirective())
      app.directive('audio-focus', createAudioFocusDirective())

      if (enableDebug) {
        console.log('🎵 Audio plugin installed with options:', options)
      }
    }
  }
}

/**
 * v-audio-sound directive
 * Triggers audio on specific DOM events
 */
const createAudioSoundDirective = (): Directive => {
  return {
    mounted(el, binding) {
      const eventName = binding.arg || 'click'
      const audioEvent = binding.value
      const options = binding.modifiers

      if (!audioEvent) {
        console.warn('v-audio-sound: No audio event specified')
        return
      }

      const handler = async (event: Event) => {
        // Respect reduced motion preference
        if (options.respectMotion && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
          return
        }

        // Don't play if element is disabled
        if ((el as HTMLElement).disabled) {
          return
        }

        try {
          await triggerAudioEvent(audioEvent, {
            source: `${el.tagName.toLowerCase()}:${eventName}`,
            data: event
          })
        } catch (error) {
          console.warn(`Failed to play audio event "${audioEvent}":`, error)
        }
      }

      el.addEventListener(eventName, handler)
      ;(el as any)._audioSoundHandler = handler
    },

    unmounted(el) {
      const handler = (el as any)._audioSoundHandler
      if (handler) {
        el.removeEventListener('click', handler)
        delete (el as any)._audioSoundHandler
      }
    }
  }
}

/**
 * v-audio-hover directive
 * Triggers audio on hover/enter events
 */
const createAudioHoverDirective = (): Directive => {
  return {
    mounted(el, binding) {
      const audioEvent = binding.value || 'ui:hover'
      const options = binding.modifiers

      let isHovering = false
      let hoverTimeout: NodeJS.Timeout | null = null

      const handleMouseEnter = async () => {
        if (isHovering) return

        // Skip if using keyboard navigation
        if (document.activeElement === el) return

        isHovering = true

        // Debounce hover sounds
        if (hoverTimeout) clearTimeout(hoverTimeout)

        hoverTimeout = setTimeout(async () => {
          try {
            await triggerAudioEvent(audioEvent, {
              source: 'hover',
              data: { element: el.tagName.toLowerCase() }
            })
          } catch (error) {
            console.warn(`Failed to play hover audio:`, error)
          }
        }, options.debounce ? 100 : 0)
      }

      const handleMouseLeave = () => {
        isHovering = false
        if (hoverTimeout) {
          clearTimeout(hoverTimeout)
          hoverTimeout = null
        }
      }

      el.addEventListener('mouseenter', handleMouseEnter)
      el.addEventListener('mouseleave', handleMouseLeave)
      el.addEventListener('focus', handleMouseEnter)
      el.addEventListener('blur', handleMouseLeave)

      // Store handlers for cleanup
      ;(el as any)._audioHoverHandlers = {
        mouseenter: handleMouseEnter,
        mouseleave: handleMouseLeave,
        focus: handleMouseEnter,
        blur: handleMouseLeave
      }
    },

    unmounted(el) {
      const handlers = (el as any)._audioHoverHandlers
      if (handlers) {
        el.removeEventListener('mouseenter', handlers.mouseenter)
        el.removeEventListener('mouseleave', handlers.mouseleave)
        el.removeEventListener('focus', handlers.focus)
        el.removeEventListener('blur', handlers.blur)
        delete (el as any)._audioHoverHandlers
      }
    }
  }
}

/**
 * v-audio-click directive
 * Triggers audio on click events
 */
const createAudioClickDirective = (): Directive => {
  return {
    mounted(el, binding) {
      const audioEvent = binding.value || 'ui:button_click'
      const options = binding.modifiers

      const handleClick = async (event: MouseEvent) => {
        // Don't play if right-clicked or disabled
        if (event.button !== 0 || (el as HTMLElement).disabled) {
          return
        }

        // Respect reduced motion
        if (options.respectMotion && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
          return
        }

        try {
          await triggerAudioEvent(audioEvent, {
            source: 'click',
            data: { element: el.tagName.toLowerCase(), event }
          })
        } catch (error) {
          console.warn(`Failed to play click audio:`, error)
        }
      }

      el.addEventListener('click', handleClick)
      ;(el as any)._audioClickHandler = handleClick
    },

    unmounted(el) {
      const handler = (el as any)._audioClickHandler
      if (handler) {
        el.removeEventListener('click', handler)
        delete (el as any)._audioClickHandler
      }
    }
  }
}

/**
 * v-audio-focus directive
 * Triggers audio on focus events
 */
const createAudioFocusDirective = (): Directive => {
  return {
    mounted(el, binding) {
      const audioEvent = binding.value || 'ui:focus'
      const options = binding.modifiers

      const handleFocus = async (event: FocusEvent) => {
        // Only play if focus is visible (keyboard navigation)
        if (options.keyboard && !(event.target as HTMLElement).matches(':focus-visible')) {
          return
        }

        try {
          await triggerAudioEvent(audioEvent, {
            source: 'focus',
            data: { element: el.tagName.toLowerCase(), event }
          })
        } catch (error) {
          console.warn(`Failed to play focus audio:`, error)
        }
      }

      el.addEventListener('focus', handleFocus)
      ;(el as any)._audioFocusHandler = handleFocus
    },

    unmounted(el) {
      const handler = (el as any)._audioFocusHandler
      if (handler) {
        el.removeEventListener('focus', handler)
        delete (el as any)._audioFocusHandler
      }
    }
  }
}

/**
 * Composable for audio functionality
 */
export const useAudioPlugin = () => {
  const audio = inject('audio') as {
    trigger: typeof triggerAudioEvent
    play: typeof triggerAudioEvent
    options: AudioPluginOptions
  }

  if (!audio) {
    throw new Error('Audio plugin not installed')
  }

  return audio
}

/**
 * Utility function to trigger events programmatically
 */
export const playAudioEvent = async (
  event: string,
  options?: Parameters<typeof triggerAudioEvent>[1]
) => {
  try {
    return await triggerAudioEvent(event, options)
  } catch (error) {
    console.error(`Failed to play audio event "${event}":`, error)
    return null
  }
}

/**
 * Predefined event triggers for common actions
 */
export const audioTriggers = {
  // System
  startup: () => playAudioEvent('system:startup'),
  shutdown: () => playAudioEvent('system:shutdown'),
  alert: () => playAudioEvent('system:alert'),

  // UI interactions
  click: () => playAudioEvent('ui:button_click'),
  menuOpen: () => playAudioEvent('ui:menu_open'),
  menuClose: () => playAudioEvent('ui:menu_close'),
  toggleOn: () => playAudioEvent('ui:toggle_on'),
  toggleOff: () => playAudioEvent('ui:toggle_off'),

  // Notifications
  notification: () => playAudioEvent('ui:notification'),
  success: () => playAudioEvent('ui:notification_success'),
  warning: () => playAudioEvent('ui:notification_warning'),
  error: () => playAudioEvent('ui:notification_error'),

  // File operations
  fileStart: () => playAudioEvent('file:operation_start'),
  fileComplete: () => playAudioEvent('file:operation_complete'),
  fileError: () => playAudioEvent('file:operation_error'),

  // Terminal
  terminalBell: () => playAudioEvent('terminal:bell'),
  commandSuccess: () => playAudioEvent('terminal:command_success'),
  commandError: () => playAudioEvent('terminal:command_error'),

  // Keyboard
  keyPress: () => playAudioEvent('keyboard:keypress'),
  enter: () => playAudioEvent('keyboard:enter'),
  backspace: () => playAudioEvent('keyboard:backspace'),

  // Network
  connected: () => playAudioEvent('network:connected'),
  disconnected: () => playAudioEvent('network:disconnected'),

  // Debounced versions for rapid events
  debouncedKeyPress: createDebouncedAudioTrigger(50),
  debouncedSliderChange: createDebouncedAudioTrigger(100),
  debouncedScroll: createDebouncedAudioTrigger(150)
}

// Export everything
export default createAudioPlugin
export type { AudioPluginOptions }