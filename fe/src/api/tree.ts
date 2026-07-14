import { apiClient } from './client';
import { TreeNode, TreeQuery, RelationQuery, Member } from '../types';
import { getActiveTreeId } from './treeScope';

export const treeApi = {
  getTree: async (query: TreeQuery): Promise<TreeNode> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/tree`, { params: query });
    return response.data.data;
  },

  getRelation: async (query: RelationQuery): Promise<TreeNode> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/tree/relation`, { params: query });
    return response.data.data;
  },

  getListView: async (): Promise<Member[]> => {
    const treeId = getActiveTreeId();
    const response = await apiClient.get(`/api/family-trees/${treeId}/tree`, { params: { style: 'list' } });
    return response.data.data;
  },
};
