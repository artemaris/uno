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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	shortURL := strings.TrimSpace(string(body))

	if !strings.HasPrefix(shortURL, baseURL) {
		t.Fatalf("shortened URL does not start with baseURL: got %q, want prefix %q", shortURL, baseURL)
	}

	id := strings.TrimPrefix(shortURL, baseURL)
	if id == "" {
		t.Fatal("shortened URL does not have an ID")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	getRec := httptest.NewRecorder()

	handler.ServeHTTP(getRec, getReq)

	getResp := getRec.Result()
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status code %d, got %d", http.StatusTemporaryRedirect, getResp.StatusCode)
	}

	location := getResp.Header.Get("Location")
	if location != originalURL {
		t.Fatalf("expected location header %q, got %q", originalURL, location)
	}
}
