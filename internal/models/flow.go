package models

import "time"

type FlowType string
type FlowState string

const (
	FlowTypeLogin               FlowType  = "login"
	FlowTypeLogout              FlowType  = "logout"
	FlowTypeRegistration        FlowType  = "registration"
	FlowStateEnterDetails       FlowState = "enter_details"
	FlowStateVerifyEmail        FlowState = "verify_email"
	FlowTypeOAuth2Login         FlowType  = "oauth2_login"
	FlowStateRedirectToProvider FlowState = "redirect_to_provider"
	FlowStateAwaitingCallback   FlowState = "awaiting_callback"
	FlowStateProcessingCallback FlowState = "processing_callback"
	FlowStateChooseMethod       FlowState = "choose_method"
	FlowStateEnterCredentials   FlowState = "enter_credentials"
	FlowStateSuccess            FlowState = "success"
	FlowStateFailed             FlowState = "failed"
	FlowStateInitiated          FlowState = "initiated"
	FlowStateConfirmLogout      FlowState = "confirm_logout"
	FlowStateLogoutComplete     FlowState = "logout_complete"
)

type FlowError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Flow struct {
	ID         string      `json:"id"`
	Type       FlowType    `json:"type"`
	State      FlowState   `json:"state"`
	ExpiresAt  time.Time   `json:"expires_at"`
	IssuedAt   time.Time   `json:"issued_at"`
	RequestURL string      `json:"request_url"`
	Errors     []FlowError `json:"errors,omitempty"`
}
