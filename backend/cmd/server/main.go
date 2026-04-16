package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	httpadapter "acc-dp/backend/internal/adapter/http"
	"acc-dp/backend/internal/config"
	"acc-dp/backend/internal/database"
	"acc-dp/backend/internal/repository"
	"acc-dp/backend/internal/service/auth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(ctx, db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	authRepository := repository.NewAuthRepository(db)
	authService, err := auth.NewService(
		authRepository,
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.SessionTTL,
	)
	if err != nil {
		log.Fatalf("create auth service: %v", err)
	}

	authHandler := httpadapter.NewAuthHandler(authService)
	router := httpadapter.NewRouter(authHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening at %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve http: %v", err)
	}

	fmt.Println("server stopped")
}
