package server

import (
	"net/http"
	"time"

	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/middleware"
)

func NewRouter(flowHandler *handlers.FlowHandler, tokenHandler *handlers.TokenHandler) *http.ServeMux {
	mux := http.NewServeMux()
	limiter := middleware.NewRateLimiter(5 * time.Minute)

	mux.HandleFunc("/refresh_token", limiter.RateLimit(tokenHandler.RefreshToken))

	//oauth2 flows
	mux.HandleFunc("/self-service/oauth/login/flows", limiter.RateLimit(flowHandler.InitiateOAuthLoginFlow))
	mux.HandleFunc("/self-service/oauth/callback", limiter.RateLimit(flowHandler.ProceedOAuthLoginFlow))

	//registration flows
	mux.HandleFunc("/self-service/registration/flows", limiter.RateLimit(flowHandler.InitiateRegistrationFlow))
	mux.HandleFunc("/self-service/registration", limiter.RateLimit(flowHandler.ProceedRegistrationFlow))

	//login flows
	mux.HandleFunc("/self-service/login/flows", limiter.RateLimit(flowHandler.InitiateLoginFlow))
	mux.HandleFunc("/self-service/login", limiter.RateLimit(flowHandler.ProceedLoginFlow))

	//logout
	mux.HandleFunc("/self-service/logout/flows", limiter.RateLimit(flowHandler.InitiateLogoutFlow))
	mux.HandleFunc("/self-service/logout", limiter.RateLimit(flowHandler.ProceedLogoutFlow))
	return mux
}
