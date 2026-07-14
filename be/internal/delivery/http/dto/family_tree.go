package dto

import "time"

type TreeIDUri struct {
	TreeID int `uri:"tree_id" binding:"required,min=1"`
}

type ShareIDUri struct {
	TreeID  int `uri:"tree_id" binding:"required,min=1"`
	ShareID int `uri:"share_id" binding:"required,min=1"`
}

type InvitationIDUri struct {
	InvitationID int `uri:"invitation_id" binding:"required,min=1"`
}

type PublicShareUri struct {
	Token string `uri:"token" binding:"required"`
}

type CreateFamilyTreeRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
}

type InviteToTreeRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	Message   *string `json:"message"`
	ExpiresAt *Date   `json:"expires_at"`
}

type CreateShareLinkRequest struct {
	ExpiresAt *Date `json:"expires_at"`
	MaxVisits *int  `json:"max_visits" binding:"omitempty,min=1"`
}

type FamilyTreeResponse struct {
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

type FamilyTreeListResponse struct {
	Trees []FamilyTreeResponse `json:"trees"`
}

type FamilyTreeInvitationResponse struct {
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

type FamilyTreeInvitationListResponse struct {
	Invitations []FamilyTreeInvitationResponse `json:"invitations"`
}

type FamilyTreeShareLinkResponse struct {
	ShareID    int        `json:"share_id"`
	TreeID     int        `json:"tree_id"`
	Token      string     `json:"token"`
	URL        string     `json:"url"`
	CreatedBy  int        `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	MaxVisits  *int       `json:"max_visits"`
	VisitCount int        `json:"visit_count"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

type FamilyTreeShareLinkListResponse struct {
	ShareLinks []FamilyTreeShareLinkResponse `json:"share_links"`
}

type PublicTreeResponse struct {
	Share FamilyTreeShareLinkResponse `json:"share"`
	Tree  *TreeNodeResponse           `json:"tree"`
}
