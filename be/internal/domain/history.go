package domain

import (
	"encoding/json"
	"time"
)

const (
	ChangeTypeInsert       = "INSERT"
	ChangeTypeUpdate       = "UPDATE"
	ChangeTypeDelete       = "DELETE"
	ChangeTypeAddSpouse    = "ADD_SPOUSE"
	ChangeTypeRemoveSpouse = "REMOVE_SPOUSE"
	ChangeTypeUpdateSpouse = "UPDATE_SPOUSE"
)

type History struct {
	HistoryID     int             `json:"history_id"`
	MemberID      int             `json:"member_id"`
	UserID        int             `json:"user_id"`
	ChangedAt     time.Time       `json:"changed_at"`
	ChangeType    string          `json:"change_type"`
	OldValues     json.RawMessage `json:"old_values"`
	NewValues     json.RawMessage `json:"new_values"`
	MemberVersion int             `json:"member_version"`
}

type HistoryWithUser struct {
	History
	UserFullName string `json:"user_full_name"`
	UserEmail    string `json:"user_email"`
}


