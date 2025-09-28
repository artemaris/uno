package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInvalidArgumentError(t *testing.T) {
	err := &InvalidArgumentError{Message: "test error"}

	// Проверяем Error()
	assert.Equal(t, "test error", err.Error())

	// Проверяем GRPCStatus()
	grpcStatus := err.GRPCStatus()
	assert.NotNil(t, grpcStatus)
	assert.Equal(t, codes.InvalidArgument, grpcStatus.Code())
	assert.Equal(t, "test error", grpcStatus.Message())
}

func TestInternalError(t *testing.T) {
	err := &InternalError{Message: "internal error"}

	// Проверяем Error()
	assert.Equal(t, "internal error", err.Error())

	// Проверяем GRPCStatus()
	grpcStatus := err.GRPCStatus()
	assert.NotNil(t, grpcStatus)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Equal(t, "internal error", grpcStatus.Message())
}

func TestErrorTypes(t *testing.T) {
	// Проверяем, что ошибки реализуют интерфейс error
	var _ error = &InvalidArgumentError{}
	var _ error = &InternalError{}

	// Проверяем, что ошибки реализуют интерфейс grpcStatus
	var _ interface{ GRPCStatus() *status.Status } = &InvalidArgumentError{}
	var _ interface{ GRPCStatus() *status.Status } = &InternalError{}
}
