import { describe, it, expect, beforeEach, vi } from 'vitest'
import { SecurityUtils, useSecurity, vValidateInput } from '~/utils/security'
import type { SecurityConfig } from '~/utils/security'

describe('SecurityUtils', () => {
  describe('sanitizeInput', () => {
    it('should handle valid input', () => {
      const result = SecurityUtils.sanitizeInput('Hello, World!')
      expect(result.isValid).toBe(true)
      expect(result.sanitized).toBe('Hello, World!')
      expect(result.errors).toHaveLength(0)
      expect(result.riskLevel).toBe(0)
    })

    it('should detect XSS attacks', () => {
      const xssInput = '<script>alert("xss")</script>'
      const result = SecurityUtils.sanitizeInput(xssInput)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Potential XSS attack detected')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should detect SQL injection attempts', () => {
      const sqlInput = "'; DROP TABLE users; --"
      const result = SecurityUtils.sanitizeInput(sqlInput)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Potential SQL injection detected')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should detect path traversal attempts', () => {
      const pathInput = '../../../etc/passwd'
      const result = SecurityUtils.sanitizeInput(pathInput)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Path traversal attempt detected')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should detect command injection attempts', () => {
      const cmdInput = '; rm -rf /'
      const result = SecurityUtils.sanitizeInput(cmdInput)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Command injection attempt detected')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should handle too long input', () => {
      const longInput = 'a'.repeat(2000)
      const result = SecurityUtils.sanitizeInput(longInput)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('too long')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should apply different sanitization levels', () => {
      const testInput = 'Hello\x00<script>world</script>'

      const minimal = SecurityUtils.sanitizeInput(testInput, { sanitizationLevel: 'minimal' })
      expect(minimal.sanitized).toContain('&lt;script&gt;')

      const standard = SecurityUtils.sanitizeInput(testInput, { sanitizationLevel: 'standard' })
      expect(standard.sanitized).not.toContain('\x00')
      expect(standard.sanitized).toContain('&lt;script&gt;')

      const strict = SecurityUtils.sanitizeInput(testInput, { sanitizationLevel: 'strict' })
      expect(strict.sanitized).not.toContain('\x00')
      expect(strict.sanitized).not.toContain('<script>')

      const paranoid = SecurityUtils.sanitizeInput(testInput, { sanitizationLevel: 'paranoid' })
      expect(paranoid.sanitized).toBe('Hello world')
    })
  })

  describe('validateURL', () => {
    it('should validate legitimate URLs', () => {
      const validURL = 'https://example.com/path?param=value'
      const result = SecurityUtils.validateURL(validURL)
      expect(result.isValid).toBe(true)
      expect(result.sanitized).toBe(validURL)
      expect(result.errors).toHaveLength(0)
    })

    it('should reject dangerous protocols', () => {
      const dangerousURL = 'javascript:alert("xss")'
      const result = SecurityUtils.validateURL(dangerousURL)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Protocol not allowed')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should reject malformed URLs', () => {
      const malformedURL = 'not-a-valid-url'
      const result = SecurityUtils.validateURL(malformedURL)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Invalid URL format')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should reject too long URLs', () => {
      const longURL = 'https://example.com/' + 'a'.repeat(3000)
      const result = SecurityUtils.validateURL(longURL)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('URL too long')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should handle relative URLs', () => {
      const relativeURL = '/api/users'
      const result = SecurityUtils.validateURL(relativeURL)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Invalid URL format')
    })
  })

  describe('validateFilename', () => {
    it('should validate safe filenames', () => {
      const safeFilename = 'document.txt'
      const result = SecurityUtils.validateFilename(safeFilename)
      expect(result.isValid).toBe(true)
      expect(result.sanitized).toBe(safeFilename)
      expect(result.errors).toHaveLength(0)
    })

    it('should reject dangerous characters', () => {
      const dangerousFilename = '../../../etc/passwd'
      const result = SecurityUtils.validateFilename(dangerousFilename)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Dangerous character')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should reject reserved names', () => {
      const reservedFilename = 'CON.txt'
      const result = SecurityUtils.validateFilename(reservedFilename)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Reserved filename')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should reject empty filenames', () => {
      const emptyFilename = ''
      const result = SecurityUtils.validateFilename(emptyFilename)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Filename cannot be empty')
      expect(result.riskLevel).toBeGreaterThan(0)
    })

    it('should sanitize filenames properly', () => {
      const dirtyFilename = 'file<>:"/\\|?*.txt'
      const result = SecurityUtils.validateFilename(dirtyFilename)
      expect(result.isValid).toBe(false) // Should be invalid due to dangerous chars
      expect(result.sanitized).toBe('file_________.txt') // Should be sanitized
    })

    it('should reject too long filenames', () => {
      const longFilename = 'a'.repeat(300)
      const result = SecurityUtils.validateFilename(longFilename)
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Filename too long')
      expect(result.riskLevel).toBeGreaterThan(0)
    })
  })

  describe('generateSecureToken', () => {
    it('should generate tokens of specified length', () => {
      const token = SecurityUtils.generateSecureToken(16)
      expect(token).toBeTruthy()
      expect(token.length).toBeGreaterThan(10) // Base64 encoding increases length
    })

    it('should generate unique tokens', () => {
      const token1 = SecurityUtils.generateSecureToken()
      const token2 = SecurityUtils.generateSecureToken()
      expect(token1).not.toBe(token2)
    })

    it('should use default length when not specified', () => {
      const token = SecurityUtils.generateSecureToken()
      expect(token).toBeTruthy()
      expect(token.length).toBeGreaterThan(0)
    })
  })

  describe('hashString', async () => {
    it('should hash strings consistently', async () => {
      const input = 'test string'
      const hash1 = await SecurityUtils.hashString(input)
      const hash2 = await SecurityUtils.hashString(input)
      expect(hash1).toBe(hash2)
      expect(hash1).toMatch(/^[a-f0-9]{64}$/) // SHA-256 hex format
    })

    it('should produce different hashes for different inputs', async () => {
      const hash1 = await SecurityUtils.hashString('string1')
      const hash2 = await SecurityUtils.hashString('string2')
      expect(hash1).not.toBe(hash2)
    })
  })
})

describe('useSecurity composable', () => {
  beforeEach(() => {
    // Reset any global state if needed
  })

  it('should provide security functions', () => {
    const security = useSecurity()
    expect(security.sanitizeInput).toBeInstanceOf(Function)
    expect(security.validateURL).toBeInstanceOf(Function)
    expect(security.validateFilename).toBeInstanceOf(Function)
    expect(security.generateToken).toBeInstanceOf(Function)
    expect(security.hashString).toBeInstanceOf(Function)
  })

  it('should wrap SecurityUtils methods correctly', () => {
    const security = useSecurity()
    const input = 'test input'

    const utilsResult = SecurityUtils.sanitizeInput(input)
    const composableResult = security.sanitizeInput(input)

    expect(composableResult).toEqual(utilsResult)
  })
})

describe('vValidateInput directive', () => {
  let mockElement: HTMLInputElement

  beforeEach(() => {
    mockElement = document.createElement('input') as HTMLInputElement
    Object.defineProperty(mockElement, 'setCustomValidity', {
      value: vi.fn(),
      writable: true
    })
    Object.defineProperty(mockElement, 'reportValidity', {
      value: vi.fn(() => true),
      writable: true
    })
  })

  it('should validate input on input event', () => {
    const directive = vValidateInput
    const mockBinding = { value: { sanitizationLevel: 'standard' } }

    // Mount directive
    directive.mounted(mockElement, mockBinding)

    // Simulate input event
    const event = new Event('input')
    Object.defineProperty(event, 'target', {
      value: {
        value: '<script>alert("xss")</script>',
        setCustomValidity: mockElement.setCustomValidity,
        reportValidity: mockElement.reportValidity
      },
      writable: true
    })

    mockElement.dispatchEvent(event)

    expect(mockElement.setCustomValidity).toHaveBeenCalled()
  })

  it('should cleanup on unmounted', () => {
    const directive = vValidateInput
    const mockBinding = { value: {} }
    const removeEventListenerSpy = vi.spyOn(mockElement, 'removeEventListener')

    // Mount directive
    directive.mounted(mockElement, mockBinding)

    // Unmount directive
    directive.unmounted(mockElement)

    expect(removeEventListenerSpy).toHaveBeenCalledWith('input', expect.any(Function))
  })
})

describe('Security edge cases', () => {
  it('should handle null/undefined inputs gracefully', () => {
    expect(() => {
      SecurityUtils.sanitizeInput('')
    }).not.toThrow()

    expect(() => {
      SecurityUtils.validateURL('')
    }).not.toThrow()

    expect(() => {
      SecurityUtils.validateFilename('')
    }).not.toThrow()
  })

  it('should handle Unicode characters correctly', () => {
    const unicodeInput = 'Hello 世界 🌍'
    const result = SecurityUtils.sanitizeInput(unicodeInput)
    expect(result.isValid).toBe(true)
    expect(result.sanitized).toContain('Hello 世界')
  })

  it('should handle special characters in strict mode', () => {
    const specialInput = '!@#$%^&*()_+-=[]{}|;:",.<>?/'
    const result = SecurityUtils.sanitizeInput(specialInput, { sanitizationLevel: 'strict' })
    expect(result.isValid).toBe(true)
    expect(result.sanitized).toBeTruthy()
  })

  it('should handle empty configuration', () => {
    const result = SecurityUtils.sanitizeInput('test', {})
    expect(result.isValid).toBe(true)
    expect(result.sanitized).toBe('test')
  })

  it('should handle custom configuration correctly', () => {
    const customConfig: Partial<SecurityConfig> = {
      maxInputLength: 10,
      sanitizationLevel: 'paranoid',
      blockXSS: true,
      blockSQLInjection: true
    }

    const result = SecurityUtils.sanitizeInput('This is a longer string with special chars <>', customConfig)
    expect(result.sanitized.length).toBeLessThanOrEqual(10)
  })
})

describe('Performance tests', () => {
  it('should handle large inputs efficiently', () => {
    const largeInput = 'a'.repeat(1000)

    const start = performance.now()
    const result = SecurityUtils.sanitizeInput(largeInput)
    const end = performance.now()

    expect(end - start).toBeLessThan(100) // Should complete in less than 100ms
    expect(result.isValid).toBe(true)
  })

  it('should handle multiple validations efficiently', () => {
    const testStrings = [
      'normal string',
      '<script>alert("xss")</script>',
      "'; DROP TABLE users; --",
      '../../../etc/passwd',
      'https://example.com/path',
      'document.txt'
    ]

    const start = performance.now()
    testStrings.forEach(str => {
      SecurityUtils.sanitizeInput(str)
      SecurityUtils.validateURL(str)
      SecurityUtils.validateFilename(str)
    })
    const end = performance.now()

    expect(end - start).toBeLessThan(50) // Should complete quickly
  })
})