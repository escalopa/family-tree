import { apiClient } from './client';
import {
  Member,
  MemberListItem,
  PaginatedHistoryResponse,
  CreateMemberRequest,
  UpdateMemberRequest,
} from '../types';
import { getActiveTreeId } from './treeScope';

export const membersApi = {
  getMember: async (memberId: number): Promise<Member> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/members/${memberId}`);
    return response.data.data;
  },

  searchMembers: async (params: {
    name?: string;
    arabic_name?: string;
    english_name?: string;
    gender?: 'M' | 'F';
    married?: boolean;
    cursor?: string;
    limit?: number;
  }): Promise<{ members: MemberListItem[]; next_cursor?: string }> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/members`, { params });
    return response.data.data;
  },

  filterMembers: async (params: {
    name?: string;
    arabic_name?: string;
    english_name?: string;
    gender?: 'M' | 'F';
    married?: boolean;
    limit?: number;
  }): Promise<{ members: MemberListItem[]; next_cursor?: string }> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/members/search`, { params });
    return response.data.data;
  },

  getMemberHistory: async (memberId: number, cursor?: string): Promise<PaginatedHistoryResponse> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/members/history`, {
      params: { member_id: memberId, cursor },
    });
    return response.data.data;
  },

  createMember: async (data: CreateMemberRequest): Promise<Member> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.post(`/api/family-trees/${treeId}/members`, data);
    return response.data.data;
  },

  updateMember: async (memberId: number, data: UpdateMemberRequest): Promise<Member> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.put(`/api/family-trees/${treeId}/members/${memberId}`, data);
    return response.data.data;
  },

  rollbackMember: async (memberId: number, historyId: number): Promise<void> => {
    const treeId = getActiveTreeId();
    await apiClient.post(`/api/family-trees/${treeId}/members/${memberId}/rollback`, { history_id: historyId });
  },

  deleteMember: async (memberId: number): Promise<void> => {
    const treeId = getActiveTreeId();
    await apiClient.delete(`/api/family-trees/${treeId}/members/${memberId}`);
  },

  uploadPicture: async (memberId: number, file: File): Promise<string> => {
    const formData = new FormData();
    formData.append('picture', file);
    const treeId = getActiveTreeId();
    const response = await apiClient.post(`/api/family-trees/${treeId}/members/${memberId}/picture`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data.data.picture_url;
  },

  deletePicture: async (memberId: number): Promise<void> => {
    const treeId = getActiveTreeId();
    await apiClient.delete(`/api/family-trees/${treeId}/members/${memberId}/picture`);
  },
};
