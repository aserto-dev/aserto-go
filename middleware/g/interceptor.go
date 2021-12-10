/*
Package grpc provides authorization middleware for gRPC servers.

The middleware intercepts incoming requests/streams and calls the Aserto authorizer service to
determine if access should be granted or denied.
*/
package g

import (
	"context"
	"fmt"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/g/internal/pbutil"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	Policy           = middleware.Policy
	AuthorizerClient = authz.AuthorizerClient
)

/*
Middleware implements unary and stream server interceptors that can be attached to gRPC servers.

To authorize incoming RPC calls, the middleware needs information about:

1. The user making the request.

2. The Aserto authorization policy to evaluate.

3. Optional, additional input data to the authorization policy.

The values for these parameters can be set globally or extracted dynamically from incoming messages.
*/
type Middleware struct {
	// Identity determines the caller identity used in authorization calls.
	Identity *IdentityBuilder

	client         AuthorizerClient
	policy         api.PolicyContext
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

// New creates middleware for the specified policy.
//
// The new middleware is created with default identity and policy path mapper.
// Those can be overridden using `Middleware.Identity` to specify the caller's identity, or using
// the middleware's ".With...()" functions to set policy path and resource mappers.
func New(client AuthorizerClient, policy Policy) *Middleware {
	policyMapper := methodPolicyMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &Middleware{
		client:         client,
		Identity:       (&IdentityBuilder{}).FromMetadata("authorization"),
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

/*
WithResourceFromFields instructs the middleware to select the specified fields from incoming messages and
use them as the resource in authorization calls. Fields are expressed as a field mask.

Note: Protobuf message fields are identified using their JSON names.

Example:

  middleware.WithResourceFromFields("product.type", "address")

This call would result in an authorization resource with the following structure:

  {
	  "product": {
		  "type": <value from message>
	  },
	  "address": <value from message>
  }

If the value of "address" is itself a message, all of its fields are included.
*/
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
			IdentityContext: m.Identity.build(ctx, req),
			PolicyContext:   &m.policy,
			ResourceContext: m.resourceMapper(ctx, req),
		},
	)
	if err != nil {
		return errors.Wrap(err, "authorization call failed")
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
