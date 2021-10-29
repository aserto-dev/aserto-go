package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/grpcc"
	"github.com/aserto-dev/aserto-go/pkg/internal"
	"google.golang.org/grpc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"

	"github.com/pkg/errors"
)

// Client gRPC connection.
type Client struct {
	conn       *grpcc.Connection
	Authorizer authz.AuthorizerClient
	Directory  dir.DirectoryClient
	Policy     policy.PolicyClient
	Info       info.InfoClient
}

// New creates an authorizer Client with the specified connection options.
func New(ctx context.Context, opts ...internal.ConnectionOption) (*Client, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:       connection,
		Authorizer: authz.NewAuthorizerClient(connection.Conn),
		Directory:  dir.NewDirectoryClient(connection.Conn),
		Policy:     policy.NewPolicyClient(connection.Conn),
		Info:       info.NewInfoClient(connection.Conn),
	}, err
}

// WithContext returns a wrapped context that includes tenant information.
func (client *Client) WithContext(ctx context.Context) context.Context {
	return client.conn.TenantID.WithContext(ctx)
}

func NewAuthorizer(ctx context.Context, opts ...internal.ConnectionOption) (authz.AuthorizerClient, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &contextualAuthorizerClient{
		authorizer: authz.NewAuthorizerClient((connection.Conn)),
		wrappers:   []internal.ContextWrapper{connection.TenantID},
	}, nil
}

type contextualAuthorizerClient struct {
	authorizer authz.AuthorizerClient
	wrappers   []internal.ContextWrapper
}

func (c *contextualAuthorizerClient) DecisionTree(
	ctx context.Context,
	in *authz.DecisionTreeRequest,
	opts ...grpc.CallOption,
) (*authz.DecisionTreeResponse, error) {
	return c.authorizer.DecisionTree(c.wrapContext(ctx), in, opts...)
}

func (c *contextualAuthorizerClient) Is(
	ctx context.Context,
	in *authz.IsRequest,
	opts ...grpc.CallOption,
) (*authz.IsResponse, error) {
	return c.authorizer.Is(c.wrapContext(ctx), in, opts...)
}

func (c *contextualAuthorizerClient) Query(
	ctx context.Context,
	in *authz.QueryRequest,
	opts ...grpc.CallOption,
) (*authz.QueryResponse, error) {
	return c.authorizer.Query(c.wrapContext(ctx), in, opts...)
}

func (c *contextualAuthorizerClient) wrapContext(ctx context.Context) context.Context {
	for _, wrapper := range c.wrappers {
		ctx = wrapper.WithContext(ctx)
	}

	return ctx
}
