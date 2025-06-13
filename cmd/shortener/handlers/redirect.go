package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"uno/cmd/shortener/storage"
)

func RedirectHandler(store storage.Storage) http.HandlerFunc {
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
