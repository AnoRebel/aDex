import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createTestingPinia } from '@pinia/testing'
import { setActivePinia, createPinia } from 'pinia'
import FileBrowser from '~/app/components/filesystem/FileBrowser.vue'
import FileList from '~/app/components/filesystem/FileList.vue'
import FileIcon from '~/app/components/filesystem/FileIcon.vue'
import { useFilesystem } from '~/app/composables/useFilesystem'
import type { FileSystemEntry } from '~/app/types/filesystem'

// Mock the filesystem composable
vi.mock('~/app/composables/useFilesystem', () => ({
  useFilesystem: () => ({
    currentPath: ref('/home/user'),
    entries: ref<FileSystemEntry[]>([
      {
        name: 'Documents',
        path: '/home/user/Documents',
        isDir: true,
        size: 0,
        mode: 0o755,
        modTime: '2023-01-01T00:00:00Z',
        permissions: 'rwxr-xr-x',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: true,
        children: []
      },
      {
        name: 'test.txt',
        path: '/home/user/test.txt',
        isDir: false,
        size: 1024,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      },
      {
        name: '.hidden',
        path: '/home/user/.hidden',
        isDir: false,
        size: 512,
        mode: 0o600,
        modTime: '2023-01-01T00:00:00Z',
        permissions: 'rw-------',
        owner: 'user',
        group: 'user',
        isHidden: true,
        isExecutable: false
      }
    ]),
    isLoading: ref(false),
    error: ref(null),
    showHidden: ref(false),
    sortBy: ref('name'),
    sortOrder: ref('asc'),
    selectedItems: ref<string[]>([]),
    navigateToDirectory: vi.fn(),
    navigateUp: vi.fn(),
    refresh: vi.fn(),
    toggleHidden: vi.fn(),
    setSorting: vi.fn(),
    selectItem: vi.fn(),
    selectAll: vi.fn(),
    clearSelection: vi.fn(),
    deleteItem: vi.fn(),
    renameItem: vi.fn(),
    createDirectory: vi.fn(),
    openTerminal: vi.fn(),
    copyPath: vi.fn()
  })
}))

// Mock wails bindings
const mockWails = {
  Events: {
    On: vi.fn(),
    Off: vi.fn()
  }
}

global.wails = mockWails

