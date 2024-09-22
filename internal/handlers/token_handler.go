package handlers

import (
	"encoding/json"
	"github.com/joakimcarlsson/yaas/internal/services"
	"github.com/joakimcarlsson/yaas/internal/utils"
	"net/http"
)

type TokenHandler struct {
	AuthService services.AuthService
}

func NewTokenHandler(authService services.AuthService) *TokenHandler {
	return &TokenHandler{
		AuthService: authService,
	}
}

func (h *TokenHandler) RefreshToken(w http.ResponseWriter, r *http.Request) { //ideally we should get it from headers / http only cookies
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
