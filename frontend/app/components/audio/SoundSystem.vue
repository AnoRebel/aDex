<template>
  <div class="sound-system">
    <div class="sound-header">
      <h3 class="glitch">AUDIO SYSTEM</h3>
      <div class="sound-controls">
        <button @click="toggleSoundPanel" class="control-btn">
          <span class="btn-icon">🔊</span>
        </button>
        <button @click="toggleMute" class="control-btn" :class="{ muted: isMuted }">
          <span class="btn-icon">{{ isMuted ? '🔇' : '🔊' }}</span>
        </button>
      </div>
    </div>

    <!-- Sound Panel -->
    <div v-if="showSoundPanel" class="sound-panel">
      <!-- Master Volume -->
      <div class="volume-control">
        <label class="volume-label">
          <span class="label-icon">🎚️</span>
          <span class="label-text">Master Volume</span>
          <span class="volume-value">{{ Math.round(masterVolume * 100) }}%</span>
        </label>
        <input
          v-model="masterVolume"
          type="range"
          min="0"
          max="1"
          step="0.01"
          @input="updateMasterVolume"
          class="volume-slider"
        />
      </div>

      <!-- Sound Categories -->
      <div class="sound-categories">
        <div
          v-for="category in soundCategories"
          :key="category.id"
          class="sound-category"
        >
          <div class="category-header">
            <span class="category-icon">{{ category.icon }}</span>
            <span class="category-name">{{ category.name }}</span>
            <div class="category-controls">
              <button
                @click="toggleCategory(category.id)"
                class="category-toggle"
                :class="{ active: category.enabled }"
              >
                {{ category.enabled ? 'ON' : 'OFF' }}
              </button>
            </div>
          </div>
          <div class="category-sounds">
            <div
              v-for="sound in category.sounds"
              :key="sound.id"
              class="sound-item"
            >
              <div class="sound-info">
                <span class="sound-name">{{ sound.name }}</span>
                <span class="sound-description">{{ sound.description }}</span>
              </div>
              <div class="sound-controls">
                <input
                  v-model="sound.volume"
                  type="range"
                  min="0"
                  max="1"
                  step="0.01"
                  @input="updateSoundVolume(sound)"
                  class="sound-slider"
                />
                <button
                  @click="playSound(sound.id)"
                  class="play-btn"
                  title="Play sound"
                >
                  ▶
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Audio Visualizer -->
      <div class="audio-visualizer">
        <div class="visualizer-header">
          <span class="visualizer-title">AUDIO VISUALIZER</span>
          <button
            @click="toggleVisualizer"
            class="visualizer-toggle"
            :class="{ active: visualizerActive }"
          >
            {{ visualizerActive ? 'ON' : 'OFF' }}
          </button>
        </div>
        <div class="visualizer-container">
          <canvas
            ref="visualizerCanvas"
            width="400"
            height="100"
            class="visualizer-canvas"
          ></canvas>
        </div>
      </div>

      <!-- Audio Settings -->
      <div class="audio-settings">
        <h4 class="settings-title">AUDIO SETTINGS</h4>
        <div class="settings-grid">
          <label class="setting-toggle">
            <input
              v-model="audioSettings.spatialAudio"
              type="checkbox"
              @change="updateAudioSettings"
            />
            <span class="setting-label">Spatial Audio</span>
          </label>
          <label class="setting-toggle">
            <input
              v-model="audioSettings.dynamicRange"
              type="checkbox"
              @change="updateAudioSettings"
            />
            <span class="setting-label">Dynamic Range</span>
          </label>
          <label class="setting-toggle">
            <input
              v-model="audioSettings.bassBoost"
              type="checkbox"
              @change="updateAudioSettings"
            />
            <span class="setting-label">Bass Boost</span>
          </label>
          <label class="setting-toggle">
            <input
              v-model="audioSettings.reverb"
              type="checkbox"
              @change="updateAudioSettings"
            />
            <span class="setting-label">Reverb Effect</span>
          </label>
        </div>
      </div>
    </div>

    <!-- Mini Player -->
    <div class="mini-player">
      <div class="player-info">
        <span class="now-playing">NOW PLAYING:</span>
        <span class="current-track">{{ currentTrack || 'None' }}</span>
      </div>
      <div class="player-controls">
        <button @click="previousTrack" class="player-btn">
          ⏮
        </button>
        <button @click="togglePlayback" class="player-btn play-pause">
          {{ isPlaying ? '⏸' : '▶' }}
        </button>
        <button @click="nextTrack" class="player-btn">
          ⏭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

