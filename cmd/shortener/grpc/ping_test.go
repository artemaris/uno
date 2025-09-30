package grpc

import (
	"context"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPing_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	svc := service.NewService(cfg, store, logger)
	server := NewServer(svc, logger)

	req := &proto.PingRequest{}
	resp, err := server.Ping(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
