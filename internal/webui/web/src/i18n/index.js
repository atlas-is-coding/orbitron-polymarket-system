import { createI18n } from 'vue-i18n'
import en from './en.json'
import ru from './ru.json'
import zh from './zh.json'
import ja from './ja.json'
import ko from './ko.json'

const saved = localStorage.getItem('lang') || 'en'

export const i18n = createI18n({
  legacy: false,
  locale: saved,
  fallbackLocale: 'en',
  messages: { en, ru, zh, ja, ko }
})

export const LANGS = ['en', 'ru', 'zh', 'ja', 'ko']
