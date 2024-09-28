package handlers

import (
	"encoding/json"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"github.com/joakimcarlsson/yaas/internal/utils"
	"net/http"
	"strconv"
	"strings"
)

type ActionAdminHandler struct {
	actionRepo repository.ActionRepository
}

func NewActionAdminHandler(actionRepo repository.ActionRepository) *ActionAdminHandler {
	return &ActionAdminHandler{actionRepo: actionRepo}
}

func (h *ActionAdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		h.CreateAction(w, r)
	case r.Method == http.MethodPut:
		h.UpdateAction(w, r)
	case r.Method == http.MethodDelete:
		h.DeleteAction(w, r)
	case r.Method == http.MethodGet:
		h.GetActions(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ActionAdminHandler) CreateAction(w http.ResponseWriter, r *http.Request) {
	var action models.Action
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.actionRepo.CreateAction(r.Context(), &action); err != nil {
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

	if err := h.actionRepo.UpdateAction(r.Context(), &action); err != nil {
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

	if err := h.actionRepo.DeleteAction(r.Context(), id); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to delete action")
		return
	}

	utils.JSONResponse(w, http.StatusOK, map[string]string{"message": "Action deleted successfully"})
}

func (h *ActionAdminHandler) GetActions(w http.ResponseWriter, r *http.Request) {
	actions, err := h.actionRepo.GetAllActions(r.Context())
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to fetch actions")
		return
	}

	utils.JSONResponse(w, http.StatusOK, actions)
}
