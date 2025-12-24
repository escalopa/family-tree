import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from './locales/en.json';
import ar from './locales/ar.json';
import ru from './locales/ru.json';

const resources = {
  en: { translation: en },
  ar: { translation: ar },
  ru: { translation: ru },
};

// Custom language detector that reads from localStorage
const customDetector = {
  name: 'customLocalStorage',
  lookup() {
    return localStorage.getItem('ui_language') || 'en';
  },
  cacheUserLanguage(lng: string) {
    localStorage.setItem('ui_language', lng);
  },
};

i18n
  .use({
    type: 'languageDetector',
    detect: customDetector.lookup,
    init: () => {},
    cacheUserLanguage: customDetector.cacheUserLanguage,
  })
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    supportedLngs: ['en', 'ar', 'ru'],
    interpolation: {
      escapeValue: false, // React already escapes
    },
    detection: {
      order: ['customLocalStorage', 'navigator'],
      caches: ['customLocalStorage'],
    },
  });

export default i18n;
