/**
 * Platform Detection and Compatibility Utilities
 * Handles cross-platform differences and provides platform-specific utilities
 */

export interface PlatformInfo {
  os: 'windows' | 'macos' | 'linux' | 'unknown'
  arch: 'x64' | 'arm64' | 'unknown'
  isWindows: boolean
  isMac: boolean
  isLinux: boolean
  isMobile: boolean
  isTablet: boolean
  isDesktop: boolean
}

export interface KeyboardShortcuts {
  copy: string
  paste: string
  cut: string
  selectAll: string
  save: string
  open: string
  close: string
  quit: string
  find: string
  undo: string
  redo: string
}

export interface FileExtensions {
  executable: string
  script: string
  library: string
  config: string
  image: string[]
  video: string[]
  audio: string[]
}

// Platform detection
export const detectOS = (): PlatformInfo['os'] => {
  if (typeof window === 'undefined') {
    return 'unknown'
  }

  const userAgent = navigator.userAgent.toLowerCase()

  if (userAgent.includes('windows')) {
    return 'windows'
  } else if (userAgent.includes('mac') && userAgent.includes('os')) {
    return 'macos'
  } else if (userAgent.includes('linux')) {
    return 'linux'
  }

  return 'unknown'
}

export const detectArch = (): PlatformInfo['arch'] => {
  if (typeof window === 'undefined') {
    return 'unknown'
  }

  const userAgent = navigator.userAgent.toLowerCase()

  if (userAgent.includes('arm64') || userAgent.includes('aarch64')) {
    return 'arm64'
  } else if (userAgent.includes('x64') || userAgent.includes('x86_64') || userAgent.includes('win64')) {
    return 'x64'
  }

  return 'unknown'
}

export const detectDeviceType = (): Pick<PlatformInfo, 'isMobile' | 'isTablet' | 'isDesktop'> => {
  if (typeof window === 'undefined') {
    return { isMobile: false, isTablet: false, isDesktop: true }
  }

  const userAgent = navigator.userAgent.toLowerCase()
  const maxWidth = window.innerWidth || 1024

  // Mobile detection
  const isMobile = /android|webos|iphone|ipod|blackberry|iemobile|opera mini/i.test(userAgent) ||
                  (maxWidth < 768 && 'ontouchstart' in window)

  // Tablet detection
  const isTablet = /ipad|android(?!.*mobile)|tablet/i.test(userAgent) ||
                   (maxWidth >= 768 && maxWidth < 1024 && 'ontouchstart' in window)

  // Desktop is anything not mobile or tablet
  const isDesktop = !isMobile && !isTablet

  return { isMobile, isTablet, isDesktop }
}

export const getPlatformInfo = (): PlatformInfo => {
  const os = detectOS()
  const arch = detectArch()
  const device = detectDeviceType()

  return {
    os,
    arch,
    isWindows: os === 'windows',
    isMac: os === 'macos',
    isLinux: os === 'linux',
    ...device
  }
}

// Platform-specific utilities
export const getOSName = (): string => {
  const os = detectOS()
  switch (os) {
    case 'windows': return 'Windows'
    case 'macos': return 'macOS'
    case 'linux': return 'Linux'
    default: return 'Unknown'
  }
}

export const isWindows = (): boolean => detectOS() === 'windows'
export const isMac = (): boolean => detectOS() === 'macos'
export const isLinux = (): boolean => detectOS() === 'linux'
export const isUnix = (): boolean => isMac() || isLinux()

// Keyboard shortcuts
export const getKeyboardShortcuts = (): KeyboardShortcuts => {
  const cmdKey = isMac() ? 'Cmd' : 'Ctrl'
  const altKey = isMac() ? 'Option' : 'Alt'

  return {
    copy: `${cmdKey}+C`,
    paste: `${cmdKey}+V`,
    cut: `${cmdKey}+X`,
    selectAll: `${cmdKey}+A`,
    save: `${cmdKey}+S`,
    open: `${cmdKey}+O`,
    close: `${cmdKey}+W`,
    quit: isMac() ? `${cmdKey}+Q` : 'Alt+F4',
    find: `${cmdKey}+F`,
    undo: `${cmdKey}+Z`,
    redo: isMac() ? `${cmdKey}+Shift+Z` : `${cmdKey}+Y`
  }
}

