<template>
  <Transition name="settings-modal">
    <div v-if="modelValue" class="settings-overlay" @click.self="handleOverlayClick">
      <div
        ref="modalRef"
        class="settings-modal"
        :style="modalStyle"
      >
        <!-- Header — doubles as the drag handle. -->
        <div class="settings-header" @pointerdown="startDrag">
          <div class="settings-header-title">
            <span class="settings-header-bracket">[</span>
            SYSTEM CONFIGURATION
            <span class="settings-header-bracket">]</span>
          </div>
          <button class="settings-close-btn" @click="cancel" title="Close settings">
            X
          </button>
        </div>

        <!-- Body: sidebar + content -->
        <div class="settings-body">
          <!-- Category Sidebar -->
          <div class="settings-sidebar">
            <div
              v-for="cat in categories"
              :key="cat.id"
              class="settings-sidebar-item"
              :class="{ active: activeCategory === cat.id }"
              @click="activeCategory = cat.id"
            >
              <span class="sidebar-marker">{{ activeCategory === cat.id ? '>' : ' ' }}</span>
              {{ cat.label }}
            </div>
          </div>

          <!-- Settings Content -->
          <div class="settings-content">
            <!-- SHELL -->
            <div v-if="activeCategory === 'terminal'" class="settings-section">
              <div class="settings-section-title">TERMINAL CONFIGURATION</div>

              <div class="settings-field">
                <label class="settings-label">Shell Path</label>
                <input
                  v-model="localSettings.shell.path"
                  type="text"
                  class="settings-input"
                  :placeholder="shellPlaceholder"
                  @change="onShellPathChange"
                />
                <p class="settings-hint">
                  Leave blank to use your default shell ($SHELL on macOS/Linux,
                  $ComSpec on Windows). New terminals pick up the change; existing
                  tabs keep their current shell.
                </p>
              </div>

              <div class="settings-field">
                <label class="settings-label">Shell Arguments</label>
                <input
                  v-model="localSettings.shell.args"
                  type="text"
                  class="settings-input"
                  placeholder="--login"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Working Directory</label>
                <input
                  v-model="localSettings.shell.workingDirectory"
                  type="text"
                  class="settings-input"
                  placeholder="/"
                />
              </div>
            </div>

            <!-- THEME -->
            <div v-if="activeCategory === 'theme'" class="settings-section">
              <div class="settings-section-title">THEME</div>

              <div class="settings-field">
                <label class="settings-label">Theme ({{ themeIndex.length }} available)</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.display.theme"
                  :items="themeMenuItems"
                  value-key="id"
                  label-key="displayName"
                  :search-input="{ placeholder: 'Search themes...' }"
                  class="settings-select"
                  @update:model-value="onThemeApply"
                />
                <p class="settings-hint">
                  Selecting a theme applies it immediately and persists
                  the choice across launches.
                </p>
              </div>

              <!-- Layout preset override — independent of theme.
                   Each layout maps to a CSS file in assets/css/layouts/.
                   `theme` (the default option here) means "follow the
                   theme's bundled layout choice" so eDex-themed presets
                   like tron-disrupted automatically pick `disrupted`. -->
              <div class="settings-field">
                <label class="settings-label">Layout preset</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.display.layout"
                  :items="layoutItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                  @update:model-value="persistCurrentSettings"
                />
                <p class="settings-hint">
                  <strong>Theme default</strong> follows whichever layout
                  the active theme ships with.<br/>
                  <strong>Default</strong> — classic left mods / terminal / right mods + keyboard.<br/>
                  <strong>Disrupted</strong> — terminal top, file manager + keyboard split bottom.<br/>
                  <strong>Notype</strong> — hide the on-screen keyboard.<br/>
                  <strong>Fulltype</strong> — full-width on-screen keyboard, file manager hidden.<br/>
                  <strong>Terminal focus</strong> — terminal fills the window; side panels, keyboard and file manager hidden.<br/>
                  <strong>Typeleft</strong> — keyboard on the left half.<br/>
                  <strong>Colorfilter</strong> — minimal chrome, single column.
                </p>
                <p class="settings-hint">
                  <strong>Custom layouts</strong> — define your own panel
                  arrangement by creating
                  <code>{{ customLayoutsPath || 'layouts.json' }}</code>.
                  Each entry needs an <code>id</code> and a
                  <code>regions</code> map of
                  <code>left</code>/<code>centre</code>/<code>right</code>/<code>bottom</code>
                  to panel names. Restart to pick up changes.
                  <span v-if="customLayoutEngine.layouts.value.length">
                    <br/>{{ customLayoutEngine.layouts.value.length }} custom
                    layout(s) loaded.
                  </span>
                </p>
              </div>

              <!-- Globe rendering style. Two genuinely different panels,
                   not a cosmetic toggle: the classic globe is the stylised
                   dot grid, the geo one projects real country outlines. -->
              <div class="settings-field">
                <label class="settings-label">Globe style</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.display.globeStyle"
                  :items="globeStyleItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                  @update:model-value="persistCurrentSettings"
                />
                <p class="settings-hint">
                  <strong>Classic</strong> (default) — the stylised rotating
                  dot-grid globe. Cheapest to draw.<br/>
                  <strong>Countries</strong> — an orthographic projection of real
                  country outlines, so the connection marker sits on a
                  recognisable landmass. Adds roughly 105&nbsp;KB of map data,
                  loaded only when this style is selected.
                </p>
                <p class="settings-hint">
                  Countries renders in a background thread, so it costs about
                  the same as Classic and cannot slow the terminal or the
                  charts even when a frame is slow.
                </p>
              </div>

              <!-- Terminal palette. Merged in from what used to be its own
                   COLOR SCHEME panel: two controls did not justify a separate
                   section, and the palette is part of theming. -->
              <div class="settings-section-title" style="margin-top: 1.2rem;">
                TERMINAL PALETTE
              </div>


              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Use theme colors</label>
                <USwitch
                  v-model="useThemeColors"
                  size="md"
                  color="primary"
                  aria-label="Use theme colors"
                />
              </div>
              <p class="settings-hint">
                When ON, the terminal palette derives from the active
                theme. Turn OFF to pick a separate color scheme.
              </p>

              <div v-if="!useThemeColors" class="settings-field">
                <label class="settings-label">Scheme</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="selectedColorScheme"
                  :items="colorSchemeItems"
                  value-key="id"
                  label-key="label"
                  :search-input="{ placeholder: 'Search schemes...' }"
                  class="settings-select"
                  @update:model-value="onColorSchemeApply"
                />
              </div>
            </div>

            <!-- KEYBOARD -->
            <div v-if="activeCategory === 'keyboard'" class="settings-section">
              <div class="settings-section-title">KEYBOARD</div>

              <div class="settings-field">
                <label class="settings-label">Layout ({{ kbIndex.length }} available)</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.display.keyboardLayout"
                  :items="kbMenuItems"
                  value-key="id"
                  label-key="displayName"
                  :search-input="{ placeholder: 'Search layouts...' }"
                  class="settings-select"
                  @update:model-value="onKeyboardLayoutApply"
                />
                <p class="settings-hint">
                  Switches the on-screen keyboard layout. Physical keyboard
                  events are passed through unchanged.
                </p>
              </div>
            </div>

            <!-- FONT -->
            <div v-if="activeCategory === 'font'" class="settings-section">
              <div class="settings-section-title">FONT</div>

              <div class="settings-field">
                <label class="settings-label">Terminal Font Size ({{ localSettings.display.terminalFontSize }}px)</label>
                <input
                  v-model.number="localSettings.display.terminalFontSize"
                  type="range"
                  class="settings-slider"
                  min="8"
                  max="24"
                  step="1"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Terminal Font Family</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.display.fontFamily"
                  :items="fontFamilyItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                />
                <p class="settings-hint">
                  Applies to the terminal renderer only.
                </p>
              </div>

              <div class="settings-field">
                <label class="settings-label">Interface Font Family</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.display.uiFontFamily"
                  :items="uiFontFamilyItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                  @update:model-value="onUiFontApply"
                />
                <p class="settings-hint">
                  Applies to the panels, settings and top bar.
                  <strong>Theme default</strong> uses whichever display font
                  the active theme ships with.
                </p>
              </div>
            </div>

            <!-- NETWORK -->
            <div v-if="activeCategory === 'network'" class="settings-section">
              <div class="settings-section-title">NETWORK</div>

              <div class="settings-field">
                <label class="settings-label">
                  Adapter
                  <span v-if="adapterItems.length <= 1" class="settings-hint" style="opacity:0.5">
                    (no adapters detected)
                  </span>
                </label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.network.adapter"
                  :items="adapterItems"
                  value-key="id"
                  label-key="label"
                  class="settings-select"
                />
                <p class="settings-hint">
                  Which interface drives the netstat panel and traffic
                  graphs. "Auto" picks the first up, non-loopback
                  interface; switch to your wifi (e.g. wlan0) when
                  ethernet is unplugged.
                </p>
              </div>

              <div class="settings-field">
                <label class="settings-label">Ping Target</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="pingTargetPreset"
                  :items="pingTargetPresetItems"
                  value-key="id"
                  label-key="label"
                  class="settings-select"
                />
                <p class="settings-hint">
                  Backend uses a TCP handshake (not ICMP) so this works
                  in sandboxed builds without CAP_NET_RAW. 8.8.8.8:53 is
                  Google DNS, 1.1.1.1:53 is Cloudflare. Pick "Custom" to
                  enter your own host:port.
                </p>
              </div>

              <div v-if="pingTargetPreset === 'custom'" class="settings-field">
                <label class="settings-label">Custom Ping Host:Port</label>
                <input
                  v-model="localSettings.network.pingTarget"
                  type="text"
                  class="settings-input"
                  placeholder="example.com:443"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">
                  Ping cadence ({{ formatSeconds(localSettings.network.pingIntervalMs, 2000) }}s)
                </label>
                <input
                  v-model.number="localSettings.network.pingIntervalMs"
                  type="range"
                  class="settings-slider"
                  min="1000"
                  max="30000"
                  step="500"
                />
                <p class="settings-hint">
                  How often the netstat panel issues a fresh TCP-handshake
                  ping. Faster = more responsive PING value but a tiny bit
                  more network chatter.
                </p>
              </div>

              <div class="settings-field">
                <label class="settings-label">
                  Traffic sample cadence ({{ formatSeconds(localSettings.network.trafficIntervalMs, 1000) }}s)
                </label>
                <input
                  v-model.number="localSettings.network.trafficIntervalMs"
                  type="range"
                  class="settings-slider"
                  min="250"
                  max="5000"
                  step="250"
                />
                <p class="settings-hint">
                  Sample rate for the up/down bandwidth chart. 1Hz is the
                  default and matches the panel's visual density; bump
                  down to 4Hz for smoother lines if your CPU can afford
                  the extra polling.
                </p>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Enable GeoIP lookup</label>
                <USwitch
                  v-model="localSettings.network.geoipEnabled"
                  size="md"
                  color="primary"
                  aria-label="Enable GeoIP lookup"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                When ON, the Globe panel resolves your public IP to a
                lat/lon and ISP via an HTTP lookup. Turn OFF for fully
                offline operation — the globe still renders, just
                without an endpoint marker.
              </p>

              <div v-if="localSettings.network.geoipEnabled" class="settings-field">
                <label class="settings-label">GeoIP endpoint (optional)</label>
                <input
                  v-model="localSettings.network.geoipEndpoint"
                  type="text"
                  class="settings-input"
                  placeholder="https://geo.kamero.ai/api/geo"
                />
                <p class="settings-hint">
                  Leave blank to use the default backend chain
                  (kamero → iplocate). Provide a custom URL only if
                  you're proxying through your own MaxMind / IPinfo
                  endpoint.
                </p>
              </div>
            </div>

            <!-- SECURITY -->
            <div v-if="activeCategory === 'security'" class="settings-section">
              <div class="settings-section-title">SECURITY</div>
              <p class="settings-hint">
                aDex runs as a single-user desktop application: it inherits
                your operating-system session and adds no separate login.
                Terminals run as your user, with your permissions.
              </p>
              <p class="settings-hint" style="opacity: 0.75;">
                The backend does apply input validation to paths and shell
                arguments — blocking directory traversal and command injection
                in filesystem and terminal requests. That protection is always
                on and has nothing to configure.
              </p>
              <div class="settings-field">
                <label class="settings-label">
                  Session Lock {{ lockConfigured ? '(enabled)' : '(not set up)' }}
                </label>
                <input
                  v-model="lockCurrent"
                  type="password"
                  class="settings-input"
                  :placeholder="lockConfigured ? 'current passphrase' : 'not required'"
                  autocomplete="off"
                />
                <input
                  v-model="lockNext"
                  type="password"
                  class="settings-input"
                  style="margin-top: 0.4rem;"
                  placeholder="new passphrase (min 4 characters)"
                  autocomplete="off"
                />
                <p v-if="lockMessage" class="settings-hint" :style="{ color: lockError ? '#ff6b6b' : undefined }">
                  {{ lockMessage }}
                </p>
                <div style="display: flex; gap: 0.5rem; margin-top: 0.5rem;">
                  <button type="button" class="settings-btn" @click="applyLockPassphrase">
                    {{ lockConfigured ? 'CHANGE PASSPHRASE' : 'SET PASSPHRASE' }}
                  </button>
                  <button
                    v-if="lockConfigured"
                    type="button"
                    class="settings-btn"
                    @click="removeLock"
                  >REMOVE LOCK</button>
                </div>
                <p class="settings-hint">
                  Locks the running session behind a passphrase — press
                  <kbd>Ctrl + Shift + L</kbd> to lock. Terminals keep running
                  while locked. The backend refuses file, terminal and process
                  access while locked, so the overlay is not the only barrier.
                </p>
              </div>

              <div v-if="lockConfigured" class="settings-field">
                <label class="settings-label">Auto-lock after idle</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="lockIdleTimeout"
                  :items="lockIdleItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                  @update:model-value="applyIdleTimeout"
                />
              </div>
            </div>

            <!-- AUDIO -->
            <div v-if="activeCategory === 'audio'" class="settings-section">
              <div class="settings-section-title">AUDIO SETTINGS</div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Enable Audio</label>
                <USwitch
                  v-model="localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Enable audio"
                  @update:model-value="onAudioEnabledChange"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Master Volume ({{ localSettings.audio.volume }}%)</label>
                <input
                  v-model.number="localSettings.audio.volume"
                  type="range"
                  class="settings-slider"
                  min="0"
                  max="100"
                  step="1"
                  :disabled="!localSettings.audio.enabled"
                  @input="onMasterVolumeInput"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Soundpack</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.audio.soundpack"
                  :items="audioPackItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                  :disabled="!localSettings.audio.enabled"
                  @update:model-value="onSoundpackChange"
                />
                <p class="settings-hint">
                  <strong>edex</strong> ships the original 13 WAV cues;
                  <strong>synth</strong> generates them via Web Audio for
                  accessibility / asset-missing modes.
                </p>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Mute in Background</label>
                <USwitch
                  v-model="localSettings.audio.muteInBackground"
                  :disabled="!localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Mute in background"
                  @update:model-value="persistCurrentSettings"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Boot Splash Audio</label>
                <USwitch
                  v-model="localSettings.audio.boot"
                  :disabled="!localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Toggle boot splash audio"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Plays a beep on every step of the launch sequence.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Shutdown Splash Audio</label>
                <USwitch
                  v-model="localSettings.audio.shutdown"
                  :disabled="!localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Toggle shutdown splash audio"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Plays a cue per stage during the quit sequence.
              </p>

              <!-- Per-category audio toggles. Each maps to a CueCategory
                   in useAdexAudio (keyboard, destructive, interface) so
                   the user can silence one class of cue without losing
                   the others. System-category cues (boot/shutdown
                   splash, alarms) are governed by the boot/shutdown
                   toggles above. -->
              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Keyboard Click</label>
                <USwitch
                  v-model="localSettings.audio.categoryKeyboard"
                  :disabled="!localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Toggle keyboard click sounds"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Mechanical-keyboard tick on every terminal keystroke.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Destructive Confirm Tone</label>
                <USwitch
                  v-model="localSettings.audio.categoryDestructive"
                  :disabled="!localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Toggle destructive-action confirm sounds"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Handgun-click cue when you confirm a destructive action
                (Quit, Kill Process, Delete).
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Interface Clicks</label>
                <USwitch
                  v-model="localSettings.audio.categoryInterface"
                  :disabled="!localSettings.audio.enabled"
                  size="md"
                  color="primary"
                  aria-label="Toggle UI click sounds"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Click-tone on buttons, menus, panel toggles, and tabs.
              </p>
            </div>

            <!-- SYSTEM -->
            <!-- KEYBINDS -->
            <div v-if="activeCategory === 'keybinds'" class="settings-section">
              <div class="settings-section-title">KEYBOARD SHORTCUTS</div>
              <p class="settings-hint">
                Shortcuts use Ctrl+Shift so they cannot collide with the
                readline bindings your shell uses inside the terminal.
              </p>

              <div class="keybind-list">
                <div v-for="bind in KEYBINDS" :key="bind.action" class="keybind-row">
                  <span class="keybind-action">{{ bind.action }}</span>
                  <kbd class="keybind-keys">{{ bind.keys }}</kbd>
                </div>
              </div>

              <p class="settings-hint" style="opacity: 0.55;">
                Shortcuts are fixed in this release; remapping is not yet
                available.
              </p>
            </div>

            <!-- MODULES -->
            <div v-if="activeCategory === 'modules'" class="settings-section">
              <div class="settings-section-title">MODULES</div>
              <p class="settings-hint">
                Show, hide and reorder individual panels. The top-bar buttons
                toggle whole columns; this is the per-panel control.
              </p>

              <div v-for="region in (['left','right','bottom'] as const)" :key="region" class="settings-field">
                <label class="settings-label">{{ regionLabel(region) }}</label>
                <div class="module-list">
                  <div
                    v-for="(mod, idx) in modulesEngine.allIn(region)"
                    :key="mod.id"
                    class="module-row"
                    :class="{ 'module-hidden': modulesEngine.isHidden(mod.id) }"
                  >
                    <USwitch
                      :model-value="!modulesEngine.isHidden(mod.id)"
                      size="sm"
                      color="primary"
                      :aria-label="'Show ' + mod.label"
                      @update:model-value="(v: boolean) => modulesEngine.setHidden(mod.id, !v)"
                    />
                    <span class="module-name">{{ mod.label }}</span>
                    <button
                      type="button"
                      class="module-move"
                      :disabled="idx === 0"
                      title="Move up"
                      @click="modulesEngine.move(mod.id, -1)"
                    >▲</button>
                    <button
                      type="button"
                      class="module-move"
                      :disabled="idx === modulesEngine.allIn(region).length - 1"
                      title="Move down"
                      @click="modulesEngine.move(mod.id, 1)"
                    >▼</button>
                  </div>
                </div>
              </div>

              <button type="button" class="settings-btn" @click="modulesEngine.reset()">
                RESET MODULE LAYOUT
              </button>
            </div>

            <div v-if="activeCategory === 'system'" class="settings-section">
              <div class="settings-section-title">SYSTEM SETTINGS</div>

              <div class="settings-field">
                <label class="settings-label">Open Terminal & File Manager In</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.system.initialCwd"
                  :items="initialCwdItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                />
                <p class="settings-hint">
                  <strong>Home</strong> ({{ startupPaths.home || '~' }}) opens
                  the file manager and new shells in your home directory.
                  <strong>Launch CWD</strong> ({{ startupPaths.cwd || '/' }})
                  uses the directory aDex-UI was launched from. Affects only
                  newly opened shells; existing tabs keep their cwd.
                </p>
              </div>

              <div class="settings-field">
                <label class="settings-label">Clock Format</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="localSettings.system.clockFormat"
                  :items="clockFormatItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  class="settings-select"
                />
              </div>

              <!-- Ping Address moved to Settings → Network → Ping Target
                   (where the live netstat panel reads from). The
                   System-tab row was a stub bound to a field that
                   wasn't in SettingsData defaults — typing into it
                   created an orphan localSettings.system.pingAddress
                   property that nothing read. Removed entirely. -->


              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Boot Animation</label>
                <USwitch
                  v-model="localSettings.system.bootAnimation"
                  size="md"
                  color="primary"
                  aria-label="Boot animation"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Show Grid Background</label>
                <USwitch
                  v-model="localSettings.system.gridBackground"
                  size="md"
                  color="primary"
                  aria-label="Show grid background"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Performance Mode</label>
                <USwitch
                  v-model="localSettings.system.performanceMode"
                  size="md"
                  color="primary"
                  aria-label="Performance mode"
                />
              </div>
            </div>

            <!-- ADVANCED -->
            <div v-if="activeCategory === 'advanced'" class="settings-section">
              <div class="settings-section-title">ADVANCED SETTINGS</div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Skip boot intro</label>
                <USwitch
                  v-model="localSettings.advanced.nointro"
                  size="md"
                  color="primary"
                  aria-label="Skip boot intro"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Bypasses the animated boot splash and goes straight to
                the shell. Equivalent to the legacy <code>--nointro</code> CLI flag.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Allow Windowed Mode</label>
                <USwitch
                  v-model="localSettings.advanced.allowWindowedMode"
                  size="md"
                  color="primary"
                  aria-label="Allow windowed mode"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Force Fullscreen</label>
                <USwitch
                  v-model="localSettings.advanced.forceFullscreen"
                  size="md"
                  color="primary"
                  aria-label="Force fullscreen on launch"
                />
              </div>
              <p class="settings-hint">
                Puts the window into true fullscreen after boot — no
                decorations, covering the display. Takes effect on next launch.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Desktop Notifications</label>
                <USwitch
                  v-model="localSettings.advanced.notifications"
                  size="md"
                  color="primary"
                  aria-label="Desktop notifications"
                  @update:model-value="onNotificationsToggle"
                />
              </div>
              <p class="settings-hint">
                Posts to your desktop's notification centre when a long-running
                operation finishes, so the result reaches you even when aDex is
                behind another window. Short operations never notify.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Disable GPU Acceleration</label>
                <USwitch
                  v-model="localSettings.advanced.disableGpu"
                  size="md"
                  color="primary"
                  aria-label="Disable GPU acceleration"
                />
              </div>
              <p class="settings-hint">
                <strong>Only turn this on if the app crashes.</strong> Some
                graphics drivers crash the rendering process — if aDex closes
                by itself, this usually stops it. The cost is a noticeably
                less responsive interface, since drawing falls back to the
                CPU. Takes effect on next launch.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Frameless Window</label>
                <USwitch
                  v-model="localSettings.advanced.frameless"
                  size="md"
                  color="primary"
                  aria-label="Frameless window"
                  @update:model-value="onFramelessApply"
                />
              </div>
              <p class="settings-hint">
                Removes the OS titlebar and border. Applies immediately; use
                the window controls in the top bar to move, resize and close.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Hide Dotfiles in File Manager</label>
                <USwitch
                  v-model="localSettings.advanced.hideDotfiles"
                  size="md"
                  color="primary"
                  aria-label="Hide dotfiles"
                  @update:model-value="persistCurrentSettings"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">File Manager: List View</label>
                <USwitch
                  v-model="localSettings.advanced.fsListView"
                  size="md"
                  color="primary"
                  aria-label="File manager list view"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                ON renders the file manager as a vertical list with
                metadata; OFF (default) keeps the grid of icon tiles.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Experimental Globe Features</label>
                <USwitch
                  v-model="localSettings.advanced.experimentalGlobeFeatures"
                  size="md"
                  color="primary"
                  aria-label="Experimental globe features"
                  @update:model-value="persistCurrentSettings"
                />
              </div>
              <p class="settings-hint">
                Connection trace lines and satellite track overlays
                on the Globe panel. Rough around the edges.
              </p>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Experimental Features</label>
                <USwitch
                  v-model="localSettings.advanced.experimentalFeatures"
                  size="md"
                  color="primary"
                  aria-label="Experimental features"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">WebGL Renderer</label>
                <USwitch
                  v-model="localSettings.advanced.webglRenderer"
                  size="md"
                  color="primary"
                  aria-label="WebGL renderer"
                />
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Debug Mode</label>
                <USwitch
                  v-model="localSettings.advanced.debugMode"
                  size="md"
                  color="primary"
                  aria-label="Debug mode"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Terminal Scrollback ({{ localSettings.advanced.scrollback }} lines)</label>
                <input
                  v-model.number="localSettings.advanced.scrollback"
                  type="range"
                  class="settings-slider"
                  min="500"
                  max="50000"
                  step="500"
                />
              </div>
            </div>

            <!-- UPDATES -->
            <div v-if="activeCategory === 'updates'" class="settings-section">
              <div class="settings-section-title">UPDATES</div>

              <div class="settings-field">
                <label class="settings-label">Current version</label>
                <div class="settings-readonly">{{ updates.currentVersion }}</div>
              </div>

              <div v-if="updates.state.value.latest" class="settings-field">
                <label class="settings-label">Latest release</label>
                <div class="settings-readonly">
                  {{ updates.state.value.latest.tag_name }}
                  <span v-if="updates.hasUpdate.value" class="settings-readonly-tag settings-readonly-tag-new">
                    UPDATE AVAILABLE
                  </span>
                  <span v-else class="settings-readonly-tag">UP TO DATE</span>
                </div>
              </div>

              <div v-if="updates.state.value.lastChecked" class="settings-field">
                <label class="settings-label">Last checked</label>
                <div class="settings-readonly">{{ formatLastChecked(updates.state.value.lastChecked) }}</div>
              </div>

              <div v-if="updates.state.value.lastError" class="settings-field">
                <label class="settings-label">Last error</label>
                <div class="settings-readonly settings-readonly-err">
                  {{ updates.state.value.lastError }}
                </div>
              </div>

              <div class="settings-field settings-field-toggle">
                <label class="settings-label">Auto-check for updates</label>
                <USwitch
                  v-model="updates.settings.value.enabled"
                  size="md"
                  color="primary"
                  aria-label="Enable update checks"
                />
              </div>

              <div class="settings-field">
                <label class="settings-label">Check frequency</label>
                <USelectMenu
                  variant="none"
                  color="neutral"
                  v-model="updates.settings.value.interval"
                  :items="updateIntervalItems"
                  value-key="id"
                  label-key="label"
                  :search-input="false"
                  :disabled="!updates.settings.value.enabled"
                  class="settings-select"
                />

                <p class="settings-hint">
                  GitHub anonymous API allows 60 requests/hour per IP,
                  so even "every launch" is comfortably within budget.
                  Use "Manual only" if you'd rather poll yourself.
                </p>
              </div>

              <div class="settings-field">
                <button
                  type="button"
                  class="settings-btn settings-btn-secondary"
                  :disabled="updates.checking.value"
                  @click="checkForUpdatesNow"
                >
                  {{ updates.checking.value ? 'CHECKING...' : 'CHECK NOW' }}
                </button>
                <button
                  v-if="updates.hasUpdate.value"
                  type="button"
                  class="settings-btn settings-btn-primary"
                  style="margin-left: 0.5vw"
                  @click="showUpdateModalFromSettings"
                >
                  VIEW UPDATE
                </button>
              </div>

              <div v-if="updates.settings.value.skippedVersions.length > 0" class="settings-field">
                <label class="settings-label">
                  Skipped versions ({{ updates.settings.value.skippedVersions.length }})
                </label>
                <div class="settings-readonly">
                  {{ updates.settings.value.skippedVersions.join(', ') }}
                </div>
                <button
                  type="button"
                  class="settings-btn settings-btn-secondary"
                  style="margin-top: 0.4vh"
                  @click="updates.unskipAll()"
                >
                  CLEAR SKIP LIST
                </button>
                <p class="settings-hint">
                  Clearing the skip list lets the update modal
                  re-surface for any version you previously dismissed.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer: action buttons -->
        <!-- Resize grip: bottom-right corner. -->
        <div
          class="settings-resize-grip"
          title="Drag to resize"
          @pointerdown="startResize"
        />

        <div class="settings-footer">
          <button class="settings-btn settings-btn-secondary" @click="resetToDefaults">
            RESET DEFAULTS
          </button>
          <div class="settings-footer-spacer" />
          <button class="settings-btn settings-btn-secondary" @click="cancel">
            CANCEL
          </button>
          <button class="settings-btn settings-btn-primary" @click="save">
            SAVE
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted } from 'vue'
import { useStorage, useDebounceFn, useEventListener } from '@vueuse/core'
import { useAdexTheme } from '~/composables/useAdexTheme'
import { useCustomLayouts } from '~/composables/useCustomLayouts'
import { useModules } from '~/composables/useModules'
import { WindowRuntime } from '~/lib/wailsjs/runtime'
import { useAdexKeyboard } from '~/composables/useAdexKeyboard'
import { useAdexAudio } from '~/composables/useAdexAudio'
import { useUpdateChecker } from '~/composables/useUpdateChecker'
import { useNetworkStore } from '~/stores/network'
import { format as formatDateFn, isValid as isValidDate } from 'date-fns'
import type { AdexThemeIndexEntry } from '~/types/adex-theme'
import type { KbLayoutIndexEntry } from '~/types/kb-layout'

