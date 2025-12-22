// API Response Types

export interface User {
  user_id: number;
  full_name: string;
  arabic_name?: string | null;
  english_name?: string | null;
  email: string;
  avatar: string | null;
  role_id: number;
  is_active: boolean;
  total_score?: number;
}

export interface AuthResponse {
  user: User;
}

export interface SpouseInfo {
  spouse_id: number;
  member_id: number;
  arabic_name: string;
  english_name: string;
  gender: 'M' | 'F';
  picture: string | null;
  marriage_date: string | null;
  divorce_date: string | null;
}

export interface Member {
  member_id: number;
  arabic_name: string;
  english_name: string;
  gender: 'M' | 'F';
  picture: string | null;
  date_of_birth: string | null;
  date_of_death: string | null;
  father_id: number | null;
  mother_id: number | null;
  nicknames: string[];
  profession: string | null;
  version: number;
  arabic_full_name?: string;
  english_full_name?: string;
  age?: number;
  generation_level?: number;
  is_married: boolean;
  spouses?: SpouseInfo[];
}

export interface ParentOption {
  member_id: number;
  arabic_name: string;
  english_name: string;
  picture: string | null;
  gender: 'M' | 'F';
}

export interface TreeNode {
  member: Member;
  children?: TreeNode[];
}

export interface HistoryRecord {
  history_id: number;
  member_id: number;
  user_id: number;
  user_full_name: string;
  user_email: string;
  changed_at: string;
  change_type: string;
  old_values: Record<string, any>;
  new_values: Record<string, any>;
  member_version: number;
}

export interface ScoreHistory {
  user_id: number;
  member_id: number;
  member_arabic_name: string;
  member_english_name: string;
  field_name: string;
  points: number;
  member_version: number;
  created_at: string;
}

export interface UserScore {
  user_id: number;
  full_name: string;
  arabic_name?: string | null;
  english_name?: string | null;
  avatar: string | null;
  total_score: number;
  rank: number;
}

// Request Types

export interface CreateMemberRequest {
  arabic_name: string;
  english_name: string;
  gender: 'M' | 'F';
  date_of_birth?: string;
  date_of_death?: string;
  father_id?: number;
  mother_id?: number;
  nicknames?: string[];
  profession?: string;
}

export interface UpdateMemberRequest extends CreateMemberRequest {
  version: number;
}

export interface CreateSpouseRequest {
  member1_id: number;
  member2_id: number;
  marriage_date?: string;
  divorce_date?: string;
}

export interface UpdateSpouseRequest extends CreateSpouseRequest {}

export interface UpdateSpouseByMemberRequest {
  spouse_id: number;
  marriage_date?: string;
  divorce_date?: string;
}

export interface UpdateRoleRequest {
  role_id: number;
}

export interface UpdateActiveRequest {
  is_active: boolean;
}

export interface UpdateNamesRequest {
  arabic_name?: string | null;
  english_name?: string | null;
}

// Search/Query Types

export interface MemberSearchQuery {
  name?: string; // Searches both Arabic and English names
  gender?: string;
  married?: number;
  cursor?: string;
  limit?: number;
}

export interface TreeQuery {
  root?: number;
  style?: 'tree' | 'list';
}

export interface RelationQuery {
  member1: number;
  member2: number;
}

// Paginated Response Types

export interface PaginatedResponse<T> {
  next_cursor?: string;
}

export interface PaginatedMembersResponse extends PaginatedResponse<Member> {
  members: Member[];
}

export interface PaginatedUsersResponse extends PaginatedResponse<User> {
  users: User[];
}

export interface PaginatedHistoryResponse extends PaginatedResponse<HistoryRecord> {
  history: HistoryRecord[];
}

export interface PaginatedScoreHistoryResponse extends PaginatedResponse<ScoreHistory> {
  scores: ScoreHistory[];
}

export interface LeaderboardResponse {
  users: UserScore[];
}

// Role constants matching backend
export const Roles = {
  NONE: 100,
  GUEST: 200,
  ADMIN: 300,
  SUPER_ADMIN: 400,
} as const;

export type RoleId = typeof Roles[keyof typeof Roles];
