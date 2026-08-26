<template>
  <div class="mod-panel mod-toplist" tabindex="0" @keydown="onKeydown">
    <div class="section-title">
      <span class="title-left">PROCESSES</span>
      <span class="title-right">
        <span class="toplist-count">{{ counterLabel }}</span>
        <button
          type="button"
          class="toplist-expand-btn"
          title="Open process manager (Ctrl+Shift+P)"
          aria-label="Open process manager"
          @click.stop="managerOpen = true"
        >
          ⤢
        </button>
      </span>
    </div>

    <!-- Filter input. Stays inline (no popover) so keyboard-only users can
         see what they typed. We deliberately don't autofocus so opening
         a tab doesn't steal focus from the terminal. -->
    <div class="toplist-filter">
      <span class="toplist-filter-prefix">/</span>
      <input
        ref="filterInputRef"
        v-model="filterRaw"
        type="text"
        class="toplist-filter-input"
        placeholder="filter (name, pid)"
        autocomplete="off"
        spellcheck="false"
        @keydown.stop
      />
    </div>

    <!-- Column headers -->
    <div class="toplist-header">
      <span class="col-pid">PID</span>
      <span class="col-name">NAME</span>
      <span class="col-cpu">CPU</span>
      <span class="col-mem">MEM</span>
    </div>

    <!-- Virtualised row container. useVirtualList renders only the rows
         in the viewport (+ a small overscan buffer) regardless of how
         many processes the system has. Without virtualization, rendering
         500+ rows on every store tick burns ~30% CPU on a budget laptop. -->
    <div ref="listEl" v-bind="containerProps" class="toplist-rows">
      <div v-bind="wrapperProps">
        <div
          v-for="{ index, data } in renderedList"
          :key="data.pid"
          class="toplist-row"
          :class="{ active: index === selectedIndex }"
          :style="{ height: ROW_HEIGHT_PX + 'px' }"
          :title="`${data.name} (PID ${data.pid})\nCPU ${data.cpu.toFixed(1)}%  MEM ${data.memory.toFixed(1)}%${data.user ? '\\nUser: ' + data.user : ''}${data.command ? '\\n' + data.command : ''}`"
          @click="selectRow(index)"
          @dblclick="openInfo(data)"
          @contextmenu.prevent="onContextMenu($event, data)"
        >
          <span class="col-pid">{{ data.pid }}</span>
          <span class="col-name">{{ data.name }}</span>
          <span class="col-cpu">{{ data.cpu.toFixed(1) }}</span>
          <span class="col-mem">{{ data.memory.toFixed(1) }}</span>
        </div>
      </div>
    </div>

    <!-- Empty state when there's no data at all -->
    <div v-if="filteredProcesses.length === 0" class="toplist-empty">
      <span style="opacity: 0.4;">{{
        rawProcesses.length === 0 ? 'Waiting for data...' : 'No matches'
      }}</span>
    </div>

    <!-- Right-click menu + supporting modals. -->
    <AdexContextMenu
      v-model="cmOpen"
      :items="cmItems"
      :x="cmX"
      :y="cmY"
      aria-label="Process context menu"
    />
    <ProcessInfoModal
      v-model="infoOpen"
      :proc="cmTarget"
    />
    <ConfirmDialog
      v-model="confirmOpen"
      :title="confirmCfg.title"
      :message="confirmCfg.message"
      :confirm-label="confirmCfg.confirmLabel"
      :danger="confirmCfg.danger"
      @confirm="onConfirmConfirm"
    />

    <!-- Expanded view. Driven by both the title-right button AND the
         global Ctrl+Shift+P keybind in pages/index.vue via defineExpose. -->
    <ProcessManagerModal v-model="managerOpen" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { useSystemStore } from '~/stores/system'
import {
  useClipboard,
  refDebounced,
  useVirtualList,
} from '@vueuse/core'
import { SignalProcess } from '~/lib/wailsjs/coordinator'
import { useAdexAudio } from '~/composables/useAdexAudio'
import AdexContextMenu, {
  type ContextMenuItems,
} from '~/components/ui/AdexContextMenu.vue'
import ProcessInfoModal from '~/components/system/ProcessInfoModal.vue'
import ProcessManagerModal from '~/components/system/ProcessManagerModal.vue'
import ConfirmDialog from '~/components/ui/ConfirmDialog.vue'

