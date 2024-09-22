package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/services"
	"github.com/joakimcarlsson/yaas/internal/utils"
)

type AuthHandler struct {
	AuthService   services.AuthService
	OAuth2Service services.OAuth2Service
}

func NewAuthHandler(authService services.AuthService, oauth2Service services.OAuth2Service) *AuthHandler {
	return &AuthHandler{
		AuthService:   authService,
		OAuth2Service: oauth2Service,
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
