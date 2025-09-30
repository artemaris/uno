package grpc

import (
	"context"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/service"
	"uno/cmd/shortener/storage"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetStats_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	svc := service.NewService(cfg, store, logger)
	server := NewServer(svc, logger)

	// Настраиваем мок
	store.On("GetStats").Return(storage.Stats{URLs: 42, Users: 15}, nil)

	req := &proto.GetStatsRequest{}
	resp, err := server.GetStats(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(42), resp.Urls)
	assert.Equal(t, int32(15), resp.Users)

	store.AssertExpectations(t)
}
