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

func (h *FlowHandler) InitiateFlow(w http.ResponseWriter, r *http.Request, flowType models.FlowType) {
	var flow *models.Flow
	var err error

	switch flowType {
	case models.FlowTypeLogin:
		flow, err = h.FlowService.InitiateFlow(r.Context(), models.FlowTypeLogin, r.URL.String())
		if err == nil {
			// Add OAuth providers to the login flow
			//oauthProviders, providerErr := h.OAuth2Service.GetAvailableProviders()
			//if providerErr != nil {
			//	utils.JSONError(w, http.StatusInternalServerError, "Failed to get OAuth providers")
			//	return
			//}
			//flow.OAuthProviders = oauthProviders
		}
	case models.FlowTypeRegistration:
		flow, err = h.FlowService.InitiateFlow(r.Context(), models.FlowTypeRegistration, r.URL.String())
	case models.FlowTypeLogout:
		flow, err = h.FlowService.InitiateFlow(r.Context(), models.FlowTypeLogout, r.URL.String())
	default:
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow type")
		return
	}

	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to initiate flow")
		return
	}

	utils.JSONResponse(w, http.StatusOK, flow)
}

func (h *FlowHandler) ProceedFlow(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Method      string `json:"method"`
		Provider    string `json:"provider,omitempty"`
		CallbackURL string `json:"callback_url,omitempty"`
		Email       string `json:"email,omitempty"`
		Password    string `json:"password,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	switch flow.Type {
	case models.FlowTypeLogin:
		switch req.Method {
		case "password":
			h.handlePasswordLogin(w, r, flow, req.Email, req.Password)
		case "oauth":
			h.handleOAuthLogin(w, r, flow, req.Provider, req.CallbackURL)
		default:
			utils.JSONError(w, http.StatusBadRequest, "Invalid authentication method")
		}
	case models.FlowTypeRegistration:
		h.ProceedRegistrationFlow(w, r, flow, req.Email, req.Password)
	case models.FlowTypeLogout:
		h.ProceedLogoutFlow(w, r, flow) // Assuming password is used as refresh token
	default:
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow type")
	}
}

func (h *FlowHandler) handlePasswordLogin(w http.ResponseWriter, r *http.Request, flow *models.Flow, email, password string) {
	user, accessToken, refreshToken, err := h.AuthService.Login(r.Context(), email, password)
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

	response := map[string]interface{}{
		"flow":         flow,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         user,
	}
	utils.JSONResponse(w, http.StatusOK, response)
}

func (h *FlowHandler) handleOAuthLogin(w http.ResponseWriter, r *http.Request, flow *models.Flow, provider, callbackURL string) {
	if provider == "" {
		utils.JSONError(w, http.StatusBadRequest, "OAuth provider is required")
		return
	}
	if callbackURL == "" {
		utils.JSONError(w, http.StatusBadRequest, "Callback URL is required")
		return
	}

	// Validate the callback URL
	//if err := h.validateCallbackURL(callbackURL); err != nil {
	//	utils.JSONError(w, http.StatusBadRequest, "Invalid callback URL")
	//	return
	//}

	stateToken, err := h.AuthService.GenerateStateTokenWithFlowID(flow.ID, callbackURL)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to generate state token")
		return
	}

	loginURL, err := h.OAuth2Service.GetLoginURL(provider, stateToken)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to get OAuth login URL")
		return
	}

	flow.State = models.FlowStateRedirectToProvider
	flow.RequestURL = loginURL
	if err := h.FlowService.UpdateFlow(r.Context(), flow); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update flow")
		return
	}

	utils.JSONResponse(w, http.StatusOK, flow)
	//redirect to loginUrl
	//http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

func (h *FlowHandler) ProceedRegistrationFlow(w http.ResponseWriter, r *http.Request, flow *models.Flow, email, password string) {
	user := &models.User{
		Email:    email,
		Password: &password,
	}

	err := h.AuthService.Register(r.Context(), user, password)
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

	response := map[string]interface{}{
		"flow": flow,
		"user": user,
	}
	utils.JSONResponse(w, http.StatusOK, response)
}

func (h *FlowHandler) ProceedLogoutFlow(w http.ResponseWriter, r *http.Request, flow *models.Flow) {
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
}

func (h *FlowHandler) ProceedOAuthLoginFlow(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if provider == "" || state == "" || code == "" {
		utils.JSONError(w, http.StatusBadRequest, "Missing parameters")
		return
	}

	flowID, callbackURL, err := h.AuthService.ValidateStateToken(state)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid state token")
		return
	}

	flow, err := h.FlowService.GetFlowByID(r.Context(), flowID)
	if err != nil || flow == nil {
		utils.JSONError(w, http.StatusNotFound, "Flow not found")
		return
	}

	if flow.State != models.FlowStateRedirectToProvider && flow.State != models.FlowStateAwaitingCallback {
		utils.JSONError(w, http.StatusBadRequest, "Invalid flow state")
		return
	}

	flow.State = models.FlowStateProcessingCallback
	h.FlowService.UpdateFlow(r.Context(), flow)

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

	flow.State = models.FlowStateSuccess
	h.FlowService.UpdateFlow(r.Context(), flow)

	redirectURL := buildRedirectURL(callbackURL, accessToken, refreshToken)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func buildRedirectURL(baseURL, accessToken, refreshToken string) string {
	return baseURL + "?accessToken=" + accessToken + "&refreshToken=" + refreshToken
}
