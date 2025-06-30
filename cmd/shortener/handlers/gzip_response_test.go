package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uno/cmd/shortener/middleware"

	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"

	"github.com/go-chi/chi/v5"
)

func setupGzipRouter(cfg *config.Config, store storage.Storage) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.WithUserID)
	r.Use(middleware.GzipMiddleware)
	r.Post("/api/shorten", APIShortenHandler(cfg, store))
	return r
}

func TestGzipResponse(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	h := setupGzipRouter(cfg, store)

	body := `{"url":"https://gzip.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)

	if res.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding, got %s", res.Header().Get("Content-Encoding"))
	}

	gzr, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gzr.Close()

	var apiResp models.APIResponse
	data, _ := io.ReadAll(gzr)
	if err := json.Unmarshal(data, &apiResp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}
