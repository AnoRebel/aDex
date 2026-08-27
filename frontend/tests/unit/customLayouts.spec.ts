import { describe, it, expect } from 'vitest'
import { validateLayout } from '../../app/composables/useCustomLayouts'

describe('custom layout validation', () => {
  it('accepts a well-formed layout', () => {
    const errors: string[] = []
    const layout = validateLayout(
      { id: 'minimal', displayName: 'Minimal', regions: { left: ['clock', 'cpu'], right: ['netstat'] } },
      errors,
    )
    expect(errors).toEqual([])
    expect(layout?.id).toBe('minimal')
    expect(layout?.regions.left).toEqual(['clock', 'cpu'])
    expect(layout?.regions.right).toEqual(['netstat'])
  })

  it('drops unknown panels but keeps the rest of the layout', () => {
    const errors: string[] = []
    const layout = validateLayout(
      { id: 'partly-bad', regions: { left: ['clock', 'not-a-real-panel', 'cpu'] } },
      errors,
    )
    expect(layout).not.toBeNull()
    expect(layout?.regions.left).toEqual(['clock', 'cpu'])
    expect(errors.some(e => e.includes('not-a-real-panel'))).toBe(true)
  })

  it('drops unknown regions but keeps valid ones', () => {
    const errors: string[] = []
    const layout = validateLayout(
      { id: 'odd-region', regions: { left: ['clock'], nowhere: ['cpu'] } },
      errors,
    )
    expect(layout?.regions.left).toEqual(['clock'])
    expect('nowhere' in (layout?.regions ?? {})).toBe(false)
    expect(errors.some(e => e.includes('nowhere'))).toBe(true)
  })

  it('rejects a layout with no id', () => {
    const errors: string[] = []
    expect(validateLayout({ regions: { left: ['clock'] } }, errors)).toBeNull()
    expect(errors.some(e => e.includes('id'))).toBe(true)
  })

  it('rejects a layout with no usable regions', () => {
    const errors: string[] = []
    expect(validateLayout({ id: 'empty', regions: {} }, errors)).toBeNull()
  })

  it('rejects non-object entries', () => {
    const errors: string[] = []
    expect(validateLayout('not-a-layout', errors)).toBeNull()
    expect(validateLayout(null, errors)).toBeNull()
  })

  it('allows a layout to omit panels entirely', () => {
    const errors: string[] = []
    const layout = validateLayout({ id: 'left-only', regions: { left: ['clock'] } }, errors)
    expect(errors).toEqual([])
    expect(layout?.regions.right).toBeUndefined()
  })
})
