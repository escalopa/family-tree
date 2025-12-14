package dto

type GoogleAuthURLResponse struct {
	URL string `json:"url"`
}

type AuthResponse struct {
	User struct {
		UserID   int     `json:"user_id"`
		FullName string  `json:"full_name"`
		Email    string  `json:"email"`
		Avatar   *string `json:"avatar"`
		RoleID   int     `json:"role_id"`
		IsActive bool    `json:"is_active"`
	} `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}


