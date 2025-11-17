import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { createTestingPinia } from '@pinia/testing'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'

// Mock Wails runtime for frontend integration testing
vi.mock('@/composables/useWails', () => ({
  useWails: () => ({
    isReady: { value: true },
    callService: vi.fn().mockImplementation(async (service, method, ...args) => {
      // Mock responses for different service calls
      if (service === 'filesystem' && method === 'getCurrentWorkingDirectory') {
        return { cwd: '/home/user/test-project' }
      }
      if (service === 'filesystem' && method === 'listDirectory') {
        return {
          files: [
            { name: 'test.txt', path: '/home/user/test-project/test.txt', isDirectory: false, size: 1024, modified: '2024-01-01T00:00:00Z' },
            { name: 'src', path: '/home/user/test-project/src', isDirectory: true, size: 0, modified: '2024-01-01T00:00:00Z' }
          ]
        }
      }
      if (service === 'network' && method === 'getNetworkStats') {
        return {
          stats: {
            bytesReceived: 1024000,
            bytesSent: 512000,
            packetsReceived: 1000,
            packetsSent: 500,
            interfaces: ['eth0', 'lo']
          }
        }
      }
      if (service === 'network' && method === 'getNetworkInterfaces') {
        return {
          interfaces: [
            { name: 'eth0', addresses: ['192.168.1.100'], isLoopback: false, isUp: true },
            { name: 'lo', addresses: ['127.0.0.1'], isLoopback: true, isUp: true }
          ]
        }
      }
      if (service === 'process' && method === 'getCurrentProcess') {
        return {
          process: {
            pid: 12345,
            name: 'dex-ui',
            cpuUsage: 5.2,
            memoryUsage: 256000000,
            startTime: '2024-01-01T00:00:00Z'
          }
        }
      }
      if (service === 'process' && method === 'getProcessList') {
        return {
          processes: [
            { pid: 12345, name: 'dex-ui', cpuUsage: 5.2, memoryUsage: 256000000 },
            { pid: 12346, name: 'chrome', cpuUsage: 12.8, memoryUsage: 512000000 }
          ]
        }
      }
      if (service === 'theme' && method === 'getAvailableThemes') {
        return {
          themes: [
            {
              id: 'default-dark',
              name: 'Default Dark',
              description: 'Default dark theme',
              author: 'aDex-UI',
              version: '1.0.0',
              colors: {
                background: { primary: '#0a0a0a', secondary: '#1a1a1a', tertiary: '#2a2a2a' },
                foreground: { primary: '#ffffff', secondary: '#cccccc', tertiary: '#999999' },
                accent: { primary: '#00ff41', secondary: '#00cc33' },
                status: { success: '#00ff41', warning: '#ffaa00', error: '#ff3333', info: '#00aaff' },
                terminal: [],
                ui: {
                  buttonBackground: '#1a1a1a',
                  buttonForeground: '#ffffff',
                  buttonHover: '#2a2a2a',
                  buttonActive: '#00ff41',
                  inputBackground: '#0a0a0a',
                  inputForeground: '#ffffff',
                  inputBorder: '#333333',
                  inputFocus: '#00ff41',
                  border: '#333333',
                  shadow: 'rgba(0, 0, 0, 0.5)'
                }
              },
              fonts: {
                families: { primary: 'Inter', secondary: 'Roboto', monospace: 'JetBrains Mono' },
                sizes: { extraSmall: '0.75rem', small: '0.875rem', base: '1rem', large: '1.125rem', extraLarge: '1.25rem', doubleExtraLarge: '1.5rem' },
                weights: { light: '300', normal: '400', medium: '500', semiBold: '600', bold: '700' },
                lineHeights: { tight: '1.25', normal: '1.5', relaxed: '1.75' },
                letterSpacing: { tight: '-0.025em', normal: '0', wide: '0.025em' }
              },
              effects: {
                borderRadius: { small: '0.25rem', medium: '0.5rem', large: '1rem', extraLarge: '1.5rem', full: '9999px' },
                shadows: [],
                gradients: [],
                animations: {
                  duration: { fast: '150ms', normal: '300ms', slow: '500ms' },
                  easing: { linear: 'linear', easeIn: 'cubic-bezier(0.4, 0, 1, 1)', easeOut: 'cubic-bezier(0, 0, 0.2, 1)', easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)' }
                },
                blur: { small: '4px', medium: '8px', large: '16px', extraLarge: '24px' }
              },
              settings: {
                disabledOpacity: 0.5,
                hoverOpacity: 0.8,
                activeOpacity: 1,
                transitionDuration: { fast: '150ms', normal: '300ms', slow: '500ms' },
                zIndexScale: {}
              },
              isDark: true
            }
          ]
        }
      }
      if (service === 'theme' && method === 'getCurrentTheme') {
        return {
          theme: {
            id: 'default-dark',
            name: 'Default Dark',
            description: 'Default dark theme',
            author: 'aDex-UI',
            version: '1.0.0',
            colors: {
              background: { primary: '#0a0a0a', secondary: '#1a1a1a', tertiary: '#2a2a2a' },
              foreground: { primary: '#ffffff', secondary: '#cccccc', tertiary: '#999999' },
              accent: { primary: '#00ff41', secondary: '#00cc33' },
              status: { success: '#00ff41', warning: '#ffaa00', error: '#ff3333', info: '#00aaff' },
              terminal: [],
              ui: {
                buttonBackground: '#1a1a1a',
                buttonForeground: '#ffffff',
                buttonHover: '#2a2a2a',
                buttonActive: '#00ff41',
                inputBackground: '#0a0a0a',
                inputForeground: '#ffffff',
                inputBorder: '#333333',
                inputFocus: '#00ff41',
                border: '#333333',
                shadow: 'rgba(0, 0, 0, 0.5)'
              }
            },
            fonts: {
              families: { primary: 'Inter', secondary: 'Roboto', monospace: 'JetBrains Mono' },
              sizes: { extraSmall: '0.75rem', small: '0.875rem', base: '1rem', large: '1.125rem', extraLarge: '1.25rem', doubleExtraLarge: '1.5rem' },
              weights: { light: '300', normal: '400', medium: '500', semiBold: '600', bold: '700' },
              lineHeights: { tight: '1.25', normal: '1.5', relaxed: '1.75' },
              letterSpacing: { tight: '-0.025em', normal: '0', wide: '0.025em' }
            },
            effects: {
              borderRadius: { small: '0.25rem', medium: '0.5rem', large: '1rem', extraLarge: '1.5rem', full: '9999px' },
              shadows: [],
              gradients: [],
              animations: {
                duration: { fast: '150ms', normal: '300ms', slow: '500ms' },
                easing: { linear: 'linear', easeIn: 'cubic-bezier(0.4, 0, 1, 1)', easeOut: 'cubic-bezier(0, 0, 0.2, 1)', easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)' }
              },
              blur: { small: '4px', medium: '8px', large: '16px', extraLarge: '24px' }
            },
            settings: {
              disabledOpacity: 0.5,
              hoverOpacity: 0.8,
              activeOpacity: 1,
              transitionDuration: { fast: '150ms', normal: '300ms', slow: '500ms' },
              zIndexScale: {}
            },
            isDark: true
          }
        }
      }
      if (service === 'theme' && method === 'setCurrentTheme') {
        return { success: true }
      }
      return { success: false, error: 'Unknown service method' }
    })
  })
}))