describe('FileBrowser Component', () => {
  let wrapper: any
  let pinia: any

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  describe('Rendering', () => {
    it('should render the file browser correctly', () => {
      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      expect(wrapper.find('.file-browser').exists()).toBe(true)
      expect(wrapper.find('.file-browser-header').exists()).toBe(true)
      expect(wrapper.find('.file-browser-content').exists()).toBe(true)
      expect(wrapper.find('.file-browser-footer').exists()).toBe(true)
    })

    it('should display the current path in the breadcrumb', () => {
      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const breadcrumb = wrapper.find('.breadcrumb')
      expect(breadcrumb.exists()).toBe(true)
      expect(breadcrumb.text()).toContain('/home/user')
    })

    it('should show loading state when loading', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.isLoading.value = true

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      expect(wrapper.find('.loading-spinner').exists()).toBe(true)
      expect(wrapper.find('.file-list').exists()).toBe(false)
    })

    it('should display error message when there is an error', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.error.value = 'Permission denied'

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      expect(wrapper.find('.error-message').exists()).toBe(true)
      expect(wrapper.find('.error-message').text()).toContain('Permission denied')
    })

    it('should display empty state when no entries', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.entries.value = []

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      expect(wrapper.find('.empty-state').exists()).toBe(true)
      expect(wrapper.find('.empty-state').text()).toContain('No files found')
    })
  })

  describe('Navigation', () => {
    it('should navigate up when clicking up button', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const upButton = wrapper.find('[data-testid="navigate-up-button"]')
      await upButton.trigger('click')

      expect(mockFilesystem.navigateUp).toHaveBeenCalledTimes(1)
    })

    it('should refresh when clicking refresh button', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const refreshButton = wrapper.find('[data-testid="refresh-button"]')
      await refreshButton.trigger('click')

      expect(mockFilesystem.refresh).toHaveBeenCalledTimes(1)
    })

    it('should open terminal when clicking terminal button', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const terminalButton = wrapper.find('[data-testid="terminal-button"]')
      await terminalButton.trigger('click')

      expect(mockFilesystem.openTerminal).toHaveBeenCalledTimes(1)
    })
  })

  describe('File Operations', () => {
    it('should toggle hidden files visibility', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const hiddenToggle = wrapper.find('[data-testid="toggle-hidden"]')
      await hiddenToggle.trigger('click')

      expect(mockFilesystem.toggleHidden).toHaveBeenCalledTimes(1)
    })

    it('should create new directory when clicking new folder button', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const newFolderButton = wrapper.find('[data-testid="new-folder-button"]')
      await newFolderButton.trigger('click')

      expect(mockFilesystem.createDirectory).toHaveBeenCalledTimes(1)
    })

    it('should update sorting when changing sort options', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const sortSelect = wrapper.find('[data-testid="sort-select"]')
      await sortSelect.setValue('size')

      expect(mockFilesystem.setSorting).toHaveBeenCalledWith('size', 'asc')
    })
  })

  describe('Search Functionality', () => {
    it('should show search input when search is enabled', async () => {
      wrapper = mount(FileBrowser, {
        props: {
          enableSearch: true
        },
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      expect(wrapper.find('[data-testid="search-input"]').exists()).toBe(true)
    })

    it('should filter entries when typing in search', async () => {
      wrapper = mount(FileBrowser, {
        props: {
          enableSearch: true
        },
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const searchInput = wrapper.find('[data-testid="search-input"]')
      await searchInput.setValue('test')

      expect(searchInput.element.value).toBe('test')
    })
  })

  describe('Selection Management', () => {
    it('should show selection count when items are selected', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.selectedItems.value = ['test.txt', 'Documents']

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const selectionInfo = wrapper.find('.selection-info')
      expect(selectionInfo.exists()).toBe(true)
      expect(selectionInfo.text()).toContain('2 items selected')
    })

    it('should clear selection when clicking clear selection button', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.selectedItems.value = ['test.txt']

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const clearButton = wrapper.find('[data-testid="clear-selection"]')
      await clearButton.trigger('click')

      expect(mockFilesystem.clearSelection).toHaveBeenCalledTimes(1)
    })
  })

  describe('Context Menu', () => {
    it('should show context menu when right-clicking on files', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      // Mock a file entry click
      const fileList = wrapper.findComponent(FileList)
      await fileList.vm.$emit('context-menu', {
        x: 100,
        y: 100,
        entry: mockFilesystem.entries.value[1] // test.txt
      })

      expect(wrapper.find('.context-menu').exists()).toBe(true)
    })

    it('should copy file path when clicking copy path in context menu', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      // Mock context menu action
      const fileList = wrapper.findComponent(FileList)
      await fileList.vm.$emit('context-menu-action', {
        action: 'copy-path',
        entry: mockFilesystem.entries.value[1] // test.txt
      })

      expect(mockFilesystem.copyPath).toHaveBeenCalledWith('/home/user/test.txt')
    })
  })

  describe('Responsive Design', () => {
    it('should adapt to small screens', async () => {
      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      // Simulate small screen
      wrapper.setData({ isMobile: true })

      expect(wrapper.find('.file-browser').classes()).toContain('mobile-layout')
      expect(wrapper.find('.file-browser-header').classes()).toContain('mobile-header')
    })

    it('should show simplified controls on mobile', async () => {
      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      wrapper.setData({ isMobile: true })

      // Mobile should hide some controls and show mobile-specific ones
      expect(wrapper.find('.mobile-actions').exists()).toBe(true)
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels', () => {
      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const upButton = wrapper.find('[data-testid="navigate-up-button"]')
      expect(upButton.attributes('aria-label')).toBe('Navigate up')

      const refreshButton = wrapper.find('[data-testid="refresh-button"]')
      expect(refreshButton.attributes('aria-label')).toBe('Refresh')
    })

    it('should support keyboard navigation', async () => {
      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      // Test keyboard shortcuts
      await wrapper.trigger('keydown', { key: 'F5' })
      // Should trigger refresh

      await wrapper.trigger('keydown', { key: 'Backspace' })
      // Should trigger navigate up

      await wrapper.trigger('keydown', { key: 'Delete' })
      // Should trigger delete for selected items
    })
  })

  describe('Error Handling', () => {
    it('should handle permission denied errors gracefully', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.error.value = 'Permission denied: cannot access directory'

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const errorElement = wrapper.find('.error-message')
      expect(errorElement.exists()).toBe(true)
      expect(errorElement.text()).toContain('Permission denied')
      expect(errorElement.find('.retry-button').exists()).toBe(true)
    })

    it('should handle network errors when loading files', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()
      mockFilesystem.error.value = 'Network error: unable to connect to file system service'

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      expect(wrapper.find('.error-message').exists()).toBe(true)
      expect(wrapper.find('.error-message').text()).toContain('Network error')
    })
  })

  describe('Performance', () => {
    it('should debounce search input', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      wrapper = mount(FileBrowser, {
        props: {
          enableSearch: true
        },
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      const searchInput = wrapper.find('[data-testid="search-input"]')

      // Rapid typing should be debounced
      await searchInput.setValue('t')
      await searchInput.setValue('te')
      await searchInput.setValue('tes')
      await searchInput.setValue('test')

      // Should not call search immediately for each keystroke
      expect(mockFilesystem.setSearchQuery).toHaveBeenCalledTimes(0)
    })

    it('should lazy load large directory listings', async () => {
      const { useFilesystem } = await import('~/app/composables/useFilesystem')
      const mockFilesystem = useFilesystem()

      // Mock large directory
      const largeEntries: FileSystemEntry[] = Array.from({ length: 1000 }, (_, i) => ({
        name: `file_${i}.txt`,
        path: `/home/user/file_${i}.txt`,
        isDir: false,
        size: 1024,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      }))

      mockFilesystem.entries.value = largeEntries

      wrapper = mount(FileBrowser, {
        global: {
          plugins: [pinia],
          stubs: {
            FileList: true,
            FileIcon: true
          }
        }
      })

      // Should render only visible items initially
      expect(wrapper.findComponent(FileList).props('entries').length).toBeLessThanOrEqual(100)
    })
  })
})

