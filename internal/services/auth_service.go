package services

import (
	"context"
	"errors"
	"github.com/joakimcarlsson/yaas/internal/executor"
	"github.com/joakimcarlsson/yaas/internal/services/oauth_providers"
	"golang.org/x/oauth2"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"github.com/joakimcarlsson/yaas/internal/utils"
)

var stateSecret = []byte("your-secure-state-secret") //todo read from config, or use a secure random generator

type StateClaims struct {
	CallbackURL string `json:"callback_url"`
	jwt.RegisteredClaims
}

type AuthService interface {
	Register(ctx context.Context, user *models.User, password string) error
	Login(ctx context.Context, email, password string) (*models.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	GenerateStateToken(callbackURL string) (string, error)
	ValidateStateToken(token string) (string, error)
	ProcessOAuthLogin(ctx context.Context, provider string, userInfo map[string]interface{}, token *oauth2.Token) (*models.User, string, string, error)
	ExecuteActions(ctx context.Context, actionType string, data map[string]interface{}) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       JWTService
	oauth2Service    OAuth2Service
	actionExecutor   *executor.ActionExecutor
}

func NewAuthService(userRepo repository.UserRepository, refreshRepo repository.RefreshTokenRepository, jwtService JWTService, oauth2Service OAuth2Service, actionExecutor *executor.ActionExecutor) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshRepo,
		jwtService:       jwtService,
		oauth2Service:    oauth2Service,
		actionExecutor:   actionExecutor,
	}
}

func (s *authService) ExecuteActions(ctx context.Context, actionType string, data map[string]interface{}) error {
	ac := &executor.ActionContext{
		Connection:  data["connection"].(string),
		User:        map[string]interface{}{},
		RequestInfo: data["request_info"].(map[string]interface{}),
	}

	if user, ok := data["user"].(map[string]interface{}); ok {
		ac.User = user
	}

	return s.actionExecutor.ExecuteActions(ctx, actionType, ac)
}

func (s *authService) Register(ctx context.Context, user *models.User, password string) error {
	preRegisterData := map[string]interface{}{
		"user": map[string]interface{}{
			"email": user.Email,
		},
		"connection": "password",
		"request_info": map[string]interface{}{
			"email": user.Email,
		},
	}
	if err := s.ExecuteActions(ctx, "pre-register", preRegisterData); err != nil {
		return err
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := utils.GenerateFromPassword(password, utils.DefaultParams)
	if err != nil {
		return err
	}
	user.Password = &hashedPassword
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Provider = "password"
	user.ProviderID = nil

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	postRegisterData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"connection": "password",
		"request_info": map[string]interface{}{
			"email": user.Email,
		},
	}
	if err := s.ExecuteActions(ctx, "post-register", postRegisterData); err != nil {
		log.Printf("Post-register action error: %v", err)
	}

	return nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	preLoginData := map[string]interface{}{
		"connection": "password",
		"request_info": map[string]interface{}{
			"email": email,
		},
	}
	if err := s.ExecuteActions(ctx, "pre-login", preLoginData); err != nil {
		return nil, "", "", err
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	match, err := utils.ComparePasswordAndHash(password, *user.Password)
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

	postLoginData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"connection": "password",
		"request_info": map[string]interface{}{
			"email": email,
		},
	}
	if err := s.ExecuteActions(ctx, "post-login", postLoginData); err != nil {
		log.Printf("Post-login action error: %v", err)
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

func (s *authService) GenerateStateToken(callbackURL string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &StateClaims{
		CallbackURL: callbackURL,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(stateSecret)
}

func (s *authService) ValidateStateToken(tokenStr string) (string, error) {
	claims := &StateClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return stateSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid or expired state token")
	}

	return claims.CallbackURL, nil
}

func (s *authService) ProcessOAuthLogin(ctx context.Context, provider string, userInfo map[string]interface{}, token *oauth2.Token) (*models.User, string, string, error) {
	preLoginData := map[string]interface{}{
		"connection":   provider,
		"request_info": userInfo,
	}
	if err := s.ExecuteActions(ctx, "pre-login", preLoginData); err != nil {
		return nil, "", "", err
	}

	providerFactory := oauth_providers.OAuthProviderFactory{}
	providerStrategy, err := providerFactory.GetProvider(provider)
	if err != nil {
		return nil, "", "", err
	}

	providerID, err := providerStrategy.GetProviderID(userInfo)
	if err != nil {
		return nil, "", "", err
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	email, err := providerStrategy.GetEmail(userInfo, client)
	if err != nil {
		return nil, "", "", err
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		user = &models.User{
			Email:      email,
			IsActive:   true,
			IsVerified: true,
			Provider:   provider,
			ProviderID: &providerID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := s.userRepo.CreateUser(ctx, user); err != nil {
			return nil, "", "", err
		}
	} else {
		if user.Provider != provider || user.ProviderID == nil || *user.ProviderID != providerID {
			return nil, "", "", errors.New("email already in use with a different provider")
		}

		now := time.Now()
		user.LastLogin = &now
		user.UpdatedAt = now
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return nil, "", "", err
		}
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

	// Execute post-login actions
	postLoginData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"connection":   provider,
		"request_info": userInfo,
	}
	if err := s.ExecuteActions(ctx, "post-login", postLoginData); err != nil {
		// Consider how to handle post-login action errors
		// For now, we'll log the error but still return the tokens
		log.Printf("Post-login action error: %v", err)
	}

	return user, accessToken, refreshToken, nil
}
