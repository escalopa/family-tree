package middleware

import (
	"github.com/escalopa/family-tree-api/internal/domain"
	"github.com/gin-gonic/gin"
)

func RequireRole(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implementation
		// 1. Get user from context
		// 2. Check if user.RoleID >= minRole
		// 3. Abort if insufficient permissions
		c.Next()
	}
}

func RequireViewer() gin.HandlerFunc {
	return RequireRole(domain.RoleViewer)
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}

func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleSuperAdmin)
}
