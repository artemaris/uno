package grpc

import (
	"context"
	"uno/api/proto"
)

// GetUserURLs обрабатывает запрос на получение URL пользователя
func (s *Server) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	userID := getUserIDFromContext(ctx)

	userURLs, err := s.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Конвертируем в gRPC формат
	items := make([]*proto.UserURLItem, 0, len(userURLs))
	for _, userURL := range userURLs {
		items = append(items, &proto.UserURLItem{
			ShortUrl:    s.service.GetBaseURL() + "/" + userURL.ShortURL,
			OriginalUrl: userURL.OriginalURL,
			Deleted:     userURL.Deleted,
		})
	}

	return &proto.GetUserURLsResponse{
		Urls: items,
	}, nil
}

// DeleteUserURLs обрабатывает запрос на удаление URL пользователя
func (s *Server) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*proto.DeleteUserURLsResponse, error) {
	userID := getUserIDFromContext(ctx)

	err := s.service.DeleteUserURLs(ctx, userID, req.Urls)
	if err != nil {
		return nil, err
	}

	return &proto.DeleteUserURLsResponse{}, nil
}
