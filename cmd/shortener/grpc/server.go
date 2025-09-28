package grpc

import (
	"uno/api/proto"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/storage"

	"go.uber.org/zap"
)

// Server реализует gRPC сервис для сокращения URL
type Server struct {
	proto.UnimplementedShortenerServiceServer
	config *config.Config
	store  storage.Storage
	logger *zap.Logger
}

// NewServer создает новый экземпляр gRPC сервера
func NewServer(cfg *config.Config, store storage.Storage, logger *zap.Logger) *Server {
	return &Server{
		config: cfg,
		store:  store,
		logger: logger,
	}
}
