package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"

	"go.uber.org/zap"
)

// DeleteRequest представляет запрос на удаление URL пользователя
type DeleteRequest struct {
	UserID string   // Идентификатор пользователя
	IDs    []string // Список сокращенных ID для удаления
}

// DeleteUserURLsHandler обрабатывает DELETE запросы для пометки URL как удаленных
// Принимает JSON массив с сокращенными ID и помещает запрос в очередь на асинхронную обработку
// Возвращает статус 202 Accepted, так как удаление выполняется асинхронно
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

// RunDeletionWorker запускает воркер для асинхронной обработки запросов на удаление URL
// Обрабатывает запросы из очереди deleteQueue и выполняет пометку URL как удаленных
// Каждый запрос обрабатывается в отдельной горутине для параллельности
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
