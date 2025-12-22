package usecase

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateRole(ctx context.Context, userID, roleID int) error
	UpdateActive(ctx context.Context, userID int, isActive bool) error
	List(ctx context.Context, cursor *string, limit int) ([]*domain.User, *string, error)
	GetWithScore(ctx context.Context, userID int) (*domain.UserWithScore, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, sessionID string) (*domain.Session, error)
	Revoke(ctx context.Context, sessionID string) error
	RevokeAllByUser(ctx context.Context, userID int) error
	CleanExpired(ctx context.Context) error
}

type MemberRepository interface {
	Create(ctx context.Context, member *domain.Member) error
	GetByID(ctx context.Context, memberID int) (*domain.Member, error)
	Update(ctx context.Context, member *domain.Member, expectedVersion int) error
	SoftDelete(ctx context.Context, memberID int) error
	UpdatePicture(ctx context.Context, memberID int, pictureURL string) error
	DeletePicture(ctx context.Context, memberID int) error
	Search(ctx context.Context, filter domain.MemberFilter, cursor *string, limit int) ([]*domain.Member, *string, error)
	GetAll(ctx context.Context) ([]*domain.Member, error)
	GetByIDs(ctx context.Context, memberIDs []int) ([]*domain.Member, error)
}

type SpouseRepository interface {
	Create(ctx context.Context, spouse *domain.Spouse) error
	Get(ctx context.Context, member1ID, member2ID int) (*domain.Spouse, error)
	Update(ctx context.Context, spouse *domain.Spouse) error
	Delete(ctx context.Context, member1ID, member2ID int) error
	GetSpousesByMemberID(ctx context.Context, memberID int) ([]int, error)
	GetAllSpouses(ctx context.Context) (map[int][]int, error)
}

type HistoryRepository interface {
	Create(ctx context.Context, history *domain.History) error
	GetByMemberID(ctx context.Context, memberID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
	GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.HistoryWithUser, *string, error)
}

type ScoreRepository interface {
	Create(ctx context.Context, score *domain.Score) error
	GetByUserID(ctx context.Context, userID int, cursor *string, limit int) ([]*domain.ScoreHistory, *string, error)
	GetLeaderboard(ctx context.Context, limit int) ([]*domain.UserScore, error)
	GetTotalByUserID(ctx context.Context, userID int) (int, error)
	DeleteByMemberAndField(ctx context.Context, memberID int, fieldName string, memberVersion int) error
}

type RoleRepository interface {
	GetByID(ctx context.Context, roleID int) (*domain.Role, error)
	GetAll(ctx context.Context) ([]*domain.Role, error)
}

type OAuthStateRepository interface {
	Create(ctx context.Context, state *domain.OAuthState) error
	Get(ctx context.Context, state string) (*domain.OAuthState, error)
	MarkUsed(ctx context.Context, state string) error
	CleanExpired(ctx context.Context) error
}

type TokenManager interface {
	GenerateAccessToken(userID int, email string, roleID int, sessionID string) (string, error)
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
