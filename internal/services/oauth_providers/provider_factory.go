package oauth_providers

import "errors"

type OAuthProviderFactory struct{}

func (f *OAuthProviderFactory) GetProvider(provider string) (OAuthProvider, error) {
	switch provider {
	case "github":
		return &GitHubProvider{}, nil
	case "google":
		return &GoogleProvider{}, nil
	default:
		return nil, errors.New("unsupported provider: " + provider)
	}
}