interface Sound {
  id: string
  name: string
  description: string
  volume: number
  enabled: boolean
  url?: string
  frequency?: number
  duration?: number
}

interface SoundCategory {
  id: string
  name: string
  icon: string
  enabled: boolean
  sounds: Sound[]
}

interface AudioSettings {
  spatialAudio: boolean
  dynamicRange: boolean
  bassBoost: boolean
  reverb: boolean
}

// Reactive data
const showSoundPanel = ref<boolean>(false)
const isMuted = ref<boolean>(false)
const masterVolume = ref<number>(0.5)
const currentTrack = ref<string>('')
const isPlaying = ref<boolean>(false)
const visualizerActive = ref<boolean>(false)
const visualizerCanvas = ref<HTMLCanvasElement>()

const audioSettings = ref<AudioSettings>({
  spatialAudio: false,
  dynamicRange: true,
  bassBoost: false,
  reverb: false
})

// Sound categories and effects
const soundCategories = ref<SoundCategory[]>([
  {
    id: 'ui',
    name: 'UI Sounds',
    icon: '🖱️',
    enabled: true,
    sounds: [
      {
        id: 'click',
        name: 'Click',
        description: 'Button click sound',
        volume: 0.3,
        enabled: true,
        frequency: 800,
        duration: 0.1
      },
      {
        id: 'hover',
        name: 'Hover',
        description: 'Element hover sound',
        volume: 0.2,
        enabled: true,
        frequency: 600,
        duration: 0.05
      },
      {
        id: 'notification',
        name: 'Notification',
        description: 'Alert notification sound',
        volume: 0.5,
        enabled: true,
        frequency: 1000,
        duration: 0.3
      },
      {
        id: 'error',
        name: 'Error',
        description: 'Error sound effect',
        volume: 0.4,
        enabled: true,
        frequency: 300,
        duration: 0.4
      }
    ]
  },
  {
    id: 'terminal',
    name: 'Terminal',
    icon: '⌨️',
    enabled: true,
    sounds: [
      {
        id: 'keystroke',
        name: 'Keystroke',
        description: 'Typing sound effect',
        volume: 0.3,
        enabled: true,
        frequency: 1200,
        duration: 0.02
      },
      {
        id: 'enter',
        name: 'Enter',
        description: 'Enter key sound',
        volume: 0.5,
        enabled: true,
        frequency: 1500,
        duration: 0.1
      },
      {
        id: 'backspace',
        name: 'Backspace',
        description: 'Backspace sound',
        volume: 0.4,
        enabled: true,
        frequency: 800,
        duration: 0.08
      },
      {
        id: 'bell',
        name: 'Bell',
        description: 'Terminal bell sound',
        volume: 0.6,
        enabled: true,
        frequency: 2000,
        duration: 0.2
      }
    ]
  },
  {
    id: 'system',
    name: 'System',
    icon: '⚙️',
    enabled: true,
    sounds: [
      {
        id: 'startup',
        name: 'Startup',
        description: 'System startup sound',
        volume: 0.7,
        enabled: true,
        frequency: 440,
        duration: 1.5
      },
      {
        id: 'shutdown',
        name: 'Shutdown',
        description: 'System shutdown sound',
        volume: 0.6,
        enabled: true,
        frequency: 220,
        duration: 1.0
      },
      {
        id: 'alert',
        name: 'Alert',
        description: 'System alert sound',
        volume: 0.8,
        enabled: true,
        frequency: 880,
        duration: 0.5
      },
      {
        id: 'success',
        name: 'Success',
        description: 'Success sound effect',
        volume: 0.5,
        enabled: true,
        frequency: 660,
        duration: 0.3
      }
    ]
  },
  {
    id: 'ambient',
    name: 'Ambient',
    icon: '🌌',
    enabled: false,
    sounds: [
      {
        id: 'hum',
        name: 'System Hum',
        description: 'Low system hum',
        volume: 0.1,
        enabled: true,
        frequency: 60,
        duration: 10
      },
      {
        id: 'atmosphere',
        name: 'Atmosphere',
        description: 'Sci-fi atmosphere',
        volume: 0.2,
        enabled: true,
        frequency: 200,
        duration: 15
      }
    ]
  }
])

// Audio context and nodes
let audioContext: AudioContext | null = null
let masterGainNode: GainNode | null = null
let analyser: AnalyserNode | null = null
let visualizerAnimation: number | null = null

