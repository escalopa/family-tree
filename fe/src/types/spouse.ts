export interface Spouse {
  member1_id: number;
  member2_id: number;
  marriage_date: string | null;
  divorce_date: string | null;
}

export interface SpouseCreateInput {
  member1_id: number;
  member2_id: number;
  marriage_date?: string;
  divorce_date?: string;
}