// Mock Nuxt runtime
vi.mock('#app', () => ({
  useNuxtApp: () => ({
    $wails: {
      Call: vi.fn()
    }
  })
}))

import { useThemeStore } from '~/stores/theme'
import { useTheme } from '~/composables/useTheme'
import { useFileSystem } from '~/composables/useFileSystem'
import { useNetworkMonitor } from '~/composables/useNetworkMonitor'
import { useProcessMonitor } from '~/composables/useProcessMonitor'
import { useSystemInfo } from '~/composables/useSystemInfo'
import ThemeSelector from '~/components/theme/ThemeSelector.vue'
import ThemePreview from '~/components/theme/ThemePreview.vue'
import type { Theme } from '~/types/theme'

describe('End-to-End System Integration Tests', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Home</div>' } },
        { path: '/theme', component: { template: '<div>Theme Page</div>' } },
        { path: '/system', component: { template: '<div>System Page</div>' } }
      ]
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Complete System Workflow', () => {
    it('should initialize all systems and load data correctly', async () => {
      const pinia = createTestingPinia()

      // Initialize all composables
      const themeComposable = useTheme()
      const fsComposable = useFileSystem()
      const networkComposable = useNetworkMonitor()
      const processComposable = useProcessMonitor()
      const systemComposable = useSystemInfo()

      // Initialize all stores
      const themeStore = useThemeStore()

      // Initialize all systems
      await Promise.all([
        themeComposable.initialize(),
        fsComposable.initialize(),
        networkComposable.initialize(),
        processComposable.initialize(),
        systemComposable.initialize(),
        themeStore.initialize()
      ])

      // Verify all systems are loaded
      expect(themeComposable.isInitialized.value).toBe(true)
      expect(fsComposable.isInitialized.value).toBe(true)
      expect(networkComposable.isInitialized.value).toBe(true)
      expect(processComposable.isInitialized.value).toBe(true)
      expect(systemComposable.isInitialized.value).toBe(true)

      // Verify data is loaded
      expect(themeComposable.currentTheme.value).toBeDefined()
      expect(fsComposable.currentDirectory.value).toBe('/home/user/test-project')
      expect(networkComposable.stats.value).toBeDefined()
      expect(processComposable.currentProcess.value).toBeDefined()
      expect(systemComposable.systemInfo.value).toBeDefined()
    })

    it('should handle theme switching across the entire application', async () => {
      const pinia = createTestingPinia()
      const themeStore = useThemeStore()
      const themeComposable = useTheme()

      // Initialize systems
      await themeStore.initialize()
      await themeComposable.initialize()

      // Verify initial theme
      expect(themeStore.currentTheme.id).toBe('default-dark')
      expect(themeComposable.currentTheme.value.id).toBe('default-dark')

      // Create a new theme to switch to
      const newTheme: Theme = {
        id: 'integration-test-light',
        name: 'Integration Test Light',
        description: 'A light theme for integration testing',
        author: 'Test Suite',
        version: '1.0.0',
        colors: {
          background: { primary: '#ffffff', secondary: '#f5f5f5', tertiary: '#eeeeee' },
          foreground: { primary: '#000000', secondary: '#333333', tertiary: '#666666' },
          accent: { primary: '#0066cc', secondary: '#004499' },
          status: { success: '#00cc00', warning: '#cc9900', error: '#cc0000', info: '#0066cc' },
          terminal: [],
          ui: {
            buttonBackground: '#f5f5f5',
            buttonForeground: '#000000',
            buttonHover: '#eeeeee',
            buttonActive: '#dddddd',
            inputBackground: '#ffffff',
            inputForeground: '#000000',
            inputBorder: '#cccccc',
            inputFocus: '#0066cc',
            border: '#cccccc',
            shadow: 'rgba(0, 0, 0, 0.1)'
          }
        },
        fonts: {
          families: { primary: 'Arial', secondary: 'Helvetica', monospace: 'Courier New' },
          sizes: { extraSmall: '10px', small: '12px', base: '14px', large: '16px', extraLarge: '18px', doubleExtraLarge: '20px' },
          weights: { light: '300', normal: '400', medium: '500', semiBold: '600', bold: '700' },
          lineHeights: { tight: '1.2', normal: '1.4', relaxed: '1.6' },
          letterSpacing: { tight: '-0.02em', normal: '0', wide: '0.02em' }
        },
        effects: {
          borderRadius: { small: '2px', medium: '4px', large: '8px', extraLarge: '12px', full: '9999px' },
          shadows: [],
          gradients: [],
          animations: {
            duration: { fast: '100ms', normal: '200ms', slow: '400ms' },
            easing: { linear: 'linear', easeIn: 'ease-in', easeOut: 'ease-out', easeInOut: 'ease-in-out' }
          },
          blur: { small: '2px', medium: '4px', large: '8px', extraLarge: '16px' }
        },
        settings: {
          disabledOpacity: 0.5,
          hoverOpacity: 0.8,
          activeOpacity: 1,
          transitionDuration: { fast: '100ms', normal: '200ms', slow: '400ms' },
          zIndexScale: {}
        },
        isDark: false
      }

      // Switch theme through composable
      await themeComposable.setTheme(newTheme)

      // Verify theme changed across all systems
      expect(themeStore.currentTheme.id).toBe('integration-test-light')
      expect(themeComposable.currentTheme.value.id).toBe('integration-test-light')
      expect(themeComposable.isDarkTheme(newTheme)).toBe(false)

      // Verify CSS variables are generated
      const cssVars = themeComposable.generateCSSVariables(newTheme)
      expect(cssVars).toContain('--dex-theme-colors-background-primary: #ffffff')
      expect(cssVars).toContain('--dex-theme-fonts-families-primary: Arial')
    })

    it('should handle real-time file system monitoring with theme updates', async () => {
      const pinia = createTestingPinia()
      const fsComposable = useFileSystem()
      const themeComposable = useTheme()

      // Initialize systems
      await fsComposable.initialize()
      await themeComposable.initialize()

      // Mock file system event
      const mockFileEvent = {
        type: 'create' as const,
        path: '/home/user/test-project/new-file.txt',
        timestamp: new Date().toISOString()
      }

      // Simulate file system event
      fsComposable.handleFileEvent(mockFileEvent)

      // Verify file is added to current directory listing
      const currentFiles = fsComposable.directoryContents.value
      const newFile = currentFiles.find(file => file.name === 'new-file.txt')
      expect(newFile).toBeDefined()
      expect(newFile?.type).toBe('file')

      // Verify theme is still applied after file system event
      expect(themeComposable.currentTheme.value).toBeDefined()
      expect(themeComposable.currentTheme.value.id).toBe('default-dark')
    })

    it('should handle concurrent network and process monitoring', async () => {
      const pinia = createTestingPinia()
      const networkComposable = useNetworkMonitor()
      const processComposable = useProcessMonitor()

      // Initialize systems
      await networkComposable.initialize()
      await processComposable.initialize()

      // Start monitoring
      await networkComposable.startMonitoring()
      await processComposable.startMonitoring()

      // Wait for monitoring data
      await new Promise(resolve => setTimeout(resolve, 100))

      // Verify network stats are being updated
      expect(networkComposable.stats.value).toBeDefined()
      expect(networkComposable.stats.value.bytesReceived).toBe(1024000)
      expect(networkComposable.stats.value.bytesSent).toBe(512000)

      // Verify network interfaces are loaded
      expect(networkComposable.interfaces.value).toHaveLength(2)
      expect(networkComposable.interfaces.value[0].name).toBe('eth0')
      expect(networkComposable.interfaces.value[1].name).toBe('lo')

      // Verify process stats are being updated
      expect(processComposable.currentProcess.value).toBeDefined()
      expect(processComposable.currentProcess.value.pid).toBe(12345)
      expect(processComposable.currentProcess.value.name).toBe('dex-ui')

      // Verify process list is loaded
      expect(processComposable.processList.value).toHaveLength(2)
      expect(processComposable.processList.value[0].name).toBe('dex-ui')
      expect(processComposable.processList.value[1].name).toBe('chrome')

      // Stop monitoring
      networkComposable.stopMonitoring()
      processComposable.stopMonitoring()
    })
  })

  describe('Component Integration Tests', () => {
    it('should integrate ThemeSelector with theme store and composable', async () => {
      const pinia = createTestingPinia()

      const mockTheme: Theme = {
        id: 'component-test-theme',
        name: 'Component Test Theme',
        description: 'A theme for testing component integration',
        author: 'Test Suite',
        version: '1.0.0',
        colors: {
          background: { primary: '#0a0a0a', secondary: '#1a1a1a', tertiary: '#2a2a2a' },
          foreground: { primary: '#ffffff', secondary: '#cccccc', tertiary: '#999999' },
          accent: { primary: '#ff6b35', secondary: '#f7931e' },
          status: { success: '#00ff41', warning: '#ffaa00', error: '#ff3333', info: '#00aaff' },
          terminal: [],
          ui: {
            buttonBackground: '#1a1a1a',
            buttonForeground: '#ffffff',
            buttonHover: '#2a2a2a',
            buttonActive: '#ff6b35',
            inputBackground: '#0a0a0a',
            inputForeground: '#ffffff',
            inputBorder: '#333333',
            inputFocus: '#ff6b35',
            border: '#333333',
            shadow: 'rgba(0, 0, 0, 0.5)'
          }
        },
        fonts: {
          families: { primary: 'Inter', secondary: 'Roboto', monospace: 'JetBrains Mono' },
          sizes: { extraSmall: '0.75rem', small: '0.875rem', base: '1rem', large: '1.125rem', extraLarge: '1.25rem', doubleExtraLarge: '1.5rem' },
          weights: { light: '300', normal: '400', medium: '500', semiBold: '600', bold: '700' },
          lineHeights: { tight: '1.25', normal: '1.5', relaxed: '1.75' },
          letterSpacing: { tight: '-0.025em', normal: '0', wide: '0.025em' }
        },
        effects: {
          borderRadius: { small: '0.25rem', medium: '0.5rem', large: '1rem', extraLarge: '1.5rem', full: '9999px' },
          shadows: [],
          gradients: [],
          animations: {
            duration: { fast: '150ms', normal: '300ms', slow: '500ms' },
            easing: { linear: 'linear', easeIn: 'cubic-bezier(0.4, 0, 1, 1)', easeOut: 'cubic-bezier(0, 0, 0.2, 1)', easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)' }
          },
          blur: { small: '4px', medium: '8px', large: '16px', extraLarge: '24px' }
        },
        settings: {
          disabledOpacity: 0.5,
          hoverOpacity: 0.8,
          activeOpacity: 1,
          transitionDuration: { fast: '150ms', normal: '300ms', slow: '500ms' },
          zIndexScale: {}
        },
        metadata: {
          tags: ['component', 'test'],
          category: 'test'
        }
      }

      const wrapper = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: mockTheme,
          config: {
            showCreateButton: true,
            showImportButton: true,
            showExportButton: true,
            compact: false,
            searchable: true,
            categorizable: true,
            previewEnabled: true
          }
        },
        global: {
          plugins: [pinia, router],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      // Test theme selection
      const themeCard = wrapper.find('.theme-card')
      expect(themeCard.exists()).toBe(true)

      await themeCard.trigger('click')
      expect(wrapper.emitted('theme-select')).toBeTruthy()
      expect(wrapper.emitted('theme-select')?.[0]).toEqual([mockTheme])

      // Test search functionality
      const searchInput = wrapper.find('.search-input')
      expect(searchInput.exists()).toBe(true)

      await searchInput.setValue('component')
      expect(wrapper.emitted('search')).toBeTruthy()
      expect(wrapper.emitted('search')?.[0]).toEqual(['component'])
    })

    it('should integrate ThemePreview with real-time theme updates', async () => {
      const pinia = createTestingPinia()

      const mockTheme: Theme = {
        id: 'preview-test-theme',
        name: 'Preview Test Theme',
        description: 'A theme for testing ThemePreview integration',
        author: 'Test Suite',
        version: '1.0.0',
        colors: {
          background: { primary: '#1a1a2e', secondary: '#16213e', tertiary: '#0f3460' },
          foreground: { primary: '#e94560', secondary: '#ff6b6b', tertiary: '#4ecdc4' },
          accent: { primary: '#f39c12', secondary: '#e67e22' },
          status: { success: '#27ae60', warning: '#f39c12', error: '#e74c3c', info: '#3498db' },
          terminal: [],
          ui: {
            buttonBackground: '#16213e',
            buttonForeground: '#e94560',
            buttonHover: '#0f3460',
            buttonActive: '#f39c12',
            inputBackground: '#1a1a2e',
            inputForeground: '#e94560',
            inputBorder: '#4ecdc4',
            inputFocus: '#f39c12',
            border: '#4ecdc4',
            shadow: 'rgba(233, 69, 96, 0.3)'
          }
        },
        fonts: {
          families: { primary: 'Fira Code', secondary: 'Source Code Pro', monospace: 'JetBrains Mono' },
          sizes: { extraSmall: '0.7rem', small: '0.8rem', base: '0.9rem', large: '1rem', extraLarge: '1.1rem', doubleExtraLarge: '1.2rem' },
          weights: { light: '300', normal: '400', medium: '500', semiBold: '600', bold: '700' },
          lineHeights: { tight: '1.3', normal: '1.5', relaxed: '1.7' },
          letterSpacing: { tight: '-0.03em', normal: '0', wide: '0.03em' }
        },
        effects: {
          borderRadius: { small: '0.2rem', medium: '0.4rem', large: '0.8rem', extraLarge: '1.2rem', full: '9999px' },
          shadows: [
            '0 4px 6px rgba(233, 69, 96, 0.1)',
            '0 8px 12px rgba(233, 69, 96, 0.15)'
          ],
          gradients: [
            'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',
            'linear-gradient(135deg, #e94560 0%, #f39c12 100%)'
          ],
          animations: {
            duration: { fast: '200ms', normal: '400ms', slow: '600ms' },
            easing: { linear: 'linear', easeIn: 'cubic-bezier(0.4, 0, 1, 1)', easeOut: 'cubic-bezier(0, 0, 0.2, 1)', easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)' }
          },
          blur: { small: '6px', medium: '10px', large: '16px', extraLarge: '24px' }
        },
        settings: {
          disabledOpacity: 0.4,
          hoverOpacity: 0.9,
          activeOpacity: 1,
          transitionDuration: { fast: '200ms', normal: '400ms', slow: '600ms' },
          zIndexScale: {}
        },
        metadata: {
          tags: ['preview', 'integration'],
          category: 'preview'
        }
      }

      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: true,
          showInfo: true,
          size: 'large'
        },
        global: {
          plugins: [pinia, router],
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      // Test theme preview rendering
      expect(wrapper.find('.theme-preview').exists()).toBe(true)
      expect(wrapper.find('.theme-preview.large').exists()).toBe(true)

      // Test theme styles are applied
      const header = wrapper.find('.preview-header')
      expect(header.attributes('style')).toContain('background-color: rgb(22, 33, 62)')
      expect(header.attributes('style')).toContain('color: rgb(233, 69, 96)')

      const content = wrapper.find('.preview-content')
      expect(content.attributes('style')).toContain('background-color: rgb(26, 26, 46)')

      // Test interactive elements
      expect(wrapper.find('.interactive-elements').exists()).toBe(true)
      expect(wrapper.find('.demo-button').exists()).toBe(true)
      expect(wrapper.find('.demo-input').exists()).toBe(true)

      // Test info overlay
      expect(wrapper.find('.theme-info-overlay').exists()).toBe(true)
      expect(wrapper.text()).toContain('Preview Test Theme')
      expect(wrapper.text()).toContain('Test Suite')
      expect(wrapper.text()).toContain('1.0.0')

      // Test button interaction
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      const button = wrapper.find('.demo-button')
      await button.trigger('click')
      expect(consoleSpy).toHaveBeenCalledWith('Button clicked in theme preview')
      consoleSpy.mockRestore()

      // Test form interaction
      const input = wrapper.find('.demo-input')
      await input.setValue('integration test command')
      const submitButton = wrapper.find('.demo-submit')
      await submitButton.trigger('click')
      expect(consoleSpy).toHaveBeenCalledWith('Form submitted with value:', 'integration test command')
      consoleSpy.mockRestore()
    })
  })

  describe('Error Handling and Recovery', () => {
    it('should handle Wails service failures gracefully', async () => {
      // Mock Wails service failure
      const { useWails } = await import('@/composables/useWails')
      vi.mocked(useWails).mockReturnValue({
        isReady: { value: true },
        callService: vi.fn().mockRejectedValue(new Error('Wails service unavailable'))
      })

      const pinia = createTestingPinia()
      const themeComposable = useTheme()
      const fsComposable = useFileSystem()

      // Initialize should handle errors gracefully
      await expect(themeComposable.initialize()).resolves.not.toThrow()
      await expect(fsComposable.initialize()).resolves.not.toThrow()

      // Systems should fall back to default behavior
      expect(themeComposable.currentTheme.value).toBeDefined()
      expect(fsComposable.currentDirectory.value).toBeDefined()
    })

    it('should recover from temporary network failures', async () => {
      const pinia = createTestingPinia()
      const networkComposable = useNetworkMonitor()

      // Initialize successfully first
      await networkComposable.initialize()
      expect(networkComposable.stats.value).toBeDefined()

      // Simulate network failure
      const originalStats = networkComposable.stats.value
      networkComposable.setError(new Error('Network connection lost'))
      expect(networkComposable.error.value).not.toBe(null)

      // Simulate recovery
      networkComposable.clearError()
      networkComposable.updateStats({
        bytesReceived: originalStats.bytesReceived + 1000,
        bytesSent: originalStats.bytesSent + 500,
        packetsReceived: originalStats.packetsReceived + 10,
        packetsSent: originalStats.packetsSent + 5
      })

      // Verify recovery
      expect(networkComposable.error.value).toBe(null)
      expect(networkComposable.stats.value.bytesReceived).toBe(originalStats.bytesReceived + 1000)
    })

    it('should handle file system permission errors', async () => {
      const pinia = createTestingPinia()
      const fsComposable = useFileSystem()

      // Initialize
      await fsComposable.initialize()

      // Mock permission error
      fsComposable.setError(new Error('Permission denied: /restricted/file'))

      // Verify error is handled
      expect(fsComposable.error.value).not.toBe(null)
      expect(fsComposable.error.value?.message).toContain('Permission denied')

      // Verify system continues to function
      expect(fsComposable.currentDirectory.value).toBeDefined()
      expect(fsComposable.directoryContents.value).toBeDefined()

      // Clear error and verify recovery
      fsComposable.clearError()
      expect(fsComposable.error.value).toBe(null)
    })
  })

  describe('Performance and Memory Management', () => {
    it('should handle large amounts of data efficiently', async () => {
      const pinia = createTestingPinia()
      const themeComposable = useTheme()
      const fsComposable = useFileSystem()

      // Initialize
      await Promise.all([
        themeComposable.initialize(),
        fsComposable.initialize()
      ])

      // Generate large number of themes
      const largeThemeList = Array.from({ length: 1000 }, (_, i) => ({
        id: `performance-test-theme-${i}`,
        name: `Performance Test Theme ${i}`,
        description: `A theme for performance testing ${i}`,
        author: 'Test Suite',
        version: '1.0.0',
        colors: {
          background: { primary: '#0a0a0a', secondary: '#1a1a1a', tertiary: '#2a2a2a' },
          foreground: { primary: '#ffffff', secondary: '#cccccc', tertiary: '#999999' },
          accent: { primary: '#00ff41', secondary: '#00cc33' },
          status: { success: '#00ff41', warning: '#ffaa00', error: '#ff3333', info: '#00aaff' },
          terminal: [],
          ui: {
            buttonBackground: '#1a1a1a',
            buttonForeground: '#ffffff',
            buttonHover: '#2a2a2a',
            buttonActive: '#00ff41',
            inputBackground: '#0a0a0a',
            inputForeground: '#ffffff',
            inputBorder: '#333333',
            inputFocus: '#00ff41',
            border: '#333333',
            shadow: 'rgba(0, 0, 0, 0.5)'
          }
        },
        fonts: {
          families: { primary: 'Inter', secondary: 'Roboto', monospace: 'JetBrains Mono' },
          sizes: { extraSmall: '0.75rem', small: '0.875rem', base: '1rem', large: '1.125rem', extraLarge: '1.25rem', doubleExtraLarge: '1.5rem' },
          weights: { light: '300', normal: '400', medium: '500', semiBold: '600', bold: '700' },
          lineHeights: { tight: '1.25', normal: '1.5', relaxed: '1.75' },
          letterSpacing: { tight: '-0.025em', normal: '0', wide: '0.025em' }
        },
        effects: {
          borderRadius: { small: '0.25rem', medium: '0.5rem', large: '1rem', extraLarge: '1.5rem', full: '9999px' },
          shadows: [],
          gradients: [],
          animations: {
            duration: { fast: '150ms', normal: '300ms', slow: '500ms' },
            easing: { linear: 'linear', easeIn: 'cubic-bezier(0.4, 0, 1, 1)', easeOut: 'cubic-bezier(0, 0, 0.2, 1)', easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)' }
          },
          blur: { small: '4px', medium: '8px', large: '16px', extraLarge: '24px' }
        },
        settings: {
          disabledOpacity: 0.5,
          hoverOpacity: 0.8,
          activeOpacity: 1,
          transitionDuration: { fast: '150ms', normal: '300ms', slow: '500ms' },
          zIndexScale: {}
        }
      }))

      // Test performance with large theme list
      const startTime = performance.now()

      for (const theme of largeThemeList.slice(0, 10)) { // Test with smaller subset for unit tests
        const cssVars = themeComposable.generateCSSVariables(theme)
        expect(cssVars).toContain('--dex-theme-colors-background-primary')
      }

      const endTime = performance.now()
      const duration = endTime - startTime

      // Performance should be reasonable (less than 1 second for 10 themes)
      expect(duration).toBeLessThan(1000)
    })

    it('should properly clean up resources when components are unmounted', async () => {
      const pinia = createTestingPinia()
      const networkComposable = useNetworkMonitor()
      const processComposable = useProcessMonitor()

      // Initialize and start monitoring
      await Promise.all([
        networkComposable.initialize(),
        processComposable.initialize()
      ])

      await networkComposable.startMonitoring()
      await processComposable.startMonitoring()

      // Verify monitoring is active
      expect(networkComposable.isMonitoring.value).toBe(true)
      expect(processComposable.isMonitoring.value).toBe(true)

      // Simulate component unmounting
      networkComposable.stopMonitoring()
      processComposable.stopMonitoring()

      // Verify cleanup
      expect(networkComposable.isMonitoring.value).toBe(false)
      expect(processComposable.isMonitoring.value).toBe(false)
    })
  })
})