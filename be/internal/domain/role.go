package domain

const (
	RoleNone       = 100
	RoleGuest      = 200
	RoleAdmin      = 300
	RoleSuperAdmin = 400
)

type Role struct {
	RoleID int    `json:"role_id"`
	Name   string `json:"name"`
}
