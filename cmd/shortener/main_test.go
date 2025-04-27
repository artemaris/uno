package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doPostRequest(t *testing.T, server http.Handler, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	return rr
}

func doGetRequest(t *testing.T, server http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	return rr
}

func TestShortenAndRedirect(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			shortenURLHandler(w, r)
		} else {
			redirectHandler(w, r)
		}
	})

	originalURL := "https://practicum.yandex.ru/"
	postResp := doPostRequest(t, handler, originalURL)

	if postResp.Code != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", postResp.Code)
	}

	shortURLBody, _ := io.ReadAll(postResp.Body)
	shortURL := strings.TrimSpace(string(shortURLBody))
	if !strings.Contains(shortURL, "http://localhost:8080/") {
		t.Fatalf("Short URL is not correct: %s", shortURL)
	}

	parts := strings.Split(shortURL, "/")
	shortID := parts[len(parts)-1]

	getResp := doGetRequest(t, handler, "/"+shortID)

	if getResp.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected status 307 Temporary Redirect, got %d", getResp.Code)
	}

	location := getResp.Header().Get("Location")
	if location != originalURL {
		t.Fatalf("Expected Location header to be %s, got %s", originalURL, location)
	}
}

func TestBadRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			shortenURLHandler(w, r)
		} else {
			redirectHandler(w, r)
		}
	})

	getResp := doGetRequest(t, handler, "/nonexistentid")

	if getResp.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found for nonexistent ID, got %d", getResp.Code)
	}
}
