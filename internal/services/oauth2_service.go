package services

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuth2Service interface {
	GetGoogleLoginURL(state string) string
	ExchangeCodeForToken(code string) (*oauth2.Token, error)
	GetGoogleUserInfo(token *oauth2.Token) (map[string]interface{}, error)
}

type oauth2Service struct {
	config      *config.Config
	oauthConfig *oauth2.Config
}

func NewOAuth2Service(cfg *config.Config) OAuth2Service {
	return &oauth2Service{
		config: cfg,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (s *oauth2Service) GetGoogleLoginURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *oauth2Service) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(context.Background(), code)
}

func (s *oauth2Service) GetGoogleUserInfo(token *oauth2.Token) (map[string]interface{}, error) {
	client := s.oauthConfig.Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get user info from Google")
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(contents, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
