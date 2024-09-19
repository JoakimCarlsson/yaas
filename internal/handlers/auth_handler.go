package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/services"
	"github.com/joakimcarlsson/yaas/internal/utils"
)

type AuthHandler struct {
	AuthService   services.AuthService
	OAuth2Service services.OAuth2Service
	stateMutex    sync.Mutex
	stateStore    map[string]stateData
}

type stateData struct {
	CallbackURL string
	Expiry      time.Time
}

func NewAuthHandler(authService services.AuthService, oauth2Service services.OAuth2Service) *AuthHandler {
	return &AuthHandler{
		AuthService:   authService,
		OAuth2Service: oauth2Service,
		stateStore:    make(map[string]stateData),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: &req.Password,
	}

	err := h.AuthService.Register(r.Context(), user, req.Password)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusCreated, map[string]string{
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	_, accessToken, refreshToken, err := h.AuthService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) { //ideally we should get it from headers / http only cookies
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	newAccessToken, newRefreshToken, err := h.AuthService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		utils.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"accessToken":  newAccessToken,
		"refreshToken": newRefreshToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.AuthService.Logout(r.Context(), req.RefreshToken); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {

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

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
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
