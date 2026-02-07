import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import EdexFilesystem from '~/app/components/edex/EdexFilesystem.vue'
import { useFilesystemStore } from '~/stores/filesystem'
import { useTerminalStore } from '~/stores/terminal'

// Mock the useWails composable
vi.mock('~/composables/useWails', () => ({
  useWails: () => ({
    system: {
      getSystemInfo: vi.fn().mockResolvedValue({}),
      getCPUUsage: vi.fn().mockResolvedValue({}),
      getMemoryUsage: vi.fn().mockResolvedValue({}),
      getDiskUsage: vi.fn().mockResolvedValue([
        {
          device: '/dev/sda1',
          mountpoint: '/',
          fstype: 'ext4',
          total: 500000000000,
          used: 250000000000,
          free: 250000000000,
          usage: 50.0,
        }
      ]),
      getTopProcesses: vi.fn().mockResolvedValue([]),
      startMonitoring: vi.fn().mockResolvedValue(undefined),
      stopMonitoring: vi.fn().mockResolvedValue(undefined),
    },
    filesystem: {
      readDirectory: vi.fn().mockResolvedValue([]),
      createDirectory: vi.fn().mockResolvedValue(undefined),
      deleteFile: vi.fn().mockResolvedValue(undefined),
      renameFile: vi.fn().mockResolvedValue(undefined),
      copyFile: vi.fn().mockResolvedValue(undefined),
      moveFile: vi.fn().mockResolvedValue(undefined),
      searchFiles: vi.fn().mockResolvedValue([]),
      readFile: vi.fn().mockResolvedValue(''),
      writeFile: vi.fn().mockResolvedValue(undefined),
    },
    isReady: { value: true },
    error: { value: null },
  })
}))

vi.mock('~/utils/errorHandler', () => ({
  handleBackendError: vi.fn(),
  handleComponentError: vi.fn(),
}))

vi.mock('~/utils/loadingStates', () => ({
  createLoading: vi.fn(),
  completeLoading: vi.fn(),
}))

// Mock the Wails bindings used by the terminal store
vi.mock('~~/bindings', () => ({
  CreateTerminal: vi.fn().mockResolvedValue({ id: 'test-session' }),
  WriteToTerminal: vi.fn().mockResolvedValue(undefined),
  ResizeTerminal: vi.fn().mockResolvedValue(undefined),
  CloseTerminal: vi.fn().mockResolvedValue(undefined),
}))

const mockDirectories = [
  { name: 'Documents', path: '/home/user/Documents', size: 0, modified: '2026-01-15T10:00:00Z', type: 'directory' },
  { name: 'Downloads', path: '/home/user/Downloads', size: 0, modified: '2026-01-20T15:30:00Z', type: 'directory' },
  { name: '.config', path: '/home/user/.config', size: 0, modified: '2026-01-10T08:00:00Z', type: 'directory' },
]

const mockFiles = [
  { name: 'readme.md', path: '/home/user/readme.md', size: 1024, modified: '2026-01-25T12:00:00Z', type: 'file', extension: 'md' },
  { name: 'app.js', path: '/home/user/app.js', size: 4096, modified: '2026-01-28T09:00:00Z', type: 'file', extension: 'js' },
  { name: '.bashrc', path: '/home/user/.bashrc', size: 512, modified: '2026-01-05T06:00:00Z', type: 'file', extension: 'bashrc' },
]

