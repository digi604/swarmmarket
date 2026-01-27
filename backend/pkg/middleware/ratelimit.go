package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/digi604/swarmmarket/backend/internal/common"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // tokens per second
	burst    int           // max tokens
	cleanup  time.Duration // cleanup interval for stale buckets
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rps,
		burst:   burst,
		cleanup: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request is allowed for the given key.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]

	if !exists {
		rl.buckets[key] = &bucket{
			tokens:    float64(rl.burst - 1),
			lastCheck: now,
		}
		return true
	}

	// Calculate tokens to add based on elapsed time
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * float64(rl.rate)

	// Cap at burst limit
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}

	b.lastCheck = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

// cleanupLoop removes stale buckets periodically.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		threshold := time.Now().Add(-rl.cleanup)
		for key, b := range rl.buckets {
			if b.lastCheck.Before(threshold) {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit creates a rate limiting middleware.
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use agent ID if authenticated, otherwise use IP
			key := r.RemoteAddr
			if agent := GetAgent(r.Context()); agent != nil {
				key = agent.ID.String()
			}

			if !limiter.Allow(key) {
				common.WriteError(w, http.StatusTooManyRequests, common.ErrTooManyRequests("rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
