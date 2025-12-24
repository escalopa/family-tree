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

type authUseCase struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	oauthStateRepo OAuthStateRepository
	oauthMgr       OAuthManager
	tokenMgr       TokenManager
}

func NewAuthUseCase(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	oauthStateRepo OAuthStateRepository,
	oauthMgr OAuthManager,
	tokenMgr TokenManager,
) *authUseCase {
	return &authUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		oauthStateRepo: oauthStateRepo,
		oauthMgr:       oauthMgr,
		tokenMgr:       tokenMgr,
	}
}

func (uc *authUseCase) GetAuthURL(ctx context.Context, provider string) (string, error) {
	state, err := uc.generateState()
	if err != nil {
		return "", domain.NewInternalError("generate state", err)
	}

	oauthState := &domain.OAuthState{
		State:     state,
		Provider:  provider,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Used:      false,
	}

	if err := uc.oauthStateRepo.Create(ctx, oauthState); err != nil {
		return "", err
	}

	url, err := uc.oauthMgr.GetAuthURL(provider, state)
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

	userInfo, err := uc.oauthMgr.GetUserInfo(ctx, provider, code)
	if err != nil {
		return nil, nil, domain.NewExternalServiceError("OAuth", err)
	}

	user, err := uc.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
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
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, nil, err
		}
	} else {
		user.FullName = userInfo.Name
		user.Avatar = &userInfo.Picture
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, nil, err
		}
	}

	sessionID := uuid.New().String()
	session := &domain.Session{
		SessionID: sessionID,
		UserID:    user.UserID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		Revoked:   false,
	}
	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, err
	}

	accessToken, err := uc.tokenMgr.GenerateAccessToken(user.UserID, sessionID)
	if err != nil {
		return nil, nil, domain.NewInternalError("generate access token", err)
	}

	refreshToken, err := uc.tokenMgr.GenerateRefreshToken(user.UserID, sessionID)
	if err != nil {
		return nil, nil, domain.NewInternalError("generate refresh token", err)
	}

	tokens := &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}

	return user, tokens, nil
}

func (uc *authUseCase) validateState(ctx context.Context, state, provider string) error {
	oauthState, err := uc.oauthStateRepo.Get(ctx, state)
	if err != nil {
		slog.Warn("authUseCase.validateState: invalid or expired OAuth state", "error", err, "state", state)
		return domain.NewUnauthorizedError("invalid or expired OAuth state", err)
	}

	if oauthState.Provider != provider {
		slog.Warn("authUseCase.validateState: state provider mismatch", "expected", provider, "actual", oauthState.Provider, "state", state)
		return domain.NewUnauthorizedError("state provider mismatch", nil)
	}

	if time.Now().After(oauthState.ExpiresAt) {
		slog.Warn("authUseCase.validateState: OAuth state expired", "state", state, "expires_at", oauthState.ExpiresAt)
		return domain.NewUnauthorizedError("OAuth state expired", nil)
	}

	if oauthState.Used {
		slog.Warn("authUseCase.validateState: OAuth state already used", "state", state)
		return domain.NewUnauthorizedError("OAuth state already used", nil)
	}

	if err := uc.oauthStateRepo.MarkUsed(ctx, state); err != nil {
		return err
	}

	return nil
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	claims, err := uc.tokenMgr.ValidateToken(refreshToken)
	if err != nil {
		slog.Warn("authUseCase.RefreshTokens: invalid refresh token", "error", err)
		return nil, domain.NewUnauthorizedError("invalid refresh token", err)
	}

	session, err := uc.sessionRepo.GetByID(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Revoked || session.ExpiresAt.Before(time.Now()) {
		slog.Warn("authUseCase.RefreshTokens: session expired or revoked", "session_id", claims.SessionID, "user_id", claims.UserID)
		return nil, domain.NewUnauthorizedError("session expired or revoked", nil)
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		slog.Warn("authUseCase.RefreshTokens: user is not active", "user_id", user.UserID)
		return nil, domain.NewForbiddenError("user is not active")
	}

	accessToken, err := uc.tokenMgr.GenerateAccessToken(user.UserID, claims.SessionID)
	if err != nil {
		return nil, domain.NewInternalError("generate access token", err)
	}

	tokens := &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    claims.SessionID,
	}

	return tokens, nil
}

func (uc *authUseCase) Logout(ctx context.Context, sessionID string) error {
	return uc.sessionRepo.Revoke(ctx, sessionID)
}

func (uc *authUseCase) LogoutAll(ctx context.Context, userID int) error {
	return uc.sessionRepo.RevokeAllByUser(ctx, userID)
}

func (uc *authUseCase) ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Revoked || session.ExpiresAt.Before(time.Now()) {
		slog.Warn("authUseCase.ValidateSession: session invalid or expired", "session_id", sessionID)
		return nil, domain.NewUnauthorizedError("session invalid or expired", nil)
	}
	return session, nil
}

func (uc *authUseCase) GetSupportedProviders(ctx context.Context) []string {
	return uc.oauthMgr.GetSupportedProviders()
}
