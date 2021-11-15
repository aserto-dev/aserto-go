package mock

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

var errNotImplemented = errors.New("not implemented")

// Mock grpc.ServerStream.
type ServerStream struct {
	Ctx context.Context
}

func (s *ServerStream) SetHeader(metadata.MD) error {
	return errNotImplemented
}

func (s *ServerStream) SendHeader(metadata.MD) error {
	return errNotImplemented
}

func (s *ServerStream) SetTrailer(metadata.MD) {
}

func (s *ServerStream) Context() context.Context {
	return s.Ctx
}

func (s *ServerStream) SendMsg(m interface{}) error {
	return errNotImplemented
}

func (s *ServerStream) RecvMsg(m interface{}) error {
	return errNotImplemented
}
