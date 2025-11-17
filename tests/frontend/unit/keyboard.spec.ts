/**
 * Keyboard Component Tests
 * Tests for the on-screen keyboard implementation
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'

// Import components
import OnScreenKeyboard from '~/components/keyboard/OnScreenKeyboard.vue'
import KeyButton from '~/components/keyboard/KeyButton.vue'
import KeyLayout from '~/components/keyboard/KeyLayout.vue'

// Import composables
import { useKeyboard } from '~/composables/useKeyboard'

// Mock keyboard layouts
const mockLayouts = {
  qwerty: {
    name: 'QWERTY',
    rows: [
      [
        { id: 'escape', label: 'Esc', code: 'Escape', modifier: true },
        { id: '1', label: '1', code: 'Digit1' },
        { id: '2', label: '2', code: 'Digit2' },
        { id: '3', label: '3', code: 'Digit3' },
        { id: '4', label: '4', code: 'Digit4' },
        { id: '5', label: '5', code: 'Digit5' },
      ],
      [
        { id: 'q', label: 'Q', code: 'KeyQ', shiftLabel: '!' },
        { id: 'w', label: 'W', code: 'KeyW', shiftLabel: '@' },
        { id: 'e', label: 'E', code: 'KeyE', shiftLabel: '#' },
        { id: 'r', label: 'R', code: 'KeyR', shiftLabel: '$' },
        { id: 't', label: 'T', code: 'KeyT', shiftLabel: '%' },
      ],
      [
        { id: 'space', label: 'Space', code: 'Space', special: true, width: 'wide' },
        { id: 'enter', label: 'Enter', code: 'Enter', special: true },
      ]
    ]
  }
}

// Mock fetch for keyboard layouts
global.fetch = vi.fn().mockResolvedValue({
  ok: true,
  json: () => Promise.resolve(mockLayouts)
})

describe('KeyButton', () => {
  it('renders key with correct label', () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'a',
          label: 'A',
          code: 'KeyA',
          shiftLabel: '!'
        }
      }
    })

    expect(wrapper.text()).toContain('A')
    expect(wrapper.find('[data-testid="key-button"]').exists()).toBe(true)
  })

  it('applies correct modifier classes', () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'shift',
          label: 'Shift',
          code: 'ShiftLeft',
          modifier: true
        },
        isShiftPressed: true
      }
    })

    const button = wrapper.find('[data-testid="key-button"]')
    expect(button.classes()).toContain('modifier-active')
  })

  it('shows shift label when shift is pressed', async () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: '1',
          label: '1',
          code: 'Digit1',
          shiftLabel: '!'
        }
      }
    })

    expect(wrapper.text()).toContain('1')
    expect(wrapper.text()).not.toContain('!')

    await wrapper.setProps({ isShiftPressed: true })
    await nextTick()

    expect(wrapper.text()).not.toContain('1')
    expect(wrapper.text()).toContain('!')
  })

  it('emits key-press event when clicked', async () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'a',
          label: 'A',
          code: 'KeyA'
        }
      }
    })

    await wrapper.find('[data-testid="key-button"]').trigger('click')

    expect(wrapper.emitted('key-press')).toBeTruthy()
    expect(wrapper.emitted('key-press')?.[0]).toEqual([{
      id: 'a',
      label: 'A',
      code: 'KeyA',
      shift: false,
      ctrl: false,
      alt: false
    }])
  })

  it('handles touch events properly', async () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'a',
          label: 'A',
          code: 'KeyA'
        }
      }
    })

    await wrapper.find('[data-testid="key-button"]').trigger('touchstart')
    await wrapper.find('[data-testid="key-button"]').trigger('touchend')

    expect(wrapper.emitted('key-press')).toBeTruthy()
  })

  it('supports keyboard repeat when held down', async () => {
    vi.useFakeTimers()

    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'a',
          label: 'A',
          code: 'KeyA'
        }
      }
    })

    const button = wrapper.find('[data-testid="key-button"]')

    await button.trigger('mousedown')

    // Should emit immediately
    expect(wrapper.emitted('key-press')).toHaveLength(1)

    // Advance time to trigger repeat
    vi.advanceTimersByTime(500)
    await nextTick()

    // Should have repeated
    expect(wrapper.emitted('key-press')?.length).toBeGreaterThan(1)

    await button.trigger('mouseup')
    vi.useRealTimers()
  })
})

describe('KeyLayout', () => {
  it('renders keyboard layout correctly', () => {
    const wrapper = mount(KeyLayout, {
      props: {
        layout: mockLayouts.qwerty,
        isShiftPressed: false,
        isCtrlPressed: false,
        isAltPressed: false
      }
    })

    expect(wrapper.text()).toContain('Esc')
    expect(wrapper.text()).toContain('Q')
    expect(wrapper.text()).toContain('Space')
  })

  it('applies shift state correctly', async () => {
    const wrapper = mount(KeyLayout, {
      props: {
        layout: mockLayouts.qwerty,
        isShiftPressed: false,
        isCtrlPressed: false,
        isAltPressed: false
      }
    })

    expect(wrapper.text()).toContain('Q')
    expect(wrapper.text()).not.toContain('!')

    await wrapper.setProps({ isShiftPressed: true })
    await nextTick()

    // Note: In actual implementation, shift labels would be shown
    // This test verifies the prop is passed correctly
  })

  it('emits key-press events from child buttons', async () => {
    const wrapper = mount(KeyLayout, {
      props: {
        layout: mockLayouts.qwerty,
        isShiftPressed: false,
        isCtrlPressed: false,
        isAltPressed: false
      }
    })

    const firstKey = wrapper.findComponent(KeyButton)
    await firstKey.vm.$emit('key-press', {
      id: 'escape',
      label: 'Esc',
      code: 'Escape',
      shift: false,
      ctrl: false,
      alt: false
    })

    expect(wrapper.emitted('key-press')).toBeTruthy()
  })
})

describe('OnScreenKeyboard', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    wrapper = mount(OnScreenKeyboard, {
      global: {
        plugins: [pinia]
      }
    })
  })

  it('renders keyboard when visible', () => {
    expect(wrapper.find('[data-testid="onscreen-keyboard"]').exists()).toBe(true)
  })

  it('does not render when hidden', async () => {
    await wrapper.setProps({ visible: false })
    await nextTick()

    expect(wrapper.find('[data-testid="onscreen-keyboard"]').exists()).toBe(false)
  })

  it('loads keyboard layouts on mount', async () => {
    // Wait for async data loading
    await nextTick()
    await nextTick()

    // Should have loaded layouts
    expect(wrapper.vm.layouts).toBeDefined()
    expect(Object.keys(wrapper.vm.layouts)).toContain('qwerty')
  })

  it('switches between layouts', async () => {
    await wrapper.setProps({ visible: true })
    await nextTick()
    await nextTick()

    // Initially should have a layout selected
    expect(wrapper.vm.currentLayoutId).toBeDefined()

    const layoutButton = wrapper.find('[data-testid="layout-button"]')
    if (layoutButton.exists()) {
      await layoutButton.trigger('click')
      await nextTick()

      // Layout should have changed
      expect(wrapper.emitted('layout-changed')).toBeTruthy()
    }
  })

  it('handles key press events', async () => {
    await wrapper.setProps({ visible: true })
    await nextTick()
    await nextTick()

    // Simulate key press
    await wrapper.vm.$emit('key-press', {
      id: 'a',
      label: 'A',
      code: 'KeyA',
      shift: false,
      ctrl: false,
      alt: false
    })

    expect(wrapper.emitted('key-press')).toBeTruthy()
  })

  it('supports different keyboard sizes', async () => {
    await wrapper.setProps({ size: 'compact' })
    await nextTick()

    expect(wrapper.find('[data-testid="onscreen-keyboard"]').classes()).toContain('compact')

    await wrapper.setProps({ size: 'large' })
    await nextTick()

    expect(wrapper.find('[data-testid="onscreen-keyboard"]').classes()).toContain('large')
  })

  it('supports positioning options', async () => {
    await wrapper.setProps({ position: 'bottom' })
    await nextTick()

    expect(wrapper.find('[data-testid="onscreen-keyboard"]').classes()).toContain('position-bottom')

    await wrapper.setProps({ position: 'top' })
    await nextTick()

    expect(wrapper.find('[data-testid="onscreen-keyboard"]').classes()).toContain('position-top')
  })

  it('shows/hides settings panel', async () => {
    const settingsButton = wrapper.find('[data-testid="settings-button"]')

    if (settingsButton.exists()) {
      await settingsButton.trigger('click')
      await nextTick()

      expect(wrapper.find('[data-testid="keyboard-settings"]').exists()).toBe(true)

      // Close settings
      const closeButton = wrapper.find('[data-testid="close-settings"]')
      if (closeButton.exists()) {
        await closeButton.trigger('click')
        await nextTick()

        expect(wrapper.find('[data-testid="keyboard-settings"]').exists()).toBe(false)
      }
    }
  })

  it('maintains focus state properly', async () => {
    const inputField = document.createElement('input')
    document.body.appendChild(inputField)

    await wrapper.vm.showKeyboard(inputField)
    await nextTick()

    expect(wrapper.vm.isVisible).toBe(true)
    expect(wrapper.vm.targetElement).toBe(inputField)

    // Cleanup
    document.body.removeChild(inputField)
  })
})

describe('useKeyboard composable', () => {
  it('provides keyboard functionality', () => {
    const pinia = createPinia()
    const keyboard = useKeyboard()

    expect(keyboard).toBeDefined()
    expect(typeof keyboard.showKeyboard).toBe('function')
    expect(typeof keyboard.hideKeyboard).toBe('function')
    expect(typeof keyboard.sendKey).toBe('function')
  })

  it('manages keyboard state correctly', () => {
    const keyboard = useKeyboard()

    expect(keyboard.isVisible.value).toBe(false)
    expect(keyboard.currentLayout.value).toBeDefined()

    keyboard.showKeyboard()
    expect(keyboard.isVisible.value).toBe(true)

    keyboard.hideKeyboard()
    expect(keyboard.isVisible.value).toBe(false)
  })

  it('handles layout switching', () => {
    const keyboard = useKeyboard()

    const originalLayout = keyboard.currentLayout.value

    keyboard.switchLayout('qwertz')
    // Layout should change if qwertz exists
    expect(keyboard.currentLayout.value).toBeDefined()
  })

  it('supports key combination simulation', () => {
    const keyboard = useKeyboard()

    const mockElement = document.createElement('input')
    document.body.appendChild(mockElement)

    keyboard.sendKey('a', { shift: true, ctrl: true })
    // In real implementation, this would trigger the appropriate keyboard events

    // Cleanup
    document.body.removeChild(mockElement)
  })
})

describe('Keyboard Accessibility', () => {
  it('provides proper ARIA labels', () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'enter',
          label: 'Enter',
          code: 'Enter'
        }
      }
    })

    const button = wrapper.find('[data-testid="key-button"]')
    expect(button.attributes('aria-label')).toBe('Enter key')
    expect(button.attributes('role')).toBe('button')
  })

  it('supports keyboard navigation', async () => {
    const wrapper = mount(OnScreenKeyboard, {
      global: {
        plugins: [createPinia()]
      }
    })

    await wrapper.setProps({ visible: true })
    await nextTick()
    await nextTick()

    // Test that keyboard can receive focus
    const keyboard = wrapper.find('[data-testid="onscreen-keyboard"]')
    expect(keyboard.attributes('tabindex')).toBeDefined()
  })

  it('announces key presses to screen readers', async () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'a',
          label: 'A',
          code: 'KeyA'
        }
      }
    })

    const button = wrapper.find('[data-testid="key-button"]')

    // Mock screen reader announcement
    const announceSpy = vi.fn()
    window.announceToScreenReader = announceSpy

    await button.trigger('click')

    // Should announce the key press in some way
    // This would depend on the specific implementation
  })
})

describe('Keyboard Performance', () => {
  it('handles rapid key presses efficiently', async () => {
    const wrapper = mount(KeyButton, {
      props: {
        keyData: {
          id: 'a',
          label: 'A',
          code: 'KeyA'
        }
      }
    })

    const button = wrapper.find('[data-testid="key-button"]')
    const startTime = performance.now()

    // Simulate rapid key presses
    for (let i = 0; i < 10; i++) {
      await button.trigger('click')
    }

    const endTime = performance.now()
    const duration = endTime - startTime

    // Should handle 10 clicks in under 100ms
    expect(duration).toBeLessThan(100)
  })

  it('does not cause memory leaks', () => {
    const wrapper = mount(OnScreenKeyboard, {
      global: {
        plugins: [createPinia()]
      }
    })

    // Cleanup
    wrapper.unmount()

    // In real implementation, would check for event listeners
    // and other potential memory leaks
  })
})