package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func NewPG(dsn string) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to postgres: %w", err)
	}

	return conn, nil
}
