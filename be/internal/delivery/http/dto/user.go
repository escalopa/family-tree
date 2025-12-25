package dto

type UpdateUserRequest struct {
	RoleID   *int  `json:"role_id,omitempty" binding:"omitempty,oneof=100 200 300 400"`
	IsActive *bool `json:"is_active,omitempty" binding:"omitempty"`
}

type UserResponse struct {
	UserID     int     `json:"user_id"`
	FullName   string  `json:"full_name"`
	Email      string  `json:"email"`
	Avatar     *string `json:"avatar"`
	RoleID     int     `json:"role_id"`
	IsActive   bool    `json:"is_active"`
	TotalScore *int    `json:"total_score,omitempty"`
}

type UserFilterQuery struct {
	Search   *string `form:"search" binding:"omitempty,max=100"`
	RoleID   *int    `form:"role_id" binding:"omitempty,oneof=100 200 300 400"`
	IsActive *bool   `form:"is_active" binding:"omitempty"`
	Cursor   *string `form:"cursor" binding:"omitempty"`
	Limit    int     `form:"limit,default=20" binding:"omitempty,min=1,max=100"`
}

type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

type LeaderboardResponse struct {
	Users []UserScore `json:"users"`
}

type UserScore struct {
	UserID     int     `json:"user_id"`
	FullName   string  `json:"full_name"`
	Avatar     *string `json:"avatar"`
	TotalScore int     `json:"total_score"`
	Rank       int     `json:"rank"`
}
