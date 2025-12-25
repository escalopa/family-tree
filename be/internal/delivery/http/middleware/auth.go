package middleware

import (
	"errors"
	"net/http"

	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

const (
	keyUserID            = "user_id"
	keyUserRole          = "user_role"
	keyIsActive          = "is_active"
	keySessionID         = "session_id"
	keyPreferredLanguage = "preferred_language"
)

type AuthMiddleware struct {
	tokenMgr      TokenManager
	authUseCase   AuthUseCase
	userRepo      UserRepository
	cookieManager CookieManager
}

func NewAuthMiddleware(tokenMgr TokenManager, authUseCase AuthUseCase, userRepo UserRepository, cookieManager CookieManager) *AuthMiddleware {
	return &AuthMiddleware{
		tokenMgr:      tokenMgr,
		authUseCase:   authUseCase,
		userRepo:      userRepo,
		cookieManager: cookieManager,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := m.cookieManager.GetAccessToken(c)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			delivery.Error(c, domain.NewUnauthorizedError("error.missing_auth_token", nil))
			c.Abort()
			return
		}

		claims, err := m.tokenMgr.ValidateToken(accessToken)
		if err != nil {
			refreshToken, err := m.cookieManager.GetRefreshToken(c)
			if err != nil {
				delivery.Error(c, domain.NewUnauthorizedError("error.invalid_or_expired_token", nil))
				c.Abort()
				return
			}

			tokens, err := m.authUseCase.RefreshTokens(c.Request.Context(), refreshToken)
			if err != nil {
				delivery.Error(c, domain.NewUnauthorizedError("error.failed_to_refresh_token", nil))
				c.Abort()
				return
			}

			m.cookieManager.SetTokenCookies(c, tokens.AccessToken, tokens.RefreshToken)

			claims, err = m.tokenMgr.ValidateToken(tokens.AccessToken)
			if err != nil {
				delivery.Error(c, domain.NewUnauthorizedError("error.invalid_refreshed_token", nil))
				c.Abort()
				return
			}
		}

		_, err = m.authUseCase.ValidateSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			delivery.Error(c, domain.NewUnauthorizedError("error.invalid_session", nil))
			c.Abort()
			return
		}

		user, err := m.userRepo.Get(c.Request.Context(), claims.UserID)
		if err != nil {
			delivery.Error(c, domain.NewUnauthorizedError("error.user_not_found", nil))
			c.Abort()
			return
		}

		c.Set(keyUserID, user.UserID)
		c.Set(keyUserRole, user.RoleID)
		c.Set(keyIsActive, user.IsActive)
		c.Set(keySessionID, claims.SessionID)
		c.Set(keyPreferredLanguage, user.PreferredLanguage)

		c.Next()
	}
}

func GetUserID(c *gin.Context) int {
	userID, exists := c.Get(keyUserID)
	if !exists {
		return 0
	}
	if id, ok := userID.(int); ok {
		return id
	}
	return 0
}

func GetUserRole(c *gin.Context) int {
	roleID, exists := c.Get(keyUserRole)
	if !exists {
		return 0
	}
	if id, ok := roleID.(int); ok {
		return id
	}
	return 0
}

func GetIsActive(c *gin.Context) bool {
	isActive, exists := c.Get(keyIsActive)
	if !exists {
		return false
	}
	if active, ok := isActive.(bool); ok {
		return active
	}
	return false
}

func GetSessionID(c *gin.Context) string {
	sessionID, exists := c.Get(keySessionID)
	if !exists {
		return ""
	}
	if id, ok := sessionID.(string); ok {
		return id
	}
	return ""
}

func GetPreferredLanguage(c *gin.Context) string {
	preferredLang, exists := c.Get(keyPreferredLanguage)
	if !exists {
		return ""
	}
	if lang, ok := preferredLang.(string); ok {
		return lang
	}
	return ""
}
