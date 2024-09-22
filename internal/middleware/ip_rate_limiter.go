package middleware

import (
	"fmt"
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
	mu       sync.RWMutex
	configs  map[string]RateLimiterConfig
	cleanup  time.Duration
}

func NewRateLimiter(cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorInfo),
		configs: map[string]RateLimiterConfig{
			"default":        {Limit: rate.Every(time.Second), Burst: 10},
			"/login":         {Limit: rate.Every(10 * time.Second), Burst: 5},
			"/register":      {Limit: rate.Every(10 * time.Minute), Burst: 3},
			"/refresh_token": {Limit: rate.Every(10 * time.Second), Burst: 5},
			"/logout":        {Limit: rate.Every(10 * time.Second), Burst: 5},
		},
		cleanup: cleanup,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(ip, endpoint string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := ip + ":" + endpoint
	v, exists := rl.visitors[key]
	if !exists {
		config, ok := rl.configs[endpoint]
		if !ok {
			config = rl.configs["default"]
		}
		limiter := rate.NewLimiter(config.Limit, config.Burst)
		rl.visitors[key] = &visitorInfo{limiter: limiter, lastSeen: time.Now()}
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
		endpoint := r.URL.Path
		limiter := rl.getVisitor(ip, endpoint)

		if !limiter.Allow() {
			config, ok := rl.configs[endpoint]
			if !ok {
				config = rl.configs["default"]
			}

			retryAfter := calculateRetryAfter(config.Limit)

			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			logger.WithFields(logger.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      http.StatusTooManyRequests,
				"ip":          ip,
				"user_agent":  r.UserAgent(),
				"retry_after": retryAfter,
			}).Warn("Rate limit exceeded")

			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func calculateRetryAfter(limit rate.Limit) int {
	if limit <= 0 {
		return 60
	}
	return int(1 / float64(limit))
}
