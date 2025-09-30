package grpc

import (
	"context"
	"uno/api/proto"
	"uno/cmd/shortener/service"
)

// BatchShortenURL обрабатывает запрос на пакетное сокращение URL
func (s *Server) BatchShortenURL(ctx context.Context, req *proto.BatchShortenURLRequest) (*proto.BatchShortenURLResponse, error) {
	userID := getUserIDFromContext(ctx)

	// Конвертируем в формат сервиса
	items := make([]service.BatchItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, service.BatchItem{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		})
	}

	results, err := s.service.BatchShortenURL(ctx, items, userID)
	if err != nil {
		return nil, err
	}

	// Конвертируем в формат gRPC
	grpcResults := make([]*proto.BatchURLResult, 0, len(results))
	for _, result := range results {
		grpcResults = append(grpcResults, &proto.BatchURLResult{
			CorrelationId: result.CorrelationID,
			ShortUrl:      result.ShortURL,
		})
	}

	return &proto.BatchShortenURLResponse{
		Items: grpcResults,
	}, nil
}
