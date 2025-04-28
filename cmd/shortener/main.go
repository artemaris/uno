package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"uno/cmd/shortener/config"

	"github.com/go-chi/chi/v5"
)

var urlStore = make(map[string]string)

const idLength = 8
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func main() {
	cfg := config.NewConfig()

	r := chi.NewRouter()
	r.Post("/", shortenURLHandler(cfg))
	r.Get("/{id}", redirectHandler)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	fmt.Println("Starting server on", cfg.Address)
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func shortenURLHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
			http.Error(w, "unsupported Content-Type", http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		originalURL := strings.TrimSpace(string(body))
		if originalURL == "" {
			http.Error(w, "empty URL", http.StatusBadRequest)
			return
		}

		shortID := generateShortID()
		urlStore[shortID] = originalURL

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, cfg.BaseURL+"/"+shortID)
	}
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")
	originalURL, ok := urlStore[shortID]
	if !ok {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func generateShortID() string {
	rand.NewSource(time.Now().UnixNano())
	id := make([]byte, idLength)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}
