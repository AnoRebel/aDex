<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="pm-overlay"
      role="dialog"
      :aria-labelledby="`pm-title-${id}`"
      aria-modal="true"
      tabindex="-1"
      @click.self="close"
      @keydown="onKeydown"
    >
      <div class="pm-card mod-panel">
        <!-- Header bar -->
        <div class="pm-header">
          <div :id="`pm-title-${id}`" class="pm-title">
            <span class="pm-bracket">[</span>
            PROCESS MANAGER
            <span class="pm-bracket">]</span>
            <span class="pm-counter">{{ counterLabel }}</span>
          </div>
          <button
            class="pm-close"
            type="button"
            aria-label="Close process manager"
            @click="close"
          >
            X
          </button>
        </div>

        <!-- Filter + sort row -->
        <div class="pm-toolbar">
          <div class="pm-filter">
            <span class="pm-filter-prefix">/</span>
            <input
              ref="filterInputRef"
              v-model="filterRaw"
              type="text"
              class="pm-filter-input"
              placeholder="filter (name, pid, user, command)"
              autocomplete="off"
              spellcheck="false"
              @keydown.stop
            />
          </div>
          <div class="pm-sort">
            <label class="pm-sort-label">SORT</label>
            <select v-model="sortKey" class="pm-sort-select">
              <option value="cpu+mem">CPU+MEM</option>
              <option value="cpu">CPU</option>
              <option value="memory">MEM</option>
              <option value="name">NAME</option>
              <option value="pid">PID</option>
              <option value="user">USER</option>
            </select>
            <button
              class="pm-sort-dir"
              type="button"
              :aria-label="sortDesc ? 'Descending' : 'Ascending'"
              @click="sortDesc = !sortDesc"
            >
              {{ sortDesc ? 'DESC' : 'ASC' }}
            </button>
          </div>
        </div>

        <!-- Column headers — sync with row template column-widths -->
        <div class="pm-header-row">
          <span class="col-pid">PID</span>
          <span class="col-name">NAME</span>
          <span class="col-user">USER</span>
          <span class="col-status">STATUS</span>
          <span class="col-cpu">CPU%</span>
          <span class="col-mem">MEM%</span>
          <span class="col-cmd">COMMAND</span>
          <span class="col-actions">ACTIONS</span>
        </div>

        <!-- Virtualised rows -->
        <div ref="listEl" v-bind="containerProps" class="pm-rows">
          <div v-bind="wrapperProps">
            <div
              v-for="{ index, data } in renderedList"
              :key="data.pid"
              class="pm-row"
              :class="{ active: index === selectedIndex }"
              :style="{ height: ROW_HEIGHT_PX + 'px' }"
              :title="rowTooltip(data)"
              @click="selectRow(index)"
              @dblclick="openInfo(data)"
              @contextmenu.prevent="onContextMenu($event, data)"
            >
              <span class="col-pid">{{ data.pid }}</span>
              <span class="col-name">{{ data.name }}</span>
              <span class="col-user">{{ data.user || '—' }}</span>
              <span class="col-status">{{ data.status || '—' }}</span>
              <span class="col-cpu">{{ data.cpu.toFixed(1) }}</span>
              <span class="col-mem">{{ data.memory.toFixed(1) }}</span>
              <span class="col-cmd">{{ data.command || data.name }}</span>
              <span class="col-actions" @click.stop>
                <button
                  type="button"
                  class="pm-act"
                  title="Process info"
                  @click="openInfo(data)"
                >
                  i
                </button>
                <button
                  type="button"
                  class="pm-act"
                  title="Copy PID"
                  @click="actCopyPid(data)"
                >
                  #
                </button>
                <button
                  type="button"
                  class="pm-act"
                  title="Terminate (SIGTERM)"
                  @click="actTerm(data)"
                >
                  T
                </button>
                <button
                  type="button"
                  class="pm-act"
                  title="Stop (SIGSTOP)"
                  @click="actStop(data)"
                >
                  ||
                </button>
                <button
                  type="button"
                  class="pm-act"
                  title="Continue (SIGCONT)"
                  @click="actCont(data)"
                >
                  &gt;
                </button>
                <button
                  type="button"
                  class="pm-act pm-act-danger"
                  title="Kill (SIGKILL)"
                  @click="actKill(data)"
                >
                  X
                </button>
              </span>
            </div>
          </div>
        </div>

        <!-- Empty state -->
        <div v-if="filteredProcesses.length === 0" class="pm-empty">
          <span style="opacity: 0.45;">
            {{ rawProcesses.length === 0 ? 'Waiting for data...' : 'No matches' }}
          </span>
        </div>

        <!-- Footer with shortcut legend -->
        <div class="pm-footer">
          <span><kbd>↑</kbd><kbd>↓</kbd> select</span>
          <span><kbd>Enter</kbd> info</span>
          <span><kbd>Del</kbd> term</span>
          <span><kbd>⇧</kbd><kbd>Del</kbd> kill</span>
          <span><kbd>/</kbd> filter</span>
          <span><kbd>Esc</kbd> close</span>
        </div>

        <!-- Sub-modals -->
        <AdexContextMenu
          v-model="cmOpen"
          :items="cmItems"
          :x="cmX"
          :y="cmY"
          aria-label="Process context menu"
        />
        <ProcessInfoModal v-model="infoOpen" :proc="cmTarget" />
        <ConfirmDialog
          v-model="confirmOpen"
          :title="confirmCfg.title"
          :message="confirmCfg.message"
          :confirm-label="confirmCfg.confirmLabel"
          :danger="confirmCfg.danger"
          @confirm="onConfirmConfirm"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick, useId } from 'vue'
