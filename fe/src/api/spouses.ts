import { apiClient } from './client';
import { CreateSpouseRequest, UpdateSpouseRequest } from '../types';
import { getActiveTreeId } from './treeScope';

export const spousesApi = {
  addSpouse: async (data: CreateSpouseRequest): Promise<void> => {
    const treeId = getActiveTreeId();
    await apiClient.post(`/api/family-trees/${treeId}/spouses`, data);
  },

  updateSpouse: async (spouseId: number, data: UpdateSpouseRequest): Promise<void> => {
    const treeId = getActiveTreeId();
    await apiClient.put(`/api/family-trees/${treeId}/spouses/${spouseId}`, data);
  },

  removeSpouse: async (spouseId: number): Promise<void> => {
    const treeId = getActiveTreeId();
    await apiClient.delete(`/api/family-trees/${treeId}/spouses/${spouseId}`);
  },
};
