package server

import (
	"database/sql"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/config"
	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/logger"
	"github.com/joakimcarlsson/yaas/internal/middleware"
	"github.com/joakimcarlsson/yaas/internal/repository/postgres"
	"github.com/joakimcarlsson/yaas/internal/services"
)

type Server struct {
	cfg    *config.Config
	router http.Handler
	db     *sql.DB
}

func NewServer(cfg *config.Config, db *sql.DB) *Server {
	logger.SetupLogger()

	s := &Server{
		cfg:    cfg,
		router: http.NewServeMux(),
		db:     db,
	}

	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)

	jwtService := services.NewJWTService(cfg)
	oauthService := services.NewOAuth2Service(cfg)
	emailService := services.NewEmailService(cfg)
	tokenService := services.NewTokenService(cfg)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtService, oauthService, emailService, tokenService)
	authHandler := handlers.NewAuthHandler(authService, oauthService)

	s.router = NewRouter(authHandler)

	s.router = middleware.AuditLogMiddleware(s.router)
	s.router = middleware.SecurityHeadersMiddleware(s.router)

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.cfg.ServerPort, s.router)
}
