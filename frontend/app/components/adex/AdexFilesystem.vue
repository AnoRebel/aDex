<template>
  <div class="adex-filesystem">
    <!-- Header -->
    <div class="fs-header">
      <div class="fs-header-left">
        <span class="fs-header-title">FILESYSTEM</span>
        <button class="fs-header-btn" @click="toggleDotfiles" :title="showDotfiles ? 'Hide dotfiles' : 'Show dotfiles'">
          {{ showDotfiles ? '[.*]' : '[a-z]' }}
        </button>
        <button class="fs-header-btn" @click="toggleViewMode" :title="viewMode === 'grid' ? 'List view' : 'Grid view'">
          {{ viewMode === 'grid' ? '=#' : '::' }}
        </button>
      </div>
      <span class="fs-header-path">{{ currentPath }}</span>
    </div>

    <!-- File display -->
    <div class="fs-display" ref="displayRef">
      <!-- Grid view -->
      <div v-if="viewMode === 'grid'" class="fs-grid">
        <!-- Show disks item -->
        <div
          class="fs-item fs-item-special"
          @click="showDisks"
          title="Show disk mounts"
        >
          <div class="fs-item-icon">[/]</div>
          <div class="fs-item-name">Show disks</div>
        </div>

        <!-- Go up item -->
        <div
          v-if="currentPath !== '/'"
          class="fs-item fs-item-special"
          @click="goUp"
          title="Go to parent directory"
        >
          <div class="fs-item-icon">[..]</div>
          <div class="fs-item-name">Go up</div>
        </div>

        <!-- Directories -->
        <div
          v-for="dir in filteredDirectories"
          :key="dir.path"
          class="fs-item fs-item-dir"
          :class="{ active: selectedItem === dir.path }"
          @click="handleItemClick(dir)"
          @dblclick="navigateToDir(dir.path)"
          :title="dir.name"
        >
          <div class="fs-item-icon">[D]</div>
          <div class="fs-item-name">{{ dir.name }}</div>
        </div>

        <!-- Files -->
        <div
          v-for="file in filteredFiles"
          :key="file.path"
          class="fs-item fs-item-file"
          :class="{ active: selectedItem === file.path }"
          @click="handleItemClick(file)"
          :title="`${file.name} (${formatSize(file.size)})`"
        >
          <div class="fs-item-icon">{{ getFileIcon(file) }}</div>
          <div class="fs-item-name">{{ file.name }}</div>
        </div>

        <!-- Empty directory message -->
        <div v-if="filteredDirectories.length === 0 && filteredFiles.length === 0" class="fs-empty">
          DIRECTORY EMPTY
        </div>
      </div>

      <!-- List view -->
      <div v-else class="fs-list">
        <!-- List header -->
        <div class="fs-list-header">
          <span class="list-name" @click="setSortBy('name')">
            NAME {{ sortBy === 'name' ? (sortOrder === 'asc' ? '^' : 'v') : '' }}
          </span>
          <span class="list-type" @click="setSortBy('type')">
            TYPE {{ sortBy === 'type' ? (sortOrder === 'asc' ? '^' : 'v') : '' }}
          </span>
          <span class="list-size" @click="setSortBy('size')">
            SIZE {{ sortBy === 'size' ? (sortOrder === 'asc' ? '^' : 'v') : '' }}
          </span>
          <span class="list-date" @click="setSortBy('modified')">
            MODIFIED {{ sortBy === 'modified' ? (sortOrder === 'asc' ? '^' : 'v') : '' }}
          </span>
        </div>

        <!-- Show disks -->
        <div class="fs-list-item fs-list-item-special" @click="showDisks">
          <span class="list-name">[/] Show disks</span>
          <span class="list-type">--</span>
          <span class="list-size">--</span>
          <span class="list-date">--</span>
        </div>

        <!-- Go up -->
        <div
          v-if="currentPath !== '/'"
          class="fs-list-item fs-list-item-special"
          @click="goUp"
        >
          <span class="list-name">[..] Go up</span>
          <span class="list-type">--</span>
          <span class="list-size">--</span>
          <span class="list-date">--</span>
        </div>

        <!-- Directories -->
        <div
          v-for="dir in filteredDirectories"
          :key="dir.path"
          class="fs-list-item fs-list-item-dir"
          :class="{ active: selectedItem === dir.path }"
          @click="handleItemClick(dir)"
          @dblclick="navigateToDir(dir.path)"
        >
          <span class="list-name">[D] {{ dir.name }}</span>
          <span class="list-type">DIR</span>
          <span class="list-size">{{ dir.itemCount != null ? dir.itemCount + ' items' : '--' }}</span>
          <span class="list-date">{{ formatDate(dir.modified) }}</span>
        </div>

        <!-- Files -->
        <div
          v-for="file in filteredFiles"
          :key="file.path"
          class="fs-list-item fs-list-item-file"
          :class="{ active: selectedItem === file.path }"
          @click="handleItemClick(file)"
        >
          <span class="list-name">{{ getFileIcon(file) }} {{ file.name }}</span>
          <span class="list-type">{{ file.extension || 'FILE' }}</span>
          <span class="list-size">{{ formatSize(file.size) }}</span>
          <span class="list-date">{{ formatDate(file.modified) }}</span>
        </div>

        <!-- Empty -->
        <div v-if="filteredDirectories.length === 0 && filteredFiles.length === 0" class="fs-list-item fs-empty-list">
          <span class="list-name">DIRECTORY EMPTY</span>
          <span class="list-type">--</span>
          <span class="list-size">--</span>
          <span class="list-date">--</span>
        </div>
      </div>

      <!-- Disk list overlay -->
      <div v-if="showingDisks" class="fs-disk-list">
        <div class="fs-disk-list-header">
          <span>MOUNT POINTS</span>
          <button class="fs-header-btn" @click="showingDisks = false">[X]</button>
        </div>
        <div
          v-for="disk in diskMounts"
          :key="disk.mountpoint"
          class="fs-disk-entry"
          @click="navigateToDir(disk.mountpoint)"
        >
          <div class="fs-disk-entry-top">
            <span class="fs-disk-device">{{ disk.device }}</span>
            <span class="fs-disk-mount">{{ disk.mountpoint }}</span>
          </div>
          <div class="fs-disk-entry-bar">
            <div class="disk-fill">
              <div
                class="disk-fill-inner"
                :style="{ width: `${disk.usage}%` }"
                :class="{ 'disk-warn': disk.usage > 80, 'disk-crit': disk.usage > 95 }"
              />
            </div>
            <span class="fs-disk-usage">{{ disk.usage.toFixed(1) }}%</span>
          </div>
          <div class="fs-disk-entry-info">
            <span>{{ formatSize(disk.used) }} / {{ formatSize(disk.total) }}</span>
            <span>{{ disk.fstype }}</span>
          </div>
        </div>
        <div v-if="diskMounts.length === 0" class="fs-disk-entry">
          <span>Loading disk information...</span>
        </div>
      </div>
    </div>

    <!-- Disk usage bar (primary mount) -->
    <div class="fs-disk-bar">
      <span class="disk-bar-label">{{ primaryMount.mountpoint }}</span>
      <div class="disk-fill">
        <div
          class="disk-fill-inner"
          :style="{ width: `${primaryMount.usage}%` }"
          :class="{ 'disk-warn': primaryMount.usage > 80, 'disk-crit': primaryMount.usage > 95 }"
        />
      </div>
      <span class="disk-bar-percent">{{ primaryMount.usage.toFixed(0) }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useFilesystemStore } from '~/stores/filesystem'
