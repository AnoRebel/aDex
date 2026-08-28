<template>
  <!-- Boot / splash overlay. enableAudio comes from the persisted audio
       setting so the boot tone respects the user's preference set on
       a previous launch (Settings → Audio → Enable). -->
  <AdexBootScreen
    v-if="booting"
    :enable-audio="bootAudioEnabled"
    @complete="onBootComplete"
  />

  <!-- Shutdown splash. Audio toggle is independent of the boot
       splash's toggle — Settings → Audio exposes both as separate
       switches so users who relaunch often can mute the shutdown
       sequence without losing the boot fanfare. -->
  <ShutdownSplash
    v-if="shuttingDown"
    :enable-audio="shutdownAudioEnabled"
    @complete="onShutdownComplete"
  />

  <!-- Main application (hidden while booting) -->
  <div
    v-show="!booting"
    v-motion
    :initial="{ opacity: 0 }"
    :visible-once="{ opacity: 1, transition: { duration: 380, ease: [0.19, 1, 0.22, 1] } }"
    class="adex-app"
    :class="{ 'perf-mode': performanceMode }"
    :data-layout="activeLayout"
  >

    <!-- Grid background pattern -->
    <div class="adex-background" :class="{ solidBackground: solidBg }" />

    <!-- Scanlines overlay (theme-controlled strength via --scanline-strength) -->
    <div class="adex-scanlines" aria-hidden="true" />

    <!-- TOP NAVIGATION BAR -->
    <nav class="adex-topbar" role="navigation" aria-label="Section navigation">
      <div
        class="topbar-section"
        :class="{ active: regions.left }"
        title="Show or hide the left panel column"
        @click="toggleRegion('left')"
      >PANEL</div>
      <div
        class="topbar-section"
        :class="{ active: regions.bottom }"
        title="Show or hide the file manager and keyboard"
        @click="toggleRegion('bottom')"
      >SYSTEM</div>
      <div
        class="topbar-section"
        :class="{ active: regions.terminal }"
        title="Show or hide the terminal"
        @click="toggleRegion('terminal')"
      >TERMINAL</div>

      <div class="topbar-spacer" />

      <!-- Shell tab labels in top bar. Double-click to rename — persists
           across launches via useStorage('adex-tab-names'). -->
      <div
        v-for="(tab, idx) in shellTabs"
        :key="tab.id"
        class="topbar-section"
        :class="{ active: activeTabIndex === idx, 'topbar-section-dead': tab.dead }"
        :title="tab.dead ? 'Click to restart this shell' : 'Double-click to rename'"
        @click="switchTab(idx)"
        @dblclick.stop="startTabRename(idx)"
      >
        <input
          v-if="renamingTabIndex === idx"
          v-model="renamingDraft"
          class="topbar-rename-input"
          autofocus
          spellcheck="false"
          @click.stop
          @keydown.enter.prevent="commitTabRename"
          @keydown.escape.prevent="cancelTabRename"
          @blur="commitTabRename"
        />
        <span v-else>{{ tab.title || tabDisplayTitle(idx) }}</span>
      </div>

      <!-- Empty tab placeholders. Click to spawn a new shell so the
           visible "EMPTY" cells are real affordances — without an
           @click these still rendered with cursor: pointer (from
           .topbar-section), which the user reported as a phantom. -->
      <div
        v-for="n in (MAX_TABS - shellTabs.length)"
        :key="'empty-' + n"
        class="topbar-section topbar-section-empty"
        :title="'New shell'"
        @click="addTab"
      >EMPTY</div>

      <div class="topbar-spacer" />

      <div
        class="topbar-section"
        :class="{ active: regions.right }"
        title="Show or hide the right panel column"
        @click="toggleRegion('right')"
      >PANEL</div>
      <div
        class="topbar-section"
        :class="{ active: regions.network }"
        title="Show or hide the network panels"
        @click="toggleRegion('network')"
      >NETWORK</div>

      <!-- Window controls — pinned right -->
      <div class="topbar-window-controls" role="group" aria-label="Window controls">
        <button
          type="button"
          class="topbar-wctl"
          aria-label="Open settings (Ctrl+,)"
          title="Settings (Ctrl+,)"
          @click="showSettings = !showSettings"
        >&#x2699;</button>
        <button
          type="button"
          class="topbar-wctl"
          aria-label="Toggle fullscreen"
          title="Toggle fullscreen"
          @click="onToggleMaximize"
        >&#x25A1;</button>
        <button
          type="button"
          class="topbar-wctl topbar-wctl-close"
          aria-label="Quit aDex-UI"
          title="Quit"
          @click="onRequestQuit"
        >&#x00D7;</button>
      </div>
    </nav>

    <!-- MAIN THREE-COLUMN AREA -->
    <div class="adex-main">

      <!-- LEFT COLUMN -->
      <aside v-show="regions.left" class="mod-column left">
        <!-- A custom layout renders this region from its definition; with no
             custom layout active the built-in composition below is used, so
             the shipped presets behave exactly as before. -->
        <template v-if="customPanelsFor('left')">
          <ErrorBoundary
            v-for="panel in customPanelsFor('left')"
            :key="'left-' + panel"
            :label="panel.toUpperCase()"
          >
            <component :is="panelComponent(panel)" />
          </ErrorBoundary>
        </template>
        <template v-else>
          <!-- Rendered from module state so Settings → Modules controls both
               which panels appear and their order. AdexHardware is included
               again but hidden by default, since it is now switchable rather
               than commented out. -->
          <ErrorBoundary
            v-for="id in modulesEngine.visibleIn('left')"
            :key="'left-' + id"
            :label="moduleLabel(id)"
          >
            <component
              :is="panelComponent(id)"
              v-bind="id === 'clock' ? { use24Hour: use24HourClock } : {}"
            />
          </ErrorBoundary>
        </template>
      </aside>

      <!-- CENTER: MAIN SHELL -->
      <section v-show="regions.terminal" class="adex-center">
        <div class="main-shell">
          <!-- Tab strip inside the shell -->
          <div class="shell-tabs">
            <div
              v-for="(tab, idx) in shellTabs"
              :key="tab.id"
              class="shell-tab"
              :class="{ active: activeTabIndex === idx, 'shell-tab-dead': tab.dead }"
              :title="tab.dead ? 'Click to restart this shell' : tab.title"
              @click="switchTab(idx)"
            >
              <span>{{ tab.title || ('SHELL ' + (idx + 1)) }}{{ tab.dead ? ' [DEAD]' : '' }}</span>
            </div>
            <!-- Empty tab slots -->
            <div
              v-for="n in (MAX_TABS - shellTabs.length)"
              :key="'empty-tab-' + n"
              class="shell-tab"
              style="opacity: 0.3;"
              @click="addTab"
            >
              <span>EMPTY</span>
            </div>
          </div>

          <!-- Terminal emulator surface (one per tab, only active is visible) -->
          <div class="shell-terminal">
            <!-- One boundary PER TAB, so a crash in one terminal cannot take
                 down the others (or the app). auto-retry remounts once, which
                 recovers transient failures such as a lost WebGL context. -->
            <ErrorBoundary
              v-for="(tab, idx) in shellTabs"
              v-show="activeTabIndex === idx"
              :key="'eb-' + tab.id"
              label="TERMINAL"
              auto-retry
            >
              <AdexTerminal
                :ref="(el: any) => terminalRefs[tab.id] = el"
                :session-id="tab.sessionId"
                :active="activeTabIndex === idx"
                @exited="onTerminalExited"
              />
            </ErrorBoundary>
            <!-- Recovery CTA: shown when boot couldn't spawn a shell.
                 Surfaces the last error so the user can see exactly
                 why createSession failed — blank-button-with-no-feedback
                 was leaving every previous attempt as guesswork. -->
            <div
              v-if="shellTabs.length === 0"
              class="shell-empty-cta-wrap"
            >
              <button
                type="button"
                class="shell-empty-cta"
                @click="addTab"
              >
                <span class="shell-empty-cta-symbol">▸</span>
                <span>NEW SHELL</span>
              </button>
              <p
                v-if="lastShellError"
                class="shell-empty-error"
              >ERROR: {{ lastShellError }}</p>
            </div>
          </div>

          <!-- Status bar -->
          <div class="shell-status">
            <span>{{ shellStatusLeft }}</span>
            <span>{{ shellStatusRight }}</span>
          </div>
        </div>
      </section>

      <!-- RIGHT COLUMN -->
      <aside v-show="regions.right" class="mod-column right">
        <template v-if="customPanelsFor('right')">
          <ErrorBoundary
            v-for="panel in customPanelsFor('right')"
            :key="'right-' + panel"
            :label="panel.toUpperCase()"
          >
            <component :is="panelComponent(panel)" />
          </ErrorBoundary>
        </template>
        <template v-else>
          <div
            v-for="id in modulesEngine.visibleIn('right')"
            v-show="id === 'globe' || regions.network"
            :key="'right-' + id"
            class="region-wrap"
          >
            <ErrorBoundary :label="moduleLabel(id)">
              <component :is="panelComponent(id)" />
            </ErrorBoundary>
          </div>
        </template>
      </aside>
    </div>

    <!-- BOTTOM SECTION: FILESYSTEM + KEYBOARD -->
    <div v-show="regions.bottom" class="adex-bottom">
      <AdexFilesystem @navigate="onFsNavigate" @open="onFsOpen" />
      <!-- AdexKeyboard sends keystrokes directly via terminalStore.sendInput;
           a stale `@key="onVirtualKey"` listener was removed because the
           component never emits `key`. The @key prop dropped here on
           purpose to avoid implying a double-injection contract. -->
      <ErrorBoundary label="KEYBOARD"><AdexKeyboard /></ErrorBoundary>
    </div>

    <!-- Session lock. Mounted at page level so it covers the whole shell,
         including modals. -->
    <LockScreen ref="lockScreenRef" />

    <!-- Settings modal (toggled via keyboard shortcut or menu) -->
    <AdexSettingsModal
      v-model="showSettings"
      @settings-changed="onSettingsChanged"
      @show-update-modal="onShowUpdateModal"
    />

    <!-- Update-available notification. Mounted at page level so it
         survives Settings closing and stays above the rest of the UI.
         Driven by `showUpdateModal` which flips either on autocheck
         or via the Settings → Updates → View Update button. -->
    <UpdateAvailableModal v-model="showUpdateModal" />

    <!-- Quit confirmation -->
    <div
      v-if="showQuitConfirm"
      v-motion
      :initial="{ opacity: 0 }"
      :enter="{ opacity: 1, transition: { duration: 150, ease: 'easeOut' } }"
      :leave="{ opacity: 0, transition: { duration: 100 } }"
      class="adex-quit-confirm"
      role="dialog"
      aria-labelledby="adex-quit-confirm-title"
      aria-modal="true"
    >
      <div
        v-motion
        :initial="{ opacity: 0, scale: 0.92, y: 8 }"
        :enter="{
          opacity: 1, scale: 1, y: 0,
          transition: { type: 'spring', stiffness: 320, damping: 24, delay: 30 }
        }"
        class="adex-quit-confirm-card mod-panel"
      >
        <h3 id="adex-quit-confirm-title" class="adex-quit-confirm-title">QUIT aDex-UI?</h3>
        <p class="adex-quit-confirm-body">All shell sessions will be terminated.</p>
        <div class="adex-quit-confirm-actions">
          <button
            type="button"
            class="adex-quit-btn"
            @click="showQuitConfirm = false"
          >CANCEL</button>
          <button
            type="button"
            class="adex-quit-btn adex-quit-btn-primary"
            data-test="quit-confirm"
            @click="onConfirmQuit"
          >QUIT</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { onKeyStroke, useStorage, useNow, useDateFormat, useIntervalFn, useTimeoutFn } from '@vueuse/core'
