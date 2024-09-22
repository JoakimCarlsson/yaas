package oauth_providers

import (
	"errors"
	"fmt"
	"net/http"
)

type GoogleProvider struct{}

func (g *GoogleProvider) GetProviderID(userInfo map[string]interface{}) (string, error) {
	fmt.Println(userInfo)
	if id, ok := userInfo["id"].(string); ok {
		return id, nil
	}
	return "", errors.New("failed to get provider ID from user info (Google)")
}

func (g *GoogleProvider) GetEmail(userInfo map[string]interface{}, _ *http.Client) (string, error) {
	if email, ok := userInfo["email"].(string); ok && email != "" {
		return email, nil
	}
	return "", errors.New("email not found in user info")
}
