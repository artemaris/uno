package service

import (
	"context"
	"net/url"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"
	"uno/cmd/shortener/utils"

	"go.uber.org/zap"
)

// Service содержит бизнес-логику для работы с URL
type Service struct {
	config *config.Config
	store  storage.Storage
	logger *zap.Logger
}

// NewService создает новый экземпляр сервиса
func NewService(cfg *config.Config, store storage.Storage, logger *zap.Logger) *Service {
	return &Service{
		config: cfg,
		store:  store,
		logger: logger,
	}
}

// ValidateURL проверяет валидность URL
func (s *Service) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return &ValidationError{Message: "URL is required"}
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return &ValidationError{Message: "Invalid URL format"}
	}

	return nil
}

// ShortenURL сокращает URL
func (s *Service) ShortenURL(ctx context.Context, originalURL, userID string) (string, error) {
	// Валидация URL
	if err := s.ValidateURL(originalURL); err != nil {
		return "", err
	}

	// Проверяем, существует ли уже такой URL
	if existingID, found := s.store.FindByOriginal(originalURL); found {
		return s.config.BaseURL + "/" + existingID, nil
	}

	// Генерируем новый сокращенный ID
	shortID, err := utils.GenerateShortID()
	if err != nil {
		s.logger.Error("Failed to generate short ID", zap.Error(err))
		return "", &InternalError{Message: "Failed to generate short ID"}
	}

	// Сохраняем URL
	s.store.Save(shortID, originalURL, userID)

	s.logger.Info("URL shortened", zap.String("original", originalURL), zap.String("short", shortID))

	return s.config.BaseURL + "/" + shortID, nil
}

// GetURL получает оригинальный URL по сокращенному
func (s *Service) GetURL(ctx context.Context, shortID string) (string, bool, bool, error) {
	if shortID == "" {
		return "", false, false, &ValidationError{Message: "ID is required"}
	}

	originalURL, deleted, exists := s.store.Get(shortID)
	return originalURL, deleted, exists, nil
}

// BatchShortenURL сокращает несколько URL пакетом
func (s *Service) BatchShortenURL(ctx context.Context, items []BatchItem, userID string) ([]BatchResult, error) {
	if len(items) == 0 {
		return nil, &ValidationError{Message: "At least one URL is required"}
	}

	// Подготавливаем данные для пакетного сохранения
	pairs := make(map[string]string)
	results := make([]BatchResult, 0, len(items))

	for _, item := range items {
		// Валидация URL
		if err := s.ValidateURL(item.OriginalURL); err != nil {
			return nil, &ValidationError{Message: err.Error() + " for correlation ID: " + item.CorrelationID}
		}

		// Проверяем, существует ли уже такой URL
		if existingID, found := s.store.FindByOriginal(item.OriginalURL); found {
			results = append(results, BatchResult{
				CorrelationID: item.CorrelationID,
				ShortURL:      s.config.BaseURL + "/" + existingID,
			})
			continue
		}

		// Генерируем новый сокращенный ID
		shortID, err := utils.GenerateShortID()
		if err != nil {
			return nil, &InternalError{Message: "Failed to generate short ID for correlation ID: " + item.CorrelationID}
		}
		pairs[shortID] = item.OriginalURL

		results = append(results, BatchResult{
			CorrelationID: item.CorrelationID,
			ShortURL:      s.config.BaseURL + "/" + shortID,
		})
	}

	// Сохраняем все URL пакетом
	if err := s.store.SaveBatch(pairs, userID); err != nil {
		s.logger.Error("Failed to save batch URLs", zap.Error(err))
		return nil, &InternalError{Message: "Failed to save URLs"}
	}

	s.logger.Info("Batch URLs shortened", zap.Int("count", len(items)))

	return results, nil
}

// GetUserURLs получает URL пользователя
func (s *Service) GetUserURLs(ctx context.Context, userID string) ([]models.UserURL, error) {
	if userID == "" {
		return []models.UserURL{}, nil
	}

	userURLs, err := s.store.GetUserURLs(userID)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, &InternalError{Message: "Failed to get user URLs"}
	}

	return userURLs, nil
}

// DeleteUserURLs удаляет URL пользователя
func (s *Service) DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error {
	if userID == "" || len(shortURLs) == 0 {
		return nil
	}

	// Извлекаем ID из полных URL
	ids := make([]string, 0, len(shortURLs))
	for _, url := range shortURLs {
		// Извлекаем ID из URL (убираем базовый URL)
		if len(url) > len(s.config.BaseURL)+1 {
			id := url[len(s.config.BaseURL)+1:]
			ids = append(ids, id)
		}
	}

	// Удаляем URL
	if err := s.store.DeleteURLs(userID, ids); err != nil {
		s.logger.Error("Failed to delete user URLs", zap.Error(err))
		return &InternalError{Message: "Failed to delete URLs"}
	}

	s.logger.Info("User URLs deleted", zap.String("userID", userID), zap.Int("count", len(ids)))

	return nil
}

// GetStats получает статистику
func (s *Service) GetStats(ctx context.Context) (storage.Stats, error) {
	stats, err := s.store.GetStats()
	if err != nil {
		s.logger.Error("Failed to get stats", zap.Error(err))
		return storage.Stats{}, &InternalError{Message: "Failed to get stats"}
	}

	return stats, nil
}

// GetBaseURL возвращает базовый URL
func (s *Service) GetBaseURL() string {
	return s.config.BaseURL
}

// BatchItem представляет элемент пакетного запроса
type BatchItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResult представляет результат пакетного запроса
type BatchResult struct {
	CorrelationID string
	ShortURL      string
}
