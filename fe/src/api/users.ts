import { apiClient } from './client';
import { ApiResponse, PaginationParams, PaginatedHistoryResponse } from '../types/api';
import { User, UserScore, PaginatedUsersResponse, PaginatedScoreHistoryResponse } from '../types/user';

export const usersApi = {
  getUser: async (userId: number): Promise<User> => {
    const response = await apiClient.get<ApiResponse<User>>(`/api/users/${userId}`);
    return response.data.data!;
  },

  listUsers: async (params?: PaginationParams): Promise<PaginatedUsersResponse> => {
    const response = await apiClient.get<ApiResponse<PaginatedUsersResponse>>('/api/users', { params });
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

  getScoreHistory: async (userId: number, params?: PaginationParams): Promise<PaginatedScoreHistoryResponse> => {
    const response = await apiClient.get<ApiResponse<PaginatedScoreHistoryResponse>>(
      `/api/users/score/${userId}`,
      { params }
    );
    return response.data.data!;
  },

  getUserChanges: async (userId: number, params?: PaginationParams): Promise<PaginatedHistoryResponse> => {
    const response = await apiClient.get<ApiResponse<PaginatedHistoryResponse>>(
      `/api/users/members/${userId}`,
      { params }
    );
    return response.data.data!;
  },
};