// ---- Types ----

interface SettingsData {
  shell: {
    path: string
    args: string
    workingDirectory: string
  }
  display: {
    theme: string
    keyboardLayout: string
    terminalFontSize: number
    fontFamily: string
    /** Interface font. Empty means follow the active theme. */
    uiFontFamily: string
    /** Layout preset override. Empty string = follow the active theme's
     *  layout. Otherwise one of the layout CSS files in
     *  assets/css/layouts/: 'default', 'disrupted', 'typeleft',
     *  'fulltype', 'notype', 'colorfilter'. Read by pages/index.vue. */
    layout: string
    /** Globe panel implementation: 'classic' dot grid or 'geo' country
     *  outlines. */
    globeStyle: string
  }
  audio: {
    enabled: boolean
    volume: number
    soundpack: string
    muteInBackground: boolean
    /** Boot splash beep-per-step. Subordinate to `enabled`. */
    boot: boolean
    /** Shutdown splash cue-per-step. Subordinate to `enabled`. */
    shutdown: boolean
    /** Per-category opt-outs. All default to true (audible).
     *  Read by pages/index.vue and forwarded into useAdexAudio's
     *  category gate so a single click toggles ALL cues of that
     *  category app-wide. */
    categoryKeyboard: boolean
    categoryDestructive: boolean
    categoryInterface: boolean
  }
  system: {
    /** Where the terminal + file manager open on launch. `home` is the
     *  user's $HOME; `cwd` is the directory aDex-UI was launched from. */
    initialCwd: 'home' | 'cwd'
    clockFormat: '12h' | '24h'
    bootAnimation: boolean
    gridBackground: boolean
    performanceMode: boolean
  }
  network: {
    /** host:port for the latency probe. Empty → backend default
     *  (8.8.8.8:53). Read by AdexNetstat via useStorage('adex-settings'). */
    pingTarget: string
    /** Active interface name override (e.g. 'wlan0'). Empty → auto-pick
     *  first up, non-loopback. */
    adapter: string
    /** Ping cadence in milliseconds (1000..30000). Read by the netstat
     *  panel's useIntervalFn. Default 2000ms (2s). */
    pingIntervalMs: number
    /** Traffic chart sample cadence in milliseconds (250..5000).
     *  Default 1000ms (1Hz). */
    trafficIntervalMs: number
    /** GeoIP enrichment for the globe / netstat endpoint info. When
     *  false, AdexGlobe shows lat/lon as Unknown and skips the HTTP
     *  fetch entirely — useful on air-gapped systems or to avoid the
     *  outbound request to the geo provider. */
    geoipEnabled: boolean
    /** Override URL for the geoIP provider. Empty → backend default
     *  (kamero + iplocate fallback). Power users with their own
     *  MaxMind-fronted service can point here. */
    geoipEndpoint: string
  }
  advanced: {
    allowWindowedMode: boolean
    experimentalFeatures: boolean
    webglRenderer: boolean
    debugMode: boolean
    scrollback: number
    /** Skip the boot splash entirely. Read by pages/index.vue on
     *  startup; equivalent to passing `--nointro` to the original
     *  eDex CLI. */
    nointro: boolean
    /** Force the window into fullscreen on launch even if the OS WM
     *  would normally honor the saved Windowed state. */
    forceFullscreen: boolean
    /** Fall back to CPU rendering. Costs responsiveness; avoids driver crashes. */
    disableGpu: boolean
    /** Post desktop notifications for long-running operations. */
    notifications: boolean
    /** Remove the OS titlebar/border. Applied live via the runtime. */
    frameless: boolean
    /** Hide dotfiles (.git, .config, etc.) in the file manager. */
    hideDotfiles: boolean
    /** Render the file manager as a list instead of the grid. */
    fsListView: boolean
    /** Show experimental controls on the Globe panel (connection
     *  trace lines, satellite track overlay, etc.). Off by default
     *  because the visuals are still rough. */
    experimentalGlobeFeatures: boolean
  }
}

