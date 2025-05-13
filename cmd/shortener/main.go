package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/storage"
	"uno/cmd/shortener/utils"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()
	store := storage.NewInMemoryStorage()

	r := chi.NewRouter()
	r.Post("/", shortenURLHandler(cfg, store))
	r.Get("/{id}", redirectHandler(store))

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: r,
	}

	log.Println("Starting server on", cfg.Address)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func shortenURLHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		shortID := utils.GenerateShortID()
		store.Save(shortID, originalURL)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, cfg.BaseURL+"/"+shortID)
	}
}

func redirectHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, "id")
		originalURL, ok := store.Get(shortID)
		if !ok {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
