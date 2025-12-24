package dto

type Response struct {
	Success   bool   `json:"success"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type PaginationQuery struct {
	Cursor *string `form:"cursor"`
	Limit  int     `form:"limit,default=20" binding:"omitempty,min=1,max=100"`
}
