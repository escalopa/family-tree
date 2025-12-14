import { apiClient } from './client';
import { SpouseCreateInput } from '../types/spouse';

export const spousesApi = {
  addSpouse: async (data: SpouseCreateInput): Promise<void> => {
    await apiClient.post('/api/spouses', data);
  },

  updateSpouse: async (data: SpouseCreateInput): Promise<void> => {
    await apiClient.put('/api/spouses', data);
  },

  removeSpouse: async (member1Id: number, member2Id: number): Promise<void> => {
    await apiClient.delete('/api/spouses', {
      data: { member1_id: member1Id, member2_id: member2Id },
    });
  },
};


