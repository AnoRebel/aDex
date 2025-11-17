export interface FileItem {
  name: string
  path: string
  size: number
  modified: Date
  created: Date
  accessed: Date
  permissions: string
  owner: string
  group: string
  isHidden: boolean
  type: string
  extension: string
  mimeType: string
  isDirectory?: false
}

export interface DirectoryItem {
  name: string
  path: string
  size: number
  modified: Date
  created: Date
  accessed: Date
  permissions: string
  owner: string
  group: string
  isHidden: boolean
  itemCount: number
  isDirectory: true
}

export type FileSystemItem = FileItem | DirectoryItem

export interface FileSystemStats {
  totalSize: number
  usedSize: number
  freeSize: number
  totalFiles: number
  totalDirectories: number
  lastModified: Date
}

export interface FileOperation {
  type: 'copy' | 'move' | 'delete' | 'create' | 'rename'
  source: string[]
  destination?: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  startTime: Date
  endTime?: Date
  error?: string
}

export interface FileWatcher {
  id: string
  path: string
  isActive: boolean
  events: FileSystemEvent[]
  createdAt: Date
}

export interface FileSystemEvent {
  type: 'create' | 'modify' | 'delete' | 'rename' | 'move'
  path: string
  oldPath?: string
  timestamp: Date
  size?: number
}

export interface FileInfo {
  path: string
  name: string
  size: number
  permissions: string
  owner: string
  group: string
  modified: Date
  created: Date
  accessed: Date
  isDirectory: boolean
  isHidden: boolean
  isExecutable: boolean
  isReadable: boolean
  isWritable: boolean
  mimeType?: string
  checksum?: string
  metadata?: Record<string, any>
}

export interface DirectoryListing {
  path: string
  directories: DirectoryItem[]
  files: FileItem[]
  totalItems: number
  totalSize: number
  permissions: {
    readable: boolean
    writable: boolean
    executable: boolean
  }
}

export interface FileSystemConfig {
  showHiddenFiles: boolean
  sortBy: 'name' | 'size' | 'modified' | 'type'
  sortOrder: 'asc' | 'desc'
  viewMode: 'list' | 'grid' | 'tree'
  itemsPerPage: number
  previewEnabled: boolean
  thumbnailsEnabled: boolean
}

export interface FileSearch {
  query: string
  path: string
  recursive: boolean
  includeHidden: boolean
  fileTypes: string[]
  maxResults: number
  startTime?: Date
  endTime?: Date
  minSize?: number
  maxSize?: number
}

export interface SearchResult {
  path: string
  name: string
  type: 'file' | 'directory'
  size: number
  modified: Date
  matches: SearchMatch[]
}

export interface SearchMatch {
  line: number
  content: string
  startIndex: number
  endIndex: number
}

export interface FilePreview {
  path: string
  type: 'text' | 'image' | 'video' | 'audio' | 'pdf' | 'binary'
  content?: string
  thumbnail?: string
  metadata?: Record<string, any>
  error?: string
}

export interface MountPoint {
  device: string
  mountpoint: string
  fstype: string
  options: string[]
  total: number
  used: number
  free: number
  usage: number
}

export interface DiskUsage {
  device: string
  mountpoint: string
  total: number
  used: number
  free: number
  usage: number
  inodes: {
    total: number
    used: number
    free: number
  }
}

export interface FileSystemPermissions {
  user: {
    read: boolean
    write: boolean
    execute: boolean
  }
  group: {
    read: boolean
    write: boolean
    execute: boolean
  }
  others: {
    read: boolean
    write: boolean
    execute: boolean
  }
  octal: string
  symbolic: string
}
