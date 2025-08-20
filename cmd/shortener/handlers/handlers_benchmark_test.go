package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"
)

func benchmarkShorten(b *testing.B, handler http.HandlerFunc) {
	b.Helper()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
		rec := httptest.NewRecorder()
		h := middleware.WithUserID(handler)
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkShortenURLHandler_InMemory(b *testing.B) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	h := ShortenURLHandler(cfg, store)
	benchmarkShorten(b, h)
}

func BenchmarkRedirectHandler_InMemory(b *testing.B) {
	store := storage.NewInMemoryStorage()
	store.Save("abc12345", "https://example.com", "user")
	h := RedirectHandler(store)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/abc12345", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
}

func BenchmarkUserURLsHandler_InMemory(b *testing.B) {
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()
	userID := "user"
	// populate
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("id-%d", i)
		url := fmt.Sprintf("https://example.com/%d", i)
		store.Save(id, url, userID)
	}
	// delete some
	var dels []string
	for i := 0; i < 200; i++ {
		dels = append(dels, fmt.Sprintf("id-%d", i))
	}
	_ = store.DeleteURLs(userID, dels)

	h := UserURLsHandler(cfg, store)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.ContextUserIDKey, userID))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
}
