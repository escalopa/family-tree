import { apiClient } from './client';
import { TreeNode, TreeQuery, RelationQuery } from '../types';

export const treeApi = {
  getTree: async (query: TreeQuery): Promise<TreeNode> => {
    const response = await apiClient.get('/api/tree', { params: query });
    return response.data;
  },

  getRelation: async (query: RelationQuery): Promise<TreeNode> => {
    const response = await apiClient.get('/api/tree/relation', { params: query });
    return response.data;
  },
};
