import { apiClient } from './client';
import { ApiResponse, PaginationParams, History } from '../types/api';
import { Member, MemberCreateInput, MemberUpdateInput, MemberSearchParams, PaginatedMembersResponse } from '../types/member';

export const membersApi = {
  createMember: async (data: MemberCreateInput): Promise<{ member_id: number; version: number }> => {
    const response = await apiClient.post<ApiResponse<{ member_id: number; version: number }>>(
      '/api/members',
      data
    );
    return response.data.data!;
  },

  updateMember: async (memberId: number, data: MemberUpdateInput): Promise<{ version: number }> => {
    const response = await apiClient.put<ApiResponse<{ version: number }>>(
      `/api/members/${memberId}`,
      data
    );
    return response.data.data!;
  },

  deleteMember: async (memberId: number): Promise<void> => {
    await apiClient.delete(`/api/members/${memberId}`);
  },

  getMember: async (memberId: number): Promise<Member> => {
    const response = await apiClient.get<ApiResponse<Member>>(`/api/members/info/${memberId}`);
    return response.data.data!;
  },

  searchMembers: async (params: MemberSearchParams): Promise<PaginatedMembersResponse> => {
    const response = await apiClient.get<ApiResponse<PaginatedMembersResponse>>('/api/members/search', { params });
    return response.data.data!;
  },

  getMemberHistory: async (memberId: number, params?: PaginationParams): Promise<History[]> => {
    const response = await apiClient.get<ApiResponse<History[]>>('/api/members/history', {
      params: { member_id: memberId, ...params },
    });
    return response.data.data!;
  },

  uploadPicture: async (memberId: number, file: File): Promise<void> => {
    const formData = new FormData();
    formData.append('picture', file);
    await apiClient.post(`/api/members/${memberId}/picture`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  deletePicture: async (memberId: number): Promise<void> => {
    await apiClient.delete(`/api/members/${memberId}/picture`);
  },
};


