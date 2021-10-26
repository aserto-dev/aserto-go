package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go"
	"github.com/aserto-dev/aserto-go/pkg/grpcc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"
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
func New(ctx context.Context, opts ...aserto.ConnectionOption) (*Client, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       connection,
		Authorizer: authz.NewAuthorizerClient(connection.Conn),
		Directory:  dir.NewDirectoryClient(connection.Conn),
		Policy:     policy.NewPolicyClient(connection.Conn),
		Info:       info.NewInfoClient(connection.Conn),
	}, err
}
