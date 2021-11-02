package authorizer

import (
	"context"
	"strings"

	"github.com/aserto-dev/aserto-go/internal"
	"github.com/aserto-dev/aserto-go/internal/grpcc"
	"github.com/aserto-dev/aserto-go/internal/pbutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	api "github.com/aserto-dev/go-grpc/aserto/api/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"

	"github.com/pkg/errors"
)

// Client provides access to services only available usign gRPC.
type Client struct {
	conn      *grpcc.Connection
	Directory dir.DirectoryClient
	Policy    policy.PolicyClient
	Info      info.InfoClient
}

// NewClient creates a Client with the specified connection options.
func NewClient(ctx context.Context, opts ...internal.ConnectionOption) (*Client, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:      connection,
		Directory: dir.NewDirectoryClient(connection.Conn),
		Policy:    policy.NewPolicyClient(connection.Conn),
		Info:      info.NewInfoClient(connection.Conn),
	}, err
}

// NewAuthorizer creates a new AuthorizerClient using the specified connection options.
func NewAuthorizer(ctx context.Context, opts ...internal.ConnectionOption) (authz.AuthorizerClient, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return authz.NewAuthorizerClient(connection.Conn), nil
}

type InterceptorOptions struct {
	identity IdentityRetriever
	policy   PolicyRetriever
	resource ResourceRetriever
}

type InterceptorOption func(*InterceptorOptions)

type IdentityRetriever func(context.Context, interface{}) *api.IdentityContext

type PolicyRetriever func(context.Context, interface{}) *api.PolicyContext

type ResourceRetriever func(context.Context, interface{}) *structpb.Struct

func UnaryServerInterceptor(a authz.AuthorizerClient, opts ...InterceptorOption) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		return handler(ctx, req)
	}
}

func WithContextMetadataIdentity(key string, identityType api.IdentityType) InterceptorOption {
	return func(opts *InterceptorOptions) {
		opts.identity = func(ctx context.Context, _ interface{}) *api.IdentityContext {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				id := md.Get(key)
				if len(id) > 0 {
					return &api.IdentityContext{
						Type:     identityType,
						Identity: id[0],
					}
				}
			}

			return &api.IdentityContext{
				Type: api.IdentityType_IDENTITY_TYPE_NONE,
			}
		}
	}
}

func WithContextValueIdentity(value string, identityType api.IdentityType) InterceptorOption {
	return func(opts *InterceptorOptions) {
		opts.identity = func(ctx context.Context, _ interface{}) *api.IdentityContext {
			identity, ok := ctx.Value(value).(string)
			if !ok {
				return &api.IdentityContext{
					Type: api.IdentityType_IDENTITY_TYPE_NONE,
				}
			}
			return &api.IdentityContext{
				Type:     identityType,
				Identity: identity,
			}
		}
	}
}

func WithIdentityRetrieverFunc(retriever IdentityRetriever) InterceptorOption {
	return func(opts *InterceptorOptions) {
		opts.identity = retriever
	}
}

func WithPolicyFromMethod(policyID string) InterceptorOption {
	return func(opts *InterceptorOptions) {
		opts.policy = func(ctx context.Context, _ interface{}) *api.PolicyContext {
			method, _ := grpc.Method(ctx)
			policyPath := strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")
			return &api.PolicyContext{
				Id:   policyID,
				Path: policyPath,
			}
		}
	}
}

func WithResourceFromMessage(fields ...string) InterceptorOption {
	return func(opts *InterceptorOptions) {
		opts.resource = func(ctx context.Context, req interface{}) *structpb.Struct {
			resource, _ := pbutil.Select(req.(protoreflect.ProtoMessage), fields...)
			return resource
		}
	}
}
