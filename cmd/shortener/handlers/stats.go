package handlers

import (
	"encoding/json"
	"net/http"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"
)

// StatsHandler возвращает статистику сервиса
func StatsHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем статистику из хранилища
		stats, err := store.GetStats()
		if err != nil {
			http.Error(w, "Failed to get stats", http.StatusInternalServerError)
			return
		}

		// Создаем ответ
		response := models.StatsResponse{
			URLs:  stats.URLs,
			Users: stats.Users,
		}

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Кодируем и отправляем ответ
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
