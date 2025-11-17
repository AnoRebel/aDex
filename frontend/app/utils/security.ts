/**
 * Security utilities for frontend input validation and sanitization
 */

import { useStorage, useThrottleFn } from '@vueuse/core'

export interface ValidationResult {
  isValid: boolean
  sanitized: string
  warnings: string[]
  errors: string[]
  riskLevel: number
}

export interface SecurityConfig {
  maxInputLength: number
  sanitizationLevel: 'minimal' | 'standard' | 'strict' | 'paranoid'
  allowHTMLTags: boolean
  allowedTags: string[]
  blockXSS: boolean
  blockSQLInjection: boolean
  maxLength: {
    filename: number
    url: number
    general: number
  }
}

export class SecurityUtils {
  private static readonly DEFAULT_CONFIG: SecurityConfig = {
    maxInputLength: 1024,
    sanitizationLevel: 'standard',
    allowHTMLTags: false,
    allowedTags: ['b', 'i', 'em', 'strong', 'br'],
    blockXSS: true,
    blockSQLInjection: true,
    maxLength: {
      filename: 255,
      url: 2048,
      general: 1024
    }
  }

  /**
   * Sanitizes user input based on configuration
   */
  static sanitizeInput(input: string, config: Partial<SecurityConfig> = {}): ValidationResult {
    const finalConfig = { ...this.DEFAULT_CONFIG, ...config }

    const result: ValidationResult = {
      isValid: true,
      sanitized: input,
      warnings: [],
      errors: [],
      riskLevel: 0
    }

    // Check input length
    if (input.length > finalConfig.maxInputLength) {
      result.isValid = false
      result.errors.push(`Input too long: ${input.length} > ${finalConfig.maxInputLength}`)
      result.riskLevel += 3
      input = input.substring(0, finalConfig.maxInputLength)
    }

    // Apply sanitization based on level
    result.sanitized = this.applySanitization(input, finalConfig)

    // Security checks
    if (finalConfig.blockXSS && this.detectXSS(result.sanitized)) {
      result.isValid = false
      result.errors.push('Potential XSS attack detected')
      result.riskLevel += 5
    }

    if (finalConfig.blockSQLInjection && this.detectSQLInjection(result.sanitized)) {
      result.isValid = false
      result.errors.push('Potential SQL injection detected')
      result.riskLevel += 5
    }

    if (this.detectPathTraversal(result.sanitized)) {
      result.isValid = false
      result.errors.push('Path traversal attempt detected')
      result.riskLevel += 4
    }

    if (this.detectCommandInjection(result.sanitized)) {
      result.isValid = false
      result.errors.push('Command injection attempt detected')
      result.riskLevel += 5
    }

    return result
  }

  /**
   * Validates a URL for security
   */
  static validateURL(url: string, config: Partial<SecurityConfig> = {}): ValidationResult {
    const finalConfig = { ...this.DEFAULT_CONFIG, ...config }

    const result: ValidationResult = {
      isValid: true,
      sanitized: url,
      warnings: [],
      errors: [],
      riskLevel: 0
    }

    // Check URL length
    if (url.length > finalConfig.maxLength.url) {
      result.isValid = false
      result.errors.push(`URL too long: ${url.length} > ${finalConfig.maxLength.url}`)
      result.riskLevel += 2
      return result
    }

    try {
      const parsedURL = new URL(url)
      const allowedProtocols = ['http:', 'https:', 'ws:', 'wss:']

      if (!allowedProtocols.includes(parsedURL.protocol)) {
        result.isValid = false
        result.errors.push(`Protocol not allowed: ${parsedURL.protocol}`)
        result.riskLevel += 3
        return result
      }

      // Check for suspicious patterns
      if (this.detectSuspiciousURL(url)) {
        result.isValid = false
        result.errors.push('Suspicious URL pattern detected')
        result.riskLevel += 4
      }

      result.sanitized = parsedURL.toString()
    } catch (error) {
      result.isValid = false
      result.errors.push(`Invalid URL format: ${error}`)
      result.riskLevel += 3
    }

    return result
  }

