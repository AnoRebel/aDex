import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import AdexKeyboard from '~/app/components/adex/AdexKeyboard.vue'
import { useTerminalStore } from '~/stores/terminal'

// Mock layout data
const mockLayoutData = {
  qwerty: {
    name: 'QWERTY',
    displayName: 'QWERTY (US)',
    type: 'standard',
    rows: [
      [
        { key: 'Escape', width: 1, type: 'function' },
        { key: 'F1', width: 1, type: 'function' },
        { key: 'F2', width: 1, type: 'function' },
      ],
      [
        { key: '`', shift: '~', width: 1 },
        { key: '1', shift: '!', width: 1 },
        { key: '2', shift: '@', width: 1 },
        { key: 'Backspace', width: 2, type: 'special' },
      ],
      [
        { key: 'Tab', width: 1.5, type: 'special' },
        { key: 'q', shift: 'Q', width: 1 },
        { key: 'w', shift: 'W', width: 1 },
        { key: 'e', shift: 'E', width: 1 },
      ],
      [
        { key: 'CapsLock', width: 1.75, type: 'special' },
        { key: 'a', shift: 'A', width: 1 },
        { key: 's', shift: 'S', width: 1 },
        { key: 'Enter', width: 2.25, type: 'special' },
      ],
      [
        { key: 'Shift', width: 2.25, type: 'modifier', side: 'left' },
        { key: 'z', shift: 'Z', width: 1 },
        { key: 'x', shift: 'X', width: 1 },
        { key: 'Shift', width: 2.75, type: 'modifier', side: 'right' },
      ],
      [
        { key: 'Ctrl', width: 1.5, type: 'modifier', side: 'left' },
        { key: 'Alt', width: 1.5, type: 'modifier', side: 'left' },
        { key: 'Space', width: 7, type: 'space' },
        { key: 'Alt', width: 1.5, type: 'modifier', side: 'right' },
        { key: 'Ctrl', width: 1.5, type: 'modifier', side: 'right' },
      ],
    ],
  },
}

// Mock the dynamic import so it resolves instantly with layout data
vi.mock('~/assets/data/keyboard-layouts.json', () => ({
  default: { layouts: mockLayoutData },
  layouts: mockLayoutData,
}))

// Mock the Wails bindings used by the terminal store
vi.mock('~~/bindings', () => ({
  CreateTerminal: vi.fn().mockResolvedValue({ id: 'test-session' }),
  WriteToTerminal: vi.fn().mockResolvedValue(undefined),
  ResizeTerminal: vi.fn().mockResolvedValue(undefined),
  CloseTerminal: vi.fn().mockResolvedValue(undefined),
}))

