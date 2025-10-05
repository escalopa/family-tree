package dto

import (
	"encoding/json"
	"time"
)

type MemberHistoryResponse struct {
	HistoryID  int             `json:"history_id"`
	MemberID   int             `json:"member_id"`
	UserID     int             `json:"user_id"`
	UserName   string          `json:"user_name"`
	Version    int             `json:"version"`
	Revision   int             `json:"revision"`
	ChangedAt  time.Time       `json:"changed_at"`
	ChangeType string          `json:"change_type"`
	OldValues  json.RawMessage `json:"old_values"`
	NewValues  json.RawMessage `json:"new_values"`
}

type MemberHistoryListResponse struct {
	History    []MemberHistoryResponse `json:"history"`
	Pagination PaginationResponse      `json:"pagination"`
}

type ActivityResponse struct {
	HistoryID  int       `json:"history_id"`
	MemberID   int       `json:"member_id"`
	MemberName string    `json:"member_name"`
	ChangeType string    `json:"change_type"`
	ChangedAt  time.Time `json:"changed_at"`
	Version    int       `json:"version"`
}
