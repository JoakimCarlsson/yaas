package server

import (
	"github.com/joakimcarlsson/yaas/internal/models"
	"net/http"
	"time"

	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/middleware"
)

func NewRouter(flowHandler *handlers.FlowHandler, tokenHandler *handlers.TokenHandler) *http.ServeMux {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(5 * time.Minute)

	mux.HandleFunc("/self-service/token/refresh", limiter.RateLimit(tokenHandler.RefreshToken))

	mux.HandleFunc("/self-service/oauth/callback", limiter.RateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			flowHandler.ProceedOAuthLoginFlow(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/self-service/registration", limiter.RateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			flowHandler.InitiateFlow(w, r, models.FlowTypeRegistration)
		} else if r.Method == http.MethodPost {
			flowHandler.ProceedFlow(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/self-service/login", limiter.RateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			flowHandler.InitiateFlow(w, r, models.FlowTypeLogin)
		} else if r.Method == http.MethodPost {
			flowHandler.ProceedFlow(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/self-service/logout", limiter.RateLimit(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			flowHandler.InitiateFlow(w, r, models.FlowTypeLogout)
		} else if r.Method == http.MethodPost {
			flowHandler.ProceedFlow(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	return mux
}
