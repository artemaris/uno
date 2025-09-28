package grpc

import (
	"testing"
	"uno/cmd/shortener/config"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()

	server := NewServer(cfg, store, logger)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, store, server.store)
	assert.Equal(t, logger, server.logger)
}
