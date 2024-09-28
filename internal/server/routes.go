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
	actionHandler *handlers.ActionAdminHandler,
) *http.ServeMux {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(5 * time.Minute)

	// Auth routes
	mux.Handle("POST /register", limiter.RateLimit(authHandler.Register))
	mux.Handle("POST /login", limiter.RateLimit(authHandler.Login))

	// Token routes
	mux.Handle("POST /logout", limiter.RateLimit(tokenHandler.Logout))
	mux.Handle("POST /refresh_token", limiter.RateLimit(tokenHandler.RefreshToken))

	// OAuth routes
	mux.Handle("GET /auth/login", limiter.RateLimit(oauthHandler.OAuthLogin))
	mux.Handle("GET /auth/callback", limiter.RateLimit(oauthHandler.OAuthCallback))

	// Actions routes
	mux.Handle("GET /actions", limiter.RateLimit(actionHandler.GetActions))
	mux.Handle("POST /actions", limiter.RateLimit(actionHandler.GetActions))
	mux.Handle("PUT /actions/{id}", limiter.RateLimit(actionHandler.UpdateAction))
	mux.Handle("DELETE /actions/{id}", limiter.RateLimit(actionHandler.DeleteAction))

	return mux
}
