import { apiClient } from './client';
import {
  User,
  PaginatedUsersResponse,
  LeaderboardResponse,
  PaginatedScoreHistoryResponse,
  PaginatedHistoryResponse,
  UpdateUserRequest,
} from '../types';

export const usersApi = {
  listUsers: async (
    cursor?: string,
    limit: number = 20,
    search?: string,
    roleId?: number,
    isActive?: boolean
  ): Promise<PaginatedUsersResponse> => {
    const response = await apiClient.get('/api/users', {
      params: {
        cursor,
        limit,
        search: search || undefined,
        role_id: roleId !== undefined ? roleId : undefined,
        is_active: isActive !== undefined ? isActive : undefined,
      },
    });
    return response.data.data;
  },

  getUser: async (userId: number): Promise<User> => {
    const response = await apiClient.get(`/api/users/${userId}`);
    return response.data.data;
  },

  getLeaderboard: async (): Promise<LeaderboardResponse> => {
    const response = await apiClient.get('/api/users/leaderboard');
    return response.data.data;
  },

  getScoreHistory: async (userId: number, cursor?: string): Promise<PaginatedScoreHistoryResponse> => {
    const response = await apiClient.get(`/api/users/score/${userId}`, {
      params: { cursor },
    });
    return response.data.data;
  },

  getUserChanges: async (userId: number, cursor?: string): Promise<PaginatedHistoryResponse> => {
    const response = await apiClient.get(`/api/users/members/${userId}`, {
      params: { cursor },
    });
    return response.data.data;
  },

  updateUser: async (userId: number, data: UpdateUserRequest): Promise<void> => {
    await apiClient.put(`/api/users/${userId}`, data);
  },
};
