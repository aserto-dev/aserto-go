package grpcmw

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/aserto-go/internal/pbutil"
	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type Config = middleware.Config

// ServerInterceptor implements gRPC unary and stream server interceptors that perform authorization.
// It provides configuration options to control how authorization parameters like caller identity, and
// policy path are extracted from incoming RPC calls.
type ServerInterceptor struct {
	client  authz.AuthorizerClient
	builder internal.IsRequestBuilder

	identityMapper StringMapper
	policyMapper   StringMapper
	resourceMapper StructMapper
}

type (
	// StringMapper functions are used to extract string values like identity and policy path from incoming messages.
	StringMapper func(context.Context, interface{}) string

	// StructMapper functions are used to extract a resource context structure from incoming messages.
	StructMapper func(context.Context, interface{}) *structpb.Struct
)

// NewServerInterceptor returns a new ServerInterceptor from the specified authorizer client and configuration.
func NewServerInterceptor(client authz.AuthorizerClient, conf Config) *ServerInterceptor {
	return &ServerInterceptor{
		client:         client,
		builder:        internal.IsRequestBuilder{Config: conf},
		identityMapper: noIdentityMapper,
		policyMapper:   methodPolicyMapper(conf.PolicyRoot),
		resourceMapper: noResourceMapper,
	}
}

// WithIdentityFromMetadata extracts caller identity from a metadata field in the incoming message.
func (interceptor *ServerInterceptor) WithIdentityFromMetadata(field string) *ServerInterceptor {
	interceptor.identityMapper = contextMetadataIdentityMapper(field)
	return interceptor
}

// WithIdentityFromContextValue extracts caller identity from a context value in the incoming message.
func (interceptor *ServerInterceptor) WithIdentityFromContextValue(value string) *ServerInterceptor {
	interceptor.identityMapper = contextValueIdentityMapper(value)
	return interceptor
}

// WithIdentityMapper takes a custom StringMapper for extracting caller identity from incoming messages.
func (interceptor *ServerInterceptor) WithIdentityMapper(mapper StringMapper) *ServerInterceptor {
	interceptor.identityMapper = mapper
	return interceptor
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

func noIdentityMapper(_ context.Context, _ interface{}) string {
	return ""
}

func contextMetadataIdentityMapper(key string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			id := md.Get(key)
			if len(id) > 0 {
				return id[0]
			}
		}

		return ""
	}
}

func contextValueIdentityMapper(value string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		identity, ok := ctx.Value(value).(string)
		if ok {
			return identity
		}

		return ""
	}
}

func policyPath(path string) StringMapper {
	return func(_ context.Context, _ interface{}) string {
		return path
	}
}

func methodPolicyMapper(policyRoot string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		method, _ := grpc.Method(ctx)
		return fmt.Sprintf("%s.%s", policyRoot, strings.ReplaceAll(strings.Trim(method, "/"), "/", "."))
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
