package grpc

import (
	"context"
	"uno/api/proto"
)

// GetStats обрабатывает запрос на получение статистики
func (s *Server) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	stats, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}
