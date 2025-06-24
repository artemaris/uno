package storage

import (
	"context"
	"errors"
	"fmt"
	"uno/cmd/shortener/models"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
            original_url varchar UNIQUE NOT NULL,
            user_id varchar NOT NULL
        )
    `)
	return err
}

func (s *PostgresStorage) Save(shortID, originalURL, userID string) {
	_, err := s.conn.Exec(context.Background(),
		`INSERT INTO public.short_urls (id, original_url, user_id) VALUES ($1, $2, $3)`,
		shortID, originalURL, userID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return
		}
	}
}

func (s *PostgresStorage) FindByOriginal(originalURL string) (string, bool) {
	var id string
	err := s.conn.QueryRow(context.Background(),
		`SELECT id FROM public.short_urls WHERE original_url = $1`, originalURL,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return id, true
}

func (s *PostgresStorage) SaveBatch(pairs map[string]string, userID string) error {
	batch := &pgx.Batch{}
	for shortID, originalURL := range pairs {
		batch.Queue(`INSERT INTO public.short_urls (id, original_url, user_id) VALUES ($1, $2, $3)
                     ON CONFLICT (id) DO NOTHING`, shortID, originalURL, userID)
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

func (s *PostgresStorage) GetUserURLs(userID string) []models.UserURL {
	rows, err := s.conn.Query(context.Background(),
		`SELECT id, original_url FROM public.short_urls WHERE user_id = $1`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []models.UserURL
	for rows.Next() {
		var id, original string
		if err := rows.Scan(&id, &original); err != nil {
			continue
		}
		result = append(result, models.UserURL{
			ShortURL:    id,
			OriginalURL: original,
		})
	}
	return result
}