// Explicit component imports for the splashes/modals. Nuxt's auto-import
// would prefix them by their directory (`UiShutdownSplash`,
// `UiAdexSettingsModal`, etc.) which silently swallows the templates
// when we write `<ShutdownSplash>` — Vue treats the unknown tag as a
// plain HTML element and the component never mounts. That's the exact
// reason the quit dialog appeared but clicking QUIT did nothing: the
// splash wasn't there to receive the v-if flip. Importing explicitly
// registers the bare name we use in templates.
import ShutdownSplash from '~/components/ui/ShutdownSplash.vue'
import UpdateAvailableModal from '~/components/ui/UpdateAvailableModal.vue'
import { useSystemStore } from '~/stores/system'
import { useTerminalStore } from '~/stores/terminal'
import { useFilesystemStore } from '~/stores/filesystem'
import { useNetworkStore } from '~/stores/network'
import { useAudioStore } from '~/stores/audio'
import { useThemeStore } from '~/stores/theme'
import { useAppStore } from '~/stores/app'
import { WindowRuntime } from '~/lib/wailsjs/runtime'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { useCustomLayouts } from '~/composables/useCustomLayouts'
import { useModules, MODULES } from '~/composables/useModules'
import ErrorBoundary from '~/components/ui/ErrorBoundary.vue'
import LockScreen from '~/components/ui/LockScreen.vue'
import AdexClock from '~/components/adex/AdexClock.vue'
import AdexSysinfo from '~/components/adex/AdexSysinfo.vue'
import AdexHardware from '~/components/adex/AdexHardware.vue'
import AdexCpuInfo from '~/components/adex/AdexCpuInfo.vue'
import AdexRamWatcher from '~/components/adex/AdexRamWatcher.vue'
import AdexToplist from '~/components/adex/AdexToplist.vue'
import AdexFilesystem from '~/components/adex/AdexFilesystem.vue'
import AdexNetstat from '~/components/adex/AdexNetstat.vue'
import AdexGlobe from '~/components/adex/AdexGlobe.vue'
import AdexGeoGlobe from '~/components/adex/AdexGeoGlobe.vue'
import AdexTraffic from '~/components/adex/AdexTraffic.vue'
import AdexKeyboard from '~/components/adex/AdexKeyboard.vue'
import { useAdexTheme } from '~/composables/useAdexTheme'
import { useAdexKeyboard } from '~/composables/useAdexKeyboard'

/** Shape emitted by AdexSettingsModal's `settings-changed`. Declared here so a
 *  future rename in the modal surfaces as a type error rather than a silent
 *  no-op — which is exactly how the previous `settings.theme` bug hid. */
interface SettingsPayload {
  shell?: { path?: string; args?: string; workingDirectory?: string }
  display?: {
    theme?: string
    keyboardLayout?: string
    terminalFontSize?: number
    fontFamily?: string
    uiFontFamily?: string
    layout?: string
  }
  audio?: {
    enabled?: boolean
    volume?: number
    soundpack?: string
    muteInBackground?: boolean
    boot?: boolean
    shutdown?: boolean
    categoryKeyboard?: boolean
    categoryDestructive?: boolean
    categoryInterface?: boolean
  }
}
import { useUpdateChecker } from '~/composables/useUpdateChecker'
import { useSettingsPersistence } from '~/composables/useSettingsPersistence'

// Stores
const systemStore = useSystemStore()
const terminalStore = useTerminalStore()
const filesystemStore = useFilesystemStore()
const networkStore = useNetworkStore()
const audioStore = useAudioStore()
const themeStore = useThemeStore()
const appStore = useAppStore()

