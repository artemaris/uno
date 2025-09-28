package grpc

import (
	"context"
	"uno/api/proto"

	"go.uber.org/zap"
)

// GetStats обрабатывает запрос на получение статистики
func (s *Server) GetStats(ctx context.Context, req *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	// Получаем статистику из хранилища
	stats, err := s.store.GetStats()
	if err != nil {
		s.logger.Error("Failed to get stats", zap.Error(err))
		return nil, &InternalError{Message: "Failed to get stats"}
	}

	s.logger.Info("Stats retrieved", zap.Int("urls", stats.URLs), zap.Int("users", stats.Users))

	return &proto.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}