interface Category {
  id: string
  label: string
}

// ---- Props / Emits ----

const props = withDefaults(
  defineProps<{
    modelValue: boolean
  }>(),
  {
    modelValue: false,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'close'): void
  (e: 'settings-changed', settings: SettingsData): void
  // Settings → Updates tab can request the host page to mount the
  // UpdateAvailableModal — keeping the modal at the page level
  // (vs nested under settings) so it survives the settings modal
  // closing and can teleport without z-index fights.
  (e: 'show-update-modal'): void
}>()

// ---- Constants ----

const STORAGE_KEY = 'adex-settings'

const categories: Category[] = [
  { id: 'terminal',    label: 'TERMINAL' },
  { id: 'theme',       label: 'THEME' },
  { id: 'keyboard',    label: 'KEYBOARD' },
  { id: 'font',        label: 'FONT' },
  { id: 'audio',       label: 'AUDIO' },
  { id: 'network',     label: 'NETWORK' },
  { id: 'security',    label: 'SECURITY' },
  { id: 'keybinds',    label: 'KEYBINDS' },
  { id: 'modules',     label: 'MODULES' },
  { id: 'system',      label: 'SYSTEM' },
  { id: 'advanced',    label: 'ADVANCED' },
  { id: 'updates',     label: 'UPDATES' },
]