// Constants
const MAX_TABS = 5

// Template refs for terminal instances
const terminalRefs = ref<Record<string, any>>({})

// Template ref to the sidebar toplist component. Used by the global
// Ctrl+Shift+P keybind to flip its expanded Process Manager modal.
const toplistRef = ref<any>(null)

// Persisted custom tab names — keyed by SLOT INDEX (0..MAX_TABS-1)
// rather than session ID because sessions are recreated each launch.
// Empty string means "no custom name, use the default 'MAIN SHELL' /
// 'SHELL N' label". Lives in adex-settings under `tabNames` so it
// persists across launches and works with the existing settings
// import/export.
const persistedTabNames = useStorage<Record<string, string>>(
  'adex-tab-names',
  {},
)
const renamingTabIndex = ref<number | null>(null)
const renamingDraft = ref('')

function defaultTabTitle(idx: number): string {
  return idx === 0 ? 'MAIN SHELL' : `SHELL ${idx + 1}`
}

function tabDisplayTitle(idx: number): string {
  const custom = (persistedTabNames.value?.[String(idx)] ?? '').trim()
  return custom || defaultTabTitle(idx)
}

function startTabRename(idx: number) {
  renamingTabIndex.value = idx
  renamingDraft.value = persistedTabNames.value?.[String(idx)] ?? ''
}

function commitTabRename() {
  if (renamingTabIndex.value === null) return
  const idx = renamingTabIndex.value
  const trimmed = renamingDraft.value.trim()
  // Mutate first so useStorage flushes synchronously.
  if (!persistedTabNames.value) persistedTabNames.value = {}
  if (trimmed) {
    persistedTabNames.value[String(idx)] = trimmed
  } else {
    // Empty input clears the custom name.
    delete persistedTabNames.value[String(idx)]
  }
  // Reflect into the live tab so it shows immediately without waiting
  // for the next render of `tabDisplayTitle`.
  const tab = shellTabs.value[idx]
  if (tab) tab.title = tabDisplayTitle(idx)
  renamingTabIndex.value = null
}

function cancelTabRename() {
  renamingTabIndex.value = null
}

// Persisted audio toggles for the two splash overlays. They live on
// the same `adex-settings.audio` object as the global enabled flag,
// but each splash has its OWN switch:
//
//   - audio.boot     → AdexBootScreen beep-per-step
//   - audio.shutdown → ShutdownSplash cue-per-step
//
// Both default to true (good first impression on first launch). We
// keep them separate from `audio.enabled` (the global mute) so the
// user can, for example, keep system sounds on but silence the
// shutdown sequence — useful if they quit/relaunch frequently.
const adexAudioSettings = useStorage<{
  audio?: {
    enabled?: boolean
    volume?: number
    boot?: boolean
    shutdown?: boolean
    categoryKeyboard?: boolean
    categoryDestructive?: boolean
    categoryInterface?: boolean
  }
}>('adex-settings', {})

const bootAudioEnabled = computed(() =>
  (adexAudioSettings.value?.audio?.enabled ?? true) &&
  (adexAudioSettings.value?.audio?.boot ?? true),
)

const shutdownAudioEnabled = computed(() =>
  (adexAudioSettings.value?.audio?.enabled ?? true) &&
  (adexAudioSettings.value?.audio?.shutdown ?? true),
)

// Sync per-category toggles into useAdexAudio. The composable owns
// its own settings object (kept in a separate `adex.audio.settings`
// storage key for legacy reasons) but reactively reads from THIS
// `adex-settings.audio.category*` set so the Settings panel is the
// single source of truth. Whenever any category toggle changes, we
// forward it into the engine's runtime gate. This watch is set up
// once on first import and self-cleans on component unmount.
watch(
  adexAudioSettings,
  (s) => {
    const audio = useAdexAudio()
    // CRITICAL: sync the master mute flag too. useAdexAudio has its
    // OWN settings.muted that persists to `adex.audio.settings` (a
    // separate localStorage key). If a user toggled audio off in a
    // previous session, that flag stays true forever — every cue's
    // effectiveVolume returns 0 even though `adex-settings.audio.enabled`
    // is true. We treat `adex-settings.audio.enabled` as the source of
    // truth: muted = !enabled. Same for muteInBackground.
    audio.setMuted(!(s?.audio?.enabled ?? true))
    audio.setCategoryEnabled('keyboard', s?.audio?.categoryKeyboard ?? true)
    audio.setCategoryEnabled('destructive', s?.audio?.categoryDestructive ?? true)
    audio.setCategoryEnabled('interface', s?.audio?.categoryInterface ?? true)
    audio.setGlobalVolume(((s?.audio?.volume ?? 50)) / 100)
  },
  { immediate: true, deep: true },
)

// Boot state. `nointro` in Settings → Advanced lets users bypass the
// boot splash entirely — handy when restarting the app frequently
// during development. Reading from useStorage means a fresh `nointro`
// flag picked up on first paint, before the boot screen even mounts.
const advancedBootSettings = useStorage<{
  advanced?: { nointro?: boolean; forceFullscreen?: boolean; frameless?: boolean; allowWindowedMode?: boolean }
  system?: {
    bootAnimation?: boolean
    showGrid?: boolean
    performanceMode?: boolean
    clockFormat?: '12h' | '24h'
  }
  display?: { globeStyle?: 'classic' | 'geo' }
}>('adex-settings', {})

// Same store, read reactively for the System-panel toggles so changing one
// takes effect without a restart.
const systemSettings = advancedBootSettings

/** Settings → System → Clock format. */
const use24HourClock = computed(
  () => (systemSettings.value?.system?.clockFormat ?? '24h') !== '12h',
)

/** Settings → System → Performance mode. Drops the decorative animations
 *  (scanline sweep, grid drift, globe spin) that dominate idle CPU. */
const performanceMode = computed(
  () => systemSettings.value?.system?.performanceMode === true,
)
const booting = ref(
  !(advancedBootSettings.value?.advanced?.nointro ?? false) &&
  (advancedBootSettings.value?.system?.bootAnimation ?? true),
)
// Settings → System → Show grid background. `solidBackground` is the
// "no grid" state, so this is the inverse of the setting.
const solidBg = computed(() => systemSettings.value?.system?.showGrid === false)
const showSettings = ref(false)
const showQuitConfirm = ref(false)

// Update checker. Single page-level handle that owns the modal toggle
// — the Settings → Updates tab fires `show-update-modal` up here when
// the user clicks "View Update" or a CHECK NOW returns a newer
// release. We also auto-trigger after boot complete (gated by the
// composable's interval setting) so first-launch picks up the latest
// release automatically.
const updateChecker = useUpdateChecker()
const showUpdateModal = ref(false)

function onShowUpdateModal() {
  showUpdateModal.value = true
}

// True once the user confirms Quit — flips on the ShutdownSplash, which
// runs its animated sequence then emits `complete` so we can fire the
// real Wails Quit. Separate from showQuitConfirm so cancelling the
// dialog doesn't briefly flash the splash.
const shuttingDown = ref(false)

// Active layout preset.
//
// Priority order (first non-empty wins):
//   1. User override in Settings → Theme → Layout (`adex-settings.display.layout`)
//      — explicit choice; survives theme switches.
//   2. Active theme's bundled `layout` field — eDex themes ship with a
//      preferred layout (e.g. `tron-disrupted` → `disrupted`).
//   3. `'default'` fallback.
//
// We track the user choice in the same `adex-settings` storage that the
// Settings modal owns so a single layout dropdown there is enough to
// drive `data-layout` reactively.
const layoutSettings = useStorage<{ display?: { layout?: string } }>(
  'adex-settings', {},
)

