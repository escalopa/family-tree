package middleware

import (
	"fmt"
	"log/slog"

	"github.com/escalopa/family-tree/internal/delivery"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
	limiter RateLimiter
	enabled bool
}

func NewRateLimiter(limiter RateLimiter, enabled bool) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		enabled: enabled,
	}
}

func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.enabled {
			c.Next()
			return
		}

		key, ip, userID := m.getKey(c)
		allowed, err := m.limiter.Allow(c.Request.Context(), key)
		if err != nil {
			slog.Error("Rate limiter",
				"error", err,
				"key", key,
				"ip", ip,
				"userID", userID,
			)
			c.Next()
			return
		}

		if !allowed {
			delivery.Error(c, domain.NewRateLimitError())
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *RateLimitMiddleware) getKey(c *gin.Context) (string, string, int) {
	ip := c.ClientIP()
	userID := GetUserID(c)
	if userID == 0 { // anonymous user (not logged in)
		return ip, ip, 0
	}
	return fmt.Sprintf("user:%d", userID), ip, userID
}
