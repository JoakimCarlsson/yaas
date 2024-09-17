package server

import (
	"database/sql"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/config"
)

type Server struct {
	cfg    *config.Config
	router *http.ServeMux
	db     *sql.DB
}

func NewServer(cfg *config.Config, db *sql.DB) *Server {
	s := &Server{
		cfg:    cfg,
		router: http.NewServeMux(),
		db:     db,
	}

	// s.routes()

	return s
}

// func (s *Server) routes() {
// 	userRepo := repository.NewPostgresUserRepository(s.db)
// 	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(s.db)

// 	authService := auth.NewAuthService(
// 		refreshTokenRepo,
// 		s.cfg.JWTAccessSecret,
// 		s.cfg.JWTRefreshSecret,
// 		time.Minute*15,
// 		time.Hour*24*7,
// 		s.cfg.BaseURL,
// 		s.cfg.BaseURL,
// 	)

// 	userHandler := handlers.NewUserHandler(userRepo, authService)

// 	s.router.HandleFunc("/login", userHandler.Login)
// 	s.router.HandleFunc("/refresh-token", userHandler.RefreshToken)
// 	s.router.HandleFunc("/register", userHandler.Register)
// }

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.cfg.ServerPort, s.router)
}
