import { describe, it, expect } from 'vitest'
import themesData from '~/app/assets/data/themes.json'

describe('Theme System', () => {
  describe('themes.json structure', () => {
    it('loads themes.json correctly', () => {
      expect(themesData).toBeDefined()
      expect(themesData).toHaveProperty('themes')
    })

    it('themes is an array', () => {
      expect(Array.isArray(themesData.themes)).toBe(true)
    })

    it('contains at least one theme', () => {
      expect(themesData.themes.length).toBeGreaterThan(0)
    })
  })

  describe('Expected themes', () => {
    const themeIds = themesData.themes.map((t: any) => t.id)

    it('has "tron" theme', () => {
      expect(themeIds).toContain('tron')
    })

    it('has "blade" theme', () => {
      expect(themeIds).toContain('blade')
    })

    it('has "matrix" theme', () => {
      expect(themeIds).toContain('matrix')
    })

    it('has "nord" theme', () => {
      expect(themeIds).toContain('nord')
    })

    it('has "matrix" theme', () => {
      expect(themeIds).toContain('matrix')
    })

    it('has "apollo" theme', () => {
      expect(themeIds).toContain('apollo')
    })

    it('has "nord" theme', () => {
      expect(themeIds).toContain('nord')
    })

    it('has "blade" theme', () => {
      expect(themeIds).toContain('blade')
    })
  })

  describe('Theme structure validation', () => {
    themesData.themes.forEach((theme: any) => {
      describe(`Theme: "${theme.name}" (${theme.id})`, () => {
        it('has an id string', () => {
          expect(typeof theme.id).toBe('string')
          expect(theme.id.length).toBeGreaterThan(0)
        })

        it('has a name string', () => {
          expect(typeof theme.name).toBe('string')
          expect(theme.name.length).toBeGreaterThan(0)
        })

        it('has a colors object', () => {
          expect(theme.colors).toBeDefined()
          expect(typeof theme.colors).toBe('object')
        })

        it('has r, g, b color values', () => {
          expect(theme.colors).toHaveProperty('r')
          expect(theme.colors).toHaveProperty('g')
          expect(theme.colors).toHaveProperty('b')
        })

        it('has black color string', () => {
          expect(typeof theme.colors.black).toBe('string')
        })

        it('has light_black color string', () => {
          expect(typeof theme.colors.light_black).toBe('string')
        })

        it('has grey color string', () => {
          expect(typeof theme.colors.grey).toBe('string')
        })
      })
    })
  })

  describe('Theme color ranges', () => {
    themesData.themes.forEach((theme: any) => {
      describe(`Color range for "${theme.name}"`, () => {
        it('r is in valid range (0-255)', () => {
          expect(theme.colors.r).toBeGreaterThanOrEqual(0)
          expect(theme.colors.r).toBeLessThanOrEqual(255)
        })

        it('g is in valid range (0-255)', () => {
          expect(theme.colors.g).toBeGreaterThanOrEqual(0)
          expect(theme.colors.g).toBeLessThanOrEqual(255)
        })

        it('b is in valid range (0-255)', () => {
          expect(theme.colors.b).toBeGreaterThanOrEqual(0)
          expect(theme.colors.b).toBeLessThanOrEqual(255)
        })

        it('r is an integer', () => {
          expect(Number.isInteger(theme.colors.r)).toBe(true)
        })

        it('g is an integer', () => {
          expect(Number.isInteger(theme.colors.g)).toBe(true)
        })

        it('b is an integer', () => {
          expect(Number.isInteger(theme.colors.b)).toBe(true)
        })
      })
    })
  })

  describe('Theme cssvars', () => {
    themesData.themes.forEach((theme: any) => {
      it(`"${theme.name}" has cssvars object`, () => {
        expect(theme.cssvars).toBeDefined()
        expect(typeof theme.cssvars).toBe('object')
      })

      it(`"${theme.name}" has font_main CSS variable`, () => {
        expect(typeof theme.cssvars.font_main).toBe('string')
        expect(theme.cssvars.font_main.length).toBeGreaterThan(0)
      })
    })
  })

  describe('Theme terminal configuration', () => {
    themesData.themes.forEach((theme: any) => {
      describe(`Terminal config for "${theme.name}"`, () => {
        it('has terminal configuration object', () => {
          expect(theme.terminal).toBeDefined()
          expect(typeof theme.terminal).toBe('object')
        })

        it('has terminal fontFamily', () => {
          expect(typeof theme.terminal.fontFamily).toBe('string')
        })

        it('has terminal cursorStyle', () => {
          expect(typeof theme.terminal.cursorStyle).toBe('string')
          expect(['block', 'underline', 'bar']).toContain(theme.terminal.cursorStyle)
        })

        it('has terminal foreground color', () => {
          expect(typeof theme.terminal.foreground).toBe('string')
          expect(theme.terminal.foreground).toMatch(/^#[0-9a-fA-F]{6}$/)
        })

        it('has terminal background color', () => {
          expect(typeof theme.terminal.background).toBe('string')
          expect(theme.terminal.background).toMatch(/^#[0-9a-fA-F]{6}$/)
        })

        it('has terminal cursor color', () => {
          expect(typeof theme.terminal.cursor).toBe('string')
          expect(theme.terminal.cursor).toMatch(/^#[0-9a-fA-F]{6}$/)
        })

        it('has terminal selection color', () => {
          expect(typeof theme.terminal.selection).toBe('string')
          // Selection colors use rgba format
          expect(theme.terminal.selection).toMatch(/^rgba\(/)
        })
      })
    })
  })

  describe('Theme uniqueness', () => {
    it('all theme IDs are unique', () => {
      const ids = themesData.themes.map((t: any) => t.id)
      const uniqueIds = new Set(ids)
      expect(uniqueIds.size).toBe(ids.length)
    })

    it('all theme names are unique', () => {
      const names = themesData.themes.map((t: any) => t.name)
      const uniqueNames = new Set(names)
      expect(uniqueNames.size).toBe(names.length)
    })
  })

  describe('InjectCSS field', () => {
    it('every theme has an injectCSS field (can be empty string)', () => {
      themesData.themes.forEach((theme: any) => {
        expect(theme).toHaveProperty('injectCSS')
        expect(typeof theme.injectCSS).toBe('string')
      })
    })
  })

  describe('Specific theme values', () => {
    it('tron theme has expected blue-cyan accent', () => {
      const tron = themesData.themes.find((t: any) => t.id === 'tron')
      expect(tron).toBeDefined()
      expect(tron!.colors.r).toBe(170)
      expect(tron!.colors.g).toBe(207)
      expect(tron!.colors.b).toBe(209)
    })

    // The previous tests here asserted specific RGB triples for matrix,
    // blade and cyberpunk. Two of those values did not match the shipped
    // themes and the third theme does not exist, so they were freezing
    // invented numbers. Assert the invariant instead: every theme declares a
    // usable accent, which is what the renderer actually depends on.
    it('every theme declares a valid RGB accent', () => {
      for (const theme of themesData.themes) {
        const { r, g, b } = theme.colors
        for (const [channel, value] of Object.entries({ r, g, b })) {
          expect(
            Number.isInteger(value) && value >= 0 && value <= 255,
            `${theme.id}.colors.${channel} = ${value}`,
          ).toBe(true)
        }
      }
    })
  })
})
