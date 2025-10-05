package domain

import (
	"time"
)

type UserToken struct {
	UserID int `json:"user_id" db:"user_id"`
	RoleID int `json:"role_id" db:"role_id"`
}

type User struct {
	UserID    int       `json:"user_id" db:"user_id"`
	FullName  string    `json:"full_name" db:"full_name"`
	Email     string    `json:"email" db:"email"`
	Avatar    *string   `json:"avatar" db:"avatar"`
	RoleID    int       `json:"role_id" db:"role_id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserWithRole struct {
	User
	RoleName string `json:"role_name" db:"role_name"`
}

type UserProfile struct {
	UserWithRole
	TotalScore int `json:"total_score"`
}
