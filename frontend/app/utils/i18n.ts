/**
 * Internationalization utilities for multi-language support
 */

import { useStorage, usePreferredLanguages } from '@vueuse/core'
import type { Ref } from 'vue'

export interface Locale {
  code: string
  name: string
  nativeName: string
  rtl: boolean
  dateFormat: string
  timeFormat: string
  numberFormat: Intl.NumberFormatOptions
  currencyFormat: Intl.NumberFormatOptions
}

export interface I18nConfig {
  defaultLocale: string
  fallbackLocale: string
  lazy: boolean
  langDir: string
  detectBrowserLanguage: boolean
  strategy: 'prefix' | 'prefix_except_default' | 'suffix' | 'no_prefix'
  differentDomains: boolean
  trailingSlash: boolean
  locales: string[]
  vueI18n: {
    legacy: boolean
    locale: string
    fallbackLocale: string
    messages: Record<string, any>
    datetimeFormats: Record<string, any>
    numberFormats: Record<string, any>
  }
}

export interface TranslationNamespace {
  [key: string]: string | TranslationNamespace
}

export class I18nManager {
  private static instance: I18nManager
  private config: Ref<I18nConfig>
  private currentLocale: Ref<string>
  private preferredLanguages: Ref<string[]>
  private loadedTranslations: Map<string, TranslationNamespace> = new Map()
  private formatters: Map<string, Intl.DateTimeFormat> = new Map()
  private numberFormatters: Map<string, Intl.NumberFormat> = new Map()

  private constructor() {
    this.config = useStorage('i18n-config', this.getDefaultConfig())
    this.currentLocale = useStorage('i18n-current-locale', this.config.value.defaultLocale)
    this.preferredLanguages = usePreferredLanguages()

    this.loadLocale(this.currentLocale.value)
  }

  static getInstance(): I18nManager {
    if (!I18nManager.instance) {
      I18nManager.instance = new I18nManager()
    }
    return I18nManager.instance
  }

  /**
   * Get current locale
   */
  getCurrentLocale(): string {
    return this.currentLocale.value
  }

  /**
   * Set current locale
   */
  setLocale(locale: string): void {
    if (this.supportedLocales.includes(locale)) {
      this.currentLocale.value = locale
      this.loadLocale(locale)
      this.updateDocumentLanguage(locale)
      this.updateDocumentDirection(locale)
    } else {
      console.warn(`Unsupported locale: ${locale}`)
    }
  }

  /**
   * Get supported locales
   */
  get supportedLocales(): string[] {
    return this.config.value.locales
  }

