package middleware

import (
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

		ip := c.ClientIP()

		allowed, err := m.limiter.Allow(c.Request.Context(), ip)
		if err != nil {
			slog.Error("Rate limiter error", "error", err, "ip", ip)
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
