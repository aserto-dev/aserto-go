package authorizer

import (
	"context"

	"github.com/aserto-dev/aserto-go/v2/client"
	authorizer "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto authorization services.
type Client struct {
	conn *client.Connection

	// Authorizer provides methods for performing authorization requests.
	Authorizer authorizer.AuthorizerClient
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
	}, err
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn.Conn
}
