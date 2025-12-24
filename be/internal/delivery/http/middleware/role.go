package middleware

import (
	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

func RequireRole(minRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole < minRole {
			delivery.Error(c, domain.NewForbiddenError("error.insufficient_permissions"))
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
			delivery.Error(c, domain.NewForbiddenError("error.account_not_activated_by_admin"))
			c.Abort()
			return
		}
		c.Next()
	}
}