import { useTerminalStore } from '~/stores/terminal'
import { useWails } from '~/composables/useWails'

const fsStore = useFilesystemStore()
const terminalStore = useTerminalStore()
const wails = useWails()

// Local state
const viewMode = ref<'grid' | 'list'>('grid')
const showDotfiles = ref(false)
const selectedItem = ref<string | null>(null)
const showingDisks = ref(false)
const diskMounts = ref<Array<{
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usage: number
}>>([])

const sortBy = ref<'name' | 'size' | 'modified' | 'type'>('name')
const sortOrder = ref<'asc' | 'desc'>('asc')

const displayRef = ref<HTMLElement | null>(null)

// Computed: current path from store
const currentPath = computed(() => fsStore.currentPath)

// Filter and sort directories
const filteredDirectories = computed(() => {
  let dirs = [...fsStore.directories]

  // Filter dotfiles
  if (!showDotfiles.value) {
    dirs = dirs.filter((d: any) => !d.name.startsWith('.'))
  }

  // Sort
  dirs.sort((a: any, b: any) => {
    let cmp = 0
    switch (sortBy.value) {
      case 'name':
        cmp = a.name.localeCompare(b.name)
        break
      case 'size':
        cmp = (a.size || 0) - (b.size || 0)
        break
      case 'modified':
        cmp = new Date(a.modified).getTime() - new Date(b.modified).getTime()
        break
      case 'type':
        cmp = 0 // dirs are all the same type
        break
    }
    return sortOrder.value === 'asc' ? cmp : -cmp
  })

  return dirs
})

