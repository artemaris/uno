package handlers

import (
	"encoding/json"
	"net/http"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"

	"github.com/google/uuid"
)

const userIDCookieName = "auth_user"

func UserURLsHandler(cfg *config.Config, store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromCookie(w, r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		urls, err := store.GetUserURLs(userID)
		var filtered []models.UserURL
		for _, url := range urls {
			if !url.Deleted {
				url.ShortURL = cfg.BaseURL + "/" + url.ShortURL
				filtered = append(filtered, url)
			}
		}
		if err != nil {
			http.Error(w, "failed to get user URLs", http.StatusInternalServerError)
			return
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

func getUserIDFromCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(userIDCookieName)
	if err != nil || cookie.Value == "" {
		newID := uuid.NewString()
		http.SetCookie(w, &http.Cookie{
			Name:  userIDCookieName,
			Value: newID,
			Path:  "/",
		})
		return newID, nil
	}
	return cookie.Value, nil
}
