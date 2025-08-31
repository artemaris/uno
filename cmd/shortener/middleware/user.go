package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// contextKey определяет тип ключа для контекста
type contextKey string

// ContextUserIDKey ключ для хранения идентификатора пользователя в контексте
const ContextUserIDKey contextKey = "userID"

// userIDCookieName имя cookie для хранения идентификатора пользователя
const userIDCookieName = "auth_user"

// WithUserID middleware добавляет идентификатор пользователя в контекст запроса
// Если у пользователя нет cookie с userID, создается новый UUID и устанавливается cookie
// Идентификатор пользователя доступен в последующих обработчиках через FromContext
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

// FromContext извлекает идентификатор пользователя из контекста запроса
// Возвращает userID и флаг успешности извлечения
func FromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextUserIDKey).(string)
	return userID, ok
}
