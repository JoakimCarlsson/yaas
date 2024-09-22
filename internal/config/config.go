package config

import (
	"fmt"
	"os"
	"time"
)

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type Config struct {
	PostgresHost          string
	PostgresUser          string
	PostgresPort          string
	PostgresPassword      string
	PostgresDB            string
	DatabaseURL           string
	JWTAccessSecret       string
	JWTRefreshSecret      string
	JWTAccessTokenExpiry  time.Duration
	JWTRefreshTokenExpiry time.Duration
	ServerPort            string
	BaseURL               string
	OAuthProviders        map[string]OAuthProviderConfig
}

func Load() *Config {
	providers := map[string]OAuthProviderConfig{
		"google": {
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:    "https://oauth2.googleapis.com/token",
			UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		},
		"github": {
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
			Scopes:       []string{"user:email"},
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
		},
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("SERVER_PORT"))
	}

	accessTokenExpiryStr := os.Getenv("JWT_ACCESS_TOKEN_EXPIRY")
	if accessTokenExpiryStr == "" {
		accessTokenExpiryStr = "15m"
	}
	accessTokenExpiry, err := time.ParseDuration(accessTokenExpiryStr)
	if err != nil {
		accessTokenExpiry = 15 * time.Minute
	}

	refreshTokenExpiryStr := os.Getenv("JWT_REFRESH_TOKEN_EXPIRY")
	if refreshTokenExpiryStr == "" {
		refreshTokenExpiryStr = "7d"
	}
	refreshTokenExpiry, err := time.ParseDuration(refreshTokenExpiryStr)
	if err != nil {
		refreshTokenExpiry = 7 * 24 * time.Hour
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		postgresUser, postgresPassword, postgresHost, postgresPort, postgresDB)

	return &Config{
		PostgresHost:          postgresHost,
		PostgresUser:          postgresUser,
		PostgresPort:          postgresPort,
		PostgresPassword:      postgresPassword,
		PostgresDB:            postgresDB,
		DatabaseURL:           databaseURL,
		JWTAccessSecret:       os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:      os.Getenv("JWT_REFRESH_SECRET"),
		JWTAccessTokenExpiry:  accessTokenExpiry,
		JWTRefreshTokenExpiry: refreshTokenExpiry,
		ServerPort:            os.Getenv("SERVER_PORT"),
		BaseURL:               baseURL,
		OAuthProviders:        providers,
	}
}