// ---- State ----

const activeCategory = ref('terminal')

// V2 engines — loaded lazily so the Settings modal can re-use the same
// running theme and keyboard runtimes the shell already booted.
const themeEngine = useAdexTheme()
const kbEngine = useAdexKeyboard()
// Update checker — exposes settings, state, currentVersion, hasUpdate,
// and the checkNow / unskipAll actions consumed by the Updates tab.
// Module-scope refs inside the composable mean every consumer
// (Settings, page-level autocheck, UpdateAvailableModal) sees the same
// data without us having to pass it down.
const updates = useUpdateChecker()

function formatLastChecked(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (!isValidDate(d)) return iso
  return formatDateFn(d, 'yyyy-MM-dd HH:mm')
}

async function checkForUpdatesNow() {
  // The composable handles the in-flight guard via its `checking` ref,
  // so a double-click here is a no-op. We just kick the request.
  await updates.checkNow()
  // If the result is a new version, surface it. Emitting up the chain
  // so the parent (pages/index.vue) can mount UpdateAvailableModal —
  // we can't open the modal from inside the settings modal because
  // it would teleport over us and steal focus from the settings UI.
  if (updates.hasUpdate.value) {
    emit('show-update-modal')
  }
}

function showUpdateModalFromSettings() {
  // The "View Update" button — same emit path as the auto-surface
  // route, but skips the network call since we already have a cached
  // release in state.
  emit('show-update-modal')
}

// Reactive indices for the USelectMenu items.
const themeIndex = computed<AdexThemeIndexEntry[]>(() => themeEngine.index.value)
const kbIndex = computed<KbLayoutIndexEntry[]>(() => kbEngine.index.value)

// USelectMenu wants `{ id, label }`-shaped items by default; we pass the
// raw index entries and tell it which key to read for value/label.
const themeMenuItems = computed(() =>
  themeIndex.value.map((t) => ({
    id: t.id,
    displayName: `${t.displayName} — ${t.layout}`,
  })),
)
const kbMenuItems = computed(() =>
  kbIndex.value.map((k) => ({
    id: k.id,
    displayName: k.displayName ?? k.id,
  })),
)

/* Movable + resizable modal ------------------------------------------------
 *
 * Position and size are persisted, so the panel reopens where the user left
 * it. Both are clamped to the viewport on every change, so a window that
 * shrinks (or a saved position from a larger display) can never strand the
 * modal off-screen with no way to reach its header. */

const modalRef = ref<HTMLElement | null>(null)

