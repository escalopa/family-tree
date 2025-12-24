import { apiClient } from './client';
import { CreateSpouseRequest, UpdateSpouseRequest } from '../types';

export const spousesApi = {
  addSpouse: async (data: CreateSpouseRequest): Promise<void> => {
    await apiClient.post('/api/spouses', data);
  },

  updateSpouse: async (data: UpdateSpouseRequest): Promise<void> => {
    await apiClient.put('/api/spouses', data);
  },

  removeSpouse: async (spouseId: number): Promise<void> => {
    await apiClient.delete('/api/spouses', {
      data: { spouse_id: spouseId },
    });
  },
};
