package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/handlers"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Глобальные переменные для информации о сборке
// Устанавливаются при сборке через флаги компилятора
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	// Выводим информацию о сборке
	printBuildInfo()

	cfg := config.NewConfig()

	var pool *pgxpool.Pool
	if cfg.DatabaseDSN != "" {
		var err error
		pool, err = pgxpool.New(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("DB connection failed: %v", err)
		}
		defer pool.Close()
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Start pprof server on :6060 only if enabled (development mode)
	if cfg.EnablePprof {
		go func() {
			log.Println("Starting pprof server on :6060 (development mode)")
			if err := http.ListenAndServe(":6060", nil); err != nil {
				log.Printf("pprof server error: %v", err)
			}
		}()
	}

	deleteQueue := make(chan handlers.DeleteRequest, 100)

	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithUserID)
	r.Use(middleware.LoggingMiddleware(logger))

	var store storage.Storage
	if pool != nil {
		store, err = storage.NewPostgresStorage(pool)
		if err != nil {
			log.Fatalf("failed to initialize PostgreSQL storage: %v", err)
		}
		if _, ok := store.(*storage.PostgresStorage); ok {
			go handlers.RunDeletionWorker(context.Background(), store, logger, deleteQueue)
		}
	} else {
		if cfg.FileStoragePath != "" {
			s, err := storage.NewFileStorage(cfg.FileStoragePath)
			if err == nil {
				store = s
				if _, ok := store.(*storage.FileStorage); ok {
					go handlers.RunDeletionWorker(context.Background(), store, logger, deleteQueue)
				}
			}
		}
		if store == nil {
			store = storage.NewInMemoryStorage()
		}
	}

	r.Post("/", handlers.ShortenURLHandler(cfg, store))
	r.Post("/api/shorten", handlers.APIShortenHandler(cfg, store))
	r.Post("/api/shorten/batch", handlers.BatchShortenHandler(cfg, store))
	r.Get("/{id}", handlers.RedirectHandler(store))
	r.Get("/ping", handlers.PingHandler(pool))
	r.Get("/api/user/urls", handlers.UserURLsHandler(cfg, store))
	r.Delete("/api/user/urls", handlers.DeleteUserURLsHandler(store, logger, deleteQueue))

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Println("Starting server on", cfg.Address)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// printBuildInfo выводит информацию о сборке приложения
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", getBuildValue(buildVersion))
	fmt.Printf("Build date: %s\n", getBuildValue(buildDate))
	fmt.Printf("Build commit: %s\n", getBuildValue(buildCommit))
}

// getBuildValue возвращает значение сборки или "N/A" если значение пустое
func getBuildValue(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