const modalBox = useStorage('adex.settings.modalBox', {
  x: null as number | null,
  y: null as number | null,
  w: null as number | null,
  h: null as number | null,
})

const modalStyle = computed(() => {
  const { x, y, w, h } = modalBox.value
  const style: Record<string, string> = {}
  if (w != null) style.width = `${w}px`
  if (h != null) style.height = `${h}px`
  if (x != null && y != null) {
    // Opt out of the overlay's centring once the user has moved it.
    style.position = 'fixed'
    style.left = `${x}px`
    style.top = `${y}px`
    style.margin = '0'
  }
  return style
})

const MIN_W = 480
const MIN_H = 320

function clampToViewport() {
  const el = modalRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const maxX = Math.max(0, window.innerWidth - r.width)
  const maxY = Math.max(0, window.innerHeight - r.height)
  const b = modalBox.value
  if (b.x != null) b.x = Math.min(Math.max(0, b.x), maxX)
  if (b.y != null) b.y = Math.min(Math.max(0, b.y), maxY)
  if (b.w != null) b.w = Math.min(Math.max(MIN_W, b.w), window.innerWidth)
  if (b.h != null) b.h = Math.min(Math.max(MIN_H, b.h), window.innerHeight)
}

function startDrag(e: PointerEvent) {
  // Ignore drags that begin on the close button or any control.
  const target = e.target as HTMLElement
  if (target.closest('button, input, select, textarea')) return

  const el = modalRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const offX = e.clientX - r.left
  const offY = e.clientY - r.top
  // Freeze the current geometry so the first move does not jump from the
  // centred position.
  modalBox.value = { ...modalBox.value, x: r.left, y: r.top, w: r.width, h: r.height }

  const move = (ev: PointerEvent) => {
    modalBox.value.x = ev.clientX - offX
    modalBox.value.y = ev.clientY - offY
    clampToViewport()
  }
  const up = () => {
    window.removeEventListener('pointermove', move)
    window.removeEventListener('pointerup', up)
  }
  window.addEventListener('pointermove', move)
  window.addEventListener('pointerup', up)
  e.preventDefault()
}

function startResize(e: PointerEvent) {
  const el = modalRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const startX = e.clientX
  const startY = e.clientY
  const startW = r.width
  const startH = r.height
  modalBox.value = { ...modalBox.value, x: r.left, y: r.top, w: startW, h: startH }

  const move = (ev: PointerEvent) => {
    modalBox.value.w = Math.max(MIN_W, startW + (ev.clientX - startX))
    modalBox.value.h = Math.max(MIN_H, startH + (ev.clientY - startY))
    clampToViewport()
  }
  const up = () => {
    window.removeEventListener('pointermove', move)
    window.removeEventListener('pointerup', up)
  }
  window.addEventListener('pointermove', move)
  window.addEventListener('pointerup', up)
  e.preventDefault()
  e.stopPropagation()
}

/** Reset to the default centred geometry. */
function resetModalGeometry() {
  modalBox.value = { x: null, y: null, w: null, h: null }
}

// A window resize can leave a saved position off-screen.
useEventListener(window, 'resize', clampToViewport)

// Globe style items — must match validGlobeStyles in the Go validator.
const globeStyleItems = [
  { id: 'classic', label: 'Classic (dot grid)' },
  { id: 'geo', label: 'Countries (outlines)' },
]

// Layout preset items — matches the CSS files in assets/css/layouts/.
// Empty id ('') means "follow the theme's bundled layout" — that's the
// sentinel pages/index.vue's computed treats as "no override".
/** Frameless applies immediately — Wails v3 can toggle decorations at
 *  runtime, so there is no reason to make the user restart. */
function onFramelessApply(value: unknown) {
  try {
    WindowRuntime.SetFrameless(value === true)
  } catch (err) {
    console.error('[Settings] failed to apply frameless:', err)
  }
}

/** The application's keyboard shortcuts, for the Keybinds panel.
 *  Kept in sync by hand with the onKeyStroke handlers in pages/index.vue. */
/* Session lock controls ---------------------------------------------------- */
const lockConfigured = ref(false)
const lockCurrent = ref('')
const lockNext = ref('')
const lockMessage = ref('')
const lockError = ref(false)
const lockIdleTimeout = ref(0)

const lockIdleItems = [
  { id: 0,    label: 'Never' },
  { id: 300,  label: 'After 5 minutes' },
  { id: 900,  label: 'After 15 minutes' },
  { id: 1800, label: 'After 30 minutes' },
  { id: 3600, label: 'After 1 hour' },
]

async function refreshLockState() {
  try {
    const { IsLockConfigured, GetLockIdleTimeout } = await import('~/lib/wailsjs/coordinator')
    lockConfigured.value = await IsLockConfigured()
    lockIdleTimeout.value = await GetLockIdleTimeout()
  } catch {
    lockConfigured.value = false
  }
}

async function applyLockPassphrase() {
  lockMessage.value = ''
  lockError.value = false
  try {
    const { SetLockPassphrase } = await import('~/lib/wailsjs/coordinator')
    await SetLockPassphrase(lockCurrent.value, lockNext.value)
    lockCurrent.value = ''
    lockNext.value = ''
    lockMessage.value = 'Passphrase saved.'
    await refreshLockState()
  } catch (err) {
    lockError.value = true
    lockMessage.value = err instanceof Error ? err.message : 'Could not set the passphrase'
  }
}

async function removeLock() {
  lockMessage.value = ''
  lockError.value = false
  try {
    const { DisableLock } = await import('~/lib/wailsjs/coordinator')
    await DisableLock(lockCurrent.value)
    lockCurrent.value = ''
    lockMessage.value = 'Lock removed.'
    await refreshLockState()
  } catch (err) {
    lockError.value = true
    lockMessage.value = err instanceof Error ? err.message : 'Could not remove the lock'
  }
}

/** Ask the platform for permission the first time notifications are enabled;
 *  without it the toggle would appear on while nothing is ever delivered. */
async function onNotificationsToggle(value: unknown) {
  if (value !== true) return
  try {
    const { RequestNotificationPermission } = await import('~/lib/wailsjs/coordinator')
    const granted = await RequestNotificationPermission()
    if (!granted) {
      localSettings.advanced.notifications = false
      console.warn('[Settings] notification permission was refused')
    }
  } catch (err) {
    console.error('[Settings] could not request notification permission:', err)
  }
}

async function applyIdleTimeout(value: unknown) {
  try {
    const { SetLockIdleTimeout } = await import('~/lib/wailsjs/coordinator')
    await SetLockIdleTimeout(Number(value) || 0)
  } catch (err) {
    console.error('[Settings] failed to set idle timeout:', err)
  }
}

const KEYBINDS = [
  { action: 'Open settings',        keys: 'Ctrl + ,' },
  { action: 'Lock session',         keys: 'Ctrl + Shift + L' },
  { action: 'New terminal tab',     keys: 'Ctrl + Shift + T' },
  { action: 'Close terminal tab',   keys: 'Ctrl + Shift + W' },
  { action: 'Switch to tab 1–5',    keys: 'Ctrl + Shift + 1…5' },
  { action: 'Quit',                 keys: 'Ctrl + Shift + Q' },
  { action: 'Toggle terminal',      keys: 'Ctrl + `' },
  { action: 'Toggle file browser',  keys: 'Ctrl + Shift + F' },
  { action: 'Toggle system monitor', keys: 'Ctrl + Shift + S' },
]

/** Interface font choices. The empty id means "follow the theme". */
const uiFontFamilyItems = computed(() => [
  { id: '', label: 'Theme default' },
  ...fontFamilyItems.value.map(f => ({ id: f.id, label: f.label })),
])

/** Applied live so the effect is visible before saving, matching theme and
 *  keyboard. An empty value restores the theme's own display font. */
function onUiFontApply(value: unknown) {
  const root = document.documentElement
  const font = typeof value === 'string' ? value.trim() : ''
  if (font) {
    root.style.setProperty('--font_main', font)
    root.style.setProperty('--font_main_light', font)
  } else {
    // Clearing the override lets the theme's value apply again.
    root.style.removeProperty('--font_main')
    root.style.removeProperty('--font_main_light')
    void themeEngine.setTheme(themeEngine.activeId.value)
  }
}

const clockFormatItems = [
  { id: '24h', label: '24-Hour' },
  { id: '12h', label: '12-Hour (AM/PM)' },
]

const updateIntervalItems = [
  { id: 'manual', label: 'Manual only' },
  { id: 'launch', label: 'On every launch' },
  { id: 'daily',  label: 'Once a day' },
  { id: 'weekly', label: 'Once a week' },
]

const BUILTIN_LAYOUTS = [
  { id: '',            label: 'Theme default' },
  { id: 'default',     label: 'Default (classic eDex)' },
  { id: 'disrupted',   label: 'Disrupted (terminal top)' },
  { id: 'typeleft',    label: 'Typeleft (keyboard left)' },
  { id: 'fulltype',    label: 'Fulltype (full-width keyboard)' },
  { id: 'notype',      label: 'Notype (no on-screen keyboard)' },
  { id: 'colorfilter', label: 'Colorfilter (minimal chrome)' },
  { id: 'terminal-focus', label: 'Terminal focus (terminal fills the window)' },
]

// Built-ins plus anything the user defined in layouts.json. Custom entries
// are suffixed so they are distinguishable in the list.
const customLayoutEngine = useCustomLayouts()
const modulesEngine = useModules()