import {
  useClipboard,
  refDebounced,
  useVirtualList,
  onKeyStroke,
} from '@vueuse/core'
import { useSystemStore } from '~/stores/system'
import { SignalProcess } from '~/lib/wailsjs/coordinator'
import { useAdexAudio } from '~/composables/useAdexAudio'
import AdexContextMenu, {
  type ContextMenuItems,
} from '~/components/ui/AdexContextMenu.vue'
import ProcessInfoModal from '~/components/system/ProcessInfoModal.vue'
import ConfirmDialog from '~/components/ui/ConfirmDialog.vue'

interface ProcessRow {
  pid: number
  name: string
  cpu: number
  memory: number
  user?: string
  status?: string
  command?: string
}

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const id = useId()
const systemStore = useSystemStore()

// Bigger than the sidebar toplist (24) because we have more vertical
// room here. Each row is ~24px so the modal viewport easily shows
// 20+ rows at typical desktop heights.
const ROW_HEIGHT_PX = 24

const rawProcesses = computed<ProcessRow[]>(() => {
  const procs = systemStore.processes ?? systemStore.topProcesses ?? []
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

// ---- Filter + sort ----

const filterRaw = ref('')
const filterDebounced = refDebounced(filterRaw, 120)

type SortKey = 'cpu+mem' | 'cpu' | 'memory' | 'name' | 'pid' | 'user'
const sortKey = ref<SortKey>('cpu+mem')
const sortDesc = ref(true)

const filteredProcesses = computed<ProcessRow[]>(() => {
  const q = filterDebounced.value.trim().toLowerCase()
  const all = rawProcesses.value
  const filtered = q
    ? all.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          String(p.pid).includes(q) ||
          (p.user ?? '').toLowerCase().includes(q) ||
          (p.command ?? '').toLowerCase().includes(q),
      )
    : all
  const dir = sortDesc.value ? -1 : 1
  return [...filtered].sort((a, b) => {
    switch (sortKey.value) {
      case 'cpu':
        return (a.cpu - b.cpu) * dir
      case 'memory':
        return (a.memory - b.memory) * dir
      case 'name':
        return a.name.localeCompare(b.name) * dir
      case 'pid':
        return (a.pid - b.pid) * dir
      case 'user':
        return (a.user ?? '').localeCompare(b.user ?? '') * dir
      case 'cpu+mem':
      default:
        return (a.cpu + a.memory - (b.cpu + b.memory)) * dir
    }
  })
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
    overscan: 10,
  })

// ---- Selection & keyboard nav ----

const selectedIndex = ref(0)
const filterInputRef = ref<HTMLInputElement | null>(null)
const listEl = ref<HTMLElement | null>(null)

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
  nextTick(() => scrollTo(next))
}

function onKeydown(e: KeyboardEvent) {
  const inFilter = e.target === filterInputRef.value
  if (inFilter && e.key !== 'Escape' && e.key !== 'Enter') return

  switch (e.key) {
    case 'Escape':
      if (inFilter && filterRaw.value) {
        filterRaw.value = ''
        e.preventDefault()
        return
      }
      e.preventDefault()
      close()
      break
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
      moveSelection(15)
      break
    case 'PageUp':
      e.preventDefault()
      moveSelection(-15)
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
      if (e.shiftKey) {
        actKill(proc)
      } else {
        actTerm(proc)
      }
      break
    }
    case '/':
      if (!inFilter) {
        e.preventDefault()
        filterInputRef.value?.focus()
      }
      break
  }
}

