<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="pi-overlay"
      role="dialog"
      :aria-labelledby="`pi-title-${id}`"
      aria-modal="true"
      @click.self="close"
      @keydown.esc="close"
    >
      <div class="pi-card mod-panel">
        <div class="pi-header">
          <div :id="`pi-title-${id}`" class="pi-title">
            <span class="pi-bracket">[</span>
            PROCESS {{ proc?.pid }}
            <span class="pi-bracket">]</span>
          </div>
          <button class="pi-close" type="button" @click="close" aria-label="Close">
            X
          </button>
        </div>

        <dl v-if="proc" class="pi-grid">
          <div class="pi-row">
            <dt>Name</dt>
            <dd>{{ proc.name }}</dd>
          </div>
          <div class="pi-row">
            <dt>PID</dt>
            <dd class="pi-mono">{{ proc.pid }}</dd>
          </div>
          <div class="pi-row">
            <dt>User</dt>
            <dd>{{ proc.user || '—' }}</dd>
          </div>
          <div class="pi-row">
            <dt>Status</dt>
            <dd>{{ proc.status || '—' }}</dd>
          </div>
          <div class="pi-row">
            <dt>CPU</dt>
            <dd class="pi-mono">{{ proc.cpu?.toFixed(1) ?? '0.0' }}%</dd>
          </div>
          <div class="pi-row">
            <dt>Memory</dt>
            <dd class="pi-mono">{{ proc.memory?.toFixed(1) ?? '0.0' }}%</dd>
          </div>
          <div class="pi-row pi-row-wide">
            <dt>Command</dt>
            <dd class="pi-mono pi-cmd">{{ proc.command || proc.name }}</dd>
          </div>
        </dl>

        <div class="pi-actions">
          <button class="pi-btn" type="button" @click="close">CLOSE</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * ProcessInfoModal — read-only details for a single process. Driven by
 * the toplist's right-click "Process Info" action.
 */
import { useId } from 'vue'

interface ProcessRow {
  pid: number
  name: string
  cpu?: number
  memory?: number
  user?: string
  status?: string
  command?: string
}

defineProps<{
  modelValue: boolean
  proc: ProcessRow | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const id = useId()

function close() {
  emit('update:modelValue', false)
}
</script>

<style scoped>
.pi-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
}

.pi-card {
  --corner-cut: 1.2vh;
  width: min(46vw, 80vh);
  max-height: 80vh;
  padding: 2vh 2vw;
  background: var(--bg-glass, rgba(5, 8, 13, 0.85));
  box-shadow: var(--glow-strong);
  display: flex;
  flex-direction: column;
}

.pi-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.4vh;
}

.pi-title {
  font-size: var(--t-xl, 1.85vh);
  letter-spacing: 0.2em;
  color: var(--accent);
}

.pi-bracket {
  opacity: 0.6;
}

.pi-close {
  background: transparent;
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  padding: 0.4vh 1vw;
  cursor: pointer;
  font-family: inherit;
  font-size: var(--t-md);
  letter-spacing: 0.2em;
}

.pi-close:hover { background: var(--surface-2); }

.pi-grid {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 0.6vh 1.5vw;
  font-size: var(--t-md);
  margin: 0 0 1.4vh 0;
  overflow-y: auto;
}

.pi-row { display: contents; }

.pi-row dt {
  text-transform: uppercase;
  letter-spacing: 0.18em;
  opacity: 0.5;
  font-size: var(--t-sm);
  align-self: center;
}

.pi-row dd { margin: 0; word-break: break-all; }

.pi-row-wide dd {
  grid-column: 1 / -1;
  margin-top: 0.5vh;
}

.pi-mono { font-family: var(--font_main); }

.pi-cmd {
  background: rgba(0, 0, 0, 0.35);
  padding: 0.6vh 0.8vw;
  border: var(--rule-width) solid var(--rule);
  font-size: var(--t-sm);
  white-space: pre-wrap;
}

.pi-actions {
  display: flex;
  justify-content: flex-end;
}

.pi-btn {
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

.pi-btn:hover { background: var(--surface-2); }
</style>
