package domain

import (
	"encoding/json"
	"time"
)

type MemberHistory struct {
	HistoryID  int             `json:"history_id" db:"history_id"`
	MemberID   int             `json:"member_id" db:"member_id"`
	UserID     int             `json:"user_id" db:"user_id"`
	Version    int             `json:"version" db:"version"`
	Revision   int             `json:"revision" db:"revision"`
	ChangedAt  time.Time       `json:"changed_at" db:"changed_at"`
	ChangeType string          `json:"change_type" db:"change_type"`
	OldValues  json.RawMessage `json:"old_values" db:"old_values"`
	NewValues  json.RawMessage `json:"new_values" db:"new_values"`
}

type MemberHistoryWithUser struct {
	MemberHistory
	UserName string `json:"user_name"`
}

const (
	ChangeTypeInsert       = "INSERT"
	ChangeTypeUpdate       = "UPDATE"
	ChangeTypeDelete       = "DELETE"
	ChangeTypeAddSpouse    = "ADD_SPOUSE"
	ChangeTypeRemoveSpouse = "REMOVE_SPOUSE"
	ChangeTypeRollback     = "ROLLBACK"
)

type Activity struct {
	HistoryID  int       `json:"history_id"`
	MemberID   int       `json:"member_id"`
	MemberName string    `json:"member_name"`
	ChangeType string    `json:"change_type"`
	ChangedAt  time.Time `json:"changed_at"`
	Version    int       `json:"version"`
}