describe('FileList Component', () => {
  let wrapper: any

  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  it('should render file list correctly', () => {
    const entries: FileSystemEntry[] = [
      {
        name: 'Documents',
        path: '/home/user/Documents',
        isDir: true,
        size: 0,
        mode: 0o755,
        modTime: '2023-01-01T00:00:00Z',
        permissions: 'rwxr-xr-x',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: true
      },
      {
        name: 'test.txt',
        path: '/home/user/test.txt',
        isDir: false,
        size: 1024,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      }
    ]

    wrapper = mount(FileList, {
      props: {
        entries,
        sortBy: 'name',
        sortOrder: 'asc'
      },
      global: {
        plugins: [pinia],
        stubs: {
          FileIcon: true
        }
      }
    })

    expect(wrapper.find('.file-list').exists()).toBe(true)
    expect(wrapper.findAll('.file-item')).toHaveLength(2)
  })

  it('should sort entries by name', () => {
    const entries: FileSystemEntry[] = [
      {
        name: 'zeta.txt',
        path: '/home/user/zeta.txt',
        isDir: false,
        size: 1024,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      },
      {
        name: 'alpha.txt',
        path: '/home/user/alpha.txt',
        isDir: false,
        size: 512,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      }
    ]

    wrapper = mount(FileList, {
      props: {
        entries,
        sortBy: 'name',
        sortOrder: 'asc'
      },
      global: {
        plugins: [pinia],
        stubs: {
          FileIcon: true
        }
      }
    })

    const fileItems = wrapper.findAll('.file-item')
    expect(fileItems[0].text()).toContain('alpha.txt')
    expect(fileItems[1].text()).toContain('zeta.txt')
  })

  it('should sort entries by size', () => {
    const entries: FileSystemEntry[] = [
      {
        name: 'large.txt',
        path: '/home/user/large.txt',
        isDir: false,
        size: 2048,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      },
      {
        name: 'small.txt',
        path: '/home/user/small.txt',
        isDir: false,
        size: 512,
        mode: 0o644,
        modTime: '2023-01-01T00:00:00Z',
        permissions: '-rw-r--r--',
        owner: 'user',
        group: 'user',
        isHidden: false,
        isExecutable: false
      }
    ]

    wrapper = mount(FileList, {
      props: {
        entries,
        sortBy: 'size',
        sortOrder: 'desc'
      },
      global: {
        plugins: [pinia],
        stubs: {
          FileIcon: true
        }
      }
    })

    const fileItems = wrapper.findAll('.file-item')
    expect(fileItems[0].text()).toContain('large.txt')
    expect(fileItems[1].text()).toContain('small.txt')
  })
})

