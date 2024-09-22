package server

import (
	"net/http"
	"time"

	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/middleware"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	oauthHandler *handlers.OAuthHandler,
	tokenHandler *handlers.TokenHandler,
) *http.ServeMux {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(5 * time.Minute)

	mux.HandleFunc("/register", limiter.RateLimit(authHandler.Register))
	mux.HandleFunc("/login", limiter.RateLimit(authHandler.Login))

	mux.HandleFunc("/logout", limiter.RateLimit(tokenHandler.Logout))
	mux.HandleFunc("/refresh_token", limiter.RateLimit(tokenHandler.RefreshToken))

	mux.HandleFunc("/auth/login", limiter.RateLimit(oauthHandler.OAuthLogin))
	mux.HandleFunc("/auth/callback", limiter.RateLimit(oauthHandler.OAuthCallback))

	return mux
}
