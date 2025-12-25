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
    return localStorage.getItem('interface_language') || 'en';
  },
  cacheUserLanguage(lng: string) {
    localStorage.setItem('interface_language', lng);
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

// Update HTML dir attribute when language changes
i18n.on('languageChanged', (lng) => {
  const dir = lng === 'ar' ? 'rtl' : 'ltr';
  document.documentElement.dir = dir;
  document.documentElement.lang = lng;
});

// Set initial direction
const initialLang = i18n.language;
const initialDir = initialLang === 'ar' ? 'rtl' : 'ltr';
document.documentElement.dir = initialDir;
document.documentElement.lang = initialLang;

export default i18n;
