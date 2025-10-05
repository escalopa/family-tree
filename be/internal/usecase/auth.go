package usecase

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/escalopa/family-tree-api/internal/pkg/oauth"
	"github.com/escalopa/family-tree-api/internal/pkg/token"
	"github.com/escalopa/family-tree-api/internal/repository"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	GetGoogleAuthURL(ctx context.Context, state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*domain.UserWithRole, *domain.UserSession, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID int) error
}

type authUseCase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	oauthClient *oauth.GoogleOAuthClient
	tokenMgr    *token.Manager
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	oauthClient *oauth.GoogleOAuthClient,
	tokenMgr *token.Manager,
) AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		oauthClient: oauthClient,
		tokenMgr:    tokenMgr,
	}
}

func (uc *authUseCase) GetGoogleAuthURL(ctx context.Context, state string) string {
	// TODO: Implementation
	return ""
}

func (uc *authUseCase) HandleGoogleCallback(ctx context.Context, code string) (*domain.UserWithRole, *domain.UserSession, error) {
	// TODO: Implementation
	return nil, nil, nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// TODO: Implementation
	return "", nil
}

func (uc *authUseCase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	// TODO: Implementation
	return nil
}

func (uc *authUseCase) LogoutAll(ctx context.Context, userID int) error {
	// TODO: Implementation
	return nil
}