const activeLayout = computed<string>(() => {
  const userPick = (layoutSettings.value?.display?.layout ?? '').trim()
  if (userPick) return userPick
  const t = themeStore as { activeLayout?: string; currentTheme?: { layout?: string } | null }
  return t.activeLayout ?? t.currentTheme?.layout ?? 'default'
})

/* Custom layouts ---------------------------------------------------------
 *
 * Built-in presets are CSS keyed on `data-layout` and are untouched by this.
 * A custom layout is DATA — which panels sit in which region, in what order —
 * so the shell renders those regions from the definition instead of the fixed
 * template. When no custom layout is active, `customPanels` is null and the
 * original markup renders exactly as before. */
const customLayoutEngine = useCustomLayouts()
const modulesEngine = useModules()
const lockScreenRef = ref<{ lock: () => Promise<void>; refresh: () => Promise<void> } | null>(null)

/** Human label for a module id, used for the error-boundary caption. */
/** Settings → Theme → Globe style. 'geo' draws real country outlines; the
 *  default is the stylised dot-grid globe. */
const globeStyle = computed(() => systemSettings.value?.display?.globeStyle ?? 'classic')

/** Resolve a panel id to its component, honouring the globe-style choice. */
function panelComponent(id: string): unknown {
  if (id === 'globe') return globeStyle.value === 'geo' ? AdexGeoGlobe : AdexGlobe
  return PANEL_COMPONENTS[id]
}

function moduleLabel(id: string): string {
  return (MODULES.find(m => m.id === id)?.label ?? id).toUpperCase()
}

/** Panel id (what users write in layouts.json) -> component.
 *  These ids are a compatibility surface: renaming one breaks existing
 *  user-authored layouts. */
const PANEL_COMPONENTS: Record<string, unknown> = {
  clock:      AdexClock,
  sysinfo:    AdexSysinfo,
  hardware:   AdexHardware,
  cpu:        AdexCpuInfo,
  ram:        AdexRamWatcher,
  toplist:    AdexToplist,
  filesystem: AdexFilesystem,
  netstat:    AdexNetstat,
  globe:      AdexGlobe,  // replaced at render time by globeComponent
  traffic:    AdexTraffic,
  keyboard:   AdexKeyboard,
}

const activeCustomLayout = computed(() => customLayoutEngine.find(activeLayout.value))

/** Panels for a region under the active custom layout, or null when the
 *  built-in template should render instead. */
function customPanelsFor(region: 'left' | 'centre' | 'right' | 'bottom'): string[] | null {
  const layout = activeCustomLayout.value
  if (!layout) return null
  return layout.regions[region] ?? []
}

function onBootComplete() {
  booting.value = false
  playSound('system_startup')
  // Second-chance maximize. The Go-side OnDomReady call fires the
  // moment the DOM is mounted, which can race with the boot screen's
  // own viewport setup on slower compositors. Calling it again here
  // (after the boot overlay is removed) gives the WM another shot
  // when the main app's actual layout is laid out.
  try { WindowRuntime.Maximise() } catch { /* non-critical */ }

  // If Settings → Advanced → Force Fullscreen is on, schedule one
  // more maximize attempt 750ms later. Some Wayland compositors
  // ignore the first WindowMaximise call when the window starts in
  // floating mode — the second call after the layout settles is the
  // one that actually sticks. We use useTimeoutFn so the timer
  // self-cleans on unmount.
  // Force Fullscreen means TRUE fullscreen — no window decorations, covering
  // the display. The previous implementation called Maximise(), which only
  // fills the work area and keeps the titlebar, so the setting never did what
  // its name promised.
  const adv = advancedBootSettings.value?.advanced
  if (adv?.forceFullscreen) {
    useTimeoutFn(() => {
      try { WindowRuntime.Fullscreen() } catch { /* non-critical */ }
    }, 750)
  } else if (adv?.allowWindowedMode === false) {
    // Windowed mode disallowed: keep the window maximised rather than letting
    // the WM restore a previous floating geometry. Previously this setting had
    // no consumer at all and did nothing.
    useTimeoutFn(() => {
      try { WindowRuntime.Maximise() } catch { /* non-critical */ }
    }, 750)
  }

  // Frameless is applied live from Settings, but must also be restored on
  // launch so the choice survives a restart.
  if (adv?.frameless) {
    try { WindowRuntime.SetFrameless(true) } catch { /* non-critical */ }
  }

  // Update check — fire-and-forget. The composable honors the user's
  // interval preference (manual / launch / daily / weekly), so this
  // call no-ops if checks aren't due. We delay by 1.5s after boot so
  // the initial paint isn't competing with a fetch + JSON parse on
  // the same frame; long enough to feel snappy, short enough that
  // the user won't see the modal pop in jarringly.
  useTimeoutFn(async () => {
    try {
      await updateChecker.maybeCheckOnLaunch()
      if (updateChecker.hasUpdate.value) {
        showUpdateModal.value = true
      }
    } catch (err) {
      // Update checks are best-effort — log and move on so a
      // GitHub outage doesn't visibly impact app startup.
      console.warn('[updates] launch check failed:', err)
    }
  }, 1500)
}

// Top bar active section (cosmetic)
/* Top-bar region toggles.
 *
 * These labels previously did nothing: setTopSection wrote a ref that only
 * five :class bindings ever read, so clicking them moved a highlight and
 * changed no layout. They now show and hide the region each one names, which
 * is what their eDEX-UI counterparts implied.
 *
 * Persisted so a hidden panel stays hidden across launches. */
const regions = useStorage('adex.regions', {
  left: true,
  right: true,
  bottom: true,
  terminal: true,
  network: true,
})

function toggleRegion(name: 'left' | 'right' | 'bottom' | 'terminal' | 'network') {
  regions.value = { ...regions.value, [name]: !regions.value[name] }
  try { useAdexAudio().playCue('panels') } catch { /* non-fatal */ }
  // Terminals size themselves to their container, so refit after the layout
  // settles or the visible tab keeps its old dimensions.
  void nextTick(() => {
    for (const ref of Object.values(terminalRefs.value)) ref?.fit?.()
  })
}


// Shell tabs
interface ShellTab {
  id: string
  sessionId: string
  title: string
  /** PTY exited (user typed `exit`, shell crashed, or backend Closed
   *  the session). Marks the tab visually so the user can click to
   *  restart in place instead of being stranded. */
  dead?: boolean
}

const shellTabs = ref<ShellTab[]>([])
const activeTabIndex = ref(0)

// Lazy session creation — IMPORTANT contract:
//
// Only the active tab gets a real PTY at boot. EMPTY slots in the topbar
// are visual placeholders that call `addTab()` on click, which is the
// ONLY path that creates a backend terminal session. Do NOT pre-create
// sessions for unused slots — each PTY is a real shell process and on a
// 5-tab layout that would mean 5 shells running idle from boot.
// Surface the LAST createSession error in the empty-state CTA so users
// see why nothing happened instead of just clicking a dead button.
const lastShellError = ref<string | null>(null)

