import { apiClient } from './client';
import { ApiResponse } from '../types/api';
import { TreeNode, TreeQueryParams, RelationQueryParams } from '../types/tree';
import { Member } from '../types/member';

export const treeApi = {
  getTree: async (params?: TreeQueryParams): Promise<TreeNode | Member[]> => {
    const response = await apiClient.get<ApiResponse<TreeNode | Member[]>>('/api/tree', { params });
    return response.data.data!;
  },

  getRelation: async (params: RelationQueryParams): Promise<Member[]> => {
    const response = await apiClient.get<ApiResponse<Member[]>>('/api/tree/relation', { params });
    return response.data.data!;
  },
};



