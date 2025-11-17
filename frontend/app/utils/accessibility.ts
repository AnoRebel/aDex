/**
 * Accessibility utilities for compliance with WCAG and assistive technologies
 */

import { useStorage, useEventListener, useThrottleFn } from '@vueuse/core'
import type { Ref } from 'vue'

export interface AccessibilityConfig {
  keyboardNavigation: boolean
  screenReaderSupport: boolean
  highContrastMode: boolean
  reducedMotion: boolean
  largeTextMode: boolean
  focusVisible: boolean
  announcements: boolean
  colorBlindSupport: boolean
  fontSize: 'small' | 'medium' | 'large' | 'extra-large'
  colorTheme: 'default' | 'high-contrast' | 'protanopia' | 'deuteranopia' | 'tritanopia'
  focusRingStyle: 'solid' | 'dashed' | 'double' | 'thick'
  announcementDuration: number
}

export interface Announcement {
  id: string
  message: string
  priority: 'polite' | 'assertive' | 'off'
  timestamp: Date
  duration?: number
}

export interface FocusRegion {
  id: string
  element: HTMLElement
  previousFocus?: HTMLElement
  trapped: boolean
}

export class AccessibilityManager {
  private static instance: AccessibilityManager
  private config: Ref<AccessibilityConfig>
  private announcements: Ref<Announcement[]> = useStorage('a11y-announcements', [])
  private focusRegions: Map<string, FocusRegion> = new Map()
  private currentFocusRegion?: FocusRegion
  private keyboardTimer?: number

  private constructor() {
    this.config = useStorage('a11y-config', this.getDefaultConfig())
    this.setupEventListeners()
    this.detectSystemPreferences()
  }

  static getInstance(): AccessibilityManager {
    if (!AccessibilityManager.instance) {
      AccessibilityManager.instance = new AccessibilityManager()
    }
    return AccessibilityManager.instance
  }

  /**
   * Gets current accessibility configuration
   */
  getConfig(): AccessibilityConfig {
    return this.config.value
  }

  /**
   * Updates accessibility configuration
   */
  updateConfig(updates: Partial<AccessibilityConfig>): void {
    this.config.value = { ...this.config.value, ...updates }
    this.applyConfigChanges()
  }

  /**
   * Announces a message to screen readers
   */
  announce(message: string, priority: 'polite' | 'assertive' | 'off' = 'polite', duration?: number): void {
    if (!this.config.value.announcements) return

    const announcement: Announcement = {
      id: this.generateAnnouncementID(),
      message,
      priority,
      timestamp: new Date(),
      duration
    }

    this.announcements.value.unshift(announcement)

    // Limit announcements in memory
    if (this.announcements.value.length > 50) {
      this.announcements.value = this.announcements.value.slice(0, 50)
    }

    this.deliverAnnouncement(announcement)
  }

  /**
   * Creates a focus trap within a container
   */
  trapFocus(element: HTMLElement, id?: string): FocusRegion {
    const focusRegion: FocusRegion = {
      id: id || this.generateRegionID(),
      element,
      trapped: true
    }

    // Store current focus
    focusRegion.previousFocus = document.activeElement as HTMLElement

    // Find all focusable elements
    const focusableElements = this.getFocusableElements(element)
    if (focusableElements.length > 0) {
      focusableElements[0].focus()
    }

    this.focusRegions.set(focusRegion.id, focusRegion)
    this.currentFocusRegion = focusRegion

    return focusRegion
  }

  /**
   * Releases a focus trap
   */
  releaseFocus(regionId: string): void {
    const region = this.focusRegions.get(regionId)
    if (!region) return

    region.trapped = false

    // Restore previous focus
    if (region.previousFocus) {
      region.previousFocus.focus()
    }

    this.focusRegions.delete(regionId)

    if (this.currentFocusRegion?.id === regionId) {
      this.currentFocusRegion = undefined
    }
  }

