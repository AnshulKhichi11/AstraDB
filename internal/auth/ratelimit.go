package auth

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*userLimit
	mu       sync.RWMutex
}

type userLimit struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*userLimit),
	}

	// Cleanup old entries every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Middleware limits requests per user
func (rl *RateLimiter) Middleware(requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				// No user in context (unauthenticated)
				next.ServeHTTP(w, r)
				return
			}

			key := user.Username

			rl.mu.Lock()
			limit, exists := rl.requests[key]
			now := time.Now()

			if !exists || now.After(limit.resetTime) {
				// New window
				rl.requests[key] = &userLimit{
					count:     1,
					resetTime: now.Add(1 * time.Minute),
				}
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if limit.count >= requestsPerMinute {
				rl.mu.Unlock()
				respondError(w, http.StatusTooManyRequests, "Rate limit exceeded")
				return
			}

			limit.count++
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, limit := range rl.requests {
		if now.After(limit.resetTime) {
			delete(rl.requests, key)
		}
	}
}