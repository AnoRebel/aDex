import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { mount } from '@vue/test-utils'
import { createTestingPinia } from '@pinia/testing'

// Mock Wails runtime
vi.mock('@/composables/useWails', () => ({
  useWails: () => ({
    isReady: { value: false },
    callService: vi.fn()
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
import ThemeSelector from '~/components/theme/ThemeSelector.vue'
import ThemePreview from '~/components/theme/ThemePreview.vue'
import type { Theme, ThemeSelectorConfig } from '~/types/theme'

describe('Theme System', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('useTheme Composable', () => {
    let themeComposable: ReturnType<typeof useTheme>

    beforeEach(() => {
      themeComposable = useTheme()
    })

    it('should initialize with default values', () => {
      expect(themeComposable.currentTheme.value).toBeDefined()
      expect(themeComposable.availableThemes.value).toBeInstanceOf(Array)
      expect(themeComposable.isLoading.value).toBe(false)
      expect(themeComposable.error.value).toBe(null)
    })

    it('should provide theme switching functionality', async () => {
      const testTheme: Theme = {
        id: 'test-theme',
        name: 'Test Theme',
        description: 'A test theme',
        author: 'Test Suite',
        version: '1.0.0',
        colors: {
          background: {
            primary: '#0a0a0a',
            secondary: '#1a1a1a',
            tertiary: '#2a2a2a'
          },
          foreground: {
            primary: '#ffffff',
            secondary: '#cccccc',
            tertiary: '#999999'
          },
          accent: {
            primary: '#00ff41',
            secondary: '#00cc33'
          },
          status: {
            success: '#00ff41',
            warning: '#ffaa00',
            error: '#ff3333',
            info: '#00aaff'
          },
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
          families: {
            primary: 'Inter, sans-serif',
            secondary: 'Roboto, sans-serif',
            monospace: 'JetBrains Mono, monospace'
          },
          sizes: {
            extraSmall: '0.75rem',
            small: '0.875rem',
            base: '1rem',
            large: '1.125rem',
            extraLarge: '1.25rem',
            doubleExtraLarge: '1.5rem'
          },
          weights: {
            light: '300',
            normal: '400',
            medium: '500',
            semiBold: '600',
            bold: '700'
          },
          lineHeights: {
            tight: '1.25',
            normal: '1.5',
            relaxed: '1.75'
          },
          letterSpacing: {
            tight: '-0.025em',
            normal: '0',
            wide: '0.025em'
          }
        },
        effects: {
          borderRadius: {
            small: '0.25rem',
            medium: '0.5rem',
            large: '1rem',
            extraLarge: '1.5rem',
            full: '9999px'
          },
          shadows: [],
          gradients: [],
          animations: {
            duration: {
              fast: '150ms',
              normal: '300ms',
              slow: '500ms'
            },
            easing: {
              linear: 'linear',
              easeIn: 'cubic-bezier(0.4, 0, 1, 1)',
              easeOut: 'cubic-bezier(0, 0, 0.2, 1)',
              easeInOut: 'cubic-bezier(0.4, 0, 0.2, 1)'
            }
          },
          blur: {
            small: '4px',
            medium: '8px',
            large: '16px',
            extraLarge: '24px'
          }
        },
        settings: {
          disabledOpacity: 0.5,
          hoverOpacity: 0.8,
          activeOpacity: 1,
          transitionDuration: {
            fast: '150ms',
            normal: '300ms',
            slow: '500ms'
          },
          zIndexScale: {}
        },
        isDark: true,
        isCustom: false
      }

      await themeComposable.setTheme(testTheme)
      expect(themeComposable.currentTheme.value.id).toBe('test-theme')
    })

    it('should handle theme loading state correctly', async () => {
      expect(themeComposable.isLoading.value).toBe(false)

      // Simulate loading
      const loadingPromise = themeComposable.initialize()
      expect(themeComposable.isLoading.value).toBe(true)

      await loadingPromise
      expect(themeComposable.isLoading.value).toBe(false)
    })

    it('should handle theme errors gracefully', async () => {
      // Mock a theme loading error
      const mockError = new Error('Failed to load theme')
      vi.spyOn(themeComposable, 'loadTheme').mockRejectedValue(mockError)

      await expect(themeComposable.loadTheme('invalid-theme')).rejects.toThrow('Failed to load theme')
      expect(themeComposable.error.value).not.toBe(null)
    })

    it('should provide CSS variable generation', () => {
      const testTheme: Theme = {
        id: 'test-theme',
        name: 'Test Theme',
        description: 'Test',
        author: 'Test',
        version: '1.0.0',
        colors: {
          background: { primary: '#ffffff', secondary: '#f0f0f0', tertiary: '#e0e0e0' },
          foreground: { primary: '#000000', secondary: '#333333', tertiary: '#666666' },
          accent: { primary: '#0066cc', secondary: '#004499' },
          status: { success: '#00cc00', warning: '#cc9900', error: '#cc0000', info: '#0066cc' },
          terminal: [],
          ui: {
            buttonBackground: '#f0f0f0',
            buttonForeground: '#000000',
            buttonHover: '#e0e0e0',
            buttonActive: '#d0d0d0',
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
        }
      }

      const cssVars = themeComposable.generateCSSVariables(testTheme)
      expect(cssVars).toContain('--dex-theme-colors-background-primary: #ffffff')
      expect(cssVars).toContain('--dex-theme-fonts-families-primary: Arial')
      expect(cssVars).toContain('--dex-theme-effects-border-radius-small: 2px')
    })

    it('should detect theme mode correctly', () => {
      const darkTheme: Theme = {
        id: 'dark-theme',
        name: 'Dark Theme',
        description: 'Dark theme',
        author: 'Test',
        version: '1.0.0',
        colors: {
          background: { primary: '#000000', secondary: '#1a1a1a', tertiary: '#2a2a2a' },
          foreground: { primary: '#ffffff', secondary: '#cccccc', tertiary: '#999999' },
          accent: { primary: '#00ff41', secondary: '#00cc33' },
          status: { success: '#00ff41', warning: '#ffaa00', error: '#ff3333', info: '#00aaff' },
          terminal: [],
          ui: {
            buttonBackground: '#1a1a1a',
            buttonForeground: '#ffffff',
            buttonHover: '#2a2a2a',
            buttonActive: '#00ff41',
            inputBackground: '#000000',
            inputForeground: '#ffffff',
            inputBorder: '#333333',
            inputFocus: '#00ff41',
            border: '#333333',
            shadow: 'rgba(0, 0, 0, 0.5)'
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
        isDark: true
      }

      const lightTheme: Theme = {
        id: 'light-theme',
        name: 'Light Theme',
        description: 'Light theme',
        author: 'Test',
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

      expect(themeComposable.isDarkTheme(darkTheme)).toBe(true)
      expect(themeComposable.isDarkTheme(lightTheme)).toBe(false)
    })
  })

  describe('Theme Store', () => {
    let store: ReturnType<typeof useThemeStore>

    beforeEach(() => {
      store = useThemeStore()
    })

    it('should initialize with default state', () => {
      expect(store.currentTheme).toBeDefined()
      expect(store.availableThemes).toBeInstanceOf(Array)
      expect(store.isLoading).toBe(false)
      expect(store.error).toBe(null)
    })

    it('should sync with composable', async () => {
      const testTheme: Theme = {
        id: 'store-test-theme',
        name: 'Store Test Theme',
        description: 'A theme for testing store functionality',
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
      }

      await store.setCurrentTheme(testTheme)
      expect(store.currentTheme.id).toBe('store-test-theme')
    })

    it('should handle loading state', async () => {
      store.setLoading(true)
      expect(store.isLoading).toBe(true)

      store.setLoading(false)
      expect(store.isLoading).toBe(false)
    })

    it('should handle errors', () => {
      const error = new Error('Test error')
      store.setError(error)
      expect(store.error).toBe(error)

      store.clearError()
      expect(store.error).toBe(null)
    })

    it('should update available themes', () => {
      const themes: Theme[] = [
        {
          id: 'theme1',
          name: 'Theme 1',
          description: 'First theme',
          author: 'Test',
          version: '1.0.0',
          colors: {
            background: { primary: '#ffffff', secondary: '#f0f0f0', tertiary: '#e0e0e0' },
            foreground: { primary: '#000000', secondary: '#333333', tertiary: '#666666' },
            accent: { primary: '#0066cc', secondary: '#004499' },
            status: { success: '#00cc00', warning: '#cc9900', error: '#cc0000', info: '#0066cc' },
            terminal: [],
            ui: {
              buttonBackground: '#f0f0f0',
              buttonForeground: '#000000',
              buttonHover: '#e0e0e0',
              buttonActive: '#d0d0d0',
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
          }
        }
      ]

      store.setAvailableThemes(themes)
      expect(store.availableThemes).toHaveLength(1)
      expect(store.availableThemes[0].id).toBe('theme1')
    })
  })

  describe('ThemeSelector Component', () => {
    const mockTheme: Theme = {
      id: 'selector-test-theme',
      name: 'Selector Test Theme',
      description: 'A theme for testing ThemeSelector component',
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
      },
      metadata: {
        tags: ['dark', 'cyberpunk'],
        category: 'dark'
      }
    }

    const config: ThemeSelectorConfig = {
      showCreateButton: true,
      showImportButton: true,
      showExportButton: true,
      compact: false,
      searchable: true,
      categorizable: true,
      previewEnabled: true
    }

    it('should render theme selector correctly', () => {
      const wrapper = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: mockTheme,
          config
        },
        global: {
          plugins: [createTestingPinia()],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      expect(wrapper.find('.theme-selector').exists()).toBe(true)
      expect(wrapper.find('.category-tabs').exists()).toBe(true)
      expect(wrapper.find('.theme-grid').exists()).toBe(true)
    })

    it('should handle theme selection', async () => {
      const wrapper = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: null,
          config
        },
        global: {
          plugins: [createTestingPinia()],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      // Find and click on a theme
      const themeCard = wrapper.find('.theme-card')
      expect(themeCard.exists()).toBe(true)

      await themeCard.trigger('click')

      // Check if the event was emitted
      expect(wrapper.emitted('theme-select')).toBeTruthy()
      expect(wrapper.emitted('theme-select')?.[0]).toEqual([mockTheme])
    })

    it('should handle search functionality', async () => {
      const wrapper = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: mockTheme,
          config: { ...config, searchable: true }
        },
        global: {
          plugins: [createTestingPinia()],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      const searchInput = wrapper.find('.search-input')
      expect(searchInput.exists()).toBe(true)

      await searchInput.setValue('test')
      expect(wrapper.emitted('search')).toBeTruthy()
      expect(wrapper.emitted('search')?.[0]).toEqual(['test'])
    })

    it('should handle category filtering', async () => {
      const wrapper = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: mockTheme,
          config: { ...config, categorizable: true }
        },
        global: {
          plugins: [createTestingPinia()],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      const categoryTab = wrapper.find('.category-tab')
      expect(categoryTab.exists()).toBe(true)

      await categoryTab.trigger('click')
      expect(wrapper.emitted('category-change')).toBeTruthy()
    })

    it('should show/hide create button based on config', () => {
      const wrapperWithCreate = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: mockTheme,
          config: { ...config, showCreateButton: true }
        },
        global: {
          plugins: [createTestingPinia()],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      expect(wrapperWithCreate.find('.create-button').exists()).toBe(true)

      const wrapperWithoutCreate = mount(ThemeSelector, {
        props: {
          availableThemes: [mockTheme],
          currentTheme: mockTheme,
          config: { ...config, showCreateButton: false }
        },
        global: {
          plugins: [createTestingPinia()],
          stubs: {
            Icon: { template: '<div></div>' },
            ThemePreview: { template: '<div class="theme-preview-stub"></div>' }
          }
        }
      })

      expect(wrapperWithoutCreate.find('.create-button').exists()).toBe(false)
    })
  })

  describe('ThemePreview Component', () => {
    const mockTheme: Theme = {
      id: 'preview-test-theme',
      name: 'Preview Test Theme',
      description: 'A theme for testing ThemePreview component',
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
      },
      metadata: {
        tags: ['dark', 'cyberpunk'],
        category: 'dark'
      }
    }

    it('should render theme preview correctly', () => {
      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: false,
          showInfo: false,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      expect(wrapper.find('.theme-preview').exists()).toBe(true)
      expect(wrapper.find('.preview-header').exists()).toBe(true)
      expect(wrapper.find('.preview-content').exists()).toBe(true)
      expect(wrapper.find('.terminal-section').exists()).toBe(true)
      expect(wrapper.find('.status-bar').exists()).toBe(true)
    })

    it('should apply theme styles correctly', () => {
      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: false,
          showInfo: false,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      const header = wrapper.find('.preview-header')
      expect(header.attributes('style')).toContain('background-color: rgb(26, 26, 26)')
      expect(header.attributes('style')).toContain('color: rgb(255, 255, 255)')

      const content = wrapper.find('.preview-content')
      expect(content.attributes('style')).toContain('background-color: rgb(10, 10, 10)')
    })

    it('should handle compact mode', () => {
      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: true,
          showLabels: true,
          showInteractive: false,
          showInfo: false,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      expect(wrapper.find('.theme-preview.compact').exists()).toBe(true)
      expect(wrapper.find('.sidebar-section').exists()).toBe(false) // Sidebar should not show in compact mode
    })

    it('should show interactive elements when enabled', () => {
      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: true,
          showInfo: false,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      expect(wrapper.find('.interactive-elements').exists()).toBe(true)
      expect(wrapper.find('.demo-button').exists()).toBe(true)
      expect(wrapper.find('.demo-input').exists()).toBe(true)
    })

    it('should handle button clicks', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: true,
          showInfo: false,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      const button = wrapper.find('.demo-button')
      await button.trigger('click')

      expect(consoleSpy).toHaveBeenCalledWith('Button clicked in theme preview')
      consoleSpy.mockRestore()
    })

    it('should handle form submission', async () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: true,
          showInfo: false,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      const input = wrapper.find('.demo-input')
      await input.setValue('test command')

      const submitButton = wrapper.find('.demo-submit')
      await submitButton.trigger('click')

      expect(consoleSpy).toHaveBeenCalledWith('Form submitted with value:', 'test command')
      consoleSpy.mockRestore()
    })

    it('should show info overlay when enabled', () => {
      const wrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: false,
          showInfo: true,
          size: 'medium'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      expect(wrapper.find('.theme-info-overlay').exists()).toBe(true)
      expect(wrapper.find('.theme-info-content').exists()).toBe(true)
      expect(wrapper.text()).toContain('Preview Test Theme')
      expect(wrapper.text()).toContain('Test Suite')
      expect(wrapper.text()).toContain('1.0.0')
    })

    it('should apply size variations', () => {
      const largeWrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: false,
          showInfo: false,
          size: 'large'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      expect(largeWrapper.find('.theme-preview.large').exists()).toBe(true)

      const smallWrapper = mount(ThemePreview, {
        props: {
          theme: mockTheme,
          compact: false,
          showLabels: true,
          showInteractive: false,
          showInfo: false,
          size: 'small'
        },
        global: {
          stubs: {
            Icon: { template: '<div></div>' }
          }
        }
      })

      expect(smallWrapper.find('.theme-preview.small').exists()).toBe(true)
    })
  })

  describe('Theme Integration', () => {
    it('should integrate composable and store correctly', async () => {
      const store = useThemeStore()
      const composable = useTheme()

      // Initialize both
      await composable.initialize()
      await store.initialize()

      // Set theme through composable
      const testTheme: Theme = {
        id: 'integration-test-theme',
        name: 'Integration Test Theme',
        description: 'Testing integration between composable and store',
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
      }

      await composable.setTheme(testTheme)

      // Store should be updated
      expect(store.currentTheme.id).toBe('integration-test-theme')
    })

    it('should handle error states across the system', async () => {
      const store = useThemeStore()
      const composable = useTheme()

      // Mock an error
      const error = new Error('Integration test error')
      vi.spyOn(composable, 'loadTheme').mockRejectedValue(error)

      await expect(composable.loadTheme('invalid-theme')).rejects.toThrow('Integration test error')

      // Error should be propagated to store
      expect(store.error).not.toBe(null)
    })
  })
})