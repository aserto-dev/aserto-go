package registry

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	registry "github.com/aserto-dev/go-grpc/aserto/registry/v1"
	"google.golang.org/grpc"

	"github.com/pkg/errors"
)

// Client provides access to the Aserto registry services.
type Client struct {
	conn *client.Connection

	// Registry provides methods for interacting with the registry service.
	Registry registry.RegistryClient
}

// NewClient creates a Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	connection, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:     connection,
		Registry: registry.NewRegistryClient(connection.Conn),
	}, err
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn.Conn
}
