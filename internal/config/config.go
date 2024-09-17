package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	ServerPort       string
	BaseURL          string
}

func Load() *Config { //todo ideally we want to read this from .json or .yaml file
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("SERVER_PORT"))
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		BaseURL:     baseURL,
	}
}
