package rate_limiter

import (
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	clientRateLimiters map[string]*rate.Limiter
	mu                 sync.Mutex
	rateLimit          rate.Limit
	burst              int
}

func NewRateLimiterConfig(rateLimit rate.Limit, burst int) *RateLimiterConfig {
	return &RateLimiterConfig{
		clientRateLimiters: make(map[string]*rate.Limiter),
		rateLimit:          rateLimit,
		burst:              burst,
	}
}
