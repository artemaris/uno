package handlers

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"
)

var deleteQueue = make(chan deleteRequest, 100)

type deleteRequest struct {
	userID string
	ids    []string
}

func DeleteUserURLsHandler(store storage.Storage, logger *zap.Logger) http.HandlerFunc {
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

func RunDeletionWorker(store storage.Storage, logger *zap.Logger) {
	go func() {
		for req := range deleteQueue {
			err := store.DeleteURLs(req.userID, req.ids)
			if err != nil {
				logger.Error("batch deletion failed", zap.Error(err))
			}
		}
	}()
}
