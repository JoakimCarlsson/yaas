package handlers

import (
	"encoding/json"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/services"
	"github.com/joakimcarlsson/yaas/internal/utils"
	"net/http"
	"strconv"
	"strings"
)

type ActionAdminHandler struct {
	actionService services.ActionService
}

func NewActionAdminHandler(actionService services.ActionService) *ActionAdminHandler {
	return &ActionAdminHandler{actionService: actionService}
}

func (h *ActionAdminHandler) CreateAction(w http.ResponseWriter, r *http.Request) {
	var action models.Action
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.actionService.CreateAction(r.Context(), &action); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create action")
		return
	}

	utils.JSONResponse(w, http.StatusCreated, action)
}

func (h *ActionAdminHandler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid action ID")
		return
	}

	var action models.Action
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	action.ID = id

	if err := h.actionService.UpdateAction(r.Context(), &action); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update action")
		return
	}

	utils.JSONResponse(w, http.StatusOK, action)
}

func (h *ActionAdminHandler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	id, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid action ID")
		return
	}

	if err := h.actionService.DeleteAction(r.Context(), id); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to delete action")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Action deleted successfully"})
}

func (h *ActionAdminHandler) GetActions(w http.ResponseWriter, r *http.Request) {
	actions, err := h.actionService.GetAllActions(r.Context())
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to fetch actions")
		return
	}

	utils.JSONResponse(w, http.StatusOK, actions)
}