  /**
   * Validates a filename for security
   */
  static validateFilename(filename: string, config: Partial<SecurityConfig> = {}): ValidationResult {
    const finalConfig = { ...this.DEFAULT_CONFIG, ...config }

    const result: ValidationResult = {
      isValid: true,
      sanitized: filename,
      warnings: [],
      errors: [],
      riskLevel: 0
    }

    // Check filename length
    if (filename.length > finalConfig.maxLength.filename) {
      result.isValid = false
      result.errors.push(`Filename too long: ${filename.length} > ${finalConfig.maxLength.filename}`)
      result.riskLevel += 2
      return result
    }

    // Check for empty filename
    if (filename.trim() === '') {
      result.isValid = false
      result.errors.push('Filename cannot be empty')
      result.riskLevel += 2
      return result
    }

    // Check for dangerous characters
    const dangerousChars = ['..', '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\x00']

    for (const char of dangerousChars) {
      if (filename.includes(char)) {
        result.isValid = false
        result.errors.push(`Dangerous character in filename: ${char}`)
        result.riskLevel += 3
      }
    }

    // Check for reserved names (Windows)
    const reservedNames = [
      'CON', 'PRN', 'AUX', 'NUL',
      'COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', 'COM9',
      'LPT1', 'LPT2', 'LPT3', 'LPT4', 'LPT5', 'LPT6', 'LPT7', 'LPT8', 'LPT9'
    ]

    const base = filename.substring(0, filename.lastIndexOf('.')) || filename
    if (reservedNames.includes(base.toUpperCase())) {
      result.isValid = false
      result.errors.push(`Reserved filename: ${base}`)
      result.riskLevel += 2
    }

    // Sanitize filename
    result.sanitized = this.sanitizeFilename(filename)

    return result
  }

  /**
   * Generates a secure random token
   */
  static generateSecureToken(length: number = 32): string {
    const array = new Uint8Array(length)
    crypto.getRandomValues(array)
    return btoa(String.fromCharCode(...array))
  }

  /**
   * Creates a hash of a string (basic implementation)
   */
  static async hashString(input: string): Promise<string> {
    const encoder = new TextEncoder()
    const data = encoder.encode(input)
    const hashBuffer = await crypto.subtle.digest('SHA-256', data)
    const hashArray = Array.from(new Uint8Array(hashBuffer))
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
  }

  /**
   * Detects potential XSS attacks
   */
  private static detectXSS(input: string): boolean {
    const lower = input.toLowerCase()
    const xssPatterns = [
      '<script', '</script>', 'javascript:', 'vbscript:', 'onload=',
      'onerror=', 'onclick=', 'onmouseover=', 'onfocus=', 'onblur=',
      'eval(', 'expression(', 'alert(', 'confirm(', 'prompt(',
      '<iframe', '<object', '<embed', '<link', '<meta', '<style',
      '@import', 'behavior:', 'binding:', 'include-source:'
    ]

    return xssPatterns.some(pattern => lower.includes(pattern))
  }

  /**
   * Detects potential SQL injection attacks
   */
  private static detectSQLInjection(input: string): boolean {
    const lower = input.toLowerCase()
    const sqlPatterns = [
      'union select', 'drop table', 'insert into', 'delete from',
      'update set', 'create table', 'alter table', 'exec(', 'execute(',
      'sp_executesql', 'xp_cmdshell', '--', '/*', '*/', "' or '1'='1",
      "' or 1=1", "' waitfor delay '", 'sleep(', 'benchmark('
    ]

    return sqlPatterns.some(pattern => lower.includes(pattern))
  }

  /**
   * Detects path traversal attacks
   */
  private static detectPathTraversal(input: string): boolean {
    const lower = input.toLowerCase()
    const pathPatterns = [
      '../', '..\\', '%2e%2e%2f', '%2e%2e\\', '..%2f', '..%5c',
      '/etc/passwd', '/etc/shadow', '/proc/', '/sys/', '/root/',
      '\\windows\\system32', 'c:\\windows', 'boot.ini'
    ]

    return pathPatterns.some(pattern => lower.includes(pattern))
  }

