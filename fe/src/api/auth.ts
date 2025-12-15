import { apiClient } from './client';
import { AuthResponse } from '../types';

export const authApi = {
  getAuthURL: async (provider: string): Promise<{ url: string; provider: string }> => {
    const response = await apiClient.get(`/auth/${provider}`);
    return response.data;
  },

  handleCallback: async (provider: string, code: string, state: string): Promise<AuthResponse> => {
    const response = await apiClient.get(`/auth/${provider}/callback`, {
      params: { code, state },
    });
    return response.data;
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout');
  },

  logoutAll: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout-all');
  },
};
