package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/pkg/grpcc"

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
func New(ctx context.Context, opts ...grpcc.ConnectionOption) (*Client, error) {
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