async function addTab() {
  console.log('[addTab] called, current tabs:', shellTabs.value.length)
  if (shellTabs.value.length >= MAX_TABS) {
    console.log('[addTab] hit MAX_TABS cap, returning early')
    return
  }
  const idx = shellTabs.value.length
  const title = tabDisplayTitle(idx)
  lastShellError.value = null

  const tryOnce = async (): Promise<string | null> => {
    try {
      const result = await terminalStore.createSession(title)
      console.log('[addTab] createSession returned:', result)
      return result
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      lastShellError.value = msg
      console.error('[addTab] createSession threw:', msg, err)
      return null
    }
  }

  let sessionId = await tryOnce()
  if (!sessionId) {
    console.log('[addTab] first attempt failed, retrying in 250ms')
    await new Promise((r) => setTimeout(r, 250))
    sessionId = await tryOnce()
  }
  if (!sessionId) {
    if (!lastShellError.value) {
      lastShellError.value = 'createSession returned null (check store error)'
    }
    console.error('[addTab] both attempts failed, lastShellError =', lastShellError.value)
    return
  }

  console.log('[addTab] success, pushing tab with sessionId:', sessionId)
  lastShellError.value = null
  shellTabs.value.push({
    id: `tab-${Date.now()}`,
    sessionId,
    title,
  })
  activeTabIndex.value = shellTabs.value.length - 1
  terminalStore.setActiveSession(sessionId)
}

function switchTab(idx: number) {
  if (idx < 0 || idx >= shellTabs.value.length) return
  const tab = shellTabs.value[idx]
  // Clicking a DEAD tab restarts it: spawn a fresh PTY into the same
  // slot so the user's tab order + persisted name are preserved.
  if (tab.dead) {
    void restartTab(idx)
    return
  }
  // 'click' cue on tab switch — routine interface feedback.
  if (idx !== activeTabIndex.value) {
    try { useAdexAudio().playCue('click') } catch { /* non-fatal */ }
  }
  activeTabIndex.value = idx
  terminalStore.setActiveSession(tab.sessionId)
  nextTick(() => {
    if (tab && terminalRefs.value[tab.id]) {
      terminalRefs.value[tab.id].focus?.()
    }
  })
}

// Backend's `terminal.exited.<id>` event bubbles up here via the
// scoped EventsOn listener inside AdexTerminal. Mark the tab so the
// strip shows it as DEAD; the user can click to restart.
function onTerminalExited(sessionId: string) {
  const tab = shellTabs.value.find((t) => t.sessionId === sessionId)
  if (!tab) return
  tab.dead = true
}

async function restartTab(idx: number) {
  const tab = shellTabs.value[idx]
  if (!tab) return
  // Free the dead session id from the store before requesting a new
  // one — the backend may still have a stale entry for it.
  try { await terminalStore.closeSession(tab.sessionId) } catch { /* already gone */ }
  delete terminalRefs.value[tab.id]
  const newSessionId = await terminalStore.createSession(tab.title)
  if (!newSessionId) return
  // Replace the tab object so :key changes, forcing AdexTerminal to
  // remount with the new sessionId — otherwise the existing
  // EventsOn listeners keep pointing at the dead session.
  shellTabs.value[idx] = {
    id: `tab-${Date.now()}`,
    sessionId: newSessionId,
    title: tab.title,
  }
  activeTabIndex.value = idx
  terminalStore.setActiveSession(newSessionId)
  await nextTick()
  const restarted = shellTabs.value[idx]
  if (restarted && terminalRefs.value[restarted.id]) {
    terminalRefs.value[restarted.id].focus?.()
  }
}

async function closeCurrentTab() {
  if (shellTabs.value.length <= 1) return
  const tab = shellTabs.value[activeTabIndex.value]
  if (!tab) return
  await terminalStore.closeSession(tab.sessionId)
  delete terminalRefs.value[tab.id]
  shellTabs.value.splice(activeTabIndex.value, 1)
  if (activeTabIndex.value >= shellTabs.value.length) {
    activeTabIndex.value = Math.max(0, shellTabs.value.length - 1)
  }
  if (shellTabs.value.length > 0) {
    terminalStore.setActiveSession(shellTabs.value[activeTabIndex.value].sessionId)
  }
}

// Live CWD per session — polled every second via the backend's
// GetTerminalCWD binding (which uses the platform-specific tracker:
// /proc/<pid>/cwd on Linux, lsof on macOS, NtQueryInformationProcess
// on Windows). Without this poll the footer kept showing the
// session's ORIGINAL spawn cwd because `session.workingDirectory`
// never updates on `cd`.
const liveCwds = ref<Record<string, string>>({})
useIntervalFn(async () => {
  const sessionId = terminalStore.activeSession?.id
  if (!sessionId) return
  try {
    const { GetTerminalCWD } = await import('~/lib/wailsjs/coordinator')
    const cwd = await GetTerminalCWD(sessionId)
    if (cwd) liveCwds.value[sessionId] = cwd
  } catch {
    // Backend may not track this session yet — fall back to the
    // spawn cwd in shellStatusLeft.
  }
}, 1000)

// Shell status bar
const shellStatusLeft = computed(() => {
  const tab = shellTabs.value[activeTabIndex.value]
  if (!tab) return ''
  const session = terminalStore.activeSession
  if (!session) return 'No active session'
  const cwd = liveCwds.value[session.id] || session.workingDirectory || '~'
  return `${session.shell ?? '/bin/bash'}  --  ${cwd}`
})

// VueUse useNow ticks every second so the status-bar clock updates
// live instead of only when the user switches tabs (the previous
// `new Date()` inside computed only re-evaluated when other deps
// changed — which is why the time looked frozen).
const _statusNow = useNow({ interval: 1000 })
const _statusTime = useDateFormat(_statusNow, 'HH:mm:ss')
const shellStatusRight = computed(() => {
  const session = terminalStore.activeSession
  const dims = session ? `${session.columns ?? 80}x${session.rows ?? 24}` : ''
  return `${dims}  ${_statusTime.value}`
})

// Filesystem events
function onFsNavigate(path: string) {
  filesystemStore.fetchDirectory(path)
}

function onFsOpen(item: { name: string; path: string; type?: string }) {
  if (item.type === 'directory') {
    filesystemStore.fetchDirectory(item.path)
    const session = terminalStore.activeSession
    if (session) {
      terminalStore.sendInput(`cd "${item.path}"\n`, session.id)
    }
  }
}

// (Removed onVirtualKey — AdexKeyboard sends keystrokes directly via
//  terminalStore.sendInput; the listener was dead code.)

// Settings
//
// Apply everything the user just saved. This previously read
// `settings.theme`, but the payload nests it as `settings.display.theme` —
// so the condition was never true and NOTHING was applied on Save. Theme and
// keyboard only appeared to work because the settings modal applies those two
// live on selection; every other setting was silently discarded.
//
// Each group is applied independently and defensively: one group failing must
// not stop the rest from being applied.
function onSettingsChanged(settings: SettingsPayload) {
  const applyGroup = (label: string, fn: () => void) => {
    try {
      fn()
    } catch (err) {
      console.error(`[settings] failed to apply ${label}:`, err)
    }
  }

  const display = settings?.display
  if (display) {
    applyGroup('display.theme', () => {
      // Live-applied by the modal on selection; re-assert on save so a value
      // restored by Cancel-then-Save still lands.
      if (display.theme) void useAdexTheme().setTheme(display.theme)
    })
    applyGroup('display.keyboardLayout', () => {
      if (display.keyboardLayout) void useAdexKeyboard().setLayout(display.keyboardLayout)
    })
    applyGroup('display.fonts', () => {
      const root = document.documentElement
      // Interface font: empty means "follow the theme", so an explicit clear
      // must remove the override rather than being skipped.
      if (typeof display.uiFontFamily === 'string') {
        if (display.uiFontFamily.trim()) {
          root.style.setProperty('--font_main', display.uiFontFamily)
          root.style.setProperty('--font_main_light', display.uiFontFamily)
        } else {
          root.style.removeProperty('--font_main')
          root.style.removeProperty('--font_main_light')
        }
      }
      if (display.fontFamily) {
        // Terminal only. --terminal_font is what AdexTerminal reads when
        // building xterm's font stack; the interface font is a separate
        // setting above, so this must not touch --font_main.
        root.style.setProperty('--terminal_font', display.fontFamily)
      }
      if (display.terminalFontSize) {
        root.style.setProperty('--terminal_font_size', `${display.terminalFontSize}px`)
      }
      // Push font changes into every live terminal. xterm holds its font in
      // instance options, so a CSS variable alone never reaches it — the size
      // was additionally hardcoded at construction.
      for (const ref of Object.values(terminalRefs.value)) {
        ref?.applyFont?.(display.fontFamily, display.terminalFontSize)
      }
    })
    // `display.layout` needs no explicit push: pages/index.vue derives
    // `activeLayout` reactively from the persisted settings store.
  }

  const audio = settings?.audio
  if (audio) {
    const engine = useAdexAudio()
    applyGroup('audio.master', () => {
      engine.setMuted(!audio.enabled)
      engine.setGlobalVolume(Math.max(0, Math.min(1, (audio.volume ?? 50) / 100)))
      if (audio.soundpack) engine.setPack(audio.soundpack as 'adex' | 'synth')
      engine.setBackgroundMuted(!!audio.muteInBackground)
    })
    applyGroup('audio.categories', () => {
      engine.setCategoryEnabled('keyboard', audio.categoryKeyboard !== false)
      engine.setCategoryEnabled('destructive', audio.categoryDestructive !== false)
      engine.setCategoryEnabled('interface', audio.categoryInterface !== false)
      // Boot and shutdown cues are the "system" category.
      engine.setCategoryEnabled('system', audio.boot !== false || audio.shutdown !== false)
    })
  }

  if (settings?.shell) {
    applyGroup('shell', () => {
      const path = settings.shell?.path?.trim()
      if (path) {
        void import('~/lib/wailsjs/coordinator').then(({ SetShellCommand }) =>
          SetShellCommand(path),
        )
      }
    })
  }
}