  /**
   * Gets all focusable elements within a container
   */
  getFocusableElements(container: HTMLElement): HTMLElement[] {
    const focusableSelectors = [
      'button:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      'textarea:not([disabled])',
      'a[href]',
      'area[href]',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable="true"]',
      'summary',
      'iframe',
      'object',
      'embed',
      'audio[controls]',
      'video[controls]',
      '[role="button"]:not([disabled])',
      '[role="link"]',
      '[role="menuitem"]',
      '[role="option"]',
      '[role="tab"]'
    ].join(', ')

    return Array.from(container.querySelectorAll<HTMLElement>(focusableSelectors))
      .filter(element => {
        // Filter out hidden elements
        if (element.offsetParent === null) return false
        if (window.getComputedStyle(element).visibility === 'hidden') return false
        if (element.getAttribute('aria-hidden') === 'true') return false

        return true
      })
  }

  /**
   * Sets focus to an element with proper announcements
   */
  setFocus(element: HTMLElement, message?: string): void {
    element.focus()

    if (message) {
      this.announce(message, 'polite')
    }
  }

  /**
   * Validates ARIA attributes and compliance
   */
  validateAria(element: HTMLElement): {
    valid: boolean
    issues: string[]
    suggestions: string[]
  } {
    const issues: string[] = []
    const suggestions: string[] = []

    // Check for required ARIA attributes
    const role = element.getAttribute('role')
    if (role) {
      const requiredAttributes = this.getRequiredAriaAttributes(role)
      requiredAttributes.forEach(attr => {
        if (!element.hasAttribute(attr)) {
          issues.push(`Missing required ARIA attribute: ${attr}`)
        }
      })
    }

    // Check for proper labeling
    if (element.tagName === 'INPUT' && element.type === 'text') {
      const hasLabel = element.hasAttribute('aria-label') ||
                       element.hasAttribute('aria-labelledby') ||
                       element.labels?.length > 0
      if (!hasLabel) {
        issues.push('Input element missing accessible label')
        suggestions.push('Add aria-label, aria-labelledby, or associate with a <label> element')
      }
    }

    // Check for duplicate IDs
    const id = element.getAttribute('id')
    if (id && document.querySelectorAll(`[id="${id}"]`).length > 1) {
      issues.push(`Duplicate ID found: ${id}`)
    }

    // Check for proper heading structure
    if (element.tagName.startsWith('H')) {
      const level = parseInt(element.tagName.substring(1))
      if (level < 1 || level > 6) {
        issues.push(`Invalid heading level: ${level}`)
      }
    }

    return {
      valid: issues.length === 0,
      issues,
      suggestions
    }
  }

  /**
   * Generates accessibility report
   */
  generateAccessibilityReport(): {
    score: number
    issues: Array<{
      element: string
      type: 'error' | 'warning' | 'info'
      message: string
      suggestion: string
    }>
    config: AccessibilityConfig
    stats: {
      totalElements: number
      focusableElements: number
      headingCount: number
      landmarkCount: number
      imageCount: number
      imagesWithAlt: number
    }
  } {
    const issues: any[] = []
    const allElements = document.querySelectorAll('*')
    const stats = {
      totalElements: allElements.length,
      focusableElements: this.getFocusableElements(document.body).length,
      headingCount: document.querySelectorAll('h1, h2, h3, h4, h5, h6').length,
      landmarkCount: document.querySelectorAll('[role="banner"], [role="navigation"], [role="main"], [role="complementary"], [role="contentinfo"]').length,
      imageCount: document.querySelectorAll('img').length,
      imagesWithAlt: document.querySelectorAll('img[alt]').length
    }

    // Check for missing alt text
    document.querySelectorAll('img:not([alt])').forEach(img => {
      issues.push({
        element: img.tagName,
        type: 'error',
        message: 'Image missing alt text',
        suggestion: 'Add descriptive alt attribute to the image'
      })
    })

    // Check for proper heading structure
    const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6')
    let previousLevel = 0
    headings.forEach(heading => {
      const level = parseInt(heading.tagName.substring(1))
      if (previousLevel > 0 && level > previousLevel + 1) {
        issues.push({
          element: heading.tagName,
          type: 'warning',
          message: `Heading level jump from H${previousLevel} to H${level}`,
          suggestion: 'Use proper heading hierarchy without skipping levels'
        })
      }
      previousLevel = level
    })

    // Check for form labeling
    document.querySelectorAll('input, select, textarea').forEach(input => {
      const hasLabel = input.hasAttribute('aria-label') ||
                       input.hasAttribute('aria-labelledby') ||
                       input.labels?.length > 0 ||
                       input.getAttribute('placeholder')
      if (!hasLabel) {
        issues.push({
          element: input.tagName,
          type: 'error',
          message: 'Form control missing accessible label',
          suggestion: 'Add label, aria-label, aria-labelledby, or meaningful placeholder'
        })
      }
    })

    // Calculate accessibility score
    const maxScore = 100
    const deductions = issues.length * 5
    const score = Math.max(0, maxScore - deductions)

    return {
      score,
      issues,
      config: this.config.value,
      stats
    }
  }

  /**
   * Applies high contrast mode
   */
  setHighContrastMode(enabled: boolean): void {
    this.updateConfig({ highContrastMode: enabled })

    if (enabled) {
      document.documentElement.classList.add('high-contrast')
      document.documentElement.setAttribute('data-theme', 'high-contrast')
    } else {
      document.documentElement.classList.remove('high-contrast')
      document.documentElement.removeAttribute('data-theme')
    }
  }

  /**
   * Sets reduced motion preference
   */
  setReducedMotion(enabled: boolean): void {
    this.updateConfig({ reducedMotion: enabled })

    if (enabled) {
      document.documentElement.setAttribute('data-reduced-motion', 'true')
    } else {
      document.documentElement.removeAttribute('data-reduced-motion')
    }
  }

  /**
   * Increases text size
   */
  setFontSize(size: 'small' | 'medium' | 'large' | 'extra-large'): void {
    this.updateConfig({ fontSize: size })
    document.documentElement.setAttribute('data-font-size', size)
  }

  /**
   * Enables focus visible styling
   */
  setFocusVisible(enabled: boolean): void {
    this.updateConfig({ focusVisible: enabled })

    if (enabled) {
      document.documentElement.classList.add('focus-visible-enabled')
    } else {
      document.documentElement.classList.remove('focus-visible-enabled')
    }
  }

  // Private helper methods

  private getDefaultConfig(): AccessibilityConfig {
    return {
      keyboardNavigation: true,
      screenReaderSupport: true,
      highContrastMode: false,
      reducedMotion: false,
      largeTextMode: false,
      focusVisible: true,
      announcements: true,
      colorBlindSupport: true,
      fontSize: 'medium',
      colorTheme: 'default',
      focusRingStyle: 'solid',
      announcementDuration: 5000
    }
  }

  private setupEventListeners(): void {
    // Keyboard navigation
    if (this.config.value.keyboardNavigation) {
      useEventListener(document, 'keydown', this.handleKeyboard.bind(this))
    }

    // Focus management
    useEventListener(document, 'focusin', this.handleFocusIn.bind(this))
    useEventListener(document, 'focusout', this.handleFocusOut.bind(this))

    // Detect preference changes
    useEventListener(window, 'change', this.detectSystemPreferences.bind(this))
  }

  private detectSystemPreferences(): void {
    // Check for reduced motion preference
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (prefersReducedMotion && !this.config.value.reducedMotion) {
      this.setReducedMotion(true)
    }

    // Check for high contrast preference
    const prefersHighContrast = window.matchMedia('(prefers-contrast: high)').matches
    if (prefersHighContrast && !this.config.value.highContrastMode) {
      this.setHighContrastMode(true)
    }

    // Check for large text preference
    const prefersLargeText = window.matchMedia('(prefers-reduced-data: reduce)').matches
    if (prefersLargeText && this.config.value.fontSize === 'medium') {
      this.setFontSize('large')
    }
  }

  private handleKeyboard(event: KeyboardEvent): void {
    // Handle Tab navigation
    if (event.key === 'Tab') {
      this.handleTabNavigation(event)
    }

    // Handle Escape key
    if (event.key === 'Escape') {
      this.handleEscapeKey(event)
    }

    // Handle keyboard shortcuts
    if (event.altKey) {
      this.handleKeyboardShortcut(event)
    }
  }

  private handleTabNavigation(event: KeyboardEvent): void {
    // Clear existing timer
    if (this.keyboardTimer) {
      clearTimeout(this.keyboardTimer)
    }

    // Set keyboard navigation class
    document.body.classList.add('keyboard-navigation')

    // Clear the class after a delay
    this.keyboardTimer = window.setTimeout(() => {
      document.body.classList.remove('keyboard-navigation')
    }, 100)

    // Handle focus trap
    if (this.currentFocusRegion?.trapped) {
      const focusableElements = this.getFocusableElements(this.currentFocusRegion.element)
      const currentIndex = focusableElements.indexOf(document.activeElement as HTMLElement)

      if (event.shiftKey) {
        // Shift+Tab - move backwards
        if (currentIndex <= 0) {
          event.preventDefault()
          focusableElements[focusableElements.length - 1].focus()
        }
      } else {
        // Tab - move forwards
        if (currentIndex >= focusableElements.length - 1) {
          event.preventDefault()
          focusableElements[0].focus()
        }
      }
    }
  }

  private handleEscapeKey(event: KeyboardEvent): void {
    // Release current focus trap
    if (this.currentFocusRegion?.trapped) {
      this.releaseFocus(this.currentFocusRegion.id)
      this.announce('Focus trap released', 'polite')
    }
  }

  private handleKeyboardShortcut(event: KeyboardEvent): void {
    // Common accessibility shortcuts
    switch (event.key) {
      case '1':
        // Alt+1: Jump to main content
        event.preventDefault()
        const main = document.querySelector('main, [role="main"]')
        if (main) {
          this.setFocus(main as HTMLElement, 'Main content')
        }
        break
      case '2':
        // Alt+2: Jump to navigation
        event.preventDefault()
        const nav = document.querySelector('nav, [role="navigation"]')
        if (nav) {
          this.setFocus(nav as HTMLElement, 'Navigation')
        }
        break
      case '3':
        // Alt+3: Jump to search
        event.preventDefault()
        const search = document.querySelector('[role="search"], input[type="search"]')
        if (search) {
          this.setFocus(search as HTMLElement, 'Search')
        }
        break
    }
  }

  private handleFocusIn(event: FocusEvent): void {
    const target = event.target as HTMLElement
    if (this.config.value.focusVisible) {
      target.classList.add('focus-visible')
    }
  }

  private handleFocusOut(event: FocusEvent): void {
    const target = event.target as HTMLElement
    if (this.config.value.focusVisible) {
      target.classList.remove('focus-visible')
    }
  }

  private applyConfigChanges(): void {
    const config = this.config.value

    // Apply font size
    this.setFontSize(config.fontSize)

    // Apply high contrast
    if (config.highContrastMode) {
      this.setHighContrastMode(true)
    }

    // Apply reduced motion
    if (config.reducedMotion) {
      this.setReducedMotion(true)
    }

    // Apply focus visible
    if (config.focusVisible) {
      this.setFocusVisible(true)
    }
  }

  private deliverAnnouncement(announcement: Announcement): void {
    const container = this.getAnnouncementContainer(announcement.priority)
    if (!container) return

    const message = document.createElement('div')
    message.setAttribute('aria-live', announcement.priority)
    message.setAttribute('aria-atomic', 'true')
    message.textContent = announcement.message
    message.style.position = 'absolute'
    message.style.left = '-10000px'
    message.style.width = '1px'
    message.style.height = '1px'
    message.style.overflow = 'hidden'

    container.appendChild(message)

    // Remove after duration
    const duration = announcement.duration || this.config.value.announcementDuration
    setTimeout(() => {
      if (message.parentNode) {
        message.parentNode.removeChild(message)
      }
    }, duration)
  }

  private getAnnouncementContainer(priority: 'polite' | 'assertive' | 'off'): HTMLElement | null {
    if (priority === 'off') return null

    const id = `a11y-announcements-${priority}`
    let container = document.getElementById(id)

    if (!container) {
      container = document.createElement('div')
      container.id = id
      container.setAttribute('aria-live', priority)
      container.setAttribute('aria-atomic', 'true')
      container.style.position = 'absolute'
      container.style.left = '-10000px'
      container.style.width = '1px'
      container.style.height = '1px'
      container.style.overflow = 'hidden'
      document.body.appendChild(container)
    }

    return container
  }

  private getRequiredAriaAttributes(role: string): string[] {
    const requiredAttributes: Record<string, string[]> = {
      'button': [],
      'link': [],
      'textbox': ['aria-label'],
      'checkbox': ['aria-checked'],
      'radio': ['aria-checked'],
      'menu': ['aria-label'],
      'menuitem': [],
      'tab': ['aria-selected'],
      'tabpanel': ['aria-labelledby'],
      'listbox': ['aria-label'],
      'option': ['aria-selected'],
      'grid': ['aria-label'],
      'gridcell': ['aria-colindex', 'aria-rowindex'],
      'tree': ['aria-label'],
      'treeitem': ['aria-level', 'aria-selected'],
      'dialog': ['aria-label'],
      'alert': ['aria-live'],
      'log': ['aria-live'],
      'status': ['aria-live'],
      'timer': ['aria-live'],
      'marquee': ['aria-live'],
      'application': ['aria-label'],
      'article': ['aria-label'],
      'document': ['aria-label'],
      'feed': ['aria-label'],
      'figure': ['aria-label'],
      'group': ['aria-label'],
      'img': ['aria-label'],
      'list': ['aria-label'],
      'listitem': [],
      'navigation': ['aria-label'],
      'region': ['aria-label'],
      'row': ['aria-rowindex'],
      'rowgroup': ['aria-label'],
      'table': ['aria-label'],
      'tablist': ['aria-label'],
      'term': [],
      'toolbar': ['aria-label'],
      'tooltip': []
    }

    return requiredAttributes[role] || []
  }

  private generateAnnouncementID(): string {
    return `announcement_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private generateRegionID(): string {
    return `region_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }
}

