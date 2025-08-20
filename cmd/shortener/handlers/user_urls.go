package handlers

import (
	"encoding/json"
	"net/http"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"
)

// UserURLsHandler обрабатывает GET запросы для получения всех URL пользователя
// Возвращает JSON массив с информацией о сокращенных URL пользователя
// Если у пользователя нет URL, возвращает статус 204 No Content
// Удаленные URL исключаются из результата
func UserURLsHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.FromContext(r.Context())
		if !ok || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		urls, err := store.GetUserURLs(userID)
		if err != nil {
			http.Error(w, "failed to get user URLs", http.StatusInternalServerError)
			return
		}

		var filtered []models.UserURL
		for _, url := range urls {
			if !url.Deleted {
				url.ShortURL = cfg.BaseURL + "/" + url.ShortURL
				filtered = append(filtered, url)
			}
		}

		if len(filtered) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(filtered)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
