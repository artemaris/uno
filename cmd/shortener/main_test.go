package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenAndRedirect(t *testing.T) {
	handler := setupRouter()

	originalURL := "https://practicum.yandex.ru/"

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	shortURL := strings.TrimSpace(string(body))

	if !strings.HasPrefix(shortURL, baseURL) {
		t.Fatalf("short URL does not start with baseURL: got %q, want prefix %q", shortURL, baseURL)
	}

	id := strings.TrimPrefix(shortURL, baseURL)
	if id == "" {
		t.Fatal("short ID is empty")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	getRec := httptest.NewRecorder()

	handler.ServeHTTP(getRec, getReq)

	getResp := getRec.Result()
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, getResp.StatusCode)
	}

	location := getResp.Header.Get("Location")
	if location != originalURL {
		t.Fatalf("expected Location header %q, got %q", originalURL, location)
	}
}
