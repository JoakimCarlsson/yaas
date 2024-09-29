package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/joakimcarlsson/yaas/internal/executor"
	"github.com/joakimcarlsson/yaas/internal/services/oauth_providers"
	"golang.org/x/oauth2"
	"log"
	"reflect"
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
	ExecuteActions(ctx context.Context, actionType string, data map[string]interface{}) (*executor.ActionResult, error)
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

func (s *authService) ExecuteActions(ctx context.Context, actionType string, data map[string]interface{}) (*executor.ActionResult, error) {
	ac := &executor.ActionContext{
		Connection:  data["connection"].(string),
		User:        map[string]interface{}{},
		RequestInfo: data["request_info"].(map[string]interface{}),
	}

	if user, ok := data["user"].(map[string]interface{}); ok {
		ac.User = user
	}

	result, err := s.actionExecutor.ExecuteActions(ctx, actionType, ac)
	if err != nil {
		return nil, err
	}

	return result, nil
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

	preRegisterResult, err := s.ExecuteActions(ctx, "pre-register", preRegisterData)
	if err != nil {
		return err
	}
	if !preRegisterResult.Allow {
		return errors.New(preRegisterResult.Message)
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

	// Apply any changes from pre-register actions
	if preRegisterResult.User != nil {
		updatedUser, err := s.updateUserFromMap(user, preRegisterResult.User)
		if err != nil {
			return err
		}
		user = updatedUser
	}

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

	postRegisterResult, err := s.ExecuteActions(ctx, "post-register", postRegisterData)
	if err != nil {
		log.Printf("Post-register action error: %v", err)
	} else if !postRegisterResult.Allow {
		log.Printf("Post-register action denied: %s", postRegisterResult.Message)
	}

	// Apply any changes from post-register actions
	if postRegisterResult != nil && postRegisterResult.User != nil {
		updatedUser, err := s.updateUserFromMap(user, postRegisterResult.User)
		if err != nil {
			log.Printf("Failed to apply post-register user updates: %v", err)
		} else {
			if err := s.userRepo.UpdateUser(ctx, updatedUser); err != nil {
				log.Printf("Failed to save post-register user updates: %v", err)
			}
		}
	}

	return nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*models.User, string, string, error) {
	preLoginData := &executor.ActionContext{
		Connection: "password",
		User:       map[string]interface{}{},
		RequestInfo: map[string]interface{}{
			"email": email,
		},
	}

	preLoginResult, err := s.actionExecutor.ExecuteActions(ctx, "pre-login", preLoginData)

	if err != nil {
		return nil, "", "", err
	}
	if !preLoginResult.Allow {
		return nil, "", "", errors.New(preLoginResult.Message)
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	match, err := utils.ComparePasswordAndHash(password, *user.Password)
	if err != nil || !match {
		return nil, "", "", errors.New("invalid email or password")
	}

	postLoginData := &executor.ActionContext{
		Connection: "password",
		User: map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		RequestInfo: map[string]interface{}{
			"email": email,
		},
	}

	postLoginResult, err := s.actionExecutor.ExecuteActions(ctx, "post-login", postLoginData)
	if err != nil {
		return nil, "", "", err
	}
	if !postLoginResult.Allow {
		return nil, "", "", errors.New(postLoginResult.Message)
	}

	if !reflect.DeepEqual(postLoginResult.User, postLoginData.User) {
		updatedUser, err := s.updateUserFromMap(user, postLoginResult.User)
		if err != nil {
			return nil, "", "", err
		}

		if err := s.userRepo.UpdateUser(ctx, updatedUser); err != nil {
			return nil, "", "", err
		}
		user = updatedUser
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

func (s *authService) updateUserFromMap(user *models.User, updates map[string]interface{}) (*models.User, error) {
	updatedUser := *user

	val := reflect.ValueOf(&updatedUser).Elem()

	for key, value := range updates {
		field := val.FieldByName(key)
		if field.IsValid() && field.CanSet() {
			switch field.Kind() {
			case reflect.String:
				if strValue, ok := value.(string); ok {
					field.SetString(strValue)
				}
			case reflect.Bool:
				if boolValue, ok := value.(bool); ok {
					field.SetBool(boolValue)
				}
			case reflect.Int, reflect.Int64:
				if intValue, ok := value.(float64); ok {
					field.SetInt(int64(intValue))
				}
			}
		}
	}

	updatedUser.ID = user.ID
	updatedUser.Email = user.Email
	updatedUser.Password = user.Password

	return &updatedUser, nil
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

	fmt.Println(email)
	preLoginData := &executor.ActionContext{
		Connection: provider,
		User:       map[string]interface{}{},
		RequestInfo: map[string]interface{}{
			"email": email,
			"ip":    "",
		},
	}

	preLoginResult, err := s.actionExecutor.ExecuteActions(ctx, "pre-login", preLoginData)
	if err != nil {
		return nil, "", "", err
	}
	if !preLoginResult.Allow {
		return nil, "", "", errors.New(preLoginResult.Message)
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

	// Apply any changes from pre-login actions
	if preLoginResult.User != nil {
		updatedUser, err := s.updateUserFromMap(user, preLoginResult.User)
		if err != nil {
			return nil, "", "", err
		}
		user = updatedUser
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

	postLoginData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
		"connection":   provider,
		"request_info": userInfo,
	}
	postLoginResult, err := s.ExecuteActions(ctx, "post-login", postLoginData)
	if err != nil {
		log.Printf("Post-login action error: %v", err)
	} else if !postLoginResult.Allow {
		log.Printf("Post-login action denied: %s", postLoginResult.Message)
	}

	if postLoginResult != nil && postLoginResult.User != nil {
		updatedUser, err := s.updateUserFromMap(user, postLoginResult.User)
		if err != nil {
			log.Printf("Failed to apply post-login user updates: %v", err)
		} else {
			user = updatedUser
			if err := s.userRepo.UpdateUser(ctx, user); err != nil {
				log.Printf("Failed to save post-login user updates: %v", err)
			}
		}
	}

	return user, accessToken, refreshToken, nil
}