  /**
   * Detects command injection attacks
   */
  private static detectCommandInjection(input: string): boolean {
    const lower = input.toLowerCase()
    const commandPatterns = [
      ';', '|', '&', '&&', '||', '`', '$(', '${',
      '>>', '>', '<', '2>&1', '/dev/null', 'nc ', 'netcat',
      'wget ', 'curl ', 'powershell', 'cmd.exe', '/bin/sh',
      '/bin/bash', 'python -c', 'perl -e', 'ruby -e'
    ]

    return commandPatterns.some(pattern => lower.includes(pattern))
  }

  /**
   * Detects suspicious URL patterns
   */
  private static detectSuspiciousURL(url: string): boolean {
    const lower = url.toLowerCase()
    const suspiciousPatterns = [
      'javascript:', 'vbscript:', 'data:', 'file:', 'ftp:',
      'mailto:', 'tel:', 'sms:', 'callto:', 'tftp:',
      'ldap:', 'gopher:', 'mms:', 'rtp:', 'rtsp:'
    ]

    return suspiciousPatterns.some(pattern => lower.startsWith(pattern))
  }

  /**
   * Applies sanitization based on level
   */
  private static applySanitization(input: string, config: SecurityConfig): string {
    let sanitized = input

    switch (config.sanitizationLevel) {
      case 'minimal':
        sanitized = this.escapeHTML(sanitized)
        break
      case 'standard':
        sanitized = this.escapeHTML(sanitized)
        sanitized = sanitized.replace(/[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]/g, '')
        break
      case 'strict':
        sanitized = this.escapeHTML(sanitized)
        sanitized = sanitized.replace(/javascript:|vbscript:|data:|file:/gi, '')
        sanitized = sanitized.replace(/[\x00-\x1F\x7F]/g, '')
        break
      case 'paranoid':
        // Allow only alphanumeric and basic punctuation
        sanitized = sanitized.replace(/[^a-zA-Z0-9\s.,!?()@#$%^&*\-_=+\[\]{}'"`]/g, '')
        sanitized = this.escapeHTML(sanitized)
        break
    }

    return sanitized
  }

  /**
   * Escapes HTML characters
   */
  private static escapeHTML(input: string): string {
    const div = document.createElement('div')
    div.textContent = input
    return div.innerHTML
  }

  /**
   * Sanitizes a filename
   */
  private static sanitizeFilename(filename: string): string {
    let sanitized = filename

    // Remove dangerous characters
    sanitized = sanitized.replace(/[<>:"/\\|?*\x00-\x1F]/g, '_')

    // Trim leading/trailing spaces and dots
    sanitized = sanitized.trim().replace(/^\.+|\.+$/g, '')

    // Replace multiple dots with single dot
    sanitized = sanitized.replace(/\.+/g, '.')

    return sanitized
  }
}

/**
 * Vue composable for security utilities
 */
export function useSecurity() {
  const sanitizeInput = (input: string, config?: Partial<SecurityConfig>) => {
    return SecurityUtils.sanitizeInput(input, config)
  }

  const validateURL = (url: string, config?: Partial<SecurityConfig>) => {
    return SecurityUtils.validateURL(url, config)
  }

  const validateFilename = (filename: string, config?: Partial<SecurityConfig>) => {
    return SecurityUtils.validateFilename(filename, config)
  }

  const generateToken = (length?: number) => {
    return SecurityUtils.generateSecureToken(length)
  }

  const hashString = (input: string) => {
    return SecurityUtils.hashString(input)
  }

  return {
    sanitizeInput,
    validateURL,
    validateFilename,
    generateToken,
    hashString
  }
}

/**
 * Input validation directives for Vue
 */
export const vValidateInput = {
  mounted(el: HTMLElement, binding: any) {
    const validateOnInput = (event: Event) => {
      const input = event.target as HTMLInputElement
      const result = SecurityUtils.sanitizeInput(input.value, binding.value)

      if (!result.isValid) {
        input.setCustomValidity(result.errors.join(', '))
        input.reportValidity()
      } else {
        input.setCustomValidity('')
        input.value = result.sanitized
      }
    }

    el.addEventListener('input', validateOnInput)

    // Store cleanup function
    ;(el as any)._validateInputCleanup = () => {
      el.removeEventListener('input', validateOnInput)
    }
  },

  unmounted(el: HTMLElement) {
    if ((el as any)._validateInputCleanup) {
      ;(el as any)._validateInputCleanup()
    }
  }
}