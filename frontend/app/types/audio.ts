// Audio event types
export type AudioEventType = 'system' | 'interaction' | 'notification' | 'error' | 'success'

// Audio event interface
export interface AudioEvent {
  id: string
  name: string
  description: string
  filePath: string
  duration: number // milliseconds
  volume: number // 0-100
  category: AudioEventType
  enabled: boolean
  createdAt: Date
}

// Audio settings interface
export interface AudioSettings {
  enabled: boolean
  volume: number // 0-100
  muted: boolean
  soundpack: string
  globalVolume: number // 0.0-1.0
  effectsVolume: number // 0.0-1.0
  notificationVolume: number // 0.0-1.0
  muteInBackground: boolean
  autoPlay: boolean
  enabledEvents: string[]
  version: string
  updatedAt: string
}

// Audio playback status interface
export interface AudioPlaybackStatus {
  isPlaying: boolean
  currentEvent: AudioEvent | null
  volume: number
  muted: boolean
  loadedEvents: string[]
  supportedFormats: string[]
}

// Audio statistics interface
export interface AudioStats {
  totalPlays: number
  eventsPlayed: Record<string, number>
  categoriesPlayed: Record<string, number>
  totalPlaytime: number
  averageVolume: number
  lastPlayed: Date
  mostPlayedEvent: string
  settings: Record<string, any>
  generatedAt: Date
}

// Soundpack interface
export interface Soundpack {
  id: string
  name: string
  displayName: string
  description: string
  version: string
  author: string
  isBuiltIn: boolean
  enabled: boolean
}

// Legacy audio interfaces that were in the original file
export interface AudioDevice {
  id: string
  name: string
  type: 'input' | 'output'
  isDefault: boolean
  isEnabled: boolean
  volume: number
  isMuted: boolean
  channels: number
  sampleRate: number
  driver: string
  description?: string
}

export interface AudioSession {
  id: string
  name: string
  processId: number
  isActive: boolean
  volume: number
  isMuted: boolean
  deviceId?: string
  format?: string
}

export interface AudioConfig {
  sampleRate: number
  bitDepth: number
  channels: number
  bufferSize: number
  latency: number
  enableEffects: boolean
  enableEqualizer: boolean
}