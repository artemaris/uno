package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"uno/cmd/shortener/middleware"
)

type deleteRequest struct {
	userID string
	ids    []string
}

func DeleteUserURLsHandler(deleteQueue chan<- deleteRequest) http.HandlerFunc {
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

		deleteQueue <- deleteRequest{userID: userID, ids: ids}

		w.WriteHeader(http.StatusAccepted)
	}
}
