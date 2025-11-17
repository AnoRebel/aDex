# aDex-UI Audio System

The aDex-UI audio system provides comprehensive sound effects support for the application, with multiple soundpacks, flexible configuration, and easy integration with Vue components.

## Features

- **Multiple Soundpacks**: Default, Minimal, Retro, and Cyberpunk themes
- **Disabled by Default**: Audio is opt-in to respect user preferences
- **Event-Driven System**: Maps UI events to appropriate sound effects
- **Cooldown and Priority**: Prevents audio spam and handles conflicts
- **Lazy Loading**: Audio files are loaded on-demand
- **Volume Controls**: Master, category, and per-event volume control
- **Web Audio API Support**: Modern audio playback with fallbacks
- **Accessibility**: Respects reduced motion preferences
- **Vue Integration**: Directives and composables for easy use

## Quick Start

### 1. Install the Audio Plugin

```typescript
// main.ts
import { createApp } from 'vue'
import App from './app.vue'
import audioPlugin from './plugins/audio'

const app = createApp(App)

app.use(audioPlugin({
  enableDebug: process.env.NODE_ENV === 'development',
  respectReducedMotion: true
}))

app.mount('#app')
```

### 2. Enable Audio in Your App

```vue
<script setup>
import { useAudio } from '~/composables/useAudio'

const audio = useAudio()

// Enable audio system
await audio.toggleEnabled()

// Play a sound effect
await audio.playEvent('button_click')
</script>
```

### 3. Use Audio Directives in Templates

```vue
<template>
  <!-- Click sound -->
  <button v-audio-click="'ui:button_click'">
    Click me!
  </button>

  <!-- Hover sound -->
  <div v-audio-hover="'ui:hover'">
    Hover over me
  </div>

  <!-- Custom event sound -->
  <input
    v-audio-sound:focus="'ui:focus'"
    v-audio-sound:blur="'ui:blur'"
    type="text"
  />
</template>
```

## Audio Events

### System Events
- `system:startup` - Application startup
- `system:shutdown` - Application shutdown
- `system:alert` - Critical system alert

### UI Interactions
- `ui:button_click` - Button clicked
- `ui:menu_open` - Menu opened
- `ui:menu_close` - Menu closed
- `ui:focus` - Element gained focus
- `ui:toggle_on` - Toggle activated
- `ui:toggle_off` - Toggle deactivated

### Terminal Events
- `terminal:bell` - Terminal bell character
- `terminal:command_success` - Command completed successfully
- `terminal:command_error` - Command failed

### File Operations
- `file:operation_start` - File operation started
- `file:operation_complete` - File operation completed
- `file:operation_error` - File operation failed

### Network Events
- `network:connected` - Network connection established
- `network:disconnected` - Network connection lost

### Notifications
- `ui:notification` - General notification
- `ui:notification_success` - Success notification
- `ui:notification_warning` - Warning notification
- `ui:notification_error` - Error notification

## Using the Audio Composable

```vue
<script setup>
import { useAudio } from '~/composables/useAudio'

const audio = useAudio()

// Enable/disable audio
const toggleAudio = async () => {
  await audio.toggleEnabled()
}

// Volume controls
const setMasterVolume = async (volume: number) => {
  await audio.setVolume(volume)
}

const setEffectsVolume = async (volume: number) => {
  await audio.setEffectsVolume(volume)
}

// Soundpack management
const changeSoundpack = async (soundpackId: string) => {
  await audio.changeSoundpack(soundpackId)
}

// Play specific events
const playNotification = async () => {
  await audio.playEvent('ui:notification')
}

// Play UI event sound (maps to appropriate audio event)
const playButtonSound = async () => {
  await audio.playUIEvent('ui:button_click')
}

// Check if event is enabled
const canPlaySound = audio.isEventEnabled('button_click')

// Get volume icon for UI
const volumeIcon = audio.getVolumeIcon()

// Access reactive state
const {
  isEnabled,
  effectiveVolume,
  currentSoundpack,
  enabledEvents,
  isLoading
} = audio
</script>
```

## Using the Audio Store

```vue
<script setup>
import { useAudioStore } from '~/stores/audio'

const audioStore = useAudioStore()

// Enable/disable audio
audioStore.toggleEnabled()

// Volume controls
audioStore.setVolume(75)
audioStore.setGlobalVolume(0.8)
audioStore.setEffectsVolume(0.6)

// Event management
audioStore.enableEvent('button_click')
audioStore.disableEvent('terminal_bell')
audioStore.enableAllEvents()

// Soundpack management
audioStore.changeSoundpack('cyberpunk')

// UI state
audioStore.setShowSettings(true)
audioStore.setShowEventTester(false)

// Access computed properties
const {
  isEnabled,
  effectiveVolume,
  eventsByCategory,
  currentSoundpack,
  volumeIcon
} = audioStore
</script>
```

## Audio Settings Component

Here's a complete example of an audio settings panel:

