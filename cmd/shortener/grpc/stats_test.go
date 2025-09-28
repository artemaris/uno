package grpc

import (
	"context"
	"errors"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"
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
	server := NewServer(cfg, store, logger)

	// Настраиваем мок
	stats := storage.Stats{
		URLs:  42,
		Users: 15,
	}
	store.On("GetStats").Return(stats, nil)

	req := &proto.GetStatsRequest{}
	resp, err := server.GetStats(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(42), resp.Urls)
	assert.Equal(t, int32(15), resp.Users)

	store.AssertExpectations(t)
}

func TestGetStats_StoreError(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - ошибка при получении статистики
	store.On("GetStats").Return(storage.Stats{}, errors.New("store error"))

	req := &proto.GetStatsRequest{}
	resp, err := server.GetStats(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Failed to get stats")

	store.AssertExpectations(t)
}
