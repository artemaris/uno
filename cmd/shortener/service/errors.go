package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Message)
}

// InternalError представляет внутреннюю ошибку сервера
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}

func (e *InternalError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Message)
}
