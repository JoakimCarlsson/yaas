package handlers

import (
	"net/http"
	"net/url"
	"time"

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

var stateSecret = []byte("your-secret-key")

type StateClaims struct {
	CallbackURL string `json:"callback_url"`
	jwt.RegisteredClaims
}

func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	callbackURL := r.URL.Query().Get("callback_url")
	if callbackURL == "" {
		http.Error(w, "Callback URL is required", http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &StateClaims{
		CallbackURL: callbackURL,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(stateSecret)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	googleURL := h.OAuth2Service.GetGoogleLoginURL(state)
	http.Redirect(w, r, googleURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "Missing state parameter", http.StatusBadRequest)
		return
	}

	claims := &StateClaims{}
	token, err := jwt.ParseWithClaims(state, claims, func(token *jwt.Token) (interface{}, error) {
		return stateSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired state token", http.StatusBadRequest)
		return
	}

	callbackURL := claims.CallbackURL

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not provided", http.StatusBadRequest)
		return
	}

	tokenData, err := h.OAuth2Service.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	_, accessToken, refreshToken, err := h.AuthService.GoogleSignIn(r.Context(), tokenData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL, err := url.Parse(callbackURL)
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
