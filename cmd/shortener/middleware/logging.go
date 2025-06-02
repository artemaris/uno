package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	data *responseData
}

func (l *loggingResponseWriter) WriteHeader(statusCode int) {
	l.data.status = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.data.size += size
	return size, err
}

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
				zap.Duration("duration", duration),
			)
		})
	}
}
