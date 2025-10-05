package dto

type AuthResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}