  /**
   * Get locale information
   */
  getLocaleInfo(locale: string): Locale | null {
    const locales: Record<string, Locale> = {
      'en': {
        code: 'en',
        name: 'English',
        nativeName: 'English',
        rtl: false,
        dateFormat: 'MM/DD/YYYY',
        timeFormat: '12h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        },
        currencyFormat: {
          style: 'currency',
          currency: 'USD'
        }
      },
      'es': {
        code: 'es',
        name: 'Spanish',
        nativeName: 'Español',
        rtl: false,
        dateFormat: 'DD/MM/YYYY',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'EUR'
        }
      },
      'fr': {
        code: 'fr',
        name: 'French',
        nativeName: 'Français',
        rtl: false,
        dateFormat: 'DD/MM/YYYY',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'EUR'
        }
      },
      'ar': {
        code: 'ar',
        name: 'Arabic',
        nativeName: 'العربية',
        rtl: true,
        dateFormat: 'DD/MM/YYYY',
        timeFormat: '12h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'SAR'
        }
      },
      'zh': {
        code: 'zh',
        name: 'Chinese',
        nativeName: '中文',
        rtl: false,
        dateFormat: 'YYYY-MM-DD',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'CNY'
        }
      },
      'ja': {
        code: 'ja',
        name: 'Japanese',
        nativeName: '日本語',
        rtl: false,
        dateFormat: 'YYYY/MM/DD',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'JPY'
        }
      },
      'de': {
        code: 'de',
        name: 'German',
        nativeName: 'Deutsch',
        rtl: false,
        dateFormat: 'DD.MM.YYYY',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'EUR'
        }
      },
      'ru': {
        code: 'ru',
        name: 'Russian',
        nativeName: 'Русский',
        rtl: false,
        dateFormat: 'DD.MM.YYYY',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'RUB'
        }
      },
      'pt': {
        code: 'pt',
        name: 'Portuguese',
        nativeName: 'Português',
        rtl: false,
        dateFormat: 'DD/MM/YYYY',
        timeFormat: '24h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'BRL'
        }
      },
      'hi': {
        code: 'hi',
        name: 'Hindi',
        nativeName: 'हिन्दी',
        rtl: false,
        dateFormat: 'DD/MM/YYYY',
        timeFormat: '12h',
        numberFormat: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currencyFormat: {
          style: 'currency',
          currency: 'INR'
        }
      }
    }

    return locales[locale] || null
  }

  /**
   * Translate a key
   */
  t(key: string, locale?: string, params?: Record<string, any>): string {
    const targetLocale = locale || this.currentLocale.value
    const translations = this.loadedTranslations.get(targetLocale)

    if (!translations) {
      console.warn(`No translations loaded for locale: ${targetLocale}`)
      return key
    }

    const value = this.getNestedValue(translations, key)
    if (typeof value !== 'string') {
      console.warn(`Translation key not found: ${key}`)
      return key
    }

    // Interpolate parameters
    if (params) {
      return this.interpolate(value, params)
    }

    return value
  }

  /**
   * Format date according to locale
   */
  formatDate(date: Date | number | string, locale?: string, options?: Intl.DateTimeFormatOptions): string {
    const targetLocale = locale || this.currentLocale.value
    const formatter = this.getDateTimeFormatter(targetLocale, options)
    return formatter.format(new Date(date))
  }

  /**
   * Format number according to locale
   */
  formatNumber(number: number, locale?: string, options?: Intl.NumberFormatOptions): string {
    const targetLocale = locale || this.currentLocale.value
    const formatter = this.getNumberFormatter(targetLocale, options)
    return formatter.format(number)
  }

  /**
   * Format currency according to locale
   */
  formatCurrency(amount: number, currency?: string, locale?: string): string {
    const targetLocale = locale || this.currentLocale.value
    const localeInfo = this.getLocaleInfo(targetLocale)

    if (!localeInfo) {
      return amount.toString()
    }

    const options = {
      ...localeInfo.currencyFormat,
      currency: currency || localeInfo.currencyFormat.currency
    }

    const formatter = this.getNumberFormatter(targetLocale, options)
    return formatter.format(amount)
  }

  /**
   * Format relative time
   */
  formatRelativeTime(value: number, unit: Intl.RelativeTimeFormatUnit, locale?: string): string {
    const targetLocale = locale || this.currentLocale.value
    const formatter = new Intl.RelativeTimeFormat(targetLocale, { numeric: 'auto' })
    return formatter.format(value, unit)
  }

  /**
   * Get plural form for a count
   */
  pluralize(count: number, key: string, locale?: string): string {
    const targetLocale = locale || this.currentLocale.value
    const pluralKey = `${key}.${this.getPluralRule(count, targetLocale)}`
    return this.t(pluralKey, targetLocale, { count })
  }

  /**
   * Detect browser language
   */
  detectBrowserLanguage(): string | null {
    if (!this.config.value.detectBrowserLanguage) {
      return null
    }

    for (const browserLang of this.preferredLanguages.value) {
      const langCode = browserLang.split('-')[0]

      // Check exact match
      if (this.supportedLocales.includes(browserLang)) {
        return browserLang
      }

      // Check language code match
      if (this.supportedLocales.includes(langCode)) {
        return langCode
      }
    }

    return null
  }

  /**
   * Get text direction for locale
   */
  getTextDirection(locale?: string): 'ltr' | 'rtl' {
    const targetLocale = locale || this.currentLocale.value
    const localeInfo = this.getLocaleInfo(targetLocale)
    return localeInfo?.rtl ? 'rtl' : 'ltr'
  }

  /**
   * Check if locale is RTL
   */
  isRTL(locale?: string): boolean {
    const targetLocale = locale || this.currentLocale.value
    const localeInfo = this.getLocaleInfo(targetLocale)
    return localeInfo?.rtl || false
  }

  /**
   * Load translations for a locale
   */
  async loadLocale(locale: string): Promise<void> {
    if (this.loadedTranslations.has(locale)) {
      return
    }

    try {
      // In a real implementation, this would load from files or API
      // For now, we'll use the imported translations
      const translations = await this.loadTranslationsFromFiles(locale)
      this.loadedTranslations.set(locale, translations)
    } catch (error) {
      console.error(`Failed to load translations for locale: ${locale}`, error)
    }
  }

  /**
   * Get all available locales
   */
  getAvailableLocales(): Locale[] {
    return this.supportedLocales.map(code => this.getLocaleInfo(code)).filter(Boolean) as Locale[]
  }

  /**
   * Update i18n configuration
   */
  updateConfig(config: Partial<I18nConfig>): void {
    this.config.value = { ...this.config.value, ...config }
  }

  // Private helper methods

  private getDefaultConfig(): I18nConfig {
    return {
      defaultLocale: 'en',
      fallbackLocale: 'en',
      lazy: true,
      langDir: 'i18n',
      detectBrowserLanguage: true,
      strategy: 'prefix_except_default',
      differentDomains: false,
      trailingSlash: false,
      locales: ['en', 'es', 'fr', 'de', 'ja', 'zh', 'ru', 'pt', 'ar', 'hi'],
      vueI18n: {
        legacy: false,
        locale: 'en',
        fallbackLocale: 'en',
        messages: {},
        datetimeFormats: {},
        numberFormats: {}
      }
    }
  }

  private async loadTranslationsFromFiles(locale: string): Promise<TranslationNamespace> {
    // This would normally load from the filesystem or API
    // For this demo, we'll return a basic structure
    try {
      const module = await import(`~/i18n/locales/${locale}.json`)
      return module.default || {}
    } catch (error) {
      console.warn(`Could not load translations for locale: ${locale}`)
      return {}
    }
  }

  private getNestedValue(obj: any, path: string): any {
    return path.split('.').reduce((current, key) => current?.[key], obj)
  }

  private interpolate(template: string, params: Record<string, any>): string {
    return template.replace(/\{(\w+)\}/g, (match, key) => {
      return params[key] !== undefined ? String(params[key]) : match
    })
  }

  private getPluralRule(count: number, locale: string): string {
    // Simple plural rules - in real implementation, use Intl.PluralRules
    const rules: Record<string, (count: number) => string> = {
      'en': (count) => count === 1 ? 'one' : 'other',
      'es': (count) => count === 1 ? 'one' : 'other',
      'fr': (count) => count === 0 || count === 1 ? 'one' : 'other',
      'ar': (count) => {
        if (count === 0) return 'zero'
        if (count === 1) return 'one'
        if (count === 2) return 'two'
        if (count % 100 >= 3 && count % 100 <= 10) return 'few'
        if (count % 100 >= 11 && count % 100 <= 99) return 'many'
        return 'other'
      },
      'zh': () => 'other',
      'ja': () => 'other',
      'ko': () => 'other',
      'de': (count) => count === 1 ? 'one' : 'other',
      'ru': (count) => {
        const mod10 = count % 10
        const mod100 = count % 100
        if (mod10 === 1 && mod100 !== 11) return 'one'
        if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return 'few'
        return 'many'
      }
    }

    const rule = rules[locale.split('-')[0]] || rules['en']
    return rule(count)
  }

  private getDateTimeFormatter(locale: string, options?: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
    const key = `${locale}-${JSON.stringify(options || {})}`

    if (!this.formatters.has(key)) {
      const defaultOptions: Intl.DateTimeFormatOptions = {
        dateStyle: 'medium',
        timeStyle: 'medium'
      }
      this.formatters.set(key, new Intl.DateTimeFormat(locale, options || defaultOptions))
    }

    return this.formatters.get(key)!
  }

  private getNumberFormatter(locale: string, options?: Intl.NumberFormatOptions): Intl.NumberFormat {
    const key = `${locale}-${JSON.stringify(options || {})}`

    if (!this.numberFormatters.has(key)) {
      const defaultOptions: Intl.NumberFormatOptions = {
        style: 'decimal',
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
      }
      this.numberFormatters.set(key, new Intl.NumberFormat(locale, options || defaultOptions))
    }

    return this.numberFormatters.get(key)!
  }

  private updateDocumentLanguage(locale: string): void {
    if (typeof document !== 'undefined') {
      document.documentElement.lang = locale
    }
  }

  private updateDocumentDirection(locale: string): void {
    if (typeof document !== 'undefined') {
      document.documentElement.dir = this.getTextDirection(locale)
    }
  }
}

