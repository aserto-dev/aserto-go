package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	authorizer "github.com/aserto-dev/go-grpc-authz/aserto/authorizer/authorizer/v1"
	directory "github.com/aserto-dev/go-grpc/aserto/authorizer/directory/v1"
	policy "github.com/aserto-dev/go-grpc/aserto/authorizer/policy/v1"
	info "github.com/aserto-dev/go-grpc/aserto/common/info/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto authorization services.
type Client struct {
	conn *client.Connection

	// Authorizer provides methods for performing authorization requests.
	Authorizer authorizer.AuthorizerClient

	// Directory provides methods for interacting with the Aserto user directory.
	// Use the Directory client to manage users, application, and roles.
	Directory directory.DirectoryClient

	// Policy provides read-only methods for listing and retrieving authorization policies defined in an Aserto account.
	Policy policy.PolicyClient

	// Info provides read-only access to system information and configuration.
	Info info.InfoClient
}

// NewClient creates a Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	connection, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:       connection,
		Authorizer: authorizer.NewAuthorizerClient(connection.Conn),
		Directory:  directory.NewDirectoryClient(connection.Conn),
		Policy:     policy.NewPolicyClient(connection.Conn),
		Info:       info.NewInfoClient(connection.Conn),
	}, err
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn.Conn
}
