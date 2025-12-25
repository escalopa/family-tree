import { apiClient } from './client';
import { Language, UserLanguagePreference } from '../types';

// Language API
// Note: Languages are static and managed via i18n translation files
// Languages cannot be created or deleted, but can be enabled/disabled
export const languageApi = {
  // Get all languages
  getLanguages: async (activeOnly = false): Promise<Language[]> => {
    const params = activeOnly ? { active: 'true' } : {};
    const response = await apiClient.get('/api/languages', { params });
    return response.data.data || response.data;
  },

  // Get a specific language
  getLanguage: async (code: string): Promise<Language> => {
    const response = await apiClient.get(`/api/languages/${code}`);
    return response.data.data || response.data;
  },

  // Toggle language active status (Super Admin only)
  toggleLanguageActive: async (
    code: string,
    isActive: boolean
  ): Promise<Language> => {
    const response = await apiClient.patch(`/api/languages/${code}/toggle`, { is_active: isActive });
    return response.data.data || response.data;
  },

  // Update language display order (Super Admin only)
  updateLanguageOrder: async (
    languages: { language_code: string; display_order: number }[]
  ): Promise<void> => {
    await apiClient.put('/api/languages/order', { languages });
  },
};

// User Language Preference API
export const userLanguagePreferenceApi = {
  // Get current user's language preference
  getPreferences: async (): Promise<UserLanguagePreference> => {
    const response = await apiClient.get('/api/users/me/preferences/languages');
    return response.data.data || response.data;
  },

  // Update current user's language preference
  updatePreferences: async (data: {
    preferred_language: string;
  }): Promise<UserLanguagePreference> => {
    const response = await apiClient.put('/api/users/me/preferences/languages', data);
    return response.data.data || response.data;
  },
};
