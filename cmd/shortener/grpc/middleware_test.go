package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestWithUserIDMiddleware(t *testing.T) {
	// Создаем тестовый обработчик
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		userID := getUserIDFromContext(ctx)
		return userID, nil
	}

	// Создаем middleware
	middleware := WithUserIDMiddleware()

	// Тестируем без userID в метаданных
	ctx := context.Background()
	req := "test request"
	info := &grpc.UnaryServerInfo{}

	resp, err := middleware(ctx, req, info, handler)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.(string)) // Должен быть сгенерирован userID
}

func TestWithUserIDMiddleware_WithMetadata(t *testing.T) {
	// Создаем тестовый обработчик
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		userID := getUserIDFromContext(ctx)
		return userID, nil
	}

	// Создаем middleware
	middleware := WithUserIDMiddleware()

	// Тестируем с userID в метаданных
	md := metadata.New(map[string]string{
		"user-id": "test-user-123",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	req := "test request"
	info := &grpc.UnaryServerInfo{}

	resp, err := middleware(ctx, req, info, handler)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-user-123", resp.(string))
}

func TestGetUserIDFromMetadata(t *testing.T) {
	// Тест без метаданных
	ctx := context.Background()
	userID := getUserIDFromMetadata(ctx)
	assert.Empty(t, userID)

	// Тест с метаданными без user-id
	md := metadata.New(map[string]string{
		"other-header": "value",
	})
	ctx = metadata.NewIncomingContext(context.Background(), md)
	userID = getUserIDFromMetadata(ctx)
	assert.Empty(t, userID)

	// Тест с user-id в метаданных
	md = metadata.New(map[string]string{
		"user-id": "test-user-456",
	})
	ctx = metadata.NewIncomingContext(context.Background(), md)
	userID = getUserIDFromMetadata(ctx)
	assert.Equal(t, "test-user-456", userID)
}

func TestGenerateUserID(t *testing.T) {
	// Генерируем несколько userID и проверяем, что они уникальны
	userID1 := generateUserID()
	userID2 := generateUserID()

	assert.NotEmpty(t, userID1)
	assert.NotEmpty(t, userID2)
	assert.NotEqual(t, userID1, userID2)
	assert.Len(t, userID1, 32) // 16 байт в hex = 32 символа
	assert.Len(t, userID2, 32)
}

func TestGetUserIDFromContext(t *testing.T) {
	// Тест без userID в контексте
	ctx := context.Background()
	userID := getUserIDFromContext(ctx)
	assert.Empty(t, userID)

	// Тест с userID в контексте
	ctx = context.WithValue(ctx, userIDKey, "test-user-789")
	userID = getUserIDFromContext(ctx)
	assert.Equal(t, "test-user-789", userID)

	// Тест с неправильным типом в контексте
	ctx = context.WithValue(ctx, userIDKey, 123)
	userID = getUserIDFromContext(ctx)
	assert.Empty(t, userID)
}
