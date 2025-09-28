package grpc

import (
	"context"
	"testing"
	"uno/api/proto"
	"uno/cmd/shortener/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestShortenURL_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок
	store.On("FindByOriginal", "https://example.com").Return("", false)
	store.On("Save", mock.AnythingOfType("string"), "https://example.com", "test-user")

	// Создаем контекст с userID
	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	// Вызываем метод
	req := &proto.ShortenURLRequest{
		Url: "https://example.com",
	}
	resp, err := server.ShortenURL(ctx, req)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Result, "http://localhost:8080/")
	assert.Len(t, resp.Result, len("http://localhost:8080/")+8) // 8 символов для короткого ID

	store.AssertExpectations(t)
}

func TestShortenURL_ExistingURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - URL уже существует
	store.On("FindByOriginal", "https://example.com").Return("existing123", true)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.ShortenURLRequest{
		Url: "https://example.com",
	}
	resp, err := server.ShortenURL(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "http://localhost:8080/existing123", resp.Result)

	store.AssertExpectations(t)
}

func TestShortenURL_InvalidURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.ShortenURLRequest{
		Url: "invalid-url",
	}
	resp, err := server.ShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid URL format")
}

func TestShortenURL_EmptyURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	ctx := context.WithValue(context.Background(), userIDKey, "test-user")

	req := &proto.ShortenURLRequest{
		Url: "",
	}
	resp, err := server.ShortenURL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "URL is required")
}

func TestGetURL_Success(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок
	store.On("Get", "test123").Return("https://example.com", false, true)

	req := &proto.GetURLRequest{
		Id: "test123",
	}
	resp, err := server.GetURL(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "https://example.com", resp.Url)
	assert.False(t, resp.Deleted)
	assert.True(t, resp.Exists)

	store.AssertExpectations(t)
}

func TestGetURL_NotFound(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	// Настраиваем мок - URL не найден
	store.On("Get", "nonexistent").Return("", false, false)

	req := &proto.GetURLRequest{
		Id: "nonexistent",
	}
	resp, err := server.GetURL(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "", resp.Url)
	assert.False(t, resp.Deleted)
	assert.False(t, resp.Exists)

	store.AssertExpectations(t)
}

func TestGetURL_EmptyID(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://localhost:8080",
	}
	store := &MockStorage{}
	logger := zap.NewNop()
	server := NewServer(cfg, store, logger)

	req := &proto.GetURLRequest{
		Id: "",
	}
	resp, err := server.GetURL(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ID is required")
}
