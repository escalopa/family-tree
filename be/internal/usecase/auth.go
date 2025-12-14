package usecase

import (
	"context"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/google/uuid"
)

type authUseCase struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
	oauthMgr    OAuthManager
	tokenMgr    TokenManager
}

func NewAuthUseCase(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	oauthMgr OAuthManager,
	tokenMgr TokenManager,
) *authUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		oauthMgr:    oauthMgr,
		tokenMgr:    tokenMgr,
	}
}

// GetAuthURL returns the OAuth URL for the specified provider
func (uc *authUseCase) GetAuthURL(provider, state string) (string, error) {
	return uc.oauthMgr.GetAuthURL(provider, state)
}

// HandleCallback handles OAuth callback for any provider
func (uc *authUseCase) HandleCallback(ctx context.Context, provider, code string) (*domain.User, *domain.AuthTokens, error) {
	// Get user info from OAuth provider (exchange happens in the provider implementation)
	userInfo, err := uc.oauthMgr.GetUserInfo(ctx, provider, code)
	if err != nil {
		return nil, nil, domain.NewExternalServiceError("OAuth", err)
	}

	// Check if user exists
	user, err := uc.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		// Create new user with "none" role
		user = &domain.User{
			FullName: userInfo.Name,
			Email:    userInfo.Email,
			Avatar:   &userInfo.Picture,
			RoleID:   domain.RoleNone,
			IsActive: false, // Needs admin approval
		}
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, nil, err
		}
	} else {
		// Update existing user info from OAuth
		user.FullName = userInfo.Name
		user.Avatar = &userInfo.Picture
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, nil, err
		}
	}

	// Create session
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

	// Generate tokens
	accessToken, err := uc.tokenMgr.GenerateAccessToken(user.UserID, user.Email, user.RoleID, sessionID)
	if err != nil {
		return nil, nil, domain.NewInternalError("failed to generate access token", err)
	}

	refreshToken, err := uc.tokenMgr.GenerateRefreshToken(user.UserID, sessionID)
	if err != nil {
		return nil, nil, domain.NewInternalError("failed to generate refresh token", err)
	}

	tokens := &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}

	return user, tokens, nil
}

func (uc *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	// Validate refresh token
	claims, err := uc.tokenMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, domain.NewUnauthorizedError("invalid refresh token", err)
	}

	// Check session
	session, err := uc.sessionRepo.GetByID(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.Revoked || session.ExpiresAt.Before(time.Now()) {
		return nil, domain.NewUnauthorizedError("session expired or revoked", nil)
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, domain.NewForbiddenError("user is not active")
	}

	// Generate new access token
	accessToken, err := uc.tokenMgr.GenerateAccessToken(user.UserID, user.Email, user.RoleID, claims.SessionID)
	if err != nil {
		return nil, domain.NewInternalError("failed to generate access token", err)
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
		return nil, domain.NewUnauthorizedError("session invalid or expired", nil)
	}
	return session, nil
}
