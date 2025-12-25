package dto

type AuthURLResponse struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
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

type ProvidersResponse struct {
	Providers []string `json:"providers"`
}

type CallbackQuery struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}
