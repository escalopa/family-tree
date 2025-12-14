package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type authMiddleware struct {
	tokenMgr    TokenManager
	authUseCase AuthUseCase
}

func NewAuthMiddleware(tokenMgr TokenManager, authUseCase AuthUseCase) *authMiddleware {
	return &authMiddleware{
		tokenMgr:    tokenMgr,
		authUseCase: authUseCase,
	}
}

func (m *authMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get access token from cookie
		accessToken, err := c.Cookie("auth_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.tokenMgr.ValidateToken(accessToken)
		if err != nil {
			// Try to refresh token
			refreshToken, err := c.Cookie("refresh_token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				c.Abort()
				return
			}

			// Refresh tokens
			tokens, err := m.authUseCase.RefreshTokens(c.Request.Context(), refreshToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to refresh token"})
				c.Abort()
				return
			}

			// Set new cookies
			c.SetCookie("auth_token", tokens.AccessToken, 3600, "/", "", false, true)
			c.SetCookie("refresh_token", tokens.RefreshToken, 7*24*3600, "/", "", false, true)

			// Validate new token
			claims, err = m.tokenMgr.ValidateToken(tokens.AccessToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token after refresh"})
				c.Abort()
				return
			}
		}

		// Validate session
		_, err = m.authUseCase.ValidateSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.RoleID)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}

func GetUserID(c *gin.Context) int {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(int)
}

func GetUserRole(c *gin.Context) int {
	roleID, exists := c.Get("user_role")
	if !exists {
		return 0
	}
	return roleID.(int)
}

func GetSessionID(c *gin.Context) string {
	sessionID, exists := c.Get("session_id")
	if !exists {
		return ""
	}
	return sessionID.(string)
}