// Filter and sort files
const filteredFiles = computed(() => {
  let files = [...fsStore.files]

  // Filter dotfiles
  if (!showDotfiles.value) {
    files = files.filter((f: any) => !f.name.startsWith('.'))
  }

  // Sort
  files.sort((a: any, b: any) => {
    let cmp = 0
    switch (sortBy.value) {
      case 'name':
        cmp = a.name.localeCompare(b.name)
        break
      case 'size':
        cmp = (a.size || 0) - (b.size || 0)
        break
      case 'modified':
        cmp = new Date(a.modified).getTime() - new Date(b.modified).getTime()
        break
      case 'type':
        cmp = (a.extension || '').localeCompare(b.extension || '')
        break
    }
    return sortOrder.value === 'asc' ? cmp : -cmp
  })

  return files
})

// Primary mount point for the disk bar
const primaryMount = computed(() => {
  if (diskMounts.value.length > 0) {
    // Find the mount that contains the current path
    const sorted = [...diskMounts.value].sort((a, b) => b.mountpoint.length - a.mountpoint.length)
    const match = sorted.find(d => currentPath.value.startsWith(d.mountpoint))
    if (match) return match
    // Fallback to root
    const root = diskMounts.value.find(d => d.mountpoint === '/')
    if (root) return root
    return diskMounts.value[0]
  }
  return { device: '/', mountpoint: '/', fstype: 'ext4', total: 1, used: 0, free: 1, usage: 0 }
})

// Actions
const toggleDotfiles = () => {
  showDotfiles.value = !showDotfiles.value
}

const toggleViewMode = () => {
  viewMode.value = viewMode.value === 'grid' ? 'list' : 'grid'
}

const navigateToDir = (path: string) => {
  showingDisks.value = false
  selectedItem.value = null
  fsStore.fetchDirectory(path)
  // Scroll display to top
  if (displayRef.value) {
    displayRef.value.scrollTop = 0
  }
}

const goUp = () => {
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  const parentPath = '/' + parts.join('/')
  navigateToDir(parentPath || '/')
}

const handleItemClick = (item: any) => {
  if (item.isDirectory || item.type === 'directory') {
    // Single click selects, double click navigates (handled by @dblclick)
    selectedItem.value = item.path
  } else {
    selectedItem.value = item.path
  }
}

const showDisks = async () => {
  showingDisks.value = true
  try {
    const disks = await wails.system.getDiskUsage()
    if (Array.isArray(disks) && disks.length > 0) {
      diskMounts.value = disks.map((d: any) => ({
        device: d.device || d.Device || '/dev/sda',
        mountpoint: d.mountpoint || d.Mountpoint || d.path || '/',
        fstype: d.fstype || d.Fstype || d.type || 'ext4',
        total: d.total || d.Total || 0,
        used: d.used || d.Used || 0,
        free: d.free || d.Free || 0,
        usage: d.usage || d.Usage || (d.total > 0 ? ((d.used || 0) / d.total) * 100 : 0)
      }))
    }
  } catch (err) {
    console.error('Failed to fetch disk usage:', err)
  }
}

