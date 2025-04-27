package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

var urlToStore = make(map[string]string)
var baseURL = "http://localhost:8080/"

func main() {
	http.ListenAndServe(":8080", setupRouter())
}

func setupRouter() http.Handler {
	r := chi.NewRouter()
	r.Post("/", shortenURLHandler)
	r.Get("/{id}", redirectHandler)
	return r
}

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	originalURL := strings.TrimSpace(string(body))

	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortID := generateShortID(originalURL)
	urlToStore[shortID] = originalURL

	shortURL := baseURL + shortID

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprint(w, shortURL)
	if err != nil {
		return
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")
	originalURL, ok := urlToStore[shortID]
	if !ok {
		http.Error(w, "Short URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateShortID(url string) string {
	if len(url) < 6 {
		return fmt.Sprintf("%x", len(url))
	}
	return fmt.Sprintf("%x", url[:6])
}
