package dto

import "time"

type ScoreHistoryResponse struct {
	UserID        int               `json:"user_id"`
	MemberID      int               `json:"member_id"`
	MemberNames   map[string]string `json:"member_names"`
	FieldName     string            `json:"field_name"`
	Points        int               `json:"points"`
	MemberVersion int               `json:"member_version"`
	CreatedAt     time.Time         `json:"created_at"`
}

type PaginatedScoreHistoryResponse struct {
	Scores     []ScoreHistoryResponse `json:"scores"`
	NextCursor *string                `json:"next_cursor,omitempty"`
}
