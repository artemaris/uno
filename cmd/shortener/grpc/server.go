package grpc

import (
	"uno/api/proto"
	"uno/cmd/shortener/service"

	"go.uber.org/zap"
)

// Server реализует gRPC сервис для сокращения URL
type Server struct {
	proto.UnimplementedShortenerServiceServer
	service *service.Service
	logger  *zap.Logger
}

// NewServer создает новый экземпляр gRPC сервера
func NewServer(svc *service.Service, logger *zap.Logger) *Server {
	return &Server{
		service: svc,
		logger:  logger,
	}
}