function regionLabel(region: 'left' | 'right' | 'bottom'): string {
  return { left: 'Left column', right: 'Right column', bottom: 'Bottom row' }[region]
}
const layoutItems = computed(() => [
  ...BUILTIN_LAYOUTS,
  ...customLayoutEngine.layouts.value.map(l => ({
    id: l.id,
    label: `${l.displayName || l.id} (custom)`,
  })),
])

/** Where users create or edit custom layouts — surfaced in the Theme panel. */
const customLayoutsPath = computed(() => customLayoutEngine.path.value)

// Apply theme/keyboard immediately on selection so the user sees the
// effect without having to hit Save. We still write to localSettings so
// the Save/Cancel/Reset buttons keep their existing semantics.
let themeApplyInFlight = false
async function onThemeApply(id: string | null | undefined) {
  if (!id || themeApplyInFlight) return
  themeApplyInFlight = true
  try {
    await themeEngine.setTheme(id)
    // 'theme' cue — the long sweep that signals a palette shift.
    try { audioEngine.playCue('theme') } catch { /* non-fatal */ }
  } catch (err) {
    console.error('[Settings] failed to apply theme:', err)
  } finally {
    themeApplyInFlight = false
  }
}

async function onKeyboardLayoutApply(id: string | null | undefined) {
  if (!id) return
  try {
    await kbEngine.setLayout(id)
  } catch (err) {
    console.error('[Settings] failed to apply keyboard layout:', err)
  }
}

const shellPlaceholder = computed(() => {
  if (typeof navigator === 'undefined') return '/bin/zsh'
  const ua = navigator.userAgent || ''
  if (/Windows/i.test(ua)) return 'pwsh.exe'
  if (/Mac/i.test(ua)) return '/bin/zsh'
  return '/bin/zsh'
})

// Persists the override in the backend so subsequent CreateTerminalIn calls
// pick up the new shell. Empty string clears the override and falls back to
// $SHELL / $ComSpec / platform default.
async function onShellPathChange() {
  try {
    const { SetShellCommand } = await import('~/lib/wailsjs/coordinator')
    await SetShellCommand((localSettings.shell.path || '').trim())
  } catch (err) {
    console.warn('[Settings] failed to set shell command:', err)
  }
}

// Network adapter list — pulled live from the network store so the
// dropdown reflects what the OS reports. The "Auto" entry stores the
// empty string, which AdexNetstat / AdexTraffic interpret as "first up,
// non-loopback".
const networkStore = useNetworkStore()
const adapterItems = computed(() => {
  const interfaces = (networkStore.metrics?.interfaces ?? []) as any[]
  const seen = new Set<string>()
  const items: Array<{ id: string; label: string }> = [
    { id: '', label: 'Auto (first up)' },
  ]
  for (const iface of interfaces) {
    if (!iface?.name || iface.name === 'lo' || seen.has(iface.name)) continue
    seen.add(iface.name)
    const tag = iface.isUp ? '' : ' (down)'
    items.push({ id: iface.name, label: `${iface.name}${tag}` })
  }
  return items
})

// Ping-target presets: Google DNS, Cloudflare DNS, or a custom
// host:port. The preset choice is derived from the actual stored value
// (so opening Settings later reflects what's saved), and writing to it
// flips localSettings.network.pingTarget.
const PING_PRESETS = {
  google:     '8.8.8.8:53',
  cloudflare: '1.1.1.1:53',
} as const

const pingTargetPresetItems = [
  { id: 'google',     label: 'Google DNS (8.8.8.8:53)' },
  { id: 'cloudflare', label: 'Cloudflare DNS (1.1.1.1:53)' },
  { id: 'custom',     label: 'Custom host:port' },
] as const

const pingTargetPreset = computed<'google' | 'cloudflare' | 'custom'>({
  get() {
    const t = (localSettings.network.pingTarget || '').trim()
    if (t === PING_PRESETS.google) return 'google'
    if (t === PING_PRESETS.cloudflare) return 'cloudflare'
    return 'custom'
  },
  set(id) {
    if (id === 'google' || id === 'cloudflare') {
      localSettings.network.pingTarget = PING_PRESETS[id]
    } else if (
      localSettings.network.pingTarget === PING_PRESETS.google ||
      localSettings.network.pingTarget === PING_PRESETS.cloudflare ||
      !localSettings.network.pingTarget
    ) {
      // Switching to "Custom" while a preset is selected: blank the
      // field so the user types their own value rather than editing
      // the preset string in place.
      localSettings.network.pingTarget = ''
    }
  },
})

// V2 audio engine — soundpack switcher (edex / synth) + global mute.
const audioEngine = useAdexAudio()
// Pack selector — controls the playback ENGINE, not an asset folder.
//   'adex'  → prefer Howler-loaded WAV samples (all cues live under
//             `assets/audio/adex/`; the URL map routes each cue ID to
//             its own file).
//   'synth' → bypass all WAVs and use the Web-Audio oscillator
//             fallbacks defined in SYNTH_RECIPES.
// User-facing labels are descriptive — the internal id is what gets
// persisted and matched against `setPack(...)`.
const audioPackItems = [
  { id: 'adex',  label: 'WAV samples (recommended)' },
  { id: 'synth', label: 'Synth fallback (Web Audio)' },
] as const

function onSoundpackChange(id: string | null | undefined) {
  if (id === 'adex' || id === 'synth') {
    audioEngine.setPack(id)
  }
}

function onAudioEnabledChange(value: boolean) {
  // USwitch's v-model has already flipped localSettings.audio.enabled
  // by the time this fires — we only run the side effects here so the
  // store and the UI stay in sync without a double-write.
  //
  // Wrap engine calls in try/catch because they touch a module-scope
  // composable: if useAdexAudio's settings ref is in a broken state
  // (e.g. localStorage has a malformed `adex.audio.settings` JSON from
  // an older schema), throwing here aborts Vue's @update:model-value
  // handler chain — and the OBSERVABLE symptom is "clicking the toggle
  // does nothing" because v-model's own update fires AFTER this
  // handler. Swallow + log so the form-state update always lands.
  try {
    audioEngine.setMuted(!value)
    if (value) {
      // Confirmation cue when turning audio back on so the user gets
      // immediate feedback that the engine is alive. The cue itself
      // would silently no-op if value is false.
      audioEngine.playCue('click')
    }
  } catch (err) {
    console.error('[settings] audio engine update failed:', err)
  }
  // Auto-persist toggle changes. Without this the user has to click
  // Save before flipping a switch sticks — and a casual "open settings,
  // toggle, close with X" cycle reverts to the stored value, which
  // confuses anyone expecting one-click stickiness for boolean toggles.
  persistCurrentSettings()
}

// Persist the current form state immediately. Used by every audio /
// category toggle so individual switches auto-save without needing
// the user to find the Save button. Text inputs (shell path, ping
// target, etc.) still wait for explicit Save because debouncing every
// keystroke would write to localStorage on every character.
// Raw persist function. Used directly by toggle-style handlers where
// the user click is the trigger and a single localStorage write is
// expected per click.
function persistCurrentSettingsNow() {
  try {
    saveToStorage(JSON.parse(JSON.stringify(localSettings)))
  } catch (err) {
    console.error('[settings] auto-persist failed:', err)
  }
}

// Debounced wrapper for high-frequency callers — sliders, text inputs
// in `@input` mode, anything where the user could fire dozens of
// updates per second. 250ms balances "feels instant" against "doesn't
// hammer localStorage". VueUse's useDebounceFn auto-cleans on
// component unmount and exposes `.flush()` if we ever need to force-
// commit before close.
const persistCurrentSettings = useDebounceFn(persistCurrentSettingsNow, 250)

// Render a millisecond value as a decimal-seconds string for slider
// labels. Defensive against undefined / NaN from older saved settings
// that didn't include the new ping/traffic-interval fields — the
// previous direct `(undefined / 1000).toFixed(1)` produced the literal
// "NaN" string that appeared in label headers (the "(NANS)" bug).
function formatSeconds(ms: number | undefined | null, fallbackMs: number): string {
  const v = Number(ms ?? fallbackMs)
  if (!Number.isFinite(v)) return (fallbackMs / 1000).toFixed(1)
  return (v / 1000).toFixed(1)
}

function onMasterVolumeInput() {
  // Clamp to 0..1 for the engine; localSettings stores 0..100 for UX.
  audioEngine.setGlobalVolume(localSettings.audio.volume / 100)
}

// ---- Font tab ----

// The font family <USelectMenu> needs `{ id, label }` records; we keep
// the underlying value identical to the previous CSS family string so
// `localSettings.display.fontFamily` continues to round-trip into the
// stylesheet without a migration.
const fontFamilyItems = [
  { id: "'Fira Code', monospace",        label: 'Fira Code' },
  { id: "'JetBrains Mono', monospace",    label: 'JetBrains Mono' },
  { id: "'Cascadia Code', monospace",     label: 'Cascadia Code' },
  { id: "'Source Code Pro', monospace",   label: 'Source Code Pro' },
  { id: "'Consolas', monospace",          label: 'Consolas' },
  { id: "'Monaco', monospace",            label: 'Monaco' },
  { id: 'monospace',                      label: 'System Monospace' },
] as const

