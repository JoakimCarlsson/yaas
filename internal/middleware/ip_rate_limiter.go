package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/joakimcarlsson/yaas/internal/logger"
	"golang.org/x/time/rate"
)

type visitorInfo struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterConfig struct {
	Limit rate.Limit
	Burst int
}

type RateLimiter struct {
	visitors map[string]*visitorInfo
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
	cleanup  time.Duration
}

func NewRateLimiter(limit time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorInfo),
		limit:    rate.Every(limit),
		burst:    burst,
		cleanup:  5 * time.Minute,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.limit, rl.burst)
		rl.visitors[ip] = &visitorInfo{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanup)

		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.cleanup {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60") // Adjustable retry time
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)

			logger.WithFields(logger.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     http.StatusTooManyRequests,
				"ip":         ip,
				"user_agent": r.UserAgent(),
			}).Warn("Rate limit exceeded")

			return
		}

		next.ServeHTTP(w, r)
	}
}