```vue
<template>
  <div class="audio-settings">
    <h2>Audio Settings</h2>

    <!-- Master toggle -->
    <div class="setting-row">
      <label>
        <input
          type="checkbox"
          v-model="audioStore.settings.enabled"
          @change="audioStore.saveToStorage()"
        />
        Enable Sound Effects
      </label>
    </div>

    <!-- Master volume -->
    <div class="setting-row">
      <label>Master Volume: {{ audioStore.settings.volume }}%</label>
      <input
        type="range"
        min="0"
        max="100"
        v-model="audioStore.settings.volume"
        @input="audioStore.setVolume($event.target.value)"
      />
    </div>

    <!-- Category volumes -->
    <div class="setting-row">
      <label>Effects Volume: {{ Math.round(audioStore.settings.effectsVolume * 100) }}%</label>
      <input
        type="range"
        min="0"
        max="1"
        step="0.1"
        v-model="audioStore.settings.effectsVolume"
        @input="audioStore.setEffectsVolume($event.target.value)"
      />
    </div>

    <!-- Soundpack selection -->
    <div class="setting-row">
      <label>Soundpack:</label>
      <select
        v-model="audioStore.settings.soundpack"
        @change="audioStore.changeSoundpack($event.target.value)"
      >
        <option
          v-for="pack in audioStore.builtinSoundpacks"
          :key="pack.id"
          :value="pack.id"
        >
          {{ pack.displayName }}
        </option>
      </select>
    </div>

    <!-- Event toggles by category -->
    <div
      v-for="(events, category) in audioStore.eventsByCategory"
      :key="category"
      class="category-section"
    >
      <h3>{{ category.charAt(0).toUpperCase() + category.slice(1) }}</h3>

      <div class="event-list">
        <label
          v-for="event in events"
          :key="event.id"
          class="event-item"
        >
          <input
            type="checkbox"
            :checked="audioStore.isEventEnabled(event.id)"
            @change="audioStore.toggleEvent(event.id)"
          />
          {{ event.name }}

          <!-- Test button -->
          <button
            @click="audio.playEvent(event.id)"
            :disabled="!audioStore.isEnabled"
            class="test-button"
          >
            Test
          </button>
        </label>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useAudioStore } from '~/stores/audio'
import { useAudio } from '~/composables/useAudio'

const audioStore = useAudioStore()
const audio = useAudio()

// Initialize store on mount
audioStore.initialize()
</script>

<style scoped>
.audio-settings {
  padding: 1rem;
  max-width: 600px;
}

.setting-row {
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.setting-row label {
  min-width: 150px;
}

.setting-row input[type="range"] {
  flex: 1;
}

.category-section {
  margin-top: 2rem;
}

.event-list {
  display: grid;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.event-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.test-button {
  margin-left: auto;
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
}
</style>
```

## Programmatic Audio Triggers

```typescript
import { audioTriggers, playAudioEvent } from '~/plugins/audio'

// Use predefined triggers
await audioTriggers.click()
await audioTriggers.success()
await audioTriggers.fileComplete()

// Use debounced triggers for rapid events
audioTriggers.debouncedKeyPress()
audioTriggers.debouncedSliderChange()

// Trigger custom events
await playAudioEvent('custom:event', {
  source: 'my-component',
  volume: 75
})
```

## Audio File Format

Audio files should be placed in `/assets/audio/{soundpack}/` directories:

```
assets/audio/
├── default/
│   ├── button_click.wav
│   ├── notification.wav
│   └── ...
├── minimal/
│   ├── soft_bell.wav
│   └── ...
├── retro/
│   ├── beep.wav
│   └── ...
└── cyberpunk/
    ├── digital_ping.wav
    └── ...
```

### Recommended Audio Specs
- **Format**: WAV, MP3, OGG, or AAC
- **Sample Rate**: 44.1kHz
- **Bit Depth**: 16-bit
- **Channels**: Mono (for UI sounds) or Stereo
- **Duration**: 50ms - 2s (keep UI sounds short)

## Custom Soundpacks

You can create custom soundpacks by adding directories and updating the soundpacks.json configuration:

```json
{
  "custom_pack": {
    "id": "custom_pack",
    "name": "Custom Pack",
    "description": "My custom sound effects",
    "version": "1.0.0",
    "author": "Your Name",
    "isBuiltIn": false,
    "events": {
      "custom_sound": {
        "id": "custom_sound",
        "name": "Custom Sound",
        "description": "My custom sound effect",
        "category": "interaction",
        "filePath": "/audio/custom_pack/custom_sound.wav",
        "duration": 200,
        "volume": 50
      }
    }
  }
}
```

## Browser Compatibility

- **Web Audio API**: Modern browsers (Chrome, Firefox, Safari, Edge)
- **HTML5 Audio**: All browsers (fallback)
- ** autoplay restrictions**: Some browsers require user interaction before playing audio

## Performance Considerations

- Audio files are loaded lazily on first use
- Caching prevents repeated network requests
- Cooldown system prevents audio spam
- Reduced motion support for accessibility
- Debounced triggers for rapid events

## Troubleshooting

### Audio not playing:
1. Check if audio is enabled: `audio.isEnabled`
2. Check browser autoplay permissions
3. Ensure audio files exist and are accessible
4. Check browser console for errors

### Audio files not loading:
1. Verify file paths in soundpacks.json
2. Check network requests in browser dev tools
3. Ensure audio files are in correct format
4. Check server MIME types for audio files

### Too many audio events:
1. Increase cooldown values in event mappings
2. Use debounced triggers for rapid events
3. Disable unnecessary event categories
4. Enable reduced motion respect option