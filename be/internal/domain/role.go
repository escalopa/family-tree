package domain

import "time"

type Role struct {
	RoleID int    `json:"role_id" db:"role_id"`
	Name   string `json:"name" db:"name"`
}

const (
	RoleNone       = 1
	RoleViewer     = 2
	RoleAdmin      = 3
	RoleSuperAdmin = 4
)

type UserRoleHistory struct {
	HistoryID  int       `json:"history_id" db:"history_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	OldRoleID  *int      `json:"old_role_id" db:"old_role_id"`
	NewRoleID  *int      `json:"new_role_id" db:"new_role_id"`
	ChangedBy  int       `json:"changed_by" db:"changed_by"`
	ChangedAt  time.Time `json:"changed_at" db:"changed_at"`
	ActionType string    `json:"action_type" db:"action_type"` // GRANT or REVOKE
}

func RoleActionType(oldRoleID *int, newRoleID int) string {
	if oldRoleID == nil { // New user
		return "GRANT"
	}
	if *oldRoleID > newRoleID {
		return "REVOKE"
	}
	return "GRANT"
}
