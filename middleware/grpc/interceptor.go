package grpc

import (
	"context"
	"fmt"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/grpc/internal/pbutil"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	Config           = middleware.Config
	AuthorizerClient = authz.AuthorizerClient
)

// ServerInterceptor implements gRPC unary and stream server interceptors that perform authorization.
// It provides configuration options to control how authorization parameters like caller identity, and
// policy path are extracted from incoming RPC calls.
type ServerInterceptor struct {
	Identity *IdentityBuilder

	client AuthorizerClient
	// builder internal.IsRequestBuilder
	request authz.IsRequest

	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (

	// StringMapper functions are used to extract string values from incoming messages.
	// They are used to define identity and policy mappers.
	StringMapper func(context.Context, interface{}) string

	// StructMapper functions are used to extract structured data from incoming message.
	// The optional resource mapper is a StructMapper.
	StructMapper func(context.Context, interface{}) *structpb.Struct
)

// NewServerInterceptor returns a new ServerInterceptor from the specified authorizer client and configuration.
func New(client AuthorizerClient, conf Config) *ServerInterceptor {
	return &ServerInterceptor{
		Identity: &IdentityBuilder{},
		client:   client,
		request: authz.IsRequest{
			IdentityContext: &api.IdentityContext{},
			PolicyContext: &api.PolicyContext{
				Id:        conf.PolicyID,
				Decisions: []string{conf.Decision},
			},
		},
		policyMapper:   methodPolicyMapper(conf.PolicyRoot),
		resourceMapper: noResourceMapper,
	}
}

// WithPolicyPath sets a path in the authorization poilcy to be used for all incoming messages.
func (interceptor *ServerInterceptor) WithPolicyPath(path string) *ServerInterceptor {
	interceptor.policyMapper = policyPath(path)
	return interceptor
}

// WithPolicyPathMapper takes a custom StringMapper for extracting the authorization policy path form
// incoming message.
func (interceptor *ServerInterceptor) WithPolicyPathMapper(mapper StringMapper) *ServerInterceptor {
	interceptor.policyMapper = mapper
	return interceptor
}

func (interceptor *ServerInterceptor) WithResourceFromFields(fields ...string) *ServerInterceptor {
	interceptor.resourceMapper = messageResourceMapper(fields...)
	return interceptor
}

// WithResourceMapper takes a custom StructMapper for extracting the authorization resource context from
// incoming messages.
func (interceptor *ServerInterceptor) WithResourceMapper(mapper StructMapper) *ServerInterceptor {
	interceptor.resourceMapper = mapper
	return interceptor
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
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

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
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
	interceptor.request.PolicyContext.Path = interceptor.policyMapper(ctx, req)
	interceptor.request.IdentityContext = interceptor.Identity.Build(ctx, req)
	interceptor.request.ResourceContext = interceptor.resourceMapper(ctx, req)

	resp, err := interceptor.client.Is(ctx, &interceptor.request)
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

func policyPath(path string) StringMapper {
	return func(_ context.Context, _ interface{}) string {
		return path
	}
}

func methodPolicyMapper(policyRoot string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		method, _ := grpc.Method(ctx)
		return fmt.Sprintf("%s.%s", policyRoot, internal.ToPolicyPath(method))
	}
}

func noResourceMapper(ctx context.Context, req interface{}) *structpb.Struct {
	resource, _ := structpb.NewStruct(nil)
	return resource
}

func messageResourceMapper(fields ...string) StructMapper {
	return func(ctx context.Context, req interface{}) *structpb.Struct {
		resource, _ := pbutil.Select(req.(protoreflect.ProtoMessage), fields...)
		return resource
	}
}
