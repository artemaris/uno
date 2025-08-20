package handlers

import (
	"net/http"
	"uno/cmd/shortener/storage"

	"github.com/go-chi/chi/v5"
)

func RedirectHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, "id")
		originalURL, deleted, exists := store.Get(shortID)

		if !exists {
			http.NotFound(w, r)
			return
		}

		if deleted {
			w.WriteHeader(http.StatusGone)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
