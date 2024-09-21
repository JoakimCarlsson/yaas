package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joakimcarlsson/yaas/internal/config"
	"github.com/joakimcarlsson/yaas/internal/models"
)

type JWTService interface {
	GenerateAccessToken(user *models.User) (string, error)
	GenerateRefreshToken(user *models.User) (string, uuid.UUID, time.Time, error)
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	ValidateRefreshToken(tokenString string) (*jwt.Token, error)
	GetJTIFromToken(token *jwt.Token) (uuid.UUID, error)
}

type jwtService struct {
	config *config.Config
}

func NewJWTService(cfg *config.Config) JWTService {
	return &jwtService{
		config: cfg,
	}
}

func (s *jwtService) GenerateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"iss": s.config.BaseURL,
		"aud": s.config.BaseURL,
		"exp": time.Now().Add(s.config.JWTAccessTokenExpiry).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTAccessSecret))
}

func (s *jwtService) GenerateRefreshToken(user *models.User) (string, uuid.UUID, time.Time, error) {
	jti := uuid.New()
	expiresAt := time.Now().Add(s.config.JWTRefreshTokenExpiry)
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
		"jti": jti.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTRefreshSecret))
	if err != nil {
		return "", uuid.Nil, time.Time{}, err
	}

	return tokenString, jti, expiresAt, nil
}

func (s *jwtService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTAccessSecret), nil
	})
}

func (s *jwtService) ValidateRefreshToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTRefreshSecret), nil
	})
}

func (s *jwtService) GetJTIFromToken(token *jwt.Token) (uuid.UUID, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}

	jtiStr, ok := claims["jti"].(string)
	if !ok {
		return uuid.Nil, errors.New("jti claim not found or invalid")
	}

	jti, err := uuid.Parse(jtiStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid jti format")
	}

	return jti, nil
}