/**
 * Vue composable for accessibility features
 */
export function useAccessibility() {
  const a11y = AccessibilityManager.getInstance()

  const announce = (message: string, priority?: 'polite' | 'assertive' | 'off', duration?: number) => {
    return a11y.announce(message, priority, duration)
  }

  const trapFocus = (element: HTMLElement, id?: string) => {
    return a11y.trapFocus(element, id)
  }

  const releaseFocus = (regionId: string) => {
    return a11y.releaseFocus(regionId)
  }

  const setFocus = (element: HTMLElement, message?: string) => {
    return a11y.setFocus(element, message)
  }

  const validateAria = (element: HTMLElement) => {
    return a11y.validateAria(element)
  }

  const generateReport = () => {
    return a11y.generateAccessibilityReport()
  }

  const getFocusableElements = (container: HTMLElement) => {
    return a11y.getFocusableElements(container)
  }

  const config = a11y.getConfig()
  const updateConfig = a11y.updateConfig.bind(a11y)

  return {
    announce,
    trapFocus,
    releaseFocus,
    setFocus,
    validateAria,
    generateReport,
    getFocusableElements,
    config: readonly(config),
    updateConfig
  }
}

/**
 * Accessibility directive for Vue
 */
export const vA11y = {
  mounted(el: HTMLElement, binding: any) {
    const { value } = binding
    if (!value) return

    // Apply ARIA attributes
    Object.entries(value).forEach(([attr, val]) => {
      if (attr.startsWith('aria-')) {
        el.setAttribute(attr, String(val))
      }
    })

    // Add keyboard event listeners if needed
    if (value.keyboardNavigation) {
      el.addEventListener('keydown', handleA11yKeydown)
      ;(el as any)._a11yKeydownHandler = handleA11yKeydown
    }
  },

  unmounted(el: HTMLElement) {
    // Clean up event listeners
    if ((el as any)._a11yKeydownHandler) {
      el.removeEventListener('keydown', (el as any)._a11yKeydownHandler)
      delete (el as any)._a11yKeydownHandler
    }
  }
}

function handleA11yKeydown(event: KeyboardEvent) {
  const el = event.target as HTMLElement

  switch (event.key) {
    case 'Enter':
    case ' ':
      if (el.getAttribute('role') === 'button' || el.tagName === 'BUTTON') {
        event.preventDefault()
        el.click()
      }
      break
  }
}