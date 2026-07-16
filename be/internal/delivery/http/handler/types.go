package handler

import (
	"context"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
)

type LanguageUseCase interface {
	Get(ctx context.Context, code string) (*domain.Language, error)
	List(ctx context.Context, activeOnly bool) ([]*domain.Language, error)
	ToggleActive(ctx context.Context, code string, isActive bool) error
	UpdatePreference(ctx context.Context, pref *domain.UserLanguagePreference) error
	UpdateDisplayOrder(ctx context.Context, orders map[string]int) error
}

type CookieManager interface {
	SetAuthCookies(c domain.CookieContext, accessToken, refreshToken, sessionID string)
	SetTokenCookies(c domain.CookieContext, accessToken, refreshToken string)
	ClearAuthCookies(c domain.CookieContext)
	GetAccessToken(c domain.CookieContext) (string, error)
	GetRefreshToken(c domain.CookieContext) (string, error)
	GetSessionID(c domain.CookieContext) (string, error)
}

type AuthUseCase interface {
	GetURL(ctx context.Context, provider string) (string, error)
	HandleCallback(ctx context.Context, provider, code, state string) (*domain.User, *domain.AuthTokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID int) error
	ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error)
	ListProviders(ctx context.Context) []string
}

type UserUseCase interface {
	Get(ctx context.Context, userID int) (*domain.User, error)
	GetWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error)
	List(ctx context.Context, filter domain.UserFilter, cursor *string, limit int) ([]*domain.User, *string, error)
	Update(ctx context.Context, userID int, roleID *int, isActive *bool, changedBy int) error
	ListLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error)
	ListScoreHistory(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error)
	ListChanges(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
}

type MemberUseCase interface {
	Create(ctx context.Context, member *domain.Member, userID int) error
	Update(ctx context.Context, member *domain.Member, expectedVersion, userID int) error
	Delete(ctx context.Context, memberID, userID int) error
	Get(ctx context.Context, memberID int) (*domain.Member, error)
	ListChildren(ctx context.Context, parentID int) ([]*domain.Member, error)
	ListSiblings(ctx context.Context, memberID int) ([]*domain.Member, error)
	List(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error)
	ListHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
	Rollback(ctx context.Context, memberID, historyID, userID int) error
	UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error)
	DeletePicture(ctx context.Context, memberID int, userID int) error
	GetPicture(ctx context.Context, memberID int) ([]byte, string, error)
	Compute(ctx context.Context, member *domain.Member, userRole int) *domain.MemberWithComputed
}

type FamilyTreeUseCase interface {
	Create(ctx context.Context, tree *domain.FamilyTree, userID int) error
	List(ctx context.Context, userID int) ([]*domain.FamilyTree, error)
	Get(ctx context.Context, treeID, userID int) (*domain.FamilyTree, error)
	EnsureAccess(ctx context.Context, treeID, userID int) error
	Invite(ctx context.Context, treeID, inviterUserID int, inviteeEmail string, message *string, expiresAt *time.Time) (*domain.FamilyTreeInvitation, error)
	ListTreeInvitations(ctx context.Context, treeID, userID int) ([]*domain.FamilyTreeInvitation, error)
	ListMyInvitations(ctx context.Context, userID int) ([]*domain.FamilyTreeInvitation, error)
	AcceptInvitation(ctx context.Context, invitationID, userID int) error
	DeclineInvitation(ctx context.Context, invitationID, userID int) error
	CreateShareLink(ctx context.Context, treeID, userID int, expiresAt *time.Time, maxVisits *int) (*domain.FamilyTreeShareLink, error)
	ListShareLinks(ctx context.Context, treeID, userID int) ([]*domain.FamilyTreeShareLink, error)
	RevokeShareLink(ctx context.Context, treeID, shareID, userID int) error
	ConsumeShareLink(ctx context.Context, token string) (*domain.FamilyTreeShareLink, error)
}

type SpouseUseCase interface {
	Get(ctx context.Context, spouseID int) (*domain.Spouse, error)
	Create(ctx context.Context, spouse *domain.Spouse, userID int) error
	Update(ctx context.Context, spouse *domain.Spouse, userID int) error
	Delete(ctx context.Context, spouseID, userID int) error
}

type TreeUseCase interface {
	Get(ctx context.Context, treeID int, rootID *int, userRole int) (*domain.MemberTreeNode, error)
	List(ctx context.Context, treeID int, rootID *int, userRole int) ([]*domain.MemberWithComputed, error)
	GetRelation(ctx context.Context, treeID, member1ID, member2ID int, userRole int) (*domain.MemberTreeNode, error)
	GetGraph(ctx context.Context, treeID int, userRole int) (*domain.FamilyGraph, error)
	GetRelationGraph(ctx context.Context, treeID, member1ID, member2ID int, userRole int) (*domain.FamilyGraph, error)
}
