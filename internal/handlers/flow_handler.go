package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/services"
	"github.com/joakimcarlsson/yaas/internal/utils"
	"net/http"
	"time"
)

type FlowHandler struct {
	FlowService   services.FlowService
	AuthService   services.AuthService
	OAuth2Service services.OAuth2Service
}

func NewFlowHandler(flowService services.FlowService, authService services.AuthService, oauth2Service services.OAuth2Service) *FlowHandler {
	return &FlowHandler{
		FlowService:   flowService,
		AuthService:   authService,
		OAuth2Service: oauth2Service,
	}
}

func (h *FlowHandler) ProceedOAuthLoginFlow(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if provider == "" || state == "" || code == "" {
		utils.JSONError(w, http.StatusBadRequest, "Missing parameters")
		return
	}

	// Validate state token and extract flow ID and callback URL
	flowID, callbackURL, err := h.AuthService.ValidateStateToken(state)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid state token")
		return
	}

	// Retrieve the flow
	flow, err := h.FlowService.GetFlowByID(r.Context(), flowID)
	if err != nil || flow == nil {
		utils.JSONError(w, http.StatusNotFound, "Flow not found")
		return
	}

	if flow.State != models.FlowStateRedirectToProvider && flow.State != models.FlowStateAwaitingCallback {
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow state")
		return
	}

	// Update flow state to 'processing_callback'
	flow.State = models.FlowStateProcessingCallback
	h.FlowService.UpdateFlow(r.Context(), flow)

	// Exchange code for token
	token, err := h.OAuth2Service.ExchangeCodeForToken(provider, code)
	if err != nil {
		flow.Errors = append(flow.Errors, models.FlowError{
			Field:   "code",
			Message: "Failed to exchange code for token",
		})
		h.FlowService.UpdateFlow(r.Context(), flow)
		utils.JSONResponse(w, http.StatusOK, flow)
		return
	}

	// Get user info
	userInfo, err := h.OAuth2Service.GetUserInfo(provider, token)
	if err != nil {
		flow.Errors = append(flow.Errors, models.FlowError{
			Field:   "user_info",
			Message: "Failed to get user info",
		})
		h.FlowService.UpdateFlow(r.Context(), flow)
		utils.JSONResponse(w, http.StatusOK, flow)
		return
	}

	// Process OAuth login
	_, accessToken, refreshToken, err := h.AuthService.ProcessOAuthLogin(r.Context(), provider, userInfo, token)
	if err != nil {
		flow.Errors = append(flow.Errors, models.FlowError{
			Field:   "oauth_login",
			Message: err.Error(),
		})
		h.FlowService.UpdateFlow(r.Context(), flow)
		utils.JSONResponse(w, http.StatusOK, flow)
		return
	}

	// Update flow state to 'success'
	flow.State = models.FlowStateSuccess
	h.FlowService.UpdateFlow(r.Context(), flow)

	// Redirect to the callback URL with tokens as query parameters
	//todo return accessToken & refreshToken as cookies
	redirectURL := fmt.Sprintf("%s?accessToken=%s&refreshToken=%s", callbackURL, accessToken, refreshToken)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *FlowHandler) InitiateOAuthLoginFlow(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	callbackURL := r.URL.Query().Get("callback_url")
	if provider == "" || callbackURL == "" {
		utils.JSONError(w, http.StatusBadRequest, "Provider and callback URL are required")
		return
	}

	requestURL := r.URL.String()
	flow, err := h.FlowService.InitiateFlow(r.Context(), models.FlowTypeOAuth2Login, requestURL)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to initiate OAuth login flow")
		return
	}

	stateToken, err := h.AuthService.GenerateStateTokenWithFlowID(flow.ID, callbackURL)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to generate state token")
		return
	}

	loginURL, err := h.OAuth2Service.GetLoginURL(provider, stateToken)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to get login URL")
		return
	}

	flow.State = models.FlowStateRedirectToProvider
	err = h.FlowService.UpdateFlow(r.Context(), flow)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update flow state")
		return
	}

	http.Redirect(w, r, loginURL, http.StatusFound)
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

func (h *FlowHandler) InitiateLogoutFlow(w http.ResponseWriter, r *http.Request) {
	requestURL := r.URL.String()
	flow, err := h.FlowService.InitiateFlow(r.Context(), models.FlowTypeLogout, requestURL)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to initiate logout flow")
		return
	}
	utils.JSONResponse(w, http.StatusOK, flow)
}

func (h *FlowHandler) ProceedLogoutFlow(w http.ResponseWriter, r *http.Request) {
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
	case models.FlowStateInitiated:
		flow.State = models.FlowStateConfirmLogout
		err = h.FlowService.UpdateFlow(r.Context(), flow)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Failed to update flow")
			return
		}
		utils.JSONResponse(w, http.StatusOK, flow)

	case models.FlowStateConfirmLogout:
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if err := h.AuthService.Logout(r.Context(), req.RefreshToken); err != nil {
			flow.Errors = append(flow.Errors, models.FlowError{
				Field:   "refreshToken",
				Message: "Failed to logout",
			})
			h.FlowService.UpdateFlow(r.Context(), flow)
			utils.JSONResponse(w, http.StatusOK, flow)
			return
		}

		flow.State = models.FlowStateLogoutComplete
		h.FlowService.UpdateFlow(r.Context(), flow)

		utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"flow":    flow,
			"message": "Successfully logged out",
		})

	default:
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow state")
	}
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
