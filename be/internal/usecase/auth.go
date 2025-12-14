package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/escalopa/family-tree/internal/pkg/oauth"
	"github.com/google/uuid"
)

type authUseCase struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
	googleOAuth oauth.OAuthProvider
	tokenMgr    TokenManager
}

func NewAuthUseCase(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	googleOAuth oauth.OAuthProvider,
	tokenMgr TokenManager,
) *authUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		googleOAuth: googleOAuth,
		tokenMgr:    tokenMgr,
	}
}

// GetAuthURL returns the OAuth URL for the specified provider
func (uc *authUseCase) GetAuthURL(provider string) (string, error) {
	// For now, only Google is supported
	if provider != "google" {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
	return uc.googleOAuth.GetAuthURL(""), nil
}

// HandleCallback handles OAuth callback for any provider
func (uc *authUseCase) HandleCallback(ctx context.Context, provider, code, state string) (*domain.User, *domain.AuthTokens, error) {
	// For now, only Google is supported
	if provider != "google" {
		return nil, nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
	return uc.handleGoogleCallback(ctx, code)
}

// Legacy method for backwards compatibility
func (uc *authUseCase) GetGoogleAuthURL(state string) string {
	return uc.googleOAuth.GetAuthURL(state)
}

func (uc *authUseCase) handleGoogleCallback(ctx context.Context, code string) (*domain.User, *domain.AuthTokens, error) {
	// Exchange code for token
	oauthToken, err := uc.googleOAuth.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := uc.googleOAuth.GetUserInfo(ctx, oauthToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check if user exists
	user, err := uc.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
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
			return nil, nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user info from OAuth
		user.FullName = userInfo.Name
		user.Avatar = &userInfo.Picture
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, nil, fmt.Errorf("failed to update user: %w", err)
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
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate tokens
	accessToken, err := uc.tokenMgr.GenerateAccessToken(user.UserID, user.Email, user.RoleID, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.tokenMgr.GenerateRefreshToken(user.UserID, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
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
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check session
	session, err := uc.sessionRepo.GetByID(ctx, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil || session.Revoked || session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session expired or revoked")
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user is not active")
	}

	// Generate new access token
	accessToken, err := uc.tokenMgr.GenerateAccessToken(user.UserID, user.Email, user.RoleID, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
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
		return nil, fmt.Errorf("session invalid or expired")
	}
	return session, nil
}
