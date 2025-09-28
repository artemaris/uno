package grpc

import (
	"context"
	"uno/api/proto"
)

// Ping обрабатывает запрос проверки доступности сервиса
func (s *Server) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}
