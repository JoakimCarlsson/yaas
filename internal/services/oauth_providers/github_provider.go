package oauth_providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type GitHubProvider struct{}

func (g *GitHubProvider) GetProviderID(userInfo map[string]interface{}) (string, error) {
	if id, ok := userInfo["id"].(float64); ok {
		return fmt.Sprintf("%.0f", id), nil
	}
	return "", errors.New("failed to get provider ID from user info (GitHub)")
}

func (g *GitHubProvider) GetEmail(userInfo map[string]interface{}, client *http.Client) (string, error) {
	if email, ok := userInfo["email"].(string); ok && email != "" {
		return email, nil
	}

	// If no email is provided, fetch it using the GitHub emails API
	response, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch GitHub user emails")
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(response.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", errors.New("no primary, verified email found for GitHub user")
}
