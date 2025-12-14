import { apiClient } from './client';
import { ApiResponse, PaginationParams, History } from '../types/api';
import { User, UserScore, ScoreHistory } from '../types/user';

export const usersApi = {
  getUser: async (userId: number): Promise<User> => {
    const response = await apiClient.get<ApiResponse<User>>(`/api/users/${userId}`);
    return response.data.data!;
  },

  listUsers: async (): Promise<User[]> => {
    const response = await apiClient.get<ApiResponse<User[]>>('/api/users');
    return response.data.data!;
  },

  updateRole: async (userId: number, roleId: number): Promise<void> => {
    await apiClient.put(`/api/users/${userId}/role`, { role_id: roleId });
  },

  updateActive: async (userId: number, isActive: boolean): Promise<void> => {
    await apiClient.put(`/api/users/${userId}/active`, { is_active: isActive });
  },

  getLeaderboard: async (limit?: number): Promise<{ users: UserScore[] }> => {
    const params = limit ? { limit } : {};
    const response = await apiClient.get<ApiResponse<{ users: UserScore[] }>>('/api/users/leaderboard', { params });
    return response.data.data!;
  },

  getScoreHistory: async (userId: number, params?: PaginationParams): Promise<ScoreHistory[]> => {
    const response = await apiClient.get<ApiResponse<ScoreHistory[]>>(
      `/api/users/score/${userId}`,
      { params }
    );
    return response.data.data!;
  },

  getUserChanges: async (userId: number, params?: PaginationParams): Promise<History[]> => {
    const response = await apiClient.get<ApiResponse<History[]>>(
      `/api/users/members/${userId}`,
      { params }
    );
    return response.data.data!;
  },
};



