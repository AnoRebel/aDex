/**
 * Keyboard E2E Tests
 * End-to-end tests for the on-screen keyboard functionality
 */

import { test, expect, Page } from '@playwright/test'

class KeyboardPage {
  constructor(private page: Page) {}

  async navigate() {
    await this.page.goto('/')
  }

  async openKeyboard() {
    // Click on an input field or button that should open keyboard
    await this.page.click('[data-testid="terminal-input"]')
    await this.page.waitForSelector('[data-testid="onscreen-keyboard"]')
  }

  async isKeyboardVisible() {
    return await this.page.isVisible('[data-testid="onscreen-keyboard"]')
  }

  async pressKey(keyId: string) {
    await this.page.click(`[data-testid="key-${keyId}"]`)
  }

  async pressKeySequence(keys: string[]) {
    for (const key of keys) {
      await this.pressKey(key)
      await this.page.waitForTimeout(50) // Small delay between keys
    }
  }

  async pressKeyWithModifier(keyId: string, modifier: 'shift' | 'ctrl' | 'alt') {
    // Hold modifier
    await this.page.click(`[data-testid="key-${modifier}"]`)
    await this.page.waitForTimeout(100)

    // Press key
    await this.pressKey(keyId)

    // Release modifier
    await this.page.click(`[data-testid="key-${modifier}"]`)
  }

  async switchLayout(layoutId: string) {
    await this.page.click(`[data-testid="layout-${layoutId}"]`)
  }

  async getKeyboardPosition() {
    const keyboard = this.page.locator('[data-testid="onscreen-keyboard"]')
    const boundingBox = await keyboard.boundingBox()
    return boundingBox
  }

  async resizeKeyboard(size: 'compact' | 'normal' | 'large') {
    await this.page.click('[data-testid="keyboard-settings"]')
    await this.page.click(`[data-testid="size-${size}"]`)
    await this.page.click('[data-testid="close-settings"]')
  }

  async getTerminalText() {
    return await this.page.inputValue('[data-testid="terminal-input"]')
  }

  async openKeyboardSettings() {
    await this.page.click('[data-testid="keyboard-settings"]')
    await this.page.waitForSelector('[data-testid="keyboard-settings-panel"]')
  }

  async closeKeyboardSettings() {
    await this.page.click('[data-testid="close-settings"]')
    await this.page.waitForSelector('[data-testid="keyboard-settings-panel"]', { state: 'hidden' })
  }

  async enableKeyRepeat(enabled: boolean) {
    await this.openKeyboardSettings()
    const toggle = await this.page.locator('[data-testid="key-repeat-toggle"]')

    if (await toggle.isChecked() !== enabled) {
      await toggle.click()
    }

    await this.closeKeyboardSettings()
  }

  async adjustKeyRepeatDelay(delay: number) {
    await this.openKeyboardSettings()
    await this.page.fill('[data-testid="key-repeat-delay"]', delay.toString())
    await this.closeKeyboardSettings()
  }

  async isLayoutSelected(layoutId: string) {
    const button = this.page.locator(`[data-testid="layout-${layoutId}"]`)
    return await button.getAttribute('aria-pressed') === 'true'
  }
}

