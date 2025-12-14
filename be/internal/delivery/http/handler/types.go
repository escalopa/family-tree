package handler

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
)

// Use case interfaces used by HTTP handlers
type AuthUseCase interface {
	GetAuthURL(provider, state string) (string, error)
	HandleCallback(ctx context.Context, provider, code string) (*domain.User, *domain.AuthTokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID int) error
	ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error)
}

type UserUseCase interface {
	GetUserByID(ctx context.Context, userID int) (*domain.User, error)
	GetUserWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error)
	ListUsers(ctx context.Context, cursor *string, limit int) ([]*domain.User, *string, error)
	UpdateUserRole(ctx context.Context, userID, newRoleID, changedBy int) error
	UpdateUserActive(ctx context.Context, userID int, isActive bool) error
	GetLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error)
	GetScoreHistory(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error)
	GetUserChanges(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
}

type MemberUseCase interface {
	CreateMember(ctx context.Context, member *domain.Member, userID int) error
	UpdateMember(ctx context.Context, member *domain.Member, expectedVersion, userID int) error
	DeleteMember(ctx context.Context, memberID, userID int) error
	GetMemberByID(ctx context.Context, memberID int) (*domain.Member, error)
	SearchMembers(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error)
	GetMemberHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
	UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error)
	DeletePicture(ctx context.Context, memberID int) error
	ComputeMemberWithExtras(ctx context.Context, member *domain.Member, userRole int) *domain.MemberWithComputed
}

type SpouseUseCase interface {
	AddSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error
	UpdateSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error
	RemoveSpouse(ctx context.Context, member1ID, member2ID, userID int) error
}

type TreeUseCase interface {
	GetTree(ctx context.Context, rootID *int, userRole int) (*domain.MemberTreeNode, error)
	GetListView(ctx context.Context, rootID *int, userRole int) ([]*domain.MemberWithComputed, error)
	GetRelationPath(ctx context.Context, member1ID, member2ID int, userRole int) ([]*domain.MemberWithComputed, error)
}
