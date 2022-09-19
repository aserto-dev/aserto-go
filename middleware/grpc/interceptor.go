/*
Package grpc provides authorization middleware for gRPC servers.

The middleware intercepts incoming requests/streams and calls the Aserto authorizer service to
determine if access should be granted or denied.
*/
package grpc

import (
	"context"
	"fmt"

	"github.com/aserto-dev/aserto-go/middleware"
	"github.com/aserto-dev/aserto-go/middleware/grpc/internal/pbutil"
	"github.com/aserto-dev/aserto-go/middleware/internal"
	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	"github.com/aserto-dev/go-grpc/aserto/api/v1"
	"github.com/aserto-dev/go-utils/cerr"
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

	client          AuthorizerClient
	policy          api.PolicyContext
	policyMapper    StringMapper
	resourceMappers []ResourceMapper
}

type (
	// StringMapper functions are used to extract string values from incoming messages.
	// They are used to define identity and policy mappers.
	StringMapper func(context.Context, interface{}) string

	// ResourceMapper functions are used to extract structured data from incoming message.
	ResourceMapper func(context.Context, interface{}, map[string]interface{})
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
		client:          client,
		Identity:        (&IdentityBuilder{}).FromMetadata("authorization"),
		policy:          *internal.DefaultPolicyContext(policy),
		policyMapper:    policyMapper,
		resourceMappers: []ResourceMapper{},
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
	m.resourceMappers = append(m.resourceMappers, messageResourceMapper(map[string][]string{}, fields...))
	return m
}

/*
WithResourceFromMessageByPath behaves similarly to `WithResourceFromFields` but allows specifying different sets
of fields for different method paths.

Example:

  middleware.WithResourceFromMessageByPath(
	  "/example.ExampleService/Method1": []string{"field1", "field2"},
	  "/example.ExampleService/Method2": []string{"field1", "field2"},
	  "id", "name",
  )

When Method1 or Method2 are called, the middleware constructs in a authorization resource with the following structure:

  {
	  "field1": <value from message>,
	  "field2": <value from message>
  }

For all other methods, the middleware constructs in a authorization resource with the following structure:

  {
	  "id": <value from message>,
	  "name": <value from message>
  }
*/
func (m *Middleware) WithResourceFromMessageByPath(fieldsByPath map[string][]string, defaults ...string) *Middleware {
	m.resourceMappers = append(m.resourceMappers, messageResourceMapper(fieldsByPath, defaults...))
	return m
}

/*
WithResourceFromContextValue instructs the middleware to read the specified value from the incoming request
context and add it to the authorization resource context.

Example:

  middleware.WithResourceFromContextValue("account_id", "account")

In each incoming request, the middlware reads the value of the "account_id" key from the request context and
adds its value to the "account" field in the authorization resource context.
*/
func (m *Middleware) WithResourceFromContextValue(ctxKey interface{}, field string) *Middleware {
	m.resourceMappers = append(m.resourceMappers, contextValueResourceMapper(ctxKey, field))
	return m
}

// WithResourceMapper takes a custom StructMapper for extracting the authorization resource context from
// incoming messages.
func (m *Middleware) WithResourceMapper(mapper ResourceMapper) *Middleware {
	m.resourceMappers = append(m.resourceMappers, mapper)
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

	resource, err := m.resourceContext(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to apply resource mapper")
	}

	resp, err := m.client.Is(
		ctx,
		&authz.IsRequest{
			IdentityContext: m.Identity.build(ctx, req),
			PolicyContext:   &m.policy,
			ResourceContext: resource,
		},
	)
	if err != nil {
		return errors.Wrap(err, "authorization call failed")
	}

	if len(resp.Decisions) == 0 {
		return cerr.ErrInvalidDecision
	}

	if !resp.Decisions[0].Is {
		return cerr.ErrAuthorizationFailed
	}

	return nil
}

func (m *Middleware) resourceContext(ctx context.Context, req interface{}) (*structpb.Struct, error) {
	res := map[string]interface{}{}
	for _, mapper := range m.resourceMappers {
		mapper(ctx, req, res)
	}

	return structpb.NewStruct(res)
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

func messageResourceMapper(fieldsByPath map[string][]string, defaults ...string) ResourceMapper {
	return func(ctx context.Context, req interface{}, res map[string]interface{}) {
		method, _ := grpc.Method(ctx)

		fields, ok := fieldsByPath[method]
		if !ok || len(fields) == 0 {
			fields = defaults
		}

		if len(fields) > 0 {
			resource, _ := pbutil.Select(req.(protoreflect.ProtoMessage), fields...)
			for k, v := range resource.AsMap() {
				res[k] = v
			}
		}
	}
}

func contextValueResourceMapper(ctxKey interface{}, field string) ResourceMapper {
	return func(ctx context.Context, _ interface{}, res map[string]interface{}) {
		if v := ctx.Value(ctxKey); v != nil {
			res[field] = v
		}
	}
}
