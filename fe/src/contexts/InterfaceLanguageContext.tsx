import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { languageApi } from '../api';
import { Language } from '../types';

interface InterfaceLanguageContextType {
  interfaceLanguage: string;
  supportedLanguages: string[];
  supportedLanguagesWithNames: Language[];
  changeInterfaceLanguage: (lang: string) => void;
  loading: boolean;
}

const InterfaceLanguageContext = createContext<InterfaceLanguageContextType | undefined>(undefined);

export const useInterfaceLanguage = (): InterfaceLanguageContextType => {
  const context = useContext(InterfaceLanguageContext);
  if (!context) {
    throw new Error('useInterfaceLanguage must be used within a InterfaceLanguageProvider');
  }
  return context;
};

interface InterfaceLanguageProviderProps {
  children: ReactNode;
}

export const InterfaceLanguageProvider: React.FC<InterfaceLanguageProviderProps> = ({ children }) => {
  const { i18n } = useTranslation();
  const [supportedLanguages, setSupportedLanguages] = useState<string[]>(['en', 'ar', 'ru']);
  const [supportedLanguagesWithNames, setSupportedLanguagesWithNames] = useState<Language[]>([]);
  const [loading, setLoading] = useState(false);

  // Load supported languages from backend
  useEffect(() => {
    const loadSupportedLanguages = async () => {
      try {
        setLoading(true);
        const languages = await languageApi.getLanguages(true); // Only active languages
        if (Array.isArray(languages)) {
          const codes = languages.map(lang => lang.language_code);
          setSupportedLanguages(codes.length > 0 ? codes : ['en', 'ar', 'ru']);
          setSupportedLanguagesWithNames(languages);
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

  // Initialize interface language from localStorage or default
  useEffect(() => {
    const savedLang = localStorage.getItem('interface_language');
    if (savedLang && supportedLanguages.includes(savedLang)) {
      i18n.changeLanguage(savedLang);
    } else if (!savedLang) {
      // No saved language, use default 'en'
      const defaultLang = 'en';
      localStorage.setItem('interface_language', defaultLang);
      i18n.changeLanguage(defaultLang);
    }
  }, [i18n, supportedLanguages]);

  const changeInterfaceLanguage = (lang: string) => {
    if (!supportedLanguages.includes(lang)) {
      console.warn(`Language ${lang} is not supported. Falling back to English.`);
      lang = 'en';
    }

    localStorage.setItem('interface_language', lang);
    i18n.changeLanguage(lang);

    // Update document direction for RTL languages
    document.documentElement.dir = lang === 'ar' ? 'rtl' : 'ltr';
    document.documentElement.lang = lang;
  };

  const value: InterfaceLanguageContextType = {
    interfaceLanguage: i18n.language,
    supportedLanguages,
    supportedLanguagesWithNames,
    changeInterfaceLanguage,
    loading,
  };

  return <InterfaceLanguageContext.Provider value={value}>{children}</InterfaceLanguageContext.Provider>;
};
