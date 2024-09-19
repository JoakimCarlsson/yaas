package server

import (
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/middleware"
	"golang.org/x/time/rate"
)

func NewRouter(authHandler *handlers.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	limiter := middleware.NewIPRateLimiter(rate.Limit(1), 1)
	mux.HandleFunc("/register", middleware.RateLimitMiddleware(limiter)(authHandler.Register))
	mux.HandleFunc("/login", middleware.RateLimitMiddleware(limiter)(authHandler.Login))
	mux.HandleFunc("/refresh_token", middleware.RateLimitMiddleware(limiter)(authHandler.RefreshToken))
	mux.HandleFunc("/logout", middleware.RateLimitMiddleware(limiter)(authHandler.Logout))

	return mux
}
