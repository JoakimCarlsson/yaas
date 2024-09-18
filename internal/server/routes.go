package server

import (
	"net/http"

	"github.com/joakimcarlsson/yaas/internal/handlers"
)

func NewRouter(authHandler *handlers.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	return mux
}
