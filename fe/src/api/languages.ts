import { apiClient } from './client';
import { Language, UserLanguagePreference } from '../types';

// Language API
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

  // Create a new language (Super Admin only)
  createLanguage: async (data: {
    language_code: string;
    language_name: string;
    display_order?: number;
  }): Promise<Language> => {
    const response = await apiClient.post('/api/languages', data);
    return response.data.data || response.data;
  },

  // Update a language (Super Admin only)
  updateLanguage: async (
    code: string,
    data: {
      language_name: string;
      is_active: boolean;
      display_order: number;
    }
  ): Promise<Language> => {
    const response = await apiClient.put(`/api/languages/${code}`, data);
    return response.data.data || response.data;
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
