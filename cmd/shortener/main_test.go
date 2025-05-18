package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"

	"uno/cmd/shortener/config"

	"github.com/go-chi/chi/v5"
)

func setupRouter(cfg *config.Config, store storage.Storage) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.GzipMiddleware)
	r.Post("/", shortenURLHandler(cfg, store))
	r.Post("/api/shorten", apiShortenHandler(cfg, store))
	r.Get("/{id}", redirectHandler(store))
	return r
}

func TestShortenAndRedirect(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	store := storage.NewInMemoryStorage()
	handler := setupRouter(cfg, store)

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

	reqEmpty := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	reqEmpty.Header.Set("Content-Type", "text/plain")
	respEmpty := httptest.NewRecorder()
	handler.ServeHTTP(respEmpty, reqEmpty)

	if respEmpty.Code != http.StatusBadRequest {
		t.Fatalf("empty body: expected %d, got %d", http.StatusBadRequest, respEmpty.Code)
	}

	reqNonexist := httptest.NewRequest(http.MethodGet, "/unknown123", nil)
	respNonexist := httptest.NewRecorder()
	handler.ServeHTTP(respNonexist, reqNonexist)

	if respNonexist.Code != http.StatusBadRequest {
		t.Fatalf("non-existent ID: expected %d, got %d", http.StatusBadRequest, respNonexist.Code)
	}
}

func TestAPIShortenHandler(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	store := storage.NewInMemoryStorage()
	handler := setupRouter(cfg, store)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.Code)
	}

	reqBad := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{}`))
	reqBad.Header.Set("Content-Type", "application/json")
	respBad := httptest.NewRecorder()
	handler.ServeHTTP(respBad, reqBad)
	if respBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty input, got %d", respBad.Code)
	}
}

func TestGzipResponse(t *testing.T) {
	cfg := &config.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	store := storage.NewInMemoryStorage()
	handler := setupRouter(cfg, store)

	jsonBody := `{"url":"https://gzip-test.example"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response, got: %s", resp.Header().Get("Content-Encoding"))
	}

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer func(gr *gzip.Reader) {
		err := gr.Close()
		if err != nil {
			t.Fatalf("failed to close gzip reader: %v", err)
		}
	}(gr)

	var apiResp models.APIResponse
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to read decompressed body: %v", err)
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		t.Fatalf("invalid JSON in decompressed response: %v", err)
	}

	if !strings.HasPrefix(apiResp.Result, cfg.BaseURL+"/") {
		t.Errorf("unexpected result: %s", apiResp.Result)
	}
}
