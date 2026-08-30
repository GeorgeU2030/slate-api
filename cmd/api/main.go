package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"

	"slate-api/internal/auth"
	"slate-api/internal/config"
	"slate-api/internal/database"
	"slate-api/internal/handlers"
	"slate-api/internal/services"
	"slate-api/internal/tmdb"
)

func main() {
	migrate := flag.Bool("migrate", false, "run migrations and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	if *migrate {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("running migrations: %v", err)
		}
		log.Println("migrations complete")
		return
	}

	if !cfg.IsProduction() {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("running migrations: %v", err)
		}
	}

	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTAccessTokenExpireMinutes, cfg.RefreshTokenExpireDays)
	googleAuthService := auth.NewGoogleAuthService(cfg.GoogleClientID, cfg.GoogleClientSecret)
	tmdbService := tmdb.NewService(cfg.TMDBAPIKey)

	authService := services.NewAuthService(db, tokenService)
	titleService := services.NewTitleService(db, tmdbService)

	authHandler := handlers.NewAuthHandler(authService, cfg)
	googleHandler := handlers.NewGoogleHandler(authService, googleAuthService, cfg)
	titleHandler := handlers.NewTitleHandler(titleService)

	router := handlers.NewRouter(handlers.RouterDeps{
		Cfg:           cfg,
		TokenService:  tokenService,
		AuthHandler:   authHandler,
		GoogleHandler: googleHandler,
		TitleHandler:  titleHandler,
	})

	registerRootRoutes(router, db, tokenService)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Slate API listening on :%s (env=%s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}

	tmdbService.Close()

	log.Println("shutdown complete")
}