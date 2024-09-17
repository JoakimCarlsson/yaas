package main

import (
	"log"

	"github.com/joakimcarlsson/yaas/internal/config"
	"github.com/joakimcarlsson/yaas/internal/server"
	"github.com/joakimcarlsson/yaas/pkg/persistence/sql"
)

func main() {
	cfg := config.Load()

	db, err := sql.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	srv := server.NewServer(cfg, db)
	log.Fatal(srv.Start())
}
