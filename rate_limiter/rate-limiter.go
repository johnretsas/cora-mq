package rate_limiter

import (
	"log"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiterConfig struct {
	clientRateLimiters map[string]*rate.Limiter
	mu                 sync.Mutex
	rateLimit          rate.Limit
	burst              int
	logger             *log.Logger
}

func NewRateLimiterConfig(rateLimit rate.Limit, burst int, logger *log.Logger) *RateLimiterConfig {
	return &RateLimiterConfig{
		clientRateLimiters: make(map[string]*rate.Limiter),
		rateLimit:          rateLimit,
		burst:              burst,
		logger:             nil,
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

func (rl *RateLimiterConfig) AllowRequest(clientId string, bypass bool) bool {
	if bypass {
		return true
	}
	return rl.getLimiter(clientId).Allow()
}
