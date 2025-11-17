export interface GeoIPInfo {
  ip: string
  country: string
  countryCode: string
  region: string
  regionCode: string
  city: string
  latitude: number
  longitude: number
  isp: string
  organization: string
  as: string
  timezone: string
  isVpn: boolean
  isProxy: boolean
  isMobile: boolean
  timestamp: string
}

export interface NetworkConnectionWithGeo {
  // Network connection fields
  localAddr: string
  remoteAddr: string
  state: string
  pid: number
  processName: string
  protocol: string
  bytesSent: number
  bytesRecv: number
  established: string
  lastActivity: string

  // GeoIP fields
  remoteGeoIP?: GeoIPInfo
  countryFlag?: string
}

export interface GeoIPSettings {
  enabled: boolean
  provider: string
  showFlags: boolean
  showCity: boolean
  showISP: boolean
  cacheTimeout: number
  autoLookup: boolean
  privacyMode: boolean
}

export interface GeoIPCacheStats {
  size: number
  maxSize: number
  lastUpdate: string
  enabled: boolean
}

export interface GeoIPProvider {
  name: string
  displayName: string
  freeTier: boolean
  authType: 'none' | 'apikey'
  description: string
}

export interface GeoIPConfig {
  enabled: boolean
  cacheTimeout: number
  requestTimeout: number
  maxCacheSize: number
  provider: string
  apiKey?: string
  updateInterval: number
}

// Helper functions for GeoIP data
export class GeoIPUtils {
  static getCountryCode(info: GeoIPInfo): string {
    if (info.countryCode.length === 3) {
      // Convert ISO 3166-1 alpha-3 to alpha-2 (basic conversion for common countries)
      const countryMap: Record<string, string> = {
        'USA': 'US',
        'CAN': 'CA',
        'GBR': 'GB',
        'AUS': 'AU',
        'DEU': 'DE',
        'FRA': 'FR',
        'ITA': 'IT',
        'JPN': 'JP',
        'CHN': 'CN',
        'IND': 'IN',
        'BRA': 'BR',
        'MEX': 'MX',
        'ESP': 'ES',
        'RUS': 'RU',
        'KOR': 'KR',
        'NLD': 'NL',
        'BEL': 'BE',
        'SWE': 'SE',
        'NOR': 'NO',
        'DNK': 'DK',
        'FIN': 'FI',
        'CHE': 'CH',
        'AUT': 'AT',
        'NZL': 'NZ',
        'IRL': 'IE',
        'PRT': 'PT',
        'GRC': 'GR',
        'TUR': 'TR',
        'POL': 'PL',
        'CZE': 'CZ',
        'HUN': 'HU',
        'ROU': 'RO',
        'BGR': 'BG',
        'HRV': 'HR',
        'SVK': 'SK',
        'SVN': 'SI',
        'EST': 'EE',
        'LVA': 'LV',
        'LTU': 'LT',
      }
      return countryMap[info.countryCode] || info.countryCode
    }
    return info.countryCode
  }

  static getCountryFlag(info: GeoIPInfo): string {
    const code = this.getCountryCode(info)
    if (code.length !== 2) {
      return ''
    }

    // Convert country code to flag emoji
    const flagOffset = 0x1F1E6 // Regional Indicator Symbol Letter A
    const asciiOffset = 0x41 // ASCII 'A'

    const r1 = String.fromCodePoint(flagOffset + (code.charCodeAt(0) - asciiOffset))
    const r2 = String.fromCodePoint(flagOffset + (code.charCodeAt(1) - asciiOffset))

    return r1 + r2
  }

  static getLocationString(info: GeoIPInfo): string {
    if (!info.city) {
      return info.country
    }
    if (!info.region) {
      return `${info.city}, ${info.country}`
    }
    return `${info.city}, ${info.region}, ${info.country}`
  }

  static isEmpty(info: GeoIPInfo): boolean {
    return !info.ip && !info.country && !info.city
  }

  static getAvailableProviders(): GeoIPProvider[] {
    return [
      {
        name: 'ipapi.co',
        displayName: 'IPAPI.co',
        freeTier: true,
        authType: 'none',
        description: 'Free IP geolocation API with no authentication required'
      },
      {
        name: 'ip-api.com',
        displayName: 'IP-API.com',
        freeTier: true,
        authType: 'none',
        description: 'Free IP geolocation API with comprehensive data'
      },
      {
        name: 'maxmind',
        displayName: 'MaxMind GeoIP2',
        freeTier: false,
        authType: 'apikey',
        description: 'Commercial GeoIP database with high accuracy'
      }
    ]
  }

  static formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
  }

  static formatSpeed(bits: number): string {
    if (!bits || bits === 0) return 'Unknown'
    const mbps = bits / 1000000
    if (mbps < 1000) {
      return `${mbps.toFixed(0)} Mbps`
    }
    const gbps = mbps / 1000
    return `${gbps.toFixed(1)} Gbps`
  }

  static extractIPFromAddr(addr: string): string {
    // Handle IPv6 addresses in brackets
    if (addr.length > 0 && addr[0] === '[') {
      const end = addr.indexOf(']')
      if (end !== -1) {
        addr = addr.substring(1, end)
      }
    }

    // Split on colon to separate IP from port
    const colonIndex = addr.lastIndexOf(':')
    if (colonIndex !== -1) {
      addr = addr.substring(0, colonIndex)
    }

    // Basic IP validation
    const ipv4Regex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
    const ipv6Regex = /^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::1$|^::$/

    if (ipv4Regex.test(addr) || ipv6Regex.test(addr)) {
      return addr
    }

    return ''
  }
}