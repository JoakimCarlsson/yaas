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

	//rate limit somewhere else, such as a load-balancer etc.
	mux.Handle("POST /register", middleware.NewRateLimiter(10*time.Second, 5).RateLimit(authHandler.Register))
	mux.Handle("POST /login", middleware.NewRateLimiter(10*time.Second, 5).RateLimit(authHandler.Login))
	mux.Handle("POST /logout", middleware.NewRateLimiter(10*time.Second, 5).RateLimit(tokenHandler.Logout))
	mux.Handle("POST /refresh_token", middleware.NewRateLimiter(10*time.Second, 5).RateLimit(tokenHandler.RefreshToken))

	mux.Handle("GET /auth/login", middleware.NewRateLimiter(10*time.Second, 5).RateLimit(oauthHandler.OAuthLogin))
	mux.Handle("GET /auth/callback", middleware.NewRateLimiter(10*time.Second, 5).RateLimit(oauthHandler.OAuthCallback))

	actionLimiter := middleware.NewRateLimiter(5*time.Second, 10)
	mux.Handle("GET /actions", actionLimiter.RateLimit(actionHandler.GetActions))
	mux.Handle("POST /actions", actionLimiter.RateLimit(actionHandler.CreateAction))
	mux.Handle("PUT /actions/{id}", actionLimiter.RateLimit(actionHandler.UpdateAction))
	mux.Handle("DELETE /actions/{id}", actionLimiter.RateLimit(actionHandler.DeleteAction))

	return mux
}
