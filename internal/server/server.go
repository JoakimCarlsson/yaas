package server

import (
	"database/sql"
	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/middleware"
	"github.com/joakimcarlsson/yaas/internal/services"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/config"
	"github.com/joakimcarlsson/yaas/internal/logger"
	"github.com/joakimcarlsson/yaas/internal/repository/postgres"
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
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtService, oauthService)

	authHandler := handlers.NewAuthHandler(authService, oauthService)
	oauthHandler := handlers.NewOAuthHandler(oauthService, authService)
	tokenHandler := handlers.NewTokenHandler(authService)

	routerWithMiddlewares := middleware.AuditLogMiddleware(NewRouter(authHandler, oauthHandler, tokenHandler))
	routerWithMiddlewares = middleware.SecurityHeadersMiddleware(routerWithMiddlewares)

	s.router = routerWithMiddlewares

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.cfg.ServerPort, s.router)
}
