package grpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// WithUserIDMiddleware добавляет userID в контекст gRPC запроса
func WithUserIDMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Извлекаем userID из метаданных или генерируем новый
		userID := getUserIDFromMetadata(ctx)
		if userID == "" {
			userID = generateUserID()
		}

		// Добавляем userID в контекст
		ctx = context.WithValue(ctx, userIDKey, userID)

		return handler(ctx, req)
	}
}

// getUserIDFromMetadata извлекает userID из gRPC метаданных
func getUserIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	userIDs := md.Get("user-id")
	if len(userIDs) > 0 {
		return userIDs[0]
	}

	return ""
}

// generateUserID генерирует новый userID
func generateUserID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getUserIDFromContext извлекает userID из контекста
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}
