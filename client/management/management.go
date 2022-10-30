package management

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	management "github.com/aserto-dev/go-grpc/aserto/management/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto Management services.
type Client struct {
	conn *client.Connection

	// Controller provides methods for interacting with the controller service.
	Controller management.ControllerClient

	// ControlPlane provides methods for interacting with the ControlPlane service.
	ControlPlane management.ControlPlaneClient
}

// NewClient creates a Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	connection, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:         connection,
		Controller:   management.NewControllerClient(connection.Conn),
		ControlPlane: management.NewControlPlaneClient(connection.Conn),
	}, err
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn.Conn
}