/**
 * Vue composable for internationalization
 */
export function useI18n() {
  const i18n = I18nManager.getInstance()

  const t = (key: string, params?: Record<string, any>) => {
    return i18n.t(key, undefined, params)
  }

  const setLocale = (locale: string) => {
    return i18n.setLocale(locale)
  }

  const getLocale = () => {
    return i18n.getCurrentLocale()
  }

  const formatDate = (date: Date | number | string, options?: Intl.DateTimeFormatOptions) => {
    return i18n.formatDate(date, undefined, options)
  }

  const formatNumber = (number: number, options?: Intl.NumberFormatOptions) => {
    return i18n.formatNumber(number, undefined, options)
  }

  const formatCurrency = (amount: number, currency?: string) => {
    return i18n.formatCurrency(amount, currency)
  }

  const formatRelativeTime = (value: number, unit: Intl.RelativeTimeFormatUnit) => {
    return i18n.formatRelativeTime(value, unit)
  }

  const pluralize = (count: number, key: string) => {
    return i18n.pluralize(count, key)
  }

  const isRTL = () => {
    return i18n.isRTL()
  }

  const getTextDirection = () => {
    return i18n.getTextDirection()
  }

  const getAvailableLocales = () => {
    return i18n.getAvailableLocales()
  }

  const detectBrowserLanguage = () => {
    return i18n.detectBrowserLanguage()
  }

  return {
    t,
    setLocale,
    getLocale,
    formatDate,
    formatNumber,
    formatCurrency,
    formatRelativeTime,
    pluralize,
    isRTL,
    getTextDirection,
    getAvailableLocales,
    detectBrowserLanguage,
    supportedLocales: i18n.supportedLocales
  }
}

