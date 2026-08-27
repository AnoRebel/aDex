<template>
  <Transition name="splash-fade">
    <div v-if="visible" class="splash-screen">
      <div class="splash-content">
        <!-- Animated Logo -->
        <div class="logo-container">
          <div class="logo-glow"></div>
          <h1 class="logo-text">
            <span class="logo-letter" v-for="(letter, i) in 'aDex'" :key="i" :style="{ animationDelay: `${i * 0.1}s` }">
              {{ letter }}
            </span>
          </h1>
          <div class="logo-subtitle">Advanced Desktop Environment</div>
        </div>

        <!-- Loading Animation -->
        <div class="loading-section">
          <div class="loading-bar">
            <div class="loading-progress" :style="{ width: `${progress}%` }"></div>
          </div>
          <div class="loading-text">{{ loadingText }}</div>
        </div>

        <!-- Version Info -->
        <div class="version-info">v2.0.0</div>

        <!-- Decorative Elements -->
        <div class="corner corner-tl"></div>
        <div class="corner corner-tr"></div>
        <div class="corner corner-bl"></div>
        <div class="corner corner-br"></div>

        <!-- Matrix Rain Effect -->
        <div class="matrix-rain">
          <div v-for="i in 20" :key="i" class="rain-column" :style="getRainStyle(i)"></div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'

interface Props {
  minDuration?: number
}

const props = withDefaults(defineProps<Props>(), {
  minDuration: 2000
})

const emit = defineEmits<{
  complete: []
}>()

const visible = ref(true)
const progress = ref(0)
const loadingText = ref('Initializing system...')

// Audio context for sci-fi sounds
let audioContext: AudioContext | null = null

// Initialize audio context
const initAudio = () => {
  try {
    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
  } catch (e) {
    console.warn('Web Audio API not supported')
  }
}

// Play sci-fi boot sound
const playBootSound = () => {
  if (!audioContext || audioContext.state === 'closed') return
  
  try {
    // Resume audio context if suspended (browser autoplay policy)
    if (audioContext.state === 'suspended') {
      audioContext.resume()
    }
    
    const now = audioContext.currentTime
    
    // Create oscillator for the main tone
    const oscillator = audioContext.createOscillator()
    const gainNode = audioContext.createGain()
    
    oscillator.connect(gainNode)
    gainNode.connect(audioContext.destination)
    
    // Sci-fi boot sound: sweeping frequency from low to high
    oscillator.type = 'sawtooth'
    oscillator.frequency.setValueAtTime(100, now)
    oscillator.frequency.exponentialRampToValueAtTime(800, now + 0.5)
    
    // Volume envelope
    gainNode.gain.setValueAtTime(0.1, now)
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.5)
    
    oscillator.start(now)
    oscillator.stop(now + 0.5)
  } catch (e) {
    console.warn('Boot sound failed:', e)
  }
}

// Play initialization step sound
const playStepSound = () => {
  if (!audioContext || audioContext.state === 'closed') return
  
  try {
    const now = audioContext.currentTime
    
    // Short blip sound
    const oscillator = audioContext.createOscillator()
    const gainNode = audioContext.createGain()
    
    oscillator.connect(gainNode)
    gainNode.connect(audioContext.destination)
    
    oscillator.type = 'sine'
    oscillator.frequency.setValueAtTime(1200, now)
    
    gainNode.gain.setValueAtTime(0.05, now)
    gainNode.gain.exponentialRampToValueAtTime(0.001, now + 0.1)
    
    oscillator.start(now)
    oscillator.stop(now + 0.1)
  } catch (e) {
    // Silently ignore audio errors
  }
}

// Play completion sound
const playCompleteSound = () => {
  if (!audioContext || audioContext.state === 'closed') return
  
  try {
    const now = audioContext.currentTime
    
    // Success chord
    const frequencies = [440, 554, 659] // A major chord
    
    frequencies.forEach((freq, i) => {
      const oscillator = audioContext!.createOscillator()
      const gainNode = audioContext!.createGain()
      
      oscillator.connect(gainNode)
      gainNode.connect(audioContext!.destination)
      
      oscillator.type = 'sine'
      oscillator.frequency.setValueAtTime(freq, now + i * 0.05)
      
      gainNode.gain.setValueAtTime(0.05, now + i * 0.05)
      gainNode.gain.exponentialRampToValueAtTime(0.001, now + 0.5 + i * 0.05)
      
      oscillator.start(now + i * 0.05)
      oscillator.stop(now + 0.5 + i * 0.05)
    })
  } catch (e) {
    // Silently ignore audio errors
  }
}

const loadingSteps = [
  { progress: 15, text: 'Loading core modules...' },
  { progress: 30, text: 'Initializing stores...' },
  { progress: 45, text: 'Connecting to backend...' },
  { progress: 60, text: 'Loading themes...' },
  { progress: 75, text: 'Preparing terminal...' },
  { progress: 90, text: 'Starting services...' },
  { progress: 100, text: 'System ready!' }
]

