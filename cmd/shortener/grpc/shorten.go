package grpc

import (
	"context"
	"net/url"
	"uno/api/proto"
	"uno/cmd/shortener/utils"

	"go.uber.org/zap"
)

// ShortenURL обрабатывает запрос на сокращение URL
func (s *Server) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	// Валидация URL
	if req.Url == "" {
		return nil, &InvalidArgumentError{Message: "URL is required"}
	}

	parsedURL, err := url.Parse(req.Url)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, &InvalidArgumentError{Message: "Invalid URL format"}
	}

	// Проверяем, существует ли уже такой URL
	if existingID, found := s.store.FindByOriginal(req.Url); found {
		return &proto.ShortenURLResponse{
			Result: s.config.BaseURL + "/" + existingID,
		}, nil
	}

	// Генерируем новый сокращенный ID
	shortID, err := utils.GenerateShortID()
	if err != nil {
		s.logger.Error("Failed to generate short ID", zap.Error(err))
		return nil, &InternalError{Message: "Failed to generate short ID"}
	}

	// Получаем userID из контекста (будет добавлено позже)
	userID := getUserIDFromContext(ctx)

	// Сохраняем URL
	s.store.Save(shortID, req.Url, userID)

	s.logger.Info("URL shortened", zap.String("original", req.Url), zap.String("short", shortID))

	return &proto.ShortenURLResponse{
		Result: s.config.BaseURL + "/" + shortID,
	}, nil
}

// GetURL обрабатывает запрос на получение оригинального URL
func (s *Server) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	if req.Id == "" {
		return nil, &InvalidArgumentError{Message: "ID is required"}
	}

	originalURL, deleted, exists := s.store.Get(req.Id)
	if !exists {
		return &proto.GetURLResponse{
			Url:     "",
			Deleted: false,
			Exists:  false,
		}, nil
	}

	return &proto.GetURLResponse{
		Url:     originalURL,
		Deleted: deleted,
		Exists:  true,
	}, nil
}
