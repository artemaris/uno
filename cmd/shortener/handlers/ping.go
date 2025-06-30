package handlers

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

func PingHandler(conn *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if conn == nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		if err := conn.Ping(ctx); err != nil {
			http.Error(w, "database connection failed", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
