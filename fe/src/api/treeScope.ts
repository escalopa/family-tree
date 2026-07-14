const ACTIVE_TREE_ID_KEY = 'active_tree_id';

export const setActiveTreeId = (treeId: number) => {
  localStorage.setItem(ACTIVE_TREE_ID_KEY, String(treeId));
};

export const getActiveTreeId = (): number => {
  const value = localStorage.getItem(ACTIVE_TREE_ID_KEY);
  const treeId = value ? Number(value) : 0;
  if (!Number.isFinite(treeId) || treeId <= 0) {
    throw new Error('No active family tree selected');
  }
  return treeId;
};
