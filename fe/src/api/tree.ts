import { apiClient } from './client';
import { TreeNode, TreeQuery, RelationQuery, Member } from '../types';

export const treeApi = {
  getTree: async (query: TreeQuery): Promise<TreeNode> => {
    const response = await apiClient.get('/api/tree', { params: query });
    return response.data.data;
  },

  getRelation: async (query: RelationQuery): Promise<TreeNode> => {
    const response = await apiClient.get('/api/tree/relation', { params: query });
    return response.data.data;
  },

  getListView: async (): Promise<Member[]> => {
    const response = await apiClient.get('/api/tree', { params: { style: 'list' } });
    return response.data.data;
  },
};
