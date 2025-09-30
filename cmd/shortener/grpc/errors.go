package grpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InvalidArgumentError представляет ошибку неверного аргумента
type InvalidArgumentError struct {
	Message string
}

func (e *InvalidArgumentError) Error() string {
	return e.Message
}

func (e *InvalidArgumentError) GRPCStatus() *status.Status {
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
