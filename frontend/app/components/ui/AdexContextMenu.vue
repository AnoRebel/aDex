<template>
  <Teleport to="body">
    <Transition name="cm">
      <div
        v-if="open"
        ref="menuRef"
        class="cm mod-panel"
        role="menu"
        :style="positionStyle"
        :aria-label="ariaLabel"
        @keydown.esc="close"
        @click.stop
      >
        <template v-for="(group, gIdx) in items" :key="gIdx">
          <template v-for="(item, iIdx) in group" :key="`${gIdx}-${iIdx}`">
            <div
              v-if="item.type === 'separator'"
              class="cm-sep"
              role="separator"
            />
            <button
              v-else
              type="button"
              role="menuitem"
              class="cm-item"
              :class="{
                'cm-item-disabled': item.disabled,
                'cm-item-danger': item.danger,
              }"
              :disabled="item.disabled"
              @click="handleSelect(item)"
              @mouseenter="hovered = item"
            >
              <span class="cm-icon" v-if="item.icon">{{ item.icon }}</span>
              <span class="cm-label">{{ item.label }}</span>
              <span class="cm-kbd" v-if="item.kbds && item.kbds.length">{{ item.kbds.join('+') }}</span>
            </button>
          </template>
        </template>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
/**
 * AdexContextMenu — themed right-click menu matching the eDex/aDex
 * aesthetic (clip-path corners, all-caps spaced labels, accent rule).
 *
 * Why not @nuxt/ui's UContextMenu directly? It ships Tailwind defaults
 * (rounded corners, soft shadows, Material-style hover) that fight the
 * sci-fi terminal look. This wrapper renders into a portal at fixed x/y
 * positions matching the right-click event's clientX/clientY, with
 * collision-edge clamping and outside-click + Esc dismissal.
 */
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { onClickOutside, useEventListener } from '@vueuse/core'

export interface ContextMenuItem {
  /** Sentinel for visual separator. */
  type?: 'separator'
  /** Visible label (caps-friendly, e.g. "OPEN", "RENAME", "DELETE"). */
  label?: string
  /** Optional left-hand glyph or symbol (kept ASCII to match eDex feel). */
  icon?: string
  /** Optional shortcut hint shown on the right (e.g. ['Del']). */
  kbds?: string[]
  /** Mark as destructive — colors the row with --err. */
  danger?: boolean
  disabled?: boolean
  /** Action invoked on click. */
  onSelect?(): void
}

export type ContextMenuItems = ContextMenuItem[][]

const props = defineProps<{
  /** Two-dimensional array of items grouped by separator. */
  items: ContextMenuItems
  /** Anchor coordinates in viewport space (clientX/clientY). */
  x: number
  y: number
  /** Whether the menu is open. */
  modelValue: boolean
  ariaLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const open = computed(() => props.modelValue)
const menuRef = ref<HTMLElement | null>(null)
const hovered = ref<ContextMenuItem | null>(null)

// Clamp the menu so it never spills off the viewport. Computed at mount
// + on resize via VueUse useEventListener.
const placement = ref({ left: 0, top: 0 })

async function reposition() {
  if (!props.modelValue) return
  await nextTick()
  const el = menuRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const vw = window.innerWidth
  const vh = window.innerHeight
  const margin = 4
  let left = props.x
  let top = props.y
  if (left + rect.width > vw - margin) left = vw - rect.width - margin
  if (top + rect.height > vh - margin) top = vh - rect.height - margin
  if (left < margin) left = margin
  if (top < margin) top = margin
  placement.value = { left, top }
}

const positionStyle = computed(() => ({
  left: `${placement.value.left}px`,
  top: `${placement.value.top}px`,
}))

watch(() => [props.modelValue, props.x, props.y], reposition, { immediate: true })
useEventListener(typeof window !== 'undefined' ? window : null, 'resize', reposition)

onClickOutside(menuRef, () => {
  if (props.modelValue) close()
})

function close() {
  emit('update:modelValue', false)
  hovered.value = null
}

function handleSelect(item: ContextMenuItem) {
  if (item.disabled || item.type === 'separator') return
  try {
    item.onSelect?.()
  } finally {
    close()
  }
}

onBeforeUnmount(() => {
  hovered.value = null
})
</script>

<style scoped>
.cm {
  --corner-cut: 0.6vh;
  position: fixed;
  z-index: 10100;
  min-width: 14vw;
  max-width: 22vw;
  padding: 0.4vh 0;
  background: var(--bg-glass, rgba(5, 8, 13, 0.92));
  box-shadow: var(--glow-mid);
  font-family: var(--font_main);
  user-select: none;
}

.cm-sep {
  height: var(--rule-width);
  margin: 0.4vh 0.6vw;
  background: var(--rule);
  opacity: 0.6;
}

.cm-item {
  display: flex;
  align-items: center;
  gap: 0.6vw;
  width: 100%;
  padding: 0.55vh 1vw;
  background: transparent;
  border: 0;
  color: var(--accent);
  font-family: inherit;
  font-size: var(--t-md);
  letter-spacing: 0.12em;
  text-transform: uppercase;
  text-align: left;
  cursor: pointer;
  transition: background 0.08s linear;
}

.cm-item:hover,
.cm-item:focus-visible {
  background: var(--surface-2);
  outline: none;
}

.cm-item-disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.cm-item-disabled:hover {
  background: transparent;
}

.cm-item-danger {
  color: var(--err);
}

.cm-item-danger:hover {
  background: rgba(239, 68, 68, 0.12);
}

.cm-icon {
  display: inline-block;
  min-width: 1.2vw;
  opacity: 0.65;
  font-size: var(--t-sm);
}

.cm-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cm-kbd {
  font-size: var(--t-xs);
  opacity: 0.5;
  letter-spacing: 0.08em;
  text-transform: none;
  font-family: var(--font_main);
}

/* Quick fade-in keeps the menu feeling responsive but on-theme. */
.cm-enter-active,
.cm-leave-active {
  transition: opacity 0.08s ease, transform 0.08s ease;
}
.cm-enter-from,
.cm-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}
</style>
