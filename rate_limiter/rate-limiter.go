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

func (rl *RateLimiterConfig) getLimiter(clientId string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clientRateLimiters[clientId]
	if !exists {
		limiter = rate.NewLimiter(rl.rateLimit, rl.burst)
		rl.clientRateLimiters[clientId] = limiter
	}

	return limiter
}

func (rl *RateLimiterConfig) AllowRequest(clientId string) bool {
	return rl.getLimiter(clientId).Allow()
}