// Window controls
//
// We toggle TRUE fullscreen (no window decorations) rather than
// Maximize toggle, per Wails v2 JS runtime docs:
//   WindowIsMaximised(): Promise<boolean>     // ← async
//   WindowMaximise(): void
//   WindowUnmaximise(): void
//
// The earlier version called WindowIsMaximised() synchronously and
// got back a Promise object (always truthy), so the toggle always
// took the Unmaximise branch — that's why the button looked broken.
async function onToggleMaximize() {
  // Wails v3 has no `window.runtime` global — the previous implementation
  // probed it and returned early every time, which is why this button did
  // nothing after the v3 port. Window controls now come from the runtime
  // facade.
  // TRUE fullscreen, not maximise. The button is labelled "Toggle fullscreen"
  // and users expect the fullscreen experience; Maximise() only fills the work
  // area and is ignored outright by tiling window managers, which already own
  // window geometry.
  try {
    if (await WindowRuntime.IsFullscreen()) {
      WindowRuntime.UnFullscreen()
    } else {
      WindowRuntime.Fullscreen()
    }
    try { useAdexAudio().playCue('click') } catch { /* non-fatal */ }
  } catch (err) {
    console.warn('[window] fullscreen toggle failed:', err)
  }
}

function onRequestQuit() {
  showQuitConfirm.value = true
  // Handgun-click "chambering" cue on dialog OPEN — primes the user
  // that they're about to confirm a destructive action. The actual
  // gunshot fires inside onConfirmQuit once they pull the trigger.
  try { useAdexAudio().playCue('destructive') } catch { /* non-fatal */ }
}

function onConfirmQuit() {
  // Two-phase teardown:
  //   1) Hide the confirm dialog, show the animated ShutdownSplash.
  //      Foreground monitoring is stopped immediately because the splash
  //      itself is the only thing the user should see updating; we don't
  //      want polling spam contending with the audio cues.
  //   2) The splash emits `complete` after its sequence — that handler
  //      (onShutdownComplete) calls the actual Wails Quit.
  //
  // Failsafe: ALSO arm a 5-second hard-quit timer in parallel. The splash
  // sequence is ~2.4s; if it crashes silently or never mounts (template
  // throw, motion-v wrapping a sibling, Teleport target missing, etc.)
  // we'd otherwise be stuck with no way out. The timer's only job is to
  // make sure the app actually exits even if everything visual fails.
  console.info('[quit] confirmed — starting shutdown splash')
  // Gunshot fires when the user pulls the trigger on the confirm.
  // The dialog-OPEN played the handgun chambering cue; this is the bang.
  try { useAdexAudio().playCue('gunshot') } catch { /* non-fatal */ }
  showQuitConfirm.value = false
  // Force-flush any setting changed in the last FLUSH_DEBOUNCE_MS so it
  // isn't lost to the pending debounce when the process exits. Fire and
  // forget: the ~2.4s shutdown splash easily outlasts the atomic write,
  // and doQuit() only runs from onShutdownComplete / the 5s failsafe.
  try { useSettingsPersistence().flushNow() } catch { /* non-fatal */ }
  try { systemStore.stopMonitoring?.() } catch { /* non-fatal */ }
  shuttingDown.value = true
  // Hard fallback. doQuit is idempotent (Wails ignores duplicate calls),
  // so if onShutdownComplete fires first, this timer's eventual call is
  // a no-op against an already-quitting process.
  setTimeout(() => {
    if (shuttingDown.value) {
      console.warn('[quit] splash sequence did not complete in 5s — forcing quit')
      doQuit()
    }
  }, 5000)
}

function onShutdownComplete() {
  console.info('[quit] splash complete — firing Wails Quit')
  doQuit()
}

// Sentinel so doQuit can only fire ONCE per session. Without this guard
// every quit attempt called runtime.Quit twice in fast succession, and
// on Linux/GTK the second call hit gtk_main_quit AFTER the first call
// had already torn down the main loop — producing the
//   "(gtk_main_quit: assertion 'main_loops != NULL' failed)"
// crash visible in the dev terminal. Wails v2's runtime.Quit is NOT
// documented as idempotent and the GTK backend treats double-call as a
// hard error. So we call exactly one path, exactly once.
let quitFired = false
function doQuit() {
  if (quitFired) return
  quitFired = true
  console.info('[quit] firing doQuit')
  // Ask the Wails runtime to quit. The Go side still controls teardown
  // order from here: ShouldQuit, then each registered service's
  // ServiceShutdown in reverse registration order, then process exit.
  //
  // Under v2 this first probed the bound `window.go.main.App.QuitApp`
  // method. Wails v3 has no `window.go` IPC global — bound methods are
  // reached through generated bindings — so the runtime call is now the
  // single path.
  try { WindowRuntime.Quit() } catch { /* runtime tearing down — ignore */ }
}

// Audio helper
function playSound(_eventId: string) {
  try {
    const store = audioStore as any
    if (store.settings?.enabled || store.isEnabled) {
      store.playEvent?.(_eventId) ?? store.playSound?.(_eventId)
    }
  } catch {
    // Non-critical
  }
}

// Keyboard shortcuts — driven by VueUse's onKeyStroke. Each handler
// double-checks the modifier state because onKeyStroke fires for the key
// regardless of modifiers; we want the chord to match exactly.
//
// Note: Ctrl+1..5 alone would hijack readline word-jumps inside the
// terminal, so all tab shortcuts require Ctrl+Shift. We also allow the
// uppercase variant since some layouts map Ctrl+Shift+t differently.
const isPlainCtrlShift = (e: KeyboardEvent) =>
  e.ctrlKey && e.shiftKey && !e.altKey && !e.metaKey

onKeyStroke(['1', '2', '3', '4', '5'], (e) => {
  if (!isPlainCtrlShift(e)) return
  const idx = parseInt(e.key, 10) - 1
  if (idx < 0 || idx >= shellTabs.value.length) return
  e.preventDefault()
  switchTab(idx)
})

