package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const ContextUserIDKey contextKey = "userID"
const userIDCookieName = "auth_user"

func WithUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(userIDCookieName)
		userID := ""

		if err != nil || cookie.Value == "" {
			userID = uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:  userIDCookieName,
				Value: userID,
				Path:  "/",
			})
		} else {
			userID = cookie.Value
		}

		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextUserIDKey).(string)
	return userID, ok
}