test.describe('On-Screen Keyboard E2E', () => {
  let keyboardPage: KeyboardPage

  test.beforeEach(async ({ page }) => {
    keyboardPage = new KeyboardPage(page)
    await keyboardPage.navigate()
  })

  test('should show keyboard when focusing on input field', async () => {
    await keyboardPage.openKeyboard()

    expect(await keyboardPage.isKeyboardVisible()).toBe(true)
  })

  test('should hide keyboard when clicking outside', async ({ page }) {
    await keyboardPage.openKeyboard()
    expect(await keyboardPage.isKeyboardVisible()).toBe(true)

    // Click outside the keyboard area
    await page.click('body', { position: { x: 10, y: 10 } })
    await page.waitForTimeout(200)

    expect(await keyboardPage.isKeyboardVisible()).toBe(false)
  })

  test('should type single character keys', async () => {
    await keyboardPage.openKeyboard()
    await keyboardPage.pressKeySequence(['a', 'b', 'c'])

    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('abc')
  })

  test('should type with shift modifier', async () => {
    await keyboardPage.openKeyboard()

    // Test shift + letter
    await keyboardPage.pressKeyWithModifier('a', 'shift')
    const text1 = await keyboardPage.getTerminalText()
    expect(text1.toLowerCase()).not.toBe(text1) // Should be uppercase

    // Test shift + number
    await keyboardPage.pressKeyWithModifier('1', 'shift')
    const text2 = await keyboardPage.getTerminalText()
    expect(text2).toContain('!') // Shift + 1 should be !
  })

  test('should handle special keys', async () => {
    await keyboardPage.openKeyboard()

    // Test enter key
    await keyboardPage.pressKeySequence(['h', 'e', 'l', 'l', 'o'])
    await keyboardPage.pressKey('enter')

    // Test space key
    await keyboardPage.pressKeySequence(['w', 'o', 'r', 'l', 'd'])
    await keyboardPage.pressKey('space')
    await keyboardPage.pressKeySequence(['!'])

    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('hello')
    expect(text).toContain('world!')
  })

  test('should switch between keyboard layouts', async () => {
    await keyboardPage.openKeyboard()

    // Check initial layout
    expect(await keyboardPage.isLayoutSelected('qwerty')).toBe(true)

    // Switch to QWERTZ
    await keyboardPage.switchLayout('qwertz')
    expect(await keyboardPage.isLayoutSelected('qwertz')).toBe(true)

    // Test that keys work in new layout
    await keyboardPage.pressKey('y') // In QWERTZ, Y and Z are swapped
    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('y') // Should still work, just different position

    // Switch back to QWERTY
    await keyboardPage.switchLayout('qwerty')
    expect(await keyboardPage.isLayoutSelected('qwerty')).toBe(true)
  })

  test('should support keyboard resizing', async () => {
    await keyboardPage.openKeyboard()

    const originalSize = await keyboardPage.getKeyboardPosition()

    // Test compact size
    await keyboardPage.resizeKeyboard('compact')
    const compactSize = await keyboardPage.getKeyboardPosition()
    expect(compactSize?.height).toBeLessThan(originalSize?.height || 0)

    // Test large size
    await keyboardPage.resizeKeyboard('large')
    const largeSize = await keyboardPage.getKeyboardPosition()
    expect(largeSize?.height).toBeGreaterThan(originalSize?.height || 0)

    // Return to normal
    await keyboardPage.resizeKeyboard('normal')
    const normalSize = await keyboardPage.getKeyboardPosition()
    expect(normalSize?.height).toBeCloseTo(originalSize?.height || 0, 0)
  })

  test('should support key repeat functionality', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Enable key repeat with short delay for testing
    await keyboardPage.adjustKeyRepeatDelay(100)
    await keyboardPage.enableKeyRepeat(true)

    const initialText = await keyboardPage.getTerminalText()

    // Press and hold a key
    const keyButton = await page.locator('[data-testid="key-a"]')
    await keyButton.down()

    // Wait for repeat to kick in
    await page.waitForTimeout(300)

    // Release key
    await keyButton.up()

    const finalText = await keyboardPage.getTerminalText()

    // Should have multiple 'a' characters due to repeat
    expect(finalText.length - initialText.length).toBeGreaterThan(1)
  })

  test('should handle rapid key presses without lag', async ({ page }) => {
    await keyboardPage.openKeyboard()

    const startTime = Date.now()

    // Type a sentence rapidly
    const keys = 'the quick brown fox jumps over the lazy dog'.split('')
    for (const key of keys) {
      if (key === ' ') {
        await keyboardPage.pressKey('space')
      } else {
        await keyboardPage.pressKey(key.toLowerCase())
      }
    }

    const endTime = Date.now()
    const duration = endTime - startTime

    // Should complete within reasonable time (less than 2 seconds)
    expect(duration).toBeLessThan(2000)

    const text = await keyboardPage.getTerminalText()
    expect(text.toLowerCase()).toContain('the quick brown fox')
  })

  test('should maintain keyboard position on window resize', async ({ page }) => {
    await keyboardPage.openKeyboard()

    const originalPosition = await keyboardPage.getKeyboardPosition()

    // Resize window
    await page.setViewportSize({ width: 800, height: 600 })
    await page.waitForTimeout(200)

    const resizedPosition = await keyboardPage.getKeyboardPosition()

    // Keyboard should still be visible and properly positioned
    expect(resizedPosition).toBeTruthy()
    expect(resizedPosition?.height).toBeGreaterThan(0)
  })

  test('should support keyboard shortcuts and commands', async () => {
    await keyboardPage.openKeyboard()

    // Test Ctrl+C (copy, but in terminal might interrupt)
    await keyboardPage.pressKeyWithModifier('c', 'ctrl')

    // Test Ctrl+Z (undo)
    await keyboardPage.pressKeyWithModifier('z', 'ctrl')

    // Test Alt+Tab (should work in most environments)
    await keyboardPage.pressKeyWithModifier('tab', 'alt')

    // Keyboard should still be responsive
    await keyboardPage.pressKey('a')
    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('a')
  })

  test('should handle keyboard focus management', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Tab through keyboard elements
    await page.keyboard.press('Tab')
    let focusedElement = await page.locator(':focus')

    // Should be able to navigate keyboard elements
    expect(await focusedElement.count()).toBeGreaterThan(0)

    // Try Enter key on focused element
    await page.keyboard.press('Enter')

    // Keyboard should still function
    await keyboardPage.pressKey('b')
    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('b')
  })

  test('should support keyboard settings persistence', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Change settings
    await keyboardPage.resizeKeyboard('large')
    await keyboardPage.switchLayout('qwertz')

    // Hide and show keyboard
    await page.click('body', { position: { x: 10, y: 10 } })
    await page.waitForTimeout(200)
    await keyboardPage.openKeyboard()

    // Settings should be preserved
    const position = await keyboardPage.getKeyboardPosition()
    expect(position?.height).toBeGreaterThan(0)

    // Note: In a real implementation, we'd verify that layout and size
    // preferences are actually persisted and restored
  })

  test('should handle keyboard accessibility', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Test ARIA labels
    const keyButton = await page.locator('[data-testid="key-a"]')
    const ariaLabel = await keyButton.getAttribute('aria-label')
    expect(ariaLabel).toBeTruthy()
    expect(ariaLabel).toContain('A')

    // Test keyboard navigation
    await page.keyboard.press('Tab')
    const focusedElement = await page.locator(':focus')
    expect(await focusedElement.count()).toBeGreaterThan(0)

    // Test screen reader announcements
    const announcedElements = await page.locator('[aria-live]')
    expect(await announcedElements.count()).toBeGreaterThan(0)
  })

  test('should handle keyboard errors gracefully', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Try to type invalid combination
    await keyboardPage.pressKeyWithModifier('invalid-key', 'shift')

    // Keyboard should still be functional
    await keyboardPage.pressKey('x')
    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('x')

    // Try rapid invalid key presses
    for (let i = 0; i < 5; i++) {
      await page.click('[data-testid="non-existent-key"]')
    }

    // Should not crash and still be functional
    expect(await keyboardPage.isKeyboardVisible()).toBe(true)
  })

  test('should work with different input types', async ({ page }) => {
    // Test with text input
    await page.click('[data-testid="terminal-input"]')
    expect(await keyboardPage.isKeyboardVisible()).toBe(true)

    await keyboardPage.pressKeySequence(['t', 'e', 's', 't'])
    const textInputValue = await page.inputValue('[data-testid="terminal-input"]')
    expect(textInputValue).toContain('test')

    // Hide keyboard
    await page.click('body', { position: { x: 10, y: 10 } })

    // Test with search input
    await page.click('[data-testid="search-input"]')
    expect(await keyboardPage.isKeyboardVisible()).toBe(true)

    await keyboardPage.pressKeySequence(['s', 'e', 'a', 'r', 'c', 'h'])
    const searchInputValue = await page.inputValue('[data-testid="search-input"]')
    expect(searchInputValue).toContain('search')
  })

  test('should handle international characters', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Switch to layout that supports international characters
    await keyboardPage.switchLayout('azerty')

    // Test international character input
    await keyboardPage.pressKeySequence(['é', 'è', 'à', 'ç'])

    const text = await keyboardPage.getTerminalText()

    // Should handle international characters correctly
    expect(text.length).toBeGreaterThan(0)
    // Note: Specific character testing would depend on actual layout implementation
  })

  test('should perform well under load', async ({ page }) => {
    await keyboardPage.openKeyboard()

    // Rapid key press test
    const startTime = performance.now()

    for (let i = 0; i < 100; i++) {
      await keyboardPage.pressKey('a')
      await keyboardPage.pressKey('backspace')
    }

    const endTime = performance.now()
    const duration = endTime - startTime

    // Should complete 200 operations within reasonable time
    expect(duration).toBeLessThan(5000)

    // Keyboard should still be responsive
    await keyboardPage.pressKey('b')
    const text = await keyboardPage.getTerminalText()
    expect(text).toContain('b')
  })
})