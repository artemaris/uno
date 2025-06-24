package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/db"
	"uno/cmd/shortener/handlers"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	var conn *pgx.Conn
	if cfg.DatabaseDSN != "" {
		var err error
		conn, err = db.NewPG(cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("DB connection failed: %v", err)
		}
		defer conn.Close(context.Background())
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not initialize zap logger: %v", err)
	}
	defer logger.Sync()

	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.LoggingMiddleware(logger))

	var store storage.Storage
	if conn != nil {
		store, err = storage.NewPostgresStorage(conn)
		if err != nil {
			log.Fatalf("failed to initialize PostgreSQL storage: %v", err)
		}
	} else {
		if cfg.FileStoragePath != "" {
			s, err := storage.NewFileStorage(cfg.FileStoragePath)
			if err == nil {
				store = s
			}
		}
		if store == nil {
			store = storage.NewInMemoryStorage()
		}
	}

	r.Use(middleware.WithUserIDMiddleware)
	r.Post("/", handlers.ShortenURLHandler(cfg, store))
	r.Post("/api/shorten", handlers.APIShortenHandler(cfg, store))
	r.Post("/api/shorten/batch", handlers.BatchShortenHandler(cfg, store))
	r.Get("/{id}", handlers.RedirectHandler(store))
	r.Get("/ping", handlers.PingHandler(conn))
	r.Get("/api/user/urls", handlers.UserURLsHandler(cfg, store))

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Println("Starting server on", cfg.Address)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
