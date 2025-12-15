import { apiClient } from './client';
import { CreateSpouseRequest, UpdateSpouseRequest } from '../types';

export const spousesApi = {
  addSpouse: async (data: CreateSpouseRequest): Promise<void> => {
    await apiClient.post('/api/spouses', data);
  },

  updateSpouse: async (data: UpdateSpouseRequest): Promise<void> => {
    await apiClient.put('/api/spouses', data);
  },

  removeSpouse: async (member1Id: number, member2Id: number): Promise<void> => {
    await apiClient.delete('/api/spouses', {
      params: { member1_id: member1Id, member2_id: member2Id },
    });
  },
};