const getRainStyle = (index: number) => {
  return {
    left: `${(index - 1) * 5}%`,
    animationDelay: `${Math.random() * 2}s`,
    animationDuration: `${1 + Math.random() * 2}s`
  }
}

const startLoading = () => {
  // Initialize audio on first user interaction
  initAudio()
  
  // Play boot sound at start
  playBootSound()
  
  let stepIndex = 0
  const stepDuration = props.minDuration / loadingSteps.length

  const nextStep = () => {
    if (stepIndex < loadingSteps.length) {
      const step = loadingSteps[stepIndex]
      progress.value = step.progress
      loadingText.value = step.text
      
      // Play step sound
      playStepSound()
      
      stepIndex++
      setTimeout(nextStep, stepDuration)
    } else {
      // Play completion sound
      playCompleteSound()
      
      setTimeout(() => {
        visible.value = false
        emit('complete')
      }, 300)
    }
  }

  nextStep()
}

onMounted(() => {
  startLoading()
})

// Expose method to manually complete
defineExpose({
  complete: () => {
    progress.value = 100
    loadingText.value = 'System ready!'
    setTimeout(() => {
      visible.value = false
      emit('complete')
    }, 300)
  }
})
</script>

<style scoped>
.splash-screen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #0a0a0a 0%, #1a1a2e 50%, #0f0f1a 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  overflow: hidden;
}

.splash-content {
  position: relative;
  text-align: center;
  z-index: 2;
}

/* Logo Animation */
.logo-container {
  position: relative;
  margin-bottom: 60px;
}

.logo-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 300px;
  height: 100px;
  background: radial-gradient(ellipse, rgba(0, 255, 255, 0.3) 0%, transparent 70%);
  filter: blur(20px);
  animation: pulse-glow 2s ease-in-out infinite;
}

.logo-text {
  font-size: 72px;
  font-weight: bold;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  color: #00ffff;
  text-shadow: 
    0 0 10px rgba(0, 255, 255, 0.8),
    0 0 20px rgba(0, 255, 255, 0.6),
    0 0 40px rgba(0, 255, 255, 0.4),
    0 0 80px rgba(0, 255, 255, 0.2);
  margin: 0;
  letter-spacing: 4px;
}

.logo-letter {
  display: inline-block;
  animation: letter-appear 0.5s ease-out forwards;
  opacity: 0;
  transform: translateY(-20px);
}

@keyframes letter-appear {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.logo-subtitle {
  font-size: 14px;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 8px;
  margin-top: 15px;
  animation: fade-in 1s ease-out 0.5s forwards;
  opacity: 0;
}

@keyframes fade-in {
  to {
    opacity: 1;
  }
}

/* Loading Bar */
.loading-section {
  width: 400px;
  margin: 0 auto;
}

.loading-bar {
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
  position: relative;
}

.loading-progress {
  height: 100%;
  background: linear-gradient(90deg, #00ffff, #00ff88, #00ffff);
  border-radius: 2px;
  transition: width 0.3s ease;
  position: relative;
}

.loading-progress::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.loading-text {
  font-size: 12px;
  color: #00ffff;
  margin-top: 15px;
  font-family: 'JetBrains Mono', monospace;
  text-transform: uppercase;
  letter-spacing: 2px;
}

/* Version */
.version-info {
  position: absolute;
  bottom: -80px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 11px;
  color: #444;
  font-family: 'JetBrains Mono', monospace;
}

/* Corner Decorations */
.corner {
  position: fixed;
  width: 60px;
  height: 60px;
  border-color: rgba(0, 255, 255, 0.3);
  border-style: solid;
  border-width: 0;
}

.corner-tl {
  top: 30px;
  left: 30px;
  border-top-width: 2px;
  border-left-width: 2px;
}

.corner-tr {
  top: 30px;
  right: 30px;
  border-top-width: 2px;
  border-right-width: 2px;
}

.corner-bl {
  bottom: 30px;
  left: 30px;
  border-bottom-width: 2px;
  border-left-width: 2px;
}

.corner-br {
  bottom: 30px;
  right: 30px;
  border-bottom-width: 2px;
  border-right-width: 2px;
}

/* Matrix Rain Effect */
.matrix-rain {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  z-index: 0;
  opacity: 0.15;
}

.rain-column {
  position: absolute;
  top: -100%;
  width: 2px;
  height: 100px;
  background: linear-gradient(to bottom, transparent, #00ffff, transparent);
  animation: rain-fall linear infinite;
}

@keyframes rain-fall {
  0% {
    transform: translateY(0);
  }
  100% {
    transform: translateY(calc(100dvh + 200px));
  }
}

@keyframes pulse-glow {
  0%, 100% {
    opacity: 0.5;
    transform: translate(-50%, -50%) scale(1);
  }
  50% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.1);
  }
}

/* Fade Transition */
.splash-fade-enter-active,
.splash-fade-leave-active {
  transition: opacity 0.5s ease;
}

.splash-fade-enter-from,
.splash-fade-leave-to {
  opacity: 0;
}
</style>
