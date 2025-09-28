package grpc

import (
	"context"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPing_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	req := &proto.PingRequest{}
	resp, err := server.Ping(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
