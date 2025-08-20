package storage

import (
	"context"
	"errors"
	"fmt"
	"uno/cmd/shortener/models"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// deleteTask представляет задачу на удаление URL для конкретного пользователя
type deleteTask struct {
	UserID    string   // Идентификатор пользователя
	ShortURLs []string // Список сокращенных ID для удаления
}

// PostgresStorage реализует интерфейс Storage с использованием PostgreSQL
// Поддерживает асинхронные операции удаления через очередь задач
type PostgresStorage struct {
	pool        *pgxpool.Pool   // Пул соединений с базой данных
	deleteQueue chan deleteTask // Канал для очереди задач на удаление
}

// NewPostgresStorage создает новый экземпляр PostgresStorage
// Инициализирует схему базы данных при создании
func NewPostgresStorage(conn *pgxpool.Pool) (Storage, error) {
	s := &PostgresStorage{
		pool:        conn,
		deleteQueue: make(chan deleteTask, 100),
	}
	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}
	return s, nil
}

// Save сохраняет связь между сокращенным ID и оригинальным URL для конкретного пользователя
// Игнорирует ошибки уникальности (URL уже существует)
func (s *PostgresStorage) Save(shortID, originalURL, userID string) {
	_, err := s.pool.Exec(context.Background(),
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

// FindByOriginal ищет существующий сокращенный ID для оригинального URL
// Возвращает пустую строку, если URL не найден или удален
func (s *PostgresStorage) FindByOriginal(originalURL string) (string, bool) {
	var id string
	err := s.pool.QueryRow(context.Background(),
		`SELECT id FROM public.short_urls WHERE original_url = $1 AND is_deleted = false`, originalURL,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return id, true
}

// SaveBatch сохраняет пакет URL для конкретного пользователя
// Использует batch операции для оптимизации производительности
func (s *PostgresStorage) SaveBatch(pairs map[string]string, userID string) error {
	batch := &pgx.Batch{}
	for shortID, originalURL := range pairs {
		batch.Queue(`INSERT INTO public.short_urls (id, original_url, userID) VALUES ($1, $2, $3)
                     ON CONFLICT (id) DO NOTHING`, shortID, originalURL, userID)
	}

	br := s.pool.SendBatch(context.Background(), batch)
	defer br.Close()

	for range pairs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

// Get возвращает оригинальный URL по сокращенному ID, флаг удаления и существования
func (s *PostgresStorage) Get(shortID string) (string, bool, bool) {
	var originalURL string
	var deleted bool

	err := s.pool.QueryRow(context.Background(),
		`SELECT original_url, is_deleted FROM public.short_urls WHERE id = $1`, shortID,
	).Scan(&originalURL, &deleted)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, false
	}
	if err != nil {
		return "", false, false
	}

	return originalURL, deleted, true
}

// GetUserURLs возвращает все не удаленные URL для конкретного пользователя
func (s *PostgresStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	rows, err := s.pool.Query(context.Background(),
		`SELECT id, original_url FROM public.short_urls WHERE user_id = $1 AND is_deleted = false`, userID)
	if err != nil {
		return nil, err
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
	return result, nil
}

// DeleteURLs помечает указанные URL как удаленные для конкретного пользователя
// Возвращает ошибку, если не было обновлено ни одной строки
func (s *PostgresStorage) DeleteURLs(userID string, shortIDs []string) error {
	commandTag, err := s.pool.Exec(context.Background(),
		`UPDATE public.short_urls SET is_deleted = true WHERE user_id = $1 AND id = ANY($2)`, userID, shortIDs)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated")
	}
	return nil
}

// RunDeletionWorker запускает воркер для асинхронной обработки задач на удаление URL
// Обрабатывает задачи из очереди deleteQueue до завершения контекста
func (s *PostgresStorage) RunDeletionWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.deleteQueue:
			if len(task.ShortURLs) == 0 {
				continue
			}
			err := s.DeleteURLs(task.UserID, task.ShortURLs)
			if err != nil {
				fmt.Printf("failed to delete URLs for user %s: %v\n", task.UserID, err)
			}
		}
	}
}

// initSchema инициализирует схему базы данных
// Создает таблицу short_urls, если она не существует
func (s *PostgresStorage) initSchema() error {
	_, err := s.pool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS public.short_urls (
            id varchar PRIMARY KEY,
            original_url varchar UNIQUE NOT NULL,
            user_id varchar NOT NULL,
            is_deleted boolean DEFAULT false
        )
    `)
	return err
}
