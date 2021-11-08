package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/config"
	"github.com/aserto-dev/aserto-go/internal/grpcc"

	authz "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
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
func NewClient(ctx context.Context, opts ...config.ConnectionOption) (*Client, error) {
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

// NewAuthorizerClient creates a new AuthorizerClient using the specified connection options.
func NewAuthorizerClient(ctx context.Context, opts ...config.ConnectionOption) (authz.AuthorizerClient, error) {
	connection, err := grpcc.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return authz.NewAuthorizerClient(connection.Conn), nil
}
