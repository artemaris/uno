package grpc

import (
	"context"
	"errors"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestBatchShortenURL_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок
	store.On("FindByOriginal", "https://google.com").Return("", false)
	store.On("FindByOriginal", "https://github.com").Return("", false)
	store.On("SaveBatch", mock.AnythingOfType("map[string]string"), "test-user").Return(nil)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://google.com",
			},
			{
				CorrelationId: "2",
				OriginalUrl:   "https://github.com",
			},
		},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "1", resp.Items[0].CorrelationId)
	assert.Contains(t, resp.Items[0].ShortUrl, "http://localhost:8080/")
	assert.Equal(t, "2", resp.Items[1].CorrelationId)
	assert.Contains(t, resp.Items[1].ShortUrl, "http://localhost:8080/")

	store.AssertExpectations(t)
}

func TestBatchShortenURL_EmptyRequest(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "At least one URL is required")
}

func TestBatchShortenURL_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "invalid-url",
			},
		},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid URL format")
}

func TestBatchShortenURL_EmptyURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "",
			},
		},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "URL is required")
}

func TestBatchShortenURL_ExistingURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - один URL уже существует
	store.On("FindByOriginal", "https://google.com").Return("existing123", true)
	store.On("FindByOriginal", "https://github.com").Return("", false)
	store.On("SaveBatch", mock.AnythingOfType("map[string]string"), "test-user").Return(nil)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://google.com",
			},
			{
				CorrelationId: "2",
				OriginalUrl:   "https://github.com",
			},
		},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "1", resp.Items[0].CorrelationId)
	assert.Equal(t, "http://localhost:8080/existing123", resp.Items[0].ShortUrl)
	assert.Equal(t, "2", resp.Items[1].CorrelationId)
	assert.Contains(t, resp.Items[1].ShortUrl, "http://localhost:8080/")

	store.AssertExpectations(t)
}

func TestBatchShortenURL_SaveError(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - ошибка при сохранении
	store.On("FindByOriginal", "https://google.com").Return("", false)
	store.On("SaveBatch", mock.AnythingOfType("map[string]string"), "test-user").Return(errors.New("save error"))

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://google.com",
			},
		},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Failed to save URLs")

	store.AssertExpectations(t)
}
