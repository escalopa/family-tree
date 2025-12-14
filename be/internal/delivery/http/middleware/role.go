package middleware

import (
	"net/http"

	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

func RequireRole(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole < minRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == domain.RoleNone {
			c.JSON(http.StatusForbidden, gin.H{"error": "account not activated by admin"})
			c.Abort()
			return
		}
		c.Next()
	}
}


