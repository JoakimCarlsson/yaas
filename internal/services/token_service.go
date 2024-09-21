package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/joakimcarlsson/yaas/internal/config"
)

type TokenService interface {
	GenerateEmailVerificationToken(userID string) (string, error)
	VerifyEmailVerificationToken(token string) (string, error)
}

type tokenService struct {
	config *config.Config
}

func NewTokenService(cfg *config.Config) TokenService {
	return &tokenService{config: cfg}
}

func (s *tokenService) GenerateEmailVerificationToken(userID string) (string, error) {
	now := time.Now().Unix()
	payload := fmt.Sprintf("%s:%d", userID, now)

	h := hmac.New(sha256.New, []byte(s.config.EmailVerificationSecret))
	h.Write([]byte(payload))
	signature := h.Sum(nil)

	token := fmt.Sprintf("%s.%s",
		base64.URLEncoding.EncodeToString([]byte(payload)),
		base64.URLEncoding.EncodeToString(signature))

	return token, nil
}

func (s *tokenService) VerifyEmailVerificationToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token format")
	}

	payload, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid payload encoding")
	}

	signature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid signature encoding")
	}

	h := hmac.New(sha256.New, []byte(s.config.EmailVerificationSecret))
	h.Write(payload)
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return "", fmt.Errorf("invalid token signature")
	}

	payloadParts := strings.Split(string(payload), ":")
	if len(payloadParts) != 2 {
		return "", fmt.Errorf("invalid payload format")
	}

	userID := payloadParts[0]
	timestamp, err := time.Parse(time.RFC3339, payloadParts[1])
	if err != nil {
		return "", fmt.Errorf("invalid timestamp in payload")
	}

	if time.Since(timestamp) > s.config.EmailVerificationTokenExpiry {
		return "", fmt.Errorf("token has expired")
	}

	return userID, nil
}
