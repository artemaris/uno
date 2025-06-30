package handlers

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"
)

type DeleteRequest struct {
	UserID string
	IDs    []string
}

func DeleteUserURLsHandler(store storage.Storage, logger *zap.Logger, deleteQueue chan<- DeleteRequest) http.HandlerFunc {
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

		deleteQueue <- DeleteRequest{UserID: userID, IDs: ids}

		w.WriteHeader(http.StatusAccepted)
	}
}

func RunDeletionWorker(
	ctx context.Context,
	store storage.Storage,
	logger *zap.Logger,
	deleteQueue <-chan DeleteRequest,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-deleteQueue:
			go func(r DeleteRequest) {
				if err := store.DeleteURLs(r.UserID, r.IDs); err != nil {
					logger.Error("batch deletion failed", zap.Error(err))
				}
			}(req)
		}
	}
}