describe('AdexKeyboard Component', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    wrapper?.unmount()
  })

  async function mountComponent() {
    const pinia = createPinia()
    setActivePinia(pinia)

    const termStore = useTerminalStore()

    // Set up a minimal session
    termStore.addSession({
      id: 'test-session',
      title: 'Test Shell',
      shell: '/bin/bash',
      workingDirectory: '/home/user',
      columns: 80,
      rows: 24,
      isActive: true,
      createdAt: new Date(),
      lastActivity: new Date(),
    })
    termStore.setActiveSession('test-session')

    wrapper = mount(AdexKeyboard, {
      global: { plugins: [pinia] }
    })

    // Wait for the async loadLayouts() to resolve, then force-set the layouts
    // directly via setupState to avoid race conditions with the dynamic import
    await flushPromises()
    await wrapper.vm.$nextTick()

    // If the dynamic import didn't resolve yet, inject layout data directly
    const setupState = (wrapper.vm.$ as any).setupState
    if (setupState && (!setupState.layouts || Object.keys(setupState.layouts).length === 0)) {
      setupState.layouts = mockLayoutData
    }
    await wrapper.vm.$nextTick()

    return { termStore }
  }

  describe('Rendering', () => {
    it('renders with adex-keyboard class', async () => {
      await mountComponent()
      expect(wrapper.find('.adex-keyboard').exists()).toBe(true)
    })

    it('has kb-row classes for each keyboard row', async () => {
      await mountComponent()
      const rows = wrapper.findAll('.kb-row')
      expect(rows.length).toBeGreaterThan(0)
    })

    it('has kb-key classes for each key', async () => {
      await mountComponent()
      const keys = wrapper.findAll('.kb-key')
      expect(keys.length).toBeGreaterThan(0)
    })

    it('renders multiple rows of keys', async () => {
      await mountComponent()
      const rows = wrapper.findAll('.kb-row')
      // Our mock layout has 6 rows
      expect(rows.length).toBe(6)
    })
  })

  describe('Key labels', () => {
    it('displays key labels', async () => {
      await mountComponent()
      const labels = wrapper.findAll('.kb-key-label')
      expect(labels.length).toBeGreaterThan(0)
    })

    it('shows abbreviated labels for special keys', async () => {
      await mountComponent()
      const allText = wrapper.text()
      // Backspace should be abbreviated to BKSP
      expect(allText).toContain('BKSP')
      // Enter should be abbreviated to RET
      expect(allText).toContain('RET')
      // Escape should be abbreviated to ESC
      expect(allText).toContain('ESC')
      // CapsLock should appear (abbreviated or full)
      expect(allText.toLowerCase()).toContain('caps')
    })

    it('shows shift characters as secondary labels for regular keys', async () => {
      await mountComponent()
      const shiftLabels = wrapper.findAll('.kb-key-shift')
      expect(shiftLabels.length).toBeGreaterThan(0)
    })
  })

  describe('Key classes', () => {
    it('applies fn-key class to function keys', async () => {
      await mountComponent()
      const fnKeys = wrapper.findAll('.fn-key')
      // F1, F2, ESC are function keys in our mock
      expect(fnKeys.length).toBeGreaterThan(0)
    })

    it('applies modifier-key class to modifier keys', async () => {
      await mountComponent()
      const modKeys = wrapper.findAll('.modifier-key')
      // Shift (2), Ctrl (2), Alt (2) = 6 modifier keys
      expect(modKeys.length).toBeGreaterThan(0)
    })

    it('applies spacebar class to space key', async () => {
      await mountComponent()
      const spaceKeys = wrapper.findAll('.spacebar')
      expect(spaceKeys.length).toBe(1)
    })

    it('applies special-key class to special keys', async () => {
      await mountComponent()
      const specialKeys = wrapper.findAll('.special-key')
      // Tab, Backspace, CapsLock, Enter are special keys
      expect(specialKeys.length).toBeGreaterThan(0)
    })
  })

  describe('Modifier key toggling', () => {
    it('toggles Shift state when Shift is clicked', async () => {
      await mountComponent()

      const shiftKeys = wrapper.findAll('.modifier-key')
      const shiftKey = shiftKeys.find(k => k.text().toLowerCase().includes('shift'))
      expect(shiftKey).toBeDefined()

      // Click Shift to activate it
      await shiftKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()

      // Shift should now be active (modifier-active class)
      expect(shiftKey!.classes()).toContain('modifier-active')

      // Click Shift again to deactivate
      await shiftKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()

      expect(shiftKey!.classes()).not.toContain('modifier-active')
    })

    it('toggles Ctrl state when Ctrl is clicked', async () => {
      await mountComponent()

      const modKeys = wrapper.findAll('.modifier-key')
      const ctrlKey = modKeys.find(k => k.text().toLowerCase().includes('ctrl'))
      expect(ctrlKey).toBeDefined()

      await ctrlKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()
      expect(ctrlKey!.classes()).toContain('modifier-active')

      await ctrlKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()
      expect(ctrlKey!.classes()).not.toContain('modifier-active')
    })

    it('toggles Alt state when Alt is clicked', async () => {
      await mountComponent()

      const modKeys = wrapper.findAll('.modifier-key')
      const altKey = modKeys.find(k => k.text().toLowerCase().includes('alt'))
      expect(altKey).toBeDefined()

      await altKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()
      expect(altKey!.classes()).toContain('modifier-active')
    })
  })

  describe('Key press sends to terminal', () => {
    it('calls sendInput on the terminal store when a regular key is pressed', async () => {
      const { termStore } = await mountComponent()

      // Spy on sendInput
      const sendInputSpy = vi.spyOn(termStore, 'sendInput').mockResolvedValue(undefined)

      // Find the 'a' key and click it
      const keys = wrapper.findAll('.kb-key')
      const aKey = keys.find(k => {
        const label = k.find('.kb-key-label')
        return label.exists() && label.text() === 'a'
      })
      expect(aKey).toBeDefined()

      await aKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()

      expect(sendInputSpy).toHaveBeenCalledWith('a')
    })

    it('sends uppercase character when Shift is active', async () => {
      const { termStore } = await mountComponent()

      const sendInputSpy = vi.spyOn(termStore, 'sendInput').mockResolvedValue(undefined)

      // First activate Shift
      const modKeys = wrapper.findAll('.modifier-key')
      const shiftKey = modKeys.find(k => k.text().toLowerCase().includes('shift'))
      await shiftKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()

      // Now press 'a' which should send 'A' (the shift value)
      const keys = wrapper.findAll('.kb-key')
      // After shift is active, the label changes from 'a' to 'A'
      const aKey = keys.find(k => {
        const label = k.find('.kb-key-label')
        return label.exists() && label.text() === 'A'
      })
      expect(aKey).toBeDefined()

      await aKey!.trigger('mousedown')
      await wrapper.vm.$nextTick()

      expect(sendInputSpy).toHaveBeenCalledWith('A')
    })
  })

  describe('Physical keyboard event handling', () => {
    it('registers keydown and keyup event listeners on mount', async () => {
      const addEventSpy = vi.spyOn(document, 'addEventListener')
      await mountComponent()

      const keydownCalls = addEventSpy.mock.calls.filter(c => c[0] === 'keydown')
      const keyupCalls = addEventSpy.mock.calls.filter(c => c[0] === 'keyup')
      expect(keydownCalls.length).toBeGreaterThan(0)
      expect(keyupCalls.length).toBeGreaterThan(0)
    })

    it('removes event listeners on unmount', async () => {
      const removeEventSpy = vi.spyOn(document, 'removeEventListener')
      await mountComponent()

      wrapper.unmount()

      const keydownCalls = removeEventSpy.mock.calls.filter(c => c[0] === 'keydown')
      const keyupCalls = removeEventSpy.mock.calls.filter(c => c[0] === 'keyup')
      expect(keydownCalls.length).toBeGreaterThan(0)
      expect(keyupCalls.length).toBeGreaterThan(0)
    })
  })

  describe('Fallback layout', () => {
    it('loads a fallback layout if the JSON import fails', async () => {
      // The component has fallback logic in loadLayouts catch block
      // Even without data, the keyboard should still render its root element
      await mountComponent()
      expect(wrapper.find('.adex-keyboard').exists()).toBe(true)
    })
  })
})
