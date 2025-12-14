export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface History {
  history_id: number;
  member_id: number;
  user_id: number;
  user_full_name: string;
  user_email: string;
  changed_at: string;
  change_type: string;
  old_values: any;
  new_values: any;
  member_version: number;
}


