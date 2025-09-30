package grpc

import (
	"testing"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	svc := service.NewService(cfg, store, logger)

	server := NewServer(svc, logger)

	assert.NotNil(t, server)
	assert.Equal(t, svc, server.service)
	assert.Equal(t, logger, server.logger)
}
