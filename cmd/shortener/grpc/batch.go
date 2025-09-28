package grpc

import (
	"context"
	"net/url"
	"uno/api/proto"
	"uno/cmd/shortener/utils"

	"go.uber.org/zap"
)

// BatchShortenURL обрабатывает запрос на пакетное сокращение URL
func (s *Server) BatchShortenURL(ctx context.Context, req *proto.BatchShortenURLRequest) (*proto.BatchShortenURLResponse, error) {
	if len(req.Items) == 0 {
		return nil, &InvalidArgumentError{Message: "At least one URL is required"}
	}

	// Получаем userID из контекста
	userID := getUserIDFromContext(ctx)

	// Подготавливаем данные для пакетного сохранения
	pairs := make(map[string]string)
	results := make([]*proto.BatchURLResult, 0, len(req.Items))

	for _, item := range req.Items {
		// Валидация URL
		if item.OriginalUrl == "" {
			return nil, &InvalidArgumentError{Message: "URL is required for correlation ID: " + item.CorrelationId}
		}

		parsedURL, err := url.Parse(item.OriginalUrl)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return nil, &InvalidArgumentError{Message: "Invalid URL format for correlation ID: " + item.CorrelationId}
		}

		// Проверяем, существует ли уже такой URL
		if existingID, found := s.store.FindByOriginal(item.OriginalUrl); found {
			results = append(results, &proto.BatchURLResult{
				CorrelationId: item.CorrelationId,
				ShortUrl:      s.config.BaseURL + "/" + existingID,
			})
			continue
		}

		// Генерируем новый сокращенный ID
		shortID, err := utils.GenerateShortID()
		if err != nil {
			return nil, &InternalError{Message: "Failed to generate short ID for correlation ID: " + item.CorrelationId}
		}
		pairs[shortID] = item.OriginalUrl

		results = append(results, &proto.BatchURLResult{
			CorrelationId: item.CorrelationId,
			ShortUrl:      s.config.BaseURL + "/" + shortID,
		})
	}

	// Сохраняем все URL пакетом
	if err := s.store.SaveBatch(pairs, userID); err != nil {
		s.logger.Error("Failed to save batch URLs", zap.Error(err))
		return nil, &InternalError{Message: "Failed to save URLs"}
	}

	s.logger.Info("Batch URLs shortened", zap.Int("count", len(req.Items)))

	return &proto.BatchShortenURLResponse{
		Items: results,
	}, nil
}