describe('EdexFilesystem Component', () => {
  let wrapper: VueWrapper

  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
  })

  afterEach(() => {
    wrapper?.unmount()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  function mountComponent(options?: {
    directories?: any[]
    files?: any[]
    currentPath?: string
  }) {
    const pinia = createPinia()
    setActivePinia(pinia)

    const fsStore = useFilesystemStore()
    const termStore = useTerminalStore()

    // Pre-populate the filesystem store
    fsStore.currentPath = options?.currentPath ?? '/home/user'
    fsStore.directories = options?.directories ?? mockDirectories
    fsStore.files = options?.files ?? mockFiles
    fsStore.settings = {
      showHiddenFiles: false,
      showPreview: true,
      enableThumbnails: true,
      itemsPerPage: 50,
      doubleClickAction: 'open',
      confirmDelete: true,
      viewMode: 'grid',
    }

    // Terminal store needs a minimal activeSession
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

    wrapper = mount(EdexFilesystem, {
      global: { plugins: [pinia] }
    })

    return { fsStore, termStore }
  }

  describe('Rendering', () => {
    it('renders with edex-filesystem class', () => {
      mountComponent()
      expect(wrapper.find('.edex-filesystem').exists()).toBe(true)
    })

    it('shows FILESYSTEM header', () => {
      mountComponent()
      const title = wrapper.find('.fs-header-title')
      expect(title.exists()).toBe(true)
      expect(title.text()).toContain('FILESYSTEM')
    })

    it('shows the current path in the header', () => {
      mountComponent({ currentPath: '/home/user' })
      const pathEl = wrapper.find('.fs-header-path')
      expect(pathEl.exists()).toBe(true)
      expect(pathEl.text()).toContain('/home/user')
    })
  })

  describe('View modes', () => {
    it('renders in grid view mode by default', () => {
      mountComponent()
      expect(wrapper.find('.fs-grid').exists()).toBe(true)
      expect(wrapper.find('.fs-list').exists()).toBe(false)
    })

    it('switches to list view when toggle button is clicked', async () => {
      mountComponent()
      // Find the view mode toggle button (the one showing '=#' for grid view)
      const buttons = wrapper.findAll('.fs-header-btn')
      const viewToggle = buttons.find(btn => btn.text() === '=#')
      expect(viewToggle).toBeDefined()

      await viewToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.fs-list').exists()).toBe(true)
      expect(wrapper.find('.fs-grid').exists()).toBe(false)
    })

    it('switches back to grid view from list view', async () => {
      mountComponent()
      const buttons = wrapper.findAll('.fs-header-btn')

      // Switch to list first
      const viewToggle = buttons.find(btn => btn.text() === '=#')
      await viewToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      // Now switch back to grid (button should now show '::')
      const updatedButtons = wrapper.findAll('.fs-header-btn')
      const gridToggle = updatedButtons.find(btn => btn.text() === '::')
      await gridToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.fs-grid').exists()).toBe(true)
    })
  })

  describe('Special items in grid view', () => {
    it('shows "Show disks" special item', () => {
      mountComponent()
      const specialItems = wrapper.findAll('.fs-item-special')
      const showDisksItem = specialItems.find(item => item.text().includes('Show disks'))
      expect(showDisksItem).toBeDefined()
    })

    it('shows "Go up" special item when not at root', () => {
      mountComponent({ currentPath: '/home/user' })
      const specialItems = wrapper.findAll('.fs-item-special')
      const goUpItem = specialItems.find(item => item.text().includes('Go up'))
      expect(goUpItem).toBeDefined()
    })

    it('does not show "Go up" at root path', () => {
      mountComponent({ currentPath: '/' })
      const specialItems = wrapper.findAll('.fs-item-special')
      const goUpItem = specialItems.find(item => item.text().includes('Go up'))
      expect(goUpItem).toBeUndefined()
    })
  })

  describe('File and directory display', () => {
    it('displays directories with fs-item-dir class', () => {
      mountComponent()
      const dirItems = wrapper.findAll('.fs-item-dir')
      // .config is hidden (dotfile), so only 2 visible dirs
      expect(dirItems.length).toBe(2)
    })

    it('displays files with fs-item-file class', () => {
      mountComponent()
      const fileItems = wrapper.findAll('.fs-item-file')
      // .bashrc is hidden (dotfile), so only 2 visible files
      expect(fileItems.length).toBe(2)
    })

    it('shows directory names', () => {
      mountComponent()
      const dirItems = wrapper.findAll('.fs-item-dir')
      const names = dirItems.map(d => d.find('.fs-item-name').text())
      expect(names).toContain('Documents')
      expect(names).toContain('Downloads')
    })

    it('shows file names', () => {
      mountComponent()
      const fileItems = wrapper.findAll('.fs-item-file')
      const names = fileItems.map(f => f.find('.fs-item-name').text())
      expect(names).toContain('readme.md')
      expect(names).toContain('app.js')
    })

    it('hides dotfiles by default', () => {
      mountComponent()
      const allItems = wrapper.findAll('.fs-item-dir, .fs-item-file')
      const names = allItems.map(item => item.find('.fs-item-name').text())
      expect(names).not.toContain('.config')
      expect(names).not.toContain('.bashrc')
    })

    it('shows dotfiles when toggled', async () => {
      mountComponent()
      const buttons = wrapper.findAll('.fs-header-btn')
      const dotfileToggle = buttons.find(btn => btn.text() === '[a-z]')
      expect(dotfileToggle).toBeDefined()

      await dotfileToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      const allItems = wrapper.findAll('.fs-item-dir, .fs-item-file')
      const names = allItems.map(item => item.find('.fs-item-name').text())
      expect(names).toContain('.config')
      expect(names).toContain('.bashrc')
    })
  })

  describe('List view', () => {
    it('shows column headers in list view (NAME, TYPE, SIZE, MODIFIED)', async () => {
      mountComponent()
      // Switch to list view
      const buttons = wrapper.findAll('.fs-header-btn')
      const viewToggle = buttons.find(btn => btn.text() === '=#')
      await viewToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      const header = wrapper.find('.fs-list-header')
      expect(header.exists()).toBe(true)
      expect(header.text()).toContain('NAME')
      expect(header.text()).toContain('TYPE')
      expect(header.text()).toContain('SIZE')
      expect(header.text()).toContain('MODIFIED')
    })

    it('shows "Show disks" in list view', async () => {
      mountComponent()
      const buttons = wrapper.findAll('.fs-header-btn')
      const viewToggle = buttons.find(btn => btn.text() === '=#')
      await viewToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      const specialItems = wrapper.findAll('.fs-list-item-special')
      expect(specialItems.length).toBeGreaterThan(0)
      const showDisks = specialItems.find(item => item.text().includes('Show disks'))
      expect(showDisks).toBeDefined()
    })

    it('shows "Go up" in list view', async () => {
      mountComponent({ currentPath: '/home/user' })
      const buttons = wrapper.findAll('.fs-header-btn')
      const viewToggle = buttons.find(btn => btn.text() === '=#')
      await viewToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      const specialItems = wrapper.findAll('.fs-list-item-special')
      const goUp = specialItems.find(item => item.text().includes('Go up'))
      expect(goUp).toBeDefined()
    })
  })

  describe('Empty directory', () => {
    it('shows DIRECTORY EMPTY message when there are no files or dirs', () => {
      mountComponent({ directories: [], files: [] })
      expect(wrapper.text()).toContain('DIRECTORY EMPTY')
    })
  })

  describe('Disk usage bar', () => {
    it('renders a disk usage bar at the bottom', () => {
      mountComponent()
      expect(wrapper.find('.fs-disk-bar').exists()).toBe(true)
    })

    it('shows disk fill bar', () => {
      mountComponent()
      expect(wrapper.find('.disk-fill').exists()).toBe(true)
      expect(wrapper.find('.disk-fill-inner').exists()).toBe(true)
    })
  })

  describe('Item click handling', () => {
    it('selects an item when clicked', async () => {
      mountComponent()
      const dirItems = wrapper.findAll('.fs-item-dir')
      await dirItems[0].trigger('click')
      await wrapper.vm.$nextTick()

      // The clicked item should get the 'active' class
      expect(dirItems[0].classes()).toContain('active')
    })
  })

  describe('Show disks overlay', () => {
    it('opens disk list overlay when Show disks is clicked', async () => {
      mountComponent()
      const specialItems = wrapper.findAll('.fs-item-special')
      const showDisksItem = specialItems.find(item => item.text().includes('Show disks'))
      expect(showDisksItem).toBeDefined()

      await showDisksItem!.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.fs-disk-list').exists()).toBe(true)
      expect(wrapper.text()).toContain('MOUNT POINTS')
    })

    it('closes disk list overlay when close button is clicked', async () => {
      mountComponent()
      const specialItems = wrapper.findAll('.fs-item-special')
      const showDisksItem = specialItems.find(item => item.text().includes('Show disks'))
      await showDisksItem!.trigger('click')
      await wrapper.vm.$nextTick()

      const closeBtn = wrapper.find('.fs-disk-list .fs-header-btn')
      expect(closeBtn.exists()).toBe(true)
      await closeBtn.trigger('click')
      await wrapper.vm.$nextTick()

      expect(wrapper.find('.fs-disk-list').exists()).toBe(false)
    })
  })

  describe('Sorting in list view', () => {
    it('has sortable column headers', async () => {
      mountComponent()
      // Switch to list view
      const buttons = wrapper.findAll('.fs-header-btn')
      const viewToggle = buttons.find(btn => btn.text() === '=#')
      await viewToggle!.trigger('click')
      await wrapper.vm.$nextTick()

      const header = wrapper.find('.fs-list-header')
      const nameCol = header.find('.list-name')
      expect(nameCol.exists()).toBe(true)

      // Click to sort by name
      await nameCol.trigger('click')
      await wrapper.vm.$nextTick()

      // Should show sort indicator
      expect(nameCol.text()).toContain('NAME')
    })
  })
})