interface TopProcess {
  pid: number
  name: string
  cpu: number
  memory: number
  // Optional rich fields surfaced by the Process Info modal. They're
  // not displayed in the row itself but the modal renders them and
  // the empty-row placeholder logic doesn't care.
  user?: string
  status?: string
  command?: string
}

const systemStore = useSystemStore()

// Row height drives the virtualizer math. Match the actual row height
// from main.css (.toplist-row padding: 0.15vh + 1vh font ≈ ~14–18px).
// Hardcoded in px because useVirtualList needs an exact integer; small
// drift is tolerated by overscan.
const ROW_HEIGHT_PX = 16

// ---- Source data ----

// All processes from the store. The system store already polls on its
// own cadence; we just consume what's there. Computing on a derived ref
// is cheap because the store's array reference only changes on update.
const rawProcesses = computed<TopProcess[]>(() => {
  const procs = systemStore.processes ?? systemStore.topProcesses ?? []
  // Accept both camelCase + PascalCase + snake_case shapes — gopsutil
  // uses one set on the Go side, the JS bindings the other, and the
  // Pinia store sometimes massages them to a third. Map them all.
  return procs.map((p: any) => ({
    pid: p.pid ?? p.PID ?? 0,
    name: p.name ?? p.Name ?? '',
    cpu: p.cpu ?? p.CPUPercent ?? p.cpuPercent ?? 0,
    memory: p.memory ?? p.MemoryPercent ?? p.memoryPercent ?? 0,
    user: p.user ?? p.User ?? p.username ?? p.Username ?? '',
    status: p.status ?? p.Status ?? p.state ?? p.State ?? '',
    command: p.command ?? p.Command ?? p.cmdline ?? p.Cmdline ?? '',
  }))
})

// ---- Filter ----

const filterRaw = ref('')
// Debounce so a fast typist doesn't trigger a re-filter on every
// keystroke; 120ms feels instant but coalesces ~5 keystrokes/sec.
const filterDebounced = refDebounced(filterRaw, 120)

const filteredProcesses = computed<TopProcess[]>(() => {
  const q = filterDebounced.value.trim().toLowerCase()
  const all = rawProcesses.value
  const filtered = q
    ? all.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          String(p.pid).includes(q)
      )
    : all
  // Sort by combined CPU+MEM weight, desc. Spread first so we never
  // mutate the store's array (would trigger reactivity loops).
  return [...filtered].sort(
    (a, b) => b.cpu + b.memory - (a.cpu + a.memory)
  )
})

const counterLabel = computed(() => {
  const total = rawProcesses.value.length
  const shown = filteredProcesses.value.length
  if (total === 0) return ''
  return shown === total ? String(total) : `${shown}/${total}`
})

// ---- Virtualised list ----

const { list: renderedList, containerProps, wrapperProps, scrollTo } =
  useVirtualList(filteredProcesses, {
    itemHeight: ROW_HEIGHT_PX,
    overscan: 6,
  })

// ---- Selection & keyboard navigation ----

const selectedIndex = ref(0)
const filterInputRef = ref<HTMLInputElement | null>(null)
const listEl = ref<HTMLElement | null>(null)

// Clamp the selection whenever the filtered list changes — without this
// a filter that shrinks the list to 3 rows could leave selectedIndex at
// 47, pointing at nothing.
watch(filteredProcesses, (list) => {
  if (selectedIndex.value >= list.length) {
    selectedIndex.value = Math.max(0, list.length - 1)
  }
}, { flush: 'post' })

function selectRow(index: number) {
  selectedIndex.value = index
}

function moveSelection(delta: number) {
  const total = filteredProcesses.value.length
  if (total === 0) return
  const next = Math.min(Math.max(0, selectedIndex.value + delta), total - 1)
  selectedIndex.value = next
  // Keep the selected row in view. scrollTo accepts an item index.
  nextTick(() => scrollTo(next))
}

