import { apiClient } from './client';
import { ApiResponse } from '../types/api';

export type OAuthProvider = 'google' | 'facebook' | 'github';

export const authApi = {
  /**
   * Get OAuth authentication URL for a specific provider
   * @param provider - The OAuth provider (e.g., 'google', 'facebook', 'github')
   */
  getAuthURL: async (provider: OAuthProvider): Promise<{ url: string; provider: string }> => {
    const response = await apiClient.get<ApiResponse<{ url: string; provider: string }>>(`/auth/${provider}`);
    return response.data.data!;
  },

  /**
   * Legacy method for backwards compatibility
   * @deprecated Use getAuthURL('google') instead
   */
  getGoogleAuthURL: async (): Promise<{ url: string; provider: string }> => {
    return authApi.getAuthURL('google');
  },

  /**
   * Handle OAuth callback by sending the authorization code to the backend
   * @param provider - The OAuth provider (e.g., 'google', 'facebook', 'github')
   * @param code - The authorization code from the OAuth provider
   */
  handleCallback: async (provider: OAuthProvider, code: string): Promise<{ user: any }> => {
    const response = await apiClient.get<ApiResponse<{ user: any }>>(
      `/auth/${provider}/callback?code=${code}`
    );
    return response.data.data!;
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout');
  },

  logoutAll: async (): Promise<void> => {
    await apiClient.post('/api/auth/logout-all');
  },
};
