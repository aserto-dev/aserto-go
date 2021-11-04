package middleware

import (
	"context"
	"errors"
	"fmt"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"google.golang.org/grpc"
)

var ErrNoDecision = errors.New("authorizer returned no decisions")

type ServerInterceptor struct {
	client  authz.AuthorizerClient
	options *InterceptorOptions
}

func NewServerInterceptor(client authz.AuthorizerClient, opts ...InterceptorOption) (*ServerInterceptor, error) {
	options, err := NewInterceptorOptions(opts...)
	if err != nil {
		return nil, err
	}

	return &ServerInterceptor{client: client, options: options}, nil
}

func (interceptor *ServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := interceptor.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (interceptor *ServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		if err := interceptor.authorize(ctx, nil); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (interceptor *ServerInterceptor) authorize(ctx context.Context, req interface{}) error {
	resp, err := interceptor.client.Is(ctx, interceptor.options.IsRequest(ctx, req))
	if err != nil {
		return fmt.Errorf("authorization call failed: %w", err)
	}

	if len(resp.Decisions) == 0 {
		return ErrNoDecision
	}

	if !resp.Decisions[0].Is {
		return interceptor.options.unauthorizedError
	}

	return nil
}
