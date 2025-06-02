package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresStorage struct {
	conn *pgx.Conn
}

func NewPostgresStorage(conn *pgx.Conn) (Storage, error) {
	s := &PostgresStorage{conn: conn}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}
	return s, nil
}

func (s *PostgresStorage) initSchema() error {
	_, err := s.conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS public.short_urls (
            id varchar PRIMARY KEY,
            original_url varchar NOT NULL
        )
    `)
	return err
}

func (s *PostgresStorage) Save(shortID, originalURL string) {
	_, _ = s.conn.Exec(context.Background(),
		`INSERT INTO public.short_urls (id, original_url) VALUES ($1, $2)
         ON CONFLICT (id) DO NOTHING`,
		shortID, originalURL,
	)
}

func (s *PostgresStorage) SaveBatch(pairs map[string]string) error {
	batch := &pgx.Batch{}
	for shortID, originalURL := range pairs {
		batch.Queue(`INSERT INTO public.short_urls (id, original_url) VALUES ($1, $2)
                     ON CONFLICT (id) DO NOTHING`, shortID, originalURL)
	}

	br := s.conn.SendBatch(context.Background(), batch)
	defer br.Close()

	for range pairs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) Get(shortID string) (string, bool) {
	var originalURL string
	err := s.conn.QueryRow(context.Background(),
		`SELECT original_url FROM public.short_urls WHERE id = $1`, shortID,
	).Scan(&originalURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return originalURL, true
}
