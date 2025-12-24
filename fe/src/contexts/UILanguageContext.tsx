import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { languageApi } from '../api';

interface UILanguageContextType {
  uiLanguage: string;
  supportedLanguages: string[];
  changeUILanguage: (lang: string) => void;
  loading: boolean;
}

const UILanguageContext = createContext<UILanguageContextType | undefined>(undefined);

export const useUILanguage = (): UILanguageContextType => {
  const context = useContext(UILanguageContext);
  if (!context) {
    throw new Error('useUILanguage must be used within a UILanguageProvider');
  }
  return context;
};

interface UILanguageProviderProps {
  children: ReactNode;
}

export const UILanguageProvider: React.FC<UILanguageProviderProps> = ({ children }) => {
  const { i18n } = useTranslation();
  const [supportedLanguages, setSupportedLanguages] = useState<string[]>(['en', 'ar', 'ru']);
  const [loading, setLoading] = useState(false);

  // Load supported languages from backend
  useEffect(() => {
    const loadSupportedLanguages = async () => {
      try {
        setLoading(true);
        const languages = await languageApi.getLanguages(true);
        if (Array.isArray(languages)) {
          const codes = languages.map(lang => lang.language_code);
          setSupportedLanguages(codes.length > 0 ? codes : ['en', 'ar', 'ru']);
        }
      } catch (err) {
        console.error('Failed to load supported languages:', err);
        // Keep default supported languages
      } finally {
        setLoading(false);
      }
    };

    loadSupportedLanguages();
  }, []);

  // Initialize UI language from localStorage or default
  useEffect(() => {
    const savedLang = localStorage.getItem('ui_language');
    if (savedLang && supportedLanguages.includes(savedLang)) {
      i18n.changeLanguage(savedLang);
    } else if (!savedLang) {
      // No saved language, use default 'en'
      const defaultLang = 'en';
      localStorage.setItem('ui_language', defaultLang);
      i18n.changeLanguage(defaultLang);
    }
  }, [i18n, supportedLanguages]);

  const changeUILanguage = (lang: string) => {
    if (!supportedLanguages.includes(lang)) {
      console.warn(`Language ${lang} is not supported. Falling back to English.`);
      lang = 'en';
    }

    localStorage.setItem('ui_language', lang);
    i18n.changeLanguage(lang);

    // Update document direction for RTL languages
    document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr';
    document.documentElement.lang = lang;
  };

  const value: UILanguageContextType = {
    uiLanguage: i18n.language,
    supportedLanguages,
    changeUILanguage,
    loading,
  };

  return <UILanguageContext.Provider value={value}>{children}</UILanguageContext.Provider>;
};
