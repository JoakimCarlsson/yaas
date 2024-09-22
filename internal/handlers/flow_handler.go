package handlers

import (
	"encoding/json"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/services"
	"github.com/joakimcarlsson/yaas/internal/utils"
	"net/http"
	"time"
)

type FlowHandler struct {
	FlowService services.FlowService
	AuthService services.AuthService
}

func NewFlowHandler(flowService services.FlowService, authService services.AuthService) *FlowHandler {
	return &FlowHandler{
		FlowService: flowService,
		AuthService: authService,
	}
}

func (h *FlowHandler) InitiateLoginFlow(w http.ResponseWriter, r *http.Request) {
	requestURL := r.URL.String()
	flow, err := h.FlowService.InitiateFlow(r.Context(), models.FlowTypeLogin, requestURL)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to initiate login flow")
		return
	}

	utils.JSONResponse(w, http.StatusOK, flow)
}

func (h *FlowHandler) InitiateRegistrationFlow(w http.ResponseWriter, r *http.Request) {
	requestURL := r.URL.String()
	flow, err := h.FlowService.InitiateFlow(r.Context(), models.FlowTypeRegistration, requestURL)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to initiate registration flow")
		return
	}

	flow.State = models.FlowStateEnterDetails
	err = h.FlowService.UpdateFlow(r.Context(), flow)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update flow")
		return
	}

	utils.JSONResponse(w, http.StatusOK, flow)
}

func (h *FlowHandler) ProceedRegistrationFlow(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Flow ID is required")
		return
	}

	flow, err := h.FlowService.GetFlowByID(r.Context(), flowID)
	if err != nil || flow == nil {
		utils.JSONError(w, http.StatusNotFound, "Flow not found")
		return
	}

	if time.Now().After(flow.ExpiresAt) {
		utils.JSONError(w, http.StatusGone, "Flow has expired")
		return
	}

	switch flow.State {
	case models.FlowStateEnterDetails:
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
			flow.Errors = append(flow.Errors, models.FlowError{
				Field:   "email",
				Message: err.Error(),
			})
			h.FlowService.UpdateFlow(r.Context(), flow)
			utils.JSONResponse(w, http.StatusOK, flow)
			return
		}

		flow.State = models.FlowStateSuccess
		h.FlowService.UpdateFlow(r.Context(), flow)

		//todo

		// Optionally, generate tokens upon successful registration
		//accessToken, refreshToken, err := h.AuthService.GenerateTokens(r.Context(), user)
		//if err != nil {
		//	utils.JSONError(w, http.StatusInternalServerError, "Failed to generate tokens")
		//	return
		//}

		response := map[string]interface{}{
			"flow": flow,
			//"accessToken":  accessToken,
			//"refreshToken": refreshToken,
			//"user":         user,
		}
		utils.JSONResponse(w, http.StatusOK, response)

	default:
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow state")
	}
}

func (h *FlowHandler) ProceedLoginFlow(w http.ResponseWriter, r *http.Request) {
	flowID := r.URL.Query().Get("flow")
	if flowID == "" {
		utils.JSONError(w, http.StatusBadRequest, "Flow ID is required")
		return
	}

	flow, err := h.FlowService.GetFlowByID(r.Context(), flowID)
	if err != nil || flow == nil {
		utils.JSONError(w, http.StatusNotFound, "Flow not found")
		return
	}

	if time.Now().After(flow.ExpiresAt) {
		utils.JSONError(w, http.StatusGone, "Flow has expired")
		return
	}

	switch flow.State {
	case models.FlowStateChooseMethod:
		// Handle method selection (e.g., password, OAuth)
		// For simplicity, we'll assume password method here
		flow.State = models.FlowStateEnterCredentials
		err = h.FlowService.UpdateFlow(r.Context(), flow)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Failed to update flow")
			return
		}
		utils.JSONResponse(w, http.StatusOK, flow)

	case models.FlowStateEnterCredentials:
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		user, accessToken, refreshToken, err := h.AuthService.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			flow.Errors = append(flow.Errors, models.FlowError{
				Field:   "email or password",
				Message: "Invalid credentials",
			})
			flow.State = models.FlowStateEnterCredentials
			h.FlowService.UpdateFlow(r.Context(), flow)
			utils.JSONResponse(w, http.StatusOK, flow)
			return
		}

		flow.State = models.FlowStateSuccess
		h.FlowService.UpdateFlow(r.Context(), flow)

		// Return tokens to the client
		response := map[string]interface{}{
			"flow":         flow,
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"user":         user,
		}
		utils.JSONResponse(w, http.StatusOK, response)

	default:
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow state")
	}
}
