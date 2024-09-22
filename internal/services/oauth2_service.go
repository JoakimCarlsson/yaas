package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/config"
	"golang.org/x/oauth2"
)

type OAuth2Service interface {
	GetLoginURL(provider, state string) (string, error)
	ExchangeCodeForToken(provider, code string) (*oauth2.Token, error)
	GetUserInfo(provider string, token *oauth2.Token) (map[string]interface{}, error)
}

type oauth2Service struct {
	config *config.Config
}

func NewOAuth2Service(cfg *config.Config) OAuth2Service {
	return &oauth2Service{
		config: cfg,
	}
}

func (s *oauth2Service) GetLoginURL(provider, state string) (string, error) {
	providerConfig, ok := s.config.OAuthProviders[provider]
	if !ok {
		return "", errors.New("unsupported provider: " + provider)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		RedirectURL:  providerConfig.RedirectURL,
		Scopes:       providerConfig.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  providerConfig.AuthURL,
			TokenURL: providerConfig.TokenURL,
		},
	}

	fmt.Println(oauthConfig.RedirectURL)

	return oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (s *oauth2Service) ExchangeCodeForToken(provider, code string) (*oauth2.Token, error) {
	providerConfig, ok := s.config.OAuthProviders[provider]
	if !ok {
		return nil, errors.New("unsupported provider: " + provider)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     providerConfig.ClientID,
		ClientSecret: providerConfig.ClientSecret,
		RedirectURL:  providerConfig.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  providerConfig.AuthURL,
			TokenURL: providerConfig.TokenURL,
		},
	}

	return oauthConfig.Exchange(context.Background(), code)
}

func (s *oauth2Service) GetUserInfo(provider string, token *oauth2.Token) (map[string]interface{}, error) {
	providerConfig, ok := s.config.OAuthProviders[provider]
	if !ok {
		return nil, errors.New("unsupported provider: " + provider)
	}

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

	response, err := client.Get(providerConfig.UserInfoURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get user info from " + provider)
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(contents, &userInfo); err != nil {
		return nil, err
	}

	if provider == "github" {
		if email, ok := userInfo["email"].(string); !ok || email == "" {
			email, err := s.getGitHubUserEmail(client)
			if err != nil {
				return nil, err
			}
			userInfo["email"] = email
		}
	}

	return userInfo, nil
}

func (s *oauth2Service) getGitHubUserEmail(client *http.Client) (string, error) {
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
