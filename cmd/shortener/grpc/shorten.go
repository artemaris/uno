package grpc

import (
	"context"
	"uno/api/proto"
)

// ShortenURL обрабатывает запрос на сокращение URL
func (s *Server) ShortenURL(ctx context.Context, req *proto.ShortenURLRequest) (*proto.ShortenURLResponse, error) {
	userID := getUserIDFromContext(ctx)

	result, err := s.service.ShortenURL(ctx, req.Url, userID)
	if err != nil {
		return nil, err
	}

	return &proto.ShortenURLResponse{
		Result: result,
	}, nil
}

// GetURL обрабатывает запрос на получение оригинального URL
func (s *Server) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	originalURL, deleted, exists, err := s.service.GetURL(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &proto.GetURLResponse{
		Url:     originalURL,
		Deleted: deleted,
		Exists:  exists,
	}, nil
}
