package middleware

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
)

type CookieManager interface {
	SetAuthCookies(c domain.CookieContext, accessToken, refreshToken, sessionID string)
	SetTokenCookies(c domain.CookieContext, accessToken, refreshToken string)
	ClearAuthCookies(c domain.CookieContext)
	GetAccessToken(c domain.CookieContext) (string, error)
	GetRefreshToken(c domain.CookieContext) (string, error)
	GetSessionID(c domain.CookieContext) (string, error)
}

type TokenManager interface {
	ValidateToken(tokenString string) (*domain.TokenClaims, error)
}

type AuthUseCase interface {
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error)
}

type UserRepository interface {
	Get(ctx context.Context, userID int) (*domain.User, error)
}

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