onKeyStroke(['t', 'T'], (e) => {
  if (!isPlainCtrlShift(e)) return
  e.preventDefault()
  addTab()
})

onKeyStroke(['w', 'W'], (e) => {
  if (!isPlainCtrlShift(e)) return
  if (shellTabs.value.length <= 1) return
  e.preventDefault()
  closeCurrentTab()
})

onKeyStroke(',', (e) => {
  if (!e.ctrlKey || e.shiftKey || e.altKey || e.metaKey) return
  e.preventDefault()
  showSettings.value = !showSettings.value
})

// Ctrl+Shift+L — lock the session. Uses the same chord shape as the tab
// shortcuts so it cannot collide with a readline binding in the terminal.
onKeyStroke(['l', 'L'], (e) => {
  if (!isPlainCtrlShift(e)) return
  e.preventDefault()
  void lockScreenRef.value?.lock()
})

// Ctrl+Shift+Q — quit aDex-UI through the same confirm dialog the
// titlebar × button uses, so the user always passes through the
// "are you sure?" gate before backend Shutdown runs.
onKeyStroke(['q', 'Q'], (e) => {
  if (!isPlainCtrlShift(e)) return
  e.preventDefault()
  onRequestQuit()
})

// Ctrl+Shift+P — toggle the expanded Process Manager modal. The modal
// itself lives inside AdexToplist (so it shares the same data + action
// handlers), and AdexToplist exposes toggleManager() via defineExpose.
// We resolve through the template ref instead of duplicating state up
// here, which would have meant maintaining two copies of "is the modal
// open" that could disagree after a route change.
onKeyStroke(['p', 'P'], (e) => {
  if (!isPlainCtrlShift(e)) return
  e.preventDefault()
  toplistRef.value?.toggleManager?.()
})

// Lifecycle
onMounted(async () => {
  // Initialise all stores in parallel (errors are non-fatal)
  const safeInit = (fn: () => Promise<any>) =>
    fn().catch((e: unknown) => console.warn('Store init:', e))

  // Rehydrate the persisted shell-path override into the backend BEFORE
  // terminalStore.initialize() spawns the first PTY — otherwise the user's
  // saved zsh / pwsh / custom shell preference is ignored on every boot.
  try {
    const stored = useStorage<{ shell?: { path?: string } }>('adex-settings', {})
    const path = stored.value?.shell?.path?.trim() ?? ''
    const { SetShellCommand } = await import('~/lib/wailsjs/coordinator')
    await SetShellCommand(path)
  } catch (err) {
    console.warn('[boot] failed to restore shell override:', err)
  }

  await Promise.all([
    safeInit(() => appStore.initialize()),
    safeInit(() => systemStore.initialize()),
    safeInit(() => themeStore.initialize()),
    // Load user-defined layouts so they are selectable and can be applied.
    safeInit(() => customLayoutEngine.load()),
    safeInit(() => terminalStore.initialize()),
    safeInit(async () => {
      // Resolve the user's "Open in" preference (Settings → System) to a
      // real path before seeding the file manager. Falls back to '/' on
      // any error so we never block boot.
      let initialPath = '/'
      try {
        const { useStartupCwd } = await import('~/composables/useStartupCwd')
        initialPath = await useStartupCwd().resolve()
      } catch {
        // resolve has its own fallbacks; this catch is just a safety net.
      }
      try {
        if (typeof filesystemStore.initialize === 'function') {
          await filesystemStore.initialize()
        }
        await filesystemStore.fetchDirectory(initialPath)
      } catch {
        // non-fatal
      }
    }),
    safeInit(async () => {
      try {
        if (typeof networkStore.startMonitoring === 'function') {
          await networkStore.startMonitoring()
        } else if (typeof (networkStore as any).initialize === 'function') {
          await (networkStore as any).initialize()
        }
      } catch {
        // Network monitoring is optional
      }
    }),
  ])

  // Audio initialisation (non-blocking, defensive)
  try {
    const store = audioStore as any
    if (typeof store.initialize === 'function') {
      store.initialize()
    }
  } catch {
    // non-critical
  }

  // Seed the first shell tab from whatever session the terminal store created
  await nextTick()
  const sessions = terminalStore.allSessions
  if (sessions.length > 0) {
    shellTabs.value = sessions.map((s, idx) => ({
      id: `tab-${s.id}`,
      sessionId: s.id,
      // Honor any custom name the user set on this slot in a previous
      // session; falls back to the default MAIN SHELL / SHELL N label.
      title: tabDisplayTitle(idx),
    }))
    activeTabIndex.value = 0
    terminalStore.setActiveSession(sessions[0].id)
  } else {
    await addTab()
  }

  // Focus the active terminal so the user can type immediately on
  // launch — the boot screen captures focus during its progress
  // animation, and without an explicit focus call here the user has
  // to click the terminal first. nextTick ensures the AdexTerminal
  // ref is mounted before we reach into it.
  await nextTick()
  const activeTab = shellTabs.value[activeTabIndex.value]
  if (activeTab && terminalRefs.value[activeTab.id]) {
    terminalRefs.value[activeTab.id].focus?.()
  }

  // Keyboard shortcuts are registered at script-setup top level via
  // VueUse's onKeyStroke() — they self-cleanup on component unmount, no
  // manual addEventListener/removeEventListener needed.
})

onUnmounted(() => {
  systemStore.stopMonitoring()
})
</script>

<style scoped>
.adex-main {
  min-height: 0;
}

/* Empty-state CTA shown when no shell session could be created at
 * boot. Centered click target with the eDEX bracket aesthetic. */
.shell-empty-cta-wrap {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5em;
}
.shell-empty-cta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1ch;
  padding: 0.8em 1.6em;
  background: transparent;
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  font: inherit;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease;
}
.shell-empty-cta:hover,
.shell-empty-cta:focus-visible {
  background: var(--surface-2, rgba(255, 255, 255, 0.06));
  color: var(--accent-strong, var(--accent));
  outline: none;
}
.shell-empty-cta-symbol {
  font-size: 1.4em;
  opacity: 0.7;
}
.shell-empty-error {
  max-width: 60ch;
  margin: 0;
  padding: 0.4em 0.8em;
  font-family: var(--font_main);
  font-size: 0.85em;
  color: #ff8a8a;
  background: rgba(255, 60, 60, 0.08);
  border: var(--border_width) solid rgba(255, 107, 107, 0.4);
  text-transform: none;
  letter-spacing: normal;
  text-align: center;
  word-break: break-word;
}

/* eDEX is keyboard-first and meant to be glanced at — every mod must be
 * visible at once, no scrollbars hiding the clock or top processes.
 * We give each mod a proportional share of the column (flex-basis 0 +
 * grow) so they collectively consume exactly the available height
 * regardless of viewport size. The "tall" mods (Toplist on left,
 * Traffic on right) get a larger grow so they keep their identity
 * while compressing the small mods (clock, sysinfo) just enough to
 * fit. Inner content owns its own scroll/scale (toplist rows, traffic
 * canvas) — never the column. */
.mod-column {
  overflow: hidden;
  min-height: 0;
  justify-content: stretch;
  gap: 0.3vh;
}

/* Layout strategy for the side columns:
 *
 *   - Small "glance" mods (clock, sysinfo, hardware, cpu, ram) take
 *     their natural content size — they don't grow OR shrink. With
 *     `flex: 0 0 auto` their inherent vh-based heights stay stable
 *     when the column resizes.
 *   - Big interactive mods (Toplist on left, Globe + Traffic on right)
 *     ARE the column's elastic content — they own all remaining space
 *     so the panel has room for many rows / a tall canvas.
 *
 * Without this, every mod was `flex: 1 1 0` and the small mods stole
 * height proportionally, leaving Toplist with ~16px of usable area. */
