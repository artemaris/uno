package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/handlers"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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

	// Создаем контекст с обработкой сигналов
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var pool *pgxpool.Pool
	if cfg.DatabaseDSN != "" {
		var err error
		pool, err = pgxpool.New(ctx, cfg.DatabaseDSN)
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
	}

	if store == nil {
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

	// Запускаем worker для удаления URL если поддерживается
	if _, ok := store.(*storage.PostgresStorage); ok {
		go handlers.RunDeletionWorker(ctx, store, logger, deleteQueue)
	} else if _, ok := store.(*storage.FileStorage); ok {
		go handlers.RunDeletionWorker(ctx, store, logger, deleteQueue)
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

	// Создаем errgroup для управления горутинами
	g, gCtx := errgroup.WithContext(ctx)

	// Запускаем pprof сервер если включен
	if cfg.EnablePprof {
		g.Go(func() error {
			log.Println("Starting pprof server on :6060 (development mode)")
			pprofSrv := &http.Server{Addr: ":6060", Handler: nil}
			go func() {
				<-gCtx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				pprofSrv.Shutdown(shutdownCtx)
			}()
			return pprofSrv.ListenAndServe()
		})
	}

	// Запускаем основной сервер
	g.Go(func() error {
		log.Println("Starting server on", cfg.Address)
		if cfg.EnableHTTPS {
			log.Printf("HTTPS enabled, using certificate: %s, key: %s", cfg.CertFile, cfg.KeyFile)
			err := srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
			// Игнорируем ошибку если сервер был остановлен через Shutdown
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		} else {
			err := srv.ListenAndServe()
			// Игнорируем ошибку если сервер был остановлен через Shutdown
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		}
	})

	// Обработка graceful shutdown
	g.Go(func() error {
		<-gCtx.Done()
		log.Println("Received shutdown signal, starting graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
			return err
		}
		log.Println("Server shutdown completed successfully")
		return nil
	})

	// Ждем завершения всех горутин
	if err := g.Wait(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
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
