package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uno/cmd/shortener/config"

	"github.com/go-chi/chi/v5"
)

func setupRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Post("/", shortenURLHandler(cfg))
	r.Get("/{id}", redirectHandler)
	return r
}

func TestShortenAndRedirect(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	handler := setupRouter(cfg)

	reqBody := "https://practicum.yandex.ru/"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.Code)
	}

	shortURL := strings.TrimSpace(resp.Body.String())
	if !strings.HasPrefix(shortURL, cfg.BaseURL+"/") {
		t.Fatalf("short URL does not have expected prefix: got %s", shortURL)
	}

	shortID := strings.TrimPrefix(shortURL, cfg.BaseURL+"/")

	getReq := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	getResp := httptest.NewRecorder()

	handler.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, getResp.Code)
	}

	location := getResp.Header().Get("Location")
	if location != reqBody {
		t.Fatalf("expected redirect to %s, got %s", reqBody, location)
	}
}
