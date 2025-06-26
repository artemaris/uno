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

func setupShortenRouter(cfg *config.Config, store storage.Storage) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.WithUserIDMiddleware("supersecret"))
	r.Post("/", handlers.ShortenURLHandler(cfg, store))
	r.Get("/{id}", handlers.RedirectHandler(store))
	return r
}

func TestShortenAndRedirect(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	handler := setupShortenRouter(cfg, store)

	orig := "https://example.com/"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(orig))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("POST / expected %d, got %d", http.StatusCreated, res.Code)
	}
	shortURL := strings.TrimSpace(res.Body.String())
	if !strings.HasPrefix(shortURL, cfg.BaseURL+"/") {
		t.Fatalf("short URL prefix: got %q", shortURL)
	}

	id := strings.TrimPrefix(shortURL, cfg.BaseURL+"/")
	getReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	getRes := httptest.NewRecorder()
	handler.ServeHTTP(getRes, getReq)

	if getRes.Code != http.StatusTemporaryRedirect {
		t.Fatalf("GET /%s expected %d, got %d", id, http.StatusTemporaryRedirect, getRes.Code)
	}
	if loc := getRes.Header().Get("Location"); loc != orig {
		t.Fatalf("redirect Location: expected %q, got %q", orig, loc)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	badRes := httptest.NewRecorder()
	handler.ServeHTTP(badRes, badReq)
	if badRes.Code != http.StatusGone {
		t.Fatalf("GET /notfound expected %d, got %d", http.StatusGone, badRes.Code)
	}
}
