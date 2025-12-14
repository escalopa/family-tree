import { Member } from './member';

export interface TreeNode {
  member: Member;
  children?: TreeNode[];
}

export interface TreeQueryParams {
  root?: number;
  style?: 'tree' | 'list';
}

export interface RelationQueryParams {
  member1: number;
  member2: number;
}


