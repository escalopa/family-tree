package dto

import "time"

type UserResponse struct {
	UserID    int       `json:"user_id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Avatar    *string   `json:"avatar"`
	RoleID    int       `json:"role_id"`
	RoleName  string    `json:"role_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UserProfileResponse struct {
	UserResponse
	TotalScore       int                `json:"total_score"`
	RecentActivities []ActivityResponse `json:"recent_activities,omitempty"`
}

type UserDetailResponse struct {
	UserResponse
	TotalScore int `json:"total_score"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
}

type UpdateRoleRequest struct {
	RoleID int `json:"role_id" binding:"required,min=1,max=4"`
}

type UpdateActiveStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

type RecentActivitiesResponse struct {
	Activities []ActivityResponse `json:"activities"`
	Pagination PaginationResponse `json:"pagination"`
}
