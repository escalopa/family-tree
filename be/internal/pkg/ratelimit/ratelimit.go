package ratelimit

import (
	"context"

	"github.com/escalopa/family-tree/internal/config"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit
	prefix  string
}

func New(client *redis.Client, cfg config.RateLimitRule) *Limiter {
	return &Limiter{
		limiter: redis_rate.NewLimiter(client),
		limit: redis_rate.Limit{
			Rate:   cfg.Requests,
			Burst:  cfg.Requests,
			Period: cfg.Window,
		},
		prefix: cfg.Prefix,
	}
}

func (l *Limiter) buildKey(key string) string {
	if l.prefix == "" {
		return key
	}
	return l.prefix + ":" + key
}

func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := l.buildKey(key)

	result, err := l.limiter.Allow(ctx, redisKey, l.limit)
	if err != nil {
		return false, err
	}

	return result.Allowed > 0, nil
}