// Global ESC to close even when no element inside has focus
onKeyStroke('Escape', (e) => {
  if (!props.modelValue) return
  e.preventDefault()
  close()
})

// Focus filter when opened so the user can immediately start typing.
watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      nextTick(() => filterInputRef.value?.focus())
    } else {
      // reset transient state on close
      filterRaw.value = ''
      cmOpen.value = false
    }
  },
)

// ---- Actions ----

const cmOpen = ref(false)
const cmX = ref(0)
const cmY = ref(0)
const cmTarget = ref<ProcessRow | null>(null)

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

function openConfirm(cfg: Partial<typeof confirmCfg> & { handler: ConfirmHandler }) {
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
    console.error('[pm] confirm action failed:', err)
  }
}

function onContextMenu(e: MouseEvent, proc: ProcessRow) {
  cmTarget.value = proc
  cmX.value = e.clientX
  cmY.value = e.clientY
  cmOpen.value = true
}

function openInfo(proc: ProcessRow) {
  cmTarget.value = proc
  infoOpen.value = true
}

const { copy: copyToClipboard } = useClipboard({ legacy: true })

async function actCopyPid(proc: ProcessRow) {
  try {
    await copyToClipboard(String(proc.pid))
  } catch (err) {
    console.error('[pm] copy pid failed:', err)
  }
}

async function sendSignal(
  proc: ProcessRow,
  signal: 'SIGTERM' | 'SIGKILL' | 'SIGSTOP' | 'SIGCONT',
) {
  try {
    await SignalProcess(proc.pid, signal)
  } catch (err) {
    console.error(`[pm] ${signal} on pid ${proc.pid} failed:`, err)
  }
}

function actTerm(proc: ProcessRow) { sendSignal(proc, 'SIGTERM') }
function actStop(proc: ProcessRow) { sendSignal(proc, 'SIGSTOP') }
function actCont(proc: ProcessRow) { sendSignal(proc, 'SIGCONT') }

function actKill(proc: ProcessRow) {
  // Handgun-click chambering on dialog OPEN.
  try { useAdexAudio().playCue('destructive') } catch { /* non-fatal */ }
  openConfirm({
    title: 'KILL PROCESS',
    message: `Send SIGKILL to "${proc.name}" (PID ${proc.pid})? This is immediate and cannot be undone.`,
    confirmLabel: 'KILL',
    danger: true,
    handler: () => {
      // Gunshot on CONFIRM — same audio narrative as quit + toplist.
      try { useAdexAudio().playCue('gunshot') } catch { /* non-fatal */ }
      return sendSignal(proc, 'SIGKILL')
    },
  })
}

const cmItems = computed<ContextMenuItems>(() => {
  const t = cmTarget.value
  if (!t) return []
  return [
    [
      { label: 'PROCESS INFO', icon: '[i]', onSelect: () => openInfo(t) },
      { label: 'COPY PID', icon: '[#]', onSelect: () => actCopyPid(t) },
    ],
    [
      { label: 'TERMINATE', icon: '[T]', onSelect: () => actTerm(t), kbds: ['SIGTERM'] },
      { label: 'STOP', icon: '[||]', onSelect: () => actStop(t), kbds: ['SIGSTOP'] },
      { label: 'CONTINUE', icon: '[>]', onSelect: () => actCont(t), kbds: ['SIGCONT'] },
    ],
    [
      { label: 'KILL', icon: '[X]', onSelect: () => actKill(t), kbds: ['SIGKILL'], danger: true },
    ],
  ]
})

function rowTooltip(p: ProcessRow): string {
  const parts: string[] = [
    `${p.name} (PID ${p.pid})`,
    `CPU ${p.cpu.toFixed(1)}%  MEM ${p.memory.toFixed(1)}%`,
  ]
  if (p.user) parts.push(`User: ${p.user}`)
  if (p.status) parts.push(`Status: ${p.status}`)
  if (p.command) parts.push(p.command)
  return parts.join('\n')
}

function close() {
  emit('update:modelValue', false)
}
</script>

<style scoped>
/* ─── Overlay ────────────────────────────────────────────────────────── */
.pm-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(3px);
  padding: clamp(0.5rem, 2vh, 2rem) clamp(0.5rem, 2vw, 2rem);
}