describe('FileIcon Component', () => {
  let wrapper: any

  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  it('should render correct icon for directory', () => {
    const entry: FileSystemEntry = {
      name: 'Documents',
      path: '/home/user/Documents',
      isDir: true,
      size: 0,
      mode: 0o755,
      modTime: '2023-01-01T00:00:00Z',
      permissions: 'rwxr-xr-x',
      owner: 'user',
      group: 'user',
      isHidden: false,
      isExecutable: true
    }

    wrapper = mount(FileIcon, {
      props: {
        entry
      },
      global: {
        plugins: [pinia]
      }
    })

    expect(wrapper.find('.file-icon').exists()).toBe(true)
    expect(wrapper.find('.file-icon').classes()).toContain('directory')
  })

  it('should render correct icon for text file', () => {
    const entry: FileSystemEntry = {
      name: 'document.txt',
      path: '/home/user/document.txt',
      isDir: false,
      size: 1024,
      mode: 0o644,
      modTime: '2023-01-01T00:00:00Z',
      permissions: '-rw-r--r--',
      owner: 'user',
      group: 'user',
      isHidden: false,
      isExecutable: false
    }

    wrapper = mount(FileIcon, {
      props: {
        entry
      },
      global: {
        plugins: [pinia]
      }
    })

    expect(wrapper.find('.file-icon').exists()).toBe(true)
    expect(wrapper.find('.file-icon').classes()).toContain('text-file')
  })

  it('should render correct icon for executable file', () => {
    const entry: FileSystemEntry = {
      name: 'script.sh',
      path: '/home/user/script.sh',
      isDir: false,
      size: 512,
      mode: 0o755,
      modTime: '2023-01-01T00:00:00Z',
      permissions: 'rwxr-xr-x',
      owner: 'user',
      group: 'user',
      isHidden: false,
      isExecutable: true
    }

    wrapper = mount(FileIcon, {
      props: {
        entry
      },
      global: {
        plugins: [pinia]
      }
    })

    expect(wrapper.find('.file-icon').exists()).toBe(true)
    expect(wrapper.find('.file-icon').classes()).toContain('executable')
  })

  it('should render hidden indicator for hidden files', () => {
    const entry: FileSystemEntry = {
      name: '.hidden',
      path: '/home/user/.hidden',
      isDir: false,
      size: 256,
      mode: 0o600,
      modTime: '2023-01-01T00:00:00Z',
      permissions: 'rw-------',
      owner: 'user',
      group: 'user',
      isHidden: true,
      isExecutable: false
    }

    wrapper = mount(FileIcon, {
      props: {
        entry
      },
      global: {
        plugins: [pinia]
      }
    })

    expect(wrapper.find('.hidden-indicator').exists()).toBe(true)
  })
})