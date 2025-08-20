package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData содержит информацию о HTTP ответе для логирования
type responseData struct {
	status      int    // HTTP статус код
	size        int    // Размер ответа в байтах
	contentType string // Тип содержимого ответа
}

// loggingResponseWriter обертка для http.ResponseWriter, которая собирает данные для логирования
type loggingResponseWriter struct {
	http.ResponseWriter
	data *responseData
}

// WriteHeader перехватывает вызов WriteHeader для записи статус кода
func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.data.status = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает вызов Write для подсчета размера ответа и определения типа содержимого
func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.data.size += size
	l.data.contentType = l.Header().Get("Content-Type")
	return size, err
}

// LoggingMiddleware middleware для логирования HTTP запросов и ответов
// Логирует метод, URI, статус код, размер ответа, тип содержимого и время выполнения
// Использует структурированное логирование через zap.Logger
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			data := &responseData{
				status: 200,
			}
			lw := &loggingResponseWriter{
				ResponseWriter: w,
				data:           data,
			}

			next.ServeHTTP(lw, r)

			duration := time.Since(start)

			logger.Info("request handled",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", data.status),
				zap.Int("size", data.size),
				zap.String("content_type", data.contentType),
				zap.Duration("duration", duration),
			)
		})
	}
}