const setSortBy = (field: 'name' | 'size' | 'modified' | 'type') => {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    sortOrder.value = 'asc'
  }
}

// Utility: format file size
const formatSize = (bytes: number): string => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

// Utility: format date
const formatDate = (date: Date | string): string => {
  if (!date) return '--'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '--'
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  return `${month}/${day} ${hours}:${minutes}`
}

// Utility: get text icon for file type
const getFileIcon = (file: any): string => {
  const ext = (file.extension || file.name?.split('.').pop() || '').toLowerCase()

  const iconMap: Record<string, string> = {
    // Code
    js: '{J}', ts: '{T}', vue: '{V}', py: '{P}', go: '{G}',
    rs: '{R}', c: '{C}', cpp: '{+}', h: '{H}', java: '{J}',
    html: '<>', css: '{S}', json: '{}', xml: '<>',
    // Text
    txt: '[T]', md: '[M]', log: '[L]', csv: '[C]',
    // Archives
    zip: '(Z)', tar: '(T)', gz: '(G)', rar: '(R)', '7z': '(7)',
    // Images
    png: '[I]', jpg: '[I]', jpeg: '[I]', gif: '[I]', svg: '[I]', webp: '[I]',
    // Audio/Video
    mp3: '[A]', wav: '[A]', flac: '[A]', mp4: '[V]', mkv: '[V]', avi: '[V]',
    // Documents
    pdf: '[P]', doc: '[D]', docx: '[D]', xls: '[X]', xlsx: '[X]',
    // Executables
    sh: '[#]', bash: '[#]', exe: '[E]', bin: '[B]',
    // Config
    yml: '{Y}', yaml: '{Y}', toml: '{=}', ini: '{I}', conf: '{=}', env: '{=}',
  }

  return iconMap[ext] || '[F]'
}

// Sync with active terminal CWD
watch(
  () => terminalStore.activeSession,
  (session) => {
    if (session?.workingDirectory && session.workingDirectory !== currentPath.value) {
      fsStore.fetchDirectory(session.workingDirectory)
    }
  },
  { deep: true }
)

// Load initial directory and disk info on mount
onMounted(async () => {
  // Load settings from store
  fsStore.loadSettings()
  showDotfiles.value = fsStore.settings.showHiddenFiles
  viewMode.value = (fsStore.settings.viewMode === 'grid' || fsStore.settings.viewMode === 'list')
    ? fsStore.settings.viewMode as 'grid' | 'list'
    : 'grid'

  // Load the current directory
  if (fsStore.files.length === 0 && fsStore.directories.length === 0) {
    await fsStore.fetchDirectory(fsStore.currentPath)
  }

  // Load disk info for the bar
  try {
    const disks = await wails.system.getDiskUsage()
    if (Array.isArray(disks) && disks.length > 0) {
      diskMounts.value = disks.map((d: any) => ({
        device: d.device || d.Device || '/dev/sda',
        mountpoint: d.mountpoint || d.Mountpoint || d.path || '/',
        fstype: d.fstype || d.Fstype || d.type || 'ext4',
        total: d.total || d.Total || 0,
        used: d.used || d.Used || 0,
        free: d.free || d.Free || 0,
        usage: d.usage || d.Usage || (d.total > 0 ? ((d.used || 0) / d.total) * 100 : 0)
      }))
    }
  } catch (err) {
    // Disk info may not be available
  }
})
</script>

<style scoped>
/* The parent CSS classes (adex-filesystem, fs-header, fs-display, fs-grid, fs-item,
   fs-list, fs-list-item, fs-disk-bar) are defined in main.css.
   Below are component-specific overrides and additions. */

.fs-header-left {
  display: flex;
  align-items: center;
  gap: 0.5vw;
}

.fs-header-title {
  text-transform: uppercase;
  letter-spacing: 0.15em;
}

.fs-header-path {
  font-size: 1vh;
  opacity: 0.7;
  max-width: 20vw;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
}