/* ─── Card ──────────────────────────────────────────────────────────── */
/* Sizing strategy:
 *   - Width grows with the viewport via clamp() so it's usable on a
 *     1280-wide laptop AND a 4K monitor. min(95vw, 1800px) caps it on
 *     ultrawides where 95vw would be unreadable line-length.
 *   - Height uses dvh (user preference over vh) so the iOS/Android-style
 *     dynamic viewport doesn't push the footer off-screen when the
 *     virtual keyboard appears. Falls back to vh via the second value.
 *   - Flex column layout means the rows region (.pm-rows) is the only
 *     flexible child — header/footer stay pinned, rows scroll. */
.pm-card {
  width: min(95vw, 1800px);
  height: min(92dvh, 92vh);
  max-height: 92dvh;
  display: flex;
  flex-direction: column;
  padding: clamp(0.8vh, 1.2vh, 1.6vh) clamp(0.8vw, 1.4vw, 2vw);
  background: var(--bg-glass, rgba(5, 8, 13, 0.88));
  box-shadow: var(--glow-strong);
  overflow: hidden;
}

/* ─── Header ────────────────────────────────────────────────────────── */
.pm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1vw;
  margin-bottom: 0.8vh;
  flex: 0 0 auto;
}

.pm-title {
  font-size: clamp(1rem, 1.85vh, 1.4rem);
  letter-spacing: 0.18em;
  color: var(--accent);
  display: flex;
  align-items: center;
  gap: 0.6vw;
  flex-wrap: wrap;
}

.pm-bracket { opacity: 0.55; }

.pm-counter {
  font-size: clamp(0.7rem, 1.1vh, 0.95rem);
  letter-spacing: 0.1em;
  opacity: 0.7;
  margin-left: 0.3vw;
}

.pm-close {
  background: transparent;
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color));
  color: var(--accent);
  padding: 0.4vh 1vw;
  cursor: pointer;
  font-family: inherit;
  font-size: clamp(0.75rem, 1.1vh, 1rem);
  letter-spacing: 0.2em;
  flex: 0 0 auto;
}

.pm-close:hover { background: var(--surface-2, rgba(255, 255, 255, 0.06)); }

/* ─── Toolbar (filter + sort) ───────────────────────────────────────── */
.pm-toolbar {
  display: flex;
  gap: 1vw;
  align-items: center;
  flex-wrap: wrap;
  padding: 0.5vh 0;
  margin-bottom: 0.5vh;
  border-bottom: var(--rule-width, 1px) solid var(--rule, var(--border-color));
  flex: 0 0 auto;
}

.pm-filter {
  display: flex;
  align-items: center;
  gap: 0.4vw;
  flex: 1 1 240px;
  min-width: 0;
}

.pm-filter-prefix { opacity: 0.5; user-select: none; }

.pm-filter-input {
  flex: 1;
  min-width: 0;
  background: transparent;
  border: none;
  outline: none;
  color: var(--accent, var(--color_accent));
  font: inherit;
  font-size: clamp(0.75rem, 1.1vh, 1rem);
  padding: 0.2vh 0;
}

.pm-filter-input::placeholder { opacity: 0.35; }

.pm-sort {
  display: flex;
  align-items: center;
  gap: 0.4vw;
  flex: 0 1 auto;
}

.pm-sort-label {
  font-size: clamp(0.65rem, 0.95vh, 0.85rem);
  letter-spacing: 0.15em;
  opacity: 0.5;
}

.pm-sort-select,
.pm-sort-dir {
  background: transparent;
  color: var(--accent);
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color));
  padding: 0.3vh 0.6vw;
  font-family: inherit;
  font-size: clamp(0.7rem, 1vh, 0.95rem);
  letter-spacing: 0.1em;
  cursor: pointer;
}

.pm-sort-dir:hover,
.pm-sort-select:hover { background: var(--surface-2, rgba(255, 255, 255, 0.06)); }

/* ─── Table ──────────────────────────────────────────────────────────
 * Grid columns are sized in fr units so they reflow naturally on any
 * card width. The wide command column (3fr) is the flex of the table
 * and will absorb extra space on large screens. */
.pm-header-row,
.pm-row {
  display: grid;
  grid-template-columns:
    minmax(56px, 0.7fr)   /* pid */
    minmax(120px, 1.4fr)  /* name */
    minmax(80px, 0.9fr)   /* user */
    minmax(70px, 0.7fr)   /* status */
    minmax(56px, 0.6fr)   /* cpu */
    minmax(56px, 0.6fr)   /* mem */
    minmax(180px, 3fr)    /* command */
    minmax(180px, 1.4fr); /* actions */
  align-items: center;
  gap: 0.8vw;
  padding: 0.2vh 0.5vw;
  font-size: clamp(0.75rem, 1.05vh, 0.95rem);
}

