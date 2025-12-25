package usecase

import (
	"context"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, userID int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateRole(ctx context.Context, userID, roleID int) error
	UpdateActive(ctx context.Context, userID int, isActive bool) error
	List(ctx context.Context, filter domain.UserFilter, cursor *string, limit int) ([]*domain.User, *string, error)
	GetWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error)
	CreateRoleHistory(ctx context.Context, userID, oldRoleID, newRoleID, changedBy int, actionType string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	Get(ctx context.Context, sessionID string) (*domain.Session, error)
	Revoke(ctx context.Context, sessionID string) error
	RevokeAllByUser(ctx context.Context, userID int) error
	CleanExpired(ctx context.Context) error
}

type MemberRepository interface {
	Create(ctx context.Context, member *domain.Member) error
	Get(ctx context.Context, memberID int) (*domain.Member, error)
	Update(ctx context.Context, member *domain.Member, expectedVersion int) error
	Delete(ctx context.Context, memberID int) (*string, error)
	UpdatePicture(ctx context.Context, memberID int, pictureURL string) error
	DeletePicture(ctx context.Context, memberID int) error
	List(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error)
	GetAll(ctx context.Context) ([]*domain.Member, error)
	GetChildrenByParentID(ctx context.Context, parentID int) ([]*domain.Member, error)
	GetChildrenByParents(ctx context.Context, fatherID, motherID int) ([]*domain.Member, error)
	GetSiblingsByMemberID(ctx context.Context, memberID int) ([]*domain.Member, error)
	HasChildrenWithParents(ctx context.Context, fatherID, motherID int) (bool, error)
}

type SpouseRepository interface {
	Create(ctx context.Context, spouse *domain.Spouse) error
	Get(ctx context.Context, spouseID int) (*domain.Spouse, error)
	GetByParents(ctx context.Context, fatherID, motherID int) (*domain.Spouse, error)
	Update(ctx context.Context, spouse *domain.Spouse) error
	Delete(ctx context.Context, spouseID int) error
	GetAllSpouses(ctx context.Context) (map[int][]int, error)
	GetByMemberID(ctx context.Context, memberID int) ([]domain.SpouseWithMemberInfo, error)
}

type HistoryRepository interface {
	Create(ctx context.Context, history *domain.History) error
	CreateBatch(ctx context.Context, histories ...*domain.History) error
	GetByMemberID(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
	GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
}

type ScoreRepository interface {
	Create(ctx context.Context, scores ...domain.Score) error
	GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error)
	GetLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error)
	GetTotalByUserID(ctx context.Context, userID int) (int, error)
	DeleteByMemberAndField(ctx context.Context, memberID int, fieldName string, memberVersion int) error
}

type RoleRepository interface {
	Get(ctx context.Context, roleID int) (*domain.Role, error)
	GetAll(ctx context.Context) ([]*domain.Role, error)
}

type OAuthStateRepository interface {
	Create(ctx context.Context, state *domain.OAuthState) error
	Get(ctx context.Context, state string) (*domain.OAuthState, error)
	MarkUsed(ctx context.Context, state string) error
	CleanExpired(ctx context.Context) error
}

type TokenManager interface {
	GenerateAccessToken(userID int, sessionID string) (string, error)
	GenerateRefreshToken(userID int, sessionID string) (string, error)
	ValidateToken(tokenString string) (*domain.TokenClaims, error)
}

type S3Client interface {
	UploadImage(ctx context.Context, data []byte, filename string) (string, error)
	DeleteImage(ctx context.Context, key string) error
	GetImage(ctx context.Context, key string) ([]byte, error)
}

type OAuthManager interface {
	GetAuthURL(providerName string, state string) (string, error)
	GetUserInfo(ctx context.Context, providerName, code string) (*domain.OAuthUserInfo, error)
	GetSupportedProviders() []string
}

type LanguageRepository interface {
	GetByCode(ctx context.Context, code string) (*domain.Language, error)
	GetAll(ctx context.Context, filter domain.LanguageFilter) ([]*domain.Language, error)
	ToggleActive(ctx context.Context, code string, isActive bool) error
	UpdateDisplayOrder(ctx context.Context, orders map[string]int) error
}

type UserLanguagePreferenceRepository interface {
	Upsert(ctx context.Context, pref *domain.UserLanguagePreference) error
}

type MarriageValidator interface {
	Create(ctx context.Context, memberAID, memberBID int) error
}

type BirthDateValidator interface {
	Update(ctx context.Context, memberID int, newBirthDate *time.Time) error
	Create(ctx context.Context, childBirth *time.Time, fatherID, motherID *int) error
}

type RelationshipValidator interface {
	CheckParents(ctx context.Context, memberID int, fatherID, motherID *int) error
}
