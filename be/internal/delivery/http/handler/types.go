package handler

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
)

type LanguageUseCase interface {
	CreateLanguage(ctx context.Context, language *domain.Language) error
	GetLanguage(ctx context.Context, code string) (*domain.Language, error)
	GetAllLanguages(ctx context.Context, activeOnly bool) ([]*domain.Language, error)
	UpdateLanguage(ctx context.Context, language *domain.Language) error
	UpdateUserLanguagePreference(ctx context.Context, pref *domain.UserLanguagePreference) error
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
	GetAuthURL(ctx context.Context, provider string) (string, error)
	HandleCallback(ctx context.Context, provider, code, state string) (*domain.User, *domain.AuthTokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID int) error
	ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error)
	GetSupportedProviders(ctx context.Context) []string
}

type UserUseCase interface {
	GetUserByID(ctx context.Context, userID int) (*domain.User, error)
	GetUserWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error)
	ListUsers(ctx context.Context, filter domain.UserFilter, cursor *string, limit int) ([]*domain.User, *string, error)
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
	GetChildrenByParentID(ctx context.Context, parentID int) ([]*domain.Member, error)
	GetSiblingsByMemberID(ctx context.Context, memberID int) ([]*domain.Member, error)
	SearchMembers(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error)
	GetMemberHistory(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
	UploadPicture(ctx context.Context, memberID int, data []byte, filename string, userID int) (string, error)
	DeletePicture(ctx context.Context, memberID int, userID int) error
	GetPicture(ctx context.Context, memberID int) ([]byte, string, error)
	ComputeMemberWithExtras(ctx context.Context, member *domain.Member, userRole int) *domain.MemberWithComputed
}

type SpouseUseCase interface {
	AddSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error
	UpdateSpouseByID(ctx context.Context, spouse *domain.Spouse, userID int) error
	UpdateSpouse(ctx context.Context, spouse *domain.Spouse, userID int) error
	RemoveSpouse(ctx context.Context, fatherID, motherID, userID int) error
	RemoveSpouseByID(ctx context.Context, spouseID, userID int) error
}

type TreeUseCase interface {
	GetTree(ctx context.Context, rootID *int, userRole int) (*domain.MemberTreeNode, error)
	GetListView(ctx context.Context, rootID *int, userRole int) ([]*domain.MemberWithComputed, error)
	GetRelationTree(ctx context.Context, member1ID, member2ID int, userRole int) (*domain.MemberTreeNode, error)
}
