package handlers

import (
	"encoding/json"
	"net/http"
	"uno/cmd/shortener/config"
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

		urls := store.GetUserURLs(userID)
		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		data, err := json.Marshal(urls)
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
