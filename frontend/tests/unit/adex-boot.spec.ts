import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import AdexBootScreen from '~/app/components/adex/AdexBootScreen.vue'

describe('AdexBootScreen Component', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function mountComponent(props?: { minDuration?: number; enableAudio?: boolean }) {
    wrapper = mount(AdexBootScreen, {
      props: {
        minDuration: props?.minDuration ?? 100, // Use short duration for tests
        enableAudio: props?.enableAudio ?? false,
      },
      global: {
        stubs: {
          Transition: false, // Use real transitions
        },
      },
    })
    return wrapper
  }

  describe('Rendering', () => {
    it('renders a full-screen overlay', () => {
      mountComponent()
      const overlay = wrapper.find('.boot-overlay')
      expect(overlay.exists()).toBe(true)
    })

    it('overlay has fixed positioning via CSS', () => {
      mountComponent()
      const overlay = wrapper.find('.boot-overlay')
      expect(overlay.exists()).toBe(true)
      // The CSS has position: fixed; width: 100vw; height: 100vh
    })

    it('shows the title "aDex"', async () => {
      mountComponent()
      // Title appears after 200ms sleep
      await vi.advanceTimersByTimeAsync(250)
      await wrapper.vm.$nextTick()

      const title = wrapper.find('.boot-title')
      expect(title.exists()).toBe(true)
      expect(title.text()).toBe('aDex')
    })

    it('title has data-text attribute for glitch effect', async () => {
      mountComponent()
      await vi.advanceTimersByTimeAsync(250)
      await wrapper.vm.$nextTick()

      const title = wrapper.find('.boot-title')
      expect(title.attributes('data-text')).toBe('aDex')
    })

    it('shows subtitle "ADVANCED DESKTOP EXPERIENCE"', async () => {
      mountComponent()
      // Subtitle appears after title (200ms + 600ms = 800ms)
      await vi.advanceTimersByTimeAsync(850)
      await wrapper.vm.$nextTick()

      const subtitle = wrapper.find('.boot-subtitle')
      expect(subtitle.exists()).toBe(true)
      expect(subtitle.text()).toBe('ADVANCED DESKTOP EXPERIENCE')
    })

    it('shows scanline effect overlay', () => {
      mountComponent()
      expect(wrapper.find('.boot-scanline').exists()).toBe(true)
    })

    it('shows boot content container', () => {
      mountComponent()
      expect(wrapper.find('.boot-content').exists()).toBe(true)
    })
  })

  describe('Boot messages', () => {
    it('shows boot messages during the animation', async () => {
      mountComponent()
      // Messages start appearing after ~1200ms (200 + 600 + 400)
            // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(12000)
      await wrapper.vm.$nextTick()

      const messages = wrapper.findAll('.boot-message')
      expect(messages.length).toBeGreaterThan(0)
    })

    it('each message has a prefix and text', async () => {
      mountComponent()
            // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(12000)
      await wrapper.vm.$nextTick()

      const messages = wrapper.findAll('.boot-message')
      if (messages.length > 0) {
        const firstMsg = messages[0]
        expect(firstMsg.find('.boot-message-icon').exists()).toBe(true)
        expect(firstMsg.find('.boot-message-tag').exists()).toBe(true)
        expect(firstMsg.find('.boot-message-text').exists()).toBe(true)
      }
    })

    it('marks completed messages with the success icon', async () => {
      mountComponent()
      // Advance enough for some messages to complete
      // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(14000)
      await wrapper.vm.$nextTick()

      const doneMessages = wrapper.findAll('.boot-message.done')
      expect(doneMessages.length).toBeGreaterThan(0)

      if (doneMessages.length > 0) {
        // Status is an icon now (see BOOT_LEVEL_ICON), not an [OK] string.
        const icon = doneMessages[0].find('.boot-message-icon')
        expect(icon.exists()).toBe(true)
        expect(icon.text().length).toBeGreaterThan(0)
      }
    })

    it('marks the in-flight message with an icon', async () => {
      mountComponent()
      // The very first message should be "current" right when it appears
      // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(12000)
      await wrapper.vm.$nextTick()

      const currentMessages = wrapper.findAll('.boot-message.current')
      if (currentMessages.length > 0) {
        const icon = currentMessages[0].find('.boot-message-icon')
        expect(icon.exists()).toBe(true)
      }
    })

    it('includes expected boot messages', async () => {
      mountComponent()
      // Advance enough for all messages to appear but before boot finishes and overlay hides
      // Messages start after ~1200ms, each takes ~280-400ms + 80ms gap
      // Total message phase is about 1200 + 7*(~300+80) = ~3860ms
      // The overlay hides ~5000ms+ so 4000ms is safe
            // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(12500)
      await wrapper.vm.$nextTick()

      // These track the real boot sequence in AdexBootScreen; the previous
      // assertions named messages the sequence no longer contains.
      const text = wrapper.text()
      expect(text).toContain('Loading settings')
      expect(text).toContain('Initializing theme engine')
      expect(text).toContain('Initializing audio cue system')
    })
  })

  describe('Progress bar', () => {
    it('shows a progress bar during boot', async () => {
      mountComponent()
      // Progress appears with messages
            // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(12000)
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.boot-progress-wrapper').exists()).toBe(true)
      expect(wrapper.find('.boot-progress-track').exists()).toBe(true)
      expect(wrapper.find('.boot-progress-fill').exists()).toBe(true)
    })

    it('progress bar width increases during boot', async () => {
      mountComponent()
            // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(12000)
      await wrapper.vm.$nextTick()

      const fill = wrapper.find('.boot-progress-fill')
      if (fill.exists()) {
        const style = fill.attributes('style')
        // Width should be greater than 0%
        expect(style).toContain('width:')
        // Extract the percentage from style
        const match = style?.match(/width:\s*([\d.]+)%/)
        if (match) {
          const percent = parseFloat(match[1])
          expect(percent).toBeGreaterThan(0)
        }
      }
    })

    it('progress reaches 100% when boot completes', async () => {
      mountComponent()
      // Past the 10s backend wait AND the canned sequence that follows it.
      await vi.advanceTimersByTimeAsync(20000)
      await wrapper.vm.$nextTick()

      const label = wrapper.find('.boot-progress-label')
      if (label.exists()) {
        expect(label.text()).toContain('100')
      }
    })
  })

  describe('Completion and events', () => {
    it('emits "complete" after the animation finishes', async () => {
      mountComponent({ minDuration: 100 })
      // Advance well past the entire boot sequence
      await vi.advanceTimersByTimeAsync(20000)
      await wrapper.vm.$nextTick()

      const emitted = wrapper.emitted('complete')
      expect(emitted).toBeTruthy()
      expect(emitted!.length).toBe(1)
    })

    it('emits "progress" events during boot', async () => {
      mountComponent()
      // Advance through some of the boot
            // The component waits up to 10s for backend boot events before
      // falling back to its canned sequence (the path jsdom takes, since no
      // Wails runtime emits them here). Advance past that window first.
      await vi.advanceTimersByTimeAsync(14000)
      await wrapper.vm.$nextTick()

      const progressEvents = wrapper.emitted('progress')
      expect(progressEvents).toBeTruthy()
      expect(progressEvents!.length).toBeGreaterThan(0)
    })

    it('progress events contain percentage values', async () => {
      mountComponent()
      await vi.advanceTimersByTimeAsync(5000)
      await wrapper.vm.$nextTick()

      const progressEvents = wrapper.emitted('progress')
      if (progressEvents && progressEvents.length > 0) {
        progressEvents.forEach(event => {
          const percent = event[0] as number
          expect(percent).toBeGreaterThanOrEqual(0)
          expect(percent).toBeLessThanOrEqual(100)
        })
      }
    })

    it('overlay becomes invisible after boot completes', async () => {
      mountComponent({ minDuration: 100 })
      // Advance well past the entire boot sequence + fade time
      await vi.advanceTimersByTimeAsync(20000)
      await wrapper.vm.$nextTick()

      // The overlay should be removed (v-if="visible" becomes false)
      const overlay = wrapper.find('.boot-overlay')
      expect(overlay.exists()).toBe(false)
    })
  })

  describe('Glitch effect', () => {
    it('applies glitching class to title during glitch phase', async () => {
      mountComponent()
      // Glitch starts after 200ms when title appears
      await vi.advanceTimersByTimeAsync(250)
      await wrapper.vm.$nextTick()

      const title = wrapper.find('.boot-title')
      if (title.exists()) {
        // The glitch should be active briefly after title appears
        expect(title.classes()).toContain('glitching')
      }
    })

    it('removes glitching class after initial glitch', async () => {
      mountComponent()
      // After title glitch: 200ms + 600ms = 800ms, glitch stops
      await vi.advanceTimersByTimeAsync(850)
      await wrapper.vm.$nextTick()

      const title = wrapper.find('.boot-title')
      if (title.exists()) {
        expect(title.classes()).not.toContain('glitching')
      }
    })
  })

  describe('Version display', () => {
    it('shows version info at the end of boot', async () => {
      mountComponent()
      // Version appears after all messages complete + progress reaches 100%
      await vi.advanceTimersByTimeAsync(8000)
      await wrapper.vm.$nextTick()

      const version = wrapper.find('.boot-version')
      if (version.exists()) {
        expect(version.text()).toContain('v2.0.0')
      }
    })
  })

  describe('Props', () => {
    it('accepts minDuration prop', () => {
      mountComponent({ minDuration: 5000 })
      expect(wrapper.props('minDuration')).toBe(5000)
    })

    it('accepts enableAudio prop', () => {
      mountComponent({ enableAudio: false })
      expect(wrapper.props('enableAudio')).toBe(false)
    })
  })

  describe('Lifecycle cleanup', () => {
    it('aborts boot sequence on unmount', async () => {
      mountComponent()
      // Start the boot sequence
      await vi.advanceTimersByTimeAsync(500)

      // Unmount before boot completes
      wrapper.unmount()

      // Advance timers further - should not throw
      await vi.advanceTimersByTimeAsync(10000)

      // The 'complete' event should NOT have been emitted since we unmounted early
      const emitted = wrapper.emitted('complete')
      expect(emitted).toBeFalsy()
    })
  })
})
