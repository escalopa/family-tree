package middleware

import (
	"context"

	"github.com/escalopa/family-tree/internal/domain"
)

// TokenManager interface used by middleware
type TokenManager interface {
	ValidateToken(tokenString string) (*domain.TokenClaims, error)
}

// AuthUseCase interface used by middleware
type AuthUseCase interface {
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthTokens, error)
	ValidateSession(ctx context.Context, sessionID string) (*domain.Session, error)
}
