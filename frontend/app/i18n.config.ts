/**
 * Nuxt 3 i18n configuration
 */

import { defineI18nConfig } from './utils/i18n'

export default defineI18nConfig({
  // Default locale
  defaultLocale: 'en',

  // Fallback locale when translation is missing
  fallbackLocale: 'en',

  // Available locales
  locales: [
    {
      code: 'en',
      name: 'English',
      file: 'en.json'
    },
    {
      code: 'es',
      name: 'Español',
      file: 'es.json'
    },
    {
      code: 'fr',
      name: 'Français',
      file: 'fr.json'
    },
    {
      code: 'de',
      name: 'Deutsch',
      file: 'de.json'
    },
    {
      code: 'ja',
      name: '日本語',
      file: 'ja.json'
    },
    {
      code: 'zh',
      name: '中文',
      file: 'zh.json'
    },
    {
      code: 'ru',
      name: 'Русский',
      file: 'ru.json'
    },
    {
      code: 'pt',
      name: 'Português',
      file: 'pt.json'
    },
    {
      code: 'ar',
      name: 'العربية',
      file: 'ar.json'
    },
    {
      code: 'hi',
      name: 'हिन्दी',
      file: 'hi.json'
    }
  ],

  // Lazy load translations
  lazy: true,

  // Directory containing translation files
  langDir: 'i18n/locales',

  // Detect browser language
  detectBrowserLanguage: {
    useCookie: true,
    cookieKey: 'aDex-UI-i18n',
    redirectOn: 'root'
  },

  // Strategy for URL routing
  strategy: 'prefix_except_default',

  // Custom routes
  pages: {
    'about': {
      en: '/about',
      es: '/acerca-de',
      fr: '/a-propos',
      de: '/uber-uns',
      ja: '/について',
      zh: '/关于',
      ru: '/о-нас',
      pt: '/sobre',
      ar: '/حول',
      hi: '/के-बारे-में'
    },
    'settings': {
      en: '/settings',
      es: '/configuracion',
      fr: '/parametres',
      de: '/einstellungen',
      ja: '/設定',
      zh: '/设置',
      ru: '/настройки',
      pt: '/configuracoes',
      ar: '/الإعدادات',
      hi: '/सेटिंग्स'
    }
  },

  // Vue I18n options
  vueI18n: {
    legacy: false,
    globalInjection: true,
    sync: true,
    silentTranslationWarn: true,
    silentFallbackWarn: true,
    formatFallbackMessages: true,

    // DateTime formats
    datetimeFormats: {
      en: {
        short: {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        },
        long: {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          weekday: 'long',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      es: {
        short: {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        },
        long: {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          weekday: 'long',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      fr: {
        short: {
          day: 'numeric',
          month: 'short',
          year: 'numeric'
        },
        long: {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      de: {
        short: {
          day: 'numeric',
          month: 'short',
          year: 'numeric'
        },
        long: {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      ja: {
        short: {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        },
        long: {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          weekday: 'long',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      zh: {
        short: {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        },
        long: {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          weekday: 'long',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      ru: {
        short: {
          day: 'numeric',
          month: 'short',
          year: 'numeric'
        },
        long: {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      pt: {
        short: {
          day: 'numeric',
          month: 'short',
          year: 'numeric'
        },
        long: {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      ar: {
        short: {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        },
        long: {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      },
      hi: {
        short: {
          year: 'numeric',
          month: 'short',
          day: 'numeric'
        },
        long: {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: 'numeric',
          minute: 'numeric'
        },
        time: {
          hour: 'numeric',
          minute: 'numeric',
          second: 'numeric'
        }
      }
    },

    // Number formats
    numberFormats: {
      en: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        },
        currency: {
          style: 'currency',
          currency: 'USD'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      es: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'EUR'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      fr: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'EUR'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      de: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'EUR'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      ja: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'JPY'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      zh: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'CNY'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      ru: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'RUB'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      pt: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'BRL'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      ar: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'SAR'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      },
      hi: {
        decimal: {
          style: 'decimal',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2,
          useGrouping: true
        },
        currency: {
          style: 'currency',
          currency: 'INR'
        },
        percent: {
          style: 'percent',
          minimumFractionDigits: 0,
          maximumFractionDigits: 2
        }
      }
    }
  }
})