package server

import (
	"net/http"
	"time"

	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/middleware"
)

func NewRouter(authHandler *handlers.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(5 * time.Minute)

	mux.HandleFunc("/register", limiter.RateLimit(authHandler.Register))
	mux.HandleFunc("/login", limiter.RateLimit(authHandler.Login))
	mux.HandleFunc("/refresh_token", limiter.RateLimit(authHandler.RefreshToken))
	mux.HandleFunc("/logout", limiter.RateLimit(authHandler.Logout))
	mux.HandleFunc("/auth/google/login", limiter.RateLimit(authHandler.GoogleLogin))
	mux.HandleFunc("/auth/google/callback", limiter.RateLimit(authHandler.GoogleCallback))

	return mux
}