function onKeydown(e: KeyboardEvent) {
  // Filter input owns its own typing — only handle nav when the panel
  // root has focus (or the filter input bubbled an Escape).
  const inFilter = e.target === filterInputRef.value
  if (inFilter && e.key !== 'Escape' && e.key !== 'Enter') return

  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      moveSelection(1)
      break
    case 'ArrowUp':
      e.preventDefault()
      moveSelection(-1)
      break
    case 'PageDown':
      e.preventDefault()
      moveSelection(10)
      break
    case 'PageUp':
      e.preventDefault()
      moveSelection(-10)
      break
    case 'Home':
      e.preventDefault()
      selectedIndex.value = 0
      nextTick(() => scrollTo(0))
      break
    case 'End': {
      e.preventDefault()
      const last = Math.max(0, filteredProcesses.value.length - 1)
      selectedIndex.value = last
      nextTick(() => scrollTo(last))
      break
    }
    case 'Enter': {
      const proc = filteredProcesses.value[selectedIndex.value]
      if (proc) openInfo(proc)
      break
    }
    case 'Delete': {
      const proc = filteredProcesses.value[selectedIndex.value]
      if (!proc) break
      cmTarget.value = proc
      // Shift-Delete = SIGKILL (force, with confirm dialog).
      // Plain Delete  = SIGTERM (graceful, no confirm).
      if (e.shiftKey) {
        actKill()
      } else {
        actTerm()
      }
      break
    }
    case 'Escape':
      if (filterRaw.value) {
        filterRaw.value = ''
        e.preventDefault()
      }
      break
    case '/':
      // Quick filter focus when not already typing in it
      if (!inFilter) {
        e.preventDefault()
        filterInputRef.value?.focus()
      }
      break
  }
}

// ---- Context menu state ----

const cmOpen = ref(false)
const cmX = ref(0)
const cmY = ref(0)
const cmTarget = ref<TopProcess | null>(null)

const infoOpen = ref(false)

type ConfirmHandler = () => Promise<void> | void
const confirmOpen = ref(false)
const confirmCfg = reactive({
  title: '',
  message: '',
  confirmLabel: 'OK',
  danger: false,
  handler: (() => undefined) as ConfirmHandler,
})

function openConfirm(cfg: Partial<typeof confirmCfg> & {
  handler: ConfirmHandler
}) {
  confirmCfg.title = cfg.title ?? ''
  confirmCfg.message = cfg.message ?? ''
  confirmCfg.confirmLabel = cfg.confirmLabel ?? 'OK'
  confirmCfg.danger = cfg.danger ?? false
  confirmCfg.handler = cfg.handler
  confirmOpen.value = true
}

async function onConfirmConfirm() {
  try {
    await confirmCfg.handler()
  } catch (err) {
    console.error('[toplist] confirm action failed:', err)
  }
}

function onContextMenu(e: MouseEvent, proc: TopProcess) {
  cmTarget.value = proc
  cmX.value = e.clientX
  cmY.value = e.clientY
  cmOpen.value = true
}

function openInfo(proc: TopProcess) {
  cmTarget.value = proc
  infoOpen.value = true
}

const { copy: copyToClipboard } = useClipboard({ legacy: true })

async function actInfo() {
  infoOpen.value = true
}

async function actSendSignal(signal: 'SIGTERM' | 'SIGKILL' | 'SIGSTOP' | 'SIGCONT') {
  const proc = cmTarget.value
  if (!proc) return
  try {
    await SignalProcess(proc.pid, signal)
  } catch (err) {
    console.error(`[toplist] ${signal} on pid ${proc.pid} failed:`, err)
  }
}

function actTerm() {
  actSendSignal('SIGTERM')
}

function actKill() {
  const proc = cmTarget.value
  if (!proc) return
  // Handgun-click chambering when dialog OPENS — the user is staring
  // at the confirm dialog deciding whether to pull the trigger.
  try { useAdexAudio().playCue('destructive') } catch { /* non-fatal */ }
  openConfirm({
    title: 'KILL PROCESS',
    message: `Send SIGKILL to "${proc.name}" (PID ${proc.pid})? This is immediate and cannot be undone.`,
    confirmLabel: 'KILL',
    danger: true,
    handler: async () => {
      // Gunshot fires on CONFIRM — the trigger pulls.
      try { useAdexAudio().playCue('gunshot') } catch { /* non-fatal */ }
      await SignalProcess(proc.pid, 'SIGKILL')
    },
  })
}

