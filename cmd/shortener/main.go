package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/db"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"
	"uno/cmd/shortener/utils"

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

	var store storage.Storage
	s, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		log.Fatalf("failed to create file storage: %v", err)
	}
	store = s

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not initialize zap logger: %v", err)
	}
	defer logger.Sync()

	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.LoggingMiddleware(logger))

	r.Post("/", shortenURLHandler(cfg, store))
	r.Post("/api/shorten", apiShortenHandler(cfg, store))
	r.Get("/{id}", redirectHandler(store))

	if conn != nil {
		r.Get("/ping", pingHandler(conn))
	}

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Println("Starting server on", cfg.Address)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func shortenURLHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		if originalURL == "" {
			http.Error(w, "empty URL", http.StatusBadRequest)
			return
		}

		shortID := utils.GenerateShortID()
		store.Save(shortID, originalURL)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, cfg.BaseURL+"/"+shortID)
	}
}

func redirectHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, "id")
		originalURL, ok := store.Get(shortID)
		if !ok {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func apiShortenHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		var req models.APIRequest
		if err := req.UnmarshalJSON(data); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		originalURL := strings.TrimSpace(req.URL)
		if originalURL == "" {
			http.Error(w, "empty URL", http.StatusBadRequest)
			return
		}

		shortID := utils.GenerateShortID()
		store.Save(shortID, originalURL)

		resp := models.APIResponse{
			Result: cfg.BaseURL + "/" + shortID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if data, err := resp.MarshalJSON(); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		} else {
			w.Write(data)
		}
	}
}

func pingHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		if err := conn.Ping(ctx); err != nil {
			http.Error(w, "database connection failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