/**
 * Nuxt 3 compatible i18n plugin configuration
 */
export const defineI18nConfig = (userConfig: Partial<I18nConfig> = {}) => {
  const i18n = I18nManager.getInstance()
  i18n.updateConfig(userConfig)

  return i18n.getConfig().value
}

/**
 * Translation directive for Vue
 */
export const vT = {
  mounted(el: HTMLElement, binding: any) {
    const { value, modifiers } = binding
    const { t } = useI18n()

    if (modifiers.plural && typeof value === 'object') {
      el.textContent = t(value.key, { count: value.count })
    } else {
      el.textContent = t(value)
    }
  },

  updated(el: HTMLElement, binding: any) {
    const { value, modifiers } = binding
    const { t } = useI18n()

    if (modifiers.plural && typeof value === 'object') {
      el.textContent = t(value.key, { count: value.count })
    } else {
      el.textContent = t(value)
    }
  }
}

/**
 * RTL/LTR utilities
 */
export const rtlUtils = {
  getDirection: () => {
    const i18n = I18nManager.getInstance()
    return i18n.getTextDirection()
  },

  isRTL: () => {
    const i18n = I18nManager.getInstance()
    return i18n.isRTL()
  },

  setDirection: (element: HTMLElement, isRTL?: boolean) => {
    const rtl = isRTL !== undefined ? isRTL : rtlUtils.isRTL()
    element.style.direction = rtl ? 'rtl' : 'ltr'
    element.setAttribute('dir', rtl ? 'rtl' : 'ltr')
  }
}

/**
 * Global types for TypeScript
 */
declare global {
  interface Window {
    __aDex_UI_I18N__?: I18nManager
  }
}