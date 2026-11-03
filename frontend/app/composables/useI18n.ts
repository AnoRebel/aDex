// composables/useI18n.ts
//
// Interface translations. Every locale is bundled (they are small), so
// switching language is instant and works offline.
//
// The language lives in `adex-settings.display.language`, persisted to disk
// with the rest of the settings. Empty means "follow the system language".
//
//   t('settings.title')                    → "SETTINGS"
//   t('updates.progress', { done, total }) → "{done} / {total}" filled in
//
// Plurals: a message may be an object keyed by Intl.PluralRules categories
// ({ "one": "...", "other": "..." }); t() picks one from params.count.
// A key missing from a locale falls back to English, then to the key itself.

import { computed } from 'vue'
import { useStorage, usePreferredLanguages } from '@vueuse/core'
import en from '../i18n/locales/en.json'
import es from '../i18n/locales/es.json'
import fr from '../i18n/locales/fr.json'
import ja from '../i18n/locales/ja.json'
import zh from '../i18n/locales/zh.json'

type Messages = { [key: string]: string | Messages }

export const LOCALES = {
  en: { name: 'English', messages: en as Messages },
  es: { name: 'Español', messages: es as Messages },
  fr: { name: 'Français', messages: fr as Messages },
  ja: { name: '日本語', messages: ja as Messages },
  zh: { name: '中文', messages: zh as Messages },
} as const

export type LocaleCode = keyof typeof LOCALES

const settings = useStorage<{ display?: { language?: string } }>('adex-settings', {})
const preferred = usePreferredLanguages()

function isLocale(code: string | undefined): code is LocaleCode {
  return !!code && Object.hasOwn(LOCALES, code)
}

/** The active locale: the user's choice, else the first supported system language. */
const locale = computed<LocaleCode>(() => {
  const chosen = settings.value.display?.language
  if (isLocale(chosen)) return chosen
  for (const tag of preferred.value) {
    const base = tag.toLowerCase().split('-')[0]
    if (isLocale(base)) return base
  }
  return 'en'
})

function lookup(messages: Messages, key: string): string | Messages | undefined {
  let node: string | Messages | undefined = messages
  for (const part of key.split('.')) {
    if (typeof node !== 'object') return undefined
    node = node[part]
  }
  return node
}

type Params = Record<string, string | number>

function resolve(code: LocaleCode, key: string, params?: Params): string | undefined {
  let msg = lookup(LOCALES[code].messages, key)
  if (msg && typeof msg === 'object') {
    const count = Number(params?.count ?? 0)
    const form = new Intl.PluralRules(code).select(count)
    msg = msg[form] ?? msg.other
  }
  return typeof msg === 'string' ? msg : undefined
}

export function translate(key: string, params?: Params): string {
  const msg = resolve(locale.value, key, params) ?? resolve('en', key, params) ?? key
  return params ? msg.replace(/\{(\w+)\}/g, (m, name) => (name in params ? String(params[name]) : m)) : msg
}

export function useI18n() {
  return { t: translate, locale, locales: LOCALES }
}
