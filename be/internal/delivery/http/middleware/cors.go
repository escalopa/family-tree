package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implementation
		// Set CORS headers
		c.Next()
	}
}
