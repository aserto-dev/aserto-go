package grpcmw

import (
	"context"
	"fmt"

	"github.com/aserto-dev/aserto-go/middleware"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type Config = middleware.Config

type ServerInterceptor struct {
	client  authz.AuthorizerClient
	builder middleware.IsRequestBuilder

	identityMapper StringMapper
	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	StringMapper func(context.Context, interface{}) string
	StructMapper func(context.Context, interface{}) *structpb.Struct
)

func NewServerInterceptor(client authz.AuthorizerClient, conf Config) *ServerInterceptor {
	return &ServerInterceptor{
		client:         client,
		builder:        middleware.IsRequestBuilder{Config: conf},
		identityMapper: NoIdentityMapper,
		resourceMapper: NoResourceMapper,
	}
}

// Generic mappers

func (interceptor *ServerInterceptor) WithIdentityFromMetadata(key string) *ServerInterceptor {
	interceptor.identityMapper = ContextMetadataIdentityMapper(key)
	return interceptor
}

func (interceptor *ServerInterceptor) WithIdentityMapper(mapper StringMapper) *ServerInterceptor {
	interceptor.identityMapper = mapper
	return interceptor
}

func (interceptor *ServerInterceptor) WithPolicyPathMapper(mapper StringMapper) *ServerInterceptor {
	interceptor.policyMapper = mapper
	return interceptor
}

func (interceptor *ServerInterceptor) WithPolicyPath(path string) *ServerInterceptor {
	interceptor.policyMapper = PolicyPath(path)
	return interceptor
}

func (interceptor *ServerInterceptor) WithResourceMapper(mapper StructMapper) *ServerInterceptor {
	interceptor.resourceMapper = mapper
	return interceptor
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
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()

		if err := interceptor.authorize(ctx, nil); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (interceptor *ServerInterceptor) authorize(ctx context.Context, req interface{}) error {
	interceptor.builder.SetPolicyPath(interceptor.policyMapper(ctx, req))
	interceptor.builder.SetIdentity(interceptor.identityMapper(ctx, req))
	interceptor.builder.SetResource(interceptor.resourceMapper(ctx, req))

	isRequest := interceptor.builder.Build()
	resp, err := interceptor.client.Is(ctx, isRequest)
	if err != nil {
		return fmt.Errorf("authorization call failed: %w", err)
	}

	if len(resp.Decisions) == 0 {
		return middleware.ErrNoDecision
	}

	if !resp.Decisions[0].Is {
		return middleware.ErrUnauthorized
	}

	return nil
}
