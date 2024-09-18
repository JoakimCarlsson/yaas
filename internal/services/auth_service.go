package services

import (
	"context"
	"errors"
	"time"

	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"github.com/joakimcarlsson/yaas/internal/utils"
)

type AuthService interface {
	Register(ctx context.Context, user *models.User, password string) error
	Login(ctx context.Context, email, password string) (*models.User, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
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

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	match, err := utils.ComparePasswordAndHash(password, user.Password)
	if err != nil || !match {
		return nil, errors.New("invalid email or password")
	}

	now := time.Now()
	user.LastLogin = &now
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
