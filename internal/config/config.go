package config

import (
	"fmt"
	"os"
	"time"
)

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
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
}

func Load() *Config {
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
		GoogleClientID:        os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:    os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:     os.Getenv("GOOGLE_REDIRECT_URL"),
	}
}
