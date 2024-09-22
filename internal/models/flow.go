package models

import "time"

type FlowType string
type FlowState string

const (
	FlowTypeLogin        FlowType = "login"
	FlowTypeRegistration FlowType = "registration"

	FlowStateEnterDetails FlowState = "enter_details"
	FlowStateVerifyEmail  FlowState = "verify_email"

	FlowStateChooseMethod     FlowState = "choose_method"
	FlowStateEnterCredentials FlowState = "enter_credentials"
	FlowStateSuccess          FlowState = "success"
	FlowStateFailed           FlowState = "failed"
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
