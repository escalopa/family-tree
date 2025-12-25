import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { Language, UserLanguagePreference } from '../types';
import { languageApi, userLanguagePreferenceApi } from '../api';
import { useAuth } from './AuthContext';

type NameProvider = { names?: Record<string, string> } | Record<string, string> | undefined;

interface LanguageContextType {
  languages: Language[];
  preferences: UserLanguagePreference | null;
  loading: boolean;
  error: string | null;
  loadLanguages: () => Promise<void>;
  loadPreferences: () => Promise<void>;
  updatePreferences: (preferred: string) => Promise<void>;
  getPreferredName: (obj: NameProvider, fallback?: string) => string;
  getAllNamesFormatted: (obj: NameProvider, separator?: string) => string;
  getAllNames: (obj: NameProvider) => Record<string, string>;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const useLanguage = (): LanguageContextType => {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};

interface LanguageProviderProps {
  children: ReactNode;
}

export const LanguageProvider: React.FC<LanguageProviderProps> = ({ children }) => {
  const { t } = useTranslation();
  const { isAuthenticated, loading: authLoading } = useAuth();
  const [languages, setLanguages] = useState<Language[]>([]);
  const [preferences, setPreferences] = useState<UserLanguagePreference | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load languages (active only)
  const loadLanguages = async () => {
    try {
      setLoading(true);
      setError(null);
      const langs = await languageApi.getLanguages(true);
      setLanguages(langs);
    } catch (err: any) {
      setError(t('apiErrors.failedToLoadLanguages'));

    } finally {
      setLoading(false);
    }
  };

  // Load user's language preference
  const loadPreferences = async () => {
    try {
      setLoading(true);
      setError(null);
      const prefs = await userLanguagePreferenceApi.getPreferences();
      setPreferences(prefs);

      // Save to localStorage for offline access
      localStorage.setItem('language_preferences', JSON.stringify(prefs));
    } catch (err: any) {
      // Try to load from localStorage as fallback
      const savedPrefs = localStorage.getItem('language_preferences');
      if (savedPrefs) {
        setPreferences(JSON.parse(savedPrefs));
      } else {
        // Default to English
        setPreferences({
          preferred_language: 'en',
        });
      }

      setError(t('apiErrors.failedToLoadPreferences'));

    } finally {
      setLoading(false);
    }
  };

  // Update user's language preference
  const updatePreferences = async (preferred: string) => {
    try {
      setLoading(true);
      setError(null);
      const updated = await userLanguagePreferenceApi.updatePreferences({
        preferred_language: preferred,
      });
      setPreferences(updated);

      // Update localStorage
      localStorage.setItem('language_preferences', JSON.stringify(updated));
    } catch (err: any) {
      setError(t('apiErrors.failedToUpdatePreferences'));

      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Extract names from object or use directly if it's a record
  const extractNames = (obj: NameProvider): Record<string, string> | undefined => {
    if (!obj) return undefined;

    // Check if it has a 'names' property (it's a Member or similar object)
    if (typeof obj === 'object' && 'names' in obj) {
      return obj.names as Record<string, string> | undefined;
    }

    // Otherwise, treat obj as Record<string, string> directly
    return obj as Record<string, string>;
  };

  // Get all names as an object
  const getAllNames = (obj: NameProvider): Record<string, string> => {
    const names = extractNames(obj);
    return names || {};
  };

  // Get name in preferred language with fallback
  const getPreferredName = (obj: NameProvider, fallback?: string): string => {
    const defaultFallback = fallback || 'N/A';
    const names = extractNames(obj);
    if (!names || Object.keys(names).length === 0) {
      return defaultFallback;
    }

    const preferredLang = preferences?.preferred_language || 'ar';
    return names[preferredLang] || names['ar'] || names['en'] || Object.values(names)[0] || defaultFallback;
  };

  // Get all names formatted for display (e.g., "أحمد | Ahmed | Ахмад")
  const getAllNamesFormatted = (obj: NameProvider, separator = ' | '): string => {
    const names = extractNames(obj);
    if (!names || Object.keys(names).length === 0) {
      return 'N/A';
    }

    // Sort languages by display order
    const sortedLanguages = languages
      .filter(lang => names[lang.language_code])
      .sort((a, b) => a.display_order - b.display_order);

    if (sortedLanguages.length === 0) {
      // Fallback: just show all names in any order
      return Object.values(names).filter(name => name).join(separator) || 'N/A';
    }

    return sortedLanguages.map(lang => names[lang.language_code]).join(separator);
  };

  // Initialize only when authenticated
  useEffect(() => {
    // Wait for auth to complete
    if (authLoading) {
      return;
    }

    // Only load if authenticated
    if (isAuthenticated) {
      loadLanguages();
      loadPreferences();
    } else {
      // Set default preferences for unauthenticated users
      setPreferences({
        preferred_language: 'en',
      });
    }
  }, [isAuthenticated, authLoading]);

  const value: LanguageContextType = {
    languages,
    preferences,
    loading,
    error,
    loadLanguages,
    loadPreferences,
    updatePreferences,
    getPreferredName,
    getAllNamesFormatted,
    getAllNames,
  };

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
};
