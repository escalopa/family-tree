package dto

import (
	"encoding/json"
	"time"
)

type HistoryQuery struct {
	MemberID *int `form:"member_id"`
	PaginationQuery
}

type HistoryResponse struct {
	HistoryID     int             `json:"history_id"`
	MemberID      int             `json:"member_id"`
	MemberName    string          `json:"member_name,omitempty"`
	UserID        int             `json:"user_id"`
	UserFullName  string          `json:"user_full_name"`
	UserEmail     string          `json:"user_email"`
	ChangedAt     time.Time       `json:"changed_at"`
	ChangeType    string          `json:"change_type"`
	OldValues     json.RawMessage `json:"old_values"`
	NewValues     json.RawMessage `json:"new_values"`
	MemberVersion int             `json:"member_version"`
}

type PaginatedHistoryResponse struct {
	History    []HistoryResponse `json:"history"`
	NextCursor *string           `json:"next_cursor,omitempty"`
}
