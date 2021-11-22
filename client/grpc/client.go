package grpc

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	dir "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"

	"github.com/pkg/errors"
)

// Client provides access to Aserto administrative services.
type Client struct {
	conn *Connection

	// Directory provides methods for interacting with the Aserto user directory.
	// Use the Directory client to manage users, application, and roles.
	Directory dir.DirectoryClient

	// Policy provides read-only methods for listing and retrieving authorization policies defined in an Aserto account.
	Policy policy.PolicyClient

	// Info provides read-only access to system information and configuration.
	Info info.InfoClient
}

// NewClient creates a Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	connection, err := NewConnection(ctx, opts...)
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
