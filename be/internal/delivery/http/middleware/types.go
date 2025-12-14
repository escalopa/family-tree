package middleware

import (
	"context"

	"github.com/escalopa/family-tree/internal/delivery/http/cookie"
	"github.com/escalopa/family-tree/internal/domain"
)

type CookieManager interface {
	SetAuthCookies(c cookie.Context, accessToken, refreshToken, sessionID string)
	SetTokenCookies(c cookie.Context, accessToken, refreshToken string)
	ClearAuthCookies(c cookie.Context)
	GetAccessToken(c cookie.Context) (string, error)
	GetRefreshToken(c cookie.Context) (string, error)
	GetSessionID(c cookie.Context) (string, error)
}

type TokenManager interface {
	ValidateToken(tokenString string) (*domain.TokenClaims, error)
}

type AuthUseCase interface {
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error)
}
