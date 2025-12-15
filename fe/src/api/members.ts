import { apiClient } from './client';
import {
  Member,
  PaginatedMembersResponse,
  PaginatedHistoryResponse,
  CreateMemberRequest,
  UpdateMemberRequest,
  MemberSearchQuery,
} from '../types';

export const membersApi = {
  getMember: async (memberId: number): Promise<Member> => {
    const response = await apiClient.get(`/api/members/info/${memberId}`);
    return response.data;
  },

  searchMembers: async (query: MemberSearchQuery): Promise<PaginatedMembersResponse> => {
    const response = await apiClient.get('/api/members/search', { params: query });
    return response.data;
  },

  getMemberHistory: async (memberId: number, cursor?: string): Promise<PaginatedHistoryResponse> => {
    const response = await apiClient.get('/api/members/history', {
      params: { member_id: memberId, cursor },
    });
    return response.data;
  },

  createMember: async (data: CreateMemberRequest): Promise<Member> => {
    const response = await apiClient.post('/api/members', data);
    return response.data;
  },

  updateMember: async (memberId: number, data: UpdateMemberRequest): Promise<Member> => {
    const response = await apiClient.put(`/api/members/${memberId}`, data);
    return response.data;
  },

  deleteMember: async (memberId: number): Promise<void> => {
    await apiClient.delete(`/api/members/${memberId}`);
  },

  uploadPicture: async (memberId: number, file: File): Promise<void> => {
    const formData = new FormData();
    formData.append('picture', file);
    await apiClient.post(`/api/members/${memberId}/picture`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },

  deletePicture: async (memberId: number): Promise<void> => {
    await apiClient.delete(`/api/members/${memberId}/picture`);
  },
};