// Methods
const initAudioContext = () => {
  if (!audioContext) {
    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
    masterGainNode = audioContext.createGain()
    analyser = audioContext.createAnalyser()

    masterGainNode.connect(analyser)
    analyser.connect(audioContext.destination)

    masterGainNode.gain.value = masterVolume.value
    analyser.fftSize = 256
  }
}

const toggleSoundPanel = () => {
  showSoundPanel.value = !showSoundPanel.value
  if (showSoundPanel.value) {
    initAudioContext()
  }
}

const toggleMute = () => {
  isMuted.value = !isMuted.value
  if (masterGainNode) {
    masterGainNode.gain.value = isMuted.value ? 0 : masterVolume.value
  }
}

const updateMasterVolume = () => {
  if (masterGainNode && !isMuted.value) {
    masterGainNode.gain.value = masterVolume.value
  }
  localStorage.setItem('masterVolume', masterVolume.value.toString())
}

const toggleCategory = (categoryId: string) => {
  const category = soundCategories.value.find(c => c.id === categoryId)
  if (category) {
    category.enabled = !category.enabled
    saveSoundSettings()
  }
}

const updateSoundVolume = (sound: Sound) => {
  saveSoundSettings()
}

const playSound = async (soundId: string) => {
  if (!audioContext || isMuted.value) return

  const sound = findSound(soundId)
  if (!sound || !sound.enabled) return

  try {
    const oscillator = audioContext.createOscillator()
    const gainNode = audioContext.createGain()

    // Apply sound settings
    oscillator.connect(gainNode)
    gainNode.connect(masterGainNode!)

    // Set frequency and type
    if (sound.frequency) {
      oscillator.frequency.value = sound.frequency
    }
    oscillator.type = 'sine'

    // Apply effects
    let finalVolume = sound.volume * masterVolume.value

    if (audioSettings.value.dynamicRange) {
      // Add some variation to make it more dynamic
      finalVolume *= (0.8 + Math.random() * 0.4)
    }

    if (audioSettings.value.bassBoost && sound.frequency && sound.frequency < 200) {
      finalVolume *= 1.5
    }

    gainNode.gain.setValueAtTime(0, audioContext.currentTime)
    gainNode.gain.linearRampToValueAtTime(finalVolume, audioContext.currentTime + 0.01)

    const fadeOutTime = sound.duration || 0.1
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + fadeOutTime)

    // Start and stop oscillator
    oscillator.start(audioContext.currentTime)
    oscillator.stop(audioContext.currentTime + fadeOutTime)

    // Update current track
    currentTrack.value = sound.name
    isPlaying.value = true

    setTimeout(() => {
      isPlaying.value = false
      currentTrack.value = ''
    }, fadeOutTime * 1000)

    // Trigger visualizer
    if (visualizerActive.value) {
      triggerVisualizer()
    }

  } catch (error) {
    console.error('Error playing sound:', error)
  }
}

const findSound = (soundId: string): Sound | null => {
  for (const category of soundCategories.value) {
    const sound = category.sounds.find(s => s.id === soundId)
    if (sound) return sound
  }
  return null
}

// Visualizer
const toggleVisualizer = () => {
  visualizerActive.value = !visualizerActive.value
  if (visualizerActive.value) {
    startVisualizer()
  } else {
    stopVisualizer()
  }
}

const startVisualizer = () => {
  if (!analyser || !visualizerCanvas.value) return

  const canvas = visualizerCanvas.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const bufferLength = analyser.frequencyBinCount
  const dataArray = new Uint8Array(bufferLength)

  const draw = () => {
    if (!visualizerActive.value) return

    visualizerAnimation = requestAnimationFrame(draw)

    analyser.getByteFrequencyData(dataArray)

    // Clear canvas
    ctx.fillStyle = 'rgba(0, 0, 0, 0.2)'
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    // Draw frequency bars
    const barWidth = (canvas.width / bufferLength) * 2.5
    let barHeight
    let x = 0

    for (let i = 0; i < bufferLength; i++) {
      barHeight = (dataArray[i] / 255) * canvas.height

      // Create gradient
      const gradient = ctx.createLinearGradient(0, canvas.height - barHeight, 0, canvas.height)
      gradient.addColorStop(0, '#0ea5e9')
      gradient.addColorStop(1, '#f59e0b')

      ctx.fillStyle = gradient
      ctx.fillRect(x, canvas.height - barHeight, barWidth, barHeight)

      x += barWidth + 1
    }
  }

  draw()
}

const stopVisualizer = () => {
  if (visualizerAnimation) {
    cancelAnimationFrame(visualizerAnimation)
    visualizerAnimation = null
  }

  // Clear canvas
  if (visualizerCanvas.value) {
    const ctx = visualizerCanvas.value.getContext('2d')
    if (ctx) {
      ctx.clearRect(0, 0, visualizerCanvas.value.width, visualizerCanvas.value.height)
    }
  }
}

