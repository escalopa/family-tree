package dto

type UpdateRoleRequest struct {
	RoleID int `json:"role_id" binding:"required,min=100,max=400"`
}

type UpdateActiveRequest struct {
	IsActive bool `json:"is_active"`
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
	Search   *string `form:"search"`
	RoleID   *int    `form:"role_id"`
	IsActive *bool   `form:"is_active"`
	Cursor   *string `form:"cursor"`
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
