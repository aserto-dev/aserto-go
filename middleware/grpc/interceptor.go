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
	Policy           = middleware.Policy
	AuthorizerClient = authz.AuthorizerClient
)

// Middleware implements gRPC unary and stream server interceptors that perform authorization.
// It provides configuration options to control how authorization parameters like caller identity, and
// policy path are extracted from incoming RPC calls.
type Middleware struct {
	Identity *IdentityBuilder

	client AuthorizerClient
	// builder internal.IsRequestBuilder
	// request authz.IsRequest

	policy api.PolicyContext

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
func New(client AuthorizerClient, policy Policy) *Middleware {
	policyMapper := methodPolicyMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &Middleware{
		client:         client,
		Identity:       &IdentityBuilder{},
		policy:         *internal.DefaultPolicyContext(policy),
		policyMapper:   policyMapper,
		resourceMapper: noResourceMapper,
	}
}

// WithPolicyPathMapper takes a custom StringMapper for extracting the authorization policy path form
// incoming message.
func (m *Middleware) WithPolicyPathMapper(mapper StringMapper) *Middleware {
	m.policyMapper = mapper
	return m
}

func (m *Middleware) WithResourceFromFields(fields ...string) *Middleware {
	m.resourceMapper = messageResourceMapper(fields...)
	return m
}

// WithResourceMapper takes a custom StructMapper for extracting the authorization resource context from
// incoming messages.
func (m *Middleware) WithResourceMapper(mapper StructMapper) *Middleware {
	m.resourceMapper = mapper
	return m
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
func (m *Middleware) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := m.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
func (m *Middleware) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()

		if err := m.authorize(ctx, nil); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (m *Middleware) authorize(ctx context.Context, req interface{}) error {
	if m.policyMapper != nil {
		m.policy.Path = m.policyMapper(ctx, req)
	}

	resp, err := m.client.Is(
		ctx,
		&authz.IsRequest{
			IdentityContext: m.Identity.Build(ctx, req),
			PolicyContext:   &m.policy,
			ResourceContext: m.resourceMapper(ctx, req),
		},
	)
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

func methodPolicyMapper(policyRoot string) StringMapper {
	return func(ctx context.Context, _ interface{}) string {
		method, _ := grpc.Method(ctx)
		path := internal.ToPolicyPath(method)

		if policyRoot == "" {
			return path
		}

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