// ---- Color Scheme tab ----
//
// The full ColorScheme manager (CRUD + import dialog) lives in
// components/colorscheme/* and writes to its own pinia store. From the
// Settings modal we expose just the picker + the "Use theme colors"
// toggle; opening the create/edit/import flows happens from the
// dedicated ColorScheme manager opened via a button in M7 follow-up.
const useThemeColors = ref(true)
const selectedColorScheme = ref<string>('default')
// Lazy-loaded list of color schemes from the coordinator. Starts empty
// so the USelectMenu doesn't fail to render before we initialize.
const colorSchemeItems = ref<{ id: string; label: string }[]>([])

async function loadColorSchemes() {
  try {
    const { GetColorSchemes } = await import('~/lib/wailsjs/coordinator')
    const schemes = (await GetColorSchemes()) as Record<string, { name?: string; displayName?: string }>
    colorSchemeItems.value = Object.entries(schemes).map(([id, s]) => ({
      id,
      label: s.displayName || s.name || id,
    }))
  } catch (err) {
    console.warn('[Settings] failed to load color schemes:', err)
    colorSchemeItems.value = []
  }
}

// ---- System tab — startup-paths picker ----
//
// We display the actual home + cwd paths next to the radio so the user
// knows what they're picking. Loaded lazily from coordinator.GetStartupPaths.
const initialCwdItems = [
  { id: 'home', label: 'Home directory' },
  { id: 'cwd',  label: 'Launch CWD' },
] as const

const startupPaths = ref<{ home: string; cwd: string }>({ home: '', cwd: '' })

async function loadStartupPaths() {
  try {
    const { GetStartupPaths } = await import('~/lib/wailsjs/coordinator')
    startupPaths.value = await GetStartupPaths()
  } catch (err) {
    console.warn('[Settings] failed to load startup paths:', err)
  }
}

async function onColorSchemeApply(id: string | null | undefined) {
  if (!id) return
  try {
    const { SetDefaultColorScheme } = await import('~/lib/wailsjs/coordinator')
    await SetDefaultColorScheme(id)
  } catch (err) {
    console.warn('[Settings] failed to set default color scheme:', err)
  }
}

function getDefaultSettings(): SettingsData {
  return {
    shell: {
      path: '/bin/bash',
      args: '--login',
      workingDirectory: '/',
    },
    display: {
      theme: 'default',
      keyboardLayout: 'en-US',
      terminalFontSize: 14,
      fontFamily: "'Fira Code', monospace",
      uiFontFamily: '',
      // Empty string = follow the theme's bundled layout. User-picked
      // values override that and persist via Settings → Theme → Layout.
      layout: '',
      // Matches the Go default in models.DefaultUISettings.
      globeStyle: 'classic',
    },
    audio: {
      // Audio defaults to ON so first-launch users get the boot fanfare
      // and key-press feedback. Previously this was `false` to be
      // conservative, but the side effect was that anyone who opened
      // Settings even once had silent boot/shutdown splashes forever
      // (since the modal's saved snapshot wrote enabled:false to the
      // same storage key the splashes read). Flip to true and let
      // users opt out via the Audio tab.
      enabled: true,
      volume: 50,
      // Default to the bundled WAV pack. We renamed the id from 'edex'
      // to 'adex' when we consolidated the asset folders — see
      // useAdexAudio's loadSettings() migration for the old-value
      // patch. New users get 'adex' here directly.
      soundpack: 'adex',
      muteInBackground: false,
      boot: true,
      shutdown: true,
      categoryKeyboard: true,
      categoryDestructive: true,
      categoryInterface: true,
    },
    system: {
      initialCwd: 'home',
      clockFormat: '24h',
      bootAnimation: true,
      gridBackground: true,
      performanceMode: false,
    },
    network: {
      pingTarget: '8.8.8.8:53',
      adapter: '',
      pingIntervalMs: 2000,
      trafficIntervalMs: 1000,
      geoipEnabled: true,
      geoipEndpoint: '',
    },
    advanced: {
      allowWindowedMode: false,
      experimentalFeatures: false,
      webglRenderer: true,
      debugMode: false,
      scrollback: 5000,
      nointro: false,
      forceFullscreen: false,
      disableGpu: false,
      notifications: true,
      frameless: false,
      hideDotfiles: true,
      fsListView: false,
      experimentalGlobeFeatures: false,
    },
  }
}

// Working form state — `reactive` so the USwitch/USelectMenu v-models
// fire synchronously without piping through the persisted ref on every
// keystroke (that would write to localStorage 30+ times during a single
// volume slider drag).
const localSettings = reactive<SettingsData>(getDefaultSettings())

// ---- Persistence ----
//
// Single source of truth on disk. VueUse's useStorage gives us:
//   - Automatic JSON (de)serialization
//   - mergeDefaults: missing keys from older saved versions get filled
//     in from getDefaultSettings(), so adding a new field doesn't break
//     existing users
//   - Cross-tab/cross-component reactivity: any other component that
//     reads `useStorage('adex-settings', …)` sees the same data
//   - SSR-safe (it short-circuits when window is undefined)
//
// We treat this ref as a transactional store: `save()` writes the
// current form snapshot to it, `cancel()` reads back from it. The form
// itself binds to `localSettings`, not this ref — so an unsaved edit
// doesn't leak into other panels until the user clicks Save.
const persistedSettings = useStorage<SettingsData>(
  STORAGE_KEY,
  getDefaultSettings(),
  undefined,
  { mergeDefaults: true },
)

// Migrate legacy soundpack ids in place — older builds wrote 'edex' or
// the leftover 'default' string into adex-settings.audio.soundpack.
// Neither matches an item in audioPackItems any more, so USelectMenu
// would render bare-text instead of the styled trigger. Patch on load.
if (persistedSettings.value?.audio?.soundpack &&
    persistedSettings.value.audio.soundpack !== 'adex' &&
    persistedSettings.value.audio.soundpack !== 'synth') {
  persistedSettings.value = {
    ...persistedSettings.value,
    audio: { ...persistedSettings.value.audio, soundpack: 'adex' },
  }
}

function loadFromStorage(): SettingsData {
  // Deep clone so callers can Object.assign without aliasing the
  // persisted ref's nested objects.
  return JSON.parse(JSON.stringify(persistedSettings.value))
}

function saveToStorage(settings: SettingsData) {
  persistedSettings.value = settings
}

// ---- Actions ----

function save() {
  const snapshot: SettingsData = JSON.parse(JSON.stringify(localSettings))
  saveToStorage(snapshot)
  // Push shell override even if the input never blurred — Save shouldn't
  // silently keep using the previous backend value.
  void onShellPathChange()
  emit('settings-changed', snapshot)
  closeModal()
}

function cancel() {
  // Revert to stored settings
  const stored = loadFromStorage()
  Object.assign(localSettings, stored)
  closeModal()
}

function resetToDefaults() {
  const defaults = getDefaultSettings()
  Object.assign(localSettings, defaults)
}

function closeModal() {
  emit('update:modelValue', false)
  emit('close')
}

function handleOverlayClick() {
  cancel()
}

// ---- Load on open ----

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      const stored = loadFromStorage()
      Object.assign(localSettings, stored)
      // Default to the Terminal tab on open. Previously this was 'shell'
      // (a stale ID from before the rename) — no v-show matched, so the
      // right pane stayed blank until the user clicked any sidebar item.
      activeCategory.value = 'terminal'
      // 'panels' cue — the modal-open whoosh.
      try { audioEngine.playCue('panels') } catch { /* non-fatal */ }
    }
  }
)

// ---- Init ----

onMounted(async () => {
  const stored = loadFromStorage()
  Object.assign(localSettings, stored)

  // Make sure the V2 engines have their indices loaded so the
  // USelectMenu lists aren't empty on first open. Both `initialize()`
  // calls early-return when an active theme/layout already exists.
  await Promise.allSettled([
    customLayoutEngine.load(),
    refreshLockState(),
    themeEngine.initialize(),
    kbEngine.initialize(),
    loadColorSchemes(),
    loadStartupPaths(),
  ])

  // Reflect the live runtime selections into the local form state so the
  // selectors don't show stale defaults from a prior session.
  if (themeEngine.activeId.value) {
    localSettings.display.theme = themeEngine.activeId.value
  }
  if (kbEngine.activeId.value) {
    localSettings.display.keyboardLayout = kbEngine.activeId.value
  }
})
</script>

<style scoped>
/* Overlay */
.settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100dvw;
  height: 100dvh;
  background: rgba(0, 0, 0, 0.75);
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Modal container */
.settings-modal {
  width: 58vw;
  max-height: 78vh;
  background: var(--color_light_black, #05080d);
  border: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  display: flex;
  flex-direction: column;
  position: relative;
  overflow: hidden;
  /* Sci-fi corner cuts */
  clip-path: polygon(
    0% 1.5vh,
    1.5vh 0%,
    calc(100% - 3vh) 0%,
    100% 3vh,
    100% calc(100% - 1.5vh),
    calc(100% - 1.5vh) 100%,
    3vh 100%,
    0% calc(100% - 3vh)
  );
}

/* Outer glow border effect */
.settings-modal::before {
  content: '';
  position: absolute;
  top: -0.15vh;
  left: -0.15vh;
  right: -0.15vh;
  bottom: -0.15vh;
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  clip-path: polygon(
    0% 1.5vh,
    1.5vh 0%,
    calc(100% - 3vh) 0%,
    100% 3vh,
    100% calc(100% - 1.5vh),
    calc(100% - 1.5vh) 100%,
    3vh 100%,
    0% calc(100% - 3vh)
  );
  pointer-events: none;
  z-index: -1;
}

/* Header */
.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 4vh;
  padding: 0 2vw;
  border-bottom: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  flex-shrink: 0;
}