// File system utilities
export const normalizePath = (path: string): string => {
  if (!path) return path

  if (isWindows()) {
    // Windows uses backslashes
    return path.replace(/\//g, '\\')
  } else {
    // Unix-like systems use forward slashes
    return path.replace(/\\/g, '/')
  }
}

export const getFileExtensions = (): FileExtensions => {
  const platform = getPlatformInfo()

  if (platform.isWindows) {
    return {
      executable: '.exe',
      script: '.bat',
      library: '.dll',
      config: '.ini',
      image: ['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.ico'],
      video: ['.mp4', '.avi', '.wmv', '.mov', '.mkv'],
      audio: ['.mp3', '.wav', '.wma', '.flac']
    }
  } else if (platform.isMac) {
    return {
      executable: '',
      script: '.sh',
      library: '.dylib',
      config: '.plist',
      image: ['.jpg', '.jpeg', '.png', '.gif', '.heic'],
      video: ['.mp4', '.mov', '.m4v', '.avi', '.mkv'],
      audio: ['.mp3', '.wav', '.m4a', '.flac', '.aac']
    }
  } else {
    return {
      executable: '',
      script: '.sh',
      library: '.so',
      config: '.conf',
      image: ['.jpg', '.jpeg', '.png', '.gif', '.webp'],
      video: ['.mp4', '.avi', '.mkv', '.webm', '.ogv'],
      audio: ['.mp3', '.wav', '.ogg', '.flac', '.opus']
    }
  }
}

// Path validation
export const isValidPath = (path: string): boolean => {
  if (!path || typeof path !== 'string') return false

  const platform = getPlatformInfo()

  if (platform.isWindows) {
    // Windows path validation
    const absolutePath = /^[a-zA-Z]:\\(?:[^\\/:*?"<>|]+\\)*[^\\/:*?"<>|]*$/
    const uncPath = /^\\\\[^\\/:*?"<>|]+(?:\\[^\\/:*?"<>|]+)*$/
    const relativePath = /^[^\\/:*?"<>|]+(?:\\[^\\/:*?"<>|]+)*$/

    return absolutePath.test(path) || uncPath.test(path) || relativePath.test(path)
  } else {
    // Unix-like path validation
    const absolutePath = /^\/(?:[^\/\0]+\/)*[^\/\0]*$/
    const relativePath = /^(?:[^\/\0]+\/)*[^\/\0]*$/
    const homePath = /^~(?:\/[^\/\0]*)*$/

    return absolutePath.test(path) || relativePath.test(path) || homePath.test(path)
  }
}

// Environment variable handling
export const getEnvironmentVariable = (name: string): string | null => {
  if (typeof window === 'undefined') {
    // Server-side
    return process.env[name] || null
  } else {
    // Client-side - would need to be provided by the backend
    return localStorage.getItem(`env_${name}`) || null
  }
}

// Performance optimization based on platform
export const getPerformanceSettings = () => {
  const platform = getPlatformInfo()
  const cores = navigator.hardwareConcurrency || 4
  const memory = (navigator as any).deviceMemory || 4

  // Base settings on hardware capabilities
  const isHighEnd = cores >= 8 && memory >= 8
  const isLowEnd = cores <= 2 && memory <= 2

  return {
    // Animation settings
    animations: {
      enabled: !isLowEnd,
      quality: isHighEnd ? 'high' : 'medium',
      fps: isHighEnd ? 60 : isLowEnd ? 30 : 45
    },

    // Terminal settings
    terminal: {
      scrollbackSize: isHighEnd ? 10000 : isLowEnd ? 1000 : 5000,
      bufferSize: isHighEnd ? 8192 : isLowEnd ? 2048 : 4096,
      maxUpdateFrequency: isHighEnd ? 120 : 60
    },

    // System monitoring
    systemMonitor: {
      updateInterval: isHighEnd ? 500 : isLowEnd ? 2000 : 1000,
      historySize: isHighEnd ? 1000 : isLowEnd ? 100 : 500
    },

    // File browser
    fileBrowser: {
      itemsPerPage: isHighEnd ? 1000 : isLowEnd ? 100 : 500,
      thumbnailGeneration: !isLowEnd,
      previewSize: isHighEnd ? 'large' : 'small'
    }
  }
}

