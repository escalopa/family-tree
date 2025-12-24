import { apiClient } from './client';
import {
  Member,
  MemberListItem,
  PaginatedHistoryResponse,
  CreateMemberRequest,
  UpdateMemberRequest,
} from '../types';

export const membersApi = {
  getMember: async (memberId: number): Promise<Member> => {
    const response = await apiClient.get(`/api/members/${memberId}`);
    return response.data.data;
  },

  searchMembers: async (params: {
    name?: string;
    gender?: 'M' | 'F';
    married?: 0 | 1;
    cursor?: string;
    limit?: number;
  }): Promise<{ members: MemberListItem[]; next_cursor?: string }> => {
    const response = await apiClient.get('/api/members', { params });
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
    await apiClient.delete(`/api/members/${memberId}`);
  },
};
