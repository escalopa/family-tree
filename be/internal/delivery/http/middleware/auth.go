package middleware

import (
	"context"

	"github.com/escalopa/family-tree-api/internal/pkg/token"
	"github.com/gin-gonic/gin"
)

// Define local interfaces rather than importing concrete repos
type SessionRepo interface {
	// subset used by middleware
	DeleteExpired(ctx context.Context) error
}

type UserRepo interface{}

type TokenManager interface {
	ValidateAuthToken(tokenString string) (*token.Claims, error)
	ValidateRefreshToken(tokenString string) (*token.Claims, error)
}

type AuthMiddleware struct {
	tokenMgr    TokenManager
	sessionRepo SessionRepo
	userRepo    UserRepo
}

func NewAuthMiddleware(
	tokenMgr TokenManager,
	sessionRepo SessionRepo,
	userRepo UserRepo,
) *AuthMiddleware {
	return &AuthMiddleware{
		tokenMgr:    tokenMgr,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implementation
		// 1. Get auth_token from cookie
		// 2. Validate token
		// 3. If expired, try refresh_token
		// 4. Get session and user
		// 5. Set user in context
		c.Next()
	}
}