const triggerVisualizer = () => {
  // Trigger a brief visualizer animation when sound plays
  if (visualizerActive.value && visualizerCanvas.value) {
    const ctx = visualizerCanvas.value.getContext('2d')
    if (ctx) {
      // Create a pulse effect
      const centerX = visualizerCanvas.value.width / 2
      const centerY = visualizerCanvas.value.height / 2
      const maxRadius = Math.max(visualizerCanvas.value.width, visualizerCanvas.value.height) / 2

      let radius = 0
      const animate = () => {
        if (radius > maxRadius) return

        ctx.strokeStyle = `rgba(14, 165, 233, ${1 - radius / maxRadius})`
        ctx.lineWidth = 2
        ctx.beginPath()
        ctx.arc(centerX, centerY, radius, 0, 2 * Math.PI)
        ctx.stroke()

        radius += 5
        requestAnimationFrame(animate)
      }

      animate()
    }
  }
}

// Mini player controls
const togglePlayback = () => {
  // Simple toggle for demo
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) {
    playSound('startup')
  }
}

const previousTrack = () => {
  playSound('click')
}

const nextTrack = () => {
  playSound('click')
}

// Settings
const updateAudioSettings = () => {
  localStorage.setItem('audioSettings', JSON.stringify(audioSettings.value))
}

const saveSoundSettings = () => {
  const settings = {
    categories: soundCategories.value.map(c => ({
      id: c.id,
      enabled: c.enabled,
      sounds: c.sounds.map(s => ({
        id: s.id,
        volume: s.volume,
        enabled: s.enabled
      }))
    }))
  }
  localStorage.setItem('soundSettings', JSON.stringify(settings))
}

const loadSoundSettings = () => {
  // Load master volume
  const savedVolume = localStorage.getItem('masterVolume')
  if (savedVolume) {
    masterVolume.value = parseFloat(savedVolume)
  }

  // Load audio settings
  const savedAudioSettings = localStorage.getItem('audioSettings')
  if (savedAudioSettings) {
    try {
      audioSettings.value = { ...audioSettings.value, ...JSON.parse(savedAudioSettings) }
    } catch (error) {
      console.error('Failed to load audio settings:', error)
    }
  }

  // Load sound settings
  const savedSoundSettings = localStorage.getItem('soundSettings')
  if (savedSoundSettings) {
    try {
      const settings = JSON.parse(savedSoundSettings)
      settings.categories?.forEach((savedCategory: any) => {
        const category = soundCategories.value.find(c => c.id === savedCategory.id)
        if (category) {
          category.enabled = savedCategory.enabled
          savedCategory.sounds?.forEach((savedSound: any) => {
            const sound = category.sounds.find(s => s.id === savedSound.id)
            if (sound) {
              sound.volume = savedSound.volume
              sound.enabled = savedSound.enabled
            }
          })
        }
      })
    } catch (error) {
      console.error('Failed to load sound settings:', error)
    }
  }
}

// Public methods for external components
const playUISound = (soundName: string) => {
  playSound(soundName)
}

const playKeystroke = () => {
  playSound('keystroke')
}

const playError = () => {
  playSound('error')
}

const playSuccess = () => {
  playSound('success')
}

defineExpose({
  playUISound,
  playKeystroke,
  playError,
  playSuccess
})

// Lifecycle
onMounted(() => {
  loadSoundSettings()
})

onUnmounted(() => {
  stopVisualizer()
  if (audioContext) {
    audioContext.close()
  }
})
</script>

<style scoped>
.sound-system {
  background: rgba(10, 10, 10, 0.95);
  border: 1px solid var(--surface-border);
  border-radius: 12px;
  padding: 20px;
  backdrop-filter: blur(10px);
  font-family: 'Fira Code', monospace;
  color: var(--text-primary);
}

.sound-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid var(--surface-border);
}

.sound-controls {
  display: flex;
  gap: 8px;
}

.control-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.control-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.control-btn.muted {
  border-color: var(--error);
  color: var(--error);
}

.btn-icon {
  font-size: 16px;
}

.sound-panel {
  margin-bottom: 20px;
}

.volume-control {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 20px;
}

.volume-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--primary-400);
}

.label-icon {
  margin-right: 8px;
}

.volume-value {
  color: var(--accent-400);
}

.volume-slider {
  width: 100%;
  height: 6px;
  background: var(--surface-elevated);
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
}

