package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func WithUserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		cookie, err := r.Cookie("auth_user")
		if err != nil || cookie.Value == "" {
			userID = uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:  "auth_user",
				Value: userID,
				Path:  "/",
			})
		} else {
			userID = cookie.Value
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}
