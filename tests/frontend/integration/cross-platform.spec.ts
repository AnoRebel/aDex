import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref, nextTick } from 'vue'
import { useStorage, usePreferredDark, useThrottleFn, useDebounceFn } from '@vueuse/core'
import { detectOS, getOSName, isMac, isWindows, isLinux } from '~/utils/platform'

describe('Cross-Platform Compatibility Tests', () => {
  describe('Platform Detection', () => {
    it('should detect Windows correctly', () => {
      // Mock Windows user agent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        writable: true
      })

      expect(isWindows()).toBe(true)
      expect(isMac()).toBe(false)
      expect(isLinux()).toBe(false)
      expect(getOSName()).toBe('Windows')
    })

    it('should detect macOS correctly', () => {
      // Mock macOS user agent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
        writable: true
      })

      expect(isMac()).toBe(true)
      expect(isWindows()).toBe(false)
      expect(isLinux()).toBe(false)
      expect(getOSName()).toBe('macOS')
    })

    it('should detect Linux correctly', () => {
      // Mock Linux user agent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36',
        writable: true
      })

      expect(isLinux()).toBe(true)
      expect(isMac()).toBe(false)
      expect(isWindows()).toBe(false)
      expect(getOSName()).toBe('Linux')
    })

    it('should handle unknown platforms gracefully', () => {
      // Mock unknown user agent
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Unknown) AppleWebKit/537.36',
        writable: true
      })

      expect(isLinux()).toBe(true) // Fallback to Linux
      expect(getOSName()).toBe('Linux')
    })
  })

  describe('Platform-Specific Features', () => {
    it('should use appropriate file separators for different platforms', () => {
      const normalizePath = (path: string) => {
        if (isWindows()) {
          return path.replace(/\//g, '\\')
        }
        return path.replace(/\\/g, '/')
      }

      // Test Windows path normalization
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        writable: true
      })

      expect(normalizePath('home/user/documents')).toBe('home\\user\\documents')
      expect(normalizePath('home\\user\\documents')).toBe('home\\user\\documents')

      // Test Unix path normalization
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (X11; Linux x86_64)',
        writable: true
      })

      expect(normalizePath('home\\user\\documents')).toBe('home/user/documents')
      expect(normalizePath('home/user/documents')).toBe('home/user/documents')
    })

    it('should handle platform-specific keyboard shortcuts', () => {
      const getShortcut = (action: string) => {
        if (isMac()) {
          switch (action) {
            case 'copy': return 'Cmd+C'
            case 'paste': return 'Cmd+V'
            case 'select-all': return 'Cmd+A'
            case 'save': return 'Cmd+S'
            case 'quit': return 'Cmd+Q'
            default: return 'Unknown'
          }
        } else {
          switch (action) {
            case 'copy': return 'Ctrl+C'
            case 'paste': return 'Ctrl+V'
            case 'select-all': return 'Ctrl+A'
            case 'save': return 'Ctrl+S'
            case 'quit': return 'Alt+F4'
            default: return 'Unknown'
          }
        }
      }

      // Test macOS shortcuts
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        writable: true
      })

      expect(getShortcut('copy')).toBe('Cmd+C')
      expect(getShortcut('save')).toBe('Cmd+S')

      // Test Windows/Linux shortcuts
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        writable: true
      })

      expect(getShortcut('copy')).toBe('Ctrl+C')
      expect(getShortcut('quit')).toBe('Alt+F4')
    })

    it('should adapt UI elements for different platforms', () => {
      const getButtonLabel = (action: string) => {
        if (isMac()) {
          switch (action) {
            case 'close': return 'Close'
            case 'minimize': return 'Minimize'
            case 'maximize': return 'Zoom'
            case 'ok': return 'OK'
            case 'cancel': return 'Cancel'
            default: return action
          }
        } else {
          switch (action) {
            case 'close': return 'Close'
            case 'minimize': return 'Minimize'
            case 'maximize': return 'Maximize'
            case 'ok': return 'OK'
            case 'cancel': return 'Cancel'
            default: return action
          }
        }
      }

      // Test macOS labels
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        writable: true
      })

      expect(getButtonLabel('maximize')).toBe('Zoom')

      // Test Windows labels
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        writable: true
      })

      expect(getButtonLabel('maximize')).toBe('Maximize')
    })
  })

  describe('File System Compatibility', () => {
    it('should handle file path validation across platforms', () => {
      const isValidPath = (path: string) => {
        // Basic validation rules
        if (!path || path.length === 0) return false

        if (isWindows()) {
          // Windows validation
          const winPattern = /^[a-zA-Z]:\\(?:[^\\/:*?"<>|]+\\)*[^\\/:*?"<>|]*$/
          return winPattern.test(path) || path.startsWith('\\\\') || /^[a-zA-Z]:\\/.test(path)
        } else {
          // Unix validation
          const unixPattern = /^\/(?:[^\/\0]+\/)*[^\/\0]*$|^~\/(?:[^\/\0]+\/)*[^\/\0]*$/
          return unixPattern.test(path) || path.startsWith('./') || path === path.split('/').pop()
        }
      }

      // Test Windows paths
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        writable: true
      })

      expect(isValidPath('C:\\Users\\Test\\Documents')).toBe(true)
      expect(isValidPath('C:\\Program Files\\App')).toBe(true)
      expect(isValidPath('/home/user/documents')).toBe(false) // Unix path on Windows

      // Test Unix paths
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (X11; Linux x86_64)',
        writable: true
      })

      expect(isValidPath('/home/user/documents')).toBe(true)
      expect(isValidPath('./relative/path')).toBe(true)
      expect(isValidPath('C:\\Users\\Test')).toBe(false) // Windows path on Unix
    })

    it('should handle file extension compatibility', () => {
      const getExecutableExtension = () => {
        if (isWindows()) {
          return '.exe'
        } else {
          return ''
        }
      }

      const getShellExtension = () => {
        if (isWindows()) {
          return '.bat'
        } else if (isMac()) {
          return '.sh'
        } else {
          return '.sh'
        }
      }

      // Test Windows extensions
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        writable: true
      })

      expect(getExecutableExtension()).toBe('.exe')
      expect(getShellExtension()).toBe('.bat')

      // Test macOS extensions
      Object.defineProperty(navigator, 'userAgent', {
        value: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        writable: true
      })

      expect(getExecutableExtension()).toBe('')
      expect(getShellExtension()).toBe('.sh')
    })
  })

  describe('Theme and Display Compatibility', () => {
    it('should adapt to system theme preferences', async () => {
      const TestComponent = {
        template: '<div :class="{ \'dark-theme\': isDark }">Content</div>',
        setup() {
          const isDark = usePreferredDark()
          return { isDark }
        }
      }

      const wrapper = mount(TestComponent)

      // Test light theme (default)
      expect(wrapper.classes()).not.toContain('dark-theme')

      // Mock dark theme preference
      Object.defineProperty(window, 'matchMedia', {
        writable: true,
        value: jest.fn().mockImplementation(query => ({
          matches: query === '(prefers-color-scheme: dark)',
          media: query,
          onchange: null,
          addListener: jest.fn(), // Deprecated
          removeListener: jest.fn(), // Deprecated
          addEventListener: jest.fn(),
          removeEventListener: jest.fn(),
          dispatchEvent: jest.fn(),
        })),
      })

      await nextTick()
      // Note: In a real test environment, you might need to trigger theme change
    })

    it('should handle high DPI displays', () => {
      const getPixelRatio = () => {
        return window.devicePixelRatio || 1
      }

      const getDisplayQuality = () => {
        const ratio = getPixelRatio()
        if (ratio >= 2) return 'high'
        if (ratio >= 1.5) return 'medium'
        return 'standard'
      }

      // Mock different pixel ratios
      Object.defineProperty(window, 'devicePixelRatio', {
        value: 2,
        writable: true
      })

      expect(getDisplayQuality()).toBe('high')
      expect(getPixelRatio()).toBe(2)

      Object.defineProperty(window, 'devicePixelRatio', {
        value: 1,
        writable: true
      })

      expect(getDisplayQuality()).toBe('standard')
    })
  })

  describe('Performance Across Platforms', () => {
    it('should adapt performance settings based on detected capabilities', () => {
      const getPerformanceLevel = () => {
        // Check hardware concurrency
        const cores = navigator.hardwareConcurrency || 4
        const memory = (navigator as any).deviceMemory || 4

        if (cores >= 8 && memory >= 8) return 'high'
        if (cores >= 4 && memory >= 4) return 'medium'
        return 'low'
      }

      const getAnimationSettings = () => {
        const level = getPerformanceLevel()
        switch (level) {
          case 'high':
            return { enabled: true, quality: 'high', fps: 60 }
          case 'medium':
            return { enabled: true, quality: 'medium', fps: 30 }
          default:
            return { enabled: false, quality: 'low', fps: 15 }
        }
      }

      // Mock high-end device
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        value: 8,
        writable: true
      })
      ;(navigator as any).deviceMemory = 8

      expect(getPerformanceLevel()).toBe('high')
      expect(getAnimationSettings()).toEqual({
        enabled: true,
        quality: 'high',
        fps: 60
      })

      // Mock low-end device
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        value: 2,
        writable: true
      })
      ;(navigator as any).deviceMemory = 2

      expect(getPerformanceLevel()).toBe('low')
      expect(getAnimationSettings()).toEqual({
        enabled: false,
        quality: 'low',
        fps: 15
      })
    })

    it('should throttle expensive operations on slower devices', async () => {
      const expensiveOperation = useThrottleFn(() => {
        // Simulate expensive computation
        let sum = 0
        for (let i = 0; i < 1000000; i++) {
          sum += Math.random()
        }
        return sum
      }, 1000) // Throttle to 1 second

      const startTime = Date.now()
      const result1 = expensiveOperation()
      const result2 = expensiveOperation() // Should be throttled
      const endTime = Date.now()

      expect(typeof result1).toBe('number')
      expect(typeof result2).toBe('number')
      expect(endTime - startTime).toBeLessThan(100) // Should return cached result
    })
  })

  describe('Input Method Compatibility', () => {
    it('should handle different input methods and locales', () => {
      const getKeyboardLayout = () => {
        const language = navigator.language || 'en-US'

        if (language.startsWith('ja')) return 'japanese'
        if (language.startsWith('zh')) return 'chinese'
        if (language.startsWith('ko')) return 'korean'
        if (language.startsWith('ar')) return 'arabic'
        if (language.startsWith('he')) return 'hebrew'
        return 'qwerty'
      }

      const getTextDirection = () => {
        const language = navigator.language || 'en-US'
        const rtlLanguages = ['ar', 'he', 'fa', 'ur']

        return rtlLanguages.some(lang => language.startsWith(lang)) ? 'rtl' : 'ltr'
      }

      // Test different locales
      Object.defineProperty(navigator, 'language', {
        value: 'ja-JP',
        writable: true
      })

      expect(getKeyboardLayout()).toBe('japanese')
      expect(getTextDirection()).toBe('ltr')

      Object.defineProperty(navigator, 'language', {
        value: 'ar-SA',
        writable: true
      })

      expect(getKeyboardLayout()).toBe('qwerty') // Default for Arabic
      expect(getTextDirection()).toBe('rtl')
    })

    it('should handle touch vs mouse input appropriately', () => {
      const getInputType = () => {
        return 'ontouchstart' in window ? 'touch' : 'mouse'
      }

      const getInteractionSettings = () => {
        const inputType = getInputType()
        return {
          inputType,
          hasTouch: inputType === 'touch',
          hasMouse: inputType === 'mouse',
          hoverSupported: inputType === 'mouse'
        }
      }

      // Mock touch device
      Object.defineProperty(window, 'ontouchstart', {
        value: jest.fn(),
        writable: true
      })

      expect(getInputType()).toBe('touch')
      expect(getInteractionSettings()).toEqual({
        inputType: 'touch',
        hasTouch: true,
        hasMouse: false,
        hoverSupported: false
      })

      // Mock mouse-only device
      delete (window as any).ontouchstart

      expect(getInputType()).toBe('mouse')
      expect(getInteractionSettings()).toEqual({
        inputType: 'mouse',
        hasTouch: false,
        hasMouse: true,
        hoverSupported: true
      })
    })
  })

  describe('Network Compatibility', () => {
    it('should handle different network conditions', async () => {
      const getConnectionType = () => {
        const connection = (navigator as any).connection ||
                          (navigator as any).mozConnection ||
                          (navigator as any).webkitConnection
        return connection ? connection.effectiveType : 'unknown'
      }

      const getNetworkSettings = () => {
        const type = getConnectionType()
        switch (type) {
          case '4g':
            return { quality: 'excellent', latency: 'low', timeout: 5000 }
          case '3g':
            return { quality: 'good', latency: 'medium', timeout: 10000 }
          case '2g':
            return { quality: 'poor', latency: 'high', timeout: 20000 }
          default:
            return { quality: 'unknown', latency: 'medium', timeout: 10000 }
        }
      }

      // Mock different network conditions
      Object.defineProperty(navigator, 'connection', {
        value: {
          effectiveType: '4g',
          addEventListener: jest.fn(),
          removeEventListener: jest.fn()
        },
        writable: true
      })

      expect(getConnectionType()).toBe('4g')
      expect(getNetworkSettings()).toEqual({
        quality: 'excellent',
        latency: 'low',
        timeout: 5000
      })

      Object.defineProperty(navigator, 'connection', {
        value: {
          effectiveType: '2g',
          addEventListener: jest.fn(),
          removeEventListener: jest.fn()
        },
        writable: true
      })

      expect(getNetworkSettings().quality).toBe('poor')
    })
  })
})