// Theme adaptation
export const getSystemTheme = () => {
  if (typeof window === 'undefined') return 'light'

  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const watchSystemTheme = (callback: (theme: 'light' | 'dark') => void) => {
  if (typeof window === 'undefined') return () => {}

  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

  const handler = (e: MediaQueryListEvent) => {
    callback(e.matches ? 'dark' : 'light')
  }

  mediaQuery.addEventListener('change', handler)

  // Return cleanup function
  return () => {
    mediaQuery.removeEventListener('change', handler)
  }
}

// Text direction for RTL languages
export const getTextDirection = (locale?: string): 'ltr' | 'rtl' => {
  const language = locale || navigator.language || 'en-US'
  const rtlLanguages = ['ar', 'he', 'fa', 'ur', 'yi', 'ckb']

  return rtlLanguages.some(lang => language.startsWith(lang)) ? 'rtl' : 'ltr'
}

// Platform-specific file dialog filters
export const getFileDialogFilters = () => {
  const platform = getPlatformInfo()

  if (platform.isWindows) {
    return [
      { name: 'All Files', extensions: ['*'] },
      { name: 'Text Files', extensions: ['txt', 'md', 'log'] },
      { name: 'Config Files', extensions: ['json', 'yaml', 'yml', 'toml', 'ini'] },
      { name: 'Script Files', extensions: ['js', 'ts', 'py', 'sh', 'bat'] }
    ]
  } else {
    return [
      { name: 'All Files', extensions: ['*'] },
      { name: 'Text Files', extensions: ['txt', 'md', 'log'] },
      { name: 'Config Files', extensions: ['json', 'yaml', 'yml', 'toml', 'conf'] },
      { name: 'Script Files', extensions: ['js', 'ts', 'py', 'sh'] }
    ]
  }
}

// Error message formatting
export const formatErrorForPlatform = (error: Error | string): string => {
  const message = typeof error === 'string' ? error : error.message
  const platform = getPlatformInfo()

  if (platform.isWindows) {
    // Windows-style error formatting
    return message.replace(/\//g, '\\').replace(/'/g, '"')
  } else {
    // Unix-style error formatting
    return message.replace(/\\/g, '/').replace(/"/g, "'")
  }
}

// Default application directories
export const getDefaultDirectories = () => {
  const platform = getPlatformInfo()

  if (platform.isWindows) {
    return {
      home: 'C:\\Users\\%USERNAME%',
      documents: 'C:\\Users\\%USERNAME%\\Documents',
      downloads: 'C:\\Users\\%USERNAME%\\Downloads',
      desktop: 'C:\\Users\\%USERNAME%\\Desktop',
      pictures: 'C:\\Users\\%USERNAME%\\Pictures',
      music: 'C:\\Users\\%USERNAME%\\Music',
      videos: 'C:\\Users\\%USERNAME%\\Videos',
      appData: 'C:\\Users\\%USERNAME%\\AppData',
      programFiles: 'C:\\Program Files',
      programFilesX86: 'C:\\Program Files (x86)',
      temp: 'C:\\Users\\%USERNAME%\\AppData\\Local\\Temp'
    }
  } else if (platform.isMac) {
    return {
      home: '/Users/%USERNAME%',
      documents: '/Users/%USERNAME%/Documents',
      downloads: '/Users/%USERNAME%/Downloads',
      desktop: '/Users/%USERNAME%/Desktop',
      pictures: '/Users/%USERNAME%/Pictures',
      music: '/Users/%USERNAME%/Music',
      videos: '/Users/%USERNAME%/Movies',
      appData: '/Users/%USERNAME%/Library/Application Support',
      temp: '/tmp'
    }
  } else {
    return {
      home: '/home/%USERNAME%',
      documents: '/home/%USERNAME%/Documents',
      downloads: '/home/%USERNAME%/Downloads',
      desktop: '/home/%USERNAME%/Desktop',
      pictures: '/home/%USERNAME%/Pictures',
      music: '/home/%USERNAME%/Music',
      videos: '/home/%USERNAME%/Videos',
      appData: '/home/%USERNAME%/.local/share',
      config: '/home/%USERNAME%/.config',
      temp: '/tmp'
    }
  }
}