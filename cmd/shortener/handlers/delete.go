package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"
)

func DeleteUserURLsHandler(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.FromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var ids []string
		if err := json.Unmarshal(body, &ids); err != nil || len(ids) == 0 {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		go func() {
			err := store.DeleteURLs(userID, ids)
			if err != nil {
				log.Println(err)
			}
		}()

		w.WriteHeader(http.StatusAccepted)
	}
}
