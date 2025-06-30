package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func NewPG(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot create pgx pool: %w", err)
	}
	// Проверим подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("cannot connect to postgres: %w", err)
	}
	return pool, nil
}
