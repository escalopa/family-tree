// API Response Types

export interface User {
  user_id: number;
  full_name: string;
  email: string;
  avatar: string | null;
  role_id: number;
  is_active: boolean;
  preferred_language: string;
  total_score?: number;
}

export interface AuthResponse {
  user: User;
}

export interface SpouseInfo {
  spouse_id: number;
  member_id: number;
  name: string; // Name in user's preferred language
  gender: 'M' | 'F';
  picture: string | null;
  marriage_date: string | null;
  divorce_date: string | null;
  married_years: number | null;
}

export interface Language {
  language_code: string;
  language_name: string;
  is_active: boolean;
  display_order: number;
}

export interface UserLanguagePreference {
  preferred_language: string;
}

export interface MemberInfo {
  member_id: number;
  name: string; // Name in user's preferred language
  picture: string | null;
  gender: 'M' | 'F';
}

// Minimal member data for list views
export interface MemberListItem {
  member_id: number;
  name: string; // Name in user's preferred language
  gender: 'M' | 'F';
  picture: string | null;
  date_of_birth: string | null;
  date_of_death: string | null;
  is_married: boolean;
}

export interface Member {
  member_id: number;
  name: string; // Name in user's preferred language
  names: Record<string, string>; // All language translations (for editing)
  full_name?: string; // Full name in user's preferred language
  full_names?: Record<string, string>; // All full name translations (for editing)
  gender: 'M' | 'F';
  picture: string | null;
  date_of_birth: string | null;
  date_of_death: string | null;
  father_id: number | null;
  mother_id: number | null;
  father?: MemberInfo;
  mother?: MemberInfo;
  nicknames: string[];
  profession: string | null;
  version: number;
  age?: number;
  generation_level?: number;
  is_married: boolean;
  spouses?: SpouseInfo[];
  children?: MemberInfo[];
  siblings?: MemberInfo[];
}

export interface TreeNode {
  member: Member;
  children?: TreeNode[];
  is_in_path?: boolean; // For relation path highlighting
}

export interface HistoryRecord {
  history_id: number;
  member_id: number;
  member_name?: string; // Member name in user's preferred language (may be missing for deleted members)
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
  member_name: string; // Member name in user's preferred language
  field_name: string;
  points: number;
  member_version: number;
  created_at: string;
}

export interface UserScore {
  user_id: number;
  full_name: string;
  avatar: string | null;
  total_score: number;
  rank: number;
}

// Request Types

export interface CreateMemberRequest {
  names: Record<string, string>; // language_code -> name
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
  father_id: number;
  mother_id: number;
  marriage_date?: string;
  divorce_date?: string;
}

export interface UpdateSpouseRequest {
  spouse_id: number;
  marriage_date?: string;
  divorce_date?: string;
}

export interface UpdateUserRequest {
  role_id?: number;
  is_active?: boolean;
}

// Search/Query Types

export interface MemberSearchQuery {
  name?: string; // Searches both Arabic and English names
  gender?: 'M' | 'F';
  married?: boolean;
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

export interface PaginatedResponse {
  next_cursor?: string;
}

export interface PaginatedMembersResponse extends PaginatedResponse {
  members: MemberListItem[];
}

export interface PaginatedUsersResponse extends PaginatedResponse {
  users: User[];
}

export interface PaginatedHistoryResponse extends PaginatedResponse {
  history: HistoryRecord[];
}

export interface PaginatedScoreHistoryResponse extends PaginatedResponse {
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
