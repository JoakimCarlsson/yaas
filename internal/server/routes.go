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

	mux.HandleFunc("/register", limiter.RateLimit(authHandler.Register)) //todo add proper http.StatusMethodNotAllowed etc.
	mux.HandleFunc("/login", limiter.RateLimit(authHandler.Login))

	mux.HandleFunc("/logout", limiter.RateLimit(tokenHandler.Logout))
	mux.HandleFunc("/refresh_token", limiter.RateLimit(tokenHandler.RefreshToken))

	mux.HandleFunc("/auth/login", limiter.RateLimit(oauthHandler.OAuthLogin))
	mux.HandleFunc("/auth/callback", limiter.RateLimit(oauthHandler.OAuthCallback))

	mux.HandleFunc("/actions", limiter.RateLimit(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			handlePreflight(w, r)
			return
		}

		// Set CORS headers for actual requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		switch r.Method {
		case http.MethodGet:
			actionHandler.GetActions(w, r)
		case http.MethodPost:
			actionHandler.CreateAction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/actions/", limiter.RateLimit(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			handlePreflight(w, r)
			return
		}

		// Set CORS headers for actual requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		switch r.Method {
		case http.MethodPut:
			actionHandler.UpdateAction(w, r)
		case http.MethodDelete:
			actionHandler.DeleteAction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	return mux
}

func handlePreflight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusNoContent)
}
