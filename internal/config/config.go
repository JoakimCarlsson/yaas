package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL           string
	JWTAccessSecret       string
	JWTRefreshSecret      string
	JWTAccessTokenExpiry  time.Duration
	JWTRefreshTokenExpiry time.Duration
	ServerPort            string
	BaseURL               string
}

func Load() *Config { //ideally we want to load the configuration from a file, but for simplicity we will use environment variables
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

	return &Config{
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		JWTAccessSecret:       os.Getenv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:      os.Getenv("JWT_REFRESH_SECRET"),
		JWTAccessTokenExpiry:  accessTokenExpiry,
		JWTRefreshTokenExpiry: refreshTokenExpiry,
		ServerPort:            os.Getenv("SERVER_PORT"),
		BaseURL:               baseURL,
	}
}
