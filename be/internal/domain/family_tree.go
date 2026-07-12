package domain

import "time"

const (
	TreeRoleOwner  = "owner"
	TreeRoleEditor = "editor"
	TreeRoleViewer = "viewer"

	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusDeclined = "declined"
	InvitationStatusRevoked  = "revoked"
)

type FamilyTree struct {
	TreeID      int       `json:"tree_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerUserID int       `json:"owner_user_id"`
	OwnerName   string    `json:"owner_name,omitempty"`
	OwnerEmail  string    `json:"owner_email,omitempty"`
	UserRole    string    `json:"user_role,omitempty"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FamilyTreeInvitation struct {
	InvitationID  int        `json:"invitation_id"`
	TreeID        int        `json:"tree_id"`
	TreeName      string     `json:"tree_name,omitempty"`
	InviterUserID int        `json:"inviter_user_id"`
	InviterName   string     `json:"inviter_name,omitempty"`
	InviteeUserID *int       `json:"invitee_user_id"`
	InviteeEmail  string     `json:"invitee_email"`
	Status        string     `json:"status"`
	Message       *string    `json:"message"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	RespondedAt   *time.Time `json:"responded_at"`
}

type FamilyTreeShareLink struct {
	ShareID    int        `json:"share_id"`
	TreeID     int        `json:"tree_id"`
	Token      string     `json:"token"`
	CreatedBy  int        `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	MaxVisits  *int       `json:"max_visits"`
	VisitCount int        `json:"visit_count"`
	RevokedAt  *time.Time `json:"revoked_at"`
}