.fs-header-btn {
  background: transparent;
  border: var(--border_width) solid rgba(var(--color_r), var(--color_g), var(--color_b), 0.3);
  color: var(--color_accent);
  font-family: var(--font_main);
  font-size: 1vh;
  padding: 0.1vh 0.3vw;
  cursor: pointer;
  transition: all 0.15s;
  line-height: 1;
}

.fs-header-btn:hover {
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.15);
  border-color: var(--color_accent);
}

/* Grid view specifics */
.fs-item-special {
  opacity: 0.7;
}

.fs-item-special:hover {
  opacity: 1;
}

.fs-item-dir .fs-item-icon {
  color: var(--color_accent);
}

.fs-item-file .fs-item-icon {
  opacity: 0.7;
}

.fs-item.active {
  border-color: var(--color_accent);
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.1);
}

/* List view specifics */
.fs-list-header {
  display: flex;
  align-items: center;
  gap: 1vw;
  padding: 0.3vh 0.5vw;
  font-size: 1vh;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  opacity: 0.5;
  border-bottom: var(--border_width) solid rgba(var(--color_r), var(--color_g), var(--color_b), 0.2);
  cursor: pointer;
  user-select: none;
}

.fs-list-header .list-name { flex: 2; }
.fs-list-header .list-type { flex: 1; }
.fs-list-header .list-size { flex: 1; text-align: right; }
.fs-list-header .list-date { flex: 1.5; text-align: right; }

.fs-list-item-special {
  opacity: 0.6;
}

.fs-list-item-special:hover {
  opacity: 1;
}

.fs-list-item.active {
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.1);
  border-left: 2px solid var(--color_accent);
}

.fs-list-item-dir .list-name {
  color: var(--color_accent);
}

.fs-empty {
  grid-column: 1 / -1;
  text-align: center;
  padding: 2vh 0;
  opacity: 0.4;
  font-size: 1.1vh;
  text-transform: uppercase;
  letter-spacing: 0.2em;
}

.fs-empty-list {
  opacity: 0.4;
}

/* Disk list overlay */
.fs-disk-list {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.92);
  z-index: 10;
  overflow-y: auto;
  padding: 0.5vh 0.5vw;
}

.fs-disk-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 1.1vh;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  padding-bottom: 0.3vh;
  border-bottom: var(--border_width) solid var(--border_color);
  margin-bottom: 0.3vh;
}

.fs-disk-entry {
  padding: 0.4vh 0.3vw;
  border-bottom: var(--border_width) solid rgba(var(--color_r), var(--color_g), var(--color_b), 0.08);
  cursor: pointer;
  transition: background 0.15s;
}

.fs-disk-entry:hover {
  background: rgba(var(--color_r), var(--color_g), var(--color_b), 0.08);
}

.fs-disk-entry-top {
  display: flex;
  justify-content: space-between;
  font-size: 1vh;
  margin-bottom: 0.15vh;
}

.fs-disk-device {
  opacity: 0.6;
}

.fs-disk-mount {
  color: var(--color_accent);
}

.fs-disk-entry-bar {
  display: flex;
  align-items: center;
  gap: 0.5vw;
  margin-bottom: 0.1vh;
}

.fs-disk-entry-bar .disk-fill {
  flex: 1;
  height: 0.5vh;
  background: var(--color_grey);
  position: relative;
  overflow: hidden;
}

.fs-disk-entry-bar .disk-fill-inner {
  height: 100%;
  background: var(--color_accent);
  transition: width 0.5s;
}

.disk-warn {
  background: var(--warning, #f59e0b) !important;
}

.disk-crit {
  background: var(--error, #ef4444) !important;
}

.fs-disk-usage {
  font-size: 0.9vh;
  min-width: 3vw;
  text-align: right;
}

.fs-disk-entry-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.85vh;
  opacity: 0.5;
}

/* Disk bar at bottom */
.disk-bar-label {
  font-size: 0.9vh;
  opacity: 0.6;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
  min-width: 3vw;
}

.disk-bar-percent {
  font-size: 0.9vh;
  min-width: 2.5vw;
  text-align: right;
}

/* Ensure the display is positioned for the overlay */
.fs-display {
  position: relative;
}
</style>
