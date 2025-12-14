export interface Member {
  member_id: number;
  arabic_name: string;
  english_name: string;
  gender: 'M' | 'F' | 'N';
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
  spouses?: number[];
}

export interface MemberCreateInput {
  arabic_name: string;
  english_name: string;
  gender: 'M' | 'F' | 'N';
  date_of_birth?: string;
  date_of_death?: string;
  father_id?: number;
  mother_id?: number;
  nicknames?: string[];
  profession?: string;
}

export interface MemberUpdateInput extends MemberCreateInput {
  version: number;
}

export interface MemberSearchParams {
  arabic_name?: string;
  english_name?: string;
  gender?: 'M' | 'F' | 'N';
  married?: 0 | 1;
  cursor?: string;
  limit?: number;
}

export interface PaginatedMembersResponse {
  members: Member[];
  next_cursor?: string;
}

