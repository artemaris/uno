package grpc

import (
	"context"
	"errors"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/models"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetUserURLs_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок
	userURLs := []models.UserURL{
		{
			ShortURL:    "abc123",
			OriginalURL: "https://example.com",
			Deleted:     false,
		},
		{
			ShortURL:    "def456",
			OriginalURL: "https://test.com",
			Deleted:     true,
		},
	}
	store.On("GetUserURLs", "test-user").Return(userURLs, nil)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.GetUserURLsRequest{}
	resp, err := server.GetUserURLs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Urls, 2)
	assert.Equal(t, "http://localhost:8080/abc123", resp.Urls[0].ShortUrl)
	assert.Equal(t, "https://example.com", resp.Urls[0].OriginalUrl)
	assert.False(t, resp.Urls[0].Deleted)
	assert.Equal(t, "http://localhost:8080/def456", resp.Urls[1].ShortUrl)
	assert.Equal(t, "https://test.com", resp.Urls[1].OriginalUrl)
	assert.True(t, resp.Urls[1].Deleted)

	store.AssertExpectations(t)
}

func TestGetUserURLs_NoUserID(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.Background() // без user-id

	req := &proto.GetUserURLsRequest{}
	resp, err := server.GetUserURLs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Urls, 0)
}

func TestGetUserURLs_StoreError(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - ошибка при получении URL
	store.On("GetUserURLs", "test-user").Return([]models.UserURL(nil), errors.New("store error"))

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.GetUserURLsRequest{}
	resp, err := server.GetUserURLs(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Failed to get user URLs")

	store.AssertExpectations(t)
}

func TestDeleteUserURLs_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

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

func TestDeleteUserURLs_NoUserID(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.Background() // без user-id

	req := &proto.DeleteUserURLsRequest{
		Urls: []string{"http://localhost:8080/abc123"},
	}
	resp, err := server.DeleteUserURLs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteUserURLs_EmptyURLs(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.DeleteUserURLsRequest{
		Urls: []string{},
	}
	resp, err := server.DeleteUserURLs(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteUserURLs_StoreError(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - ошибка при удалении
	store.On("DeleteURLs", "test-user", []string{"abc123"}).Return(errors.New("delete error"))

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.DeleteUserURLsRequest{
		Urls: []string{"http://localhost:8080/abc123"},
	}
	resp, err := server.DeleteUserURLs(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Failed to delete URLs")

	store.AssertExpectations(t)
}
