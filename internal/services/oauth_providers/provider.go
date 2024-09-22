package oauth_providers

import "net/http"

type OAuthProvider interface {
	GetProviderID(userInfo map[string]interface{}) (string, error)
	GetEmail(userInfo map[string]interface{}, client *http.Client) (string, error)
}
