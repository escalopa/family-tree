import { apiClient } from './client';
import { ApiResponse } from '../types/api';
import { User } from '../types/user';

interface AuthResponse {
  user: User;
}

export type OAuthProvider = 'google' | 'facebook' | 'github';

export const authApi = {
  /**
   * Get OAuth authentication URL for a specific provider
   * @param provider - The OAuth provider (e.g., 'google', 'facebook', 'github')
   */
  getAuthURL: async (provider: OAuthProvider): Promise<{ url: string }> => {
    const response = await apiClient.get<ApiResponse<{ url: string }>>(`/auth/${provider}`);
    return response.data.data!;
  },

  /**
   * Legacy method for backwards compatibility
   * @deprecated Use getAuthURL('google') instead
   */
  getGoogleAuthURL: async (): Promise<{ url: string }> => {
    return authApi.getAuthURL('google');
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout');
  },

  logoutAll: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout-all');
  },
};

