package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/google/uuid"
)

type (
	authUseCaseRepo struct {
		user       UserRepository
		session    SessionRepository
		oauthState OAuthStateRepository
	}

	authUseCaseManager struct {
		oauth OAuthManager
		token TokenManager
	}

	authUseCase struct {
		repo authUseCaseRepo
		mgr  authUseCaseManager
	}
)

func NewAuthUseCase(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	oauthStateRepo OAuthStateRepository,
	oauthMgr OAuthManager,
	tokenMgr TokenManager,
) *authUseCase {
	return &authUseCase{
		repo: authUseCaseRepo{
			user:       userRepo,
			session:    sessionRepo,
			oauthState: oauthStateRepo,
		},
		mgr: authUseCaseManager{
			oauth: oauthMgr,
			token: tokenMgr,
		},
	}
}

func (uc *authUseCase) GetURL(ctx context.Context, provider string) (string, error) {
	state, err := uc.generateState()
	if err != nil {
		return "", domain.NewInternalError(err)
	}

	oauthState := &domain.OAuthState{
		State:     state,
		Provider:  provider,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Used:      false,
	}

	if err := uc.repo.oauthState.Create(ctx, oauthState); err != nil {
		return "", err
	}

	url, err := uc.mgr.oauth.GetAuthURL(provider, state)
	if err != nil {
		return "", err
	}

	return url, nil
}

// generateState creates a random state token for OAuth CSRF protection
func (uc *authUseCase) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (uc *authUseCase) HandleCallback(ctx context.Context, provider, code, state string) (*domain.User, *domain.AuthTokens, error) {
	if err := uc.validateState(ctx, state, provider); err != nil {
		return nil, nil, err
	}

	userInfo, err := uc.mgr.oauth.GetUserInfo(ctx, provider, code)
	if err != nil {
		return nil, nil, domain.NewExternalServiceError(err)
	}

	user, err := uc.repo.user.GetByEmail(ctx, userInfo.Email)
	if err != nil && !domain.IsDomainError(err, domain.ErrCodeNotFound) {
		return nil, nil, err
	}

	if user == nil {
		user = &domain.User{
			FullName: userInfo.Name,
			Email:    userInfo.Email,
			Avatar:   &userInfo.Picture,
			RoleID:   domain.RoleNone,
			IsActive: false, // needs admin approval
		}
	} else {
		user.FullName = userInfo.Name
		user.Avatar = &userInfo.Picture
	}

	if err := uc.repo.user.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	sessionID := uuid.New().String()
	session := &domain.Session{
		SessionID: sessionID,
		UserID:    user.UserID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		Revoked:   false,
	}
	if err := uc.repo.session.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	accessToken, err := uc.mgr.token.GenerateAccessToken(user.UserID, sessionID)
	if err != nil {
		return nil, nil, domain.NewInternalError(err)
	}

	refreshToken, err := uc.mgr.token.GenerateRefreshToken(user.UserID, sessionID)
	if err != nil {
		return nil, nil, domain.NewInternalError(err)
	}

	tokens := &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}

	return user, tokens, nil
}

func (uc *authUseCase) validateState(ctx context.Context, state, provider string) error {
	oauthState, err := uc.repo.oauthState.Get(ctx, state)
	if err != nil {
		slog.Warn("authUseCase.validateState: invalid or expired OAuth state", "error", err, "state", state)
		return domain.NewInvalidOAuthStateError()
	}

	if oauthState.Provider != provider {
		slog.Warn("authUseCase.validateState: state provider mismatch", "expected", provider, "actual", oauthState.Provider, "state", state)
		return domain.NewInvalidOAuthStateError()
	}

	if time.Now().After(oauthState.ExpiresAt) {
		slog.Warn("authUseCase.validateState: OAuth state expired", "state", state, "expires_at", oauthState.ExpiresAt)
		return domain.NewInvalidOAuthStateError()
	}

	if oauthState.Used {
		slog.Warn("authUseCase.validateState: OAuth state already used", "state", state)
		return domain.NewInvalidOAuthStateError()
	}

	if err := uc.repo.oauthState.MarkUsed(ctx, state); err != nil {
		return err
	}

	return nil
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	claims, err := uc.mgr.token.ValidateToken(refreshToken)
	if err != nil {
		slog.Warn("authUseCase.RefreshTokens: invalid refresh token", "error", err)
		return nil, domain.NewUnauthorizedError("error.invalid_token", err)
	}

	session, err := uc.repo.session.Get(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Revoked || session.ExpiresAt.Before(time.Now()) {
		slog.Warn("authUseCase.RefreshTokens: session expired or revoked", "session_id", claims.SessionID, "user_id", claims.UserID)
		return nil, domain.NewUnauthorizedError("error.session_expired", nil)
	}

	user, err := uc.repo.user.Get(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		slog.Warn("authUseCase.RefreshTokens: user is not active", "user_id", user.UserID)
		return nil, domain.NewForbiddenError("error.user.inactive")
	}

	accessToken, err := uc.mgr.token.GenerateAccessToken(user.UserID, claims.SessionID)
	if err != nil {
		return nil, domain.NewInternalError(err)
	}

	tokens := &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    claims.SessionID,
	}

	return tokens, nil
}

func (uc *authUseCase) Logout(ctx context.Context, sessionID string) error {
	return uc.repo.session.Revoke(ctx, sessionID)
}

func (uc *authUseCase) LogoutAll(ctx context.Context, userID int) error {
	return uc.repo.session.RevokeAllByUser(ctx, userID)
}

func (uc *authUseCase) ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, err := uc.repo.session.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Revoked || session.ExpiresAt.Before(time.Now()) {
		slog.Warn("authUseCase.ValidateSession: session invalid or expired", "session_id", sessionID)
		return nil, domain.NewUnauthorizedError("error.session_expired", nil)
	}
	return session, nil
}

func (uc *authUseCase) ListProviders(ctx context.Context) []string {
	return uc.mgr.oauth.GetSupportedProviders()
}
