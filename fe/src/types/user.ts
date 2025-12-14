export interface User {
  user_id: number;
  full_name: string;
  email: string;
  avatar: string | null;
  role_id: number;
  is_active: boolean;
  total_score?: number;
}

export interface UserScore {
  user_id: number;
  full_name: string;
  avatar: string | null;
  total_score: number;
  rank: number;
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

export const ROLES = {
  NONE: 100,
  GUEST: 200,
  ADMIN: 300,
  SUPER_ADMIN: 400,
} as const;


