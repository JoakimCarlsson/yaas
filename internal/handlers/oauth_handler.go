package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joakimcarlsson/yaas/internal/services"
)

type OAuthHandler struct {
	OAuth2Service services.OAuth2Service
	AuthService   services.AuthService
}

func NewOAuthHandler(oauth2Service services.OAuth2Service, authService services.AuthService) *OAuthHandler {
	return &OAuthHandler{
		OAuth2Service: oauth2Service,
		AuthService:   authService,
	}
}

type StateClaims struct {
	CallbackURL string `json:"callback_url"`
	jwt.RegisteredClaims
}

func (h *OAuthHandler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	callbackURL := r.URL.Query().Get("callback_url")
	if provider == "" || callbackURL == "" {
		http.Error(w, "Provider and callback URL are required", http.StatusBadRequest)
		return
	}

	stateToken, err := h.AuthService.GenerateStateToken(callbackURL)
	if err != nil {
		http.Error(w, "Failed to generate state token", http.StatusInternalServerError)
		return
	}

	loginURL, err := h.OAuth2Service.GetLoginURL(provider, stateToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if provider == "" || state == "" || code == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	callbackURL, err := h.AuthService.ValidateStateToken(state)
	if err != nil {
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	token, err := h.OAuth2Service.ExchangeCodeForToken(provider, code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	userInfo, err := h.OAuth2Service.GetUserInfo(provider, token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	_, accessToken, refreshToken, err := h.AuthService.ProcessOAuthLogin(r.Context(), provider, userInfo, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, callbackURL+"?accessToken="+accessToken+"&refreshToken="+refreshToken, http.StatusFound)
}
