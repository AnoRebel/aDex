<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fp-overlay"
      role="dialog"
      aria-labelledby="fp-title"
      aria-modal="true"
      @click.self="close"
      @keydown.esc="close"
    >
      <div class="fp-card mod-panel">
        <div class="fp-header">
          <div id="fp-title" class="fp-title">
            <span class="fp-bracket">[</span>
            PROPERTIES
            <span class="fp-bracket">]</span>
          </div>
          <button class="fp-close" type="button" @click="close" aria-label="Close">
            X
          </button>
        </div>

        <div v-if="loading" class="fp-state">Loading…</div>
        <div v-else-if="error" class="fp-state fp-state-error">{{ error }}</div>
        <dl v-else class="fp-grid">
          <div class="fp-row">
            <dt>Name</dt>
            <dd class="fp-truncate">{{ info?.name }}</dd>
          </div>
          <div class="fp-row">
            <dt>Path</dt>
            <dd class="fp-truncate" :title="info?.path">{{ info?.path }}</dd>
          </div>
          <div class="fp-row">
            <dt>Type</dt>
            <dd>{{ info?.isDir ? 'Directory' : 'File' }}</dd>
          </div>
          <div class="fp-row">
            <dt>Size</dt>
            <dd>{{ humanSize(info?.size) }}</dd>
          </div>
          <div class="fp-row">
            <dt>Modified</dt>
            <dd>{{ formatDate(info?.modTime) }}</dd>
          </div>
          <div class="fp-row">
            <dt>Permissions</dt>
            <dd class="fp-mono">{{ info?.permissions || '—' }}</dd>
          </div>
          <div class="fp-row">
            <dt>Owner</dt>
            <dd>{{ info?.owner || '—' }}</dd>
          </div>
          <div class="fp-row">
            <dt>Group</dt>
            <dd>{{ info?.group || '—' }}</dd>
          </div>
        </dl>

        <div class="fp-actions">
          <button class="fp-btn" type="button" @click="close">CLOSE</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * FilePropertiesModal — read-only file/folder info dialog. Driven by the
 * file-manager context menu's "Properties / Info" action.
 *
 * Backend round-trip:
 *   coordinator.GetFileInfo(path)  →  models.FileInfo (with json tags
 *   matching what we render below; see backend/services/filesystem/service.go).
 */
import { ref, watch } from 'vue'
import { GetFileInfo } from '~/lib/wailsjs/coordinator'

interface FileInfo {
  name?: string
  path?: string
  size?: number
  isDir?: boolean
  mode?: number
  modTime?: string
  permissions?: string
  owner?: string
  group?: string
}

const props = defineProps<{
  modelValue: boolean
  path: string | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const info = ref<FileInfo | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

function close() {
  emit('update:modelValue', false)
}

watch(
  () => [props.modelValue, props.path] as const,
  async ([open, path]) => {
    if (!open || !path) return
    loading.value = true
    error.value = null
    info.value = null
    try {
      info.value = (await GetFileInfo(path)) as FileInfo
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to load info'
    } finally {
      loading.value = false
    }
  },
  { immediate: true },
)

function humanSize(bytes: number | undefined): string {
  if (bytes === undefined || bytes === null) return '—'
  if (bytes < 1024) return `${bytes} B`
  const units = ['KiB', 'MiB', 'GiB', 'TiB']
  let v = bytes / 1024
  let i = 0
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(2)} ${units[i]}`
}

function formatDate(iso: string | undefined): string {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}
</script>

<style scoped>
.fp-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
}

.fp-card {
  --corner-cut: 1.2vh;
  width: min(46vw, 80vh);
  max-height: 80vh;
  padding: 2vh 2vw;
  background: var(--bg-glass, rgba(5, 8, 13, 0.85));
  box-shadow: var(--glow-strong);
  display: flex;
  flex-direction: column;
}

.fp-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.4vh;
}

.fp-title {
  font-size: var(--t-xl, 1.85vh);
  letter-spacing: 0.2em;
  color: var(--accent);
}

.fp-bracket {
  opacity: 0.6;
}

.fp-close {
  background: transparent;
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  padding: 0.4vh 1vw;
  cursor: pointer;
  font-family: inherit;
  font-size: var(--t-md);
  letter-spacing: 0.2em;
}

.fp-close:hover {
  background: var(--surface-2);
}

.fp-state {
  font-size: var(--t-md);
  opacity: 0.7;
  padding: 2vh 0;
}

.fp-state-error {
  color: var(--err);
  opacity: 1;
}

.fp-grid {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.6vh 1.5vw;
  font-size: var(--t-md);
  margin: 0 0 1.4vh 0;
  overflow-y: auto;
}

.fp-row {
  display: contents;
}

.fp-row dt {
  text-transform: uppercase;
  letter-spacing: 0.18em;
  opacity: 0.5;
  font-size: var(--t-sm);
  align-self: center;
}

.fp-row dd {
  margin: 0;
  word-break: break-all;
}

.fp-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fp-mono {
  font-family: var(--font_main);
}

.fp-actions {
  display: flex;
  justify-content: flex-end;
}

.fp-btn {
  padding: 0.7vh 2vw;
  background: transparent;
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  font-family: inherit;
  font-size: var(--t-md);
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
}

.fp-btn:hover {
  background: var(--surface-2);
}
</style>