function actStop() { actSendSignal('SIGSTOP') }
function actCont() { actSendSignal('SIGCONT') }

async function actCopyPid() {
  const proc = cmTarget.value
  if (!proc) return
  try {
    await copyToClipboard(String(proc.pid))
  } catch (err) {
    console.error('[toplist] copy pid failed:', err)
  }
}

// Expanded process manager — opened from the title-bar button OR the
// global Ctrl+Shift+P shortcut. Exposed so pages/index.vue can flip it
// without us having to register a duplicate global listener here.
const managerOpen = ref(false)

defineExpose({
  openManager: () => { managerOpen.value = true },
  closeManager: () => { managerOpen.value = false },
  toggleManager: () => { managerOpen.value = !managerOpen.value },
})

const cmItems = computed<ContextMenuItems>(() => {
  if (!cmTarget.value) return []
  return [
    [
      { label: 'PROCESS INFO', icon: '[i]', onSelect: actInfo },
      { label: 'COPY PID',     icon: '[#]', onSelect: actCopyPid },
    ],
    [
      { label: 'TERMINATE',    icon: '[T]', onSelect: actTerm,
        kbds: ['SIGTERM'] },
      { label: 'STOP',         icon: '[||]', onSelect: actStop,
        kbds: ['SIGSTOP'] },
      { label: 'CONTINUE',     icon: '[>]', onSelect: actCont,
        kbds: ['SIGCONT'] },
    ],
    [
      { label: 'KILL', icon: '[X]', onSelect: actKill,
        kbds: ['SIGKILL'], danger: true },
    ],
  ]
})
</script>

<style scoped>
.toplist-filter {
  display: flex;
  align-items: center;
  gap: 0.3vw;
  padding: 0.2vh 0.3vw;
  margin-bottom: 0.2vh;
  border-bottom: var(--border_width) solid var(--border_color);
  font-size: 1vh;
}

.toplist-filter-prefix {
  opacity: 0.5;
  user-select: none;
}

.toplist-filter-input {
  flex: 1;
  min-width: 0;
  background: transparent;
  border: none;
  outline: none;
  color: var(--accent, var(--color_accent));
  font: inherit;
  padding: 0;
}

.toplist-filter-input::placeholder {
  opacity: 0.35;
}

.toplist-rows {
  flex: 1 1 0;
  /* Even on a tightly squeezed column we render at least ~10 rows so
   * the panel is useful — useVirtualList handles the rest via its
   * internal scroll. min-height: 0 alone could collapse the area to
   * a single row when sibling mods take their full natural height. */
  min-height: calc(10 * 16px);
  overflow: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.25) transparent;
}

.toplist-rows::-webkit-scrollbar {
  width: 0.3vw;
}

.toplist-rows::-webkit-scrollbar-track {
  background: transparent;
}

.toplist-rows::-webkit-scrollbar-thumb {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.25);
}

.toplist-row.active {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.18);
  outline: var(--border_width) solid var(--color_accent);
  outline-offset: -1px;
}

.toplist-empty {
  padding: 0.3vh 0.5vw;
  font-size: 0.95vh;
}

/* Title-right cluster: count + expand button live side-by-side without
 * fighting for room. flex on the right span keeps the original eDex
 * layout intact. */
.title-right :deep(.toplist-count),
.toplist-count {
  opacity: 0.7;
  margin-right: 0.4vw;
}

.toplist-expand-btn {
  background: transparent;
  border: none;
  color: var(--accent, var(--color_accent));
  cursor: pointer;
  font-family: inherit;
  font-size: 1.2vh;
  line-height: 1;
  padding: 0 0.2vw;
  opacity: 0.55;
  transition: opacity 120ms;
}

.toplist-expand-btn:hover,
.toplist-expand-btn:focus-visible { opacity: 1; outline: none; }
</style>
