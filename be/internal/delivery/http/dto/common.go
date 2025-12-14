package dto

type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type PaginationQuery struct {
	Cursor *string `form:"cursor"`
	Limit  int     `form:"limit" binding:"omitempty,min=1,max=100"`
}
