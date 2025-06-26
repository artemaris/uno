package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uno/cmd/shortener/config"
	"uno/cmd/shortener/handlers"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"

	"github.com/go-chi/chi/v5"
)

func setupAPIShortenRouter(cfg *config.Config, store storage.Storage) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.WithUserIDMiddleware("supersecret"))
	r.Post("/api/shorten", handlers.APIShortenHandler(cfg, store))
	return r
}

func TestAPIShortenHandler_Valid(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	h := setupAPIShortenRouter(cfg, store)

	body := `{"url":"https://foo.bar"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, res.Code)
	}
}

func TestAPIShortenHandler_InvalidJSON(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	h := setupAPIShortenRouter(cfg, store)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected %d for empty payload, got %d", http.StatusBadRequest, res.Code)
	}
}
