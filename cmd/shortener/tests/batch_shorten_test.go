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

func setupBatchRouter(cfg *config.Config, store storage.Storage) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.WithUserIDMiddleware("supersecret"))
	r.Post("/api/shorten/batch", handlers.BatchShortenHandler(cfg, store))
	return r
}

func TestBatchShortenHandler(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	h := setupBatchRouter(cfg, store)

	cases := []struct {
		body   string
		status int
	}{
		{`[{"correlation_id":"1","original_url":"https://a.com"}]`, http.StatusCreated},
		{`[]`, http.StatusBadRequest},
		{`not-json`, http.StatusBadRequest},
	}

	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(c.body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != c.status {
			t.Errorf("body=%q expected %d, got %d", c.body, c.status, res.Code)
		}
	}
}
