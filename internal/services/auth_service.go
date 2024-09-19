package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"github.com/joakimcarlsson/yaas/internal/utils"
)

type AuthService interface {
	Register(ctx context.Context, user *models.User, password string) error
	Login(ctx context.Context, email, password string) (*models.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       JWTService
}

func NewAuthService(userRepo repository.UserRepository, refreshRepo repository.RefreshTokenRepository, jwtService JWTService) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshRepo,
		jwtService:       jwtService,
	}
}

func (s *authService) Register(ctx context.Context, user *models.User, password string) error {
	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := utils.GenerateFromPassword(password, utils.DefaultParams)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Provider = "password"

	return s.userRepo.CreateUser(ctx, user)
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	match, err := utils.ComparePasswordAndHash(password, user.Password)
	if err != nil || !match {
		return nil, "", "", errors.New("invalid email or password")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, jti, expiresAt, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", "", err
	}

	if err := s.refreshTokenRepo.Create(ctx, user.ID, jti, expiresAt); err != nil {
		return nil, "", "", err
	}

	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	token, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	jti, err := s.jwtService.GetJTIFromToken(token)
	if err != nil {
		return "", "", errors.New("invalid jti in token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", "", errors.New("invalid user ID in token")
	}

	storedToken, err := s.refreshTokenRepo.GetByJTI(ctx, jti)
	if err != nil || storedToken == nil {
		return "", "", errors.New("refresh token not found")
	}

	if err := s.refreshTokenRepo.Delete(ctx, jti); err != nil {
		return "", "", errors.New("failed to delete old refresh token")
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	newAccessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", errors.New("failed to generate access token")
	}

	newRefreshToken, newJTI, expiresAt, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}

	if err := s.refreshTokenRepo.Create(ctx, user.ID, newJTI, expiresAt); err != nil {
		return "", "", errors.New("failed to store refresh token")
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	token, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	jti, err := s.jwtService.GetJTIFromToken(token)
	if err != nil {
		return errors.New("invalid jti in token")
	}

	return s.refreshTokenRepo.Delete(ctx, jti)
}
