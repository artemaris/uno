package grpc

import (
	"context"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/service"

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
	svc := service.NewService(cfg, store, logger)
	server := NewServer(svc, logger)

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
	svc := service.NewService(cfg, store, logger)
	server := NewServer(svc, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{},
	}
	resp, err := server.BatchShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "At least one URL is required")
}
