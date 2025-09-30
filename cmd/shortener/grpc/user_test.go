package grpc

import (
	"context"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/service"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetUserURLs_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	svc := service.NewService(cfg, store, logger)
	server := NewServer(svc, logger)

	// Настраиваем мок
	store.On("GetUserURLs", "test-user").Return([]models.UserURL{}, nil)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.GetUserURLsRequest{}
	resp, err := server.GetUserURLs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Urls, 0)

	store.AssertExpectations(t)
}

func TestDeleteUserURLs_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	svc := service.NewService(cfg, store, logger)
	server := NewServer(svc, logger)

	// Настраиваем мок
	store.On("DeleteURLs", "test-user", []string{"abc123", "def456"}).Return(nil)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.DeleteUserURLsRequest{
		Urls: []string{
			"http://localhost:8080/abc123",
			"http://localhost:8080/def456",
		},
	}
	resp, err := server.DeleteUserURLs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	store.AssertExpectations(t)
}
