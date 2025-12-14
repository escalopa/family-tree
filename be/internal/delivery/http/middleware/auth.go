package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	keyUserID    = "user_id"
	keyUserEmail = "user_email"
	keyUserRole  = "user_role"
	keySessionID = "session_id"
)

type authMiddleware struct {
	tokenMgr      TokenManager
	authUseCase   AuthUseCase
	cookieManager CookieManager
}

func NewAuthMiddleware(tokenMgr TokenManager, authUseCase AuthUseCase, cookieManager CookieManager) *authMiddleware {
	return &authMiddleware{
		tokenMgr:      tokenMgr,
		authUseCase:   authUseCase,
		cookieManager: cookieManager,
	}
}

func (m *authMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := m.cookieManager.GetAccessToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			c.Abort()
			return
		}

		claims, err := m.tokenMgr.ValidateToken(accessToken)
		if err != nil {
			refreshToken, err := m.cookieManager.GetRefreshToken(c)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				c.Abort()
				return
			}

			tokens, err := m.authUseCase.RefreshTokens(c.Request.Context(), refreshToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to refresh token"})
				c.Abort()
				return
			}

			m.cookieManager.SetTokenCookies(c, tokens.AccessToken, tokens.RefreshToken)

			claims, err = m.tokenMgr.ValidateToken(tokens.AccessToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token after refresh"})
				c.Abort()
				return
			}
		}

		_, err = m.authUseCase.ValidateSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		c.Set(keyUserID, claims.UserID)
		c.Set(keyUserEmail, claims.Email)
		c.Set(keyUserRole, claims.RoleID)
		c.Set(keySessionID, claims.SessionID)

		c.Next()
	}
}

func GetUserID(c *gin.Context) int {
	userID, exists := c.Get(keyUserID)
	if !exists {
		return 0
	}
	return userID.(int)
}

func GetUserRole(c *gin.Context) int {
	roleID, exists := c.Get(keyUserRole)
	if !exists {
		return 0
	}
	return roleID.(int)
}

func GetSessionID(c *gin.Context) string {
	sessionID, exists := c.Get(keySessionID)
	if !exists {
		return ""
	}
	return sessionID.(string)
}