.settings-header-title {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.3vh;
  color: var(--color_accent, rgb(170, 207, 209));
  text-transform: uppercase;
  letter-spacing: 0.3em;
}

.settings-header-bracket {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
}

.settings-close-btn {
  background: none;
  border: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  width: 2.5vh;
  height: 2.5vh;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.settings-close-btn:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Body */
.settings-body {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* Sidebar */
.settings-sidebar {
  width: 12vw;
  border-right: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  padding: 1vh 0;
  flex-shrink: 0;
  overflow-y: auto;
}

.settings-sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.5vw;
  padding: 0.8vh 1.2vw;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
  text-transform: uppercase;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  user-select: none;
}

.settings-sidebar-item:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.08);
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-sidebar-item.active {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  color: var(--color_accent, rgb(170, 207, 209));
}

.sidebar-marker {
  font-size: 1.2vh;
  width: 1vw;
  flex-shrink: 0;
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Content */
.settings-content {
  flex: 1;
  padding: 1.5vh 2vw;
  overflow-y: auto;
  min-height: 0;
}

.settings-section-title {
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.2vh;
  color: var(--color_accent, rgb(170, 207, 209));
  text-transform: uppercase;
  letter-spacing: 0.2em;
  padding-bottom: 0.8vh;
  margin-bottom: 1.5vh;
  border-bottom: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
}

/* Inline help text below a settings field. Uses the canonical `--accent`
 * with reduced opacity so it doesn't fight the active theme. */
.settings-hint {
  margin: 0.5vh 0 0;
  font-size: 1.05vh;
  line-height: 1.5;
  color: var(--accent, rgb(170, 207, 209));
  opacity: 0.55;
  letter-spacing: 0.04em;
}

.settings-hint strong {
  opacity: 1;
  font-weight: normal;
  color: var(--accent, rgb(170, 207, 209));
}

/* Fields */
.settings-field {
  margin-bottom: 1.5vh;
}

.settings-field-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.settings-label {
  display: block;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.05vh;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.7);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin-bottom: 0.5vh;
}

.settings-field-toggle .settings-label {
  margin-bottom: 0;
}

/* Input */
.settings-input {
  width: 100%;
  height: 3vh;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  padding: 0 0.8vw;
  outline: none;
  transition: border-color 0.15s, background 0.15s;
}

.settings-input:focus {
  border-color: var(--color_accent, rgb(170, 207, 209));
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.08);
}

.settings-input::placeholder {
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.25);
}

/* Read-only value rows — used by the Updates tab to display current
 * version, latest release tag, last-checked timestamp, etc. without
 * looking like an editable field. */
.settings-readonly {
  width: 100%;
  padding: 0.4vh 0.6vw;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.04);
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.05vh;
  display: flex;
  align-items: center;
  gap: 0.6vw;
}

.settings-readonly-err {
  color: var(--err, #ef4444);
  border-color: rgba(239, 68, 68, 0.35);
}

.settings-readonly-tag {
  font-size: 0.85em;
  letter-spacing: 0.15em;
  padding: 0.05vh 0.4vw;
  border: 1px solid currentColor;
  opacity: 0.7;
}

.settings-readonly-tag-new {
  color: var(--ok, #10b981);
  opacity: 1;
}

/* Keybind list (Settings -> Keybinds) */
.keybind-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.8rem;
}

.keybind-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.35rem 0.6rem;
  border: var(--border_width, 1px) solid var(--color_accent_dimmed, rgba(170, 207, 209, 0.2));
  background: var(--color_light_black, #05080d);
}

.keybind-action {
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 0.78rem;
}

.keybind-keys {
  padding: 0.15rem 0.55rem;
  border: var(--border_width, 1px) solid var(--color_accent_dimmed, rgba(170, 207, 209, 0.35));
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 0.72rem;
  white-space: nowrap;
}

/* Module list (Settings -> Modules) */
.module-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.module-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.3rem 0.5rem;
  border: var(--border_width, 1px) solid var(--color_accent_dimmed, rgba(170, 207, 209, 0.25));
  background: var(--color_light_black, #05080d);
}

.module-row.module-hidden .module-name { opacity: 0.4; }

.module-name {
  flex: 1 1 auto;
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, monospace);
  font-size: 0.78rem;
}

.module-move {
  flex: 0 0 auto;
  padding: 0.1rem 0.45rem;
  background: transparent;
  border: var(--border_width, 1px) solid var(--color_accent_dimmed, rgba(170, 207, 209, 0.3));
  color: var(--color_accent, rgb(170, 207, 209));
  font-size: 0.7rem;
  cursor: pointer;
}
.module-move:hover:not(:disabled) { background: var(--color_accent_glow, rgba(170, 207, 209, 0.15)); }
.module-move:disabled { opacity: 0.25; cursor: not-allowed; }

/* Drag + resize affordances */
.settings-header {
  cursor: move;
  user-select: none;
  touch-action: none;
}

.settings-resize-grip {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 18px;
  height: 18px;
  cursor: nwse-resize;
  touch-action: none;
  z-index: 3;
  /* Two short strokes, the conventional resize-corner hint. */
  background:
    linear-gradient(135deg, transparent 0 45%,
      var(--color_accent, rgb(170, 207, 209)) 45% 55%, transparent 55% 100%),
    linear-gradient(135deg, transparent 0 70%,
      var(--color_accent, rgb(170, 207, 209)) 70% 80%, transparent 80% 100%);
  opacity: 0.45;
}
.settings-resize-grip:hover { opacity: 0.9; }

/* Select */
.settings-select {
  width: 100%;
  height: 3vh;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.05);
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  color: var(--color_accent, rgb(170, 207, 209));
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.1vh;
  padding: 0 0.8vw;
  outline: none;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23aacfd1' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.6vw center;
  background-size: 1vh;
  transition: border-color 0.15s;
}

.settings-select:focus {
  border-color: var(--color_accent, rgb(170, 207, 209));
}

.settings-select:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.settings-select option {
  background: var(--color_light_black, #05080d);
  color: var(--color_accent, rgb(170, 207, 209));
}

/* Slider */
.settings-slider {
  width: 100%;
  height: 0.5vh;
  -webkit-appearance: none;
  appearance: none;
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  outline: none;
  border-radius: 0;
  margin-top: 0.5vh;
}

.settings-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 1.2vh;
  height: 2vh;
  background: var(--color_accent, rgb(170, 207, 209));
  cursor: pointer;
  border: none;
}

.settings-slider::-moz-range-thumb {
  width: 1.2vh;
  height: 2vh;
  background: var(--color_accent, rgb(170, 207, 209));
  cursor: pointer;
  border: none;
  border-radius: 0;
}

.settings-slider:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.settings-slider:disabled::-webkit-slider-thumb {
  cursor: not-allowed;
}

/* Footer */
.settings-footer {
  display: flex;
  align-items: center;
  gap: 1vw;
  padding: 1vh 2vw;
  border-top: var(--border_width, 0.092vh) solid var(--border_color, rgba(170, 207, 209, 0.5));
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.03);
  flex-shrink: 0;
}

.settings-footer-spacer {
  flex: 1;
}

.settings-btn {
  height: 3vh;
  padding: 0 1.5vw;
  font-family: var(--font_main, 'Fira Code', monospace);
  font-size: 1.05vh;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: all 0.15s;
  border: var(--border_width, 0.092vh) solid rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
  background: transparent;
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.7);
}

.settings-btn:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.1);
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-btn-primary {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.15);
  border-color: var(--color_accent, rgb(170, 207, 209));
  color: var(--color_accent, rgb(170, 207, 209));
}

.settings-btn-primary:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.3);
}

.settings-btn-secondary {
  border-color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.5);
}

.settings-btn-secondary:hover {
  border-color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
  color: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.8);
}

/* Modal transition */
.settings-modal-enter-active {
  transition: all 0.3s cubic-bezier(0.19, 1, 0.22, 1);
}

.settings-modal-leave-active {
  transition: all 0.2s ease-in;
}

.settings-modal-enter-from {
  opacity: 0;
}

.settings-modal-enter-from .settings-modal {
  transform: scale(0.95) translateY(1vh);
  opacity: 0;
}

.settings-modal-leave-to {
  opacity: 0;
}

.settings-modal-leave-to .settings-modal {
  transform: scale(0.95) translateY(1vh);
  opacity: 0;
}

/* Scrollbar in settings */
.settings-content::-webkit-scrollbar,
.settings-sidebar::-webkit-scrollbar {
  width: 0.3vw;
}

.settings-content::-webkit-scrollbar-track,
.settings-sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.settings-content::-webkit-scrollbar-thumb,
.settings-sidebar::-webkit-scrollbar-thumb {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.2);
}

.settings-content::-webkit-scrollbar-thumb:hover,
.settings-sidebar::-webkit-scrollbar-thumb:hover {
  background: rgba(var(--color_r, 170), var(--color_g, 207), var(--color_b, 209), 0.4);
}
</style>
