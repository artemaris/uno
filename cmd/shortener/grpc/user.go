package grpc

import (
	"context"
	"uno/api/proto"

	"go.uber.org/zap"
)

// GetUserURLs обрабатывает запрос на получение URL пользователя
func (s *Server) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	// Получаем userID из контекста
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return &proto.GetUserURLsResponse{
			Urls: []*proto.UserURLItem{},
		}, nil
	}

	// Получаем URL пользователя
	userURLs, err := s.store.GetUserURLs(userID)
	if err != nil {
		s.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, &InternalError{Message: "Failed to get user URLs"}
	}

	// Конвертируем в gRPC формат
	items := make([]*proto.UserURLItem, 0, len(userURLs))
	for _, userURL := range userURLs {
		items = append(items, &proto.UserURLItem{
			ShortUrl:    s.config.BaseURL + "/" + userURL.ShortURL,
			OriginalUrl: userURL.OriginalURL,
			Deleted:     userURL.Deleted,
		})
	}

	s.logger.Info("User URLs retrieved", zap.String("userID", userID), zap.Int("count", len(items)))

	return &proto.GetUserURLsResponse{
		Urls: items,
	}, nil
}

// DeleteUserURLs обрабатывает запрос на удаление URL пользователя
func (s *Server) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*proto.DeleteUserURLsResponse, error) {
	// Получаем userID из контекста
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		return &proto.DeleteUserURLsResponse{}, nil
	}

	if len(req.Urls) == 0 {
		return &proto.DeleteUserURLsResponse{}, nil
	}

	// Извлекаем ID из полных URL
	ids := make([]string, 0, len(req.Urls))
	for _, url := range req.Urls {
		// Извлекаем ID из URL (убираем базовый URL)
		if len(url) > len(s.config.BaseURL)+1 {
			id := url[len(s.config.BaseURL)+1:]
			ids = append(ids, id)
		}
	}

	// Удаляем URL
	if err := s.store.DeleteURLs(userID, ids); err != nil {
		s.logger.Error("Failed to delete user URLs", zap.Error(err))
		return nil, &InternalError{Message: "Failed to delete URLs"}
	}

	s.logger.Info("User URLs deleted", zap.String("userID", userID), zap.Int("count", len(ids)))

	return &proto.DeleteUserURLsResponse{}, nil
}