.mod-column > * {
  /* `flex: 0 0 auto` meant the small mods could not shrink at all, so with
     every panel enabled their natural heights exceed the column and the last
     one (Processes) is pushed out of sight. Allow shrinking: they keep their
     natural size while there is room and compress only when there is not. */
  flex: 0 1 auto;
  min-height: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
}
/* The tall mods still claim the leftover space, but their floors are now a
   fraction of the column too, so a fixed vh cannot exceed what is available
   on a short window. */
/* Text-only panels must never be compressed: they have no canvas to scale
   down, so `flex-shrink` just hides their last rows behind the panel's
   `overflow: hidden` — SYSTEM INFO lost its RAM line and HARDWARE INSPECTOR
   its CHASSIS line exactly this way. Pin them to their content height and
   let the panels that CAN scale (the charts and the process list) absorb
   the pressure instead.

   The left column renders panels as DIRECT children (no .region-wrap), so
   `.mod-column > *` applies to the panel itself and this override needs the
   same shape to beat it. */
.mod-column > :deep(.mod-sysinfo),
.mod-column > :deep(.mod-hardware),
.mod-column > :deep(.mod-clock-panel) {
  flex: 0 0 auto;
}

.mod-column :deep(.mod-toplist) { flex: 1 1 auto; min-height: min(22vh, 30%); }
.mod-column :deep(.mod-traffic) { flex: 1 1 auto; min-height: min(16vh, 25%); }
/* The globe's wrapper grows to claim the column's leftover height, so the
   panel inside must grow with it — left at `flex: 0 1 auto` the panel kept
   its natural height and the surplus showed up as a dead gap between the
   world view and the traffic chart. `1 1 auto` makes the panel fill the
   wrapper it was already given. */
.mod-column :deep(.mod-globe) { flex: 1 1 auto; min-height: min(14vh, 22%); }

/* The countries globe sizes itself: a square canvas plus four readout rows.
   Imposing a vh cap here fought that intrinsic height and clipped whichever
   of the two lost, so let the panel take the height its content needs. */
.mod-column :deep(.mod-geo-globe) {
  flex: 0 0 auto;
}

/* Region wrappers (added so v-show has a real element to act on) sit between
   the column and its panels. The `.mod-column > *` rule above gives them
   `flex: 0 1 auto`, which stopped the panel INSIDE from growing — the right
   column left ~23% of its height unused and the traffic graph looked squashed.
   These must be declared here, in the scoped block: a plain `.region-wrap`
   rule in main.css loses to the data-v attribute selector above. */
.mod-column > .region-wrap {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
/* A wrapper around a panel that should not grow must not grow either.
   ErrorBoundary sits between the wrapper and the panel, so match the panel
   at any depth rather than as a direct child. Without this the countries
   globe's wrapper stretched while the fixed-height panel inside it did not,
   leaving a large empty gap above the traffic chart. */
.mod-column > .region-wrap:has(> .mod-netstat) {
  flex: 0 1 auto;
}

/* The countries globe has a fixed intrinsic height (a square canvas plus
   four readout rows), so its wrapper must not stretch — the extra height
   would just be an empty gap above the traffic chart.

   This has to live in the SCOPED block, immediately after the
   `.mod-column > .region-wrap` rule it overrides: that rule carries the
   page's data-v attribute (specificity 0,3,1), so the same rule written
   unscoped in main.css loses the cascade and does nothing. `:deep()` is
   what lets :has() match .mod-geo-globe, which carries the child
   component's scope id rather than this page's. */
.mod-column > .region-wrap:has(:deep(.mod-geo-globe)) {
  flex: 0 0 auto;
}

/* Mod inner contents must respect their parent's height so the canvas
 * elements (cpu chart, ram bar, traffic graph) never overflow. */
.mod-column :deep(.mod-panel) {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.mod-column :deep(.mod-panel canvas) {
  max-height: 100%;
  max-width: 100%;
}

/* Ensure the largest mods (Toplist on the left, Traffic on the right)
 * are the ones that scroll internally when the column runs out of room.
 * Without this they expand to fit their content and push siblings off
 * the bottom of the column. */
/* The mod-toplist / mod-traffic flex sizing is now defined above
 * (see `Layout strategy for the side columns`). The previous
 * duplicate rule here used `flex: 1 1 0` which competed with the
 * `flex: 0 0 auto` baseline and let small mods steal height. */

/* Compact mode for laptop-class screens (sub-1080 height). The eDEX
 * baseline assumes 1080p; under that we shrink type + padding so the
 * mods still fit without truncating content rows mid-letter. */
@media (max-height: 900px) {
  .mod-column {
    gap: 0.2vh;
  }
  .mod-column :deep(.section-title) {
    height: 1.8vh;
    font-size: 1vh;
    margin-bottom: 0.2vh;
  }
  .mod-column :deep(.mod-panel) {
    padding: 0.4vh 0.5vw;
  }
}

/* .adex-app fade-in is handled by v-motion (see template :visible-once) */

/* Window control buttons in the top bar */
.topbar-window-controls {
  display: flex;
  align-items: stretch;
  margin-left: 0.5vw;
  border-left: var(--rule-width) solid var(--rule);
}

.topbar-wctl {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.5vh;
  padding: 0 0.8vw;
  font-size: 1.6vh;
  line-height: 1;
  /* The buttons used to draw at full --accent against an --accent-tinted
   * topbar background, which made them blend in on every theme. We pin
   * a strong accent foreground here so the cog/maximize/close glyphs are
   * always legible — and keep the hover/focus surface bright. */
  color: var(--accent-strong, var(--accent));
  background: transparent;
  border: 0;
  border-right: var(--rule-width) solid var(--rule);
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.topbar-wctl:last-child {
  border-right: 0;
}

.topbar-wctl:hover,
.topbar-wctl:focus-visible {
  background: var(--surface-3, rgba(255, 255, 255, 0.08));
  color: var(--accent);
  outline: none;
}

.topbar-wctl-close:hover,
.topbar-wctl-close:focus-visible {
  background: rgba(255, 60, 60, 0.25);
  color: #ff8a8a;
}

.topbar-wctl-close:hover {
  background: rgba(239, 68, 68, 0.18); /* var(--err) tinted */
  color: var(--err);
}

.topbar-wctl:focus-visible {
  outline: var(--rule-width) solid var(--accent);
  outline-offset: -2px;
}

/* Quit confirm dialog */
.adex-quit-confirm {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
  /* fade and card mount handled by v-motion in the template */
}

.adex-quit-confirm-card {
  --corner-cut: 1.2vh;
  width: min(40vw, 60vh);
  padding: 2vh 2vw;
  background: var(--bg-glass);
  box-shadow: var(--glow-strong);
}

.adex-quit-confirm-title {
  font-size: var(--t-xl);
  letter-spacing: 0.2em;
  margin-bottom: 1vh;
  color: var(--accent);
}

.adex-quit-confirm-body {
  font-size: var(--t-md);
  opacity: 0.75;
  margin-bottom: 2vh;
}

.adex-quit-confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1vw;
}

.adex-quit-btn {
  padding: 0.7vh 2vw;
  background: transparent;
  border: var(--rule-width) solid var(--rule);
  color: var(--accent);
  font-family: inherit;
  font-size: var(--t-md);
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background 0.15s ease;
}

.adex-quit-btn:hover {
  background: var(--surface-2);
}

.adex-quit-btn-primary {
  background: rgba(239, 68, 68, 0.15);
  border-color: var(--err);
  color: var(--err);
}

.adex-quit-btn-primary:hover {
  background: rgba(239, 68, 68, 0.3);
}
</style>
