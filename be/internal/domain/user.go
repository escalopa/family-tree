package domain

import "time"

type User struct {
	UserID    int       `json:"user_id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Avatar    *string   `json:"avatar"`
	RoleID    int       `json:"role_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWithScore struct {
	User
	TotalScore int `json:"total_score"`
}



