// Every locale must carry exactly the keys English does, and every key the
// app asks for must exist. A missing key would otherwise show English (or the
// raw key) to a user who picked another language, with nothing failing.

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { LOCALES } from '~/composables/useI18n'

type Messages = { [key: string]: string | Messages }
const PLURAL = /^(zero|one|two|few|many|other)$/

function keys(m: Messages, prefix = ''): string[] {
  return Object.entries(m).flatMap(([k, v]) => {
    const path = prefix ? `${prefix}.${k}` : k
    if (typeof v === 'string') return [path]
    // A plural message is one key, whichever forms a language needs.
    if (Object.keys(v).every(f => PLURAL.test(f))) return [path]
    return keys(v, path)
  })
}

function placeholders(m: Messages, key: string): string[] {
  let node: string | Messages = m
  for (const p of key.split('.')) node = (node as Messages)[p]!
  const text = typeof node === 'string' ? node : Object.values(node).join(' ')
  return [...new Set(text.match(/\{\w+\}/g) ?? [])].sort()
}

function sources(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name)
    if (statSync(path).isDirectory()) return sources(path)
    return /\.(vue|ts)$/.test(name) ? [path] : []
  })
}

const en = LOCALES.en.messages
const enKeys = keys(en).sort()

describe('locales', () => {
  for (const [code, { messages }] of Object.entries(LOCALES)) {
    if (code === 'en') continue

    it(`${code} has exactly the English keys`, () => {
      expect(keys(messages).sort()).toEqual(enKeys)
    })

    it(`${code} keeps every {placeholder}`, () => {
      for (const key of enKeys) {
        expect(placeholders(messages, key), `${code}: ${key}`).toEqual(placeholders(en, key))
      }
    })
  }

  it('every t() key used by the app exists in English', () => {
    const root = join(import.meta.dirname, '../../app')
    const missing: string[] = []
    // useI18n.ts itself only mentions keys in its usage comment.
    for (const file of sources(root).filter(f => !f.endsWith('useI18n.ts'))) {
      for (const [, key] of readFileSync(file, 'utf8').matchAll(/\bt\(\s*['"`]([\w.]+)['"`]/g)) {
        if (!enKeys.includes(key!)) missing.push(`${file.slice(root.length + 1)}: ${key}`)
      }
    }
    expect(missing).toEqual([])
  })
})

describe('useI18n', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.resetModules()
  })

  it('follows the chosen language and falls back to English', async () => {
    localStorage.setItem('adex-settings', JSON.stringify({ display: { language: 'fr' } }))
    const { useI18n } = await import('~/composables/useI18n')
    const { t, locale } = useI18n()
    expect(locale.value).toBe('fr')
    expect(t('settings.display.language')).toBe('Langue')
    expect(t('no.such.key')).toBe('no.such.key')
  })

  it('ignores an unknown language', async () => {
    localStorage.setItem('adex-settings', JSON.stringify({ display: { language: 'xx' } }))
    const { useI18n } = await import('~/composables/useI18n')
    expect(['en', 'es', 'fr', 'ja', 'zh']).toContain(useI18n().locale.value)
  })
})
