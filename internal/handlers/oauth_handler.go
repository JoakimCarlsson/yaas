package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/joakimcarlsson/yaas/internal/services"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type OAuthHandler struct {
	OAuth2Service services.OAuth2Service
	AuthService   services.AuthService
	stateStore    map[string]stateData
	stateMutex    sync.Mutex
}

func NewOAuthHandler(oauth2Service services.OAuth2Service, authService services.AuthService) *OAuthHandler {
	return &OAuthHandler{
		OAuth2Service: oauth2Service,
		AuthService:   authService,
		stateStore:    make(map[string]stateData),
	}
}

func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {

	callbackURL := r.URL.Query().Get("callback_url")
	if callbackURL == "" {
		http.Error(w, "Callback URL is required", http.StatusBadRequest)
		return
	}

	//validate callback url, and check that the domain is allowed

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	state := base64.URLEncoding.EncodeToString(b)

	h.stateMutex.Lock()
	h.stateStore[state] = stateData{
		CallbackURL: callbackURL,
		Expiry:      time.Now().Add(15 * time.Minute),
	}
	h.stateMutex.Unlock()

	googleURL := h.OAuth2Service.GetGoogleLoginURL(state)

	http.Redirect(w, r, googleURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	h.stateMutex.Lock()
	data, exists := h.stateStore[state]
	delete(h.stateStore, state)
	h.stateMutex.Unlock()

	if !exists || time.Now().After(data.Expiry) {
		http.Error(w, "Invalid or expired state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not provided", http.StatusBadRequest)
		return
	}

	token, err := h.OAuth2Service.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	_, accessToken, refreshToken, err := h.AuthService.GoogleSignIn(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL, err := url.Parse(data.CallbackURL)
	if err != nil {
		http.Error(w, "Invalid callback URL", http.StatusInternalServerError)
		return
	}

	query := redirectURL.Query()
	query.Set("accessToken", accessToken)
	query.Set("refreshToken", refreshToken)
	redirectURL.RawQuery = query.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}
