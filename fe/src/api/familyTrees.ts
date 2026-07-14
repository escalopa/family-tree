import { apiClient } from './client';
import {
  FamilyTree,
  FamilyTreeInvitation,
  FamilyTreeShareLink,
  CreateFamilyTreeRequest,
  InviteToTreeRequest,
  CreateShareLinkRequest,
  PublicTreeResponse,
} from '../types';

const withFrontendShareURL = (link: FamilyTreeShareLink): FamilyTreeShareLink => ({
  ...link,
  url: `${window.location.origin}/public/trees/${link.token}`,
});

export const familyTreesApi = {
  list: async (): Promise<FamilyTree[]> => {
    const response = await apiClient.get('/api/family-trees');
    return response.data.data.trees || [];
  },

  create: async (data: CreateFamilyTreeRequest): Promise<FamilyTree> => {
    const response = await apiClient.post('/api/family-trees', data);
    return response.data.data;
  },

  get: async (treeId: number): Promise<FamilyTree> => {
    const response = await apiClient.get(`/api/family-trees/${treeId}`);
    return response.data.data;
  },

  listMyInvitations: async (): Promise<FamilyTreeInvitation[]> => {
    const response = await apiClient.get('/api/family-trees/invitations');
    return response.data.data.invitations || [];
  },

  acceptInvitation: async (invitationId: number): Promise<void> => {
    await apiClient.post(`/api/family-trees/invitations/${invitationId}/accept`);
  },

  declineInvitation: async (invitationId: number): Promise<void> => {
    await apiClient.post(`/api/family-trees/invitations/${invitationId}/decline`);
  },

  invite: async (treeId: number, data: InviteToTreeRequest): Promise<FamilyTreeInvitation> => {
    const response = await apiClient.post(`/api/family-trees/${treeId}/invitations`, data);
    return response.data.data;
  },

  listInvitations: async (treeId: number): Promise<FamilyTreeInvitation[]> => {
    const response = await apiClient.get(`/api/family-trees/${treeId}/invitations`);
    return response.data.data.invitations || [];
  },

  createShareLink: async (treeId: number, data: CreateShareLinkRequest): Promise<FamilyTreeShareLink> => {
    const response = await apiClient.post(`/api/family-trees/${treeId}/share-links`, data);
    return withFrontendShareURL(response.data.data);
  },

  listShareLinks: async (treeId: number): Promise<FamilyTreeShareLink[]> => {
    const response = await apiClient.get(`/api/family-trees/${treeId}/share-links`);
    return (response.data.data.share_links || []).map(withFrontendShareURL);
  },

  revokeShareLink: async (treeId: number, shareId: number): Promise<void> => {
    await apiClient.delete(`/api/family-trees/${treeId}/share-links/${shareId}`);
  },

  getPublicTree: async (token: string): Promise<PublicTreeResponse> => {
    const response = await apiClient.get(`/public/trees/${token}`);
    return response.data.data;
  },
};
