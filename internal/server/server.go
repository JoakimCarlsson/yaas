package server

import (
	"database/sql"
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/config"
	"github.com/joakimcarlsson/yaas/internal/handlers"
	"github.com/joakimcarlsson/yaas/internal/repository/postgres"
	"github.com/joakimcarlsson/yaas/internal/services"
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

	userRepo := postgres.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	s.router = NewRouter(authHandler)

	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.cfg.ServerPort, s.router)
}
