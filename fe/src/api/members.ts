import { apiClient } from './client';
import {
  Member,
  ParentOption,
  PaginatedMembersResponse,
  PaginatedHistoryResponse,
  CreateMemberRequest,
  UpdateMemberRequest,
  MemberSearchQuery,
} from '../types';

export const membersApi = {
  getMember: async (memberId: number): Promise<Member> => {
    const response = await apiClient.get(`/api/members/info/${memberId}`);
    return response.data.data;
  },

  getChildren: async (parentId: number): Promise<Member[]> => {
    const response = await apiClient.get(`/api/members/${parentId}/children`);
    return response.data.data;
  },

  searchMembers: async (query: MemberSearchQuery): Promise<PaginatedMembersResponse> => {
    const response = await apiClient.get('/api/members/search', { params: query });
    return response.data.data;
  },

  getMemberHistory: async (memberId: number, cursor?: string): Promise<PaginatedHistoryResponse> => {
    const response = await apiClient.get('/api/members/history', {
      params: { member_id: memberId, cursor },
    });
    return response.data.data;
  },

  createMember: async (data: CreateMemberRequest): Promise<Member> => {
    const response = await apiClient.post('/api/members', data);
    return response.data.data;
  },

  updateMember: async (memberId: number, data: UpdateMemberRequest): Promise<Member> => {
    const response = await apiClient.put(`/api/members/${memberId}`, data);
    return response.data.data;
  },

  deleteMember: async (memberId: number): Promise<void> => {
    await apiClient.delete(`/api/members/${memberId}`);
  },

  uploadPicture: async (memberId: number, file: File): Promise<string> => {
    const formData = new FormData();
    formData.append('picture', file);
    const response = await apiClient.post(`/api/members/${memberId}/picture`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data.data.picture_url;
  },

  deletePicture: async (memberId: number): Promise<void> => {
    await apiClient.delete(`/api/members/${memberId}/picture`);
  },

  searchParents: async (query: string, gender: 'M' | 'F'): Promise<ParentOption[]> => {
    const response = await apiClient.get('/api/members/search-parents', {
      params: { q: query, gender, limit: 20 },
    });
    return response.data.data;
  },
};