.pm-header-row {
  font-size: clamp(0.65rem, 0.95vh, 0.85rem);
  letter-spacing: 0.18em;
  opacity: 0.6;
  text-transform: uppercase;
  border-bottom: var(--rule-width, 1px) solid var(--rule, var(--border-color));
  padding-bottom: 0.4vh;
  flex: 0 0 auto;
}

.pm-rows {
  flex: 1 1 0;
  min-height: 0;
  overflow: auto;
  scrollbar-width: thin;
  scrollbar-color:
    rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.25)
    transparent;
}

.pm-rows::-webkit-scrollbar { width: 0.5vw; }
.pm-rows::-webkit-scrollbar-track { background: transparent; }
.pm-rows::-webkit-scrollbar-thumb {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.25);
}

.pm-row {
  cursor: pointer;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.pm-row:hover { background: rgba(255, 255, 255, 0.04); }

.pm-row.active {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.16);
  outline: var(--rule-width, 1px) solid var(--accent);
  outline-offset: -1px;
}

.col-pid,
.col-cpu,
.col-mem { font-family: var(--font_main); }

.col-name,
.col-user,
.col-status,
.col-cmd {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-cmd { opacity: 0.75; }

.col-actions {
  display: flex;
  gap: 0.3vw;
  justify-content: flex-end;
  flex-wrap: nowrap;
}

.pm-act {
  background: transparent;
  border: var(--rule-width, 1px) solid var(--rule, var(--border-color));
  color: var(--accent);
  font-family: inherit;
  font-size: clamp(0.65rem, 0.95vh, 0.85rem);
  letter-spacing: 0.05em;
  padding: 0 0.5vw;
  min-width: 1.6em;
  height: 1.8em;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.pm-act:hover { background: var(--surface-2, rgba(255, 255, 255, 0.08)); }

.pm-act-danger {
  border-color: var(--error, #ef4444);
  color: var(--error, #ef4444);
}

.pm-act-danger:hover { background: rgba(239, 68, 68, 0.18); }

/* ─── Empty + Footer ────────────────────────────────────────────────── */
.pm-empty {
  padding: 1vh 0.5vw;
  text-align: center;
  font-size: clamp(0.75rem, 1.05vh, 0.95rem);
}

.pm-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4vw 1.2vw;
  padding-top: 0.6vh;
  margin-top: 0.4vh;
  border-top: var(--rule-width, 1px) solid var(--rule, var(--border-color));
  font-size: clamp(0.65rem, 0.95vh, 0.8rem);
  letter-spacing: 0.08em;
  opacity: 0.55;
  flex: 0 0 auto;
}

.pm-footer kbd {
  display: inline-block;
  padding: 0 0.4em;
  margin: 0 0.15em;
  border: 1px solid currentColor;
  border-radius: 2px;
  font-family: var(--font_main);
  font-size: 0.85em;
}

/* ─── Responsive breakpoints ────────────────────────────────────────── */
/* Tablet: drop USER + STATUS columns, keep everything else readable */
@media (max-width: 960px) {
  .pm-header-row,
  .pm-row {
    grid-template-columns:
      minmax(48px, 0.7fr)
      minmax(110px, 1.5fr)
      minmax(50px, 0.6fr)
      minmax(50px, 0.6fr)
      minmax(140px, 2.6fr)
      minmax(150px, 1.4fr);
  }
  .col-user,
  .col-status { display: none; }
  .pm-card { padding: 1vh 1.2vw; }
}

/* Phone: command column also disappears, action buttons go compact */
@media (max-width: 640px) {
  .pm-header-row,
  .pm-row {
    grid-template-columns:
      minmax(40px, 0.7fr)
      minmax(90px, 1.6fr)
      minmax(44px, 0.55fr)
      minmax(44px, 0.55fr)
      minmax(140px, 1.5fr);
    gap: 0.5vw;
  }
  .col-cmd { display: none; }
  .pm-act {
    padding: 0 0.35em;
    min-width: 1.4em;
    height: 1.6em;
  }
  .pm-footer { gap: 0.3vw 0.8vw; }
  .pm-toolbar { gap: 0.6vw; }
  .pm-sort-label { display: none; }
}

/* Very narrow: hide the lower-priority action buttons (STOP/CONT)
 * so TERM + KILL + INFO still fit on a small screen. */
@media (max-width: 480px) {
  .col-actions .pm-act:nth-child(4),
  .col-actions .pm-act:nth-child(5) { display: none; }
  .pm-title { font-size: 0.9rem; }
  .pm-card { padding: 0.8vh 0.8vw; }
}
</style>