.volume-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  background: var(--primary-500);
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.5);
}

.volume-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  background: var(--primary-500);
  border-radius: 50%;
  cursor: pointer;
  border: none;
  box-shadow: 0 0 10px rgba(14, 165, 233, 0.5);
}

.sound-categories {
  display: grid;
  gap: 15px;
  margin-bottom: 20px;
}

.sound-category {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  overflow: hidden;
}

.category-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 15px;
  background: var(--surface-elevated);
  border-bottom: 1px solid var(--surface-border);
}

.category-icon {
  font-size: 18px;
  margin-right: 8px;
}

.category-name {
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--primary-400);
  flex: 1;
}

.category-toggle {
  padding: 4px 8px;
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  font-size: 9px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s ease;
}

.category-toggle.active {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

.category-sounds {
  padding: 15px;
}

.sound-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--surface-elevated);
}

.sound-item:last-child {
  border-bottom: none;
}

.sound-info {
  flex: 1;
}

.sound-name {
  font-size: 11px;
  font-weight: bold;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.sound-description {
  font-size: 9px;
  color: var(--text-secondary);
}

.sound-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.sound-slider {
  width: 60px;
  height: 4px;
  background: var(--surface-elevated);
  border-radius: 2px;
  outline: none;
  -webkit-appearance: none;
  appearance: none;
  cursor: pointer;
}

.sound-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 12px;
  height: 12px;
  background: var(--accent-500);
  border-radius: 50%;
  cursor: pointer;
}

.sound-slider::-moz-range-thumb {
  width: 12px;
  height: 12px;
  background: var(--accent-500);
  border-radius: 50%;
  cursor: pointer;
  border: none;
}

.play-btn {
  width: 24px;
  height: 24px;
  background: var(--accent-500);
  border: none;
  border-radius: 50%;
  color: var(--text-primary);
  font-size: 10px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s ease;
}

.play-btn:hover {
  background: var(--accent-600);
  transform: scale(1.1);
}

.audio-visualizer {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 20px;
}

.visualizer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.visualizer-title {
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--primary-400);
}

.visualizer-toggle {
  padding: 4px 8px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 4px;
  font-size: 9px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s ease;
}

.visualizer-toggle.active {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

.visualizer-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100px;
}

.visualizer-canvas {
  width: 100%;
  height: 100%;
  border-radius: 4px;
  background: var(--background);
}

.audio-settings {
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 20px;
}

.settings-title {
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--primary-400);
  margin-bottom: 15px;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 10px;
}

.setting-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 10px;
  color: var(--text-primary);
}

.setting-toggle input[type="checkbox"] {
  accent-color: var(--primary-500);
}

.setting-label {
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.mini-player {
  background: var(--surface);
  border: 1px solid var(--surface-border);
  border-radius: 8px;
  padding: 12px 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.player-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.now-playing {
  font-size: 9px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.current-track {
  font-size: 11px;
  font-weight: bold;
  color: var(--primary-400);
}

.player-controls {
  display: flex;
  gap: 8px;
}

.player-btn {
  width: 32px;
  height: 32px;
  background: var(--surface-elevated);
  border: 1px solid var(--surface-border);
  border-radius: 50%;
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.player-btn:hover {
  border-color: var(--primary-500);
  color: var(--primary-400);
  transform: scale(1.1);
}

.player-btn.play-pause {
  background: var(--primary-500);
  border-color: var(--primary-500);
  color: var(--text-primary);
}

/* Glitch effect for title */
.glitch {
  position: relative;
  color: var(--primary-400);
  font-size: 14px;
  font-weight: bold;
  text-transform: uppercase;
  text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  animation: glitch 2s infinite;
}

@keyframes glitch {
  0%, 90%, 100% {
    text-shadow: 2px 2px 0 var(--accent-500), -2px -2px 0 var(--primary-600);
  }
  95% {
    text-shadow: -2px 2px 0 var(--accent-500), 2px -2px 0 var(--primary-600);
  }
}

/* Responsive design */
@media (max-width: 768px) {
  .sound-categories {
    gap: 10px;
  }

  .sound-item {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .sound-controls {
    justify-content: space-between;
  }

  .settings-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .mini-player {
    flex-direction: column;
    gap: 10px;
    text-align: center;
  }
}

@media (max-width: 480px) {
  .category-header {
    flex-direction: column;
    gap: 8px;
    text-align: center;
  }

  .sound-slider {
    width: 40px;
  }

  .settings-grid {
    grid-template-columns: 1fr;
  }
}
</style>