package registrytenant

import (
	"context"

	"github.com/aserto-dev/aserto-go/client"
	registry_tenant "github.com/aserto-dev/go-grpc/aserto/registry_tenant/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto registry tenant services.
type Client struct {
	conn *client.Connection

	// Tenant provides methods for interacting with the tenant service.
	Tenant registry_tenant.TenantClient

	// Policy provides methods for interacting with the policy service.
	Policy registry_tenant.PolicyClient

	// PolicyRepo provides methods for interacting with the policy repository service.
	PolicyRepo registry_tenant.PolicyRepoClient
}

// NewClient creates a Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	connection, err := client.NewConnection(ctx, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:       connection,
		Tenant:     registry_tenant.NewTenantClient(connection.Conn),
		Policy:     registry_tenant.NewPolicyClient(connection.Conn),
		PolicyRepo: registry_tenant.NewPolicyRepoClient(connection.Conn),
	}, err
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn.Conn
}
